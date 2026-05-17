package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FourCardPokerGame is the Four Card Poker game interface consumed by the
// usecase layer.
type FourCardPokerGame interface {
	BaseGame
	// Reset clears per-round state and reshuffles.
	Reset()
	// Bet places the ante (mandatory) and optional Aces Up sidebet, then deals.
	Bet(ante, acesUp int) error
	// Play places the Play bet at the requested ante multiplier (1, 2, or 3) and resolves.
	Play(multiplier int) error
	// Fold forfeits the ante (Aces Up is still evaluated).
	Fold() error

	// GetPlayerHand returns the 5-card player hand.
	GetPlayerHand() []*domain.Card
	// GetDealerHand returns the 6-card dealer hand.
	GetDealerHand() []*domain.Card
	// GetPlayerBest returns the best 4-card subset of the player's hand.
	GetPlayerBest() []*domain.Card
	// GetDealerBest returns the best 4-card subset of the dealer's hand.
	GetDealerBest() []*domain.Card
	// GetDealerUpCard returns the dealer's face-up card.
	GetDealerUpCard() *domain.Card
	// GetPhase returns the current phase.
	GetPhase() int
	// GetGameEndFlag reports whether the round has ended.
	GetGameEndFlag() bool
	// GetAnteBet returns the ante bet.
	GetAnteBet() int
	// GetAcesUpBet returns the Aces Up sidebet.
	GetAcesUpBet() int
	// GetPlayBet returns the Play bet.
	GetPlayBet() int
	// GetPlayMultiplier returns the chosen Play multiplier.
	GetPlayMultiplier() int
	// GetResult returns the round result.
	GetResult() domain.GameResult
	// GetAntePayout returns the ante payout.
	GetAntePayout() int
	// GetPlayPayout returns the play payout.
	GetPlayPayout() int
	// GetAnteBonusPayout returns the automatic ante bonus payout.
	GetAnteBonusPayout() int
	// GetAcesUpPayout returns the Aces Up sidebet payout.
	GetAcesUpPayout() int
	// GetTotalPayout returns the sum of all payouts.
	GetTotalPayout() int
	// GetPlayerHandRank returns the player's hand rank.
	GetPlayerHandRank() int
	// GetDealerHandRank returns the dealer's hand rank.
	GetDealerHandRank() int
	// GetChips returns the current chip stack.
	GetChips() int
}
