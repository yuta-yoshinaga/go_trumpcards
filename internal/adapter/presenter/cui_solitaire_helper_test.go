//go:build test
// +build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// pile builds a pile of n placeholder cards. Only the length matters here --
// the summary counts cards, it never inspects them.
func pile(n int) []*domain.Card {
	p := make([]*domain.Card, n)
	for i := range p {
		p[i] = &domain.Card{}
	}
	return p
}

func TestCuiCountPileCards(t *testing.T) {
	t.Run("sums every pile", func(t *testing.T) {
		assert.Equal(t, 6, cuiCountPileCards(pile(1), pile(2), pile(3)))
	})

	t.Run("no piles is zero", func(t *testing.T) {
		assert.Equal(t, 0, cuiCountPileCards())
	})

	t.Run("empty piles are zero", func(t *testing.T) {
		assert.Equal(t, 0, cuiCountPileCards(pile(0), pile(0)))
	})

	t.Run("nil pile contributes nothing", func(t *testing.T) {
		assert.Equal(t, 2, cuiCountPileCards(nil, pile(2)))
	})
}

func TestCuiSolitaireProgressPercent(t *testing.T) {
	cases := []struct {
		name         string
		count, total int
		want         int
	}{
		{"none reached", 0, 52, 0},
		{"all reached", 52, 52, 100},
		{"half reached", 26, 52, 50},
		{"rounds to nearest", 17, 52, 33},    // 32.69 -> 33
		{"rounds up at .5", 26, 104, 25},     // 25.0 exactly
		{"double deck partial", 51, 104, 49}, // 49.03 -> 49
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, cuiSolitaireProgressPercent(c.count, c.total))
		})
	}

	t.Run("zero total does not divide by zero", func(t *testing.T) {
		assert.Equal(t, 0, cuiSolitaireProgressPercent(3, 0))
	})

	t.Run("negative total is treated as zero", func(t *testing.T) {
		assert.Equal(t, 0, cuiSolitaireProgressPercent(3, -1))
	})
}

func TestCuiSolitaireGameOverSummary(t *testing.T) {
	t.Run("includes count, total and percent", func(t *testing.T) {
		got := cuiSolitaireGameOverSummary(26, 52)
		assert.Contains(t, got, "26")
		assert.Contains(t, got, "52")
		assert.Contains(t, got, "50")
	})

	t.Run("zero total still renders without dividing by zero", func(t *testing.T) {
		got := cuiSolitaireGameOverSummary(0, 0)
		assert.Contains(t, got, "0")
	})
}

func TestCuiSolitaireGameOverFaces(t *testing.T) {
	t.Run("includes completed and total faces", func(t *testing.T) {
		got := cuiSolitaireGameOverFaces(7, 12)
		assert.Contains(t, got, "7")
		assert.Contains(t, got, "12")
	})
}
