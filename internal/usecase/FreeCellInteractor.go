//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FreeCellInteractorIF フリーセルインタラクターインタフェース
type FreeCellInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションにカードを移動
	MoveTableauToFoundation(col int) string
	// MoveTableauToFreeCell タブローからフリーセルにカードを移動
	MoveTableauToFreeCell(col, cell int) string
	// MoveFreeCellToTableau フリーセルからタブローにカードを移動
	MoveFreeCellToTableau(cell, col int) string
	// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
	MoveFreeCellToFoundation(cell int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// FreeCellInteractor フリーセルインタラクタークラス
type FreeCellInteractor struct {
	GameBase[interfaces.FreeCellGame]
	fp presenter.FreeCellPresenter
	solitaireActions[interfaces.FreeCellGame]
}

// NewFreeCellInteractor コンストラクタ
func NewFreeCellInteractor(f interfaces.FreeCellGame, fp presenter.FreeCellPresenter) *FreeCellInteractor {
	mustNotNil("FreeCellInteractor", map[string]any{"f": f, "fp": fp})
	return &FreeCellInteractor{
		GameBase:         GameBase[interfaces.FreeCellGame]{Game: f},
		fp:               fp,
		solitaireActions: newSolitaireActions[interfaces.FreeCellGame](f, fp),
	}
}

// Reset ゲーム初期化
func (fi *FreeCellInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.fp, fi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (fi *FreeCellInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (fi *FreeCellInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveTableauToFoundation(col) })
}

// MoveTableauToFreeCell タブローからフリーセルにカードを移動
func (fi *FreeCellInteractor) MoveTableauToFreeCell(col, cell int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveTableauToFreeCell(col, cell) })
}

// MoveFreeCellToTableau フリーセルからタブローにカードを移動
func (fi *FreeCellInteractor) MoveFreeCellToTableau(cell, col int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveFreeCellToTableau(cell, col) })
}

// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
func (fi *FreeCellInteractor) MoveFreeCellToFoundation(cell int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveFreeCellToFoundation(cell) })
}

// Hint ヒント取得
func (fi *FreeCellInteractor) Hint() string {
	return fi.fp.HintOutput(fi.Game)
}

// ActionLog 棋譜を出力する
func (fi *FreeCellInteractor) ActionLog() string {
	return fi.fp.ActionLogOutput(fi.Game)
}

// RestoreFreeCellInteractor deserialises JSON into a FreeCellInteractor.
func RestoreFreeCellInteractor(data []byte, fp presenter.FreeCellPresenter) (*FreeCellInteractor, error) {
	return restoreAndBuild[domain.FreeCell](data, func(g *domain.FreeCell) *FreeCellInteractor {
		return &FreeCellInteractor{
			GameBase:         GameBase[interfaces.FreeCellGame]{Game: g},
			fp:               fp,
			solitaireActions: newSolitaireActions[interfaces.FreeCellGame](g, fp),
		}
	})
}
