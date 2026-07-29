//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SetteEMezzoGame セッテ・エ・メッツォ ゲームインタフェース
type SetteEMezzoGame interface {
	BaseGame
	// Reset 新しい局を始める（親は引き継ぐ）
	Reset()
	// PlaceBet ベットを置いて配る
	PlaceBet(bet int) error
	// StartAsBanker 人間が親の局を配る（ベットなし）
	StartAsBanker() error
	// Hit 1 枚引く
	Hit() error
	// Stand 引き止める
	Stand() error
	// SetMattaValue マッタに割り当てる値を選ぶ（半点単位）
	SetMattaValue(halves int) error
	// BankerHit 人間が親のときに 1 枚引く
	BankerHit() error
	// BankerStand 人間が親のときに引き止めて精算する
	BankerStand() error

	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetChips 人間のチップを取得する
	GetChips() int
	// GetSeats 全席を取得する
	GetSeats() []*domain.SetteEMezzoSeat
	// GetBankerIdx 親の席番号を取得する
	GetBankerIdx() int
	// IsHumanBanker 人間が親か
	IsHumanBanker() bool
	// GetBankerHand 親の手を取得する
	GetBankerHand() *domain.SetteEMezzoHand
	// GetActiveSeat 手番の席を取得する
	GetActiveSeat() int
	// GetNextBanker 次局の親を取得する（未定なら -1）
	GetNextBanker() int
	// GetLastResult 直近の精算の要約を取得する
	GetLastResult() string
	// GetGameEndFlag 局が終わっているか
	GetGameEndFlag() bool
	// GetHandHalves 手の合計を半点単位で取得する
	GetHandHalves(h *domain.SetteEMezzoHand) int
	// FormatHalves 半点単位を表示文字列にする
	FormatHalves(halves int) string
	// CanHit 引けるか
	CanHit() bool
	// CanStand 止められるか
	CanStand() bool
	// CanSetMatta マッタの値を選べるか
	CanSetMatta() bool
}
