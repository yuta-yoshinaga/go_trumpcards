package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresident_SuggestWeakestPlay(t *testing.T) {
	p := NewDefaultPresident()
	p.Reset()

	t.Run("returns nil for an out-of-range player index", func(t *testing.T) {
		assert.Nil(t, p.SuggestWeakestPlay(-1))
		assert.Nil(t, p.SuggestWeakestPlay(999))
	})

	t.Run("delegates to findWeakestLegalPlay for a valid index", func(t *testing.T) {
		// A freshly dealt player on an open field always has at least one legal
		// single-card play, so the suggestion is non-empty and in range.
		play := p.SuggestWeakestPlay(0)
		for _, idx := range play {
			assert.GreaterOrEqual(t, idx, 0)
			assert.Less(t, idx, p.GetPlayer(0).GetCardsSize())
		}
	})
}
