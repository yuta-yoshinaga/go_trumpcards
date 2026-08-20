//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ZhengInteractorIF 争上游インタラクターインタフェース
type ZhengInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Play 人間プレイヤーがカードを出す (または パスする)
	Play(indices []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.ZhengConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.ZhengConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ZhengInteractor 争上游インタラクタークラス
type ZhengInteractor struct {
	GameBase[interfaces.ZhengGame]
	zp presenter.ZhengPresenter
}

// NewZhengInteractor コンストラクタ
func NewZhengInteractor(zg interfaces.ZhengGame, zp presenter.ZhengPresenter) *ZhengInteractor {
	mustNotNil("ZhengInteractor", map[string]any{"zg": zg, "zp": zp})
	return &ZhengInteractor{
		GameBase: GameBase[interfaces.ZhengGame]{Game: zg},
		zp:       zp,
	}
}

// Reset ゲーム初期化
func (zi *ZhengInteractor) Reset() string {
	zi.Game.Reset()
	zi.runCpuTurns()
	return zi.zp.Output(zi.Game, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
func (zi *ZhengInteractor) Play(indices []int) string {
	if out, blocked := guardNotPlayable(zi.Game, zi.zp); blocked {
		return out
	}
	err := zi.Game.PlayerPlay(indices)
	if err == nil && !zi.Game.GetGameEndFlag() && !zi.Game.HasPendingAction() {
		zi.runCpuTurns()
	}
	return zi.zp.Output(zi.Game, err)
}

// GetConfig 現在の設定を返す
func (zi *ZhengInteractor) GetConfig() domain.ZhengConfig {
	return zi.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (zi *ZhengInteractor) ResetWithConfig(config domain.ZhengConfig) string {
	return resetWithValidatedConfig(zi.Game, zi.zp, config, zi.Game.SetConfig, zi.Reset)
}

// ActionLog 棋譜を出力する
func (zi *ZhengInteractor) ActionLog() string {
	return zi.zp.ActionLogOutput(zi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (zi *ZhengInteractor) runCpuTurns() {
	runCpuTurnsCapped(zi.Game, zi.Game.CpuPlay)
}

// RestoreZhengInteractor deserialises JSON into a ZhengInteractor.
func RestoreZhengInteractor(data []byte, zp presenter.ZhengPresenter) (*ZhengInteractor, error) {
	return restoreAndBuild[domain.Zheng](data, func(g *domain.Zheng) *ZhengInteractor {
		return &ZhengInteractor{GameBase: GameBase[interfaces.ZhengGame]{Game: g}, zp: zp}
	})
}
