package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SevensInteractorIF 7並べインタラクターインタフェース
type SevensInteractorIF interface {
	Reset() string
	ResetWithConfig(tunnelEnabled bool, jokerCount int, cpuStrategy bool, maxPasses int, noJokerFinish bool) string
	Play(idx int) string
	PlayJoker(cardIdx, targetSuit, targetValue int) string
}

// SevensInteractor 7並べインタラクタークラス
type SevensInteractor struct {
	s  interfaces.SevensGame
	sp presenter.SevensPresenter
}

// NewSevensInteractor コンストラクタ
func NewSevensInteractor(s interfaces.SevensGame, sp presenter.SevensPresenter) *SevensInteractor {
	if s == nil {
		panic("SevensInteractor: s must not be nil")
	}
	if sp == nil {
		panic("SevensInteractor: sp must not be nil")
	}
	return &SevensInteractor{
		s:  s,
		sp: sp,
	}
}

// ResetWithConfig 設定付きゲーム初期化
func (si *SevensInteractor) ResetWithConfig(tunnelEnabled bool, jokerCount int, cpuStrategy bool, maxPasses int, noJokerFinish bool) string {
	if jokerCount < 0 {
		jokerCount = 0
	}
	if jokerCount > 2 {
		jokerCount = 2
	}
	if maxPasses < 0 {
		maxPasses = 0
	}
	config := domain.SevensConfig{
		TunnelEnabled: tunnelEnabled,
		JokerCount:    jokerCount,
		CpuStrategy:   cpuStrategy,
		MaxPasses:     maxPasses,
		NoJokerFinish: noJokerFinish,
	}
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	si.s = domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
	si.s.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.s, nil)
}

// Reset ゲーム初期化
func (si *SevensInteractor) Reset() string {
	si.s.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.s, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
// idx: 出すカードのインデックス。-1 の場合はパス。
func (si *SevensInteractor) Play(idx int) string {
	if si.s.GetGameEndFlag() {
		return si.sp.Output(si.s, nil)
	}
	if !si.s.IsHumanTurn() {
		return si.sp.Output(si.s, nil)
	}
	err := si.s.PlayerPlay(idx)
	if err == nil && !si.s.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.sp.Output(si.s, err)
}

// PlayJoker 人間プレイヤーがジョーカーを指定ポジションに出す
func (si *SevensInteractor) PlayJoker(cardIdx, targetSuit, targetValue int) string {
	if si.s.GetGameEndFlag() {
		return si.sp.Output(si.s, nil)
	}
	if !si.s.IsHumanTurn() {
		return si.sp.Output(si.s, nil)
	}
	err := si.s.PlayerPlayJoker(cardIdx, targetSuit, targetValue)
	if err == nil && !si.s.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.sp.Output(si.s, err)
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
