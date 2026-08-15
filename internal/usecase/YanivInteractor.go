//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// YanivInteractorIF Yaniv インタラクターインタフェース
type YanivInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.YanivConfig) string
	// Discard カードの組を捨てる
	Discard(cardIndices []int) string
	// DeclareYaniv Yaniv を宣言する
	DeclareYaniv() string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromPickup 直前の捨て札の端から引く
	DrawFromPickup(end int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.YanivConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// YanivInteractor Yaniv インタラクタークラス
type YanivInteractor struct {
	GameBase[interfaces.YanivGame]
	gp presenter.YanivPresenter
}

// NewYanivInteractor コンストラクタ
func NewYanivInteractor(g interfaces.YanivGame, gp presenter.YanivPresenter) *YanivInteractor {
	mustNotNil("YanivInteractor", map[string]any{"g": g, "gp": gp})
	return &YanivInteractor{GameBase: GameBase[interfaces.YanivGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *YanivInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *YanivInteractor) ResetWithConfig(cfg domain.YanivConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Discard カードの組を捨てる
func (ci *YanivInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndices); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DeclareYaniv Yaniv を宣言する
func (ci *YanivInteractor) DeclareYaniv() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDeclareYaniv(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawFromStock 山札からカードを引く
func (ci *YanivInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromStock(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawFromPickup 直前の捨て札の端から引く
func (ci *YanivInteractor) DrawFromPickup(end int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromPickup(end); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *YanivInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *YanivInteractor) GetConfig() domain.YanivConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *YanivInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// yanivMaxCpuSteps bounds runCpuTurns so a malformed state can never spin the
// CPU loop forever (defensive — normal play always reaches a human turn, round
// end, or game end well within this limit).
const yanivMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (ci *YanivInteractor) runCpuTurns() {
	for step := 0; step < yanivMaxCpuSteps && !ci.Game.GetGameEndFlag(); step++ {
		phase := ci.Game.GetPhase()
		if phase == domain.YanivPhaseRoundEnd || phase == domain.YanivPhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

// RestoreYanivInteractor deserialises JSON into a YanivInteractor.
func RestoreYanivInteractor(data []byte, gp presenter.YanivPresenter) (*YanivInteractor, error) {
	return restoreAndBuild[domain.Yaniv](data, func(g *domain.Yaniv) *YanivInteractor {
		return &YanivInteractor{GameBase: GameBase[interfaces.YanivGame]{Game: g}, gp: gp}
	})
}
