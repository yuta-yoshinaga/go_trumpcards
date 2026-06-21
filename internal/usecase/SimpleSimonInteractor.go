//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SimpleSimonInteractorIF シンプル・サイモンのインタラクターインタフェース。
type SimpleSimonInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化。
	Reset() string
	// MoveSequence 列 fromCol の cardIndex 以降を列 toCol へ移す。
	MoveSequence(fromCol, cardIndex, toCol int) string
	// GiveUp 投了する。
	GiveUp() string
	// Hint ヒント取得。
	Hint() string
	// ActionLog 棋譜出力。
	ActionLog() string
	// Undo アンドゥ。
	Undo() string
	// UndoN n 回アンドゥ。
	UndoN(n int) string
}

// SimpleSimonInteractor シンプル・サイモンのインタラクタークラス。
type SimpleSimonInteractor struct {
	GameBase[interfaces.SimpleSimonGame]
	op presenter.SimpleSimonPresenter
}

// NewSimpleSimonInteractor コンストラクタ。
func NewSimpleSimonInteractor(g interfaces.SimpleSimonGame, op presenter.SimpleSimonPresenter) *SimpleSimonInteractor {
	mustNotNil("SimpleSimonInteractor", map[string]any{"g": g, "op": op})
	return &SimpleSimonInteractor{GameBase: GameBase[interfaces.SimpleSimonGame]{Game: g}, op: op}
}

// Reset ゲーム初期化。
func (si *SimpleSimonInteractor) Reset() string {
	return runAndPresent(si.Game, si.op, si.Game.Reset)
}

// MoveSequence 列移動。
func (si *SimpleSimonInteractor) MoveSequence(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.op, func() error { return si.Game.MoveSequence(fromCol, cardIndex, toCol) })
}

// GiveUp 投了する。
func (si *SimpleSimonInteractor) GiveUp() string {
	return runAndPresent(si.Game, si.op, si.Game.GiveUp)
}

// Hint ヒント取得。
func (si *SimpleSimonInteractor) Hint() string {
	return si.op.HintOutput(si.Game)
}

// ActionLog 棋譜出力。
func (si *SimpleSimonInteractor) ActionLog() string {
	return si.op.ActionLogOutput(si.Game)
}

// Undo アンドゥ。
func (si *SimpleSimonInteractor) Undo() string {
	return execAndPresent(si.Game, si.op, si.Game.Undo)
}

// UndoN n 回アンドゥ。
func (si *SimpleSimonInteractor) UndoN(n int) string {
	return execAndPresent(si.Game, si.op, func() error { return si.Game.UndoN(n) })
}

// RestoreSimpleSimonInteractor deserialises JSON into a SimpleSimonInteractor.
func RestoreSimpleSimonInteractor(data []byte, op presenter.SimpleSimonPresenter) (*SimpleSimonInteractor, error) {
	return restoreAndBuild[domain.SimpleSimon](data, func(g *domain.SimpleSimon) *SimpleSimonInteractor {
		return &SimpleSimonInteractor{GameBase: GameBase[interfaces.SimpleSimonGame]{Game: g}, op: op}
	})
}
