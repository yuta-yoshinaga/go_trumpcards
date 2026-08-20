//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// acesOnMockFoundations mirrors the opening board: one Ace on each foundation.
func acesOnMockFoundations() [domain.AuldLangSyneFoundationCnt][]*domain.Card {
	var foundations [domain.AuldLangSyneFoundationCnt][]*domain.Card
	for i := range domain.AuldLangSyneFoundationCnt {
		foundations[i] = []*domain.Card{domain.NewCard(i, 1, false)}
	}
	return foundations
}

func setupAuldLangSyneCuiMockDefaults(g *interfaces.MockAuldLangSyneGame) {
	g.On("GetPhase").Return(domain.AuldLangSynePhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("GetFoundations").Return(acesOnMockFoundations()).Maybe()

	var wastes [domain.AuldLangSyneWasteCnt][]*domain.Card
	wastes[0] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 11, false)}
	g.On("GetWastes").Return(wastes).Maybe()
}

func TestAuldLangSyneCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		setupAuldLangSyneCuiMockDefaults(g)
		p := new(AuldLangSyneCuiPresenter)

		result := p.Output(g, nil)
		assert.Contains(t, result, "Auld Lang Syne")
		assert.Contains(t, result, "[F0")
		assert.Contains(t, result, "[W0]")
		assert.Contains(t, result, "ストック")
	})

	// The deal count is the readout Sir Tommy's "next card" line cannot be: the
	// player cannot see what is coming, only how many deals remain.
	t.Run("stock line reports deals remaining, not the next card", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetStockCount").Return(44)
		setupAuldLangSyneCuiMockDefaults(g)
		p := new(AuldLangSyneCuiPresenter)

		result := p.Output(g, nil)
		assert.Contains(t, result, "44")
		assert.Contains(t, result, "11", "44 cards / 4 wastes = 11 deals left")
	})

	t.Run("next required rank per foundation", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)

		var foundations [domain.AuldLangSyneFoundationCnt][]*domain.Card
		// F0: empty -> A (only reachable from a corrupt snapshot), F1: on a 5 -> 6,
		// F2: 13 cards -> complete, F3: on an Ace -> 2.
		foundations[1] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)}
		for v := 1; v <= domain.CardValueMax; v++ {
			foundations[2] = append(foundations[2], domain.NewCard(domain.CardDesignClover, v, false))
		}
		foundations[3] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		// **Register before the defaults.** testify uses the first matching
		// expectation, so calling setup first would let the `.Maybe()` pile win.
		g.On("GetFoundations").Return(foundations)
		setupAuldLangSyneCuiMockDefaults(g)
		p := new(AuldLangSyneCuiPresenter)

		result := p.Output(g, nil)
		assert.Contains(t, result, "完成", "a 13-card foundation reads as complete")
		assert.Contains(t, result, "次:", "the other foundations announce their next rank")
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0).Maybe()
		setupAuldLangSyneCuiMockDefaults(g)
		p := new(AuldLangSyneCuiPresenter)

		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game clear", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetPhase").Return(domain.AuldLangSynePhaseGameClear)
		setupAuldLangSyneCuiMockDefaults(g)
		p := new(AuldLangSyneCuiPresenter)

		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetPhase").Return(domain.AuldLangSynePhaseGameOver)
		setupAuldLangSyneCuiMockDefaults(g)
		p := new(AuldLangSyneCuiPresenter)

		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("error is surfaced", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		setupAuldLangSyneCuiMockDefaults(g)
		p := new(AuldLangSyneCuiPresenter)

		assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
	})

	t.Run("empty wastes and stock", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		var wastes [domain.AuldLangSyneWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes)
		g.On("GetStockCount").Return(0)
		setupAuldLangSyneCuiMockDefaults(g)
		p := new(AuldLangSyneCuiPresenter)

		assert.Contains(t, p.Output(g, nil), "(空)")
	})
}

func TestAuldLangSyneCuiPresenter_HintOutput(t *testing.T) {
	t.Run("waste hint", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetHint").Return(&domain.AuldLangSyneHint{WasteIdx: 2, FoundationIdx: 1})
		p := new(AuldLangSyneCuiPresenter)

		out := p.HintOutput(g)
		assert.Contains(t, out, "2")
		assert.Contains(t, out, "1")
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetHint").Return(nil)
		p := new(AuldLangSyneCuiPresenter)

		assert.NotEmpty(t, p.HintOutput(g))
	})
}

func TestAuldLangSyneCuiPresenter_ActionLogOutput(t *testing.T) {
	// While the game is still running the log is withheld, matching every other
	// solitaire: the transcript is a post-mortem, not a live crib sheet.
	t.Run("withheld while playing", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetPhase").Return(domain.AuldLangSynePhasePlaying)
		p := new(AuldLangSyneCuiPresenter)

		assert.Equal(t, actionLogToText(nil), p.ActionLogOutput(g))
	})

	t.Run("emitted once ended", func(t *testing.T) {
		g := new(interfaces.MockAuldLangSyneGame)
		g.On("GetPhase").Return(domain.AuldLangSynePhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "deal", Detail: "4枚を配りました"},
		})
		p := new(AuldLangSyneCuiPresenter)

		assert.Contains(t, p.ActionLogOutput(g), "deal")
	})
}
