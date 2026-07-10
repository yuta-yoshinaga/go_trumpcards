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

func setupPrsiCuiMock() *interfaces.MockPrsiGame {
	m := new(interfaces.MockPrsiGame)
	m.On("GetDrawPileCount").Return(11)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetPenaltyDrawCount").Return(0)
	m.On("GetPendingSkips").Return(0)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PrsiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupPrsiCuiMockWithPlayers() (*interfaces.MockPrsiGame, []*domain.PrsiPlayer) {
	m := setupPrsiCuiMock()
	players := makePrsiPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPrsiCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PrsiCuiPresenter)

	t.Run("play phase renders header, hand, prompt", func(t *testing.T) {
		m, players := setupPrsiCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 8, false))

		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "11") // stock count
	})

	t.Run("discard top with penalty", func(t *testing.T) {
		m, _ := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPenaltyDrawCount")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		m.On("GetPenaltyDrawCount").Return(4)

		out := p.Output(m, nil)
		assert.Contains(t, out, "4") // penalty count appears
	})

	t.Run("error block rendered", func(t *testing.T) {
		m, _ := setupPrsiCuiMockWithPlayers()
		out := p.Output(m, errors.New("boom"))
		assert.Contains(t, out, "boom")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})
}

func TestPrsiCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PrsiCuiPresenter)
	m := new(interfaces.MockPrsiGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays SPADE 7"},
	})
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "play")
}
