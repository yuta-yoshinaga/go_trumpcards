package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PyramidInteractorIF ピラミッドインタラクターインタフェース
type PyramidInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// RemovePair ピラミッド上の2枚を合計13で除去
	RemovePair(row1, col1, row2, col2 int) string
	// RemoveKing ピラミッド上のKを単独除去
	RemoveKing(row, col int) string
	// RemoveWithWaste ウェイストとピラミッドのペア除去
	RemoveWithWaste(row, col int) string
	// RemoveWasteKing ウェイストのK除去
	RemoveWasteKing() string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// PyramidInteractor ピラミッドインタラクタークラス
type PyramidInteractor struct {
	GameBase[interfaces.PyramidGame]
	pp presenter.PyramidPresenter
}

// NewPyramidInteractor コンストラクタ
func NewPyramidInteractor(p interfaces.PyramidGame, pp presenter.PyramidPresenter) *PyramidInteractor {
	mustNotNil("PyramidInteractor", map[string]any{"p": p, "pp": pp})
	return &PyramidInteractor{GameBase: GameBase[interfaces.PyramidGame]{Game: p}, pp: pp}
}

// Reset ゲーム初期化
func (pi *PyramidInteractor) Reset() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (pi *PyramidInteractor) Draw() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Draw)
}

// RemovePair ピラミッド上の2枚を合計13で除去
func (pi *PyramidInteractor) RemovePair(row1, col1, row2, col2 int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.RemovePair(row1, col1, row2, col2) })
}

// RemoveKing ピラミッド上のKを単独除去
func (pi *PyramidInteractor) RemoveKing(row, col int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.RemoveKing(row, col) })
}

// RemoveWithWaste ウェイストとピラミッドのペア除去
func (pi *PyramidInteractor) RemoveWithWaste(row, col int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.RemoveWithWaste(row, col) })
}

// RemoveWasteKing ウェイストのK除去
func (pi *PyramidInteractor) RemoveWasteKing() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.RemoveWasteKing)
}

// GiveUp ギブアップ
func (pi *PyramidInteractor) GiveUp() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.GiveUp)
}

// Hint ヒント取得
func (pi *PyramidInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *PyramidInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// Undo アンドゥ
func (pi *PyramidInteractor) Undo() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Undo)
}

// UndoN n回連続アンドゥ
func (pi *PyramidInteractor) UndoN(n int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.UndoN(n) })
}

// RestorePyramidInteractor deserialises JSON into a PyramidInteractor.
func RestorePyramidInteractor(data []byte, pp presenter.PyramidPresenter) (*PyramidInteractor, error) {
	return restoreAndBuild[domain.Pyramid](data, func(g *domain.Pyramid) *PyramidInteractor {
		return &PyramidInteractor{GameBase: GameBase[interfaces.PyramidGame]{Game: g}, pp: pp}
	})
}
