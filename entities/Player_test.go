package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

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
	t.Run("success ShuffleCards does not change size", func(t *testing.T) {
		p := entities.NewPlayer()
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		p.ShuffleCards()
		assert.Equal(t, 3, p.GetCardsSize())
	})
	t.Run("success ShuffleCards randomizes order", func(t *testing.T) {
		p := entities.NewPlayer()
		for i := 2; i <= 10; i++ {
			p.AddCard(entities.NewCard(entities.CardDesignSpade, i, false))
		}
		original := make([]int, p.GetCardsSize())
		for i := range original {
			original[i] = p.GetCard(i).GetValue()
		}
		// 多数回シャッフルして順番が変わることを確認
		changed := false
		for attempt := 0; attempt < 100; attempt++ {
			p2 := entities.NewPlayer()
			for _, v := range original {
				p2.AddCard(entities.NewCard(entities.CardDesignSpade, v, false))
			}
			p2.ShuffleCards()
			for i := range original {
				if p2.GetCard(i).GetValue() != original[i] {
					changed = true
					break
				}
			}
			if changed {
				break
			}
		}
		assert.True(t, changed, "ShuffleCards should change card order")
	})
	t.Run("success ShuffleCards on empty hand", func(t *testing.T) {
		p := entities.NewPlayer()
		p.ShuffleCards()
		assert.Equal(t, 0, p.GetCardsSize())
	})
	t.Run("success ShuffleCards on single card hand", func(t *testing.T) {
		p := entities.NewPlayer()
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		p.ShuffleCards()
		assert.Equal(t, 1, p.GetCardsSize())
		assert.Equal(t, 2, p.GetCard(0).GetValue())
	})
}
