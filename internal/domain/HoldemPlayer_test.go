package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHoldemPlayer(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	assert.Equal(t, 0, p.GetChips())
	assert.False(t, p.GetFolded())
	assert.False(t, p.GetAllIn())
	assert.Equal(t, 0, p.GetCurrentBet())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestHoldemPlayer_ChipOperations(t *testing.T) {
	p := NewHoldemPlayer(false, HoldemStyleLAP)

	p.SetChips(1000)
	assert.Equal(t, 1000, p.GetChips())

	p.AddChips(500)
	assert.Equal(t, 1500, p.GetChips())

	ok := p.SubtractChips(200)
	assert.True(t, ok)
	assert.Equal(t, 1300, p.GetChips())

	ok = p.SubtractChips(2000)
	assert.False(t, ok)
	assert.Equal(t, 1300, p.GetChips())
}

func TestHoldemPlayer_SettersGetters(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)

	p.SetFolded(true)
	assert.True(t, p.GetFolded())

	p.SetAllIn(true)
	assert.True(t, p.GetAllIn())

	p.SetCurrentBet(100)
	assert.Equal(t, 100, p.GetCurrentBet())
}

func TestHoldemPlayer_GetPlayStyleName(t *testing.T) {
	tests := []struct {
		style HoldemPlayStyle
		name  string
	}{
		{HoldemStyleTAG, "TAG"},
		{HoldemStyleLAP, "LAP"},
		{HoldemStyleTAP, "TAP"},
		{HoldemStyleLAG, "LAG"},
		{HoldemPlayStyle(99), "Unknown"},
	}
	for _, tt := range tests {
		p := NewHoldemPlayer(false, tt.style)
		assert.Equal(t, tt.name, p.GetPlayStyleName())
	}
}

func TestHoldemPlayer_EvalBestHand_HighCard(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 2, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))

	community := []*Card{
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandHighCard, rank)
	assert.Equal(t, PokerHandHighCard, p.GetHandRank())
	assert.NotNil(t, p.GetBestHand())
	assert.Equal(t, 5, len(p.GetBestHand()))
}

func TestHoldemPlayer_EvalBestHand_OnePair(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))

	community := []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandOnePair, rank)
}

func TestHoldemPlayer_EvalBestHand_TwoPair(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))

	community := []*Card{
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandTwoPair, rank)
}

func TestHoldemPlayer_EvalBestHand_ThreeOfAKind(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))

	community := []*Card{
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 13, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandThreeOfAKind, rank)
}

func TestHoldemPlayer_EvalBestHand_Straight(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 6, false))
	p.AddCard(NewCard(CardDesignHeart, 7, false))

	community := []*Card{
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandStraight, rank)
}

func TestHoldemPlayer_EvalBestHand_Flush(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 2, false))
	p.AddCard(NewCard(CardDesignSpade, 5, false))

	community := []*Card{
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 7, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandFlush, rank)
}

func TestHoldemPlayer_EvalBestHand_FullHouse(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))

	community := []*Card{
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 7, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandFullHouse, rank)
}

func TestHoldemPlayer_EvalBestHand_FourOfAKind(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))

	community := []*Card{
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 7, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandFourOfAKind, rank)
}

func TestHoldemPlayer_EvalBestHand_StraightFlush(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignSpade, 6, false))

	community := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandStraightFlush, rank)
}

func TestHoldemPlayer_EvalBestHand_RoyalFlush(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignSpade, 13, false))

	community := []*Card{
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandRoyalFlush, rank)
}

func TestHoldemPlayer_EvalBestHand_LessThan5Cards(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))

	community := []*Card{
		NewCard(CardDesignClover, 3, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandHighCard, rank)
	assert.Nil(t, p.GetBestHand())
}

func TestCombinations(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
	}

	// C(4,2) = 6
	combos := combinations(cards, 2)
	assert.Equal(t, 6, len(combos))

	// C(4,5) = 0 (k > n)
	combos = combinations(cards, 5)
	assert.Equal(t, 0, len(combos))

	// C(7,5) = 21
	cards7 := make([]*Card, 7)
	for i := 0; i < 7; i++ {
		cards7[i] = NewCard(CardDesignSpade, i+1, false)
	}
	combos = combinations(cards7, 5)
	assert.Equal(t, 21, len(combos))
}

func TestCompareHighCardsSlice(t *testing.T) {
	a := []*Card{
		NewCard(CardDesignSpade, 1, false),  // 14
		NewCard(CardDesignHeart, 13, false), // 13
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 3, false),
	}
	b := []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 3, false),
	}

	// a has Ace (14), b has King (13), so a > b
	assert.Equal(t, 1, compareHighCardsSlice(a, b))
	assert.Equal(t, -1, compareHighCardsSlice(b, a))

	// Equal hands
	assert.Equal(t, 0, compareHighCardsSlice(a, a))

	// Empty slices
	assert.Equal(t, 0, compareHighCardsSlice(nil, a))
	assert.Equal(t, 0, compareHighCardsSlice(a, nil))
	assert.Equal(t, 0, compareHighCardsSlice(nil, nil))

	// Wheel (A-2-3-4-5) should lose to 6-high straight (2-3-4-5-6)
	wheel := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
	}
	sixHigh := []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 6, false),
	}
	assert.Equal(t, -1, compareHighCardsSlice(wheel, sixHigh))
	assert.Equal(t, 1, compareHighCardsSlice(sixHigh, wheel))
	assert.Equal(t, 0, compareHighCardsSlice(wheel, wheel))
}

func TestCompareHighCardsSlice_OnePairTieBreak(t *testing.T) {
	// Pair of 4s with low kickers vs Pair of 3s with high kickers
	// Pair value must be compared first
	a := []*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignSpade, 2, false),
	}
	b := []*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignClover, 11, false),
	}
	// Pair of 4s > Pair of 3s
	assert.Equal(t, 1, compareHighCardsSlice(a, b))
	assert.Equal(t, -1, compareHighCardsSlice(b, a))
}

func TestCompareHighCardsSlice_TwoPairTieBreak(t *testing.T) {
	// 10-10-5-5-A vs 9-9-8-8-A
	a := []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 1, false),
	}
	b := []*Card{
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignClover, 1, false),
	}
	// Top pair: 10 > 9
	assert.Equal(t, 1, compareHighCardsSlice(a, b))
	assert.Equal(t, -1, compareHighCardsSlice(b, a))
}

func TestCompareHighCardsSlice_FullHouseTieBreak(t *testing.T) {
	// 3-3-3-K-K vs 2-2-2-A-A
	// Trips value (3 > 2) should decide, not the pair
	a := []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignSpade, 13, false),
	}
	b := []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 1, false),
	}
	// Trips: 3 > 2
	assert.Equal(t, 1, compareHighCardsSlice(a, b))
	assert.Equal(t, -1, compareHighCardsSlice(b, a))
}

func TestIsWheelHand(t *testing.T) {
	wheel := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
	}
	assert.True(t, isWheelHand(wheel))

	notWheel := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 6, false),
	}
	assert.False(t, isWheelHand(notWheel))

	// Not 5 cards
	assert.False(t, isWheelHand([]*Card{NewCard(CardDesignSpade, 1, false)}))
}

func TestHoldemPlayer_EvalBestHand_Wheel(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))

	community := []*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignClover, 12, false),
	}

	rank := p.EvalBestHand(community)
	assert.Equal(t, PokerHandStraight, rank)
}

func TestHoldemPlayer_EvalBestHand_ChoosesBestFromSeven(t *testing.T) {
	p := NewHoldemPlayer(true, HoldemStyleTAG)
	// hole: pair of Aces
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 1, false))

	// community: 3 more Aces impossible, but pair of Kings + junk
	community := []*Card{
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 9, false),
	}

	rank := p.EvalBestHand(community)
	// AA + KK = Two Pair
	assert.Equal(t, PokerHandTwoPair, rank)
}

