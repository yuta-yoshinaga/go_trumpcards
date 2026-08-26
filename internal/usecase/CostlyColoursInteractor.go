//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CostlyColoursInteractorIF はコストリー・カラーズのインタラクターインタフェース。
type CostlyColoursInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.CostlyColoursConfig) string
	// Mog 交換に応じるかを決める
	Mog(accept bool) string
	// Play 手札を 1 枚出す
	Play(handIdx int) string
	// NextDeal 次のディールへ進む
	NextDeal() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.CostlyColoursConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CostlyColoursInteractor はコストリー・カラーズのインタラクター。
type CostlyColoursInteractor struct {
	GameBase[interfaces.CostlyColoursGame]
	cp presenter.CostlyColoursPresenter
}

// NewCostlyColoursInteractor コンストラクタ。
func NewCostlyColoursInteractor(g interfaces.CostlyColoursGame, cp presenter.CostlyColoursPresenter) *CostlyColoursInteractor {
	mustNotNil("CostlyColoursInteractor", map[string]any{"g": g, "cp": cp})
	return &CostlyColoursInteractor{GameBase: GameBase[interfaces.CostlyColoursGame]{Game: g}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ci *CostlyColoursInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ci *CostlyColoursInteractor) ResetWithConfig(config domain.CostlyColoursConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, config, ci.Game.SetConfig, ci.Reset)
}

// Mog 交換に応じるかを決める。
func (ci *CostlyColoursInteractor) Mog(accept bool) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerMog(accept)
	if err == nil {
		ci.runCpuTurns()
	}
	return ci.cp.Output(ci.Game, err)
}

// Play 手札を 1 枚出す。
func (ci *CostlyColoursInteractor) Play(handIdx int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(handIdx)
	if err == nil {
		ci.runCpuTurns()
	}
	return ci.cp.Output(ci.Game, err)
}

// NextDeal 次のディールへ進む。
//
// **終局とショー以外での呼び出しはドメインが弾く。** CostlyColours.NextDeal が
// gameEndFlag とフェーズの両方を見ているので、ここで同じ検査は重ねない。
func (ci *CostlyColoursInteractor) NextDeal() string {
	ci.Game.NextDeal()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を返す。
func (ci *CostlyColoursInteractor) GetConfig() domain.CostlyColoursConfig {
	return ci.Game.GetConfig()
}

// Hint ヒントを出力する。
func (ci *CostlyColoursInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する。
func (ci *CostlyColoursInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// costlyColoursMaxCpuIterations は runCpuTurns の防御的な反復上限。
const costlyColoursMaxCpuIterations = 1000

// runCpuTurns は人間の手番か、ショーか、終局まで CPU を回す。
func (ci *CostlyColoursInteractor) runCpuTurns() {
	for i := 0; i < costlyColoursMaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() || ci.Game.IsHumanTurn() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase != domain.CostlyColoursPhaseMog && phase != domain.CostlyColoursPhasePlay {
			// ショーでは人間が次へ進める操作を待つ。
			return
		}
		ci.Game.CpuAct()
	}
}

// RestoreCostlyColoursInteractor deserialises JSON into an interactor.
func RestoreCostlyColoursInteractor(data []byte, cp presenter.CostlyColoursPresenter) (*CostlyColoursInteractor, error) {
	return restoreAndBuild[domain.CostlyColours](data, func(g *domain.CostlyColours) *CostlyColoursInteractor {
		return &CostlyColoursInteractor{
			GameBase: GameBase[interfaces.CostlyColoursGame]{Game: g}, cp: cp,
		}
	})
}
