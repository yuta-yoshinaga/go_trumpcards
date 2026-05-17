package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGapsPersistence_RoundTripPreservesState(t *testing.T) {
	g := NewDefaultGaps()
	g.Reset()
	// Force a known state: increment redeals and execute a redeal so history exists.
	if err := g.Redeal(); err != nil {
		t.Fatalf("Redeal failed: %v", err)
	}
	g.SetIsStalemate(true) // exercise the persistence of the bool field

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var restored Gaps
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if restored.GetMoveCount() != g.GetMoveCount() {
		t.Errorf("moveCount mismatch: got %d, want %d", restored.GetMoveCount(), g.GetMoveCount())
	}
	if restored.GetRedealsUsed() != g.GetRedealsUsed() {
		t.Errorf("redealsUsed mismatch: got %d, want %d", restored.GetRedealsUsed(), g.GetRedealsUsed())
	}
	if restored.GetPhase() != g.GetPhase() {
		t.Errorf("phase mismatch")
	}
	if restored.IsStalemate() != g.IsStalemate() {
		t.Errorf("isStalemate mismatch")
	}
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			a := g.GetGrid()[r][c]
			b := restored.GetGrid()[r][c]
			if (a == nil) != (b == nil) {
				t.Errorf("cell (%d,%d) gap-status differs", r, c)
				continue
			}
			if a != nil && (a.GetDesign() != b.GetDesign() || a.GetValue() != b.GetValue()) {
				t.Errorf("cell (%d,%d) card differs", r, c)
			}
		}
	}
}

func TestGapsPersistence_NilDeckFallsBackToDefault(t *testing.T) {
	var g Gaps
	if err := json.Unmarshal([]byte(`{}`), &g); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	// trumpCards must default to a usable deck so Reset() can be called.
	g.Reset()
	if g.GetPhase() != GapsPhasePlaying {
		t.Errorf("expected Playing after Reset on restored zero state")
	}
}

func TestGapsPersistence_RejectsOversizedHistory(t *testing.T) {
	// Build a payload claiming a huge history array — should be rejected.
	var entries []string
	for i := 0; i <= gapsMaxSliceLen; i++ {
		entries = append(entries, `{}`)
	}
	payload := `{"hi":[` + strings.Join(entries, ",") + `]}`
	var g Gaps
	if err := json.Unmarshal([]byte(payload), &g); err == nil {
		t.Error("expected oversized-history error")
	}
}
