//go:build !js || !wasm || extra3

package domain

// PopeJoanConfig ポープ・ジョーンのゲーム設定
type PopeJoanConfig struct {
	// TargetDeals これだけディールを終えたら決着。
	TargetDeals int `json:"td"`
}

// DefaultPopeJoanConfig デフォルト設定を返す
func DefaultPopeJoanConfig() PopeJoanConfig {
	return PopeJoanConfig{
		TargetDeals: 5,
	}
}

// Validate 設定値のドメインバリデーション
func (c PopeJoanConfig) Validate() error {
	return ValidateRange("target deals", c.TargetDeals, 1, 100)
}
