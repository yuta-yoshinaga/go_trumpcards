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

func makeGongZhuPlayers() []*domain.GongZhuPlayer {
	return []*domain.GongZhuPlayer{
		domain.NewGongZhuPlayer(true),
		domain.NewGongZhuPlayer(false),
		domain.NewGongZhuPlayer(false),
		domain.NewGongZhuPlayer(false),
	}
}

func setupGongZhuWebMock() *interfaces.MockGongZhuGame {
	m := new(interfaces.MockGongZhuGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetHeartsBroken").Return(false)
	m.On("GetCurrentTrick").Return([]*domain.GongZhuTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GongZhuPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetExposure").Return(domain.GongZhuExposure{})
	m.On("GetExposableIndices", 0).Return([]int{})
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultGongZhuConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupGongZhuWebMockWithPlayers() (*interfaces.MockGongZhuGame, []*domain.GongZhuPlayer) {
	m := setupGongZhuWebMock()
	players := makeGongZhuPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestGongZhuWebPresenter_Output(t *testing.T) {
	p := new(presenter.GongZhuWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupGongZhuWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, 1, resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, "gongzhu.playPhase.lead", resObj.MessageCode)
		// human cards visible, cpu hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("exposure flags & exposable indices", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetExposure")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetExposableIndices")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetExposure").Return(domain.GongZhuExposure{Pig: true, Sheep: true})
		m.On("GetExposableIndices", 0).Return([]int{1, 4})
		m.On("GetPhase").Return(domain.GongZhuPhaseExpose)

		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.Exposed.Pig)
		assert.True(t, resObj.Exposed.Sheep)
		assert.False(t, resObj.Exposed.Ace)
		assert.Equal(t, []int{1, 4}, resObj.ExposableIndices)
		assert.Equal(t, "gongzhu.exposePhase", resObj.MessageCode)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.GongZhuCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 1000, resObj.Config.PointLimit)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.GongZhuTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "gongzhu.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end / round end message codes", func(t *testing.T) {
		for phase, code := range map[domain.GongZhuPhase]string{
			domain.GongZhuPhaseTrickEnd: "gongzhu.trickEnd",
			domain.GongZhuPhaseRoundEnd: "gongzhu.roundEnd",
		} {
			m, _ := setupGongZhuWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			result := p.Output(m, nil)
			var resObj controller.GongZhuWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
			assert.Equal(t, code, resObj.MessageCode)
		}
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "gongzhu.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)
		result := p.Output(m, nil)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "gongzhu.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})
}

func TestGongZhuWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.GongZhuWebPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.On("GetHint").Return(&domain.GongZhuHint{CardIndices: []int{2}, Reason: "follow_suit"})
		result := p.HintOutput(m)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_suit", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGongZhuWebMockWithPlayers()
		m.On("GetHint").Return((*domain.GongZhuHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.GongZhuWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestGongZhuWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GongZhuWebPresenter)
	m := new(interfaces.MockGongZhuGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠5"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}
