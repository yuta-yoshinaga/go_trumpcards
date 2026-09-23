//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPasurPlayerPlayReturnsCodedError(t *testing.T) {
	p := NewDefaultPasur()
	p.Reset()
	p.phase, p.currentPlayerIdx = PasurPhasePlay, 0
	p.players[0].ResetGame()
	err := p.play(0, -1, nil)
	require.ErrorIs(t, err, ErrInvalidCard)
	de, ok := err.(*DomainError)
	require.True(t, ok)
	require.Empty(t, de.Message)
	require.Equal(t, "pasur.errCardIndexOutOfRange", de.MessageCode())
}
