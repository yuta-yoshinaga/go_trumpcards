package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TriPeaksInteractorIF トリピークスインタラクターインタフェース
type TriPeaksInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// Remove タブローのカードを除去
	Remove(row, col int) string
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

// TriPeaksInteractor トリピークスインタラクタークラス
type TriPeaksInteractor struct {
	GameBase[interfaces.TriPeaksGame]
	tp presenter.TriPeaksPresenter
}

// NewTriPeaksInteractor コンストラクタ
func NewTriPeaksInteractor(t interfaces.TriPeaksGame, tp presenter.TriPeaksPresenter) *TriPeaksInteractor {
	mustNotNil("TriPeaksInteractor", map[string]any{"t": t, "tp": tp})
	return &TriPeaksInteractor{GameBase: GameBase[interfaces.TriPeaksGame]{Game: t}, tp: tp}
}

// Reset ゲーム初期化
func (ti *TriPeaksInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.tp, ti.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (ti *TriPeaksInteractor) Draw() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Draw)
}

// Remove タブローのカードを除去
func (ti *TriPeaksInteractor) Remove(row, col int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.Remove(row, col) })
}

// GiveUp ギブアップ
func (ti *TriPeaksInteractor) GiveUp() string {
	return runAndPresent(ti.Game, ti.tp, ti.Game.GiveUp)
}

// Hint ヒント取得
func (ti *TriPeaksInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TriPeaksInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// Undo アンドゥ
func (ti *TriPeaksInteractor) Undo() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Undo)
}

// UndoN n回連続アンドゥ
func (ti *TriPeaksInteractor) UndoN(n int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.UndoN(n) })
}

// RestoreTriPeaksInteractor deserialises JSON into a TriPeaksInteractor.
func RestoreTriPeaksInteractor(data []byte, tp presenter.TriPeaksPresenter) (*TriPeaksInteractor, error) {
	return restoreAndBuild[domain.TriPeaks](data, func(g *domain.TriPeaks) *TriPeaksInteractor {
		return &TriPeaksInteractor{GameBase: GameBase[interfaces.TriPeaksGame]{Game: g}, tp: tp}
	})
}
