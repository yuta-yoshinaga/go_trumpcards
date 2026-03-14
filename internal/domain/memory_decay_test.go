package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// testMemoryEntry テスト用のMemoryEntry実装
type testMemoryEntry struct {
	turnSeen int
}

func (e testMemoryEntry) GetTurnSeen() int { return e.turnSeen }

func TestDecayMemories_ForgetProbAboveOne(t *testing.T) {
	entries := []testMemoryEntry{
		{turnSeen: 0},
		{turnSeen: 1},
	}
	// decayRate=1.0, currentTurn=2 → age=2, forgetProb=2.0 >= 1.0 → always forget entry[0]
	// age=1, forgetProb=1.0 >= 1.0 → always forget entry[1]
	result := DecayMemories(entries, 2, 1.0)
	assert.Empty(t, result)
}

func TestDecayMemories_DecayRateZero_NeverForgets(t *testing.T) {
	entries := []testMemoryEntry{
		{turnSeen: 0},
		{turnSeen: 5},
		{turnSeen: 100},
	}
	result := DecayMemories(entries, 1000, 0.0)
	assert.Len(t, result, 3)
}

func TestDecayMemories_EmptySlice(t *testing.T) {
	result := DecayMemories([]testMemoryEntry{}, 10, 0.5)
	assert.Empty(t, result)
}

func TestDecayMemories_NilSlice(t *testing.T) {
	var entries []testMemoryEntry
	result := DecayMemories(entries, 10, 0.5)
	assert.Empty(t, result)
}

func TestDecayMemories_ProbabilisticDecay(t *testing.T) {
	// With moderate decay, over many attempts we should see both kept and forgotten outcomes
	keptAtLeastOnce := false
	forgotAtLeastOnce := false

	for attempt := 0; attempt < 1000; attempt++ {
		entries := []testMemoryEntry{{turnSeen: 9}} // age=1, forgetProb=0.5
		result := DecayMemories(entries, 10, 0.5)
		if len(result) == 1 {
			keptAtLeastOnce = true
		} else {
			forgotAtLeastOnce = true
		}
		if keptAtLeastOnce && forgotAtLeastOnce {
			return
		}
	}
	t.Fatal("expected both kept and forgotten outcomes within 1000 attempts")
}
