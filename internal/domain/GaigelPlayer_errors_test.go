//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGaigelPlayerInvalidTeamErrorHasMessageCode(t *testing.T) {
	var player GaigelPlayer
	err := json.Unmarshal([]byte(`{"tm":-1}`), &player)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Equal(t, "gaigel.errTeamOutOfRange", err.(*DomainError).MessageCode())
}
