//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSedmaInvalidCardErrorHasMessageCode(t *testing.T) {
	g := newSedGame(true)
	g.SetPhase(SedmaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	sedSetHand(g.GetPlayer(0), sedCard(CardDesignClover, 8))
	err := g.PlayerPlay(-1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCard)
	assert.Equal(t, "sedma.errCardIndexOutOfRange", err.(*DomainError).MessageCode())
}
