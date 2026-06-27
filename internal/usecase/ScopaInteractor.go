//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ScopaInteractorIF スコパインタラクターインタフェース。
type ScopaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextRound 次のラウンド開始
	NextRound() string
	// Play 手札を出す (tableIdxs が空なら場に置く、それ以外は捕獲)
	Play(handIdx int, tableIdxs []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.ScopaConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.ScopaConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ScopaInteractor スコパインタラクター。
type ScopaInteractor struct {
	GameBase[interfaces.ScopaGame]
	sp presenter.ScopaPresenter
}

// NewScopaInteractor コンストラクタ。
func NewScopaInteractor(sg interfaces.ScopaGame, sp presenter.ScopaPresenter) *ScopaInteractor {
	mustNotNil("ScopaInteractor", map[string]any{"sg": sg, "sp": sp})
	return &ScopaInteractor{
		GameBase: GameBase[interfaces.ScopaGame]{Game: sg},
		sp:       sp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)
func (si *ScopaInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound 次のラウンド開始。
func (si *ScopaInteractor) NextRound() string {
	if si.Game.GetGameEndFlag() {
		return si.sp.Output(si.Game, nil)
	}
	si.Game.NextRound()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// Play 手札を出す (捕獲または場に置く)。
func (si *ScopaInteractor) Play(handIdx int, tableIdxs []int) string {
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
func (si *ScopaInteractor) GetConfig() domain.ScopaConfig { return si.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (si *ScopaInteractor) ResetWithConfig(config domain.ScopaConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, config, si.Game.SetConfig, si.Reset)
}

// ActionLog 棋譜を出力する。
func (si *ScopaInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// Hint ヒントを出力する。
func (si *ScopaInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// scopaMaxCpuIterations は runCpuTurns の防御的な反復上限。
// 通常 1 ラウンドで CPU が動くのは数十回で、1000 を超えるなら CpuPlay または
// NextRound が手番を進めていない可能性が高い。
const scopaMaxCpuIterations = 1000

// runCpuTurns ゲームが終わるか人間の手番になるまで CPU ターンを回す。
// ラウンド境界に到達した場合は自動的に NextRound し、続行する。
func (si *ScopaInteractor) runCpuTurns() {
	for i := 0; i < scopaMaxCpuIterations; i++ {
		if si.Game.GetGameEndFlag() || si.Game.IsHumanTurn() {
			return
		}
		si.Game.CpuPlay()
		// Round 終了後に NextRound を自動実行 (ゲーム終了で無ければ)。
		if !si.Game.GetGameEndFlag() && si.Game.GetPhase() != domain.ScopaPhasePlayerTurn {
			si.Game.NextRound()
		}
	}
}

// RestoreScopaInteractor deserialises JSON into a ScopaInteractor.
func RestoreScopaInteractor(data []byte, sp presenter.ScopaPresenter) (*ScopaInteractor, error) {
	return restoreAndBuild[domain.Scopa](data, func(g *domain.Scopa) *ScopaInteractor {
		return &ScopaInteractor{GameBase: GameBase[interfaces.ScopaGame]{Game: g}, sp: sp}
	})
}
