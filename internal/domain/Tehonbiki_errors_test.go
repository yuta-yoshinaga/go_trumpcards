//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTehonbikiInvalidHalfGroupErrorHasMessageCode(t *testing.T) {
	g := NewDefaultTehonbiki()
	err := g.PlaceBet([]int{1, 2, 4}, TehonbikiBetHalf, 50)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Equal(t, "tehonbiki.errInvalidHalfGroup", err.(*DomainError).MessageCode())
}
