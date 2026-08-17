//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CalculationGame カルキュレーションゲームインタフェース
type CalculationGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// PlayStockToFoundation ストック最上段をファンデーションに置く
	PlayStockToFoundation(fIdx int) error
	// PlayStockToWaste ストック最上段を指定ウェイストパイルに置く
	PlayStockToWaste(wasteIdx int) error
	// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
	PlayWasteToFoundation(wasteIdx, fIdx int) error
	// GetHint ヒントを取得する
	GetHint() *domain.CalculationHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CalculationPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount ストック枚数を取得する
	GetStockCount() int
	// GetStockTop ストック最上段カードを取得する
	GetStockTop() *domain.Card
	// GetWastes ウェイスト一覧を取得する
	GetWastes() [domain.CalculationWasteCnt][]*domain.Card
	// GetFoundations ファンデーション一覧を取得する
	GetFoundations() [domain.CalculationFoundationCnt][]*domain.Card
	// GetNextFoundationRank そのファンデーションに次に置けるランクを取得する (0=置けない)
	GetNextFoundationRank(fIdx int) int
	// GetUpcomingFoundationRanks これから必要になるランクを最大 max 件返す (#5551)
	GetUpcomingFoundationRanks(fIdx, max int) []int
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
