//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GolfInteractorIF ゴルフソリティアインタラクターインタフェース
type GolfInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// Remove タブローのカードを除去
	Remove(col int) string
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
	// ResetNineHole 9ホールスコアをリセット
	ResetNineHole() string
}

// GolfInteractor ゴルフソリティアインタラクタークラス
type GolfInteractor struct {
	GameBase[interfaces.GolfGame]
	gp presenter.GolfPresenter
}

// NewGolfInteractor コンストラクタ
func NewGolfInteractor(g interfaces.GolfGame, gp presenter.GolfPresenter) *GolfInteractor {
	mustNotNil("GolfInteractor", map[string]any{"g": g, "gp": gp})
	return &GolfInteractor{GameBase: GameBase[interfaces.GolfGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (gi *GolfInteractor) Reset() string {
	return runAndPresent(gi.Game, gi.gp, gi.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (gi *GolfInteractor) Draw() string {
	return execAndPresent(gi.Game, gi.gp, gi.Game.Draw)
}

// Remove タブローのカードを除去
func (gi *GolfInteractor) Remove(col int) string {
	return execAndPresent(gi.Game, gi.gp, func() error { return gi.Game.Remove(col) })
}

// GiveUp ギブアップ
func (gi *GolfInteractor) GiveUp() string {
	return runAndPresent(gi.Game, gi.gp, gi.Game.GiveUp)
}

// Hint ヒント取得
func (gi *GolfInteractor) Hint() string {
	return gi.gp.HintOutput(gi.Game)
}

// ActionLog 棋譜を出力する
func (gi *GolfInteractor) ActionLog() string {
	return gi.gp.ActionLogOutput(gi.Game)
}

// Undo アンドゥ
func (gi *GolfInteractor) Undo() string {
	return execAndPresent(gi.Game, gi.gp, gi.Game.Undo)
}

// ResetNineHole 9ホールスコアをリセット
func (gi *GolfInteractor) ResetNineHole() string {
	return gi.gp.ResetNineHole(gi.Game)
}

// UndoN n回連続アンドゥ
func (gi *GolfInteractor) UndoN(n int) string {
	return execAndPresent(gi.Game, gi.gp, func() error { return gi.Game.UndoN(n) })
}

// RestoreGolfInteractor deserialises JSON into a GolfInteractor.
func RestoreGolfInteractor(data []byte, gp presenter.GolfPresenter) (*GolfInteractor, error) {
	return restoreAndBuild[domain.Golf](data, func(g *domain.Golf) *GolfInteractor {
		return &GolfInteractor{GameBase: GameBase[interfaces.GolfGame]{Game: g}, gp: gp}
	})
}
