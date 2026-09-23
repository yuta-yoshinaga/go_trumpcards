//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertEuchreCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func newEuchreErrorTestGame() *Euchre {
	players := []*EuchrePlayer{
		NewEuchrePlayer(true, 0), NewEuchrePlayer(false, 1),
		NewEuchrePlayer(false, 0), NewEuchrePlayer(false, 1),
	}
	return NewEuchre(NewTrumpCardsEuchre(), players, DefaultEuchreConfig())
}

func TestEuchreDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("face-up suit", func(t *testing.T) {
		g := newEuchreErrorTestGame()
		g.SetPhase(EuchrePhaseCallTrump)
		g.SetBidPlayerIdx(0)
		g.SetFaceUpCard(NewCard(CardDesignHeart, 10, false))
		assertEuchreCodedError(t, g.PlayerCallTrump(CardDesignHeart, false), ErrInvalidPlay, "euchre.errFaceUpSuit")
	})
	t.Run("invalid suit", func(t *testing.T) {
		g := newEuchreErrorTestGame()
		g.SetPhase(EuchrePhaseCallTrump)
		g.SetBidPlayerIdx(0)
		assertEuchreCodedError(t, g.PlayerCallTrump(99, false), ErrInvalidPlay, "euchre.errInvalidSuit")
	})
	t.Run("dealer must choose a suit", func(t *testing.T) {
		g := newEuchreErrorTestGame()
		g.SetPhase(EuchrePhaseCallTrump)
		g.SetDealerIdx(0)
		g.SetBidPlayerIdx(0)
		assertEuchreCodedError(t, g.PlayerPassCall(), ErrCannotPass, "euchre.errDealerMustChooseSuit")
	})
	t.Run("card index", func(t *testing.T) {
		g := newEuchreErrorTestGame()
		g.SetPhase(EuchrePhasePlay)
		g.SetCurrentPlayerIdx(0)
		assertEuchreCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "euchre.errCardIndexOutOfRange")
	})
	t.Run("must follow lead suit", func(t *testing.T) {
		g := newEuchreErrorTestGame()
		g.SetPhase(EuchrePhasePlay)
		g.SetTrumpSuit(CardDesignHeart)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignClover, 9, false)}})
		g.GetPlayer(0).AddCard(NewCard(CardDesignClover, 10, false))
		g.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 10, false))
		assertEuchreCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "euchre.errFollowLeadSuit")
	})
}
