//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DoudizhuInteractorIF 斗地主インタラクターインタフェース
type DoudizhuInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bid 人間プレイヤーがビッドする
	Bid(value int) string
	// Play 人間プレイヤーがカードを出す (または パスする)
	Play(indices []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.DoudizhuConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.DoudizhuConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DoudizhuInteractor 斗地主インタラクタークラス
type DoudizhuInteractor struct {
	GameBase[interfaces.DoudizhuGame]
	dgp presenter.DoudizhuPresenter
}

// NewDoudizhuInteractor コンストラクタ
func NewDoudizhuInteractor(dg interfaces.DoudizhuGame, dgp presenter.DoudizhuPresenter) *DoudizhuInteractor {
	mustNotNil("DoudizhuInteractor", map[string]any{"dg": dg, "dgp": dgp})
	return &DoudizhuInteractor{
		GameBase: GameBase[interfaces.DoudizhuGame]{Game: dg},
		dgp:      dgp,
	}
}

// Reset ゲーム初期化
func (di *DoudizhuInteractor) Reset() string {
	di.Game.Reset()
	di.runCpuTurns()
	return di.dgp.Output(di.Game, nil)
}

// Bid 人間プレイヤーがビッドする
func (di *DoudizhuInteractor) Bid(value int) string {
	if out, blocked := guardNotPlayable(di.Game, di.dgp); blocked {
		return out
	}
	err := di.Game.PlayerBid(value)
	if err == nil && !di.Game.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dgp.Output(di.Game, err)
}

// Play 人間プレイヤーがカードを出す (または パスする)
func (di *DoudizhuInteractor) Play(indices []int) string {
	if out, blocked := guardNotPlayable(di.Game, di.dgp); blocked {
		return out
	}
	err := di.Game.PlayerPlay(indices)
	if err == nil && !di.Game.GetGameEndFlag() && !di.Game.HasPendingAction() {
		di.runCpuTurns()
	}
	return di.dgp.Output(di.Game, err)
}

// GetConfig 現在の設定を返す
func (di *DoudizhuInteractor) GetConfig() domain.DoudizhuConfig {
	return di.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (di *DoudizhuInteractor) ResetWithConfig(config domain.DoudizhuConfig) string {
	return resetWithValidatedConfig(di.Game, di.dgp, config, di.Game.SetConfig, di.Reset)
}

// ActionLog 棋譜を出力する
func (di *DoudizhuInteractor) ActionLog() string {
	return di.dgp.ActionLogOutput(di.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (di *DoudizhuInteractor) runCpuTurns() {
	runCpuTurnsCapped(di.Game, di.Game.CpuPlay)
}

// RestoreDoudizhuInteractor deserialises JSON into a DoudizhuInteractor.
func RestoreDoudizhuInteractor(data []byte, dgp presenter.DoudizhuPresenter) (*DoudizhuInteractor, error) {
	return restoreAndBuild[domain.Doudizhu](data, func(g *domain.Doudizhu) *DoudizhuInteractor {
		return &DoudizhuInteractor{GameBase: GameBase[interfaces.DoudizhuGame]{Game: g}, dgp: dgp}
	})
}
