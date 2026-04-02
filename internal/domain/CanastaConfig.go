package domain

// CanastaCpuDifficulty CPU の難易度レベル
type CanastaCpuDifficulty int

// CanastaのCPU難易度定数
const (
	// CanastaCpuDifficultyEasy 低難易度
	CanastaCpuDifficultyEasy CanastaCpuDifficulty = iota
	// CanastaCpuDifficultyNormal 中難易度
	CanastaCpuDifficultyNormal
	// CanastaCpuDifficultyHard 高難易度
	CanastaCpuDifficultyHard
)

// CanastaConfig カナスタゲーム設定
type CanastaConfig struct {
	CpuDifficulty CanastaCpuDifficulty `json:"cd"`
	PointLimit    int                  `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultCanastaConfig デフォルト設定を返す
func DefaultCanastaConfig() CanastaConfig {
	return CanastaConfig{
		CpuDifficulty: CanastaCpuDifficultyNormal,
		PointLimit:    5000,
	}
}

// Validate 設定値のドメインバリデーション
func (c CanastaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CanastaCpuDifficultyEasy), int(CanastaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
