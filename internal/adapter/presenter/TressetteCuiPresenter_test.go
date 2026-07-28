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

func setupTressetteCuiMock() *interfaces.MockTressetteGame {
	m := new(interfaces.MockTressetteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTeamScores").Return([domain.TressetteTeamCnt]int{0, 0})
	m.On("GetTeamRoundThirds").Return([domain.TressetteTeamCnt]int{0, 0})
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TressettePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupTressetteCuiMockWithPlayers() (*interfaces.MockTressetteGame, []*domain.TressettePlayer) {
	m := setupTressetteCuiMock()
	players := makeTressettePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestTressetteCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TressetteCuiPresenter)

	t.Run("play phase shows player info and prompt", func(t *testing.T) {
		m, players := setupTressetteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Tressette")
		assert.NotEmpty(t, result)
		// The thirds-conversion rule is explained alongside the score line.
		assert.Contains(t, result, "サード: 3トリック")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TressettePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows team breakdown and last-trick team", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamRoundThirds")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLeadPlayerIdx")
		m.On("GetPhase").Return(domain.TressettePhaseRoundEnd)
		m.On("GetTeamRoundThirds").Return([domain.TressetteTeamCnt]int{7, 4})
		m.On("GetLeadPlayerIdx").Return(1) // team B took the last trick
		result := p.Output(m, nil)
		assert.Contains(t, result, strings.Split(i18n.T("tressette.roundBreakdown"), "{{")[0])
		// Each team's thirds appear (7 and 4).
		assert.Contains(t, result, "thirds7")
		assert.Contains(t, result, "thirds4")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestTressetteCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TressetteCuiPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, players := setupTressetteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.On("GetHint").Return(&domain.TressetteHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.TressetteHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestTressetteCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TressetteCuiPresenter)
	m := new(interfaces.MockTressetteGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠3"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
