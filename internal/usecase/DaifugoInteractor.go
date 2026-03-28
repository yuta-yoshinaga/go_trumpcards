package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DaifugoInteractorIF 大富豪インタラクターインタフェース
type DaifugoInteractorIF interface {
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
	dg  interfaces.DaifugoGame
	dgp presenter.DaifugoPresenter
}

// NewDaifugoInteractor コンストラクタ
func NewDaifugoInteractor(dg interfaces.DaifugoGame, dgp presenter.DaifugoPresenter) *DaifugoInteractor {
	mustNotNil("DaifugoInteractor", map[string]any{"dg": dg, "dgp": dgp})
	return &DaifugoInteractor{
		dg:  dg,
		dgp: dgp,
	}
}

// Reset ゲーム初期化
func (di *DaifugoInteractor) Reset() string {
	di.dg.Reset()
	di.runCpuTurns()
	return di.dgp.Output(di.dg, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
// indices: 出すカードのインデックス。空の場合はパス。
func (di *DaifugoInteractor) Play(indices []int) string {
	if out, blocked := guardNotPlayable(di.dg, di.dgp); blocked {
		return out
	}
	err := di.dg.PlayerPlay(indices)
	if err == nil && !di.dg.GetGameEndFlag() && !di.dg.HasPendingAction() {
		di.runCpuTurns()
	}
	return di.dgp.Output(di.dg, err)
}

// GetConfig 現在の設定を返す
func (di *DaifugoInteractor) GetConfig() domain.DaifugoConfig {
	return di.dg.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (di *DaifugoInteractor) ResetWithConfig(config domain.DaifugoConfig) string {
	return resetWithValidatedConfig(di.dg, di.dgp, config, di.dg.SetConfig, di.Reset)
}

// Sort 手札ソートモードを変更
func (di *DaifugoInteractor) Sort(mode domain.DaifugoSortMode) string {
	return execAndPresent(di.dg, di.dgp, func() error { return di.dg.SortHumanHand(mode) })
}

// ActionLog 棋譜を出力する
func (di *DaifugoInteractor) ActionLog() string {
	return di.dgp.ActionLogOutput(di.dg)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (di *DaifugoInteractor) runCpuTurns() {
	for !di.dg.GetGameEndFlag() && !di.dg.IsHumanTurn() {
		di.dg.CpuPlay()
	}
}

// Snapshot serialises the game state to JSON for KV persistence.
func (di *DaifugoInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(di.dg)
}

// RestoreDaifugoInteractor deserialises JSON into a DaifugoInteractor.
func RestoreDaifugoInteractor(data []byte, dgp presenter.DaifugoPresenter) (*DaifugoInteractor, error) {
	var dg domain.Daifugo
	if err := json.Unmarshal(data, &dg); err != nil {
		return nil, err
	}
	return &DaifugoInteractor{dg: &dg, dgp: dgp}, nil
}
