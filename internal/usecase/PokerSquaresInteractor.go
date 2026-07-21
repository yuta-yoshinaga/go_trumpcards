//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PokerSquaresInteractorIF はポーカー・スクエアズインタラクターのインタフェース。
type PokerSquaresInteractorIF interface {
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

// PokerSquaresInteractor はポーカー・スクエアズインタラクター。
type PokerSquaresInteractor struct {
	GameBase[interfaces.PokerSquaresGame]
	pp presenter.PokerSquaresPresenter
}

// NewPokerSquaresInteractor はコンストラクタ。
func NewPokerSquaresInteractor(p interfaces.PokerSquaresGame, pp presenter.PokerSquaresPresenter) *PokerSquaresInteractor {
	mustNotNil("PokerSquaresInteractor", map[string]any{"p": p, "pp": pp})
	return &PokerSquaresInteractor{GameBase: GameBase[interfaces.PokerSquaresGame]{Game: p}, pp: pp}
}

// Reset ゲーム初期化
func (pi *PokerSquaresInteractor) Reset() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.Reset)
}

// Place カード配置
func (pi *PokerSquaresInteractor) Place(row, col int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.Place(row, col) })
}

// Undo アンドゥ
func (pi *PokerSquaresInteractor) Undo() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Undo)
}

// GiveUp ギブアップ
func (pi *PokerSquaresInteractor) GiveUp() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.GiveUp)
}

// Hint 現在のカードを置く最善のセルのヒントを出力する
func (pi *PokerSquaresInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *PokerSquaresInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// RestorePokerSquaresInteractor は JSON から PokerSquaresInteractor を復元する。
func RestorePokerSquaresInteractor(data []byte, pp presenter.PokerSquaresPresenter) (*PokerSquaresInteractor, error) {
	return restoreAndBuild[domain.PokerSquares](data, func(g *domain.PokerSquares) *PokerSquaresInteractor {
		return &PokerSquaresInteractor{GameBase: GameBase[interfaces.PokerSquaresGame]{Game: g}, pp: pp}
	})
}
