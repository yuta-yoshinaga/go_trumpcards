//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertOmiCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestOmiDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("invalid suit", func(t *testing.T) {
		g := newTestOmi()
		g.SetPhase(domain.OmiPhaseCallTrump)
		g.SetTrumpCallerIdx(0)
		g.SetCurrentPlayerIdx(0)
		assertOmiCodedError(t, g.PlayerCallTrump(0), domain.ErrInvalidPlay, "omi.errInvalidSuit")
	})

	t.Run("must follow lead suit", func(t *testing.T) {
		g := newTestOmi()
		g.SetPhase(domain.OmiPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)}})
		setupOmiHand(g, 0, []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
		})
		assertOmiCodedError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "omi.errFollowLeadSuit")
	})
}
