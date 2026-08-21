//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// StalactitesInteractorIF フリーセルインタラクターインタフェース
type StalactitesInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションにカードを移動
	MoveTableauToFoundation(col int) string
	// MoveTableauToStalactites タブローからフリーセルにカードを移動
	MoveTableauToStalactites(col, cell int) string
	// MoveStalactitesToTableau フリーセルからタブローにカードを移動
	MoveStalactitesToTableau(cell, col int) string
	// MoveStalactitesToFoundation フリーセルからファンデーションにカードを移動
	MoveStalactitesToFoundation(cell int) string
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

// StalactitesInteractor フリーセルインタラクタークラス
type StalactitesInteractor struct {
	GameBase[interfaces.StalactitesGame]
	fp presenter.StalactitesPresenter
	solitaireActions[interfaces.StalactitesGame]
}

// NewStalactitesInteractor コンストラクタ
func NewStalactitesInteractor(f interfaces.StalactitesGame, fp presenter.StalactitesPresenter) *StalactitesInteractor {
	mustNotNil("StalactitesInteractor", map[string]any{"f": f, "fp": fp})
	return &StalactitesInteractor{
		GameBase:         GameBase[interfaces.StalactitesGame]{Game: f},
		fp:               fp,
		solitaireActions: newSolitaireActions[interfaces.StalactitesGame](f, fp),
	}
}

// Reset ゲーム初期化
func (fi *StalactitesInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.fp, fi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (fi *StalactitesInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (fi *StalactitesInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveTableauToFoundation(col) })
}

// MoveTableauToStalactites タブローからフリーセルにカードを移動
func (fi *StalactitesInteractor) MoveTableauToStalactites(col, cell int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveTableauToStalactites(col, cell) })
}

// MoveStalactitesToTableau フリーセルからタブローにカードを移動
func (fi *StalactitesInteractor) MoveStalactitesToTableau(cell, col int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveStalactitesToTableau(cell, col) })
}

// MoveStalactitesToFoundation フリーセルからファンデーションにカードを移動
func (fi *StalactitesInteractor) MoveStalactitesToFoundation(cell int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.MoveStalactitesToFoundation(cell) })
}

// Hint ヒント取得
func (fi *StalactitesInteractor) Hint() string {
	return fi.fp.HintOutput(fi.Game)
}

// ActionLog 棋譜を出力する
func (fi *StalactitesInteractor) ActionLog() string {
	return fi.fp.ActionLogOutput(fi.Game)
}

// RestoreStalactitesInteractor deserialises JSON into a StalactitesInteractor.
func RestoreStalactitesInteractor(data []byte, fp presenter.StalactitesPresenter) (*StalactitesInteractor, error) {
	return restoreAndBuild[domain.Stalactites](data, func(g *domain.Stalactites) *StalactitesInteractor {
		return &StalactitesInteractor{
			GameBase:         GameBase[interfaces.StalactitesGame]{Game: g},
			fp:               fp,
			solitaireActions: newSolitaireActions[interfaces.StalactitesGame](g, fp),
		}
	})
}
