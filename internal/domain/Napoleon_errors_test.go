//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertNapoleonDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func TestNapoleonDomainErrorsHaveMessageCodes(t *testing.T) {
	n := newTestNapoleon()
	n.Reset()
	assertNapoleonDomainError(t, n.PlayerBid(11), domain.ErrInvalidPlay, "napoleon.errBidRange")
	n.SetHighestBid(13)
	assertNapoleonDomainError(t, n.PlayerBid(13), domain.ErrInvalidPlay, "napoleon.errBidHigherThanHighest")

	setupDeclare := func() {
		n.Reset()
		n.SetPhase(domain.NapoleonPhaseTrumpDeclaration)
		n.SetNapoleonIdx(0)
		n.GetPlayer(0).SetIsNapoleon(true)
	}
	setupDeclare()
	assertNapoleonDomainError(t, n.PlayerDeclareTrump(0, domain.CardDesignHeart, 1), domain.ErrInvalidPlay, "napoleon.errInvalidSuit")
	setupDeclare()
	assertNapoleonDomainError(t, n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignJoker, 2), domain.ErrInvalidPlay, "napoleon.errJokerValue")
	setupDeclare()
	assertNapoleonDomainError(t, n.PlayerDeclareTrump(domain.CardDesignSpade, 5, 1), domain.ErrInvalidPlay, "napoleon.errInvalidAdjutantSuit")
	setupDeclare()
	assertNapoleonDomainError(t, n.PlayerDeclareTrump(domain.CardDesignSpade, domain.CardDesignHeart, 0), domain.ErrInvalidPlay, "napoleon.errInvalidAdjutantValue")

	n.Reset()
	n.SetPhase(domain.NapoleonPhaseKittyExchange)
	n.SetNapoleonIdx(0)
	n.GetPlayer(0).SetIsNapoleon(true)
	assertNapoleonDomainError(t, n.PlayerExchangeKitty(-1), domain.ErrInvalidCard, "napoleon.errCardIndexOutOfRange")

	n.Reset()
	n.SetPhase(domain.NapoleonPhasePlay)
	n.SetCurrentPlayerIdx(0)
	n.SetTrickNumber(2)
	n.SetTrumpSuit(domain.CardDesignSpade)
	n.SetNapoleonIdx(0)
	assertNapoleonDomainError(t, n.PlayerPlay(-1), domain.ErrInvalidCard, "napoleon.errCardIndexOutOfRange")
	n.SetCurrentTrick([]*domain.NapoleonTrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)}})
	n.GetPlayer(0).Reset()
	n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	n.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	assertNapoleonDomainError(t, n.PlayerPlay(0), domain.ErrInvalidPlay, "napoleon.errMustFollowLeadSuit")
}
