//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CurdsAndWheyInteractorIF カーズ・アンド・ホエイのインタラクターインタフェース。
type CurdsAndWheyInteractorIF interface {
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

// CurdsAndWheyInteractor カーズ・アンド・ホエイのインタラクタークラス。
type CurdsAndWheyInteractor struct {
	GameBase[interfaces.CurdsAndWheyGame]
	op presenter.CurdsAndWheyPresenter
}

// NewCurdsAndWheyInteractor コンストラクタ。
func NewCurdsAndWheyInteractor(g interfaces.CurdsAndWheyGame, op presenter.CurdsAndWheyPresenter) *CurdsAndWheyInteractor {
	mustNotNil("CurdsAndWheyInteractor", map[string]any{"g": g, "op": op})
	return &CurdsAndWheyInteractor{GameBase: GameBase[interfaces.CurdsAndWheyGame]{Game: g}, op: op}
}

// Reset ゲーム初期化。
func (si *CurdsAndWheyInteractor) Reset() string {
	return runAndPresent(si.Game, si.op, si.Game.Reset)
}

// MoveSequence 列移動。
func (si *CurdsAndWheyInteractor) MoveSequence(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.op, func() error { return si.Game.MoveSequence(fromCol, cardIndex, toCol) })
}

// GiveUp 投了する。
func (si *CurdsAndWheyInteractor) GiveUp() string {
	return runAndPresent(si.Game, si.op, si.Game.GiveUp)
}

// Hint ヒント取得。
func (si *CurdsAndWheyInteractor) Hint() string {
	return si.op.HintOutput(si.Game)
}

// ActionLog 棋譜出力。
func (si *CurdsAndWheyInteractor) ActionLog() string {
	return si.op.ActionLogOutput(si.Game)
}

// Undo アンドゥ。
func (si *CurdsAndWheyInteractor) Undo() string {
	return execAndPresent(si.Game, si.op, si.Game.Undo)
}

// UndoN n 回アンドゥ。
func (si *CurdsAndWheyInteractor) UndoN(n int) string {
	return execAndPresent(si.Game, si.op, func() error { return si.Game.UndoN(n) })
}

// RestoreCurdsAndWheyInteractor deserialises JSON into a CurdsAndWheyInteractor.
func RestoreCurdsAndWheyInteractor(data []byte, op presenter.CurdsAndWheyPresenter) (*CurdsAndWheyInteractor, error) {
	return restoreAndBuild[domain.CurdsAndWhey](data, func(g *domain.CurdsAndWhey) *CurdsAndWheyInteractor {
		return &CurdsAndWheyInteractor{GameBase: GameBase[interfaces.CurdsAndWheyGame]{Game: g}, op: op}
	})
}
