//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BoliviaInteractorIF ボリビアインタラクターインタフェース
type BoliviaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BoliviaConfig) string
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
	GetConfig() domain.BoliviaConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BoliviaInteractor ボリビアインタラクタークラス
type BoliviaInteractor struct {
	GameBase[interfaces.BoliviaGame]
	gp presenter.BoliviaPresenter
}

// NewBoliviaInteractor コンストラクタ
func NewBoliviaInteractor(g interfaces.BoliviaGame, gp presenter.BoliviaPresenter) *BoliviaInteractor {
	mustNotNil("BoliviaInteractor", map[string]any{"g": g, "gp": gp})
	return &BoliviaInteractor{GameBase: GameBase[interfaces.BoliviaGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *BoliviaInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *BoliviaInteractor) ResetWithConfig(cfg domain.BoliviaConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *BoliviaInteractor) DrawFromStock() string {
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
func (ci *BoliviaInteractor) DrawFromDiscard(naturalPairIndices []int) string {
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
func (ci *BoliviaInteractor) Meld(meldGroups [][]int) string {
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
func (ci *BoliviaInteractor) SkipMeld() string {
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
func (ci *BoliviaInteractor) Discard(cardIndex int) string {
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
func (ci *BoliviaInteractor) GoOut() string {
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
func (ci *BoliviaInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *BoliviaInteractor) GetConfig() domain.BoliviaConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *BoliviaInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行
func (ci *BoliviaInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.BoliviaPhaseRoundEnd || phase == domain.BoliviaPhaseGameEnd || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreBoliviaInteractor deserialises JSON into a BoliviaInteractor.
func RestoreBoliviaInteractor(data []byte, gp presenter.BoliviaPresenter) (*BoliviaInteractor, error) {
	return restoreAndBuild[domain.Bolivia](data, func(g *domain.Bolivia) *BoliviaInteractor {
		return &BoliviaInteractor{GameBase: GameBase[interfaces.BoliviaGame]{Game: g}, gp: gp}
	})
}
