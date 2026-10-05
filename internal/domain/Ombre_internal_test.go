//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOmbre_GetHighestBidDuringAuction(t *testing.T) {
	g := NewDefaultOmbre()
	g.Reset()
	g.phase = OmbrePhaseBid
	assert.Equal(t, OmbreBidNone, g.GetHighestBid(), "everyone has passed or not declared")
	assert.True(t, g.isBidLegal(OmbreBidEntrar), "Entrar is legal when there is no bid")

	g.bids[0] = OmbreBidEntrar
	assert.Equal(t, OmbreBidEntrar, g.GetHighestBid())
	assert.False(t, g.isBidLegal(OmbreBidEntrar), "Entrar cannot match the current highest bid")

	g.bids[1] = OmbreBidSolo
	assert.Equal(t, OmbreBidSolo, g.GetHighestBid(), "Solo outranks Entrar")
	assert.Equal(t, OmbreBidNone, g.GetWinningBid(), "the winning bid is unset during the auction")
}

func TestOmbre_GetHighestBidderIdxUsesForehandTieBreak(t *testing.T) {
	g := NewDefaultOmbre()
	g.forehandIdx = 1
	g.phase = OmbrePhaseBid
	g.bids[0] = OmbreBidSolo
	g.bids[1] = OmbreBidSolo
	assert.Equal(t, 1, g.GetHighestBidderIdx())
	g.bids[1] = OmbreBidNone
	assert.Equal(t, 0, g.GetHighestBidderIdx())
}
