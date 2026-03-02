package domain

import "errors"

// Sentinel errors for domain-layer validation failures.
var (
	// ErrWrongPhase indicates an action was attempted in the wrong game phase.
	ErrWrongPhase = errors.New("wrong game phase")
	// ErrInvalidAmount indicates a bet or raise amount is out of range.
	ErrInvalidAmount = errors.New("invalid amount")
	// ErrInsufficientChips indicates the player lacks chips for the action.
	ErrInsufficientChips = errors.New("insufficient chips")
	// ErrGameEnded indicates the game has already finished.
	ErrGameEnded = errors.New("game has ended")
	// ErrNotHumanTurn indicates it is not the human player's turn.
	ErrNotHumanTurn = errors.New("not human player's turn")
	// ErrInvalidCard indicates the selected card index is out of range or invalid.
	ErrInvalidCard = errors.New("invalid card")
	// ErrInvalidPlay indicates the attempted play violates game rules.
	ErrInvalidPlay = errors.New("invalid play")
	// ErrDeckExhausted indicates the deck has no more cards to draw.
	ErrDeckExhausted = errors.New("deck exhausted")
	// ErrCannotPass indicates the player is not allowed to pass.
	ErrCannotPass = errors.New("cannot pass")
	// ErrHandFinished indicates the hand has already been resolved.
	ErrHandFinished = errors.New("hand already finished")
	// ErrInvalidIndices indicates the provided index permutation is invalid.
	ErrInvalidIndices = errors.New("invalid indices")
)

// DomainError wraps a sentinel error with a user-facing message.
// Error() returns only the user-facing message, while Unwrap() returns
// the sentinel so that errors.Is() works for programmatic checks.
type DomainError struct {
	Sentinel error
	Message  string
}

// Error returns the user-facing message without the sentinel prefix.
func (de *DomainError) Error() string {
	return de.Message
}

// Unwrap returns the sentinel error for use with errors.Is().
func (de *DomainError) Unwrap() error {
	return de.Sentinel
}

// NewDomainError creates a DomainError wrapping the given sentinel.
func NewDomainError(sentinel error, message string) *DomainError {
	return &DomainError{Sentinel: sentinel, Message: message}
}
