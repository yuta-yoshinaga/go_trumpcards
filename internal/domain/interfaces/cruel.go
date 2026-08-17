//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CruelGame クルーエルゲームインタフェース
type CruelGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToTableau タブロー間で最上段の1枚を移動する
	MoveTableauToTableau(fromCol, toCol int) error
	// MoveTableauToFoundation タブロー最上段のカードをファウンデーションへ移動する
	MoveTableauToFoundation(col int) error
	// Shift タブロー残カードを収集し、左から12列×4枚へ配り直す
	Shift() error
	// CanAutoComplete 今オートコンプリートで動かせる札があるか
	CanAutoComplete() bool
	// GetHint ヒントを取得する
	GetHint() *domain.CruelHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CruelPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	// GetFoundation ファウンデーションを取得する
	GetFoundation() [domain.CruelFoundationCnt][]*domain.Card
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
