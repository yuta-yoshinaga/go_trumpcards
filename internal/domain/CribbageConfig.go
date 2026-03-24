package domain

// CribbageCpuDifficulty CPU の難易度レベル
type CribbageCpuDifficulty int

// CribbageのCPU難易度定数
const (
	// CribbageCpuDifficultyEasy 低難易度
	CribbageCpuDifficultyEasy CribbageCpuDifficulty = iota
	// CribbageCpuDifficultyNormal 中難易度
	CribbageCpuDifficultyNormal
	// CribbageCpuDifficultyHard 高難易度
	CribbageCpuDifficultyHard
)

// CribbageConfig クリベッジゲーム設定
type CribbageConfig struct {
	CpuDifficulty CribbageCpuDifficulty
	PointLimit    int // ゲーム終了スコア (先に到達したプレイヤーが勝利、デフォルト121)
}

// DefaultCribbageConfig デフォルト設定を返す
func DefaultCribbageConfig() CribbageConfig {
	return CribbageConfig{
		CpuDifficulty: CribbageCpuDifficultyNormal,
		PointLimit:    121,
	}
}

// Validate 設定値のドメインバリデーション
func (c CribbageConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CribbageCpuDifficultyEasy), int(CribbageCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
