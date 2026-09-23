//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertMacauCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestMacauDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("card index out of range", func(t *testing.T) {
		g := newTestMacau()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		assertMacauCodedError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "macau.errCardIndexOutOfRange")
	})

	t.Run("card cannot be played", func(t *testing.T) {
		g := newTestMacau()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		assertMacauCodedError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "macau.errCardNotPlayable")
	})

	t.Run("suit out of range", func(t *testing.T) {
		g := newTestMacau()
		g.SetPhase(domain.MacauPhaseChooseSuit)
		g.SetCurrentPlayerIdx(0)
		assertMacauCodedError(t, g.PlayerChooseSuit(0), domain.ErrInvalidPlay, "macau.errSuitOutOfRange")
	})
}
