//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MemoryInteractorIF 神経衰弱インタラクターインタフェース
type MemoryInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MemoryConfig) string
	// Flip カードをめくる
	Flip(pos int) string
	// Next 結果を解決し、CPUターンを実行する
	Next() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MemoryConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MemoryInteractor 神経衰弱インタラクタークラス
type MemoryInteractor struct {
	GameBase[interfaces.MemoryGame]
	mp presenter.MemoryPresenter
}

// NewMemoryInteractor コンストラクタ
func NewMemoryInteractor(m interfaces.MemoryGame, mp presenter.MemoryPresenter) *MemoryInteractor {
	mustNotNil("MemoryInteractor", map[string]any{"m": m, "mp": mp})
	return &MemoryInteractor{GameBase: GameBase[interfaces.MemoryGame]{Game: m}, mp: mp}
}

// Reset ゲーム初期化
func (mi *MemoryInteractor) Reset() string {
	return runAndPresent(mi.Game, mi.mp, mi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *MemoryInteractor) ResetWithConfig(cfg domain.MemoryConfig) string {
	mi.Game.SetConfig(cfg)
	return mi.Reset()
}

// Flip カードをめくる
func (mi *MemoryInteractor) Flip(pos int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	err := mi.Game.PlayerFlip(pos)
	if err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	return mi.mp.Output(mi.Game, nil)
}

// Next 結果を解決し、CPU ターンを実行する
func (mi *MemoryInteractor) Next() string {
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.ResolveFlip()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *MemoryInteractor) GetConfig() domain.MemoryConfig {
	return mi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (mi *MemoryInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (mi *MemoryInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if mi.Game.GetGameEndFlag() {
			return
		}
		if mi.Game.GetPhase() != domain.MemoryPhaseFlip1 {
			break
		}
		if mi.Game.IsHumanTurn() {
			break
		}
		mi.Game.CpuFlip()
		mi.Game.ResolveFlip()
	}
}

// RestoreMemoryInteractor deserialises JSON into a MemoryInteractor.
func RestoreMemoryInteractor(data []byte, mp presenter.MemoryPresenter) (*MemoryInteractor, error) {
	return restoreAndBuild[domain.Memory](data, func(g *domain.Memory) *MemoryInteractor {
		return &MemoryInteractor{GameBase: GameBase[interfaces.MemoryGame]{Game: g}, mp: mp}
	})
}
