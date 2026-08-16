//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ThreeThirteenInteractorIF スリー・サーティーンインタラクターインタフェース
type ThreeThirteenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ThreeThirteenConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札トップからカードを引く
	DrawFromDiscard() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// Knock カードを捨ててノックする
	Knock(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ThreeThirteenConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ThreeThirteenInteractor スリー・サーティーンインタラクター
type ThreeThirteenInteractor struct {
	GameBase[interfaces.ThreeThirteenGame]
	gp presenter.ThreeThirteenPresenter
}

// NewThreeThirteenInteractor コンストラクタ
func NewThreeThirteenInteractor(g interfaces.ThreeThirteenGame, gp presenter.ThreeThirteenPresenter) *ThreeThirteenInteractor {
	mustNotNil("ThreeThirteenInteractor", map[string]any{"g": g, "gp": gp})
	return &ThreeThirteenInteractor{GameBase: GameBase[interfaces.ThreeThirteenGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *ThreeThirteenInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ThreeThirteenInteractor) ResetWithConfig(cfg domain.ThreeThirteenConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *ThreeThirteenInteractor) DrawFromStock() string {
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
func (ci *ThreeThirteenInteractor) DrawFromDiscard() string {
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
func (ci *ThreeThirteenInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Knock カードを捨ててノックする
func (ci *ThreeThirteenInteractor) Knock(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerKnock(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *ThreeThirteenInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *ThreeThirteenInteractor) GetConfig() domain.ThreeThirteenConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *ThreeThirteenInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *ThreeThirteenInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.ThreeThirteenPhaseRoundEnd || phase == domain.ThreeThirteenPhaseGameEnd || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreThreeThirteenInteractor JSON から ThreeThirteenInteractor を復元する
func RestoreThreeThirteenInteractor(data []byte, gp presenter.ThreeThirteenPresenter) (*ThreeThirteenInteractor, error) {
	return restoreAndBuild[domain.ThreeThirteen](data, func(g *domain.ThreeThirteen) *ThreeThirteenInteractor {
		return &ThreeThirteenInteractor{GameBase: GameBase[interfaces.ThreeThirteenGame]{Game: g}, gp: gp}
	})
}
