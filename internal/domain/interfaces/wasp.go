package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// WaspGame ワスプゲームインタフェース
type WaspGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Deal ストックから先頭3列にカードを配る
	Deal() error
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// GetHint ヒントを取得する
	GetHint() *domain.WaspHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.WaspPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount ストックの残り枚数を取得する
	GetStockCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
	// GetCompletedSuits 完成スート数を取得する
	GetCompletedSuits() int
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
