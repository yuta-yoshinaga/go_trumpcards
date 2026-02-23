package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SevensInteractorIF 7並べインタラクターインタフェース
type SevensInteractorIF interface {
	Reset() string
	ResetWithConfig(tunnelEnabled bool, jokerCount int, cpuStrategy bool) string
	Play(idx int) string
	PlayJoker(cardIdx, targetSuit, targetValue int) string
}

// SevensInteractor 7並べインタラクタークラス
type SevensInteractor struct {
	s  *domain.Sevens
	sp presenter.SevensPresenter
}

// NewSevensInteractor コンストラクタ
func NewSevensInteractor(sp presenter.SevensPresenter) *SevensInteractor {
	config := domain.DefaultSevensConfig()
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),  // player 0: 人間
		domain.NewSevensPlayer(false), // player 1: CPU
		domain.NewSevensPlayer(false), // player 2: CPU
		domain.NewSevensPlayer(false), // player 3: CPU
	}
	return &SevensInteractor{
		s:  domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config),
		sp: sp,
	}
}

// ResetWithConfig 設定付きゲーム初期化
func (si *SevensInteractor) ResetWithConfig(tunnelEnabled bool, jokerCount int, cpuStrategy bool) string {
	if jokerCount < 0 {
		jokerCount = 0
	}
	if jokerCount > 2 {
		jokerCount = 2
	}
	config := domain.SevensConfig{
		TunnelEnabled: tunnelEnabled,
		JokerCount:    jokerCount,
		CpuStrategy:   cpuStrategy,
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
	return si.sp.Output(si.s)
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

// PlayJoker 人間プレイヤーがジョーカーを指定ポジションに出す
func (si *SevensInteractor) PlayJoker(cardIdx, targetSuit, targetValue int) string {
	if si.s.GetGameEndFlag() {
		return si.sp.Output(si.s)
	}
	if !si.s.IsHumanTurn() {
		return si.sp.Output(si.s)
	}
	si.s.PlayerPlayJoker(cardIdx, targetSuit, targetValue)
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
