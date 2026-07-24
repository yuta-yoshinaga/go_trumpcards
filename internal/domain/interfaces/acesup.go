//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AcesUpGame エースアップ（四つ葉のクローバー）ゲームインタフェース
type AcesUpGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 各列にカードを配る
	Draw() error
	// Remove 列の一番上のカードを除去する
	Remove(col int) error
	// Move 列の一番上のカードを空き列へ移動する
	Move(col int) error
	// GiveUp ギブアップする
	GiveUp()
	// GetHint ヒントを取得する
	GetHint() *domain.AcesUpHint
	// Undo 操作を元に戻す
	Undo() error

	// CanUndo 元に戻す操作が可能かを返す
	CanUndo() bool
	// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す
	UndoToEscape() int
	// UndoN n回連続でアンドゥを実行する
	UndoN(n int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.AcesUpPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetDiscardCount 除去済みの枚数を取得する
	GetDiscardCount() int
	// GetDiscardTop 捨て札の一番上（直近に除去した札）を取得する
	GetDiscardTop() *domain.Card
	// GetColumns 場札の列を取得する
	GetColumns() [domain.AcesUpColCnt][]*domain.Card
	// CanRemove 指定列の一番上のカードが除去可能かを返す
	CanRemove(col int) bool
	// CanMove 指定列の一番上のカードが空き列へ移動可能かを返す
	CanMove(col int) bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
