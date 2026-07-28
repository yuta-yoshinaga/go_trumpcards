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

func setupKlaverjasWebMock() *interfaces.MockKlaverjasGame {
	m := new(interfaces.MockKlaverjasGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTeamScores").Return([domain.KlaverjasTeamCnt]int{0, 0})
	m.On("GetRoundCardPoints").Return([domain.KlaverjasTeamCnt]int{0, 0})
	m.On("GetRoundRoem").Return([domain.KlaverjasTeamCnt]int{0, 0})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultKlaverjasConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupKlaverjasWebMockWithPlayers() (*interfaces.MockKlaverjasGame, []*domain.KlaverjasPlayer) {
	m := setupKlaverjasWebMock()
	players := makeKlaverjasPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestKlaverjasWebPresenter_Output(t *testing.T) {
	p := new(presenter.KlaverjasWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupKlaverjasWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.KlaverjasPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, "klaverjas.playPhase.lead", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.KlaverjasCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 1501, resObj.Config.TargetPoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.KlaverjasPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "klaverjas.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.KlaverjasPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "klaverjas.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.KlaverjasPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "klaverjas.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human, team 0; winner = 0
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "klaverjas.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human, team 0; winner = team 1 (CPU)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "klaverjas.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "1"}, resObj.MessageParams)
	})

	t.Run("team scores propagated to players", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScores")
		m.On("GetTeamScores").Return([domain.KlaverjasTeamCnt]int{300, 150})
		result := p.Output(m, nil)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		// player 0 is team 0, score 300
		assert.Equal(t, 300, resObj.Players[0].TeamScore)
		// player 1 is team 1, score 150
		assert.Equal(t, 150, resObj.Players[1].TeamScore)
	})
}

func TestKlaverjasWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.KlaverjasWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		m.On("GetHint").Return(&domain.KlaverjasHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupKlaverjasWebMockWithPlayers()
		m.On("GetHint").Return((*domain.KlaverjasHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.KlaverjasWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestKlaverjasWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KlaverjasWebPresenter)
	m := new(interfaces.MockKlaverjasGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}
