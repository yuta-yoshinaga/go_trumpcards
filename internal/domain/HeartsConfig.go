package domain

import "fmt"

// HeartsCpuDifficulty CPU の難易度レベル
type HeartsCpuDifficulty int

const (
	// HeartsCpuDifficultyEasy 低難易度
	HeartsCpuDifficultyEasy HeartsCpuDifficulty = iota
	// HeartsCpuDifficultyNormal 中難易度
	HeartsCpuDifficultyNormal
	// HeartsCpuDifficultyHard 高難易度
	HeartsCpuDifficultyHard
)

// HeartsConfig ハーツゲーム設定
type HeartsConfig struct {
	CpuDifficulty HeartsCpuDifficulty
	PointLimit    int // ゲーム終了スコア (いずれかのプレイヤーがこの点数に達したら終了)
}

// DefaultHeartsConfig デフォルト設定を返す
func DefaultHeartsConfig() HeartsConfig {
	return HeartsConfig{CpuDifficulty: HeartsCpuDifficultyNormal, PointLimit: 100}
}

// Validate 設定値のドメインバリデーション
func (c HeartsConfig) Validate() error {
	if c.CpuDifficulty < HeartsCpuDifficultyEasy || c.CpuDifficulty > HeartsCpuDifficultyHard {
		return fmt.Errorf("CPU difficulty must be %d-%d, got %d", int(HeartsCpuDifficultyEasy), int(HeartsCpuDifficultyHard), int(c.CpuDifficulty))
	}
	if c.PointLimit < 1 {
		return fmt.Errorf("point limit must be >= 1, got %d", c.PointLimit)
	}
	return nil
}
