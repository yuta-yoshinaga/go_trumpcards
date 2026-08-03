//go:build !js || !wasm || extra2

package domain

// BidEuchreCpuDifficulty CPU の難易度レベル
type BidEuchreCpuDifficulty int

// Bid Euchre の CPU 難易度定数
const (
	// BidEuchreCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	BidEuchreCpuDifficultyNormal BidEuchreCpuDifficulty = iota
)

// BidEuchreConfig ビッド・ユーカーのゲーム設定
type BidEuchreConfig struct {
	CpuDifficulty BidEuchreCpuDifficulty `json:"cd"`
	// AllowNoTrump はノートランプの宣言を許すか。
	AllowNoTrump bool `json:"nt"`
}

// DefaultBidEuchreConfig デフォルト設定を返す
func DefaultBidEuchreConfig() BidEuchreConfig {
	return BidEuchreConfig{
		CpuDifficulty: BidEuchreCpuDifficultyNormal,
		AllowNoTrump:  true,
	}
}

// Validate 設定値のドメインバリデーション
func (c BidEuchreConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(BidEuchreCpuDifficultyNormal), int(BidEuchreCpuDifficultyNormal))
}
