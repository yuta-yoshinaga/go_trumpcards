//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertViraDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestViraDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("invalid bid", func(t *testing.T) {
		g := newTestVira(t)
		assertViraDomainError(t, g.applyBid(0, ViraBid(-1)), ErrInvalidPlay, "vira.errInvalidBid", map[string]string{"bid": "-1"})
	})

	t.Run("bid already placed", func(t *testing.T) {
		g := newTestVira(t)
		require.NoError(t, g.applyBid(0, ViraBidPass))
		assertViraDomainError(t, g.applyBid(0, ViraBidGask), ErrInvalidPlay, "vira.errBidAlreadyPlaced", nil)
	})

	t.Run("bid must outrank", func(t *testing.T) {
		g := newTestVira(t)
		require.NoError(t, g.applyBid(0, ViraBidSolo))
		assertViraDomainError(t, g.applyBid(1, ViraBidGask), ErrInvalidPlay, "vira.errBidMustOutrank", map[string]string{"bid": "Solo"})
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newTestVira(t)
		g.SetPhase(ViraPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assertViraDomainError(t, g.PlayerPlay(-1), ErrInvalidCard, "vira.errCardIndexOutOfRange", nil)
	})

	t.Run("card missing", func(t *testing.T) {
		g := newTestVira(t)
		setViraHand(t, g, 0, nil)
		g.SetPhase(ViraPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assertViraDomainError(t, g.PlayerPlay(0), ErrInvalidCard, "vira.errCardMissing", nil)
	})

	t.Run("must follow lead suit", func(t *testing.T) {
		g := newTestVira(t)
		setViraHand(t, g, 0, NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 7, false))
		g.SetPhase(ViraPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 13, false)}}
		assertViraDomainError(t, g.PlayerPlay(1), ErrInvalidPlay, "vira.errFollowLeadSuit", nil)
	})

	t.Run("player index out of range", func(t *testing.T) {
		g := newTestVira(t)
		assertViraDomainError(t, g.ForcePassForTest(-1), ErrInvalidPlay, "vira.errPlayerIndexOutOfRange", nil)
	})
}
