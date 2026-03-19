package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOmahaPlayer(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	assert.Equal(t, 0, p.GetChips())
	assert.False(t, p.GetFolded())
	assert.False(t, p.GetAllIn())
	assert.Equal(t, 0, p.GetCurrentBet())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestOmahaPlayer_GetPlayStyleName(t *testing.T) {
	tests := []struct {
		style HoldemPlayStyle
		name  string
	}{
		{HoldemStyleTAG, "TAG"},
		{HoldemStyleLAP, "LAP"},
		{HoldemStyleTAP, "TAP"},
		{HoldemStyleLAG, "LAG"},
		{HoldemStyleGTO, "GTO"},
		{HoldemPlayStyle(99), "Unknown"},
	}
	for _, tt := range tests {
		p := NewOmahaPlayer(false, tt.style)
		assert.Equal(t, tt.name, p.GetPlayStyleName())
	}
}

func TestOmahaPlayer_EvalBestHand_OmahaRule(t *testing.T) {
	t.Run("must use exactly 2 hole + 3 community", func(t *testing.T) {
		p := NewOmahaPlayer(true, HoldemStyleTAG)
		// Hole: 4♠ 4♥ 4♦ 4♣ (four-of-a-kind in hand)
		p.AddCard(NewCard(CardDesignSpade, 4, false))
		p.AddCard(NewCard(CardDesignHeart, 4, false))
		p.AddCard(NewCard(CardDesignDiamond, 4, false))
		p.AddCard(NewCard(CardDesignClover, 4, false))

		// Community: A♠ K♠ Q♠ J♠ T♠ (royal flush on board)
		community := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignSpade, 12, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 10, false),
		}

		rank := p.EvalBestHand(community)
		// In Omaha, must use exactly 2 hole cards (both are 4s) + 3 community
		// Best hand is 4-4-A-K-Q → OnePair, NOT four-of-a-kind or straight flush
		assert.Equal(t, PokerHandOnePair, rank)
		assert.NotNil(t, p.GetBestHand())
		assert.Equal(t, 5, len(p.GetBestHand()))
	})
}

func TestOmahaPlayer_EvalBestHand_HighCard(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 2, false))
	p.AddCard(NewCard(CardDesignHeart, 4, false))
	p.AddCard(NewCard(CardDesignClover, 6, false))
	p.AddCard(NewCard(CardDesignDiamond, 8, false))

	community := []*Card{
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandHighCard, rank)
	assert.NotNil(t, p.GetBestHand())
	assert.Equal(t, 5, len(p.GetBestHand()))
}

func TestOmahaPlayer_EvalBestHand_OnePair(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))
	p.AddCard(NewCard(CardDesignClover, 2, false))
	p.AddCard(NewCard(CardDesignDiamond, 3, false))

	community := []*Card{
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandOnePair, rank)
}

func TestOmahaPlayer_EvalBestHand_TwoPair(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	// Hole: 10♠ 5♣ + other cards
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignClover, 5, false))
	p.AddCard(NewCard(CardDesignDiamond, 3, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))

	// Community: 10♦ 5♦ 8♠ (use 10♠+5♣ from hole + 10♦+5♦+8♠ from community = TwoPair)
	community := []*Card{
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandTwoPair, rank)
}

func TestOmahaPlayer_EvalBestHand_ThreeOfAKind(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))
	p.AddCard(NewCard(CardDesignClover, 2, false))
	p.AddCard(NewCard(CardDesignDiamond, 3, false))

	community := []*Card{
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandThreeOfAKind, rank)
}

func TestOmahaPlayer_EvalBestHand_Straight(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 6, false))
	p.AddCard(NewCard(CardDesignHeart, 7, false))
	p.AddCard(NewCard(CardDesignClover, 2, false))
	p.AddCard(NewCard(CardDesignDiamond, 3, false))

	community := []*Card{
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandStraight, rank)
}

func TestOmahaPlayer_EvalBestHand_Flush(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 2, false))
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignHeart, 3, false))
	p.AddCard(NewCard(CardDesignClover, 4, false))

	community := []*Card{
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 9, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandFlush, rank)
}

func TestOmahaPlayer_EvalBestHand_FullHouse(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))
	p.AddCard(NewCard(CardDesignClover, 2, false))
	p.AddCard(NewCard(CardDesignDiamond, 3, false))

	community := []*Card{
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 9, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandFullHouse, rank)
}

func TestOmahaPlayer_EvalBestHand_FourOfAKind(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))
	p.AddCard(NewCard(CardDesignClover, 2, false))
	p.AddCard(NewCard(CardDesignDiamond, 3, false))

	community := []*Card{
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 9, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandFourOfAKind, rank)
}

func TestOmahaPlayer_EvalBestHand_StraightFlush(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignSpade, 6, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))
	p.AddCard(NewCard(CardDesignClover, 3, false))

	community := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandStraightFlush, rank)
}

func TestOmahaPlayer_EvalBestHand_RoyalFlush(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignSpade, 13, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))
	p.AddCard(NewCard(CardDesignClover, 3, false))

	community := []*Card{
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignClover, 5, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandRoyalFlush, rank)
}

func TestOmahaPlayer_EvalBestHand_LessThan2HoleCards(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))

	community := []*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignClover, 12, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandHighCard, rank)
	assert.Nil(t, p.GetBestHand())
}

func TestOmahaPlayer_EvalBestHand_LessThan3CommunityCards(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))
	p.AddCard(NewCard(CardDesignClover, 3, false))
	p.AddCard(NewCard(CardDesignDiamond, 4, false))

	community := []*Card{
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignDiamond, 6, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandHighCard, rank)
	assert.Nil(t, p.GetBestHand())
}

func TestOmahaPlayer_EvalBestHand_Wheel(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))
	p.AddCard(NewCard(CardDesignClover, 8, false))
	p.AddCard(NewCard(CardDesignDiamond, 9, false))

	community := []*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignClover, 12, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandStraight, rank)
}

func TestOmahaPlayer_EvalBestHand_BoardFlushNotUsable(t *testing.T) {
	t.Run("flush on board but no suited hole cards", func(t *testing.T) {
		p := NewOmahaPlayer(true, HoldemStyleTAG)
		// No spade hole cards
		p.AddCard(NewCard(CardDesignHeart, 2, false))
		p.AddCard(NewCard(CardDesignHeart, 3, false))
		p.AddCard(NewCard(CardDesignClover, 4, false))
		p.AddCard(NewCard(CardDesignDiamond, 6, false))

		// 5 spades on community
		community := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignSpade, 11, false),
		}

		rank := p.EvalBestHand(community)
		// Must use 2 hole cards → no flush possible
		assert.Less(t, rank, PokerHandFlush)
	})
}

func TestOmahaPlayer_HUDStats(t *testing.T) {
	p := NewOmahaPlayer(false, HoldemStyleTAG)

	t.Run("initial stats are zero", func(t *testing.T) {
		assert.Equal(t, 0, p.GetTotalHands())
		assert.Equal(t, 0, p.GetVPIPCount())
		assert.Equal(t, 0, p.GetPFRCount())
		assert.Equal(t, 0, p.GetVPIP())
		assert.Equal(t, 0, p.GetPFR())
	})

	t.Run("VPIP and PFR with totalHands", func(t *testing.T) {
		p.IncrementTotalHands()
		p.IncrementTotalHands()
		p.IncrementTotalHands()
		p.IncrementTotalHands() // totalHands=4
		p.IncrementVPIP()
		p.IncrementVPIP()
		p.IncrementVPIP() // vpipCount=3
		p.IncrementPFR()  // pfrCount=1

		assert.Equal(t, 4, p.GetTotalHands())
		assert.Equal(t, 3, p.GetVPIPCount())
		assert.Equal(t, 1, p.GetPFRCount())
		assert.Equal(t, 75, p.GetVPIP()) // 3*100/4=75
		assert.Equal(t, 25, p.GetPFR())  // 1*100/4=25
	})
}

func TestOmahaPlayer_ThreeBetStats(t *testing.T) {
	t.Run("initial stats are zero", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		assert.Equal(t, 0, p.GetThreeBetOpportunity())
		assert.Equal(t, 0, p.GetThreeBetCount())
		assert.Equal(t, 0, p.GetThreeBet())
	})

	t.Run("3Bet percentage with opportunities", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.IncrementThreeBetOpportunity()
		p.IncrementThreeBetOpportunity()
		p.IncrementThreeBetOpportunity()
		p.IncrementThreeBetOpportunity() // 4 opportunities
		p.IncrementThreeBet()            // 1 3bet
		assert.Equal(t, 4, p.GetThreeBetOpportunity())
		assert.Equal(t, 1, p.GetThreeBetCount())
		assert.Equal(t, 25, p.GetThreeBet()) // 1*100/4=25
	})

	t.Run("3Bet percentage zero when no opportunities", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		assert.Equal(t, 0, p.GetThreeBet())
	})
}

func TestOmahaPlayer_AFStats(t *testing.T) {
	t.Run("initial AF display is dash", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		assert.Equal(t, 0, p.GetPostFlopBetRaise())
		assert.Equal(t, 0, p.GetPostFlopCall())
		assert.Equal(t, "-", p.GetAFDisplay())
	})

	t.Run("AF infinity when bets but no calls", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		assert.Equal(t, 2, p.GetPostFlopBetRaise())
		assert.Equal(t, 0, p.GetPostFlopCall())
		assert.Equal(t, "∞", p.GetAFDisplay())
	})

	t.Run("AF normal ratio", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise() // 5
		p.IncrementPostFlopCall()
		p.IncrementPostFlopCall() // 2
		assert.Equal(t, "2.5", p.GetAFDisplay())
	})

	t.Run("AF integer ratio", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise() // 4
		p.IncrementPostFlopCall()
		p.IncrementPostFlopCall() // 2
		assert.Equal(t, "2.0", p.GetAFDisplay())
	})

	t.Run("AF dash when no postflop actions", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		assert.Equal(t, "-", p.GetAFDisplay())
	})

	t.Run("AF zero when only calls", func(t *testing.T) {
		p := NewOmahaPlayer(false, HoldemStyleTAG)
		p.IncrementPostFlopCall()
		p.IncrementPostFlopCall()
		assert.Equal(t, "0.0", p.GetAFDisplay())
	})
}

func TestOmahaPlayer_GetComparisonCards(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 13, false))
	p.AddCard(NewCard(CardDesignClover, 12, false))
	p.AddCard(NewCard(CardDesignDiamond, 11, false))

	community := []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
	}

	p.EvalBestHand(community)
	cards := p.GetComparisonCards()
	assert.Equal(t, 5, len(cards))
	// Verify it's a copy
	cards[0] = nil
	assert.NotNil(t, p.GetBestHand()[0])
}

func TestOmahaPlayer_SetBestHand(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	hand := []*Card{NewCard(CardDesignSpade, 1, false)}
	p.SetBestHand(hand)
	assert.Equal(t, hand, p.GetBestHand())
}
