//go:build !js || !wasm || classic

package domain

// ReversisRoundsMin ラウンド数の下限
const ReversisRoundsMin = 1

// ReversisRoundsMax ラウンド数の上限
const ReversisRoundsMax = 12

// ReversisRoundsDefault 既定のラウンド数（4 人が 1 回ずつ配り手を務める）
const ReversisRoundsDefault = 4

// ReversisConfig レヴェルシ ゲーム設定
type ReversisConfig struct {
	// Rounds 何ラウンドで打ち切るか
	Rounds int `json:"rd"`
}

// DefaultReversisConfig デフォルト設定を返す
func DefaultReversisConfig() ReversisConfig {
	return ReversisConfig{Rounds: ReversisRoundsDefault}
}

// Validate 設定値のドメインバリデーション
func (c ReversisConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, ReversisRoundsMin, ReversisRoundsMax)
}
