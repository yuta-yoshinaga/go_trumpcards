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

// sideBetPayoutsGolden mirrors the structure of
// frontend/src/utils/__fixtures__/blackjackSideBetPayouts.golden.json.
type sideBetPayoutsGolden struct {
	PerfectPairs struct {
		Mixed   int `json:"mixed"`
		Colored int `json:"colored"`
		Perfect int `json:"perfect"`
	} `json:"perfectPairs"`
	TwentyOnePlus3 struct {
		Flush         int `json:"flush"`
		Straight      int `json:"straight"`
		Trips         int `json:"trips"`
		StraightFlush int `json:"straightFlush"`
		SuitedTrips   int `json:"suitedTrips"`
	} `json:"twentyOnePlus3"`
}

// TestBlackJackSideBetPayouts_GoldenValues asserts that the golden fixture
// used by the TypeScript front-end matches the authoritative Go domain
// constants.  Both sides are stated independently so that changing either the
// JSON or the Go constants alone is guaranteed to fail this test.
//
// Source of truth: internal/domain/BlackJackSideBet.go
func TestBlackJackSideBetPayouts_GoldenValues(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "utils", "__fixtures__", "blackjackSideBetPayouts.golden.json"))
	require.NoError(t, err, "golden fixture must be readable")

	var g sideBetPayoutsGolden
	require.NoError(t, json.Unmarshal(raw, &g), "golden fixture must parse as JSON")

	// Perfect Pairs — each value is stated twice (fixture + constant) so that
	// changing only one side causes a mismatch.
	assert.Equal(t, domain.BJPPMixedPairPayout, g.PerfectPairs.Mixed,
		"perfectPairs.mixed must match BJPPMixedPairPayout")
	assert.Equal(t, domain.BJPPColoredPairPayout, g.PerfectPairs.Colored,
		"perfectPairs.colored must match BJPPColoredPairPayout")
	assert.Equal(t, domain.BJPPPerfectPairPayout, g.PerfectPairs.Perfect,
		"perfectPairs.perfect must match BJPPPerfectPairPayout")

	// 21+3 (Poker Hand Bonus)
	assert.Equal(t, domain.BJT3FlushPayout, g.TwentyOnePlus3.Flush,
		"twentyOnePlus3.flush must match BJT3FlushPayout")
	assert.Equal(t, domain.BJT3StraightPayout, g.TwentyOnePlus3.Straight,
		"twentyOnePlus3.straight must match BJT3StraightPayout")
	assert.Equal(t, domain.BJT3ThreeOfAKindPayout, g.TwentyOnePlus3.Trips,
		"twentyOnePlus3.trips must match BJT3ThreeOfAKindPayout")
	assert.Equal(t, domain.BJT3StraightFlushPayout, g.TwentyOnePlus3.StraightFlush,
		"twentyOnePlus3.straightFlush must match BJT3StraightFlushPayout")
	assert.Equal(t, domain.BJT3SuitedTripsPayout, g.TwentyOnePlus3.SuitedTrips,
		"twentyOnePlus3.suitedTrips must match BJT3SuitedTripsPayout")
}
