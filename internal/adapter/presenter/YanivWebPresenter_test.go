//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeYanivPlayers() []*domain.YanivPlayer {
	return []*domain.YanivPlayer{
		domain.NewYanivPlayer(true),
		domain.NewYanivPlayer(false),
		domain.NewYanivPlayer(false),
		domain.NewYanivPlayer(false),
	}
}

func setupYanivWebMock() (*interfaces.MockYanivGame, []*domain.YanivPlayer) {
	m := new(interfaces.MockYanivGame)
	players := makeYanivPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(39)
	m.On("GetPickupCards").Return([]*domain.Card{})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.YanivPhaseDiscard)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultYanivConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetCallerIdx").Return(-1)
	m.On("GetAsafWinnerIdx").Return(-1)
	m.On("GetIsAsaf").Return(false)
	m.On("GetRoundScores").Return([]int{})
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestYanivWebPresenter_Output(t *testing.T) {
	p := new(presenter.YanivWebPresenter)

	t.Run("discard phase hides CPU cards", func(t *testing.T) {
		m, players := setupYanivWebMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		var resObj controller.YanivWebOutput
		require := json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.NoError(t, require)
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, 39, resObj.DrawPileCount)
		assert.Equal(t, "yaniv.discardPhase", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1) // human shown
		assert.Len(t, resObj.Players[1].Cards, 0) // CPU hidden
		assert.Equal(t, 5, resObj.Players[0].HandTotal)
		assert.Equal(t, 0, resObj.Players[1].HandTotal) // hidden total
	})

	t.Run("draw phase message", func(t *testing.T) {
		m, _ := setupYanivWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.YanivPhaseDraw)
		var resObj controller.YanivWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "yaniv.drawPhase", resObj.MessageCode)
	})

	t.Run("pickup cards serialized", func(t *testing.T) {
		m, _ := setupYanivWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickupCards")
		m.On("GetPickupCards").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 4, false),
			domain.NewCard(domain.CardDesignSpade, 6, false),
		})
		var resObj controller.YanivWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Len(t, resObj.PickupCards, 2)
		assert.Equal(t, 4, resObj.PickupCards[0].Value)
	})

	t.Run("round end yaniv reveal", func(t *testing.T) {
		m, players := setupYanivWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCallerIdx")
		m.On("GetPhase").Return(domain.YanivPhaseRoundEnd)
		m.On("GetCallerIdx").Return(0)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
		var resObj controller.YanivWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1) // revealed
		assert.Equal(t, 10, resObj.Players[1].HandTotal)
		assert.Equal(t, "yaniv.yanivResult", resObj.MessageCode)
	})

	t.Run("round end asaf message", func(t *testing.T) {
		m, _ := setupYanivWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsAsaf")
		m.On("GetPhase").Return(domain.YanivPhaseRoundEnd)
		m.On("GetIsAsaf").Return(true)
		var resObj controller.YanivWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "yaniv.asafResult", resObj.MessageCode)
	})

	t.Run("round end no contest", func(t *testing.T) {
		m, _ := setupYanivWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.YanivPhaseRoundEnd)
		var resObj controller.YanivWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "yaniv.roundEnd", resObj.MessageCode)
	})

	t.Run("game end has winner message", func(t *testing.T) {
		m, _ := setupYanivWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.YanivPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		var resObj controller.YanivWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, 0, resObj.WinnerIdx)
		assert.NotEmpty(t, resObj.MessageCode)
	})

	t.Run("eliminated flag surfaced", func(t *testing.T) {
		m, players := setupYanivWebMock()
		players[3].SetEliminated(true)
		players[3].SetScore(210)
		var resObj controller.YanivWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.True(t, resObj.Players[3].IsEliminated)
		assert.Equal(t, 210, resObj.Players[3].Score)
	})

	t.Run("error surfaced", func(t *testing.T) {
		m, _ := setupYanivWebMock()
		var resObj controller.YanivWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, errors.New("oops"))), &resObj)
		assert.Equal(t, "oops", resObj.Message)
	})
}

func TestYanivWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.YanivWebPresenter)
	m := new(interfaces.MockYanivGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "yaniv", Detail: "calls Yaniv"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)
	assert.Contains(t, p.ActionLogOutput(m), "yaniv")
}
