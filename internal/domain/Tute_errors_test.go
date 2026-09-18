//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertTuteDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestTuteDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("card index out of range", func(t *testing.T) {
		g := newTuteGame(true)
		g.Reset()
		g.SetPhase(TutePhasePlay)
		g.SetCurrentPlayerIdx(0)
		tuteSetHand(g.GetPlayer(0), tuteCard(CardDesignSpade, 7))
		assertTuteDomainError(t, g.PlayerPlay(1), ErrInvalidCard, "tute.errCardIndexOutOfRange")
	})

	t.Run("invalid marriage", func(t *testing.T) {
		g := newTuteGame(true)
		g.Reset()
		g.SetPhase(TutePhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetLeadPlayerIdx(0)
		g.SetCurrentTrick(nil)
		tuteSetHand(g.GetPlayer(0), tuteCard(CardDesignSpade, 13))
		assertTuteDomainError(t, g.PlayerDeclareMarriage(CardDesignSpade), ErrInvalidPlay, "tute.errInvalidMarriage")
	})

	t.Run("cannot declare tute", func(t *testing.T) {
		g := newTuteGame(true)
		g.Reset()
		g.SetPhase(TutePhasePlay)
		g.SetCurrentPlayerIdx(0)
		tuteSetHand(g.GetPlayer(0), tuteCard(CardDesignSpade, 13), tuteCard(CardDesignClover, 12))
		assertTuteDomainError(t, g.PlayerDeclareTute(), ErrInvalidPlay, "tute.errCannotDeclareTute")
	})
}
