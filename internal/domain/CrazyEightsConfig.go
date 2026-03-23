package domain

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
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CrazyEightsCpuDifficultyEasy), int(CrazyEightsCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
