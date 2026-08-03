//go:build !js || !wasm || solo

package domain

// LiteratureCpuDifficulty CPU の難易度レベル
type LiteratureCpuDifficulty int

// リテラチャーの CPU 難易度定数
const (
	// LiteratureCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	LiteratureCpuDifficultyNormal LiteratureCpuDifficulty = iota
)

// LiteratureConfig リテラチャーのゲーム設定
type LiteratureConfig struct {
	CpuDifficulty LiteratureCpuDifficulty `json:"cd"`
}

// DefaultLiteratureConfig デフォルト設定を返す
func DefaultLiteratureConfig() LiteratureConfig {
	return LiteratureConfig{CpuDifficulty: LiteratureCpuDifficultyNormal}
}

// Validate 設定値のドメインバリデーション
func (c LiteratureConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(LiteratureCpuDifficultyNormal), int(LiteratureCpuDifficultyNormal))
}
