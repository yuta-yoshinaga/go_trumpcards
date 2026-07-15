package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSevenBridge_SuggestMeldAndDiscard(t *testing.T) {
	g := NewDefaultSevenBridge()
	g.Reset()

	t.Run("out-of-range indices are guarded", func(t *testing.T) {
		assert.Nil(t, g.SuggestMeld(-1))
		assert.Nil(t, g.SuggestMeld(999))
		assert.Equal(t, -1, g.SuggestDiscard(-1))
		assert.Equal(t, -1, g.SuggestDiscard(999))
	})

	t.Run("SuggestDiscard returns an in-range index for a non-empty hand", func(t *testing.T) {
		d := g.SuggestDiscard(0)
		assert.GreaterOrEqual(t, d, 0)
		assert.Less(t, d, g.GetPlayer(0).GetCardsSize())
	})

	t.Run("SuggestMeld finds a guaranteed set of three", func(t *testing.T) {
		human := g.GetPlayer(0)
		human.AddCard(NewCard(CardDesignSpade, 3, false))
		human.AddCard(NewCard(CardDesignClover, 3, false))
		human.AddCard(NewCard(CardDesignHeart, 3, false))
		meld := g.SuggestMeld(0)
		assert.GreaterOrEqual(t, len(meld), SevenBridgeMeldMinSize)
	})
}
