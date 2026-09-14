//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertBoliviaError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func newErrorBolivia() *domain.Bolivia {
	players := make([]*domain.BoliviaPlayer, domain.BoliviaPlayerCnt)
	for i := range players {
		players[i] = domain.NewBoliviaPlayer(i == 0, i%domain.BoliviaTeamCnt)
	}
	return domain.NewBolivia(domain.NewTrumpCardsWithDecks(3, 6), players, domain.DefaultBoliviaConfig())
}

func boliviaErrorGame(phase domain.BoliviaPhase) (*domain.Bolivia, *domain.BoliviaPlayer) {
	g := newErrorBolivia()
	g.Reset()
	g.SetPhase(phase)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	p.Reset()
	return g, p
}

func TestBolivia_ValidationErrorsHaveCodes(t *testing.T) {
	card := func(d, v int) *domain.Card { return domain.NewCard(d, v, false) }
	pair := func(p *domain.BoliviaPlayer, v int) {
		p.AddCard(card(domain.CardDesignHeart, v))
		p.AddCard(card(domain.CardDesignDiamond, v))
	}
	for _, tc := range []struct {
		name    string
		pile    []*domain.Card
		indices []int
		code    string
	}{
		{"empty", nil, nil, "bolivia.errDiscardPileEmpty"},
		{"black three", []*domain.Card{card(domain.CardDesignSpade, 3)}, []int{0, 1}, "bolivia.errBlackThreeCannotTakeDiscardPile"},
		{"wild", []*domain.Card{card(domain.CardDesignJoker, 1)}, []int{0, 1}, "bolivia.errWildCardCannotTakeDiscardPile"},
		{"pair count", []*domain.Card{card(domain.CardDesignSpade, 7)}, nil, "bolivia.errNaturalPairIndicesRequired"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := boliviaErrorGame(domain.BoliviaPhaseDraw)
			g.SetDiscardPile(tc.pile)
			assertBoliviaError(t, g.PlayerDrawFromDiscard(tc.indices), domain.ErrInvalidPlay, tc.code)
		})
	}
	g, p := boliviaErrorGame(domain.BoliviaPhaseDraw)
	g.SetDiscardPile([]*domain.Card{card(domain.CardDesignSpade, 7)})
	pair(p, 7)
	assertBoliviaError(t, g.PlayerDrawFromDiscard([]int{-1, 1}), domain.ErrInvalidCard, "bolivia.errCardIndexOutOfRange")
	g, p = boliviaErrorGame(domain.BoliviaPhaseDraw)
	g.SetDiscardPile([]*domain.Card{card(domain.CardDesignSpade, 7)})
	pair(p, 7)
	assertBoliviaError(t, g.PlayerDrawFromDiscard([]int{0, 0}), domain.ErrInvalidCard, "bolivia.errSameCard")
	g, p = boliviaErrorGame(domain.BoliviaPhaseDraw)
	g.SetDiscardPile([]*domain.Card{card(domain.CardDesignSpade, 7)})
	p.AddCard(card(domain.CardDesignJoker, 1))
	p.AddCard(card(domain.CardDesignHeart, 7))
	assertBoliviaError(t, g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrInvalidPlay, "bolivia.errPairMustBeNatural")
	g, p = boliviaErrorGame(domain.BoliviaPhaseDraw)
	g.SetDiscardPile([]*domain.Card{card(domain.CardDesignSpade, 7)})
	pair(p, 8)
	assertBoliviaError(t, g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrInvalidPlay, "bolivia.errPairRankMismatch")
	g, p = boliviaErrorGame(domain.BoliviaPhaseDraw)
	g.SetDiscardPile([]*domain.Card{card(domain.CardDesignSpade, 4)})
	pair(p, 4)
	assertBoliviaError(t, g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrInvalidPlay, "bolivia.errInitialMeldMinimumNotMet")

	g, p = boliviaErrorGame(domain.BoliviaPhaseMeld)
	p.AddCard(card(domain.CardDesignSpade, 7))
	assertBoliviaError(t, g.PlayerMeld([][]int{{-1, 0, 1}}), domain.ErrInvalidCard, "bolivia.errCardIndexOutOfRange")
	g, p = boliviaErrorGame(domain.BoliviaPhaseMeld)
	p.AddCard(card(domain.CardDesignSpade, 7))
	p.AddCard(card(domain.CardDesignHeart, 7))
	assertBoliviaError(t, g.PlayerMeld([][]int{{0, 1}}), domain.ErrInvalidPlay, "bolivia.errMeldNeedsAtLeastThreeCards")
	g, p = boliviaErrorGame(domain.BoliviaPhaseMeld)
	pair(p, 7)
	p.AddCard(card(domain.CardDesignDiamond, 7))
	assertBoliviaError(t, g.PlayerMeld([][]int{{0, 1, 2}, {0, 1, 2}}), domain.ErrInvalidCard, "bolivia.errDuplicateCardIndex")
	g, p = boliviaErrorGame(domain.BoliviaPhaseMeld)
	p.AddCard(card(domain.CardDesignSpade, 7))
	p.AddCard(card(domain.CardDesignSpade, 8))
	p.AddCard(card(domain.CardDesignSpade, 7))
	assertBoliviaError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "bolivia.errSequenceMeldCannotDuplicateCard")
	g, p = boliviaErrorGame(domain.BoliviaPhaseMeld)
	p.SetHasInitMeld(true)
	p.AddCard(card(domain.CardDesignSpade, 3))
	p.AddCard(card(domain.CardDesignHeart, 3))
	p.AddCard(card(domain.CardDesignClover, 3))
	assertBoliviaError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "bolivia.errBlackThreeCannotMeld")
	g, p = boliviaErrorGame(domain.BoliviaPhaseMeld)
	p.AddCard(card(domain.CardDesignSpade, 4))
	p.AddCard(card(domain.CardDesignSpade, 4))
	p.AddCard(card(domain.CardDesignSpade, 4))
	assertBoliviaError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "bolivia.errInitialMeldMinimumNotMet")
	for _, tc := range []struct {
		name, code string
		cards      []*domain.Card
	}{
		{"four wild", "bolivia.errMeldAllowsAtMostThreeWildCards", []*domain.Card{card(domain.CardDesignSpade, 7), card(domain.CardDesignHeart, 7), card(domain.CardDesignJoker, 1), card(domain.CardDesignJoker, 2), card(domain.CardDesignJoker, 3), card(domain.CardDesignJoker, 4)}},
		{"wild exceeds natural", "bolivia.errWildCardsCannotExceedNaturalCards", []*domain.Card{card(domain.CardDesignSpade, 7), card(domain.CardDesignHeart, 7), card(domain.CardDesignJoker, 1), card(domain.CardDesignJoker, 2), card(domain.CardDesignJoker, 3)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, p := boliviaErrorGame(domain.BoliviaPhaseMeld)
			p.SetHasInitMeld(true)
			for _, c := range tc.cards {
				p.AddCard(c)
			}
			idx := make([]int, len(tc.cards))
			for i := range idx {
				idx[i] = i
			}
			assertBoliviaError(t, g.PlayerMeld([][]int{idx}), domain.ErrInvalidPlay, tc.code)
		})
	}

	g, p = boliviaErrorGame(domain.BoliviaPhaseMeld)
	p.SetHasInitMeld(true)
	p.SetMelds([]*domain.BoliviaMeld{{Cards: []*domain.Card{card(domain.CardDesignSpade, 7), card(domain.CardDesignHeart, 7)}, Kind: domain.BoliviaMeldSet, IsNatural: true}})
	p.AddCard(card(domain.CardDesignDiamond, 7))
	p.AddCard(card(domain.CardDesignDiamond, 8))
	assertBoliviaError(t, g.PlayerMeld([][]int{{0, 1}}), domain.ErrInvalidPlay, "bolivia.errSequenceNeedsAtLeastThreeCards")
	g, p = boliviaErrorGame(domain.BoliviaPhaseDraw)
	g.SetDiscardPile([]*domain.Card{card(domain.CardDesignSpade, 7)})
	p.SetHasInitMeld(true)
	pair(p, 7)
	require.NoError(t, g.PlayerDrawFromDiscard([]int{0, 1}))
	assertBoliviaError(t, g.PlayerSkipMeld(), domain.ErrInvalidPlay, "bolivia.errTopCardMustBeMelded")
	g, p = boliviaErrorGame(domain.BoliviaPhaseDiscard)
	assertBoliviaError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "bolivia.errCardIndexOutOfRange")
	p.AddCard(card(domain.CardDesignHeart, 3))
	assertBoliviaError(t, g.PlayerDiscard(0), domain.ErrInvalidPlay, "bolivia.errRedThreeCannotBeDiscarded")
	g, p = boliviaErrorGame(domain.BoliviaPhaseDiscard)
	p.AddCard(card(domain.CardDesignHeart, 4))
	p.AddCard(card(domain.CardDesignHeart, 5))
	assertBoliviaError(t, g.PlayerGoOut(), domain.ErrInvalidPlay, "bolivia.errCompletedMeldsRequiredToGoOut")
	_ = p
}

func TestBolivia_SequenceErrorCodes(t *testing.T) {
	card := func(d, v int) *domain.Card { return domain.NewCard(d, v, false) }
	g := newErrorBolivia()
	for _, tc := range []struct {
		name, code string
		cards      []*domain.Card
	}{
		{"three", "bolivia.errThreeCannotBeUsedInSequence", []*domain.Card{card(domain.CardDesignSpade, 3), card(domain.CardDesignSpade, 4), card(domain.CardDesignSpade, 5)}},
		{"duplicate", "bolivia.errSequenceMeldCannotDuplicateCard", []*domain.Card{card(domain.CardDesignSpade, 4), card(domain.CardDesignSpade, 4), card(domain.CardDesignSpade, 5)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g.Reset()
			g.SetPhase(domain.BoliviaPhaseMeld)
			p := g.GetPlayer(0)
			p.Reset()
			p.SetHasInitMeld(true)
			for _, c := range tc.cards {
				p.AddCard(c)
			}
			idx := make([]int, len(tc.cards))
			for i := range idx {
				idx[i] = i
			}
			assertBoliviaError(t, g.PlayerMeld([][]int{idx}), domain.ErrInvalidPlay, tc.code)
		})
	}
}
