//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func simonCard(design, value int) *Card { return NewCard(design, value, true) }

func newSsGame() *SimpleSimon {
	g := NewDefaultSimpleSimon()
	g.Reset()
	return g
}

// ssClear empties the board for deterministic setups.
func ssClear(g *SimpleSimon) {
	for i := range g.columns {
		g.columns[i] = nil
	}
	g.completedSuits = 0
}

func TestSimpleSimon_ResetDeals(t *testing.T) {
	g := newSsGame()
	if g.GetPhase() != SimpleSimonPhasePlaying {
		t.Fatalf("phase = %d, want Playing", g.GetPhase())
	}
	total := 0
	cols := g.GetColumns()
	for i, want := range simpleSimonDeal {
		if len(cols[i]) != want {
			t.Errorf("col %d = %d cards, want %d", i, len(cols[i]), want)
		}
		total += len(cols[i])
	}
	if total != 52 {
		t.Errorf("dealt %d, want 52", total)
	}
}

func TestSimpleSimon_MoveSingleAndRun(t *testing.T) {
	g := newSsGame()
	ssClear(g)
	// col0: 9♠; col1: 8♠ -> move 8♠ onto 9♠.
	g.columns[0] = []*Card{simonCard(CardDesignSpade, 9)}
	g.columns[1] = []*Card{simonCard(CardDesignSpade, 8)}
	if err := g.MoveSequence(1, 0, 0); err != nil {
		t.Fatalf("move single: %v", err)
	}
	if len(g.columns[0]) != 2 || len(g.columns[1]) != 0 {
		t.Errorf("unexpected state: %d / %d", len(g.columns[0]), len(g.columns[1]))
	}
	// Move a same-suit run (7♠6♠) from col2 onto the 8♠ now on col0.
	g.columns[2] = []*Card{simonCard(CardDesignSpade, 7), simonCard(CardDesignSpade, 6)}
	if err := g.MoveSequence(2, 0, 0); err != nil {
		t.Fatalf("move run: %v", err)
	}
	if len(g.columns[0]) != 4 {
		t.Errorf("col0 = %d, want 4", len(g.columns[0]))
	}
}

func TestSimpleSimon_IllegalMoves(t *testing.T) {
	g := newSsGame()
	ssClear(g)
	g.columns[0] = []*Card{simonCard(CardDesignSpade, 9)}
	// Non-run group (8♠ 6♠) cannot move together.
	g.columns[1] = []*Card{simonCard(CardDesignSpade, 8), simonCard(CardDesignSpade, 6)}
	if err := g.MoveSequence(1, 0, 0); err == nil {
		t.Error("expected error: not a run")
	}
	// Wrong rank placement (7♠ onto 9♠).
	g.columns[2] = []*Card{simonCard(CardDesignSpade, 7)}
	if err := g.MoveSequence(2, 0, 0); err == nil {
		t.Error("expected error: 7 cannot sit on 9")
	}
	// Same column.
	if err := g.MoveSequence(0, 0, 0); err == nil {
		t.Error("expected error: same column")
	}
	// Out-of-range column.
	if err := g.MoveSequence(0, 0, 99); err == nil {
		t.Error("expected error: bad column")
	}
	// Bad card index.
	if err := g.MoveSequence(0, 9, 1); err == nil {
		t.Error("expected error: bad card index")
	}
}

func TestSimpleSimon_PlaceAnySuitOnHigher(t *testing.T) {
	g := newSsGame()
	ssClear(g)
	// A run only moves same-suit, but placement is any-suit: 8♥ onto 9♠ is legal.
	g.columns[0] = []*Card{simonCard(CardDesignSpade, 9)}
	g.columns[1] = []*Card{simonCard(CardDesignHeart, 8)}
	if err := g.MoveSequence(1, 0, 0); err != nil {
		t.Fatalf("any-suit placement should be legal: %v", err)
	}
}

func TestSimpleSimon_CompleteSuitRemoved(t *testing.T) {
	g := newSsGame()
	ssClear(g)
	// col0 holds K..2 of spades; move the lone A♠ on top to complete K..A.
	col := make([]*Card, 0, 13)
	for v := 13; v >= 2; v-- {
		col = append(col, simonCard(CardDesignSpade, v))
	}
	g.columns[0] = col
	g.columns[1] = []*Card{simonCard(CardDesignSpade, 1)}
	if err := g.MoveSequence(1, 0, 0); err != nil {
		t.Fatalf("move A onto 2: %v", err)
	}
	if g.GetCompletedSuits() != 1 {
		t.Errorf("completed suits = %d, want 1", g.GetCompletedSuits())
	}
	if len(g.columns[0]) != 0 {
		t.Errorf("col0 should be empty after removal, got %d", len(g.columns[0]))
	}
}

func TestSimpleSimon_WinClearsGame(t *testing.T) {
	g := newSsGame()
	ssClear(g)
	g.completedSuits = 3
	// Build the fourth suit (hearts) K..A and complete it.
	col := make([]*Card, 0, 13)
	for v := 13; v >= 2; v-- {
		col = append(col, simonCard(CardDesignHeart, v))
	}
	g.columns[0] = col
	g.columns[1] = []*Card{simonCard(CardDesignHeart, 1)}
	if err := g.MoveSequence(1, 0, 0); err != nil {
		t.Fatalf("final move: %v", err)
	}
	if g.GetPhase() != SimpleSimonPhaseGameClear {
		t.Errorf("phase = %d, want GameClear", g.GetPhase())
	}
}

func TestSimpleSimon_GameOverWhenStuck(t *testing.T) {
	g := newSsGame()
	ssClear(g)
	// All 10 columns filled (no empties) with tops of ranks {A,A,A,A,3,3,3,3,5,5}.
	// No rank is one higher than another present, so no placement is possible
	// and there is no empty column to drop a card onto -> genuinely stuck.
	suits := []int{CardDesignSpade, CardDesignHeart, CardDesignDiamond, CardDesignClover}
	for i, s := range suits {
		g.columns[i] = []*Card{simonCard(s, 1)}
	}
	for i, s := range suits {
		g.columns[4+i] = []*Card{simonCard(s, 3)}
	}
	g.columns[8] = []*Card{simonCard(CardDesignSpade, 5)}
	g.columns[9] = []*Card{simonCard(CardDesignHeart, 5)}
	if g.hasAnyLegalMove() {
		t.Fatal("setup should have no legal move")
	}
	g.checkGameOver()
	if g.GetPhase() != SimpleSimonPhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestSimpleSimon_GiveUp(t *testing.T) {
	g := newSsGame()
	g.GiveUp()
	if g.GetPhase() != SimpleSimonPhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestSimpleSimon_UndoAndHint(t *testing.T) {
	g := newSsGame()
	ssClear(g)
	g.columns[0] = []*Card{simonCard(CardDesignSpade, 9)}
	g.columns[1] = []*Card{simonCard(CardDesignSpade, 8)}
	if g.CanUndo() {
		t.Error("should not be able to undo before a move")
	}
	h := g.GetHint()
	if h == nil {
		t.Fatal("expected a hint")
	}
	// The suggested move must be legal.
	if err := g.MoveSequence(h.FromCol, h.CardIndex, h.ToCol); err != nil {
		t.Fatalf("hinted move should be legal: %v", err)
	}
	_ = g.Undo()
	_ = g.MoveSequence(1, 0, 0)
	if !g.CanUndo() {
		t.Fatal("should be able to undo")
	}
	if err := g.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if len(g.columns[1]) != 1 {
		t.Errorf("undo did not restore col1")
	}
	if err := g.Undo(); err == nil {
		t.Error("expected error: nothing to undo")
	}
	if err := g.UndoN(3); err != nil {
		t.Errorf("UndoN past head: %v", err)
	}
}

func TestSimpleSimon_JSONRoundTrip(t *testing.T) {
	g := newSsGame()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	g2 := NewDefaultSimpleSimon()
	if err := json.Unmarshal(data, g2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g2.GetColumns()[0] == nil || len(g2.GetColumns()[0]) != len(g.GetColumns()[0]) {
		t.Error("column mismatch after round-trip")
	}
	// Reject an out-of-range completed-suit count.
	if err := json.Unmarshal([]byte(`{"cs":9}`), NewDefaultSimpleSimon()); err == nil {
		t.Error("expected error for invalid completed suits")
	}
}

// **まとめて掴めるのは末尾の同スート降順だけ** (#5679)。画面はその境界を出すので、
// 位置を返すこの関数がずれると、掴めない札まで掴めるように見える。
func TestSimpleSimonMovableFrom(t *testing.T) {
	card := func(d, v int) *Card { return NewCard(d, v, false) }

	t.Run("an empty column has no boundary", func(t *testing.T) {
		assert.Equal(t, 0, SimpleSimonMovableFrom(nil))
	})

	t.Run("a single card is the whole run", func(t *testing.T) {
		assert.Equal(t, 0, SimpleSimonMovableFrom([]*Card{card(CardDesignSpade, 5)}))
	})

	// ♠K ♥9 ♠8 ♠7 → 末尾 2 枚だけが同スート降順。
	t.Run("stops where the suit breaks", func(t *testing.T) {
		col := []*Card{
			card(CardDesignSpade, 13),
			card(CardDesignHeart, 9),
			card(CardDesignSpade, 8),
			card(CardDesignSpade, 7),
		}
		assert.Equal(t, 2, SimpleSimonMovableFrom(col))
	})

	// ♠9 ♠8 ♠6 → ランクが飛ぶので末尾 1 枚だけ。
	t.Run("stops where the rank skips", func(t *testing.T) {
		col := []*Card{
			card(CardDesignSpade, 9),
			card(CardDesignSpade, 8),
			card(CardDesignSpade, 6),
		}
		assert.Equal(t, 2, SimpleSimonMovableFrom(col))
	})

	// 列全体が 1 つの並びなら先頭から掴める。
	t.Run("a column that is entirely one run starts at zero", func(t *testing.T) {
		col := []*Card{
			card(CardDesignSpade, 5),
			card(CardDesignSpade, 4),
			card(CardDesignSpade, 3),
		}
		assert.Equal(t, 0, SimpleSimonMovableFrom(col))
	})
}
