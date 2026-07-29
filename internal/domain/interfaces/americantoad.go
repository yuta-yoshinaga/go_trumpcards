//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AmericanToadGame アメリカン・トード ゲームインタフェース
type AmericanToadGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から捨て札へ 1 枚めくる（山札が空ならめくり直す）
	Draw() error
	// MoveReserveToFoundation リザーブから基礎札へ移動する
	MoveReserveToFoundation() error
	// MoveReserveToTableau リザーブからタブローへ移動する
	MoveReserveToTableau(col int) error
	// MoveWasteToFoundation 捨て札から基礎札へ移動する
	MoveWasteToFoundation() error
	// MoveWasteToTableau 捨て札からタブローへ移動する
	MoveWasteToTableau(col int) error
	// MoveTableauToFoundation タブローから基礎札へ移動する
	MoveTableauToFoundation(col int) error
	// MoveTableauToTableau タブロー間で移動する（cardIndex 以降の連番をまとめて）
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// GetHint ヒントを取得する
	GetHint() *domain.AmericanToadHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.AmericanToadPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste 捨て札を取得する
	GetWaste() []*domain.Card
	// GetReserve リザーブを取得する
	GetReserve() []*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.AmericanToadTableauCnt][]*domain.AmericanToadTableauCard
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.AmericanToadFoundationCnt][]*domain.Card
	// GetBaseRank 基礎札の開始ランクを取得する
	GetBaseRank() int
	// GetPassesUsed 山札を通した回数を取得する
	GetPassesUsed() int
	// CanRedeal もう一度めくり直せるかを返す
	CanRedeal() bool
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
