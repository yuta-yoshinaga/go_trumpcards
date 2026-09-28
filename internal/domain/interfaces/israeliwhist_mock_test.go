//go:build test

package interfaces

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockIsraeliWhistGameGetCurrentTrickWinnerIdx(t *testing.T) {
	game := new(MockIsraeliWhistGame)
	game.On("GetCurrentTrickWinnerIdx").Return(2).Once()

	require.Equal(t, 2, game.GetCurrentTrickWinnerIdx())
	game.AssertExpectations(t)
}
