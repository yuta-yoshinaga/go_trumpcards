//go:build !js || !wasm || extra2

package domain

// SjavsConfig は Sjavs のゲーム設定。
type SjavsConfig struct{}

// DefaultSjavsConfig はデフォルト設定を返す。
func DefaultSjavsConfig() SjavsConfig { return SjavsConfig{} }

// Validate は設定値のドメインバリデーション。
func (SjavsConfig) Validate() error { return nil }
