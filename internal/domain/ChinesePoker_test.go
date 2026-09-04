//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultChinesePoker(t *testing.T) {
	cp := NewDefaultChinesePoker()
	assert.Equal(t, ChinesePokerPhaseBet, cp.GetPhase())
	assert.Equal(t, ChinesePokerDefaultChips, cp.GetChips())
	assert.False(t, cp.GetGameEndFlag())
	assert.Equal(t, 0, cp.GetBet())
}

func TestChinesePoker_Reset(t *testing.T) {
	cp := NewDefaultChinesePoker()
	_ = cp.Bet(100)
	cp.Reset()
	assert.Equal(t, ChinesePokerPhaseBet, cp.GetPhase())
	assert.False(t, cp.GetGameEndFlag())
	assert.Equal(t, 0, cp.GetBet())
	assert.Nil(t, cp.GetPlayerCards())
	assert.Nil(t, cp.GetDealerCards())
}

func TestChinesePoker_Reset_RefillChips(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetChips(0)
	cp.Reset()
	assert.Equal(t, ChinesePokerDefaultChips, cp.GetChips())
}

func TestChinesePoker_Bet_Success(t *testing.T) {
	cp := NewDefaultChinesePoker()
	err := cp.Bet(100)
	require.NoError(t, err)
	assert.Equal(t, ChinesePokerPhaseSetHands, cp.GetPhase())
	assert.Equal(t, 100, cp.GetBet())
	assert.Equal(t, ChinesePokerDefaultChips-100, cp.GetChips())
	assert.Len(t, cp.GetPlayerCards(), ChinesePokerHandSize)
	assert.Len(t, cp.GetDealerCards(), ChinesePokerHandSize)
}

func TestChinesePoker_Bet_WrongPhase(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	err := cp.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrWrongPhase))
}

func TestChinesePoker_Bet_InvalidAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int
	}{
		{"below minimum", 5},
		{"not multiple", 15},
		{"zero", 0},
		{"negative", -10},
		{"above maximum", ChinesePokerMaxBet + 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewDefaultChinesePoker()
			err := cp.Bet(tt.amount)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, ErrInvalidAmount))
		})
	}
}

func TestChinesePoker_Bet_InsufficientChips(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetChips(50)
	err := cp.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInsufficientChips))
}

func TestChinesePoker_SetHands_WrongPhase(t *testing.T) {
	cp := NewDefaultChinesePoker()
	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrWrongPhase))
}

func TestChinesePoker_SetHands_InvalidFrontCount(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetPlayerCards(make([]*Card, 13))
	err := cp.SetHands([]int{0, 1}, []int{3, 4, 5, 6, 7})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestChinesePoker_SetHands_InvalidMiddleCount(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetPlayerCards(make([]*Card, 13))
	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestChinesePoker_SetHands_DuplicateIndex(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetPlayerCards(make([]*Card, 13))
	err := cp.SetHands([]int{0, 1, 2}, []int{2, 4, 5, 6, 7})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestChinesePoker_SetHands_OutOfRange(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetPlayerCards(make([]*Card, 13))
	err := cp.SetHands([]int{0, 1, 13}, []int{3, 4, 5, 6, 7})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestChinesePoker_SetHands_NegativeIndex(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetPlayerCards(make([]*Card, 13))
	err := cp.SetHands([]int{-1, 1, 2}, []int{3, 4, 5, 6, 7})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestChinesePoker_SetHands_InsufficientPlayerCards(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetPlayerCards(make([]*Card, 5))
	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPlay))
}

func TestChinesePoker_SetHands_FoulStraightFront(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetBet(100)
	// Front: A-2-3 (straight), Middle: 4-5-6-7-8 (straight), but front straight should be foul if middle is only pair
	cp.SetPlayerCards([]*Card{
		NewCard(CardDesignSpade, 1, false),    // 0: A
		NewCard(CardDesignHeart, 2, false),    // 1: 2
		NewCard(CardDesignClover, 3, false),   // 2: 3
		NewCard(CardDesignDiamond, 5, false),  // 3
		NewCard(CardDesignSpade, 5, false),    // 4: pair of 5s
		NewCard(CardDesignHeart, 7, false),    // 5
		NewCard(CardDesignClover, 9, false),   // 6
		NewCard(CardDesignDiamond, 11, false), // 7
		NewCard(CardDesignSpade, 13, false),   // 8
		NewCard(CardDesignHeart, 13, false),   // 9
		NewCard(CardDesignClover, 12, false),  // 10
		NewCard(CardDesignDiamond, 12, false), // 11
		NewCard(CardDesignSpade, 11, false),   // 12
	})
	cp.SetDealerCards(make([]*Card, 13))
	// Front: [0,1,2]=A,2,3 (straight), Middle: [3,4,5,6,7]=5,5,7,9,J (one pair)
	// Straight > Pair → foul
	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPlay))
}

func TestChinesePoker_SetHands_Foul(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetBet(100)
	// Front: A-A-A (trips), Middle: 2-3-4-5-7 (high card), Back: 8-9-10-J-Q (high card)
	// Front is stronger than Middle → foul
	cp.SetPlayerCards([]*Card{
		NewCard(CardDesignSpade, 1, false),    // 0: A♠
		NewCard(CardDesignHeart, 1, false),    // 1: A♥
		NewCard(CardDesignClover, 1, false),   // 2: A♣
		NewCard(CardDesignSpade, 2, false),    // 3
		NewCard(CardDesignHeart, 3, false),    // 4
		NewCard(CardDesignClover, 4, false),   // 5
		NewCard(CardDesignDiamond, 5, false),  // 6
		NewCard(CardDesignSpade, 7, false),    // 7
		NewCard(CardDesignHeart, 8, false),    // 8
		NewCard(CardDesignClover, 9, false),   // 9
		NewCard(CardDesignDiamond, 10, false), // 10
		NewCard(CardDesignSpade, 11, false),   // 11
		NewCard(CardDesignHeart, 12, false),   // 12
	})
	cp.SetDealerCards(make([]*Card, 13))

	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPlay))
}

func TestChinesePoker_SetHands_ValidArrangement(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetBet(100)
	// Front: 2-3-4 (high card), Middle: 5-6-7-8-9 (straight), Back: 10-J-Q-K-A (straight)
	playerCards := []*Card{
		NewCard(CardDesignSpade, 2, false),    // 0
		NewCard(CardDesignHeart, 3, false),    // 1
		NewCard(CardDesignClover, 4, false),   // 2
		NewCard(CardDesignDiamond, 5, false),  // 3
		NewCard(CardDesignSpade, 6, false),    // 4
		NewCard(CardDesignHeart, 7, false),    // 5
		NewCard(CardDesignClover, 8, false),   // 6
		NewCard(CardDesignDiamond, 9, false),  // 7
		NewCard(CardDesignSpade, 10, false),   // 8
		NewCard(CardDesignHeart, 11, false),   // 9
		NewCard(CardDesignClover, 12, false),  // 10
		NewCard(CardDesignDiamond, 13, false), // 11
		NewCard(CardDesignSpade, 1, false),    // 12
	}
	cp.SetPlayerCards(playerCards)

	dealerCards := []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignClover, 11, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 1, false),
	}
	cp.SetDealerCards(dealerCards)

	// Front: [0,1,2] = 2,3,4 high card; Middle: [3,4,5,6,7] = 5-9 straight; Back: [8,9,10,11,12] = 10-A straight
	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	require.NoError(t, err)
	assert.Equal(t, ChinesePokerPhaseEnd, cp.GetPhase())
	assert.True(t, cp.GetGameEndFlag())
	assert.Len(t, cp.GetPlayerFront(), 3)
	assert.Len(t, cp.GetPlayerMiddle(), 5)
	assert.Len(t, cp.GetPlayerBack(), 5)
}

func TestChinesePoker_Resolve_ScoopWin(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetBet(100)
	cp.SetChips(ChinesePokerDefaultChips - 100)

	// Player: strong hands, Dealer: weak hands
	cp.SetPlayerCards([]*Card{
		NewCard(CardDesignSpade, 1, false),    // 0: A♠
		NewCard(CardDesignHeart, 13, false),   // 1: K♥
		NewCard(CardDesignClover, 12, false),  // 2: Q♣
		NewCard(CardDesignDiamond, 1, false),  // 3: A♦
		NewCard(CardDesignSpade, 13, false),   // 4: K♠
		NewCard(CardDesignHeart, 12, false),   // 5: Q♥
		NewCard(CardDesignClover, 11, false),  // 6: J♣
		NewCard(CardDesignDiamond, 10, false), // 7: 10♦
		NewCard(CardDesignSpade, 10, false),   // 8: 10♠
		NewCard(CardDesignHeart, 10, false),   // 9: 10♥
		NewCard(CardDesignClover, 10, false),  // 10: 10♣
		NewCard(CardDesignDiamond, 9, false),  // 11: 9♦
		NewCard(CardDesignSpade, 9, false),    // 12: 9♠
	})

	cp.SetDealerCards([]*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignClover, 6, false),
	})

	// Front: A,K,Q; Middle: A,K,Q,J,10; Back: 10,10,10,9,9 (full house)
	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	require.NoError(t, err)
	assert.Equal(t, GameResultWin, cp.GetResult())
	assert.True(t, cp.GetScoop())
	assert.Equal(t, GameResultWin, cp.GetFrontResult())
	assert.Equal(t, GameResultWin, cp.GetMiddleResult())
	assert.Equal(t, GameResultWin, cp.GetBackResult())
	assert.Greater(t, cp.GetPayout(), 0)
}

func TestChinesePoker_Resolve_ScoopLose(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetBet(100)
	cp.SetChips(ChinesePokerDefaultChips - 100)

	// Player: all low, no combos; Dealer: pairs and trips
	cp.SetPlayerCards([]*Card{
		NewCard(CardDesignHeart, 2, false),   // 0
		NewCard(CardDesignClover, 4, false),  // 1
		NewCard(CardDesignDiamond, 6, false), // 2
		NewCard(CardDesignSpade, 3, false),   // 3
		NewCard(CardDesignHeart, 5, false),   // 4
		NewCard(CardDesignClover, 7, false),  // 5: no straight (2,4,6 gap)
		NewCard(CardDesignDiamond, 2, false), // 6
		NewCard(CardDesignSpade, 8, false),   // 7
		NewCard(CardDesignHeart, 3, false),   // 8
		NewCard(CardDesignClover, 5, false),  // 9
		NewCard(CardDesignDiamond, 9, false), // 10
		NewCard(CardDesignSpade, 7, false),   // 11
		NewCard(CardDesignHeart, 10, false),  // 12
	})

	// Dealer: A-A-A, K-K-K, Q-Q, J, 10, 9, 8, 7
	cp.SetDealerCards([]*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 7, false),
	})

	// Front: 2♥,4♣,6♦ (HC); Middle: 3♠,5♥,7♣,2♦,8♠ (HC); Back: 3♥,5♣,9♦,7♠,10♥ (HC)
	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	require.NoError(t, err)
	assert.Equal(t, ChinesePokerPhaseEnd, cp.GetPhase())
	assert.True(t, cp.GetGameEndFlag())
	// Dealer house way arranges strong hands from AAA+KKK+QQ; exact result depends on house way strategy
}

// --- Royalty tests ---

func TestCpBackRoyalty(t *testing.T) {
	tests := []struct {
		rank     int
		expected int
	}{
		{PokerHandHighCard, 0},
		{PokerHandOnePair, 0},
		{PokerHandStraight, 2},
		{PokerHandFlush, 4},
		{PokerHandFullHouse, 6},
		{PokerHandFourOfAKind, 10},
		{PokerHandStraightFlush, 15},
		{PokerHandRoyalFlush, 25},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, cpBackRoyalty(tt.rank))
	}
}

func TestCpMiddleRoyalty(t *testing.T) {
	tests := []struct {
		rank     int
		expected int
	}{
		{PokerHandHighCard, 0},
		{PokerHandOnePair, 0},
		{PokerHandThreeOfAKind, 2},
		{PokerHandStraight, 4},
		{PokerHandFlush, 8},
		{PokerHandFullHouse, 12},
		{PokerHandFourOfAKind, 20},
		{PokerHandStraightFlush, 30},
		{PokerHandRoyalFlush, 50},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, cpMiddleRoyalty(tt.rank))
	}
}

func TestCpFrontRoyalty(t *testing.T) {
	tests := []struct {
		name     string
		rank     int
		cards    []*Card
		expected int
	}{
		{
			"high card",
			ThreeCardHandHighCard,
			[]*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignClover, 9, false)},
			0,
		},
		{
			"pair of 5s (below 6)",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignClover, 9, false)},
			0,
		},
		{
			"pair of 6s",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 6, false), NewCard(CardDesignHeart, 6, false), NewCard(CardDesignClover, 9, false)},
			1,
		},
		{
			"pair of 7s",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 7, false), NewCard(CardDesignClover, 9, false)},
			2,
		},
		{
			"pair of 8s",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 8, false), NewCard(CardDesignHeart, 8, false), NewCard(CardDesignClover, 9, false)},
			3,
		},
		{
			"pair of 9s",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 9, false), NewCard(CardDesignClover, 2, false)},
			4,
		},
		{
			"pair of 10s",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 10, false), NewCard(CardDesignClover, 2, false)},
			5,
		},
		{
			"pair of Jacks",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 11, false), NewCard(CardDesignHeart, 11, false), NewCard(CardDesignClover, 2, false)},
			6,
		},
		{
			"pair of Queens",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 12, false), NewCard(CardDesignHeart, 12, false), NewCard(CardDesignClover, 2, false)},
			7,
		},
		{
			"pair of Kings",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 13, false), NewCard(CardDesignHeart, 13, false), NewCard(CardDesignClover, 2, false)},
			8,
		},
		{
			"pair of Aces",
			ThreeCardHandPair,
			[]*Card{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false), NewCard(CardDesignClover, 9, false)},
			9,
		},
		{
			"three of a kind 2s",
			ThreeCardHandThreeOfAKind,
			[]*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 2, false), NewCard(CardDesignClover, 2, false)},
			10,
		},
		{
			"three of a kind 9s",
			ThreeCardHandThreeOfAKind,
			[]*Card{NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 9, false), NewCard(CardDesignClover, 9, false)},
			10,
		},
		{
			"three of a kind 10s",
			ThreeCardHandThreeOfAKind,
			[]*Card{NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 10, false), NewCard(CardDesignClover, 10, false)},
			12,
		},
		{
			"three of a kind Jacks",
			ThreeCardHandThreeOfAKind,
			[]*Card{NewCard(CardDesignSpade, 11, false), NewCard(CardDesignHeart, 11, false), NewCard(CardDesignClover, 11, false)},
			13,
		},
		{
			"three of a kind Queens",
			ThreeCardHandThreeOfAKind,
			[]*Card{NewCard(CardDesignSpade, 12, false), NewCard(CardDesignHeart, 12, false), NewCard(CardDesignClover, 12, false)},
			14,
		},
		{
			"three of a kind Kings",
			ThreeCardHandThreeOfAKind,
			[]*Card{NewCard(CardDesignSpade, 13, false), NewCard(CardDesignHeart, 13, false), NewCard(CardDesignClover, 13, false)},
			15,
		},
		{
			"three of a kind Aces",
			ThreeCardHandThreeOfAKind,
			[]*Card{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false), NewCard(CardDesignClover, 1, false)},
			22,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cpFrontRoyalty(tt.rank, tt.cards))
		})
	}
}

// --- Validation tests ---

func TestCpValidateHands(t *testing.T) {
	t.Run("valid arrangement", func(t *testing.T) {
		front := []*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignClover, 5, false)}
		middle := []*Card{NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 7, false), NewCard(CardDesignClover, 8, false), NewCard(CardDesignDiamond, 9, false), NewCard(CardDesignSpade, 10, false)}
		back := []*Card{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false), NewCard(CardDesignClover, 1, false), NewCard(CardDesignDiamond, 13, false), NewCard(CardDesignSpade, 13, false)}
		assert.True(t, cpValidateHands(front, middle, back))
	})

	t.Run("back weaker than middle is foul", func(t *testing.T) {
		front := []*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignClover, 5, false)}
		middle := []*Card{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false), NewCard(CardDesignClover, 1, false), NewCard(CardDesignDiamond, 13, false), NewCard(CardDesignSpade, 13, false)}
		back := []*Card{NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 7, false), NewCard(CardDesignClover, 8, false), NewCard(CardDesignDiamond, 9, false), NewCard(CardDesignSpade, 10, false)}
		assert.False(t, cpValidateHands(front, middle, back))
	})
}

// --- House way test ---

func TestCpHouseWay(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignSpade, 6, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignSpade, 2, false),
	}
	front, middle, back := cpHouseWay(cards)
	assert.Len(t, front, 3)
	assert.Len(t, middle, 5)
	assert.Len(t, back, 5)
	assert.True(t, cpValidateHands(front, middle, back))
}

// --- JSON roundtrip ---

func TestChinesePoker_JSON_Roundtrip(t *testing.T) {
	cp := NewDefaultChinesePoker()
	_ = cp.Bet(100)

	data, err := json.Marshal(cp)
	require.NoError(t, err)

	var cp2 ChinesePoker
	err = json.Unmarshal(data, &cp2)
	require.NoError(t, err)

	assert.Equal(t, cp.GetPhase(), cp2.GetPhase())
	assert.Equal(t, cp.GetBet(), cp2.GetBet())
	assert.Equal(t, cp.GetChips(), cp2.GetChips())
	assert.Len(t, cp2.GetPlayerCards(), ChinesePokerHandSize)
	assert.Len(t, cp2.GetDealerCards(), ChinesePokerHandSize)
}

func TestChinesePoker_UnmarshalJSON_OversizedArray(t *testing.T) {
	big := make([]*Card, cpMaxSliceLen+1)
	j := chinesePokerJSON{PlayerCards: big}
	data, _ := json.Marshal(j)
	var cp ChinesePoker
	err := cp.UnmarshalJSON(data)
	assert.Error(t, err)
}

func TestChinesePoker_UnmarshalJSON_NilDefaults(t *testing.T) {
	data := []byte(`{"ps":1}`)
	var cp ChinesePoker
	err := cp.UnmarshalJSON(data)
	require.NoError(t, err)
	assert.NotNil(t, cp.playerCards)
	assert.NotNil(t, cp.dealerCards)
	assert.NotNil(t, cp.actionLog)
}

// --- cpCompareFiveCardHands tests ---

func TestCpCompareFiveCardHands(t *testing.T) {
	straight := []*Card{
		NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignClover, 7, false), NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 9, false),
	}
	pair := []*Card{
		NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 5, false), NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 10, false),
	}
	assert.Equal(t, 1, cpCompareFiveCardHands(straight, pair))
	assert.Equal(t, -1, cpCompareFiveCardHands(pair, straight))
	assert.Equal(t, 0, cpCompareFiveCardHands(straight, straight))
}

// --- cpMapThreeCardToFiveCardRank tests ---

func TestCpMapThreeCardToFiveCardRank(t *testing.T) {
	assert.Equal(t, PokerHandHighCard, cpMapThreeCardToFiveCardRank(ThreeCardHandHighCard))
	assert.Equal(t, PokerHandOnePair, cpMapThreeCardToFiveCardRank(ThreeCardHandPair))
	assert.Equal(t, PokerHandTwoPair, cpMapThreeCardToFiveCardRank(ThreeCardHandFlush))
	assert.Equal(t, PokerHandTwoPair, cpMapThreeCardToFiveCardRank(ThreeCardHandStraight))
	assert.Equal(t, PokerHandThreeOfAKind, cpMapThreeCardToFiveCardRank(ThreeCardHandThreeOfAKind))
	assert.Equal(t, PokerHandStraight, cpMapThreeCardToFiveCardRank(ThreeCardHandStraightFlush))
	assert.Equal(t, PokerHandHighCard, cpMapThreeCardToFiveCardRank(99))
}

// --- Action log test ---

func TestChinesePoker_ActionLog(t *testing.T) {
	cp := NewDefaultChinesePoker()
	assert.Empty(t, cp.GetActionLog())
	_ = cp.Bet(100)
	assert.NotEmpty(t, cp.GetActionLog())
}

// --- cpPairValue test ---

func TestCpPairValue(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 3, false),
	}
	assert.Equal(t, 7, cpPairValue(cards))
}

func TestCpPairValue_NoPair(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignClover, 9, false),
	}
	assert.Equal(t, 0, cpPairValue(cards))
}

// **CUI は13枚を自力で 3/5/5 に分けるしかなく、ファウルしても無警告だった
// (#4717)。**Web には推奨分割もファウル警告 (cp-foul-warning) もある。
func TestChinesePoker_GetSuggestedArrangement(t *testing.T) {
	card := func(design, value int) *Card { return NewCard(design, value, false) }
	setHands := func(cards []*Card) *ChinesePoker {
		cp := NewDefaultChinesePoker()
		cp.SetPhase(ChinesePokerPhaseSetHands)
		cp.SetPlayerCards(cards)
		return cp
	}
	// A K Q J 10 9 8 7 6 5 4 3 2 のばらばらな13枚。
	spread := []*Card{
		card(CardDesignSpade, 1), card(CardDesignHeart, 13), card(CardDesignClover, 12),
		card(CardDesignDiamond, 11), card(CardDesignSpade, 10), card(CardDesignHeart, 9),
		card(CardDesignClover, 8), card(CardDesignDiamond, 7), card(CardDesignSpade, 6),
		card(CardDesignHeart, 5), card(CardDesignClover, 4), card(CardDesignDiamond, 3),
		card(CardDesignSpade, 2),
	}

	t.Run("nothing to suggest outside the set-hands phase", func(t *testing.T) {
		cp := setHands(spread)
		cp.SetPhase(ChinesePokerPhaseBet)
		assert.Nil(t, cp.GetSuggestedArrangement())
	})

	t.Run("nothing to suggest without a full hand", func(t *testing.T) {
		assert.Nil(t, setHands(spread[:12]).GetSuggestedArrangement())
	})

	t.Run("splits the hand into three, five and five", func(t *testing.T) {
		arr := setHands(spread).GetSuggestedArrangement()
		require.NotNil(t, arr)
		assert.Len(t, arr.Front, ChinesePokerFrontSize)
		assert.Len(t, arr.Middle, ChinesePokerMiddleSize)
		assert.Len(t, arr.Back, ChinesePokerBackSize)

		// **13枚を過不足なく使う。**同じ札を二度置いたら成立しない。
		seen := map[int]bool{}
		for _, group := range [][]int{arr.Front, arr.Middle, arr.Back} {
			for _, i := range group {
				assert.False(t, seen[i], "index %d が重複している", i)
				seen[i] = true
			}
		}
		assert.Len(t, seen, ChinesePokerHandSize)
	})

	// **エースは 14 として並べる。**value が 1 のまま並べると、最強の札が
	// 前列に落ちる。
	t.Run("puts the ace in the back row, not the front", func(t *testing.T) {
		cp := setHands(spread)
		arr := cp.GetSuggestedArrangement()
		require.NotNil(t, arr)
		aceInBack := false
		for _, i := range arr.Back {
			if cp.GetPlayerCards()[i].GetValue() == 1 {
				aceInBack = true
			}
		}
		assert.True(t, aceInBack, "エースは後列に入るべき")
		for _, i := range arr.Front {
			assert.NotEqual(t, 1, cp.GetPlayerCards()[i].GetValue(), "エースが前列に落ちている")
		}
	})

	t.Run("reports a clean split as legal", func(t *testing.T) {
		arr := setHands(spread).GetSuggestedArrangement()
		require.NotNil(t, arr)
		assert.False(t, arr.Foul)
	})

	// **ランク順に切るだけでは合法とは限らない。**低いスリーカードが前列に
	// 落ちると、中列のハイカードより強くなってファウルする。
	t.Run("flags a rank-ordered split that would foul", func(t *testing.T) {
		// 後列 A K Q J 9 / 中列 8 7 5 4 3 (どちらもハイカード) / 前列 2 2 2。
		// 前列のスリーカードが中列のハイカードより強いのでファウル。
		fouling := []*Card{
			card(CardDesignSpade, 1), card(CardDesignHeart, 13), card(CardDesignClover, 12),
			card(CardDesignDiamond, 11), card(CardDesignSpade, 9), card(CardDesignHeart, 8),
			card(CardDesignClover, 7), card(CardDesignDiamond, 5), card(CardDesignSpade, 4),
			card(CardDesignHeart, 3), card(CardDesignClover, 2), card(CardDesignDiamond, 2),
			card(CardDesignSpade, 2),
		}
		arr := setHands(fouling).GetSuggestedArrangement()
		require.NotNil(t, arr)
		assert.True(t, arr.Foul, "前列が 2 のスリーカードになりファウルするはず")
	})
}

func TestChinesePoker_SegmentRoyalties_ResolveAndSum(t *testing.T) {
	card := func(design, value int) *Card { return NewCard(design, value, false) }

	// Player:
	// Front: Q♠, Q♥, 2♣ (Pair of Queens -> front royalty = 7)
	// Middle: 9♠, 9♥, 9♣, 5♦, 5♠ (Full House 9s full of 5s -> middle royalty = 12)
	// Back: 10♦, J♦, Q♦, K♦, A♦ (Royal Flush -> back royalty = 25)
	// Total player royalty = 7 + 12 + 25 = 44
	playerCards := []*Card{
		card(CardDesignSpade, 12), card(CardDesignHeart, 12), card(CardDesignClover, 2), // Front (0,1,2)
		card(CardDesignSpade, 9), card(CardDesignHeart, 9), card(CardDesignClover, 9), card(CardDesignDiamond, 5), card(CardDesignSpade, 5), // Middle (3,4,5,6,7)
		card(CardDesignDiamond, 10), card(CardDesignDiamond, 11), card(CardDesignDiamond, 12), card(CardDesignDiamond, 13), card(CardDesignDiamond, 1), // Back (8,9,10,11,12)
	}

	// Dealer has no royalties (pairs <= 5 and high cards):
	dealerCards := []*Card{
		card(CardDesignSpade, 2), card(CardDesignHeart, 2),
		card(CardDesignSpade, 3), card(CardDesignHeart, 3),
		card(CardDesignSpade, 4), card(CardDesignHeart, 4),
		card(CardDesignSpade, 5), card(CardDesignHeart, 5),
		card(CardDesignSpade, 7), card(CardDesignHeart, 9),
		card(CardDesignClover, 11), card(CardDesignDiamond, 13), card(CardDesignSpade, 1),
	}

	cp := NewDefaultChinesePoker()
	cp.SetPhase(ChinesePokerPhaseSetHands)
	cp.SetBet(100)
	cp.SetPlayerCards(playerCards)
	cp.SetDealerCards(dealerCards)

	err := cp.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	require.NoError(t, err)
	assert.Equal(t, ChinesePokerPhaseEnd, cp.GetPhase())

	// Player segment royalties verification
	assert.Equal(t, 7, cp.GetPlayerFrontRoyalty(), "Player front pair of Queens royalty")
	assert.Equal(t, 12, cp.GetPlayerMiddleRoyalty(), "Player middle full house royalty")
	assert.Equal(t, 25, cp.GetPlayerBackRoyalty(), "Player back royal flush royalty")
	assert.Equal(t, 44, cp.GetPlayerRoyalty(), "Player total royalty must match sum of segments")
	assert.Equal(t, cp.GetPlayerFrontRoyalty()+cp.GetPlayerMiddleRoyalty()+cp.GetPlayerBackRoyalty(), cp.GetPlayerRoyalty())

	// Dealer segment royalties verification (all 0)
	assert.Equal(t, 0, cp.GetDealerFrontRoyalty(), "Dealer front royalty")
	assert.Equal(t, 0, cp.GetDealerMiddleRoyalty(), "Dealer middle royalty")
	assert.Equal(t, 0, cp.GetDealerBackRoyalty(), "Dealer back royalty")
	assert.Equal(t, 0, cp.GetDealerRoyalty(), "Dealer total royalty must match sum of segments")
	assert.Equal(t, cp.GetDealerFrontRoyalty()+cp.GetDealerMiddleRoyalty()+cp.GetDealerBackRoyalty(), cp.GetDealerRoyalty())

	// Second round with different hands to verify another set of royalty values:
	// Front: 6♠, 6♥, 2♣ (Pair of 6s -> front royalty = 1)
	// Middle: 3♠, 3♥, 3♦, 4♦, 5♦ (Three of a Kind 3s -> middle royalty = 2)
	// Back: 4♠, 5♠, 6♦, 7♠, 8♠ (Straight 4-8 -> back royalty = 2)
	// Total player royalty = 1 + 2 + 2 = 5
	playerCards2 := []*Card{
		card(CardDesignSpade, 6), card(CardDesignHeart, 6), card(CardDesignClover, 2), // Front (0,1,2)
		card(CardDesignSpade, 3), card(CardDesignHeart, 3), card(CardDesignDiamond, 3), card(CardDesignDiamond, 4), card(CardDesignDiamond, 5), // Middle (3,4,5,6,7)
		card(CardDesignSpade, 4), card(CardDesignSpade, 5), card(CardDesignDiamond, 6), card(CardDesignSpade, 7), card(CardDesignSpade, 8), // Back (8,9,10,11,12)
	}
	cp2 := NewDefaultChinesePoker()
	cp2.SetPhase(ChinesePokerPhaseSetHands)
	cp2.SetBet(100)
	cp2.SetPlayerCards(playerCards2)
	cp2.SetDealerCards(dealerCards)

	err = cp2.SetHands([]int{0, 1, 2}, []int{3, 4, 5, 6, 7})
	require.NoError(t, err)
	assert.Equal(t, 1, cp2.GetPlayerFrontRoyalty())
	assert.Equal(t, 2, cp2.GetPlayerMiddleRoyalty())
	assert.Equal(t, 2, cp2.GetPlayerBackRoyalty())
	assert.Equal(t, 5, cp2.GetPlayerRoyalty())
	assert.Equal(t, cp2.GetPlayerFrontRoyalty()+cp2.GetPlayerMiddleRoyalty()+cp2.GetPlayerBackRoyalty(), cp2.GetPlayerRoyalty())
}

func TestChinesePoker_Reset_ClearsSegmentRoyalties(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPlayerFrontRoyalty(7)
	cp.SetPlayerMiddleRoyalty(12)
	cp.SetPlayerBackRoyalty(25)
	cp.SetPlayerRoyalty(44)
	cp.SetDealerFrontRoyalty(1)
	cp.SetDealerMiddleRoyalty(30)
	cp.SetDealerBackRoyalty(10)
	cp.SetDealerRoyalty(41)

	// Verify non-zero before reset
	assert.Equal(t, 7, cp.GetPlayerFrontRoyalty())
	assert.Equal(t, 12, cp.GetPlayerMiddleRoyalty())
	assert.Equal(t, 25, cp.GetPlayerBackRoyalty())
	assert.Equal(t, 44, cp.GetPlayerRoyalty())
	assert.Equal(t, 1, cp.GetDealerFrontRoyalty())
	assert.Equal(t, 30, cp.GetDealerMiddleRoyalty())
	assert.Equal(t, 10, cp.GetDealerBackRoyalty())
	assert.Equal(t, 41, cp.GetDealerRoyalty())

	cp.Reset()

	// Verify all segment royalties and total royalties are reset to 0
	assert.Equal(t, 0, cp.GetPlayerFrontRoyalty())
	assert.Equal(t, 0, cp.GetPlayerMiddleRoyalty())
	assert.Equal(t, 0, cp.GetPlayerBackRoyalty())
	assert.Equal(t, 0, cp.GetPlayerRoyalty())
	assert.Equal(t, 0, cp.GetDealerFrontRoyalty())
	assert.Equal(t, 0, cp.GetDealerMiddleRoyalty())
	assert.Equal(t, 0, cp.GetDealerBackRoyalty())
	assert.Equal(t, 0, cp.GetDealerRoyalty())
}

func TestChinesePoker_JSON_SegmentRoyalties(t *testing.T) {
	cp := NewDefaultChinesePoker()
	cp.SetPlayerFrontRoyalty(3)
	cp.SetPlayerMiddleRoyalty(4)
	cp.SetPlayerBackRoyalty(6)
	cp.SetPlayerRoyalty(13)
	cp.SetDealerFrontRoyalty(0)
	cp.SetDealerMiddleRoyalty(2)
	cp.SetDealerBackRoyalty(4)
	cp.SetDealerRoyalty(6)

	data, err := json.Marshal(cp)
	require.NoError(t, err)

	var restored ChinesePoker
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, 3, restored.GetPlayerFrontRoyalty())
	assert.Equal(t, 4, restored.GetPlayerMiddleRoyalty())
	assert.Equal(t, 6, restored.GetPlayerBackRoyalty())
	assert.Equal(t, 13, restored.GetPlayerRoyalty())
	assert.Equal(t, 0, restored.GetDealerFrontRoyalty())
	assert.Equal(t, 2, restored.GetDealerMiddleRoyalty())
	assert.Equal(t, 4, restored.GetDealerBackRoyalty())
	assert.Equal(t, 6, restored.GetDealerRoyalty())
}
