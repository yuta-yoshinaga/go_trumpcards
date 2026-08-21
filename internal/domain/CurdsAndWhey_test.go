//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func simonCardCW(design, value int) *Card { return NewCard(design, value, true) }

func newSsGameCW() *CurdsAndWhey {
	g := NewDefaultCurdsAndWhey()
	g.Reset()
	return g
}

// ssClearCW empties the board for deterministic setups.
func ssClearCW(g *CurdsAndWhey) {
	for i := range g.columns {
		g.columns[i] = nil
	}
	g.completedSuits = 0
}

// Curds and Whey deals 13 columns of 4 -- Simple Simon, which this was cloned
// from, deals 10 columns of 8/8/8/7/6/5/4/3/2/1.
func TestCurdsAndWhey_Deal(t *testing.T) {
	g := newSsGameCW()
	g.Reset()

	assert.Equal(t, 13, CurdsAndWheyColCnt, "thirteen columns")
	total := 0
	for i := 0; i < CurdsAndWheyColCnt; i++ {
		assert.Equal(t, 4, len(g.columns[i]), "column %d holds 4 cards", i)
		total += len(g.columns[i])
	}
	assert.Equal(t, 52, total, "13 x 4 = the whole deck, evenly")
}

// The tableau accepts the SAME SUIT one rank higher, or the SAME RANK. Simple
// Simon accepts any suit one rank higher, so both halves are a divergence and
// each gets a negative control.
func TestCurdsAndWhey_CanPlace(t *testing.T) {
	g := newSsGameCW()
	g.Reset()
	for i := 0; i < CurdsAndWheyColCnt; i++ {
		g.columns[i] = nil
	}
	g.columns[0] = []*Card{simonCardCW(CardDesignSpade, 7)}

	assert.True(t, g.canPlace(simonCardCW(CardDesignSpade, 6), 0), "same suit, one lower")
	assert.True(t, g.canPlace(simonCardCW(CardDesignHeart, 7), 0), "same rank")

	// Simple Simon would accept this; Curds and Whey must not.
	assert.False(t, g.canPlace(simonCardCW(CardDesignHeart, 6), 0), "one lower but wrong suit")
	assert.False(t, g.canPlace(simonCardCW(CardDesignSpade, 5), 0), "same suit but two lower")
	assert.False(t, g.canPlace(simonCardCW(CardDesignSpade, 8), 0), "same suit but higher")

	// An empty column still takes anything.
	assert.True(t, g.canPlace(simonCardCW(CardDesignSpade, 5), 1), "empty column takes any card")
}

func TestCurdsAndWhey_ResetDeals(t *testing.T) {
	g := newSsGameCW()
	if g.GetPhase() != CurdsAndWheyPhasePlaying {
		t.Fatalf("phase = %d, want Playing", g.GetPhase())
	}
	total := 0
	cols := g.GetColumns()
	for i, want := range curdsAndWheyDeal {
		if len(cols[i]) != want {
			t.Errorf("col %d = %d cards, want %d", i, len(cols[i]), want)
		}
		total += len(cols[i])
	}
	if total != 52 {
		t.Errorf("dealt %d, want 52", total)
	}
}

func TestCurdsAndWhey_MoveSingleAndRun(t *testing.T) {
	g := newSsGameCW()
	ssClearCW(g)
	// col0: 9♠; col1: 8♠ -> move 8♠ onto 9♠.
	g.columns[0] = []*Card{simonCardCW(CardDesignSpade, 9)}
	g.columns[1] = []*Card{simonCardCW(CardDesignSpade, 8)}
	if err := g.MoveSequence(1, 0, 0); err != nil {
		t.Fatalf("move single: %v", err)
	}
	if len(g.columns[0]) != 2 || len(g.columns[1]) != 0 {
		t.Errorf("unexpected state: %d / %d", len(g.columns[0]), len(g.columns[1]))
	}
	// Move a same-suit run (7♠6♠) from col2 onto the 8♠ now on col0.
	g.columns[2] = []*Card{simonCardCW(CardDesignSpade, 7), simonCardCW(CardDesignSpade, 6)}
	if err := g.MoveSequence(2, 0, 0); err != nil {
		t.Fatalf("move run: %v", err)
	}
	if len(g.columns[0]) != 4 {
		t.Errorf("col0 = %d, want 4", len(g.columns[0]))
	}
}

func TestCurdsAndWhey_IllegalMoves(t *testing.T) {
	g := newSsGameCW()
	ssClearCW(g)
	g.columns[0] = []*Card{simonCardCW(CardDesignSpade, 9)}
	// Non-run group (8♠ 6♠) cannot move together.
	g.columns[1] = []*Card{simonCardCW(CardDesignSpade, 8), simonCardCW(CardDesignSpade, 6)}
	if err := g.MoveSequence(1, 0, 0); err == nil {
		t.Error("expected error: not a run")
	}
	// Wrong rank placement (7♠ onto 9♠).
	g.columns[2] = []*Card{simonCardCW(CardDesignSpade, 7)}
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

func TestCurdsAndWhey_PlacementRequiresSuitOrRank(t *testing.T) {
	g := newSsGameCW()
	ssClearCW(g)
	// Simple Simon accepted ANY suit one rank higher, and this test asserted it.
	// Curds and Whey needs the same suit -- or the same rank.
	g.columns[0] = []*Card{simonCardCW(CardDesignSpade, 9)}
	g.columns[1] = []*Card{simonCardCW(CardDesignHeart, 8)}
	if err := g.MoveSequence(1, 0, 0); err == nil {
		t.Fatal("8H onto 9S must be rejected: right rank, wrong suit")
	}

	// Same suit, one lower: legal.
	g.columns[1] = []*Card{simonCardCW(CardDesignSpade, 8)}
	if err := g.MoveSequence(1, 0, 0); err != nil {
		t.Fatalf("8S onto 9S should be legal: %v", err)
	}

	// Same rank: legal (the "temporary stack" move).
	ssClearCW(g)
	g.columns[0] = []*Card{simonCardCW(CardDesignSpade, 9)}
	g.columns[1] = []*Card{simonCardCW(CardDesignHeart, 9)}
	if err := g.MoveSequence(1, 0, 0); err != nil {
		t.Fatalf("9H onto 9S should be legal (same rank): %v", err)
	}
}

func TestCurdsAndWhey_CompleteSuitRemoved(t *testing.T) {
	g := newSsGameCW()
	ssClearCW(g)
	// col0 holds K..2 of spades; move the lone A♠ on top to complete K..A.
	col := make([]*Card, 0, 13)
	for v := 13; v >= 2; v-- {
		col = append(col, simonCardCW(CardDesignSpade, v))
	}
	g.columns[0] = col
	g.columns[1] = []*Card{simonCardCW(CardDesignSpade, 1)}
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

func TestCurdsAndWhey_WinClearsGame(t *testing.T) {
	g := newSsGameCW()
	ssClearCW(g)
	g.completedSuits = 3
	// Build the fourth suit (hearts) K..A and complete it.
	col := make([]*Card, 0, 13)
	for v := 13; v >= 2; v-- {
		col = append(col, simonCardCW(CardDesignHeart, v))
	}
	g.columns[0] = col
	g.columns[1] = []*Card{simonCardCW(CardDesignHeart, 1)}
	if err := g.MoveSequence(1, 0, 0); err != nil {
		t.Fatalf("final move: %v", err)
	}
	if g.GetPhase() != CurdsAndWheyPhaseGameClear {
		t.Errorf("phase = %d, want GameClear", g.GetPhase())
	}
}

func TestCurdsAndWhey_GameOverWhenStuck(t *testing.T) {
	g := newSsGameCW()
	ssClearCW(g)
	// **All THIRTEEN columns must be occupied** -- the Simple Simon version
	// filled ten and would now leave three empty columns, each of which accepts
	// any card. Ranks are 1..13 so no two tops share a rank (same-rank stacking
	// is legal here), and the suit cycles with the rank so any two adjacent
	// ranks land in different suits (same-suit descending is the other legal
	// move). Nothing connects.
	suits := []int{CardDesignSpade, CardDesignHeart, CardDesignDiamond, CardDesignClover}
	for i := 0; i < CurdsAndWheyColCnt; i++ {
		rank := i + 1
		g.columns[i] = []*Card{simonCardCW(suits[rank%len(suits)], rank)}
	}
	if g.hasAnyLegalMove() {
		t.Fatal("setup should have no legal move")
	}
	g.checkGameOver()
	if g.GetPhase() != CurdsAndWheyPhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestCurdsAndWhey_GiveUp(t *testing.T) {
	g := newSsGameCW()
	g.GiveUp()
	if g.GetPhase() != CurdsAndWheyPhaseGameOver {
		t.Errorf("phase = %d, want GameOver", g.GetPhase())
	}
}

func TestCurdsAndWhey_UndoAndHint(t *testing.T) {
	g := newSsGameCW()
	ssClearCW(g)
	g.columns[0] = []*Card{simonCardCW(CardDesignSpade, 9)}
	g.columns[1] = []*Card{simonCardCW(CardDesignSpade, 8)}
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

func TestCurdsAndWhey_JSONRoundTrip(t *testing.T) {
	g := newSsGameCW()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	g2 := NewDefaultCurdsAndWhey()
	if err := json.Unmarshal(data, g2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g2.GetColumns()[0] == nil || len(g2.GetColumns()[0]) != len(g.GetColumns()[0]) {
		t.Error("column mismatch after round-trip")
	}
	// Reject an out-of-range completed-suit count.
	if err := json.Unmarshal([]byte(`{"cs":9}`), NewDefaultCurdsAndWhey()); err == nil {
		t.Error("expected error for invalid completed suits")
	}
}

// **まとめて掴めるのは末尾の同スート降順だけ** (#5679)。画面はその境界を出すので、
// 位置を返すこの関数がずれると、掴めない札まで掴めるように見える。
func TestCurdsAndWheyMovableFrom(t *testing.T) {
	card := func(d, v int) *Card { return NewCard(d, v, false) }

	t.Run("an empty column has no boundary", func(t *testing.T) {
		assert.Equal(t, 0, CurdsAndWheyMovableFrom(nil))
	})

	t.Run("a single card is the whole run", func(t *testing.T) {
		assert.Equal(t, 0, CurdsAndWheyMovableFrom([]*Card{card(CardDesignSpade, 5)}))
	})

	// ♠K ♥9 ♠8 ♠7 → 末尾 2 枚だけが同スート降順。
	t.Run("stops where the suit breaks", func(t *testing.T) {
		col := []*Card{
			card(CardDesignSpade, 13),
			card(CardDesignHeart, 9),
			card(CardDesignSpade, 8),
			card(CardDesignSpade, 7),
		}
		assert.Equal(t, 2, CurdsAndWheyMovableFrom(col))
	})

	// ♠9 ♠8 ♠6 → ランクが飛ぶので末尾 1 枚だけ。
	t.Run("stops where the rank skips", func(t *testing.T) {
		col := []*Card{
			card(CardDesignSpade, 9),
			card(CardDesignSpade, 8),
			card(CardDesignSpade, 6),
		}
		assert.Equal(t, 2, CurdsAndWheyMovableFrom(col))
	})

	// 列全体が 1 つの並びなら先頭から掴める。
	t.Run("a column that is entirely one run starts at zero", func(t *testing.T) {
		col := []*Card{
			card(CardDesignSpade, 5),
			card(CardDesignSpade, 4),
			card(CardDesignSpade, 3),
		}
		assert.Equal(t, 0, CurdsAndWheyMovableFrom(col))
	})
}
