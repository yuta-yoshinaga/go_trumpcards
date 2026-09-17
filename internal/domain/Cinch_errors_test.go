//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertCinchDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestCinchDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("dealer cannot pass when all pass", func(t *testing.T) {
		g := newTestCinch(t, domain.CinchDifficultyEasy)
		g.SetDealerIdx(0)
		g.SetBidPlayerIdx(0)
		g.SetCurrentBid(0)
		for i := 1; i < domain.CinchPlayerCnt; i++ {
			g.GetPlayer(i).SetBid(domain.CinchPassBid)
		}
		assertCinchDomainError(t, g.PlayerBid(domain.CinchPassBid), domain.ErrInvalidPlay, "cinch.errDealerCannotPassWhenAllPass", nil)
	})

	t.Run("bid range", func(t *testing.T) {
		g := newTestCinch(t, domain.CinchDifficultyEasy)
		assertCinchDomainError(t, g.PlayerBid(-1), domain.ErrInvalidPlay, "cinch.errBidRange", map[string]string{"min": "1", "max": "14"})
	})

	t.Run("bid must exceed current", func(t *testing.T) {
		g := newTestCinch(t, domain.CinchDifficultyEasy)
		g.SetCurrentBid(3)
		assertCinchDomainError(t, g.PlayerBid(3), domain.ErrInvalidPlay, "cinch.errBidMustExceedCurrent", map[string]string{"bid": "3"})
	})

	t.Run("trump suit range", func(t *testing.T) {
		g := newTestCinch(t, domain.CinchDifficultyEasy)
		g.SetPhase(domain.CinchPhaseNameTrump)
		g.SetBidWinnerIdx(0)
		assertCinchDomainError(t, g.NameTrump(0), domain.ErrInvalidPlay, "cinch.errTrumpSuitRange", nil)
	})

	t.Run("card index range", func(t *testing.T) {
		g := newTestCinch(t, domain.CinchDifficultyEasy)
		g.SetPhase(domain.CinchPhasePlay)
		g.SetCurrentTurn(0)
		setCinchHand(g, 0, cinchCard(domain.CardDesignSpade, 1))
		assertCinchDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "cinch.errCardIndexOutOfRange", nil)
	})

	t.Run("must follow trump when trump is led", func(t *testing.T) {
		g := newTestCinch(t, domain.CinchDifficultyEasy)
		g.SetPhase(domain.CinchPhasePlay)
		g.SetCurrentTurn(0)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: cinchCard(domain.CardDesignHeart, 9)}})
		setCinchHand(g, 0, cinchCard(domain.CardDesignHeart, 2), cinchCard(domain.CardDesignSpade, 3))
		assertCinchDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "cinch.errFollowTrumpLead", nil)
	})

	t.Run("must follow lead or trump", func(t *testing.T) {
		g := newTestCinch(t, domain.CinchDifficultyEasy)
		g.SetPhase(domain.CinchPhasePlay)
		g.SetCurrentTurn(0)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: cinchCard(domain.CardDesignClover, 9)}})
		setCinchHand(g, 0, cinchCard(domain.CardDesignClover, 2), cinchCard(domain.CardDesignSpade, 3))
		assertCinchDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "cinch.errFollowLeadOrTrump", nil)
	})
}
