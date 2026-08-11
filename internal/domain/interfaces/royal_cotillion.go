//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// RoyalCotillionGame ロイヤルコティヨン ゲームインタフェース
type RoyalCotillionGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() error
	// MoveTableauToFoundation タブローから基礎札へ移動する
	MoveTableauToFoundation(pile int) error
	// MoveReserveToFoundation リザーブの一番上から基礎札へ移動する
	MoveReserveToFoundation(pile int) error
	// MoveWasteToFoundation 捨て札から基礎札へ移動する
	MoveWasteToFoundation() error
	// MoveWasteToTableau 捨て札からタブローへ移動する
	MoveWasteToTableau(pile int) error
	// MoveStockToTableau 山札から空き山へ直接置く
	MoveStockToTableau(pile int) error
	// GetHint ヒントを取得する
	GetHint() *domain.RoyalCotillionHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.RoyalCotillionPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste 捨て札を取得する
	GetWaste() []*domain.Card
	// GetTableau タブロー枠を取得する（1 枠 1 枚、空きは nil）
	GetTableau() [domain.RoyalCotillionTableauCnt]*domain.Card
	// GetReserve リザーブを取得する
	GetReserve() [domain.RoyalCotillionReserveCnt][]*domain.Card
	// IsOddFoundation その基礎札が A 始まりかを返す
	IsOddFoundation(fIdx int) bool
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.RoyalCotillionFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
