package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewDefaultRedDog(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	assert.Equal(t, domain.RedDogPhaseBet, rd.GetPhase())
	assert.Equal(t, domain.RedDogDefaultChips, rd.GetChips())
	assert.False(t, rd.GetGameEndFlag())
	assert.Nil(t, rd.GetInitialCards())
	assert.Nil(t, rd.GetThirdCard())
}

func TestRedDog_Reset(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetPhase(domain.RedDogPhaseEnd)
	rd.Reset()
	assert.Equal(t, domain.RedDogPhaseBet, rd.GetPhase())
	assert.False(t, rd.GetGameEndFlag())
	assert.Nil(t, rd.GetInitialCards())
	assert.Nil(t, rd.GetThirdCard())
	assert.Equal(t, 0, rd.GetAnte())
	assert.Equal(t, 0, rd.GetRaise())
	assert.Equal(t, 0, rd.GetSpread())
}

func TestRedDog_Reset_RefillChips(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	rd.SetChips(5) // below MinBet
	rd.Reset()
	assert.Equal(t, domain.RedDogDefaultChips, rd.GetChips())
}

func TestRedDog_Reset_NoRefillAboveThreshold(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	rd.SetChips(500)
	rd.Reset()
	assert.Equal(t, 500, rd.GetChips())
}

func TestRedDog_Bet_WrongPhase(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	rd.SetPhase(domain.RedDogPhaseSpreadDecision)
	err := rd.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRedDog_Bet_InvalidAmount(t *testing.T) {
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
			rd := domain.NewDefaultRedDog()
			err := rd.Bet(tt.amount)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestRedDog_Bet_InsufficientChips(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	rd.SetChips(50)
	err := rd.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestRedDog_Bet_Consecutive_Push(t *testing.T) {
	// Force consecutive cards manually
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	// Override: 5 and 6 → consecutive → push, ante refunded immediately
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 6},
	))
	rd.ResolveInitial()
	assert.Equal(t, domain.RedDogPhaseEnd, rd.GetPhase())
	assert.True(t, rd.GetGameEndFlag())
	assert.Equal(t, domain.GameResultDraw, rd.GetResult())
	assert.Equal(t, 100, rd.GetTotalPayout()) // ante refunded
	assert.Equal(t, domain.RedDogDefaultChips, rd.GetChips())
}

func TestRedDog_Bet_PairThenMatch_Win(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignClover, 7},
	))
	rd.ResolveInitial()
	assert.Equal(t, domain.RedDogPhasePairThird, rd.GetPhase())
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	rd.ResolveThird()
	assert.Equal(t, domain.GameResultWin, rd.GetResult())
	// 11:1 on ante: 100 + 100*11 = 1200
	assert.Equal(t, 1200, rd.GetTotalPayout())
}

func TestRedDog_Bet_PairThenNoMatch_Push(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignClover, 7},
	))
	rd.ResolveInitial()
	assert.Equal(t, domain.RedDogPhasePairThird, rd.GetPhase())
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 9, true))
	rd.ResolveThird()
	assert.Equal(t, domain.GameResultDraw, rd.GetResult())
	assert.Equal(t, 100, rd.GetTotalPayout()) // ante refunded
}

func TestRedDog_SpreadDecision_Spread1_Win(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	// 5 and 7 → spread = |7-5|-1 = 1
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 7},
	))
	rd.ResolveInitial()
	assert.Equal(t, domain.RedDogPhaseSpreadDecision, rd.GetPhase())
	assert.Equal(t, 1, rd.GetSpread())

	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 6, true))
	require.NoError(t, rd.Stay())
	assert.Equal(t, domain.GameResultWin, rd.GetResult())
	// spread 1 → 5:1 on ante (no raise): 100 + 100*5 = 600
	assert.Equal(t, 600, rd.GetTotalPayout())
}

func TestRedDog_SpreadDecision_Spread2_Win(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 8},
	))
	rd.ResolveInitial()
	assert.Equal(t, 2, rd.GetSpread())
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	require.NoError(t, rd.Stay())
	assert.Equal(t, domain.GameResultWin, rd.GetResult())
	// spread 2 → 4:1: 100 + 100*4 = 500
	assert.Equal(t, 500, rd.GetTotalPayout())
}

func TestRedDog_SpreadDecision_Spread3_Win(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 9},
	))
	rd.ResolveInitial()
	assert.Equal(t, 3, rd.GetSpread())
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	require.NoError(t, rd.Stay())
	// spread 3 → 2:1: 100 + 100*2 = 300
	assert.Equal(t, 300, rd.GetTotalPayout())
}

func TestRedDog_SpreadDecision_Spread4_Win(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 10},
	))
	rd.ResolveInitial()
	assert.Equal(t, 4, rd.GetSpread())
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	require.NoError(t, rd.Stay())
	// spread 4+ → 1:1: 100 + 100 = 200
	assert.Equal(t, 200, rd.GetTotalPayout())
}

func TestRedDog_SpreadDecision_LargeSpread(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	// 2 and Ace(14) → spread = 11
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignClover, 1}, // Ace high
	))
	rd.ResolveInitial()
	assert.Equal(t, 11, rd.GetSpread())
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	require.NoError(t, rd.Stay())
	assert.Equal(t, 200, rd.GetTotalPayout()) // 1:1
}

func TestRedDog_SpreadDecision_LossOutside(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 9},
	))
	rd.ResolveInitial()
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 13, true)) // K outside
	require.NoError(t, rd.Stay())
	assert.Equal(t, domain.GameResultLose, rd.GetResult())
	assert.Equal(t, 0, rd.GetTotalPayout())
}

func TestRedDog_SpreadDecision_LossOnBoundary(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 9},
	))
	rd.ResolveInitial()
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 5, true)) // matches lower
	require.NoError(t, rd.Stay())
	assert.Equal(t, domain.GameResultLose, rd.GetResult())
}

func TestRedDog_Raise_Success(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 10},
	))
	rd.ResolveInitial()
	chipsBefore := rd.GetChips()
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	require.NoError(t, rd.Raise(100))
	assert.Equal(t, 100, rd.GetRaise())
	// total bet 200, multiplier 1: payout = 200 + 200*1 = 400
	assert.Equal(t, 400, rd.GetTotalPayout())
	assert.Equal(t, chipsBefore-100+400, rd.GetChips())
}

func TestRedDog_Raise_LossLosesAllBets(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 9},
	))
	rd.ResolveInitial()
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 13, true))
	chipsBefore := rd.GetChips()
	require.NoError(t, rd.Raise(100))
	assert.Equal(t, domain.GameResultLose, rd.GetResult())
	assert.Equal(t, 0, rd.GetTotalPayout())
	assert.Equal(t, chipsBefore-100, rd.GetChips())
}

func TestRedDog_Raise_InvalidAmount(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 10},
	))
	rd.ResolveInitial()
	tests := []int{0, -1, 101, 5}
	for _, amount := range tests {
		err := rd.Raise(amount)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	}
}

func TestRedDog_Raise_InsufficientChips(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	rd.SetChips(150)
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 10},
	))
	rd.ResolveInitial()
	err := rd.Raise(100) // only 50 chips left
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestRedDog_Raise_WrongPhase(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	err := rd.Raise(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRedDog_Stay_WrongPhase(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	err := rd.Stay()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestRedDog_Bet_Success_NaturalFlow(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	assert.Equal(t, 100, rd.GetAnte())
	assert.Len(t, rd.GetInitialCards(), 2)
	assert.Equal(t, domain.RedDogPhaseInitialDealt, rd.GetPhase())
	assert.Equal(t, domain.RedDogDefaultChips-100, rd.GetChips())
}

func TestRedDog_AceLowerCard(t *testing.T) {
	// Ace (1 in card value) high (rank 14) — higher than K (13)
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 13}, // K
		cd{domain.CardDesignClover, 1}, // Ace
	))
	rd.ResolveInitial()
	// K(13) and A(14) → consecutive, push
	assert.Equal(t, domain.GameResultDraw, rd.GetResult())
}

func TestRedDog_GetActionLog(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	assert.NotEmpty(t, rd.GetActionLog())
}

func TestRedDog_JSONRoundTrip(t *testing.T) {
	rd := domain.NewDefaultRedDog()
	require.NoError(t, rd.Bet(100))
	rd.SetInitialCards(makeHand(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 9},
	))
	rd.ResolveInitial()
	rd.SetThirdCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	require.NoError(t, rd.Raise(50))

	data, err := json.Marshal(rd)
	require.NoError(t, err)
	var restored domain.RedDog
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, rd.GetPhase(), restored.GetPhase())
	assert.Equal(t, rd.GetChips(), restored.GetChips())
	assert.Equal(t, rd.GetAnte(), restored.GetAnte())
	assert.Equal(t, rd.GetRaise(), restored.GetRaise())
	assert.Equal(t, rd.GetSpread(), restored.GetSpread())
	assert.Equal(t, rd.GetResult(), restored.GetResult())
	assert.Equal(t, rd.GetTotalPayout(), restored.GetTotalPayout())
}

func TestRedDog_UnmarshalJSON_InvalidData(t *testing.T) {
	var rd domain.RedDog
	err := rd.UnmarshalJSON([]byte("not json"))
	assert.Error(t, err)
}

func TestRedDog_UnmarshalJSON_NilFields(t *testing.T) {
	data := []byte(`{"tc":null,"ic":null,"tr":null,"ch":null,"an":0,"rs":0,"sp":0,"ps":1,"ge":false,"gr":0,"tp":0,"al":null}`)
	var rd domain.RedDog
	require.NoError(t, json.Unmarshal(data, &rd))
	assert.NotNil(t, rd.GetInitialCards())
	assert.NotNil(t, rd.GetActionLog())
}
