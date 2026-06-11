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

func setupCribbageWebMock() *interfaces.MockCribbageGame {
	m := new(interfaces.MockCribbageGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetStarter").Return((*domain.Card)(nil))
	m.On("GetPegCount").Return(0)
	m.On("GetPegPlayedCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CribbagePhaseDiscard)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultCribbageConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHandScoreDetails").Return([3]*domain.CribbageScoreDetail{})
	m.On("GetCrib").Return(([]*domain.Card)(nil))
	m.On("GetShowPhaseStep").Return(0)
	return m
}

func setupCribbageWebMockWithPlayers() (*interfaces.MockCribbageGame, []*domain.CribbagePlayer) {
	m := setupCribbageWebMock()
	players := makeCribbagePlayers()
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestCribbageWebPresenter_Output(t *testing.T) {
	p := new(presenter.CribbageWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupCribbageWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.CribbageWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 0, resObj.DealerIdx)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Nil(t, resObj.Starter)
		assert.Equal(t, "", resObj.Message)
	})

	t.Run("human cards shown, CPU cards hidden in discard phase", func(t *testing.T) {
		m, players := setupCribbageWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsHuman)
		assert.Equal(t, 1, humanPlayer.CardCount)
		assert.Len(t, humanPlayer.Cards, 1)
		assert.Equal(t, "SPADE", humanPlayer.Cards[0].Design)
		assert.Equal(t, 5, humanPlayer.Cards[0].Value)

		cpu1 := resObj.Players[1]
		assert.False(t, cpu1.IsHuman)
		assert.Equal(t, 1, cpu1.CardCount)
		assert.Len(t, cpu1.Cards, 0)
	})

	t.Run("CPU cards shown in show phase", func(t *testing.T) {
		m, players := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CribbagePhaseShow)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.Len(t, cpu1.Cards, 1)
	})

	t.Run("CPU cards shown in round end phase", func(t *testing.T) {
		m, players := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CribbagePhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.Len(t, cpu1.Cards, 1)
	})

	t.Run("CPU cards shown in game end phase", func(t *testing.T) {
		m, players := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.CribbagePhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.Len(t, cpu1.Cards, 1)
	})

	t.Run("player scores", func(t *testing.T) {
		m, players := setupCribbageWebMockWithPlayers()
		players[1].SetCumulativeScore(80)
		players[1].SetRoundScore(15)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 80, resObj.Players[1].CumulativeScore)
		assert.Equal(t, 15, resObj.Players[1].RoundScore)
	})

	t.Run("starter populated", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetStarter")
		starter := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetStarter").Return(starter)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.Starter)
		assert.Equal(t, "HEART", resObj.Starter.Design)
		assert.Equal(t, 7, resObj.Starter.Value)
	})

	t.Run("crib shown in show phase", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCrib")
		m.On("GetPhase").Return(domain.CribbagePhaseShow)
		m.On("GetCrib").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
		})

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.Crib, 2)
	})

	t.Run("crib hidden in discard phase", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCrib")
		m.On("GetCrib").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
		})

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.Crib, 0)
	})

	t.Run("peg played cards", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPegPlayedCards")
		m.On("GetPegPlayedCards").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
		})

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.PegPlayedCards, 1)
		assert.Equal(t, "SPADE", resObj.PegPlayedCards[0].Design)
	})

	t.Run("hand score details", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHandScoreDetails")
		m.On("GetHandScoreDetails").Return([3]*domain.CribbageScoreDetail{
			{Fifteens: 4, Pairs: 2, Runs: 3, Flush: 0, Nobs: 1, Total: 10},
			nil,
			nil,
		})

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.HandScoreDetails[0])
		assert.Equal(t, 4, resObj.HandScoreDetails[0].Fifteens)
		assert.Equal(t, 2, resObj.HandScoreDetails[0].Pairs)
		assert.Equal(t, 3, resObj.HandScoreDetails[0].Runs)
		assert.Equal(t, 0, resObj.HandScoreDetails[0].Flush)
		assert.Equal(t, 1, resObj.HandScoreDetails[0].Nobs)
		assert.Equal(t, 10, resObj.HandScoreDetails[0].Total)
		assert.Nil(t, resObj.HandScoreDetails[1])
		assert.Nil(t, resObj.HandScoreDetails[2])
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.CribbageConfig{
			CpuDifficulty: domain.CribbageCpuDifficultyHard,
			PointLimit:    200,
		})

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.CribbageCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 200, resObj.Config.PointLimit)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		m.On("GetPhase").Return(domain.CribbagePhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "あなた")
		assert.Equal(t, "cribbage.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)
		m.On("GetPhase").Return(domain.CribbagePhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 1")
		assert.Equal(t, "cribbage.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "1"}, resObj.MessageParams)
	})

	t.Run("game end nil player at winnerIdx", func(t *testing.T) {
		m := setupCribbageWebMock()
		m.On("GetPlayer", 0).Return((*domain.CribbagePlayer)(nil))
		m.On("GetPlayer", 1).Return((*domain.CribbagePlayer)(nil))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.CribbagePlayer)(nil))
		m.On("GetPhase").Return(domain.CribbagePhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 99")
		assert.Equal(t, "cribbage.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "99"}, resObj.MessageParams)
	})

	t.Run("discard phase messageCode", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "cribbage.discardPhase", resObj.MessageCode)
	})

	t.Run("pegging phase messageCode", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CribbagePhasePegging)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "cribbage.peggingPhase", resObj.MessageCode)
	})

	t.Run("show phase messageCode", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CribbagePhaseShow)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "cribbage.showPhase", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CribbagePhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "cribbage.roundEnd", resObj.MessageCode)
	})

	t.Run("unrecognized phase no messageCode", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CribbagePhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupCribbageWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.CribbageWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.CribbageCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 121, resObj.Config.PointLimit)
	})
}

func TestCribbageWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CribbageWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "discard", Detail: "discarded 2 cards", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"discard"`)
		assert.Contains(t, result, `"detail":"discarded 2 cards"`)
		assert.Contains(t, result, `"turnNumber":1`)
		assert.Contains(t, result, `"playerIdx":0`)
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockCribbageGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

func TestCribbageWebPresenter_HintOutput_DelegatesToOutput(t *testing.T) {
	p := new(presenter.CribbageWebPresenter)
	m, players := setupCribbageWebMockWithPlayers()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	// The Web GUI computes hints in the frontend, so the server hint command
	// simply re-emits the current state.
	assert.Equal(t, p.Output(m, nil), p.HintOutput(m))
}
