//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CuarentaInteractorIF クアレンタインタラクターインタフェース。
type CuarentaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextRound 次のラウンド開始
	NextRound() string
	// Play 手札を出す (同ランク捕獲、なければ場に置く)
	Play(handIdx int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.CuarentaConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.CuarentaConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CuarentaInteractor クアレンタインタラクター。
type CuarentaInteractor struct {
	GameBase[interfaces.CuarentaGame]
	sp presenter.CuarentaPresenter
}

// NewCuarentaInteractor コンストラクタ。
func NewCuarentaInteractor(cg interfaces.CuarentaGame, sp presenter.CuarentaPresenter) *CuarentaInteractor {
	mustNotNil("CuarentaInteractor", map[string]any{"cg": cg, "sp": sp})
	return &CuarentaInteractor{
		GameBase: GameBase[interfaces.CuarentaGame]{Game: cg},
		sp:       sp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ci *CuarentaInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.sp.Output(ci.Game, nil)
}

// NextRound 次のラウンド開始。
func (ci *CuarentaInteractor) NextRound() string {
	if ci.Game.GetGameEndFlag() {
		return ci.sp.Output(ci.Game, nil)
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.sp.Output(ci.Game, nil)
}

// Play 手札を出す (捕獲または場に置く)。
func (ci *CuarentaInteractor) Play(handIdx int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.sp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(handIdx)
	if err == nil && !ci.Game.GetGameEndFlag() {
		ci.runCpuTurns()
	}
	return ci.sp.Output(ci.Game, err)
}

// GetConfig 現在の設定を返す。
func (ci *CuarentaInteractor) GetConfig() domain.CuarentaConfig { return ci.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (ci *CuarentaInteractor) ResetWithConfig(config domain.CuarentaConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.sp, config, ci.Game.SetConfig, ci.Reset)
}

// ActionLog 棋譜を出力する。
func (ci *CuarentaInteractor) ActionLog() string {
	return ci.sp.ActionLogOutput(ci.Game)
}

// cuarentaMaxCpuIterations は runCpuTurns の防御的な反復上限。
const cuarentaMaxCpuIterations = 1000

// runCpuTurns ゲームが終わるか人間の手番になるまで CPU ターンを回す。
// ラウンド境界に到達した場合は自動的に NextRound し、続行する。
func (ci *CuarentaInteractor) runCpuTurns() {
	for i := 0; i < cuarentaMaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() || ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
		// Stop at a non-Play phase (e.g. RoundEnd) so the player can see the
		// round result. The frontend advances explicitly via the NextRound
		// command rather than auto-skipping the score screen.
		if ci.Game.GetPhase() != int(domain.CuarentaPhasePlay) {
			return
		}
	}
}

// RestoreCuarentaInteractor deserialises JSON into a CuarentaInteractor.
func RestoreCuarentaInteractor(data []byte, sp presenter.CuarentaPresenter) (*CuarentaInteractor, error) {
	return restoreAndBuild[domain.Cuarenta](data, func(g *domain.Cuarenta) *CuarentaInteractor {
		return &CuarentaInteractor{GameBase: GameBase[interfaces.CuarentaGame]{Game: g}, sp: sp}
	})
}
