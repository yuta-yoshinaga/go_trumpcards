//go:build !js || !wasm || extra3

package domain

// DesmocheCpuDifficulty CPU の難易度レベル
type DesmocheCpuDifficulty int

// Desmoche の CPU 難易度定数
const (
	// DesmocheCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	DesmocheCpuDifficultyNormal DesmocheCpuDifficulty = iota
)

// DesmocheConfig デスモチェのゲーム設定
type DesmocheConfig struct {
	CpuDifficulty DesmocheCpuDifficulty `json:"cd"`
}

// DefaultDesmocheConfig デフォルト設定を返す
func DefaultDesmocheConfig() DesmocheConfig {
	return DesmocheConfig{CpuDifficulty: DesmocheCpuDifficultyNormal}
}

// Validate 設定値のドメインバリデーション
func (c DesmocheConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(DesmocheCpuDifficultyNormal), int(DesmocheCpuDifficultyNormal))
}
