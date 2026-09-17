//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertFrenchTarotDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func newFrenchTarotErrorGame(phase domain.FrenchTarotPhase) *domain.FrenchTarot {
	g := domain.NewDefaultFrenchTarot()
	g.Reset()
	g.SetPhase(phase)
	g.SetCurrentPlayerIdx(0)
	g.SetBidPlayerIdx(0)
	g.SetDeclarerIdx(0)
	g.GetPlayer(0).ResetRound()
	return g
}

func addFrenchTarotCards(g *domain.FrenchTarot, cards ...*domain.Card) {
	for _, card := range cards {
		g.GetPlayer(0).AddCard(card)
	}
}

func frenchTarotDiscardCards() []*domain.Card {
	return []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	}
}

func TestFrenchTarotDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newFrenchTarotErrorGame(domain.FrenchTarotPhaseBid)
	assertFrenchTarotDomainError(t, g.PlayerBid(domain.FrenchTarotBidPass), domain.ErrInvalidPlay, "frenchtarot.errInvalidBid")
	g.SetHighestBid(domain.FrenchTarotBidPetite)
	assertFrenchTarotDomainError(t, g.PlayerBid(domain.FrenchTarotBidPetite), domain.ErrInvalidPlay, "frenchtarot.errBidHigherThanHighest")

	for _, test := range []struct {
		name     string
		cards    []*domain.Card
		pick     []int
		sentinel error
		code     string
	}{
		{"count", nil, []int{0}, domain.ErrInvalidCard, "frenchtarot.errScartoCount"},
		{"index", frenchTarotDiscardCards(), []int{-1, 0, 1, 2, 3, 4}, domain.ErrInvalidCard, "frenchtarot.errCardIndexOutOfRange"},
		{"duplicate", frenchTarotDiscardCards(), []int{0, 1, 2, 3, 4, 4}, domain.ErrInvalidCard, "frenchtarot.errDuplicateCardIndex"},
		{"excuse", append([]*domain.Card{domain.NewCard(domain.FrenchTarotExcuseDesign, domain.FrenchTarotExcuseValue, false)}, frenchTarotDiscardCards()...), []int{0, 1, 2, 3, 4, 5}, domain.ErrInvalidPlay, "frenchtarot.errDiscardExcuse"},
		{"bout", append([]*domain.Card{domain.NewCard(domain.FrenchTarotTrumpDesign, 1, false)}, frenchTarotDiscardCards()...), []int{0, 1, 2, 3, 4, 5}, domain.ErrInvalidPlay, "frenchtarot.errDiscardBout"},
		{"king", append([]*domain.Card{domain.NewCard(domain.CardDesignSpade, domain.FrenchTarotKingValue, false)}, frenchTarotDiscardCards()...), []int{0, 1, 2, 3, 4, 5}, domain.ErrInvalidPlay, "frenchtarot.errDiscardKing"},
		{"trump", append([]*domain.Card{domain.NewCard(domain.FrenchTarotTrumpDesign, 2, false)}, frenchTarotDiscardCards()...), []int{0, 1, 2, 3, 4, 5}, domain.ErrInvalidPlay, "frenchtarot.errDiscardTrump"},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := newFrenchTarotErrorGame(domain.FrenchTarotPhaseChien)
			addFrenchTarotCards(g, test.cards...)
			assertFrenchTarotDomainError(t, g.PlayerDiscard(test.pick), test.sentinel, test.code)
		})
	}

	g = newFrenchTarotErrorGame(domain.FrenchTarotPhasePlay)
	assertFrenchTarotDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "frenchtarot.errCardIndexOutOfRange")
}
