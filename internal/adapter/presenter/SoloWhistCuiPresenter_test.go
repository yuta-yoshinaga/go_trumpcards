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

func makeSoloWhistPlayers() []*domain.SoloWhistPlayer {
	return []*domain.SoloWhistPlayer{
		domain.NewSoloWhistPlayer(true),
		domain.NewSoloWhistPlayer(false),
		domain.NewSoloWhistPlayer(false),
		domain.NewSoloWhistPlayer(false),
	}
}

func setupSoloWhistCuiMock() *interfaces.MockSoloWhistGame {
	m := new(interfaces.MockSoloWhistGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SoloWhistPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.SoloWhistBidSolo)
	m.On("GetBids").Return([domain.SoloWhistPlayerCnt]domain.SoloWhistBid{domain.SoloWhistBidSolo, domain.SoloWhistBidPass, domain.SoloWhistBidPass, domain.SoloWhistBidPass})
	m.On("GetBidDone").Return([domain.SoloWhistPlayerCnt]bool{true, true, false, false})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayerScores").Return([domain.SoloWhistPlayerCnt]int{0, 0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSoloWhistCuiMockWithPlayers() (*interfaces.MockSoloWhistGame, []*domain.SoloWhistPlayer) {
	m := setupSoloWhistCuiMock()
	players := makeSoloWhistPlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestSoloWhistCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SoloWhistCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupSoloWhistCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Solo Whist")
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupSoloWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.On("GetPhase").Return(domain.SoloWhistPhaseBid)
		m.On("GetDeclarerIdx").Return(-1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// The bid-status line lists every player; an un-bid player renders "-".
		assert.Contains(t, result, strings.Split(i18n.T("solowhist.bidHistory"), "{{")[0])
		assert.Contains(t, result, "=-")
	})

	t.Run("misere shows no trump", func(t *testing.T) {
		m, _ := setupSoloWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupSoloWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SoloWhistPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupSoloWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SoloWhistPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupSoloWhistCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupSoloWhistCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestSoloWhistCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SoloWhistCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSoloWhistCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.SoloWhistHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupSoloWhistCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.SoloWhistHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupSoloWhistCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.SoloWhistHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestSoloWhistCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SoloWhistCuiPresenter)
	m := new(interfaces.MockSoloWhistGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
