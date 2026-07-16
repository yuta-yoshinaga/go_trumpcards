//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeCalabresellaPlayers() []*domain.CalabresellaPlayer {
	return []*domain.CalabresellaPlayer{
		domain.NewCalabresellaPlayer(true),
		domain.NewCalabresellaPlayer(false),
		domain.NewCalabresellaPlayer(false),
	}
}

func setupCalabresellaCuiMock() *interfaces.MockCalabresellaGame {
	m := new(interfaces.MockCalabresellaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetWinningBid").Return(domain.CalabresellaBidChiamo)
	m.On("GetCurrentTrick").Return(([]*domain.CalabresellaTrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CalabresellaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetCurrentBidderIdx").Return(1)
	m.On("GetSoloistIdx").Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetRoundThirds").Return([domain.CalabresellaPlayerCnt]int{0, 0, 0})
	m.On("GetPlayerScores").Return([domain.CalabresellaPlayerCnt]int{0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupCalabresellaCuiMockWithPlayers() (*interfaces.MockCalabresellaGame, []*domain.CalabresellaPlayer) {
	m := setupCalabresellaCuiMock()
	players := makeCalabresellaPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestCalabresellaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CalabresellaCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupCalabresellaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "カラブレセッラ") // translated helpTitle
		assert.Contains(t, result, "マストフォロー") // play-phase help mentions must-follow
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseBid)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "ビッド") // translated bid prompt/help
	})

	t.Run("discard phase prompt", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseDiscard)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "モンテ") // translated discard prompt/help
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CalabresellaPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end lists all players thirds with roles", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundThirds")
		m.On("GetPhase").Return(domain.CalabresellaPhaseRoundEnd)
		// Soloist (index 0) takes 5, the two coalition players 3 each → 11 total.
		m.On("GetRoundThirds").Return([domain.CalabresellaPlayerCnt]int{5, 3, 3})
		result := p.Output(m, nil)
		assert.Contains(t, result, i18n.T("calabresella.roleCoalition"))
		// Each player's thirds appear in the breakdown: soloist 5, coalition 3 each.
		assert.Contains(t, result, "5/3")
		assert.Contains(t, result, "3/3")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestCalabresellaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CalabresellaCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.CalabresellaHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupCalabresellaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.CalabresellaHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupCalabresellaCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.CalabresellaHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestCalabresellaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CalabresellaCuiPresenter)
	m := new(interfaces.MockCalabresellaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
