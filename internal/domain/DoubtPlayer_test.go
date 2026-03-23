package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewDoubtPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := domain.NewDoubtPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
		assert.Equal(t, 0, p.GetCardsSize())
	})

	t.Run("CPU player", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		assert.False(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
	})
}

func TestDoubtPlayer_ResetMemory(t *testing.T) {
	p := domain.NewDoubtPlayer(false)
	// Record some cards first
	p.RecordRevealedCard(5, 1.0, 0)
	p.RecordRevealedCard(5, 1.0, 0)
	assert.Equal(t, 2, p.CountKnownCards(5))

	p.ResetMemory()
	assert.Equal(t, 0, p.CountKnownCards(5))
}

func TestDoubtPlayer_CountKnownCards(t *testing.T) {
	t.Run("includes hand cards", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		// 2 in hand, 0 in memory
		assert.Equal(t, 2, p.CountKnownCards(5))
	})

	t.Run("combines memory and hand", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		p.RecordRevealedCard(3, 1.0, 0)
		p.RecordRevealedCard(3, 1.0, 0)
		// 1 in hand + 2 in memory = 3
		assert.Equal(t, 3, p.CountKnownCards(3))
	})

	t.Run("unknown value returns zero", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		assert.Equal(t, 0, p.CountKnownCards(13))
	})
}
