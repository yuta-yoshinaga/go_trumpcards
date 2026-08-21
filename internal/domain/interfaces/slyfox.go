//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SlyFoxGame スライ・フォックス ゲームインタフェース
type SlyFoxGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// DealToPile 山札から 1 枚めくって、選んだリザーブ枠に置く
	DealToPile(pile int) error
	// DealToFoundation 山札から 1 枚めくって、そのまま基礎札へ送る（周に数えない）
	DealToFoundation(fIdx int) error
	// MoveTableauToFoundation リザーブから基礎札へ移動する
	MoveTableauToFoundation(pile int) error
	// GetHint ヒントを取得する
	GetHint() *domain.SlyFoxHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SlyFoxPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// DealtThisCycle この周でリザーブに置いた枚数を取得する
	DealtThisCycle() int
	// ReserveIsLocked この周を配り切るまでリザーブが閉じているかを取得する
	ReserveIsLocked() bool
	// GetTableau リザーブ枠を取得する
	GetTableau() [domain.SlyFoxTableauCnt][]*domain.Card
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.SlyFoxFoundationCnt][]*domain.Card
	// IsAscendingFoundation その基礎札が A からの昇順かを返す
	IsAscendingFoundation(fIdx int) bool
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
