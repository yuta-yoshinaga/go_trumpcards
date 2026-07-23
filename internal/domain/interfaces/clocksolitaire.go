//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ClockSolitaireGame クロックソリティアゲームインタフェース
type ClockSolitaireGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Step 1ステップ実行する
	Step() error
	// AutoPlay 自動プレイ（全ステップ実行）
	AutoPlay() error
	// Undo 直前のステップを取り消す
	Undo() error
	// CanUndo アンドゥ可能か
	CanUndo() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ClockSolitairePhase
	// GetPiles パイルを取得する
	GetPiles() [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
	// GetFaceUpCount 各パイルの表向き枚数を取得する
	GetFaceUpCount() [domain.ClockSolitairePileCount]int
	// GetStepCount ステップ数を取得する
	GetStepCount() int
	// GetCurrentCard 現在のカードを取得する
	GetCurrentCard() *domain.Card
}
