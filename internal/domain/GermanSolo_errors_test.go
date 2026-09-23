//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertGermanSoloCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestGermanSoloDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newTestGermanSolo()
	assertGermanSoloCodedError(t, g.PlayerBid(domain.GermanSoloBid(99), -1), domain.ErrInvalidPlay, "germansolo.errBidTooLow")

	g = newTestGermanSolo()
	assertGermanSoloCodedError(t, g.PlayerBid(domain.GermanSoloBidFrage, 0), domain.ErrInvalidPlay, "germansolo.errInvalidTrumpSuit")

	g = newTestGermanSolo()
	g.SetPhase(domain.GermanSoloPhaseAceCall)
	g.SetDeclarerIdx(0)
	assertGermanSoloCodedError(t, g.DeclareAce(1, domain.CardDesignSpade), domain.ErrInvalidPlay, "germansolo.errAceCallerOnly")

	g = newTestGermanSolo()
	g.SetPhase(domain.GermanSoloPhaseAceCall)
	g.SetDeclarerIdx(0)
	assertGermanSoloCodedError(t, g.DeclareAce(0, 0), domain.ErrInvalidCard, "germansolo.errInvalidSuit")

	g = newTestGermanSolo()
	g.SetPhase(domain.GermanSoloPhaseAceCall)
	g.SetDeclarerIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	setGermanSoloHand(g, 0, germanSoloCard(domain.CardDesignSpade, 1))
	assertGermanSoloCodedError(t, g.DeclareAce(0, domain.CardDesignSpade), domain.ErrInvalidPlay, "germansolo.errOwnAceCannotBeCalled")

	g = newTestGermanSolo()
	g.SetPhase(domain.GermanSoloPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertGermanSoloCodedError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "germansolo.errCardIndexOutOfRange")

	g = newTestGermanSolo()
	g.SetPhase(domain.GermanSoloPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	setGermanSoloHand(g, 0, germanSoloCard(domain.CardDesignClover, 7), germanSoloCard(domain.CardDesignHeart, 7))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: germanSoloCard(domain.CardDesignClover, 8)}})
	assertGermanSoloCodedError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "germansolo.errFollowLeadSuit")
}
