//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CirullaInteractorIF はチルッラのインタラクターインタフェース。
type CirullaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.CirullaConfig) string
	// Play 手札を出す (captureIdxs が空なら場に置く)
	Play(handIdx int, captureIdxs []int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.CirullaConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CirullaInteractor はチルッラのインタラクター。
type CirullaInteractor struct {
	GameBase[interfaces.CirullaGame]
	cp presenter.CirullaPresenter
}

// NewCirullaInteractor コンストラクタ。
func NewCirullaInteractor(g interfaces.CirullaGame, cp presenter.CirullaPresenter) *CirullaInteractor {
	mustNotNil("CirullaInteractor", map[string]any{"g": g, "cp": cp})
	return &CirullaInteractor{GameBase: GameBase[interfaces.CirullaGame]{Game: g}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ci *CirullaInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ci *CirullaInteractor) ResetWithConfig(config domain.CirullaConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, config, ci.Game.SetConfig, ci.Reset)
}

// Play 手札を出す。
func (ci *CirullaInteractor) Play(handIdx int, captureIdxs []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(handIdx, captureIdxs)
	if err == nil {
		ci.runCpuTurns()
	}
	return ci.cp.Output(ci.Game, err)
}

// NextRound 次のラウンドへ進む。
//
// **終局と区切り以外での呼び出しはドメインが弾く。** Cirulla.NextRound が
// gameEndFlag とフェーズの両方を見ているので、ここで同じ検査を重ねると
// 到達しない分岐が増えるだけになる。
func (ci *CirullaInteractor) NextRound() string {
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を返す。
func (ci *CirullaInteractor) GetConfig() domain.CirullaConfig { return ci.Game.GetConfig() }

// Hint ヒントを出力する。
func (ci *CirullaInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する。
func (ci *CirullaInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// cirullaMaxCpuIterations は runCpuTurns の防御的な反復上限。
const cirullaMaxCpuIterations = 1000

// runCpuTurns は人間の手番か、ラウンド終了か、終局まで CPU を回す。
func (ci *CirullaInteractor) runCpuTurns() {
	for i := 0; i < cirullaMaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() || ci.Game.IsHumanTurn() {
			return
		}
		if ci.Game.GetPhase() != domain.CirullaPhasePlay {
			// ラウンド終了では人間が次へ進める操作を待つ。
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestoreCirullaInteractor deserialises JSON into an interactor.
func RestoreCirullaInteractor(data []byte, cp presenter.CirullaPresenter) (*CirullaInteractor, error) {
	return restoreAndBuild[domain.Cirulla](data, func(g *domain.Cirulla) *CirullaInteractor {
		return &CirullaInteractor{GameBase: GameBase[interfaces.CirullaGame]{Game: g}, cp: cp}
	})
}
