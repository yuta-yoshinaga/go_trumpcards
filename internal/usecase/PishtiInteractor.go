//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PishtiInteractorIF は Pişti インタラクターインタフェース。
type PishtiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextRound 次のゲームを開始する
	NextRound() string
	// Play 手札を場へ出す
	Play(cardIndex int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.PishtiConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.PishtiConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PishtiInteractor は Pişti インタラクター。
type PishtiInteractor struct {
	GameBase[interfaces.PishtiGame]
	pp presenter.PishtiPresenter
}

// NewPishtiInteractor コンストラクタ。
func NewPishtiInteractor(pg interfaces.PishtiGame, pp presenter.PishtiPresenter) *PishtiInteractor {
	mustNotNil("PishtiInteractor", map[string]any{"pg": pg, "pp": pp})
	return &PishtiInteractor{
		GameBase: GameBase[interfaces.PishtiGame]{Game: pg},
		pp:       pp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (pi *PishtiInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// NextRound 次のゲームを開始する。
func (pi *PishtiInteractor) NextRound() string {
	pi.Game.NextRound()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// Play 手札を場へ出す。
func (pi *PishtiInteractor) Play(cardIndex int) string {
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
func (pi *PishtiInteractor) GetConfig() domain.PishtiConfig { return pi.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (pi *PishtiInteractor) ResetWithConfig(config domain.PishtiConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.pp, config, pi.Game.SetConfig, pi.Reset)
}

// ActionLog 棋譜を出力する。
func (pi *PishtiInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// pishtiMaxCpuIterations は runCpuTurns の防御的な反復上限。
// 1 ゲームの総ターン数は高々 52 (山札) なので、これを大きく超える場合は
// CpuPlay が手番を進めていない可能性が高い。無限ループ防止のための保険。
const pishtiMaxCpuIterations = 1000

// runCpuTurns はゲームが終わるか人間の手番になるまで CPU ターンを回す。
func (pi *PishtiInteractor) runCpuTurns() {
	for i := 0; i < pishtiMaxCpuIterations; i++ {
		if pi.Game.GetGameEndFlag() || pi.Game.IsHumanTurn() {
			return
		}
		pi.Game.CpuPlay()
	}
}

// RestorePishtiInteractor deserialises JSON into a PishtiInteractor.
func RestorePishtiInteractor(data []byte, pp presenter.PishtiPresenter) (*PishtiInteractor, error) {
	return restoreAndBuild[domain.Pishti](data, func(g *domain.Pishti) *PishtiInteractor {
		return &PishtiInteractor{GameBase: GameBase[interfaces.PishtiGame]{Game: g}, pp: pp}
	})
}
