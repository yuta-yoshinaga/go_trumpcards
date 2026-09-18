//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertThreeThirteenDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestThreeThirteenDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("discard pile empty", func(t *testing.T) {
		g := newTestThreeThirteen(2)
		g.SetPhase(ThreeThirteenPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		g.SetDiscardPile(nil)
		assertThreeThirteenDomainError(t, g.PlayerDrawFromDiscard(), ErrInvalidPlay, "threethirteen.errDiscardPileEmpty")
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newTestThreeThirteen(2)
		g.SetPhase(ThreeThirteenPhaseDiscard)
		g.SetCurrentPlayerIdx(0)
		ttSetHand(g.GetPlayer(0), ttCard(CardDesignSpade, 5))
		assertThreeThirteenDomainError(t, g.PlayerDiscard(-1), ErrInvalidCard, "threethirteen.errCardIndexOutOfRange")
	})

	t.Run("already knocked", func(t *testing.T) {
		g := newTestThreeThirteen(2)
		g.SetPhase(ThreeThirteenPhaseDiscard)
		g.SetCurrentPlayerIdx(0)
		ttSetHand(g.GetPlayer(0), ttCard(CardDesignSpade, 7), ttCard(CardDesignHeart, 7), ttCard(CardDesignDiamond, 7), ttCard(CardDesignClover, 9))
		require.NoError(t, g.PlayerKnock(3))
		g.SetPhase(ThreeThirteenPhaseDiscard)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).AddCard(ttCard(CardDesignClover, 10))
		assertThreeThirteenDomainError(t, g.PlayerKnock(0), ErrInvalidPlay, "threethirteen.errAlreadyKnocked")
	})

	t.Run("cannot knock", func(t *testing.T) {
		g := newTestThreeThirteen(2)
		g.SetPhase(ThreeThirteenPhaseDiscard)
		g.SetCurrentPlayerIdx(0)
		ttSetHand(g.GetPlayer(0), ttCard(CardDesignSpade, 7), ttCard(CardDesignHeart, 8))
		assertThreeThirteenDomainError(t, g.PlayerKnock(0), ErrInvalidPlay, "threethirteen.errCannotKnock")
	})
}
