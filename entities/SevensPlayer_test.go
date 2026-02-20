package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestSevensPlayer_Method(t *testing.T) {
	t.Run("success NewSevensPlayer human", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		assert.NotNil(t, p)
		assert.True(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
		assert.Equal(t, -1, p.GetRank())
		assert.Equal(t, 0, p.GetPassesUsed())
		assert.Equal(t, entities.SevensMaxPasses, p.GetMaxPasses())
		assert.True(t, p.CanPass())
	})

	t.Run("success NewSevensPlayer cpu", func(t *testing.T) {
		p := entities.NewSevensPlayer(false)
		assert.False(t, p.GetIsHuman())
	})

	t.Run("success SetIsFinished and GetIsFinished", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		p.SetIsFinished(true)
		assert.True(t, p.GetIsFinished())
		p.SetIsFinished(false)
		assert.False(t, p.GetIsFinished())
	})

	t.Run("success SetRank and GetRank", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		p.SetRank(2)
		assert.Equal(t, 2, p.GetRank())
	})

	t.Run("success IncrPassesUsed and CanPass", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		assert.True(t, p.CanPass())
		for i := 0; i < entities.SevensMaxPasses; i++ {
			p.IncrPassesUsed()
		}
		assert.False(t, p.CanPass())
		assert.Equal(t, entities.SevensMaxPasses, p.GetPassesUsed())
	})

	t.Run("success ResetPasses", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		p.IncrPassesUsed()
		p.IncrPassesUsed()
		p.ResetPasses()
		assert.Equal(t, 0, p.GetPassesUsed())
		assert.True(t, p.CanPass())
	})

	t.Run("success RemoveCard valid index", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		card := p.RemoveCard(0)
		assert.NotNil(t, card)
		assert.Equal(t, entities.CardDesignSpade, card.GetDesign())
		assert.Equal(t, 3, card.GetValue())
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("success RemoveCard invalid index returns nil", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		card := p.RemoveCard(5)
		assert.Nil(t, card)
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("success RemoveSevens removes all 7s", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 8, false))
		removed := p.RemoveSevens()
		assert.Len(t, removed, 2)
		assert.Equal(t, 2, p.GetCardsSize())
		for _, c := range removed {
			assert.Equal(t, 7, c.GetValue())
		}
	})

	t.Run("success RemoveSevens no sevens", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		removed := p.RemoveSevens()
		assert.Len(t, removed, 0)
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("success SortCards sorts by suit then value", func(t *testing.T) {
		p := entities.NewSevensPlayer(true)
		p.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		p.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		p.SortCards()
		assert.Equal(t, entities.CardDesignSpade, p.GetCard(0).GetDesign())
		assert.Equal(t, 3, p.GetCard(0).GetValue())
		assert.Equal(t, entities.CardDesignSpade, p.GetCard(1).GetDesign())
		assert.Equal(t, 10, p.GetCard(1).GetValue())
		assert.Equal(t, entities.CardDesignClover, p.GetCard(2).GetDesign())
		assert.Equal(t, 2, p.GetCard(2).GetValue())
		assert.Equal(t, entities.CardDesignHeart, p.GetCard(3).GetDesign())
		assert.Equal(t, 5, p.GetCard(3).GetValue())
	})
}
