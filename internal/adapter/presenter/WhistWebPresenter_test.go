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

func setupWhistWebMock() *interfaces.MockWhistGame {
	m := new(interfaces.MockWhistGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.WhistPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultWhistConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeWhistPlayers() []*domain.WhistPlayer {
	return []*domain.WhistPlayer{
		domain.NewWhistPlayer(true, 0),
		domain.NewWhistPlayer(false, 1),
		domain.NewWhistPlayer(false, 0),
		domain.NewWhistPlayer(false, 1),
	}
}

func setupWhistWebMockWithPlayers() (*interfaces.MockWhistGame, []*domain.WhistPlayer) {
	m := setupWhistWebMock()
	players := makeWhistPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestWhistWebPresenter_Output(t *testing.T) {
	p := new(presenter.WhistWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupWhistWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.WhistWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, 0, resObj.Phase) // WhistPhasePlay
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.TrickNumber)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 0, resObj.LeadPlayerIdx)
		assert.Equal(t, "", resObj.Message)
		assert.Empty(t, resObj.CurrentTrick)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupWhistWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsHuman)
		assert.Equal(t, 1, humanPlayer.CardCount)
		assert.Len(t, humanPlayer.Cards, 1)
		assert.Equal(t, "SPADE", humanPlayer.Cards[0].Design)
		assert.Equal(t, 5, humanPlayer.Cards[0].Value)

		cpu1 := resObj.Players[1]
		assert.False(t, cpu1.IsHuman)
		assert.Equal(t, 1, cpu1.CardCount)
		assert.Len(t, cpu1.Cards, 0)
	})

	t.Run("player team info", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 0, resObj.Players[0].Team)
		assert.Equal(t, 1, resObj.Players[1].Team)
		assert.Equal(t, 0, resObj.Players[2].Team)
		assert.Equal(t, 1, resObj.Players[3].Team)
	})

	t.Run("team scores", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.On("GetTeamScore", 0).Return(3)
		m.On("GetTeamScore", 1).Return(2)

		result := p.Output(m, nil)
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, [2]int{3, 2}, resObj.TeamScores)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
	})

	t.Run("game end message", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Contains(t, resObj.Message, "チーム0")
		assert.Equal(t, "whist.result.team0Win", resObj.MessageCode)
	})

	t.Run("play phase lead message", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "whist.playPhase.lead", resObj.MessageCode)
	})

	t.Run("play phase follow message", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		}
		m.On("GetCurrentTrick").Return(trick)

		result := p.Output(m, nil)
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "whist.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WhistPhaseTrickEnd)

		result := p.Output(m, nil)
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "whist.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WhistPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.WhistWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "whist.roundEnd", resObj.MessageCode)
	})
}

func TestWhistWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.WhistWebPresenter)

	t.Run("with hint", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()
		cardIdx := 3
		m.On("GetHint").Return(&domain.WhistHint{CardIndex: &cardIdx, Reason: "follow_suit"})

		result := p.HintOutput(m)
		var resObj controller.WhistWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, 3, *resObj.Hint.CardIndex)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupWhistWebMockWithPlayers()
		m.On("GetHint").Return((*domain.WhistHint)(nil))

		result := p.HintOutput(m)
		var resObj controller.WhistWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.Hint)
	})
}

func TestWhistWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WhistWebPresenter)
	m := setupWhistWebMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "test"},
	})

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
