//go:build !js || !wasm || extra3

package domain

// SkitgubbeConfig はシートグッベのゲーム設定。
type SkitgubbeConfig struct{}

// DefaultSkitgubbeConfig はデフォルト設定を返す。
func DefaultSkitgubbeConfig() SkitgubbeConfig { return SkitgubbeConfig{} }

// Validate は設定値のドメインバリデーション。
func (SkitgubbeConfig) Validate() error { return nil }
