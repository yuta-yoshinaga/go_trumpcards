//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CanfieldGame キャンフィールドゲームインタフェース
type CanfieldGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札からカードをめくる
	Draw() error
	// MoveWasteToTableau ウェイストからタブローに移動
	MoveWasteToTableau(col int) error
	// MoveWasteToFoundation ウェイストからファンデーションに移動
	MoveWasteToFoundation() error
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションに移動
	MoveTableauToFoundation(col int) error
	// MoveReserveToTableau リザーブからタブローに移動
	MoveReserveToTableau(col int) error
	// MoveReserveToFoundation リザーブからファンデーションに移動
	MoveReserveToFoundation() error
	// GiveUp ギブアップ
	GiveUp()
	// GetHint ヒント取得
	GetHint() *domain.CanfieldHint
	// AutoComplete 自動完了
	AutoComplete() error
	// Undo アンドゥ
	Undo() error
	// CanUndo アンドゥ可能か
	CanUndo() bool
	// UndoN n回アンドゥ
	UndoN(n int) error

	// GetPhase フェーズ取得
	GetPhase() domain.CanfieldPhase
	// GetMoveCount 移動回数
	GetMoveCount() int
	// GetStockCount ストック残枚数
	GetStockCount() int
	// GetWaste ウェイスト取得
	GetWaste() []*domain.Card
	// GetReserve リザーブ取得
	GetReserve() []*domain.Card
	// GetTableau タブロー取得
	GetTableau() [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
	// GetFoundation ファンデーション取得
	GetFoundation() [domain.CanfieldFoundationCnt][]*domain.Card
	// GetBaseRank ベースランク取得
	GetBaseRank() int
}
