package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PresidentInteractorIF プレジデントインタラクターインタフェース
type PresidentInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Play 人間プレイヤーがカードを出す (または パスする)
	Play(indices []int) string
	// Hint ヒント取得
	Hint() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.PresidentConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.PresidentConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PresidentInteractor プレジデントインタラクタークラス
type PresidentInteractor struct {
	GameBase[interfaces.PresidentGame]
	pp presenter.PresidentPresenter
}

// NewPresidentInteractor コンストラクタ
func NewPresidentInteractor(pg interfaces.PresidentGame, pp presenter.PresidentPresenter) *PresidentInteractor {
	mustNotNil("PresidentInteractor", map[string]any{"pg": pg, "pp": pp})
	return &PresidentInteractor{
		GameBase: GameBase[interfaces.PresidentGame]{Game: pg},
		pp:       pp,
	}
}

// Reset ゲーム初期化
func (pi *PresidentInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
func (pi *PresidentInteractor) Play(indices []int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerPlay(indices)
	if err == nil && !pi.Game.GetGameEndFlag() {
		pi.runCpuTurns()
	}
	return pi.pp.Output(pi.Game, err)
}

// Hint ヒント取得
func (pi *PresidentInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// GetConfig 現在の設定を返す
func (pi *PresidentInteractor) GetConfig() domain.PresidentConfig {
	return pi.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (pi *PresidentInteractor) ResetWithConfig(config domain.PresidentConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.pp, config, pi.Game.SetConfig, pi.Reset)
}

// ActionLog 棋譜を出力する
func (pi *PresidentInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (pi *PresidentInteractor) runCpuTurns() {
	for !pi.Game.GetGameEndFlag() && !pi.Game.IsHumanTurn() {
		pi.Game.CpuPlay()
	}
}

// RestorePresidentInteractor deserialises JSON into a PresidentInteractor.
func RestorePresidentInteractor(data []byte, pp presenter.PresidentPresenter) (*PresidentInteractor, error) {
	return restoreAndBuild[domain.President](data, func(g *domain.President) *PresidentInteractor {
		return &PresidentInteractor{GameBase: GameBase[interfaces.PresidentGame]{Game: g}, pp: pp}
	})
}
