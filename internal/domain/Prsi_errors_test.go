//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPrsiCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de := err.(*DomainError)
	assert.Equal(t, code, de.MessageCode())
}

func TestPrsiDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultPrsi()
	g.Reset()
	g.SetPhase(PrsiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertPrsiCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "prsi.errCardIndexOutOfRange")

	g.GetPlayer(0).Reset()
	g.SetDiscardPile([]*Card{NewCard(CardDesignSpade, 7, false)})
	g.GetPlayer(0).AddCard(NewCard(CardDesignHeart, 8, false))
	assertPrsiCodedError(t, g.PlayerPlay(0), ErrInvalidPlay, "prsi.errCardNotPlayable")
}
