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
	g.On("GetBankerIdx").Return(3).Maybe()
	g.On("GetLastResult").Return("").Maybe()
	g.On("GetGameEndFlag").Return(false).Maybe()
	g.On("GetRankLabel", mock.Anything).Return("牛牛").Maybe()
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

	// A hidden hand must not reach the wire -- the same hole Pontoon had before
	// #4485. The banker's hand is the one that decides the round.
	t.Run("a hidden hand carries no cards, rank or combo", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuWebMockDefaults(g)

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))

		require.NotNil(t, result.BankerHand)
		assert.True(t, result.BankerHand.Hidden)
		assert.Zero(t, result.BankerHand.Rank)
		assert.Empty(t, result.BankerHand.RankLabel)
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
		assert.Equal(t, "牛牛", own.RankLabel)
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
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetLastResult")
		g.On("GetGameEndFlag").Return(true)
		g.On("GetPhase").Return(domain.NiuNiuPhaseEnd)
		g.On("GetLastResult").Return("親: 牛牛")

		result := parseNiuNiuOutput(t, new(NiuNiuWebPresenter).Output(g, nil))
		assert.False(t, result.BankerHand.Hidden)
		assert.NotNil(t, result.BankerHand.Cards[0])
		assert.False(t, result.Seats[1].Hand.Hidden)
		assert.Equal(t, "niuniu.roundOver", result.MessageCode)
		assert.Equal(t, "親: 牛牛", result.Message)
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
