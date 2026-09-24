//go:build !js || !wasm || extra8

package domain

// ShengJiConfig 升级のゲーム設定
type ShengJiConfig struct{}

// DefaultShengJiConfig デフォルト設定を返す
func DefaultShengJiConfig() ShengJiConfig {
	return ShengJiConfig{}
}

// Validate 設定値のドメインバリデーション
func (c ShengJiConfig) Validate() error {
	return nil
}
