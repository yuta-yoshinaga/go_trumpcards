//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BeleagueredCastleGame Beleaguered Castle ゲームインタフェース
type BeleagueredCastleGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションにカードを移動する
	MoveTableauToFoundation(col int) error
	// GetHint ヒントを取得する
	GetHint() *domain.BeleagueredCastleHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BeleagueredCastlePhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.BeleagueredCastleFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
