//go:build !js || !wasm || extra10

package domain

// KaiserConfig カイザーのゲーム設定
type KaiserConfig struct {
	// AllowNoTrump はノートランプ系のビッドを許すか。
	AllowNoTrump bool `json:"nt"`
}

// DefaultKaiserConfig デフォルト設定を返す
func DefaultKaiserConfig() KaiserConfig {
	return KaiserConfig{
		AllowNoTrump: true,
	}
}

// Validate 設定値のドメインバリデーション
func (c KaiserConfig) Validate() error {
	return nil
}
