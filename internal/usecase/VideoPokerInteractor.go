//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// VideoPokerInteractorIF ビデオポーカーインタラクターインタフェース
type VideoPokerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet ベット
	Bet(amount int) string
	// Hold ホールド＆ドロー
	Hold(indices []int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// VideoPokerInteractor ビデオポーカーインタラクタークラス
type VideoPokerInteractor struct {
	GameBase[interfaces.VideoPokerGame]
	vpp presenter.VideoPokerPresenter
}

// NewVideoPokerInteractor コンストラクタ
func NewVideoPokerInteractor(vp interfaces.VideoPokerGame, vpp presenter.VideoPokerPresenter) *VideoPokerInteractor {
	mustNotNil("VideoPokerInteractor", map[string]any{"vp": vp, "vpp": vpp})
	return &VideoPokerInteractor{
		GameBase: GameBase[interfaces.VideoPokerGame]{Game: vp},
		vpp:      vpp,
	}
}

// Reset ゲーム初期化
func (vi *VideoPokerInteractor) Reset() string {
	return runAndPresent(vi.Game, vi.vpp, vi.Game.Reset)
}

// Bet ベット
func (vi *VideoPokerInteractor) Bet(amount int) string {
	return execAndPresent(vi.Game, vi.vpp, func() error { return vi.Game.Bet(amount) })
}

// Hold ホールド＆ドロー
func (vi *VideoPokerInteractor) Hold(indices []int) string {
	return execAndPresent(vi.Game, vi.vpp, func() error { return vi.Game.Hold(indices) })
}

// ActionLog 棋譜を出力する
func (vi *VideoPokerInteractor) ActionLog() string {
	return vi.vpp.ActionLogOutput(vi.Game)
}

// Hint ヒントを出力する
func (vi *VideoPokerInteractor) Hint() string {
	return vi.vpp.HintOutput(vi.Game)
}

// RestoreVideoPokerInteractor deserialises JSON into a VideoPokerInteractor.
func RestoreVideoPokerInteractor(data []byte, vpp presenter.VideoPokerPresenter) (*VideoPokerInteractor, error) {
	return restoreAndBuild[domain.VideoPoker](data, func(g *domain.VideoPoker) *VideoPokerInteractor {
		return &VideoPokerInteractor{GameBase: GameBase[interfaces.VideoPokerGame]{Game: g}, vpp: vpp}
	})
}
