//go:build !js || !wasm || extra2

package domain

// ChineseTenCpuDifficulty は CPU の難易度レベル。
type ChineseTenCpuDifficulty int

// ChineseTenのCPU難易度定数
const (
	// ChineseTenCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	ChineseTenCpuDifficultyNormal ChineseTenCpuDifficulty = iota
)

// ChineseTenConfig は撿紅點のゲーム設定。
type ChineseTenConfig struct {
	CpuDifficulty ChineseTenCpuDifficulty `json:"cd"`
}

// DefaultChineseTenConfig はデフォルト設定を返す。
func DefaultChineseTenConfig() ChineseTenConfig {
	return ChineseTenConfig{CpuDifficulty: ChineseTenCpuDifficultyNormal}
}

// Validate は設定値のドメインバリデーション。
func (c ChineseTenConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(ChineseTenCpuDifficultyNormal), int(ChineseTenCpuDifficultyNormal))
}
