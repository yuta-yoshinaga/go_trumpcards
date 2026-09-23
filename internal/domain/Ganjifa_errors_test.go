//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertGanjifaCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	assert.Equal(t, code, err.(*DomainError).MessageCode())
}

func TestGanjifaDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultGanjifa()
	g.Reset()
	g.SetPhase(GanjifaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertGanjifaCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "ganjifa.errCardIndexOutOfRange")
	assertGanjifaCodedError(t, g.validatePlay(0, nil), ErrInvalidCard, "ganjifa.errCardMissing")

	g.GetPlayer(0).ResetRound()
	g.GetPlayer(0).AddCard(NewCard(1, 2, false))
	g.GetPlayer(0).AddCard(NewCard(2, 2, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(1, 3, false)}}
	assertGanjifaCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "ganjifa.errFollowLeadSuit")
}
