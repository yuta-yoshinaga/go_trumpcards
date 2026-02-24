package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPokerPlayer_GetHandName_Unknown(t *testing.T) {
	tpp := NewPokerPlayer()
	// Set handRank to invalid value via direct field access
	tpp.handRank = -1
	assert.Equal(t, "Unknown", tpp.GetHandName())
	tpp.handRank = 99
	assert.Equal(t, "Unknown", tpp.GetHandName())
}

func TestPokerPlayer_SetHandRank_Internal(t *testing.T) {
	tpp := NewPokerPlayer()
	tpp.handRank = PokerHandFlush
	assert.Equal(t, PokerHandFlush, tpp.GetHandRank())
	tpp.handRank = PokerHandRoyalFlush
	assert.Equal(t, PokerHandRoyalFlush, tpp.GetHandRank())
	tpp.handRank = 0
	assert.Equal(t, PokerHandHighCard, tpp.GetHandRank())
}

// TestPoker_DealerRespondToBet_FoldBranch_Deterministic tests the dealer fold
// path in dealerRespondToBet by calling the unexported method directly with
// controlled state. The fold decision uses rand.Intn(100), so a statistical
// loop verifies both fold and non-fold outcomes occur.
//
// This replaces the old TestPoker_PlayerBet_DealerFolds which looped up to
// 100 times through the full Reset()+PlayerBet() path.
func TestPoker_DealerRespondToBet_FoldBranch_Deterministic(t *testing.T) {
	t.Run("first fold branch: weak high card + bad pot odds + large bet", func(t *testing.T) {
		// Conditions for first fold branch (70% fold rate):
		//   rank == PokerHandHighCard
		//   !hasHighCard (no A/Q/K)
		//   potOdds > dealerFoldPotOddsThreshold (0.4)
		//   diff > PokerMinBet * dealerFoldBetMultiplierWeak (20)
		foldCount := 0
		noFoldCount := 0
		for i := 0; i < 200; i++ {
			tc := NewTrumpCards(0)
			player := NewPokerPlayer()
			dealer := NewPokerPlayer()
			player.SetChips(PokerDefaultChips)
			dealer.SetChips(PokerDefaultChips)
			p := NewPoker(tc, player, dealer)

			// Set up dealer with weak HighCard (no A/Q/K)
			dealer.cards = []*Card{
				NewCard(CardDesignClover, 2, false),
				NewCard(CardDesignHeart, 4, false),
				NewCard(CardDesignDiamond, 6, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 10, false),
			}

			// Set bets for the right conditions:
			// diff = playerBet - dealerBet = 100 - 0 = 100
			// potOdds = 100 / (20 + 100) = 0.833 > 0.4
			// diff = 100 > 20 (PokerMinBet * dealerFoldBetMultiplierWeak)
			p.phase = PokerPhaseDeal
			p.pot = 20
			p.playerBet = 100
			p.dealerBet = 0

			p.dealerRespondToBet()

			if p.folded == PokerFoldByDealer {
				foldCount++
			} else {
				noFoldCount++
			}
		}
		assert.True(t, foldCount > 0, "dealer should fold at least once (70% rate)")
		assert.True(t, noFoldCount > 0, "dealer should NOT fold at least once (30% rate)")
	})

	t.Run("second fold branch: weak high card + strong bet but good pot odds", func(t *testing.T) {
		// Conditions for second fold branch (50% fold rate):
		//   rank == PokerHandHighCard
		//   !hasHighCard (no A/Q/K)
		//   diff > PokerMinBet * dealerFoldBetMultiplierStrong (30)
		//   potOdds <= 0.4 (skips first branch)
		foldCount := 0
		noFoldCount := 0
		for i := 0; i < 200; i++ {
			tc := NewTrumpCards(0)
			player := NewPokerPlayer()
			dealer := NewPokerPlayer()
			player.SetChips(PokerDefaultChips)
			dealer.SetChips(PokerDefaultChips)
			p := NewPoker(tc, player, dealer)

			// Set up dealer with weak HighCard (no A/Q/K)
			dealer.cards = []*Card{
				NewCard(CardDesignClover, 2, false),
				NewCard(CardDesignHeart, 4, false),
				NewCard(CardDesignDiamond, 6, false),
				NewCard(CardDesignClover, 8, false),
				NewCard(CardDesignHeart, 10, false),
			}

			// To hit the second branch but not the first:
			// potOdds = diff / (pot + diff) <= 0.4  =>  pot >= diff * 1.5
			// diff = playerBet - dealerBet = 55 - 20 = 35 > 30
			// pot = 75, potOdds = 35 / (75 + 35) = 0.318 <= 0.4
			p.phase = PokerPhaseDeal
			p.pot = 75
			p.playerBet = 55
			p.dealerBet = 20

			p.dealerRespondToBet()

			if p.folded == PokerFoldByDealer {
				foldCount++
			} else {
				noFoldCount++
			}
		}
		assert.True(t, foldCount > 0, "dealer should fold at least once (50% rate)")
		assert.True(t, noFoldCount > 0, "dealer should NOT fold at least once (50% rate)")
	})
}
