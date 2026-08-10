//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BlackHoleGame ブラックホール (論理パズル系ソリティア) のゲームインタフェース。
type BlackHoleGame interface {
	BaseGame
	// GetGameEndFlag プレイ中でなければ true。
	GetGameEndFlag() bool
	// Reset ゲームを初期化する。
	Reset()
	// MoveFanToBlackHole 扇のトップをブラックホールへ積む。
	MoveFanToBlackHole(idx int) error
	// GiveUp 投了する。
	GiveUp()
	// Undo 直近の手を取り消す。
	Undo() error
	// CanUndo Undo 可能か。
	CanUndo() bool
	// UndoN n 回 Undo する。
	UndoN(n int) error
	// GetHint ヒントを取得する。
	GetHint() *domain.BlackHoleHint
	// IsStalemate 合法手がない状態か。
	IsStalemate() bool

	// GetPhase 現在のフェーズを取得する。
	GetPhase() domain.BlackHolePhase
	// GetMoveCount 累計手数を取得する。
	GetMoveCount() int
	// GetFans 扇の一覧を取得する。
	GetFans() [][]*domain.Card
	// AcceptableRanks いまブラックホールが受け付けるランク
	AcceptableRanks() []int
	// PlayableFans いま積める扇の番号
	PlayableFans() []int
	// GetBlackHole ブラックホールの積み上げを取得する。
	GetBlackHole() []*domain.Card
}
