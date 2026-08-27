//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SeahavenTowersGame シーヘイブンタワーズゲームインタフェース
type SeahavenTowersGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// CanAutoComplete はいまオートコンプリートを押せば最後まで通るかを返す。
	CanAutoComplete() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションにカードを移動する
	MoveTableauToFoundation(col int) error
	// MoveTableauToFreeCell タブローからリザーブセルにカードを移動する
	MoveTableauToFreeCell(col, cell int) error
	// MoveFreeCellToTableau リザーブセルからタブローにカードを移動する
	MoveFreeCellToTableau(cell, col int) error
	// MoveFreeCellToFoundation リザーブセルからファンデーションにカードを移動する
	MoveFreeCellToFoundation(cell int) error
	// GetHint ヒントを取得する
	GetHint() *domain.SeahavenTowersHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SeahavenTowersPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.SeahavenTowersTableauCnt][]*domain.Card
	// GetFreeCells リザーブセルを取得する
	GetFreeCells() [domain.SeahavenTowersCellCnt]*domain.Card
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.SeahavenTowersFoundationCnt][]*domain.Card
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
