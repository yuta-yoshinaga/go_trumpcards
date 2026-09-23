//go:build !js || !wasm || extra3

package domain

// VintConfig ヴィントのゲーム設定
type VintConfig struct{}

// DefaultVintConfig デフォルト設定を返す
func DefaultVintConfig() VintConfig {
	return VintConfig{}
}

// Validate 設定値のドメインバリデーション
func (c VintConfig) Validate() error {
	return nil
}
