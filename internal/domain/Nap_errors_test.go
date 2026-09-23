//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNapErrorsReturnMessageCodes(t *testing.T) {
	players := make([]*domain.NapPlayer, domain.NapPlayerCnt)
	for i := range players {
		players[i] = domain.NewNapPlayer(true)
	}
	g := domain.NewNap(domain.NewTrumpCards(0), players, domain.DefaultNapConfig())
	g.Reset()

	err := g.PlayerBid(domain.NapBid(1))
	assertErrorCode(t, err, domain.ErrInvalidPlay, "nap.errInvalidBid")

	require.NoError(t, g.PlayerBid(domain.NapBidNap))
	err = g.PlayerBid(domain.NapBidThree)
	assertErrorCode(t, err, domain.ErrInvalidPlay, "nap.errBidMustExceed")

	g.SetPhase(domain.NapPhasePlay)
	g.SetCurrentPlayerIdx(0)
	err = g.PlayerPlay(-1)
	assertErrorCode(t, err, domain.ErrInvalidCard, "nap.errCardIndexOutOfRange")
}

func assertErrorCode(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	var domainErr *domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Empty(t, domainErr.Message)
	assert.Equal(t, code, domainErr.MessageCode())
}
