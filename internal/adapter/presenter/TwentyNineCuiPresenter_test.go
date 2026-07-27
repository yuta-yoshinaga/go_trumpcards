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

func makeTwentyNinePlayers() []*domain.TwentyNinePlayer {
	return []*domain.TwentyNinePlayer{
		domain.NewTwentyNinePlayer(true),
		domain.NewTwentyNinePlayer(false),
		domain.NewTwentyNinePlayer(false),
		domain.NewTwentyNinePlayer(false),
	}
}

func setupTwentyNineCuiMock() *interfaces.MockTwentyNineGame {
	m := new(interfaces.MockTwentyNineGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpRevealed").Return(true)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TwentyNinePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetContract").Return(domain.TwentyNineBidTwenty)
	m.On("GetBids").Return([domain.TwentyNinePlayerCnt]domain.TwentyNineBid{domain.TwentyNineBidTwenty, domain.TwentyNineBidPass, domain.TwentyNineBidPass, domain.TwentyNineBidPass})
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetTeamScores").Return([domain.TwentyNineTeamCnt]int{0, 0})
	m.On("GetRoundTeamPoints").Return([domain.TwentyNineTeamCnt]int{0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupTwentyNineCuiMockWithPlayers() (*interfaces.MockTwentyNineGame, []*domain.TwentyNinePlayer) {
	m := setupTwentyNineCuiMock()
	players := makeTwentyNinePlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestTwentyNineCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TwentyNineCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupTwentyNineCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Twenty-Nine")
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt hides trump", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpRevealed")
		m.On("GetPhase").Return(domain.TwentyNinePhaseBid)
		m.On("GetDeclarerIdx").Return(-1)
		m.On("GetTrumpSuit").Return(0)
		m.On("GetTrumpRevealed").Return(false)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwentyNinePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TwentyNinePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestTwentyNineCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TwentyNineCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.TwentyNineHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupTwentyNineCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.TwentyNineHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupTwentyNineCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.TwentyNineHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestTwentyNineCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TwentyNineCuiPresenter)
	m := new(interfaces.MockTwentyNineGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
