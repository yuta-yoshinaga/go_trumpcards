//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BassetInteractorIF はバセットインタラクターのインタフェース。
type BassetInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// NextRound 次のディールを開始
	NextRound() string
	// PlaceBet ランクにベットを置く
	PlaceBet(rank, amount int) string
	// DealTurn 2枚をめくってベットを解決
	DealTurn() string
	// TakeWinnings receives a hit at the current payout.
	TakeWinnings() string
	// PressParoli keeps the stake and advances its payout stage.
	PressParoli() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BassetInteractor はバセットインタラクター。
type BassetInteractor struct {
	GameBase[interfaces.BassetGame]
	cp presenter.BassetPresenter
}

// NewBassetInteractor コンストラクタ。
func NewBassetInteractor(f interfaces.BassetGame, cp presenter.BassetPresenter) *BassetInteractor {
	mustNotNil("BassetInteractor", map[string]any{"f": f, "cp": cp})
	return &BassetInteractor{
		GameBase: GameBase[interfaces.BassetGame]{Game: f},
		cp:       cp,
	}
}

// Reset ゲーム初期化。
func (fi *BassetInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.cp, fi.Game.Reset)
}

// NextRound 次のディールを開始。
func (fi *BassetInteractor) NextRound() string {
	return runAndPresent(fi.Game, fi.cp, fi.Game.NextRound)
}

// PlaceBet ランクにベットを置く。
func (fi *BassetInteractor) PlaceBet(rank, amount int) string {
	return execAndPresent(fi.Game, fi.cp, func() error { return fi.Game.PlayerPlaceBet(rank, amount) })
}

// DealTurn 2枚をめくってベットを解決。
func (fi *BassetInteractor) DealTurn() string {
	return execAndPresent(fi.Game, fi.cp, fi.Game.PlayerDealTurn)
}

// TakeWinnings receives the current payout.
func (fi *BassetInteractor) TakeWinnings() string {
	return execAndPresent(fi.Game, fi.cp, fi.Game.PlayerTakeWinnings)
}

// PressParoli advances the payout stage.
func (fi *BassetInteractor) PressParoli() string {
	return execAndPresent(fi.Game, fi.cp, fi.Game.PlayerPressParoli)
}

// ActionLog 棋譜を出力する。
func (fi *BassetInteractor) ActionLog() string {
	return fi.cp.ActionLogOutput(fi.Game)
}

// RestoreBassetInteractor deserialises JSON into a BassetInteractor.
func RestoreBassetInteractor(data []byte, cp presenter.BassetPresenter) (*BassetInteractor, error) {
	return restoreAndBuild[domain.Basset](data, func(g *domain.Basset) *BassetInteractor {
		return &BassetInteractor{GameBase: GameBase[interfaces.BassetGame]{Game: g}, cp: cp}
	})
}
