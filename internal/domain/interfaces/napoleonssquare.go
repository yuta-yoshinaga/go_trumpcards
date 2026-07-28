//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// NapoleonsSquareGame ナポレオンズ・スクエア ゲームインタフェース
type NapoleonsSquareGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札からウェイストへ 1 枚めくる
	Draw() error
	// MoveWasteToTableau ウェイストからタブローへ移動する
	MoveWasteToTableau(col int) error
	// MoveWasteToFoundation ウェイストから基礎札へ移動する
	MoveWasteToFoundation() error
	// MoveTableauToTableau タブロー間で移動する（cardIndex 以降の連番をまとめて）
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローから基礎札へ移動する
	MoveTableauToFoundation(col int) error
	// GetHint ヒントを取得する
	GetHint() *domain.NapoleonsSquareHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.NapoleonsSquarePhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste ウェイストを取得する
	GetWaste() []*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.NapoleonsSquareTableauCnt][]*domain.NapoleonsSquareTableauCard
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.NapoleonsSquareFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
