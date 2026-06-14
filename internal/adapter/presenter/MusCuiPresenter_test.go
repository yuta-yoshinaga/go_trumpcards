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
)

func makeMusPlayers() []*domain.MusPlayer {
	return []*domain.MusPlayer{
		domain.NewMusPlayer(true),
		domain.NewMusPlayer(false),
		domain.NewMusPlayer(false),
		domain.NewMusPlayer(false),
	}
}

func setupMusCuiMock() *interfaces.MockMusGame {
	m := new(interfaces.MockMusGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetAmarrakos").Return([domain.MusTeamCnt]int{0, 0})
	m.On("GetPhase").Return(domain.MusPhaseMus)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetMusTurn").Return(0)
	m.On("GetDiscardTurn").Return(0)
	m.On("GetBetTeam").Return(0)
	m.On("GetPendingStake").Return(0)
	m.On("GetLastBettorTeam").Return(-1)
	m.On("GetManoIdx").Return(0)
	m.On("GetMusCycle").Return(0)
	for ri := 0; ri < domain.MusRoundCnt; ri++ {
		m.On("GetResult", ri).Return(domain.MusRoundResult{Team: -1})
	}
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupMusCuiMockWithPlayers() (*interfaces.MockMusGame, []*domain.MusPlayer) {
	m := setupMusCuiMock()
	players := makeMusPlayers()
	m.On("GetPlayerCnt").Return(domain.MusPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestMusCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MusCuiPresenter)

	t.Run("mus phase shows prompt", func(t *testing.T) {
		m, players := setupMusCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Mus")
		assert.NotEmpty(t, result)
	})

	t.Run("discard phase shows prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseDiscard)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("grande betting phase shows bet prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseGrande)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("juego betting phase shows bet prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseJuego)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("showdown phase shows prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseShowdown)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end phase shows prompt", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("betting results shown when in betting phase", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseChica)
		for ri := 0; ri < domain.MusRoundCnt; ri++ {
			m.On("GetResult", ri).Return(domain.MusRoundResult{Kind: domain.MusResultDeferred, Stake: 1, Team: -1})
		}
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})
}

func TestMusCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MusCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.MusHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("mus phase hint - exchange", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.MusHint{Mus: true, Reason: "mus_exchange"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("mus phase hint - cut", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.MusHint{Mus: false, Reason: "mus_cut"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("discard phase hint with indices", func(t *testing.T) {
		m, players := setupMusCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseDiscard)
		m.On("GetHint").Return(&domain.MusHint{Indices: []int{1}, Reason: "discard_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("discard phase hint no cards to discard", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseDiscard)
		m.On("GetHint").Return(&domain.MusHint{Indices: []int{}, Reason: "discard_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("bet hint - paso", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhaseGrande)
		m.On("GetHint").Return(&domain.MusHint{Action: domain.MusActionPaso, Reason: "bet_paso"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("bet hint - envido", func(t *testing.T) {
		m, _ := setupMusCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MusPhasePares)
		m.On("GetHint").Return(&domain.MusHint{Action: domain.MusActionEnvido, Amount: 2, Reason: "bet_envido"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestMusCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MusCuiPresenter)
	m := new(interfaces.MockMusGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "mus", Detail: "You wants mus"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "mus")
}
