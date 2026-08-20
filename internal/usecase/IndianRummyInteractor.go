//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// IndianRummyInteractorIF インドラミーインタラクターインタフェース
type IndianRummyInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.IndianRummyConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札トップからカードを引く
	DrawFromDiscard() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// Declare 宣言する
	Declare(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.IndianRummyConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// IndianRummyInteractor インドラミーインタラクター
type IndianRummyInteractor struct {
	GameBase[interfaces.IndianRummyGame]
	gp presenter.IndianRummyPresenter
}

// NewIndianRummyInteractor コンストラクタ
func NewIndianRummyInteractor(g interfaces.IndianRummyGame, gp presenter.IndianRummyPresenter) *IndianRummyInteractor {
	mustNotNil("IndianRummyInteractor", map[string]any{"g": g, "gp": gp})
	return &IndianRummyInteractor{GameBase: GameBase[interfaces.IndianRummyGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *IndianRummyInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *IndianRummyInteractor) ResetWithConfig(cfg domain.IndianRummyConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *IndianRummyInteractor) DrawFromStock() string {
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
func (ci *IndianRummyInteractor) DrawFromDiscard() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromDiscard(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨てる
func (ci *IndianRummyInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Declare 宣言する
func (ci *IndianRummyInteractor) Declare(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDeclare(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *IndianRummyInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *IndianRummyInteractor) GetConfig() domain.IndianRummyConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *IndianRummyInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *IndianRummyInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.IndianRummyPhaseRoundEnd || phase == domain.IndianRummyPhaseGameEnd || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreIndianRummyInteractor JSON から IndianRummyInteractor を復元する
func RestoreIndianRummyInteractor(data []byte, gp presenter.IndianRummyPresenter) (*IndianRummyInteractor, error) {
	return restoreAndBuild[domain.IndianRummy](data, func(g *domain.IndianRummy) *IndianRummyInteractor {
		return &IndianRummyInteractor{GameBase: GameBase[interfaces.IndianRummyGame]{Game: g}, gp: gp}
	})
}
