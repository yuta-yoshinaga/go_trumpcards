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

func TestDoubtPlayer_SetIsFinished(t *testing.T) {
	p := domain.NewDoubtPlayer(true)
	assert.False(t, p.GetIsFinished())
	p.SetIsFinished(true)
	assert.True(t, p.GetIsFinished())
	p.SetIsFinished(false)
	assert.False(t, p.GetIsFinished())
}

func TestDoubtPlayer_RemoveCards(t *testing.T) {
	makePlayer := func() *domain.DoubtPlayer {
		p := domain.NewDoubtPlayer(false)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // 0
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false)) // 1
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // 2
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false)) // 3
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // 4
		return p
	}

	t.Run("empty indices returns empty slice", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{})
		assert.Empty(t, removed)
		assert.Equal(t, 5, p.GetCardsSize())
	})

	t.Run("single index removal", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{2})
		assert.Len(t, removed, 1)
		assert.Equal(t, 3, removed[0].GetValue())
		assert.Equal(t, 4, p.GetCardsSize())
	})

	t.Run("multiple indices removal in order", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{0, 2, 4})
		assert.Len(t, removed, 3)
		assert.Equal(t, 1, removed[0].GetValue())
		assert.Equal(t, 3, removed[1].GetValue())
		assert.Equal(t, 5, removed[2].GetValue())
		assert.Equal(t, 2, p.GetCardsSize())
	})

	t.Run("duplicate indices are deduplicated", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{1, 1, 3})
		assert.Len(t, removed, 2)
		assert.Equal(t, 2, removed[0].GetValue())
		assert.Equal(t, 4, removed[1].GetValue())
		assert.Equal(t, 3, p.GetCardsSize())
	})

	t.Run("out-of-range index is ignored", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{0, 99})
		assert.Len(t, removed, 1)
		assert.Equal(t, 1, removed[0].GetValue())
		assert.Equal(t, 4, p.GetCardsSize())
	})

	t.Run("negative index is ignored", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{-1, 0})
		assert.Len(t, removed, 1)
		assert.Equal(t, 1, removed[0].GetValue())
		assert.Equal(t, 4, p.GetCardsSize())
	})

	t.Run("remove all cards", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{0, 1, 2, 3, 4})
		assert.Len(t, removed, 5)
		assert.Equal(t, 0, p.GetCardsSize())
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

func TestDoubtPlayer_RecordRevealedCard(t *testing.T) {
	t.Run("full retention - always records", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		p.RecordRevealedCard(7, 1.0, 0)
		p.RecordRevealedCard(7, 1.0, 0)
		assert.Equal(t, 2, p.CountKnownCards(7))
	})

	t.Run("zero retention - never records", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		for i := 0; i < 100; i++ {
			p.RecordRevealedCard(3, 0.0, 0)
		}
		assert.Equal(t, 0, p.CountKnownCards(3))
	})

	t.Run("partial retention - sometimes records", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		for attempt := 0; attempt < 1000; attempt++ {
			p.RecordRevealedCard(9, 0.5, 0)
			if p.CountKnownCards(9) > 0 {
				return // retention branch hit
			}
		}
		t.Fatal("partial retention never recorded after 1000 attempts")
	})
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

func TestDoubtPlayer_DecayMemories(t *testing.T) {
	t.Run("decayRate=0 never forgets", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		p.RecordRevealedCard(5, 1.0, 0)
		p.RecordRevealedCard(7, 1.0, 0)
		p.DecayMemories(100, 0.0) // rate=0 → 忘れない
		assert.Equal(t, 1, p.CountKnownCards(5))
		assert.Equal(t, 1, p.CountKnownCards(7))
	})

	t.Run("forgetProb >= 1.0 always forgets", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		p.RecordRevealedCard(5, 1.0, 0) // age = 10-0 = 10, forgetProb = 1.0*10 = 10.0 >= 1.0
		p.DecayMemories(10, 1.0)
		assert.Equal(t, 0, p.CountKnownCards(5))
	})

	t.Run("new memories survive, old memories forgotten", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		p.RecordRevealedCard(5, 1.0, 0)  // old: age=20
		p.RecordRevealedCard(7, 1.0, 20) // new: age=0, forgetProb=0
		p.DecayMemories(20, 1.0)         // rate=1.0, old: forgetProb=20 >= 1 → forget, new: forgetProb=0 → keep
		assert.Equal(t, 0, p.CountKnownCards(5))
		assert.Equal(t, 1, p.CountKnownCards(7))
	})

	t.Run("probabilistic decay - sometimes forgets", func(t *testing.T) {
		forgotten := false
		for attempt := 0; attempt < 1000; attempt++ {
			p := domain.NewDoubtPlayer(false)
			p.RecordRevealedCard(5, 1.0, 0) // age=5, forgetProb=0.1*5=0.5
			p.DecayMemories(5, 0.1)
			if p.CountKnownCards(5) == 0 {
				forgotten = true
				break
			}
		}
		assert.True(t, forgotten, "probabilistic decay should sometimes forget")
	})

	t.Run("probabilistic decay - sometimes remembers", func(t *testing.T) {
		remembered := false
		for attempt := 0; attempt < 1000; attempt++ {
			p := domain.NewDoubtPlayer(false)
			p.RecordRevealedCard(5, 1.0, 0) // age=5, forgetProb=0.1*5=0.5
			p.DecayMemories(5, 0.1)
			if p.CountKnownCards(5) == 1 {
				remembered = true
				break
			}
		}
		assert.True(t, remembered, "probabilistic decay should sometimes remember")
	})

	t.Run("empty memories - no panic", func(t *testing.T) {
		p := domain.NewDoubtPlayer(false)
		p.DecayMemories(10, 0.5) // 空スライス → no-op
		assert.Equal(t, 0, p.CountKnownCards(1))
	})
}
