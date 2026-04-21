package interfaces

// SolitaireGame groups the methods shared by the solitaire-family interfaces
// (Klondike, Spider, FreeCell, FortyThieves, Yukon, Scorpion). Extracted per
// issue #1461 to reduce duplication across the 10+ solitaire interfaces.
//
// Games missing one of these methods (e.g. Pyramid has no AutoComplete,
// Canfield has no UndoToEscape) do not embed SolitaireGame — they keep
// their current method lists. Adding a partial-SolitaireGame variant would
// introduce combinatorial interface sprawl without clear benefit, so the
// pragmatic choice is "embed when the full set applies".
type SolitaireGame interface {
	BaseGame
	// GiveUp marks the current game as abandoned.
	GiveUp()
	// Undo reverses the last move; returns an error if nothing can be undone.
	Undo() error
	// CanUndo reports whether at least one move is available to undo.
	CanUndo() bool
	// UndoN reverses the last n moves; no-ops past the head of the history.
	UndoN(n int) error
	// UndoToEscape returns the number of undos required to escape a stalemate.
	UndoToEscape() int
	// AutoComplete runs the auto-complete routine when the board is trivially
	// solvable (all cards face-up, no blocking moves).
	AutoComplete() error
}
