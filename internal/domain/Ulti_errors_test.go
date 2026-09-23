//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertUltiDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestUltiDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("invalid contract", func(t *testing.T) {
		g := newTestUlti()
		assertUltiDomainError(t, g.PlayerBid(domain.UltiContractNone, -1), domain.ErrInvalidPlay, "ulti.errInvalidContract")
	})

	t.Run("contract trump suit required", func(t *testing.T) {
		g := newTestUlti()
		assertUltiDomainError(t, g.PlayerBid(domain.UltiContractParty, -1), domain.ErrInvalidPlay, "ulti.errContractTrumpSuitRequired")
	})

	t.Run("discard count", func(t *testing.T) {
		g := newTestUlti()
		require.NoError(t, g.PlayerBid(domain.UltiContractBetli, -1))
		assertUltiDomainError(t, g.PlayerDiscard([]int{0}), domain.ErrInvalidPlay, "ulti.errDiscardCount")
	})

	t.Run("discard card index range", func(t *testing.T) {
		g := newTestUlti()
		require.NoError(t, g.PlayerBid(domain.UltiContractBetli, -1))
		assertUltiDomainError(t, g.PlayerDiscard([]int{0, 99}), domain.ErrInvalidCard, "ulti.errCardIndexOutOfRange")
	})

	t.Run("duplicate discard card", func(t *testing.T) {
		g := newTestUlti()
		require.NoError(t, g.PlayerBid(domain.UltiContractBetli, -1))
		assertUltiDomainError(t, g.PlayerDiscard([]int{3, 3}), domain.ErrInvalidPlay, "ulti.errDuplicateCard")
	})

	t.Run("play card index range", func(t *testing.T) {
		g := newTestUlti()
		g.SetPhase(domain.UltiPhasePlay)
		g.SetCurrentPlayerIdx(0)
		setUltiHand(g, 0, ultiCard(domain.CardDesignSpade, 13))
		assertUltiDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "ulti.errCardIndexOutOfRange")
	})

	t.Run("must follow lead or overtrump", func(t *testing.T) {
		g := newTestUlti()
		g.SetPhase(domain.UltiPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetContract(domain.UltiContractParty)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: ultiCard(domain.CardDesignClover, 13)}})
		setUltiHand(g, 0, ultiCard(domain.CardDesignClover, 9), ultiCard(domain.CardDesignSpade, 8))
		assertUltiDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "ulti.errFollowLeadOrOvertrump")
	})
}
