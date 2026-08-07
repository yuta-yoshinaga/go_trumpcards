//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// UltimateTexasHoldemInteractorIF アルティメット・テキサスホールデムインタラクターインタフェース
type UltimateTexasHoldemInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテ＋同額のブラインド＋オプションのトリップス
	Bet(ante, trips int) string
	// Play プレイベット（プリフロップ 3/4、フロップ 2、リバー 1）
	Play(multiplier int) string
	// Check プリフロップまたはフロップでチェック
	Check() string
	// Fold リバーでフォールド
	Fold() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// UltimateTexasHoldemInteractor アルティメット・テキサスホールデムインタラクタークラス
type UltimateTexasHoldemInteractor struct {
	GameBase[interfaces.UltimateTexasHoldemGame]
	up presenter.UltimateTexasHoldemPresenter
}

// NewUltimateTexasHoldemInteractor コンストラクタ
func NewUltimateTexasHoldemInteractor(g interfaces.UltimateTexasHoldemGame, up presenter.UltimateTexasHoldemPresenter) *UltimateTexasHoldemInteractor {
	mustNotNil("UltimateTexasHoldemInteractor", map[string]any{"g": g, "up": up})
	return &UltimateTexasHoldemInteractor{
		GameBase: GameBase[interfaces.UltimateTexasHoldemGame]{Game: g},
		up:       up,
	}
}

// Reset ゲーム初期化
func (ui *UltimateTexasHoldemInteractor) Reset() string {
	return runAndPresent(ui.Game, ui.up, ui.Game.Reset)
}

// Bet アンテ＋ブラインド＋オプションのトリップス
func (ui *UltimateTexasHoldemInteractor) Bet(ante, trips int) string {
	return execAndPresent(ui.Game, ui.up, func() error { return ui.Game.Bet(ante, trips) })
}

// Play プレイベット
func (ui *UltimateTexasHoldemInteractor) Play(multiplier int) string {
	return execAndPresent(ui.Game, ui.up, func() error { return ui.Game.Play(multiplier) })
}

// Check チェック
func (ui *UltimateTexasHoldemInteractor) Check() string {
	return execAndPresent(ui.Game, ui.up, ui.Game.Check)
}

// Fold フォールド
func (ui *UltimateTexasHoldemInteractor) Fold() string {
	return execAndPresent(ui.Game, ui.up, ui.Game.Fold)
}

// Hint ヒント取得
func (ui *UltimateTexasHoldemInteractor) Hint() string {
	return ui.up.HintOutput(ui.Game)
}

// ActionLog 棋譜を出力する
func (ui *UltimateTexasHoldemInteractor) ActionLog() string {
	return ui.up.ActionLogOutput(ui.Game)
}

// RestoreUltimateTexasHoldemInteractor deserialises JSON into a UltimateTexasHoldemInteractor.
func RestoreUltimateTexasHoldemInteractor(data []byte, up presenter.UltimateTexasHoldemPresenter) (*UltimateTexasHoldemInteractor, error) {
	return restoreAndBuild[domain.UltimateTexasHoldem](data, func(g *domain.UltimateTexasHoldem) *UltimateTexasHoldemInteractor {
		return &UltimateTexasHoldemInteractor{GameBase: GameBase[interfaces.UltimateTexasHoldemGame]{Game: g}, up: up}
	})
}
