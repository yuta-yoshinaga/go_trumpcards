//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpadesAppendBidLogSeparatesNilFromNumericBid(t *testing.T) {
	players := []*SpadesPlayer{NewSpadesPlayer(true)}
	s := NewSpades(NewTrumpCards(0), players, DefaultSpadesConfig())

	s.appendBidLog(0, 0)
	assert.Equal(t, "spades.log.bidNil", s.actionLog[0].DetailCode)
	assert.Equal(t, map[string]string{"name": "You"}, s.actionLog[0].DetailParams)

	s.actionLog = nil
	s.appendBidLog(0, 3)
	assert.Equal(t, "spades.log.bid", s.actionLog[0].DetailCode)
	assert.Equal(t, map[string]string{"name": "You", "bid": "3"}, s.actionLog[0].DetailParams)
}
