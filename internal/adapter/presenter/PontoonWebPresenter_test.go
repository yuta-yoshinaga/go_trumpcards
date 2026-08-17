//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupPontoonWebMockDefaults(g *interfaces.MockPontoonGame) {
	g.On("GetPhase").Return(domain.PontoonPhasePlayerTurn).Maybe()
	g.On("GetChips").Return(900).Maybe()
	g.On("GetBankerIdx").Return(1).Maybe()
	g.On("IsHumanBanker").Return(false).Maybe()
	g.On("GetActiveSeat").Return(0).Maybe()
	g.On("GetActiveHand").Return(0).Maybe()
	g.On("GetNextBanker").Return(-1).Maybe()
	g.On("GetLastResult").Return("").Maybe()
	g.On("GetGameEndFlag").Return(false).Maybe()
	g.On("CanStick").Return(true).Maybe()
	g.On("CanTwist").Return(true).Maybe()
	g.On("CanBuy").Return(false).Maybe()
	g.On("CanSplit").Return(false).Maybe()
	g.On("GetHandTotal", mock.Anything).Return(18).Maybe()
	g.On("GetHandRank", mock.Anything).Return(domain.PontoonRankPoints).Maybe()

	g.On("GetSeats").Return([]*domain.PontoonSeat{
		pontoonSeatFromJSON(`{"nm":"あなた","cp":false,"hd":[{"cd":[{"d":1,"v":10,"f":true},{"d":2,"v":8,"f":true}],"bt":100}]}`),
		pontoonSeatFromJSON(`{"nm":"CPU1","cp":true}`),
		pontoonSeatFromJSON(`{"nm":"CPU2","cp":true,"hd":[{"cd":[{"d":3,"v":9,"f":true},{"d":4,"v":7,"f":true}],"bt":20}]}`),
	}).Maybe()
	g.On("GetBankerHand").Return(
		pontoonHandFromJSON(`{"cd":[{"d":1,"v":10,"f":true},{"d":2,"v":9,"f":true}]}`)).Maybe()
}

// PontoonSeat and PontoonHand keep unexported fields, so the only way to build
// one from outside the domain package is through its own codec. That is also
// what the Worker does, so the fixture exercises the real path.
func pontoonSeatFromJSON(src string) *domain.PontoonSeat {
	s := new(domain.PontoonSeat)
	if err := json.Unmarshal([]byte(src), s); err != nil {
		panic(err)
	}
	return s
}

func pontoonHandFromJSON(src string) *domain.PontoonHand {
	h := new(domain.PontoonHand)
	if err := json.Unmarshal([]byte(src), h); err != nil {
		panic(err)
	}
	return h
}

func parsePontoonOutput(t *testing.T, jsonStr string) *controller.PontoonWebOutput {
	t.Helper()
	var out controller.PontoonWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

func TestPontoonWebPresenter_Output(t *testing.T) {
	t.Run("player turn", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		assert.Equal(t, domain.PontoonPhasePlayerTurn, result.Phase)
		assert.Equal(t, 900, result.Chips)
		assert.Len(t, result.Seats, 3)
		assert.Equal(t, 1, result.BankerIdx)
		assert.False(t, result.IsHumanBanker)
		assert.Equal(t, -1, result.NextBanker)
		assert.Equal(t, "pontoon.yourTurn", result.MessageCode)
	})

	// The rules are involved enough that the server decides what is legal --
	// the client must not re-derive "cannot stick below 15" or "no buy after a
	// twist".
	t.Run("the legal actions ride on the wire", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		assert.True(t, result.CanStick)
		assert.True(t, result.CanTwist)
		assert.False(t, result.CanBuy)
		assert.False(t, result.CanSplit)
	})

	// A hidden hand must not reach the wire at all. Rendering backs on the page
	// is not enough: anyone reading the response would see the banker's hole
	// cards, which is the one thing this game is built on not being visible.
	t.Run("a hidden hand carries no cards, total or rank", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))

		require.NotNil(t, result.BankerHand)
		assert.True(t, result.BankerHand.Hidden)
		assert.Zero(t, result.BankerHand.Total)
		assert.Zero(t, result.BankerHand.Rank)
		for i, c := range result.BankerHand.Cards {
			assert.Nil(t, c, "banker card %d leaked", i)
		}
		// The COUNT survives -- Twist and Buy change the hand size in view of
		// the table, and the page needs it to draw the backs.
		assert.Len(t, result.BankerHand.Cards, 2)

		cpu := result.Seats[2].Hands[0]
		assert.True(t, cpu.Hidden)
		for i, c := range cpu.Cards {
			assert.Nil(t, c, "cpu card %d leaked", i)
		}
	})

	// The human's own hand is theirs to see, at every phase.
	t.Run("the human's own hand is never hidden", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		own := result.Seats[0].Hands[0]
		assert.False(t, own.Hidden)
		assert.NotNil(t, own.Cards[0])
		assert.Equal(t, 18, own.Total)
	})

	t.Run("a settled round reveals every hand", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetGameEndFlag").Return(true)
		g.On("GetPhase").Return(domain.PontoonPhaseEnd)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		assert.False(t, result.BankerHand.Hidden)
		assert.NotNil(t, result.BankerHand.Cards[0])
		assert.False(t, result.Seats[2].Hands[0].Hidden)
	})

	// When the human banks, the banker's hand IS the human's hand.
	t.Run("the human banker sees their own hand", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
		g.On("IsHumanBanker").Return(true)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		assert.False(t, result.BankerHand.Hidden)
		assert.NotNil(t, result.BankerHand.Cards[0])
	})

	// The total and the rank are computed server-side so the client never has to
	// re-implement the ace's 1/11 or the five-card rule.
	t.Run("each hand carries its total and rank", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		hand := result.Seats[0].Hands[0]
		assert.Equal(t, 18, hand.Total)
		assert.Equal(t, int(domain.PontoonRankPoints), hand.Rank)
		assert.Equal(t, 100, hand.Bet)
		assert.Len(t, hand.Cards, 2)
	})

	t.Run("a seat with no hand renders as an empty list", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		assert.NotNil(t, result.Seats[1].Hands)
		assert.Empty(t, result.Seats[1].Hands)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name        string
		phase       int
		humanBanker bool
		code        string
	}{
		{"betting", domain.PontoonPhaseBet, false, "pontoon.placeBet"},
		{"betting while banking", domain.PontoonPhaseBet, true, "pontoon.dealAsBanker"},
		{"banker turn", domain.PontoonPhaseBankerTurn, true, "pontoon.bankerTurn"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockPontoonGame)
			setupPontoonWebMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
			g.On("GetPhase").Return(tc.phase)
			g.On("IsHumanBanker").Return(tc.humanBanker)

			result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}

	t.Run("round over", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetLastResult")
		g.On("GetPhase").Return(domain.PontoonPhaseEnd)
		g.On("GetLastResult").Return("親は 19")

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		assert.Equal(t, "pontoon.roundOver", result.MessageCode)
		assert.Equal(t, "親は 19", result.Message)
	})

	// The bank changing hands is the headline event of the round, so it gets its
	// own message rather than being buried in the summary.
	t.Run("the bank passing has its own message", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		setupPontoonWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetNextBanker")
		g.On("GetPhase").Return(domain.PontoonPhaseEnd)
		g.On("GetNextBanker").Return(0)

		result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
		assert.Equal(t, "pontoon.bankPasses", result.MessageCode)
		assert.Equal(t, "0", result.MessageParams["seat"])
	})
}

func TestPontoonWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("mid-round returns an empty log", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		g.On("GetGameEndFlag").Return(false)
		assert.Contains(t, new(PontoonWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("a settled round returns the log", func(t *testing.T) {
		g := new(interfaces.MockPontoonGame)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "deal", Detail: "test"},
		})
		assert.Contains(t, new(PontoonWebPresenter).ActionLogOutput(g), "deal")
	})
}

// #5565: 停止ラインの数字はドメインの定数から来ること。訳文にも TS にも
// 焼き込まないので、ここが落ちるとクライアントは「0 以上で止まる」と書く。
func TestPontoonWebPresenter_CarriesBothStickThresholds(t *testing.T) {
	g := new(interfaces.MockPontoonGame)
	setupPontoonWebMockDefaults(g)

	result := parsePontoonOutput(t, new(PontoonWebPresenter).Output(g, nil))
	assert.Equal(t, domain.PontoonStickMin, result.StickMin)
	assert.Equal(t, domain.PontoonCpuStickMin, result.CpuStickMin)
	// ゼロ値と区別が付かない検査にしない。
	assert.NotZero(t, result.CpuStickMin)
}
