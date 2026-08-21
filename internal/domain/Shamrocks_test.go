//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func llCardSH(design, value int) *Card { return NewCard(design, value, true) }

func newLlGameSH() *Shamrocks {
	g := NewDefaultShamrocks()
	g.Reset()
	return g
}

// llClearSH empties the board for deterministic setups.
func llClearSH(g *Shamrocks) {
	g.fans = nil
	for i := range g.foundation {
		g.foundation[i] = nil
	}
}

// Shamrocks diverges from Shamrocks in four ways. Each gets a test and a
// negative control.
// TestShamrocks_MoveSearchMatchesTheMoveRules pins hasAnyLegalMove/GetHint to
// the same rules MoveFanToFan enforces.
//
// Shamrocks's search only asked shamrocksCanStack, which knows nothing
// about Shamrocks' two new constraints. That leaves two defects:
//
//   - an EMPTY fan is never proposed, because fanTop is nil and canStack says
//     no -- yet an empty fan legally takes any card. Since Shamrocks has no
//     redeals, checkGameOver fires straight off hasAnyLegalMove, so a board
//     whose only move is into an empty fan was declared lost.
//   - a FULL fan is proposed, because the search never checks the three-card
//     cap, so the hint can point at a move MoveFanToFan rejects.
func TestShamrocks_MoveSearchMatchesTheMoveRules(t *testing.T) {
	t.Run("an empty fan counts as a legal move", func(t *testing.T) {
		g := newLlGameSH()
		g.Reset()
		llClearSH(g)
		// Two fans, no legal stack between them, but one is empty.
		g.fans = [][]*Card{{llCardSH(CardDesignSpade, 9)}, {}}
		assert.True(t, g.hasAnyLegalMove(), "the 9 can move into the empty fan")
		require.NotNil(t, g.GetHint(), "and the hint should say so")
		g.checkGameOver()
		assert.Equal(t, ShamrocksPhasePlaying, g.GetPhase(), "not a game over")
	})

	t.Run("a full fan is never proposed", func(t *testing.T) {
		g := newLlGameSH()
		g.Reset()
		llClearSH(g)
		// The 6 and the 7 are rank-adjacent, so each would take the other -- but
		// BOTH fans hold three, so neither can receive. Filling only one leaves
		// the reverse move legal, which is what the first draft of this fixture
		// missed.
		g.fans = [][]*Card{
			{llCardSH(CardDesignHeart, 11), llCardSH(CardDesignHeart, 13), llCardSH(CardDesignHeart, 6)},
			{llCardSH(CardDesignSpade, 2), llCardSH(CardDesignSpade, 4), llCardSH(CardDesignSpade, 7)},
		}
		assert.False(t, g.hasAnyLegalMove(), "the only rank-legal target is full")
		assert.Nil(t, g.GetHint(), "so there is nothing to hint")
	})
}

func TestShamrocks_Deal(t *testing.T) {
	g := newLlGameSH()
	g.Reset()

	// 52 = 3*17 + 1, so the last fan holds the single leftover card. The issue
	// claimed "no remainder"; there is one, and it is a fan of its own.
	assert.Equal(t, 18, len(g.fans), "seventeen fans of three plus one of one")
	threes, ones, total := 0, 0, 0
	for _, f := range g.fans {
		switch len(f) {
		case 3:
			threes++
		case 1:
			ones++
		}
		total += len(f)
	}
	assert.Equal(t, 17, threes)
	assert.Equal(t, 1, ones)
	assert.Equal(t, 52, total)
}

func TestShamrocks_StackIgnoresSuitAndGoesBothWays(t *testing.T) {
	// Shamrocks needs the same suit AND descending. Shamrocks ignores suit
	// and accepts either direction.
	assert.True(t, shamrocksCanStack(llCardSH(CardDesignHeart, 6), llCardSH(CardDesignSpade, 7)), "one lower, other suit")
	assert.True(t, shamrocksCanStack(llCardSH(CardDesignHeart, 8), llCardSH(CardDesignSpade, 7)), "one higher, other suit")
	assert.True(t, shamrocksCanStack(llCardSH(CardDesignSpade, 6), llCardSH(CardDesignSpade, 7)), "same suit still fine")

	// Negative controls: adjacency is still required.
	assert.False(t, shamrocksCanStack(llCardSH(CardDesignHeart, 5), llCardSH(CardDesignSpade, 7)), "two lower")
	assert.False(t, shamrocksCanStack(llCardSH(CardDesignHeart, 7), llCardSH(CardDesignSpade, 7)), "same rank")
}

func TestShamrocks_FanCapIsThree(t *testing.T) {
	g := newLlGameSH()
	g.Reset()
	llClearSH(g)
	// A fan already at the cap cannot take a fourth card, even on a legal rank.
	g.fans = [][]*Card{
		{llCardSH(CardDesignHeart, 6)},
		{llCardSH(CardDesignSpade, 3), llCardSH(CardDesignSpade, 4), llCardSH(CardDesignSpade, 7)},
	}
	assert.Error(t, g.MoveFanToFan(0, 1), "fan of three is full")

	// One card shorter and the same move is legal.
	g.fans = [][]*Card{
		{llCardSH(CardDesignHeart, 6)},
		{llCardSH(CardDesignSpade, 4), llCardSH(CardDesignSpade, 7)},
	}
	assert.NoError(t, g.MoveFanToFan(0, 1))
}

func TestShamrocks_EmptyFanTakesAnyCard(t *testing.T) {
	g := newLlGameSH()
	g.Reset()
	llClearSH(g)
	g.fans = [][]*Card{{llCardSH(CardDesignHeart, 6)}, {}}
	assert.NoError(t, g.MoveFanToFan(0, 1), "an empty fan accepts any single card")
	assert.Equal(t, 1, len(g.fans[1]))
}

func TestShamrocks_HasNoRedeals(t *testing.T) {
	g := newLlGameSH()
	g.Reset()
	// La Belle Lucie allows three redeals; Shamrocks has none.
	assert.Equal(t, 0, g.GetRedealsLeft(), "Shamrocks does not redeal")
	assert.Error(t, g.Redeal(), "redeal must be rejected")
}

func TestShamrocks_ResetDeals(t *testing.T) {
	g := newLlGameSH()
	if g.GetPhase() != ShamrocksPhasePlaying {
		t.Fatalf("phase = %d, want Playing", g.GetPhase())
	}
	if g.GetRedealsLeft() != ShamrocksMaxRedeals {
		t.Errorf("redeals = %d, want %d", g.GetRedealsLeft(), ShamrocksMaxRedeals)
	}
	total := 0
	for _, fan := range g.GetFans() {
		total += len(fan)
	}
	if total != 52 {
		t.Errorf("dealt %d cards, want 52", total)
	}
	// 17 fans of 3 + 1 fan of 1 = 18 fans.
	if len(g.GetFans()) != 18 {
		t.Errorf("fans = %d, want 18", len(g.GetFans()))
	}
}

func TestShamrocks_MoveFanToFan(t *testing.T) {
	g := newLlGameSH()
	llClearSH(g)
	g.fans = [][]*Card{{llCardSH(CardDesignSpade, 9)}, {llCardSH(CardDesignSpade, 8)}}
	// 8♠ onto 9♠: same suit, descending -> legal.
	if err := g.MoveFanToFan(1, 0); err != nil {
		t.Fatalf("MoveFanToFan: %v", err)
	}
	if len(g.fans[0]) != 2 || len(g.fans[1]) != 0 {
		t.Errorf("unexpected fan state: %d / %d", len(g.fans[0]), len(g.fans[1]))
	}
	// **Different suit is LEGAL in Shamrocks** -- La Belle Lucie required the
	// same suit, and this assertion is inverted rather than deleted so the flip
	// stays visible.
	g.fans = [][]*Card{{llCardSH(CardDesignHeart, 9)}, {llCardSH(CardDesignSpade, 8)}}
	if err := g.MoveFanToFan(1, 0); err != nil {
		t.Errorf("8S onto 9H should be legal (suit ignored): %v", err)
	}
	// Still illegal: not rank-adjacent.
	g.fans = [][]*Card{{llCardSH(CardDesignHeart, 9)}, {llCardSH(CardDesignSpade, 6)}}
	if err := g.MoveFanToFan(1, 0); err == nil {
		t.Error("expected error: 6 is not adjacent to 9")
	}
	// Onto itself.
	if err := g.MoveFanToFan(0, 0); err == nil {
		t.Error("expected error moving fan onto itself")
	}
	// Empty source.
	g.fans = [][]*Card{{}, {llCardSH(CardDesignSpade, 8)}}
	if err := g.MoveFanToFan(0, 1); err == nil {
		t.Error("expected error for empty source")
	}
}

func TestShamrocks_FoundationAndClear(t *testing.T) {
	g := newLlGameSH()
	llClearSH(g)
	// One fan holds A♦; play it to a foundation.
	g.fans = [][]*Card{{llCardSH(CardDesignDiamond, 1)}}
	if err := g.MoveFanToFoundation(0); err != nil {
		t.Fatalf("MoveFanToFoundation: %v", err)
	}
	if len(g.foundation[0]) != 1 {
		t.Errorf("foundation size = %d, want 1", len(g.foundation[0]))
	}
	// A non-startable card errors.
	g.fans = [][]*Card{{llCardSH(CardDesignClover, 5)}}
	if err := g.MoveFanToFoundation(0); err == nil {
		t.Error("expected error: 5 cannot open a foundation")
	}
}

func TestShamrocks_WinByFillingFoundations(t *testing.T) {
	g := newLlGameSH()
	llClearSH(g)
	// Pre-fill three foundations fully and the fourth up to the King-1.
	g.foundation[0] = fullSuitSH(CardDesignSpade)
	g.foundation[1] = fullSuitSH(CardDesignHeart)
	g.foundation[2] = fullSuitSH(CardDesignClover)
	g.foundation[3] = fullSuitSH(CardDesignDiamond)[:12]  // A..Q
	g.fans = [][]*Card{{llCardSH(CardDesignDiamond, 13)}} // K♦ to finish
	if err := g.MoveFanToFoundation(0); err != nil {
		t.Fatalf("final move: %v", err)
	}
	if g.GetPhase() != ShamrocksPhaseGameClear {
		t.Errorf("phase = %d, want GameClear", g.GetPhase())
	}
}

func fullSuitSH(design int) []*Card {
	s := make([]*Card, 13)
	for v := 1; v <= 13; v++ {
		s[v-1] = llCardSH(design, v)
	}
	return s
}

// Shamrocks's redeal test is gone: Shamrocks has no redeal at all, so
// there is no "spend one, then run out" behaviour to cover. That the feature is
// unavailable is asserted in TestShamrocks_HasNoRedeals; keeping the original
// here would have meant asserting a redeal that can never succeed.

func TestShamrocks_GameOverWhenStuck(t *testing.T) {
	g := newLlGameSH()
	llClearSH(g)
	g.redealsLeft = 0
	// Tops K♠ / 5♥ / 7♦ have no legal interaction and nothing reaches a foundation.
	g.fans = [][]*Card{
		{llCardSH(CardDesignSpade, 13)},
		{llCardSH(CardDesignHeart, 5)},
		{llCardSH(CardDesignClover, 2), llCardSH(CardDesignDiamond, 7)},
	}
	g.checkGameOver()
	if g.GetPhase() != ShamrocksPhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestShamrocks_GiveUp(t *testing.T) {
	g := newLlGameSH()
	g.GiveUp()
	if g.GetPhase() != ShamrocksPhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestShamrocks_UndoRestores(t *testing.T) {
	g := newLlGameSH()
	llClearSH(g)
	g.fans = [][]*Card{{llCardSH(CardDesignDiamond, 1)}}
	if g.CanUndo() {
		t.Error("should not be able to undo before any move")
	}
	_ = g.MoveFanToFoundation(0)
	if !g.CanUndo() {
		t.Fatal("should be able to undo")
	}
	if err := g.Undo(); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if len(g.fans[0]) != 1 || len(g.foundation[0]) != 0 {
		t.Error("undo did not restore the board")
	}
	if err := g.Undo(); err == nil {
		t.Error("expected error: nothing to undo")
	}
	// UndoN past the head is a no-op.
	if err := g.UndoN(5); err != nil {
		t.Errorf("UndoN: %v", err)
	}
}

func TestShamrocks_Hint(t *testing.T) {
	g := newLlGameSH()
	llClearSH(g)
	g.fans = [][]*Card{{llCardSH(CardDesignSpade, 1)}}
	h := g.GetHint()
	if h == nil || !h.ToFoundation {
		t.Fatal("expected a foundation hint for an exposed Ace")
	}
	// Fan-to-fan hint.
	llClearSH(g)
	g.fans = [][]*Card{{llCardSH(CardDesignSpade, 9)}, {llCardSH(CardDesignSpade, 8)}}
	h = g.GetHint()
	if h == nil || h.ToFoundation {
		t.Fatal("expected a fan-to-fan hint")
	}
}

func TestShamrocks_AutoComplete(t *testing.T) {
	g := newLlGameSH()
	llClearSH(g)
	g.foundation[0] = fullSuitSH(CardDesignSpade)[:1] // A♠
	g.fans = [][]*Card{{llCardSH(CardDesignSpade, 2)}, {llCardSH(CardDesignSpade, 3)}}
	if err := g.AutoComplete(); err != nil {
		t.Fatalf("AutoComplete: %v", err)
	}
	if len(g.foundation[0]) != 3 {
		t.Errorf("foundation size = %d, want 3 (A-2-3)", len(g.foundation[0]))
	}
}

func TestShamrocks_JSONRoundTrip(t *testing.T) {
	g := newLlGameSH()
	_ = g.Redeal()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	g2 := NewDefaultShamrocks()
	if err := json.Unmarshal(data, g2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g2.GetRedealsLeft() != g.GetRedealsLeft() {
		t.Errorf("redeals mismatch after round-trip")
	}
	if len(g2.GetFans()) != len(g.GetFans()) {
		t.Errorf("fan count mismatch after round-trip")
	}
	// Reject an out-of-range redeal count.
	if err := json.Unmarshal([]byte(`{"rd":99}`), NewDefaultShamrocks()); err == nil {
		t.Error("expected error for invalid redeal count")
	}
}
