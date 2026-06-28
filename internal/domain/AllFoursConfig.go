package domain

// AllFoursCpuDifficulty CPU の難易度レベル
type AllFoursCpuDifficulty int

// AllFoursのCPU難易度定数
const (
	// AllFoursCpuDifficultyEasy 低難易度
	AllFoursCpuDifficultyEasy AllFoursCpuDifficulty = iota
	// AllFoursCpuDifficultyNormal 中難易度
	AllFoursCpuDifficultyNormal
	// AllFoursCpuDifficultyHard 高難易度
	AllFoursCpuDifficultyHard
)

// AllFoursConfig All Fours (Seven Up) ゲーム設定
type AllFoursConfig struct {
	CpuDifficulty AllFoursCpuDifficulty `json:"cd"`
	PointLimit    int                   `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利, デフォルト7)
}

// DefaultAllFoursConfig デフォルト設定を返す
func DefaultAllFoursConfig() AllFoursConfig {
	return AllFoursConfig{
		CpuDifficulty: AllFoursCpuDifficultyNormal,
		PointLimit:    7,
	}
}

// Validate 設定値のドメインバリデーション
func (c AllFoursConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(AllFoursCpuDifficultyEasy), int(AllFoursCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
