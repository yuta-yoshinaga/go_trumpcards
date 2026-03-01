package presenter

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestCardToOutput(t *testing.T) {
	t.Run("SPADE", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignSpade, 1, false))
		assert.Equal(t, "SPADE", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("CLOVER", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignClover, 5, false))
		assert.Equal(t, "CLOVER", card.Design)
		assert.Equal(t, 5, card.Value)
	})
	t.Run("HEART", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignHeart, 10, false))
		assert.Equal(t, "HEART", card.Design)
		assert.Equal(t, 10, card.Value)
	})
	t.Run("DIAMOND", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignDiamond, 13, false))
		assert.Equal(t, "DIAMOND", card.Design)
		assert.Equal(t, 13, card.Value)
	})
	t.Run("JOKER", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		assert.Equal(t, "JOKER", card.Design)
		assert.Equal(t, 0, card.Value)
	})
	t.Run("nil input", func(t *testing.T) {
		assert.Nil(t, cardToOutput(nil))
	})
}

func TestCardsToOutput(t *testing.T) {
	t.Run("normal slice", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 13, false),
		}
		result := cardsToOutput(cards)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "SPADE", result[0].Design)
		assert.Equal(t, 1, result[0].Value)
		assert.Equal(t, "HEART", result[1].Design)
		assert.Equal(t, 13, result[1].Value)
	})
	t.Run("nil slice", func(t *testing.T) {
		assert.Nil(t, cardsToOutput(nil))
	})
	t.Run("empty slice", func(t *testing.T) {
		result := cardsToOutput([]*domain.Card{})
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})
}
