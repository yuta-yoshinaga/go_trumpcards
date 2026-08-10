//go:build !js || !wasm || solo || extra || extra3 || extra2 || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"

// solitairePopulatable is the runtime contract every solitaire game type
// satisfies for the five fields shared by their Web JSON output. Phase is
// intentionally NOT here because each game declares its own typed enum
// (e.g. domain.CalculationPhase, domain.AccordionPhase); the presenter
// passes the int-cast value at the call site so the enum's compile-time
// safety is preserved.
//
// Defined locally rather than added to interfaces.SolitaireGame because
// the four methods listed here are sufficient for *output* purposes —
// the broader SolitaireGame interface drags in GiveUp/Undo/AutoComplete
// which the helper does not need.
type solitairePopulatable interface {
	GetMoveCount() int
	CanUndo() bool
	IsStalemate() bool
	UndoToEscape() int
}

// populateSolitaireBase fills the five common solitaire fields on the
// embedded SolitaireWebOutputBase. Replaces the 5-line copy/paste that
// previously lived in every `<Game>WebPresenter.buildBase`. See issue
// #1563.
func populateSolitaireBase(b *controller.SolitaireWebOutputBase, g solitairePopulatable, phase int) {
	b.Phase = phase
	b.MoveCount = g.GetMoveCount()
	b.CanUndo = g.CanUndo()
	b.IsStalemate = g.IsStalemate()
	b.UndoToEscape = g.UndoToEscape()
}
