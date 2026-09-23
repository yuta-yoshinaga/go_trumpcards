//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertOmbreDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func allHumanOmbre() *domain.Ombre {
	players := make([]*domain.OmbrePlayer, domain.OmbrePlayerCnt)
	for i := range players {
		players[i] = domain.NewOmbrePlayer(true)
	}
	return domain.NewOmbre(domain.NewTrumpCards(0), players, domain.DefaultOmbreConfig())
}

func TestOmbreDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("trump suit required", func(t *testing.T) {
		g := allHumanOmbre()
		g.Reset()
		g.SetPhase(domain.OmbrePhaseBid)
		for g.GetCurrentBidderIdx() != 0 {
			require.NoError(t, g.PlayerBid(domain.OmbreBidNone, -1))
		}
		assertOmbreDomainError(t, g.PlayerBid(domain.OmbreBidEntrar, -1), domain.ErrInvalidPlay, "ombre.errTrumpSuitRequired")
	})

	t.Run("bid too low", func(t *testing.T) {
		g := allHumanOmbre()
		g.Reset()
		g.SetPhase(domain.OmbrePhaseBid)
		require.Equal(t, 1, g.GetCurrentBidderIdx())
		require.NoError(t, g.PlayerBid(domain.OmbreBidNone, -1))
		require.NoError(t, g.PlayerBid(domain.OmbreBidEntrar, domain.CardDesignHeart))
		assertOmbreDomainError(t, g.PlayerBid(domain.OmbreBidEntrar, domain.CardDesignHeart), domain.ErrInvalidPlay, "ombre.errBidTooLow")
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newTestOmbre()
		g.SetPhase(domain.OmbrePhasePlay)
		g.SetCurrentPlayerIdx(0)
		setOmbreHand(g, 0, ombreCard(domain.CardDesignSpade, 7))
		assertOmbreDomainError(t, g.PlayerPlay(1), domain.ErrInvalidCard, "ombre.errCardIndexOutOfRange")
	})

	t.Run("must follow lead suit", func(t *testing.T) {
		g := newTestOmbre()
		g.SetPhase(domain.OmbrePhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: ombreCard(domain.CardDesignSpade, 13)}})
		setOmbreHand(g, 0, ombreCard(domain.CardDesignSpade, 7), ombreCard(domain.CardDesignClover, 6))
		assertOmbreDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "ombre.errFollowLeadSuit")
	})
}
