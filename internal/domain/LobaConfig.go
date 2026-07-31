//go:build !js || !wasm || extra2

package domain

// LobaCpuDifficulty は CPU の難易度レベル。
type LobaCpuDifficulty int

// Loba の CPU 難易度定数
const (
	// LobaCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	LobaCpuDifficultyNormal LobaCpuDifficulty = iota
)

// LobaConfig は Loba のゲーム設定。
type LobaConfig struct {
	CpuDifficulty LobaCpuDifficulty `json:"cd"`
}

// DefaultLobaConfig はデフォルト設定を返す。
func DefaultLobaConfig() LobaConfig {
	return LobaConfig{CpuDifficulty: LobaCpuDifficultyNormal}
}

// Validate は設定値のドメインバリデーション。
func (c LobaConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(LobaCpuDifficultyNormal), int(LobaCpuDifficultyNormal))
}
