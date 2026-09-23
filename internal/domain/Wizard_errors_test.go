//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertWizardCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de := err.(*DomainError)
	assert.Equal(t, code, de.MessageCode())
}

func TestWizardDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultWizard()
	g.Reset()
	g.SetBidPlayerIdx(0)
	assertWizardCodedError(t, g.PlayerBid(-1), ErrInvalidPlay, "wizard.errBidRange")

	g.SetPhase(WizardPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertWizardCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "wizard.errCardIndexOutOfRange")
	g.GetPlayer(0).ResetRound()
	g.GetPlayer(0).AddCard(NewCard(CardDesignHeart, 2, false))
	g.GetPlayer(0).AddCard(NewCard(CardDesignClover, 2, false))
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 3, false)}})
	assertWizardCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "wizard.errFollowLeadSuit")
}
