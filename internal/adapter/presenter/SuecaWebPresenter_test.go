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

func setupSuecaWebMock() *interfaces.MockSuecaGame {
	m := new(interfaces.MockSuecaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SuecaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTeamGamePoints").Return([domain.SuecaTeamCnt]int{0, 0})
	m.On("GetRoundCardPoints").Return([domain.SuecaTeamCnt]int{0, 0})
	m.On("GetRoundWinnerTeam").Return(-1)
	m.On("GetRoundGamePoints").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultSuecaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSuecaWebMockWithPlayers() (*interfaces.MockSuecaGame, []*domain.SuecaPlayer) {
	m := setupSuecaWebMock()
	players := makeSuecaPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestSuecaWebPresenter_Output(t *testing.T) {
	p := new(presenter.SuecaWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupSuecaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, int(domain.SuecaPhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, "sueca.playPhase.lead", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.SuecaCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 4, resObj.Config.TargetGamePoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.SuecaPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "sueca.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SuecaPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sueca.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SuecaPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sueca.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human, team 0; winner = 0
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "sueca.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		// player 0 is human, team 0; winner = team 1 (CPU)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sueca.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "1"}, resObj.MessageParams)
	})

	t.Run("team game points propagated to players", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamGamePoints")
		m.On("GetTeamGamePoints").Return([domain.SuecaTeamCnt]int{3, 1})
		result := p.Output(m, nil)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		// player 0 is team 0, game points 3
		assert.Equal(t, 3, resObj.Players[0].TeamGamePoints)
		// player 1 is team 1, game points 1
		assert.Equal(t, 1, resObj.Players[1].TeamGamePoints)
	})
}

func TestSuecaWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SuecaWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		m.On("GetHint").Return(&domain.SuecaHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSuecaWebMockWithPlayers()
		m.On("GetHint").Return((*domain.SuecaHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.SuecaWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestSuecaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SuecaWebPresenter)
	m := new(interfaces.MockSuecaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}
