//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DilotiInteractorIF はディロティのインタラクターインタフェース。
type DilotiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.DilotiConfig) string
	// Play 1 手打つ (action は capture / declare / trail)
	Play(handIdx int, action string, tableIdxs, declIdxs []int, declValue int) string
	// NextRound 次の局へ進む
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.DilotiConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DilotiInteractor はディロティのインタラクター。
type DilotiInteractor struct {
	GameBase[interfaces.DilotiGame]
	dp presenter.DilotiPresenter
}

// NewDilotiInteractor コンストラクタ。
func NewDilotiInteractor(g interfaces.DilotiGame, dp presenter.DilotiPresenter) *DilotiInteractor {
	mustNotNil("DilotiInteractor", map[string]any{"g": g, "dp": dp})
	return &DilotiInteractor{GameBase: GameBase[interfaces.DilotiGame]{Game: g}, dp: dp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (di *DilotiInteractor) Reset() string {
	di.Game.Reset()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (di *DilotiInteractor) ResetWithConfig(config domain.DilotiConfig) string {
	return resetWithValidatedConfig(di.Game, di.dp, config, di.Game.SetConfig, di.Reset)
}

// Play 1 手打つ。
func (di *DilotiInteractor) Play(handIdx int, action string, tableIdxs, declIdxs []int, declValue int) string {
	if out, blocked := guardNotPlayable(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.PlayerPlay(handIdx, action, tableIdxs, declIdxs, declValue)
	if err == nil {
		di.runCpuTurns()
	}
	return di.dp.Output(di.Game, err)
}

// NextRound 次の局へ進む。
//
// **終局と区切り以外での呼び出しはドメインが弾く。** Diloti.NextRound が
// gameEndFlag とフェーズの両方を見ているので、ここで同じ検査は重ねない。
func (di *DilotiInteractor) NextRound() string {
	di.Game.NextRound()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// GetConfig 現在の設定を返す。
func (di *DilotiInteractor) GetConfig() domain.DilotiConfig { return di.Game.GetConfig() }

// Hint ヒントを出力する。
func (di *DilotiInteractor) Hint() string { return di.dp.HintOutput(di.Game) }

// ActionLog 棋譜を出力する。
func (di *DilotiInteractor) ActionLog() string { return di.dp.ActionLogOutput(di.Game) }

// dilotiMaxCpuIterations は runCpuTurns の防御的な反復上限。
const dilotiMaxCpuIterations = 1000

// runCpuTurns は人間の手番か、局の区切りか、終局まで CPU を回す。
func (di *DilotiInteractor) runCpuTurns() {
	for i := 0; i < dilotiMaxCpuIterations; i++ {
		if di.Game.GetGameEndFlag() || di.Game.IsHumanTurn() {
			return
		}
		if di.Game.GetPhase() != domain.DilotiPhasePlay {
			// 局の区切りでは人間が次へ進める操作を待つ。
			return
		}
		di.Game.CpuPlay()
	}
}

// RestoreDilotiInteractor deserialises JSON into an interactor.
func RestoreDilotiInteractor(data []byte, dp presenter.DilotiPresenter) (*DilotiInteractor, error) {
	return restoreAndBuild[domain.Diloti](data, func(g *domain.Diloti) *DilotiInteractor {
		return &DilotiInteractor{GameBase: GameBase[interfaces.DilotiGame]{Game: g}, dp: dp}
	})
}
