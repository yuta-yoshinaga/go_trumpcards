//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func bhCard(design, value int) *Card { return NewCard(design, value, true) }

func newBhGame() *BlackHole {
	g := NewDefaultBlackHole()
	g.Reset()
	return g
}

// bhClear empties the board for deterministic setups.
func bhClear(g *BlackHole) {
	g.fans = nil
	g.blackHole = nil
}

func TestBlackHole_ResetDeals(t *testing.T) {
	g := newBhGame()
	if g.GetPhase() != BlackHolePhasePlaying {
		t.Fatalf("phase = %d, want Playing", g.GetPhase())
	}
	// The black hole starts with the Ace of Spades.
	bh := g.GetBlackHole()
	if len(bh) != 1 || bh[0].GetDesign() != CardDesignSpade || bh[0].GetValue() != 1 {
		t.Fatalf("black hole should start with ♠A, got %+v", bh)
	}
	fans := g.GetFans()
	total := 0
	for _, f := range fans {
		if len(f) > BlackHoleFanSize {
			t.Errorf("fan has %d cards, want <= %d", len(f), BlackHoleFanSize)
		}
		total += len(f)
	}
	if total != 51 {
		t.Errorf("fans hold %d cards, want 51", total)
	}
	if len(fans) != BlackHoleFanCnt {
		t.Errorf("fan count = %d, want %d", len(fans), BlackHoleFanCnt)
	}
}

func TestBlackHole_AdjacencyNoWrap(t *testing.T) {
	if !blackHoleAdjacent(bhCard(CardDesignHeart, 5), bhCard(CardDesignClover, 4)) {
		t.Error("5 and 4 should be adjacent")
	}
	if !blackHoleAdjacent(bhCard(CardDesignHeart, 5), bhCard(CardDesignClover, 6)) {
		t.Error("5 and 6 should be adjacent")
	}
	if blackHoleAdjacent(bhCard(CardDesignHeart, 5), bhCard(CardDesignClover, 7)) {
		t.Error("5 and 7 should not be adjacent")
	}
	// No K-A wrap.
	if blackHoleAdjacent(bhCard(CardDesignHeart, 13), bhCard(CardDesignClover, 1)) {
		t.Error("K and A must NOT wrap around")
	}
}

func TestBlackHole_PlayAdjacentCard(t *testing.T) {
	g := newBhGame()
	bhClear(g)
	g.blackHole = []*Card{bhCard(CardDesignSpade, 5)}
	g.fans = [][]*Card{
		{bhCard(CardDesignHeart, 9), bhCard(CardDesignClover, 6)}, // top = 6 (playable)
		{bhCard(CardDesignDiamond, 10)},                           // top = 10 (not playable)
	}
	if err := g.MoveFanToBlackHole(0); err != nil {
		t.Fatalf("play adjacent card: %v", err)
	}
	if g.blackHoleTop().GetValue() != 6 {
		t.Errorf("black hole top = %d, want 6", g.blackHoleTop().GetValue())
	}
	if len(g.fans[0]) != 1 {
		t.Errorf("fan 0 should have 1 card left, got %d", len(g.fans[0]))
	}
	// Non-adjacent move is rejected.
	if err := g.MoveFanToBlackHole(1); err == nil {
		t.Error("expected error: 10 is not adjacent to 6")
	}
}

func TestBlackHole_WinClearsGame(t *testing.T) {
	g := newBhGame()
	bhClear(g)
	// Pre-fill the black hole with 51 cards; one playable card remains in a fan.
	g.blackHole = make([]*Card, 51)
	for i := range g.blackHole {
		g.blackHole[i] = bhCard(CardDesignSpade, 1)
	}
	g.blackHole[50] = bhCard(CardDesignHeart, 7)
	g.fans = [][]*Card{{bhCard(CardDesignClover, 8)}}
	if err := g.MoveFanToBlackHole(0); err != nil {
		t.Fatalf("final move: %v", err)
	}
	if g.GetPhase() != BlackHolePhaseGameClear {
		t.Errorf("phase = %d, want GameClear", g.GetPhase())
	}
}

func TestBlackHole_StalemateEndsGame(t *testing.T) {
	g := newBhGame()
	bhClear(g)
	g.blackHole = []*Card{bhCard(CardDesignSpade, 5)}
	// Only non-adjacent cards remain -> stalemate after a move triggers game over.
	g.fans = [][]*Card{{bhCard(CardDesignHeart, 6)}, {bhCard(CardDesignClover, 10)}}
	if err := g.MoveFanToBlackHole(0); err != nil {
		t.Fatalf("move: %v", err)
	}
	// Now top is 6, remaining fan top is 10 (not adjacent) -> game over.
	if g.GetPhase() != BlackHolePhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestBlackHole_HintUndoGiveUp(t *testing.T) {
	g := newBhGame()
	bhClear(g)
	g.blackHole = []*Card{bhCard(CardDesignSpade, 5)}
	g.fans = [][]*Card{{bhCard(CardDesignClover, 10)}, {bhCard(CardDesignHeart, 4)}}
	h := g.GetHint()
	if h == nil || h.Fan != 1 {
		t.Fatalf("expected hint for fan 1, got %+v", h)
	}
	if g.CanUndo() {
		t.Error("should not be able to undo before a move")
	}
	_ = g.MoveFanToBlackHole(1)
	if !g.CanUndo() {
		t.Fatal("should be able to undo after a move")
	}
	if err := g.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if g.blackHoleTop().GetValue() != 5 || len(g.fans[1]) != 1 {
		t.Error("undo did not restore the pre-move state")
	}
	if err := g.UndoN(5); err != nil {
		t.Errorf("UndoN past head: %v", err)
	}
	g.GiveUp()
	if g.GetPhase() != BlackHolePhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestBlackHole_IsStalemate(t *testing.T) {
	g := newBhGame()
	bhClear(g)
	g.blackHole = []*Card{bhCard(CardDesignSpade, 5)}
	g.fans = [][]*Card{{bhCard(CardDesignClover, 10)}}
	if !g.IsStalemate() {
		t.Error("expected stalemate (no adjacent cards)")
	}
}

func TestBlackHole_JSONRoundTrip(t *testing.T) {
	g := newBhGame()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	g2 := NewDefaultBlackHole()
	if err := json.Unmarshal(data, g2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(g2.GetBlackHole()) != len(g.GetBlackHole()) || len(g2.GetFans()) != len(g.GetFans()) {
		t.Error("state mismatch after round-trip")
	}
	if err := json.Unmarshal([]byte(`{"ph":9}`), NewDefaultBlackHole()); err == nil {
		t.Error("expected error for invalid phase")
	}
}

// **±1 のクランプと合法手数 (#4818)。**Web の acceptableRanks / legalFans と
// 同じ値になること。
func TestBlackHole_AcceptableRanksAndPlayableFans(t *testing.T) {
	g := NewDefaultBlackHole()
	if got := g.AcceptableRanks(); got != nil {
		t.Errorf("穴が空なら候補なし: got %v", got)
	}

	restore := func(js string) {
		t.Helper()
		if err := json.Unmarshal([]byte(js), g); err != nil {
			t.Fatalf("restore: %v", err)
		}
	}
	eq := func(want, got []int, msg string) {
		t.Helper()
		if len(want) != len(got) {
			t.Fatalf("%s: want %v got %v", msg, want, got)
		}
		for i := range want {
			if want[i] != got[i] {
				t.Fatalf("%s: want %v got %v", msg, want, got)
			}
		}
	}

	restore(`{"ph":0,"bh":[{"d":1,"v":7}],"fn":[[{"d":3,"v":8}],[{"d":2,"v":2}],[{"d":4,"v":6}]]}`)
	eq([]int{6, 8}, g.AcceptableRanks(), "±1")
	eq([]int{0, 2}, g.PlayableFans(), "♥8 と ♦6 が積める")

	// A は下側が無い、K は上側が無い (ラップしない)。
	restore(`{"ph":0,"bh":[{"d":1,"v":1}],"fn":[]}`)
	eq([]int{2}, g.AcceptableRanks(), "A の隣は 2 だけ")
	restore(`{"ph":0,"bh":[{"d":1,"v":13}],"fn":[]}`)
	eq([]int{12}, g.AcceptableRanks(), "K の隣は Q だけ")

	// 積める扇が無い。
	restore(`{"ph":0,"bh":[{"d":1,"v":7}],"fn":[[{"d":2,"v":2}]]}`)
	if got := g.PlayableFans(); got != nil {
		t.Errorf("積める扇は無いはず: got %v", got)
	}
}
