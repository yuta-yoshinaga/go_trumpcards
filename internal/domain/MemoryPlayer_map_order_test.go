//go:build test

package domain

import "testing"

func TestMemoryPlayerFindAnyKnownPairMapOrder(t *testing.T) {
	p := NewMemoryPlayer(false)
	p.cardMemories = []memoryCardEntry{{rank: 9, position: 0}, {rank: 9, position: 1}, {rank: 4, position: 2}, {rank: 4, position: 3}}
	for i := 0; i < 50; i++ {
		a, b, ok := p.FindAnyKnownPair()
		if !ok || a != 2 || b != 3 {
			t.Fatalf("iteration %d: (%d,%d,%v)", i, a, b, ok)
		}
	}
}
