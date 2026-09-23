//go:build !js || !wasm || extra6

package domain

// MushiConfig は虫のゲーム設定。
type MushiConfig struct {
	// TargetRounds は 1 ゲームの局数。虫は通常 12 局。
	TargetRounds int `json:"tr"`
}

// DefaultMushiConfig はデフォルト設定を返す。
func DefaultMushiConfig() MushiConfig {
	return MushiConfig{
		TargetRounds: MushiMaxRounds,
	}
}

// Validate は設定値のドメインバリデーション。
func (c MushiConfig) Validate() error {
	return ValidateRange("target rounds", c.TargetRounds, 1, MushiMaxRounds)
}
