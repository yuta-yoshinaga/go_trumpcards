//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertDoubtDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestDoubtDomainErrorsHaveMessageCodes(t *testing.T) {
	tests := []struct {
		name     string
		play     func(*domain.Doubt) error
		sentinel error
		code     string
	}{
		{
			name:     "claimed value out of range",
			play:     func(g *domain.Doubt) error { return g.PlayerPlay([]int{0}, 0, 0) },
			sentinel: domain.ErrInvalidPlay,
			code:     "doubt.errClaimedValueOutOfRange",
		},
		{
			name:     "no cards specified",
			play:     func(g *domain.Doubt) error { return g.PlayerPlay(nil, 1, 0) },
			sentinel: domain.ErrInvalidPlay,
			code:     "doubt.errNoCardsSpecified",
		},
		{
			name:     "duplicate card index",
			play:     func(g *domain.Doubt) error { return g.PlayerPlay([]int{0, 0}, 1, 0) },
			sentinel: domain.ErrInvalidCard,
			code:     "doubt.errDuplicateCardIndex",
		},
		{
			name:     "card index out of range",
			play:     func(g *domain.Doubt) error { return g.PlayerPlay([]int{1}, 1, 0) },
			sentinel: domain.ErrInvalidCard,
			code:     "doubt.errCardIndexOutOfRange",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, players := makeDoubtGame()
			g.SetPhase(domain.DoubtPhasePlay)
			players[0].Reset()
			players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
			assertDoubtDomainError(t, tt.play(g), tt.sentinel, tt.code)
		})
	}
}
