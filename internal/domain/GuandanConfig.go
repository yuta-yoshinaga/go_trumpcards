//go:build !js || !wasm || extra2

package domain

// GuandanCpuDifficulty CPU の難易度レベル
type GuandanCpuDifficulty int

// 掼蛋の CPU 難易度定数
const (
	// GuandanCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	GuandanCpuDifficultyNormal GuandanCpuDifficulty = iota
)

// GuandanConfig 掼蛋のゲーム設定
type GuandanConfig struct {
	CpuDifficulty GuandanCpuDifficulty `json:"cd"`
}

// DefaultGuandanConfig デフォルト設定を返す
func DefaultGuandanConfig() GuandanConfig {
	return GuandanConfig{CpuDifficulty: GuandanCpuDifficultyNormal}
}

// Validate 設定値のドメインバリデーション
func (c GuandanConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(GuandanCpuDifficultyNormal), int(GuandanCpuDifficultyNormal))
}
