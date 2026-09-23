//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertChinchonDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestChinchonDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("discard pile empty", func(t *testing.T) {
		g := newTestChinchon(2)
		g.Reset()
		chClearState(g)
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.ChinchonPhaseDraw)
		assertChinchonDomainError(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay, "chinchon.errDiscardPileEmpty", nil)
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newTestChinchon(2)
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.ChinchonPhaseDiscard)
		assertChinchonDomainError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "chinchon.errCardIndexOutOfRange", nil)
	})

	t.Run("knock deadwood too high", func(t *testing.T) {
		g := newTestChinchon(2)
		g.Reset()
		chClearState(g)
		chSetHand(g.GetPlayer(0),
			chCard(domain.CardDesignSpade, 13), chCard(domain.CardDesignHeart, 12),
			chCard(domain.CardDesignDiamond, 11), chCard(domain.CardDesignClover, 7),
			chCard(domain.CardDesignSpade, 5), chCard(domain.CardDesignHeart, 3),
			chCard(domain.CardDesignDiamond, 1))
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.ChinchonPhaseDiscard)
		assertChinchonDomainError(t, g.PlayerKnock(0), domain.ErrInvalidPlay, "chinchon.errKnockDeadwoodTooHigh", map[string]string{"threshold": "5", "deadwood": "36"})
	})

	t.Run("duplicate card index", func(t *testing.T) {
		g := newTestChinchon(2)
		g.Reset()
		chClearState(g)
		chSetHand(g.GetPlayer(0), chCard(domain.CardDesignSpade, 1))
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.ChinchonPhaseLayoff)
		assertChinchonDomainError(t, g.PlayerLayoff([]int{0, 0}), domain.ErrInvalidCard, "chinchon.errDuplicateCardIndex", nil)
	})

	t.Run("layoff card cannot add", func(t *testing.T) {
		g := newTestChinchon(2)
		g.Reset()
		chClearState(g)
		g.SetKnockerMelds([][]*domain.Card{{chCard(domain.CardDesignSpade, 1), chCard(domain.CardDesignSpade, 2), chCard(domain.CardDesignSpade, 3)}})
		chSetHand(g.GetPlayer(1), chCard(domain.CardDesignHeart, 13))
		g.SetCurrentPlayerIdx(1)
		g.SetPhase(domain.ChinchonPhaseLayoff)
		assertChinchonDomainError(t, g.PlayerLayoff([]int{0}), domain.ErrInvalidPlay, "chinchon.errLayoffCardCannotAdd", map[string]string{"card": "♥K"})
	})
}
