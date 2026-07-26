//go:build test

package domain

import (
	"reflect"
	"testing"
)

// handStub is a minimal cardIndexer for exercising the tactical helpers.
type handStub struct{ cards []*Card }

func (h handStub) GetCard(i int) *Card { return h.cards[i] }

func newHandStub(cards ...*Card) handStub { return handStub{cards: cards} }

func TestFilterByDesign(t *testing.T) {
	h := newHandStub(
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignSpade, 12, false),
	)
	got := filterByDesign(h, []int{0, 1, 2}, CardDesignSpade)
	if !reflect.DeepEqual(got, []int{0, 2}) {
		t.Fatalf("filterByDesign: want [0 2], got %v", got)
	}
}

func TestFilterAboveBelow(t *testing.T) {
	h := newHandStub(
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignSpade, 11, false),
	)
	if got := filterAbove(h, []int{0, 1, 2}, 8, nil); !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("filterAbove(>8): want [2], got %v", got)
	}
	if got := filterBelow(h, []int{0, 1, 2}, 8, nil); !reflect.DeepEqual(got, []int{0}) {
		t.Fatalf("filterBelow(<8): want [0], got %v", got)
	}
}

func TestPickLowestHighest(t *testing.T) {
	h := newHandStub(
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignHeart, 14, false),
	)
	if got := pickLowest(h, []int{0, 1, 2}, nil); got != 1 {
		t.Fatalf("pickLowest: want 1, got %d", got)
	}
	if got := pickHighest(h, []int{0, 1, 2}, nil); got != 2 {
		t.Fatalf("pickHighest: want 2, got %d", got)
	}
	if got := pickLowest(h, nil, nil); got != -1 {
		t.Fatalf("pickLowest(empty): want -1, got %d", got)
	}
	if got := pickHighest(h, nil, nil); got != -1 {
		t.Fatalf("pickHighest(empty): want -1, got %d", got)
	}
}

func TestPickTiesKeepEarlier(t *testing.T) {
	h := newHandStub(
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignHeart, 7, false),
	)
	if got := pickLowest(h, []int{0, 1}, nil); got != 0 {
		t.Fatalf("pickLowest tie: want earlier 0, got %d", got)
	}
	if got := pickHighest(h, []int{0, 1}, nil); got != 0 {
		t.Fatalf("pickHighest tie: want earlier 0, got %d", got)
	}
}

func TestHelpersCustomRank(t *testing.T) {
	// Inverted rank: raw 3 outranks raw 14.
	invert := func(c *Card) int { return -c.GetValue() }
	h := newHandStub(
		NewCard(CardDesignClover, 14, false),
		NewCard(CardDesignClover, 3, false),
	)
	if got := pickLowest(h, []int{0, 1}, invert); got != 0 {
		t.Fatalf("pickLowest(invert): want 0 (raw 14 = lowest inverted rank), got %d", got)
	}
	if got := pickHighest(h, []int{0, 1}, invert); got != 1 {
		t.Fatalf("pickHighest(invert): want 1 (raw 3 = highest inverted rank), got %d", got)
	}
}
