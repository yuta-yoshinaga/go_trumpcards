//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// StHelenaGame セント・ヘレナ・ソリティアのゲームインタフェース。
type StHelenaGame interface {
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
	GetHint() *domain.StHelenaHint
	// GetPhase 現在のフェーズを取得する。
	GetPhase() domain.StHelenaPhase
	// GetMoveCount 移動回数を取得する。
	GetMoveCount() int
	// GetRedealsRemaining 残り再配り回数を取得する。
	GetRedealsRemaining() int
	// RestrictionsActive 初回の配りの送り先制限がまだ効いているかを取得する。
	// 上 4 列は K 段だけ、下 4 列は A 段だけに送れる状態。
	RestrictionsActive() bool
	// GetTableau タブローを取得する。
	GetTableau() [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	// GetFoundation ファンデーションを取得する。
	GetFoundation() [domain.StHelenaFoundationCnt][]*domain.Card
	// IsStalemate 手詰まり状態を取得する。
	IsStalemate() bool
}
