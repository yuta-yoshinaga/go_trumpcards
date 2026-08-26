//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MrsMopGame ミセス・モップソリティアゲームインタフェース
type MrsMopGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// ResetWithConfig 指定設定でゲームを初期化する
	ResetWithConfig(cfg domain.MrsMopConfig)
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// GetHint ヒントを取得する
	GetHint() *domain.MrsMopHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MrsMopPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.MrsMopTableauCnt][]*domain.MrsMopTableauCard
	// GetCompletedSuits 完成スート数を取得する
	GetCompletedSuits() int
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// GetScore 現在のスコアを取得する
	GetScore() int
	// GetDifficulty 難易度を取得する
	GetDifficulty() domain.MrsMopDifficulty
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
