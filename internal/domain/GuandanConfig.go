//go:build !js || !wasm || extra7

package domain

// GuandanConfig 掼蛋のゲーム設定
type GuandanConfig struct {
}

// DefaultGuandanConfig デフォルト設定を返す
func DefaultGuandanConfig() GuandanConfig {
	return GuandanConfig{}
}

// Validate 設定値のドメインバリデーション
func (c GuandanConfig) Validate() error {
	return nil
}
