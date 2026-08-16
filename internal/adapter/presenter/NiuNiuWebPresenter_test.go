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

// NiuNiuSeat/Hand keep unexported fields, so the only way to build one from
// outside the domain package is through its own codec -- the same path the
// Worker takes out of KV.
func nnSeatFromJSON(src string) *domain.NiuNiuSeat {
	s := new(domain.NiuNiuSeat)
	if err := json.Unmarshal([]byte(src), s); err != nil {
		panic(err)
	}
	return s
}

func nnHandFromJSON(src string) *domain.NiuNiuHand {
	h := new(domain.NiuNiuHand)
	if err := json.Unmarshal([]byte(src), h); err != nil {
		panic(err)
	}
	return h
}

const nnFiveCards = `[{"d":1,"v":10,"f":true},{"d":2,"v":10,"f":true},{"d":3,"v":10,"f":true},{"d":4,"v":5,"f":true},{"d":1,"v":5,"f":true}]`

func setupNiuNiuWebMockDefaults(g *interfaces.MockNiuNiuGame) {
	g.On("GetPhase").Return(domain.NiuNiuPhaseBet).Maybe()
	g.On("GetChips").Return(900).Maybe()
	g.On("GetMaxMultiplier").Return(domain.NiuNiuMaxMultiplier).Maybe()
	g.On("GetBankerIdx").Return(3).Maybe()
	g.On("GetBankerRankKey").Return("none").Maybe()
	g.On("GetGameEndFlag").Return(false).Maybe()
	g.On("GetMultiplier", mock.Anything).Return(3).Maybe()

	g.On("GetSeats").Return([]*domain.NiuNiuSeat{
		nnSeatFromJSON(`{"nm":"あなた","cp":false,"hd":{"cd":` + nnFiveCards + `,"bt":100,"ci":[0,1,2],"rk":10}}`),
		nnSeatFromJSON(`{"nm":"CPU1","cp":true,"hd":{"cd":` + nnFiveCards + `,"bt":20,"ci":[0,1,2],"rk":10}}`),
		nnSeatFromJSON(`{"nm":"CPU2","cp":true}`),
		nnSeatFromJSON(`{"nm":"親","cp":true}`),
	}).Maybe()
	g.On("GetBankerHand").Return(
		nnHandFromJSON(`{"cd":` + nnFiveCards + `,"ci":[0,1,2],"rk":10}`)).Maybe()
}

func parseNiuNiuOutput(t *testing.T, jsonStr string) *controller.NiuNiuWebOutput {
	t.Helper()
	var out controller.NiuNiuWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

func TestNiuNiuWebPresenter_Output(t *testing.T) {
	t.Run("betting phase", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))
		assert.Equal(t, domain.NiuNiuPhaseBet, result.Phase)
		assert.Equal(t, 900, result.Chips)
		assert.Len(t, result.Seats, 4)
		assert.Equal(t, 3, result.BankerIdx)
		assert.Equal(t, "niuniu.placeBet", result.MessageCode)
	})

	// The client caps its stake buttons at chips / maxMultiplier, so the figure
	// has to come from the server rather than being hardcoded twice.
	t.Run("the worst-case multiplier rides the wire", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))
		assert.Equal(t, domain.NiuNiuMaxMultiplier, result.MaxMultiplier)
		assert.Equal(t, 3, result.MaxMultiplier)
	})

	// A hidden hand must not reach the wire.
	//
	// This drives the presenter through a state the DOMAIN cannot currently
	// produce -- `deal` settles in the same call, so a hand never coexists with
	// an unfinished round. The mock builds it anyway, because the guard exists
	// so the presenter does not depend on that invariant holding forever.
	t.Run("a hidden hand carries no cards, rank or combo", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))

		require.NotNil(t, result.BankerHand)
		assert.True(t, result.BankerHand.Hidden)
		assert.Zero(t, result.BankerHand.Rank)
		assert.Empty(t, result.BankerHand.RankKey)
		assert.Empty(t, result.BankerHand.ComboIdx)
		for i, c := range result.BankerHand.Cards {
			assert.Nil(t, c, "banker card %d leaked", i)
		}
		// 枚数だけは残る。
		assert.Len(t, result.BankerHand.Cards, domain.NiuNiuHandSize)

		cpu := result.Seats[1].Hand
		require.NotNil(t, cpu)
		assert.True(t, cpu.Hidden)
		for i, c := range cpu.Cards {
			assert.Nil(t, c, "cpu card %d leaked", i)
		}
	})

	t.Run("the human's own hand is never hidden", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))
		own := result.Seats[0].Hand
		require.NotNil(t, own)
		assert.False(t, own.Hidden)
		assert.NotNil(t, own.Cards[0])
		assert.Equal(t, 10, own.Rank)
		assert.Equal(t, "niuniu", own.RankKey)
		assert.Equal(t, 3, own.Multiplier)
	})

	// The three cards that made the bull ride along so the client never has to
	// redo the combination search to highlight them.
	t.Run("a revealed hand carries the winning combo", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))
		assert.Equal(t, []int{0, 1, 2}, result.Seats[0].Hand.ComboIdx)
	})

	t.Run("a settled round reveals every hand", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetGameEndFlag").Return(true)
		g.On("GetPhase").Return(domain.NiuNiuPhaseEnd)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBankerRankKey")
		g.On("GetBankerRankKey").Return("niuniu")

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))
		assert.False(t, result.BankerHand.Hidden)
		assert.NotNil(t, result.BankerHand.Cards[0])
		assert.False(t, result.Seats[1].Hand.Hidden)
		assert.Equal(t, "niuniu.roundOverNiuNiu", result.MessageCode)
	})

	// **完成済みの日本語を params に載せない。** 以前は "親: 牛牛" をそのまま
	// messageParams["result"] に入れており、英語ロケールでもそれが出ていた (#5567)。
	t.Run("the round-over message carries no pre-formatted label", func(t *testing.T) {
		for _, c := range []struct {
			rankKey  string
			wantCode string
			wantN    string
		}{
			{"none", "niuniu.roundOverNone", ""},
			{"niuniu", "niuniu.roundOverNiuNiu", ""},
			{"n7", "niuniu.roundOverN", "7"},
			// 役が確定していないキーでは何も送らない。default に流すと n が空の
			// まま roundOverN が出て、画面に「親: 牛」が残る。
			{"", "", ""},
			{"n0", "", ""},
			{"bogus", "", ""},
		} {
			g := new(interfaces.MockNiuNiuGame)
			setupNiuNiuWebMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBankerRankKey")
			g.On("GetGameEndFlag").Return(true)
			g.On("GetPhase").Return(domain.NiuNiuPhaseEnd)
			g.On("GetBankerRankKey").Return(c.rankKey)

			result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))
			assert.Equal(t, c.wantCode, result.MessageCode, "rank %s", c.rankKey)
			if c.wantN == "" {
				assert.Empty(t, result.MessageParams, "数字の無い格は params を持たない")
			} else {
				assert.Equal(t, c.wantN, result.MessageParams["n"])
			}
			// 日本語が params に混ざっていたら、それはロケールを無視する経路。
			for k, v := range result.MessageParams {
				for _, r := range v {
					assert.Less(t, r, rune(128), "messageParams[%s] に非ASCII: %q", k, v)
				}
			}
		}
	})

	t.Run("a seat with no hand sends none", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))
		assert.Nil(t, result.Seats[2].Hand)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

func TestNiuNiuWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("mid-round returns an empty log", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		g.On("GetGameEndFlag").Return(false)
		assert.Contains(t, new(NiuNiuWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("a settled round returns the log", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "deal", Detail: "test"},
		})
		assert.Contains(t, new(NiuNiuWebPresenter).ActionLogOutput(g), "deal")
	})
}
