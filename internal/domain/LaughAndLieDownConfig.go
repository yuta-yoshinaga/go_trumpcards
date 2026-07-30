//go:build !js || !wasm || extra2

package domain

// LaughAndLieDownCpuDifficulty は CPU の難易度レベル。
type LaughAndLieDownCpuDifficulty int

// Laugh and Lie Down の CPU 難易度定数
const (
	// LaughAndLieDownCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	LaughAndLieDownCpuDifficultyNormal LaughAndLieDownCpuDifficulty = iota
)

// LaughAndLieDownConfig は Laugh and Lie Down のゲーム設定。
type LaughAndLieDownConfig struct {
	CpuDifficulty LaughAndLieDownCpuDifficulty `json:"cd"`
}

// DefaultLaughAndLieDownConfig はデフォルト設定を返す。
func DefaultLaughAndLieDownConfig() LaughAndLieDownConfig {
	return LaughAndLieDownConfig{CpuDifficulty: LaughAndLieDownCpuDifficultyNormal}
}

// Validate は設定値のドメインバリデーション。
func (c LaughAndLieDownConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(LaughAndLieDownCpuDifficultyNormal), int(LaughAndLieDownCpuDifficultyNormal))
}
