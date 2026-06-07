//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// EightOffInteractorIF エイトオフインタラクターインタフェース
type EightOffInteractorIF interface {
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

// EightOffInteractor エイトオフインタラクタークラス
type EightOffInteractor struct {
	GameBase[interfaces.EightOffGame]
	ep presenter.EightOffPresenter
	solitaireActions[interfaces.EightOffGame]
}

// NewEightOffInteractor コンストラクタ
func NewEightOffInteractor(e interfaces.EightOffGame, ep presenter.EightOffPresenter) *EightOffInteractor {
	mustNotNil("EightOffInteractor", map[string]any{"e": e, "ep": ep})
	return &EightOffInteractor{
		GameBase:         GameBase[interfaces.EightOffGame]{Game: e},
		ep:               ep,
		solitaireActions: newSolitaireActions[interfaces.EightOffGame](e, ep),
	}
}

// Reset ゲーム初期化
func (ei *EightOffInteractor) Reset() string {
	return runAndPresent(ei.Game, ei.ep, ei.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (ei *EightOffInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ei.Game, ei.ep, func() error {
		return ei.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ei *EightOffInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ei.Game, ei.ep, func() error { return ei.Game.MoveTableauToFoundation(col) })
}

// MoveTableauToFreeCell タブローからフリーセルにカードを移動
func (ei *EightOffInteractor) MoveTableauToFreeCell(col, cell int) string {
	return execAndPresent(ei.Game, ei.ep, func() error { return ei.Game.MoveTableauToFreeCell(col, cell) })
}

// MoveFreeCellToTableau フリーセルからタブローにカードを移動
func (ei *EightOffInteractor) MoveFreeCellToTableau(cell, col int) string {
	return execAndPresent(ei.Game, ei.ep, func() error { return ei.Game.MoveFreeCellToTableau(cell, col) })
}

// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
func (ei *EightOffInteractor) MoveFreeCellToFoundation(cell int) string {
	return execAndPresent(ei.Game, ei.ep, func() error { return ei.Game.MoveFreeCellToFoundation(cell) })
}

// Hint ヒント取得
func (ei *EightOffInteractor) Hint() string {
	return ei.ep.HintOutput(ei.Game)
}

// ActionLog 棋譜を出力する
func (ei *EightOffInteractor) ActionLog() string {
	return ei.ep.ActionLogOutput(ei.Game)
}

// RestoreEightOffInteractor deserialises JSON into an EightOffInteractor.
func RestoreEightOffInteractor(data []byte, ep presenter.EightOffPresenter) (*EightOffInteractor, error) {
	return restoreAndBuild[domain.EightOff](data, func(g *domain.EightOff) *EightOffInteractor {
		return &EightOffInteractor{
			GameBase:         GameBase[interfaces.EightOffGame]{Game: g},
			ep:               ep,
			solitaireActions: newSolitaireActions[interfaces.EightOffGame](g, ep),
		}
	})
}
