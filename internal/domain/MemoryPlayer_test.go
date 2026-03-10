package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := NewMemoryPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetPairCount())
		assert.Nil(t, p.GetPairs())
		assert.Equal(t, 0, p.GetMemoryCount())
	})
	t.Run("CPU player", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		assert.False(t, p.GetIsHuman())
	})
}

func TestMemoryPlayerPairs(t *testing.T) {
	p := NewMemoryPlayer(false)
	c1 := NewCard(CardDesignSpade, 5, false)
	c2 := NewCard(CardDesignHeart, 5, false)
	p.AddPair(c1, c2)
	assert.Equal(t, 1, p.GetPairCount())
	assert.Len(t, p.GetPairs(), 1)
	assert.Equal(t, c1, p.GetPairs()[0][0])
	assert.Equal(t, c2, p.GetPairs()[0][1])

	// Add another pair
	c3 := NewCard(CardDesignClover, 10, false)
	c4 := NewCard(CardDesignDiamond, 10, false)
	p.AddPair(c3, c4)
	assert.Equal(t, 2, p.GetPairCount())
}

func TestMemoryPlayerSetPairCount(t *testing.T) {
	p := NewMemoryPlayer(false)
	p.SetPairCount(5)
	assert.Equal(t, 5, p.GetPairCount())
}

func TestMemoryPlayerResetGame(t *testing.T) {
	p := NewMemoryPlayer(false)
	c1 := NewCard(CardDesignSpade, 5, false)
	c2 := NewCard(CardDesignHeart, 5, false)
	p.AddPair(c1, c2)
	p.RecordRevealedCard(0, 5, 1.0, 1)
	p.ResetGame()
	assert.Equal(t, 0, p.GetPairCount())
	assert.Nil(t, p.GetPairs())
	assert.Equal(t, 0, p.GetMemoryCount())
}

func TestMemoryPlayerRecordRevealedCard(t *testing.T) {
	t.Run("records with 100% chance", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(5, 3, 1.0, 1)
		assert.Equal(t, 1, p.GetMemoryCount())
	})
	t.Run("does not duplicate same position", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(5, 3, 1.0, 1)
		p.RecordRevealedCard(5, 3, 1.0, 2)
		assert.Equal(t, 1, p.GetMemoryCount())
	})
	t.Run("0% chance does not record", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(5, 3, 0.0, 1)
		assert.Equal(t, 0, p.GetMemoryCount())
	})
	t.Run("random retention branch", func(t *testing.T) {
		// ensure that with 0.5 chance, sometimes recorded and sometimes not
		recorded := false
		notRecorded := false
		for i := 0; i < 1000; i++ {
			p := NewMemoryPlayer(false)
			p.RecordRevealedCard(i, 3, 0.5, 1)
			if p.GetMemoryCount() > 0 {
				recorded = true
			} else {
				notRecorded = true
			}
			if recorded && notRecorded {
				return
			}
		}
		t.Fatal("expected both recorded and not-recorded outcomes")
	})
}

func TestMemoryPlayerResetMemory(t *testing.T) {
	p := NewMemoryPlayer(false)
	p.RecordRevealedCard(0, 5, 1.0, 1)
	p.ResetMemory()
	assert.Equal(t, 0, p.GetMemoryCount())
}

func TestMemoryPlayerDecayMemories(t *testing.T) {
	t.Run("high decay forgets old memories", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(0, 5, 1.0, 1)
		// decayRate=1.0, age=10 => forgetProb=10.0 >= 1.0 => definitely forget
		p.DecayMemories(11, 1.0)
		assert.Equal(t, 0, p.GetMemoryCount())
	})
	t.Run("zero decay keeps memories", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(0, 5, 1.0, 1)
		p.DecayMemories(100, 0.0)
		assert.Equal(t, 1, p.GetMemoryCount())
	})
	t.Run("random decay branch", func(t *testing.T) {
		kept := false
		forgotten := false
		for i := 0; i < 1000; i++ {
			p := NewMemoryPlayer(false)
			p.RecordRevealedCard(0, 5, 1.0, 1)
			p.DecayMemories(2, 0.5) // forgetProb=0.5
			if p.GetMemoryCount() > 0 {
				kept = true
			} else {
				forgotten = true
			}
			if kept && forgotten {
				return
			}
		}
		t.Fatal("expected both kept and forgotten outcomes")
	})
}

func TestMemoryPlayerFindKnownMatch(t *testing.T) {
	t.Run("finds pair", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(3, 7, 1.0, 1)
		p.RecordRevealedCard(10, 7, 1.0, 2)
		pos1, pos2, found := p.FindKnownMatch(7)
		assert.True(t, found)
		assert.Equal(t, 3, pos1)
		assert.Equal(t, 10, pos2)
	})
	t.Run("no pair with one known", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(3, 7, 1.0, 1)
		_, _, found := p.FindKnownMatch(7)
		assert.False(t, found)
	})
	t.Run("no match for unknown rank", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(3, 7, 1.0, 1)
		_, _, found := p.FindKnownMatch(5)
		assert.False(t, found)
	})
}

func TestMemoryPlayerFindAnyKnownPair(t *testing.T) {
	t.Run("finds pair among mixed memories", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(0, 3, 1.0, 1)
		p.RecordRevealedCard(5, 7, 1.0, 2)
		p.RecordRevealedCard(10, 7, 1.0, 3)
		pos1, pos2, found := p.FindAnyKnownPair()
		assert.True(t, found)
		assert.Equal(t, 5, pos1)
		assert.Equal(t, 10, pos2)
	})
	t.Run("no pair", func(t *testing.T) {
		p := NewMemoryPlayer(false)
		p.RecordRevealedCard(0, 3, 1.0, 1)
		p.RecordRevealedCard(5, 7, 1.0, 2)
		_, _, found := p.FindAnyKnownPair()
		assert.False(t, found)
	})
}

func TestMemoryPlayerRemoveMemoryAt(t *testing.T) {
	p := NewMemoryPlayer(false)
	p.RecordRevealedCard(0, 3, 1.0, 1)
	p.RecordRevealedCard(5, 7, 1.0, 2)
	p.RemoveMemoryAt(0)
	assert.Equal(t, 1, p.GetMemoryCount())
	// removing non-existent position does nothing
	p.RemoveMemoryAt(99)
	assert.Equal(t, 1, p.GetMemoryCount())
}
