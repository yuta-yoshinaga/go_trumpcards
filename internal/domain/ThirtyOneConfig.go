//go:build !js || !wasm || solo

package domain

// ThirtyOneCpuDifficulty CPU の難易度レベル
type ThirtyOneCpuDifficulty int

// ThirtyOne の CPU 難易度定数
const (
	// ThirtyOneCpuDifficultyEasy 低難易度
	ThirtyOneCpuDifficultyEasy ThirtyOneCpuDifficulty = iota
	// ThirtyOneCpuDifficultyNormal 中難易度
	ThirtyOneCpuDifficultyNormal
	// ThirtyOneCpuDifficultyHard 高難易度
	ThirtyOneCpuDifficultyHard
)

// ThirtyOneMinLives 設定可能な初期ライフの下限
const ThirtyOneMinLives = 1

// ThirtyOneMaxLives 設定可能な初期ライフの上限
const ThirtyOneMaxLives = 10

// ThirtyOneConfig ThirtyOne ゲーム設定
type ThirtyOneConfig struct {
	CpuDifficulty ThirtyOneCpuDifficulty `json:"cd"`
	InitialLives  int                    `json:"il"` // 各プレイヤーの初期ライフ数
}

// DefaultThirtyOneConfig デフォルト設定を返す
func DefaultThirtyOneConfig() ThirtyOneConfig {
	return ThirtyOneConfig{
		CpuDifficulty: ThirtyOneCpuDifficultyNormal,
		InitialLives:  3,
	}
}

// Validate 設定値のドメインバリデーション
func (c ThirtyOneConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ThirtyOneCpuDifficultyEasy), int(ThirtyOneCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("initial lives", c.InitialLives, ThirtyOneMinLives, ThirtyOneMaxLives); err != nil {
		return err
	}
	return nil
}
