//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertCourtPieceCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de := err.(*DomainError)
	assert.Equal(t, code, de.MessageCode())
}

func TestCourtPieceDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultCourtPiece()
	g.Reset()
	g.SetPhase(CourtPiecePhaseTrumpDeclaration)
	g.SetCallerIdx(0)
	assertCourtPieceCodedError(t, g.PlayerDeclareTrump(0), ErrInvalidPlay, "courtpiece.errTrumpSuitOutOfRange")

	g.SetPhase(CourtPiecePhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertCourtPieceCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "courtpiece.errCardIndexOutOfRange")
}
