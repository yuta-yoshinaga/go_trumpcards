//go:build test

package domain

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSevenCardStudPlayer(t *testing.T) {
	p := NewSevenCardStudPlayer(true, HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	assert.Equal(t, "TAG", p.GetPlayStyleName())
	assert.Empty(t, p.GetHoleCards())
	assert.Empty(t, p.GetDoorCards())
	assert.Nil(t, p.GetBestHand())
}

func TestSevenCardStudPlayer_CardManagement(t *testing.T) {
	p := NewSevenCardStudPlayer(false, HoldemStyleLAP)

	h1 := NewCard(CardDesignSpade, 10, true)
	h2 := NewCard(CardDesignHeart, 11, true)
	d1 := NewCard(CardDesignDiamond, 12, true)

	p.AddHoleCard(h1)
	p.AddHoleCard(h2)
	p.AddDoorCard(d1)

	assert.Len(t, p.GetHoleCards(), 2)
	assert.Len(t, p.GetDoorCards(), 1)
	assert.Len(t, p.GetAllCards(), 3)

	p.ClearCards()
	assert.Empty(t, p.GetHoleCards())
	assert.Empty(t, p.GetDoorCards())
	assert.Nil(t, p.GetBestHand())
}

func TestSevenCardStudPlayer_EvalBestHand(t *testing.T) {
	t.Run("7 cards - finds best 5", func(t *testing.T) {
		p := NewSevenCardStudPlayer(true, HoldemStyleTAG)
		// Royal flush in spades + 2 junk cards
		p.AddHoleCard(NewCard(CardDesignSpade, 1, true))  // A♠
		p.AddHoleCard(NewCard(CardDesignSpade, 13, true)) // K♠
		p.AddHoleCard(NewCard(CardDesignHeart, 2, true))  // 2♥ (junk)
		p.AddDoorCard(NewCard(CardDesignSpade, 12, true)) // Q♠
		p.AddDoorCard(NewCard(CardDesignSpade, 11, true)) // J♠
		p.AddDoorCard(NewCard(CardDesignSpade, 10, true)) // 10♠
		p.AddDoorCard(NewCard(CardDesignClover, 3, true)) // 3♣ (junk)

		rank := p.EvalBestHand()
		assert.Equal(t, PokerHandRoyalFlush, rank)
		assert.Len(t, p.GetBestHand(), 5)
	})

	t.Run("fewer than 5 cards", func(t *testing.T) {
		p := NewSevenCardStudPlayer(true, HoldemStyleTAG)
		p.AddHoleCard(NewCard(CardDesignSpade, 1, true))
		p.AddDoorCard(NewCard(CardDesignHeart, 2, true))

		rank := p.EvalBestHand()
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, p.GetBestHand())
	})

	t.Run("full house", func(t *testing.T) {
		p := NewSevenCardStudPlayer(true, HoldemStyleTAG)
		p.AddHoleCard(NewCard(CardDesignSpade, 8, true))
		p.AddHoleCard(NewCard(CardDesignHeart, 8, true))
		p.AddHoleCard(NewCard(CardDesignDiamond, 8, true))
		p.AddDoorCard(NewCard(CardDesignSpade, 5, true))
		p.AddDoorCard(NewCard(CardDesignHeart, 5, true))
		p.AddDoorCard(NewCard(CardDesignClover, 2, true))
		p.AddDoorCard(NewCard(CardDesignDiamond, 3, true))

		rank := p.EvalBestHand()
		assert.Equal(t, PokerHandFullHouse, rank)
	})
}

func TestSevenCardStudPlayer_EvalVisibleHand(t *testing.T) {
	tests := []struct {
		name     string
		doors    []*Card
		wantRank int
	}{
		{
			name:     "no door cards",
			doors:    nil,
			wantRank: PokerHandHighCard,
		},
		{
			name: "single card",
			doors: []*Card{
				NewCard(CardDesignSpade, 10, true),
			},
			wantRank: PokerHandHighCard,
		},
		{
			name: "pair",
			doors: []*Card{
				NewCard(CardDesignSpade, 10, true),
				NewCard(CardDesignHeart, 10, true),
			},
			wantRank: PokerHandOnePair,
		},
		{
			name: "two pair",
			doors: []*Card{
				NewCard(CardDesignSpade, 10, true),
				NewCard(CardDesignHeart, 10, true),
				NewCard(CardDesignSpade, 5, true),
				NewCard(CardDesignHeart, 5, true),
			},
			wantRank: PokerHandTwoPair,
		},
		{
			name: "three of a kind",
			doors: []*Card{
				NewCard(CardDesignSpade, 10, true),
				NewCard(CardDesignHeart, 10, true),
				NewCard(CardDesignDiamond, 10, true),
			},
			wantRank: PokerHandThreeOfAKind,
		},
		{
			name: "four of a kind",
			doors: []*Card{
				NewCard(CardDesignSpade, 10, true),
				NewCard(CardDesignHeart, 10, true),
				NewCard(CardDesignDiamond, 10, true),
				NewCard(CardDesignClover, 10, true),
			},
			wantRank: PokerHandFourOfAKind,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewSevenCardStudPlayer(true, HoldemStyleTAG)
			for _, c := range tt.doors {
				p.AddDoorCard(c)
			}
			assert.Equal(t, tt.wantRank, p.EvalVisibleHand())
		})
	}
}

func TestCompareVisibleHands(t *testing.T) {
	t.Run("pair beats high card", func(t *testing.T) {
		a := NewSevenCardStudPlayer(true, HoldemStyleTAG)
		a.AddDoorCard(NewCard(CardDesignSpade, 10, true))
		a.AddDoorCard(NewCard(CardDesignHeart, 10, true))

		b := NewSevenCardStudPlayer(false, HoldemStyleTAG)
		b.AddDoorCard(NewCard(CardDesignSpade, 1, true))
		b.AddDoorCard(NewCard(CardDesignHeart, 13, true))

		assert.Equal(t, 1, CompareVisibleHands(a, b))
		assert.Equal(t, -1, CompareVisibleHands(b, a))
	})

	t.Run("same rank - high card tiebreak", func(t *testing.T) {
		a := NewSevenCardStudPlayer(true, HoldemStyleTAG)
		a.AddDoorCard(NewCard(CardDesignSpade, 1, true)) // Ace

		b := NewSevenCardStudPlayer(false, HoldemStyleTAG)
		b.AddDoorCard(NewCard(CardDesignHeart, 13, true)) // King

		assert.Equal(t, 1, CompareVisibleHands(a, b)) // Ace (14) > King (13)
	})

	t.Run("equal hands", func(t *testing.T) {
		a := NewSevenCardStudPlayer(true, HoldemStyleTAG)
		a.AddDoorCard(NewCard(CardDesignSpade, 10, true))

		b := NewSevenCardStudPlayer(false, HoldemStyleTAG)
		b.AddDoorCard(NewCard(CardDesignHeart, 10, true))

		assert.Equal(t, 0, CompareVisibleHands(a, b))
	})
}

func TestSuitRank(t *testing.T) {
	assert.Equal(t, 4, SuitRank(CardDesignSpade))
	assert.Equal(t, 3, SuitRank(CardDesignHeart))
	assert.Equal(t, 2, SuitRank(CardDesignDiamond))
	assert.Equal(t, 1, SuitRank(CardDesignClover))
	assert.Equal(t, 0, SuitRank(CardDesignJoker))
}

func TestSevenCardStudPlayer_HUDStats(t *testing.T) {
	p := NewSevenCardStudPlayer(true, HoldemStyleTAG)

	// Initial state
	assert.Equal(t, 0, p.GetVPIP())
	assert.Equal(t, 0, p.GetPFR())
	assert.Equal(t, 0, p.GetThreeBet())
	assert.Equal(t, "-", p.GetAFDisplay())

	// After some hands
	p.IncrementTotalHands()
	p.IncrementTotalHands()
	p.IncrementVPIP()
	p.IncrementPFR()
	assert.Equal(t, 50, p.GetVPIP())
	assert.Equal(t, 50, p.GetPFR())

	// 3Bet
	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBet()
	assert.Equal(t, 50, p.GetThreeBet())

	// AF
	p.IncrementPostFlopBetRaise()
	assert.Equal(t, "∞", p.GetAFDisplay()) // bet-raise > 0, call = 0
	p.IncrementPostFlopCall()
	assert.Equal(t, "1.0", p.GetAFDisplay())
}

func TestSevenCardStudPlayer_GetComparisonCards(t *testing.T) {
	p := NewSevenCardStudPlayer(true, HoldemStyleTAG)
	// Set up a 7-card hand and evaluate
	p.AddHoleCard(NewCard(CardDesignSpade, 1, true))
	p.AddHoleCard(NewCard(CardDesignHeart, 1, true))
	p.AddHoleCard(NewCard(CardDesignDiamond, 5, true))
	p.AddDoorCard(NewCard(CardDesignSpade, 13, true))
	p.AddDoorCard(NewCard(CardDesignHeart, 13, true))
	p.AddDoorCard(NewCard(CardDesignClover, 7, true))
	p.AddDoorCard(NewCard(CardDesignDiamond, 9, true))

	p.EvalBestHand()
	comp := p.GetComparisonCards()
	assert.Len(t, comp, 5)
	// Should be a copy, not the same slice
	assert.NotSame(t, &p.bestHand[0], &comp[0])
}

func TestSevenCardStudPlayer_JSON(t *testing.T) {
	p := NewSevenCardStudPlayer(true, HoldemStyleLAG)
	p.SetChips(500)
	p.AddHoleCard(NewCard(CardDesignSpade, 1, true))
	p.AddHoleCard(NewCard(CardDesignHeart, 13, true))
	p.AddDoorCard(NewCard(CardDesignDiamond, 10, true))
	p.IncrementTotalHands()
	p.IncrementVPIP()

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored SevenCardStudPlayer
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, HoldemStyleLAG, restored.GetPlayStyle())
	assert.Equal(t, 500, restored.GetChips())
	assert.Len(t, restored.GetHoleCards(), 2)
	assert.Len(t, restored.GetDoorCards(), 1)
	assert.Equal(t, 1, restored.GetTotalHands())
	assert.Equal(t, 1, restored.GetVPIPCount())
}

func TestSevenCardStudPlayer_JSON_Empty(t *testing.T) {
	p := NewSevenCardStudPlayer(false, HoldemStyleGTO)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored SevenCardStudPlayer
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.False(t, restored.GetIsHuman())
	assert.NotNil(t, restored.GetHoleCards())
	assert.NotNil(t, restored.GetDoorCards())
	assert.Empty(t, restored.GetHoleCards())
	assert.Empty(t, restored.GetDoorCards())
}

func TestEvalPartialHand(t *testing.T) {
	tests := []struct {
		name     string
		cards    []*Card
		wantRank int
	}{
		{
			name:     "empty",
			cards:    nil,
			wantRank: PokerHandHighCard,
		},
		{
			name: "high card",
			cards: []*Card{
				NewCard(CardDesignSpade, 10, true),
				NewCard(CardDesignHeart, 5, true),
			},
			wantRank: PokerHandHighCard,
		},
		{
			name: "one pair",
			cards: []*Card{
				NewCard(CardDesignSpade, 10, true),
				NewCard(CardDesignHeart, 10, true),
			},
			wantRank: PokerHandOnePair,
		},
		{
			name: "full house from 4 cards",
			cards: []*Card{
				NewCard(CardDesignSpade, 10, true),
				NewCard(CardDesignHeart, 10, true),
				NewCard(CardDesignDiamond, 10, true),
				NewCard(CardDesignClover, 5, true),
			},
			wantRank: PokerHandThreeOfAKind,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantRank, evalPartialHand(tt.cards))
		})
	}
}

func TestEvalBestHandRazz_PicksLowestCombo(t *testing.T) {
	p := NewSevenCardStudPlayer(true, 0)
	// 7 cards: A,2,3,4,5,K,Q — should pick A-2-3-4-5 (wheel, HighCard)
	p.AddHoleCard(NewCard(CardDesignSpade, 1, false))
	p.AddHoleCard(NewCard(CardDesignHeart, 2, false))
	p.AddHoleCard(NewCard(CardDesignClover, 13, false)) // K
	p.AddDoorCard(NewCard(CardDesignDiamond, 3, false))
	p.AddDoorCard(NewCard(CardDesignSpade, 4, false))
	p.AddDoorCard(NewCard(CardDesignHeart, 5, false))
	p.AddDoorCard(NewCard(CardDesignClover, 12, false)) // Q

	rank := p.EvalBestHandRazz()
	assert.Equal(t, PokerHandHighCard, rank)
	// Best hand should be the wheel (A-2-3-4-5)
	vals := make([]int, 5)
	for i, c := range p.GetBestHand() {
		vals[i] = c.GetValue()
	}
	sort.Ints(vals)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, vals)
}

func TestEvalBestHandRazz_IgnoresFlushAndStraight(t *testing.T) {
	p := NewSevenCardStudPlayer(true, 0)
	// All spades A-2-3-4-5 + 6s + 7s — would be straight flush in normal poker
	// In Razz, still evaluates as HighCard
	p.AddHoleCard(NewCard(CardDesignSpade, 1, false))
	p.AddHoleCard(NewCard(CardDesignSpade, 2, false))
	p.AddHoleCard(NewCard(CardDesignSpade, 6, false))
	p.AddDoorCard(NewCard(CardDesignSpade, 3, false))
	p.AddDoorCard(NewCard(CardDesignSpade, 4, false))
	p.AddDoorCard(NewCard(CardDesignSpade, 5, false))
	p.AddDoorCard(NewCard(CardDesignSpade, 7, false))

	rank := p.EvalBestHandRazz()
	assert.Equal(t, PokerHandHighCard, rank)
}

func TestEvalBestHandRazz_PairIsWorst(t *testing.T) {
	p := NewSevenCardStudPlayer(true, 0)
	// 7 cards: 3,3,5,7,9,J,K — best Razz hand avoids the pair
	p.AddHoleCard(NewCard(CardDesignSpade, 3, false))
	p.AddHoleCard(NewCard(CardDesignHeart, 3, false))
	p.AddHoleCard(NewCard(CardDesignClover, 5, false))
	p.AddDoorCard(NewCard(CardDesignDiamond, 7, false))
	p.AddDoorCard(NewCard(CardDesignSpade, 9, false))
	p.AddDoorCard(NewCard(CardDesignHeart, 11, false))  // J
	p.AddDoorCard(NewCard(CardDesignClover, 13, false)) // K

	rank := p.EvalBestHandRazz()
	// Best combo avoids pair: 3,5,7,9,J = HighCard
	assert.Equal(t, PokerHandHighCard, rank)
}

func TestEvalBestHandRazz_LessThan5Cards(t *testing.T) {
	p := NewSevenCardStudPlayer(true, 0)
	p.AddHoleCard(NewCard(CardDesignSpade, 1, false))
	p.AddDoorCard(NewCard(CardDesignHeart, 2, false))

	rank := p.EvalBestHandRazz()
	assert.Equal(t, PokerHandHighCard, rank)
	assert.Nil(t, p.GetBestHand())
}

func TestSevenCardStudRazzBestLow_CompleteFive(t *testing.T) {
	// 7 cards; best low is the five lowest distinct ranks: 2-3-4-5-7.
	cards := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignClover, 13, false),
	}
	best, complete := SevenCardStudRazzBestLow(cards)
	assert.True(t, complete)
	assert.Len(t, best, 5)
	vals := map[int]bool{}
	for _, c := range best {
		vals[c.GetValue()] = true
	}
	for _, v := range []int{2, 3, 4, 5, 7} {
		assert.True(t, vals[v], "expected %d in best low", v)
	}
	assert.False(t, vals[11] || vals[13], "high cards must be excluded")
}

func TestSevenCardStudRazzBestLow_IncompleteSortedAsc(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 1, false), // Ace low
		NewCard(CardDesignClover, 4, false),
	}
	best, complete := SevenCardStudRazzBestLow(cards)
	assert.False(t, complete)
	assert.Len(t, best, 3)
	// Sorted ascending with Ace (1) first.
	assert.Equal(t, 1, best[0].GetValue())
	assert.Equal(t, 4, best[1].GetValue())
	assert.Equal(t, 7, best[2].GetValue())
}

func TestCompareVisibleHandsLow_LowerRankWins(t *testing.T) {
	a := NewSevenCardStudPlayer(true, 0)
	a.AddDoorCard(NewCard(CardDesignSpade, 2, false)) // HighCard

	b := NewSevenCardStudPlayer(false, 0)
	b.AddDoorCard(NewCard(CardDesignHeart, 5, false))
	b.AddDoorCard(NewCard(CardDesignClover, 5, false)) // OnePair

	// a has HighCard (rank 0), b has OnePair (rank 1). Lower rank = stronger in low.
	result := CompareVisibleHandsLow(a, b)
	assert.Equal(t, 1, result, "HighCard should beat OnePair in lowball")
}

func TestCompareVisibleHandsLow_SameRankLowerCardsWin(t *testing.T) {
	a := NewSevenCardStudPlayer(true, 0)
	a.AddDoorCard(NewCard(CardDesignSpade, 1, false)) // Ace=1
	a.AddDoorCard(NewCard(CardDesignHeart, 3, false))

	b := NewSevenCardStudPlayer(false, 0)
	b.AddDoorCard(NewCard(CardDesignClover, 5, false))
	b.AddDoorCard(NewCard(CardDesignDiamond, 7, false))

	// Both HighCard. a has [3,1], b has [7,5]. a's highest card 3 < 7 => a wins.
	result := CompareVisibleHandsLow(a, b)
	assert.Equal(t, 1, result, "Lower cards should win in lowball")
}

func TestCompareVisibleHandsLow_EqualCards(t *testing.T) {
	a := NewSevenCardStudPlayer(true, 0)
	a.AddDoorCard(NewCard(CardDesignSpade, 5, false))

	b := NewSevenCardStudPlayer(false, 0)
	b.AddDoorCard(NewCard(CardDesignHeart, 5, false))

	result := CompareVisibleHandsLow(a, b)
	assert.Equal(t, 0, result)
}
