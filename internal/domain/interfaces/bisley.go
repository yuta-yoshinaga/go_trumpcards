//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BisleyGame ビズリー ゲームインタフェース
type BisleyGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, toCol int) error
	// MoveTableauToAceFoundation タブローから昇順（エース側）の基礎札にカードを移動する
	MoveTableauToAceFoundation(col int) error
	// MoveTableauToKingFoundation タブローから降順（キング側）の基礎札にカードを移動する
	MoveTableauToKingFoundation(col int) error
	// GetHint ヒントを取得する
	GetHint() *domain.BisleyHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BisleyPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.BisleyTableauCnt][]*domain.BisleyTableauCard
	// GetAceFoundations 昇順基礎札を取得する
	GetAceFoundations() [domain.BisleyFoundationCnt][]*domain.Card
	// GetKingFoundations 降順基礎札を取得する
	GetKingFoundations() [domain.BisleyFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
