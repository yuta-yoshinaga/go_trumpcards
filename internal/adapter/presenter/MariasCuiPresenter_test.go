//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeMariasPlayers() []*domain.MariasPlayer {
	return []*domain.MariasPlayer{
		domain.NewMariasPlayer(true),
		domain.NewMariasPlayer(false),
		domain.NewMariasPlayer(false),
	}
}

func setupMariasCuiMock() *interfaces.MockMariasGame {
	m := new(interfaces.MockMariasGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.MariasTrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MariasPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetSoloistIdx").Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetRoundCardPoints").Return([domain.MariasPlayerCnt]int{0, 0, 0})
	m.On("GetPlayerScores").Return([domain.MariasPlayerCnt]int{0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupMariasCuiMockWithPlayers() (*interfaces.MockMariasGame, []*domain.MariasPlayer) {
	m := setupMariasCuiMock()
	players := makeMariasPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestMariasCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MariasCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupMariasCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Mariáš")
		assert.Contains(t, result, "マリッジ") // play-phase help explains the marriage bonus
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MariasPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows defenders total", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundCardPoints")
		m.On("GetPhase").Return(domain.MariasPhaseRoundEnd)
		// Soloist (idx 0) took 40; the two defenders took 30 + 20 = 50.
		m.On("GetRoundCardPoints").Return([domain.MariasPlayerCnt]int{40, 30, 20})
		result := p.Output(m, nil)
		assert.Contains(t, result, strings.Split(i18n.T("marias.promptRoundEndDefenders"), "{{")[0])
		assert.Contains(t, result, "50")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestMariasCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MariasCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.MariasHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupMariasCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.MariasHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupMariasCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.MariasHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestMariasCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MariasCuiPresenter)
	m := new(interfaces.MockMariasGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
