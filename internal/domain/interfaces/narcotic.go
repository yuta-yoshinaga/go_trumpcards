//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// NarcoticGame ナルコティック（Perpetual Motion）ゲームインタフェース
type NarcoticGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 各列にカードを配る
	Draw() error
	// Remove 露出4枚のランクが揃っているとき、その4枚を捨てる
	Remove() error
	// Move 列の露出札を、同ランクを露出する最も左の列へ重ねる
	Move(col int) error
	// Redeal 山札が尽きたとき、右の列から集めて無シャッフルで配り直す
	Redeal() error
	// GetRedealCount 配り直した回数 (上限なし)
	GetRedealCount() int
	// GiveUp ギブアップする
	GiveUp()
	// GetHint ヒントを取得する
	GetHint() *domain.NarcoticHint
	// Undo 操作を元に戻す
	Undo() error

	// CanUndo 元に戻す操作が可能かを返す
	CanUndo() bool
	// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す
	UndoToEscape() int
	// UndoN n回連続でアンドゥを実行する
	UndoN(n int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.NarcoticPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetDiscardCount 除去済みの枚数を取得する
	GetDiscardCount() int
	// GetDiscardTop 捨て札の一番上（直近に除去した札）を取得する
	GetDiscardTop() *domain.Card
	// GetColumns 場札の列を取得する
	GetColumns() [domain.NarcoticColCnt][]*domain.Card
	// CanRemoveSet 露出4枚のランクが揃っているかを返す
	CanRemoveSet() bool
	// CanMove 列の露出札を重ねられるかを返す
	CanMove(col int) bool
	// MoveTarget 列の露出札の行き先 (-1 = 無し)
	MoveTarget(col int) int
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
