package domain

// MemoryCpuDifficulty CPU の難易度レベル
type MemoryCpuDifficulty int

const (
	// MemoryCpuDifficultyEasy 低難易度
	MemoryCpuDifficultyEasy MemoryCpuDifficulty = iota
	// MemoryCpuDifficultyNormal 中難易度
	MemoryCpuDifficultyNormal
	// MemoryCpuDifficultyHard 高難易度
	MemoryCpuDifficultyHard
)

// MemoryConfig 神経衰弱ゲーム設定
type MemoryConfig struct {
	CpuDifficulty MemoryCpuDifficulty
}

// DefaultMemoryConfig デフォルト設定を返す
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{CpuDifficulty: MemoryCpuDifficultyNormal}
}
