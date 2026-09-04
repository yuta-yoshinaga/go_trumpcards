//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// The Web has warned about a moon attempt since #4483 via
// frontend/src/utils/heartsShootMoonAlert.ts. These cases are that file's rules,
// restated here so the two surfaces cannot drift into disagreeing about what
// counts as "in progress".
func TestHeartsShootTheMoonAlertIdx(t *testing.T) {
	t.Run("flags the player holding every point once the threshold is passed", func(t *testing.T) {
		idx, ok := domain.HeartsShootTheMoonAlertIdx([]int{0, 14, 0, 0})
		assert.True(t, ok)
		assert.Equal(t, 1, idx)
	})

	t.Run("stays silent below the threshold", func(t *testing.T) {
		_, ok := domain.HeartsShootTheMoonAlertIdx([]int{0, 12, 0, 0})
		assert.False(t, ok, "12 points is not yet a moon attempt")
	})

	t.Run("stays silent while the points are split", func(t *testing.T) {
		_, ok := domain.HeartsShootTheMoonAlertIdx([]int{0, 13, 5, 0})
		assert.False(t, ok, "someone else holds points, so no one is shooting")
	})

	t.Run("stays silent when a negative score could be hiding penalties", func(t *testing.T) {
		// Omnibus J-diamond folds a -10 bonus into roundScore, so a negative
		// number cannot be read as "took no penalty cards".
		_, ok := domain.HeartsShootTheMoonAlertIdx([]int{-10, 20, 0, 0})
		assert.False(t, ok)
	})

	t.Run("stays silent on a fresh round", func(t *testing.T) {
		_, ok := domain.HeartsShootTheMoonAlertIdx([]int{0, 0, 0, 0})
		assert.False(t, ok)
	})
}
