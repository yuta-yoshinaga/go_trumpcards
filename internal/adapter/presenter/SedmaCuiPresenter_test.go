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

func makeSedmaPlayers() []*domain.SedmaPlayer {
	return []*domain.SedmaPlayer{
		domain.NewSedmaPlayer(true),
		domain.NewSedmaPlayer(false),
		domain.NewSedmaPlayer(false),
		domain.NewSedmaPlayer(false),
	}
}

func setupSedmaCuiMock() *interfaces.MockSedmaGame {
	m := new(interfaces.MockSedmaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.SedmaTrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SedmaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetRoundCardPoints").Return([domain.SedmaTeamCnt]int{0, 0})
	m.On("GetTeamScores").Return([domain.SedmaTeamCnt]int{0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSedmaCuiMockWithPlayers() (*interfaces.MockSedmaGame, []*domain.SedmaPlayer) {
	m := setupSedmaCuiMock()
	players := makeSedmaPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestSedmaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SedmaCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupSedmaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end shows winner and captured points", func(t *testing.T) {
		m, _ := setupSedmaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetPhase").Return(domain.SedmaPhaseTrickEnd)
		m.On("GetLeadPlayerIdx").Return(2)
		// One ace and one ten (10 pts each) plus a plain card (0 pts) → 20 pts,
		// exercising all three branches of the point-counting loop.
		m.On("GetCurrentTrick").Return([]*domain.SedmaTrickCard{
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 8, false)},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "20")
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupSedmaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SedmaPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupSedmaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupSedmaCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestSedmaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SedmaCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSedmaCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.SedmaHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupSedmaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.SedmaHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupSedmaCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.SedmaHint{CardIndices: nil, Reason: "capture"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestSedmaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SedmaCuiPresenter)
	m := new(interfaces.MockSedmaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
