//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPreferenceCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	assert.Equal(t, code, err.(*DomainError).MessageCode())
}

func TestPreferenceDomainErrorsHaveMessageCodes(t *testing.T) {
	g := NewDefaultPreference()
	g.Reset()
	g.phase = PreferencePhaseBid
	g.currentPlayerIdx = 0
	assertPreferenceCodedError(t, g.PlayerBid(PreferenceBid(-1)), ErrInvalidPlay, "preference.errInvalidBid")

	g.bids[1] = PreferenceBidSix
	g.bidDone[1] = true
	assertPreferenceCodedError(t, g.PlayerBid(PreferenceBidSix), ErrInvalidPlay, "preference.errBidMustExceed")

	g.phase = PreferencePhasePlay
	g.currentPlayerIdx = 0
	assertPreferenceCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "preference.errCardIndexOutOfRange")
}
