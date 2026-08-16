package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ShitheadInteractorIF シットヘッドインタラクターインタフェース
type ShitheadInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Play 人間プレイヤーがカードを出す (空 indices = ピックアップ)
	Play(indices []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.ShitheadConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.ShitheadConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ShitheadInteractor シットヘッドインタラクタークラス
type ShitheadInteractor struct {
	GameBase[interfaces.ShitheadGame]
	pp presenter.ShitheadPresenter
}

// NewShitheadInteractor コンストラクタ
func NewShitheadInteractor(sg interfaces.ShitheadGame, pp presenter.ShitheadPresenter) *ShitheadInteractor {
	mustNotNil("ShitheadInteractor", map[string]any{"sg": sg, "pp": pp})
	return &ShitheadInteractor{
		GameBase: GameBase[interfaces.ShitheadGame]{Game: sg},
		pp:       pp,
	}
}

// Reset ゲーム初期化
func (si *ShitheadInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.pp.Output(si.Game, nil)
}

// Play 人間プレイヤーがカードを出す
func (si *ShitheadInteractor) Play(indices []int) string {
	if out, blocked := guardNotPlayable(si.Game, si.pp); blocked {
		return out
	}
	err := si.Game.PlayerPlay(indices)
	if err == nil && !si.Game.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.pp.Output(si.Game, err)
}

// GetConfig 現在の設定を返す
func (si *ShitheadInteractor) GetConfig() domain.ShitheadConfig {
	return si.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (si *ShitheadInteractor) ResetWithConfig(config domain.ShitheadConfig) string {
	return resetWithValidatedConfig(si.Game, si.pp, config, si.Game.SetConfig, si.Reset)
}

// ActionLog 棋譜を出力する
func (si *ShitheadInteractor) ActionLog() string {
	return si.pp.ActionLogOutput(si.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (si *ShitheadInteractor) runCpuTurns() {
	runCpuTurnsCapped(si.Game, si.Game.CpuPlay)
}

// RestoreShitheadInteractor deserialises JSON into a ShitheadInteractor.
func RestoreShitheadInteractor(data []byte, pp presenter.ShitheadPresenter) (*ShitheadInteractor, error) {
	return restoreAndBuild[domain.Shithead](data, func(g *domain.Shithead) *ShitheadInteractor {
		return &ShitheadInteractor{GameBase: GameBase[interfaces.ShitheadGame]{Game: g}, pp: pp}
	})
}
