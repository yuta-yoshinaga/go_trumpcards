//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TichuInteractorIF ティチューインタラクターインタフェース
type TichuInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Declare 人間プレイヤーが宣言する (0=なし, 1=ティチュー, 2=グランド)
	Declare(declType int) string
	// Play 人間プレイヤーがカードを出す (または パスする)
	Play(indices []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.TichuConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.TichuConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TichuInteractor ティチューインタラクタークラス
type TichuInteractor struct {
	GameBase[interfaces.TichuGame]
	tgp presenter.TichuPresenter
}

// NewTichuInteractor コンストラクタ
func NewTichuInteractor(tg interfaces.TichuGame, tgp presenter.TichuPresenter) *TichuInteractor {
	mustNotNil("TichuInteractor", map[string]any{"tg": tg, "tgp": tgp})
	return &TichuInteractor{
		GameBase: GameBase[interfaces.TichuGame]{Game: tg},
		tgp:      tgp,
	}
}

// Reset ゲーム初期化
func (ti *TichuInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tgp.Output(ti.Game, nil)
}

// Declare 人間プレイヤーが宣言する
func (ti *TichuInteractor) Declare(declType int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tgp); blocked {
		return out
	}
	err := ti.Game.PlayerDeclare(declType)
	if err == nil && !ti.Game.GetGameEndFlag() {
		ti.runCpuTurns()
	}
	return ti.tgp.Output(ti.Game, err)
}

// Play 人間プレイヤーがカードを出す (または パスする)
func (ti *TichuInteractor) Play(indices []int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tgp); blocked {
		return out
	}
	err := ti.Game.PlayerPlay(indices)
	if err == nil && !ti.Game.GetGameEndFlag() && !ti.Game.HasPendingAction() {
		ti.runCpuTurns()
	}
	return ti.tgp.Output(ti.Game, err)
}

// GetConfig 現在の設定を返す
func (ti *TichuInteractor) GetConfig() domain.TichuConfig {
	return ti.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (ti *TichuInteractor) ResetWithConfig(config domain.TichuConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tgp, config, ti.Game.SetConfig, ti.Reset)
}

// ActionLog 棋譜を出力する
func (ti *TichuInteractor) ActionLog() string {
	return ti.tgp.ActionLogOutput(ti.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行する
func (ti *TichuInteractor) runCpuTurns() {
	runCpuTurnsCapped(ti.Game, ti.Game.CpuPlay)
}

// RestoreTichuInteractor deserialises JSON into a TichuInteractor.
func RestoreTichuInteractor(data []byte, tgp presenter.TichuPresenter) (*TichuInteractor, error) {
	return restoreAndBuild[domain.Tichu](data, func(g *domain.Tichu) *TichuInteractor {
		return &TichuInteractor{GameBase: GameBase[interfaces.TichuGame]{Game: g}, tgp: tgp}
	})
}
