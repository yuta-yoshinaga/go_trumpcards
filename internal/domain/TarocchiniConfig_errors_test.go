//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTarocchiniConfigTargetRoundsNotMultipleErrorHasMessageCode(t *testing.T) {
	err := (TarocchiniConfig{CpuDifficulty: TarocchiniCpuDifficultyNormal, TargetRounds: TarocchiniPlayerCnt + 1}).Validate()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Equal(t, "tarocchini.errTargetRoundsNotMultiple", err.(*DomainError).MessageCode())
	assert.Equal(t, map[string]string{"min": "4", "rounds": "5"}, err.(*DomainError).MessageParams())
}
