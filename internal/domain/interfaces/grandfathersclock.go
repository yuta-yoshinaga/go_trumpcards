//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GrandfathersClockGame グランドファーザーズ・クロック ゲームインタフェース
type GrandfathersClockGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToFoundation タブローから文字盤へ移動する
	MoveTableauToFoundation(col, fIdx int) error
	// MoveTableauToTableau タブロー間で移動する
	MoveTableauToTableau(fromCol, toCol int) error
	// GetHint ヒントを取得する
	GetHint() *domain.GrandfathersClockHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GrandfathersClockPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetFoundation 文字盤を取得する
	GetFoundation() [domain.GrandfathersClockFoundationCnt][]*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.GrandfathersClockTableauCnt][]*domain.GrandfathersClockTableauCard
	// IsFoundationComplete 文字盤が目標ランクに達しているか
	IsFoundationComplete(fIdx int) bool
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
