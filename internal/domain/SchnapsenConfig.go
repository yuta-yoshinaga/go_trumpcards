//go:build !js || !wasm || solo

package domain

// SchnapsenConfig シュナプセン (Schnapsen / Sixty-Six) ゲーム設定
type SchnapsenConfig struct {
}

// DefaultSchnapsenConfig デフォルト設定を返す
func DefaultSchnapsenConfig() SchnapsenConfig {
	return SchnapsenConfig{}
}

// Validate 設定値のドメインバリデーション
func (c SchnapsenConfig) Validate() error {
	return nil
}
