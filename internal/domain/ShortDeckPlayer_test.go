package domain_test

import (
	"testing"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewShortDeckPlayer(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, domain.HoldemStyleTAG, p.GetPlayStyle())
	assert.Equal(t, 0, p.GetChips())
	assert.False(t, p.GetFolded())
	assert.False(t, p.GetAllIn())
	assert.Equal(t, 0, p.GetCurrentBet())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestShortDeckPlayer_GetPlayStyleName(t *testing.T) {
	tests := []struct {
		style domain.HoldemPlayStyle
		name  string
	}{
		{domain.HoldemStyleTAG, "TAG"},
		{domain.HoldemStyleLAP, "LAP"},
		{domain.HoldemStyleTAP, "TAP"},
		{domain.HoldemStyleLAG, "LAG"},
		{domain.HoldemStyleGTO, "GTO"},
		{domain.HoldemPlayStyle(99), "Unknown"},
	}
	for _, tt := range tests {
		p := domain.NewShortDeckPlayer(false, tt.style)
		assert.Equal(t, tt.name, p.GetPlayStyleName())
	}
}

func TestShortDeckPlayer_EvalBestHand_HighCard(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandHighCard, rank)
	assert.Equal(t, domain.ShortDeckHandHighCard, p.GetHandRank())
	assert.NotNil(t, p.GetBestHand())
	assert.Equal(t, 5, len(p.GetBestHand()))
}

func TestShortDeckPlayer_EvalBestHand_OnePair(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandOnePair, rank)
}

func TestShortDeckPlayer_EvalBestHand_TwoPair(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandTwoPair, rank)
}

func TestShortDeckPlayer_EvalBestHand_ThreeOfAKind(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandThreeOfAKind, rank)
}

func TestShortDeckPlayer_EvalBestHand_Straight(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandStraight, rank)
}

func TestShortDeckPlayer_EvalBestHand_FullHouse(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandFullHouse, rank)
}

func TestShortDeckPlayer_EvalBestHand_Flush(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandFlush, rank)
}

func TestShortDeckPlayer_EvalBestHand_FourOfAKind(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 10, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandFourOfAKind, rank)
}

func TestShortDeckPlayer_EvalBestHand_StraightFlush(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandStraightFlush, rank)
}

func TestShortDeckPlayer_EvalBestHand_RoyalFlush(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandRoyalFlush, rank)
}

func TestShortDeckPlayer_EvalBestHand_LessThan5Cards(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 7, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandHighCard, rank)
	assert.Nil(t, p.GetBestHand())
}

func TestShortDeckPlayer_EvalBestHand_FlushBeatsFullHouse(t *testing.T) {
	// In ShortDeck, Flush (rank 6) > FullHouse (rank 5)
	flushPlayer := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	flushPlayer.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	flushPlayer.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
	}

	flushRank := flushPlayer.EvalBestHand(community)
	assert.Equal(t, domain.ShortDeckHandFlush, flushRank)

	fullHousePlayer := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	fullHousePlayer.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	fullHousePlayer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))

	community2 := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}

	fullHouseRank := fullHousePlayer.EvalBestHand(community2)
	assert.Equal(t, domain.ShortDeckHandFullHouse, fullHouseRank)

	// Flush rank (6) must be higher than FullHouse rank (5) in ShortDeck
	assert.Greater(t, flushRank, fullHouseRank)
}

func TestShortDeckPlayer_EvalBestHand_ChoosesBestFromSeven(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	// hole: pair of Aces
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

	// community: pair of Kings + junk
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
	}

	rank := p.EvalBestHand(community)
	// AA + KK = Two Pair
	assert.Equal(t, domain.ShortDeckHandTwoPair, rank)
}

func TestShortDeckPlayer_HUDStats(t *testing.T) {
	p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)

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

func TestShortDeckPlayer_ThreeBetStats(t *testing.T) {
	t.Run("initial stats are zero", func(t *testing.T) {
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		assert.Equal(t, 0, p.GetThreeBetOpportunity())
		assert.Equal(t, 0, p.GetThreeBetCount())
		assert.Equal(t, 0, p.GetThreeBet())
	})

	t.Run("3Bet percentage with opportunities", func(t *testing.T) {
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
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
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		assert.Equal(t, 0, p.GetThreeBet())
	})
}

func TestShortDeckPlayer_AFStats(t *testing.T) {
	t.Run("initial AF display is dash", func(t *testing.T) {
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		assert.Equal(t, 0, p.GetPostFlopBetRaise())
		assert.Equal(t, 0, p.GetPostFlopCall())
		assert.Equal(t, "-", p.GetAFDisplay())
	})

	t.Run("AF infinity when bets but no calls", func(t *testing.T) {
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		assert.Equal(t, 2, p.GetPostFlopBetRaise())
		assert.Equal(t, 0, p.GetPostFlopCall())
		assert.Equal(t, "∞", p.GetAFDisplay())
	})

	t.Run("AF normal ratio", func(t *testing.T) {
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
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
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise() // 4
		p.IncrementPostFlopCall()
		p.IncrementPostFlopCall() // 2
		assert.Equal(t, "2.0", p.GetAFDisplay())
	})

	t.Run("AF dash when no postflop actions", func(t *testing.T) {
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		assert.Equal(t, "-", p.GetAFDisplay())
	})

	t.Run("AF zero when only calls", func(t *testing.T) {
		p := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		p.IncrementPostFlopCall()
		p.IncrementPostFlopCall()
		assert.Equal(t, "0.0", p.GetAFDisplay())
	})
}

func TestShortDeckPlayer_SetBestHand(t *testing.T) {
	p := domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	hand := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	p.SetBestHand(hand)
	assert.Equal(t, hand, p.GetBestHand())
}
