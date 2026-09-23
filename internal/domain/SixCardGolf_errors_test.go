//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertSixCardGolfDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func setupSixCardGolfPlayerTurn(g *domain.SixCardGolf) {
	g.Reset()
	for _, pos := range []int{0, 1} {
		_ = g.FlipInitial(pos)
	}
	for _, pos := range []int{0, 1} {
		_ = g.FlipInitial(pos)
	}
}

func TestSixCardGolfDomainErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultSixCardGolf()
	g.Reset()
	assertSixCardGolfDomainError(t, g.FlipInitial(-1), domain.ErrInvalidCard, "sixcardgolf.errCardIndexOutOfRange")
	assert.NoError(t, g.FlipInitial(0))
	assertSixCardGolfDomainError(t, g.FlipInitial(0), domain.ErrInvalidPlay, "sixcardgolf.errAlreadyFaceUp")

	setupSixCardGolfPlayerTurn(g)
	g.SetDrawPile(nil)
	assertSixCardGolfDomainError(t, g.DrawStock(), domain.ErrInvalidPlay, "sixcardgolf.errStockEmpty")
	g.SetDiscardPile(nil)
	assertSixCardGolfDomainError(t, g.DrawDiscard(), domain.ErrInvalidPlay, "sixcardgolf.errDiscardPileEmpty")

	g.SetPhase(domain.SixCardGolfPhaseDrawPending)
	g.SetDrawnCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	assertSixCardGolfDomainError(t, g.SwapCard(-1), domain.ErrInvalidCard, "sixcardgolf.errCardIndexOutOfRange")

	g.SetPhase(domain.SixCardGolfPhasePlayerTurn)
	g.SetCanFlip(false)
	assertSixCardGolfDomainError(t, g.FlipCard(0), domain.ErrInvalidPlay, "sixcardgolf.errCannotFlip")
	g.SetCanFlip(true)
	g.GetPlayer(0).Grid[0].FaceUp = true
	assertSixCardGolfDomainError(t, g.FlipCard(0), domain.ErrInvalidPlay, "sixcardgolf.errAlreadyFaceUp")
	g.SetCanFlip(false)
	assertSixCardGolfDomainError(t, g.SkipFlip(), domain.ErrInvalidPlay, "sixcardgolf.errCannotSkip")
}
