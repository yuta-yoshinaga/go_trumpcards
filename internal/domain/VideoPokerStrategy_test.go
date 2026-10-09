//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestRecommendVideoPokerHoldRepresentative(t *testing.T) {
	c := func(s, v int) *domain.Card { return domain.NewCard(s, v, false) }
	tests := []struct {
		name, variant, key string
		hand               []*domain.Card
		keep               []int
	}{
		{"royal draw precedes high pair", "jacksorbetter", "royalDraw4", []*domain.Card{c(1, 1), c(1, 13), c(1, 12), c(1, 11), c(3, 11)}, []int{0, 1, 2, 3}},
		{"straight flush draw", "jacksorbetter", "straightFlushDraw4", []*domain.Card{c(3, 5), c(3, 6), c(3, 7), c(3, 8), c(2, 8)}, []int{0, 1, 2, 3}},
		{"wild is kept", "deuceswild", "keepFourDeuces", []*domain.Card{c(1, 2), c(2, 2), c(3, 2), c(4, 2), c(3, 9)}, []int{0, 1, 2, 3}},
		{"joker stays", "jokerpoker", "keepHighPair", []*domain.Card{c(0, 1), c(1, 13), c(2, 3), c(3, 7), c(4, 9)}, []int{0, 1}},
		{"royal draw outranks high pair", "jacksorbetter", "royalDraw4", []*domain.Card{c(1, 1), c(1, 13), c(1, 12), c(1, 11), c(3, 11)}, []int{0, 1, 2, 3}},
		{"inside straight draw", "jacksorbetter", "insideStraightDraw", []*domain.Card{c(1, 1), c(2, 13), c(4, 12), c(3, 11), c(1, 9)}, []int{0, 1, 2, 3}},
		{"low pair and flush draw chooses flush", "jacksorbetter", "flushDraw4", []*domain.Card{c(1, 7), c(4, 7), c(1, 3), c(1, 5), c(1, 10)}, []int{0, 2, 3, 4}},
		{"low pair and outside draw keeps pair", "jacksorbetter", "keepLowPair", []*domain.Card{c(1, 7), c(4, 7), c(1, 8), c(2, 9), c(3, 10)}, []int{0, 1}},
		{"high cards are Q K", "jacksorbetter", "highCards", []*domain.Card{c(1, 12), c(3, 13), c(4, 3), c(2, 5), c(1, 7)}, []int{0, 1}},
		{"high cards choose K Q over A", "jacksorbetter", "highCards", []*domain.Card{c(1, 1), c(3, 13), c(4, 12), c(2, 3), c(1, 7)}, []int{1, 2}},
		{"Deuces Wild keeps one 2", "deuceswild", "keepWilds", []*domain.Card{c(1, 2), c(3, 5), c(4, 7), c(2, 9), c(1, 13)}, []int{0}},
		{"Deuces Wild holds high pair in a two pair", "deuceswild", "keepPair", []*domain.Card{c(1, 10), c(3, 10), c(4, 7), c(2, 7), c(1, 3)}, []int{0, 1}},
		{"Deuces Wild two suited tens royal draw", "deuceswild", "royalDraw2", []*domain.Card{c(1, 10), c(1, 11), c(2, 3), c(3, 5), c(4, 7)}, []int{0, 1}},
		{"Deuces Wild ace king not royal draw", "deuceswild", "drawAll", []*domain.Card{c(1, 1), c(1, 13), c(2, 3), c(3, 5), c(4, 7)}, []int{}},
		{"Deuces Wild ace low inside is excluded", "deuceswild", "drawAll", []*domain.Card{c(1, 1), c(2, 3), c(3, 4), c(4, 5), c(1, 9)}, []int{}},
		{"Deuces Wild ace-low inside with no deuces is excluded", "deuceswild", "drawAll", []*domain.Card{c(1, 1), c(3, 3), c(4, 4), c(2, 5), c(3, 9)}, []int{}},
		{"Joker and K form high pair", "jokerpoker", "keepHighPair", []*domain.Card{c(0, 1), c(1, 13), c(2, 3), c(3, 7), c(4, 9)}, []int{0, 1}},
		{"Joker and middle card", "jokerpoker", "keepWildAndMiddleCard", []*domain.Card{c(0, 1), c(1, 7), c(2, 3), c(3, 5), c(4, 11)}, []int{0, 1}},
		{"jokerless J ten royal draw", "jokerpoker", "royalDraw2", []*domain.Card{c(1, 11), c(1, 10), c(2, 3), c(3, 5), c(4, 7)}, []int{0, 1}},
		{"jokerless K J royal draw", "jokerpoker", "royalDraw2", []*domain.Card{c(1, 13), c(1, 11), c(2, 3), c(3, 5), c(4, 7)}, []int{0, 1}},
		{"Joker remains with middle card", "jokerpoker", "keepWildAndMiddleCard", []*domain.Card{c(0, 1), c(1, 7), c(2, 3), c(3, 5), c(4, 11)}, []int{0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.RecommendVideoPokerHold(tt.variant, tt.hand)
			require.Equal(t, tt.key, got.RuleKey)
			for i := range tt.hand {
				want := false
				for _, k := range tt.keep {
					if i == k {
						want = true
					}
				}
				require.Equal(t, want, got.Hold[i])
			}
			for i := 0; i < 100; i++ {
				require.Equal(t, got, domain.RecommendVideoPokerHold(tt.variant, tt.hand))
			}
		})
	}
}

func TestRecommendVideoPokerHoldInvalidAndFallback(t *testing.T) {
	got := domain.RecommendVideoPokerHold("bad", nil)
	require.Equal(t, "drawAll", got.RuleKey)
	h := []*domain.Card{domain.NewCard(1, 1, false), domain.NewCard(1, 13, false), domain.NewCard(1, 12, false), domain.NewCard(1, 11, false), domain.NewCard(3, 11, false)}
	require.Equal(t, domain.RecommendVideoPokerHold("jacksorbetter", h), domain.RecommendVideoPokerHold("bad", h))
}

func TestRecommendedHoldOutsideDrawIsZero(t *testing.T) {
	vp := domain.NewDefaultVideoPoker()
	require.Equal(t, domain.VideoPokerHoldAdvice{}, vp.RecommendedHold())
}
