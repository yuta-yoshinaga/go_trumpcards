//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GutsInteractorIF はガッツ (Guts) のインタラクターインタフェース。
type GutsInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.GutsConfig) string
	// Declare 人間の in/out 宣言を受け付けラウンドを解決する
	Declare(stay bool) string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.GutsConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GutsInteractor はガッツのインタラクター。
type GutsInteractor struct {
	GameBase[interfaces.GutsGame]
	sp presenter.GutsPresenter
}

// NewGutsInteractor コンストラクタ。
func NewGutsInteractor(g interfaces.GutsGame, sp presenter.GutsPresenter) *GutsInteractor {
	mustNotNil("GutsInteractor", map[string]any{"g": g, "sp": sp})
	return &GutsInteractor{GameBase: GameBase[interfaces.GutsGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ti *GutsInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.Reset)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ti *GutsInteractor) ResetWithConfig(cfg domain.GutsConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Declare 人間の in/out 宣言を受け付けラウンドを解決する。
func (ti *GutsInteractor) Declare(stay bool) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, func() error {
		return ti.Game.Declare(stay)
	})
}

// NextRound 次のラウンドを配る。
func (ti *GutsInteractor) NextRound() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.NextRound)
}

// GetConfig 現在の設定を返す。
func (ti *GutsInteractor) GetConfig() domain.GutsConfig {
	return ti.Game.GetConfig()
}

// Hint ヒントを出力する。
func (ti *GutsInteractor) Hint() string { return ti.sp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する。
func (ti *GutsInteractor) ActionLog() string { return ti.sp.ActionLogOutput(ti.Game) }

// RestoreGutsInteractor deserialises JSON into a GutsInteractor.
func RestoreGutsInteractor(data []byte, sp presenter.GutsPresenter) (*GutsInteractor, error) {
	return restoreAndBuild[domain.Guts](data, func(g *domain.Guts) *GutsInteractor {
		return &GutsInteractor{GameBase: GameBase[interfaces.GutsGame]{Game: g}, sp: sp}
	})
}
