//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertQuadrilleCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func quadrilleBidReady() *domain.Quadrille {
	g := newTestQuadrille()
	// The first three deals start with a CPU forehand; deal four gives the human
	// player the bid turn without relying on shuffled cards.
	g.SetPhase(domain.QuadrillePhaseRoundEnd)
	g.NextRound()
	g.SetPhase(domain.QuadrillePhaseRoundEnd)
	g.NextRound()
	g.SetPhase(domain.QuadrillePhaseRoundEnd)
	g.NextRound()
	return g
}

func TestQuadrilleDomainErrorsHaveMessageCodes(t *testing.T) {
	g := quadrilleBidReady()
	assertQuadrilleCodedError(t, g.PlayerBid(domain.QuadrilleBid(99), -1), domain.ErrInvalidPlay, "quadrille.errBidTooLow")

	g = quadrilleBidReady()
	assertQuadrilleCodedError(t, g.PlayerBid(domain.QuadrilleBidEntrar, 0), domain.ErrInvalidPlay, "quadrille.errInvalidTrumpSuit")

	g = newTestQuadrille()
	g.SetPhase(domain.QuadrillePhaseKingCall)
	g.SetQuadrilleIdx(0)
	assertQuadrilleCodedError(t, g.DeclareKing(1, domain.CardDesignSpade), domain.ErrInvalidPlay, "quadrille.errKingCallerOnly")

	g = newTestQuadrille()
	g.SetPhase(domain.QuadrillePhaseKingCall)
	g.SetQuadrilleIdx(0)
	assertQuadrilleCodedError(t, g.DeclareKing(0, 0), domain.ErrInvalidCard, "quadrille.errInvalidSuit")

	g = newTestQuadrille()
	g.SetPhase(domain.QuadrillePhaseKingCall)
	g.SetQuadrilleIdx(0)
	setQuadrilleHand(g, 0, quadrilleCard(domain.CardDesignSpade, 13))
	assertQuadrilleCodedError(t, g.DeclareKing(0, domain.CardDesignSpade), domain.ErrInvalidPlay, "quadrille.errOwnKingCannotBeCalled")

	g = newTestQuadrille()
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertQuadrilleCodedError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "quadrille.errCardIndexOutOfRange")

	g = newTestQuadrille()
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	setQuadrilleHand(g, 0, quadrilleCard(domain.CardDesignClover, 2), quadrilleCard(domain.CardDesignHeart, 2))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignClover, 3)}})
	assertQuadrilleCodedError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "quadrille.errFollowLeadSuit")
}
