//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryRandomAvailablePosition(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)

	t.Run("returns valid position", func(t *testing.T) {
		pos := m.randomAvailablePosition(-1)
		assert.GreaterOrEqual(t, pos, 0)
		assert.Less(t, pos, MemoryBoardSize)
	})

	t.Run("excludes specified position", func(t *testing.T) {
		// Mark all but positions 0 and 1 as taken
		for i := 2; i < MemoryBoardSize; i++ {
			m.board[i].Taken = true
		}
		pos := m.randomAvailablePosition(0)
		assert.Equal(t, 1, pos)
	})

	t.Run("returns -1 when no positions available", func(t *testing.T) {
		for i := 0; i < MemoryBoardSize; i++ {
			m.board[i].Taken = true
		}
		pos := m.randomAvailablePosition(-1)
		assert.Equal(t, -1, pos)
	})
}

func TestMemoryAllTaken(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)

	assert.False(t, m.allTaken())
	for i := 0; i < MemoryBoardSize; i++ {
		m.board[i].Taken = true
	}
	assert.True(t, m.allTaken())
}

func TestMemoryFlipPhaseTransitions(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)

	assert.Equal(t, MemoryPhaseFlip1, m.GetPhase())
	err := m.flip(0)
	assert.NoError(t, err)
	assert.Equal(t, MemoryPhaseFlip2, m.GetPhase())
	err = m.flip(2)
	assert.NoError(t, err)
	assert.Equal(t, MemoryPhaseResult, m.GetPhase())
}

func TestMemoryCpuFlipNoAvailableSecondCard(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)
	m.SetCurrentPlayerIdx(1)

	// Mark all but one card as taken → CPU can flip first but not second
	for i := 1; i < MemoryBoardSize; i++ {
		m.board[i].Taken = true
	}
	m.CpuFlip()
	// Only one card available, so CPU flips it and then can't find a second
	assert.Equal(t, MemoryPhaseFlip2, m.GetPhase())
}

func TestMemoryCpuFlipNoAvailableFirstCard(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)
	m.SetCurrentPlayerIdx(1)

	// Mark all cards as taken
	for i := 0; i < MemoryBoardSize; i++ {
		m.board[i].Taken = true
	}
	m.CpuFlip()
	assert.Equal(t, MemoryPhaseFlip1, m.GetPhase()) // nothing happened
}

func TestMemoryCpuFlipKnownPairAlreadyTaken(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)
	m.SetCurrentPlayerIdx(1)

	// CPU thinks it knows a pair, but those cards are already taken
	cpu := m.GetPlayer(1)
	cpu.RecordRevealedCard(0, 1, 1.0, 0)
	cpu.RecordRevealedCard(1, 1, 1.0, 0)
	m.board[0].Taken = true
	m.board[1].Taken = true

	m.CpuFlip()
	// Falls through to random pick
	assert.Equal(t, MemoryPhaseResult, m.GetPhase())
}
