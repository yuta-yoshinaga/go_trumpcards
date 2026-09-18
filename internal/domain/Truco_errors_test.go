//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertTrucoCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de := err.(*DomainError)
	assert.Equal(t, code, de.MessageCode())
}

func TestTrucoDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultTruco()
	g.Reset()
	g.SetPhase(TrucoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertTrucoCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "truco.errCardIndexOutOfRange")

	g.SetAcceptedLevel(TrucoMaxLevel)
	assertTrucoCodedError(t, g.DeclareTruco(), ErrWrongPhase, "truco.errCannotRaiseNow")
}
