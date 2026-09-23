//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllFoursPlayerPlayReturnsCodedErrors(t *testing.T) {
	a := NewDefaultAllFours()
	a.Reset()
	a.phase, a.currentPlayerIdx = AllFoursPhasePlay, 0
	a.players[0].ResetRound()
	assertAllFoursErrorCode(t, a.PlayerPlay(-1), ErrInvalidCard, "allfours.errCardIndexOutOfRange")
	a.players[0].AddCard(NewCard(CardDesignHeart, 2, false))
	a.players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	a.trumpSuit = CardDesignDiamond
	a.currentTrick = []*TrickCard{{Card: NewCard(CardDesignHeart, 4, false)}}
	assertAllFoursErrorCode(t, a.PlayerPlay(1), ErrInvalidPlay, "allfours.errMustFollowLeadSuit")
}

func assertAllFoursErrorCode(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok)
	require.Empty(t, de.Message)
	require.Equal(t, code, de.MessageCode())
}
