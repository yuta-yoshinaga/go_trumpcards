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

func tcbSetupWebMock() *interfaces.MockThreeCardBragGame {
	m := new(interfaces.MockThreeCardBragGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetPot").Return(4)
	m.On("GetStake").Return(1)
	m.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetMatchWinnerIdx").Return(-1)
	m.On("IsShowdown").Return(false)
	m.On("CanShow").Return(false)
	m.On("IsHumanTurn").Return(true)
	m.On("GetConfig").Return(domain.DefaultThreeCardBragConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func tcbSetupWebMockWithPlayers() (*interfaces.MockThreeCardBragGame, []*domain.ThreeCardBragPlayer) {
	m := tcbSetupWebMock()
	players := tcbMakePlayers()
	m.On("GetPlayerCnt").Return(domain.ThreeCardBragPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestThreeCardBragWebPresenter_Output(t *testing.T) {
	p := new(presenter.ThreeCardBragWebPresenter)

	t.Run("betting phase hides CPU cards", func(t *testing.T) {
		m, players := tcbSetupWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		result := p.Output(m, nil)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, domain.ThreeCardBragPlayerCnt)
		assert.Equal(t, int(domain.ThreeCardBragPhaseBetting), resObj.Phase)
		assert.Equal(t, "threecardbrag.bettingPhase", resObj.MessageCode)
		assert.Equal(t, 4, resObj.Pot)
		assert.Equal(t, 1, resObj.Stake)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 30, resObj.Players[0].Chips)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := tcbSetupWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.ThreeCardBragCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, domain.ThreeCardBragDefaultAnte, resObj.Config.Ante)
		assert.Equal(t, domain.ThreeCardBragDefaultStartingChips, resObj.Config.StartingChips)
	})

	t.Run("showdown reveals non-folded hands with hand name", func(t *testing.T) {
		m, players := tcbSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThreeCardBragPhaseShowdown)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsShowdown")
		m.On("IsShowdown").Return(true)
		// CPU (player 1) gets a prial 5-5-5 -> handName set.
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		// player 2 folded -> stays hidden.
		players[2].SetFolded(true)
		players[2].AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		result := p.Output(m, nil)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "threecardbrag.showdownPhase", resObj.MessageCode)
		assert.Len(t, resObj.Players[1].Cards, 3)
		assert.Equal(t, "prial", resObj.Players[1].HandName)
		assert.Len(t, resObj.Players[2].Cards, 0)
	})

	t.Run("round end human win message", func(t *testing.T) {
		m, _ := tcbSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThreeCardBragPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetRoundWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "threecardbrag.roundEndHumanWin", resObj.MessageCode)
	})

	t.Run("round end cpu win message", func(t *testing.T) {
		m, _ := tcbSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThreeCardBragPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetRoundWinnerIdx").Return(1)
		result := p.Output(m, nil)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "threecardbrag.roundEndCpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "1"}, resObj.MessageParams)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := tcbSetupWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human win", func(t *testing.T) {
		m, _ := tcbSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchWinnerIdx")
		m.On("GetMatchWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "threecardbrag.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end cpu win", func(t *testing.T) {
		m, _ := tcbSetupWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchWinnerIdx")
		m.On("GetMatchWinnerIdx").Return(2)
		result := p.Output(m, nil)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "threecardbrag.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"player": "2"}, resObj.MessageParams)
	})
}

func TestThreeCardBragWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.ThreeCardBragWebPresenter)

	t.Run("hint present", func(t *testing.T) {
		m, _ := tcbSetupWebMockWithPlayers()
		m.On("GetHint").Return(&domain.ThreeCardBragHint{Action: "raise", Reason: "strong_hand"})
		result := p.HintOutput(m)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, "raise", resObj.Hint.Action)
		assert.Equal(t, "strong_hand", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := tcbSetupWebMockWithPlayers()
		m.On("GetHint").Return((*domain.ThreeCardBragHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.ThreeCardBragWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestThreeCardBragWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ThreeCardBragWebPresenter)
	m := new(interfaces.MockThreeCardBragGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "You bets 1"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"bet"`)
}
