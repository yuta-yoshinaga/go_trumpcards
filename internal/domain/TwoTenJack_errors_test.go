//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTwoTenJackPlayerActionsReturnCodedErrors(t *testing.T) {
	g := NewDefaultTwoTenJack()
	g.Reset()
	g.phase = TwoTenJackPhaseDeclare
	assertTwoTenJackErrorCode(t, g.PlayerDeclareTrump(-1), ErrInvalidPlay, "twotenjack.errInvalidTrumpSuit")
	g.phase, g.currentPlayerIdx = TwoTenJackPhasePlay, 0
	g.players[0].ResetRound()
	assertTwoTenJackErrorCode(t, g.PlayerPlay(-1), ErrInvalidCard, "twotenjack.errCardIndexOutOfRange")
}

func assertTwoTenJackErrorCode(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok)
	require.Empty(t, de.Message)
	require.Equal(t, code, de.MessageCode())
}
