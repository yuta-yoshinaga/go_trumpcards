//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func memCard(design, value int, faceUp, taken, visited bool) *domain.MemoryBoardCard {
	return &domain.MemoryBoardCard{
		Card:    domain.NewCard(design, value, false),
		FaceUp:  faceUp,
		Taken:   taken,
		Visited: visited,
	}
}

// Mirrors frontend/src/utils/memoryKnownMatch.ts so the two surfaces cannot
// drift into disagreeing about when a match is known.
func TestMemoryKnownMatchIdx(t *testing.T) {
	t.Run("points at a seen face-down card matching the single face-up one", func(t *testing.T) {
		board := []*domain.MemoryBoardCard{
			memCard(domain.CardDesignHeart, 5, true, false, true),
			memCard(domain.CardDesignSpade, 9, false, false, true),
			memCard(domain.CardDesignHeart, 5, false, false, true),
		}
		idx, ok := domain.MemoryKnownMatchIdx(board)
		assert.True(t, ok)
		assert.Equal(t, 2, idx)
	})

	t.Run("ignores a matching card that was never turned over", func(t *testing.T) {
		board := []*domain.MemoryBoardCard{
			memCard(domain.CardDesignHeart, 5, true, false, true),
			memCard(domain.CardDesignHeart, 5, false, false, false),
		}
		_, ok := domain.MemoryKnownMatchIdx(board)
		assert.False(t, ok, "recall only covers cards the player actually saw")
	})

	t.Run("stays silent with two cards face up", func(t *testing.T) {
		board := []*domain.MemoryBoardCard{
			memCard(domain.CardDesignHeart, 5, true, false, true),
			memCard(domain.CardDesignSpade, 9, true, false, true),
			memCard(domain.CardDesignHeart, 5, false, false, true),
		}
		_, ok := domain.MemoryKnownMatchIdx(board)
		assert.False(t, ok, "the turn cannot be played, so the hint would mislead")
	})

	t.Run("stays silent with nothing face up", func(t *testing.T) {
		board := []*domain.MemoryBoardCard{
			memCard(domain.CardDesignHeart, 5, false, false, true),
			memCard(domain.CardDesignHeart, 5, false, false, true),
		}
		_, ok := domain.MemoryKnownMatchIdx(board)
		assert.False(t, ok)
	})

	t.Run("ignores taken pairs", func(t *testing.T) {
		board := []*domain.MemoryBoardCard{
			memCard(domain.CardDesignHeart, 5, true, false, true),
			memCard(domain.CardDesignHeart, 5, false, true, true),
		}
		_, ok := domain.MemoryKnownMatchIdx(board)
		assert.False(t, ok)
	})
}
