package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandAndFoot_SuggestMelds(t *testing.T) {
	g := NewDefaultHandAndFoot()
	g.Reset()

	t.Run("out-of-range index returns nil", func(t *testing.T) {
		assert.Nil(t, g.SuggestMelds(-1))
		assert.Nil(t, g.SuggestMelds(999))
	})

	t.Run("finds a meld from a guaranteed triple", func(t *testing.T) {
		p := g.GetPlayer(0)
		p.AddCard(NewCard(CardDesignSpade, 7, false))
		p.AddCard(NewCard(CardDesignHeart, 7, false))
		p.AddCard(NewCard(CardDesignClover, 7, false))
		melds := g.SuggestMelds(0)
		assert.NotEmpty(t, melds)
	})
}
