package interfaces

// SolitaireGame groups the undo/give-up/autocomplete methods shared by solitaire-family interfaces.
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
