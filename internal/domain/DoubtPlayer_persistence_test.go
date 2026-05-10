//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDoubtPlayer_JSONRoundTrip_PreservesCardMemories verifies that #1655's
// CPU memory survives a Marshal/Unmarshal cycle for Doubt. Before the fix
// `cardMemories` was intentionally dropped on restore so the Hard CPU lost
// every value it had counted.
func TestDoubtPlayer_JSONRoundTrip_PreservesCardMemories(t *testing.T) {
	p := NewDoubtPlayer(false)
	p.AddMemory(cardMemoryEntry{value: 5, turnSeen: 1})
	p.AddMemory(cardMemoryEntry{value: 5, turnSeen: 3})
	p.AddMemory(cardMemoryEntry{value: 11, turnSeen: 4})

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored DoubtPlayer
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, 2, restored.CountKnownCards(5))
	assert.Equal(t, 1, restored.CountKnownCards(11))
	assert.Equal(t, 0, restored.CountKnownCards(7))
}
