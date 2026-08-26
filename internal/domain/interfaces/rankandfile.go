//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// RankAndFileGame ランク・アンド・ファイルゲームインタフェース
type RankAndFileGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
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
	GetHint() *domain.RankAndFileHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.RankAndFilePhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste ウェイストのカード一覧を取得する
	GetWaste() []*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.RankAndFileFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
