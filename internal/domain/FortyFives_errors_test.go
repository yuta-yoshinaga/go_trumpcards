//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertFortyFivesCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestFortyFivesDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("invalid bid", func(t *testing.T) {
		g := newFfAllHuman()
		g.Reset()
		assertFortyFivesCodedError(t, g.PlayerBid(FortyFivesBid(99)), ErrInvalidPlay, "fortyfives.errInvalidBid")
	})

	t.Run("bid must exceed current bid", func(t *testing.T) {
		g := newFfAllHuman()
		g.Reset()
		require.NoError(t, g.PlayerBid(FortyFivesBidTwenty))
		assertFortyFivesCodedError(t, g.PlayerBid(FortyFivesBidFifteen), ErrInvalidPlay, "fortyfives.errBidMustExceedCurrent")
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newFfGame(true)
		g.SetPhase(FortyFivesPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assertFortyFivesCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "fortyfives.errCardIndexOutOfRange")
	})

	t.Run("must follow trump", func(t *testing.T) {
		g := newFfGame(true)
		g.SetPhase(FortyFivesPhasePlay)
		g.SetTrumpSuit(CardDesignDiamond)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: ffCard(CardDesignDiamond, 7)}})
		ffSetHand(g.GetPlayer(0), ffCard(CardDesignDiamond, 9), ffCard(CardDesignClover, 9))
		assertFortyFivesCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "fortyfives.errMustFollowTrump")
	})

	t.Run("must follow lead suit", func(t *testing.T) {
		g := newFfGame(true)
		g.SetPhase(FortyFivesPhasePlay)
		g.SetTrumpSuit(CardDesignDiamond)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: ffCard(CardDesignClover, 7)}})
		ffSetHand(g.GetPlayer(0), ffCard(CardDesignClover, 9), ffCard(CardDesignHeart, 9))
		assertFortyFivesCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "fortyfives.errMustFollowLeadSuit")
	})
}
