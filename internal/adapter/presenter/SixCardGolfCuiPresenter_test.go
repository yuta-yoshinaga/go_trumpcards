//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// sixCardGolfDrawPending returns a human-turn DrawPending game whose grid is
// filled with the given face-up value and with drawnCard set.
func sixCardGolfDrawPending(gridValue, drawnValue int) *domain.SixCardGolf {
	g := domain.NewDefaultSixCardGolf()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.SixCardGolfPhaseDrawPending)
	p := g.GetPlayer(0)
	for i := range p.Grid {
		p.Grid[i] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignSpade, gridValue, false), FaceUp: true}
	}
	g.SetDrawnCard(domain.NewCard(domain.CardDesignSpade, drawnValue, false))
	return g
}

func TestSixCardGolfCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SixCardGolfCuiPresenter)

	t.Run("swap recommended for a low drawn card", func(t *testing.T) {
		g := sixCardGolfDrawPending(11, 1) // Jacks in grid, drew an Ace
		assert.Contains(t, p.HintOutput(g), "sw")
	})

	t.Run("discard recommended for a high drawn card", func(t *testing.T) {
		g := sixCardGolfDrawPending(1, 11) // Aces in grid, drew a Jack
		assert.Contains(t, p.HintOutput(g), i18n.T("sixcardgolf.hintDiscard"))
	})

	t.Run("column pair is called out", func(t *testing.T) {
		g := sixCardGolfDrawPending(9, 5)
		g.GetPlayer(0).Grid[3] = domain.SixCardGolfSlot{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true}
		assert.Contains(t, p.HintOutput(g), i18n.Tf("sixcardgolf.hintSwapPair", "pos", "0"))
	})

	t.Run("draw from discard for a low top card", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.SixCardGolfPhasePlayerTurn)
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)}) // King = 0
		assert.Contains(t, p.HintOutput(g), "dd")
	})

	t.Run("draw from stock for a high top card", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.SixCardGolfPhasePlayerTurn)
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)}) // 9 > 3
		assert.Contains(t, p.HintOutput(g), i18n.T("sixcardgolf.hintDrawStock"))
	})

	t.Run("no hint on a CPU turn", func(t *testing.T) {
		g := sixCardGolfDrawPending(11, 1)
		g.SetCurrentPlayerIdx(1) // CPU
		assert.Contains(t, p.HintOutput(g), i18n.T("sixcardgolf.hintNone"))
	})

	t.Run("no hint outside a decision phase", func(t *testing.T) {
		g := domain.NewDefaultSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.SixCardGolfPhaseSetup)
		assert.Contains(t, p.HintOutput(g), i18n.T("sixcardgolf.hintNone"))
	})
}
