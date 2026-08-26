//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FlowerGardenGame Flower Garden ゲームインタフェース
type FlowerGardenGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションにカードを移動する
	MoveTableauToFoundation(col int) error
	// MoveReserveToTableau リザーブからタブローにカードを移動する
	MoveReserveToTableau(reserveIdx, toCol int) error
	// MoveReserveToFoundation リザーブからファンデーションにカードを移動する
	MoveReserveToFoundation(reserveIdx int) error
	// GetHint ヒントを取得する
	GetHint() *domain.FlowerGardenHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FlowerGardenPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	// GetReserve リザーブを取得する
	GetReserve() []*domain.Card
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.FlowerGardenFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
