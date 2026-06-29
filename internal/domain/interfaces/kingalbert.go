//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KingAlbertGame King Albert ゲームインタフェース
type KingAlbertGame interface {
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
	GetHint() *domain.KingAlbertHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KingAlbertPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	// GetReserve リザーブを取得する
	GetReserve() []*domain.Card
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.KingAlbertFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
