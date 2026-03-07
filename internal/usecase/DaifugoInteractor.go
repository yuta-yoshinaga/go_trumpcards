package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DaifugoInteractorIF 大富豪インタラクターインタフェース
type DaifugoInteractorIF interface {
	Reset() string
	Play(indices []int) string
	ResetWithConfig(config domain.DaifugoConfig) string
	Sort(mode domain.DaifugoSortMode) string
}

// DaifugoInteractor 大富豪インタラクタークラス
type DaifugoInteractor struct {
	dg  interfaces.DaifugoGame
	dgp presenter.DaifugoPresenter
}

// NewDaifugoInteractor コンストラクタ
func NewDaifugoInteractor(dg interfaces.DaifugoGame, dgp presenter.DaifugoPresenter) *DaifugoInteractor {
	mustNotNil("DaifugoInteractor", map[string]any{"dg": dg, "dgp": dgp})
	return &DaifugoInteractor{
		dg:  dg,
		dgp: dgp,
	}
}

// Reset ゲーム初期化
func (di *DaifugoInteractor) Reset() string {
	di.dg.Reset()
	di.runCpuTurns()
	return di.dgp.Output(di.dg, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
// indices: 出すカードのインデックス。空の場合はパス。
func (di *DaifugoInteractor) Play(indices []int) string {
	if di.dg.GetGameEndFlag() {
		return di.dgp.Output(di.dg, nil)
	}
	if !di.dg.IsHumanTurn() {
		return di.dgp.Output(di.dg, nil)
	}
	err := di.dg.PlayerPlay(indices)
	if err == nil && !di.dg.GetGameEndFlag() && !di.dg.HasPendingAction() {
		di.runCpuTurns()
	}
	return di.dgp.Output(di.dg, err)
}

// ResetWithConfig 設定を変更してゲームを初期化
func (di *DaifugoInteractor) ResetWithConfig(config domain.DaifugoConfig) string {
	di.dg.SetConfig(config)
	di.dg.Reset()
	di.runCpuTurns()
	return di.dgp.Output(di.dg, nil)
}

// Sort 手札ソートモードを変更
func (di *DaifugoInteractor) Sort(mode domain.DaifugoSortMode) string {
	err := di.dg.SortHumanHand(mode)
	return di.dgp.Output(di.dg, err)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (di *DaifugoInteractor) runCpuTurns() {
	for !di.dg.GetGameEndFlag() && !di.dg.IsHumanTurn() {
		di.dg.CpuPlay()
	}
}
