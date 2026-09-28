//go:build !js || !wasm || extra3

package domain

// NainJauneConfig ル・ナン・ジョーヌのゲーム設定
type NainJauneConfig struct {
	// TargetDeals これだけディールを終えたら決着。
	TargetDeals int `json:"td"`
}

// DefaultNainJauneConfig デフォルト設定を返す
func DefaultNainJauneConfig() NainJauneConfig {
	return NainJauneConfig{
		TargetDeals: 5,
	}
}

// Validate 設定値のドメインバリデーション
func (c NainJauneConfig) Validate() error {
	return ValidateRange("target deals", c.TargetDeals, 1, 100)
}
