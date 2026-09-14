//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertKoenigrufenDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func newKoenigrufenErrorGame(phase domain.KoenigrufenPhase) *domain.Koenigrufen {
	g := domain.NewDefaultKoenigrufen()
	g.Reset()
	g.SetPhase(phase)
	g.SetCurrentPlayerIdx(0)
	g.SetDeclarerIdx(0)
	g.SetBidPlayerIdx(0)
	g.GetPlayer(0).ResetRound()
	return g
}

func addKoenigrufenCards(g *domain.Koenigrufen, cards ...*domain.Card) {
	for _, card := range cards {
		g.GetPlayer(0).AddCard(card)
	}
}

func TestKoenigrufenDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newKoenigrufenErrorGame(domain.KoenigrufenPhaseBid)
	assertKoenigrufenDomainError(t, g.PlayerBid(domain.KoenigrufenBid(0)), domain.ErrInvalidPlay, "koenigrufen.errInvalidBid")
	g.SetHighestBid(domain.KoenigrufenBidRufer)
	assertKoenigrufenDomainError(t, g.PlayerBid(domain.KoenigrufenBidRufer), domain.ErrInvalidPlay, "koenigrufen.errBidHigherThanHighest")

	g = newKoenigrufenErrorGame(domain.KoenigrufenPhaseCall)
	assertKoenigrufenDomainError(t, g.PlayerCallKing(0), domain.ErrInvalidPlay, "koenigrufen.errInvalidSuit")
	addKoenigrufenCards(g, domain.NewCard(1, domain.KoenigrufenKingValue, false))
	assertKoenigrufenDomainError(t, g.PlayerCallKing(1), domain.ErrInvalidPlay, "koenigrufen.errKingInOwnHand")

	for _, test := range []struct {
		name     string
		cards    []*domain.Card
		pick     []int
		sentinel error
		code     string
	}{
		{"count", nil, []int{0}, domain.ErrInvalidCard, "koenigrufen.errScartoCount"},
		{"index", []*domain.Card{domain.NewCard(1, 1, false)}, []int{-1, 0, 0, 0, 0, 0}, domain.ErrInvalidCard, "koenigrufen.errCardIndexOutOfRange"},
		{"duplicate", []*domain.Card{domain.NewCard(1, 1, false), domain.NewCard(1, 2, false), domain.NewCard(1, 3, false), domain.NewCard(1, 4, false), domain.NewCard(2, 1, false), domain.NewCard(2, 2, false)}, []int{0, 1, 2, 3, 4, 4}, domain.ErrInvalidCard, "koenigrufen.errDuplicateCardIndex"},
		{"king", []*domain.Card{domain.NewCard(1, domain.KoenigrufenKingValue, false), domain.NewCard(1, 1, false), domain.NewCard(1, 2, false), domain.NewCard(1, 3, false), domain.NewCard(2, 1, false), domain.NewCard(2, 2, false)}, []int{0, 1, 2, 3, 4, 5}, domain.ErrInvalidPlay, "koenigrufen.errDiscardKing"},
		{"trull", []*domain.Card{domain.NewCard(domain.KoenigrufenTrumpDesign, 1, false), domain.NewCard(1, 1, false), domain.NewCard(1, 2, false), domain.NewCard(1, 3, false), domain.NewCard(2, 1, false), domain.NewCard(2, 2, false)}, []int{0, 1, 2, 3, 4, 5}, domain.ErrInvalidPlay, "koenigrufen.errDiscardTrull"},
		{"trump", []*domain.Card{domain.NewCard(domain.KoenigrufenTrumpDesign, 2, false), domain.NewCard(1, 1, false), domain.NewCard(1, 2, false), domain.NewCard(1, 3, false), domain.NewCard(2, 1, false), domain.NewCard(2, 2, false), domain.NewCard(2, 3, false)}, []int{0, 1, 2, 3, 4, 5}, domain.ErrInvalidPlay, "koenigrufen.errDiscardTrump"},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := newKoenigrufenErrorGame(domain.KoenigrufenPhaseTalon)
			addKoenigrufenCards(g, test.cards...)
			assertKoenigrufenDomainError(t, g.PlayerDiscard(test.pick), test.sentinel, test.code)
		})
	}

	g = newKoenigrufenErrorGame(domain.KoenigrufenPhasePlay)
	assertKoenigrufenDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "koenigrufen.errCardIndexOutOfRange")
}
