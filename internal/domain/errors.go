package domain

import "errors"

// Sentinel errors for domain-layer validation failures.
var (
	ErrWrongPhase        = errors.New("wrong game phase")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInsufficientChips = errors.New("insufficient chips")
	ErrGameEnded         = errors.New("game has ended")
	ErrNotHumanTurn      = errors.New("not human player's turn")
	ErrInvalidCard       = errors.New("invalid card")
	ErrInvalidPlay       = errors.New("invalid play")
	ErrDeckExhausted     = errors.New("deck exhausted")
	ErrCannotPass        = errors.New("cannot pass")
	ErrHandFinished      = errors.New("hand already finished")
)
