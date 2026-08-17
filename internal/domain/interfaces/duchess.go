//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DuchessGame ダッチェス ゲームインタフェース
type DuchessGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// ChooseBaseRank リザーブ扇から最初の基礎札を選び、開始ランクを決める
	ChooseBaseRank(fanIdx int) error
	// Draw 山札からウェイストへ 1 枚めくる
	Draw() error
	// MoveReserveToFoundation リザーブから基礎札へ移動する
	MoveReserveToFoundation(fanIdx int) error
	// MoveReserveToTableau リザーブからタブローへ移動する
	MoveReserveToTableau(fanIdx, col int) error
	// MoveWasteToFoundation ウェイストから基礎札へ移動する
	MoveWasteToFoundation() error
	// MoveWasteToTableau ウェイストからタブローへ移動する
	MoveWasteToTableau(col int) error
	// MoveTableauToFoundation タブローから基礎札へ移動する
	MoveTableauToFoundation(col int) error
	// MoveTableauToTableau タブロー間で移動する（cardIndex 以降の連番をまとめて）
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// GetHint ヒントを取得する
	GetHint() *domain.DuchessHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.DuchessPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste ウェイストを取得する
	GetWaste() []*domain.Card
	// GetReserve リザーブ扇を取得する
	GetReserve() [domain.DuchessReserveCnt][]*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.DuchessTableauCnt][]*domain.DuchessTableauCard
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.DuchessFoundationCnt][]*domain.Card
	// IsAwaitingBaseRank 開始ランクがまだ選ばれていないか
	IsAwaitingBaseRank() bool
	// GetBaseRank 基礎札の開始ランク（未選択なら 0）
	GetBaseRank() int
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
	// CanAutoComplete いま AutoComplete が 1 枚でも動かせるか (#5557)
	CanAutoComplete() bool
}
