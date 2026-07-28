//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FaroInteractorIF はファロインタラクターのインタフェース。
type FaroInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// NextRound 次のディールを開始
	NextRound() string
	// PlaceBet ランクにベットを置く
	PlaceBet(rank, amount int, copper bool) string
	// ClearBet 指定ランクのベットを取り消す
	ClearBet(rank int) string
	// ClearAll すべてのベットを取り消す
	ClearAll() string
	// DealTurn 2枚をめくってベットを解決
	DealTurn() string
	// Call 残り3枚の順序を予想
	Call(order []int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// FaroInteractor はファロインタラクター。
type FaroInteractor struct {
	GameBase[interfaces.FaroGame]
	cp presenter.FaroPresenter
}

// NewFaroInteractor コンストラクタ。
func NewFaroInteractor(f interfaces.FaroGame, cp presenter.FaroPresenter) *FaroInteractor {
	mustNotNil("FaroInteractor", map[string]any{"f": f, "cp": cp})
	return &FaroInteractor{
		GameBase: GameBase[interfaces.FaroGame]{Game: f},
		cp:       cp,
	}
}

// Reset ゲーム初期化。
func (fi *FaroInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.cp, fi.Game.Reset)
}

// NextRound 次のディールを開始。
func (fi *FaroInteractor) NextRound() string {
	return runAndPresent(fi.Game, fi.cp, fi.Game.NextRound)
}

// PlaceBet ランクにベットを置く。
func (fi *FaroInteractor) PlaceBet(rank, amount int, copper bool) string {
	return execAndPresent(fi.Game, fi.cp, func() error { return fi.Game.PlayerPlaceBet(rank, amount, copper) })
}

// ClearBet 指定ランクのベットを取り消す。
func (fi *FaroInteractor) ClearBet(rank int) string {
	return execAndPresent(fi.Game, fi.cp, func() error { return fi.Game.PlayerClearBet(rank) })
}

// ClearAll すべてのベットを取り消す。
func (fi *FaroInteractor) ClearAll() string {
	return execAndPresent(fi.Game, fi.cp, fi.Game.PlayerClearAll)
}

// DealTurn 2枚をめくってベットを解決。
func (fi *FaroInteractor) DealTurn() string {
	return execAndPresent(fi.Game, fi.cp, fi.Game.PlayerDealTurn)
}

// Call 残り3枚の順序を予想。
func (fi *FaroInteractor) Call(order []int) string {
	return execAndPresent(fi.Game, fi.cp, func() error { return fi.Game.PlayerCall(order) })
}

// ActionLog 棋譜を出力する。
func (fi *FaroInteractor) ActionLog() string {
	return fi.cp.ActionLogOutput(fi.Game)
}

// RestoreFaroInteractor deserialises JSON into a FaroInteractor.
func RestoreFaroInteractor(data []byte, cp presenter.FaroPresenter) (*FaroInteractor, error) {
	return restoreAndBuild[domain.Faro](data, func(g *domain.Faro) *FaroInteractor {
		return &FaroInteractor{GameBase: GameBase[interfaces.FaroGame]{Game: g}, cp: cp}
	})
}
