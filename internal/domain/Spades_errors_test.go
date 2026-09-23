//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertSpadesDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestSpadesDomainErrorsHaveMessageCodes(t *testing.T) {
	s := newTestSpades()
	setupSpadesBidPhase(s, 0)
	assertSpadesDomainError(t, s.PlayerBid(-1), domain.ErrInvalidPlay, "spades.errBidOutOfRange", map[string]string{"max": "13"})

	s = newTestSpades()
	setupSpadesPlayPhase(s, 0, 0, 1)
	assertSpadesDomainError(t, s.PlayerPlay(-1), domain.ErrInvalidCard, "spades.errCardIndexOutOfRange", nil)

	s = newTestSpades()
	setupSpadesPlayPhase(s, 0, 0, 1)
	s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
	s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	assertSpadesDomainError(t, s.PlayerPlay(1), domain.ErrInvalidPlay, "spades.errLeadTwoOfClubs", nil)

	s = newTestSpades()
	setupSpadesPlayPhase(s, 0, 0, 2)
	s.SetSpadesBroken(false)
	s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	assertSpadesDomainError(t, s.PlayerPlay(1), domain.ErrInvalidPlay, "spades.errSpadesNotBroken", nil)

	s = newTestSpades()
	setupSpadesPlayPhase(s, 0, 0, 2)
	s.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 3, false)}})
	s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	s.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
	assertSpadesDomainError(t, s.PlayerPlay(0), domain.ErrInvalidPlay, "spades.errFollowLeadSuit", nil)
}
