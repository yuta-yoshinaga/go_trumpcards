//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PaiGowInteractorIF パイガオポーカーインタラクターインタフェース
type PaiGowInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet ベット
	Bet(amount int) string
	// SetHands ハンド分割
	SetHands(lowIdx0, lowIdx1 int) string
	// AutoSetHands ハウスウェイで自動分割する
	AutoSetHands() string
	// Hint 推奨分割を出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PaiGowInteractor パイガオポーカーインタラクタークラス
type PaiGowInteractor struct {
	GameBase[interfaces.PaiGowGame]
	pp presenter.PaiGowPresenter
}

// NewPaiGowInteractor コンストラクタ
func NewPaiGowInteractor(pg interfaces.PaiGowGame, pp presenter.PaiGowPresenter) *PaiGowInteractor {
	mustNotNil("PaiGowInteractor", map[string]any{"pg": pg, "pp": pp})
	return &PaiGowInteractor{
		GameBase: GameBase[interfaces.PaiGowGame]{Game: pg},
		pp:       pp,
	}
}

// Reset ゲーム初期化
func (pi *PaiGowInteractor) Reset() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.Reset)
}

// Bet ベット
func (pi *PaiGowInteractor) Bet(amount int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.Bet(amount) })
}

// SetHands ハンド分割
func (pi *PaiGowInteractor) SetHands(lowIdx0, lowIdx1 int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.SetHands(lowIdx0, lowIdx1) })
}

// AutoSetHands ハウスウェイで自動分割する
func (pi *PaiGowInteractor) AutoSetHands() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.AutoSetHands)
}

// Hint 推奨分割を出力する
func (pi *PaiGowInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *PaiGowInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// RestorePaiGowInteractor deserialises JSON into a PaiGowInteractor.
func RestorePaiGowInteractor(data []byte, pp presenter.PaiGowPresenter) (*PaiGowInteractor, error) {
	return restoreAndBuild[domain.PaiGow](data, func(g *domain.PaiGow) *PaiGowInteractor {
		return &PaiGowInteractor{GameBase: GameBase[interfaces.PaiGowGame]{Game: g}, pp: pp}
	})
}
