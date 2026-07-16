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

func makeSuecaPlayers() []*domain.SuecaPlayer {
	return []*domain.SuecaPlayer{
		domain.NewSuecaPlayer(true),
		domain.NewSuecaPlayer(false),
		domain.NewSuecaPlayer(false),
		domain.NewSuecaPlayer(false),
	}
}

func setupSuecaCuiMock() *interfaces.MockSuecaGame {
	m := new(interfaces.MockSuecaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.SuecaTrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SuecaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetRoundCardPoints").Return([domain.SuecaTeamCnt]int{0, 0})
	m.On("GetTeamGamePoints").Return([domain.SuecaTeamCnt]int{0, 0})
	m.On("GetLeadPlayerIdx").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupSuecaCuiMockWithPlayers() (*interfaces.MockSuecaGame, []*domain.SuecaPlayer) {
	m := setupSuecaCuiMock()
	players := makeSuecaPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestSuecaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SuecaCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupSuecaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Sueca")
		assert.NotEmpty(t, result)
	})

	t.Run("trick end shows the winning player and team A", func(t *testing.T) {
		m, _ := setupSuecaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SuecaPhaseTrickEnd)
		// Lead player 0 (you) -> team A won the trick.
		result := p.Output(m, nil)
		assert.Contains(t, result, "（チームA）がトリックを獲得")
	})

	t.Run("trick end shows team B for an odd-index winner", func(t *testing.T) {
		m, _ := setupSuecaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLeadPlayerIdx")
		m.On("GetPhase").Return(domain.SuecaPhaseTrickEnd)
		m.On("GetLeadPlayerIdx").Return(1) // CPU 1 -> team B
		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1（チームB）がトリックを獲得")
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupSuecaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SuecaPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupSuecaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupSuecaCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestSuecaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SuecaCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupSuecaCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.SuecaHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupSuecaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.SuecaHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupSuecaCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.SuecaHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestSuecaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SuecaCuiPresenter)
	m := new(interfaces.MockSuecaGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
