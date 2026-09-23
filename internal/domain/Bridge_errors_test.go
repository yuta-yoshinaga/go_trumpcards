//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertBridgeDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func TestBridgeBidErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultBridge()
	g.Reset()
	g.SetBidPlayerIdx(0)
	assertBridgeDomainError(t, g.PlayerBid(99, 1, domain.BridgeBidSuitClub), domain.ErrInvalidPlay, "bridge.errInvalidBidType")
	assertBridgeDomainError(t, g.PlayerBid(int(domain.BridgeBidNormal), 0, domain.BridgeBidSuitClub), domain.ErrInvalidPlay, "bridge.errBidLevelRange")
	assertBridgeDomainError(t, g.PlayerBid(int(domain.BridgeBidNormal), 1, 99), domain.ErrInvalidPlay, "bridge.errInvalidBidSuit")
	assert.NoError(t, g.PlayerBid(int(domain.BridgeBidNormal), 1, domain.BridgeBidSuitClub))
	g.SetBidPlayerIdx(0)
	assertBridgeDomainError(t, g.PlayerBid(int(domain.BridgeBidNormal), 1, domain.BridgeBidSuitClub), domain.ErrInvalidPlay, "bridge.errBidHigherThanContract")
	assertBridgeDomainError(t, g.PlayerBid(int(domain.BridgeBidDouble), 1, domain.BridgeBidSuitClub), domain.ErrInvalidPlay, "bridge.errCannotDoubleOwnTeam")
	g.SetDoubled(1)
	assertBridgeDomainError(t, g.PlayerBid(int(domain.BridgeBidDouble), 1, domain.BridgeBidSuitClub), domain.ErrInvalidPlay, "bridge.errAlreadyDoubled")
	g.SetDoubled(0)
	g.SetContractLevel(0)
	assertBridgeDomainError(t, g.PlayerBid(int(domain.BridgeBidDouble), 1, domain.BridgeBidSuitClub), domain.ErrInvalidPlay, "bridge.errCannotDoubleWithoutBid")
	assertBridgeDomainError(t, g.PlayerBid(int(domain.BridgeBidRedouble), 1, domain.BridgeBidSuitClub), domain.ErrInvalidPlay, "bridge.errCannotRedoubleWithoutDouble")
}

// リダブルできるのはダブルされた側だけ。相手側が試みる分岐は、人間以外が
// 最後にビッドした局面でしか起きない (PlayerBid は人間の手番しか通さない)。
func TestBridgeRedoubleByTheDoublingTeamHasMessageCode(t *testing.T) {
	g := domain.NewDefaultBridge()
	g.Reset()
	g.SetPhase(domain.BridgePhaseBid)
	g.SetBidPlayerIdx(0)
	g.SetContractLevel(1)
	g.SetDoubled(1)

	opponents := 1 - g.GetPlayer(0).GetTeam()
	g.SetLastBidTeam(opponents)
	require.NotEqual(t, g.GetPlayer(0).GetTeam(), g.GetLastBidTeam())

	assertBridgeDomainError(t, g.PlayerBid(int(domain.BridgeBidRedouble), 1, domain.BridgeBidSuitClub), domain.ErrInvalidPlay, "bridge.errCannotRedoubleOwnTeam")
}

func TestBridgePlayErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultBridge()
	g.Reset()
	g.SetPhase(domain.BridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).ResetRound()
	assertBridgeDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "bridge.errCardIndexOutOfRange")
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)}})
	assertBridgeDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "bridge.errFollowLeadSuit")
}
