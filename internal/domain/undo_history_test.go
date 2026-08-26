//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

// TestAppendSnapshotCapsHistory pins the cap itself.
func TestAppendSnapshotCapsHistory(t *testing.T) {
	var h []int
	for i := 0; i < MaxUndoHistory*3; i++ {
		h = appendSnapshot(h, i)
		if len(h) > MaxUndoHistory {
			t.Fatalf("history grew to %d after %d appends, cap is %d", len(h), i+1, MaxUndoHistory)
		}
	}
	// The newest entries are the ones worth keeping: undo walks backwards.
	if got, want := h[len(h)-1], MaxUndoHistory*3-1; got != want {
		t.Errorf("newest = %d, want %d -- the cap dropped the wrong end", got, want)
	}
	if got, want := h[0], MaxUndoHistory*3-MaxUndoHistory; got != want {
		t.Errorf("oldest kept = %d, want %d", got, want)
	}
}

// TestAppendSnapshotTrimsAnOversizedHistoryInOneStep covers the sessions that
// are already in KV with hundreds of snapshots. Dropping one per move would
// leave them over budget -- and therefore still failing -- for another hundred
// moves.
func TestAppendSnapshotTrimsAnOversizedHistoryInOneStep(t *testing.T) {
	legacy := make([]int, 200)
	got := appendSnapshot(legacy, 999)
	if len(got) != MaxUndoHistory {
		t.Fatalf("len = %d after appending to a 200-entry history, want %d", len(got), MaxUndoHistory)
	}
	if got[len(got)-1] != 999 {
		t.Errorf("the new snapshot was trimmed away")
	}
}

// TestKlondikeKVPayloadStaysBounded is the regression test for the reported
// bug: klondike answered 503 / Cloudflare 1102 on move 117 because the payload
// it writes to KV grew without limit. It asserts the payload, not the slice,
// because the payload is what costs the Worker its CPU budget.
func TestKlondikeKVPayloadStaysBounded(t *testing.T) {
	k := NewDefaultKlondike()
	k.Reset()

	first, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	moves := 0
	for i := 0; i < 400; i++ {
		if err := k.Draw(); err != nil {
			break
		}
		moves++
	}
	// A test that never got the game moving would pass this vacuously.
	if moves < 150 {
		t.Fatalf("only %d draws succeeded; this cannot exercise the growth", moves)
	}

	last, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(k.history) > MaxUndoHistory {
		t.Errorf("history = %d snapshots after %d moves, cap is %d", len(k.history), moves, MaxUndoHistory)
	}

	// 217 KB is what actually drew the 1102 in production; with the cap this
	// reaches about 99 KB at 400 moves. The remainder is the action log, still
	// unbounded at roughly 107 bytes a move -- 14 times slower than the history
	// it replaces as the leading term, and left for its own change because
	// capping it moves turn numbering, which 22 games depend on. The budget sits
	// above the measured figure and well under what failed, so it catches a
	// regression in the cap without tripping whenever a board field is added.
	const budget = 120 * 1024
	if len(last) > budget {
		t.Errorf("KV payload = %d bytes after %d moves, budget is %d (it was %d at deal)",
			len(last), moves, budget, len(first))
	}
	t.Logf("payload: deal=%d bytes, after %d moves=%d bytes, history=%d",
		len(first), moves, len(last), len(k.history))
}
