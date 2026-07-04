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

func makeSambaPlayers() []*domain.SambaPlayer {
	return []*domain.SambaPlayer{
		domain.NewSambaPlayer(true, 0),
		domain.NewSambaPlayer(false, 1),
		domain.NewSambaPlayer(false, 0),
		domain.NewSambaPlayer(false, 1),
	}
}

func setupSambaWebMock() *interfaces.MockSambaGame {
	m := new(interfaces.MockSambaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(80)
	m.On("GetDiscardPileCount").Return(0)
	m.On("GetIsFrozen").Return(false)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SambaPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultSambaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetTeamCount").Return(2)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	return m
}

func setupSambaWebMockWithPlayers() (*interfaces.MockSambaGame, []*domain.SambaPlayer) {
	m := setupSambaWebMock()
	players := makeSambaPlayers()
	m.On("GetPlayerCnt").Return(4)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestSambaWebPresenter_Output(t *testing.T) {
	p := new(presenter.SambaWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupSambaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(m, nil)
		var resObj controller.SambaWebOutput
		require := assert.New(t)
		require.NoError(json.Unmarshal([]byte(result), &resObj))
		require.Equal(4, len(resObj.Players))
		require.False(resObj.GameEndFlag)
		require.Equal(0, resObj.Phase)
		require.Equal(80, resObj.DrawPileCount)
		require.Equal(-1, resObj.WinnerIdx)
		require.Equal(2, len(resObj.TeamScores))
		require.Equal(0, resObj.Players[0].Team)
		require.Equal(1, resObj.Players[1].Team)
	})

	t.Run("human cards shown, CPU hidden in draw phase", func(t *testing.T) {
		m, players := setupSambaWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

		result := p.Output(m, nil)
		var resObj controller.SambaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.True(t, resObj.Players[0].IsHuman)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 1, resObj.Players[1].CardCount)
	})

	t.Run("CPU cards shown at round end", func(t *testing.T) {
		m, players := setupSambaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SambaPhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

		result := p.Output(m, nil)
		var resObj controller.SambaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
	})

	t.Run("set and sequence melds distinguished", func(t *testing.T) {
		m, players := setupSambaWebMockWithPlayers()
		players[0].SetMelds([]*domain.SambaMeld{
			{Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
			}, Kind: domain.SambaMeldSet, IsNatural: true},
			{Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignHeart, 4, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
				domain.NewCard(domain.CardDesignHeart, 6, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignHeart, 8, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
				domain.NewCard(domain.CardDesignHeart, 10, false),
			}, Kind: domain.SambaMeldSequence, IsNatural: true},
		})

		result := p.Output(m, nil)
		var resObj controller.SambaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		melds := resObj.Players[0].Melds
		require := assert.New(t)
		require.Len(melds, 2)
		require.Equal(int(domain.SambaMeldSet), melds[0].Kind)
		require.False(melds[0].IsSamba)
		require.Equal(int(domain.SambaMeldSequence), melds[1].Kind)
		require.True(melds[1].IsSamba)
		require.True(resObj.Players[0].HasSamba)
	})

	t.Run("team scores exposed", func(t *testing.T) {
		m, _ := setupSambaWebMockWithPlayers()
		// setupSambaWebMock registers two GetTeamScore defaults (arg 0 and arg 1),
		// and removeMockCall drops only the first match, so remove twice before
		// re-adding — otherwise the leftover GetTeamScore(1)->0 default shadows
		// the new GetTeamScore(1)->300 (testify matches the first expectation).
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScore")
		m.On("GetTeamScore", 0).Return(1500)
		m.On("GetTeamScore", 1).Return(300)

		result := p.Output(m, nil)
		var resObj controller.SambaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, []int{1500, 300}, resObj.TeamScores)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupSambaWebMockWithPlayers()
		result := p.Output(m, errors.New("test error"))
		var resObj controller.SambaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human team wins", func(t *testing.T) {
		m, _ := setupSambaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		m.On("GetPhase").Return(domain.SambaPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.SambaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "あなた")
		assert.Equal(t, "samba.result.humanWin", resObj.MessageCode)
	})

	t.Run("game end CPU team wins", func(t *testing.T) {
		m, _ := setupSambaWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)
		m.On("GetPhase").Return(domain.SambaPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.SambaWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "samba.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "1"}, resObj.MessageParams)
	})

	t.Run("phase message codes", func(t *testing.T) {
		cases := []struct {
			phase domain.SambaPhase
			code  string
		}{
			{domain.SambaPhaseDraw, "samba.drawPhase"},
			{domain.SambaPhaseMeld, "samba.meldPhase"},
			{domain.SambaPhaseDiscard, "samba.discardPhase"},
			{domain.SambaPhaseRoundEnd, "samba.roundEnd"},
		}
		for _, c := range cases {
			m, _ := setupSambaWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(c.phase)
			result := p.Output(m, nil)
			var resObj controller.SambaWebOutput
			_ = json.Unmarshal([]byte(result), &resObj)
			assert.Equal(t, c.code, resObj.MessageCode)
		}
	})
}

func TestSambaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SambaWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockSambaGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "drew"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"draw_stock"`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockSambaGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}
