//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertSchafkopfCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestSchafkopfDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("invalid contract", func(t *testing.T) {
		g := newSKGame(true)
		g.SetPhase(SchafkopfPhasePick)
		g.SetCurrentPlayerIdx(0)
		assertSchafkopfCodedError(t, g.PlayerDeclare(true, SchafkopfContract(99), 0), ErrInvalidPlay, "schafkopf.errInvalidContract")
	})
	t.Run("solo requires a trump suit", func(t *testing.T) {
		g := newSKGame(true)
		g.SetPhase(SchafkopfPhasePick)
		g.SetCurrentPlayerIdx(0)
		assertSchafkopfCodedError(t, g.PlayerDeclare(true, SchafkopfContractSolo, 0), ErrInvalidCard, "schafkopf.errTrumpSuitRequired")
	})
	t.Run("contract must be higher", func(t *testing.T) {
		g := newSKGame(true)
		g.SetPhase(SchafkopfPhasePick)
		g.resolvePick(1, true, SchafkopfContractWenz, 0)
		g.SetCurrentPlayerIdx(0)
		assertSchafkopfCodedError(t, g.PlayerDeclare(true, SchafkopfContractRufspiel, 0), ErrInvalidPlay, "schafkopf.errContractNotHigher")
	})
	t.Run("invalid call suit", func(t *testing.T) {
		g := newSKGame(true)
		g.SetPhase(SchafkopfPhaseCall)
		g.SetPickerIdx(0)
		g.SetCurrentPlayerIdx(0)
		assertSchafkopfCodedError(t, g.PlayerCall(CardDesignClover), ErrInvalidPlay, "schafkopf.errInvalidCallSuit")
	})
	t.Run("card index", func(t *testing.T) {
		g := newSKGame(true)
		g.SetPhase(SchafkopfPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assertSchafkopfCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "schafkopf.errCardIndexOutOfRange")
	})
	t.Run("must follow lead suit", func(t *testing.T) {
		g := newSKGame(true)
		g.SetPhase(SchafkopfPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: skCard(CardDesignClover, 1)}})
		skSetHand(g.GetPlayer(0), skCard(CardDesignClover, 10), skCard(CardDesignSpade, 10))
		assertSchafkopfCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "schafkopf.errFollowLeadSuit")
	})
}
