//go:build !js || !wasm || extra5

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TongitsInteractorIF Tongitsインタラクターインタフェース
type TongitsInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TongitsConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札からカードを引く
	DrawFromDiscard() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// Knock ノックする
	Knock(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TongitsConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TongitsInteractor Tongitsインタラクタークラス
type TongitsInteractor struct {
	GameBase[interfaces.TongitsGame]
	gp presenter.TongitsPresenter
}

// NewTongitsInteractor コンストラクタ
func NewTongitsInteractor(g interfaces.TongitsGame, gp presenter.TongitsPresenter) *TongitsInteractor {
	mustNotNil("TongitsInteractor", map[string]any{"g": g, "gp": gp})
	return &TongitsInteractor{GameBase: GameBase[interfaces.TongitsGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *TongitsInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *TongitsInteractor) ResetWithConfig(cfg domain.TongitsConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *TongitsInteractor) DrawFromStock() string {
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
func (ci *TongitsInteractor) DrawFromDiscard() string {
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
func (ci *TongitsInteractor) Discard(cardIndex int) string {
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
func (ci *TongitsInteractor) Knock(cardIndex int) string {
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

// NextRound 次のラウンドへ進む
func (ci *TongitsInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *TongitsInteractor) GetConfig() domain.TongitsConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *TongitsInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行
func (ci *TongitsInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.TongitsPhaseRoundEnd || phase == domain.TongitsPhaseGameEnd || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreTongitsInteractor deserialises JSON into a TongitsInteractor.
func RestoreTongitsInteractor(data []byte, gp presenter.TongitsPresenter) (*TongitsInteractor, error) {
	return restoreAndBuild[domain.Tongits](data, func(g *domain.Tongits) *TongitsInteractor {
		return &TongitsInteractor{GameBase: GameBase[interfaces.TongitsGame]{Game: g}, gp: gp}
	})
}
