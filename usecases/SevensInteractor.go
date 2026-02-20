package usecases

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"
)

// SevensInteractorIF 7並べインタラクターインタフェース
type SevensInteractorIF interface {
	Reset() string
	Play(idx int) string
}

// SevensInteractor 7並べインタラクタークラス
type SevensInteractor struct {
	s  *entities.Sevens
	sp presenters.SevensPresenter
}

// NewSevensInteractor コンストラクタ
func NewSevensInteractor(sp presenters.SevensPresenter) *SevensInteractor {
	players := []*entities.SevensPlayer{
		entities.NewSevensPlayer(true),  // player 0: 人間
		entities.NewSevensPlayer(false), // player 1: CPU
		entities.NewSevensPlayer(false), // player 2: CPU
		entities.NewSevensPlayer(false), // player 3: CPU
	}
	return &SevensInteractor{
		s:  entities.NewSevens(entities.NewTrumpCards(0), players),
		sp: sp,
	}
}

// Reset ゲーム初期化
func (si *SevensInteractor) Reset() string {
	si.s.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.s)
}

// Play 人間プレイヤーがカードを出す (または パスする)
// idx: 出すカードのインデックス。-1 の場合はパス。
func (si *SevensInteractor) Play(idx int) string {
	if si.s.GetGameEndFlag() {
		return si.sp.Output(si.s)
	}
	if !si.s.IsHumanTurn() {
		return si.sp.Output(si.s)
	}
	si.s.PlayerPlay(idx)
	if !si.s.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.sp.Output(si.s)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
// 人間の手番になった場合でも選択肢がなければ自動処理する
func (si *SevensInteractor) runCpuTurns() {
	for !si.s.GetGameEndFlag() {
		if si.s.IsHumanTurn() {
			// 人間に選択肢がなければ自動処理 (失格)
			if !si.s.HasAnyOption(si.s.GetCurrentTurn()) {
				si.s.AutoHandleNoOption()
			} else {
				break
			}
		} else {
			si.s.CpuPlay()
		}
	}
}
