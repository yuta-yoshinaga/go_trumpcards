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

func makeSheepsheadPlayers() []*domain.SheepsheadPlayer {
	cfg := domain.DefaultSheepsheadConfig()
	return []*domain.SheepsheadPlayer{
		domain.NewSheepsheadPlayer(true, cfg.StartChips),
		domain.NewSheepsheadPlayer(false, cfg.StartChips),
		domain.NewSheepsheadPlayer(false, cfg.StartChips),
		domain.NewSheepsheadPlayer(false, cfg.StartChips),
		domain.NewSheepsheadPlayer(false, cfg.StartChips),
	}
}

func setupSheepsheadCuiMock() *interfaces.MockSheepsheadGame {
	m := new(interfaces.MockSheepsheadGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetPickerIdx").Return(-1)
	m.On("GetPartnerIdx").Return(-1)
	m.On("GetCalledSuit").Return(0)
	m.On("GetBlind").Return([]*domain.Card(nil))
	m.On("GetPassCount").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.SheepsheadTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SheepsheadPhasePick)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("IsPartnerRevealed").Return(false)
	m.On("GetRoundPickerPoints").Return(0)
	m.On("GetRoundMultiplier").Return(1)
	m.On("GetRoundPickerWon").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSheepsheadCuiMockWithPlayers() (*interfaces.MockSheepsheadGame, []*domain.SheepsheadPlayer) {
	m := setupSheepsheadCuiMock()
	players := makeSheepsheadPlayers()
	m.On("GetPlayerCnt").Return(5)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	m.On("GetPlayer", 4).Return(players[4])
	return m, players
}

func TestSheepsheadCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SheepsheadCuiPresenter)

	t.Run("pick phase shows blind count and prompt", func(t *testing.T) {
		m, players := setupSheepsheadCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Sheepshead")
		assert.NotEmpty(t, result)
	})

	t.Run("bury phase shows picker info", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPickerIdx").Return(0)
		m.On("GetPhase").Return(domain.SheepsheadPhaseBury)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("call phase shows picker prompt", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPickerIdx").Return(0)
		m.On("GetPhase").Return(domain.SheepsheadPhaseCall)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("play phase shows current player", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SheepsheadPhasePlay)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SheepsheadPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows picker won result", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundPickerWon")
		m.On("GetPhase").Return(domain.SheepsheadPhaseRoundEnd)
		m.On("GetRoundPickerWon").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, i18n.T("sheepshead.roundPickerWon"))
	})

	t.Run("round end shows picker lost result", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SheepsheadPhaseRoundEnd)
		result := p.Output(m, nil) // default GetRoundPickerWon = false
		assert.Contains(t, result, i18n.T("sheepshead.roundPickerLost"))
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("picker with partner revealed", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCalledSuit")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsPartnerRevealed")
		m.On("GetPickerIdx").Return(0)
		m.On("GetPartnerIdx").Return(2)
		m.On("GetCalledSuit").Return(domain.CardDesignClover)
		m.On("IsPartnerRevealed").Return(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})
}

func TestSheepsheadCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SheepsheadCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.SheepsheadHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("pick hint", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.SheepsheadHint{Pick: true, Reason: "pick_take"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("bury hint with card indices", func(t *testing.T) {
		m, players := setupSheepsheadCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPickerIdx").Return(0)
		m.On("GetPhase").Return(domain.SheepsheadPhaseBury)
		m.On("GetHint").Return(&domain.SheepsheadHint{CardIndices: []int{0, 1}, Reason: "bury_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("call suit hint", func(t *testing.T) {
		m, _ := setupSheepsheadCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.SheepsheadHint{Suit: domain.CardDesignClover, Reason: "call_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("play hint", func(t *testing.T) {
		m, players := setupSheepsheadCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.On("GetHint").Return(&domain.SheepsheadHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestSheepsheadCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SheepsheadCuiPresenter)
	m := new(interfaces.MockSheepsheadGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "pick", Detail: "You picks up the blind"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "pick")
}
