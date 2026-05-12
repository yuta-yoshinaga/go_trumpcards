package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CrescentGame クレセント・ソリティアのゲームインタフェース。
type CrescentGame interface {
	SolitaireGame
	// GetGameEndFlag プレイ中でなくなったかを返す。
	GetGameEndFlag() bool
	// Reset ゲームを初期化する。
	Reset()
	// MoveTableauToTableau タブロー間でカードを移動する (最上段 1 枚のみ)。
	MoveTableauToTableau(fromCol, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションへカードを移動する。
	MoveTableauToFoundation(fromCol, foundationIdx int) error
	// Redeal 再配り (シャッフル) を実行する。
	Redeal() error
	// GetHint ヒントを取得する。
	GetHint() *domain.CrescentHint
	// GetPhase 現在のフェーズを取得する。
	GetPhase() domain.CrescentPhase
	// GetMoveCount 移動回数を取得する。
	GetMoveCount() int
	// GetRedealsRemaining 残り再配り回数を取得する。
	GetRedealsRemaining() int
	// GetTableau タブローを取得する。
	GetTableau() [domain.CrescentTableauCnt][]*domain.CrescentTableauCard
	// GetFoundation ファンデーションを取得する。
	GetFoundation() [domain.CrescentFoundationCnt][]*domain.Card
	// IsStalemate 手詰まり状態を取得する。
	IsStalemate() bool
}
