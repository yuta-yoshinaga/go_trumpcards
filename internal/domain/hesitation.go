package domain

import "math"

// 迷い時間ディレイ (ミリ秒) の共通範囲定数
const (
	hesitationFastMin   = 300 // 速い反応 (ペア成立, ブラフ急ぎ)
	hesitationFastMax   = 500
	hesitationMediumMin = 600 // 中間の反応 (正直, 通常)
	hesitationMediumMax = 1000
)

// Meta-AI adaptation tuning constants, shared by every HumanProfile.
const (
	// hesitationMinPlays 迷い時間ブーストを有効にするための最小データ点数
	hesitationMinPlays = 3
	// hesitationZThreshold z-scoreがこの値を超えた場合にブーストが発生する
	hesitationZThreshold = 1.0
	// hesitationWeight z-score超過分に対するブースト重み
	hesitationWeight = 0.05
	// maxHesitationBoost 迷い時間ブーストの上限
	maxHesitationBoost = 0.10
	// metaAIMaxAdaptGames 適応が最大に達するゲーム数
	metaAIMaxAdaptGames = 5
	// metaAIAdaptPerGame 1ゲームあたりの適応強度
	metaAIAdaptPerGame = 0.04
)

// Welford's online algorithm and meta-AI adaptation helpers shared by every
// HumanProfile that tracks hesitation time. The three profile structs
// (Betting, IndianPoker, Doubt) all need:
//
//   - Streaming mean/variance of hesitation times (numerically stable).
//   - A bounded linear ramp from GamesPlayed → adapt strength.
//   - A z-score-based probability boost once enough samples exist.
//
// All helpers below group the Welford accumulator triple (count, mean, m2)
// first and put the per-call sample `ms` last, so call sites can skim
// signatures consistently.

// welfordUpdate applies one sample of Welford's online algorithm to the
// given streaming accumulators. Samples where ms <= 0 are ignored (used by
// callers that cannot measure, e.g. the CUI).
//
// Mutates *count, *mean, and *m2 in place so that the caller's JSON-tagged
// fields remain the source of truth.
func welfordUpdate(count *int, mean, m2 *float64, ms int) {
	if ms <= 0 {
		return
	}
	*count++
	x := float64(ms)
	delta := x - *mean
	*mean += delta / float64(*count)
	delta2 := x - *mean
	*m2 += delta * delta2
}

// welfordStdDev returns the sample standard deviation from Welford
// accumulators. Returns 0 when count < 2 (undefined, not a real deviation).
func welfordStdDev(count int, m2 float64) float64 {
	if count < 2 {
		return 0
	}
	return math.Sqrt(m2 / float64(count-1))
}

// welfordZScore returns (ms - mean) / stddev for the current sample against
// the accumulator, or 0 when the stddev is 0 (insufficient data).
func welfordZScore(count int, mean, m2 float64, ms int) float64 {
	sd := welfordStdDev(count, m2)
	if sd == 0 {
		return 0
	}
	return (float64(ms) - mean) / sd
}

// hesitationBoost returns a positive probability-adjustment when the sample
// ms is sufficiently slower than the player's own baseline. Returns 0 when
// data is insufficient (count < hesitationMinPlays), z-score is at/below
// threshold, or clamped at maxHesitationBoost.
//
// Shared by every HumanProfile that applies a "player hesitated → lean
// toward a call/doubt/bluff" adjustment.
func hesitationBoost(count int, mean, m2 float64, ms int) float64 {
	if count < hesitationMinPlays {
		return 0
	}
	z := welfordZScore(count, mean, m2, ms)
	if z <= hesitationZThreshold {
		return 0
	}
	boost := (z - hesitationZThreshold) * hesitationWeight
	if boost > maxHesitationBoost {
		return maxHesitationBoost
	}
	return boost
}

// computeAdaptStrength ramps linearly from 0 to metaAIMaxAdaptGames *
// metaAIAdaptPerGame as the session progresses. Games beyond
// metaAIMaxAdaptGames add no further adaptation (clamped).
func computeAdaptStrength(gamesPlayed int) float64 {
	games := min(gamesPlayed, metaAIMaxAdaptGames)
	return float64(games) * metaAIAdaptPerGame
}
