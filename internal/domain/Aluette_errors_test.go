//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAluetteInvalidCardErrorHasMessageCode(t *testing.T) {
	g := newTestAluette(t)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerPlay(-1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCard)
	assert.Equal(t, "aluette.errCardIndexOutOfRange", err.(*DomainError).MessageCode())
}
