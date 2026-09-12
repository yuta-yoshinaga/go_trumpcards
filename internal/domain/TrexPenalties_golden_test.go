//go:build test

package domain_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

type trexPenaltiesGolden struct {
	KingOfHearts int `json:"kingOfHearts"`
	Diamond      int `json:"diamond"`
	Queen        int `json:"queen"`
}

// TestTrexPenalties_GoldenValues verifies that the penalty values used in frontend
// (frontend/src/utils/__fixtures__/trexPenalties.golden.json) match the domain constants.
func TestTrexPenalties_GoldenValues(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "utils", "__fixtures__", "trexPenalties.golden.json"))
	require.NoError(t, err)

	var golden trexPenaltiesGolden
	require.NoError(t, json.Unmarshal(raw, &golden))

	assert.Equal(t, domain.TrexKingOfHeartsPenalty, golden.KingOfHearts)
	assert.Equal(t, domain.TrexDiamondPenalty, golden.Diamond)
	assert.Equal(t, domain.TrexQueenPenalty, golden.Queen)
}
