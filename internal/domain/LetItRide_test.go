package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewDefaultLetItRide(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	assert.Equal(t, domain.LetItRidePhaseBet, lir.GetPhase())
	assert.Equal(t, domain.LetItRideDefaultChips, lir.GetChips())
	assert.False(t, lir.GetGameEndFlag())
	assert.Nil(t, lir.GetPlayerHand())
	assert.Nil(t, lir.GetCommunityCards())
}

func TestLetItRide_Reset(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))
	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())
	assert.Equal(t, domain.LetItRidePhaseEnd, lir.GetPhase())

	lir.Reset()
	assert.Equal(t, domain.LetItRidePhaseBet, lir.GetPhase())
	assert.False(t, lir.GetGameEndFlag())
	assert.Nil(t, lir.GetPlayerHand())
	assert.Nil(t, lir.GetCommunityCards())
	assert.Equal(t, 0, lir.GetBetAmount())
	assert.False(t, lir.GetBet1Active())
	assert.False(t, lir.GetBet2Active())
	assert.False(t, lir.GetBet3Active())
}

func TestLetItRide_Reset_RefillChips(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	lir.SetChips(20) // below MinBet * 3 = 30
	lir.Reset()
	assert.Equal(t, domain.LetItRideDefaultChips, lir.GetChips())
}

func TestLetItRide_Reset_NoRefillAboveThreshold(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	lir.SetChips(500)
	lir.Reset()
	assert.Equal(t, 500, lir.GetChips())
}

func TestLetItRide_Bet_WrongPhase(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	lir.SetPhase(domain.LetItRidePhaseFirstDecision)
	err := lir.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestLetItRide_Bet_InvalidAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int
	}{
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
		{"Zero", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lir := domain.NewDefaultLetItRide()
			err := lir.Bet(tt.amount)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestLetItRide_Bet_InsufficientChips(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	lir.SetChips(200) // needs 300 (100 * 3)
	err := lir.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestLetItRide_Bet_Success(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	err := lir.Bet(100)
	assert.NoError(t, err)
	assert.Equal(t, domain.LetItRidePhaseFirstDecision, lir.GetPhase())
	assert.Equal(t, 100, lir.GetBetAmount())
	assert.True(t, lir.GetBet1Active())
	assert.True(t, lir.GetBet2Active())
	assert.True(t, lir.GetBet3Active())
	assert.Len(t, lir.GetPlayerHand(), 3)
	assert.Len(t, lir.GetCommunityCards(), 2)
	assert.Equal(t, domain.LetItRideDefaultChips-300, lir.GetChips())
}

func TestLetItRide_Pull_WrongPhase(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	err := lir.Pull()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestLetItRide_Pull_FirstDecision(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))
	chipsBefore := lir.GetChips()
	require.NoError(t, lir.Pull())

	assert.Equal(t, domain.LetItRidePhaseSecondDecision, lir.GetPhase())
	assert.False(t, lir.GetBet3Active())
	assert.True(t, lir.GetBet2Active())
	assert.True(t, lir.GetBet1Active())
	assert.Equal(t, chipsBefore+100, lir.GetChips()) // bet3 refunded
}

func TestLetItRide_Pull_SecondDecision(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))
	require.NoError(t, lir.LetItRideAction()) // first decision: let it ride
	chipsBefore := lir.GetChips()
	require.NoError(t, lir.Pull()) // second decision: pull bet 2

	assert.Equal(t, domain.LetItRidePhaseEnd, lir.GetPhase())
	assert.True(t, lir.GetGameEndFlag())
	assert.False(t, lir.GetBet2Active())
	assert.True(t, lir.GetBet3Active()) // was not pulled in first decision
	// Chips change depends on hand result + bet2 refund
	_ = chipsBefore
}

func TestLetItRide_LetItRide_WrongPhase(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	err := lir.LetItRideAction()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestLetItRide_LetItRide_FirstDecision(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))
	chipsBefore := lir.GetChips()
	require.NoError(t, lir.LetItRideAction())

	assert.Equal(t, domain.LetItRidePhaseSecondDecision, lir.GetPhase())
	assert.True(t, lir.GetBet3Active()) // stayed
	assert.Equal(t, chipsBefore, lir.GetChips())
}

func TestLetItRide_LetItRide_SecondDecision(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))
	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())

	assert.Equal(t, domain.LetItRidePhaseEnd, lir.GetPhase())
	assert.True(t, lir.GetGameEndFlag())
	assert.True(t, lir.GetBet2Active()) // stayed
}

func TestLetItRide_Pull_EndPhase(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	lir.SetPhase(domain.LetItRidePhaseEnd)
	err := lir.Pull()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestLetItRide_LetItRide_EndPhase(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	lir.SetPhase(domain.LetItRidePhaseEnd)
	err := lir.LetItRideAction()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestLetItRide_Payout_AllBetsActive_TensOrBetter(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))

	// Player hand: 10, 10 → pair of tens (+ community doesn't help)
	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 10},
		cd{domain.CardDesignClover, 10},
		cd{domain.CardDesignHeart, 3},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignDiamond, 5},
		cd{domain.CardDesignSpade, 7},
	))

	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())

	assert.Equal(t, domain.GameResultWin, lir.GetResult())
	assert.Equal(t, domain.PokerHandOnePair, lir.GetHandRank())
	// Each active bet pays 1:1 → betAmount + betAmount*1 = 200 each
	assert.Equal(t, 200, lir.GetBet1Payout())
	assert.Equal(t, 200, lir.GetBet2Payout())
	assert.Equal(t, 200, lir.GetBet3Payout())
	assert.Equal(t, 600, lir.GetTotalPayout())
}

func TestLetItRide_Payout_PullBoth_TensOrBetter(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))

	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 10},
		cd{domain.CardDesignClover, 10},
		cd{domain.CardDesignHeart, 3},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignDiamond, 5},
		cd{domain.CardDesignSpade, 7},
	))

	require.NoError(t, lir.Pull()) // pull bet3
	require.NoError(t, lir.Pull()) // pull bet2

	assert.Equal(t, domain.GameResultWin, lir.GetResult())
	assert.Equal(t, 200, lir.GetBet1Payout()) // only bet1 active
	assert.Equal(t, 0, lir.GetBet2Payout())
	assert.Equal(t, 0, lir.GetBet3Payout())
	assert.Equal(t, 200, lir.GetTotalPayout())
}

func TestLetItRide_Payout_LowPair_NoPayment(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))

	// Pair of 5s → below tens, no payout
	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignHeart, 3},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 12},
	))

	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())

	assert.Equal(t, domain.GameResultLose, lir.GetResult())
	assert.Equal(t, 0, lir.GetTotalPayout())
}

func TestLetItRide_Payout_HighCard_NoPayment(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))

	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignHeart, 8},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignDiamond, 11},
		cd{domain.CardDesignSpade, 13},
	))

	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())

	assert.Equal(t, domain.GameResultLose, lir.GetResult())
	assert.Equal(t, 0, lir.GetTotalPayout())
}

func TestLetItRide_PayoutMultipliers(t *testing.T) {
	tests := []struct {
		name       string
		player     []*domain.Card
		community  []*domain.Card
		multiplier int
	}{
		{
			name: "RoyalFlush",
			player: makeHand(
				cd{domain.CardDesignSpade, 1},
				cd{domain.CardDesignSpade, 13},
				cd{domain.CardDesignSpade, 12},
			),
			community: makeHand(
				cd{domain.CardDesignSpade, 11},
				cd{domain.CardDesignSpade, 10},
			),
			multiplier: domain.LetItRidePayRoyalFlush,
		},
		{
			name: "StraightFlush",
			player: makeHand(
				cd{domain.CardDesignClover, 5},
				cd{domain.CardDesignClover, 6},
				cd{domain.CardDesignClover, 7},
			),
			community: makeHand(
				cd{domain.CardDesignClover, 8},
				cd{domain.CardDesignClover, 9},
			),
			multiplier: domain.LetItRidePayStraightFlush,
		},
		{
			name: "FourOfAKind",
			player: makeHand(
				cd{domain.CardDesignSpade, 7},
				cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7},
			),
			community: makeHand(
				cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 2},
			),
			multiplier: domain.LetItRidePayFourOfAKind,
		},
		{
			name: "FullHouse",
			player: makeHand(
				cd{domain.CardDesignSpade, 7},
				cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7},
			),
			community: makeHand(
				cd{domain.CardDesignDiamond, 2},
				cd{domain.CardDesignSpade, 2},
			),
			multiplier: domain.LetItRidePayFullHouse,
		},
		{
			name: "Flush",
			player: makeHand(
				cd{domain.CardDesignSpade, 2},
				cd{domain.CardDesignSpade, 5},
				cd{domain.CardDesignSpade, 7},
			),
			community: makeHand(
				cd{domain.CardDesignSpade, 9},
				cd{domain.CardDesignSpade, 11},
			),
			multiplier: domain.LetItRidePayFlush,
		},
		{
			name: "Straight",
			player: makeHand(
				cd{domain.CardDesignSpade, 5},
				cd{domain.CardDesignClover, 6},
				cd{domain.CardDesignHeart, 7},
			),
			community: makeHand(
				cd{domain.CardDesignDiamond, 8},
				cd{domain.CardDesignClover, 9},
			),
			multiplier: domain.LetItRidePayStraight,
		},
		{
			name: "ThreeOfAKind",
			player: makeHand(
				cd{domain.CardDesignSpade, 7},
				cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7},
			),
			community: makeHand(
				cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2},
			),
			multiplier: domain.LetItRidePayThreeOfAKind,
		},
		{
			name: "TwoPair",
			player: makeHand(
				cd{domain.CardDesignSpade, 7},
				cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 4},
			),
			community: makeHand(
				cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2},
			),
			multiplier: domain.LetItRidePayTwoPair,
		},
		{
			name: "TensOrBetter",
			player: makeHand(
				cd{domain.CardDesignSpade, 11},
				cd{domain.CardDesignClover, 11},
				cd{domain.CardDesignHeart, 4},
			),
			community: makeHand(
				cd{domain.CardDesignDiamond, 6},
				cd{domain.CardDesignSpade, 2},
			),
			multiplier: domain.LetItRidePayTensOrBetter,
		},
		{
			name: "AcePairPays",
			player: makeHand(
				cd{domain.CardDesignSpade, 1},
				cd{domain.CardDesignClover, 1},
				cd{domain.CardDesignHeart, 4},
			),
			community: makeHand(
				cd{domain.CardDesignDiamond, 6},
				cd{domain.CardDesignSpade, 8},
			),
			multiplier: domain.LetItRidePayTensOrBetter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lir := domain.NewDefaultLetItRide()
			lir.SetChips(100000)
			require.NoError(t, lir.Bet(100))
			lir.SetPlayerHand(tt.player)
			lir.SetCommunityCards(tt.community)
			require.NoError(t, lir.LetItRideAction())
			require.NoError(t, lir.LetItRideAction())

			assert.Equal(t, domain.GameResultWin, lir.GetResult())
			// All 3 bets active, each pays betAmount + betAmount*multiplier
			expectedEach := 100 + 100*tt.multiplier
			assert.Equal(t, expectedEach, lir.GetBet1Payout())
			assert.Equal(t, expectedEach, lir.GetBet2Payout())
			assert.Equal(t, expectedEach, lir.GetBet3Payout())
			assert.Equal(t, expectedEach*3, lir.GetTotalPayout())
		})
	}
}

func TestLetItRide_NinePairNoPayout(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))

	// Pair of 9s → below tens, no payout
	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignClover, 9},
		cd{domain.CardDesignHeart, 3},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignDiamond, 6},
		cd{domain.CardDesignSpade, 12},
	))

	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())

	assert.Equal(t, domain.GameResultLose, lir.GetResult())
	assert.Equal(t, 0, lir.GetTotalPayout())
}

func TestLetItRide_GetActionLog(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))
	require.NoError(t, lir.Pull())
	require.NoError(t, lir.LetItRideAction())
	assert.NotEmpty(t, lir.GetActionLog())
}

func TestLetItRide_JSONRoundTrip(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))
	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())

	data, err := json.Marshal(lir)
	require.NoError(t, err)

	var restored domain.LetItRide
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, lir.GetPhase(), restored.GetPhase())
	assert.Equal(t, lir.GetChips(), restored.GetChips())
	assert.Equal(t, lir.GetBetAmount(), restored.GetBetAmount())
	assert.Equal(t, lir.GetBet1Active(), restored.GetBet1Active())
	assert.Equal(t, lir.GetBet2Active(), restored.GetBet2Active())
	assert.Equal(t, lir.GetBet3Active(), restored.GetBet3Active())
	assert.Equal(t, lir.GetResult(), restored.GetResult())
	assert.Equal(t, lir.GetHandRank(), restored.GetHandRank())
	assert.Equal(t, lir.GetTotalPayout(), restored.GetTotalPayout())
}

func TestLetItRide_UnmarshalJSON_InvalidData(t *testing.T) {
	var lir domain.LetItRide
	err := lir.UnmarshalJSON([]byte("not json"))
	assert.Error(t, err)
}

func TestLetItRide_UnmarshalJSON_NilFields(t *testing.T) {
	// Minimal valid JSON with null fields
	data := []byte(`{"tc":null,"ph":null,"cc":null,"ch":null,"ba":0,"b1":false,"b2":false,"b3":false,"ps":1,"ge":false,"rs":0,"hr":0,"p1":0,"p2":0,"p3":0,"tp":0,"al":null}`)
	var lir domain.LetItRide
	require.NoError(t, json.Unmarshal(data, &lir))
	assert.NotNil(t, lir.GetPlayerHand())
	assert.NotNil(t, lir.GetCommunityCards())
	assert.NotNil(t, lir.GetActionLog())
}

func TestLetItRide_SettersExposeFields(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	lir.SetPlayerHand(makeHand(cd{domain.CardDesignSpade, 1}))
	lir.SetCommunityCards(makeHand(cd{domain.CardDesignClover, 13}))
	lir.SetBetAmount(100)
	lir.SetBet1Active(true)
	lir.SetBet2Active(false)
	lir.SetBet3Active(true)
	assert.Len(t, lir.GetPlayerHand(), 1)
	assert.Len(t, lir.GetCommunityCards(), 1)
	assert.Equal(t, 100, lir.GetBetAmount())
	assert.True(t, lir.GetBet1Active())
	assert.False(t, lir.GetBet2Active())
	assert.True(t, lir.GetBet3Active())
}

func TestLetItRide_ChipsAfterWinAllRide(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100)) // 1000 - 300 = 700
	assert.Equal(t, 700, lir.GetChips())

	// Two pair → multiplier 2
	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignClover, 7},
		cd{domain.CardDesignHeart, 4},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignDiamond, 4},
		cd{domain.CardDesignSpade, 2},
	))

	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())

	// Each bet pays 100 + 100*2 = 300, total 900
	assert.Equal(t, 900, lir.GetTotalPayout())
	assert.Equal(t, 700+900, lir.GetChips())
}

func TestLetItRide_ChipsAfterLossAllRide(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100)) // 1000 - 300 = 700

	// High card → no payout
	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignHeart, 8},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignDiamond, 11},
		cd{domain.CardDesignSpade, 13},
	))

	require.NoError(t, lir.LetItRideAction())
	require.NoError(t, lir.LetItRideAction())

	assert.Equal(t, 0, lir.GetTotalPayout())
	assert.Equal(t, 700, lir.GetChips())
}

func TestLetItRide_ChipsAfterPullBoth(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100)) // 1000 - 300 = 700

	// Two pair → win
	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignClover, 7},
		cd{domain.CardDesignHeart, 4},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignDiamond, 4},
		cd{domain.CardDesignSpade, 2},
	))

	require.NoError(t, lir.Pull()) // pull bet3 → +100 = 800
	assert.Equal(t, 800, lir.GetChips())
	require.NoError(t, lir.Pull()) // pull bet2 → +100 = 900

	// Only bet1 active, pays 100 + 100*2 = 300
	assert.Equal(t, 300, lir.GetBet1Payout())
	assert.Equal(t, 0, lir.GetBet2Payout())
	assert.Equal(t, 0, lir.GetBet3Payout())
	assert.Equal(t, 900+300, lir.GetChips())
}

func TestLetItRide_FullFlow_PullThenRide(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	require.NoError(t, lir.Bet(100))

	// Flush → multiplier 8
	lir.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignSpade, 7},
	))
	lir.SetCommunityCards(makeHand(
		cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignSpade, 11},
	))

	require.NoError(t, lir.Pull())            // pull bet3
	require.NoError(t, lir.LetItRideAction()) // keep bet2

	assert.Equal(t, domain.GameResultWin, lir.GetResult())
	// bet1 and bet2 active, each pays 100 + 100*8 = 900
	assert.Equal(t, 900, lir.GetBet1Payout())
	assert.Equal(t, 900, lir.GetBet2Payout())
	assert.Equal(t, 0, lir.GetBet3Payout())
	assert.Equal(t, 1800, lir.GetTotalPayout())
}

// **Pull はリスクを「下げる」操作。**issue #4699 は「掛け金を1つ引き上げる」と
// 書いているが、ドメインは1口ぶん取り下げて手元に戻す。取り消せないのはそこで、
// 危険だからではない。
func TestLetItRide_GetPullPreview(t *testing.T) {
	newDecisionGame := func(t *testing.T) *domain.LetItRide {
		t.Helper()
		lir := domain.NewDefaultLetItRide()
		lir.Reset()
		require.NoError(t, lir.Bet(100))
		return lir
	}

	t.Run("returns nothing before any bet is placed", func(t *testing.T) {
		lir := domain.NewDefaultLetItRide()
		lir.Reset()
		assert.Nil(t, lir.GetPullPreview())
	})

	t.Run("first decision returns one bet and lowers the stake", func(t *testing.T) {
		lir := newDecisionGame(t)
		pv := lir.GetPullPreview()
		require.NotNil(t, pv)
		assert.Equal(t, 100, pv.Returned)
		assert.Equal(t, 300, pv.RiskBefore)
		assert.Equal(t, 200, pv.RiskAfter, "リスクは下がる (上がらない)")
	})

	t.Run("second decision reflects the bet already pulled", func(t *testing.T) {
		lir := newDecisionGame(t)
		require.NoError(t, lir.Pull())
		pv := lir.GetPullPreview()
		require.NotNil(t, pv)
		assert.Equal(t, 200, pv.RiskBefore, "1口取り下げた後の残りを見ていること")
		assert.Equal(t, 100, pv.RiskAfter)
	})

	t.Run("returns nothing once the round is resolved", func(t *testing.T) {
		lir := newDecisionGame(t)
		require.NoError(t, lir.Pull())
		require.NoError(t, lir.Pull())
		assert.Nil(t, lir.GetPullPreview())
	})
}
