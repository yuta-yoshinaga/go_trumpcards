//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertMaoCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	assert.Equal(t, code, err.(*DomainError).MessageCode())
}

func TestMaoDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultMao()
	g.Reset()
	g.SetPhase(MaoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertMaoCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "mao.errCardIndexOutOfRange")

	g.GetPlayer(0).ResetRound()
	g.SetDiscardPile([]*Card{NewCard(CardDesignSpade, 5, false)})
	g.GetPlayer(0).AddCard(NewCard(CardDesignHeart, 6, false))
	assertMaoCodedError(t, g.PlayerPlay(0), ErrInvalidPlay, "mao.errCardNotPlayable")

	g.SetPhase(MaoPhaseChooseSuit)
	assertMaoCodedError(t, g.PlayerChooseSuit(0), ErrInvalidPlay, "mao.errSuitOutOfRange")
}
