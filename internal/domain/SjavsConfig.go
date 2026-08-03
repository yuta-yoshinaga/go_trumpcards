//go:build !js || !wasm || extra2

package domain

// SjavsCpuDifficulty は CPU の難易度レベル。
type SjavsCpuDifficulty int

// Sjavs の CPU 難易度定数
const (
	// SjavsCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	SjavsCpuDifficultyNormal SjavsCpuDifficulty = iota
)

// SjavsConfig は Sjavs のゲーム設定。
type SjavsConfig struct {
	CpuDifficulty SjavsCpuDifficulty `json:"cd"`
}

// DefaultSjavsConfig はデフォルト設定を返す。
func DefaultSjavsConfig() SjavsConfig {
	return SjavsConfig{CpuDifficulty: SjavsCpuDifficultyNormal}
}

// Validate は設定値のドメインバリデーション。
func (c SjavsConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(SjavsCpuDifficultyNormal), int(SjavsCpuDifficultyNormal))
}
