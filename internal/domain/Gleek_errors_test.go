//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertGleekDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func newGleekErrorGame() *domain.Gleek {
	players := []*domain.GleekPlayer{
		domain.NewGleekPlayer(true),
		domain.NewGleekPlayer(false),
		domain.NewGleekPlayer(false),
	}
	return domain.NewGleek(domain.NewTrumpCards(0), players, domain.DefaultGleekConfig())
}

func TestGleekDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("invalid bid", func(t *testing.T) {
		g := newGleekErrorGame()
		g.SetPhase(domain.GleekPhaseBid)
		assertGleekDomainError(t, g.PlayerBid(14), domain.ErrInvalidPlay, "gleek.errInvalidBid", nil)
	})

	t.Run("discard count", func(t *testing.T) {
		g := newTestGleek()
		g.SetPhase(domain.GleekPhaseDiscard)
		g.SetBuyerIdx(0)
		assertGleekDomainError(t, g.PlayerDiscard(nil), domain.ErrInvalidPlay, "gleek.errDiscardCardCount", map[string]string{"count": "7"})
	})

	t.Run("discard index", func(t *testing.T) {
		g := newTestGleek()
		g.SetPhase(domain.GleekPhaseDiscard)
		g.SetBuyerIdx(0)
		setGleekHand(g, 0,
			gleekCard(domain.CardDesignSpade, 4), gleekCard(domain.CardDesignSpade, 5),
			gleekCard(domain.CardDesignSpade, 6), gleekCard(domain.CardDesignSpade, 7),
			gleekCard(domain.CardDesignSpade, 8), gleekCard(domain.CardDesignSpade, 9),
			gleekCard(domain.CardDesignSpade, 10), gleekCard(domain.CardDesignHeart, 4),
		)
		assertGleekDomainError(t, g.PlayerDiscard([]int{0, 1, 2, 3, 4, 5, 8}), domain.ErrInvalidCard, "gleek.errCardIndexOutOfRange", nil)
	})

	t.Run("duplicate discard", func(t *testing.T) {
		g := newTestGleek()
		g.SetPhase(domain.GleekPhaseDiscard)
		g.SetBuyerIdx(0)
		cards := make([]*domain.Card, 8)
		for i := range cards {
			cards[i] = gleekCard(domain.CardDesignSpade, 4+i)
		}
		setGleekHand(g, 0, cards...)
		assertGleekDomainError(t, g.PlayerDiscard([]int{0, 0, 1, 2, 3, 4, 5}), domain.ErrInvalidPlay, "gleek.errDuplicateDiscard", nil)
	})

	t.Run("play index", func(t *testing.T) {
		g := newTestGleek()
		g.SetPhase(domain.GleekPhasePlay)
		g.SetCurrentPlayerIdx(0)
		setGleekHand(g, 0, gleekCard(domain.CardDesignSpade, 4))
		assertGleekDomainError(t, g.PlayerPlay(1), domain.ErrInvalidCard, "gleek.errCardIndexOutOfRange", nil)
	})

	t.Run("follow lead suit", func(t *testing.T) {
		g := newTestGleek()
		g.SetPhase(domain.GleekPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetTrumpSuit(domain.CardDesignHeart)
		setGleekHand(g, 0, gleekCard(domain.CardDesignClover, 4), gleekCard(domain.CardDesignSpade, 5))
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: gleekCard(domain.CardDesignSpade, 13)}})
		assertGleekDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "gleek.errFollowLeadSuit", nil)
	})
}
