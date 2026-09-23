//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertThirtyOneCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	assert.Equal(t, code, err.(*DomainError).MessageCode())
}

func TestThirtyOneDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultThirtyOne()
	g.Reset()
	g.SetPhase(ThirtyOnePhaseDraw)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile(nil)
	assertThirtyOneCodedError(t, g.PlayerDrawFromDiscard(), ErrInvalidPlay, "thirtyone.errDiscardPileEmpty")

	g.SetPhase(ThirtyOnePhaseDiscard)
	assertThirtyOneCodedError(t, g.PlayerDiscard(-1), ErrInvalidCard, "thirtyone.errCardIndexOutOfRange")

	g.SetPhase(ThirtyOnePhaseDraw)
	g.SetKnockerIdx(-1)
	assert.NoError(t, g.PlayerKnock())
	g.SetCurrentPlayerIdx(0)
	assertThirtyOneCodedError(t, g.PlayerKnock(), ErrInvalidPlay, "thirtyone.errAlreadyKnocked")
}
