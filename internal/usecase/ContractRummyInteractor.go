//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ContractRummyInteractorIF コントラクトラミーインタラクターインタフェース
type ContractRummyInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ContractRummyConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札トップからカードを引く
	DrawFromDiscard() string
	// MeldContract コントラクトを達成するメルドを場に出す
	MeldContract(indicesPerSlot [][]int) string
	// MeldExtra コントラクト達成後の追加メルドを場に出す
	MeldExtra(indices []int) string
	// Layoff 既存メルドにカードを 1 枚追加する
	Layoff(targetPlayerIdx, meldIdx, cardIndex int) string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ContractRummyConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ContractRummyInteractor コントラクトラミーインタラクター
type ContractRummyInteractor struct {
	GameBase[interfaces.ContractRummyGame]
	gp presenter.ContractRummyPresenter
}

// NewContractRummyInteractor コンストラクタ
func NewContractRummyInteractor(g interfaces.ContractRummyGame, gp presenter.ContractRummyPresenter) *ContractRummyInteractor {
	mustNotNil("ContractRummyInteractor", map[string]any{"g": g, "gp": gp})
	return &ContractRummyInteractor{GameBase: GameBase[interfaces.ContractRummyGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *ContractRummyInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ContractRummyInteractor) ResetWithConfig(cfg domain.ContractRummyConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *ContractRummyInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromStock(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawFromDiscard 捨て札トップからカードを引く
func (ci *ContractRummyInteractor) DrawFromDiscard() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromDiscard(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// MeldContract コントラクトを達成するメルドを場に出す
func (ci *ContractRummyInteractor) MeldContract(indicesPerSlot [][]int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerMeldContract(indicesPerSlot); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// MeldExtra コントラクト達成後の追加メルドを場に出す
func (ci *ContractRummyInteractor) MeldExtra(indices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerMeldExtra(indices); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Layoff 既存メルドにカードを 1 枚追加する
func (ci *ContractRummyInteractor) Layoff(targetPlayerIdx, meldIdx, cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨てる
func (ci *ContractRummyInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *ContractRummyInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *ContractRummyInteractor) GetConfig() domain.ContractRummyConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *ContractRummyInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *ContractRummyInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase == domain.ContractRummyPhaseRoundEnd || phase == domain.ContractRummyPhaseGameEnd {
			return
		}
		if ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestoreContractRummyInteractor JSON から ContractRummyInteractor を復元する
func RestoreContractRummyInteractor(data []byte, gp presenter.ContractRummyPresenter) (*ContractRummyInteractor, error) {
	return restoreAndBuild[domain.ContractRummy](data, func(g *domain.ContractRummy) *ContractRummyInteractor {
		return &ContractRummyInteractor{GameBase: GameBase[interfaces.ContractRummyGame]{Game: g}, gp: gp}
	})
}
