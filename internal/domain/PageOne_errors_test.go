//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func pageOneErrorGame(t *testing.T) *domain.PageOne {
	t.Helper()
	g := domain.NewDefaultPageOne()
	g.Reset()
	g.GetPlayer(0).ResetRound()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.PageOnePhasePlay)
	return g
}

func assertPageOneError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	assert.Equal(t, code, err.(*domain.DomainError).MessageCode())
}

func TestPageOneActionErrorsHaveMessageCodes(t *testing.T) {
	t.Run("card index out of range", func(t *testing.T) {
		g := pageOneErrorGame(t)
		assertPageOneError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "pageone.errCardIndexOutOfRange")
	})

	t.Run("card cannot be played", func(t *testing.T) {
		g := pageOneErrorGame(t)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		assertPageOneError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "pageone.errInvalidPlay")
	})

	t.Run("cannot draw with a playable card", func(t *testing.T) {
		g := pageOneErrorGame(t)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		assertPageOneError(t, g.PlayerDraw(), domain.ErrInvalidPlay, "pageone.errCannotDrawWithPlayableCard")
	})
}
