//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertTongitsDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func tongitsErrorGame() *domain.Tongits {
	g := domain.NewDefaultTongits()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.GetPlayer(0).Reset()
	return g
}

func TestTongitsDomainErrorsHaveMessageCodes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*domain.Tongits)
		action   func(*domain.Tongits) error
		sentinel error
		code     string
	}{
		{"discard pile empty", func(g *domain.Tongits) { g.SetPhase(domain.TongitsPhaseDraw); g.SetDiscardPile(nil) }, func(g *domain.Tongits) error { return g.PlayerDrawFromDiscard() }, domain.ErrInvalidPlay, "tongits.errDiscardPileEmpty"},
		{"discard index", func(g *domain.Tongits) { g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) }, func(g *domain.Tongits) error { return g.PlayerDiscard(-1) }, domain.ErrInvalidCard, "tongits.errCardIndexOutOfRange"},
		{"meld minimum", func(g *domain.Tongits) { g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) }, func(g *domain.Tongits) error { return g.PlayerMeld([]int{0, 1}) }, domain.ErrInvalidPlay, "tongits.errMeldNeedsAtLeastThreeCards"},
		{"invalid meld", func(g *domain.Tongits) {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		}, func(g *domain.Tongits) error { return g.PlayerMeld([]int{0, 1, 2}) }, domain.ErrInvalidPlay, "tongits.errInvalidMeld"},
		{"target player", nil, func(g *domain.Tongits) error { return g.PlayerSapaw(-1, 0, 0) }, domain.ErrInvalidPlay, "tongits.errTargetPlayerInvalid"},
		{"target meld", nil, func(g *domain.Tongits) error { return g.PlayerSapaw(1, 0, 0) }, domain.ErrInvalidPlay, "tongits.errTargetMeldInvalid"},
		{"sapaw card index", func(g *domain.Tongits) {
			g.GetPlayer(1).AppendMeld([]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			})
		}, func(g *domain.Tongits) error { return g.PlayerSapaw(1, 0, 999) }, domain.ErrInvalidCard, "tongits.errCardIndexOutOfRange"},
		{"challenge response count", nil, func(g *domain.Tongits) error { return g.PlayerChallenge(nil) }, domain.ErrInvalidPlay, "tongits.errResponseCount"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tongitsErrorGame()
			if tt.setup != nil {
				tt.setup(g)
			}
			assertTongitsDomainError(t, tt.action(g), tt.sentinel, tt.code)
		})
	}

	g := tongitsErrorGame()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	g.GetPlayer(1).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	})
	assertTongitsDomainError(t, g.PlayerSapaw(1, 0, 0), domain.ErrInvalidPlay, "tongits.errLayoffCardCannotAdd")
}
