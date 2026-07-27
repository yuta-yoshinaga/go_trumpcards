//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupUltiCuiMock() *interfaces.MockUltiGame {
	m := new(interfaces.MockUltiGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetContract").Return(domain.UltiContractParty)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.UltiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetOutcome").Return(domain.UltiOutcomeWin)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayerCoins").Return([domain.UltiPlayerCnt]int{0, 0, 0})
	m.On("GetCardPoints", mock.Anything).Return(0)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupUltiCuiMockWithPlayers() (*interfaces.MockUltiGame, []*domain.UltiPlayer) {
	m := setupUltiCuiMock()
	players := makeUltiPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestUltiCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.UltiCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupUltiCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "ウルティ") // translated helpTitle
		assert.Contains(t, result, "契約")   // round line mentions contract
		assert.Contains(t, result, "切り札")  // trump
		assert.Contains(t, result, "トリック") // trick
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupUltiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.UltiPhaseBid)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "契約宣言")
	})

	t.Run("discard phase prompt", func(t *testing.T) {
		m, _ := setupUltiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.UltiPhaseDiscard)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "タロン")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupUltiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.UltiPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupUltiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.UltiPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "成功") // outcome win label
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupUltiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupUltiCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestUltiCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.UltiCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupUltiCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.UltiHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupUltiCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.UltiHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("discard hint uses declarer hand", func(t *testing.T) {
		m, players := setupUltiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.UltiPhaseDiscard)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.UltiHint{CardIndices: []int{0}, Reason: "discard_weak"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupUltiCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.UltiHint{CardIndices: nil, Reason: "bid_party"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestUltiCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.UltiCuiPresenter)
	m := new(interfaces.MockUltiGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}
