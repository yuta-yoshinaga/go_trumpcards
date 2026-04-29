package controller

// SolitaireWebOutputBase carries the five JSON fields that every solitaire
// WebOutput exposes: Phase, MoveCount, CanUndo, IsStalemate, UndoToEscape.
// Solitaire-family WebOutput structs embed this type so:
//
//   - Adding a new field is a one-line change here, not 11 controller edits.
//   - The presenters' `buildBase` helpers populate the bundle through a
//     single helper (see populateSolitaireBase in the presenter package),
//     replacing the 5-line copy/paste that previously lived in each
//     `internal/adapter/presenter/<Game>WebPresenter.go`.
//
// Embedded struct fields are flattened by encoding/json, so the wire
// format stays exactly the same as the prior "five top-level fields"
// shape — frontend and tests do not need any change.
//
// See issue #1563.
type SolitaireWebOutputBase struct {
	// Phase is the current game phase as a domain-defined enum, normalized
	// to int because each solitaire has its own Phase enum type.
	Phase int `json:"phase"`
	// MoveCount is the total number of moves the player has made this game.
	MoveCount int `json:"moveCount"`
	// CanUndo reports whether at least one move is available to undo.
	CanUndo bool `json:"canUndo"`
	// IsStalemate reports whether the board has no legal forward moves.
	IsStalemate bool `json:"isStalemate"`
	// UndoToEscape is the number of undos required to escape a stalemate;
	// 0 when the game is not in stalemate.
	UndoToEscape int `json:"undoToEscape"`
}
