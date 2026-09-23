//go:build !js || !wasm || extra6

package domain

// ChineseTenConfig は撿紅點のゲーム設定。
type ChineseTenConfig struct {
}

// DefaultChineseTenConfig はデフォルト設定を返す。
func DefaultChineseTenConfig() ChineseTenConfig {
	return ChineseTenConfig{}
}

// Validate は設定値のドメインバリデーション。
func (c ChineseTenConfig) Validate() error {
	return nil
}
