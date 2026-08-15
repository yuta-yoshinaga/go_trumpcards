//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KalookiInteractorIF カルーキインタラクターインタフェース
type KalookiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KalookiConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札トップからカードを引く
	DrawFromDiscard() string
	// Meld メルド群を場に出す（オープニング要件チェックを含む）
	Meld(meldGroups [][]int) string
	// Layoff 既存メルドにカードを 1 枚追加する
	Layoff(targetPlayerIdx, meldIdx, cardIndex int) string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KalookiConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KalookiInteractor カルーキインタラクター
type KalookiInteractor struct {
	GameBase[interfaces.KalookiGame]
	gp presenter.KalookiPresenter
}

// NewKalookiInteractor コンストラクタ
func NewKalookiInteractor(g interfaces.KalookiGame, gp presenter.KalookiPresenter) *KalookiInteractor {
	mustNotNil("KalookiInteractor", map[string]any{"g": g, "gp": gp})
	return &KalookiInteractor{GameBase: GameBase[interfaces.KalookiGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *KalookiInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *KalookiInteractor) ResetWithConfig(cfg domain.KalookiConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *KalookiInteractor) DrawFromStock() string {
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
func (ci *KalookiInteractor) DrawFromDiscard() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromDiscard(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Meld メルド群を場に出す
func (ci *KalookiInteractor) Meld(meldGroups [][]int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerMeld(meldGroups); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Layoff 既存メルドにカードを 1 枚追加する
func (ci *KalookiInteractor) Layoff(targetPlayerIdx, meldIdx, cardIndex int) string {
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
func (ci *KalookiInteractor) Discard(cardIndex int) string {
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
func (ci *KalookiInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *KalookiInteractor) GetConfig() domain.KalookiConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *KalookiInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *KalookiInteractor) runCpuTurns() {
	for !ci.Game.GetGameEndFlag() {
		phase := ci.Game.GetPhase()
		if phase == domain.KalookiPhaseRoundEnd || phase == domain.KalookiPhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

// RestoreKalookiInteractor JSON から KalookiInteractor を復元する
func RestoreKalookiInteractor(data []byte, gp presenter.KalookiPresenter) (*KalookiInteractor, error) {
	return restoreAndBuild[domain.Kalooki](data, func(g *domain.Kalooki) *KalookiInteractor {
		return &KalookiInteractor{GameBase: GameBase[interfaces.KalookiGame]{Game: g}, gp: gp}
	})
}
