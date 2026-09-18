//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestThreeCardRebetErrorHasMessageCode(t *testing.T) {
	err := domain.NewDefaultThreeCard().Rebet()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	assert.Equal(t, "threecard.errCannotRebet", err.(*domain.DomainError).MessageCode())
}
