//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertTarocchiniCodedError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestTarocchiniDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("discard count", func(t *testing.T) {
		g := newTestTarocchini(t)
		assertTarocchiniCodedError(t, g.PlayerScarto(nil), ErrInvalidIndices, "tarocchini.errDiscardCount", map[string]string{"count": "2"})
	})

	t.Run("discard card index", func(t *testing.T) {
		g := newTestTarocchini(t)
		assertTarocchiniCodedError(t, g.PlayerScarto([]int{0, 99}), ErrInvalidCard, "tarocchini.errCardIndexOutOfRange", nil)
	})

	t.Run("duplicate discard", func(t *testing.T) {
		g := newTestTarocchini(t)
		assertTarocchiniCodedError(t, g.PlayerScarto([]int{0, 0}), ErrInvalidIndices, "tarocchini.errDuplicateCard", nil)
	})

	t.Run("trump cannot be discarded", func(t *testing.T) {
		g := newTestTarocchini(t)
		dealer := g.players[g.dealerIdx]
		dealer.Reset()
		dealer.AddCard(NewCard(TarocchiniTrumpDesign, 7, false))
		dealer.AddCard(NewCard(1, 6, false))
		assertTarocchiniCodedError(t, g.PlayerScarto([]int{0, 1}), ErrInvalidPlay, "tarocchini.errCannotDiscardTrumpOrMatto", nil)
	})

	t.Run("play card index", func(t *testing.T) {
		g := newTestTarocchini(t)
		g.SetPhase(TarocchiniPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assertTarocchiniCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "tarocchini.errCardIndexOutOfRange", nil)
	})

	t.Run("must follow lead suit", func(t *testing.T) {
		g := newTestTarocchini(t)
		g.SetPhase(TarocchiniPhasePlay)
		g.SetCurrentPlayerIdx(0)
		p := g.players[0]
		p.Reset()
		p.AddCard(NewCard(1, 9, false))
		p.AddCard(NewCard(2, 14, false))
		g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(1, 6, false)}}
		assertTarocchiniCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "tarocchini.errFollowLeadSuit", nil)
	})
}
