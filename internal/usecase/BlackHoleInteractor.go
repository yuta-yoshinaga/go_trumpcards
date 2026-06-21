//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BlackHoleInteractorIF ブラックホールのインタラクターインタフェース。
type BlackHoleInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化。
	Reset() string
	// MoveFanToBlackHole 扇のトップをブラックホールへ積む。
	MoveFanToBlackHole(idx int) string
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

// BlackHoleInteractor ブラックホールのインタラクタークラス。
type BlackHoleInteractor struct {
	GameBase[interfaces.BlackHoleGame]
	op presenter.BlackHolePresenter
}

// NewBlackHoleInteractor コンストラクタ。
func NewBlackHoleInteractor(g interfaces.BlackHoleGame, op presenter.BlackHolePresenter) *BlackHoleInteractor {
	mustNotNil("BlackHoleInteractor", map[string]any{"g": g, "op": op})
	return &BlackHoleInteractor{GameBase: GameBase[interfaces.BlackHoleGame]{Game: g}, op: op}
}

// Reset ゲーム初期化。
func (li *BlackHoleInteractor) Reset() string {
	return runAndPresent(li.Game, li.op, li.Game.Reset)
}

// MoveFanToBlackHole 扇のトップをブラックホールへ積む。
func (li *BlackHoleInteractor) MoveFanToBlackHole(idx int) string {
	return execAndPresent(li.Game, li.op, func() error { return li.Game.MoveFanToBlackHole(idx) })
}

// GiveUp 投了する。
func (li *BlackHoleInteractor) GiveUp() string {
	return runAndPresent(li.Game, li.op, li.Game.GiveUp)
}

// Hint ヒント取得。
func (li *BlackHoleInteractor) Hint() string {
	return li.op.HintOutput(li.Game)
}

// ActionLog 棋譜出力。
func (li *BlackHoleInteractor) ActionLog() string {
	return li.op.ActionLogOutput(li.Game)
}

// Undo アンドゥ。
func (li *BlackHoleInteractor) Undo() string {
	return execAndPresent(li.Game, li.op, li.Game.Undo)
}

// UndoN n 回アンドゥ。
func (li *BlackHoleInteractor) UndoN(n int) string {
	return execAndPresent(li.Game, li.op, func() error { return li.Game.UndoN(n) })
}

// RestoreBlackHoleInteractor deserialises JSON into a BlackHoleInteractor.
func RestoreBlackHoleInteractor(data []byte, op presenter.BlackHolePresenter) (*BlackHoleInteractor, error) {
	return restoreAndBuild[domain.BlackHole](data, func(g *domain.BlackHole) *BlackHoleInteractor {
		return &BlackHoleInteractor{GameBase: GameBase[interfaces.BlackHoleGame]{Game: g}, op: op}
	})
}
