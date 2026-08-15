//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// cardHint consolidates the hint-construction step repeated verbatim in 93
// presenter methods — see issue #5368.
func TestCardHint(t *testing.T) {
	t.Run("copies indices and reason", func(t *testing.T) {
		got := cardHint([]int{1, 3}, "lead low")
		assert.Equal(t, []int{1, 3}, got.CardIndices)
		assert.Equal(t, "lead low", got.Reason)
	})

	// A hint with no indices is still a hint: some games advise "pass" with an
	// empty selection and a reason, and dropping it would lose the reason.
	t.Run("keeps a reason with no indices", func(t *testing.T) {
		got := cardHint(nil, "no legal play, pass")
		assert.Empty(t, got.CardIndices)
		assert.Equal(t, "no legal play, pass", got.Reason)
	})

	// Never returns nil: the caller's own nil check decides whether to assign,
	// and a nil return here would silently blank a hint the game did provide.
	t.Run("always returns a value", func(t *testing.T) {
		assert.NotNil(t, cardHint(nil, ""))
	})
}
