//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AnacondaInteractorIF はアナコンダ (Anaconda) のインタラクターインタフェース。
type AnacondaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.AnacondaConfig) string
	// Pass 人間が選んだ札を左隣へ渡す
	Pass(indices []int) string
	// Keep 人間が残す 5 枚を選ぶ
	Keep(indices []int) string
	// Call 人間がコール (チェック含む) する
	Call() string
	// Raise 人間がレイズする
	Raise() string
	// Fold 人間がフォールドする
	Fold() string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.AnacondaConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// AnacondaInteractor はアナコンダのインタラクター。
type AnacondaInteractor struct {
	GameBase[interfaces.AnacondaGame]
	sp presenter.AnacondaPresenter
}

// NewAnacondaInteractor コンストラクタ。
func NewAnacondaInteractor(g interfaces.AnacondaGame, sp presenter.AnacondaPresenter) *AnacondaInteractor {
	mustNotNil("AnacondaInteractor", map[string]any{"g": g, "sp": sp})
	return &AnacondaInteractor{GameBase: GameBase[interfaces.AnacondaGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ti *AnacondaInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.Reset)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ti *AnacondaInteractor) ResetWithConfig(cfg domain.AnacondaConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Pass 人間が選んだ札を左隣へ渡す。
func (ti *AnacondaInteractor) Pass(indices []int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, func() error {
		return ti.Game.Pass(indices)
	})
}

// Keep 人間が残す 5 枚を選ぶ。
func (ti *AnacondaInteractor) Keep(indices []int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, func() error {
		return ti.Game.Keep(indices)
	})
}

// Call 人間がコール (チェック含む) する。
func (ti *AnacondaInteractor) Call() string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, ti.Game.PlayerCall)
}

// Raise 人間がレイズする。
func (ti *AnacondaInteractor) Raise() string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, ti.Game.PlayerRaise)
}

// Fold 人間がフォールドする。
func (ti *AnacondaInteractor) Fold() string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, ti.Game.PlayerFold)
}

// NextRound 次のラウンドを配る。
func (ti *AnacondaInteractor) NextRound() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.NextRound)
}

// GetConfig 現在の設定を返す。
func (ti *AnacondaInteractor) GetConfig() domain.AnacondaConfig {
	return ti.Game.GetConfig()
}

// Hint ヒントを出力する。
func (ti *AnacondaInteractor) Hint() string { return ti.sp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する。
func (ti *AnacondaInteractor) ActionLog() string { return ti.sp.ActionLogOutput(ti.Game) }

// RestoreAnacondaInteractor deserialises JSON into an AnacondaInteractor.
func RestoreAnacondaInteractor(data []byte, sp presenter.AnacondaPresenter) (*AnacondaInteractor, error) {
	return restoreAndBuild[domain.Anaconda](data, func(g *domain.Anaconda) *AnacondaInteractor {
		return &AnacondaInteractor{GameBase: GameBase[interfaces.AnacondaGame]{Game: g}, sp: sp}
	})
}
