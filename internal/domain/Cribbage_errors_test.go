//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertCribbageDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestCribbageDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newTestCribbage()
	setupDiscardPhase(g)
	assertCribbageDomainError(t, g.PlayerDiscard([]int{0}), ErrInvalidIndices, "cribbage.errDiscardCount", map[string]string{"count": "2"})

	g = newTestCribbage()
	setupDiscardPhase(g)
	assertCribbageDomainError(t, g.PlayerDiscard([]int{-1, 0}), ErrInvalidCard, "cribbage.errCardIndexOutOfRange", nil)

	g = newTestCribbage()
	setupDiscardPhase(g)
	assertCribbageDomainError(t, g.PlayerDiscard([]int{0, 0}), ErrInvalidIndices, "cribbage.errDuplicateCardIndex", nil)

	g = newTestCribbage()
	setupPeggingPhase(g)
	assertCribbageDomainError(t, g.PlayerPeg(-1), ErrInvalidCard, "cribbage.errCardIndexOutOfRange", nil)

	g = newTestCribbage()
	setupPeggingPhase(g)
	g.SetPegCount(25)
	assertCribbageDomainError(t, g.PlayerPeg(3), ErrInvalidPlay, "cribbage.errPeggingLimitExceeded", map[string]string{"limit": "31"})

	g = newTestCribbage()
	setupPeggingPhase(g)
	assertCribbageDomainError(t, g.PlayerGo(), ErrInvalidPlay, "cribbage.errPlayableCardsRemain", nil)
}
