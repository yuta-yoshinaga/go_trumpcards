//go:build !js || !wasm || classic

package domain

// ShengJiCpuDifficulty CPU の難易度レベル
type ShengJiCpuDifficulty int

// 升级の CPU 難易度定数
const (
	// ShengJiCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	ShengJiCpuDifficultyNormal ShengJiCpuDifficulty = iota
)

// ShengJiConfig 升级のゲーム設定
type ShengJiConfig struct {
	CpuDifficulty ShengJiCpuDifficulty `json:"cd"`
}

// DefaultShengJiConfig デフォルト設定を返す
func DefaultShengJiConfig() ShengJiConfig {
	return ShengJiConfig{CpuDifficulty: ShengJiCpuDifficultyNormal}
}

// Validate 設定値のドメインバリデーション
func (c ShengJiConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(ShengJiCpuDifficultyNormal), int(ShengJiCpuDifficultyNormal))
}
