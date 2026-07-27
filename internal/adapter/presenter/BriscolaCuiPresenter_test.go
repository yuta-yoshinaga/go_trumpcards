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

func setupBriscolaCuiMock(trumpCard *domain.Card) *interfaces.MockBriscolaGame {
	m := new(interfaces.MockBriscolaGame)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BriscolaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetStockRemaining").Return(33)
	m.On("GetWinnerIdx").Return(-1)
	return m
}

func setupBriscolaCuiMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockBriscolaGame, []*domain.BriscolaPlayer) {
	m := setupBriscolaCuiMock(trumpCard)
	players := []*domain.BriscolaPlayer{
		domain.NewBriscolaPlayer(true),
		domain.NewBriscolaPlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayerPoints", 0).Return(15)
	m.On("GetPlayerPoints", 1).Return(5)
	return m, players
}

func TestBriscolaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.BriscolaCuiPresenter)

	t.Run("initial state shows header, trump, points", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, players := setupBriscolaCuiMockWithPlayers(trump)
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))

		out := p.Output(m, nil)
		assert.Contains(t, out, "Briscola")
		assert.Contains(t, out, "トリック: 1")
		assert.Contains(t, out, "山札: 33枚")
		assert.Contains(t, out, "得点: あなた=15  CPU=5")
		assert.Contains(t, out, "あなた:")
		assert.Contains(t, out, "CPU 1:")
		assert.Contains(t, out, "play <idx>")
	})

	t.Run("trump card exhausted", func(t *testing.T) {
		m, _ := setupBriscolaCuiMockWithPlayers(nil)
		out := p.Output(m, nil)
		assert.Contains(t, out, "使い切り")
	})

	t.Run("error is rendered", func(t *testing.T) {
		m, _ := setupBriscolaCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		out := p.Output(m, errors.New("kaboom"))
		assert.Contains(t, out, "kaboom")
	})

	t.Run("trick-end prompt", func(t *testing.T) {
		m, _ := setupBriscolaCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BriscolaPhaseTrickEnd)
		out := p.Output(m, nil)
		assert.Contains(t, out, "トリック終了")
	})

	t.Run("game end p0 banner", func(t *testing.T) {
		m, _ := setupBriscolaCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.On("GetPlayerPoints", 0).Return(70)
		m.On("GetPlayerPoints", 1).Return(50)

		out := p.Output(m, nil)
		assert.Contains(t, out, "あなたの勝利")
		assert.Contains(t, out, "(70-50)")
	})

	t.Run("game end p1 banner", func(t *testing.T) {
		m, _ := setupBriscolaCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.On("GetPlayerPoints", 0).Return(40)
		m.On("GetPlayerPoints", 1).Return(80)

		out := p.Output(m, nil)
		assert.Contains(t, out, "CPUの勝利")
	})

	t.Run("game end tie banner", func(t *testing.T) {
		m, _ := setupBriscolaCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(-1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.On("GetPlayerPoints", 0).Return(60)
		m.On("GetPlayerPoints", 1).Return(60)

		out := p.Output(m, nil)
		assert.Contains(t, out, "引き分け")
	})
}

func TestBriscolaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BriscolaCuiPresenter)

	t.Run("hint shows card and reason", func(t *testing.T) {
		m, players := setupBriscolaCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.BriscolaHint{CardIndex: &idx, Reason: "follow_cut"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
		assert.Contains(t, out, "トランプでカット")
	})

	t.Run("hint nil falls back to hintNone", func(t *testing.T) {
		m, _ := setupBriscolaCuiMockWithPlayers(nil)
		m.On("GetHint").Return((*domain.BriscolaHint)(nil))
		out := p.HintOutput(m)
		assert.Contains(t, out, "ヒントはありません")
	})

	t.Run("hint with unknown reason falls back to shared lookup", func(t *testing.T) {
		m, players := setupBriscolaCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.BriscolaHint{CardIndex: &idx, Reason: "unknown_reason"})
		out := p.HintOutput(m)
		assert.NotEmpty(t, out)
	})
}

func TestBriscolaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BriscolaCuiPresenter)
	m, _ := setupBriscolaCuiMockWithPlayers(nil)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotNil(t, out)
}
