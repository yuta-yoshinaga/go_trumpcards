//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MonteCarloGame はモンテカルロ・ソリティアのインタフェース。
type MonteCarloGame interface {
	// CountRemovablePairs 盤面に残っている取り除ける組の数
	CountRemovablePairs() int
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Remove 隣接する同ランクのペアを取り除く
	Remove(r1, c1, r2, c2 int) error
	// Deal 盤面を詰め直し、空きを山札から補充する
	Deal() error
	// Undo 直前の操作を取り消す
	Undo() error
	// CanUndo アンドゥ可能かを返す
	CanUndo() bool
	// GiveUp ギブアップする
	GiveUp()
	// Hint 推奨手を返す。playing 以外では nil。
	Hint() *domain.MonteCarloHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MonteCarloPhase
	// GetBoard ボードを取得する
	GetBoard() [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card
	// GetStockCount 残りの山札枚数を返す
	GetStockCount() int
	// GetRemovedCount 取り除いた累計枚数を返す
	GetRemovedCount() int
	// GetDealCount Deal を実行した回数を返す
	GetDealCount() int
	// IsComplete ゲームクリア状態かを返す
	IsComplete() bool
	// IsStalemate 手詰まり状態かを返す
	IsStalemate() bool
}
