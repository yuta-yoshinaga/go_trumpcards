package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestDaifugoPlayer_Method(t *testing.T) {
	t.Run("success NewDaifugoPlayer human", func(t *testing.T) {
		p := domain.NewDaifugoPlayer(true)
		assert.NotNil(t, p)
		assert.True(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
		assert.Equal(t, -1, p.GetRank())
		assert.Equal(t, 0, p.GetCardsSize())
	})

	t.Run("success NewDaifugoPlayer cpu", func(t *testing.T) {
		p := domain.NewDaifugoPlayer(false)
		assert.False(t, p.GetIsHuman())
	})

	t.Run("success SortCards sorts by Daifugo strength", func(t *testing.T) {
		p := domain.NewDaifugoPlayer(true)
		// Add cards in non-sorted order: A(1), K(13), 3, 2
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // Ace → strength 14
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // K → strength 13
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))  // 3 → strength 3
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))  // 2 → strength 15
		p.SortCards()
		// Expected order: 3, K, A, 2
		assert.Equal(t, 3, p.GetCard(0).GetValue())
		assert.Equal(t, 13, p.GetCard(1).GetValue())
		assert.Equal(t, 1, p.GetCard(2).GetValue())
		assert.Equal(t, 2, p.GetCard(3).GetValue())
	})

}

func TestDaifugoPlayer_SortCardsByStrength(t *testing.T) {
	t.Run("sorts by custom strength function (revolution order)", func(t *testing.T) {
		p := domain.NewDaifugoPlayer(true)
		// Add cards: 3, K(13), A(1), 2
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))  // revolution strength 15 (strongest)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // revolution strength 5
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // revolution strength 4
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))  // revolution strength 3 (weakest)
		p.SortCardsByStrength(func(c *domain.Card) int {
			return domain.DaifugoCardStrengthRevolution(c.GetValue())
		})
		// Expected order by revolution strength (weakest first): 2, A, K, 3
		assert.Equal(t, 2, p.GetCard(0).GetValue())
		assert.Equal(t, 1, p.GetCard(1).GetValue())
		assert.Equal(t, 13, p.GetCard(2).GetValue())
		assert.Equal(t, 3, p.GetCard(3).GetValue())
	})

	t.Run("success SetPrevRank and GetPrevRank", func(t *testing.T) {
		p := domain.NewDaifugoPlayer(true)
		assert.Equal(t, -1, p.GetPrevRank())
		p.SetPrevRank(2)
		assert.Equal(t, 2, p.GetPrevRank())
	})

	t.Run("success SortCards with joker puts joker last", func(t *testing.T) {
		p := domain.NewDaifugoPlayer(true)
		p.AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.SortCards()
		// Expected: 3(str=3), 2(str=15), Joker(str=16)
		assert.Equal(t, 3, p.GetCard(0).GetValue())
		assert.Equal(t, 2, p.GetCard(1).GetValue())
		assert.Equal(t, domain.CardDesignJoker, p.GetCard(2).GetDesign())
	})
}

func TestDaifugoCardStrength(t *testing.T) {
	t.Run("3 is weakest playable card", func(t *testing.T) {
		assert.Equal(t, 3, domain.DaifugoCardStrength(3))
	})
	t.Run("Ace is strength 14", func(t *testing.T) {
		assert.Equal(t, 14, domain.DaifugoCardStrength(1))
	})
	t.Run("2 is strength 15 (strongest)", func(t *testing.T) {
		assert.Equal(t, 15, domain.DaifugoCardStrength(2))
	})
	t.Run("Jack is 11", func(t *testing.T) {
		assert.Equal(t, 11, domain.DaifugoCardStrength(11))
	})
	t.Run("King is 13", func(t *testing.T) {
		assert.Equal(t, 13, domain.DaifugoCardStrength(13))
	})
}

func TestDaifugoCardStrengthRevolution(t *testing.T) {
	t.Run("3 is strongest in revolution", func(t *testing.T) {
		assert.Equal(t, 15, domain.DaifugoCardStrengthRevolution(3))
	})
	t.Run("2 is weakest in revolution", func(t *testing.T) {
		assert.Equal(t, 3, domain.DaifugoCardStrengthRevolution(2))
	})
	t.Run("Ace is revolution strength 4", func(t *testing.T) {
		assert.Equal(t, 4, domain.DaifugoCardStrengthRevolution(1))
	})
	t.Run("King is revolution strength 5", func(t *testing.T) {
		assert.Equal(t, 5, domain.DaifugoCardStrengthRevolution(13))
	})
	t.Run("Jack is revolution strength 7", func(t *testing.T) {
		assert.Equal(t, 7, domain.DaifugoCardStrengthRevolution(11))
	})
	t.Run("4 is revolution strength 14", func(t *testing.T) {
		assert.Equal(t, 14, domain.DaifugoCardStrengthRevolution(4))
	})
}
