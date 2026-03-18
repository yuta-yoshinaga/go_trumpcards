package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FreeCellGame フリーセルゲームインタフェース
type FreeCellGame interface {
	// interactor が呼ぶメソッド
	Reset()
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	MoveTableauToFoundation(col int) error
	MoveTableauToFreeCell(col, cell int) error
	MoveFreeCellToTableau(cell, col int) error
	MoveFreeCellToFoundation(cell int) error
	GiveUp()
	GetHint() *domain.FreeCellHint
	AutoComplete() error
	Undo() error

	// state readers
	CanUndo() bool
	GetPhase() domain.FreeCellPhase
	GetMoveCount() int
	GetTableau() [domain.FreeCellTableauCnt][]*domain.Card
	GetFreeCells() [domain.FreeCellCellCnt]*domain.Card
	GetFoundation() [domain.FreeCellFoundationCnt][]*domain.Card
	GetActionLog() []*domain.ActionLogEntry
}
