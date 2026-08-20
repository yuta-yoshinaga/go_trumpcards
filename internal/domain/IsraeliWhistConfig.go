//go:build !js || !wasm || extra4

package domain

// IsraeliWhistRoundsMin ラウンド数の下限
const IsraeliWhistRoundsMin = 1

// IsraeliWhistRoundsMax ラウンド数の上限
const IsraeliWhistRoundsMax = 20

// IsraeliWhistRoundsDefault 既定のラウンド数
const IsraeliWhistRoundsDefault = 4

// IsraeliWhistConfig イスラエリホイスト ゲーム設定
type IsraeliWhistConfig struct {
	// Rounds ゲーム終了までのラウンド数
	Rounds int `json:"rd"`
}

// DefaultIsraeliWhistConfig デフォルト設定を返す
func DefaultIsraeliWhistConfig() IsraeliWhistConfig {
	return IsraeliWhistConfig{Rounds: IsraeliWhistRoundsDefault}
}

// Validate 設定値のドメインバリデーション
func (c IsraeliWhistConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, IsraeliWhistRoundsMin, IsraeliWhistRoundsMax)
}
