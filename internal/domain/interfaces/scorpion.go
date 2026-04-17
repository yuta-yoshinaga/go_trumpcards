package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ScorpionGame スコーピオンゲームインタフェース
type ScorpionGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Deal ストックから先頭3列にカードを配る
	Deal() error
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// GiveUp ギブアップする
	GiveUp()
	// GetHint ヒントを取得する
	GetHint() *domain.ScorpionHint
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
	GetPhase() domain.ScorpionPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount ストックの残り枚数を取得する
	GetStockCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard
	// GetCompletedSuits 完成スート数を取得する
	GetCompletedSuits() int
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
