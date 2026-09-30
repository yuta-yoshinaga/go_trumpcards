//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKoenigrufenScoreDealPartnershipWinZeroSum(t *testing.T) {
	// team captured 60 (> 53) → won. diff 7, base 17.
	bd := koenigrufenScoreDeal(60, false, 1)
	assert.True(t, bd.Won)
	assert.False(t, bd.Solo)
	assert.Equal(t, 53, bd.Threshold)
	assert.Equal(t, 7, bd.Diff)
	assert.Equal(t, 17, bd.Base)
	assert.Equal(t, 17, bd.DeclarerScore)
	assert.Equal(t, 17, bd.PartnerScore)
	assert.Equal(t, -17, bd.OpponentScore)
	// zero-sum: declarer + partner + 2 opponents.
	assert.Equal(t, 0, bd.DeclarerScore+bd.PartnerScore+2*bd.OpponentScore)
}

func TestKoenigrufenScoreDealPartnershipLoss(t *testing.T) {
	bd := koenigrufenScoreDeal(40, false, 1)
	assert.False(t, bd.Won)
	assert.Equal(t, 13, bd.Diff)
	assert.Equal(t, 23, bd.Base)
	assert.Equal(t, -23, bd.DeclarerScore)
	assert.Equal(t, -23, bd.PartnerScore)
	assert.Equal(t, 23, bd.OpponentScore)
	assert.Equal(t, 0, bd.DeclarerScore+bd.PartnerScore+2*bd.OpponentScore)
}

func TestKoenigrufenScoreDealSoloZeroSum(t *testing.T) {
	bd := koenigrufenScoreDeal(60, true, 1)
	assert.True(t, bd.Won)
	assert.True(t, bd.Solo)
	assert.Equal(t, 51, bd.DeclarerScore) // 3 × 17
	assert.Equal(t, 0, bd.PartnerScore)
	assert.Equal(t, -17, bd.OpponentScore)
	// zero-sum: declarer + 3 opponents.
	assert.Equal(t, 0, bd.DeclarerScore+3*bd.OpponentScore)
}

func TestKoenigrufenScoreDealExactHalfIsLoss(t *testing.T) {
	// exactly half (53 of 106) is NOT more than half → loss.
	bd := koenigrufenScoreDeal(53, false, 1)
	assert.False(t, bd.Won)
	assert.Equal(t, 0, bd.Diff)
	assert.Equal(t, 10, bd.Base)
}

// --- Bidding ---
