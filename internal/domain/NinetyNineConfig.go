package domain

// NinetyNineCpuDifficulty CPU の難易度レベル
type NinetyNineCpuDifficulty int

// NinetyNineのCPU難易度定数
const (
	// NinetyNineCpuDifficultyEasy 低難易度
	NinetyNineCpuDifficultyEasy NinetyNineCpuDifficulty = iota
	// NinetyNineCpuDifficultyNormal 中難易度
	NinetyNineCpuDifficultyNormal
	// NinetyNineCpuDifficultyHard 高難易度
	NinetyNineCpuDifficultyHard
)

// NinetyNineConfig ナインティナインゲーム設定
type NinetyNineConfig struct {
	CpuDifficulty NinetyNineCpuDifficulty `json:"cd"`
	TargetScore   int                     `json:"ts"` // ゲーム勝利に必要な累積スコア (デフォルト100)
}

// DefaultNinetyNineConfig デフォルト設定を返す
func DefaultNinetyNineConfig() NinetyNineConfig {
	return NinetyNineConfig{
		CpuDifficulty: NinetyNineCpuDifficultyNormal,
		TargetScore:   100,
	}
}

// Validate 設定値のドメインバリデーション
func (c NinetyNineConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(NinetyNineCpuDifficultyEasy), int(NinetyNineCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("target score", c.TargetScore, 10, 1000); err != nil {
		return err
	}
	return nil
}
