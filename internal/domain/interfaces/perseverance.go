//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PerseveranceGame パーシビアランスゲームインタフェース
type PerseveranceGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションにカードを移動する
	MoveTableauToFoundation(col int) error
	// GetHint ヒントを取得する
	GetHint() *domain.PerseveranceHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PerseverancePhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.PerseveranceTableauCnt][]*domain.PerseveranceTableauCard
	// LegalTargets 列 fromCol の一番下の札を置ける先 (タブロー列 / 組札)
	LegalTargets(fromCol int) ([]int, []int)
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.PerseveranceFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// Redeal 集めて配り直す (最大2回)
	Redeal() error
	// GetRedealsLeft 残りの再配り回数
	GetRedealsLeft() int
}
