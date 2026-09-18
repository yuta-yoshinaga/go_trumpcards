//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestMadrassoPlayerPlayInvalidCardReturnsMessageCode(t *testing.T) {
	g := domain.NewDefaultMadrasso()
	g.Reset()

	err := g.PlayerPlay(-1)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
	var domainErr *domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Empty(t, domainErr.Message)
	assert.Equal(t, "madrasso.errCardIndexOutOfRange", domainErr.MessageCode())
}
