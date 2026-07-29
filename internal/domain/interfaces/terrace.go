//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TerraceGame テラス ゲームインタフェース
type TerraceGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() error
	// MoveReserveToFoundation テラスの一番上を基礎札へ送る
	MoveReserveToFoundation() error
	// MoveWasteToFoundation 捨て札から基礎札へ移動する
	MoveWasteToFoundation() error
	// MoveWasteToTableau 捨て札からタブローへ移動する
	MoveWasteToTableau(pile int) error
	// MoveTableauToFoundation タブローから基礎札へ移動する
	MoveTableauToFoundation(pile int) error
	// MoveTableauToTableau タブロー間で 1 枚移動する
	MoveTableauToTableau(fromPile, toPile int) error
	// GetHint ヒントを取得する
	GetHint() *domain.TerraceHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TerracePhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste 捨て札を取得する
	GetWaste() []*domain.Card
	// GetReserve テラス（リザーブ）を取得する
	GetReserve() []*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.TerraceTableauCnt][]*domain.Card
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.TerraceFoundationCnt][]*domain.Card
	// GetBaseRank 基礎札の開始ランク（未選択なら 0）
	GetBaseRank() int
	// IsAwaitingBaseRank 開始ランクがまだ選ばれていないか
	IsAwaitingBaseRank() bool
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
