package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDramahaPlayer(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	assert.Equal(t, 0, p.GetChips())
	assert.False(t, p.GetFolded())
	assert.False(t, p.GetAllIn())
	assert.Equal(t, 0, p.GetCurrentBet())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestDramahaPlayer_GetPlayStyleName(t *testing.T) {
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
		p := NewDramahaPlayer(false, tt.style)
		assert.Equal(t, tt.name, p.GetPlayStyleName())
	}
}

func TestDramahaPlayer_EvalBestHand_DramahaRule(t *testing.T) {
	t.Run("must use exactly 2 hole + 3 community", func(t *testing.T) {
		p := NewDramahaPlayer(true, HoldemStyleTAG)
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
		// In Dramaha, must use exactly 2 hole cards (both are 4s) + 3 community
		// Best hand is 4-4-A-K-Q → OnePair, NOT four-of-a-kind or straight flush
		assert.Equal(t, PokerHandOnePair, rank)
		assert.NotNil(t, p.GetBestHand())
		assert.Equal(t, 5, len(p.GetBestHand()))
	})
}

func TestDramahaPlayer_EvalBestHand_HighCard(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_OnePair(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_TwoPair(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_ThreeOfAKind(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_Straight(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_Flush(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_FullHouse(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_FourOfAKind(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_StraightFlush(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_RoyalFlush(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_LessThan2HoleCards(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_LessThan3CommunityCards(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_Wheel(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_EvalBestHand_BoardFlushNotUsable(t *testing.T) {
	t.Run("flush on board but no suited hole cards", func(t *testing.T) {
		p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_HUDStats(t *testing.T) {
	p := NewDramahaPlayer(false, HoldemStyleTAG)

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

func TestDramahaPlayer_ThreeBetStats(t *testing.T) {
	t.Run("initial stats are zero", func(t *testing.T) {
		p := NewDramahaPlayer(false, HoldemStyleTAG)
		assert.Equal(t, 0, p.GetThreeBetOpportunity())
		assert.Equal(t, 0, p.GetThreeBetCount())
		assert.Equal(t, 0, p.GetThreeBet())
	})

	t.Run("3Bet percentage with opportunities", func(t *testing.T) {
		p := NewDramahaPlayer(false, HoldemStyleTAG)
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
		p := NewDramahaPlayer(false, HoldemStyleTAG)
		assert.Equal(t, 0, p.GetThreeBet())
	})
}

func TestDramahaPlayer_AFStats(t *testing.T) {
	t.Run("initial AF display is dash", func(t *testing.T) {
		p := NewDramahaPlayer(false, HoldemStyleTAG)
		assert.Equal(t, 0, p.GetPostFlopBetRaise())
		assert.Equal(t, 0, p.GetPostFlopCall())
		assert.Equal(t, "-", p.GetAFDisplay())
	})

	t.Run("AF infinity when bets but no calls", func(t *testing.T) {
		p := NewDramahaPlayer(false, HoldemStyleTAG)
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		assert.Equal(t, 2, p.GetPostFlopBetRaise())
		assert.Equal(t, 0, p.GetPostFlopCall())
		assert.Equal(t, "∞", p.GetAFDisplay())
	})

	t.Run("AF normal ratio", func(t *testing.T) {
		p := NewDramahaPlayer(false, HoldemStyleTAG)
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
		p := NewDramahaPlayer(false, HoldemStyleTAG)
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise()
		p.IncrementPostFlopBetRaise() // 4
		p.IncrementPostFlopCall()
		p.IncrementPostFlopCall() // 2
		assert.Equal(t, "2.0", p.GetAFDisplay())
	})

	t.Run("AF dash when no postflop actions", func(t *testing.T) {
		p := NewDramahaPlayer(false, HoldemStyleTAG)
		assert.Equal(t, "-", p.GetAFDisplay())
	})

	t.Run("AF zero when only calls", func(t *testing.T) {
		p := NewDramahaPlayer(false, HoldemStyleTAG)
		p.IncrementPostFlopCall()
		p.IncrementPostFlopCall()
		assert.Equal(t, "0.0", p.GetAFDisplay())
	})
}

func TestDramahaPlayer_GetComparisonCards(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
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

func TestDramahaPlayer_SetBestHand(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
	hand := []*Card{NewCard(CardDesignSpade, 1, false)}
	p.SetBestHand(hand)
	assert.Equal(t, hand, p.GetBestHand())
}

// ---------------------------------------------------------------------------
// Draw side -- the five hole cards evaluated as they are. This is the half of
// the pot the clone (Omaha Hi-Lo, whose EvalBestLowHand this replaced) had no
// equivalent for: there is no qualifier, and the board is never consulted.
// ---------------------------------------------------------------------------

// TestDramahaPlayer_EvalDrawHand_IgnoresTheBoard is the test that separates the
// two evaluations. The same five cards rank as a flush on their own, while the
// Omaha rule (exactly 2 hole + exactly 3 board) can only reach high card on
// this board -- so the two sides of the pot must disagree.
func TestDramahaPlayer_EvalDrawHand_IgnoresTheBoard(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
	for _, c := range []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 9, false),
	} {
		p.AddCard(c)
	}
	// A board that pairs nothing in the hole and offers only one heart, so the
	// hole's flush is unreachable through the Omaha rule.
	board := []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 7, false),
	}

	omahaRank := p.EvalBestHand(board)
	drawRank := p.EvalDrawHand()

	assert.Equal(t, PokerHandHighCard, omahaRank, "the Omaha side can only play the board's high cards")
	assert.Equal(t, PokerHandFlush, drawRank, "the draw side reads the five hearts as a flush")
	assert.NotEqual(t, omahaRank, drawRank, "the two sides of the pot must be evaluated separately")

	assert.Equal(t, drawRank, p.GetDrawRank())
	assert.Len(t, p.GetDrawBestHand(), DramahaHoleCards)
	assert.Len(t, p.GetBestHand(), 5)

	// The draw hand is the hole, card for card; the Omaha hand is not.
	for i, c := range p.GetDrawBestHand() {
		assert.Same(t, p.GetCard(i), c, "draw card %d must be hole card %d", i, i)
	}
	usesBoard := false
	for _, c := range p.GetBestHand() {
		for _, b := range board {
			if c.GetDesign() == b.GetDesign() && c.GetValue() == b.GetValue() {
				usesBoard = true
			}
		}
	}
	assert.True(t, usesBoard, "the Omaha hand has to take three cards off the board")
}

// TestDramahaPlayer_EvalDrawHand_IsIndependentOfTheBoardEntirely varies the
// board while holding the hole fixed: the draw rank must not move.
func TestDramahaPlayer_EvalDrawHand_IsIndependentOfTheBoardEntirely(t *testing.T) {
	hole := []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 9, false),
	}
	boards := [][]*Card{
		nil,
		{NewCard(CardDesignSpade, 13, false), NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 11, false)},
		{
			NewCard(CardDesignSpade, 1, false), NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 10, false),
		},
	}
	for i, board := range boards {
		p := NewDramahaPlayer(true, HoldemStyleTAG)
		for _, c := range hole {
			p.AddCard(NewCard(c.GetDesign(), c.GetValue(), false))
		}
		p.EvalBestHand(board)
		assert.Equal(t, PokerHandFlush, p.EvalDrawHand(), "board %d must not change the draw hand", i)
	}
}

func TestDramahaPlayer_EvalDrawHand_RequiresExactlyFiveCards(t *testing.T) {
	for _, n := range []int{0, 1, 4, 6} {
		p := NewDramahaPlayer(true, HoldemStyleTAG)
		for i := 0; i < n; i++ {
			p.AddCard(NewCard(CardDesignHeart, i+2, false))
		}
		assert.Equal(t, PokerHandHighCard, p.EvalDrawHand(), "%d cards is not a draw hand", n)
		assert.Nil(t, p.GetDrawBestHand(), "%d cards must leave no draw hand behind", n)

		rank, best := p.PeekDrawHand()
		assert.Equal(t, PokerHandHighCard, rank, "%d cards", n)
		assert.Nil(t, best, "%d cards", n)
	}
}

func TestDramahaPlayer_EvalDrawHand_ClearsAStaleResult(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
	for _, c := range []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 9, false),
	} {
		p.AddCard(c)
	}
	assert.Equal(t, PokerHandFlush, p.EvalDrawHand())

	p.RemoveCard(0)

	assert.Equal(t, PokerHandHighCard, p.EvalDrawHand(),
		"a hand that is no longer five cards must not keep its old rank")
	assert.Nil(t, p.GetDrawBestHand())
}

func TestDramahaPlayer_PeekDrawHand_LeavesTheStateAlone(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
	for _, c := range []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 9, false),
	} {
		p.AddCard(c)
	}

	rank, best := p.PeekDrawHand()

	assert.Equal(t, PokerHandFlush, rank, "peeking still reports the real rank")
	assert.Len(t, best, DramahaHoleCards)
	assert.Equal(t, PokerHandHighCard, p.GetDrawRank(), "peeking must not write the rank")
	assert.Nil(t, p.GetDrawBestHand(), "peeking must not write the hand")
}

func TestDramahaPlayer_GetLowComparisonCards_ReturnsACopyOfTheDrawHand(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
	for _, c := range []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 9, false),
	} {
		p.AddCard(c)
	}
	p.EvalDrawHand()

	cards := p.GetLowComparisonCards()
	assert.Len(t, cards, DramahaHoleCards)

	cards[0] = nil
	assert.NotNil(t, p.GetDrawBestHand()[0], "the caller must not be able to blank the stored hand")
}

func TestDramahaPlayer_ReplaceCard(t *testing.T) {
	newHand := func() *DramahaPlayer {
		p := NewDramahaPlayer(true, HoldemStyleTAG)
		for i := 0; i < DramahaHoleCards; i++ {
			p.AddCard(NewCard(CardDesignHeart, i+2, false))
		}
		return p
	}

	t.Run("replaces the named card and nothing else", func(t *testing.T) {
		p := newHand()
		assert.True(t, p.ReplaceCard(2, NewCard(CardDesignSpade, 1, false)))
		assert.Equal(t, DramahaHoleCards, p.GetCardsSize())
		assert.Equal(t, CardDesignSpade, p.GetCard(2).GetDesign())
		assert.Equal(t, 1, p.GetCard(2).GetValue())
		assert.Equal(t, 2, p.GetCard(0).GetValue())
		assert.Equal(t, 6, p.GetCard(4).GetValue())
	})

	t.Run("refuses an index outside the hand", func(t *testing.T) {
		for _, idx := range []int{-1, DramahaHoleCards, 99} {
			p := newHand()
			assert.False(t, p.ReplaceCard(idx, NewCard(CardDesignSpade, 1, false)), "index %d", idx)
			assert.Equal(t, DramahaHoleCards, p.GetCardsSize(), "index %d", idx)
			for i := 0; i < DramahaHoleCards; i++ {
				assert.Equal(t, CardDesignHeart, p.GetCard(i).GetDesign(), "index %d, card %d", idx, i)
			}
		}
	})

	t.Run("refuses a nil card", func(t *testing.T) {
		p := newHand()
		assert.False(t, p.ReplaceCard(0, nil))
		assert.NotNil(t, p.GetCard(0), "a nil card would make the draw hand unevaluable")
	})
}

func TestDramahaPlayer_HoleCardsCopy_IsADetachedCopy(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
	for i := 0; i < DramahaHoleCards; i++ {
		p.AddCard(NewCard(CardDesignHeart, i+2, false))
	}

	cards := p.HoleCardsCopy()
	assert.Len(t, cards, DramahaHoleCards)

	cards[0] = nil
	assert.NotNil(t, p.GetCard(0), "the copy must not alias the hand")
}

// TestDramahaPlayer_EvalBestHand_SearchesAllFiveHoleCards guards the widened
// hole. With four cards the search was C(4,2)=6 pairs; with five it is C(5,2)
// =10, and the winning pair here includes the fifth card -- a search that
// stopped at the first four would report high card.
func TestDramahaPlayer_EvalBestHand_SearchesAllFiveHoleCards(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleTAG)
	for _, c := range []*Card{
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignHeart, 13, false), // the only card that pairs the board
	} {
		p.AddCard(c)
	}
	board := []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 6, false),
	}

	assert.Equal(t, PokerHandOnePair, p.EvalBestHand(board),
		"the fifth hole card must be in the search")

	paired := false
	for _, c := range p.GetBestHand() {
		if c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
			paired = true
		}
	}
	assert.True(t, paired, "the best hand must actually contain the fifth hole card")
}
