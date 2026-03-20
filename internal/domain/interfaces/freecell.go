package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FreeCellGame フリーセルゲームインタフェース
type FreeCellGame interface {
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションにカードを移動する
	MoveTableauToFoundation(col int) error
	// MoveTableauToFreeCell タブローからフリーセルにカードを移動する
	MoveTableauToFreeCell(col, cell int) error
	// MoveFreeCellToTableau フリーセルからタブローにカードを移動する
	MoveFreeCellToTableau(cell, col int) error
	// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動する
	MoveFreeCellToFoundation(cell int) error
	// GiveUp ギブアップする
	GiveUp()
	// GetHint ヒントを取得する
	GetHint() *domain.FreeCellHint
	// AutoComplete 自動完了を実行する
	AutoComplete() error
	// Undo 操作を元に戻す
	Undo() error

	// CanUndo 元に戻す操作が可能かを返す
	CanUndo() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FreeCellPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.FreeCellTableauCnt][]*domain.Card
	// GetFreeCells フリーセルを取得する
	GetFreeCells() [domain.FreeCellCellCnt]*domain.Card
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.FreeCellFoundationCnt][]*domain.Card
	// GetActionLog 棋譜を取得する
	GetActionLog() []*domain.ActionLogEntry
}
