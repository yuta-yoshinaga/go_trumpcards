package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PigsTailInteractorIF ぶたのしっぽインタラクターインタフェース
type PigsTailInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset(config domain.PigsTailConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.PigsTailConfig
	// Action 人間プレイヤーのアクション (山札から1枚引く)
	Action(actionIdx int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PigsTailInteractor ぶたのしっぽインタラクタークラス
type PigsTailInteractor struct {
	GameBase[interfaces.PigsTailGame]
	ptp presenter.PigsTailPresenter
}

// NewPigsTailInteractor コンストラクタ
func NewPigsTailInteractor(pt interfaces.PigsTailGame, ptp presenter.PigsTailPresenter) *PigsTailInteractor {
	mustNotNil("PigsTailInteractor", map[string]any{"pt": pt, "ptp": ptp})
	return &PigsTailInteractor{
		GameBase: GameBase[interfaces.PigsTailGame]{Game: pt},
		ptp:      ptp,
	}
}

// GetConfig 現在の設定を返す
func (pi *PigsTailInteractor) GetConfig() domain.PigsTailConfig {
	return pi.Game.GetConfig()
}

// Reset ゲーム初期化
func (pi *PigsTailInteractor) Reset(config domain.PigsTailConfig) string {
	if err := config.Validate(); err != nil {
		return pi.ptp.Output(pi.Game, err)
	}
	pi.Game.SetConfig(config)
	pi.Game.Reset()
	pi.runCpuTurns()
	return pi.ptp.Output(pi.Game, nil)
}

// Action 人間プレイヤーのアクション (山札から1枚引く)
func (pi *PigsTailInteractor) Action(actionIdx int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.ptp); blocked {
		return out
	}
	err := pi.Game.PlayerAction(actionIdx)
	if err == nil && !pi.Game.GetGameEndFlag() {
		pi.runCpuTurns()
	}
	return pi.ptp.Output(pi.Game, err)
}

// ActionLog 棋譜を出力する
func (pi *PigsTailInteractor) ActionLog() string {
	return pi.ptp.ActionLogOutput(pi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (pi *PigsTailInteractor) runCpuTurns() {
	runCpuTurnsCapped(pi.Game, func() { _ = pi.Game.CpuAction() })
}

// RestorePigsTailInteractor deserialises JSON into a PigsTailInteractor.
func RestorePigsTailInteractor(data []byte, ptp presenter.PigsTailPresenter) (*PigsTailInteractor, error) {
	return restoreAndBuild[domain.PigsTail](data, func(g *domain.PigsTail) *PigsTailInteractor {
		return &PigsTailInteractor{GameBase: GameBase[interfaces.PigsTailGame]{Game: g}, ptp: ptp}
	})
}
