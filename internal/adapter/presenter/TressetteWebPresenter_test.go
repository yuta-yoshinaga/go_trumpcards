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

func makeTressettePlayers() []*domain.TressettePlayer {
	return []*domain.TressettePlayer{
		domain.NewTressettePlayer(true),
		domain.NewTressettePlayer(false),
		domain.NewTressettePlayer(false),
		domain.NewTressettePlayer(false),
	}
}

func setupTressetteWebMock() *interfaces.MockTressetteGame {
	m := new(interfaces.MockTressetteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TressetteTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TressettePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetTeamScores").Return([domain.TressetteTeamCnt]int{0, 0})
	m.On("GetTeamRoundThirds").Return([domain.TressetteTeamCnt]int{0, 0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("GetConfig").Return(domain.DefaultTressetteConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupTressetteWebMockWithPlayers() (*interfaces.MockTressetteGame, []*domain.TressettePlayer) {
	m := setupTressetteWebMock()
	players := makeTressettePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestTressetteWebPresenter_Output(t *testing.T) {
	p := new(presenter.TressetteWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupTressetteWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TressetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerTeam)
		assert.Equal(t, 0, resObj.Players[0].TeamID)
		assert.Equal(t, 1, resObj.Players[1].TeamID)
		assert.Equal(t, "tressette.playPhase.lead", resObj.MessageCode)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		// human cards visible, cpu hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupTressetteWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TressetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.TressetteCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 21, resObj.Config.TargetPoints)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupTressetteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TressetteTrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 3, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.TressetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "tressette.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end / round end message codes", func(t *testing.T) {
		for phase, code := range map[domain.TressettePhase]string{
			domain.TressettePhaseTrickEnd: "tressette.trickEnd",
			domain.TressettePhaseRoundEnd: "tressette.roundEnd",
		} {
			m, _ := setupTressetteWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			result := p.Output(m, nil)
			var resObj controller.TressetteWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
			assert.Equal(t, code, resObj.MessageCode)
		}
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupTressetteWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.TressetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupTressetteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		var resObj controller.TressetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "tressette.result.humanTeamWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "A"}, resObj.MessageParams)
	})

	t.Run("game end cpu team wins", func(t *testing.T) {
		m, _ := setupTressetteWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		result := p.Output(m, nil)
		var resObj controller.TressetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "tressette.result.cpuTeamWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"team": "B"}, resObj.MessageParams)
	})
}

func TestTressetteWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TressetteWebPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, _ := setupTressetteWebMockWithPlayers()
		m.On("GetHint").Return(&domain.TressetteHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.TressetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTressetteWebMockWithPlayers()
		m.On("GetHint").Return((*domain.TressetteHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.TressetteWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestTressetteWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TressetteWebPresenter)
	m := new(interfaces.MockTressetteGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠5"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}
