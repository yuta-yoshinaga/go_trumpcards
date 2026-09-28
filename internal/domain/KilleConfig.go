//go:build !js || !wasm || extra10

package domain

// KilleConfig キッレのゲーム設定
type KilleConfig struct {
	// Stake は 1 ラウンドの掛け金。
	Stake int `json:"st"`
}

// DefaultKilleConfig デフォルト設定を返す
func DefaultKilleConfig() KilleConfig {
	return KilleConfig{
		Stake: 1,
	}
}

// Validate 設定値のドメインバリデーション
func (c KilleConfig) Validate() error {
	return ValidateRange("stake", c.Stake, 1, 100)
}
