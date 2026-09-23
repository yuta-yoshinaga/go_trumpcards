//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertTarneebDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func newTarneebErrorTest() *domain.Tarneeb {
	players := []*domain.TarneebPlayer{
		domain.NewTarneebPlayer(true, 0), domain.NewTarneebPlayer(false, 1),
		domain.NewTarneebPlayer(false, 0), domain.NewTarneebPlayer(false, 1),
	}
	return domain.NewTarneeb(domain.NewTrumpCards(0), players, domain.DefaultTarneebConfig())
}

func TestTarneebDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("bid range", func(t *testing.T) {
		tn := newTarneebErrorTest()
		tn.SetBidPlayerIdx(0)
		assertTarneebDomainError(t, tn.PlayerBid(6), domain.ErrInvalidPlay, "tarneeb.errBidRange", map[string]string{"min": "7", "max": "13"})
	})

	t.Run("bid must exceed highest", func(t *testing.T) {
		tn := newTarneebErrorTest()
		tn.SetBidPlayerIdx(0)
		tn.SetHighestBid(9)
		assertTarneebDomainError(t, tn.PlayerBid(9), domain.ErrInvalidPlay, "tarneeb.errBidHigherThanHighest", map[string]string{"bid": "9"})
	})

	t.Run("invalid trump suit", func(t *testing.T) {
		tn := newTarneebErrorTest()
		tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
		tn.SetBidWinnerIdx(0)
		assertTarneebDomainError(t, tn.PlayerDeclareTrump(0), domain.ErrInvalidPlay, "tarneeb.errInvalidTrumpSuit", nil)
	})

	t.Run("card index out of range", func(t *testing.T) {
		tn := newTarneebErrorTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetCurrentPlayerIdx(0)
		assertTarneebDomainError(t, tn.PlayerPlay(-1), domain.ErrInvalidCard, "tarneeb.errCardIndexOutOfRange", nil)
	})
}
