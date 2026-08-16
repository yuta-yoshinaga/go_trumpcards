//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// NiuNiuGame 闘牛（ニウニウ）ゲームインタフェース
type NiuNiuGame interface {
	BaseGame
	// Reset 新しい局を始める
	Reset()
	// PlaceBet ベットを置いて配り、精算まで進める
	PlaceBet(bet int) error

	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetChips 人間のチップを取得する
	GetChips() int
	// GetMaxMultiplier 最大の配当倍率を取得する
	GetMaxMultiplier() int
	// GetSeats 全席を取得する
	GetSeats() []*domain.NiuNiuSeat
	// GetBankerIdx 親の席番号を取得する
	GetBankerIdx() int
	// GetBankerHand 親の手を取得する
	GetBankerHand() *domain.NiuNiuHand
	// GetLastResult 直近の精算の要約を取得する
	GetLastResult() string
	// GetBankerRankKey は親の格をロケール非依存の識別子で返す。
	// ドメインの NiuNiuRankLabel が返す表示文字列と違い、presenter 側で i18n に通すためのもの。
	GetBankerRankKey() string
	// GetGameEndFlag 局が終わっているか
	GetGameEndFlag() bool
	// GetMultiplier 格の配当倍率を取得する
	GetMultiplier(rank domain.NiuNiuRank) int
}
