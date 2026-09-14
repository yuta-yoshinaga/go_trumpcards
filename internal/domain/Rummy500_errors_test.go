//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertRummy500DomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func newRummy500ErrorGame(phase domain.Rummy500Phase) *domain.Rummy500 {
	g := domain.NewDefaultRummy500()
	g.Reset()
	g.SetPhase(phase)
	g.SetCurrentPlayerIdx(0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).ResetRound()
	}
	return g
}

func TestRummy500DomainErrorsHaveMessageCodes(t *testing.T) {
	g := newRummy500ErrorGame(domain.Rummy500PhaseDraw)
	g.SetDiscardPile(nil)
	assertRummy500DomainError(t, g.PlayerDrawFromDiscard(0), domain.ErrInvalidPlay, "rummy500.errDiscardPileEmpty")
	g.SetDiscardPile([]*domain.Card{domain.NewCard(1, 1, false)})
	assertRummy500DomainError(t, g.PlayerDrawFromDiscard(-1), domain.ErrInvalidCard, "rummy500.errCardIndexOutOfRange")

	g = newRummy500ErrorGame(domain.Rummy500PhasePlay)
	assertRummy500DomainError(t, g.PlayerMeld(nil), domain.ErrInvalidPlay, "rummy500.errMeldMinimumCards")
	g.GetPlayer(0).AddCard(domain.NewCard(1, 1, false))
	assertRummy500DomainError(t, g.PlayerMeld([]int{0, 1, 2}), domain.ErrInvalidCard, "rummy500.errCardIndexOutOfRange")
	g.GetPlayer(0).AddCard(domain.NewCard(1, 2, false))
	assertRummy500DomainError(t, g.PlayerMeld([]int{0, 1, 1}), domain.ErrInvalidCard, "rummy500.errDuplicateCardIndex")
	g.GetPlayer(0).AddCard(domain.NewCard(2, 7, false))
	assertRummy500DomainError(t, g.PlayerMeld([]int{0, 1, 2}), domain.ErrInvalidPlay, "rummy500.errInvalidMeld")

	assertRummy500DomainError(t, g.PlayerLayoff(-1, 0, 0), domain.ErrInvalidPlay, "rummy500.errTargetPlayerInvalid")
	g.GetPlayer(1).SetLaidMelds([][]*domain.Card{{domain.NewCard(1, 5, false), domain.NewCard(2, 5, false), domain.NewCard(3, 5, false)}})
	assertRummy500DomainError(t, g.PlayerLayoff(1, 1, 0), domain.ErrInvalidPlay, "rummy500.errTargetMeldInvalid")
	assertRummy500DomainError(t, g.PlayerLayoff(1, 0, -1), domain.ErrInvalidCard, "rummy500.errCardIndexOutOfRange")
	g.GetPlayer(0).AddCard(domain.NewCard(4, 9, false))
	assertRummy500DomainError(t, g.PlayerLayoff(1, 0, 0), domain.ErrInvalidPlay, "rummy500.errLayoffCardCannotAdd")
	assertRummy500DomainError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "rummy500.errCardIndexOutOfRange")
}
