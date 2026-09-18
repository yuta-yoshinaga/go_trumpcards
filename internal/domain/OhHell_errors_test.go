//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertOhHellDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestOhHellDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("bid out of range", func(t *testing.T) {
		o := newTestOhHell()
		setupOhHellBidPhase(o, 0)
		o.SetHandSize(5)
		assertOhHellDomainError(t, o.PlayerBid(-1), domain.ErrInvalidPlay, "ohhell.errBidOutOfRange", map[string]string{"max": "5"})
	})

	t.Run("restricted dealer bid", func(t *testing.T) {
		o := newTestOhHell()
		setupOhHellBidPhase(o, 0)
		o.SetHandSize(5)
		o.GetPlayer(1).SetBid(1)
		o.GetPlayer(2).SetBid(1)
		o.GetPlayer(3).SetBid(1)
		assertOhHellDomainError(t, o.PlayerBid(2), domain.ErrInvalidPlay, "ohhell.errRestrictedBid", map[string]string{"bid": "2"})
	})

	t.Run("card index out of range", func(t *testing.T) {
		o := newTestOhHell()
		setupOhHellPlayPhase(o, 0, 0, 1)
		assertOhHellDomainError(t, o.PlayerPlay(0), domain.ErrInvalidCard, "ohhell.errCardIndexOutOfRange", nil)
	})

	t.Run("must follow lead suit", func(t *testing.T) {
		o := newTestOhHell()
		setupOhHellPlayPhase(o, 0, 0, 1)
		o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		o.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)}})
		assertOhHellDomainError(t, o.PlayerPlay(1), domain.ErrInvalidPlay, "ohhell.errFollowLeadSuit", nil)
	})
}
