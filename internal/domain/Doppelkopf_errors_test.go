//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertDoppelkopfCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	assert.Equal(t, code, err.(*DomainError).MessageCode())
}

func TestDoppelkopfDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultDoppelkopf()
	g.Reset()
	g.SetPhase(DoppelkopfPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertDoppelkopfCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "doppelkopf.errCardIndexOutOfRange")

	g.SetTrickNumber(2)
	assertDoppelkopfCodedError(t, g.PlayerAnnounce(), ErrInvalidPlay, "doppelkopf.errAnnounceUnavailable")

	g.GetPlayer(0).ResetRound()
	g.SetTrickNumber(1)
	g.GetPlayer(0).AddCard(NewCard(CardDesignHeart, 2, false))
	g.GetPlayer(0).AddCard(NewCard(CardDesignClover, 2, false))
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 3, false)}})
	assertDoppelkopfCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "doppelkopf.errFollowLeadSuit")
}
