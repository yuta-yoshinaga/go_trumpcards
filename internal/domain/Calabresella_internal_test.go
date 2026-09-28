//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalabresella_GetHighestBidDuringAuction(t *testing.T) {
	g := NewDefaultCalabresella()
	g.Reset()
	g.phase = CalabresellaPhaseBid

	assert.Equal(t, CalabresellaBidNone, g.GetHighestBid(), "all players have passed")
	assert.Equal(t, CalabresellaBidNone, g.GetWinningBid(), "winning bid remains unset during the auction")
	assert.True(t, g.isBidLegal(CalabresellaBidChiamo), "chiamo is legal when there is no bid")

	g.bids = [CalabresellaPlayerCnt]CalabresellaBid{
		CalabresellaBidChiamo,
		CalabresellaBidNone,
		CalabresellaBidNone,
	}
	assert.Equal(t, CalabresellaBidChiamo, g.GetHighestBid())
	assert.False(t, g.isBidLegal(CalabresellaBidChiamo), "chiamo cannot match the current highest bid")
	assert.Equal(t, CalabresellaBidNone, g.GetWinningBid(), "winning bid remains unset during the auction")

	g.bids[1] = CalabresellaBidSolo
	assert.Equal(t, CalabresellaBidSolo, g.GetHighestBid(), "solo is higher than chiamo")
	assert.Equal(t, CalabresellaBidNone, g.GetWinningBid(), "winning bid remains unset during the auction")
}
