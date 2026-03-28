package domain

// EuchreCpuDifficulty CPU の難易度レベル
type EuchreCpuDifficulty int

// EuchreのCPU難易度定数
const (
	// EuchreCpuDifficultyEasy 低難易度
	EuchreCpuDifficultyEasy EuchreCpuDifficulty = iota
	// EuchreCpuDifficultyNormal 中難易度
	EuchreCpuDifficultyNormal
	// EuchreCpuDifficultyHard 高難易度
	EuchreCpuDifficultyHard
)

// EuchreConfig ユーカーゲーム設定
type EuchreConfig struct {
	CpuDifficulty EuchreCpuDifficulty `json:"cd"`
	PointLimit    int                 `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト10)
}

// DefaultEuchreConfig デフォルト設定を返す
func DefaultEuchreConfig() EuchreConfig {
	return EuchreConfig{
		CpuDifficulty: EuchreCpuDifficultyNormal,
		PointLimit:    10,
	}
}

// Validate 設定値のドメインバリデーション
func (c EuchreConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(EuchreCpuDifficultyEasy), int(EuchreCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
