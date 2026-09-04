//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BadugiGame is the boundary interface between the usecase layer and the
// Badugi domain implementation. Kept small (only what PokerInteractor-style
// callers actually need) to make mocks cheap.
type BadugiGame interface {
	BaseGame

	// Reset clears the hand state and deals fresh cards. Returns domain
	// validation errors surfaced to the caller as-is.
	Reset() error
	// PlayerAction applies a human betting action. humanPlayMs is the
	// deliberation time in milliseconds (0 = not measured).
	PlayerAction(action, amount, humanPlayMs int) error
	// PlayerExchange replaces the cards at the given hand indices during the
	// draw phase. humanPlayMs is the deliberation time in milliseconds (0 = not measured).
	PlayerExchange(indices []int, humanPlayMs int) error
	// PlayerStand stands pat during a draw phase (shorthand for
	// PlayerExchange(nil, humanPlayMs)).
	PlayerStand(humanPlayMs int) error

	// GetPlayers returns the seated players.
	GetPlayers() []*domain.BadugiPlayer
	// GetPhase returns the current phase constant (BadugiPhase*).
	GetPhase() int
	// GetDrawIndex returns the current draw round counter (0..3).
	GetDrawIndex() int
	// GetPot returns the current pot value.
	GetPot() int
	// GetSidePots returns the side pots (populated at showdown).
	GetSidePots() []domain.SidePot
	// GetDealerIdx returns the button seat index.
	GetDealerIdx() int
	// GetCurrentTurn returns the index of the seat expected to act next.
	GetCurrentTurn() int
	// GetGameEndFlag reports whether the current hand has been resolved.
	GetGameEndFlag() bool
	// GetLastBet returns the last bet size in the current round.
	GetLastBet() int
	// GetMinRaise returns the minimum legal raise increment.
	GetMinRaise() int
	// GetRaiseCount returns the raise count for the current round.
	GetRaiseCount() int
	// GetAnte returns the configured ante value.
	GetAnte() int
	// GetRoundResults returns the showdown results for the most recent hand.
	GetRoundResults() []domain.BadugiResult
	// GetCpuActions returns the CPU action log for the current hand.
	GetCpuActions() []domain.BadugiCpuAction
	// GetCpuExchanges returns the CPU draw log for the current hand.
	GetCpuExchanges() []domain.BadugiCpuExchange
	// GetConfig returns a copy of the active configuration.
	GetConfig() domain.BadugiConfig
	// SetConfig replaces the active configuration.
	SetConfig(cfg domain.BadugiConfig)

	// GetHumanProfile returns the meta-AI profile (nil if disabled).
	GetHumanProfile() *domain.BettingHumanProfile
	// ResetProfile clears the meta-AI profile.
	ResetProfile()
	// ExportProfile returns a marshalable copy of the profile, or nil.
	ExportProfile() any
	// ImportProfile loads a profile from JSON bytes.
	ImportProfile(data []byte) error
}
