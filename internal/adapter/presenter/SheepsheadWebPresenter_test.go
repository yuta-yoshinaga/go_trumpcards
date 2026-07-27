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

func setupSheepsheadWebMock() *interfaces.MockSheepsheadGame {
	m := new(interfaces.MockSheepsheadGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SheepsheadPhasePick)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetPickerIdx").Return(-1)
	m.On("GetPartnerIdx").Return(-1)
	m.On("GetCalledSuit").Return(0)
	m.On("IsPartnerRevealed").Return(false)
	m.On("GetPassCount").Return(0)
	m.On("GetBlind").Return([]*domain.Card(nil))
	m.On("GetBuried").Return([]*domain.Card(nil))
	m.On("GetCallableSuits").Return([]int(nil))
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("GetRoundPickerPoints").Return(0)
	m.On("GetRoundMultiplier").Return(1)
	m.On("GetRoundPickerWon").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultSheepsheadConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSheepsheadWebMockWithPlayers() (*interfaces.MockSheepsheadGame, []*domain.SheepsheadPlayer) {
	m := setupSheepsheadWebMock()
	players := makeSheepsheadPlayers()
	m.On("GetPlayerCnt").Return(5)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	m.On("GetPlayer", 4).Return(players[4])
	return m, players
}

func TestSheepsheadWebPresenter_Output(t *testing.T) {
	p := new(presenter.SheepsheadWebPresenter)

	t.Run("initial state pick phase", func(t *testing.T) {
		m, players := setupSheepsheadWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 5)
		assert.Equal(t, int(domain.SheepsheadPhasePick), resObj.Phase)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, -1, resObj.PickerIdx)
		assert.Equal(t, -1, resObj.PartnerIdx)
		assert.Equal(t, "sheepshead.pickPhase", resObj.MessageCode)
		// human cards visible, CPU hidden
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, int(domain.SheepsheadCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 2, resObj.Config.BaseChips)
		assert.Equal(t, 20, resObj.Config.StartChips)
		assert.Equal(t, 40, resObj.Config.TargetChips)
	})

	t.Run("bury phase message code", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SheepsheadPhaseBury)
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sheepshead.buryPhase", resObj.MessageCode)
	})

	t.Run("call phase returns callable suits", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCallableSuits")
		m.On("GetPhase").Return(domain.SheepsheadPhaseCall)
		m.On("GetCallableSuits").Return([]int{domain.CardDesignClover, domain.CardDesignSpade})
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sheepshead.callPhase", resObj.MessageCode)
		assert.Equal(t, []int{domain.CardDesignClover, domain.CardDesignSpade}, resObj.CallableSuits)
	})

	t.Run("play phase lead message", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SheepsheadPhasePlay)
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sheepshead.playPhase.lead", resObj.MessageCode)
		assert.Equal(t, []int{0}, resObj.PlayableIndices)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.SheepsheadPhasePlay)
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
		})
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.CurrentTrick, 1)
		assert.Equal(t, "sheepshead.playPhase.follow", resObj.MessageCode)
	})

	t.Run("trick end message code", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SheepsheadPhaseTrickEnd)
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sheepshead.trickEnd", resObj.MessageCode)
	})

	t.Run("round end reveals buried cards and partner", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBuried")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickerIdx")
		m.On("GetPhase").Return(domain.SheepsheadPhaseRoundEnd)
		m.On("GetPickerIdx").Return(0)
		m.On("GetPartnerIdx").Return(2)
		m.On("GetBuried").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
		})
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sheepshead.roundEnd", resObj.MessageCode)
		assert.Len(t, resObj.Buried, 2)
		assert.Equal(t, 2, resObj.PartnerIdx)
	})

	t.Run("partner idx hidden during play until revealed", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerIdx")
		m.On("GetPhase").Return(domain.SheepsheadPhasePlay)
		m.On("GetPartnerIdx").Return(3)
		// IsPartnerRevealed is false in default mock
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, -1, resObj.PartnerIdx)
	})

	t.Run("partner idx shown when partner revealed", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsPartnerRevealed")
		m.On("GetPhase").Return(domain.SheepsheadPhasePlay)
		m.On("GetPartnerIdx").Return(3)
		m.On("IsPartnerRevealed").Return(true)
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 3, resObj.PartnerIdx)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "sheepshead.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end cpu wins", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)
		result := p.Output(m, nil)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "sheepshead.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})
}

func TestSheepsheadWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SheepsheadWebPresenter)

	t.Run("hint with card indices", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.On("GetHint").Return(&domain.SheepsheadHint{CardIndices: []int{2}, Suit: 0, Pick: false, Reason: "follow_win"})
		result := p.HintOutput(m)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, []int{2}, resObj.Hint.CardIndices)
		assert.Equal(t, "follow_win", resObj.Hint.Reason)
	})

	t.Run("pick hint", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.On("GetHint").Return(&domain.SheepsheadHint{Pick: true, Reason: "pick_take"})
		result := p.HintOutput(m)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.NotNil(t, resObj.Hint)
		assert.True(t, resObj.Hint.Pick)
		assert.Equal(t, "pick_take", resObj.Hint.Reason)
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSheepsheadWebMockWithPlayers()
		m.On("GetHint").Return((*domain.SheepsheadHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.SheepsheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Nil(t, resObj.Hint)
	})
}

func TestSheepsheadWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SheepsheadWebPresenter)
	m := new(interfaces.MockSheepsheadGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "pick", Detail: "You picks up the blind"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, `"actionType":"pick"`)
}
