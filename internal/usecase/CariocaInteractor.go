//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CariocaInteractorIF カリオカインタラクターインタフェース
type CariocaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CariocaConfig) string
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
	GetConfig() domain.CariocaConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CariocaInteractor カリオカインタラクター
type CariocaInteractor struct {
	GameBase[interfaces.CariocaGame]
	gp presenter.CariocaPresenter
}

// NewCariocaInteractor コンストラクタ
func NewCariocaInteractor(g interfaces.CariocaGame, gp presenter.CariocaPresenter) *CariocaInteractor {
	mustNotNil("CariocaInteractor", map[string]any{"g": g, "gp": gp})
	return &CariocaInteractor{GameBase: GameBase[interfaces.CariocaGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *CariocaInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CariocaInteractor) ResetWithConfig(cfg domain.CariocaConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *CariocaInteractor) DrawFromStock() string {
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
func (ci *CariocaInteractor) DrawFromDiscard() string {
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
func (ci *CariocaInteractor) MeldContract(indicesPerSlot [][]int) string {
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
func (ci *CariocaInteractor) MeldExtra(indices []int) string {
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
func (ci *CariocaInteractor) Layoff(targetPlayerIdx, meldIdx, cardIndex int) string {
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
func (ci *CariocaInteractor) Discard(cardIndex int) string {
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
func (ci *CariocaInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *CariocaInteractor) GetConfig() domain.CariocaConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *CariocaInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *CariocaInteractor) runCpuTurns() {
	for !ci.Game.GetGameEndFlag() {
		phase := ci.Game.GetPhase()
		if phase == domain.CariocaPhaseRoundEnd || phase == domain.CariocaPhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

// RestoreCariocaInteractor JSON から CariocaInteractor を復元する
func RestoreCariocaInteractor(data []byte, gp presenter.CariocaPresenter) (*CariocaInteractor, error) {
	return restoreAndBuild[domain.Carioca](data, func(g *domain.Carioca) *CariocaInteractor {
		return &CariocaInteractor{GameBase: GameBase[interfaces.CariocaGame]{Game: g}, gp: gp}
	})
}
