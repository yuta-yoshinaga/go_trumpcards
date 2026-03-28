package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// VideoPokerInteractorIF ビデオポーカーインタラクターインタフェース
type VideoPokerInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Bet ベット
	Bet(amount int) string
	// Hold ホールド＆ドロー
	Hold(indices []int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// VideoPokerInteractor ビデオポーカーインタラクタークラス
type VideoPokerInteractor struct {
	vp  interfaces.VideoPokerGame
	vpp presenter.VideoPokerPresenter
}

// NewVideoPokerInteractor コンストラクタ
func NewVideoPokerInteractor(vp interfaces.VideoPokerGame, vpp presenter.VideoPokerPresenter) *VideoPokerInteractor {
	mustNotNil("VideoPokerInteractor", map[string]any{"vp": vp, "vpp": vpp})
	return &VideoPokerInteractor{
		vp:  vp,
		vpp: vpp,
	}
}

// Reset ゲーム初期化
func (vi *VideoPokerInteractor) Reset() string {
	return runAndPresent(vi.vp, vi.vpp, vi.vp.Reset)
}

// Bet ベット
func (vi *VideoPokerInteractor) Bet(amount int) string {
	return execAndPresent(vi.vp, vi.vpp, func() error { return vi.vp.Bet(amount) })
}

// Hold ホールド＆ドロー
func (vi *VideoPokerInteractor) Hold(indices []int) string {
	return execAndPresent(vi.vp, vi.vpp, func() error { return vi.vp.Hold(indices) })
}

// ActionLog 棋譜を出力する
func (vi *VideoPokerInteractor) ActionLog() string {
	return vi.vpp.ActionLogOutput(vi.vp)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (vi *VideoPokerInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(vi.vp)
}

// RestoreVideoPokerInteractor deserialises JSON into a VideoPokerInteractor.
func RestoreVideoPokerInteractor(data []byte, vpp presenter.VideoPokerPresenter) (*VideoPokerInteractor, error) {
	var vp domain.VideoPoker
	if err := json.Unmarshal(data, &vp); err != nil {
		return nil, err
	}
	return &VideoPokerInteractor{vp: &vp, vpp: vpp}, nil
}
