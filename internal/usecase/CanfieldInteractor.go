package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CanfieldInteractorIF キャンフィールドインタラクターインタフェース
type CanfieldInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// MoveWasteToTableau ウェイストからタブローに移動
	MoveWasteToTableau(col int) string
	// MoveWasteToFoundation ウェイストからファンデーションに移動
	MoveWasteToFoundation() string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションに移動
	MoveTableauToFoundation(col int) string
	// MoveReserveToTableau リザーブからタブローに移動
	MoveReserveToTableau(col int) string
	// MoveReserveToFoundation リザーブからファンデーションに移動
	MoveReserveToFoundation() string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜出力
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// CanfieldInteractor キャンフィールドインタラクタークラス
type CanfieldInteractor struct {
	GameBase[interfaces.CanfieldGame]
	cp presenter.CanfieldPresenter
}

// NewCanfieldInteractor コンストラクタ
func NewCanfieldInteractor(c interfaces.CanfieldGame, cp presenter.CanfieldPresenter) *CanfieldInteractor {
	mustNotNil("CanfieldInteractor", map[string]any{"c": c, "cp": cp})
	return &CanfieldInteractor{GameBase: GameBase[interfaces.CanfieldGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *CanfieldInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (ci *CanfieldInteractor) Draw() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Draw)
}

// MoveWasteToTableau ウェイストからタブローに移動
func (ci *CanfieldInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストからファンデーションに移動
func (ci *CanfieldInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.MoveWasteToFoundation)
}

// MoveTableauToTableau タブロー間で移動
func (ci *CanfieldInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションに移動
func (ci *CanfieldInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(col) })
}

// MoveReserveToTableau リザーブからタブローに移動
func (ci *CanfieldInteractor) MoveReserveToTableau(col int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveReserveToTableau(col) })
}

// MoveReserveToFoundation リザーブからファンデーションに移動
func (ci *CanfieldInteractor) MoveReserveToFoundation() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.MoveReserveToFoundation)
}

// GiveUp ギブアップ
func (ci *CanfieldInteractor) GiveUp() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.GiveUp)
}

// Hint ヒント取得
func (ci *CanfieldInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// AutoComplete オートコンプリート
func (ci *CanfieldInteractor) AutoComplete() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.AutoComplete)
}

// ActionLog 棋譜出力
func (ci *CanfieldInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// Undo アンドゥ
func (ci *CanfieldInteractor) Undo() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Undo)
}

// UndoN n回連続アンドゥ
func (ci *CanfieldInteractor) UndoN(n int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.UndoN(n) })
}

// RestoreCanfieldInteractor deserialises JSON into a CanfieldInteractor.
func RestoreCanfieldInteractor(data []byte, cp presenter.CanfieldPresenter) (*CanfieldInteractor, error) {
	return restoreAndBuild[domain.Canfield](data, func(g *domain.Canfield) *CanfieldInteractor {
		return &CanfieldInteractor{GameBase: GameBase[interfaces.CanfieldGame]{Game: g}, cp: cp}
	})
}
