//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// Mirrors frontend/src/utils/spadesBid.ts so the badge and the CUI line cannot
// disagree about whether a nil is still alive.
func TestSpadesBidProgressOf(t *testing.T) {
	t.Run("nil holds while the player has taken nothing", func(t *testing.T) {
		p := domain.SpadesBidProgressOf(0, 0)
		assert.Equal(t, domain.SpadesBidNilOk, p.Kind)
	})

	t.Run("nil fails on the very first trick", func(t *testing.T) {
		p := domain.SpadesBidProgressOf(0, 1)
		assert.Equal(t, domain.SpadesBidNilFail, p.Kind, "one trick is already fatal to a nil")
	})

	t.Run("a positive bid counts down what is still owed", func(t *testing.T) {
		p := domain.SpadesBidProgressOf(4, 1)
		assert.Equal(t, domain.SpadesBidRemaining, p.Kind)
		assert.Equal(t, 3, p.Remaining)
	})

	t.Run("meeting the bid exactly is made with no bags", func(t *testing.T) {
		p := domain.SpadesBidProgressOf(4, 4)
		assert.Equal(t, domain.SpadesBidMade, p.Kind)
		assert.Equal(t, 0, p.Bags)
	})

	t.Run("overshooting counts the excess as bags", func(t *testing.T) {
		p := domain.SpadesBidProgressOf(4, 6)
		assert.Equal(t, domain.SpadesBidMade, p.Kind)
		assert.Equal(t, 2, p.Bags)
	})

	t.Run("a bid not yet placed is treated as nil, not as a met contract", func(t *testing.T) {
		// GetBid() returns -1 before bidding; the Web's `bid <= 0` branch covers
		// it the same way, so the two surfaces stay in step.
		p := domain.SpadesBidProgressOf(-1, 0)
		assert.Equal(t, domain.SpadesBidNilOk, p.Kind)
	})
}
