//go:build !js || !wasm || extra2

package domain

// LobaConfig は Loba のゲーム設定。
type LobaConfig struct{}

// DefaultLobaConfig はデフォルト設定を返す。
func DefaultLobaConfig() LobaConfig {
	return LobaConfig{}
}

// Validate は設定値のドメインバリデーション。
func (c LobaConfig) Validate() error {
	return nil
}
