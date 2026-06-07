//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GapsInteractorIF はGapsゲームのインタラクターインタフェース。
type GapsInteractorIF interface {
	// Snapshot はゲーム状態をKV永続化用にシリアライズする。
	Snapshot() ([]byte, error)
	// Reset はゲームを初期化する。
	Reset() string
	// Move はカードを移動する。
	Move(fromRow, fromCol, toRow, toCol int) string
	// Redeal は再配りする。
	Redeal() string
	// Undo はアンドゥする。
	Undo() string
	// UndoN はn回連続でアンドゥする。
	UndoN(n int) string
	// GiveUp はギブアップする。
	GiveUp() string
	// Hint はヒントを取得する。
	Hint() string
	// ActionLog は棋譜を出力する。
	ActionLog() string
}

// GapsInteractor はGapsゲームのインタラクター。
type GapsInteractor struct {
	GameBase[interfaces.GapsGame]
	gp presenter.GapsPresenter
}

// NewGapsInteractor はGapsInteractorを生成する。
func NewGapsInteractor(g interfaces.GapsGame, gp presenter.GapsPresenter) *GapsInteractor {
	mustNotNil("GapsInteractor", map[string]any{"g": g, "gp": gp})
	return &GapsInteractor{GameBase: GameBase[interfaces.GapsGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化。
func (gi *GapsInteractor) Reset() string {
	return runAndPresent(gi.Game, gi.gp, gi.Game.Reset)
}

// Move カード移動。
func (gi *GapsInteractor) Move(fromRow, fromCol, toRow, toCol int) string {
	return execAndPresent(gi.Game, gi.gp, func() error {
		return gi.Game.Move(fromRow, fromCol, toRow, toCol)
	})
}

// Redeal 再配り。
func (gi *GapsInteractor) Redeal() string {
	return execAndPresent(gi.Game, gi.gp, gi.Game.Redeal)
}

// Undo アンドゥ。
func (gi *GapsInteractor) Undo() string {
	return execAndPresent(gi.Game, gi.gp, gi.Game.Undo)
}

// UndoN n回アンドゥ。
func (gi *GapsInteractor) UndoN(n int) string {
	return execAndPresent(gi.Game, gi.gp, func() error { return gi.Game.UndoN(n) })
}

// GiveUp ギブアップ。
func (gi *GapsInteractor) GiveUp() string {
	return runAndPresent(gi.Game, gi.gp, gi.Game.GiveUp)
}

// Hint ヒント取得。
func (gi *GapsInteractor) Hint() string {
	return gi.gp.HintOutput(gi.Game)
}

// ActionLog 棋譜出力。
func (gi *GapsInteractor) ActionLog() string {
	return gi.gp.ActionLogOutput(gi.Game)
}

// RestoreGapsInteractor はJSONからGapsInteractorを復元する。
func RestoreGapsInteractor(data []byte, gp presenter.GapsPresenter) (*GapsInteractor, error) {
	return restoreAndBuild[domain.Gaps](data, func(g *domain.Gaps) *GapsInteractor {
		return &GapsInteractor{GameBase: GameBase[interfaces.GapsGame]{Game: g}, gp: gp}
	})
}
