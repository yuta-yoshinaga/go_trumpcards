package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FourCardPokerInteractorIF is the Four Card Poker use-case interface.
type FourCardPokerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset resets the game.
	Reset() string
	// Bet places the ante and optional Aces Up sidebet.
	Bet(ante, acesUp int) string
	// Play places the Play bet at the requested ante multiplier (1, 2, or 3).
	Play(multiplier int) string
	// Fold folds the hand.
	Fold() string
	// ActionLog returns the rendered action log.
	ActionLog() string
}

// FourCardPokerInteractor is the Four Card Poker interactor.
type FourCardPokerInteractor struct {
	GameBase[interfaces.FourCardPokerGame]
	fp presenter.FourCardPokerPresenter
}

// NewFourCardPokerInteractor constructs the interactor.
func NewFourCardPokerInteractor(fc interfaces.FourCardPokerGame, fp presenter.FourCardPokerPresenter) *FourCardPokerInteractor {
	mustNotNil("FourCardPokerInteractor", map[string]any{"fc": fc, "fp": fp})
	return &FourCardPokerInteractor{
		GameBase: GameBase[interfaces.FourCardPokerGame]{Game: fc},
		fp:       fp,
	}
}

// Reset resets the round.
func (fi *FourCardPokerInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.fp, fi.Game.Reset)
}

// Bet places the ante and optional Aces Up sidebet.
func (fi *FourCardPokerInteractor) Bet(ante, acesUp int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.Bet(ante, acesUp) })
}

// Play places the Play bet at the requested ante multiplier.
func (fi *FourCardPokerInteractor) Play(multiplier int) string {
	return execAndPresent(fi.Game, fi.fp, func() error { return fi.Game.Play(multiplier) })
}

// Fold folds the hand.
func (fi *FourCardPokerInteractor) Fold() string {
	return execAndPresent(fi.Game, fi.fp, fi.Game.Fold)
}

// ActionLog renders the action log via the presenter.
func (fi *FourCardPokerInteractor) ActionLog() string {
	return fi.fp.ActionLogOutput(fi.Game)
}

// RestoreFourCardPokerInteractor deserialises JSON into an interactor.
func RestoreFourCardPokerInteractor(data []byte, fp presenter.FourCardPokerPresenter) (*FourCardPokerInteractor, error) {
	return restoreAndBuild[domain.FourCardPoker](data, func(g *domain.FourCardPoker) *FourCardPokerInteractor {
		return &FourCardPokerInteractor{GameBase: GameBase[interfaces.FourCardPokerGame]{Game: g}, fp: fp}
	})
}
