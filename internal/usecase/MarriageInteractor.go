//go:build !js || !wasm || extra5

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MarriageInteractorIF マリッジインタラクターインタフェース
type MarriageInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MarriageConfig) string
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
	GetConfig() domain.MarriageConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MarriageInteractor マリッジインタラクター
type MarriageInteractor struct {
	GameBase[interfaces.MarriageGame]
	gp presenter.MarriagePresenter
}

// NewMarriageInteractor コンストラクタ
func NewMarriageInteractor(g interfaces.MarriageGame, gp presenter.MarriagePresenter) *MarriageInteractor {
	mustNotNil("MarriageInteractor", map[string]any{"g": g, "gp": gp})
	return &MarriageInteractor{GameBase: GameBase[interfaces.MarriageGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *MarriageInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *MarriageInteractor) ResetWithConfig(cfg domain.MarriageConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *MarriageInteractor) DrawFromStock() string {
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
func (ci *MarriageInteractor) DrawFromDiscard() string {
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
func (ci *MarriageInteractor) Discard(cardIndex int) string {
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
func (ci *MarriageInteractor) Declare(cardIndex int) string {
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
func (ci *MarriageInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *MarriageInteractor) GetConfig() domain.MarriageConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *MarriageInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *MarriageInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.MarriagePhaseRoundEnd || phase == domain.MarriagePhaseGameEnd || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreMarriageInteractor JSON から MarriageInteractor を復元する
func RestoreMarriageInteractor(data []byte, gp presenter.MarriagePresenter) (*MarriageInteractor, error) {
	return restoreAndBuild[domain.Marriage](data, func(g *domain.Marriage) *MarriageInteractor {
		return &MarriageInteractor{GameBase: GameBase[interfaces.MarriageGame]{Game: g}, gp: gp}
	})
}
