//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertHandAndFootError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}
func handAndFootErrorGame(phase domain.HandAndFootPhase) (*domain.HandAndFoot, *domain.HandAndFootPlayer) {
	g := newTestHandAndFoot()
	g.Reset()
	g.SetPhase(phase)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	p.Reset()
	return g, p
}

func TestHandAndFoot_ValidationErrorsHaveCodes(t *testing.T) {
	c := func(d, v int) *domain.Card { return domain.NewCard(d, v, false) }
	g, _ := handAndFootErrorGame(domain.HandAndFootPhaseDraw)
	g.SetDiscardPile(nil)
	assertHandAndFootError(t, g.PlayerDrawFromDiscard(nil), domain.ErrInvalidPlay, "handandfoot.errDiscardPileEmpty")
	for _, tc := range []struct {
		pile []*domain.Card
		idx  []int
		code string
	}{{[]*domain.Card{c(domain.CardDesignSpade, 3)}, []int{0, 1}, "handandfoot.errBlackThreeCannotTakeDiscardPile"}, {[]*domain.Card{c(domain.CardDesignJoker, 1)}, []int{0, 1}, "handandfoot.errWildCardCannotTakeDiscardPile"}} {
		g, p := handAndFootErrorGame(domain.HandAndFootPhaseDraw)
		g.SetDiscardPile(tc.pile)
		p.AddCard(c(domain.CardDesignHeart, 7))
		p.AddCard(c(domain.CardDesignDiamond, 7))
		assertHandAndFootError(t, g.PlayerDrawFromDiscard(tc.idx), domain.ErrInvalidPlay, tc.code)
	}
	g, p := handAndFootErrorGame(domain.HandAndFootPhaseDraw)
	g.SetDiscardPile([]*domain.Card{c(domain.CardDesignSpade, 7)})
	assertHandAndFootError(t, g.PlayerDrawFromDiscard(nil), domain.ErrInvalidPlay, "handandfoot.errNaturalPairIndicesRequired")
	p.AddCard(c(domain.CardDesignHeart, 7))
	p.AddCard(c(domain.CardDesignDiamond, 7))
	assertHandAndFootError(t, g.PlayerDrawFromDiscard([]int{-1, 1}), domain.ErrInvalidCard, "handandfoot.errCardIndexOutOfRange")
	assertHandAndFootError(t, g.PlayerDrawFromDiscard([]int{0, 0}), domain.ErrInvalidCard, "handandfoot.errSameCard")
	g, p = handAndFootErrorGame(domain.HandAndFootPhaseMeld)
	p.AddCard(c(domain.CardDesignSpade, 7))
	assertHandAndFootError(t, g.PlayerMeld([][]int{{-1, 0, 1}}), domain.ErrInvalidCard, "handandfoot.errCardIndexOutOfRange")
	p.Reset()
	p.AddCard(c(domain.CardDesignSpade, 7))
	p.AddCard(c(domain.CardDesignHeart, 7))
	assertHandAndFootError(t, g.PlayerMeld([][]int{{0, 1}}), domain.ErrInvalidPlay, "handandfoot.errMeldNeedsAtLeastThreeCards")
	p.Reset()
	p.AddCard(c(domain.CardDesignSpade, 7))
	p.AddCard(c(domain.CardDesignHeart, 7))
	p.AddCard(c(domain.CardDesignDiamond, 7))
	assertHandAndFootError(t, g.PlayerMeld([][]int{{0, 1, 2}, {0, 1, 2}}), domain.ErrInvalidCard, "handandfoot.errDuplicateCardIndex")
	g, p = handAndFootErrorGame(domain.HandAndFootPhaseDraw)
	g.SetDiscardPile([]*domain.Card{c(domain.CardDesignSpade, 7)})
	p.AddCard(c(domain.CardDesignHeart, 7))
	p.AddCard(c(domain.CardDesignDiamond, 7))
	require.NoError(t, g.PlayerDrawFromDiscard([]int{0, 1}))
	assertHandAndFootError(t, g.PlayerSkipMeld(), domain.ErrInvalidPlay, "handandfoot.errTopCardMustBeMelded")
	g, p = handAndFootErrorGame(domain.HandAndFootPhaseDiscard)
	assertHandAndFootError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "handandfoot.errCardIndexOutOfRange")
	p.AddCard(c(domain.CardDesignHeart, 3))
	assertHandAndFootError(t, g.PlayerDiscard(0), domain.ErrInvalidPlay, "handandfoot.errRedThreeCannotBeDiscarded")
	g, p = handAndFootErrorGame(domain.HandAndFootPhaseDiscard)
	p.AddCard(c(domain.CardDesignHeart, 4))
	p.AddCard(c(domain.CardDesignHeart, 5))
	assertHandAndFootError(t, g.PlayerGoOut(), domain.ErrInvalidPlay, "handandfoot.errGoOutRequirementsNotMet")
}
