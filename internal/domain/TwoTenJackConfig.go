package domain

// TwoTenJackCpuDifficulty CPUの難易度レベル
type TwoTenJackCpuDifficulty int

// TwoTenJackのCPU難易度定数
const (
	// TwoTenJackCpuDifficultyEasy 低難易度
	TwoTenJackCpuDifficultyEasy TwoTenJackCpuDifficulty = iota
	// TwoTenJackCpuDifficultyNormal 中難易度
	TwoTenJackCpuDifficultyNormal
	// TwoTenJackCpuDifficultyHard 高難易度
	TwoTenJackCpuDifficultyHard
)

// TwoTenJackConfig ツーテンジャックゲーム設定
type TwoTenJackConfig struct {
	CpuDifficulty TwoTenJackCpuDifficulty `json:"cd"`
	PointLimit    int                     `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利)
}

// DefaultTwoTenJackConfig デフォルト設定を返す
func DefaultTwoTenJackConfig() TwoTenJackConfig {
	return TwoTenJackConfig{
		CpuDifficulty: TwoTenJackCpuDifficultyNormal,
		PointLimit:    50,
	}
}

// Validate 設定値のドメインバリデーション
func (c TwoTenJackConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TwoTenJackCpuDifficultyEasy), int(TwoTenJackCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
