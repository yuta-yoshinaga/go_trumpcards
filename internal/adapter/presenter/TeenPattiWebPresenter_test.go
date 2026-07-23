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

func tpSetupWebMock() *interfaces.MockTeenPattiGame {
	m := new(interfaces.MockTeenPattiGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetPot").Return(4)
	m.On("GetStake").Return(1)
	m.On("GetPhase").Return(domain.TeenPattiPhaseBetting)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetMatchWinnerIdx").Return(-1)
	m.On("IsShowdown").Return(false)
	m.On("CanShow").Return(false)
	m.On("CanRequestSideShow").Return(false)
	m.On("GetSideShowRequester").Return(-1)
	m.On("GetSideShowTarget").Return(-1)
	m.On("GetLastSideShow").Return(-1, -1, -1, false)
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultTeenPattiConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func tpSetupWebMockWithPlayers() (*interfaces.MockTeenPattiGame, []*domain.TeenPattiPlayer) {
	m := tpSetupWebMock()
	players := tpMakePlayers()
	m.On("GetPlayerCnt").Return(domain.TeenPattiPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestTeenPattiWebPresenter_Output(t *testing.T) {
	p := new(presenter.TeenPattiWebPresenter)

	t.Run("betting phase hides CPU cards", func(t *testing.T) {
		m, players := tpSetupWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, domain.TeenPattiPlayerCnt)
		assert.Equal(t, int(domain.TeenPattiPhaseBetting), resObj.Phase)
		assert.Equal(t, "teenpatti.bettingPhase", resObj.MessageCode)
		assert.Equal(t, 4, resObj.Pot)
		assert.Equal(t, 1, resObj.Stake)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 30, resObj.Players[0].Chips)
		assert.Equal(t, -1, resObj.SideShowRequester)
		assert.Equal(t, -1, resObj.SideShowTarget)
		assert.False(t, resObj.CanRequestSideShow)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.TeenPattiCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.TeenPattiDefaultAnte, resObj.Config.Ante)
		assert.Equal(t, domain.TeenPattiDefaultStartingChips, resObj.Config.StartingChips)
	})

	t.Run("side show phase message and fields", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TeenPattiPhaseSideShow)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "CanRequestSideShow")
		m.On("CanRequestSideShow").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSideShowRequester")
		m.On("GetSideShowRequester").Return(2)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetSideShowTarget")
		m.On("GetSideShowTarget").Return(1)
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "teenpatti.sideShowPhase", resObj.MessageCode)
		assert.True(t, resObj.CanRequestSideShow)
		assert.Equal(t, 2, resObj.SideShowRequester)
		assert.Equal(t, 1, resObj.SideShowTarget)
	})

	t.Run("showdown reveals non-folded hands with hand name", func(t *testing.T) {
		m, players := tpSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TeenPattiPhaseShowdown)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsShowdown")
		m.On("IsShowdown").Return(true)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[2].SetFolded(true)
		players[2].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "teenpatti.showdownPhase", resObj.MessageCode)
		assert.Len(t, resObj.Players[1].Cards, 3)
		assert.Equal(t, "trail", resObj.Players[1].HandName)
		assert.Len(t, resObj.Players[2].Cards, 0)
	})

	t.Run("round end human win message", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TeenPattiPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetRoundWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "teenpatti.roundEndHumanWin", resObj.MessageCode)
	})

	t.Run("round end cpu win message", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TeenPattiPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetRoundWinnerIdx").Return(1)
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "teenpatti.roundEndCpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("human-involved side show reveals both hands", func(t *testing.T) {
		m, players := tpSetupWebMockWithPlayers()
		// 人間 (0) が申請者、CPU (2) が対象。人間の勝ち (トレイル) → 対象が敗者。
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[2].SetFolded(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastSideShow")
		m.On("GetLastSideShow").Return(0, 2, 2, true)
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.LastSideShow)
		assert.Equal(t, 0, resObj.LastSideShow.RequesterIdx)
		assert.Equal(t, 2, resObj.LastSideShow.TargetIdx)
		assert.Equal(t, 0, resObj.LastSideShow.WinnerIdx)
		assert.Equal(t, 2, resObj.LastSideShow.LoserIdx)
		assert.Len(t, resObj.LastSideShow.Requester.Cards, 3)
		assert.Equal(t, "trail", resObj.LastSideShow.Requester.HandName)
		assert.Len(t, resObj.LastSideShow.Target.Cards, 3)
		assert.Equal(t, "highcard", resObj.LastSideShow.Target.HandName)
	})

	t.Run("cpu-vs-cpu side show stays hidden", func(t *testing.T) {
		m, players := tpSetupWebMockWithPlayers()
		// CPU (1) 申請者・CPU (2) 対象 → 人間非関与のため非公開。
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastSideShow")
		m.On("GetLastSideShow").Return(1, 2, 2, true)
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.LastSideShow)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human win", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchWinnerIdx")
		m.On("GetMatchWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "teenpatti.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end cpu win", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchWinnerIdx")
		m.On("GetMatchWinnerIdx").Return(2)
		result := p.Output(m, nil)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "teenpatti.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "2"}, resObj.MessageParams)
	})
}

func TestTeenPattiWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TeenPattiWebPresenter)

	t.Run("hint present", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		m.On("GetHint").Return(&domain.TeenPattiHint{Action: "raise", Reason: "strong_hand"})
		result := p.HintOutput(m)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, "raise", resObj.Hint.Action)
		assert.Equal(t, "strong_hand", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := tpSetupWebMockWithPlayers()
		m.On("GetHint").Return((*domain.TeenPattiHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.TeenPattiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestTeenPattiWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TeenPattiWebPresenter)
	m := new(interfaces.MockTeenPattiGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "You bets 1"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"bet"`)
}
