package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MonteCarloInteractorIF はモンテカルロ・ソリティアインタラクターのインタフェース。
type MonteCarloInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Remove ペアを取り除く
	Remove(r1, c1, r2, c2 int) string
	// Deal 山札からの補充 (盤面詰め直し)
	Deal() string
	// Undo アンドゥ
	Undo() string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MonteCarloInteractor はモンテカルロ・ソリティアインタラクター。
type MonteCarloInteractor struct {
	GameBase[interfaces.MonteCarloGame]
	mp presenter.MonteCarloPresenter
}

// NewMonteCarloInteractor はコンストラクタ。
func NewMonteCarloInteractor(g interfaces.MonteCarloGame, mp presenter.MonteCarloPresenter) *MonteCarloInteractor {
	mustNotNil("MonteCarloInteractor", map[string]any{"g": g, "mp": mp})
	return &MonteCarloInteractor{GameBase: GameBase[interfaces.MonteCarloGame]{Game: g}, mp: mp}
}

// Reset ゲーム初期化
func (mi *MonteCarloInteractor) Reset() string {
	return runAndPresent(mi.Game, mi.mp, mi.Game.Reset)
}

// Remove ペア取り除き
func (mi *MonteCarloInteractor) Remove(r1, c1, r2, c2 int) string {
	return execAndPresent(mi.Game, mi.mp, func() error { return mi.Game.Remove(r1, c1, r2, c2) })
}

// Deal 山札からの補充
func (mi *MonteCarloInteractor) Deal() string {
	return execAndPresent(mi.Game, mi.mp, mi.Game.Deal)
}

// Undo アンドゥ
func (mi *MonteCarloInteractor) Undo() string {
	return execAndPresent(mi.Game, mi.mp, mi.Game.Undo)
}

// GiveUp ギブアップ
func (mi *MonteCarloInteractor) GiveUp() string {
	return runAndPresent(mi.Game, mi.mp, mi.Game.GiveUp)
}

// Hint ヒント取得
func (mi *MonteCarloInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *MonteCarloInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// RestoreMonteCarloInteractor は JSON から MonteCarloInteractor を復元する。
func RestoreMonteCarloInteractor(data []byte, mp presenter.MonteCarloPresenter) (*MonteCarloInteractor, error) {
	return restoreAndBuild[domain.MonteCarlo](data, func(g *domain.MonteCarlo) *MonteCarloInteractor {
		return &MonteCarloInteractor{GameBase: GameBase[interfaces.MonteCarloGame]{Game: g}, mp: mp}
	})
}
