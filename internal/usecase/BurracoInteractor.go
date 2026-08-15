//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BurracoInteractorIF ブラーコインタラクターインタフェース
type BurracoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BurracoConfig) string
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
	GetConfig() domain.BurracoConfig
	// Hint 現在手番に対する推奨アクションを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BurracoInteractor ブラーコインタラクタークラス
type BurracoInteractor struct {
	GameBase[interfaces.BurracoGame]
	gp presenter.BurracoPresenter
}

// NewBurracoInteractor コンストラクタ
func NewBurracoInteractor(g interfaces.BurracoGame, gp presenter.BurracoPresenter) *BurracoInteractor {
	mustNotNil("BurracoInteractor", map[string]any{"g": g, "gp": gp})
	return &BurracoInteractor{GameBase: GameBase[interfaces.BurracoGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *BurracoInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *BurracoInteractor) ResetWithConfig(cfg domain.BurracoConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *BurracoInteractor) DrawFromStock() string {
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
func (ci *BurracoInteractor) DrawFromDiscard(naturalPairIndices []int) string {
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
func (ci *BurracoInteractor) Meld(meldGroups [][]int) string {
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
func (ci *BurracoInteractor) SkipMeld() string {
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
func (ci *BurracoInteractor) Discard(cardIndex int) string {
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
func (ci *BurracoInteractor) GoOut() string {
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
func (ci *BurracoInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *BurracoInteractor) GetConfig() domain.BurracoConfig {
	return ci.Game.GetConfig()
}

// Hint 現在手番に対する推奨アクションを出力する
func (ci *BurracoInteractor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *BurracoInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行
func (ci *BurracoInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase == domain.BurracoPhaseRoundEnd || phase == domain.BurracoPhaseGameEnd {
			return
		}
		if ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestoreBurracoInteractor deserialises JSON into a BurracoInteractor.
func RestoreBurracoInteractor(data []byte, gp presenter.BurracoPresenter) (*BurracoInteractor, error) {
	return restoreAndBuild[domain.Burraco](data, func(g *domain.Burraco) *BurracoInteractor {
		return &BurracoInteractor{GameBase: GameBase[interfaces.BurracoGame]{Game: g}, gp: gp}
	})
}
