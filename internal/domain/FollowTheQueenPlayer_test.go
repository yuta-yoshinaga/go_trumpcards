//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFollowTheQueenPlayer(t *testing.T) {
	p := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	assert.Equal(t, "TAG", p.GetPlayStyleName())
	assert.Empty(t, p.GetHoleCards())
	assert.Empty(t, p.GetDoorCards())
	assert.Nil(t, p.GetBestHand())
}

func TestFollowTheQueenPlayer_CardManagement(t *testing.T) {
	p := NewFollowTheQueenPlayer(false, HoldemStyleLAP)

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

func TestFollowTheQueenPlayer_EvalBestHand(t *testing.T) {
	t.Run("7 cards - finds best 5", func(t *testing.T) {
		p := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
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
		p := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
		p.AddHoleCard(NewCard(CardDesignSpade, 1, true))
		p.AddDoorCard(NewCard(CardDesignHeart, 2, true))

		rank := p.EvalBestHand()
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, p.GetBestHand())
	})

	t.Run("full house", func(t *testing.T) {
		p := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
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

func TestFollowTheQueenPlayer_EvalVisibleHand(t *testing.T) {
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
			p := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
			for _, c := range tt.doors {
				p.AddDoorCard(c)
			}
			assert.Equal(t, tt.wantRank, p.EvalVisibleHand())
		})
	}
}

func TestFollowTheQueenPlayer_CompareVisibleHands(t *testing.T) {
	t.Run("pair beats high card", func(t *testing.T) {
		a := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
		a.AddDoorCard(NewCard(CardDesignSpade, 10, true))
		a.AddDoorCard(NewCard(CardDesignHeart, 10, true))

		b := NewFollowTheQueenPlayer(false, HoldemStyleTAG)
		b.AddDoorCard(NewCard(CardDesignSpade, 1, true))
		b.AddDoorCard(NewCard(CardDesignHeart, 13, true))

		assert.Equal(t, 1, followTheQueenCompareVisibleHands(a, b))
		assert.Equal(t, -1, followTheQueenCompareVisibleHands(b, a))
	})

	t.Run("same rank - high card tiebreak", func(t *testing.T) {
		a := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
		a.AddDoorCard(NewCard(CardDesignSpade, 1, true)) // Ace

		b := NewFollowTheQueenPlayer(false, HoldemStyleTAG)
		b.AddDoorCard(NewCard(CardDesignHeart, 13, true)) // King

		assert.Equal(t, 1, followTheQueenCompareVisibleHands(a, b)) // Ace (14) > King (13)
	})

	t.Run("equal hands", func(t *testing.T) {
		a := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
		a.AddDoorCard(NewCard(CardDesignSpade, 10, true))

		b := NewFollowTheQueenPlayer(false, HoldemStyleTAG)
		b.AddDoorCard(NewCard(CardDesignHeart, 10, true))

		assert.Equal(t, 0, followTheQueenCompareVisibleHands(a, b))
	})
}

func TestFollowTheQueenPlayer_SuitRank(t *testing.T) {
	assert.Equal(t, 4, SuitRank(CardDesignSpade))
	assert.Equal(t, 3, SuitRank(CardDesignHeart))
	assert.Equal(t, 2, SuitRank(CardDesignDiamond))
	assert.Equal(t, 1, SuitRank(CardDesignClover))
	assert.Equal(t, 0, SuitRank(CardDesignJoker))
}

func TestFollowTheQueenPlayer_HUDStats(t *testing.T) {
	p := NewFollowTheQueenPlayer(true, HoldemStyleTAG)

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

func TestFollowTheQueenPlayer_GetComparisonCards(t *testing.T) {
	p := NewFollowTheQueenPlayer(true, HoldemStyleTAG)
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

func TestFollowTheQueenPlayer_JSON(t *testing.T) {
	p := NewFollowTheQueenPlayer(true, HoldemStyleLAG)
	p.SetChips(500)
	p.AddHoleCard(NewCard(CardDesignSpade, 1, true))
	p.AddHoleCard(NewCard(CardDesignHeart, 13, true))
	p.AddDoorCard(NewCard(CardDesignDiamond, 10, true))
	p.IncrementTotalHands()
	p.IncrementVPIP()

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored FollowTheQueenPlayer
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

func TestFollowTheQueenPlayer_JSON_Empty(t *testing.T) {
	p := NewFollowTheQueenPlayer(false, HoldemStyleGTO)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored FollowTheQueenPlayer
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.False(t, restored.GetIsHuman())
	assert.NotNil(t, restored.GetHoleCards())
	assert.NotNil(t, restored.GetDoorCards())
	assert.Empty(t, restored.GetHoleCards())
	assert.Empty(t, restored.GetDoorCards())
}

func TestFollowTheQueenPlayer_EvalPartialHand(t *testing.T) {
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
			assert.Equal(t, tt.wantRank, followTheQueenEvalPartialHand(tt.cards))
		})
	}
}
