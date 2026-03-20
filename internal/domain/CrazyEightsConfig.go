package domain

import "fmt"

// CrazyEightsCpuDifficulty CPU の難易度レベル
type CrazyEightsCpuDifficulty int

// CrazyEightsのCPU難易度定数
const (
	// CrazyEightsCpuDifficultyEasy 低難易度
	CrazyEightsCpuDifficultyEasy CrazyEightsCpuDifficulty = iota
	// CrazyEightsCpuDifficultyNormal 中難易度
	CrazyEightsCpuDifficultyNormal
	// CrazyEightsCpuDifficultyHard 高難易度
	CrazyEightsCpuDifficultyHard
)

// CrazyEightsConfig クレイジーエイトゲーム設定
type CrazyEightsConfig struct {
	CpuDifficulty CrazyEightsCpuDifficulty
	PointLimit    int // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultCrazyEightsConfig デフォルト設定を返す
func DefaultCrazyEightsConfig() CrazyEightsConfig {
	return CrazyEightsConfig{
		CpuDifficulty: CrazyEightsCpuDifficultyNormal,
		PointLimit:    200,
	}
}

// Validate 設定値のドメインバリデーション
func (c CrazyEightsConfig) Validate() error {
	if c.CpuDifficulty < CrazyEightsCpuDifficultyEasy || c.CpuDifficulty > CrazyEightsCpuDifficultyHard {
		return fmt.Errorf("CPU difficulty must be %d-%d, got %d", int(CrazyEightsCpuDifficultyEasy), int(CrazyEightsCpuDifficultyHard), int(c.CpuDifficulty))
	}
	if c.PointLimit < 1 {
		return fmt.Errorf("point limit must be >= 1, got %d", c.PointLimit)
	}
	return nil
}
