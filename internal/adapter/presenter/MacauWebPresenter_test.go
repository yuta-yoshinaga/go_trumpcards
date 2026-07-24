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

func setupMacauWebMock() *interfaces.MockMacauGame {
	m := new(interfaces.MockMacauGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(30)
	m.On("GetDirection").Return(1)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetChosenSuit").Return(-1)
	m.On("GetPenaltyDrawCount").Return(0)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MacauPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultMacauConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupMacauWebMockWithPlayers() (*interfaces.MockMacauGame, []*domain.MacauPlayer) {
	m := setupMacauWebMock()
	players := makeMacauPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestMacauWebPresenter_Output(t *testing.T) {
	p := new(presenter.MacauWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupMacauWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		var resObj controller.MacauWebOutput
		err := json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 30, resObj.DrawPileCount)
		assert.Equal(t, 1, resObj.Direction)
		assert.Equal(t, 0, resObj.PenaltyDrawCount)
		assert.Equal(t, "macau.playPhase", resObj.MessageCode)
	})

	t.Run("human cards shown, CPU hidden", func(t *testing.T) {
		m, players := setupMacauWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("HintOutput falls back to the state output", func(t *testing.T) {
		m, players := setupMacauWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		// The web presenter computes hints client-side, so HintOutput mirrors Output.
		assert.Equal(t, p.Output(m, nil), p.HintOutput(m))
	})

	t.Run("penalty and direction populated", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPenaltyDrawCount")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDirection")
		m.On("GetPenaltyDrawCount").Return(6)
		m.On("GetDirection").Return(-1)

		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, 6, resObj.PenaltyDrawCount)
		assert.Equal(t, -1, resObj.Direction)
	})

	t.Run("hasDeclared reflected", func(t *testing.T) {
		m, players := setupMacauWebMockWithPlayers()
		players[0].SetHasDeclared(true)

		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.True(t, resObj.Players[0].HasDeclared)
	})

	t.Run("discard top populated", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))

		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, "HEART", resObj.DiscardTop.Design)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, errors.New("test error"))), &resObj)
		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "macau.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)

		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "macau.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})

	t.Run("choose suit phase messageCode", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MacauPhaseChooseSuit)
		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "macau.chooseSuitPhase", resObj.MessageCode)
	})

	t.Run("must declare phase messageCode", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MacauPhaseMustDeclare)
		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "macau.mustDeclarePhase", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MacauPhaseRoundEnd)
		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "macau.roundEnd", resObj.MessageCode)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupMacauWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.MacauConfig{CpuDifficulty: domain.MacauCpuDifficultyHard, PointLimit: 300})
		var resObj controller.MacauWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, int(domain.MacauCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 300, resObj.Config.PointLimit)
	})
}

func TestMacauWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MacauWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockMacauGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"play"`)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockMacauGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
	})
}
