//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertEcarteDomainError(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestEcarteDomainErrorsHaveMessageCodes(t *testing.T) {
	e := newTestEcarte(true)
	e.SetPhase(domain.EcartePhaseExchange)
	e.SetNegStep(domain.EcarteNegElderDiscard)
	e.SetCurrentPlayerIdx(0)
	assertEcarteDomainError(t, e.PlayerDiscard(make([]int, e.GetStockRemaining()+1)), "ecarte.errExchangeExceedsStock")

	e = newTestEcarte(true)
	e.SetPhase(domain.EcartePhaseExchange)
	e.SetNegStep(domain.EcarteNegElderDiscard)
	e.SetCurrentPlayerIdx(0)
	assertEcarteDomainError(t, e.PlayerDiscard([]int{0, 0}), "ecarte.errDiscardIndexInvalid")

	e = newTestEcarte(true)
	e.SetPhase(domain.EcartePhasePlay)
	e.SetCurrentPlayerIdx(0)
	ecSetHand(e.GetPlayer(0), ecCard(domain.CardDesignHeart, 7), ecCard(domain.CardDesignHeart, 13))
	assertEcarteDomainError(t, e.PlayerPlay(-1), "ecarte.errCardIndexOutOfRange")

	e = newTestEcarte(true)
	e.SetPhase(domain.EcartePhasePlay)
	e.SetCurrentPlayerIdx(0)
	ecSetHand(e.GetPlayer(0), nil)
	assertEcarteDomainError(t, e.PlayerPlay(0), "ecarte.errCardNil")

	e = newTestEcarte(true)
	e.SetPhase(domain.EcartePhasePlay)
	e.SetCurrentPlayerIdx(0)
	e.SetTrumpSuit(domain.CardDesignSpade)
	e.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: ecCard(domain.CardDesignHeart, 12)}})
	ecSetHand(e.GetPlayer(0), ecCard(domain.CardDesignHeart, 7), ecCard(domain.CardDesignHeart, 13))
	assertEcarteDomainError(t, e.PlayerPlay(0), "ecarte.errFollowRule")
}
