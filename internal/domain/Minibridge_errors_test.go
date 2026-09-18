//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMinibridgePlayerPlayReturnsCodedErrors(t *testing.T) {
	m := NewDefaultMinibridge()
	m.Reset()
	m.phase, m.currentPlayerIdx = MinibridgePhasePlay, 0
	m.players[0].ResetRound()
	assertMinibridgeErrorCode(t, m.play(0, -1), ErrInvalidCard, "minibridge.errCardIndexOutOfRange")
	m.players[0].AddCard(nil)
	assertMinibridgeErrorCode(t, m.play(0, 0), ErrInvalidCard, "minibridge.errCardMissing")
}

func assertMinibridgeErrorCode(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok)
	require.Empty(t, de.Message)
	require.Equal(t, code, de.MessageCode())
	require.True(t, errors.Is(err, sentinel))
}
