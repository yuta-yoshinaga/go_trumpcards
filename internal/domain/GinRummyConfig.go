package domain

import "fmt"

// GinRummyCpuDifficulty CPU の難易度レベル
type GinRummyCpuDifficulty int

// GinRummyのCPU難易度定数
const (
	// GinRummyCpuDifficultyEasy 低難易度
	GinRummyCpuDifficultyEasy GinRummyCpuDifficulty = iota
	// GinRummyCpuDifficultyNormal 中難易度
	GinRummyCpuDifficultyNormal
	// GinRummyCpuDifficultyHard 高難易度
	GinRummyCpuDifficultyHard
)

// GinRummyConfig ジンラミーゲーム設定
type GinRummyConfig struct {
	CpuDifficulty GinRummyCpuDifficulty
	PointLimit    int // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultGinRummyConfig デフォルト設定を返す
func DefaultGinRummyConfig() GinRummyConfig {
	return GinRummyConfig{
		CpuDifficulty: GinRummyCpuDifficultyNormal,
		PointLimit:    100,
	}
}

// Validate 設定値のドメインバリデーション
func (c GinRummyConfig) Validate() error {
	if c.CpuDifficulty < GinRummyCpuDifficultyEasy || c.CpuDifficulty > GinRummyCpuDifficultyHard {
		return fmt.Errorf("CPU difficulty must be %d-%d, got %d", int(GinRummyCpuDifficultyEasy), int(GinRummyCpuDifficultyHard), int(c.CpuDifficulty))
	}
	if c.PointLimit < 1 {
		return fmt.Errorf("point limit must be >= 1, got %d", c.PointLimit)
	}
	return nil
}
