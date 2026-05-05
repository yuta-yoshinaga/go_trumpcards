//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryPlayer_JSONRoundTrip_PreservesCardMemories verifies that #1655's
// memoryManager state survives a Marshal/Unmarshal cycle. Before the fix the
// CPU memory was intentionally dropped, causing Hard difficulty CPUs to forget
// every revealed card after a session restore.
func TestMemoryPlayer_JSONRoundTrip_PreservesCardMemories(t *testing.T) {
	p := NewMemoryPlayer(false)
	p.AddMemory(memoryCardEntry{position: 3, rank: 7, turnSeen: 2})
	p.AddMemory(memoryCardEntry{position: 12, rank: 7, turnSeen: 4})
	p.AddMemory(memoryCardEntry{position: 25, rank: 13, turnSeen: 5})

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored MemoryPlayer
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, 3, restored.GetMemoryCount())
	pos1, pos2, found := restored.FindKnownMatch(7)
	assert.True(t, found)
	assert.Equal(t, 3, pos1)
	assert.Equal(t, 12, pos2)
}
