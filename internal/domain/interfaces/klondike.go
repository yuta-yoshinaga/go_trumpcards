//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KlondikeGame クロンダイクゲームインタフェース
type KlondikeGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// ResetWithConfig 指定設定でゲームを初期化する
	ResetWithConfig(cfg domain.KlondikeConfig)
	// Draw 山札からカードをめくる
	Draw() error
	// MoveWasteToTableau ウェイストからタブローにカードを移動する
	MoveWasteToTableau(col int) error
	// MoveWasteToFoundation ウェイストからファンデーションにカードを移動する
	MoveWasteToFoundation() error
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションにカードを移動する
	MoveTableauToFoundation(col int) error
	// GetHint ヒントを取得する
	GetHint() *domain.KlondikeHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KlondikePhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste ウェイストのカード一覧を取得する
	GetWaste() []*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.KlondikeFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// CanAutoComplete いまオートコンプリートが実行できるかを返す
	CanAutoComplete() bool
	// GetDrawCount ドロー枚数設定を取得する
	GetDrawCount() int
	// GetScore 現在のスコアを取得する
	GetScore() int
	// GetScoringMode スコアリングモードを取得する
	GetScoringMode() domain.KlondikeScoringMode
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
