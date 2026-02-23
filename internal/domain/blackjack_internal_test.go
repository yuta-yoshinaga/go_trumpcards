package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBlackJack_PlayerDoubleDown_Bust_Deterministic exercises lines 234-236 of
// BlackJack.go: when score >= 22 after the double-down draw, the hand should be
// marked as busted. Instead of retrying random shuffles, the deck is directly
// stacked so the next drawn card is guaranteed to be a King (BJ value 10),
// pushing the player's score from 20 to 30 (bust).
func TestBlackJack_PlayerDoubleDown_Bust_Deterministic(t *testing.T) {
	tc := NewTrumpCards(0)
	player := NewBlackJackPlayer()
	dealer := NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := NewBlackJack(tc, player, dealer)

	// Set up player hand with score 20 (10 + King=10)
	hand := bj.playerHands[0]
	hand.SetBet(100)
	hand.AddCard(NewCard(CardDesignSpade, 10, false))
	hand.AddCard(NewCard(CardDesignHeart, 13, false)) // King = BJ value 10, score = 20

	// Set up dealer cards (score 17, so dealer won't draw if reached)
	dealer.AddCard(NewCard(CardDesignClover, 10, false))
	dealer.AddCard(NewCard(CardDesignDiamond, 7, false))

	// Stack the deck: place a King at the current draw position so that
	// drawCard() returns a card with BJ value 10, causing 20 + 10 = 30 (bust).
	bj.trumpCards.deck[0] = NewCard(CardDesignClover, 13, false) // King = BJ value 10
	bj.trumpCards.deckDrawCnt = 0

	// Set phase to action so PlayerDoubleDown is allowed
	bj.phase = BJPhaseAction

	err := bj.PlayerDoubleDown()
	assert.NoError(t, err)

	// Verify double-down mechanics
	assert.Equal(t, 200, hand.GetBet(), "bet should be doubled")
	assert.True(t, hand.IsDoubled(), "hand should be marked as doubled")
	assert.Equal(t, 3, hand.GetCardsSize(), "hand should have 3 cards after DD")

	// Verify bust-specific state (lines 234-236)
	assert.True(t, hand.IsBusted(), "hand should be busted when score >= 22")
	assert.False(t, hand.IsStood(), "busted hand should not be stood")
	assert.True(t, hand.GetScore() >= 22, "score should be >= 22 for bust")

	// Game should end (single hand, all finished → dealerPlay → allBusted → endGame)
	assert.True(t, bj.gameEndFlag)
	assert.Equal(t, BJPhaseEnd, bj.phase)

	// Player should lose (busted)
	assert.Equal(t, GameResultLose, bj.GameJudgment())
}

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
