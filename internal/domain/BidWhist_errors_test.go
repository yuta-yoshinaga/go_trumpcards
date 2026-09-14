//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertBidWhistDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	actualCode, actualParams := domain.ErrorMessageCode(err)
	assert.Equal(t, code, actualCode)
	assert.Equal(t, params, actualParams)
}

func newBidWhistErrorGame() *domain.BidWhist {
	players := []*domain.BidWhistPlayer{
		domain.NewBidWhistPlayer(true, 0), domain.NewBidWhistPlayer(false, 1),
		domain.NewBidWhistPlayer(false, 0), domain.NewBidWhistPlayer(false, 1),
	}
	g := domain.NewBidWhist(domain.NewTrumpCards(2), players, domain.DefaultBidWhistConfig())
	g.SetBidPlayerIdx(0)
	return g
}

func TestBidWhistDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("invalid bid", func(t *testing.T) {
		g := newBidWhistErrorGame()
		g.SetPhase(domain.BidWhistPhaseBid)
		assertBidWhistDomainError(t, g.PlayerBid(0, domain.BidWhistDirectionUptown), domain.ErrInvalidPlay, "bidwhist.errInvalidBid", nil)
	})

	t.Run("bid higher than highest", func(t *testing.T) {
		g := newBidWhistErrorGame()
		g.SetPhase(domain.BidWhistPhaseBid)
		g.SetDealerIdx(3)
		assert.NoError(t, g.PlayerBid(1, domain.BidWhistDirectionUptown))
		g.SetBidPlayerIdx(0)
		assertBidWhistDomainError(t, g.PlayerBid(1, domain.BidWhistDirectionUptown), domain.ErrInvalidPlay, "bidwhist.errBidHigherThanHighest", map[string]string{"bid": "1"})
	})

	t.Run("invalid trump suit", func(t *testing.T) {
		g := newBidWhistErrorGame()
		g.SetPhase(domain.BidWhistPhaseTrumpDeclaration)
		g.SetDeclarerIdx(0)
		assertBidWhistDomainError(t, g.PlayerDeclareTrump(99), domain.ErrInvalidPlay, "bidwhist.errInvalidSuit", nil)
	})

	t.Run("kitty exchange", func(t *testing.T) {
		g := newBidWhistErrorGame()
		g.SetPhase(domain.BidWhistPhaseKittyExchange)
		g.SetDeclarerIdx(0)
		for i := 0; i < 6; i++ {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2+i, false))
		}
		assertBidWhistDomainError(t, g.PlayerExchangeKitty([]int{0}), domain.ErrInvalidCard, "bidwhist.errKittyCardCount", map[string]string{"count": "6"})
		assertBidWhistDomainError(t, g.PlayerExchangeKitty([]int{0, 1, 2, 3, 4, 6}), domain.ErrInvalidCard, "bidwhist.errCardIndexOutOfRange", nil)
		assertBidWhistDomainError(t, g.PlayerExchangeKitty([]int{0, 1, 2, 3, 4, 4}), domain.ErrInvalidCard, "bidwhist.errSameCard", nil)
	})

	t.Run("play card index", func(t *testing.T) {
		g := newBidWhistErrorGame()
		g.SetPhase(domain.BidWhistPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		assertBidWhistDomainError(t, g.PlayerPlay(1), domain.ErrInvalidCard, "bidwhist.errCardIndexOutOfRange", nil)
	})

	t.Run("no trump joker lead", func(t *testing.T) {
		g := newBidWhistErrorGame()
		g.SetPhase(domain.BidWhistPhasePlay)
		g.SetContract(1, domain.BidWhistDirectionNoTrump, -1)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		assertBidWhistDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "bidwhist.errNoTrumpJokerLead", nil)
	})

	t.Run("follow lead suit", func(t *testing.T) {
		g := newBidWhistErrorGame()
		g.SetPhase(domain.BidWhistPhasePlay)
		g.SetContract(1, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
		g.SetTrumpSuit(domain.CardDesignSpade)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 3, false)}})
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		assertBidWhistDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "bidwhist.errFollowLeadSuit", nil)
	})
}
