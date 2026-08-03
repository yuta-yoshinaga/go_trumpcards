//go:build !js || !wasm || extra3

package domain

// VintCpuDifficulty CPU の難易度レベル
type VintCpuDifficulty int

// Vint の CPU 難易度定数
const (
	// VintCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	VintCpuDifficultyNormal VintCpuDifficulty = iota
)

// VintConfig ヴィントのゲーム設定
type VintConfig struct {
	CpuDifficulty VintCpuDifficulty `json:"cd"`
}

// DefaultVintConfig デフォルト設定を返す
func DefaultVintConfig() VintConfig {
	return VintConfig{CpuDifficulty: VintCpuDifficultyNormal}
}

// Validate 設定値のドメインバリデーション
func (c VintConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(VintCpuDifficultyNormal), int(VintCpuDifficultyNormal))
}
