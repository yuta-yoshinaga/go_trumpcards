//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HandAndFootInteractorIF ハンドアンドフットインタラクターインタフェース
type HandAndFootInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.HandAndFootConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札の山を取る
	DrawFromDiscard(naturalPairIndices []int) string
	// Meld メルドを出す
	Meld(meldGroups [][]int) string
	// SkipMeld メルドフェーズをスキップ
	SkipMeld() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// GoOut 上がる
	GoOut() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.HandAndFootConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// HandAndFootInteractor ハンドアンドフットインタラクタークラス
type HandAndFootInteractor struct {
	GameBase[interfaces.HandAndFootGame]
	gp presenter.HandAndFootPresenter
}

// NewHandAndFootInteractor コンストラクタ
func NewHandAndFootInteractor(g interfaces.HandAndFootGame, gp presenter.HandAndFootPresenter) *HandAndFootInteractor {
	mustNotNil("HandAndFootInteractor", map[string]any{"g": g, "gp": gp})
	return &HandAndFootInteractor{GameBase: GameBase[interfaces.HandAndFootGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *HandAndFootInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *HandAndFootInteractor) ResetWithConfig(cfg domain.HandAndFootConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *HandAndFootInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDrawFromStock()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawFromDiscard 捨て札の山を取る
func (ci *HandAndFootInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDrawFromDiscard(naturalPairIndices)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Meld メルドを出す
func (ci *HandAndFootInteractor) Meld(meldGroups [][]int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerMeld(meldGroups)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// SkipMeld メルドフェーズをスキップ
func (ci *HandAndFootInteractor) SkipMeld() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerSkipMeld()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨てる
func (ci *HandAndFootInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDiscard(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// GoOut 上がる
func (ci *HandAndFootInteractor) GoOut() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerGoOut()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *HandAndFootInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *HandAndFootInteractor) GetConfig() domain.HandAndFootConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *HandAndFootInteractor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *HandAndFootInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行
func (ci *HandAndFootInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.HandAndFootPhaseRoundEnd || phase == domain.HandAndFootPhaseGameEnd || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreHandAndFootInteractor deserialises JSON into a HandAndFootInteractor.
func RestoreHandAndFootInteractor(data []byte, gp presenter.HandAndFootPresenter) (*HandAndFootInteractor, error) {
	return restoreAndBuild[domain.HandAndFoot](data, func(g *domain.HandAndFoot) *HandAndFootInteractor {
		return &HandAndFootInteractor{GameBase: GameBase[interfaces.HandAndFootGame]{Game: g}, gp: gp}
	})
}
