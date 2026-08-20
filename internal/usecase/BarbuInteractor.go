//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BarbuInteractorIF はバルブインタラクターインタフェース。
type BarbuInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextDeal 次のディール開始
	NextDeal() string
	// SelectContract ディーラーがコントラクトを選択する (trumpSuit は Trumps のみ)
	SelectContract(contract, trumpSuit int) string
	// Play 手札を出す (Dominoes では handIdx == -1 でパス)
	Play(handIdx int, tableIdxs []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.BarbuConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.BarbuConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BarbuInteractor はバルブインタラクター。
type BarbuInteractor struct {
	GameBase[interfaces.BarbuGame]
	bp presenter.BarbuPresenter
}

// NewBarbuInteractor コンストラクタ。
func NewBarbuInteractor(bg interfaces.BarbuGame, bp presenter.BarbuPresenter) *BarbuInteractor {
	mustNotNil("BarbuInteractor", map[string]any{"bg": bg, "bp": bp})
	return &BarbuInteractor{
		GameBase: GameBase[interfaces.BarbuGame]{Game: bg},
		bp:       bp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (bi *BarbuInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextDeal 次のディール開始。
func (bi *BarbuInteractor) NextDeal() string {
	if bi.Game.GetGameEndFlag() {
		return bi.bp.Output(bi.Game, nil)
	}
	bi.Game.NextDeal()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// SelectContract ディーラーがコントラクトを選択する。
func (bi *BarbuInteractor) SelectContract(contract, trumpSuit int) string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.SelectContract(contract, trumpSuit)
	if err == nil && !bi.Game.GetGameEndFlag() {
		bi.runCpuTurns()
	}
	return bi.bp.Output(bi.Game, err)
}

// Play 手札を出す。
func (bi *BarbuInteractor) Play(handIdx int, tableIdxs []int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerPlay(handIdx, tableIdxs)
	if err == nil && !bi.Game.GetGameEndFlag() {
		bi.runCpuTurns()
	}
	return bi.bp.Output(bi.Game, err)
}

// GetConfig 現在の設定を返す。
func (bi *BarbuInteractor) GetConfig() domain.BarbuConfig { return bi.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (bi *BarbuInteractor) ResetWithConfig(config domain.BarbuConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, config, bi.Game.SetConfig, bi.Reset)
}

// ActionLog 棋譜を出力する。
func (bi *BarbuInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// Hint ヒントを出力する。
func (bi *BarbuInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// barbuMaxCpuIterations は runCpuTurns の防御的な反復上限。
const barbuMaxCpuIterations = 1000

// runCpuTurns はゲーム終了・人間の手番・ディール終了のいずれかに到達するまで
// CPU ステップを回す。ディール終了 (BarbuPhaseDealEnd) では人間が次ディールへ
// 進める操作を待つため、ここでは自動進行しない。
func (bi *BarbuInteractor) runCpuTurns() {
	for i := 0; i < barbuMaxCpuIterations; i++ {
		if bi.Game.GetGameEndFlag() || bi.Game.IsHumanTurn() {
			return
		}
		if bi.Game.GetPhase() == domain.BarbuPhaseDealEnd {
			return
		}
		bi.Game.CpuPlay()
	}
}

// RestoreBarbuInteractor deserialises JSON into a BarbuInteractor.
func RestoreBarbuInteractor(data []byte, bp presenter.BarbuPresenter) (*BarbuInteractor, error) {
	return restoreAndBuild[domain.Barbu](data, func(g *domain.Barbu) *BarbuInteractor {
		return &BarbuInteractor{GameBase: GameBase[interfaces.BarbuGame]{Game: g}, bp: bp}
	})
}
