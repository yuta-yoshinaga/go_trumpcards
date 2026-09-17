//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertLooDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestLooDomainErrorsHaveMessageCodes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*domain.Loo)
		play     int
		code     string
		sentinel error
	}{
		{
			name:     "card index out of range",
			setup:    func(g *domain.Loo) { setupLooPlay(g, nil, nil) },
			play:     -1,
			code:     "loo.errCardIndexOutOfRange",
			sentinel: domain.ErrInvalidCard,
		},
		{
			name: "follow lead suit",
			setup: func(g *domain.Loo) {
				setupLooPlay(g, []*domain.Card{
					looCard(domain.CardDesignHeart, 2), looCard(domain.CardDesignSpade, 3),
				}, []*domain.TrickCard{{PlayerIdx: 1, Card: looCard(domain.CardDesignHeart, 9)}})
			},
			play:     1,
			code:     "loo.errFollowLeadSuit",
			sentinel: domain.ErrInvalidPlay,
		},
		{
			name: "must head",
			setup: func(g *domain.Loo) {
				setupLooPlay(g, []*domain.Card{
					looCard(domain.CardDesignHeart, 9), looCard(domain.CardDesignHeart, 12),
				}, []*domain.TrickCard{{PlayerIdx: 1, Card: looCard(domain.CardDesignHeart, 10)}})
			},
			play:     0,
			code:     "loo.errMustHead",
			sentinel: domain.ErrInvalidPlay,
		},
		{
			name: "play trump",
			setup: func(g *domain.Loo) {
				setupLooPlay(g, []*domain.Card{
					looCard(domain.CardDesignSpade, 2), looCard(domain.CardDesignHeart, 3),
				}, []*domain.TrickCard{{PlayerIdx: 1, Card: looCard(domain.CardDesignClover, 9)}})
			},
			play:     0,
			code:     "loo.errPlayTrump",
			sentinel: domain.ErrInvalidPlay,
		},
		{
			name: "must head trump",
			setup: func(g *domain.Loo) {
				setupLooPlay(g, []*domain.Card{
					looCard(domain.CardDesignHeart, 9), looCard(domain.CardDesignHeart, 1),
				}, []*domain.TrickCard{
					{PlayerIdx: 1, Card: looCard(domain.CardDesignClover, 10)},
					{PlayerIdx: 2, Card: looCard(domain.CardDesignHeart, 13)},
				})
			},
			play:     0,
			code:     "loo.errMustHeadTrump",
			sentinel: domain.ErrInvalidPlay,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestLoo(t, domain.LooCpuDifficultyEasy)
			g.SetTrumpSuit(domain.CardDesignHeart)
			tt.setup(g)
			assertLooDomainError(t, g.PlayerPlay(tt.play), tt.sentinel, tt.code)
		})
	}
}

func setupLooPlay(g *domain.Loo, hand []*domain.Card, trick []*domain.TrickCard) {
	g.SetPhase(domain.LooPhasePlay)
	g.SetCurrentTurn(0)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentTrick(trick)
	g.GetPlayer(0).Reset()
	for _, card := range hand {
		g.GetPlayer(0).AddCard(card)
	}
}
