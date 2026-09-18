//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestThreeCardRummyRebetErrorHasMessageCode(t *testing.T) {
	err := domain.NewDefaultThreeCardRummy().Rebet()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	assert.Equal(t, "threecardrummy.errCannotRebet", err.(*domain.DomainError).MessageCode())
}
