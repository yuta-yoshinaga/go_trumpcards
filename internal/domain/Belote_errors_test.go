//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertBeloteDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func TestBeloteDomainErrorsHaveMessageCodes(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetPhase(domain.BelotePhaseBidCallTrump)
	b.SetBidPlayerIdx(0)
	b.SetFaceUpCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	assertBeloteDomainError(t, b.PlayerCallTrump(domain.CardDesignHeart), domain.ErrInvalidPlay, "belote.errFaceUpSuit")
	assertBeloteDomainError(t, b.PlayerCallTrump(99), domain.ErrInvalidPlay, "belote.errInvalidSuit")

	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	assertBeloteDomainError(t, b.PlayerPlay(-1), domain.ErrInvalidCard, "belote.errCardIndexOutOfRange")

	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 13, false)}})
	setupBeloteHand(b, 0, []*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false), domain.NewCard(domain.CardDesignSpade, 7, false)})
	assertBeloteDomainError(t, b.PlayerPlay(1), domain.ErrInvalidPlay, "belote.errMustFollowTrump")

	b.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 1, false)}})
	setupBeloteHand(b, 0, []*domain.Card{domain.NewCard(domain.CardDesignClover, 7, false), domain.NewCard(domain.CardDesignHeart, 7, false)})
	assertBeloteDomainError(t, b.PlayerPlay(0), domain.ErrInvalidPlay, "belote.errMustPlayTrump")

	b.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 1, false)}})
	setupBeloteHand(b, 0, []*domain.Card{domain.NewCard(domain.CardDesignClover, 7, false), domain.NewCard(domain.CardDesignSpade, 7, false)})
	assertBeloteDomainError(t, b.PlayerPlay(0), domain.ErrInvalidPlay, "belote.errFollowLeadSuit")

	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
	})
	setupBeloteHand(b, 0, []*domain.Card{domain.NewCard(domain.CardDesignHeart, 8, false), domain.NewCard(domain.CardDesignHeart, 11, false)})
	assertBeloteDomainError(t, b.PlayerPlay(0), domain.ErrInvalidPlay, "belote.errMustOvertrump")
}
