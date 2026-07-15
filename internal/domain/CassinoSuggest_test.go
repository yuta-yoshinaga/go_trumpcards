package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func cassinoCard(v int) *Card { return NewCard(CardDesignSpade, v, false) }

func TestSuggestCassinoMove(t *testing.T) {
	t.Run("nil for an empty hand", func(t *testing.T) {
		assert.Nil(t, SuggestCassinoMove(nil, []*Card{cassinoCard(3)}, nil))
	})

	t.Run("takes a loose-card sum capture", func(t *testing.T) {
		hand := []*Card{cassinoCard(5)}
		table := []*Card{cassinoCard(2), cassinoCard(3), cassinoCard(9)}
		h := SuggestCassinoMove(hand, table, nil)
		assert.Equal(t, CassinoHintTake, h.Action)
		assert.Equal(t, 0, h.HandIdx)
		assert.ElementsMatch(t, []int{0, 1}, h.TableIdxs) // 2 + 3 = 5
	})

	t.Run("prefers the capture that takes the most cards", func(t *testing.T) {
		// 6 can take {6} (1 card) or {2,4} (2 cards) — prefer the latter.
		hand := []*Card{cassinoCard(6)}
		table := []*Card{cassinoCard(6), cassinoCard(2), cassinoCard(4)}
		h := SuggestCassinoMove(hand, table, nil)
		assert.Equal(t, CassinoHintTake, h.Action)
		assert.Len(t, h.TableIdxs, 2)
	})

	t.Run("takes matching face cards by rank", func(t *testing.T) {
		hand := []*Card{NewCard(CardDesignHeart, 13, false)} // King
		table := []*Card{NewCard(CardDesignSpade, 13, false), cassinoCard(4)}
		h := SuggestCassinoMove(hand, table, nil)
		assert.Equal(t, CassinoHintTake, h.Action)
		assert.Equal(t, []int{0}, h.TableIdxs)
	})

	t.Run("takes a build whose declared value matches the hand card", func(t *testing.T) {
		hand := []*Card{cassinoCard(8)}
		builds := []*CassinoBuild{NewCassinoBuild(1, 8, []*Card{cassinoCard(5), cassinoCard(3)})}
		h := SuggestCassinoMove(hand, nil, builds)
		assert.Equal(t, CassinoHintTake, h.Action)
		assert.Equal(t, []int{0}, h.BuildIdxs)
	})

	t.Run("builds when a combined value is held in hand", func(t *testing.T) {
		// Play 3 onto a table 5 to declare an 8, holding the other 8.
		hand := []*Card{cassinoCard(3), cassinoCard(8)}
		table := []*Card{cassinoCard(5)}
		h := SuggestCassinoMove(hand, table, nil)
		assert.Equal(t, CassinoHintBuild, h.Action)
		assert.Equal(t, 0, h.HandIdx)
		assert.Equal(t, 8, h.Value)
	})

	t.Run("trails the lowest card when nothing can be captured or built", func(t *testing.T) {
		hand := []*Card{cassinoCard(9), cassinoCard(4)}
		table := []*Card{cassinoCard(7)}
		h := SuggestCassinoMove(hand, table, nil)
		assert.Equal(t, CassinoHintTrail, h.Action)
		assert.Equal(t, 1, h.HandIdx) // the 4 is lower than the 9
	})
}
