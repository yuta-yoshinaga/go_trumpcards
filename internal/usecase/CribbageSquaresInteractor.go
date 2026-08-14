//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CribbageSquaresInteractorIF はクリベッジ・スクエアズインタラクターのインタフェース。
type CribbageSquaresInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Place カードを配置
	Place(row, col int) string
	// Undo アンドゥ
	Undo() string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint 現在のカードを置く最善のセルのヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CribbageSquaresInteractor はクリベッジ・スクエアズインタラクター。
type CribbageSquaresInteractor struct {
	GameBase[interfaces.CribbageSquaresGame]
	pp presenter.CribbageSquaresPresenter
}

// NewCribbageSquaresInteractor はコンストラクタ。
func NewCribbageSquaresInteractor(p interfaces.CribbageSquaresGame, pp presenter.CribbageSquaresPresenter) *CribbageSquaresInteractor {
	mustNotNil("CribbageSquaresInteractor", map[string]any{"p": p, "pp": pp})
	return &CribbageSquaresInteractor{GameBase: GameBase[interfaces.CribbageSquaresGame]{Game: p}, pp: pp}
}

// Reset ゲーム初期化
func (pi *CribbageSquaresInteractor) Reset() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.Reset)
}

// Place カード配置
func (pi *CribbageSquaresInteractor) Place(row, col int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.Place(row, col) })
}

// Undo アンドゥ
func (pi *CribbageSquaresInteractor) Undo() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Undo)
}

// GiveUp ギブアップ
func (pi *CribbageSquaresInteractor) GiveUp() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.GiveUp)
}

// Hint 現在のカードを置く最善のセルのヒントを出力する
func (pi *CribbageSquaresInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *CribbageSquaresInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// RestoreCribbageSquaresInteractor は JSON から CribbageSquaresInteractor を復元する。
func RestoreCribbageSquaresInteractor(data []byte, pp presenter.CribbageSquaresPresenter) (*CribbageSquaresInteractor, error) {
	return restoreAndBuild[domain.CribbageSquares](data, func(g *domain.CribbageSquares) *CribbageSquaresInteractor {
		return &CribbageSquaresInteractor{GameBase: GameBase[interfaces.CribbageSquaresGame]{Game: g}, pp: pp}
	})
}
