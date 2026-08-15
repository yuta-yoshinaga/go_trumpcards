//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GinRummyInteractorIF ジンラミーインタラクターインタフェース
type GinRummyInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.GinRummyConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札からカードを引く
	DrawFromDiscard() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// Knock ノックする
	Knock(cardIndex int) string
	// Layoff レイオフする
	Layoff(cardIndices []int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.GinRummyConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GinRummyInteractor ジンラミーインタラクタークラス
type GinRummyInteractor struct {
	GameBase[interfaces.GinRummyGame]
	gp presenter.GinRummyPresenter
}

// NewGinRummyInteractor コンストラクタ
func NewGinRummyInteractor(g interfaces.GinRummyGame, gp presenter.GinRummyPresenter) *GinRummyInteractor {
	mustNotNil("GinRummyInteractor", map[string]any{"g": g, "gp": gp})
	return &GinRummyInteractor{GameBase: GameBase[interfaces.GinRummyGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *GinRummyInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *GinRummyInteractor) ResetWithConfig(cfg domain.GinRummyConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *GinRummyInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDrawFromStock()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawFromDiscard 捨て札からカードを引く
func (ci *GinRummyInteractor) DrawFromDiscard() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDrawFromDiscard()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨てる
func (ci *GinRummyInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDiscard(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Knock ノックする
func (ci *GinRummyInteractor) Knock(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerKnock(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Layoff レイオフする
func (ci *GinRummyInteractor) Layoff(cardIndices []int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerLayoff(cardIndices)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *GinRummyInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *GinRummyInteractor) GetConfig() domain.GinRummyConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *GinRummyInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行
func (ci *GinRummyInteractor) runCpuTurns() {
	for !ci.Game.GetGameEndFlag() {
		phase := ci.Game.GetPhase()
		if phase == domain.GinRummyPhaseRoundEnd || phase == domain.GinRummyPhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

// RestoreGinRummyInteractor deserialises JSON into a GinRummyInteractor.
func RestoreGinRummyInteractor(data []byte, gp presenter.GinRummyPresenter) (*GinRummyInteractor, error) {
	return restoreAndBuild[domain.GinRummy](data, func(g *domain.GinRummy) *GinRummyInteractor {
		return &GinRummyInteractor{GameBase: GameBase[interfaces.GinRummyGame]{Game: g}, gp: gp}
	})
}
