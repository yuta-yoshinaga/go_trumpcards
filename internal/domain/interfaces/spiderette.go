//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SpideretteGame スパイダレットゲームインタフェース
type SpideretteGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Deal ストックからタブローに配る
	Deal() error
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// GetHint ヒントを取得する
	GetHint() *domain.SpideretteHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SpiderettePhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetDealsRemaining 「配る」をあと何回押せるかを取得する
	GetDealsRemaining() int
	// GetTableau タブローを取得する
	GetTableau() [domain.SpideretteTableauCnt][]*domain.SpideretteTableauCard
	// GetCompletedSuits 完成スート数を取得する
	GetCompletedSuits() int
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// GetScore 現在のスコアを取得する
	GetScore() int
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
