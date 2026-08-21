//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FourteenOutGame はモンテカルロ・ソリティアのインタフェース。
type FourteenOutGame interface {
	// CountRemovablePairs 盤面に残っている取り除ける組の数
	CountRemovablePairs() int
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Remove 2 列の末尾札を、合計が 14 なら取り除く
	Remove(c1, c2 int) error
	// Undo 直前の操作を取り消す
	Undo() error
	// CanUndo アンドゥ可能かを返す
	CanUndo() bool
	// GiveUp ギブアップする
	GiveUp()
	// Hint 推奨手を返す。playing 以外では nil。
	Hint() *domain.FourteenOutHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FourteenOutPhase
	// GetColumns 各列を取得する (末尾が露出している札)
	GetColumns() [][]*domain.Card
	// GetRemovedCount 取り除いた累計枚数を返す
	GetRemovedCount() int
	// IsComplete ゲームクリア状態かを返す
	IsComplete() bool
	// IsStalemate 手詰まり状態かを返す
	IsStalemate() bool
}
