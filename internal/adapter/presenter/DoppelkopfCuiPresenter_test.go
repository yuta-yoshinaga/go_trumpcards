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

func makeDoppelkopfPlayers() []*domain.DoppelkopfPlayer {
	cfg := domain.DefaultDoppelkopfConfig()
	return []*domain.DoppelkopfPlayer{
		domain.NewDoppelkopfPlayer(true, cfg.StartChips),
		domain.NewDoppelkopfPlayer(false, cfg.StartChips),
		domain.NewDoppelkopfPlayer(false, cfg.StartChips),
		domain.NewDoppelkopfPlayer(false, cfg.StartChips),
	}
}

func setupDoppelkopfCuiMock() *interfaces.MockDoppelkopfGame {
	m := new(interfaces.MockDoppelkopfGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("CanHumanAnnounce").Return(false)
	m.On("GetRoundRePoints").Return(0)
	m.On("GetRoundReWon").Return(false)
	m.On("GetRoundGamePoints").Return(0)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupDoppelkopfCuiMockWithPlayers() (*interfaces.MockDoppelkopfGame, []*domain.DoppelkopfPlayer) {
	m := setupDoppelkopfCuiMock()
	players := makeDoppelkopfPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestDoppelkopfCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.DoppelkopfCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupDoppelkopfCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Doppelkopf")
		assert.NotEmpty(t, result)
	})

	t.Run("play phase with announce prompt", func(t *testing.T) {
		m, _ := setupDoppelkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "CanHumanAnnounce")
		m.On("CanHumanAnnounce").Return(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupDoppelkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.DoppelkopfPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt shows localized Kontra-wins outcome", func(t *testing.T) {
		m, _ := setupDoppelkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.DoppelkopfPhaseRoundEnd)
		// GetRoundReWon defaults to false -> Kontra wins. Default locale is ja.
		assert.Contains(t, p.Output(m, nil), "Kontra の勝ち")

		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		assert.Contains(t, p.Output(m, nil), "Kontra wins")
	})

	t.Run("round end prompt shows localized Re-wins outcome", func(t *testing.T) {
		m, _ := setupDoppelkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.DoppelkopfPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundReWon")
		m.On("GetRoundReWon").Return(true)
		assert.Contains(t, p.Output(m, nil), "Re の勝ち")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupDoppelkopfCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupDoppelkopfCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestDoppelkopfCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.DoppelkopfCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupDoppelkopfCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.DoppelkopfHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupDoppelkopfCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.On("GetHint").Return(&domain.DoppelkopfHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupDoppelkopfCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.DoppelkopfHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestDoppelkopfCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.DoppelkopfCuiPresenter)
	m := new(interfaces.MockDoppelkopfGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠Q"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
