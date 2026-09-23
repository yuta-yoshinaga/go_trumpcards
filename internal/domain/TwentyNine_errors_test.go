//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertTwentyNineCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestTwentyNineDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newTnAllHuman()

	t.Run("invalid bid", func(t *testing.T) {
		assertTwentyNineCodedError(t, g.applyBid(0, TwentyNineBid(-1)), ErrInvalidPlay, "twentynine.errInvalidBid")
	})

	t.Run("bid must exceed current bid", func(t *testing.T) {
		require.NoError(t, g.applyBid(0, TwentyNineBidTwenty))
		assertTwentyNineCodedError(t, g.applyBid(1, TwentyNineBidSixteen), ErrInvalidPlay, "twentynine.errBidMustExceed")
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newTnGame(true)
		g.SetPhase(TwentyNinePhasePlay)
		g.SetCurrentPlayerIdx(0)
		tnSetHand(g.GetPlayer(0), tnCard(CardDesignSpade, 7))
		assertTwentyNineCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "twentynine.errCardIndexOutOfRange")
	})
}
