//go:build !js || !wasm || extra9

package domain

// BidEuchreConfig ビッド・ユーカーのゲーム設定
type BidEuchreConfig struct {
	// AllowNoTrump はノートランプの宣言を許すか。
	AllowNoTrump bool `json:"nt"`
}

// DefaultBidEuchreConfig デフォルト設定を返す
func DefaultBidEuchreConfig() BidEuchreConfig {
	return BidEuchreConfig{
		AllowNoTrump: true,
	}
}

// Validate 設定値のドメインバリデーション
func (c BidEuchreConfig) Validate() error {
	return nil
}
