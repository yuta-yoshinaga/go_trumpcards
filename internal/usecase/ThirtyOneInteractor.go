//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ThirtyOneInteractorIF ThirtyOne インタラクターインタフェース
type ThirtyOneInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ThirtyOneConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札からカードを引く
	DrawFromDiscard() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// Knock ノックする
	Knock() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ThirtyOneConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// ThirtyOneInteractor ThirtyOne インタラクタークラス
type ThirtyOneInteractor struct {
	GameBase[interfaces.ThirtyOneGame]
	gp presenter.ThirtyOnePresenter
}

// NewThirtyOneInteractor コンストラクタ
func NewThirtyOneInteractor(g interfaces.ThirtyOneGame, gp presenter.ThirtyOnePresenter) *ThirtyOneInteractor {
	mustNotNil("ThirtyOneInteractor", map[string]any{"g": g, "gp": gp})
	return &ThirtyOneInteractor{GameBase: GameBase[interfaces.ThirtyOneGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *ThirtyOneInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ThirtyOneInteractor) ResetWithConfig(cfg domain.ThirtyOneConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *ThirtyOneInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromStock(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawFromDiscard 捨て札からカードを引く
func (ci *ThirtyOneInteractor) DrawFromDiscard() string {
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
func (ci *ThirtyOneInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Knock ノックする
func (ci *ThirtyOneInteractor) Knock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerKnock(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *ThirtyOneInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *ThirtyOneInteractor) GetConfig() domain.ThirtyOneConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *ThirtyOneInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// Hint ヒントを出力する
func (ci *ThirtyOneInteractor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// thirtyOneMaxCpuSteps bounds runCpuTurns so a malformed state can never spin
// the CPU loop forever (defensive — normal play always reaches a human turn,
// round end, or game end well within this limit).
const thirtyOneMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (ci *ThirtyOneInteractor) runCpuTurns() {
	for step := 0; step < thirtyOneMaxCpuSteps && !ci.Game.GetGameEndFlag(); step++ {
		phase := ci.Game.GetPhase()
		if phase == domain.ThirtyOnePhaseRoundEnd || phase == domain.ThirtyOnePhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

// RestoreThirtyOneInteractor deserialises JSON into a ThirtyOneInteractor.
func RestoreThirtyOneInteractor(data []byte, gp presenter.ThirtyOnePresenter) (*ThirtyOneInteractor, error) {
	return restoreAndBuild[domain.ThirtyOne](data, func(g *domain.ThirtyOne) *ThirtyOneInteractor {
		return &ThirtyOneInteractor{GameBase: GameBase[interfaces.ThirtyOneGame]{Game: g}, gp: gp}
	})
}
