//go:build !js || !wasm || extra3

package domain

// BuraConfig ブラのゲーム設定
type BuraConfig struct {
}

// DefaultBuraConfig デフォルト設定を返す
func DefaultBuraConfig() BuraConfig {
	return BuraConfig{}
}

// Validate 設定値のドメインバリデーション
func (c BuraConfig) Validate() error {
	return nil
}
