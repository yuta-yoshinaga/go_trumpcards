//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertGinRummyDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestGinRummyDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newTestGinRummy()
	setupGinRummyDrawPhase(g, 0)
	g.SetDiscardPile(nil)
	assertGinRummyDomainError(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay, "ginrummy.errDiscardPileEmpty", nil)

	g = newTestGinRummy()
	setupGinRummyDiscardPhase(g, 0)
	assertGinRummyDomainError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "ginrummy.errCardIndexOutOfRange", nil)

	g = newTestGinRummy()
	setupGinRummyDiscardPhase(g, 0)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	err := g.PlayerKnock(0)
	assertGinRummyDomainError(t, err, domain.ErrInvalidPlay, "ginrummy.errKnockDeadwoodTooHigh", map[string]string{
		"threshold": "10",
		"deadwood":  "20",
	})

	g = newTestGinRummy()
	g.SetPhase(domain.GinRummyPhaseLayoff)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	assertGinRummyDomainError(t, g.PlayerLayoff([]int{0, 0}), domain.ErrInvalidCard, "ginrummy.errDuplicateCardIndex", nil)

	g = newTestGinRummy()
	g.SetPhase(domain.GinRummyPhaseLayoff)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	g.SetKnockerMelds([][]*domain.Card{{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignSpade, 4, false),
	}})
	assertGinRummyDomainError(t, g.PlayerLayoff([]int{0}), domain.ErrInvalidPlay, "ginrummy.errLayoffCardCannotAdd", map[string]string{"card": "♥10"})
}
