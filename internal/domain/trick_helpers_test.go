//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func tc(playerIdx, design, value int) *TrickCard {
	return &TrickCard{PlayerIdx: playerIdx, Card: NewCard(design, value, false)}
}

func TestResolveTrickWinner_Empty(t *testing.T) {
	if got := ResolveTrickWinner(nil, -1, nil); got != 0 {
		t.Fatalf("empty trick: want 0, got %d", got)
	}
}

func TestResolveTrickWinner_NoTrump(t *testing.T) {
	// Lead suit = Heart. Highest Heart wins; the higher off-suit Diamond loses.
	trick := []*TrickCard{
		tc(0, CardDesignHeart, 9),
		tc(1, CardDesignDiamond, 14),
		tc(2, CardDesignHeart, 11),
		tc(3, CardDesignClover, 13),
	}
	if got := ResolveTrickWinner(trick, -1, nil); got != 2 {
		t.Fatalf("no-trump: want seat 2, got %d", got)
	}
}

func TestResolveTrickWinner_TrumpBeatsNonTrump(t *testing.T) {
	// Lead Heart, trump = Spade. A low Spade beats the high Heart.
	trick := []*TrickCard{
		tc(0, CardDesignHeart, 14),
		tc(1, CardDesignSpade, 2),
		tc(2, CardDesignHeart, 13),
	}
	if got := ResolveTrickWinner(trick, CardDesignSpade, nil); got != 1 {
		t.Fatalf("trump beats non-trump: want seat 1, got %d", got)
	}
}

func TestResolveTrickWinner_HighestTrumpWins(t *testing.T) {
	trick := []*TrickCard{
		tc(0, CardDesignHeart, 14),
		tc(1, CardDesignSpade, 5),
		tc(2, CardDesignSpade, 10),
	}
	if got := ResolveTrickWinner(trick, CardDesignSpade, nil); got != 2 {
		t.Fatalf("highest trump: want seat 2, got %d", got)
	}
}

func TestResolveTrickWinner_TieKeepsEarlierCard(t *testing.T) {
	// Two equal-rank lead-suit cards: the earlier-played seat keeps the trick.
	trick := []*TrickCard{
		tc(0, CardDesignHeart, 10),
		tc(1, CardDesignHeart, 10),
	}
	if got := ResolveTrickWinner(trick, CardDesignSpade, nil); got != 0 {
		t.Fatalf("tie: want earlier seat 0, got %d", got)
	}
}

func TestResolveTrickWinner_CustomRank(t *testing.T) {
	// Inverted rank: lower raw value is stronger. Lead-suit 3 beats lead-suit 14.
	invert := func(c *Card) int { return -c.GetValue() }
	trick := []*TrickCard{
		tc(0, CardDesignClover, 14),
		tc(1, CardDesignClover, 3),
	}
	if got := ResolveTrickWinner(trick, -1, invert); got != 1 {
		t.Fatalf("custom rank: want seat 1, got %d", got)
	}
}

// TestTrickCard_WireShape pins the serialized field names.
//
// This one type now carries the trick in 51 games' KV snapshots (issue #4363),
// each of which previously declared its own struct with these exact tags. A
// change to either tag would therefore break restore for all 51 at once, and
// silently: encoding/json simply leaves an unmatched field at its zero value,
// so an in-progress trick would come back empty rather than erroring.
func TestTrickCard_WireShape(t *testing.T) {
	data, err := json.Marshal(&TrickCard{PlayerIdx: 3, Card: NewCard(2, 11, false)})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for _, key := range []string{"pi", "c"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("TrickCard must serialize %q; got %s", key, data)
		}
	}
	if len(raw) != 2 {
		t.Errorf("TrickCard must serialize exactly pi + c; got %s", data)
	}

	// And it must round-trip, since restore is the direction that matters.
	var back TrickCard
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("round-trip Unmarshal: %v", err)
	}
	if back.PlayerIdx != 3 || back.Card == nil || back.Card.GetValue() != 11 {
		t.Errorf("round-trip lost data: %+v", back)
	}
}
