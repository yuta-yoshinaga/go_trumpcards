//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PyramidGame ピラミッドゲームインタフェース
type PyramidGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札からカードをめくる
	Draw() error
	// RemovePair ピラミッド上の2枚を合計13で除去する
	RemovePair(row1, col1, row2, col2 int) error
	// RemoveKing ピラミッド上のKを単独除去する
	RemoveKing(row, col int) error
	// RemoveWithWaste ウェイストとピラミッドのカードをペアで除去する
	RemoveWithWaste(row, col int) error
	// RemoveWasteKing ウェイストのKを単独除去する
	RemoveWasteKing() error
	// GiveUp ギブアップする
	GiveUp()
	// GetHint ヒントを取得する
	GetHint() *domain.PyramidHint
	// Undo 操作を元に戻す
	Undo() error

	// CanUndo 元に戻す操作が可能かを返す
	CanUndo() bool
	// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す
	UndoToEscape() int
	// UndoN n回連続でアンドゥを実行する
	UndoN(n int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PyramidPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste ウェイストのカード一覧を取得する
	GetWaste() []*domain.Card
	// GetPyramid ピラミッドを取得する
	GetPyramid() [domain.PyramidRowCnt][]*domain.PyramidCard
	// IsExposed カードが露出しているかを返す
	IsExposed(row, col int) bool
	// AllRemoved 全ピラミッドカードが除去されたかを返す
	AllRemoved() bool
	// IsRemovableKing いま単独で除去できるキングかを返す
	IsRemovableKing(row, col int) bool
	// IsWasteKingRemovable ウェイストのトップが単独で除去できるキングかを返す
	IsWasteKingRemovable() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
