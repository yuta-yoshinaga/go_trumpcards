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

func setupPreferenceWebMock() *interfaces.MockPreferenceGame {
	m := new(interfaces.MockPreferenceGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PreferencePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.PreferenceBidSix)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetBids").Return([domain.PreferencePlayerCnt]domain.PreferenceBid{domain.PreferenceBidSix, domain.PreferenceBidPass, domain.PreferenceBidPass})
	m.On("GetPlayerScores").Return([domain.PreferencePlayerCnt]int{0, 0, 0})
	m.On("GetRoundTricks").Return([domain.PreferencePlayerCnt]int{0, 0, 0})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultPreferenceConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupPreferenceWebMockWithPlayers() (*interfaces.MockPreferenceGame, []*domain.PreferencePlayer) {
	m := setupPreferenceWebMock()
	players := makePreferencePlayers()
	m.On("GetPlayerCnt").Return(3)
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestPreferenceWebPresenter_Output(t *testing.T) {
	p := new(presenter.PreferenceWebPresenter)

	t.Run("initial state play phase lead", func(t *testing.T) {
		m, players := setupPreferenceWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 3)
		assert.Equal(t, int(domain.PreferencePhasePlay), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerPlayer)
		assert.Equal(t, "preference.playPhase.lead", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.True(t, resObj.Players[0].IsDeclarer)
		assert.False(t, resObj.Players[1].IsDeclarer)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
		assert.Equal(t, domain.CardDesignSpade, resObj.TrumpSuit)
		assert.Equal(t, int(domain.PreferenceBidSix), resObj.Contract)
		assert.Equal(t, int(domain.PreferenceBidSix), resObj.Bids[0])
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.PreferenceCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 30, resObj.Config.TargetPoints)
	})

	t.Run("bid phase message code", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanBidTurn")
		m.On("GetPhase").Return(domain.PreferencePhaseBid)
		m.On("IsHumanBidTurn").Return(true)
		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "preference.bidPhase", resObj.MessageCode)
		assert.True(t, resObj.IsHumanBidTurn)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "preference.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PreferencePhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "preference.trickEnd", resObj.MessageCode)
	})

	t.Run("round end message code", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PreferencePhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "preference.roundEnd", resObj.MessageCode)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "preference.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(1)
		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "preference.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("player scores propagated to players", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerScores")
		m.On("GetPlayerScores").Return([domain.PreferencePlayerCnt]int{4, 2, 0})
		result := p.Output(m, nil)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 4, resObj.Players[0].Score)
		assert.Equal(t, 2, resObj.Players[1].Score)
	})
}

func TestPreferenceWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.PreferenceWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.On("GetHint").Return(&domain.PreferenceHint{CardIndices: []int{2}, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupPreferenceWebMockWithPlayers()
		m.On("GetHint").Return((*domain.PreferenceHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.PreferenceWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestPreferenceWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PreferenceWebPresenter)
	m := new(interfaces.MockPreferenceGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"play"`)
}
