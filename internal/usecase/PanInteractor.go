//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PanInteractorIF パングインゲインタラクターインタフェース
type PanInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PanConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札トップからカードを引く
	DrawFromDiscard() string
	// Meld 手札のカードでメルドを場に出す
	Meld(cardIndices []int) string
	// Layoff 既存メルドにカードを追加する
	Layoff(meldOwner, meldIdx, cardIndex int) string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PanConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PanInteractor パングインゲインタラクター
type PanInteractor struct {
	GameBase[interfaces.PanGame]
	gp presenter.PanPresenter
}

// NewPanInteractor コンストラクタ
func NewPanInteractor(g interfaces.PanGame, gp presenter.PanPresenter) *PanInteractor {
	mustNotNil("PanInteractor", map[string]any{"g": g, "gp": gp})
	return &PanInteractor{GameBase: GameBase[interfaces.PanGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *PanInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *PanInteractor) ResetWithConfig(cfg domain.PanConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *PanInteractor) DrawFromStock() string {
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
func (ci *PanInteractor) DrawFromDiscard() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromDiscard(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Meld メルド
func (ci *PanInteractor) Meld(cardIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerMeld(cardIndices); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Layoff レイオフ
func (ci *PanInteractor) Layoff(meldOwner, meldIdx, cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerLayoff(meldOwner, meldIdx, cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨てる
func (ci *PanInteractor) Discard(cardIndex int) string {
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
func (ci *PanInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *PanInteractor) GetConfig() domain.PanConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *PanInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *PanInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase == domain.PanPhaseRoundEnd || phase == domain.PanPhaseGameEnd {
			return
		}
		if ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestorePanInteractor JSON から PanInteractor を復元する
func RestorePanInteractor(data []byte, gp presenter.PanPresenter) (*PanInteractor, error) {
	return restoreAndBuild[domain.Pan](data, func(g *domain.Pan) *PanInteractor {
		return &PanInteractor{GameBase: GameBase[interfaces.PanGame]{Game: g}, gp: gp}
	})
}
