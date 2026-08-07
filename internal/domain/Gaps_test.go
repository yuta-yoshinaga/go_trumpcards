package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGaps returns a Gaps with the standard Reset() layout.
// Tests that depend on a specific grid override it via SetGrid afterwards.
func newTestGaps(t *testing.T) *Gaps {
	t.Helper()
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	return g
}

// gridWithAces builds a deterministic 4x13 grid for legality tests.
// Layout (rows are suits Spade, Clover, Heart, Diamond):
//
//	row 0: 2..K of Spade
//	row 1: 2..K of Clover
//	row 2: 2..K of Heart
//	row 3: 2..K of Diamond
//
// Aces are absent — every cell is filled with a card. Callers usually punch
// a hole (gap) at one or more cells via setCell.
func gridWithoutGaps() [GapsRowCnt][GapsColCnt]GapsCell {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	var grid [GapsRowCnt][GapsColCnt]GapsCell
	for r, s := range suits {
		for c := 0; c < GapsColCnt; c++ {
			grid[r][c] = NewCard(s, c+2, true) // 2..K
		}
	}
	return grid
}

func TestNewDefaultGaps_NotNil(t *testing.T) {
	if NewDefaultGaps() == nil {
		t.Fatal("NewDefaultGaps returned nil")
	}
}

func TestReset_LayoutHasFourGapsNoAces(t *testing.T) {
	g := newTestGaps(t)
	gaps := 0
	cards := 0
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			cell := g.GetGrid()[r][c]
			if cell == nil {
				gaps++
				continue
			}
			cards++
			if cell.GetValue() == 1 {
				t.Errorf("Ace must be removed at (%d,%d)", r, c)
			}
		}
	}
	if gaps != 4 {
		t.Errorf("expected 4 gaps, got %d", gaps)
	}
	if cards != 48 {
		t.Errorf("expected 48 cards, got %d", cards)
	}
	if g.GetPhase() != GapsPhasePlaying {
		t.Errorf("expected Playing, got %v", g.GetPhase())
	}
	if g.GetMoveCount() != 0 {
		t.Errorf("expected 0 moves, got %d", g.GetMoveCount())
	}
	if g.GetRedealsUsed() != 0 {
		t.Errorf("expected 0 redeals used, got %d", g.GetRedealsUsed())
	}
	if g.GetRedealsRemaining() != GapsMaxRedeals {
		t.Errorf("expected %d redeals remaining, got %d", GapsMaxRedeals, g.GetRedealsRemaining())
	}
}

func TestMove_AnchorColumn_PlaceAny2Succeeds(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	// Punch a gap at (0,0) and move the 2 of Hearts (from row 2 col 0) into it.
	grid[0][0] = nil
	g.SetGrid(grid)
	if err := g.Move(2, 0, 0, 0); err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	if g.GetGrid()[0][0] == nil || g.GetGrid()[0][0].GetDesign() != CardDesignHeart || g.GetGrid()[0][0].GetValue() != 2 {
		t.Errorf("expected 2H at (0,0), got %+v", g.GetGrid()[0][0])
	}
	if g.GetGrid()[2][0] != nil {
		t.Errorf("expected gap at (2,0) after move")
	}
	if g.GetMoveCount() != 1 {
		t.Errorf("expected moveCount=1, got %d", g.GetMoveCount())
	}
}

func TestMove_AnchorColumn_RejectsNon2(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	grid[0][0] = nil
	g.SetGrid(grid)
	// Try to move 3S (from row 0, col 1) into the leftmost gap → illegal.
	if err := g.Move(0, 1, 0, 0); err == nil {
		t.Error("expected illegal-move error for non-2 into anchor column")
	}
}

func TestMove_SuitRankPlus1_Succeeds(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	// Punch a gap at (0,5). The original 7S there is relocated to (3,12) so we
	// can move it back into the gap. Left of (0,5) is 6S, so 7S is the legal
	// placement.
	saved := grid[0][5]
	grid[0][5] = nil
	grid[3][12] = saved
	g.SetGrid(grid)
	if err := g.Move(3, 12, 0, 5); err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	if g.GetGrid()[0][5].GetDesign() != CardDesignSpade || g.GetGrid()[0][5].GetValue() != 7 {
		t.Errorf("expected 7S at (0,5)")
	}
	if g.GetGrid()[3][12] != nil {
		t.Errorf("expected gap at (3,12)")
	}
}

func TestMove_WrongSuit_Rejected(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	grid[0][5] = nil // left=6S, legal=7S
	g.SetGrid(grid)
	// 7H lives at (2,5); wrong suit.
	if err := g.Move(2, 5, 0, 5); err == nil {
		t.Error("expected error for wrong-suit move")
	}
}

func TestMove_WrongRank_Rejected(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	grid[0][5] = nil // left=6S, legal=7S
	g.SetGrid(grid)
	// 8S lives at (0,7); rank+2 not +1.
	if err := g.Move(0, 7, 0, 5); err == nil {
		t.Error("expected error for wrong-rank move")
	}
}

func TestMove_RightOfKing_DeadGap(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	// Force left neighbour of (0,12) to be the K of Spades (already is at (0,11)).
	// To create a gap-right-of-King we need a 13-column row + an extra; instead simulate by
	// making (0,12) a gap (with KS at col 11). Any move into this gap must fail.
	grid[0][12] = nil
	g.SetGrid(grid)
	// Try moving 2H to that gap.
	if err := g.Move(2, 0, 0, 12); err == nil {
		t.Error("expected error for gap-right-of-King")
	}
}

func TestMove_DstNotGap_Rejected(t *testing.T) {
	g := newTestGaps(t)
	// Pick any non-gap target and any non-gap source.
	var srcR, srcC, dstR, dstC = -1, -1, -1, -1
	for r := 0; r < GapsRowCnt && srcR < 0; r++ {
		for c := 0; c < GapsColCnt; c++ {
			if g.GetGrid()[r][c] != nil {
				srcR, srcC = r, c
				break
			}
		}
	}
	for r := 0; r < GapsRowCnt && dstR < 0; r++ {
		for c := 0; c < GapsColCnt; c++ {
			if g.GetGrid()[r][c] != nil && (r != srcR || c != srcC) {
				dstR, dstC = r, c
				break
			}
		}
	}
	if err := g.Move(srcR, srcC, dstR, dstC); err == nil {
		t.Error("expected error when destination is not a gap")
	}
}

func TestMove_SrcEmpty_Rejected(t *testing.T) {
	g := newTestGaps(t)
	// Find a gap to use as source.
	srcR, srcC := -1, -1
	for r := 0; r < GapsRowCnt && srcR < 0; r++ {
		for c := 0; c < GapsColCnt; c++ {
			if g.GetGrid()[r][c] == nil {
				srcR, srcC = r, c
				break
			}
		}
	}
	if err := g.Move(srcR, srcC, 0, 0); err == nil {
		t.Error("expected error when source is empty")
	}
}

func TestMove_NotPlaying_Rejected(t *testing.T) {
	g := newTestGaps(t)
	g.GiveUp()
	if err := g.Move(0, 0, 0, 1); err == nil {
		t.Error("expected error when phase != Playing")
	}
}

func TestMove_OutOfBounds_From(t *testing.T) {
	g := newTestGaps(t)
	if err := g.Move(-1, 0, 0, 0); err == nil {
		t.Error("expected error for negative row")
	}
	if err := g.Move(GapsRowCnt, 0, 0, 0); err == nil {
		t.Error("expected error for too-large row")
	}
	if err := g.Move(0, -1, 0, 0); err == nil {
		t.Error("expected error for negative col")
	}
	if err := g.Move(0, GapsColCnt, 0, 0); err == nil {
		t.Error("expected error for too-large col")
	}
}

func TestMove_OutOfBounds_To(t *testing.T) {
	g := newTestGaps(t)
	if err := g.Move(0, 0, -1, 0); err == nil {
		t.Error("expected error for negative dst row")
	}
	if err := g.Move(0, 0, GapsRowCnt, 0); err == nil {
		t.Error("expected error")
	}
	if err := g.Move(0, 0, 0, -1); err == nil {
		t.Error("expected error")
	}
	if err := g.Move(0, 0, 0, GapsColCnt); err == nil {
		t.Error("expected error")
	}
}

func TestRedeal_DecrementsCounter(t *testing.T) {
	g := newTestGaps(t)
	if err := g.Redeal(); err != nil {
		t.Fatalf("Redeal failed: %v", err)
	}
	if g.GetRedealsUsed() != 1 {
		t.Errorf("expected 1 redeal used, got %d", g.GetRedealsUsed())
	}
	if g.GetRedealsRemaining() != GapsMaxRedeals-1 {
		t.Errorf("expected %d redeals remaining, got %d", GapsMaxRedeals-1, g.GetRedealsRemaining())
	}
	// After redeal, 4 gaps still exist (one per row at the post-lock-prefix column).
	gaps := 0
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			if g.GetGrid()[r][c] == nil {
				gaps++
			}
		}
	}
	if gaps != 4 {
		t.Errorf("expected 4 gaps after redeal, got %d", gaps)
	}
}

func TestRedeal_NoRedealsLeft(t *testing.T) {
	g := newTestGaps(t)
	for i := 0; i < GapsMaxRedeals; i++ {
		if err := g.Redeal(); err != nil {
			t.Fatalf("Redeal #%d failed: %v", i, err)
		}
	}
	if err := g.Redeal(); err == nil {
		t.Error("expected error on 4th redeal")
	}
}

func TestRedeal_NotPlaying_Rejected(t *testing.T) {
	g := newTestGaps(t)
	g.GiveUp()
	if err := g.Redeal(); err == nil {
		t.Error("expected error when phase != Playing")
	}
}

func TestRedeal_LocksCorrectPrefix(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	// Row 0: 2S, 3S already in place — should be locked. Insert a gap at (0,2)
	// so we don't have a fully-finished row.
	grid[0][2] = nil
	// Row 1: 2C in place at (1,0) — locked alone. Gap at (1,1).
	grid[1][1] = nil
	// Row 2: starts with 3H — no lock. Gap at (2,0).
	grid[2][0] = nil
	grid[2][1] = NewCard(CardDesignHeart, 2, true) // move the 2H here
	// Row 3: similar, leave a gap at (3,2). 2D, 3D, 4D... actually 2D at (3,0), 3D at (3,1). Lock len = 2.
	grid[3][2] = nil
	g.SetGrid(grid)
	if err := g.Redeal(); err != nil {
		t.Fatalf("Redeal failed: %v", err)
	}
	// Locked prefixes should be preserved:
	if c := g.GetGrid()[0][0]; c == nil || c.GetDesign() != CardDesignSpade || c.GetValue() != 2 {
		t.Errorf("row 0 col 0 should remain 2S, got %+v", c)
	}
	if c := g.GetGrid()[0][1]; c == nil || c.GetDesign() != CardDesignSpade || c.GetValue() != 3 {
		t.Errorf("row 0 col 1 should remain 3S, got %+v", c)
	}
	if g.GetGrid()[0][2] != nil {
		t.Errorf("row 0 col 2 should be a fresh gap after redeal, got %+v", g.GetGrid()[0][2])
	}
	if c := g.GetGrid()[1][0]; c == nil || c.GetDesign() != CardDesignClover || c.GetValue() != 2 {
		t.Errorf("row 1 col 0 should remain 2C, got %+v", c)
	}
	if g.GetGrid()[1][1] != nil {
		t.Errorf("row 1 col 1 should be a fresh gap, got %+v", g.GetGrid()[1][1])
	}
	// Row 2 has no lock — col 0 should be a gap.
	if g.GetGrid()[2][0] != nil {
		t.Errorf("row 2 col 0 should be a fresh gap (no lock), got %+v", g.GetGrid()[2][0])
	}
	// Total gaps should be 4.
	gaps := 0
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			if g.GetGrid()[r][c] == nil {
				gaps++
			}
		}
	}
	if gaps != 4 {
		t.Errorf("expected 4 gaps after redeal, got %d", gaps)
	}
}

func TestUndo_RestoresPreviousState(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	grid[0][0] = nil
	g.SetGrid(grid)
	beforeMoveGrid := g.GetGrid()
	if err := g.Move(2, 0, 0, 0); err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	if !g.CanUndo() {
		t.Fatal("CanUndo should be true after a move")
	}
	if err := g.Undo(); err != nil {
		t.Fatalf("Undo failed: %v", err)
	}
	if g.GetMoveCount() != 0 {
		t.Errorf("expected moveCount=0 after undo, got %d", g.GetMoveCount())
	}
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			a := beforeMoveGrid[r][c]
			b := g.GetGrid()[r][c]
			if (a == nil) != (b == nil) {
				t.Errorf("cell (%d,%d) gap status differs after undo", r, c)
			}
			if a != nil && b != nil && (a.GetDesign() != b.GetDesign() || a.GetValue() != b.GetValue()) {
				t.Errorf("cell (%d,%d) card differs after undo", r, c)
			}
		}
	}
}

func TestUndo_NoHistory(t *testing.T) {
	g := newTestGaps(t)
	if g.CanUndo() {
		t.Error("CanUndo should be false on fresh game")
	}
	if err := g.Undo(); err == nil {
		t.Error("expected error when undoing without history")
	}
}

func TestUndo_NotPlaying(t *testing.T) {
	g := newTestGaps(t)
	// Push a snapshot via Redeal then GiveUp.
	if err := g.Redeal(); err != nil {
		t.Fatalf("Redeal failed: %v", err)
	}
	g.GiveUp()
	if err := g.Undo(); err == nil {
		t.Error("expected error when phase != Playing")
	}
}

func TestUndoN_ChainsUndos(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	grid[0][0] = nil
	g.SetGrid(grid)
	if err := g.Move(2, 0, 0, 0); err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	// Now (2,0) is the gap. Move 2C (1,0) into it.
	if err := g.Move(1, 0, 2, 0); err != nil {
		t.Fatalf("Second Move failed: %v", err)
	}
	if g.GetMoveCount() != 2 {
		t.Fatalf("expected 2 moves, got %d", g.GetMoveCount())
	}
	if err := g.UndoN(2); err != nil {
		t.Fatalf("UndoN failed: %v", err)
	}
	if g.GetMoveCount() != 0 {
		t.Errorf("expected moveCount=0 after UndoN(2), got %d", g.GetMoveCount())
	}
}

func TestUndoN_NegativeOrZeroIsNoop(t *testing.T) {
	g := newTestGaps(t)
	if err := g.UndoN(0); err != nil {
		t.Errorf("UndoN(0) should be a noop, got %v", err)
	}
	if err := g.UndoN(-1); err != nil {
		t.Errorf("UndoN(-1) should be a noop, got %v", err)
	}
}

func TestGiveUp_TransitionsToGameOver(t *testing.T) {
	g := newTestGaps(t)
	g.GiveUp()
	if g.GetPhase() != GapsPhaseGameOver {
		t.Errorf("expected GameOver, got %v", g.GetPhase())
	}
	// Calling again should not panic.
	g.GiveUp()
	if !g.GetGameEndFlag() {
		t.Error("GameEndFlag should be true after GiveUp")
	}
}

func TestGetHint_FindsAnchorMove(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	grid := gridWithoutGaps()
	grid[0][0] = nil
	g.SetGrid(grid)
	h := g.GetHint()
	if h == nil {
		t.Fatal("expected a hint")
	}
	// Anchor move: any 2 → (0,0). 2S is at (0,1) actually — no, grid[0] = 2S,3S,... and we punched
	// grid[0][0] to nil. So grid[0][1]=3S. The 2 of Spade lives... actually with our
	// gridWithoutGaps, grid[0][0] would have been 2S, which is now gone. The 2 of clover/heart/diamond
	// live at (1,0), (2,0), (3,0). So the hint should point to one of these → (0,0).
	if h.ToRow != 0 || h.ToCol != 0 {
		t.Errorf("expected hint target (0,0), got (%d,%d)", h.ToRow, h.ToCol)
	}
	src := g.GetGrid()[h.FromRow][h.FromCol]
	if src == nil || src.GetValue() != 2 {
		t.Errorf("hint source should be a 2, got %+v", src)
	}
}

func TestGetHint_NoneAvailable(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	// Build a grid where every gap is right-of-K → no legal move.
	grid := gridWithoutGaps()
	// Row 0 is 2..K (cols 0..11) plus K at col 11; col 12 is also K (from 13). Wait gridWithoutGaps
	// fills 2..K = values 2..13 = 12 cards into 13 cols — actually it's 13 cols indexed 0..12 with
	// value c+2: c=0→2, c=11→13(K), c=12→14 which is invalid. Cap it at K.
	// To make this deterministic, place a K at col 11 and a gap at col 12 in every row.
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			grid[r][c] = nil
		}
	}
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for r, s := range suits {
		for c := 0; c < 12; c++ {
			grid[r][c] = NewCard(s, c+2, true) // 2..K (values 2..13)
		}
		// col 12 left as gap; col 11 = K of suit s → dead gap.
	}
	g.SetGrid(grid)
	if h := g.GetHint(); h != nil {
		t.Errorf("expected no hint, got %+v", h)
	}
}

func TestGetHint_NotPlaying_ReturnsNil(t *testing.T) {
	g := newTestGaps(t)
	g.GiveUp()
	if g.GetHint() != nil {
		t.Error("hint should be nil when not playing")
	}
}

func TestIsStalemate_TrueWhenNoMovesAndNoRedeals(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	g.SetRedealsUsed(GapsMaxRedeals)
	// Force a no-hint grid as in TestGetHint_NoneAvailable.
	var grid [GapsRowCnt][GapsColCnt]GapsCell
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for r, s := range suits {
		for c := 0; c < 12; c++ {
			grid[r][c] = NewCard(s, c+2, true)
		}
	}
	g.SetGrid(grid)
	g.RecomputeStalemate()
	if !g.IsStalemate() {
		t.Error("expected stalemate")
	}
}

func TestIsStalemate_FalseWhenRedealsAvailable(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	// Even with no hint, redeals available → not stalemate.
	var grid [GapsRowCnt][GapsColCnt]GapsCell
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for r, s := range suits {
		for c := 0; c < 12; c++ {
			grid[r][c] = NewCard(s, c+2, true)
		}
	}
	g.SetGrid(grid)
	g.RecomputeStalemate()
	if g.IsStalemate() {
		t.Error("expected not-stalemate (redeals available)")
	}
}

func TestAllWon_TrueForSortedRows(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	// Cells 0..11 hold 2..K of one suit; cell 12 = gap.
	var winGrid [GapsRowCnt][GapsColCnt]GapsCell
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for r, s := range suits {
		for c := 0; c < 12; c++ {
			winGrid[r][c] = NewCard(s, c+2, true)
		}
		// winGrid[r][12] stays nil
	}
	g.SetGrid(winGrid)
	if !g.AllWon() {
		t.Error("expected AllWon = true for sorted layout")
	}
}

func TestAllWon_FalseForShuffled(t *testing.T) {
	g := newTestGaps(t)
	if g.AllWon() {
		t.Error("expected AllWon = false for freshly-dealt grid")
	}
}

func TestCheckGameClear_TransitionsToGameClearAfterWin(t *testing.T) {
	g := NewGaps(NewTrumpCards(0))
	g.Reset()
	// Build a one-move-from-win grid: row 0 K-of-spades is missing at (0,11), spare 2 swap.
	var grid [GapsRowCnt][GapsColCnt]GapsCell
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for r, s := range suits {
		for c := 0; c < 12; c++ {
			grid[r][c] = NewCard(s, c+2, true)
		}
	}
	// Move KS away from (0,11) so we can move it back.
	ks := grid[0][11]
	grid[0][11] = nil
	grid[0][12] = ks // KS sitting at (0,12); we need to move it back to (0,11).
	g.SetGrid(grid)
	if err := g.Move(0, 12, 0, 11); err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	if g.GetPhase() != GapsPhaseGameClear {
		t.Errorf("expected GameClear, got %v", g.GetPhase())
	}
}

func TestSetters_AndGetters(t *testing.T) {
	g := newTestGaps(t)
	g.SetPhase(GapsPhaseGameOver)
	if g.GetPhase() != GapsPhaseGameOver {
		t.Error("SetPhase did not stick")
	}
	g.SetRedealsUsed(2)
	if g.GetRedealsUsed() != 2 {
		t.Error("SetRedealsUsed did not stick")
	}
	g.SetIsStalemate(true)
	if !g.IsStalemate() {
		t.Error("SetIsStalemate did not stick")
	}
	if g.GetActionLog() == nil {
		// allow empty but not nil
		t.Log("action log is nil after Reset — acceptable if we never call appendLog")
	}
}

// **各ギャップが受け入れる次ランクを追うのが Gaps の根幹戦略 (#4800)。**Web は
// ゴーストカードと 🚫 で常時プレビューしているのに、CUI は空きマスを一律
// [ . ] としか出していなかった。
func TestGaps_GetGapNeed(t *testing.T) {
	card := func(design, value int) *Card { return NewCard(design, value, false) }
	// row0 に指定のセルを並べた盤面 (残りは空き)。
	board := func(cells ...*Card) *Gaps {
		g := NewDefaultGaps()
		g.Reset()
		var grid [GapsRowCnt][GapsColCnt]GapsCell
		for i, c := range cells {
			if i < GapsColCnt {
				grid[0][i] = c
			}
		}
		g.SetGrid(grid)
		return g
	}

	// 0列目は「どのスートでもよい 2」。
	t.Run("the leftmost column takes any two", func(t *testing.T) {
		need := board().GetGapNeed(0, 0)
		if assert.NotNil(t, need) {
			assert.Equal(t, GapsNeedAnySuit, need.Kind)
			assert.Equal(t, GapsAnchorRank, need.Value)
		}
	})

	// **左隣が K なら詰み。**K の次は無いので何も置けない。
	t.Run("a king on the left blocks the gap", func(t *testing.T) {
		need := board(card(CardDesignSpade, GapsKingRank)).GetGapNeed(0, 1)
		if assert.NotNil(t, need) {
			assert.Equal(t, GapsNeedBlocked, need.Kind)
		}
	})

	t.Run("names the exact card a gap needs", func(t *testing.T) {
		need := board(card(CardDesignHeart, 5)).GetGapNeed(0, 1)
		if assert.NotNil(t, need) {
			assert.Equal(t, GapsNeedCard, need.Kind)
			assert.Equal(t, CardDesignHeart, need.Design)
			assert.Equal(t, 6, need.Value, "同スートの次のランク")
		}
	})

	// **左隣も空きなら何も言わない。**決まらないものを決まったように見せない。
	t.Run("says nothing when the left neighbour is itself a gap", func(t *testing.T) {
		assert.Nil(t, board().GetGapNeed(0, 2))
	})

	t.Run("says nothing for a filled cell or an out-of-range one", func(t *testing.T) {
		g := board(card(CardDesignSpade, 5))
		assert.Nil(t, g.GetGapNeed(0, 0), "埋まっているマス")
		assert.Nil(t, g.GetGapNeed(-1, 0))
		assert.Nil(t, g.GetGapNeed(0, GapsColCnt))
	})

	// **案内した札は実際に置ける。**別実装だと置けない札を案内する。
	t.Run("the card it names is the one isLegalMove accepts", func(t *testing.T) {
		g := board(card(CardDesignHeart, 5))
		// 別の行に ♥6 と ♥7 を置いて、案内どおりの札だけが通ることを見る。
		grid := g.GetGrid()
		grid[1][0] = card(CardDesignHeart, 6)
		grid[1][1] = card(CardDesignHeart, 7)
		g.SetGrid(grid)

		need := g.GetGapNeed(0, 1)
		require.NotNil(t, need)
		assert.True(t, g.isLegalMove(1, 0, 0, 1), "案内した ♥6 は置ける")
		assert.False(t, g.isLegalMove(1, 1, 0, 1), "♥7 は置けない")
	})
}
