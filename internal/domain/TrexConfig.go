//go:build !js || !wasm || extra3

package domain

// TrexCpuDifficulty は CPU の難易度レベル。
type TrexCpuDifficulty int

// Trex の CPU 難易度定数
const (
	// TrexCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	TrexCpuDifficultyNormal TrexCpuDifficulty = iota
)

// TrexConfig は Trex のゲーム設定。
type TrexConfig struct {
	CpuDifficulty TrexCpuDifficulty `json:"cd"`
}

// DefaultTrexConfig はデフォルト設定を返す。
func DefaultTrexConfig() TrexConfig {
	return TrexConfig{CpuDifficulty: TrexCpuDifficultyNormal}
}

// Validate は設定値のドメインバリデーション。
func (c TrexConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(TrexCpuDifficultyNormal), int(TrexCpuDifficultyNormal))
}
