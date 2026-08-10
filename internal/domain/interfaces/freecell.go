//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FreeCellGame フリーセルゲームインタフェース
type FreeCellGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
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
	// GetHint ヒントを取得する
	GetHint() *domain.FreeCellHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FreeCellPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.FreeCellTableauCnt][]*domain.Card
	// GetMaxMovableCards いま一度に動かせる最大枚数を取得する
	GetMaxMovableCards() int
	// GetMaxMovableCardsToEmptyColumn 空き列へ動かすときの上限を取得する
	GetMaxMovableCardsToEmptyColumn() int
	// GetFreeCells フリーセルを取得する
	GetFreeCells() [domain.FreeCellCellCnt]*domain.Card
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.FreeCellFoundationCnt][]*domain.Card
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
