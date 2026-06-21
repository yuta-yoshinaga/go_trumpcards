//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ScoponeInteractorIF スコポーネインタラクターインタフェース。
type ScoponeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextRound 次のラウンド開始
	NextRound() string
	// Play 手札を出す (tableIdxs が空なら場に置く、それ以外は捕獲)
	Play(handIdx int, tableIdxs []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.ScoponeConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.ScoponeConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ScoponeInteractor スコポーネインタラクター。
type ScoponeInteractor struct {
	GameBase[interfaces.ScoponeGame]
	sp presenter.ScoponePresenter
}

// NewScoponeInteractor コンストラクタ。
func NewScoponeInteractor(sg interfaces.ScoponeGame, sp presenter.ScoponePresenter) *ScoponeInteractor {
	mustNotNil("ScoponeInteractor", map[string]any{"sg": sg, "sp": sp})
	return &ScoponeInteractor{
		GameBase: GameBase[interfaces.ScoponeGame]{Game: sg},
		sp:       sp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)
func (si *ScoponeInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound 次のラウンド開始。
func (si *ScoponeInteractor) NextRound() string {
	if si.Game.GetGameEndFlag() {
		return si.sp.Output(si.Game, nil)
	}
	si.Game.NextRound()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// Play 手札を出す (捕獲または場に置く)。
func (si *ScoponeInteractor) Play(handIdx int, tableIdxs []int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.PlayerPlay(handIdx, tableIdxs)
	if err == nil && !si.Game.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.sp.Output(si.Game, err)
}

// GetConfig 現在の設定を返す。
func (si *ScoponeInteractor) GetConfig() domain.ScoponeConfig { return si.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (si *ScoponeInteractor) ResetWithConfig(config domain.ScoponeConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, config, si.Game.SetConfig, si.Reset)
}

// ActionLog 棋譜を出力する。
func (si *ScoponeInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// scoponeMaxCpuIterations は runCpuTurns の防御的な反復上限。
const scoponeMaxCpuIterations = 2000

// runCpuTurns はプレイヤーターン中に CPU の手番を進める。
// 人間の手番、ラウンド終了、ゲーム終了のいずれかに達したら停止する。
// ラウンド終了は自動進行せず、明示的な NextRound を待つ (NextRound が再度この
// ループを回す)。
func (si *ScoponeInteractor) runCpuTurns() {
	for guard := 0; guard < scoponeMaxCpuIterations; guard++ {
		if si.Game.GetGameEndFlag() {
			return
		}
		if si.Game.GetPhase() == domain.ScoponePhasePlayerTurn && !si.Game.IsHumanTurn() {
			si.Game.CpuPlay()
			continue
		}
		// 人間の手番・ラウンド終了・ゲーム終了 → 停止。
		return
	}
}

// RestoreScoponeInteractor deserialises JSON into a ScoponeInteractor.
func RestoreScoponeInteractor(data []byte, sp presenter.ScoponePresenter) (*ScoponeInteractor, error) {
	return restoreAndBuild[domain.Scopone](data, func(g *domain.Scopone) *ScoponeInteractor {
		return &ScoponeInteractor{GameBase: GameBase[interfaces.ScoponeGame]{Game: g}, sp: sp}
	})
}
