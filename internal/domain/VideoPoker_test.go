//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Helper ---

func newTestVideoPoker() *VideoPoker {
	return NewDefaultVideoPoker()
}

// makeHand creates a 5-card hand from (design, value) pairs.
func makeHand(pairs [][2]int) []*Card {
	cards := make([]*Card, len(pairs))
	for i, p := range pairs {
		cards[i] = NewCard(p[0], p[1], true)
	}
	return cards
}

// --- Constructor ---

func TestNewDefaultVideoPoker(t *testing.T) {
	vp := newTestVideoPoker()
	assert.NotNil(t, vp)
	assert.Equal(t, VideoPokerDefaultChips, vp.GetChips())
	assert.Equal(t, VideoPokerPhaseBet, vp.GetPhase())
	assert.False(t, vp.GetGameEndFlag())
}

func TestNewVideoPoker(t *testing.T) {
	tc := NewTrumpCards(0)
	vp := NewVideoPoker(tc)
	assert.NotNil(t, vp)
	assert.Equal(t, VideoPokerPhaseBet, vp.GetPhase())
	assert.Equal(t, 0, vp.GetChips()) // chips not set by NewVideoPoker
}

// --- Bet ---

func TestVideoPoker_Bet_Success(t *testing.T) {
	vp := newTestVideoPoker()
	err := vp.Bet(3)
	assert.NoError(t, err)
	assert.Equal(t, VideoPokerPhaseDraw, vp.GetPhase())
	assert.Equal(t, 3, vp.GetBetAmount())
	assert.Len(t, vp.GetHand(), VideoPokerHandSize)
	assert.Equal(t, VideoPokerDefaultChips-3, vp.GetChips())
}

func TestVideoPoker_Bet_WrongPhase(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	err := vp.Bet(1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrWrongPhase))
}

func TestVideoPoker_Bet_InvalidAmount_Zero(t *testing.T) {
	vp := newTestVideoPoker()
	err := vp.Bet(0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAmount))
}

func TestVideoPoker_Bet_InvalidAmount_TooHigh(t *testing.T) {
	vp := newTestVideoPoker()
	err := vp.Bet(6)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAmount))
}

func TestVideoPoker_Bet_InvalidAmount_Negative(t *testing.T) {
	vp := newTestVideoPoker()
	err := vp.Bet(-1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAmount))
}

func TestVideoPoker_Bet_InsufficientChips(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetChips(0)
	err := vp.Bet(1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInsufficientChips))
}

// --- Hold ---

func TestVideoPoker_Hold_WrongPhase(t *testing.T) {
	vp := newTestVideoPoker()
	err := vp.Hold([]int{0, 1})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrWrongPhase))
}

func TestVideoPoker_Hold_InvalidIndex_OutOfRange(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetHand(makeHand([][2]int{{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}}))
	err := vp.Hold([]int{5})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestVideoPoker_Hold_InvalidIndex_Negative(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetHand(makeHand([][2]int{{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}}))
	err := vp.Hold([]int{-1})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestVideoPoker_Hold_InvalidIndex_Duplicate(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetHand(makeHand([][2]int{{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}}))
	err := vp.Hold([]int{0, 0})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestVideoPoker_Hold_AllCards(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	// Royal Flush: A♠ 10♠ J♠ Q♠ K♠
	hand := makeHand([][2]int{
		{CardDesignSpade, 1}, {CardDesignSpade, 10}, {CardDesignSpade, 11},
		{CardDesignSpade, 12}, {CardDesignSpade, 13},
	})
	vp.SetHand(hand)
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, VideoPokerPhaseResult, vp.GetPhase())
	assert.True(t, vp.GetGameEndFlag())
	// Hand should be unchanged (all held)
	assert.Equal(t, hand, vp.GetHand())
}

func TestVideoPoker_Hold_EmptyIndices(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetHand(makeHand([][2]int{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {1, 6}}))
	err := vp.Hold([]int{})
	assert.NoError(t, err)
	assert.Equal(t, VideoPokerPhaseResult, vp.GetPhase())
	// All cards replaced (we can't assert specific cards due to shuffle)
}

// --- Payout: Royal Flush ---

func TestVideoPoker_Payout_RoyalFlush_MaxBet(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(5)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 1}, {CardDesignSpade, 10}, {CardDesignSpade, 11},
		{CardDesignSpade, 12}, {CardDesignSpade, 13},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, GameResultWin, vp.GetResult())
	assert.Equal(t, 4000, vp.GetPayout()) // 5 * 800
	assert.Equal(t, 4000, vp.GetChips())
	assert.Equal(t, "Royal Flush", vp.GetHandName())
}

func TestVideoPoker_Payout_RoyalFlush_NonMaxBet(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(3)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignHeart, 1}, {CardDesignHeart, 10}, {CardDesignHeart, 11},
		{CardDesignHeart, 12}, {CardDesignHeart, 13},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 750, vp.GetPayout()) // 3 * 250
}

// --- Payout: Straight Flush ---

func TestVideoPoker_Payout_StraightFlush(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(2)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignClover, 5}, {CardDesignClover, 6}, {CardDesignClover, 7},
		{CardDesignClover, 8}, {CardDesignClover, 9},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 100, vp.GetPayout()) // 2 * 50
	assert.Equal(t, "Straight Flush", vp.GetHandName())
}

// --- Payout: Four of a Kind ---

func TestVideoPoker_Payout_FourOfAKind(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 7}, {CardDesignClover, 7}, {CardDesignHeart, 7},
		{CardDesignDiamond, 7}, {CardDesignSpade, 3},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 25, vp.GetPayout()) // 1 * 25
	assert.Equal(t, "Four of a Kind", vp.GetHandName())
}

// --- Payout: Full House ---

func TestVideoPoker_Payout_FullHouse(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 3}, {CardDesignClover, 3}, {CardDesignHeart, 3},
		{CardDesignDiamond, 5}, {CardDesignSpade, 5},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 9, vp.GetPayout()) // 1 * 9
	assert.Equal(t, "Full House", vp.GetHandName())
}

// --- Payout: Flush ---

func TestVideoPoker_Payout_Flush(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignHeart, 2}, {CardDesignHeart, 5}, {CardDesignHeart, 7},
		{CardDesignHeart, 9}, {CardDesignHeart, 11},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 6, vp.GetPayout()) // 1 * 6
	assert.Equal(t, "Flush", vp.GetHandName())
}

// --- Payout: Straight ---

func TestVideoPoker_Payout_Straight(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 4}, {CardDesignClover, 5}, {CardDesignHeart, 6},
		{CardDesignDiamond, 7}, {CardDesignSpade, 8},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 4, vp.GetPayout()) // 1 * 4
	assert.Equal(t, "Straight", vp.GetHandName())
}

// --- Payout: Three of a Kind ---

func TestVideoPoker_Payout_ThreeOfAKind(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 9}, {CardDesignClover, 9}, {CardDesignHeart, 9},
		{CardDesignDiamond, 3}, {CardDesignSpade, 5},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 3, vp.GetPayout()) // 1 * 3
	assert.Equal(t, "Three of a Kind", vp.GetHandName())
}

// --- Payout: Two Pair ---

func TestVideoPoker_Payout_TwoPair(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 4}, {CardDesignClover, 4}, {CardDesignHeart, 8},
		{CardDesignDiamond, 8}, {CardDesignSpade, 2},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 2, vp.GetPayout()) // 1 * 2
	assert.Equal(t, "Two Pair", vp.GetHandName())
}

// --- Payout: Jacks or Better ---

func TestVideoPoker_Payout_JacksOrBetter_Jacks(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 11}, {CardDesignClover, 11}, {CardDesignHeart, 3},
		{CardDesignDiamond, 5}, {CardDesignSpade, 9},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, GameResultWin, vp.GetResult())
	assert.Equal(t, 1, vp.GetPayout()) // 1 * 1
	assert.Equal(t, "Jacks or Better", vp.GetHandName())
}

func TestVideoPoker_Payout_JacksOrBetter_Aces(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(2)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 1}, {CardDesignClover, 1}, {CardDesignHeart, 3},
		{CardDesignDiamond, 5}, {CardDesignSpade, 9},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, GameResultWin, vp.GetResult())
	assert.Equal(t, 2, vp.GetPayout()) // 2 * 1
}

func TestVideoPoker_Payout_JacksOrBetter_Kings(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 13}, {CardDesignClover, 13}, {CardDesignHeart, 3},
		{CardDesignDiamond, 5}, {CardDesignSpade, 9},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, GameResultWin, vp.GetResult())
	assert.Equal(t, 1, vp.GetPayout())
}

// --- Payout: Low Pair (no payout) ---

func TestVideoPoker_Payout_LowPair_NoPayout(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 2}, {CardDesignClover, 2}, {CardDesignHeart, 5},
		{CardDesignDiamond, 8}, {CardDesignSpade, 10},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, GameResultLose, vp.GetResult())
	assert.Equal(t, 0, vp.GetPayout())
	assert.Empty(t, vp.GetHandName())
}

func TestVideoPoker_Payout_HighCard_NoPayout(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 2}, {CardDesignClover, 5}, {CardDesignHeart, 7},
		{CardDesignDiamond, 9}, {CardDesignSpade, 11},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, GameResultLose, vp.GetResult())
	assert.Equal(t, 0, vp.GetPayout())
}

// --- Reset ---

func TestVideoPoker_Reset_ReplenishChips(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetChips(0)
	vp.Reset()
	assert.Equal(t, VideoPokerDefaultChips, vp.GetChips())
	assert.Equal(t, VideoPokerPhaseBet, vp.GetPhase())
	assert.False(t, vp.GetGameEndFlag())
}

func TestVideoPoker_Reset_KeepChips(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetChips(500)
	vp.Reset()
	assert.Equal(t, 500, vp.GetChips())
}

func TestVideoPoker_Reset_ClearsState(t *testing.T) {
	vp := newTestVideoPoker()
	_ = vp.Bet(3)
	vp.Reset()
	assert.Nil(t, vp.GetHand())
	assert.Equal(t, 0, vp.GetBetAmount())
	assert.Equal(t, 0, vp.GetPayout())
	assert.Empty(t, vp.GetHandName())
	assert.Equal(t, [VideoPokerHandSize]bool{}, vp.GetHeldIndices())
}

// --- Action Log ---

func TestVideoPoker_ActionLog(t *testing.T) {
	vp := newTestVideoPoker()
	assert.Nil(t, vp.GetActionLog())
	_ = vp.Bet(1)
	log := vp.GetActionLog()
	assert.Len(t, log, 2) // bet + deal
	assert.Equal(t, "bet", log[0].ActionType)
	assert.Equal(t, "deal", log[1].ActionType)
}

// --- Getters ---

func TestVideoPoker_Getters(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetHandRank(PokerHandFlush)
	assert.Equal(t, PokerHandFlush, vp.GetHandRank())
	vp.SetHandName("Flush")
	assert.Equal(t, "Flush", vp.GetHandName())
	vp.SetPayout(42)
	assert.Equal(t, 42, vp.GetPayout())
	vp.SetResult(GameResultWin)
	assert.Equal(t, GameResultWin, vp.GetResult())
	held := [VideoPokerHandSize]bool{true, false, true, false, true}
	vp.SetHeldIndices(held)
	assert.Equal(t, held, vp.GetHeldIndices())
}

// --- getHandDisplayName edge ---

func TestVideoPoker_getHandDisplayName_Queens(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 12}, {CardDesignClover, 12}, {CardDesignHeart, 3},
		{CardDesignDiamond, 5}, {CardDesignSpade, 9},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, "Jacks or Better", vp.GetHandName())
}

// --- Wheel straight (A-2-3-4-5) ---

func TestVideoPoker_Payout_WheelStraight(t *testing.T) {
	vp := newTestVideoPoker()
	vp.SetPhase(VideoPokerPhaseDraw)
	vp.SetBetAmount(1)
	vp.SetChips(0)
	vp.SetHand(makeHand([][2]int{
		{CardDesignSpade, 1}, {CardDesignClover, 2}, {CardDesignHeart, 3},
		{CardDesignDiamond, 4}, {CardDesignSpade, 5},
	}))
	err := vp.Hold([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, 4, vp.GetPayout())
	assert.Equal(t, "Straight", vp.GetHandName())
}
