//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertMinchiateDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func setupMinchiateScartoErrorGame(t *testing.T) *Minchiate {
	t.Helper()
	g := newTestMinchiate(t)
	dealer := g.GetPlayer(g.GetDealerIdx())
	dealer.Reset()
	for i := 0; i < MinchiateHandSize+MinchiateSurplus; i++ {
		dealer.AddCard(NewCard(1, i+1, false))
	}
	return g
}

func TestMinchiateDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("scarto count", func(t *testing.T) {
		g := setupMinchiateScartoErrorGame(t)
		assertMinchiateDomainError(t, g.PlayerScarto(nil), ErrInvalidIndices, "minchiate.errScartoCardCount", map[string]string{"count": "13"})
	})

	t.Run("scarto index", func(t *testing.T) {
		g := setupMinchiateScartoErrorGame(t)
		indices := make([]int, MinchiateSurplus)
		for i := range indices {
			indices[i] = i
		}
		indices[len(indices)-1] = MinchiateHandSize + MinchiateSurplus
		assertMinchiateDomainError(t, g.PlayerScarto(indices), ErrInvalidCard, "minchiate.errCardIndexOutOfRange", nil)
	})

	t.Run("duplicate scarto", func(t *testing.T) {
		g := setupMinchiateScartoErrorGame(t)
		indices := make([]int, MinchiateSurplus)
		for i := range indices {
			indices[i] = i
		}
		indices[1] = indices[0]
		assertMinchiateDomainError(t, g.PlayerScarto(indices), ErrInvalidIndices, "minchiate.errDuplicateScarto", nil)
	})

	t.Run("cannot discard trump or Matto", func(t *testing.T) {
		g := setupMinchiateScartoErrorGame(t)
		g.GetPlayer(g.GetDealerIdx()).RemoveCard(0)
		g.GetPlayer(g.GetDealerIdx()).AddCard(NewCard(MinchiateTrumpDesign, 1, false))
		indices := make([]int, MinchiateSurplus)
		for i := range indices {
			indices[i] = i
		}
		indices[len(indices)-1] = MinchiateHandSize + MinchiateSurplus - 1
		assertMinchiateDomainError(t, g.PlayerScarto(indices), ErrInvalidPlay, "minchiate.errCannotDiscardTrumpOrMatto", nil)
	})

	t.Run("play index", func(t *testing.T) {
		g := newTestMinchiate(t)
		g.SetPhase(MinchiatePhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(NewCard(1, 1, false))
		assertMinchiateDomainError(t, g.PlayerPlay(1), ErrInvalidCard, "minchiate.errCardIndexOutOfRange", nil)
	})

	t.Run("follow lead suit", func(t *testing.T) {
		g := newTestMinchiate(t)
		g.SetPhase(MinchiatePhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(NewCard(2, 1, false))
		g.GetPlayer(0).AddCard(NewCard(1, 1, false))
		g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(1, 14, false)}}
		assertMinchiateDomainError(t, g.PlayerPlay(0), ErrInvalidPlay, "minchiate.errFollowLeadSuit", nil)
	})
}
