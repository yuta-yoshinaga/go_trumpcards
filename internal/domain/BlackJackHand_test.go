package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewBlackJackHand(t *testing.T) {
	h := domain.NewBlackJackHand()
	assert.NotNil(t, h)
	assert.Equal(t, 0, h.GetCardsSize())
	assert.Equal(t, 0, h.GetScore())
	assert.Equal(t, 0, h.GetBet())
	assert.False(t, h.IsStood())
	assert.False(t, h.IsDoubled())
	assert.False(t, h.IsBusted())
	assert.False(t, h.IsBlackJack())
	assert.False(t, h.CanSplit())
	assert.False(t, h.IsFinished())
}

func TestBlackJackHand_Score(t *testing.T) {
	t.Run("score with ace as 11", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		assert.Equal(t, 20, h.GetScore())
	})
	t.Run("score with ace as 1", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		h.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		assert.Equal(t, 15, h.GetScore())
	})
	t.Run("score with face cards as 10", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 12, false))
		assert.Equal(t, 20, h.GetScore())
	})
	t.Run("score with two aces", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		assert.Equal(t, 12, h.GetScore())
	})
}

func TestBlackJackHand_IsBlackJack(t *testing.T) {
	t.Run("natural blackjack ace+10", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		assert.True(t, h.IsBlackJack())
	})
	t.Run("natural blackjack ace+king", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
		assert.True(t, h.IsBlackJack())
	})
	t.Run("not blackjack with 3 cards", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		h.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		assert.False(t, h.IsBlackJack())
	})
	t.Run("not blackjack score not 21", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		assert.False(t, h.IsBlackJack())
	})
}

func TestBlackJackHand_CanSplit(t *testing.T) {
	t.Run("can split same value", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		assert.True(t, h.CanSplit())
	})
	t.Run("can split face cards king+queen", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 12, false))
		assert.True(t, h.CanSplit())
	})
	t.Run("can split face cards jack+10", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		assert.True(t, h.CanSplit())
	})
	t.Run("cannot split different values", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		assert.False(t, h.CanSplit())
	})
	t.Run("cannot split with 3 cards", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		assert.False(t, h.CanSplit())
	})
	t.Run("cannot split with 1 card", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		assert.False(t, h.CanSplit())
	})
}

func TestBlackJackHand_GetCard(t *testing.T) {
	h := domain.NewBlackJackHand()
	h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	t.Run("valid index", func(t *testing.T) {
		card := h.GetCard(0)
		assert.NotNil(t, card)
		assert.Equal(t, 5, card.GetValue())
	})
	t.Run("out of range index", func(t *testing.T) {
		assert.Nil(t, h.GetCard(-1))
		assert.Nil(t, h.GetCard(1))
	})
}

func TestBlackJackHand_GetCards(t *testing.T) {
	h := domain.NewBlackJackHand()
	h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	h.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	cards := h.GetCards()
	assert.Equal(t, 2, len(cards))
}

func TestBlackJackHand_SettersGetters(t *testing.T) {
	h := domain.NewBlackJackHand()
	h.SetBet(100)
	assert.Equal(t, 100, h.GetBet())
	h.SetStood(true)
	assert.True(t, h.IsStood())
	assert.True(t, h.IsFinished())
	h.SetDoubled(true)
	assert.True(t, h.IsDoubled())
	h.SetBusted(true)
	assert.True(t, h.IsBusted())
	assert.True(t, h.IsFinished())
}

func TestBlackJackHand_Reset(t *testing.T) {
	h := domain.NewBlackJackHand()
	h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	h.SetBet(100)
	h.SetStood(true)
	h.SetDoubled(true)
	h.SetBusted(true)
	h.SetFromSplit(true)
	h.Reset()
	assert.Equal(t, 0, h.GetCardsSize())
	assert.Equal(t, 0, h.GetBet())
	assert.False(t, h.IsStood())
	assert.False(t, h.IsDoubled())
	assert.False(t, h.IsBusted())
	assert.False(t, h.IsFromSplit())
}

func TestBlackJackHand_FromSplit(t *testing.T) {
	h := domain.NewBlackJackHand()
	assert.False(t, h.IsFromSplit())
	h.SetFromSplit(true)
	assert.True(t, h.IsFromSplit())
	h.SetFromSplit(false)
	assert.False(t, h.IsFromSplit())
}

func TestBlackJackHand_IsFinished(t *testing.T) {
	t.Run("not finished by default", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		assert.False(t, h.IsFinished())
	})
	t.Run("finished when stood", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.SetStood(true)
		assert.True(t, h.IsFinished())
	})
	t.Run("finished when busted", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.SetBusted(true)
		assert.True(t, h.IsFinished())
	})
}

func TestBlackJackHand_Surrender(t *testing.T) {
	t.Run("new hand not surrendered", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		assert.False(t, h.IsSurrendered())
		assert.False(t, h.CanSurrender())
	})
	t.Run("can surrender with 2 cards", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		assert.True(t, h.CanSurrender())
	})
	t.Run("cannot surrender with 3 cards", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		h.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		assert.False(t, h.CanSurrender())
	})
	t.Run("cannot surrender when stood", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		h.SetStood(true)
		assert.False(t, h.CanSurrender())
	})
	t.Run("cannot surrender when busted", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		h.SetBusted(true)
		assert.False(t, h.CanSurrender())
	})
	t.Run("cannot surrender when already surrendered", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		h.SetSurrendered(true)
		assert.False(t, h.CanSurrender())
		assert.True(t, h.IsSurrendered())
		assert.True(t, h.IsFinished())
	})
	t.Run("reset clears surrendered", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		h.SetSurrendered(true)
		h.Reset()
		assert.False(t, h.IsSurrendered())
		assert.False(t, h.IsFinished())
	})
}

func TestBlackJackHand_IsFinished_Surrendered(t *testing.T) {
	h := domain.NewBlackJackHand()
	assert.False(t, h.IsFinished())
	h.SetSurrendered(true)
	assert.True(t, h.IsFinished())
}

func TestBlackJackHand_IsSoft(t *testing.T) {
	t.Run("soft hand ace counting as 11", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		assert.True(t, h.IsSoft())
	})
	t.Run("hard hand ace forced to 1", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		h.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		assert.False(t, h.IsSoft())
	})
	t.Run("no ace is hard", func(t *testing.T) {
		h := domain.NewBlackJackHand()
		h.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		h.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		assert.False(t, h.IsSoft())
	})
}
