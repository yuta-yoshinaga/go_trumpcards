//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViraConfigTargetRoundsNotMultipleErrorHasMessageCode(t *testing.T) {
	err := (ViraConfig{CpuDifficulty: ViraCpuDifficultyNormal, TargetRounds: ViraPlayerCnt + 1}).Validate()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Equal(t, "vira.errTargetRoundsNotMultiple", err.(*DomainError).MessageCode())
	assert.Equal(t, map[string]string{"min": "3", "rounds": "4"}, err.(*DomainError).MessageParams())
}
