package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestDaifugoPlayer_Method(t *testing.T) {
	t.Run("success NewDaifugoPlayer human", func(t *testing.T) {
		p := entities.NewDaifugoPlayer(true)
		assert.NotNil(t, p)
		assert.True(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
		assert.Equal(t, -1, p.GetRank())
		assert.Equal(t, 0, p.GetCardsSize())
	})

	t.Run("success NewDaifugoPlayer cpu", func(t *testing.T) {
		p := entities.NewDaifugoPlayer(false)
		assert.False(t, p.GetIsHuman())
	})

	t.Run("success SetIsFinished", func(t *testing.T) {
		p := entities.NewDaifugoPlayer(true)
		p.SetIsFinished(true)
		assert.True(t, p.GetIsFinished())
		p.SetIsFinished(false)
		assert.False(t, p.GetIsFinished())
	})

	t.Run("success SetRank and GetRank", func(t *testing.T) {
		p := entities.NewDaifugoPlayer(true)
		p.SetRank(1)
		assert.Equal(t, 1, p.GetRank())
		p.SetRank(4)
		assert.Equal(t, 4, p.GetRank())
	})

	t.Run("success SortCards sorts by Daifugo strength", func(t *testing.T) {
		p := entities.NewDaifugoPlayer(true)
		// Add cards in non-sorted order: A(1), K(13), 3, 2
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))  // Ace → strength 14
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 13, false)) // K → strength 13
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))  // 3 → strength 3
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))  // 2 → strength 15
		p.SortCards()
		// Expected order: 3, K, A, 2
		assert.Equal(t, 3, p.GetCard(0).GetValue())
		assert.Equal(t, 13, p.GetCard(1).GetValue())
		assert.Equal(t, 1, p.GetCard(2).GetValue())
		assert.Equal(t, 2, p.GetCard(3).GetValue())
	})

	t.Run("success RemoveCards single", func(t *testing.T) {
		p := entities.NewDaifugoPlayer(true)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		removed := p.RemoveCards([]int{1})
		assert.Len(t, removed, 1)
		assert.Equal(t, 5, removed[0].GetValue())
		assert.Equal(t, 2, p.GetCardsSize())
		assert.Equal(t, 3, p.GetCard(0).GetValue())
		assert.Equal(t, 7, p.GetCard(1).GetValue())
	})

	t.Run("success RemoveCards multiple", func(t *testing.T) {
		p := entities.NewDaifugoPlayer(true)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		removed := p.RemoveCards([]int{0, 2})
		assert.Len(t, removed, 2)
		assert.Equal(t, 3, removed[0].GetValue())
		assert.Equal(t, 7, removed[1].GetValue())
		assert.Equal(t, 2, p.GetCardsSize())
		assert.Equal(t, 5, p.GetCard(0).GetValue())
		assert.Equal(t, 9, p.GetCard(1).GetValue())
	})

	t.Run("success RemoveCards indices unordered", func(t *testing.T) {
		p := entities.NewDaifugoPlayer(true)
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 3, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		p.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		// Remove indices 2,0 (unordered)
		removed := p.RemoveCards([]int{2, 0})
		assert.Len(t, removed, 2)
		assert.Equal(t, 3, removed[0].GetValue())
		assert.Equal(t, 7, removed[1].GetValue())
		assert.Equal(t, 1, p.GetCardsSize())
		assert.Equal(t, 5, p.GetCard(0).GetValue())
	})
}

func TestDaifugoCardStrength(t *testing.T) {
	t.Run("3 is weakest playable card", func(t *testing.T) {
		assert.Equal(t, 3, entities.DaifugoCardStrength(3))
	})
	t.Run("Ace is strength 14", func(t *testing.T) {
		assert.Equal(t, 14, entities.DaifugoCardStrength(1))
	})
	t.Run("2 is strength 15 (strongest)", func(t *testing.T) {
		assert.Equal(t, 15, entities.DaifugoCardStrength(2))
	})
	t.Run("Jack is 11", func(t *testing.T) {
		assert.Equal(t, 11, entities.DaifugoCardStrength(11))
	})
	t.Run("King is 13", func(t *testing.T) {
		assert.Equal(t, 13, entities.DaifugoCardStrength(13))
	})
}
