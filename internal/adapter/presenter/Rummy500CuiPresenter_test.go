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

func setupRummy500CuiMock() *interfaces.MockRummy500Game {
	m := new(interfaces.MockRummy500Game)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(25)
	m.On("GetDiscardPile").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.Rummy500PhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeRummy500Players() []*domain.Rummy500Player {
	return []*domain.Rummy500Player{
		domain.NewRummy500Player(true),
		domain.NewRummy500Player(false),
	}
}

func setupRummy500CuiMockWithPlayers() (*interfaces.MockRummy500Game, []*domain.Rummy500Player) {
	m := setupRummy500CuiMock()
	players := makeRummy500Players()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestRummy500CuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.Rummy500CuiPresenter)

	t.Run("initial draw phase shows header and players", func(t *testing.T) {
		m, players := setupRummy500CuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Rummy 500")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 25枚")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "手番: あなた")
	})

	t.Run("discard pile shown", func(t *testing.T) {
		m, _ := setupRummy500CuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardPile")
		pile := []*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)}
		m.On("GetDiscardPile").Return(pile)
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札")
	})

	t.Run("empty discard pile shows empty line", func(t *testing.T) {
		m, _ := setupRummy500CuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: なし")
	})

	t.Run("play phase prompts", func(t *testing.T) {
		m, _ := setupRummy500CuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.Rummy500PhasePlay)
		result := p.Output(m, nil)
		assert.Contains(t, result, "プレイフェーズ")
	})

	t.Run("round end shows prompt", func(t *testing.T) {
		m, _ := setupRummy500CuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.Rummy500PhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end shows banner", func(t *testing.T) {
		m, _ := setupRummy500CuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupRummy500CuiMockWithPlayers()
		result := p.Output(m, errors.New("oops"))
		assert.Contains(t, result, "oops")
	})

	t.Run("player with melds shown", func(t *testing.T) {
		m, players := setupRummy500CuiMockWithPlayers()
		players[0].AddLaidMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "メルド[0]")
	})
}

func TestRummy500CuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.Rummy500CuiPresenter)
	m := new(interfaces.MockRummy500Game)
	m.On("GetGameEndFlag").Return(false)
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}

func TestRummy500CuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.Rummy500CuiPresenter)

	playPhase := func() (*interfaces.MockRummy500Game, []*domain.Rummy500Player) {
		m, players := setupRummy500CuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.Rummy500PhasePlay)
		m.On("IsHumanTurn").Return(true)
		return m, players
	}

	t.Run("lists meld candidates", func(t *testing.T) {
		m, players := playPhase()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		out := p.HintOutput(m)
		prefix := strings.SplitN(i18n.T("rummy500.hintMeld"), "{{", 2)[0]
		assert.Contains(t, out, prefix)
	})

	t.Run("advises discarding when no meld is possible", func(t *testing.T) {
		m, players := playPhase()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		assert.Contains(t, p.HintOutput(m), i18n.T("rummy500.hintNoMeld"))
	})

	t.Run("no hint during the draw phase", func(t *testing.T) {
		m, _ := setupRummy500CuiMockWithPlayers()
		m.On("IsHumanTurn").Return(true)
		assert.Contains(t, p.HintOutput(m), i18n.T("rummy500.hintNone"))
	})

	t.Run("no hint on a CPU turn", func(t *testing.T) {
		m, _ := playPhase()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("IsHumanTurn").Return(false)
		assert.Contains(t, p.HintOutput(m), i18n.T("rummy500.hintNone"))
	})
}
