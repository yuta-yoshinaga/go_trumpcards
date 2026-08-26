//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FourteenOutInteractorIF はフォーティーンアウト・ソリティアインタラクターのインタフェース。
type FourteenOutInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Remove 2 列の末尾札を、合計が 14 なら取り除く
	Remove(c1, c2 int) string
	// Undo アンドゥ
	Undo() string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// FourteenOutInteractor はフォーティーンアウト・ソリティアインタラクター。
type FourteenOutInteractor struct {
	GameBase[interfaces.FourteenOutGame]
	mp presenter.FourteenOutPresenter
}

// NewFourteenOutInteractor はコンストラクタ。
func NewFourteenOutInteractor(g interfaces.FourteenOutGame, mp presenter.FourteenOutPresenter) *FourteenOutInteractor {
	mustNotNil("FourteenOutInteractor", map[string]any{"g": g, "mp": mp})
	return &FourteenOutInteractor{GameBase: GameBase[interfaces.FourteenOutGame]{Game: g}, mp: mp}
}

// Reset ゲーム初期化
func (mi *FourteenOutInteractor) Reset() string {
	return runAndPresent(mi.Game, mi.mp, mi.Game.Reset)
}

// Remove ペア取り除き
func (mi *FourteenOutInteractor) Remove(c1, c2 int) string {
	return execAndPresent(mi.Game, mi.mp, func() error { return mi.Game.Remove(c1, c2) })
}

// Undo アンドゥ
func (mi *FourteenOutInteractor) Undo() string {
	return execAndPresent(mi.Game, mi.mp, mi.Game.Undo)
}

// GiveUp ギブアップ
func (mi *FourteenOutInteractor) GiveUp() string {
	return runAndPresent(mi.Game, mi.mp, mi.Game.GiveUp)
}

// Hint ヒント取得
func (mi *FourteenOutInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *FourteenOutInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// RestoreFourteenOutInteractor は JSON から FourteenOutInteractor を復元する。
func RestoreFourteenOutInteractor(data []byte, mp presenter.FourteenOutPresenter) (*FourteenOutInteractor, error) {
	return restoreAndBuild[domain.FourteenOut](data, func(g *domain.FourteenOut) *FourteenOutInteractor {
		return &FourteenOutInteractor{GameBase: GameBase[interfaces.FourteenOutGame]{Game: g}, mp: mp}
	})
}
