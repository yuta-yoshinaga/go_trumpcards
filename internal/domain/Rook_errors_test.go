//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertRookCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func rookErrorExchange() *domain.Rook {
	g := rookNewGame()
	g.SetDeclarerIdx(0)
	g.GetPlayer(0).SetIsDeclarer(true)
	g.GetPlayer(0).Reset()
	for i := 0; i < domain.RookHandSize+domain.RookNestSize; i++ {
		g.GetPlayer(0).AddCard(rookCard((i%4)+1, (i%13)+1))
	}
	g.SetPhase(domain.RookPhaseNestExchange)
	g.SetCurrentPlayerIdx(0)
	return g
}

func rookErrorPlay() *domain.Rook {
	g := rookNewGame()
	g.SetPhase(domain.RookPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetTrumpColor(4)
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(rookCard(1, 7))
	g.GetPlayer(0).AddCard(rookCard(2, 7))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: rookCard(1, 10)}})
	return g
}

func TestRookDomainErrorsHaveMessageCodes(t *testing.T) {
	g := rookNewGame()
	g.SetBidPlayerIdx(0)
	assertRookCodedError(t, g.PlayerBid(71), domain.ErrInvalidPlay, "rook.errInvalidBid")

	g = rookNewGame()
	g.SetBidPlayerIdx(0)
	g.SetHighestBid(80)
	assertRookCodedError(t, g.PlayerBid(80), domain.ErrInvalidPlay, "rook.errHigherBidRequired")

	g = rookErrorExchange()
	assertRookCodedError(t, g.PlayerExchangeNest([]int{0, 1, 2, 3, 4}, 0), domain.ErrInvalidPlay, "rook.errInvalidTrumpColor")
	g = rookErrorExchange()
	assertRookCodedError(t, g.PlayerExchangeNest([]int{0, 1, 2}, 1), domain.ErrInvalidCard, "rook.errDiscardFiveCards")
	g = rookErrorExchange()
	assertRookCodedError(t, g.PlayerExchangeNest([]int{0, 1, 2, 3, 99}, 1), domain.ErrInvalidCard, "rook.errCardIndexOutOfRange")
	g = rookErrorExchange()
	assertRookCodedError(t, g.PlayerExchangeNest([]int{0, 1, 2, 3, 3}, 1), domain.ErrInvalidCard, "rook.errDuplicateCardIndex")

	g = rookNewGame()
	g.SetPhase(domain.RookPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertRookCodedError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "rook.errCardIndexOutOfRange")

	g = rookErrorPlay()
	assertRookCodedError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "rook.errFollowLeadSuit")
}
