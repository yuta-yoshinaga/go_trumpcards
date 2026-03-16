//go:build test
// +build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryManager_AddMemory(t *testing.T) {
	t.Run("add to empty", func(t *testing.T) {
		m := &memoryManager[testMemoryEntry]{}
		m.AddMemory(testMemoryEntry{turnSeen: 5})
		assert.Len(t, m.cardMemories, 1)
		assert.Equal(t, 5, m.cardMemories[0].turnSeen)
	})

	t.Run("add to existing", func(t *testing.T) {
		m := &memoryManager[testMemoryEntry]{
			cardMemories: []testMemoryEntry{{turnSeen: 1}},
		}
		m.AddMemory(testMemoryEntry{turnSeen: 3})
		assert.Len(t, m.cardMemories, 2)
		assert.Equal(t, 3, m.cardMemories[1].turnSeen)
	})
}

func TestMemoryManager_ResetMemory(t *testing.T) {
	t.Run("reset clears all memories", func(t *testing.T) {
		m := &memoryManager[testMemoryEntry]{
			cardMemories: []testMemoryEntry{
				{turnSeen: 1},
				{turnSeen: 2},
			},
		}
		m.ResetMemory()
		assert.Nil(t, m.cardMemories)
	})

	t.Run("reset on nil is no-op", func(t *testing.T) {
		m := &memoryManager[testMemoryEntry]{}
		m.ResetMemory()
		assert.Nil(t, m.cardMemories)
	})
}

func TestMemoryManager_DecayMemories(t *testing.T) {
	t.Run("decay removes old memories with forgetProb >= 1.0", func(t *testing.T) {
		m := &memoryManager[testMemoryEntry]{
			cardMemories: []testMemoryEntry{
				{turnSeen: 1},  // age=9, prob=0.9*9=8.1 (>=1.0, always forgotten)
				{turnSeen: 10}, // age=0, prob=0.9*0=0.0 (never forgotten)
			},
		}
		m.DecayMemories(10, 0.9)
		assert.Len(t, m.cardMemories, 1)
		assert.Equal(t, 10, m.cardMemories[0].turnSeen)
	})

	t.Run("decay with zero rate keeps all", func(t *testing.T) {
		m := &memoryManager[testMemoryEntry]{
			cardMemories: []testMemoryEntry{
				{turnSeen: 1},
				{turnSeen: 2},
			},
		}
		m.DecayMemories(10, 0)
		assert.Len(t, m.cardMemories, 2)
	})

	t.Run("decay on empty memories", func(t *testing.T) {
		m := &memoryManager[testMemoryEntry]{}
		m.DecayMemories(10, 0.5)
		assert.Empty(t, m.cardMemories)
	})
}
