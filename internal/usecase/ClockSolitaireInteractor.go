//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ClockSolitaireInteractorIF クロックソリティアインタラクターインタフェース
type ClockSolitaireInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Step 1ステップ実行
	Step() string
	// AutoPlay 自動プレイ
	AutoPlay() string
	// Undo 直前のステップを取り消す
	Undo() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ClockSolitaireInteractor クロックソリティアインタラクタークラス
type ClockSolitaireInteractor struct {
	GameBase[interfaces.ClockSolitaireGame]
	p presenter.ClockSolitairePresenter
}

// NewClockSolitaireInteractor コンストラクタ
func NewClockSolitaireInteractor(g interfaces.ClockSolitaireGame, p presenter.ClockSolitairePresenter) *ClockSolitaireInteractor {
	mustNotNil("ClockSolitaireInteractor", map[string]any{"g": g, "p": p})
	return &ClockSolitaireInteractor{GameBase: GameBase[interfaces.ClockSolitaireGame]{Game: g}, p: p}
}

// Reset ゲーム初期化
func (ci *ClockSolitaireInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.p, ci.Game.Reset)
}

// Step 1ステップ実行
func (ci *ClockSolitaireInteractor) Step() string {
	return execAndPresent(ci.Game, ci.p, ci.Game.Step)
}

// AutoPlay 自動プレイ
func (ci *ClockSolitaireInteractor) AutoPlay() string {
	return execAndPresent(ci.Game, ci.p, ci.Game.AutoPlay)
}

// Undo 直前のステップを取り消す
func (ci *ClockSolitaireInteractor) Undo() string {
	return execAndPresent(ci.Game, ci.p, ci.Game.Undo)
}

// ActionLog 棋譜を出力する
func (ci *ClockSolitaireInteractor) ActionLog() string {
	return ci.p.ActionLogOutput(ci.Game)
}

// RestoreClockSolitaireInteractor deserialises JSON into a ClockSolitaireInteractor.
func RestoreClockSolitaireInteractor(data []byte, p presenter.ClockSolitairePresenter) (*ClockSolitaireInteractor, error) {
	return restoreAndBuild[domain.ClockSolitaire](data, func(g *domain.ClockSolitaire) *ClockSolitaireInteractor {
		return &ClockSolitaireInteractor{GameBase: GameBase[interfaces.ClockSolitaireGame]{Game: g}, p: p}
	})
}
