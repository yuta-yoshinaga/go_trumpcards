package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// NewPokerPlayer
// ---------------------------------------------------------------------------

func TestNewPokerPlayer(t *testing.T) {
	t.Run("human player with Conservative style", func(t *testing.T) {
		p := NewPokerPlayer(true, PokerStyleConservative)
		assert.NotNil(t, p)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, PokerStyleConservative, p.GetPlayStyle())
		assert.Equal(t, 0, p.GetCardsSize())
		assert.Equal(t, 0, p.GetChips())
		assert.Equal(t, 0, p.GetHandRank())
		assert.False(t, p.GetFolded())
		assert.False(t, p.GetAllIn())
		assert.Equal(t, 0, p.GetCurrentBet())
		assert.Equal(t, 0, p.GetExchangeCount())
	})

	t.Run("CPU player with Bluffer style", func(t *testing.T) {
		p := NewPokerPlayer(false, PokerStyleBluffer)
		assert.False(t, p.GetIsHuman())
		assert.Equal(t, PokerStyleBluffer, p.GetPlayStyle())
	})
}

// ---------------------------------------------------------------------------
// AddCard
// ---------------------------------------------------------------------------

func TestPokerPlayer_AddCard(t *testing.T) {
	p := NewPokerPlayer(true, PokerStyleBalanced)
	assert.Equal(t, 0, p.GetCardsSize())

	p.AddCard(NewCard(CardDesignSpade, 1, false))
	assert.Equal(t, 1, p.GetCardsSize())

	p.AddCard(NewCard(CardDesignHeart, 10, false))
	assert.Equal(t, 2, p.GetCardsSize())
}

// ---------------------------------------------------------------------------
// ExchangeCard
// ---------------------------------------------------------------------------

func TestPokerPlayer_ExchangeCard(t *testing.T) {
	setup := func() *PokerPlayer {
		p := NewPokerPlayer(true, PokerStyleBalanced)
		p.AddCard(NewCard(CardDesignSpade, 2, false))
		p.AddCard(NewCard(CardDesignClover, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 7, false))
		return p
	}

	t.Run("valid index replaces card", func(t *testing.T) {
		p := setup()
		replacement := NewCard(CardDesignDiamond, 13, false)
		p.ExchangeCard(1, replacement)
		assert.Equal(t, 13, p.GetCard(1).GetValue())
		assert.Equal(t, CardDesignDiamond, p.GetCard(1).GetDesign())
	})

	t.Run("negative index does nothing", func(t *testing.T) {
		p := setup()
		p.ExchangeCard(-1, NewCard(CardDesignDiamond, 13, false))
		// cards unchanged
		assert.Equal(t, 2, p.GetCard(0).GetValue())
		assert.Equal(t, 5, p.GetCard(1).GetValue())
		assert.Equal(t, 7, p.GetCard(2).GetValue())
	})

	t.Run("index equal to len does nothing", func(t *testing.T) {
		p := setup()
		p.ExchangeCard(3, NewCard(CardDesignDiamond, 13, false))
		assert.Equal(t, 3, p.GetCardsSize())
		assert.Equal(t, 2, p.GetCard(0).GetValue())
	})

	t.Run("index greater than len does nothing", func(t *testing.T) {
		p := setup()
		p.ExchangeCard(100, NewCard(CardDesignDiamond, 13, false))
		assert.Equal(t, 3, p.GetCardsSize())
	})
}

// ---------------------------------------------------------------------------
// EvalHand
// ---------------------------------------------------------------------------

func TestPokerPlayer_EvalHand(t *testing.T) {
	t.Run("evaluates hand and stores rank", func(t *testing.T) {
		p := NewPokerPlayer(true, PokerStyleBalanced)
		// One Pair of 5s
		p.AddCard(NewCard(CardDesignSpade, 5, false))
		p.AddCard(NewCard(CardDesignClover, 5, false))
		p.AddCard(NewCard(CardDesignHeart, 7, false))
		p.AddCard(NewCard(CardDesignDiamond, 9, false))
		p.AddCard(NewCard(CardDesignSpade, 11, false))

		rank := p.EvalHand()
		assert.Equal(t, PokerHandOnePair, rank)
		assert.Equal(t, PokerHandOnePair, p.GetHandRank())
	})
}

// ---------------------------------------------------------------------------
// GetHandRank / SetHandRank
// ---------------------------------------------------------------------------

func TestPokerPlayer_HandRank(t *testing.T) {
	p := NewPokerPlayer(true, PokerStyleBalanced)
	assert.Equal(t, PokerHandHighCard, p.GetHandRank())

	p.SetHandRank(PokerHandRoyalFlush)
	assert.Equal(t, PokerHandRoyalFlush, p.GetHandRank())
}

// ---------------------------------------------------------------------------
// GetHandName
// ---------------------------------------------------------------------------

func TestPokerPlayer_GetHandName(t *testing.T) {
	p := NewPokerPlayer(true, PokerStyleBalanced)

	tests := []struct {
		rank int
		name string
	}{
		{PokerHandHighCard, "High Card"},
		{PokerHandOnePair, "One Pair"},
		{PokerHandTwoPair, "Two Pair"},
		{PokerHandThreeOfAKind, "Three of a Kind"},
		{PokerHandStraight, "Straight"},
		{PokerHandFlush, "Flush"},
		{PokerHandFullHouse, "Full House"},
		{PokerHandFourOfAKind, "Four of a Kind"},
		{PokerHandStraightFlush, "Straight Flush"},
		{PokerHandRoyalFlush, "Royal Flush"},
		{PokerHandFiveOfAKind, "Five of a Kind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.SetHandRank(tt.rank)
			assert.Equal(t, tt.name, p.GetHandName())
		})
	}

	t.Run("negative rank returns Unknown", func(t *testing.T) {
		p.SetHandRank(-1)
		assert.Equal(t, "Unknown", p.GetHandName())
	})

	t.Run("rank beyond array returns Unknown", func(t *testing.T) {
		p.SetHandRank(len(PokerHandNames))
		assert.Equal(t, "Unknown", p.GetHandName())
	})

	t.Run("large rank returns Unknown", func(t *testing.T) {
		p.SetHandRank(999)
		assert.Equal(t, "Unknown", p.GetHandName())
	})
}

// ---------------------------------------------------------------------------
// GetChips / SetChips / AddChips
// ---------------------------------------------------------------------------

func TestPokerPlayer_Chips(t *testing.T) {
	p := NewPokerPlayer(true, PokerStyleBalanced)
	assert.Equal(t, 0, p.GetChips())

	p.SetChips(500)
	assert.Equal(t, 500, p.GetChips())

	p.AddChips(200)
	assert.Equal(t, 700, p.GetChips())
}

// ---------------------------------------------------------------------------
// SubtractChips
// ---------------------------------------------------------------------------

func TestPokerPlayer_SubtractChips(t *testing.T) {
	t.Run("sufficient chips returns true", func(t *testing.T) {
		p := NewPokerPlayer(true, PokerStyleBalanced)
		p.SetChips(100)
		ok := p.SubtractChips(60)
		assert.True(t, ok)
		assert.Equal(t, 40, p.GetChips())
	})

	t.Run("exact chips returns true", func(t *testing.T) {
		p := NewPokerPlayer(true, PokerStyleBalanced)
		p.SetChips(100)
		ok := p.SubtractChips(100)
		assert.True(t, ok)
		assert.Equal(t, 0, p.GetChips())
	})

	t.Run("insufficient chips returns false and chips unchanged", func(t *testing.T) {
		p := NewPokerPlayer(true, PokerStyleBalanced)
		p.SetChips(10)
		ok := p.SubtractChips(50)
		assert.False(t, ok)
		assert.Equal(t, 10, p.GetChips())
	})
}

// ---------------------------------------------------------------------------
// GetIsHuman / GetFolded / SetFolded / GetAllIn / SetAllIn
// ---------------------------------------------------------------------------

func TestPokerPlayer_BoolFlags(t *testing.T) {
	p := NewPokerPlayer(false, PokerStyleAggressive)

	assert.False(t, p.GetIsHuman())
	assert.False(t, p.GetFolded())
	assert.False(t, p.GetAllIn())

	p.SetFolded(true)
	assert.True(t, p.GetFolded())
	p.SetFolded(false)
	assert.False(t, p.GetFolded())

	p.SetAllIn(true)
	assert.True(t, p.GetAllIn())
	p.SetAllIn(false)
	assert.False(t, p.GetAllIn())
}

// ---------------------------------------------------------------------------
// GetCurrentBet / SetCurrentBet
// ---------------------------------------------------------------------------

func TestPokerPlayer_CurrentBet(t *testing.T) {
	p := NewPokerPlayer(true, PokerStyleBalanced)
	assert.Equal(t, 0, p.GetCurrentBet())

	p.SetCurrentBet(50)
	assert.Equal(t, 50, p.GetCurrentBet())
}

// ---------------------------------------------------------------------------
// GetPlayStyle / GetPlayStyleName
// ---------------------------------------------------------------------------

func TestPokerPlayer_PlayStyle(t *testing.T) {
	tests := []struct {
		style PokerPlayStyle
		name  string
	}{
		{PokerStyleConservative, "Conservative"},
		{PokerStyleBalanced, "Balanced"},
		{PokerStyleAggressive, "Aggressive"},
		{PokerStyleBluffer, "Bluffer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPokerPlayer(false, tt.style)
			assert.Equal(t, tt.style, p.GetPlayStyle())
			assert.Equal(t, tt.name, p.GetPlayStyleName())
		})
	}

	t.Run("out-of-range style returns Unknown", func(t *testing.T) {
		p := NewPokerPlayer(false, PokerPlayStyle(99))
		assert.Equal(t, "Unknown", p.GetPlayStyleName())
	})
}

// ---------------------------------------------------------------------------
// GetExchangeCount / SetExchangeCount
// ---------------------------------------------------------------------------

func TestPokerPlayer_ExchangeCount(t *testing.T) {
	p := NewPokerPlayer(true, PokerStyleBalanced)
	assert.Equal(t, 0, p.GetExchangeCount())

	p.SetExchangeCount(3)
	assert.Equal(t, 3, p.GetExchangeCount())
}

// ---------------------------------------------------------------------------
// evalFiveCardHandWithJokers
// ---------------------------------------------------------------------------

func TestEvalFiveCardHandWithJokers(t *testing.T) {
	t.Run("not 5 cards returns HighCard", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignClover, 2, false),
		}
		assert.Equal(t, PokerHandHighCard, evalFiveCardHandWithJokers(cards))
	})

	t.Run("empty slice returns HighCard", func(t *testing.T) {
		assert.Equal(t, PokerHandHighCard, evalFiveCardHandWithJokers([]*Card{}))
	})

	t.Run("6 cards returns HighCard", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignDiamond, 4, false),
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 6, false),
		}
		assert.Equal(t, PokerHandHighCard, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers delegates to evalFiveCardHand - HighCard", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		}
		assert.Equal(t, PokerHandHighCard, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - OnePair", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		}
		assert.Equal(t, PokerHandOnePair, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - TwoPair", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 9, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		}
		assert.Equal(t, PokerHandTwoPair, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - ThreeOfAKind", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		}
		assert.Equal(t, PokerHandThreeOfAKind, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - Straight", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 8, false),
			NewCard(CardDesignSpade, 9, false),
		}
		assert.Equal(t, PokerHandStraight, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - Straight A-2-3-4-5", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignDiamond, 4, false),
			NewCard(CardDesignSpade, 5, false),
		}
		assert.Equal(t, PokerHandStraight, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - Straight A-10-J-Q-K mixed suit", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignClover, 10, false),
			NewCard(CardDesignHeart, 11, false),
			NewCard(CardDesignDiamond, 12, false),
			NewCard(CardDesignSpade, 13, false),
		}
		assert.Equal(t, PokerHandStraight, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - Flush", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignSpade, 11, false),
		}
		assert.Equal(t, PokerHandFlush, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - FullHouse", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignDiamond, 3, false),
			NewCard(CardDesignSpade, 3, false),
		}
		assert.Equal(t, PokerHandFullHouse, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - FourOfAKind", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignDiamond, 6, false),
			NewCard(CardDesignSpade, 9, false),
		}
		assert.Equal(t, PokerHandFourOfAKind, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - StraightFlush", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignSpade, 7, false),
		}
		assert.Equal(t, PokerHandStraightFlush, evalFiveCardHandWithJokers(cards))
	})

	t.Run("no jokers - RoyalFlush", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 12, false),
			NewCard(CardDesignSpade, 13, false),
		}
		assert.Equal(t, PokerHandRoyalFlush, evalFiveCardHandWithJokers(cards))
	})

	// ----- 1 joker tests -----

	t.Run("1 joker makes OnePair from HighCard", func(t *testing.T) {
		// Joker + 4 distinct non-sequential, non-suited cards => at least OnePair
		cards := []*Card{
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignDiamond, 11, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.True(t, rank >= PokerHandOnePair)
	})

	t.Run("1 joker makes ThreeOfAKind from OnePair", func(t *testing.T) {
		// Pair of 5s + joker => Three of a Kind (at least)
		cards := []*Card{
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignDiamond, 11, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.True(t, rank >= PokerHandThreeOfAKind)
	})

	t.Run("1 joker makes FourOfAKind from ThreeOfAKind", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 2, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.True(t, rank >= PokerHandFourOfAKind)
	})

	t.Run("1 joker makes FiveOfAKind from FourOfAKind", func(t *testing.T) {
		// 4 sixes + joker => FiveOfAKind
		cards := []*Card{
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignDiamond, 6, false),
			NewCard(CardDesignJoker, 1, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.Equal(t, PokerHandFiveOfAKind, rank)
	})

	t.Run("1 joker completes StraightFlush", func(t *testing.T) {
		// Spade 3,4,5,7 + joker => joker as Spade 6 => StraightFlush
		cards := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignSpade, 7, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.Equal(t, PokerHandStraightFlush, rank)
	})

	t.Run("1 joker completes RoyalFlush", func(t *testing.T) {
		// Spade A,10,J,K + joker => joker as Spade Q => RoyalFlush
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignSpade, 13, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.Equal(t, PokerHandRoyalFlush, rank)
	})

	t.Run("1 joker restores joker card after evaluation", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignDiamond, 11, false),
		}
		evalFiveCardHandWithJokers(cards)
		// After evaluation, the joker card should be restored
		assert.Equal(t, CardDesignJoker, cards[0].GetDesign())
	})

	t.Run("1 joker preserves original pointer identity after evaluation", func(t *testing.T) {
		jokerCard := NewCard(CardDesignJoker, 1, false)
		cards := []*Card{
			jokerCard,
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignDiamond, 11, false),
		}
		evalFiveCardHandWithJokers(cards)
		// The original pointer must be restored, not a newly created card
		assert.Same(t, jokerCard, cards[0])
	})

	// ----- 2 jokers tests -----

	t.Run("2 jokers improve hand", func(t *testing.T) {
		// 2 jokers + 3 distinct cards => at least ThreeOfAKind
		cards := []*Card{
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignJoker, 2, false),
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignHeart, 11, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.True(t, rank >= PokerHandThreeOfAKind)
	})

	t.Run("2 jokers make FiveOfAKind from ThreeOfAKind", func(t *testing.T) {
		// 3 eights + 2 jokers => FiveOfAKind
		cards := []*Card{
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignJoker, 2, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.Equal(t, PokerHandFiveOfAKind, rank)
	})

	t.Run("2 jokers complete RoyalFlush", func(t *testing.T) {
		// Spade A,10,11 + 2 jokers => jokers as Spade 12 and 13 => RoyalFlush
		cards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignJoker, 2, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.Equal(t, PokerHandRoyalFlush, rank)
	})

	t.Run("2 jokers restore joker cards after evaluation", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignJoker, 2, false),
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignHeart, 11, false),
		}
		evalFiveCardHandWithJokers(cards)
		assert.Equal(t, CardDesignJoker, cards[0].GetDesign())
		assert.Equal(t, CardDesignJoker, cards[1].GetDesign())
	})

	t.Run("2 jokers preserve original pointer identity after evaluation", func(t *testing.T) {
		jokerCard0 := NewCard(CardDesignJoker, 1, false)
		jokerCard1 := NewCard(CardDesignJoker, 2, false)
		cards := []*Card{
			jokerCard0,
			jokerCard1,
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignHeart, 11, false),
		}
		evalFiveCardHandWithJokers(cards)
		// The original pointers must be restored, not newly created cards
		assert.Same(t, jokerCard0, cards[0])
		assert.Same(t, jokerCard1, cards[1])
	})

	t.Run("2 jokers with pair does not trigger FiveOfAKind", func(t *testing.T) {
		// Pair of 5s + 2 jokers + other card: best = FourOfAKind (2+2=4, not 5)
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 9, false),
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignJoker, 2, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		// 2 jokers + pair => FourOfAKind, but not FiveOfAKind (count 2+2 = 4 < 5)
		assert.True(t, rank >= PokerHandFourOfAKind)
		assert.True(t, rank < PokerHandFiveOfAKind)
	})

	// ----- FiveOfAKind boundary: bestRank < FourOfAKind does not check FiveOfAKind -----

	t.Run("1 joker with low hand skips FiveOfAKind check", func(t *testing.T) {
		// Joker + 4 cards that form at best a Straight (rank < FourOfAKind)
		// Spade 5,6,7,8 + Joker => StraightFlush at best (joker = Spade 4 or 9)
		// But let's ensure no FiveOfAKind is falsely triggered
		cards := []*Card{
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 8, false),
		}
		rank := evalFiveCardHandWithJokers(cards)
		assert.True(t, rank < PokerHandFiveOfAKind)
	})
}

// ---------------------------------------------------------------------------
// checkFiveOfAKind
// ---------------------------------------------------------------------------

func TestCheckFiveOfAKind(t *testing.T) {
	t.Run("4 of same value + 1 joker returns true", func(t *testing.T) {
		nonJokers := []*Card{
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignDiamond, 6, false),
		}
		assert.True(t, checkFiveOfAKind(nonJokers, 1))
	})

	t.Run("3 of same value + 2 jokers returns true", func(t *testing.T) {
		nonJokers := []*Card{
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignHeart, 8, false),
		}
		assert.True(t, checkFiveOfAKind(nonJokers, 2))
	})

	t.Run("2 of same value + 1 joker returns false", func(t *testing.T) {
		nonJokers := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 9, false),
			NewCard(CardDesignDiamond, 11, false),
		}
		assert.False(t, checkFiveOfAKind(nonJokers, 1))
	})

	t.Run("no matching values returns false", func(t *testing.T) {
		nonJokers := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 8, false),
		}
		assert.False(t, checkFiveOfAKind(nonJokers, 1))
	})

	t.Run("pair + 2 jokers returns false (2+2=4 < 5)", func(t *testing.T) {
		nonJokers := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 9, false),
		}
		assert.False(t, checkFiveOfAKind(nonJokers, 2))
	})

	t.Run("empty nonJokers returns false", func(t *testing.T) {
		assert.False(t, checkFiveOfAKind([]*Card{}, 2))
	})
}

// ---------------------------------------------------------------------------
// Full hand evaluation via EvalHand (integration through PokerPlayer)
// ---------------------------------------------------------------------------

func TestPokerPlayer_EvalHand_AllRanks(t *testing.T) {
	makePlayer := func(cards ...*Card) *PokerPlayer {
		p := NewPokerPlayer(true, PokerStyleBalanced)
		for _, c := range cards {
			p.AddCard(c)
		}
		return p
	}

	t.Run("HighCard", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		)
		assert.Equal(t, PokerHandHighCard, p.EvalHand())
		assert.Equal(t, "High Card", p.GetHandName())
	})

	t.Run("OnePair", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		)
		assert.Equal(t, PokerHandOnePair, p.EvalHand())
		assert.Equal(t, "One Pair", p.GetHandName())
	})

	t.Run("TwoPair", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignHeart, 9, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		)
		assert.Equal(t, PokerHandTwoPair, p.EvalHand())
		assert.Equal(t, "Two Pair", p.GetHandName())
	})

	t.Run("ThreeOfAKind", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignClover, 7, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		)
		assert.Equal(t, PokerHandThreeOfAKind, p.EvalHand())
		assert.Equal(t, "Three of a Kind", p.GetHandName())
	})

	t.Run("Straight", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 8, false),
			NewCard(CardDesignSpade, 9, false),
		)
		assert.Equal(t, PokerHandStraight, p.EvalHand())
		assert.Equal(t, "Straight", p.GetHandName())
	})

	t.Run("Flush", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignHeart, 2, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignHeart, 9, false),
			NewCard(CardDesignHeart, 11, false),
		)
		assert.Equal(t, PokerHandFlush, p.EvalHand())
		assert.Equal(t, "Flush", p.GetHandName())
	})

	t.Run("FullHouse", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignDiamond, 3, false),
			NewCard(CardDesignSpade, 3, false),
		)
		assert.Equal(t, PokerHandFullHouse, p.EvalHand())
		assert.Equal(t, "Full House", p.GetHandName())
	})

	t.Run("FourOfAKind", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignDiamond, 6, false),
			NewCard(CardDesignSpade, 9, false),
		)
		assert.Equal(t, PokerHandFourOfAKind, p.EvalHand())
		assert.Equal(t, "Four of a Kind", p.GetHandName())
	})

	t.Run("StraightFlush", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignClover, 3, false),
			NewCard(CardDesignClover, 4, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignClover, 7, false),
		)
		assert.Equal(t, PokerHandStraightFlush, p.EvalHand())
		assert.Equal(t, "Straight Flush", p.GetHandName())
	})

	t.Run("RoyalFlush", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignDiamond, 1, false),
			NewCard(CardDesignDiamond, 10, false),
			NewCard(CardDesignDiamond, 11, false),
			NewCard(CardDesignDiamond, 12, false),
			NewCard(CardDesignDiamond, 13, false),
		)
		assert.Equal(t, PokerHandRoyalFlush, p.EvalHand())
		assert.Equal(t, "Royal Flush", p.GetHandName())
	})

	t.Run("FiveOfAKind with 1 joker", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignDiamond, 6, false),
			NewCard(CardDesignJoker, 1, false),
		)
		assert.Equal(t, PokerHandFiveOfAKind, p.EvalHand())
		assert.Equal(t, "Five of a Kind", p.GetHandName())
	})

	t.Run("FiveOfAKind with 2 jokers", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignJoker, 2, false),
		)
		assert.Equal(t, PokerHandFiveOfAKind, p.EvalHand())
		assert.Equal(t, "Five of a Kind", p.GetHandName())
	})

	t.Run("less than 5 cards returns HighCard", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignClover, 2, false),
		)
		assert.Equal(t, PokerHandHighCard, p.EvalHand())
	})

	t.Run("joker making RoyalFlush via EvalHand", func(t *testing.T) {
		p := makePlayer(
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignHeart, 10, false),
			NewCard(CardDesignJoker, 1, false),
			NewCard(CardDesignHeart, 12, false),
			NewCard(CardDesignHeart, 13, false),
		)
		assert.Equal(t, PokerHandRoyalFlush, p.EvalHand())
	})
}
