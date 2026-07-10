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

func makePrsiPlayers() []*domain.PrsiPlayer {
	return []*domain.PrsiPlayer{
		domain.NewPrsiPlayer(true),
		domain.NewPrsiPlayer(false),
		domain.NewPrsiPlayer(false),
		domain.NewPrsiPlayer(false),
	}
}

func setupPrsiWebMock() *interfaces.MockPrsiGame {
	m := new(interfaces.MockPrsiGame)
	m.On("GetDrawPileCount").Return(11)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetPenaltyDrawCount").Return(0)
	m.On("GetPendingSkips").Return(0)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PrsiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultPrsiConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupPrsiWebMockWithPlayers() (*interfaces.MockPrsiGame, []*domain.PrsiPlayer) {
	m := setupPrsiWebMock()
	players := makePrsiPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPrsiWebPresenter_Output(t *testing.T) {
	p := new(presenter.PrsiWebPresenter)

	t.Run("initial state, human cards shown, CPU hidden", func(t *testing.T) {
		m, players := setupPrsiWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

		result := p.Output(m, nil)
		var resObj controller.PrsiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 4)
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 11, resObj.DrawPileCount)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Nil(t, resObj.DiscardTop)
		assert.Equal(t, "prsi.playPhase", resObj.MessageCode)

		assert.True(t, resObj.Players[0].IsHuman)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.False(t, resObj.Players[1].IsHuman)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 1, resObj.Players[1].CardCount)
	})

	t.Run("discard top and penalty populated", func(t *testing.T) {
		m, _ := setupPrsiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPenaltyDrawCount")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		m.On("GetPenaltyDrawCount").Return(4)

		result := p.Output(m, nil)
		var resObj controller.PrsiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, "HEART", resObj.DiscardTop.Design)
		assert.Equal(t, 7, resObj.DiscardTop.Value)
		assert.Equal(t, 4, resObj.PenaltyDrawCount)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupPrsiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.PrsiConfig{CpuDifficulty: domain.PrsiCpuDifficultyHard})

		result := p.Output(m, nil)
		var resObj controller.PrsiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, int(domain.PrsiCpuDifficultyHard), resObj.Config.CpuDifficulty)
	})

	t.Run("error message takes priority", func(t *testing.T) {
		m, _ := setupPrsiWebMockWithPlayers()
		result := p.Output(m, errors.New("test error"))
		var resObj controller.PrsiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupPrsiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.PrsiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "prsi.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupPrsiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)

		result := p.Output(m, nil)
		var resObj controller.PrsiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "prsi.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})

	t.Run("game end nil winner player", func(t *testing.T) {
		m := setupPrsiWebMock()
		m.On("GetPlayerCnt").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.PrsiPlayer)(nil))

		result := p.Output(m, nil)
		var resObj controller.PrsiWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, "prsi.result.cpuWin", resObj.MessageCode)
	})

	t.Run("non-play, non-end phase has no messageCode", func(t *testing.T) {
		m, _ := setupPrsiWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PrsiPhase(99))

		result := p.Output(m, nil)
		var resObj controller.PrsiWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Empty(t, resObj.MessageCode)
	})
}

func TestPrsiWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PrsiWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockPrsiGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays SPADE 7", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, true)}},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"play"`)
		assert.Contains(t, result, `"turnNumber":1`)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockPrsiGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		assert.Contains(t, p.ActionLogOutput(m), `"entries":[]`)
	})
}
