package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FreeCellInteractorIF フリーセルインタラクターインタフェース
type FreeCellInteractorIF interface {
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
}

// FreeCellInteractor フリーセルインタラクタークラス
type FreeCellInteractor struct {
	f  interfaces.FreeCellGame
	fp presenter.FreeCellPresenter
}

// NewFreeCellInteractor コンストラクタ
func NewFreeCellInteractor(f interfaces.FreeCellGame, fp presenter.FreeCellPresenter) *FreeCellInteractor {
	mustNotNil("FreeCellInteractor", map[string]any{"f": f, "fp": fp})
	return &FreeCellInteractor{f: f, fp: fp}
}

// Reset ゲーム初期化
func (fi *FreeCellInteractor) Reset() string {
	fi.f.Reset()
	return fi.fp.Output(fi.f, nil)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (fi *FreeCellInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	err := fi.f.MoveTableauToTableau(fromCol, cardIndex, toCol)
	return fi.fp.Output(fi.f, err)
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (fi *FreeCellInteractor) MoveTableauToFoundation(col int) string {
	err := fi.f.MoveTableauToFoundation(col)
	return fi.fp.Output(fi.f, err)
}

// MoveTableauToFreeCell タブローからフリーセルにカードを移動
func (fi *FreeCellInteractor) MoveTableauToFreeCell(col, cell int) string {
	err := fi.f.MoveTableauToFreeCell(col, cell)
	return fi.fp.Output(fi.f, err)
}

// MoveFreeCellToTableau フリーセルからタブローにカードを移動
func (fi *FreeCellInteractor) MoveFreeCellToTableau(cell, col int) string {
	err := fi.f.MoveFreeCellToTableau(cell, col)
	return fi.fp.Output(fi.f, err)
}

// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
func (fi *FreeCellInteractor) MoveFreeCellToFoundation(cell int) string {
	err := fi.f.MoveFreeCellToFoundation(cell)
	return fi.fp.Output(fi.f, err)
}

// GiveUp ギブアップ
func (fi *FreeCellInteractor) GiveUp() string {
	fi.f.GiveUp()
	return fi.fp.Output(fi.f, nil)
}

// Hint ヒント取得
func (fi *FreeCellInteractor) Hint() string {
	return fi.fp.HintOutput(fi.f)
}

// AutoComplete オートコンプリート
func (fi *FreeCellInteractor) AutoComplete() string {
	err := fi.f.AutoComplete()
	return fi.fp.Output(fi.f, err)
}

// ActionLog 棋譜を出力する
func (fi *FreeCellInteractor) ActionLog() string {
	return fi.fp.ActionLogOutput(fi.f)
}

// Undo アンドゥ
func (fi *FreeCellInteractor) Undo() string {
	err := fi.f.Undo()
	return fi.fp.Output(fi.f, err)
}
