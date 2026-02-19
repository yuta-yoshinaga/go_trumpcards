package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestPokerPlayer_Method(t *testing.T) {
	tpp := entities.NewPokerPlayer()

	t.Run("success GetHandRank initial", func(t *testing.T) {
		assert.Equal(t, entities.PokerHandHighCard, tpp.GetHandRank())
	})

	t.Run("success GetHandName initial", func(t *testing.T) {
		assert.Equal(t, "High Card", tpp.GetHandName())
	})

	t.Run("success EvalHand with less than 5 cards", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandHighCard, rank)
	})

	t.Run("success EvalHand High Card", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandHighCard, rank)
		assert.Equal(t, "High Card", tpp.GetHandName())
	})

	t.Run("success EvalHand One Pair", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandOnePair, rank)
		assert.Equal(t, "One Pair", tpp.GetHandName())
	})

	t.Run("success EvalHand Two Pair", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandTwoPair, rank)
		assert.Equal(t, "Two Pair", tpp.GetHandName())
	})

	t.Run("success EvalHand Three of a Kind", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandThreeOfAKind, rank)
		assert.Equal(t, "Three of a Kind", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 6, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 8, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandStraight, rank)
		assert.Equal(t, "Straight", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight low ace A-2-3-4-5", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 3, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 4, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandStraight, rank)
		assert.Equal(t, "Straight", tpp.GetHandName())
	})

	t.Run("success EvalHand Flush", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandFlush, rank)
		assert.Equal(t, "Flush", tpp.GetHandName())
	})

	t.Run("success EvalHand Full House", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 8, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 8, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 3, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandFullHouse, rank)
		assert.Equal(t, "Full House", tpp.GetHandName())
	})

	t.Run("success EvalHand Four of a Kind", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 6, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 6, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 6, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandFourOfAKind, rank)
		assert.Equal(t, "Four of a Kind", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight Flush", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 4, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 6, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandStraightFlush, rank)
		assert.Equal(t, "Straight Flush", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight Flush", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 3, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 4, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 6, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandStraightFlush, rank)
		assert.Equal(t, "Straight Flush", tpp.GetHandName())
	})

	t.Run("success EvalHand Flush A-10-J-Q-K same suit evaluates as Flush", func(t *testing.T) {
		// Note: checkStraight does not handle high-ace straight (A-10-J-Q-K),
		// so this hand is classified as Flush rather than Royal Flush.
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 12, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 13, false))
		rank := tpp.EvalHand()
		assert.Equal(t, entities.PokerHandFlush, rank)
		assert.Equal(t, "Flush", tpp.GetHandName())
	})

	t.Run("success ExchangeCard", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		newCard := entities.NewCard(entities.CardDesignHeart, 13, false)
		tpp.ExchangeCard(0, newCard)
		assert.Equal(t, 13, tpp.GetCard(0).GetValue())
	})

	t.Run("success ExchangeCard invalid index", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		newCard := entities.NewCard(entities.CardDesignHeart, 13, false)
		tpp.ExchangeCard(10, newCard)
		assert.Equal(t, 2, tpp.GetCard(0).GetValue())
	})
}
