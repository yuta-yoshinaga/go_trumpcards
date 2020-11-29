package entities_test

import (
	"testing"
	"go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestPlayer_Method(t *testing.T) {
	tp := entities.NewPlayer()
	t.Run("success AddCard", func(t *testing.T) {
		tp.AddCard(entities.NewCard(entities.CardDesignJoker, entities.CardValueJoker, false))
		assert.Equal(t, 1, tp.GetCardsSize())
	})
	t.Run("success GetCard", func(t *testing.T) {
		card := tp.GetCard(0)
		assert.Equal(t, entities.CardDesignJoker, card.GetDesign())
		assert.Equal(t, entities.CardValueJoker, card.GetValue())
		assert.Equal(t, false, card.GetDraw())
	})
	t.Run("failed GetCard idx -1", func(t *testing.T) {
		card := tp.GetCard(-1)
		assert.Empty(t, card)
	})
	t.Run("failed GetCard idx 1", func(t *testing.T) {
		card := tp.GetCard(1)
		assert.Empty(t, card)
	})
	t.Run("success Reset", func(t *testing.T) {
		tp.Reset()
		assert.Equal(t, 0, tp.GetCardsSize())
	})
}
