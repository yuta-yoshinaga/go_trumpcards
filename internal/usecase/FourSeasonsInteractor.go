//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FourSeasonsInteractorIF フォーシーズンズインタラクターインタフェース
type FourSeasonsInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// MoveWasteToTableau ウェイストからタブローに移動
	MoveWasteToTableau(col int) string
	// MoveWasteToFoundation ウェイストからファンデーションに移動
	MoveWasteToFoundation(fIdx int) string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションに移動
	MoveTableauToFoundation(col, fIdx int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜出力
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// FourSeasonsInteractor フォーシーズンズインタラクタークラス
type FourSeasonsInteractor struct {
	GameBase[interfaces.FourSeasonsGame]
	fp presenter.FourSeasonsPresenter
}

// NewFourSeasonsInteractor コンストラクタ
func NewFourSeasonsInteractor(f interfaces.FourSeasonsGame, fp presenter.FourSeasonsPresenter) *FourSeasonsInteractor {
	mustNotNil("FourSeasonsInteractor", map[string]any{"f": f, "fp": fp})
	return &FourSeasonsInteractor{GameBase: GameBase[interfaces.FourSeasonsGame]{Game: f}, fp: fp}
}

// Reset ゲーム初期化
func (fi *FourSeasonsInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.fp, fi.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (fi *FourSeasonsInteractor) Draw() string {
	return execAndPresent(fi.Game, fi.fp, fi.Game.Draw)
}

// MoveWasteToTableau ウェイストからタブローに移動
func (fi *FourSeasonsInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストからファンデーションに移動
func (fi *FourSeasonsInteractor) MoveWasteToFoundation(fIdx int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveWasteToFoundation(fIdx) })
}

// MoveTableauToTableau タブロー間で移動
func (fi *FourSeasonsInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveTableauToTableau(fromCol, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションに移動
func (fi *FourSeasonsInteractor) MoveTableauToFoundation(col, fIdx int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveTableauToFoundation(col, fIdx) })
}

// GiveUp ギブアップ
func (fi *FourSeasonsInteractor) GiveUp() string {
	return runAndPresent(fi.Game, fi.fp, fi.Game.GiveUp)
}

// Hint ヒント取得
func (fi *FourSeasonsInteractor) Hint() string {
	return fi.fp.HintOutput(fi.Game)
}

// AutoComplete オートコンプリート
func (fi *FourSeasonsInteractor) AutoComplete() string {
	return execAndPresent(fi.Game, fi.fp, fi.Game.AutoComplete)
}

// ActionLog 棋譜出力
func (fi *FourSeasonsInteractor) ActionLog() string {
	return fi.fp.ActionLogOutput(fi.Game)
}

// Undo アンドゥ
func (fi *FourSeasonsInteractor) Undo() string {
	return execAndPresent(fi.Game, fi.fp, fi.Game.Undo)
}

// UndoN n回連続アンドゥ
func (fi *FourSeasonsInteractor) UndoN(n int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.UndoN(n) })
}

// RestoreFourSeasonsInteractor deserialises JSON into a FourSeasonsInteractor.
func RestoreFourSeasonsInteractor(data []byte, fp presenter.FourSeasonsPresenter) (*FourSeasonsInteractor, error) {
	return restoreAndBuild[domain.FourSeasons](data, func(g *domain.FourSeasons) *FourSeasonsInteractor {
		return &FourSeasonsInteractor{GameBase: GameBase[interfaces.FourSeasonsGame]{Game: g}, fp: fp}
	})
}
