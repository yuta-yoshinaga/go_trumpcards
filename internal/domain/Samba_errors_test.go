//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertSambaError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}
func sambaErrorGame(phase domain.SambaPhase) (*domain.Samba, *domain.SambaPlayer) {
	g := newTestSamba()
	g.Reset()
	g.SetPhase(phase)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	p.Reset()
	return g, p
}

func TestSamba_ValidationErrorsHaveCodes(t *testing.T) {
	c := func(d, v int) *domain.Card { return domain.NewCard(d, v, false) }
	g, _ := sambaErrorGame(domain.SambaPhaseDraw)
	g.SetDiscardPile(nil)
	assertSambaError(t, g.PlayerDrawFromDiscard(nil), domain.ErrInvalidPlay, "samba.errDiscardPileEmpty")
	for _, tc := range []struct {
		pile []*domain.Card
		idx  []int
		code string
	}{{[]*domain.Card{c(domain.CardDesignSpade, 3)}, []int{0, 1}, "samba.errBlackThreeCannotTakeDiscardPile"}, {[]*domain.Card{c(domain.CardDesignJoker, 1)}, []int{0, 1}, "samba.errWildCardCannotTakeDiscardPile"}, {[]*domain.Card{c(domain.CardDesignSpade, 7)}, nil, "samba.errNaturalPairIndicesRequired"}} {
		g, p := sambaErrorGame(domain.SambaPhaseDraw)
		g.SetDiscardPile(tc.pile)
		if tc.idx == nil {
			assertSambaError(t, g.PlayerDrawFromDiscard(nil), domain.ErrInvalidPlay, tc.code)
			continue
		}
		p.AddCard(c(domain.CardDesignHeart, 7))
		p.AddCard(c(domain.CardDesignDiamond, 7))
		assertSambaError(t, g.PlayerDrawFromDiscard(tc.idx), domain.ErrInvalidPlay, tc.code)
	}
	g, p := sambaErrorGame(domain.SambaPhaseDraw)
	g.SetDiscardPile([]*domain.Card{c(domain.CardDesignSpade, 7)})
	p.AddCard(c(domain.CardDesignHeart, 7))
	p.AddCard(c(domain.CardDesignDiamond, 7))
	assertSambaError(t, g.PlayerDrawFromDiscard([]int{-1, 1}), domain.ErrInvalidCard, "samba.errCardIndexOutOfRange")
	assertSambaError(t, g.PlayerDrawFromDiscard([]int{0, 0}), domain.ErrInvalidCard, "samba.errSameCard")
	g, p = sambaErrorGame(domain.SambaPhaseMeld)
	p.AddCard(c(domain.CardDesignSpade, 7))
	assertSambaError(t, g.PlayerMeld([][]int{{-1, 0, 1}}), domain.ErrInvalidCard, "samba.errCardIndexOutOfRange")
	g, p = sambaErrorGame(domain.SambaPhaseMeld)
	p.SetHasInitMeld(true)
	p.AddCard(c(domain.CardDesignSpade, 7))
	p.AddCard(c(domain.CardDesignSpade, 8))
	p.AddCard(c(domain.CardDesignSpade, 7))
	assertSambaError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "samba.errSequenceMeldCannotDuplicateCard")
	g, p = sambaErrorGame(domain.SambaPhaseMeld)
	p.SetHasInitMeld(true)
	p.AddCard(c(domain.CardDesignSpade, 7))
	p.AddCard(c(domain.CardDesignHeart, 7))
	for i := 0; i < 4; i++ {
		p.AddCard(c(domain.CardDesignJoker, i+1))
	}
	assertSambaError(t, g.PlayerMeld([][]int{{0, 1, 2, 3, 4, 5}}), domain.ErrInvalidPlay, "samba.errMeldAllowsAtMostThreeWildCards")
	g, p = sambaErrorGame(domain.SambaPhaseMeld)
	p.SetHasInitMeld(true)
	p.SetMelds([]*domain.SambaMeld{{Cards: []*domain.Card{c(domain.CardDesignSpade, 7), c(domain.CardDesignHeart, 7)}, Kind: domain.SambaMeldSet, IsNatural: true}})
	p.AddCard(c(domain.CardDesignDiamond, 7))
	p.AddCard(c(domain.CardDesignDiamond, 8))
	assertSambaError(t, g.PlayerMeld([][]int{{0, 1}}), domain.ErrInvalidPlay, "samba.errSequenceNeedsAtLeastThreeCards")
	g, p = sambaErrorGame(domain.SambaPhaseDraw)
	g.SetDiscardPile([]*domain.Card{c(domain.CardDesignSpade, 7)})
	p.SetHasInitMeld(true)
	p.AddCard(c(domain.CardDesignHeart, 7))
	p.AddCard(c(domain.CardDesignDiamond, 7))
	require.NoError(t, g.PlayerDrawFromDiscard([]int{0, 1}))
	assertSambaError(t, g.PlayerSkipMeld(), domain.ErrInvalidPlay, "samba.errTopCardMustBeMelded")
	g, p = sambaErrorGame(domain.SambaPhaseDiscard)
	assertSambaError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "samba.errCardIndexOutOfRange")
	p.AddCard(c(domain.CardDesignHeart, 3))
	assertSambaError(t, g.PlayerDiscard(0), domain.ErrInvalidPlay, "samba.errRedThreeCannotBeDiscarded")
}
