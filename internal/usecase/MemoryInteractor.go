package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MemoryInteractorIF 神経衰弱インタラクターインタフェース
type MemoryInteractorIF interface {
	Reset() string
	ResetWithConfig(cfg domain.MemoryConfig) string
	Flip(pos int) string
	Next() string
	GetConfig() domain.MemoryConfig
	ActionLog() string
}

// MemoryInteractor 神経衰弱インタラクタークラス
type MemoryInteractor struct {
	m  interfaces.MemoryGame
	mp presenter.MemoryPresenter
}

// NewMemoryInteractor コンストラクタ
func NewMemoryInteractor(m interfaces.MemoryGame, mp presenter.MemoryPresenter) *MemoryInteractor {
	mustNotNil("MemoryInteractor", map[string]any{"m": m, "mp": mp})
	return &MemoryInteractor{m: m, mp: mp}
}

// Reset ゲーム初期化
func (mi *MemoryInteractor) Reset() string {
	mi.m.Reset()
	return mi.mp.Output(mi.m, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *MemoryInteractor) ResetWithConfig(cfg domain.MemoryConfig) string {
	mi.m.SetConfig(cfg)
	return mi.Reset()
}

// Flip カードをめくる
func (mi *MemoryInteractor) Flip(pos int) string {
	if mi.m.GetGameEndFlag() {
		return mi.mp.Output(mi.m, nil)
	}
	if !mi.m.IsHumanTurn() {
		return mi.mp.Output(mi.m, nil)
	}
	err := mi.m.PlayerFlip(pos)
	if err != nil {
		return mi.mp.Output(mi.m, err)
	}
	return mi.mp.Output(mi.m, nil)
}

// Next 結果を解決し、CPU ターンを実行する
func (mi *MemoryInteractor) Next() string {
	if mi.m.GetGameEndFlag() {
		return mi.mp.Output(mi.m, nil)
	}
	mi.m.ResolveFlip()
	mi.runCpuTurns()
	return mi.mp.Output(mi.m, nil)
}

// GetConfig 現在の設定を取得
func (mi *MemoryInteractor) GetConfig() domain.MemoryConfig {
	return mi.m.GetConfig()
}

// ActionLog 棋譜を出力する
func (mi *MemoryInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.m)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (mi *MemoryInteractor) runCpuTurns() {
	for !mi.m.GetGameEndFlag() {
		if mi.m.GetPhase() != domain.MemoryPhaseFlip1 {
			break
		}
		if mi.m.IsHumanTurn() {
			break
		}
		mi.m.CpuFlip()
		mi.m.ResolveFlip()
	}
}
