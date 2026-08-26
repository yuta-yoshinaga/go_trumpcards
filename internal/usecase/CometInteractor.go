//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CometInteractorIF はコメットのインタラクターインタフェース。
type CometInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.CometConfig) string
	// Play 手札を 1 枚出す
	Play(handIdx int) string
	// Pass パスする
	Pass() string
	// NextRound 次の局へ進む
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.CometConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CometInteractor はコメットのインタラクター。
type CometInteractor struct {
	GameBase[interfaces.CometGame]
	cp presenter.CometPresenter
}

// NewCometInteractor コンストラクタ。
func NewCometInteractor(g interfaces.CometGame, cp presenter.CometPresenter) *CometInteractor {
	mustNotNil("CometInteractor", map[string]any{"g": g, "cp": cp})
	return &CometInteractor{GameBase: GameBase[interfaces.CometGame]{Game: g}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ci *CometInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ci *CometInteractor) ResetWithConfig(config domain.CometConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, config, ci.Game.SetConfig, ci.Reset)
}

// Play 手札を 1 枚出す。
func (ci *CometInteractor) Play(handIdx int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(handIdx)
	if err == nil {
		ci.runCpuTurns()
	}
	return ci.cp.Output(ci.Game, err)
}

// Pass パスする。
func (ci *CometInteractor) Pass() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerPass()
	if err == nil {
		ci.runCpuTurns()
	}
	return ci.cp.Output(ci.Game, err)
}

// NextRound 次の局へ進む。
//
// **終局と区切り以外での呼び出しはドメインが弾く。** Comet.NextRound が
// gameEndFlag とフェーズの両方を見ているので、ここで同じ検査は重ねない。
func (ci *CometInteractor) NextRound() string {
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を返す。
func (ci *CometInteractor) GetConfig() domain.CometConfig { return ci.Game.GetConfig() }

// Hint ヒントを出力する。
func (ci *CometInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する。
func (ci *CometInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// cometMaxCpuIterations は runCpuTurns の防御的な反復上限。
const cometMaxCpuIterations = 1000

// runCpuTurns は人間の手番か、局の区切りか、終局まで CPU を回す。
func (ci *CometInteractor) runCpuTurns() {
	for i := 0; i < cometMaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() || ci.Game.IsHumanTurn() {
			return
		}
		if ci.Game.GetPhase() != domain.CometPhasePlay {
			// 局の区切りでは人間が次へ進める操作を待つ。
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestoreCometInteractor deserialises JSON into an interactor.
func RestoreCometInteractor(data []byte, cp presenter.CometPresenter) (*CometInteractor, error) {
	return restoreAndBuild[domain.Comet](data, func(g *domain.Comet) *CometInteractor {
		return &CometInteractor{GameBase: GameBase[interfaces.CometGame]{Game: g}, cp: cp}
	})
}
