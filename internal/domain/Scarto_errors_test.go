//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertScartoDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func scartoErrorGame(cards ...*domain.Card) *domain.Scarto {
	g := domain.NewDefaultScarto()
	g.Reset()
	g.GetPlayer(0).Reset()
	for _, card := range cards {
		g.GetPlayer(0).AddCard(card)
	}
	return g
}

func TestScartoDomainErrorsHaveMessageCodes(t *testing.T) {
	low := func(value int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, value, false) }
	tests := []struct {
		name     string
		game     *domain.Scarto
		indices  []int
		sentinel error
		code     string
	}{
		{"scarto count", scartoErrorGame(low(2), low(3)), []int{0, 1}, domain.ErrInvalidCard, "scarto.errScartoCount"},
		{"card index", scartoErrorGame(low(2), low(3), low(4)), []int{-1, 1, 2}, domain.ErrInvalidCard, "scarto.errCardIndexOutOfRange"},
		{"same card", scartoErrorGame(low(2), low(3), low(4)), []int{0, 0, 1}, domain.ErrInvalidCard, "scarto.errSameCard"},
		{"invalid card", scartoErrorGame(nil, low(3), low(4)), []int{0, 1, 2}, domain.ErrInvalidCard, "scarto.errScartoInvalidCard"},
		{"excuse", scartoErrorGame(domain.NewCard(domain.ScartoExcuseDesign, domain.ScartoExcuseValue, false), low(3), low(4)), []int{0, 1, 2}, domain.ErrInvalidPlay, "scarto.errDiscardHonour"},
		{"bout", scartoErrorGame(domain.NewCard(domain.ScartoTrumpDesign, domain.ScartoPetitValue, false), low(3), low(4)), []int{0, 1, 2}, domain.ErrInvalidPlay, "scarto.errDiscardHonour"},
		{"trump", scartoErrorGame(low(2), low(3), low(4), domain.NewCard(domain.ScartoTrumpDesign, 2, false)), []int{0, 1, 3}, domain.ErrInvalidPlay, "scarto.errDiscardTrump"},
		{"court", scartoErrorGame(domain.NewCard(domain.CardDesignSpade, domain.ScartoKingValue, false), low(3), low(4)), []int{0, 1, 2}, domain.ErrInvalidPlay, "scarto.errDiscardCourt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScartoDomainError(t, tt.game.PlayerScarto(tt.indices), tt.sentinel, tt.code)
		})
	}
}
