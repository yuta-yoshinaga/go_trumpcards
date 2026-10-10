package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThreeCard_BetSetsPlayerHandRankDuringAction(t *testing.T) {
	tc := NewThreeCard(NewTrumpCards(0))
	tc.chips.SetChips(ThreeCardDefaultChips)
	tc.trumpCards.deck = []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 7, false),
	}
	tc.trumpCards.deckCnt = len(tc.trumpCards.deck)
	tc.trumpCards.deckInit()

	require.NoError(t, tc.Bet(100, 0))

	require.Equal(t, ThreeCardPhaseAction, tc.phase)
	require.Equal(t, ThreeCardHandPair, tc.playerHandRank)
	require.Equal(t, evalThreeCardHand(tc.playerHand), tc.playerHandRank)
	require.NotZero(t, tc.playerHandRank)
	require.Zero(t, tc.dealerHandRank)
}
