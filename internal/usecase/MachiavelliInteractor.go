//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MachiavelliInteractorIF マキャヴェッリインタラクターインタフェース
type MachiavelliInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MachiavelliConfig) string
	// Draw 山札からカードを引く（ターン終了）
	Draw() string
	// Play 新しい場（メルド群）と追加する手札インデックスを提出する
	Play(refs [][]domain.MachiavelliCardRef, handIndices []int) string
	// NewMeld 手札インデックスから新しいメルドを 1 つ場に出す
	NewMeld(handIndices []int) string
	// Layoff 手札 1 枚を既存メルドに追加する
	Layoff(meldIdx, handIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MachiavelliConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MachiavelliInteractor マキャヴェッリインタラクター
type MachiavelliInteractor struct {
	GameBase[interfaces.MachiavelliGame]
	gp presenter.MachiavelliPresenter
}

// NewMachiavelliInteractor コンストラクタ
func NewMachiavelliInteractor(g interfaces.MachiavelliGame, gp presenter.MachiavelliPresenter) *MachiavelliInteractor {
	mustNotNil("MachiavelliInteractor", map[string]any{"g": g, "gp": gp})
	return &MachiavelliInteractor{GameBase: GameBase[interfaces.MachiavelliGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *MachiavelliInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *MachiavelliInteractor) ResetWithConfig(cfg domain.MachiavelliConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Draw 山札からカードを引く（ターン終了）
func (ci *MachiavelliInteractor) Draw() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDraw(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Play 新しい場（メルド群）と追加する手札インデックスを提出する
func (ci *MachiavelliInteractor) Play(refs [][]domain.MachiavelliCardRef, handIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(refs, handIndices); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NewMeld 手札インデックスから新しいメルドを 1 つ場に出す
func (ci *MachiavelliInteractor) NewMeld(handIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerNewMeld(handIndices); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Layoff 手札 1 枚を既存メルドに追加する
func (ci *MachiavelliInteractor) Layoff(meldIdx, handIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerLayoff(meldIdx, handIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *MachiavelliInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *MachiavelliInteractor) GetConfig() domain.MachiavelliConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *MachiavelliInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *MachiavelliInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase == domain.MachiavelliPhaseRoundEnd || phase == domain.MachiavelliPhaseGameEnd {
			return
		}
		if ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestoreMachiavelliInteractor JSON から MachiavelliInteractor を復元する
func RestoreMachiavelliInteractor(data []byte, gp presenter.MachiavelliPresenter) (*MachiavelliInteractor, error) {
	return restoreAndBuild[domain.Machiavelli](data, func(g *domain.Machiavelli) *MachiavelliInteractor {
		return &MachiavelliInteractor{GameBase: GameBase[interfaces.MachiavelliGame]{Game: g}, gp: gp}
	})
}
