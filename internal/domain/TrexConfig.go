//go:build !js || !wasm || extra3

package domain

// TrexConfig は Trex のゲーム設定。
type TrexConfig struct{}

// DefaultTrexConfig はデフォルト設定を返す。
func DefaultTrexConfig() TrexConfig {
	return TrexConfig{}
}

// Validate は設定値のドメインバリデーション。
func (c TrexConfig) Validate() error {
	return nil
}
