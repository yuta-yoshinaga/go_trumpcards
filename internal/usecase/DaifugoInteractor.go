package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DaifugoInteractorIF 大富豪インタラクターインタフェース
type DaifugoInteractorIF interface {
	Reset() string
	Play(indices []int) string
}

// DaifugoInteractor 大富豪インタラクタークラス
type DaifugoInteractor struct {
	dg  *domain.Daifugo
	dgp presenter.DaifugoPresenter
}

// NewDaifugoInteractor コンストラクタ
func NewDaifugoInteractor(dgp presenter.DaifugoPresenter) *DaifugoInteractor {
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),  // player 0: 人間
		domain.NewDaifugoPlayer(false), // player 1: CPU
		domain.NewDaifugoPlayer(false), // player 2: CPU
		domain.NewDaifugoPlayer(false), // player 3: CPU
	}
	return &DaifugoInteractor{
		dg:  domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config),
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
	if err == nil && !di.dg.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dgp.Output(di.dg, err)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (di *DaifugoInteractor) runCpuTurns() {
	for !di.dg.GetGameEndFlag() && !di.dg.IsHumanTurn() {
		di.dg.CpuPlay()
	}
}
