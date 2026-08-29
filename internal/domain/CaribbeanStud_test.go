package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// helper: build a 5-card hand from (design, value) pairs.
type cd struct {
	d, v int
}

func makeHand(specs ...cd) []*domain.Card {
	cards := make([]*domain.Card, len(specs))
	for i, s := range specs {
		cards[i] = domain.NewCard(s.d, s.v, false)
	}
	return cards
}

func TestNewDefaultCaribbeanStud(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	assert.Equal(t, domain.CaribbeanStudPhaseBet, cs.GetPhase())
	assert.Equal(t, domain.CaribbeanStudDefaultChips, cs.GetChips())
	assert.False(t, cs.GetGameEndFlag())
	assert.Nil(t, cs.GetPlayerHand())
	assert.Nil(t, cs.GetDealerHand())
}

func TestCaribbeanStud_Reset(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Play())
	assert.Equal(t, domain.CaribbeanStudPhaseEnd, cs.GetPhase())

	cs.Reset()
	assert.Equal(t, domain.CaribbeanStudPhaseBet, cs.GetPhase())
	assert.False(t, cs.GetGameEndFlag())
	assert.Nil(t, cs.GetPlayerHand())
	assert.Nil(t, cs.GetDealerHand())
	assert.Equal(t, 0, cs.GetAnteBet())
	assert.Equal(t, 0, cs.GetJackpotBet())
	assert.Equal(t, 0, cs.GetPlayBet())
}

func TestCaribbeanStud_Reset_RefillChips(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	cs.SetChips(5)
	cs.Reset()
	assert.Equal(t, domain.CaribbeanStudDefaultChips, cs.GetChips())
}

func TestCaribbeanStud_Bet_WrongPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	cs.SetPhase(domain.CaribbeanStudPhaseAction)
	err := cs.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCaribbeanStud_Bet_InvalidAnteAmount(t *testing.T) {
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
			cs := domain.NewDefaultCaribbeanStud()
			err := cs.Bet(tt.ante, 0)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestCaribbeanStud_Bet_InvalidJackpotAmount(t *testing.T) {
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
			cs := domain.NewDefaultCaribbeanStud()
			err := cs.Bet(100, tt.jackpot)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestCaribbeanStud_Bet_InsufficientChips(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	cs.SetChips(50)
	err := cs.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestCaribbeanStud_Bet_Success(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	err := cs.Bet(100, 50)
	assert.NoError(t, err)
	assert.Equal(t, domain.CaribbeanStudPhaseAction, cs.GetPhase())
	assert.Equal(t, 100, cs.GetAnteBet())
	assert.Equal(t, 50, cs.GetJackpotBet())
	assert.Len(t, cs.GetPlayerHand(), 5)
	assert.Len(t, cs.GetDealerHand(), 5)
	assert.Equal(t, domain.CaribbeanStudDefaultChips-150, cs.GetChips())
}

func TestCaribbeanStud_Bet_AnteOnlyNoJackpot(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	err := cs.Bet(100, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, cs.GetJackpotBet())
	assert.Equal(t, domain.CaribbeanStudDefaultChips-100, cs.GetChips())
}

func TestCaribbeanStud_Play_WrongPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	err := cs.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCaribbeanStud_Play_InsufficientChips(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))
	cs.SetChips(0)
	err := cs.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestCaribbeanStud_Play_PlacesDoubleAnte(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Play())
	assert.Equal(t, 200, cs.GetPlayBet())
	assert.Equal(t, domain.CaribbeanStudPhaseEnd, cs.GetPhase())
	assert.True(t, cs.GetGameEndFlag())
}

func TestCaribbeanStud_Fold_WrongPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	err := cs.Fold()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCaribbeanStud_Fold_Success(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))
	chipsBefore := cs.GetChips()
	require.NoError(t, cs.Fold())
	assert.Equal(t, domain.CaribbeanStudPhaseEnd, cs.GetPhase())
	assert.True(t, cs.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, cs.GetResult())
	assert.Equal(t, chipsBefore, cs.GetChips())
}

func TestCaribbeanStud_Fold_JackpotStillEvaluated(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 10))

	// Player hand: a flush (jackpot pays 50:1)
	cs.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignSpade, 11},
	))
	chipsBefore := cs.GetChips()
	require.NoError(t, cs.Fold())
	assert.Equal(t, 10*domain.CaribbeanStudJackpotFlush, cs.GetJackpotPayout())
	assert.Equal(t, chipsBefore+10*domain.CaribbeanStudJackpotFlush, cs.GetChips())
}

func TestCaribbeanStud_DealerQualification(t *testing.T) {
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
			cs := domain.NewDefaultCaribbeanStud()
			require.NoError(t, cs.Bet(100, 0))

			// Weak player hand (Ace high) so dealer wins via kicker if needed
			cs.SetPlayerHand(makeHand(
				cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 4},
				cd{domain.CardDesignHeart, 6}, cd{domain.CardDesignDiamond, 8},
				cd{domain.CardDesignSpade, 10}))
			cs.SetDealerHand(tt.dealer)
			require.NoError(t, cs.Play())
			assert.Equal(t, tt.qualified, cs.GetDealerQualified())
		})
	}
}

func TestCaribbeanStud_Payouts_DealerNotQualified(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))

	// Player: high card; Dealer: Queen high (does not qualify)
	cs.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignHeart, 6}, cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 10}))
	cs.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 12}, cd{domain.CardDesignHeart, 5},
		cd{domain.CardDesignClover, 3}, cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignDiamond, 9}))

	require.NoError(t, cs.Play())
	assert.False(t, cs.GetDealerQualified())
	assert.Equal(t, 200, cs.GetAntePayout()) // ante 1:1
	assert.Equal(t, 200, cs.GetPlayPayout()) // play push (returned, equal to playBet=200)
}

func TestCaribbeanStud_Payouts_PlayerWinsFlush(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))

	// Player: flush; Dealer: pair (qualifies but loses)
	cs.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignSpade, 11}))
	cs.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 4}, cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
		cd{domain.CardDesignDiamond, 10}))

	require.NoError(t, cs.Play())
	assert.True(t, cs.GetDealerQualified())
	assert.Equal(t, domain.GameResultWin, cs.GetResult())
	assert.Equal(t, 200, cs.GetAntePayout())
	// playBet=200, multiplier=5 → returns 200 + 200*5 = 1200
	assert.Equal(t, 1200, cs.GetPlayPayout())
}

func TestCaribbeanStud_Payouts_PlayerLoses(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))

	// Player: high card; Dealer: pair (qualifies and wins)
	cs.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignHeart, 6}, cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 10}))
	cs.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 4}, cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
		cd{domain.CardDesignDiamond, 11}))

	require.NoError(t, cs.Play())
	assert.True(t, cs.GetDealerQualified())
	assert.Equal(t, domain.GameResultLose, cs.GetResult())
	assert.Equal(t, 0, cs.GetAntePayout())
	assert.Equal(t, 0, cs.GetPlayPayout())
}

func TestCaribbeanStud_Payouts_Push(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))

	// Identical hand ranks and identical kickers → tie
	cs.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 5}, cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 9},
		cd{domain.CardDesignSpade, 11}))
	cs.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignHeart, 5},
		cd{domain.CardDesignClover, 7}, cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignDiamond, 11}))

	require.NoError(t, cs.Play())
	assert.True(t, cs.GetDealerQualified())
	assert.Equal(t, domain.GameResultDraw, cs.GetResult())
	assert.Equal(t, 100, cs.GetAntePayout()) // push (return)
	assert.Equal(t, 200, cs.GetPlayPayout()) // push (return)
}

func TestCaribbeanStud_PlayMultipliers(t *testing.T) {
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
			multiplier: domain.CaribbeanStudPayRoyalFlush,
		},
		{
			name: "StraightFlush",
			hand: makeHand(
				cd{domain.CardDesignClover, 5}, cd{domain.CardDesignClover, 6},
				cd{domain.CardDesignClover, 7}, cd{domain.CardDesignClover, 8},
				cd{domain.CardDesignClover, 9}),
			multiplier: domain.CaribbeanStudPayStraightFlush,
		},
		{
			name: "FourOfAKind",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanStudPayFourOfAKind,
		},
		{
			name: "FullHouse",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 2},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanStudPayFullHouse,
		},
		{
			name: "Straight",
			hand: makeHand(
				cd{domain.CardDesignSpade, 5}, cd{domain.CardDesignClover, 6},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 8},
				cd{domain.CardDesignClover, 9}),
			multiplier: domain.CaribbeanStudPayStraight,
		},
		{
			name: "ThreeOfAKind",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanStudPayThreeOfAKind,
		},
		{
			name: "TwoPair",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 4}, cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanStudPayTwoPair,
		},
		{
			name: "OnePair",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 5}, cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanStudPayPair,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := domain.NewDefaultCaribbeanStud()
			cs.SetChips(100000)
			require.NoError(t, cs.Bet(100, 0))
			cs.SetPlayerHand(tt.hand)
			// Dealer: Queen high (does not qualify) → force play multiplier path manually
			// Use a weaker dealer hand BUT qualifying so player wins.
			cs.SetDealerHand(makeHand(
				cd{domain.CardDesignDiamond, 3}, cd{domain.CardDesignHeart, 3},
				cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
				cd{domain.CardDesignDiamond, 10}))
			require.NoError(t, cs.Play())
			assert.Equal(t, domain.GameResultWin, cs.GetResult())
			expected := 200 + 200*tt.multiplier
			assert.Equal(t, expected, cs.GetPlayPayout())
		})
	}
}

func TestCaribbeanStud_JackpotPayouts(t *testing.T) {
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
			multiplier: domain.CaribbeanStudJackpotRoyalFlush,
		},
		{
			name: "StraightFlushJackpot",
			hand: makeHand(
				cd{domain.CardDesignClover, 5}, cd{domain.CardDesignClover, 6},
				cd{domain.CardDesignClover, 7}, cd{domain.CardDesignClover, 8},
				cd{domain.CardDesignClover, 9}),
			multiplier: domain.CaribbeanStudJackpotStraightFlush,
		},
		{
			name: "FourOfAKindJackpot",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanStudJackpotFourOfAKind,
		},
		{
			name: "FullHouseJackpot",
			hand: makeHand(
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignHeart, 7}, cd{domain.CardDesignDiamond, 2},
				cd{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanStudJackpotFullHouse,
		},
		{
			name: "FlushJackpot",
			hand: makeHand(
				cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignSpade, 5},
				cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignSpade, 9},
				cd{domain.CardDesignSpade, 11}),
			multiplier: domain.CaribbeanStudJackpotFlush,
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
			cs := domain.NewDefaultCaribbeanStud()
			cs.SetChips(100000)
			require.NoError(t, cs.Bet(100, 10))
			cs.SetPlayerHand(tt.hand)
			cs.SetDealerHand(makeHand(
				cd{domain.CardDesignDiamond, 3}, cd{domain.CardDesignHeart, 3},
				cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
				cd{domain.CardDesignDiamond, 10}))
			require.NoError(t, cs.Play())
			assert.Equal(t, 10*tt.multiplier, cs.GetJackpotPayout())
		})
	}
}

func TestCaribbeanStud_TotalPayoutZero(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	assert.Equal(t, 0, cs.GetTotalPayout())
}

func TestCaribbeanStud_GetActionLog(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Play())
	assert.NotEmpty(t, cs.GetActionLog())
}

func TestCaribbeanStud_JSONRoundTrip(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	require.NoError(t, cs.Bet(100, 10))
	require.NoError(t, cs.Play())

	data, err := json.Marshal(cs)
	require.NoError(t, err)

	var restored domain.CaribbeanStud
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, cs.GetPhase(), restored.GetPhase())
	assert.Equal(t, cs.GetChips(), restored.GetChips())
	assert.Equal(t, cs.GetAnteBet(), restored.GetAnteBet())
	assert.Equal(t, cs.GetJackpotBet(), restored.GetJackpotBet())
	assert.Equal(t, cs.GetPlayBet(), restored.GetPlayBet())
	assert.Equal(t, cs.GetResult(), restored.GetResult())
	assert.Equal(t, cs.GetTotalPayout(), restored.GetTotalPayout())
	assert.Equal(t, cs.GetPlayerHandRank(), restored.GetPlayerHandRank())
	assert.Equal(t, cs.GetDealerHandRank(), restored.GetDealerHandRank())
}

func TestCaribbeanStud_UnmarshalJSON_InvalidData(t *testing.T) {
	var cs domain.CaribbeanStud
	err := cs.UnmarshalJSON([]byte("not json"))
	assert.Error(t, err)
}

func TestCaribbeanStud_SetSettersExposeFields(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	cs.SetPlayerHand(makeHand(cd{domain.CardDesignSpade, 1}))
	cs.SetDealerHand(makeHand(cd{domain.CardDesignClover, 13}))
	cs.SetAnteBet(100)
	cs.SetJackpotBet(20)
	cs.SetPlayBet(200)
	assert.Len(t, cs.GetPlayerHand(), 1)
	assert.Len(t, cs.GetDealerHand(), 1)
	assert.Equal(t, 100, cs.GetAnteBet())
	assert.Equal(t, 20, cs.GetJackpotBet())
	assert.Equal(t, 200, cs.GetPlayBet())
}

func TestCaribbeanStud_TotalPayoutConsistency(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	cs.SetChips(100000)
	require.NoError(t, cs.Bet(100, 10))

	// Player: flush (pays jackpot); Dealer: pair (qualifies but loses)
	cs.SetPlayerHand(makeHand(
		cd{domain.CardDesignSpade, 2}, cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignSpade, 7}, cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignSpade, 11}))
	cs.SetDealerHand(makeHand(
		cd{domain.CardDesignDiamond, 4}, cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignClover, 6}, cd{domain.CardDesignSpade, 8},
		cd{domain.CardDesignDiamond, 10}))

	require.NoError(t, cs.Play())

	// Ensure jackpot is non-zero so we are actually testing consistency with it
	assert.Greater(t, cs.GetJackpotPayout(), 0)

	totalExpected := cs.GetAntePayout() + cs.GetPlayPayout() + cs.GetJackpotPayout()
	assert.Equal(t, totalExpected, cs.GetTotalPayout())
}
