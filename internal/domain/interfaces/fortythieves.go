package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FortyThievesGame フォーティシーブスゲームインタフェース
type FortyThievesGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札からカードをめくる
	Draw() error
	// MoveWasteToTableau ウェイストからタブローにカードを移動する
	MoveWasteToTableau(col int) error
	// MoveWasteToFoundation ウェイストからファンデーションにカードを移動する
	MoveWasteToFoundation() error
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションにカードを移動する
	MoveTableauToFoundation(col int) error
	// GiveUp ギブアップする
	GiveUp()
	// GetHint ヒントを取得する
	GetHint() *domain.FortyThievesHint
	// AutoComplete 自動完了を実行する
	AutoComplete() error
	// Undo 操作を元に戻す
	Undo() error

	// CanUndo 元に戻す操作が可能かを返す
	CanUndo() bool
	// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す
	UndoToEscape() int
	// UndoN n回連続でアンドゥを実行する
	UndoN(n int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FortyThievesPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste ウェイストのカード一覧を取得する
	GetWaste() []*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.FortyThievesFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
