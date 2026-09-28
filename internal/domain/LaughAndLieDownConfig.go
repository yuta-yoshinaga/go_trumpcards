//go:build !js || !wasm || extra2

package domain

// LaughAndLieDownConfig は Laugh and Lie Down のゲーム設定。
type LaughAndLieDownConfig struct {
}

// DefaultLaughAndLieDownConfig はデフォルト設定を返す。
func DefaultLaughAndLieDownConfig() LaughAndLieDownConfig {
	return LaughAndLieDownConfig{}
}

// Validate は設定値のドメインバリデーション。
func (c LaughAndLieDownConfig) Validate() error {
	return nil
}
