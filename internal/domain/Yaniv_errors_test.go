//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertYanivDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestYanivDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("Yaniv call too high", func(t *testing.T) {
		g := newTestYaniv()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(YanivPhaseDiscard)
		setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 10, false))
		assertYanivDomainError(t, g.PlayerDeclareYaniv(), ErrInvalidPlay, "yaniv.errYanivCallTooHigh", map[string]string{"threshold": "5"})
	})

	t.Run("empty pickup", func(t *testing.T) {
		g := newTestYaniv()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(YanivPhaseDraw)
		g.SetPickupCards(nil)
		assertYanivDomainError(t, g.PlayerDrawFromPickup(0), ErrInvalidPlay, "yaniv.errPickupEmpty", nil)
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newTestYaniv()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(YanivPhaseDiscard)
		setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 8, false))
		assertYanivDomainError(t, g.PlayerDiscard([]int{1}), ErrInvalidCard, "yaniv.errCardIndexOutOfRange", nil)
	})

	t.Run("invalid combo", func(t *testing.T) {
		g := newTestYaniv()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(YanivPhaseDiscard)
		setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 8, false), NewCard(CardDesignHeart, 3, false))
		assertYanivDomainError(t, g.PlayerDiscard([]int{0, 1}), ErrInvalidPlay, "yaniv.errInvalidCombo", nil)
	})
}
