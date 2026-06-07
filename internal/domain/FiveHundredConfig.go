//go:build !js || !wasm || solo

package domain

// FiveHundredCpuDifficulty CPU の難易度レベル
type FiveHundredCpuDifficulty int

// FiveHundred の CPU 難易度定数
const (
	// FiveHundredCpuDifficultyEasy 低難易度
	FiveHundredCpuDifficultyEasy FiveHundredCpuDifficulty = iota
	// FiveHundredCpuDifficultyNormal 中難易度
	FiveHundredCpuDifficultyNormal
	// FiveHundredCpuDifficultyHard 高難易度
	FiveHundredCpuDifficultyHard
)

// FiveHundredDefaultTargetScore ゲーム終了スコア (先に到達したチームが勝利)
const FiveHundredDefaultTargetScore = 500

// FiveHundredConfig 500 (Five Hundred) ゲーム設定
type FiveHundredConfig struct {
	CpuDifficulty FiveHundredCpuDifficulty `json:"cd"`
	TargetScore   int                      `json:"ts"` // 勝利スコア (デフォルト500)
}

// DefaultFiveHundredConfig デフォルト設定を返す
func DefaultFiveHundredConfig() FiveHundredConfig {
	return FiveHundredConfig{
		CpuDifficulty: FiveHundredCpuDifficultyNormal,
		TargetScore:   FiveHundredDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c FiveHundredConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(FiveHundredCpuDifficultyEasy), int(FiveHundredCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	return nil
}
