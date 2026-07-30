//go:build !js || !wasm || extra2

package domain

// MushiCpuDifficulty は CPU の難易度レベル。
type MushiCpuDifficulty int

// MushiのCPU難易度定数
const (
	// MushiCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	MushiCpuDifficultyNormal MushiCpuDifficulty = iota
)

// MushiConfig は虫のゲーム設定。
type MushiConfig struct {
	CpuDifficulty MushiCpuDifficulty `json:"cd"`
	// TargetRounds は 1 ゲームの局数。虫は通常 12 局。
	TargetRounds int `json:"tr"`
}

// DefaultMushiConfig はデフォルト設定を返す。
func DefaultMushiConfig() MushiConfig {
	return MushiConfig{
		CpuDifficulty: MushiCpuDifficultyNormal,
		TargetRounds:  MushiMaxRounds,
	}
}

// Validate は設定値のドメインバリデーション。
func (c MushiConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(MushiCpuDifficultyNormal), int(MushiCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("target rounds", c.TargetRounds, 1, MushiMaxRounds)
}
