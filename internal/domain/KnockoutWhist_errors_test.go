//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertKnockoutWhistCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestKnockoutWhistDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("trump suit out of range", func(t *testing.T) {
		g := newKoGame(true)
		g.SetPhase(KnockoutWhistPhaseTrumpSelect)
		g.SetLeadPlayerIdx(0)
		assertKnockoutWhistCodedError(t, g.PlayerSelectTrump(0), ErrInvalidPlay, "knockoutwhist.errTrumpSuitOutOfRange")
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newKoGame(true)
		g.SetPhase(KnockoutWhistPhasePlay)
		g.SetCurrentPlayerIdx(0)
		koSetHand(g.GetPlayer(0), koCard(CardDesignSpade, 7))
		assertKnockoutWhistCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "knockoutwhist.errCardIndexOutOfRange")
	})
}
