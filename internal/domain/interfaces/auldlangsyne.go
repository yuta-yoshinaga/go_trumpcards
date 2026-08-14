//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AuldLangSyneGame オールド・ラング・サインゲームインタフェース
type AuldLangSyneGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Deal ストックから各ウェイストへ1枚ずつ配る
	Deal() error
	// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
	PlayWasteToFoundation(wasteIdx, fIdx int) error
	// GetHint ヒントを取得する
	GetHint() *domain.AuldLangSyneHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.AuldLangSynePhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount ストック枚数を取得する
	GetStockCount() int
	// GetWastes ウェイスト一覧を取得する
	GetWastes() [domain.AuldLangSyneWasteCnt][]*domain.Card
	// GetFoundations ファンデーション一覧を取得する
	GetFoundations() [domain.AuldLangSyneFoundationCnt][]*domain.Card
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
