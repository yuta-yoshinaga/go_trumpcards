//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertBeziqueCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de := err.(*DomainError)
	assert.Equal(t, code, de.MessageCode())
}

func TestBeziqueDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultBezique()
	g.Reset()
	g.SetPhase(BeziquePhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertBeziqueCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "bezique.errCardIndexOutOfRange")

	g.GetPlayer(0).ResetGame()
	g.SetPhase(BeziquePhaseMeld)
	assertBeziqueCodedError(t, g.PlayerDeclareMeld(0), ErrInvalidPlay, "bezique.errNoMeldAvailable")
}
