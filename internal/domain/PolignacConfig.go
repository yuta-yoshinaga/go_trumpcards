//go:build !js || !wasm || extra2

package domain

// PolignacRoundsMin ラウンド数の下限
const PolignacRoundsMin = 1

// PolignacRoundsMax ラウンド数の上限
const PolignacRoundsMax = 12

// PolignacRoundsDefault 既定のラウンド数（4 人が 1 回ずつ配り手を務める）
const PolignacRoundsDefault = 4

// PolignacConfig ポリニャック ゲーム設定
type PolignacConfig struct {
	// Rounds 何ラウンドで打ち切るか
	Rounds int `json:"rd"`
}

// DefaultPolignacConfig デフォルト設定を返す
func DefaultPolignacConfig() PolignacConfig {
	return PolignacConfig{Rounds: PolignacRoundsDefault}
}

// Validate 設定値のドメインバリデーション
func (c PolignacConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, PolignacRoundsMin, PolignacRoundsMax)
}
