//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// EscobaInteractorIF エスコバインタラクターインタフェース。
type EscobaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextRound 次のラウンド開始
	NextRound() string
	// Play 手札を出す (tableIdxs が空なら場に置く、それ以外は捕獲)
	Play(handIdx int, tableIdxs []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.EscobaConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.EscobaConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// EscobaInteractor エスコバインタラクター。
type EscobaInteractor struct {
	GameBase[interfaces.EscobaGame]
	sp presenter.EscobaPresenter
}

// NewEscobaInteractor コンストラクタ。
func NewEscobaInteractor(eg interfaces.EscobaGame, sp presenter.EscobaPresenter) *EscobaInteractor {
	mustNotNil("EscobaInteractor", map[string]any{"eg": eg, "sp": sp})
	return &EscobaInteractor{
		GameBase: GameBase[interfaces.EscobaGame]{Game: eg},
		sp:       sp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)
func (ei *EscobaInteractor) Reset() string {
	ei.Game.Reset()
	ei.runCpuTurns()
	return ei.sp.Output(ei.Game, nil)
}

// NextRound 次のラウンド開始。
func (ei *EscobaInteractor) NextRound() string {
	if ei.Game.GetGameEndFlag() {
		return ei.sp.Output(ei.Game, nil)
	}
	ei.Game.NextRound()
	ei.runCpuTurns()
	return ei.sp.Output(ei.Game, nil)
}

// Play 手札を出す (捕獲または場に置く)。
func (ei *EscobaInteractor) Play(handIdx int, tableIdxs []int) string {
	if out, blocked := guardNotPlayable(ei.Game, ei.sp); blocked {
		return out
	}
	err := ei.Game.PlayerPlay(handIdx, tableIdxs)
	if err == nil && !ei.Game.GetGameEndFlag() {
		ei.runCpuTurns()
	}
	return ei.sp.Output(ei.Game, err)
}

// GetConfig 現在の設定を返す。
func (ei *EscobaInteractor) GetConfig() domain.EscobaConfig { return ei.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (ei *EscobaInteractor) ResetWithConfig(config domain.EscobaConfig) string {
	return resetWithValidatedConfig(ei.Game, ei.sp, config, ei.Game.SetConfig, ei.Reset)
}

// ActionLog 棋譜を出力する。
func (ei *EscobaInteractor) ActionLog() string {
	return ei.sp.ActionLogOutput(ei.Game)
}

// escobaMaxCpuIterations は runCpuTurns の防御的な反復上限。
const escobaMaxCpuIterations = 2000

// runCpuTurns はプレイヤーターン中に CPU の手番を進める。
// 人間の手番、ラウンド終了、ゲーム終了のいずれかに達したら停止する。
// ラウンド終了は自動進行せず、明示的な NextRound を待つ (NextRound が再度この
// ループを回す)。
func (ei *EscobaInteractor) runCpuTurns() {
	for guard := 0; guard < escobaMaxCpuIterations; guard++ {
		if ei.Game.GetGameEndFlag() {
			return
		}
		if ei.Game.GetPhase() == domain.EscobaPhasePlayerTurn && !ei.Game.IsHumanTurn() {
			ei.Game.CpuPlay()
			continue
		}
		// 人間の手番・ラウンド終了・ゲーム終了 → 停止。
		return
	}
}

// RestoreEscobaInteractor deserialises JSON into an EscobaInteractor.
func RestoreEscobaInteractor(data []byte, sp presenter.EscobaPresenter) (*EscobaInteractor, error) {
	return restoreAndBuild[domain.Escoba](data, func(g *domain.Escoba) *EscobaInteractor {
		return &EscobaInteractor{GameBase: GameBase[interfaces.EscobaGame]{Game: g}, sp: sp}
	})
}
