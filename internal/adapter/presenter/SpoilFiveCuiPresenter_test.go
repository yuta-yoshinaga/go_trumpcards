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

func makeSpoilFivePlayers() []*domain.SpoilFivePlayer {
	return []*domain.SpoilFivePlayer{
		domain.NewSpoilFivePlayer(true),
		domain.NewSpoilFivePlayer(false),
		domain.NewSpoilFivePlayer(false),
		domain.NewSpoilFivePlayer(false),
		domain.NewSpoilFivePlayer(false),
	}
}

func setupSpoilFiveCuiMock() *interfaces.MockSpoilFiveGame {
	m := new(interfaces.MockSpoilFiveGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetPot").Return(5)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetCurrentTrick").Return(([]*domain.SpoilFiveTrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSpoilFiveCuiMockWithPlayers() (*interfaces.MockSpoilFiveGame, []*domain.SpoilFivePlayer) {
	m := setupSpoilFiveCuiMock()
	players := makeSpoilFivePlayers()
	m.On("GetPlayerCnt").Return(5)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestSpoilFiveCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SpoilFiveCuiPresenter)

	t.Run("play phase shows current player and lead marker", func(t *testing.T) {
		m, players := setupSpoilFiveCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// The lead player (seat 0) is flagged; no round winner during play.
		assert.Contains(t, result, i18n.T("spoilfive.leaderMark"))
		assert.NotContains(t, result, i18n.T("spoilfive.winnerMark"))
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupSpoilFiveCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpoilFivePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt with winner", func(t *testing.T) {
		m, _ := setupSpoilFiveCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetPhase").Return(domain.SpoilFivePhaseRoundEnd)
		m.On("GetRoundWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// Seat 0 is the round winner → the winner marker appears.
		assert.Contains(t, result, i18n.T("spoilfive.winnerMark"))
	})

	t.Run("round end spoil prompt", func(t *testing.T) {
		m, _ := setupSpoilFiveCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SpoilFivePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupSpoilFiveCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupSpoilFiveCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestSpoilFiveCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SpoilFiveCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSpoilFiveCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.SpoilFiveHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupSpoilFiveCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.SpoilFiveHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupSpoilFiveCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.SpoilFiveHint{CardIndices: nil, Reason: "take_trick"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestSpoilFiveCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpoilFiveCuiPresenter)
	m := new(interfaces.MockSpoilFiveGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
