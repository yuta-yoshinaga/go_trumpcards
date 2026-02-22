package usecases

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"
)

// DaifugoInteractorIF 大富豪インタラクターインタフェース
type DaifugoInteractorIF interface {
	Reset() string
	Play(indices []int) string
}

// DaifugoInteractor 大富豪インタラクタークラス
type DaifugoInteractor struct {
	dg  *entities.Daifugo
	dgp presenters.DaifugoPresenter
}

// NewDaifugoInteractor コンストラクタ
func NewDaifugoInteractor(dgp presenters.DaifugoPresenter) *DaifugoInteractor {
	config := entities.DefaultDaifugoConfig()
	players := []*entities.DaifugoPlayer{
		entities.NewDaifugoPlayer(true),  // player 0: 人間
		entities.NewDaifugoPlayer(false), // player 1: CPU
		entities.NewDaifugoPlayer(false), // player 2: CPU
		entities.NewDaifugoPlayer(false), // player 3: CPU
	}
	return &DaifugoInteractor{
		dg:  entities.NewDaifugo(entities.NewTrumpCards(config.JokerCount), players, config),
		dgp: dgp,
	}
}

// Reset ゲーム初期化
func (di *DaifugoInteractor) Reset() string {
	di.dg.Reset()
	di.runCpuTurns()
	return di.dgp.Output(di.dg)
}

// Play 人間プレイヤーがカードを出す (または パスする)
// indices: 出すカードのインデックス。空の場合はパス。
func (di *DaifugoInteractor) Play(indices []int) string {
	if di.dg.GetGameEndFlag() {
		return di.dgp.Output(di.dg)
	}
	if !di.dg.IsHumanTurn() {
		return di.dgp.Output(di.dg)
	}
	di.dg.PlayerPlay(indices)
	if !di.dg.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dgp.Output(di.dg)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (di *DaifugoInteractor) runCpuTurns() {
	for !di.dg.GetGameEndFlag() && !di.dg.IsHumanTurn() {
		di.dg.CpuPlay()
	}
}
