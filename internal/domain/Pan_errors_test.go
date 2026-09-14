//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertPanDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func newPanErrorGame(phase domain.PanPhase) *domain.Pan {
	g := domain.NewDefaultPan()
	g.Reset()
	g.SetPhase(phase)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).ResetRound()
	return g
}

func TestPanDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newPanErrorGame(domain.PanPhaseDraw)
	g.SetDiscardPile(nil)
	assertPanDomainError(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay, "pan.errDiscardPileEmpty")

	g = newPanErrorGame(domain.PanPhasePlay)
	assertPanDomainError(t, g.PlayerMeld(nil), domain.ErrInvalidPlay, "pan.errMeldNeedsAtLeastThreeCards")
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	assertPanDomainError(t, g.PlayerMeld([]int{0, 1, 2}), domain.ErrInvalidCard, "pan.errCardIndexOutOfRange")
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	assertPanDomainError(t, g.PlayerMeld([]int{0, 1, 1}), domain.ErrInvalidCard, "pan.errDuplicateCardIndex")
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	assertPanDomainError(t, g.PlayerMeld([]int{0, 1, 2}), domain.ErrInvalidPlay, "pan.errInvalidMeld")

	assertPanDomainError(t, g.PlayerLayoff(-1, 0, 0), domain.ErrInvalidPlay, "pan.errTargetPlayerInvalid")
	g.GetPlayer(1).SetLaidMelds([][]*domain.Card{{domain.NewCard(domain.CardDesignSpade, 5, false), domain.NewCard(domain.CardDesignHeart, 5, false), domain.NewCard(domain.CardDesignClover, 5, false)}})
	assertPanDomainError(t, g.PlayerLayoff(1, 1, 0), domain.ErrInvalidPlay, "pan.errTargetMeldInvalid")
	assertPanDomainError(t, g.PlayerLayoff(1, 0, -1), domain.ErrInvalidCard, "pan.errCardIndexOutOfRange")
	assertPanDomainError(t, g.PlayerLayoff(1, 0, 0), domain.ErrInvalidPlay, "pan.errLayoffCardCannotAdd")
	assertPanDomainError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "pan.errCardIndexOutOfRange")
}
