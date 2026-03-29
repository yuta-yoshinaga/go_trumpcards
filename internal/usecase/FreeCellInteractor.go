package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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
	return runAndPresent(fi.f, fi.fp, fi.f.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (fi *FreeCellInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(fi.f, fi.fp, func() error { return fi.f.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (fi *FreeCellInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(fi.f, fi.fp, func() error { return fi.f.MoveTableauToFoundation(col) })
}

// MoveTableauToFreeCell タブローからフリーセルにカードを移動
func (fi *FreeCellInteractor) MoveTableauToFreeCell(col, cell int) string {
	return execAndPresent(fi.f, fi.fp, func() error { return fi.f.MoveTableauToFreeCell(col, cell) })
}

// MoveFreeCellToTableau フリーセルからタブローにカードを移動
func (fi *FreeCellInteractor) MoveFreeCellToTableau(cell, col int) string {
	return execAndPresent(fi.f, fi.fp, func() error { return fi.f.MoveFreeCellToTableau(cell, col) })
}

// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
func (fi *FreeCellInteractor) MoveFreeCellToFoundation(cell int) string {
	return execAndPresent(fi.f, fi.fp, func() error { return fi.f.MoveFreeCellToFoundation(cell) })
}

// GiveUp ギブアップ
func (fi *FreeCellInteractor) GiveUp() string {
	return runAndPresent(fi.f, fi.fp, fi.f.GiveUp)
}

// Hint ヒント取得
func (fi *FreeCellInteractor) Hint() string {
	return fi.fp.HintOutput(fi.f)
}

// AutoComplete オートコンプリート
func (fi *FreeCellInteractor) AutoComplete() string {
	return execAndPresent(fi.f, fi.fp, fi.f.AutoComplete)
}

// ActionLog 棋譜を出力する
func (fi *FreeCellInteractor) ActionLog() string {
	return fi.fp.ActionLogOutput(fi.f)
}

// Undo アンドゥ
func (fi *FreeCellInteractor) Undo() string {
	return execAndPresent(fi.f, fi.fp, fi.f.Undo)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (fi *FreeCellInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(fi.f)
}

// RestoreFreeCellInteractor deserialises JSON into a FreeCellInteractor.
func RestoreFreeCellInteractor(data []byte, fp presenter.FreeCellPresenter) (*FreeCellInteractor, error) {
	fc, err := restoreGame[domain.FreeCell](data)
	if err != nil {
		return nil, err
	}
	return &FreeCellInteractor{f: fc, fp: fp}, nil
}
