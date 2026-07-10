package domain

// CatchTenCpuDifficulty CPU の難易度レベル
type CatchTenCpuDifficulty int

// CatchTenのCPU難易度定数
const (
	// CatchTenCpuDifficultyEasy 低難易度
	CatchTenCpuDifficultyEasy CatchTenCpuDifficulty = iota
	// CatchTenCpuDifficultyNormal 中難易度
	CatchTenCpuDifficultyNormal
	// CatchTenCpuDifficultyHard 高難易度
	CatchTenCpuDifficultyHard
)

// CatchTenConfig Catch the Ten (Scotch Whist) ゲーム設定
type CatchTenConfig struct {
	CpuDifficulty CatchTenCpuDifficulty `json:"cd"`
	PointLimit    int                   `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト41)
}

// DefaultCatchTenConfig デフォルト設定を返す
func DefaultCatchTenConfig() CatchTenConfig {
	return CatchTenConfig{
		CpuDifficulty: CatchTenCpuDifficultyNormal,
		PointLimit:    41,
	}
}

// Validate 設定値のドメインバリデーション
func (c CatchTenConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CatchTenCpuDifficultyEasy), int(CatchTenCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
