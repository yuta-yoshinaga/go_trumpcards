//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertWhistCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de := err.(*DomainError)
	assert.Equal(t, code, de.MessageCode())
}

func TestWhistDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultWhist()
	g.Reset()
	g.SetPhase(WhistPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertWhistCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "whist.errCardIndexOutOfRange")

	g.GetPlayer(0).ResetRound()
	g.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 9, false))
	g.GetPlayer(0).AddCard(NewCard(CardDesignHeart, 7, false))
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}})
	assertWhistCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "whist.errFollowLeadSuit")
}
