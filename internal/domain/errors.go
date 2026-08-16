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
//
// A DomainError carries EITHER a finished phrase (Message) or an i18n key
// (Code, with optional Params). The phrase form is the older one and is fixed
// in whatever language it was written in -- MissMilligan's were English, most
// other games' are Japanese -- and every presenter prints it verbatim, so it
// ignores the player's locale (#5556). Prefer the key form for new errors and
// let the presenter's i18n build the sentence.
type DomainError struct {
	Sentinel error
	Message  string
	Code     string
	Params   map[string]string
}

// Error returns the user-facing message without the sentinel prefix. When the
// error names an i18n key instead, the key is returned: a caller that does not
// know about codes still gets something identifiable in a log, and an
// untranslated key on screen is obvious rather than silently blank.
func (de *DomainError) Error() string {
	if de.Message == "" {
		return de.Code
	}
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

// NewDomainErrorCode creates a DomainError that names an i18n key instead of a
// finished phrase, so the presenter renders it in the player's locale.
func NewDomainErrorCode(sentinel error, code string, params map[string]string) *DomainError {
	return &DomainError{Sentinel: sentinel, Code: code, Params: params}
}

// MessageCode returns the i18n key this error names, or "" when it carries a
// finished phrase instead. An empty answer means "do not try to translate me":
// looking up "" would silently render the key itself.
func (de *DomainError) MessageCode() string { return de.Code }

// MessageParams returns the interpolation values for MessageCode.
func (de *DomainError) MessageParams() map[string]string { return de.Params }

// ErrorMessageCode reports the i18n key and params an error names, if any.
// Plain errors and phrase-carrying DomainErrors answer "", nil -- which is how
// a presenter decides between translating and printing what it was handed.
func ErrorMessageCode(err error) (string, map[string]string) {
	var de *DomainError
	if errors.As(err, &de) && de.Code != "" {
		return de.Code, de.Params
	}
	return "", nil
}
