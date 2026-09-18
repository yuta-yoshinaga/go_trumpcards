//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertBriscolaCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de := err.(*DomainError)
	assert.Equal(t, code, de.MessageCode())
}

func TestBriscolaDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultBriscola()
	g.Reset()
	g.SetPhase(BriscolaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertBriscolaCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "briscola.errCardIndexOutOfRange")

	g.GetPlayer(0).ResetGame()
	g.GetPlayer(0).AddCard(nil)
	assertBriscolaCodedError(t, g.PlayerPlay(0), ErrInvalidCard, "briscola.errCardNil")
}
