//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SutdaInteractorIF はソッタのインタラクターインタフェース。
type SutdaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.SutdaConfig) string
	// Action 人間が 1 手打つ (call / raise / fold)
	Action(action string) string
	// NextHand 次のハンドへ進む
	NextHand() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.SutdaConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SutdaInteractor はソッタのインタラクター。
type SutdaInteractor struct {
	GameBase[interfaces.SutdaGame]
	sp presenter.SutdaPresenter
}

// NewSutdaInteractor コンストラクタ。
func NewSutdaInteractor(g interfaces.SutdaGame, sp presenter.SutdaPresenter) *SutdaInteractor {
	mustNotNil("SutdaInteractor", map[string]any{"g": g, "sp": sp})
	return &SutdaInteractor{GameBase: GameBase[interfaces.SutdaGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (si *SutdaInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (si *SutdaInteractor) ResetWithConfig(config domain.SutdaConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, config, si.Game.SetConfig, si.Reset)
}

// Action 人間が 1 手打つ。
func (si *SutdaInteractor) Action(action string) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.PlayerAction(action)
	if err == nil {
		si.runCpuTurns()
	}
	return si.sp.Output(si.Game, err)
}

// NextHand 次のハンドへ進む。
func (si *SutdaInteractor) NextHand() string {
	if si.Game.GetGameEndFlag() {
		return si.sp.Output(si.Game, nil)
	}
	si.Game.NextHand()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を返す。
func (si *SutdaInteractor) GetConfig() domain.SutdaConfig { return si.Game.GetConfig() }

// Hint ヒントを出力する。
func (si *SutdaInteractor) Hint() string { return si.sp.HintOutput(si.Game) }

// ActionLog 棋譜を出力する。
func (si *SutdaInteractor) ActionLog() string { return si.sp.ActionLogOutput(si.Game) }

// sutdaMaxCpuIterations は runCpuTurns の防御的な反復上限。
const sutdaMaxCpuIterations = 1000

// runCpuTurns は人間の番か、ショーダウンか、終局まで CPU を回す。
func (si *SutdaInteractor) runCpuTurns() {
	for i := 0; i < sutdaMaxCpuIterations; i++ {
		if si.Game.GetGameEndFlag() || si.Game.IsHumanTurn() {
			return
		}
		if si.Game.GetPhase() != domain.SutdaPhaseBet {
			// ショーダウンでは人間が次へ進める操作を待つ。
			return
		}
		si.Game.CpuAction()
	}
}

// RestoreSutdaInteractor deserialises JSON into an interactor.
func RestoreSutdaInteractor(data []byte, sp presenter.SutdaPresenter) (*SutdaInteractor, error) {
	return restoreAndBuild[domain.Sutda](data, func(g *domain.Sutda) *SutdaInteractor {
		return &SutdaInteractor{GameBase: GameBase[interfaces.SutdaGame]{Game: g}, sp: sp}
	})
}
