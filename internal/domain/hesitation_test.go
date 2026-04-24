package domain

import (
	"math"
	"testing"
)

// Welford's algorithm must be numerically stable — a sanity property test
// that the shared helpers match a two-pass mean/variance on well-behaved
// input.
func TestWelfordUpdate_MatchesTwoPass(t *testing.T) {
	samples := []int{450, 520, 610, 300, 780, 900, 1100, 420, 560, 680}

	// Reference pass: naive mean/variance.
	sum := 0.0
	for _, s := range samples {
		sum += float64(s)
	}
	wantMean := sum / float64(len(samples))
	var ssq float64
	for _, s := range samples {
		diff := float64(s) - wantMean
		ssq += diff * diff
	}
	wantSD := math.Sqrt(ssq / float64(len(samples)-1))

	// Streaming pass via the helper under test.
	var count int
	var mean, m2 float64
	for _, s := range samples {
		welfordUpdate(&count, &mean, &m2, s)
	}
	gotSD := welfordStdDev(count, m2)

	if count != len(samples) {
		t.Errorf("count = %d, want %d", count, len(samples))
	}
	if math.Abs(mean-wantMean) > 1e-9 {
		t.Errorf("mean = %f, want %f", mean, wantMean)
	}
	if math.Abs(gotSD-wantSD) > 1e-9 {
		t.Errorf("stddev = %f, want %f", gotSD, wantSD)
	}
}

// welfordUpdate must ignore non-positive samples (CUI can't measure) so
// counts never get inflated and the mean stays pure.
func TestWelfordUpdate_IgnoresNonPositive(t *testing.T) {
	tests := []struct {
		name   string
		sample int
	}{
		{"zero", 0},
		{"negative", -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int
			var mean, m2 float64
			welfordUpdate(&count, &mean, &m2, tt.sample)
			if count != 0 || mean != 0 || m2 != 0 {
				t.Errorf("non-positive sample %d mutated accumulators: count=%d mean=%f m2=%f",
					tt.sample, count, mean, m2)
			}
		})
	}
}

// welfordStdDev must return 0 when count < 2 (sample variance needs n-1).
func TestWelfordStdDev_InsufficientData(t *testing.T) {
	if got := welfordStdDev(0, 0); got != 0 {
		t.Errorf("count=0 stddev = %f, want 0", got)
	}
	if got := welfordStdDev(1, 123.4); got != 0 {
		t.Errorf("count=1 stddev = %f, want 0 (single sample has undefined variance)", got)
	}
}

// welfordZScore must return 0 when the accumulator has no variance signal.
func TestWelfordZScore_InsufficientData(t *testing.T) {
	if got := welfordZScore(1, 500, 0, 500); got != 0 {
		t.Errorf("single-sample z-score = %f, want 0", got)
	}
}

// computeAdaptStrength clamps games above metaAIMaxAdaptGames so a
// long-running session never pushes adapt strength past its ceiling.
func TestComputeAdaptStrength_ClampsAtMax(t *testing.T) {
	cap := float64(metaAIMaxAdaptGames) * metaAIAdaptPerGame
	tests := []struct {
		games int
		want  float64
	}{
		{0, 0},
		{1, metaAIAdaptPerGame},
		{metaAIMaxAdaptGames, cap},
		{metaAIMaxAdaptGames + 1, cap}, // clamped
		{1_000_000, cap},               // still clamped
	}
	for _, tt := range tests {
		if got := computeAdaptStrength(tt.games); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("computeAdaptStrength(%d) = %f, want %f", tt.games, got, tt.want)
		}
	}
}

// hesitationBoost must return 0 until there are at least hesitationMinPlays
// samples, regardless of how extreme the z-score would otherwise be.
func TestHesitationBoost_MinimumSamples(t *testing.T) {
	// Populate with just under the threshold so a z-score CAN be computed
	// but the boost path short-circuits on count < hesitationMinPlays.
	var count int
	var mean, m2 float64
	for i := 0; i < hesitationMinPlays-1; i++ {
		welfordUpdate(&count, &mean, &m2, 500+i*100)
	}
	if got := hesitationBoost(count, mean, m2, 5000); got != 0 {
		t.Errorf("hesitationBoost with count=%d = %f, want 0 (below hesitationMinPlays=%d)",
			count, got, hesitationMinPlays)
	}
}

// hesitationBoost must cap at maxHesitationBoost even when the raw z-score
// overshoot is enormous.
func TestHesitationBoost_CapsAtMax(t *testing.T) {
	var count int
	var mean, m2 float64
	for _, s := range []int{500, 510, 490, 505, 495} {
		welfordUpdate(&count, &mean, &m2, s)
	}
	// 100x the mean → z-score in the tens; boost must still clamp.
	got := hesitationBoost(count, mean, m2, 50_000)
	if got != maxHesitationBoost {
		t.Errorf("hesitationBoost(huge-ms) = %f, want capped at %f", got, maxHesitationBoost)
	}
}

// hesitationBoost must return 0 when the z-score sits at or below the
// configured threshold — a player who didn't visibly hesitate gives the
// CPU no signal, even if enough samples exist. Covers the branch that the
// "cap" and "minimum samples" tests miss.
func TestHesitationBoost_BelowThreshold(t *testing.T) {
	var count int
	var mean, m2 float64
	for _, s := range []int{500, 510, 490, 505, 495} {
		welfordUpdate(&count, &mean, &m2, s)
	}
	// 501 ms is barely above the ~500 ms cluster — z-score well below 1.0.
	if got := hesitationBoost(count, mean, m2, 501); got != 0 {
		t.Errorf("hesitationBoost(barely-above-mean) = %f, want 0 (z <= threshold)", got)
	}
}

// hesitationBoost must return a positive, uncapped value when the z-score
// is moderately above the threshold — the linear ramp before the max clamp
// needs coverage too.
func TestHesitationBoost_PositiveBelowCap(t *testing.T) {
	var count int
	var mean, m2 float64
	for _, s := range []int{500, 510, 490, 505, 495} {
		welfordUpdate(&count, &mean, &m2, s)
	}
	// ~520 ms is ≈2 stddevs above the tight cluster around 500 — above the
	// 1.0 threshold but nowhere near the boost cap.
	got := hesitationBoost(count, mean, m2, 520)
	if got <= 0 {
		t.Errorf("hesitationBoost(520) = %f, want positive", got)
	}
	if got >= maxHesitationBoost {
		t.Errorf("hesitationBoost(520) = %f, must stay strictly below cap %f", got, maxHesitationBoost)
	}
}
