//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ConquianInteractorIF コンキャンインタラクターインタフェース
type ConquianInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ConquianConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札からカードを引く
	DrawFromDiscard() string
	// Meld メルドを並べる/付ける
	Meld(meldGroups [][]int) string
	// MeldWithTargets 延長先メルドの指定つきでメルドする
	MeldWithTargets(meldGroups [][]int, extendTargets []int) string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ConquianConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ConquianInteractor コンキャンインタラクタークラス
type ConquianInteractor struct {
	GameBase[interfaces.ConquianGame]
	gp presenter.ConquianPresenter
}

// NewConquianInteractor コンストラクタ
func NewConquianInteractor(g interfaces.ConquianGame, gp presenter.ConquianPresenter) *ConquianInteractor {
	mustNotNil("ConquianInteractor", map[string]any{"g": g, "gp": gp})
	return &ConquianInteractor{GameBase: GameBase[interfaces.ConquianGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *ConquianInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ConquianInteractor) ResetWithConfig(cfg domain.ConquianConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *ConquianInteractor) DrawFromStock() string {
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
func (ci *ConquianInteractor) DrawFromDiscard() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromDiscard(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Meld メルドを並べる/付ける
func (ci *ConquianInteractor) Meld(meldGroups [][]int) string {
	return ci.MeldWithTargets(meldGroups, nil)
}

// MeldWithTargets は延長先メルドの指定つきでメルドする。extendTargets[i] は
// meldGroups[i] の延長先。指定が無ければ従来どおり最初に延長できるメルドへ。
func (ci *ConquianInteractor) MeldWithTargets(meldGroups [][]int, extendTargets []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerMeldWithTargets(meldGroups, extendTargets); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨てる
func (ci *ConquianInteractor) Discard(cardIndex int) string {
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
func (ci *ConquianInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *ConquianInteractor) GetConfig() domain.ConquianConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *ConquianInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行
func (ci *ConquianInteractor) runCpuTurns() {
	for !ci.Game.GetGameEndFlag() {
		phase := ci.Game.GetPhase()
		if phase == domain.ConquianPhaseRoundEnd || phase == domain.ConquianPhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

// RestoreConquianInteractor deserialises JSON into a ConquianInteractor.
func RestoreConquianInteractor(data []byte, gp presenter.ConquianPresenter) (*ConquianInteractor, error) {
	return restoreAndBuild[domain.Conquian](data, func(g *domain.Conquian) *ConquianInteractor {
		return &ConquianInteractor{GameBase: GameBase[interfaces.ConquianGame]{Game: g}, gp: gp}
	})
}
