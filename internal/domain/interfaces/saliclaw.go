//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SalicLawGame サリカ法典 ゲームインタフェース
type SalicLawGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から 1 枚めくって今の列に置く（K なら次の列を開く）
	Draw() error
	// MoveTableauToFoundation タブローから基礎札へ移動する
	MoveTableauToFoundation(pile int) error
	// MoveTableauToTableau 「K だけの列」へ 1 枚移動する
	MoveTableauToTableau(fromPile, toPile int) error
	// GetHint ヒントを取得する
	GetHint() *domain.SalicLawHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SalicLawPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetQueens 場から抜いたクイーン 8 枚を取得する（飾り。動かせない）
	GetQueens() []*domain.Card
	// GetOpenPiles 土台の K が据わって使えるようになった列の数を取得する
	GetOpenPiles() int
	// GetTableau タブローを取得する
	GetTableau() [domain.SalicLawTableauCnt][]*domain.Card
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.SalicLawFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
