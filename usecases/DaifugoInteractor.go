package usecases

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"
)

// DaifugoInteractorIF 大富豪インタラクターインタフェース
type DaifugoInteractorIF interface {
	Reset() string
	Play(cardIndices []int) string
	Pass() string
}

// DaifugoInteractor 大富豪インタラクタークラス
type DaifugoInteractor struct {
	d  *entities.Daifugo
	dp presenters.DaifugoPresenter
}

// NewDaifugoInteractor コンストラクタ
func NewDaifugoInteractor(dp presenters.DaifugoPresenter) *DaifugoInteractor {
	players := []*entities.DaifugoPlayer{
		entities.NewDaifugoPlayer(true),  // player 0: 人間
		entities.NewDaifugoPlayer(false), // player 1: CPU
		entities.NewDaifugoPlayer(false), // player 2: CPU
		entities.NewDaifugoPlayer(false), // player 3: CPU
	}
	// ジョーカー2枚で初期化
	return &DaifugoInteractor{
		d:  entities.NewDaifugo(entities.NewTrumpCards(2), players),
		dp: dp,
	}
}

// Reset ゲーム初期化
func (di *DaifugoInteractor) Reset() string {
	di.d.Reset()
	di.runCpuTurns()
	return di.dp.Output(di.d)
}

// Play 人間プレイヤーがカードを出す
func (di *DaifugoInteractor) Play(cardIndices []int) string {
	if di.d.GetGameEndFlag() {
		return di.dp.Output(di.d)
	}
	if !di.d.GetPlayers()[di.d.GetCurrentTurn()].GetIsHuman() {
		return di.dp.Output(di.d)
	}
	di.d.PlayCards(di.d.GetCurrentTurn(), cardIndices)
	if !di.d.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dp.Output(di.d)
}

// Pass 人間プレイヤーがパスする
func (di *DaifugoInteractor) Pass() string {
	if di.d.GetGameEndFlag() {
		return di.dp.Output(di.d)
	}
	if !di.d.GetPlayers()[di.d.GetCurrentTurn()].GetIsHuman() {
		return di.dp.Output(di.d)
	}
	di.d.Pass(di.d.GetCurrentTurn())
	if !di.d.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dp.Output(di.d)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (di *DaifugoInteractor) runCpuTurns() {
	for !di.d.GetGameEndFlag() && !di.d.GetPlayers()[di.d.GetCurrentTurn()].GetIsHuman() {
		di.d.CpuPlay()
	}
}
