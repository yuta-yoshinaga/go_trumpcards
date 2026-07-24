package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// Note: makeHand and cd are declared in CaribbeanStud_test.go (same package).

func TestNewDefaultOasisPoker(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	assert.Equal(t, domain.OasisPokerPhaseBet, op.GetPhase())
	assert.Equal(t, domain.OasisPokerDefaultChips, op.GetChips())
	assert.False(t, op.GetGameEndFlag())
	assert.Nil(t, op.GetPlayerHand())
	assert.Nil(t, op.GetDealerHand())
}

func TestOasisPoker_Reset(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Stand())
	require.NoError(t, op.Play())
	assert.Equal(t, domain.OasisPokerPhaseEnd, op.GetPhase())

	op.Reset()
	assert.Equal(t, domain.OasisPokerPhaseBet, op.GetPhase())
	assert.False(t, op.GetGameEndFlag())
	assert.Nil(t, op.GetPlayerHand())
	assert.Nil(t, op.GetDealerHand())
	assert.Equal(t, 0, op.GetAnteBet())
	assert.Equal(t, 0, op.GetJackpotBet())
	assert.Equal(t, 0, op.GetPlayBet())
	assert.Equal(t, 0, op.GetExchangeCount())
	assert.Equal(t, 0, op.GetExchangeFee())
}

func TestOasisPoker_Reset_RefillChips(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	op.SetChips(5)
	op.Reset()
	assert.Equal(t, domain.OasisPokerDefaultChips, op.GetChips())
}

func TestOasisPoker_Bet_WrongPhase(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	op.SetPhase(domain.OasisPokerPhaseAction)
	err := op.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestOasisPoker_Bet_InvalidAnteAmount(t *testing.T) {
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
			op := domain.NewDefaultOasisPoker()
			err := op.Bet(tt.ante, 0)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestOasisPoker_Bet_InvalidJackpotAmount(t *testing.T) {
	tests := []struct {
		name    string
		jackpot int
	}{
		{"Negative", -10},
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := domain.NewDefaultOasisPoker()
			err := op.Bet(100, tt.jackpot)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestOasisPoker_Bet_InsufficientChips(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	op.SetChips(50)
	err := op.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestOasisPoker_Bet_Success(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	err := op.Bet(100, 50)
	assert.NoError(t, err)
	assert.Equal(t, domain.OasisPokerPhaseExchange, op.GetPhase())
	assert.Equal(t, 100, op.GetAnteBet())
	assert.Equal(t, 50, op.GetJackpotBet())
	assert.Len(t, op.GetPlayerHand(), 5)
	assert.Len(t, op.GetDealerHand(), 5)
	assert.Equal(t, domain.OasisPokerDefaultChips-150, op.GetChips())
}

func TestOasisPoker_Exchange_WrongPhase(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	err := op.Exchange([]int{0})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestOasisPoker_Exchange_TooManyIndices(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	err := op.Exchange([]int{0, 1, 2, 3, 4, 0})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
}

func TestOasisPoker_Exchange_OutOfRange(t *testing.T) {
	for _, idx := range []int{-1, 5, 99} {
		op := domain.NewDefaultOasisPoker()
		require.NoError(t, op.Bet(100, 0))
		err := op.Exchange([]int{idx})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	}
}

func TestOasisPoker_Exchange_Duplicate(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	err := op.Exchange([]int{1, 1})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
}

func TestOasisPoker_Exchange_InsufficientFee(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	// After Bet: chips = 1000 - 100 = 900. Force chips to 50 — can't afford fee 100.
	op.SetChips(50)
	err := op.Exchange([]int{0})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestOasisPoker_Exchange_DeductsFee(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	chipsBefore := op.GetChips()
	require.NoError(t, op.Exchange([]int{0, 1, 2}))
	assert.Equal(t, domain.OasisPokerPhaseAction, op.GetPhase())
	assert.Equal(t, 3, op.GetExchangeCount())
	assert.Equal(t, 300, op.GetExchangeFee())
	assert.Equal(t, chipsBefore-300, op.GetChips())
}

func TestOasisPoker_Stand_NoFee(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	chipsBefore := op.GetChips()
	require.NoError(t, op.Stand())
	assert.Equal(t, domain.OasisPokerPhaseAction, op.GetPhase())
	assert.Equal(t, 0, op.GetExchangeCount())
	assert.Equal(t, 0, op.GetExchangeFee())
	assert.Equal(t, chipsBefore, op.GetChips())
}

func TestOasisPoker_Exchange_ReplacesCards(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	originalHand := append([]*domain.Card(nil), op.GetPlayerHand()...)
	require.NoError(t, op.Exchange([]int{0, 4}))
	newHand := op.GetPlayerHand()
	// 0 and 4 should have been redrawn; 1,2,3 must be untouched.
	assert.Same(t, originalHand[1], newHand[1])
	assert.Same(t, originalHand[2], newHand[2])
	assert.Same(t, originalHand[3], newHand[3])
}

func TestOasisPoker_Play_WrongPhase(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	err := op.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestOasisPoker_Play_InsufficientChips(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Stand())
	op.SetChips(0)
	err := op.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestOasisPoker_Play_PlacesDoubleAnte(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Stand())
	require.NoError(t, op.Play())
	assert.Equal(t, 200, op.GetPlayBet())
	assert.Equal(t, domain.OasisPokerPhaseEnd, op.GetPhase())
	assert.True(t, op.GetGameEndFlag())
}

func TestOasisPoker_Fold_WrongPhase(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	err := op.Fold()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestOasisPoker_Fold_Success(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Stand())
	chipsBefore := op.GetChips()
	require.NoError(t, op.Fold())
	assert.Equal(t, domain.OasisPokerPhaseEnd, op.GetPhase())
	assert.True(t, op.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, op.GetResult())
	assert.Equal(t, chipsBefore, op.GetChips())
}

func TestOasisPoker_Fold_JackpotStillEvaluated(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 10))
	require.NoError(t, op.Stand())

	op.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignSpade, 11},
	))
	chipsBefore := op.GetChips()
	require.NoError(t, op.Fold())
	assert.Equal(t, 10*domain.OasisPokerJackpotFlush, op.GetJackpotPayout())
	assert.Equal(t, chipsBefore+10*domain.OasisPokerJackpotFlush, op.GetChips())
}

func TestOasisPoker_DealerQualification(t *testing.T) {
	tests := []struct {
		name      string
		dealer    []*domain.Card
		qualified bool
	}{
		{
			name: "PairOfTwosQualifies",
			dealer: makeHand(
				cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignClover, 2},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 9}),
			qualified: true,
		},
		{
			name: "AceKingHighQualifies",
			dealer: makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 13},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 9}),
			qualified: true,
		},
		{
			name: "AceQueenHighDoesNotQualify",
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
			op := domain.NewDefaultOasisPoker()
			require.NoError(t, op.Bet(100, 0))
			require.NoError(t, op.Stand())
			op.SetPlayerHand(makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 4},
				cd{domain.CardDesignHeart, 6}, cd{domain.CardDesignDiamond, 8},
				cd{domain.CardDesignSpade, 10}))
			op.SetDealerHand(tt.dealer)
			require.NoError(t, op.Play())
			assert.Equal(t, tt.qualified, op.GetDealerQualified())
		})
	}
}

func TestOasisPoker_Payouts_DealerNotQualified(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Stand())

	op.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignHeart, 6}, cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 10}))
	op.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 12}, cd{domain.CardDesignHeart, 5},
		cd{domain.CardDesignClover, 3}, cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignDiamond, 9}))

	require.NoError(t, op.Play())
	assert.False(t, op.GetDealerQualified())
	assert.Equal(t, 200, op.GetAntePayout())
	assert.Equal(t, 200, op.GetPlayPayout())
}

func TestOasisPoker_Payouts_PlayerWinsFlush(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Stand())

	op.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignSpade, 11}))
	op.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 4}, cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
		cd{domain.CardDesignDiamond, 10}))

	require.NoError(t, op.Play())
	assert.True(t, op.GetDealerQualified())
	assert.Equal(t, domain.GameResultWin, op.GetResult())
	assert.Equal(t, 200, op.GetAntePayout())
	assert.Equal(t, 1200, op.GetPlayPayout())
}

func TestOasisPoker_Payouts_PlayerLoses(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Stand())

	op.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignHeart, 6}, cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 10}))
	op.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 4}, cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
		cd{domain.CardDesignDiamond, 11}))

	require.NoError(t, op.Play())
	assert.True(t, op.GetDealerQualified())
	assert.Equal(t, domain.GameResultLose, op.GetResult())
	assert.Equal(t, 0, op.GetAntePayout())
	assert.Equal(t, 0, op.GetPlayPayout())
}

func TestOasisPoker_Payouts_Push(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Stand())

	op.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 5}, cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 9},
		cd{domain.CardDesignSpade, 11}))
	op.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignHeart, 5},
		cd{domain.CardDesignClover, 7}, cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignDiamond, 11}))

	require.NoError(t, op.Play())
	assert.True(t, op.GetDealerQualified())
	assert.Equal(t, domain.GameResultDraw, op.GetResult())
	assert.Equal(t, 100, op.GetAntePayout())
	assert.Equal(t, 200, op.GetPlayPayout())
}

func TestOasisPoker_PlayMultipliers(t *testing.T) {
	tests := []struct {
		name       string
		hand       []*domain.Card
		multiplier int
	}{
		{
			name: "RoyalFlush",
			hand: makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignSpade, 10},
				cd{domain.CardDesignSpade, 11}, cd{domain.CardDesignSpade, 12},
				cd{domain.CardDesignSpade, 13}),
			multiplier: domain.OasisPokerPayRoyalFlush,
		},
		{
			name: "StraightFlush",
			hand: makeHand(
				cd{domain.CardDesignClover, 5}, cd{domain.CardDesignClover, 6},
				cd{domain.CardDesignClover, 7}, cd{domain.CardDesignClover, 8},
				cd{domain.CardDesignClover, 9}),
			multiplier: domain.OasisPokerPayStraightFlush,
		},
		{
			name: "FourOfAKind",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.OasisPokerPayFourOfAKind,
		},
		{
			name: "FullHouse",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 2},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.OasisPokerPayFullHouse,
		},
		{
			name: "Straight",
			hand: makeHand(
				cd{domain.CardDesignSpade, 5}, cd{domain.CardDesignClover, 6},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 8},
				cd{domain.CardDesignClover, 9}),
			multiplier: domain.OasisPokerPayStraight,
		},
		{
			name: "ThreeOfAKind",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.OasisPokerPayThreeOfAKind,
		},
		{
			name: "TwoPair",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 4}, cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.OasisPokerPayTwoPair,
		},
		{
			name: "OnePair",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.OasisPokerPayPair,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := domain.NewDefaultOasisPoker()
			op.SetChips(100000)
			require.NoError(t, op.Bet(100, 0))
			require.NoError(t, op.Stand())
			op.SetPlayerHand(tt.hand)
			op.SetDealerHand(makeHand(
				cd{domain.CardDesignDiamond, 3}, cd{domain.CardDesignHeart, 3},
				cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
				cd{domain.CardDesignDiamond, 10}))
			require.NoError(t, op.Play())
			assert.Equal(t, domain.GameResultWin, op.GetResult())
			expected := 200 + 200*tt.multiplier
			assert.Equal(t, expected, op.GetPlayPayout())
		})
	}
}

func TestOasisPoker_JackpotPayouts(t *testing.T) {
	tests := []struct {
		name       string
		hand       []*domain.Card
		multiplier int
	}{
		{
			name: "RoyalFlushJackpot",
			hand: makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignSpade, 10},
				cd{domain.CardDesignSpade, 11}, cd{domain.CardDesignSpade, 12},
				cd{domain.CardDesignSpade, 13}),
			multiplier: domain.OasisPokerJackpotRoyalFlush,
		},
		{
			name: "StraightFlushJackpot",
			hand: makeHand(
				cd{domain.CardDesignClover, 5}, cd{domain.CardDesignClover, 6},
				cd{domain.CardDesignClover, 7}, cd{domain.CardDesignClover, 8},
				cd{domain.CardDesignClover, 9}),
			multiplier: domain.OasisPokerJackpotStraightFlush,
		},
		{
			name: "FourOfAKindJackpot",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.OasisPokerJackpotFourOfAKind,
		},
		{
			name: "FullHouseJackpot",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 2},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.OasisPokerJackpotFullHouse,
		},
		{
			name: "FlushJackpot",
			hand: makeHand(
				cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignSpade, 5},
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignSpade, 9},
				cd{domain.CardDesignSpade, 11}),
			multiplier: domain.OasisPokerJackpotFlush,
		},
		{
			name: "NoJackpotPair",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2}),
			multiplier: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := domain.NewDefaultOasisPoker()
			op.SetChips(100000)
			require.NoError(t, op.Bet(100, 10))
			require.NoError(t, op.Stand())
			op.SetPlayerHand(tt.hand)
			op.SetDealerHand(makeHand(
				cd{domain.CardDesignDiamond, 3}, cd{domain.CardDesignHeart, 3},
				cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
				cd{domain.CardDesignDiamond, 10}))
			require.NoError(t, op.Play())
			assert.Equal(t, 10*tt.multiplier, op.GetJackpotPayout())
		})
	}
}

func TestOasisPoker_TotalPayoutZero(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	assert.Equal(t, 0, op.GetTotalPayout())
}

func TestOasisPoker_GetActionLog(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 0))
	require.NoError(t, op.Exchange([]int{0}))
	require.NoError(t, op.Play())
	assert.NotEmpty(t, op.GetActionLog())
}

func TestOasisPoker_JSONRoundTrip(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	require.NoError(t, op.Bet(100, 10))
	require.NoError(t, op.Exchange([]int{2}))
	require.NoError(t, op.Play())

	data, err := json.Marshal(op)
	require.NoError(t, err)

	var restored domain.OasisPoker
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, op.GetPhase(), restored.GetPhase())
	assert.Equal(t, op.GetChips(), restored.GetChips())
	assert.Equal(t, op.GetAnteBet(), restored.GetAnteBet())
	assert.Equal(t, op.GetJackpotBet(), restored.GetJackpotBet())
	assert.Equal(t, op.GetPlayBet(), restored.GetPlayBet())
	assert.Equal(t, op.GetExchangeCount(), restored.GetExchangeCount())
	assert.Equal(t, op.GetExchangeFee(), restored.GetExchangeFee())
	assert.Equal(t, op.GetResult(), restored.GetResult())
	assert.Equal(t, op.GetTotalPayout(), restored.GetTotalPayout())
	assert.Equal(t, op.GetPlayerHandRank(), restored.GetPlayerHandRank())
	assert.Equal(t, op.GetDealerHandRank(), restored.GetDealerHandRank())
}

func TestOasisPoker_UnmarshalJSON_InvalidData(t *testing.T) {
	var op domain.OasisPoker
	err := op.UnmarshalJSON([]byte("not json"))
	assert.Error(t, err)
}

func TestOasisPoker_SetSettersExposeFields(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	op.SetPlayerHand(makeHand(cd{domain.CardDesignSpade, 1}))
	op.SetDealerHand(makeHand(cd{domain.CardDesignClover, 13}))
	op.SetAnteBet(100)
	op.SetJackpotBet(20)
	op.SetPlayBet(200)
	assert.Len(t, op.GetPlayerHand(), 1)
	assert.Len(t, op.GetDealerHand(), 1)
	assert.Equal(t, 100, op.GetAnteBet())
	assert.Equal(t, 20, op.GetJackpotBet())
	assert.Equal(t, 200, op.GetPlayBet())
}

func TestOasisPoker_RecommendPlay(t *testing.T) {
	t.Run("play with a pair", func(t *testing.T) {
		op := domain.NewDefaultOasisPoker()
		op.SetPlayerHand(makeHand(
			cd{domain.CardDesignSpade, 8}, cd{domain.CardDesignHeart, 8},
			cd{domain.CardDesignClover, 3}, cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignSpade, 9},
		))
		assert.True(t, op.RecommendPlay())
	})
	t.Run("play with an ace high", func(t *testing.T) {
		op := domain.NewDefaultOasisPoker()
		op.SetPlayerHand(makeHand(
			cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignHeart, 7},
			cd{domain.CardDesignClover, 3}, cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignSpade, 9},
		))
		assert.True(t, op.RecommendPlay())
	})
	t.Run("fold with junk", func(t *testing.T) {
		op := domain.NewDefaultOasisPoker()
		op.SetPlayerHand(makeHand(
			cd{domain.CardDesignSpade, 3}, cd{domain.CardDesignHeart, 7},
			cd{domain.CardDesignClover, 9}, cd{domain.CardDesignDiamond, 11}, cd{domain.CardDesignSpade, 5},
		))
		assert.False(t, op.RecommendPlay())
	})
}
