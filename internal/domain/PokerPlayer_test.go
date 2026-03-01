package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPokerPlayer_Method(t *testing.T) {
	tpp := domain.NewPokerPlayer()

	t.Run("success GetHandRank initial", func(t *testing.T) {
		assert.Equal(t, domain.PokerHandHighCard, tpp.GetHandRank())
	})

	t.Run("success GetHandName initial", func(t *testing.T) {
		assert.Equal(t, "High Card", tpp.GetHandName())
	})

	t.Run("success EvalHand with less than 5 cards", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandHighCard, rank)
	})

	t.Run("success EvalHand High Card", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandHighCard, rank)
		assert.Equal(t, "High Card", tpp.GetHandName())
	})

	t.Run("success EvalHand One Pair", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandOnePair, rank)
		assert.Equal(t, "One Pair", tpp.GetHandName())
	})

	t.Run("success EvalHand Two Pair", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandTwoPair, rank)
		assert.Equal(t, "Two Pair", tpp.GetHandName())
	})

	t.Run("success EvalHand Three of a Kind", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandThreeOfAKind, rank)
		assert.Equal(t, "Three of a Kind", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandStraight, rank)
		assert.Equal(t, "Straight", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight low ace A-2-3-4-5", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandStraight, rank)
		assert.Equal(t, "Straight", tpp.GetHandName())
	})

	t.Run("success EvalHand Flush", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandFlush, rank)
		assert.Equal(t, "Flush", tpp.GetHandName())
	})

	t.Run("success EvalHand Full House", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandFullHouse, rank)
		assert.Equal(t, "Full House", tpp.GetHandName())
	})

	t.Run("success EvalHand Four of a Kind", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandFourOfAKind, rank)
		assert.Equal(t, "Four of a Kind", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight Flush", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandStraightFlush, rank)
		assert.Equal(t, "Straight Flush", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight Flush", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandStraightFlush, rank)
		assert.Equal(t, "Straight Flush", tpp.GetHandName())
	})

	t.Run("success EvalHand Royal Flush A-10-J-Q-K same suit", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandRoyalFlush, rank)
		assert.Equal(t, "Royal Flush", tpp.GetHandName())
	})

	t.Run("success EvalHand Straight high ace A-10-J-Q-K mixed suit", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 12, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		rank := tpp.EvalHand()
		assert.Equal(t, domain.PokerHandStraight, rank)
		assert.Equal(t, "Straight", tpp.GetHandName())
	})

	t.Run("success ExchangeCard", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		newCard := domain.NewCard(domain.CardDesignHeart, 13, false)
		tpp.ExchangeCard(0, newCard)
		assert.Equal(t, 13, tpp.GetCard(0).GetValue())
	})

	t.Run("success ExchangeCard invalid index", func(t *testing.T) {
		tpp.Reset()
		tpp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		newCard := domain.NewCard(domain.CardDesignHeart, 13, false)
		tpp.ExchangeCard(10, newCard)
		assert.Equal(t, 2, tpp.GetCard(0).GetValue())
	})

	t.Run("success GetChips initial", func(t *testing.T) {
		assert.Equal(t, 0, tpp.GetChips())
	})

	t.Run("success SetChips", func(t *testing.T) {
		tpp.SetChips(1000)
		assert.Equal(t, 1000, tpp.GetChips())
	})

	t.Run("success AddChips", func(t *testing.T) {
		tpp.SetChips(100)
		tpp.AddChips(50)
		assert.Equal(t, 150, tpp.GetChips())
	})

	t.Run("success SubtractChips sufficient", func(t *testing.T) {
		tpp.SetChips(100)
		ok := tpp.SubtractChips(30)
		assert.True(t, ok)
		assert.Equal(t, 70, tpp.GetChips())
	})

	t.Run("success SubtractChips insufficient", func(t *testing.T) {
		tpp.SetChips(10)
		ok := tpp.SubtractChips(50)
		assert.False(t, ok)
		assert.Equal(t, 10, tpp.GetChips())
	})
}
