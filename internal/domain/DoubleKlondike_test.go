//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func dblkCard(design, value int) *Card { return NewCard(design, value, true) }

func newDkGame() *DoubleKlondike {
	g := NewDefaultDoubleKlondike()
	g.Reset()
	return g
}

// dkClear empties the board for deterministic setups.
func dkClear(g *DoubleKlondike) {
	for i := range g.tableau {
		g.tableau[i] = nil
	}
	for i := range g.foundation {
		g.foundation[i] = nil
	}
	g.stock = nil
	g.waste = nil
}

func TestDoubleKlondike_ResetDeals(t *testing.T) {
	g := newDkGame()
	if g.GetPhase() != DoubleKlondikePhasePlaying {
		t.Fatalf("phase = %d, want Playing", g.GetPhase())
	}
	total := 0
	cols := g.GetTableau()
	for i := 0; i < DoubleKlondikeTableauCnt; i++ {
		if len(cols[i]) != i+1 {
			t.Errorf("col %d = %d cards, want %d", i, len(cols[i]), i+1)
		}
		// Only the top card is face up.
		for j, tc := range cols[i] {
			if tc.FaceUp != (j == len(cols[i])-1) {
				t.Errorf("col %d card %d FaceUp=%v unexpected", i, j, tc.FaceUp)
			}
		}
		total += len(cols[i])
	}
	if total != 45 {
		t.Errorf("tableau total = %d, want 45", total)
	}
	if g.GetStockCount() != 104-45 {
		t.Errorf("stock = %d, want %d", g.GetStockCount(), 104-45)
	}
}

func TestDoubleKlondike_Draw(t *testing.T) {
	g := newDkGame()
	before := g.GetStockCount()
	if err := g.Draw(); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if len(g.GetWaste()) != DoubleKlondikeDrawCount {
		t.Errorf("waste = %d, want %d", len(g.GetWaste()), DoubleKlondikeDrawCount)
	}
	if g.GetStockCount() != before-DoubleKlondikeDrawCount {
		t.Errorf("stock = %d, want %d", g.GetStockCount(), before-DoubleKlondikeDrawCount)
	}
	// Exhaust stock then recycle.
	for g.GetStockCount() > 0 {
		_ = g.Draw()
	}
	wasteBefore := len(g.GetWaste())
	if wasteBefore == 0 {
		t.Skip("waste empty (unexpected)")
	}
	if err := g.Draw(); err != nil {
		t.Fatalf("recycle: %v", err)
	}
	if g.GetStockCount() != wasteBefore {
		t.Errorf("after recycle stock = %d, want %d", g.GetStockCount(), wasteBefore)
	}
}

func TestDoubleKlondike_TableauRules(t *testing.T) {
	g := newDkGame()
	dkClear(g)
	// Empty column only accepts a King.
	g.tableau[0] = nil
	if !g.canPlaceOnTableau(dblkCard(CardDesignSpade, 13), 0) {
		t.Error("empty column should accept a King")
	}
	if g.canPlaceOnTableau(dblkCard(CardDesignSpade, 12), 0) {
		t.Error("empty column should reject a non-King")
	}
	// Alternating colour descending.
	g.tableau[1] = []*DoubleKlondikeTableauCard{{Card: dblkCard(CardDesignSpade, 8), FaceUp: true}}
	if !g.canPlaceOnTableau(dblkCard(CardDesignHeart, 7), 1) {
		t.Error("red 7 on black 8 should be legal")
	}
	if g.canPlaceOnTableau(dblkCard(CardDesignClover, 7), 1) {
		t.Error("black 7 on black 8 should be illegal")
	}
}

func TestDoubleKlondike_FoundationAcrossTwoPiles(t *testing.T) {
	g := newDkGame()
	dkClear(g)
	// Two separate spade Aces should each open a different foundation.
	g.waste = []*Card{dblkCard(CardDesignSpade, 1)}
	if err := g.MoveWasteToFoundation(); err != nil {
		t.Fatalf("first ace: %v", err)
	}
	g.waste = []*Card{dblkCard(CardDesignSpade, 1)}
	if err := g.MoveWasteToFoundation(); err != nil {
		t.Fatalf("second ace (other pile): %v", err)
	}
	used := 0
	for _, f := range g.GetFoundation() {
		if len(f) > 0 {
			used++
		}
	}
	if used != 2 {
		t.Errorf("two spade aces should occupy 2 foundations, got %d", used)
	}
}

func TestDoubleKlondike_TableauToFoundationAndFlip(t *testing.T) {
	g := newDkGame()
	dkClear(g)
	// col0: [face-down 5♣, face-up A♦]; moving the Ace exposes (flips) the 5♣.
	g.tableau[0] = []*DoubleKlondikeTableauCard{
		{Card: dblkCard(CardDesignClover, 5), FaceUp: false},
		{Card: dblkCard(CardDesignDiamond, 1), FaceUp: true},
	}
	if err := g.MoveTableauToFoundation(0); err != nil {
		t.Fatalf("move A to foundation: %v", err)
	}
	if len(g.tableau[0]) != 1 || !g.tableau[0][0].FaceUp {
		t.Error("the exposed card should be auto-flipped face up")
	}
}

func TestDoubleKlondike_TableauToTableauRun(t *testing.T) {
	g := newDkGame()
	dkClear(g)
	g.tableau[0] = []*DoubleKlondikeTableauCard{{Card: dblkCard(CardDesignSpade, 9), FaceUp: true}}
	g.tableau[1] = []*DoubleKlondikeTableauCard{
		{Card: dblkCard(CardDesignHeart, 8), FaceUp: true},
		{Card: dblkCard(CardDesignClover, 7), FaceUp: true},
	}
	// Move the 8♥7♣ run onto the 9♠.
	if err := g.MoveTableauToTableau(1, 0, 0); err != nil {
		t.Fatalf("move run: %v", err)
	}
	if len(g.tableau[0]) != 3 || len(g.tableau[1]) != 0 {
		t.Errorf("unexpected state: %d / %d", len(g.tableau[0]), len(g.tableau[1]))
	}
	// Face-down card cannot move.
	g.tableau[2] = []*DoubleKlondikeTableauCard{{Card: dblkCard(CardDesignSpade, 6), FaceUp: false}}
	if err := g.MoveTableauToTableau(2, 0, 0); err == nil {
		t.Error("expected error: face-down card")
	}
}

func TestDoubleKlondike_WinClearsGame(t *testing.T) {
	g := newDkGame()
	dkClear(g)
	// Fill 7 foundations completely and the 8th up to the Queen.
	for i := 0; i < 7; i++ {
		g.foundation[i] = dkFullSuit(CardDesignSpade)
	}
	g.foundation[7] = dkFullSuit(CardDesignHeart)[:12]
	g.waste = []*Card{dblkCard(CardDesignHeart, 13)}
	if err := g.MoveWasteToFoundation(); err != nil {
		t.Fatalf("final move: %v", err)
	}
	if g.GetPhase() != DoubleKlondikePhaseGameClear {
		t.Errorf("phase = %d, want GameClear", g.GetPhase())
	}
}

func dkFullSuit(design int) []*Card {
	s := make([]*Card, 13)
	for v := 1; v <= 13; v++ {
		s[v-1] = dblkCard(design, v)
	}
	return s
}

func TestDoubleKlondike_HintUndoGiveUpStalemate(t *testing.T) {
	g := newDkGame()
	dkClear(g)
	g.waste = []*Card{dblkCard(CardDesignSpade, 1)}
	h := g.GetHint()
	if h == nil || h.ToZone != "foundation" {
		t.Fatalf("expected a foundation hint for an exposed Ace, got %+v", h)
	}
	if g.CanUndo() {
		t.Error("should not be able to undo before a move")
	}
	_ = g.MoveWasteToFoundation()
	if !g.CanUndo() {
		t.Fatal("should be able to undo")
	}
	if err := g.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if len(g.waste) != 1 {
		t.Error("undo did not restore the waste")
	}
	if err := g.UndoN(5); err != nil {
		t.Errorf("UndoN past head: %v", err)
	}
	// Stalemate: no stock/waste and a single non-King card with no moves.
	dkClear(g)
	g.tableau[0] = []*DoubleKlondikeTableauCard{{Card: dblkCard(CardDesignHeart, 5), FaceUp: true}}
	if !g.IsStalemate() {
		t.Error("expected stalemate")
	}
	g.GiveUp()
	if g.GetPhase() != DoubleKlondikePhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestDoubleKlondike_AutoComplete(t *testing.T) {
	g := newDkGame()
	dkClear(g)
	g.foundation[0] = dkFullSuit(CardDesignSpade)[:1] // A♠
	g.tableau[0] = []*DoubleKlondikeTableauCard{{Card: dblkCard(CardDesignSpade, 2), FaceUp: true}}
	g.waste = []*Card{dblkCard(CardDesignSpade, 3)}
	if err := g.AutoComplete(); err != nil {
		t.Fatalf("autocomplete: %v", err)
	}
	if len(g.foundation[0]) != 3 {
		t.Errorf("foundation = %d, want 3 (A-2-3)", len(g.foundation[0]))
	}
}

func TestDoubleKlondike_JSONRoundTrip(t *testing.T) {
	g := newDkGame()
	_ = g.Draw()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	g2 := NewDefaultDoubleKlondike()
	if err := json.Unmarshal(data, g2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g2.GetStockCount() != g.GetStockCount() || len(g2.GetWaste()) != len(g.GetWaste()) {
		t.Error("stock/waste mismatch after round-trip")
	}
	if err := json.Unmarshal([]byte(`{"ph":9}`), NewDefaultDoubleKlondike()); err == nil {
		t.Error("expected error for invalid phase")
	}
}
