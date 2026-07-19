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

func setupGaigelWebMock() *interfaces.MockGaigelGame {
	m := new(interfaces.MockGaigelGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.GaigelTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GaigelPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetTrumpCard").Return((*domain.Card)(nil))
	m.On("GetStockRemaining").Return(27)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetRoundPoints", 0).Return(0)
	m.On("GetRoundPoints", 1).Return(0)
	m.On("GetRoundMarriagePoints", 0).Return(0)
	m.On("GetRoundMarriagePoints", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetMarriageIndices", 0).Return([]int(nil))
	m.On("GetConfig").Return(domain.DefaultGaigelConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupGaigelWebMockWithPlayers() (*interfaces.MockGaigelGame, []*domain.GaigelPlayer) {
	m := setupGaigelWebMock()
	players := makeGaigelPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestGaigelWebPresenter_Output(t *testing.T) {
	p := new(presenter.GaigelWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupGaigelWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))

		result := p.Output(m, nil)
		var resObj controller.GaigelWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, int(domain.GaigelPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 27, resObj.StockRemaining)
	})

	t.Run("turn-up trump card populated", func(t *testing.T) {
		m, players := setupGaigelWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpCard")
		m.On("GetTrumpCard").Return(domain.NewCard(domain.CardDesignHeart, 10, true))

		result := p.Output(m, nil)
		var resObj controller.GaigelWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.TrumpCard)
		assert.Equal(t, "HEART", resObj.TrumpCard.Design)
		assert.Equal(t, 10, resObj.TrumpCard.Value)
	})

	t.Run("turn-up trump card omitted once drawn", func(t *testing.T) {
		m, players := setupGaigelWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		// setupGaigelWebMock already returns a nil trump card.
		result := p.Output(m, nil)
		var resObj controller.GaigelWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.TrumpCard)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupGaigelWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.GaigelWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
	})

	t.Run("round end phase", func(t *testing.T) {
		m, _ := setupGaigelWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GaigelPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "gaigel.roundEnd")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupGaigelWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "gaigel.result.team0Win")
	})
}

func TestGaigelWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.GaigelWebPresenter)
	m, players := setupGaigelWebMockWithPlayers()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	idx := 0
	m.On("GetHint").Return(&domain.GaigelHint{CardIndex: &idx, Reason: "follow_win"})
	result := p.HintOutput(m)
	var resObj controller.GaigelWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
	assert.NotNil(t, resObj.Hint)
	assert.Equal(t, "follow_win", resObj.Hint.Reason)
}

func TestGaigelWebPresenter_HintOutput_Nil(t *testing.T) {
	p := new(presenter.GaigelWebPresenter)
	m, _ := setupGaigelWebMockWithPlayers()
	m.On("GetHint").Return((*domain.GaigelHint)(nil))
	result := p.HintOutput(m)
	var resObj controller.GaigelWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
	assert.Nil(t, resObj.Hint)
}

func TestGaigelWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GaigelWebPresenter)
	m := setupGaigelWebMock()
	assert.NotNil(t, p.ActionLogOutput(m))
}
