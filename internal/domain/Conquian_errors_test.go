//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertConquianDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func TestConquianDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newTestConquian()
	g.Reset()
	g.SetDiscardPile(nil)
	assertConquianDomainError(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay, "conquian.errDiscardPileEmpty")

	g.Reset()
	cqSetHand(g.GetPlayer(0), cqCard(domain.CardDesignSpade, 1), cqCard(domain.CardDesignHeart, 4), cqCard(domain.CardDesignDiamond, 7))
	g.SetDiscardPile([]*domain.Card{cqCard(domain.CardDesignClover, 13)})
	assertConquianDomainError(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay, "conquian.errDiscardCardCannotBeMelded")

	g.Reset()
	g.SetPhase(domain.ConquianPhaseMeld)
	cqSetHand(g.GetPlayer(0), cqCard(domain.CardDesignSpade, 1))
	assertConquianDomainError(t, g.PlayerMeld([][]int{{}}), domain.ErrInvalidPlay, "conquian.errInvalidMeld")
	assertConquianDomainError(t, g.PlayerMeld([][]int{{2}}), domain.ErrInvalidCard, "conquian.errCardIndexOutOfRange")
	assertConquianDomainError(t, g.PlayerMeld([][]int{{0, 0}}), domain.ErrInvalidCard, "conquian.errDuplicateCardIndex")
	assertConquianDomainError(t, g.PlayerMeld([][]int{{0}}), domain.ErrInvalidPlay, "conquian.errInvalidMeld")

	g.Reset()
	cqSetHand(g.GetPlayer(0), cqCard(domain.CardDesignHeart, 5), cqCard(domain.CardDesignDiamond, 5), cqCard(domain.CardDesignClover, 2))
	g.SetDiscardPile([]*domain.Card{cqCard(domain.CardDesignSpade, 5)})
	require.NoError(t, g.PlayerDrawFromDiscard())
	assertConquianDomainError(t, g.PlayerMeld(nil), domain.ErrInvalidPlay, "conquian.errMeldRequired")

	g.Reset()
	g.SetPhase(domain.ConquianPhaseMeld)
	cqSetHand(g.GetPlayer(0), cqCard(domain.CardDesignSpade, 1))
	assertConquianDomainError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "conquian.errCardIndexOutOfRange")
}
