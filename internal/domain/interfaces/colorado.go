//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ColoradoGame コロラド ゲームインタフェース
type ColoradoGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() error
	// MoveTableauToFoundation タブローから基礎札へ移動する
	MoveTableauToFoundation(pile int) error
	// MoveWasteToFoundation 捨て札から基礎札へ移動する
	MoveWasteToFoundation() error
	// MoveWasteToTableau 捨て札からタブローへ移動する
	MoveWasteToTableau(pile int) error
	// MoveStockToTableau 山札から空き山へ直接置く
	MoveStockToTableau(pile int) error
	// GetHint ヒントを取得する
	GetHint() *domain.ColoradoHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ColoradoPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste 捨て札を取得する
	GetWaste() []*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.ColoradoTableauCnt][]*domain.Card
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.ColoradoFoundationCnt][]*domain.Card
	// IsAscendingFoundation その基礎札が A からの昇順かを返す
	IsAscendingFoundation(fIdx int) bool
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
