package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DaifugoInteractorIF 大富豪インタラクターインタフェース
type DaifugoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Play 人間プレイヤーがカードを出す (または パスする)
	Play(indices []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.DaifugoConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.DaifugoConfig
	// Sort 手札ソートモードを変更
	Sort(mode domain.DaifugoSortMode) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DaifugoInteractor 大富豪インタラクタークラス
type DaifugoInteractor struct {
	GameBase[interfaces.DaifugoGame]
	dgp presenter.DaifugoPresenter
}

// NewDaifugoInteractor コンストラクタ
func NewDaifugoInteractor(dg interfaces.DaifugoGame, dgp presenter.DaifugoPresenter) *DaifugoInteractor {
	mustNotNil("DaifugoInteractor", map[string]any{"dg": dg, "dgp": dgp})
	return &DaifugoInteractor{
		GameBase: GameBase[interfaces.DaifugoGame]{Game: dg},
		dgp:      dgp,
	}
}

// Reset ゲーム初期化
func (di *DaifugoInteractor) Reset() string {
	di.Game.Reset()
	di.runCpuTurns()
	return di.dgp.Output(di.Game, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
// indices: 出すカードのインデックス。空の場合はパス。
func (di *DaifugoInteractor) Play(indices []int) string {
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
func (di *DaifugoInteractor) GetConfig() domain.DaifugoConfig {
	return di.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (di *DaifugoInteractor) ResetWithConfig(config domain.DaifugoConfig) string {
	return resetWithValidatedConfig(di.Game, di.dgp, config, di.Game.SetConfig, di.Reset)
}

// Sort 手札ソートモードを変更
func (di *DaifugoInteractor) Sort(mode domain.DaifugoSortMode) string {
	return execAndPresent(di.Game, di.dgp, func() error { return di.Game.SortHumanHand(mode) })
}

// ActionLog 棋譜を出力する
func (di *DaifugoInteractor) ActionLog() string {
	return di.dgp.ActionLogOutput(di.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (di *DaifugoInteractor) runCpuTurns() {
	runCpuTurnsCapped(di.Game, di.Game.CpuPlay)
}

// RestoreDaifugoInteractor deserialises JSON into a DaifugoInteractor.
func RestoreDaifugoInteractor(data []byte, dgp presenter.DaifugoPresenter) (*DaifugoInteractor, error) {
	return restoreAndBuild[domain.Daifugo](data, func(g *domain.Daifugo) *DaifugoInteractor {
		return &DaifugoInteractor{GameBase: GameBase[interfaces.DaifugoGame]{Game: g}, dgp: dgp}
	})
}
