//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func llCard(design, value int) *Card { return NewCard(design, value, true) }

func newLlGame() *LaBelleLucie {
	g := NewDefaultLaBelleLucie()
	g.Reset()
	return g
}

// llClear empties the board for deterministic setups.
func llClear(g *LaBelleLucie) {
	g.fans = nil
	for i := range g.foundation {
		g.foundation[i] = nil
	}
}

func TestLaBelleLucie_ResetDeals(t *testing.T) {
	g := newLlGame()
	if g.GetPhase() != LaBelleLuciePhasePlaying {
		t.Fatalf("phase = %d, want Playing", g.GetPhase())
	}
	if g.GetRedealsLeft() != LaBelleLucieMaxRedeals {
		t.Errorf("redeals = %d, want %d", g.GetRedealsLeft(), LaBelleLucieMaxRedeals)
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

func TestLaBelleLucie_MoveFanToFan(t *testing.T) {
	g := newLlGame()
	llClear(g)
	g.fans = [][]*Card{{llCard(CardDesignSpade, 9)}, {llCard(CardDesignSpade, 8)}}
	// 8♠ onto 9♠: same suit, descending -> legal.
	if err := g.MoveFanToFan(1, 0); err != nil {
		t.Fatalf("MoveFanToFan: %v", err)
	}
	if len(g.fans[0]) != 2 || len(g.fans[1]) != 0 {
		t.Errorf("unexpected fan state: %d / %d", len(g.fans[0]), len(g.fans[1]))
	}
	// Illegal: different suit.
	g.fans = [][]*Card{{llCard(CardDesignHeart, 9)}, {llCard(CardDesignSpade, 8)}}
	if err := g.MoveFanToFan(1, 0); err == nil {
		t.Error("expected error for different-suit stack")
	}
	// Onto itself.
	if err := g.MoveFanToFan(0, 0); err == nil {
		t.Error("expected error moving fan onto itself")
	}
	// Empty source.
	g.fans = [][]*Card{{}, {llCard(CardDesignSpade, 8)}}
	if err := g.MoveFanToFan(0, 1); err == nil {
		t.Error("expected error for empty source")
	}
}

func TestLaBelleLucie_FoundationAndClear(t *testing.T) {
	g := newLlGame()
	llClear(g)
	// One fan holds A♦; play it to a foundation.
	g.fans = [][]*Card{{llCard(CardDesignDiamond, 1)}}
	if err := g.MoveFanToFoundation(0); err != nil {
		t.Fatalf("MoveFanToFoundation: %v", err)
	}
	if len(g.foundation[0]) != 1 {
		t.Errorf("foundation size = %d, want 1", len(g.foundation[0]))
	}
	// A non-startable card errors.
	g.fans = [][]*Card{{llCard(CardDesignClover, 5)}}
	if err := g.MoveFanToFoundation(0); err == nil {
		t.Error("expected error: 5 cannot open a foundation")
	}
}

func TestLaBelleLucie_WinByFillingFoundations(t *testing.T) {
	g := newLlGame()
	llClear(g)
	// Pre-fill three foundations fully and the fourth up to the King-1.
	g.foundation[0] = fullSuit(CardDesignSpade)
	g.foundation[1] = fullSuit(CardDesignHeart)
	g.foundation[2] = fullSuit(CardDesignClover)
	g.foundation[3] = fullSuit(CardDesignDiamond)[:12]  // A..Q
	g.fans = [][]*Card{{llCard(CardDesignDiamond, 13)}} // K♦ to finish
	if err := g.MoveFanToFoundation(0); err != nil {
		t.Fatalf("final move: %v", err)
	}
	if g.GetPhase() != LaBelleLuciePhaseGameClear {
		t.Errorf("phase = %d, want GameClear", g.GetPhase())
	}
}

func fullSuit(design int) []*Card {
	s := make([]*Card, 13)
	for v := 1; v <= 13; v++ {
		s[v-1] = llCard(design, v)
	}
	return s
}

func TestLaBelleLucie_Redeal(t *testing.T) {
	g := newLlGame()
	before := g.GetRedealsLeft()
	if err := g.Redeal(); err != nil {
		t.Fatalf("Redeal: %v", err)
	}
	if g.GetRedealsLeft() != before-1 {
		t.Errorf("redeals = %d, want %d", g.GetRedealsLeft(), before-1)
	}
	// Exhaust the remaining redeals, then the next one errors.
	for g.GetRedealsLeft() > 0 {
		_ = g.Redeal()
	}
	if err := g.Redeal(); err == nil {
		t.Error("expected error: no redeals left")
	}
}

func TestLaBelleLucie_GameOverWhenStuck(t *testing.T) {
	g := newLlGame()
	llClear(g)
	g.redealsLeft = 0
	// Tops K♠ / 5♥ / 7♦ have no legal interaction and nothing reaches a foundation.
	g.fans = [][]*Card{
		{llCard(CardDesignSpade, 13)},
		{llCard(CardDesignHeart, 5)},
		{llCard(CardDesignClover, 2), llCard(CardDesignDiamond, 7)},
	}
	g.checkGameOver()
	if g.GetPhase() != LaBelleLuciePhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestLaBelleLucie_GiveUp(t *testing.T) {
	g := newLlGame()
	g.GiveUp()
	if g.GetPhase() != LaBelleLuciePhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestLaBelleLucie_UndoRestores(t *testing.T) {
	g := newLlGame()
	llClear(g)
	g.fans = [][]*Card{{llCard(CardDesignDiamond, 1)}}
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

func TestLaBelleLucie_Hint(t *testing.T) {
	g := newLlGame()
	llClear(g)
	g.fans = [][]*Card{{llCard(CardDesignSpade, 1)}}
	h := g.GetHint()
	if h == nil || !h.ToFoundation {
		t.Fatal("expected a foundation hint for an exposed Ace")
	}
	// Fan-to-fan hint.
	llClear(g)
	g.fans = [][]*Card{{llCard(CardDesignSpade, 9)}, {llCard(CardDesignSpade, 8)}}
	h = g.GetHint()
	if h == nil || h.ToFoundation {
		t.Fatal("expected a fan-to-fan hint")
	}
}

func TestLaBelleLucie_AutoComplete(t *testing.T) {
	g := newLlGame()
	llClear(g)
	g.foundation[0] = fullSuit(CardDesignSpade)[:1] // A♠
	g.fans = [][]*Card{{llCard(CardDesignSpade, 2)}, {llCard(CardDesignSpade, 3)}}
	if err := g.AutoComplete(); err != nil {
		t.Fatalf("AutoComplete: %v", err)
	}
	if len(g.foundation[0]) != 3 {
		t.Errorf("foundation size = %d, want 3 (A-2-3)", len(g.foundation[0]))
	}
}

func TestLaBelleLucie_JSONRoundTrip(t *testing.T) {
	g := newLlGame()
	_ = g.Redeal()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	g2 := NewDefaultLaBelleLucie()
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
	if err := json.Unmarshal([]byte(`{"rd":99}`), NewDefaultLaBelleLucie()); err == nil {
		t.Error("expected error for invalid redeal count")
	}
}
