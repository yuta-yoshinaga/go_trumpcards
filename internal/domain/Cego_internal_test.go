//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCegoScoreDealZeroSum(t *testing.T) {
	for _, points := range []int{0, 20, 53, 54, 70, 106} {
		breakdown := cegoScoreDeal(points, 1)
		assert.Equalf(t, 0, breakdown.DeclarerScore+3*breakdown.OpponentScore, "points=%d", points)
		assert.Equal(t, points > 53, breakdown.Won)
	}
}

func TestCegoScoreDealThreshold(t *testing.T) {
	assert.False(t, cegoScoreDeal(53, 1).Won)
	assert.True(t, cegoScoreDeal(54, 1).Won)
	assert.Equal(t, 53, cegoScoreDeal(54, 1).Threshold)
}
