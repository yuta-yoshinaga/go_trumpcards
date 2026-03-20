package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CrazyEightsInteractorIF クレイジーエイトインタラクターインタフェース
type CrazyEightsInteractorIF interface {
	Reset() string
	ResetWithConfig(cfg domain.CrazyEightsConfig) string
	Play(cardIndex int) string
	ChooseSuit(suit int) string
	Draw() string
	NextRound() string
	GetConfig() domain.CrazyEightsConfig
	ActionLog() string
}

// CrazyEightsInteractor クレイジーエイトインタラクタークラス
type CrazyEightsInteractor struct {
	g  interfaces.CrazyEightsGame
	gp presenter.CrazyEightsPresenter
}

// NewCrazyEightsInteractor コンストラクタ
func NewCrazyEightsInteractor(g interfaces.CrazyEightsGame, gp presenter.CrazyEightsPresenter) *CrazyEightsInteractor {
	mustNotNil("CrazyEightsInteractor", map[string]any{"g": g, "gp": gp})
	return &CrazyEightsInteractor{g: g, gp: gp}
}

// Reset ゲーム初期化
func (ci *CrazyEightsInteractor) Reset() string {
	ci.g.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CrazyEightsInteractor) ResetWithConfig(cfg domain.CrazyEightsConfig) string {
	if err := cfg.Validate(); err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.g.SetConfig(cfg)
	return ci.Reset()
}

// Play カードをプレイ
func (ci *CrazyEightsInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerPlay(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// ChooseSuit スートを選択 (8を出した後)
func (ci *CrazyEightsInteractor) ChooseSuit(suit int) string {
	if out, blocked := guardGameEnd(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerChooseSuit(suit)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// Draw カードを引く
func (ci *CrazyEightsInteractor) Draw() string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerDraw()
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *CrazyEightsInteractor) NextRound() string {
	ci.g.ScoreRound()
	if out, blocked := guardGameEnd(ci.g, ci.gp); blocked {
		return out
	}
	ci.g.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// GetConfig 現在の設定を取得
func (ci *CrazyEightsInteractor) GetConfig() domain.CrazyEightsConfig {
	return ci.g.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *CrazyEightsInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.g)
}

// runCpuTurns ゲームが終わるか人間の手番またはラウンド/ゲーム終了になるまでCPUターンを実行
func (ci *CrazyEightsInteractor) runCpuTurns() {
	for !ci.g.GetGameEndFlag() {
		phase := ci.g.GetPhase()
		if phase == CrazyEightsPhaseRoundEnd || phase == CrazyEightsPhaseGameEnd {
			break
		}
		if phase == domain.CrazyEightsPhaseChooseSuit {
			if ci.g.IsHumanTurn() {
				break
			}
			ci.g.CpuChooseSuit()
			continue
		}
		if phase != domain.CrazyEightsPhasePlay {
			break
		}
		if ci.g.IsHumanTurn() {
			break
		}
		ci.g.CpuPlay()
	}
}

// CrazyEightsPhaseRoundEnd is imported from domain for convenience in this file.
const (
	CrazyEightsPhaseRoundEnd = domain.CrazyEightsPhaseRoundEnd
	CrazyEightsPhaseGameEnd  = domain.CrazyEightsPhaseGameEnd
)
