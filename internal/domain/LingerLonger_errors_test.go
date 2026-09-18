//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLingerLongerPlayerPlayReturnsCodedErrors(t *testing.T) {
	l := NewDefaultLingerLonger()
	l.Reset()
	l.phase, l.currentPlayerIdx = LingerLongerPhasePlay, 0
	l.players[0].ResetGame()
	assertLingerLongerErrorCode(t, l.play(0, -1), ErrInvalidCard, "lingerlonger.errCardIndexOutOfRange")
	l.players[0].AddCard(nil)
	assertLingerLongerErrorCode(t, l.play(0, 0), ErrInvalidCard, "lingerlonger.errCardMissing")
}

func assertLingerLongerErrorCode(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok)
	require.Empty(t, de.Message)
	require.Equal(t, code, de.MessageCode())
}
