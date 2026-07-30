//go:build !js || !wasm || extra3

package domain

// SkitgubbeCpuDifficulty は CPU の難易度レベル。
type SkitgubbeCpuDifficulty int

// SkitgubbeのCPU難易度定数
const (
	// SkitgubbeCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	SkitgubbeCpuDifficultyNormal SkitgubbeCpuDifficulty = iota
)

// SkitgubbeConfig はシートグッベのゲーム設定。
type SkitgubbeConfig struct {
	CpuDifficulty SkitgubbeCpuDifficulty `json:"cd"`
}

// DefaultSkitgubbeConfig はデフォルト設定を返す。
func DefaultSkitgubbeConfig() SkitgubbeConfig {
	return SkitgubbeConfig{CpuDifficulty: SkitgubbeCpuDifficultyNormal}
}

// Validate は設定値のドメインバリデーション。
func (c SkitgubbeConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(SkitgubbeCpuDifficultyNormal), int(SkitgubbeCpuDifficultyNormal))
}
