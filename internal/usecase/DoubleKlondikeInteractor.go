//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DoubleKlondikeInteractorIF ダブル・クロンダイクのインタラクターインタフェース。
type DoubleKlondikeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化。
	Reset() string
	// Draw ストックからカードをめくる。
	Draw() string
	// MoveWasteToTableau ウェイストからタブローに移す。
	MoveWasteToTableau(col int) string
	// MoveWasteToFoundation ウェイストからファウンデーションに移す。
	MoveWasteToFoundation() string
	// MoveTableauToTableau タブロー間でカードを移す。
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファウンデーションに移す。
	MoveTableauToFoundation(col int) string
	// GiveUp 投了する。
	GiveUp() string
	// AutoComplete オートコンプリート。
	AutoComplete() string
	// Hint ヒント取得。
	Hint() string
	// ActionLog 棋譜出力。
	ActionLog() string
	// Undo アンドゥ。
	Undo() string
	// UndoN n 回アンドゥ。
	UndoN(n int) string
}

// DoubleKlondikeInteractor ダブル・クロンダイクのインタラクタークラス。
type DoubleKlondikeInteractor struct {
	GameBase[interfaces.DoubleKlondikeGame]
	op presenter.DoubleKlondikePresenter
}

// NewDoubleKlondikeInteractor コンストラクタ。
func NewDoubleKlondikeInteractor(g interfaces.DoubleKlondikeGame, op presenter.DoubleKlondikePresenter) *DoubleKlondikeInteractor {
	mustNotNil("DoubleKlondikeInteractor", map[string]any{"g": g, "op": op})
	return &DoubleKlondikeInteractor{GameBase: GameBase[interfaces.DoubleKlondikeGame]{Game: g}, op: op}
}

// Reset ゲーム初期化。
func (di *DoubleKlondikeInteractor) Reset() string {
	return runAndPresent(di.Game, di.op, di.Game.Reset)
}

// Draw ストックからカードをめくる。
func (di *DoubleKlondikeInteractor) Draw() string {
	return execAndPresent(di.Game, di.op, di.Game.Draw)
}

// MoveWasteToTableau ウェイストからタブローに移す。
func (di *DoubleKlondikeInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(di.Game, di.op, func() error { return di.Game.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストからファウンデーションに移す。
func (di *DoubleKlondikeInteractor) MoveWasteToFoundation() string {
	return execAndPresent(di.Game, di.op, di.Game.MoveWasteToFoundation)
}

// MoveTableauToTableau タブロー間でカードを移す。
func (di *DoubleKlondikeInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(di.Game, di.op, func() error { return di.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファウンデーションに移す。
func (di *DoubleKlondikeInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(di.Game, di.op, func() error { return di.Game.MoveTableauToFoundation(col) })
}

// GiveUp 投了する。
func (di *DoubleKlondikeInteractor) GiveUp() string {
	return runAndPresent(di.Game, di.op, di.Game.GiveUp)
}

// AutoComplete オートコンプリート。
func (di *DoubleKlondikeInteractor) AutoComplete() string {
	return execAndPresent(di.Game, di.op, di.Game.AutoComplete)
}

// Hint ヒント取得。
func (di *DoubleKlondikeInteractor) Hint() string {
	return di.op.HintOutput(di.Game)
}

// ActionLog 棋譜出力。
func (di *DoubleKlondikeInteractor) ActionLog() string {
	return di.op.ActionLogOutput(di.Game)
}

// Undo アンドゥ。
func (di *DoubleKlondikeInteractor) Undo() string {
	return execAndPresent(di.Game, di.op, di.Game.Undo)
}

// UndoN n 回アンドゥ。
func (di *DoubleKlondikeInteractor) UndoN(n int) string {
	return execAndPresent(di.Game, di.op, func() error { return di.Game.UndoN(n) })
}

// RestoreDoubleKlondikeInteractor deserialises JSON into a DoubleKlondikeInteractor.
func RestoreDoubleKlondikeInteractor(data []byte, op presenter.DoubleKlondikePresenter) (*DoubleKlondikeInteractor, error) {
	return restoreAndBuild[domain.DoubleKlondike](data, func(g *domain.DoubleKlondike) *DoubleKlondikeInteractor {
		return &DoubleKlondikeInteractor{GameBase: GameBase[interfaces.DoubleKlondikeGame]{Game: g}, op: op}
	})
}
