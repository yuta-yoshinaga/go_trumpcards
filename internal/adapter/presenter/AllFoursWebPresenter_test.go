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

func makeAllFoursPlayers() []*domain.AllFoursPlayer {
	return []*domain.AllFoursPlayer{
		domain.NewAllFoursPlayer(true),
		domain.NewAllFoursPlayer(false),
	}
}

func setupAllFoursWebMock() *interfaces.MockAllFoursGame {
	m := new(interfaces.MockAllFoursGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetDealerIdx").Return(1)
	m.On("GetNonDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetTurnUp").Return((*domain.Card)(nil))
	m.On("GetRunCount").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.AllFoursTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.AllFoursPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultAllFoursConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1})
	return m
}

func setupAllFoursWebMockWithPlayers() (*interfaces.MockAllFoursGame, []*domain.AllFoursPlayer) {
	m := setupAllFoursWebMock()
	players := makeAllFoursPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestAllFoursWebPresenter_Output(t *testing.T) {
	p := new(presenter.AllFoursWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupAllFoursWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		var resObj controller.AllFoursWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 2, len(resObj.Players))
		assert.Equal(t, int(domain.AllFoursPhasePlay), resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.DealerIdx)
		assert.Equal(t, 0, resObj.NonDealerIdx)
		assert.Equal(t, []int{0, 1}, resObj.ValidPlayIndices)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupAllFoursWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.Players[0].IsHuman)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.False(t, resObj.Players[1].IsHuman)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 1, resObj.Players[1].CardCount)
	})

	t.Run("error message included", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		result := p.Output(m, errors.New("invalid play"))
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "invalid play", resObj.Message)
	})

	t.Run("game end shows winner code", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "allfours.result.humanWin", resObj.MessageCode)
	})
}

func TestAllFoursWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.AllFoursWebPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		m.On("GetHint").Return((*domain.AllFoursHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})

	t.Run("with beg hint", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		beg := true
		m.On("GetHint").Return(&domain.AllFoursHint{Beg: &beg, Reason: "beg_beg"})
		result := p.HintOutput(m)
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, "beg_beg", resObj.Hint.Reason)
		assert.True(t, *resObj.Hint.Beg)
	})
}

func TestAllFoursWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.AllFoursWebPresenter)
	m := new(interfaces.MockAllFoursGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "stand", Detail: "You stand"},
	})
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "You stand")
}
