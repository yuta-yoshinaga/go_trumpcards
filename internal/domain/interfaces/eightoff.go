//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// EightOffGame エイトオフゲームインタフェース
type EightOffGame interface {
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
	GetHint() *domain.EightOffHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.EightOffPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.EightOffTableauCnt][]*domain.Card
	// GetFreeCells フリーセルを取得する
	GetFreeCells() [domain.EightOffCellCnt]*domain.Card
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.EightOffFoundationCnt][]*domain.Card
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
