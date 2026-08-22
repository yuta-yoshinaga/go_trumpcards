//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RistikontraInteractorIF はリスティコントラ インタラクターインタフェース。
type RistikontraInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextRound 次のゲームを開始する
	NextRound() string
	// Play 手札を場へ出す
	Play(cardIndex int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.RistikontraConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.RistikontraConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// RistikontraInteractor はリスティコントラ インタラクター。
type RistikontraInteractor struct {
	GameBase[interfaces.RistikontraGame]
	pp presenter.RistikontraPresenter
}

// NewRistikontraInteractor コンストラクタ。
func NewRistikontraInteractor(pg interfaces.RistikontraGame, pp presenter.RistikontraPresenter) *RistikontraInteractor {
	mustNotNil("RistikontraInteractor", map[string]any{"pg": pg, "pp": pp})
	return &RistikontraInteractor{
		GameBase: GameBase[interfaces.RistikontraGame]{Game: pg},
		pp:       pp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (pi *RistikontraInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// NextRound 次のゲームを開始する。
func (pi *RistikontraInteractor) NextRound() string {
	pi.Game.NextRound()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// Play 手札を場へ出す。
func (pi *RistikontraInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerPlay(cardIndex)
	if err == nil && !pi.Game.GetGameEndFlag() {
		pi.runCpuTurns()
	}
	return pi.pp.Output(pi.Game, err)
}

// GetConfig 現在の設定を返す。
func (pi *RistikontraInteractor) GetConfig() domain.RistikontraConfig { return pi.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (pi *RistikontraInteractor) ResetWithConfig(config domain.RistikontraConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.pp, config, pi.Game.SetConfig, pi.Reset)
}

// ActionLog 棋譜を出力する。
func (pi *RistikontraInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// ristikontraMaxCpuIterations は runCpuTurns の防御的な反復上限。
// 1 ゲームの総ターン数は高々 52 (山札) なので、これを大きく超える場合は
// CpuPlay が手番を進めていない可能性が高い。無限ループ防止のための保険。
const ristikontraMaxCpuIterations = 1000

// runCpuTurns はゲームが終わるか人間の手番になるまで CPU ターンを回す。
func (pi *RistikontraInteractor) runCpuTurns() {
	for i := 0; i < ristikontraMaxCpuIterations; i++ {
		if pi.Game.GetGameEndFlag() || pi.Game.IsHumanTurn() {
			return
		}
		pi.Game.CpuPlay()
	}
}

// RestoreRistikontraInteractor deserialises JSON into a RistikontraInteractor.
func RestoreRistikontraInteractor(data []byte, pp presenter.RistikontraPresenter) (*RistikontraInteractor, error) {
	return restoreAndBuild[domain.Ristikontra](data, func(g *domain.Ristikontra) *RistikontraInteractor {
		return &RistikontraInteractor{GameBase: GameBase[interfaces.RistikontraGame]{Game: g}, pp: pp}
	})
}
