//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PontoonGame ポンツーン ゲームインタフェース
type PontoonGame interface {
	BaseGame
	// Reset 新しい局を始める（親は引き継ぐ）
	Reset()
	// PlaceBet ベットを置いて配る
	PlaceBet(bet int) error
	// StartAsBanker 人間が親の局を配る（ベットなし）
	StartAsBanker() error
	// Stick 今の手を打ち止めにする（15 未満では不可）
	Stick() error
	// Twist 表向きに 1 枚引く
	Twist() error
	// Buy 賭け金を上乗せして 1 枚引く（Twist 後は不可）
	Buy(extra int) error
	// Split 同ランク 2 枚を 2 つの手に分ける
	Split() error
	// BankerTwist 人間が親のときに 1 枚引く
	BankerTwist() error
	// BankerStay 人間が親のときに引くのをやめて精算する
	BankerStay() error

	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetChips 人間のチップを取得する
	GetChips() int
	// GetSeats 全席を取得する
	GetSeats() []*domain.PontoonSeat
	// GetBankerIdx 親の席番号を取得する
	GetBankerIdx() int
	// IsHumanBanker 人間が親か
	IsHumanBanker() bool
	// GetBankerHand 親の手を取得する
	GetBankerHand() *domain.PontoonHand
	// GetActiveSeat 手番の席を取得する
	GetActiveSeat() int
	// GetActiveHand 手番の手を取得する
	GetActiveHand() int
	// GetNextBanker 次局の親を取得する（未定なら -1）
	GetNextBanker() int
	// GetLastResult 直近の精算の要約を取得する
	GetLastResult() string
	// GetGameEndFlag 局が終わっているか
	GetGameEndFlag() bool
	// GetHandTotal 手の合計を取得する
	GetHandTotal(cards []*domain.Card) int
	// GetHandRank 手の格を取得する
	GetHandRank(cards []*domain.Card) domain.PontoonRank
	// CanStick Stick を宣言できるか
	CanStick() bool
	// CanTwist Twist できるか
	CanTwist() bool
	// CanBuy Buy できるか
	CanBuy() bool
	// CanSplit 手を割れるか
	CanSplit() bool
}
