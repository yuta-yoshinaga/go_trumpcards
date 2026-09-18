//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertNinetyNineDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestNinetyNineDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("card index out of range while playing", func(t *testing.T) {
		o := newTestNinetyNine()
		o.SetPhase(domain.NinetyNinePhasePlay)
		o.SetCurrentPlayerIdx(0)
		assertNinetyNineDomainError(t, o.PlayerPlay(0), domain.ErrInvalidCard, "ninetynine.errCardIndexOutOfRange", nil)
	})

	t.Run("bury count", func(t *testing.T) {
		o := newTestNinetyNine()
		o.SetPhase(domain.NinetyNinePhaseBid)
		o.SetBidPlayerIdx(0)
		assertNinetyNineDomainError(t, o.PlayerBid([]int{0, 1}), domain.ErrInvalidPlay, "ninetynine.errBuryCount", map[string]string{"count": "3"})
	})

	t.Run("bury card index out of range", func(t *testing.T) {
		o := newTestNinetyNine()
		o.SetPhase(domain.NinetyNinePhaseBid)
		o.SetBidPlayerIdx(0)
		for i := 0; i < 3; i++ {
			o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7+i, false))
		}
		assertNinetyNineDomainError(t, o.PlayerBid([]int{0, 1, 3}), domain.ErrInvalidCard, "ninetynine.errCardIndexOutOfRange", nil)
	})

	t.Run("duplicate bury card index", func(t *testing.T) {
		o := newTestNinetyNine()
		o.SetPhase(domain.NinetyNinePhaseBid)
		o.SetBidPlayerIdx(0)
		for i := 0; i < 3; i++ {
			o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7+i, false))
		}
		assertNinetyNineDomainError(t, o.PlayerBid([]int{0, 0, 1}), domain.ErrInvalidPlay, "ninetynine.errDuplicateCardIndex", nil)
	})
}
