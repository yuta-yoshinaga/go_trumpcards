package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PenguinInteractorIF ペンギンインタラクターインタフェース
type PenguinInteractorIF interface {
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

// PenguinInteractor ペンギンインタラクタークラス
type PenguinInteractor struct {
	GameBase[interfaces.PenguinGame]
	pp presenter.PenguinPresenter
	solitaireActions[interfaces.PenguinGame]
}

// NewPenguinInteractor コンストラクタ
func NewPenguinInteractor(p interfaces.PenguinGame, pp presenter.PenguinPresenter) *PenguinInteractor {
	mustNotNil("PenguinInteractor", map[string]any{"p": p, "pp": pp})
	return &PenguinInteractor{
		GameBase:         GameBase[interfaces.PenguinGame]{Game: p},
		pp:               pp,
		solitaireActions: newSolitaireActions[interfaces.PenguinGame](p, pp),
	}
}

// Reset ゲーム初期化
func (pi *PenguinInteractor) Reset() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (pi *PenguinInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(pi.Game, pi.pp, func() error {
		return pi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (pi *PenguinInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.MoveTableauToFoundation(col) })
}

// MoveTableauToFreeCell タブローからフリーセルにカードを移動
func (pi *PenguinInteractor) MoveTableauToFreeCell(col, cell int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.MoveTableauToFreeCell(col, cell) })
}

// MoveFreeCellToTableau フリーセルからタブローにカードを移動
func (pi *PenguinInteractor) MoveFreeCellToTableau(cell, col int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.MoveFreeCellToTableau(cell, col) })
}

// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
func (pi *PenguinInteractor) MoveFreeCellToFoundation(cell int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.MoveFreeCellToFoundation(cell) })
}

// Hint ヒント取得
func (pi *PenguinInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *PenguinInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// RestorePenguinInteractor deserialises JSON into a PenguinInteractor.
func RestorePenguinInteractor(data []byte, pp presenter.PenguinPresenter) (*PenguinInteractor, error) {
	return restoreAndBuild[domain.Penguin](data, func(g *domain.Penguin) *PenguinInteractor {
		return &PenguinInteractor{
			GameBase:         GameBase[interfaces.PenguinGame]{Game: g},
			pp:               pp,
			solitaireActions: newSolitaireActions[interfaces.PenguinGame](g, pp),
		}
	})
}
