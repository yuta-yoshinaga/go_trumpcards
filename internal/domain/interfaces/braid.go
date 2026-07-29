//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BraidGame ブレイド ゲームインタフェース
type BraidGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から捨て札へ 1 枚めくる（山札が空ならめくり直す）
	Draw() error
	// ChooseDirection 基礎札を積む向きを決める
	ChooseDirection(ascending bool) error
	// MoveBraidToFoundation ブレイドの末尾を基礎札へ送る
	MoveBraidToFoundation() error
	// MoveFieldToFoundation ブレイド札の枠から基礎札へ送る
	MoveFieldToFoundation(idx int) error
	// MoveHelperToFoundation ヘルパー枠から基礎札へ送る
	MoveHelperToFoundation(idx int) error
	// MoveWasteToFoundation 捨て札から基礎札へ送る
	MoveWasteToFoundation() error
	// MoveWasteToHelper 捨て札で空のヘルパー枠を埋める
	MoveWasteToHelper(idx int) error
	// GetHint ヒントを取得する
	GetHint() *domain.BraidHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BraidPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste 捨て札を取得する
	GetWaste() []*domain.Card
	// GetBraid 三つ編みを取得する
	GetBraid() []*domain.Card
	// GetFields ブレイド札の枠を取得する
	GetFields() [domain.BraidFieldCnt]*domain.Card
	// GetHelpers ヘルパー枠を取得する
	GetHelpers() [domain.BraidHelperCnt]*domain.Card
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.BraidFoundationCnt][]*domain.Card
	// GetBaseRank 基礎札の開始ランク
	GetBaseRank() int
	// GetDirection 基礎札を積む向きを取得する
	GetDirection() domain.BraidDirection
	// IsAwaitingDirection 積む向きがまだ選ばれていないか
	IsAwaitingDirection() bool
	// GetPassesUsed 山札を通した回数を取得する
	GetPassesUsed() int
	// CanRedeal もう一度めくり直せるか
	CanRedeal() bool
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
