package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OldMaidInteractorIF ババ抜きインタラクターインタフェース
type OldMaidInteractorIF interface {
	Reset() string
	Draw(cardIdx int) string
}

// OldMaidInteractor ババ抜きインタラクタークラス
type OldMaidInteractor struct {
	om  *domain.OldMaid
	omp presenter.OldMaidPresenter
}

// NewOldMaidInteractor コンストラクタ
func NewOldMaidInteractor(omp presenter.OldMaidPresenter) *OldMaidInteractor {
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),  // player 0: 人間
		domain.NewOldMaidPlayer(false), // player 1: CPU
		domain.NewOldMaidPlayer(false), // player 2: CPU
		domain.NewOldMaidPlayer(false), // player 3: CPU
	}
	return &OldMaidInteractor{
		om:  domain.NewOldMaid(domain.NewTrumpCards(1), players),
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
// cardIdx: 引くカードのインデックス。-1 の場合はランダム選択。
func (oi *OldMaidInteractor) Draw(cardIdx int) string {
	if oi.om.GetGameEndFlag() {
		return oi.omp.Output(oi.om)
	}
	if !oi.om.IsHumanTurn() {
		return oi.omp.Output(oi.om)
	}
	oi.om.PlayerDraw(cardIdx)
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
