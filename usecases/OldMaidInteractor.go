package usecases

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"
)

// OldMaidInteractorIF ババ抜きインタラクターインタフェース
type OldMaidInteractorIF interface {
	Reset() string
	Draw() string
}

// OldMaidInteractor ババ抜きインタラクタークラス
type OldMaidInteractor struct {
	om  *entities.OldMaid
	omp presenters.OldMaidPresenter
}

// NewOldMaidInteractor コンストラクタ
func NewOldMaidInteractor(omp presenters.OldMaidPresenter) *OldMaidInteractor {
	players := []*entities.OldMaidPlayer{
		entities.NewOldMaidPlayer(true),  // player 0: 人間
		entities.NewOldMaidPlayer(false), // player 1: CPU
		entities.NewOldMaidPlayer(false), // player 2: CPU
		entities.NewOldMaidPlayer(false), // player 3: CPU
	}
	return &OldMaidInteractor{
		om:  entities.NewOldMaid(entities.NewTrumpCards(1), players),
		omp: omp,
	}
}

// Reset ゲーム初期化
func (oi *OldMaidInteractor) Reset() string {
	oi.om.Reset()
	oi.runCpuTurns()
	return oi.omp.Output(oi.om)
}

// Draw 人間プレイヤーがカードを引く
func (oi *OldMaidInteractor) Draw() string {
	if oi.om.GetGameEndFlag() {
		return oi.omp.Output(oi.om)
	}
	if !oi.om.IsHumanTurn() {
		return oi.omp.Output(oi.om)
	}
	oi.om.PlayerDraw()
	if !oi.om.GetGameEndFlag() {
		oi.runCpuTurns()
	}
	return oi.omp.Output(oi.om)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (oi *OldMaidInteractor) runCpuTurns() {
	for !oi.om.GetGameEndFlag() && !oi.om.IsHumanTurn() {
		oi.om.CpuDraw()
	}
}
