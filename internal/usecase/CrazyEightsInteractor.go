package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CrazyEightsInteractorIF クレイジーエイトインタラクターインタフェース
type CrazyEightsInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CrazyEightsConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// ChooseSuit スートを選択 (8を出した後)
	ChooseSuit(suit int) string
	// Draw カードを引く
	Draw() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CrazyEightsConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CrazyEightsInteractor クレイジーエイトインタラクタークラス
type CrazyEightsInteractor struct {
	GameBase[interfaces.CrazyEightsGame]
	gp presenter.CrazyEightsPresenter
}

// NewCrazyEightsInteractor コンストラクタ
func NewCrazyEightsInteractor(g interfaces.CrazyEightsGame, gp presenter.CrazyEightsPresenter) *CrazyEightsInteractor {
	mustNotNil("CrazyEightsInteractor", map[string]any{"g": g, "gp": gp})
	return &CrazyEightsInteractor{GameBase: GameBase[interfaces.CrazyEightsGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *CrazyEightsInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CrazyEightsInteractor) ResetWithConfig(cfg domain.CrazyEightsConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードをプレイ
func (ci *CrazyEightsInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ChooseSuit スートを選択 (8を出した後)
func (ci *CrazyEightsInteractor) ChooseSuit(suit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerChooseSuit(suit)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Draw カードを引く
func (ci *CrazyEightsInteractor) Draw() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDraw()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *CrazyEightsInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *CrazyEightsInteractor) GetConfig() domain.CrazyEightsConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *CrazyEightsInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはラウンド/ゲーム終了になるまでCPUターンを実行
func (ci *CrazyEightsInteractor) runCpuTurns() {
	for !ci.Game.GetGameEndFlag() {
		phase := ci.Game.GetPhase()
		if phase == CrazyEightsPhaseRoundEnd || phase == CrazyEightsPhaseGameEnd {
			break
		}
		if phase == domain.CrazyEightsPhaseChooseSuit {
			if ci.Game.IsHumanTurn() {
				break
			}
			ci.Game.CpuChooseSuit()
			continue
		}
		if phase != domain.CrazyEightsPhasePlay {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

const (
	// CrazyEightsPhaseRoundEnd ラウンド終了フェーズ (domain からの再エクスポート)
	CrazyEightsPhaseRoundEnd = domain.CrazyEightsPhaseRoundEnd
	// CrazyEightsPhaseGameEnd ゲーム終了フェーズ (domain からの再エクスポート)
	CrazyEightsPhaseGameEnd = domain.CrazyEightsPhaseGameEnd
)

// RestoreCrazyEightsInteractor deserialises JSON into a CrazyEightsInteractor.
func RestoreCrazyEightsInteractor(data []byte, gp presenter.CrazyEightsPresenter) (*CrazyEightsInteractor, error) {
	return restoreAndBuild[domain.CrazyEights](data, func(g *domain.CrazyEights) *CrazyEightsInteractor {
		return &CrazyEightsInteractor{GameBase: GameBase[interfaces.CrazyEightsGame]{Game: g}, gp: gp}
	})
}
