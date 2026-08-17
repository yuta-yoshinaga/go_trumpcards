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

// SetteEMezzoSeat/Hand keep unexported fields, so the only way to build one from
// outside the domain package is through its own codec -- the same path the
// Worker takes out of KV.
func semSeatFromJSON(src string) *domain.SetteEMezzoSeat {
	s := new(domain.SetteEMezzoSeat)
	if err := json.Unmarshal([]byte(src), s); err != nil {
		panic(err)
	}
	return s
}

func semHandFromJSON(src string) *domain.SetteEMezzoHand {
	h := new(domain.SetteEMezzoHand)
	if err := json.Unmarshal([]byte(src), h); err != nil {
		panic(err)
	}
	return h
}

func setupSemWebMockDefaults(g *interfaces.MockSetteEMezzoGame) {
	g.On("GetPhase").Return(domain.SetteEMezzoPhasePlayerTurn).Maybe()
	g.On("GetChips").Return(900).Maybe()
	g.On("GetBankerIdx").Return(1).Maybe()
	g.On("IsHumanBanker").Return(false).Maybe()
	g.On("GetActiveSeat").Return(0).Maybe()
	g.On("GetNextBanker").Return(-1).Maybe()
	g.On("GetLastResult").Return("").Maybe()
	g.On("GetGameEndFlag").Return(false).Maybe()
	g.On("CanHit").Return(true).Maybe()
	g.On("CanStand").Return(true).Maybe()
	g.On("CanSetMatta").Return(false).Maybe()
	g.On("GetHandHalves", mock.Anything).Return(9).Maybe()
	g.On("FormatHalves", mock.Anything).Return("4.5").Maybe()

	g.On("GetSeats").Return([]*domain.SetteEMezzoSeat{
		semSeatFromJSON(`{"nm":"あなた","cp":false,"hd":{"cd":[{"d":1,"v":4,"f":true}],"bt":100}}`),
		semSeatFromJSON(`{"nm":"CPU1","cp":true}`),
		semSeatFromJSON(`{"nm":"CPU2","cp":true,"hd":{"cd":[{"d":3,"v":5,"f":true}],"bt":20}}`),
	}).Maybe()
	g.On("GetBankerHand").Return(
		semHandFromJSON(`{"cd":[{"d":2,"v":5,"f":true}]}`)).Maybe()
}

func parseSemOutput(t *testing.T, jsonStr string) *controller.SetteEMezzoWebOutput {
	t.Helper()
	var out controller.SetteEMezzoWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

func TestSetteEMezzoWebPresenter_Output(t *testing.T) {
	t.Run("player turn", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
		assert.Equal(t, domain.SetteEMezzoPhasePlayerTurn, result.Phase)
		assert.Equal(t, 900, result.Chips)
		assert.Len(t, result.Seats, 3)
		assert.Equal(t, -1, result.NextBanker)
		assert.Equal(t, "settemezzo.yourTurn", result.MessageCode)
	})

	// The target rides on the wire in halves so the client compares integers
	// rather than reconstructing 7.5 from a label.
	t.Run("the target is sent in halves", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
		assert.Equal(t, domain.SetteEMezzoTargetHalves, result.TargetHalves)
		assert.Equal(t, 15, result.TargetHalves)
	})

	// A hidden hand must not reach the wire -- the same hole Pontoon had before
	// #4485. Rendering backs on the page is not enough.
	t.Run("a hidden hand carries no cards or total", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))

		require.NotNil(t, result.BankerHand)
		assert.True(t, result.BankerHand.Hidden)
		assert.Zero(t, result.BankerHand.TotalHalves)
		assert.Empty(t, result.BankerHand.TotalLabel)
		for i, c := range result.BankerHand.Cards {
			assert.Nil(t, c, "banker card %d leaked", i)
		}
		// 枚数だけは残る。引いた枚数は卓上で見えている情報。
		assert.Len(t, result.BankerHand.Cards, 1)

		cpu := result.Seats[2].Hand
		require.NotNil(t, cpu)
		assert.True(t, cpu.Hidden)
		for i, c := range cpu.Cards {
			assert.Nil(t, c, "cpu card %d leaked", i)
		}
	})

	t.Run("the human's own hand is never hidden", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
		own := result.Seats[0].Hand
		require.NotNil(t, own)
		assert.False(t, own.Hidden)
		assert.NotNil(t, own.Cards[0])
		assert.Equal(t, 9, own.TotalHalves)
		assert.Equal(t, "4.5", own.TotalLabel)
	})

	t.Run("a settled round reveals every hand", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetGameEndFlag").Return(true)
		g.On("GetPhase").Return(domain.SetteEMezzoPhaseEnd)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
		assert.False(t, result.BankerHand.Hidden)
		assert.NotNil(t, result.BankerHand.Cards[0])
		assert.False(t, result.Seats[2].Hand.Hidden)
	})

	t.Run("the human banker sees their own hand", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
		g.On("IsHumanBanker").Return(true)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
		assert.False(t, result.BankerHand.Hidden)
		assert.NotNil(t, result.BankerHand.Cards[0])
	})

	t.Run("a seat with no hand sends none", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
		assert.Nil(t, result.Seats[1].Hand)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name        string
		phase       int
		humanBanker bool
		code        string
	}{
		{"betting", domain.SetteEMezzoPhaseBet, false, "settemezzo.placeBet"},
		{"betting while banking", domain.SetteEMezzoPhaseBet, true, "settemezzo.dealAsBanker"},
		{"banker turn", domain.SetteEMezzoPhaseBankerTurn, true, "settemezzo.bankerTurn"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockSetteEMezzoGame)
			setupSemWebMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsHumanBanker")
			g.On("GetPhase").Return(tc.phase)
			g.On("IsHumanBanker").Return(tc.humanBanker)

			result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}

	t.Run("round over", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetLastResult")
		g.On("GetPhase").Return(domain.SetteEMezzoPhaseEnd)
		g.On("GetLastResult").Return("親は 6.5")

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
		assert.Equal(t, "settemezzo.roundOver", result.MessageCode)
		assert.Equal(t, "親は 6.5", result.Message)
	})

	// Landing exactly on 7.5 is the only way the bank moves, so it gets its own
	// message rather than being buried in the summary.
	t.Run("the bank passing has its own message", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		setupSemWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetNextBanker")
		g.On("GetPhase").Return(domain.SetteEMezzoPhaseEnd)
		g.On("GetNextBanker").Return(0)

		result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
		assert.Equal(t, "settemezzo.bankPasses", result.MessageCode)
	})
}

func TestSetteEMezzoWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("mid-round returns an empty log", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		g.On("GetGameEndFlag").Return(false)
		assert.Contains(t, new(SetteEMezzoWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("a settled round returns the log", func(t *testing.T) {
		g := new(interfaces.MockSetteEMezzoGame)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "deal", Detail: "test"},
		})
		assert.Contains(t, new(SetteEMezzoWebPresenter).ActionLogOutput(g), "deal")
	})
}

// #5566: 停止ラインの数字はドメインの定数から来ること。訳文にも TS にも
// 焼き込まないので、ここが落ちるとクライアントは「0 点で止まる」と書く。
func TestSetteEMezzoWebPresenter_CarriesTheCpuStandThreshold(t *testing.T) {
	g := new(interfaces.MockSetteEMezzoGame)
	setupSemWebMockDefaults(g)

	result := parseSemOutput(t, new(SetteEMezzoWebPresenter).Output(g, nil))
	assert.Equal(t, domain.SetteEMezzoCpuStandHalves, result.CpuStandHalves)
	assert.NotZero(t, result.CpuStandHalves)
	// 目標点とは別の数字であること。同じなら CPU は一度も引かない。
	assert.NotEqual(t, result.TargetHalves, result.CpuStandHalves)
}
