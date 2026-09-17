//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertTonkDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestTonkDomainErrorsHaveMessageCodes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*domain.Tonk)
		play     func(*domain.Tonk) error
		sentinel error
		code     string
		params   map[string]string
	}{
		{
			name: "discard pile empty",
			setup: func(g *domain.Tonk) {
				setupTonkDrawPhase(g, 0)
				g.SetDiscardPile(nil)
			},
			play:     func(g *domain.Tonk) error { return g.PlayerDrawFromDiscard() },
			sentinel: domain.ErrInvalidPlay,
			code:     "tonk.errDiscardPileEmpty",
		},
		{
			name: "discard card index out of range",
			setup: func(g *domain.Tonk) {
				setupTonkDiscardPhase(g, 0)
				giveHand(g.GetPlayer(0), []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
			},
			play:     func(g *domain.Tonk) error { return g.PlayerDiscard(-1) },
			sentinel: domain.ErrInvalidCard,
			code:     "tonk.errCardIndexOutOfRange",
		},
		{
			name: "knock card index out of range",
			setup: func(g *domain.Tonk) {
				setupTonkDiscardPhase(g, 0)
				giveHand(g.GetPlayer(0), []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
			},
			play:     func(g *domain.Tonk) error { return g.PlayerKnock(-1) },
			sentinel: domain.ErrInvalidCard,
			code:     "tonk.errCardIndexOutOfRange",
		},
		{
			name: "knock deadwood too high",
			setup: func(g *domain.Tonk) {
				setupTonkDiscardPhase(g, 0)
				giveHand(g.GetPlayer(0), []*domain.Card{
					domain.NewCard(domain.CardDesignSpade, 13, false),
					domain.NewCard(domain.CardDesignHeart, 12, false),
				})
			},
			play:     func(g *domain.Tonk) error { return g.PlayerKnock(0) },
			sentinel: domain.ErrInvalidPlay,
			code:     "tonk.errKnockDeadwoodTooHigh",
			params:   map[string]string{"threshold": "5", "deadwood": "10"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestTonk()
			tt.setup(g)
			assertTonkDomainError(t, tt.play(g), tt.sentinel, tt.code, tt.params)
		})
	}
}
