//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TriPeaksGame トリピークスゲームインタフェース
type TriPeaksGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札からカードをめくる
	Draw() error
	// Remove タブローのカードを除去する
	Remove(row, col int) error
	// GiveUp ギブアップする
	GiveUp()
	// GetHint ヒントを取得する
	GetHint() *domain.TriPeaksHint
	// Undo 操作を元に戻す
	Undo() error

	// CanUndo 元に戻す操作が可能かを返す
	CanUndo() bool
	// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す
	UndoToEscape() int
	// UndoN n回連続でアンドゥを実行する
	UndoN(n int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TriPeaksPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetScore は累計点を取得する。
	GetScore() int
	// GetCombo は途切れていない連続除去の長さを取得する。
	GetCombo() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste ウェイストのカード一覧を取得する
	GetWaste() []*domain.Card
	// GetLayout レイアウトを取得する
	GetLayout() [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard
	// IsExposed カードが露出しているかを返す
	IsExposed(row, col int) bool
	// AllRemoved 全タブローカードが除去されたかを返す
	AllRemoved() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
