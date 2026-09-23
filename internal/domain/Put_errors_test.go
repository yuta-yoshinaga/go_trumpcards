//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPutCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de := err.(*DomainError)
	assert.Equal(t, code, de.MessageCode())
}

func TestPutDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultPut()
	g.Reset()
	g.SetPhase(PutPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertPutCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "put.errCardIndexOutOfRange")

	g.SetAcceptedLevel(PutMaxLevel)
	assertPutCodedError(t, g.DeclarePut(), ErrWrongPhase, "put.errCannotRaiseNow")
}
