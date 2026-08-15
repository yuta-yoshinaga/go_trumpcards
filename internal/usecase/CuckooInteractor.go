//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CuckooInteractorIF Cuckoo (カッコー) のインタラクターインタフェース
type CuckooInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CuckooConfig) string
	// Keep 手札を保持する
	Keep() string
	// Swap スワップを試みる
	Swap() string
	// Refuse King 隣人がスワップを拒否する
	Refuse() string
	// AcceptSwap King 隣人がスワップを受け入れる
	AcceptSwap() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CuckooConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CuckooInteractor Cuckoo (カッコー) のインタラクタークラス
type CuckooInteractor struct {
	GameBase[interfaces.CuckooGame]
	gp presenter.CuckooPresenter
}

// NewCuckooInteractor コンストラクタ
func NewCuckooInteractor(g interfaces.CuckooGame, gp presenter.CuckooPresenter) *CuckooInteractor {
	mustNotNil("CuckooInteractor", map[string]any{"g": g, "gp": gp})
	return &CuckooInteractor{GameBase: GameBase[interfaces.CuckooGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *CuckooInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CuckooInteractor) ResetWithConfig(cfg domain.CuckooConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Keep 手札を保持する
func (ci *CuckooInteractor) Keep() string {
	return ci.act(ci.Game.PlayerKeep)
}

// Swap スワップを試みる
func (ci *CuckooInteractor) Swap() string {
	return ci.act(ci.Game.PlayerSwap)
}

// Refuse King 隣人がスワップを拒否する
func (ci *CuckooInteractor) Refuse() string {
	return ci.act(ci.Game.PlayerRefuse)
}

// AcceptSwap King 隣人がスワップを受け入れる
func (ci *CuckooInteractor) AcceptSwap() string {
	return ci.act(ci.Game.PlayerAcceptSwap)
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (ci *CuckooInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *CuckooInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *CuckooInteractor) GetConfig() domain.CuckooConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *CuckooInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// cuckooMaxCpuSteps bounds runCpuTurns so a malformed state can never spin
// the CPU loop forever (defensive — normal play always reaches a human turn,
// round end, or game end well within this limit).
const cuckooMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (ci *CuckooInteractor) runCpuTurns() {
	for step := 0; step < cuckooMaxCpuSteps && !ci.Game.GetGameEndFlag(); step++ {
		phase := ci.Game.GetPhase()
		if phase == domain.CuckooPhaseRoundEnd || phase == domain.CuckooPhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

// RestoreCuckooInteractor deserialises JSON into a CuckooInteractor.
func RestoreCuckooInteractor(data []byte, gp presenter.CuckooPresenter) (*CuckooInteractor, error) {
	return restoreAndBuild[domain.Cuckoo](data, func(g *domain.Cuckoo) *CuckooInteractor {
		return &CuckooInteractor{GameBase: GameBase[interfaces.CuckooGame]{Game: g}, gp: gp}
	})
}
