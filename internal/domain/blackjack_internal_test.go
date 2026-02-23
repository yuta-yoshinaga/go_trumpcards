package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBlackJack_PlayerBet_DealerAceTriggersInsurance_Deterministic exercises
// lines 131-134 of BlackJack.go: when the dealer's first card is an Ace
// (value==1), insuranceAvailable is set to true and phase becomes
// BJPhaseInsurance. Instead of relying on random shuffles, the deck is
// directly stacked so the dealer's first dealt card is guaranteed to be an Ace.
func TestBlackJack_PlayerBet_DealerAceTriggersInsurance_Deterministic(t *testing.T) {
	tc := NewTrumpCards(0)
	player := NewBlackJackPlayer()
	dealer := NewBlackJackPlayer()
	player.SetChips(BJDefaultChips)
	dealer.SetChips(BJDefaultChips)
	bj := NewBlackJack(tc, player, dealer)

	// Stack the deck so that the deal produces a known layout.
	// PlayerBet deals cards in this order (2 iterations of the loop):
	//   deck[0] -> player, deck[1] -> dealer,
	//   deck[2] -> player, deck[3] -> dealer
	// We place an Ace at deck[1] so dealer's first card is an Ace.
	bj.trumpCards.deck[0] = NewCard(CardDesignSpade, 10, false)  // player card 1
	bj.trumpCards.deck[1] = NewCard(CardDesignHeart, 1, false)   // dealer card 1 (Ace)
	bj.trumpCards.deck[2] = NewCard(CardDesignClover, 9, false)  // player card 2
	bj.trumpCards.deck[3] = NewCard(CardDesignDiamond, 7, false) // dealer card 2 (non-10 to avoid natural BJ)
	bj.trumpCards.deckDrawCnt = 0

	// Reset phase to Bet so PlayerBet can proceed
	bj.phase = BJPhaseBet

	err := bj.PlayerBet(BJMinBet)
	assert.NoError(t, err)

	// Verify insurance-specific state
	assert.True(t, bj.insuranceAvailable, "insuranceAvailable should be true when dealer shows Ace")
	assert.Equal(t, BJPhaseInsurance, bj.phase)

	// Confirm dealer's first card is an Ace
	dealerCard := bj.dealer.GetCard(0)
	assert.NotNil(t, dealerCard)
	assert.Equal(t, 1, dealerCard.GetValue(), "dealer's first card should be an Ace")

	// Confirm game has not ended yet (insurance decision pending)
	assert.False(t, bj.gameEndFlag)
}
