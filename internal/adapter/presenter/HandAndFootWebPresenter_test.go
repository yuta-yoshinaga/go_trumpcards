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

func setupHandAndFootWebMock() *interfaces.MockHandAndFootGame {
	m := new(interfaces.MockHandAndFootGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(75)
	m.On("GetDiscardPileCount").Return(0)
	m.On("GetIsFrozen").Return(false)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.HandAndFootPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetConfig").Return(domain.DefaultHandAndFootConfig())
	m.On("GetTeamMelds", 0).Return(([]*domain.CanastaMeld)(nil))
	m.On("GetTeamMelds", 1).Return(([]*domain.CanastaMeld)(nil))
	m.On("GetTeamRed3s", 0).Return(([]*domain.Card)(nil))
	m.On("GetTeamRed3s", 1).Return(([]*domain.Card)(nil))
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupHandAndFootWebMockWithPlayers() (*interfaces.MockHandAndFootGame, []*domain.HandAndFootPlayer) {
	m := setupHandAndFootWebMock()
	players := makeHandAndFootPlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestHandAndFootWebPresenter_Output(t *testing.T) {
	p := new(presenter.HandAndFootWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupHandAndFootWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		require := assert.New(t)
		require.NoError(json.Unmarshal([]byte(result), &resObj))
		require.Equal(4, len(resObj.Players))
		require.Equal(2, len(resObj.Teams))
		require.False(resObj.GameEndFlag)
		require.Equal(0, resObj.CurrentPlayerIdx)
		require.Equal(1, resObj.RoundNumber)
		require.Equal(75, resObj.DrawPileCount)
		require.Equal(-1, resObj.WinnerTeam)
		require.Nil(resObj.DiscardTop)
		require.Equal("handandfoot.drawPhase", resObj.MessageCode)
	})

	t.Run("human cards shown CPU hidden in draw phase", func(t *testing.T) {
		m, players := setupHandAndFootWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.Players[0].IsHuman)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Equal(t, 0, resObj.Players[0].Team)
		assert.Equal(t, 1, resObj.Players[1].Team)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 1, resObj.Players[1].CardCount)
	})

	t.Run("CPU cards shown in round end", func(t *testing.T) {
		m, players := setupHandAndFootWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HandAndFootPhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
	})

	t.Run("foot count and inFoot", func(t *testing.T) {
		m, players := setupHandAndFootWebMockWithPlayers()
		players[0].AddFootCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[0].SetInFoot(false)
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, 1, resObj.Players[0].FootCount)
		assert.False(t, resObj.Players[0].InFoot)
	})

	t.Run("team meld output", func(t *testing.T) {
		m, _ := setupHandAndFootWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamMelds")
		meld := &domain.CanastaMeld{
			Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			},
			IsNatural: true,
		}
		m.On("GetTeamMelds", 0).Return([]*domain.CanastaMeld{meld})
		m.On("GetTeamMelds", 1).Return(([]*domain.CanastaMeld)(nil))
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Teams[0].Melds, 1)
		assert.Equal(t, 7, resObj.Teams[0].Melds[0].Rank)
		assert.True(t, resObj.Teams[0].Melds[0].IsNatural)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupHandAndFootWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.HandAndFootConfig{CpuDifficulty: domain.HandAndFootCpuDifficultyHard, PointLimit: 7500, RedCanastasToGoOut: 1, BlackCanastasToGoOut: 1})
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, int(domain.HandAndFootCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 7500, resObj.Config.PointLimit)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupHandAndFootWebMockWithPlayers()
		result := p.Output(m, errors.New("test error"))
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupHandAndFootWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		m.On("GetPhase").Return(domain.HandAndFootPhaseGameEnd)
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "handandfoot.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end CPU team wins", func(t *testing.T) {
		m, _ := setupHandAndFootWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(1)
		m.On("GetPhase").Return(domain.HandAndFootPhaseGameEnd)
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "handandfoot.result.cpuWin", resObj.MessageCode)
	})

	t.Run("meld phase messageCode", func(t *testing.T) {
		m, _ := setupHandAndFootWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HandAndFootPhaseMeld)
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "handandfoot.meldPhase", resObj.MessageCode)
	})

	t.Run("discard phase messageCode", func(t *testing.T) {
		m, _ := setupHandAndFootWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HandAndFootPhaseDiscard)
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "handandfoot.discardPhase", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupHandAndFootWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.HandAndFootPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.HandAndFootWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "handandfoot.roundEnd", resObj.MessageCode)
	})
}

func TestHandAndFootWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.HandAndFootWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockHandAndFootGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "drew", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"draw_stock"`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockHandAndFootGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

func TestHandAndFootWebPresenter_HintOutput(t *testing.T) {
	// Web hints are client-side, so HintOutput mirrors Output.
	p := new(presenter.HandAndFootWebPresenter)
	m, _ := setupHandAndFootWebMockWithPlayers()
	assert.Equal(t, p.Output(m, nil), p.HintOutput(m))
}
