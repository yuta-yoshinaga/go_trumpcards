package domain

// TonkCpuDifficulty CPU の難易度レベル
type TonkCpuDifficulty int

// TonkのCPU難易度定数
const (
	// TonkCpuDifficultyEasy 低難易度
	TonkCpuDifficultyEasy TonkCpuDifficulty = iota
	// TonkCpuDifficultyNormal 中難易度
	TonkCpuDifficultyNormal
	// TonkCpuDifficultyHard 高難易度
	TonkCpuDifficultyHard
)

// TonkConfig Tonkゲーム設定
type TonkConfig struct {
	CpuDifficulty TonkCpuDifficulty `json:"cd"`
	PointLimit    int               `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultTonkConfig デフォルト設定を返す
func DefaultTonkConfig() TonkConfig {
	return TonkConfig{
		CpuDifficulty: TonkCpuDifficultyNormal,
		PointLimit:    50,
	}
}

// Validate 設定値のドメインバリデーション
func (c TonkConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TonkCpuDifficultyEasy), int(TonkCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
