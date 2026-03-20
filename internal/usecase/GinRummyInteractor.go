package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GinRummyInteractorIF ジンラミーインタラクターインタフェース
type GinRummyInteractorIF interface {
	Reset() string
	ResetWithConfig(cfg domain.GinRummyConfig) string
	DrawFromStock() string
	DrawFromDiscard() string
	Discard(cardIndex int) string
	Knock(cardIndex int) string
	Layoff(cardIndices []int) string
	NextRound() string
	GetConfig() domain.GinRummyConfig
	ActionLog() string
}

// GinRummyInteractor ジンラミーインタラクタークラス
type GinRummyInteractor struct {
	g  interfaces.GinRummyGame
	gp presenter.GinRummyPresenter
}

// NewGinRummyInteractor コンストラクタ
func NewGinRummyInteractor(g interfaces.GinRummyGame, gp presenter.GinRummyPresenter) *GinRummyInteractor {
	mustNotNil("GinRummyInteractor", map[string]any{"g": g, "gp": gp})
	return &GinRummyInteractor{g: g, gp: gp}
}

// Reset ゲーム初期化
func (ci *GinRummyInteractor) Reset() string {
	ci.g.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *GinRummyInteractor) ResetWithConfig(cfg domain.GinRummyConfig) string {
	if err := cfg.Validate(); err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.g.SetConfig(cfg)
	return ci.Reset()
}

// DrawFromStock 山札からカードを引く
func (ci *GinRummyInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerDrawFromStock()
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// DrawFromDiscard 捨て札からカードを引く
func (ci *GinRummyInteractor) DrawFromDiscard() string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerDrawFromDiscard()
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// Discard カードを捨てる
func (ci *GinRummyInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerDiscard(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// Knock ノックする
func (ci *GinRummyInteractor) Knock(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerKnock(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// Layoff レイオフする
func (ci *GinRummyInteractor) Layoff(cardIndices []int) string {
	if out, blocked := guardGameEnd(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerLayoff(cardIndices)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// NextRound 次のラウンドへ進む
func (ci *GinRummyInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ci.g, ci.gp); blocked {
		return out
	}
	ci.g.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// GetConfig 現在の設定を取得
func (ci *GinRummyInteractor) GetConfig() domain.GinRummyConfig {
	return ci.g.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *GinRummyInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.g)
}

// runCpuTurns CPUターンを実行
func (ci *GinRummyInteractor) runCpuTurns() {
	for !ci.g.GetGameEndFlag() {
		phase := ci.g.GetPhase()
		if phase == domain.GinRummyPhaseRoundEnd || phase == domain.GinRummyPhaseGameEnd {
			break
		}
		if ci.g.IsHumanTurn() {
			break
		}
		ci.g.CpuPlay()
	}
}
