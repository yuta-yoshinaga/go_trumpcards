//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CribbageInteractorIF クリベッジインタラクターインタフェース
type CribbageInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CribbageConfig) string
	// Discard クリブに2枚捨てる
	Discard(cardIndices []int) string
	// Cut 非ディーラーがデッキをカットしてスターターを公開する
	Cut() string
	// Peg ペギングでカードを出す
	Peg(cardIndex int) string
	// Go Goを宣言する
	Go() string
	// ShowNext ショーフェーズの次のスコア計算
	ShowNext() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CribbageConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒント取得
	Hint() string
}

// CribbageInteractor クリベッジインタラクタークラス
type CribbageInteractor struct {
	GameBase[interfaces.CribbageGame]
	gp presenter.CribbagePresenter
}

// NewCribbageInteractor コンストラクタ
func NewCribbageInteractor(g interfaces.CribbageGame, gp presenter.CribbagePresenter) *CribbageInteractor {
	mustNotNil("CribbageInteractor", map[string]any{"g": g, "gp": gp})
	return &CribbageInteractor{GameBase: GameBase[interfaces.CribbageGame]{Game: g}, gp: gp}
}

// Hint ヒント取得
func (ci *CribbageInteractor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// Reset ゲーム初期化
func (ci *CribbageInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CribbageInteractor) ResetWithConfig(cfg domain.CribbageConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Discard クリブに2枚捨てる
func (ci *CribbageInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDiscard(cardIndices)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Cut 非ディーラーがデッキをカットしてスターターを公開する
func (ci *CribbageInteractor) Cut() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerCut()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Peg ペギングでカードを出す
func (ci *CribbageInteractor) Peg(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerPeg(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Go Goを宣言する
func (ci *CribbageInteractor) Go() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerGo()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ShowNext ショーフェーズの次のスコア計算
func (ci *CribbageInteractor) ShowNext() string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.ShowNext()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *CribbageInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *CribbageInteractor) GetConfig() domain.CribbageConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *CribbageInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行
func (ci *CribbageInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.CribbagePhaseRoundEnd || phase == domain.CribbagePhaseGameEnd || phase == domain.CribbagePhaseShow || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreCribbageInteractor deserialises JSON into a CribbageInteractor.
func RestoreCribbageInteractor(data []byte, gp presenter.CribbagePresenter) (*CribbageInteractor, error) {
	return restoreAndBuild[domain.Cribbage](data, func(g *domain.Cribbage) *CribbageInteractor {
		return &CribbageInteractor{GameBase: GameBase[interfaces.CribbageGame]{Game: g}, gp: gp}
	})
}
