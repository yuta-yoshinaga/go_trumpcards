//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DoubleKlondikeGame ダブル・クロンダイク (ガルガンチュア) のゲームインタフェース。
type DoubleKlondikeGame interface {
	BaseGame
	// GetGameEndFlag プレイ中でなければ true。
	GetGameEndFlag() bool
	// Reset ゲームを初期化する。
	Reset()
	// Draw ストックからカードをめくる。
	Draw() error
	// MoveWasteToTableau ウェイストからタブローに移す。
	MoveWasteToTableau(col int) error
	// MoveWasteToFoundation ウェイストからファウンデーションに移す。
	MoveWasteToFoundation() error
	// MoveTableauToTableau タブロー間でカードを移す。
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファウンデーションに移す。
	MoveTableauToFoundation(col int) error
	// GiveUp 投了する。
	GiveUp()
	// AutoComplete 全カード表向きのとき自動でファウンデーションへ出し切る。
	AutoComplete() error
	// Undo 直近の手を取り消す。
	Undo() error
	// CanUndo Undo 可能か。
	CanUndo() bool
	// UndoN n 回 Undo する。
	UndoN(n int) error
	// GetHint ヒントを取得する。
	GetHint() *domain.DoubleKlondikeHint

	// GetPhase 現在のフェーズを取得する。
	GetPhase() domain.DoubleKlondikePhase
	// GetMoveCount 累計手数を取得する。
	GetMoveCount() int
	// GetStockCount ストック残枚数を取得する。
	GetStockCount() int
	// GetWaste ウェイストを取得する。
	GetWaste() []*domain.Card
	// GetTableau タブローを取得する。
	GetTableau() [domain.DoubleKlondikeTableauCnt][]*domain.DoubleKlondikeTableauCard
	// GetFoundation ファウンデーションを取得する。
	GetFoundation() [domain.DoubleKlondikeFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す。
	AllFaceUp() bool
	// IsStalemate 手詰まりかを返す。
	IsStalemate() bool
}
