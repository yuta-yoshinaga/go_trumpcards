package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestOldMaidPlayer_Method(t *testing.T) {
	t.Run("success NewOldMaidPlayer human", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
	})

	t.Run("success NewOldMaidPlayer cpu", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		assert.False(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
	})

	t.Run("success SetIsFinished true", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.SetIsFinished(true)
		assert.True(t, p.GetIsFinished())
	})

	t.Run("success SetIsFinished false", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.SetIsFinished(true)
		p.SetIsFinished(false)
		assert.False(t, p.GetIsFinished())
	})

	t.Run("success DiscardPairs no pairs", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		discardedCards, discardedCount := p.DiscardPairs()
		assert.Equal(t, 0, discardedCount)
		assert.Equal(t, 0, len(discardedCards))
		assert.Equal(t, 3, p.GetCardsSize())
	})

	t.Run("success DiscardPairs one pair", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		discardedCards, discardedCount := p.DiscardPairs()
		assert.Equal(t, 1, discardedCount)
		assert.Equal(t, 2, len(discardedCards))
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("success DiscardPairs two pairs", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 9, false))
		p.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		discardedCards, discardedCount := p.DiscardPairs()
		assert.Equal(t, 2, discardedCount)
		assert.Equal(t, 4, len(discardedCards))
		assert.Equal(t, 0, p.GetCardsSize())
	})

	t.Run("success DiscardPairs joker not paired", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		discardedCards, discardedCount := p.DiscardPairs()
		// joker should not be paired, but the two 5s should be discarded
		assert.Equal(t, 1, discardedCount)
		assert.Equal(t, 2, len(discardedCards))
		assert.Equal(t, 1, p.GetCardsSize())
		assert.Equal(t, entities.CardDesignJoker, p.GetCard(0).GetDesign())
	})

	t.Run("success DiscardPairs empty hand", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		discardedCards, discardedCount := p.DiscardPairs()
		assert.Equal(t, 0, discardedCount)
		assert.Equal(t, 0, len(discardedCards))
		assert.Equal(t, 0, p.GetCardsSize())
	})

	t.Run("success RemoveCard valid index", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		card := p.RemoveCard(1)
		assert.NotNil(t, card)
		assert.Equal(t, 5, card.GetValue())
		assert.Equal(t, 2, p.GetCardsSize())
	})

	t.Run("success RemoveCard first index", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		card := p.RemoveCard(0)
		assert.NotNil(t, card)
		assert.Equal(t, 2, card.GetValue())
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("success RemoveCard last index", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		card := p.RemoveCard(1)
		assert.NotNil(t, card)
		assert.Equal(t, 5, card.GetValue())
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("success RemoveCard invalid negative index returns nil", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		card := p.RemoveCard(-1)
		assert.Nil(t, card)
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("success RemoveCard invalid out-of-bounds index returns nil", func(t *testing.T) {
		p := entities.NewOldMaidPlayer(false)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		card := p.RemoveCard(5)
		assert.Nil(t, card)
		assert.Equal(t, 1, p.GetCardsSize())
	})
}
