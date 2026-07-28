//go:build !js || !wasm || solo

package domain

// MemoryCpuDifficulty CPU の難易度レベル
type MemoryCpuDifficulty int

// MemoryのCPU難易度定数
const (
	// MemoryCpuDifficultyEasy 低難易度
	MemoryCpuDifficultyEasy MemoryCpuDifficulty = iota
	// MemoryCpuDifficultyNormal 中難易度
	MemoryCpuDifficultyNormal
	// MemoryCpuDifficultyHard 高難易度
	MemoryCpuDifficultyHard
)

// ペア数の許容範囲 (ADR-0035)。
//
// 上限 26 はフルデッキ 52 枚（従来動作）。下限 6 は 12 枚で、記憶ゲームとして
// 成立する最小規模として選んだ。モバイル縦 375x667 では 44x44 のタップターゲット
// を守ると 7 列 x 6 行 = 42 枚しか置けないため、フロントエンドは狭幅で 20 ペアを
// 既定にする。経緯は docs/adr/0035-memory-mobile-pair-count.md。
const (
	// MemoryMinPairCount 選択できる最小ペア数
	MemoryMinPairCount = 6
	// MemoryMaxPairCount 選択できる最大ペア数（フルデッキ）
	MemoryMaxPairCount = MemoryBoardSize / 2
)

// MemoryConfig 神経衰弱ゲーム設定
type MemoryConfig struct {
	CpuDifficulty MemoryCpuDifficulty `json:"cd"`
	// PairCount 盤面に並べるペア数。0 は「未設定」を意味し既定として扱う。
	// ゼロ値でも従来どおりフルデッキで開始できるようにするためで、
	// PairCount を知らない既存の KV スナップショットもこれで動き続ける。
	PairCount int `json:"pc"`
}

// DefaultMemoryConfig デフォルト設定を返す
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{CpuDifficulty: MemoryCpuDifficultyNormal, PairCount: MemoryMaxPairCount}
}

// NormalizedPairCount 許容範囲に収めたペア数を返す。
//
// 0（未設定）は既定へ、範囲外はクランプする。エラーにしないのは、この値が UI の
// 選択肢に由来するもので、不正値でゲーム開始そのものを失敗させる理由がないため。
func (c MemoryConfig) NormalizedPairCount() int {
	switch {
	case c.PairCount == 0:
		return MemoryMaxPairCount
	case c.PairCount < MemoryMinPairCount:
		return MemoryMinPairCount
	case c.PairCount > MemoryMaxPairCount:
		return MemoryMaxPairCount
	default:
		return c.PairCount
	}
}
