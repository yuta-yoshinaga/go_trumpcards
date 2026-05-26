package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewDefaultRussianPoker(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	assert.Equal(t, domain.RussianPokerPhaseBet, rp.GetPhase())
	assert.Equal(t, domain.RussianPokerDefaultChips, rp.GetChips())
	assert.False(t, rp.GetGameEndFlag())
	assert.Nil(t, rp.GetPlayerHand())
	assert.Nil(t, rp.GetDealerHand())
}

func TestRussianPoker_Reset(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	require.NoError(t, rp.Play())
	// might be ForceQualify or End depending on dealer hand
	if rp.GetPhase() == domain.RussianPokerPhaseForceQualify {
		require.NoError(t, rp.Decline())
	}
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())

	rp.Reset()
	assert.Equal(t, domain.RussianPokerPhaseBet, rp.GetPhase())
	assert.False(t, rp.GetGameEndFlag())
	assert.Nil(t, rp.GetPlayerHand())
	assert.Nil(t, rp.GetDealerHand())
	assert.Equal(t, 0, rp.GetAnteBet())
	assert.Equal(t, 0, rp.GetPlayBet())
	assert.Equal(t, 0, rp.GetExchangeCount())
	assert.Equal(t, 0, rp.GetExchangeFee())
	assert.False(t, rp.GetBought6th())
	assert.Equal(t, 0, rp.GetBuy6thFee())
	assert.False(t, rp.GetForceExchanged())
	assert.Equal(t, 0, rp.GetForceExchangeFee())
}

func TestRussianPoker_Reset_RefillChips(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(5)
	rp.Reset()
	assert.Equal(t, domain.RussianPokerDefaultChips, rp.GetChips())
}

func TestRussianPoker_Bet_WrongPhase(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetPhase(domain.RussianPokerPhaseAction)
	err := rp.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRussianPoker_Bet_InvalidAnteAmount(t *testing.T) {
	tests := []struct {
		name string
		ante int
	}{
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
		{"Zero", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := domain.NewDefaultRussianPoker()
			err := rp.Bet(tt.ante)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestRussianPoker_Bet_InsufficientChips(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(50)
	err := rp.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestRussianPoker_Bet_Success(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	err := rp.Bet(100)
	assert.NoError(t, err)
	assert.Equal(t, domain.RussianPokerPhaseAction, rp.GetPhase())
	assert.Equal(t, 100, rp.GetAnteBet())
	assert.Len(t, rp.GetPlayerHand(), 5)
	assert.Len(t, rp.GetDealerHand(), 5)
	assert.Equal(t, domain.RussianPokerDefaultChips-100, rp.GetChips())
}

// --- Exchange ---

func TestRussianPoker_Exchange_WrongPhase(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	err := rp.Exchange([]int{0})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRussianPoker_Exchange_EmptyIndices(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	err := rp.Exchange([]int{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
}

func TestRussianPoker_Exchange_TooManyIndices(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	err := rp.Exchange([]int{0, 1, 2, 3, 4, 0})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
}

func TestRussianPoker_Exchange_OutOfRange(t *testing.T) {
	for _, idx := range []int{-1, 5, 99} {
		rp := domain.NewDefaultRussianPoker()
		require.NoError(t, rp.Bet(100))
		err := rp.Exchange([]int{idx})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	}
}

func TestRussianPoker_Exchange_Duplicate(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	err := rp.Exchange([]int{1, 1})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
}

func TestRussianPoker_Exchange_InsufficientFee(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	rp.SetChips(50)
	err := rp.Exchange([]int{0})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestRussianPoker_Exchange_DeductsFee(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	chipsBefore := rp.GetChips()
	require.NoError(t, rp.Exchange([]int{0, 1, 2}))
	assert.Equal(t, domain.RussianPokerPhasePostAction, rp.GetPhase())
	assert.Equal(t, 3, rp.GetExchangeCount())
	assert.Equal(t, 300, rp.GetExchangeFee())
	assert.Equal(t, chipsBefore-300, rp.GetChips())
}

func TestRussianPoker_Exchange_ReplacesCards(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	originalHand := append([]*domain.Card(nil), rp.GetPlayerHand()...)
	require.NoError(t, rp.Exchange([]int{0, 4}))
	newHand := rp.GetPlayerHand()
	assert.Same(t, originalHand[1], newHand[1])
	assert.Same(t, originalHand[2], newHand[2])
	assert.Same(t, originalHand[3], newHand[3])
}

// --- Buy6th ---

func TestRussianPoker_Buy6th_WrongPhase(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	err := rp.Buy6th()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRussianPoker_Buy6th_InsufficientChips(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	rp.SetChips(50)
	err := rp.Buy6th()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestRussianPoker_Buy6th_Success(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	chipsBefore := rp.GetChips()
	require.NoError(t, rp.Buy6th())
	assert.Equal(t, domain.RussianPokerPhaseSelect, rp.GetPhase())
	assert.True(t, rp.GetBought6th())
	assert.Equal(t, 100, rp.GetBuy6thFee())
	assert.Len(t, rp.GetPlayerHand(), 6)
	assert.Equal(t, chipsBefore-100, rp.GetChips())
}

// --- Select ---

func TestRussianPoker_Select_WrongPhase(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	err := rp.Select(0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRussianPoker_Select_OutOfRange(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	require.NoError(t, rp.Buy6th())
	assert.Equal(t, domain.RussianPokerPhaseSelect, rp.GetPhase())

	err := rp.Select(-1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))

	err = rp.Select(6)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
}

func TestRussianPoker_Select_Success(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	require.NoError(t, rp.Buy6th())
	assert.Len(t, rp.GetPlayerHand(), 6)

	require.NoError(t, rp.Select(2))
	assert.Equal(t, domain.RussianPokerPhasePostAction, rp.GetPhase())
	assert.Len(t, rp.GetPlayerHand(), 5)
}

// --- Play ---

func TestRussianPoker_Play_WrongPhase(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	err := rp.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRussianPoker_Play_InsufficientChips(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	rp.SetChips(0)
	err := rp.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestRussianPoker_Play_FromAction(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignSpade, 13},
		cd{domain.CardDesignSpade, 12}, cd{domain.CardDesignSpade, 11},
		cd{domain.CardDesignSpade, 10},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignHeart, 1}, cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignClover, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignClover, 9},
	))
	require.NoError(t, rp.Play())
	assert.Equal(t, 200, rp.GetPlayBet())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.True(t, rp.GetGameEndFlag())
}

func TestRussianPoker_Play_FromPostAction(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	require.NoError(t, rp.Exchange([]int{0}))
	assert.Equal(t, domain.RussianPokerPhasePostAction, rp.GetPhase())

	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignSpade, 13},
		cd{domain.CardDesignSpade, 12}, cd{domain.CardDesignSpade, 11},
		cd{domain.CardDesignSpade, 10},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignHeart, 1}, cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignClover, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignClover, 9},
	))
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.True(t, rp.GetGameEndFlag())
}

// --- Fold ---

func TestRussianPoker_Fold_WrongPhase(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	err := rp.Fold()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRussianPoker_Fold_FromAction(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	chipsBefore := rp.GetChips()
	require.NoError(t, rp.Fold())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.True(t, rp.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, rp.GetResult())
	assert.Equal(t, chipsBefore, rp.GetChips())
}

func TestRussianPoker_Fold_FromPostAction(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	require.NoError(t, rp.Exchange([]int{0}))
	require.NoError(t, rp.Fold())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.True(t, rp.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, rp.GetResult())
}

// --- Dealer Qualification ---

func TestRussianPoker_DealerQualification(t *testing.T) {
	tests := []struct {
		name      string
		dealer    []*domain.Card
		qualified bool
	}{
		{
			name: "PairQualifies",
			dealer: makeHand(
				cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 2},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 9}),
			qualified: true,
		},
		{
			name: "AceKingQualifies",
			dealer: makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 13},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 9}),
			qualified: true,
		},
		{
			name: "AceQueenDoesNotQualify",
			dealer: makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 12},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 9}),
			qualified: false,
		},
		{
			name: "KingHighDoesNotQualify",
			dealer: makeHand(
				cd{domain.CardDesignSpade, 13}, cd{domain.CardDesignClover, 12},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 9}),
			qualified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := domain.NewDefaultRussianPoker()
			rp.SetChips(10000)
			require.NoError(t, rp.Bet(100))
			rp.SetPlayerHand(makeHand(
				cd{domain.CardDesignHeart, 1}, cd{domain.CardDesignHeart, 13},
				cd{domain.CardDesignHeart, 12}, cd{domain.CardDesignHeart, 11},
				cd{domain.CardDesignHeart, 10},
			))
			rp.SetDealerHand(tt.dealer)
			require.NoError(t, rp.Play())
			if tt.qualified {
				assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
				assert.True(t, rp.GetDealerQualified())
			} else {
				assert.Equal(t, domain.RussianPokerPhaseForceQualify, rp.GetPhase())
				assert.False(t, rp.GetDealerQualified())
			}
		})
	}
}

// --- ForceExchange ---

func TestRussianPoker_ForceExchange_WrongPhase(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	err := rp.ForceExchange()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRussianPoker_ForceExchange_InsufficientChips(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignHeart, 1}, cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignHeart, 12}, cd{domain.CardDesignHeart, 11},
		cd{domain.CardDesignHeart, 10},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 13}, cd{domain.CardDesignClover, 12},
		cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9}),
	)
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseForceQualify, rp.GetPhase())
	rp.SetChips(0)
	err := rp.ForceExchange()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestRussianPoker_ForceExchange_Success(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignHeart, 1}, cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignHeart, 12}, cd{domain.CardDesignHeart, 11},
		cd{domain.CardDesignHeart, 10},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 13}, cd{domain.CardDesignClover, 12},
		cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9}),
	)
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseForceQualify, rp.GetPhase())

	require.NoError(t, rp.ForceExchange())
	assert.True(t, rp.GetForceExchanged())
	assert.Equal(t, 100, rp.GetForceExchangeFee())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.True(t, rp.GetGameEndFlag())
}

// --- Decline ---

func TestRussianPoker_Decline_WrongPhase(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	err := rp.Decline()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRussianPoker_Decline_Success(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignHeart, 1}, cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignHeart, 12}, cd{domain.CardDesignHeart, 11},
		cd{domain.CardDesignHeart, 10},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 13}, cd{domain.CardDesignClover, 12},
		cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9}),
	)
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseForceQualify, rp.GetPhase())

	chipsBefore := rp.GetChips()
	require.NoError(t, rp.Decline())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.True(t, rp.GetGameEndFlag())
	// Decline: ante 1:1 + play returned
	assert.Equal(t, chipsBefore+100*2+200, rp.GetChips())
}

// --- Payouts ---

func TestRussianPoker_Payouts_PlayerWins_Qualified(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	// Player: Full House (3+2)
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 10}, cd{domain.CardDesignClover, 10},
		cd{domain.CardDesignHeart, 10}, cd{domain.CardDesignDiamond, 5},
		cd{domain.CardDesignSpade, 5},
	))
	// Dealer: Pair of aces (qualifies)
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 1},
		cd{domain.CardDesignHeart, 3}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignClover, 9},
	))
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.Equal(t, domain.GameResultWin, rp.GetResult())
	assert.True(t, rp.GetDealerQualified())
	// ante: 100*2=200, play: 200 + 200*7=1600
	assert.Equal(t, 200, rp.GetAntePayout())
	assert.Equal(t, 1600, rp.GetPlayPayout())
}

func TestRussianPoker_Payouts_DealerWins_Qualified(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	// Player: High card
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 9},
		cd{domain.CardDesignSpade, 11},
	))
	// Dealer: Pair of aces
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 1},
		cd{domain.CardDesignHeart, 3}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignClover, 9},
	))
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.Equal(t, domain.GameResultLose, rp.GetResult())
	assert.Equal(t, 0, rp.GetAntePayout())
	assert.Equal(t, 0, rp.GetPlayPayout())
}

func TestRussianPoker_Payouts_Push(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	hand := makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 13},
		cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9},
	)
	hand2 := makeHand(
		cd{domain.CardDesignHeart, 1}, cd{domain.CardDesignDiamond, 13},
		cd{domain.CardDesignClover, 5}, cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignHeart, 9},
	)
	rp.SetPlayerHand(hand)
	rp.SetDealerHand(hand2)
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.GameResultDraw, rp.GetResult())
	assert.Equal(t, 100, rp.GetAntePayout())
	assert.Equal(t, 200, rp.GetPlayPayout())
}

func TestRussianPoker_Payouts_DealerNotQualified_Decline(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 9},
		cd{domain.CardDesignSpade, 11},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 13}, cd{domain.CardDesignClover, 12},
		cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9}),
	)
	require.NoError(t, rp.Play())
	require.Equal(t, domain.RussianPokerPhaseForceQualify, rp.GetPhase())
	require.NoError(t, rp.Decline())
	// ante 1:1 returned + play returned
	assert.Equal(t, 200, rp.GetAntePayout())
	assert.Equal(t, 200, rp.GetPlayPayout())
}

// --- Play multiplier ---

func TestRussianPoker_PlayMultiplier(t *testing.T) {
	tests := []struct {
		name       string
		player     []*domain.Card
		multiplier int
	}{
		{
			name: "RoyalFlush",
			player: makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignSpade, 13},
				cd{domain.CardDesignSpade, 12}, cd{domain.CardDesignSpade, 11},
				cd{domain.CardDesignSpade, 10}),
			multiplier: domain.RussianPokerPayRoyalFlush,
		},
		{
			name: "StraightFlush",
			player: makeHand(
				cd{domain.CardDesignSpade, 9}, cd{domain.CardDesignSpade, 8},
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignSpade, 6},
				cd{domain.CardDesignSpade, 5}),
			multiplier: domain.RussianPokerPayStraightFlush,
		},
		{
			name: "FourOfAKind",
			player: makeHand(
				cd{domain.CardDesignSpade, 5}, cd{domain.CardDesignClover, 5},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 5},
				cd{domain.CardDesignSpade, 9}),
			multiplier: domain.RussianPokerPayFourOfAKind,
		},
		{
			name: "FullHouse",
			player: makeHand(
				cd{domain.CardDesignSpade, 10}, cd{domain.CardDesignClover, 10},
				cd{domain.CardDesignHeart, 10}, cd{domain.CardDesignDiamond, 5},
				cd{domain.CardDesignSpade, 5}),
			multiplier: domain.RussianPokerPayFullHouse,
		},
		{
			name: "Flush",
			player: makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignSpade, 10},
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignSpade, 5},
				cd{domain.CardDesignSpade, 3}),
			multiplier: domain.RussianPokerPayFlush,
		},
		{
			name: "Straight",
			player: makeHand(
				cd{domain.CardDesignSpade, 9}, cd{domain.CardDesignClover, 8},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 6},
				cd{domain.CardDesignSpade, 5}),
			multiplier: domain.RussianPokerPayStraight,
		},
		{
			name: "ThreeOfAKind",
			player: makeHand(
				cd{domain.CardDesignSpade, 8}, cd{domain.CardDesignClover, 8},
				cd{domain.CardDesignHeart, 8}, cd{domain.CardDesignDiamond, 5},
				cd{domain.CardDesignSpade, 3}),
			multiplier: domain.RussianPokerPayThreeOfAKind,
		},
		{
			name: "TwoPair",
			player: makeHand(
				cd{domain.CardDesignSpade, 10}, cd{domain.CardDesignClover, 10},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 5},
				cd{domain.CardDesignSpade, 3}),
			multiplier: domain.RussianPokerPayTwoPair,
		},
		{
			name: "Pair",
			player: makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 1},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 9}),
			multiplier: domain.RussianPokerPayPair,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := domain.NewDefaultRussianPoker()
			rp.SetChips(10000)
			require.NoError(t, rp.Bet(100))
			rp.SetPlayerHand(tt.player)
			rp.SetDealerHand(makeHand(
				cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 2},
				cd{domain.CardDesignHeart, 3}, cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignClover, 6},
			))
			require.NoError(t, rp.Play())
			// Play payout = playBet + playBet * multiplier
			assert.Equal(t, 200+200*tt.multiplier, rp.GetPlayPayout())
		})
	}
}

// --- Getters ---

func TestRussianPoker_GetTotalPayout(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 1},
		cd{domain.CardDesignHeart, 3}, cd{domain.CardDesignDiamond, 4},
		cd{domain.CardDesignClover, 6},
	))
	require.NoError(t, rp.Play())
	assert.Equal(t, rp.GetAntePayout()+rp.GetPlayPayout(), rp.GetTotalPayout())
}

// --- JSON round-trip ---

func TestRussianPoker_JSON_RoundTrip(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))

	data, err := json.Marshal(rp)
	require.NoError(t, err)

	var restored domain.RussianPoker
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, rp.GetPhase(), restored.GetPhase())
	assert.Equal(t, rp.GetChips(), restored.GetChips())
	assert.Equal(t, rp.GetAnteBet(), restored.GetAnteBet())
	assert.Len(t, restored.GetPlayerHand(), 5)
	assert.Len(t, restored.GetDealerHand(), 5)
}

func TestRussianPoker_JSON_OversizeInput(t *testing.T) {
	huge := make([]*domain.Card, 1001)
	for i := range huge {
		huge[i] = domain.NewCard(domain.CardDesignSpade, 1, false)
	}
	data, _ := json.Marshal(struct {
		PlayerHand []*domain.Card `json:"ph"`
	}{PlayerHand: huge})
	var rp domain.RussianPoker
	assert.Error(t, json.Unmarshal(data, &rp))
}

func TestRussianPoker_JSON_NilFields(t *testing.T) {
	data := []byte(`{"ps":1}`)
	var rp domain.RussianPoker
	require.NoError(t, json.Unmarshal(data, &rp))
	assert.NotNil(t, rp.GetPlayerHand())
	assert.NotNil(t, rp.GetDealerHand())
	assert.NotNil(t, rp.GetActionLog())
}

// --- Full game flow ---

func TestRussianPoker_FullFlow_PlayDirect(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))

	// Set deterministic hands
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 1},
		cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignHeart, 3}, cd{domain.CardDesignDiamond, 4},
		cd{domain.CardDesignClover, 6},
	))
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.True(t, rp.GetGameEndFlag())
	assert.True(t, rp.GetDealerQualified())
	assert.Equal(t, domain.GameResultWin, rp.GetResult())
}

func TestRussianPoker_FullFlow_ExchangeThenPlay(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	require.NoError(t, rp.Exchange([]int{0}))
	assert.Equal(t, domain.RussianPokerPhasePostAction, rp.GetPhase())

	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 1},
		cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignHeart, 3}, cd{domain.CardDesignDiamond, 4},
		cd{domain.CardDesignClover, 6},
	))
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
}

func TestRussianPoker_FullFlow_Buy6thSelectThenPlay(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	require.NoError(t, rp.Buy6th())
	assert.Equal(t, domain.RussianPokerPhaseSelect, rp.GetPhase())
	assert.Len(t, rp.GetPlayerHand(), 6)

	require.NoError(t, rp.Select(0))
	assert.Equal(t, domain.RussianPokerPhasePostAction, rp.GetPhase())
	assert.Len(t, rp.GetPlayerHand(), 5)

	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 1},
		cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignHeart, 3}, cd{domain.CardDesignDiamond, 4},
		cd{domain.CardDesignClover, 6},
	))
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
}

func TestRussianPoker_FullFlow_ForceExchangeRoute(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	rp.SetChips(10000)
	require.NoError(t, rp.Bet(100))
	rp.SetPlayerHand(makeHand(
		cd{domain.CardDesignHeart, 1}, cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignHeart, 12}, cd{domain.CardDesignHeart, 11},
		cd{domain.CardDesignHeart, 10},
	))
	rp.SetDealerHand(makeHand(
		cd{domain.CardDesignSpade, 13}, cd{domain.CardDesignClover, 12},
		cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignDiamond, 7},
		cd{domain.CardDesignSpade, 9}),
	)
	require.NoError(t, rp.Play())
	assert.Equal(t, domain.RussianPokerPhaseForceQualify, rp.GetPhase())

	require.NoError(t, rp.ForceExchange())
	assert.Equal(t, domain.RussianPokerPhaseEnd, rp.GetPhase())
	assert.True(t, rp.GetGameEndFlag())
	assert.True(t, rp.GetForceExchanged())
}

// --- ActionLog ---

func TestRussianPoker_ActionLog(t *testing.T) {
	rp := domain.NewDefaultRussianPoker()
	require.NoError(t, rp.Bet(100))
	log := rp.GetActionLog()
	assert.NotEmpty(t, log)
}
