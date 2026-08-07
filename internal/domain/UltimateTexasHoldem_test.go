package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// uthSetup builds an UltimateTexasHoldem positioned at the river phase with the
// given hands and bets, so the showdown payout logic can be exercised
// deterministically by calling Play(1) or Fold.
func uthSetup(t *testing.T, ante int, community, playerHand, dealerHand []*domain.Card, extraChipsForPlay int) *domain.UltimateTexasHoldem {
	t.Helper()
	u := domain.NewDefaultUltimateTexasHoldem()
	u.SetPhase(domain.UltimateTexasHoldemPhaseRiver)
	u.SetAnteBet(ante)
	u.SetBlindBet(ante)
	u.SetCommunity(community)
	u.SetPlayerHand(playerHand)
	u.SetDealerHand(dealerHand)
	// Reserve enough chips to pay the river play bet (1x ante) plus a generous
	// margin so accumulated payouts can be compared independently of starting
	// chips.
	u.SetChips(extraChipsForPlay)
	return u
}

func TestNewDefaultUltimateTexasHoldem(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	assert.Equal(t, domain.UltimateTexasHoldemPhaseBet, u.GetPhase())
	assert.Equal(t, domain.UltimateTexasHoldemDefaultChips, u.GetChips())
	assert.False(t, u.GetGameEndFlag())
	assert.Nil(t, u.GetPlayerHand())
	assert.Nil(t, u.GetDealerHand())
	assert.Nil(t, u.GetCommunity())
}

func TestUltimateTexasHoldem_Reset(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	require.NoError(t, u.Check())                                     // preflop -> flop
	require.NoError(t, u.Check())                                     // flop -> river
	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x)) // river -> end
	assert.Equal(t, domain.UltimateTexasHoldemPhaseEnd, u.GetPhase())

	u.Reset()
	assert.Equal(t, domain.UltimateTexasHoldemPhaseBet, u.GetPhase())
	assert.False(t, u.GetGameEndFlag())
	assert.Nil(t, u.GetPlayerHand())
	assert.Nil(t, u.GetDealerHand())
	assert.Nil(t, u.GetCommunity())
	assert.Equal(t, 0, u.GetAnteBet())
	assert.Equal(t, 0, u.GetBlindBet())
	assert.Equal(t, 0, u.GetTripsBet())
	assert.Equal(t, 0, u.GetPlayBet())
	assert.False(t, u.GetFolded())
}

func TestUltimateTexasHoldem_Reset_RefillChips(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	u.SetChips(5)
	u.Reset()
	assert.Equal(t, domain.UltimateTexasHoldemDefaultChips, u.GetChips())
}

func TestUltimateTexasHoldem_Bet_WrongPhase(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	u.SetPhase(domain.UltimateTexasHoldemPhasePreFlop)
	err := u.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestUltimateTexasHoldem_Bet_InvalidAnteAmount(t *testing.T) {
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
			u := domain.NewDefaultUltimateTexasHoldem()
			err := u.Bet(tt.ante, 0)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestUltimateTexasHoldem_Bet_InvalidTripsAmount(t *testing.T) {
	tests := []struct {
		name  string
		trips int
	}{
		{"Negative", -10},
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := domain.NewDefaultUltimateTexasHoldem()
			err := u.Bet(100, tt.trips)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestUltimateTexasHoldem_Bet_InsufficientChips(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	u.SetChips(100) // need 200 (ante 100 + blind 100)
	err := u.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestUltimateTexasHoldem_Bet_Success(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 50))
	assert.Equal(t, domain.UltimateTexasHoldemPhasePreFlop, u.GetPhase())
	assert.Equal(t, 100, u.GetAnteBet())
	assert.Equal(t, 100, u.GetBlindBet())
	assert.Equal(t, 50, u.GetTripsBet())
	assert.Len(t, u.GetPlayerHand(), 2)
	assert.Len(t, u.GetDealerHand(), 2)
	// 1000 - (ante 100 + blind 100 + trips 50) = 750
	assert.Equal(t, domain.UltimateTexasHoldemDefaultChips-250, u.GetChips())
}

func TestUltimateTexasHoldem_Play_WrongPhase(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	err := u.Play(3)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestUltimateTexasHoldem_Play_InvalidMultiplierPerPhase(t *testing.T) {
	tests := []struct {
		name  string
		phase int
		mult  int
	}{
		{"PreFlop2x", domain.UltimateTexasHoldemPhasePreFlop, 2},
		{"PreFlop1x", domain.UltimateTexasHoldemPhasePreFlop, 1},
		{"PreFlop5x", domain.UltimateTexasHoldemPhasePreFlop, 5},
		{"Flop3x", domain.UltimateTexasHoldemPhaseFlop, 3},
		{"Flop1x", domain.UltimateTexasHoldemPhaseFlop, 1},
		{"River2x", domain.UltimateTexasHoldemPhaseRiver, 2},
		{"River4x", domain.UltimateTexasHoldemPhaseRiver, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := domain.NewDefaultUltimateTexasHoldem()
			u.SetPhase(tt.phase)
			u.SetAnteBet(100)
			err := u.Play(tt.mult)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestUltimateTexasHoldem_Play_InsufficientChips(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	u.SetChips(0)
	err := u.Play(domain.UltimateTexasHoldemPlayPreFlop4x)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestUltimateTexasHoldem_Play_PreFlop4x_DealsAllCommunity(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	chipsBefore := u.GetChips()
	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayPreFlop4x))
	assert.Equal(t, 400, u.GetPlayBet())
	assert.Equal(t, domain.UltimateTexasHoldemPhaseEnd, u.GetPhase())
	assert.True(t, u.GetGameEndFlag())
	assert.Len(t, u.GetCommunity(), 5)
	// chips dropped by 400 (play bet) and were credited with any payouts
	// determined by the random showdown. Just sanity-check upper bound.
	assert.LessOrEqual(t, u.GetChips(), chipsBefore-400+5000)
}

func TestUltimateTexasHoldem_Play_PreFlop3x(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayPreFlop3x))
	assert.Equal(t, 300, u.GetPlayBet())
	assert.Equal(t, domain.UltimateTexasHoldemPhaseEnd, u.GetPhase())
	assert.Len(t, u.GetCommunity(), 5)
}

func TestUltimateTexasHoldem_Check_PreFlopToFlop(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	require.NoError(t, u.Check())
	assert.Equal(t, domain.UltimateTexasHoldemPhaseFlop, u.GetPhase())
	assert.Len(t, u.GetCommunity(), 3)
}

func TestUltimateTexasHoldem_Check_FlopToRiver_DealsTurnAndRiver(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	require.NoError(t, u.Check())
	require.NoError(t, u.Check())
	assert.Equal(t, domain.UltimateTexasHoldemPhaseRiver, u.GetPhase())
	assert.Len(t, u.GetCommunity(), 5)
}

func TestUltimateTexasHoldem_Check_WrongPhase(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	err := u.Check()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestUltimateTexasHoldem_Play_Flop2x(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	require.NoError(t, u.Check())
	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayFlop2x))
	assert.Equal(t, 200, u.GetPlayBet())
	assert.Equal(t, domain.UltimateTexasHoldemPhaseEnd, u.GetPhase())
	assert.Len(t, u.GetCommunity(), 5)
}

func TestUltimateTexasHoldem_Fold_WrongPhase(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	err := u.Fold()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestUltimateTexasHoldem_Fold_River_LosesAnteBlind(t *testing.T) {
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	chipsAfterBet := u.GetChips()
	require.NoError(t, u.Check())
	require.NoError(t, u.Check())
	require.NoError(t, u.Fold())
	assert.True(t, u.GetFolded())
	assert.Equal(t, domain.GameResultLose, u.GetResult())
	assert.Equal(t, domain.UltimateTexasHoldemPhaseEnd, u.GetPhase())
	// Chips unchanged from after-bet (no play bet placed, no payout because
	// trips bet is 0).
	assert.Equal(t, chipsAfterBet, u.GetChips())
}

func TestUltimateTexasHoldem_Fold_River_PaysTripsRegardlessOfFold(t *testing.T) {
	// Player folds at the river but their hole+community make at least three
	// of a kind, so the Trips side bet still pays. Use uthSetup so we can pin
	// the community and hole cards.
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignHeart, 3},
		cd{domain.CardDesignClover, 9},
		cd{domain.CardDesignDiamond, 11},
		cd{domain.CardDesignSpade, 12},
	)
	playerHand := makeHand(
		cd{domain.CardDesignHeart, 3}, // three 3s with the two on the board
		cd{domain.CardDesignDiamond, 2},
	)
	dealerHand := makeHand(
		cd{domain.CardDesignClover, 13},
		cd{domain.CardDesignHeart, 13},
	)
	u := uthSetup(t, 100, community, playerHand, dealerHand, 1000)
	u.SetTripsBet(20)

	require.NoError(t, u.Fold())
	assert.Equal(t, domain.GameResultLose, u.GetResult())
	assert.True(t, u.GetFolded())
	// Trips pays 3:1 on three-of-a-kind: 20 + 20*3 = 80.
	assert.Equal(t, 20+20*domain.UltimateTexasHoldemTripsPayThreeOfAKind, u.GetTripsPayout())
	// Chips reflect the trips payout being credited.
	assert.Equal(t, 1000+u.GetTripsPayout(), u.GetChips())
}

func TestUltimateTexasHoldem_Showdown_PlayerWinsWithStraight_DealerQualifies(t *testing.T) {
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 6},
		cd{domain.CardDesignSpade, 13},
	)
	playerHand := makeHand(
		cd{domain.CardDesignHeart, 7}, // straight 3-4-5-6-7
		cd{domain.CardDesignDiamond, 2},
	)
	dealerHand := makeHand(
		cd{domain.CardDesignHeart, 13}, // pair of kings (qualifies)
		cd{domain.CardDesignClover, 11},
	)
	u := uthSetup(t, 100, community, playerHand, dealerHand, 500)

	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x))
	assert.Equal(t, domain.GameResultWin, u.GetResult())
	assert.True(t, u.GetDealerQualified())
	// Ante 1:1 -> 200, Play 1:1 -> 200, Blind 1:1 (straight) -> 200, total 600.
	assert.Equal(t, 200, u.GetAntePayout())
	assert.Equal(t, 200, u.GetPlayPayout())
	assert.Equal(t, 200, u.GetBlindPayout())
	assert.Equal(t, 600, u.GetTotalPayout())
}

func TestUltimateTexasHoldem_Showdown_DealerDoesNotQualify_AntePushes(t *testing.T) {
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 6},
		cd{domain.CardDesignSpade, 9},
	)
	playerHand := makeHand(
		cd{domain.CardDesignHeart, 7}, // straight 3-4-5-6-7
		cd{domain.CardDesignDiamond, 2},
	)
	dealerHand := makeHand(
		cd{domain.CardDesignDiamond, 10},
		cd{domain.CardDesignClover, 11},
	)
	u := uthSetup(t, 100, community, playerHand, dealerHand, 500)

	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x))
	assert.Equal(t, domain.GameResultWin, u.GetResult())
	assert.False(t, u.GetDealerQualified(), "dealer high card should NOT qualify")
	// Ante pushes (just returned) -> 100, Play 1:1 -> 200, Blind 1:1 -> 200.
	assert.Equal(t, 100, u.GetAntePayout())
	assert.Equal(t, 200, u.GetPlayPayout())
	assert.Equal(t, 200, u.GetBlindPayout())
}

func TestUltimateTexasHoldem_Showdown_PlayerLoses(t *testing.T) {
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 6},
		cd{domain.CardDesignSpade, 13},
	)
	playerHand := makeHand(
		cd{domain.CardDesignHeart, 8},
		cd{domain.CardDesignDiamond, 10},
	) // K-high through community; dealer pair of kings beats.
	dealerHand := makeHand(
		cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignClover, 11},
	)
	u := uthSetup(t, 100, community, playerHand, dealerHand, 500)

	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x))
	assert.Equal(t, domain.GameResultLose, u.GetResult())
	assert.Equal(t, 0, u.GetAntePayout())
	assert.Equal(t, 0, u.GetBlindPayout())
	assert.Equal(t, 0, u.GetPlayPayout())
}

func TestUltimateTexasHoldem_Showdown_Push(t *testing.T) {
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 6},
		cd{domain.CardDesignSpade, 13},
	)
	playerHand := makeHand(
		cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignDiamond, 11},
	) // pair of kings using community 13.
	dealerHand := makeHand(
		cd{domain.CardDesignClover, 13},
		cd{domain.CardDesignSpade, 11},
	) // same pair of kings + same kickers (11, 6, 5).
	u := uthSetup(t, 100, community, playerHand, dealerHand, 500)

	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x))
	assert.Equal(t, domain.GameResultDraw, u.GetResult())
	// Push: each leg returns its stake.
	assert.Equal(t, 100, u.GetAntePayout())
	assert.Equal(t, 100, u.GetBlindPayout())
	assert.Equal(t, 100, u.GetPlayPayout()) // playBet was 100 (1x).
}

func TestUltimateTexasHoldem_Blind_PaysFlush3to2(t *testing.T) {
	// Player makes a flush; dealer can't qualify with a pair of 9s? – let's
	// just ensure the flush+win path: spade flush for player, no dealer pair.
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignDiamond, 6},
		cd{domain.CardDesignClover, 4},
	)
	playerHand := makeHand(
		cd{domain.CardDesignSpade, 11},
		cd{domain.CardDesignSpade, 13},
	)
	dealerHand := makeHand(
		cd{domain.CardDesignDiamond, 13}, // dealer K-high; doesn't qualify
		cd{domain.CardDesignClover, 12},
	)
	u := uthSetup(t, 100, community, playerHand, dealerHand, 500)
	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x))
	assert.Equal(t, domain.GameResultWin, u.GetResult())
	// Blind 3:2 on flush of 100 -> profit 150 + return 100 = 250
	assert.Equal(t, 250, u.GetBlindPayout())
}

func TestUltimateTexasHoldem_Blind_BelowStraightPushesOnWin(t *testing.T) {
	// Player wins with two pair (below straight). Blind should push (return only).
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignHeart, 3},
		cd{domain.CardDesignClover, 9},
		cd{domain.CardDesignDiamond, 2},
		cd{domain.CardDesignSpade, 12},
	)
	playerHand := makeHand(
		cd{domain.CardDesignHeart, 9},
		cd{domain.CardDesignDiamond, 7},
	) // two pair: 9s and 3s.
	dealerHand := makeHand(
		cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignDiamond, 5},
	) // pair of 3s only (qualifies but loses to two pair).
	u := uthSetup(t, 100, community, playerHand, dealerHand, 500)
	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x))
	assert.Equal(t, domain.GameResultWin, u.GetResult())
	assert.Equal(t, 100, u.GetBlindPayout(), "Blind pushes on win when hand is below straight")
}

func TestUltimateTexasHoldem_Trips_StraightPayout(t *testing.T) {
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 6},
		cd{domain.CardDesignSpade, 9},
	)
	playerHand := makeHand(
		cd{domain.CardDesignHeart, 7},
		cd{domain.CardDesignDiamond, 2},
	)
	dealerHand := makeHand(
		cd{domain.CardDesignDiamond, 10},
		cd{domain.CardDesignClover, 11},
	)
	u := uthSetup(t, 100, community, playerHand, dealerHand, 500)
	u.SetTripsBet(20)
	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x))
	// Straight pays 4:1 on trips: 20 + 20*4 = 100
	assert.Equal(t, 20+20*domain.UltimateTexasHoldemTripsPayStraight, u.GetTripsPayout())
}

func TestUltimateTexasHoldem_Trips_NoPayoutForLessThanThreeOfAKind(t *testing.T) {
	community := makeHand(
		cd{domain.CardDesignSpade, 3},
		cd{domain.CardDesignHeart, 7},
		cd{domain.CardDesignClover, 9},
		cd{domain.CardDesignDiamond, 12},
		cd{domain.CardDesignSpade, 13},
	)
	playerHand := makeHand(
		cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignDiamond, 5},
	) // K-high, nothing.
	dealerHand := makeHand(
		cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignClover, 6},
	)
	u := uthSetup(t, 100, community, playerHand, dealerHand, 500)
	u.SetTripsBet(20)
	require.NoError(t, u.Play(domain.UltimateTexasHoldemPlayRiver1x))
	assert.Equal(t, 0, u.GetTripsPayout())
}

func TestUltimateTexasHoldem_PlayerHandRank_PopulatedAtFlop(t *testing.T) {
	// Drive a hand to FLOP with a deterministic pair so the frontend hint at
	// FLOP can read a meaningful rank without waiting for showdown.
	u := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, u.Bet(100, 0))
	u.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignHeart, 2},
	))
	u.SetCommunity(makeHand(
		cd{domain.CardDesignClover, 7},
		cd{domain.CardDesignDiamond, 9},
		cd{domain.CardDesignSpade, 13},
	))
	u.SetPhase(domain.UltimateTexasHoldemPhaseFlop)
	require.NoError(t, u.Check()) // deals turn + river and updates mid-rank.
	require.GreaterOrEqual(t, u.GetPlayerHandRank(), domain.PokerHandOnePair)
}

func TestUltimateTexasHoldem_JSONRoundTrip(t *testing.T) {
	src := domain.NewDefaultUltimateTexasHoldem()
	require.NoError(t, src.Bet(100, 20))
	require.NoError(t, src.Play(domain.UltimateTexasHoldemPlayPreFlop3x))

	data, err := json.Marshal(src)
	require.NoError(t, err)

	dst := new(domain.UltimateTexasHoldem)
	require.NoError(t, json.Unmarshal(data, dst))

	assert.Equal(t, src.GetPhase(), dst.GetPhase())
	assert.Equal(t, src.GetAnteBet(), dst.GetAnteBet())
	assert.Equal(t, src.GetBlindBet(), dst.GetBlindBet())
	assert.Equal(t, src.GetTripsBet(), dst.GetTripsBet())
	assert.Equal(t, src.GetPlayBet(), dst.GetPlayBet())
	assert.Equal(t, src.GetChips(), dst.GetChips())
	assert.Equal(t, src.GetResult(), dst.GetResult())
	assert.Equal(t, src.GetAntePayout(), dst.GetAntePayout())
	assert.Equal(t, src.GetBlindPayout(), dst.GetBlindPayout())
	assert.Equal(t, src.GetPlayPayout(), dst.GetPlayPayout())
	assert.Equal(t, src.GetTripsPayout(), dst.GetTripsPayout())
	assert.Len(t, dst.GetPlayerHand(), 2)
	assert.Len(t, dst.GetDealerHand(), 2)
	assert.Len(t, dst.GetCommunity(), 5)
}

func TestUltimateTexasHoldem_JSONUnmarshal_RejectsHugeSlices(t *testing.T) {
	// Build a payload with an oversized action log.
	type entry struct {
		TurnNumber int    `json:"TurnNumber"`
		PlayerIdx  int    `json:"PlayerIdx"`
		ActionType string `json:"ActionType"`
		Detail     string `json:"Detail"`
	}
	huge := make([]entry, 1001)
	for i := range huge {
		huge[i] = entry{TurnNumber: i, PlayerIdx: 0, ActionType: "bet", Detail: "x"}
	}
	payload, err := json.Marshal(map[string]any{"al": huge})
	require.NoError(t, err)
	dst := new(domain.UltimateTexasHoldem)
	err = json.Unmarshal(payload, dst)
	assert.Error(t, err)
}

// **CUI には 4x/3x/2x/1x/check/fold を選ぶ材料が何も無かった (#4709)。**
// Web はプリフロップの強さで 4x / 3x ボタンを光らせている。
func TestUltimateTexasHoldem_RecommendPlay(t *testing.T) {
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }
	preflop := func(hand ...*domain.Card) *domain.UltimateTexasHoldem {
		u := domain.NewDefaultUltimateTexasHoldem()
		u.SetPhase(domain.UltimateTexasHoldemPhasePreFlop)
		u.SetPlayerHand(hand)
		return u
	}

	t.Run("no recommendation before the cards are out", func(t *testing.T) {
		u := domain.NewDefaultUltimateTexasHoldem()
		u.SetPhase(domain.UltimateTexasHoldemPhaseBet)
		assert.Equal(t, "", u.RecommendPlay())
	})

	t.Run("4x on a pocket pair", func(t *testing.T) {
		assert.Equal(t, domain.UTHRecommendPlay4x,
			preflop(card(domain.CardDesignSpade, 7), card(domain.CardDesignHeart, 7)).RecommendPlay())
	})

	// **エースは 14 として数える。**value は 1 なので、素朴な大小比較だと
	// 最強のスターターを「弱い」と判定して 4x を逃す。
	t.Run("4x on any ace", func(t *testing.T) {
		assert.Equal(t, domain.UTHRecommendPlay4x,
			preflop(card(domain.CardDesignSpade, 1), card(domain.CardDesignHeart, 4)).RecommendPlay())
	})

	t.Run("3x on a king with a low kicker", func(t *testing.T) {
		assert.Equal(t, domain.UTHRecommendPlay3x,
			preflop(card(domain.CardDesignSpade, 13), card(domain.CardDesignHeart, 4)).RecommendPlay())
	})

	// **スーテッドかどうかで段が変わる。**同じ K-4 でも同スートなら 4x。
	t.Run("a suited king moves up from 3x to 4x", func(t *testing.T) {
		assert.Equal(t, domain.UTHRecommendPlay4x,
			preflop(card(domain.CardDesignSpade, 13), card(domain.CardDesignSpade, 4)).RecommendPlay())
	})

	t.Run("check on a weak offsuit holding", func(t *testing.T) {
		assert.Equal(t, domain.UTHRecommendCheck,
			preflop(card(domain.CardDesignSpade, 8), card(domain.CardDesignHeart, 3)).RecommendPlay())
	})

	t.Run("2x on the flop once a pair is made", func(t *testing.T) {
		u := preflop(card(domain.CardDesignSpade, 8), card(domain.CardDesignHeart, 3))
		u.SetPhase(domain.UltimateTexasHoldemPhaseFlop)
		u.SetCommunity([]*domain.Card{
			card(domain.CardDesignClover, 8), card(domain.CardDesignDiamond, 11), card(domain.CardDesignSpade, 2),
		})
		assert.Equal(t, domain.UTHRecommendPlay2x, u.RecommendPlay())
	})

	t.Run("check on the flop with nothing made", func(t *testing.T) {
		u := preflop(card(domain.CardDesignSpade, 8), card(domain.CardDesignHeart, 3))
		u.SetPhase(domain.UltimateTexasHoldemPhaseFlop)
		u.SetCommunity([]*domain.Card{
			card(domain.CardDesignClover, 9), card(domain.CardDesignDiamond, 11), card(domain.CardDesignSpade, 2),
		})
		assert.Equal(t, domain.UTHRecommendCheck, u.RecommendPlay())
	})

	t.Run("1x on the river with a made hand", func(t *testing.T) {
		u := preflop(card(domain.CardDesignSpade, 8), card(domain.CardDesignHeart, 3))
		u.SetPhase(domain.UltimateTexasHoldemPhaseRiver)
		u.SetCommunity([]*domain.Card{
			card(domain.CardDesignClover, 8), card(domain.CardDesignDiamond, 11), card(domain.CardDesignSpade, 2),
			card(domain.CardDesignHeart, 5), card(domain.CardDesignClover, 6),
		})
		assert.Equal(t, domain.UTHRecommendPlay1x, u.RecommendPlay())
	})

	// **リバーで役が無ければ降りる。**チェックは選べないので check とは違う。
	t.Run("fold on the river with nothing made", func(t *testing.T) {
		u := preflop(card(domain.CardDesignSpade, 8), card(domain.CardDesignHeart, 3))
		u.SetPhase(domain.UltimateTexasHoldemPhaseRiver)
		u.SetCommunity([]*domain.Card{
			card(domain.CardDesignClover, 9), card(domain.CardDesignDiamond, 11), card(domain.CardDesignSpade, 2),
			card(domain.CardDesignHeart, 5), card(domain.CardDesignClover, 7),
		})
		assert.Equal(t, domain.UTHRecommendFold, u.RecommendPlay())
	})
}
