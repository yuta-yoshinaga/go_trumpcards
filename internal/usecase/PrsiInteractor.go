package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PrsiInteractorIF プルシーインタラクターインタフェース
type PrsiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PrsiConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// Draw カードを引く
	Draw() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PrsiConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PrsiInteractor プルシーインタラクタークラス
type PrsiInteractor struct {
	GameBase[interfaces.PrsiGame]
	gp presenter.PrsiPresenter
}

// NewPrsiInteractor コンストラクタ
func NewPrsiInteractor(g interfaces.PrsiGame, gp presenter.PrsiPresenter) *PrsiInteractor {
	mustNotNil("PrsiInteractor", map[string]any{"g": g, "gp": gp})
	return &PrsiInteractor{GameBase: GameBase[interfaces.PrsiGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *PrsiInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *PrsiInteractor) ResetWithConfig(cfg domain.PrsiConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードをプレイ
func (ci *PrsiInteractor) Play(cardIndex int) string {
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

// Draw カードを引く
func (ci *PrsiInteractor) Draw() string {
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

// GetConfig 現在の設定を取得
func (ci *PrsiInteractor) GetConfig() domain.PrsiConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *PrsiInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行する。
// 勝敗判定・スキップ・ペナルティはすべてドメインで処理されるため、どの席
// (人間/CPU) が終端アクションを起こしても正しく処理される。
func (ci *PrsiInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.PrsiPhasePlay || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestorePrsiInteractor deserialises JSON into a PrsiInteractor.
func RestorePrsiInteractor(data []byte, gp presenter.PrsiPresenter) (*PrsiInteractor, error) {
	return restoreAndBuild[domain.Prsi](data, func(g *domain.Prsi) *PrsiInteractor {
		return &PrsiInteractor{GameBase: GameBase[interfaces.PrsiGame]{Game: g}, gp: gp}
	})
}
