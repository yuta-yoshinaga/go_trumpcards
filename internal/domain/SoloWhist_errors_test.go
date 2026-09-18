//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSoloWhistPlayerActionsReturnCodedErrors(t *testing.T) {
	g := NewDefaultSoloWhist()
	g.Reset()
	g.phase, g.currentPlayerIdx = SoloWhistPhaseBid, 0
	g.bids[0] = SoloWhistBidSolo
	g.bidDone[0] = true
	assertSoloWhistErrorCode(t, g.PlayerBid(SoloWhistBidSolo), ErrInvalidPlay, "solowhist.errBidMustExceed")
	g.phase = SoloWhistPhasePlay
	g.players[0].ResetRound()
	assertSoloWhistErrorCode(t, g.PlayerPlay(-1), ErrInvalidCard, "solowhist.errCardIndexOutOfRange")
}

func assertSoloWhistErrorCode(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok)
	require.Empty(t, de.Message)
	require.Equal(t, code, de.MessageCode())
}
