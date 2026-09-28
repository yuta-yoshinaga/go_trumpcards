//go:build !js || !wasm || extra4

package domain

// LiteratureConfig リテラチャーのゲーム設定
type LiteratureConfig struct{}

// DefaultLiteratureConfig デフォルト設定を返す
func DefaultLiteratureConfig() LiteratureConfig {
	return LiteratureConfig{}
}

// Validate 設定値のドメインバリデーション
func (c LiteratureConfig) Validate() error {
	return nil
}
