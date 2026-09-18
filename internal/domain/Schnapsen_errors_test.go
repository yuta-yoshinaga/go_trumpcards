//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertSchnapsenCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestSchnapsenDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("card index out of range", func(t *testing.T) {
		s := newTestSchnapsen()
		s.SetCurrentPlayerIdx(0)
		assertSchnapsenCodedError(t, s.PlayerPlay(-1), domain.ErrInvalidCard, "schnapsen.errCardIndexOutOfRange")
	})

	t.Run("marriage only when leading", func(t *testing.T) {
		s := newTestSchnapsen()
		s.SetCurrentPlayerIdx(0)
		s.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: schnCard(domain.CardDesignSpade, 1)}})
		assertSchnapsenCodedError(t, s.PlayerDeclareMarriage(0), domain.ErrInvalidPlay, "schnapsen.errMarriageLeadOnly")
	})

	t.Run("marriage unavailable for card", func(t *testing.T) {
		s := newTestSchnapsen()
		s.SetCurrentPlayerIdx(0)
		s.SetCurrentTrick(nil)
		schnSetHand(s.GetPlayer(0), schnCard(domain.CardDesignSpade, 1))
		assertSchnapsenCodedError(t, s.PlayerDeclareMarriage(0), domain.ErrInvalidPlay, "schnapsen.errMarriageUnavailable")
	})
}
