package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// Helper to build the full game in the action phase with deterministic hands.
func setupActionPhase(t *testing.T, player []*domain.Card, dealer []*domain.Card, ante, acesUp int) *domain.FourCardPoker {
	t.Helper()
	fcp := domain.NewDefaultFourCardPoker()
	err := fcp.Bet(ante, acesUp)
	require.NoError(t, err)
	fcp.SetPlayerHand(player)
	fcp.SetDealerHand(dealer)
	return fcp
}

func dealerLowSix() []*domain.Card {
	// All low cards, no qualifying combinations — high card 9.
	return []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
	}
}

func TestNewDefaultFourCardPoker(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	assert.Equal(t, domain.FourCardPokerPhaseBet, fcp.GetPhase())
	assert.Equal(t, domain.FourCardPokerDefaultChips, fcp.GetChips())
	assert.False(t, fcp.GetGameEndFlag())
	assert.Nil(t, fcp.GetPlayerHand())
	assert.Nil(t, fcp.GetDealerHand())
}

func TestFourCardPoker_Reset(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	require.NoError(t, fcp.Bet(100, 0))
	require.NoError(t, fcp.Play(1))
	require.Equal(t, domain.FourCardPokerPhaseEnd, fcp.GetPhase())

	fcp.Reset()
	assert.Equal(t, domain.FourCardPokerPhaseBet, fcp.GetPhase())
	assert.False(t, fcp.GetGameEndFlag())
	assert.Nil(t, fcp.GetPlayerHand())
	assert.Nil(t, fcp.GetDealerHand())
	assert.Equal(t, 0, fcp.GetAnteBet())
	assert.Equal(t, 0, fcp.GetAcesUpBet())
	assert.Equal(t, 0, fcp.GetPlayBet())
	assert.Equal(t, 0, fcp.GetPlayMultiplier())
}

func TestFourCardPoker_Reset_RefillChips(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	fcp.SetChips(5)
	fcp.Reset()
	assert.Equal(t, domain.FourCardPokerDefaultChips, fcp.GetChips())
}

func TestFourCardPoker_Bet_WrongPhase(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	fcp.SetPhase(domain.FourCardPokerPhaseAction)
	err := fcp.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestFourCardPoker_Bet_InvalidAnte(t *testing.T) {
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
			fcp := domain.NewDefaultFourCardPoker()
			err := fcp.Bet(tt.ante, 0)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestFourCardPoker_Bet_InvalidAcesUp(t *testing.T) {
	tests := []struct {
		name   string
		acesUp int
	}{
		{"Negative", -10},
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fcp := domain.NewDefaultFourCardPoker()
			err := fcp.Bet(100, tt.acesUp)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestFourCardPoker_Bet_InsufficientChips(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	fcp.SetChips(50)
	err := fcp.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestFourCardPoker_Bet_Success(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	err := fcp.Bet(100, 50)
	assert.NoError(t, err)
	assert.Equal(t, domain.FourCardPokerPhaseAction, fcp.GetPhase())
	assert.Equal(t, 100, fcp.GetAnteBet())
	assert.Equal(t, 50, fcp.GetAcesUpBet())
	assert.Len(t, fcp.GetPlayerHand(), domain.FourCardPokerPlayerCards)
	assert.Len(t, fcp.GetDealerHand(), domain.FourCardPokerDealerCards)
	assert.Equal(t, domain.FourCardPokerDefaultChips-150, fcp.GetChips())
	assert.NotNil(t, fcp.GetDealerUpCard())
}

func TestFourCardPoker_Play_WrongPhase(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	err := fcp.Play(1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestFourCardPoker_Play_InvalidMultiplier(t *testing.T) {
	for _, mul := range []int{0, -1, 4, 99} {
		fcp := domain.NewDefaultFourCardPoker()
		require.NoError(t, fcp.Bet(100, 0))
		err := fcp.Play(mul)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	}
}

func TestFourCardPoker_Play_InsufficientChips(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	require.NoError(t, fcp.Bet(100, 0))
	fcp.SetChips(50) // less than ante*2
	err := fcp.Play(2)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestFourCardPoker_Play_Success(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	require.NoError(t, fcp.Bet(100, 0))
	require.NoError(t, fcp.Play(2))
	assert.Equal(t, domain.FourCardPokerPhaseEnd, fcp.GetPhase())
	assert.True(t, fcp.GetGameEndFlag())
	assert.Equal(t, 200, fcp.GetPlayBet())
	assert.Equal(t, 2, fcp.GetPlayMultiplier())
	assert.Len(t, fcp.GetPlayerBest(), 4)
	assert.Len(t, fcp.GetDealerBest(), 4)
}

func TestFourCardPoker_Fold_WrongPhase(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	err := fcp.Fold()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestFourCardPoker_Fold_Success(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	require.NoError(t, fcp.Bet(100, 0))
	chipsBefore := fcp.GetChips()
	require.NoError(t, fcp.Fold())
	assert.Equal(t, domain.FourCardPokerPhaseEnd, fcp.GetPhase())
	assert.True(t, fcp.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, fcp.GetResult())
	assert.Equal(t, chipsBefore, fcp.GetChips())
}

func TestFourCardPoker_Fold_AcesUpStillEvaluated(t *testing.T) {
	// Player has pair of aces in best 4 → Aces Up pays even on fold.
	player := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
	}
	fcp := setupActionPhase(t, player, dealerLowSix(), 100, 50)
	require.NoError(t, fcp.Fold())
	// Aces Up 1:1: 50 wager returned + 50 bonus = 100
	assert.Equal(t, 100, fcp.GetAcesUpPayout())
}

func TestFourCardPoker_Payouts_PlayerWins(t *testing.T) {
	// Player: 9-9-9 trips in 4-card → beats dealer high
	player := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}
	fcp := setupActionPhase(t, player, dealerLowSix(), 100, 0)
	require.NoError(t, fcp.Play(1))
	assert.Equal(t, domain.GameResultWin, fcp.GetResult())
	assert.Equal(t, 200, fcp.GetAntePayout()) // 100 + 100 win
	assert.Equal(t, 200, fcp.GetPlayPayout()) // 100 + 100 win
	// Ante Bonus: 3-of-a-kind 2:1 → 100 * 2 = 200
	assert.Equal(t, 200, fcp.GetAnteBonusPayout())
}

func TestFourCardPoker_Payouts_PlayerLoses(t *testing.T) {
	// Player: pure high card, Dealer: pair of kings
	player := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignClover, 11, false),
	}
	dealer := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignDiamond, 6, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	}
	fcp := setupActionPhase(t, player, dealer, 100, 0)
	require.NoError(t, fcp.Play(1))
	assert.Equal(t, domain.GameResultLose, fcp.GetResult())
	assert.Equal(t, 0, fcp.GetAntePayout())
	assert.Equal(t, 0, fcp.GetPlayPayout())
	assert.Equal(t, 0, fcp.GetAnteBonusPayout())
}

func TestFourCardPoker_Payouts_Draw(t *testing.T) {
	// Both player and dealer make identical pair-of-fives + same kickers.
	player := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}
	dealer := []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 2, false),
	}
	fcp := setupActionPhase(t, player, dealer, 100, 0)
	require.NoError(t, fcp.Play(1))
	assert.Equal(t, domain.GameResultDraw, fcp.GetResult())
	assert.Equal(t, 100, fcp.GetAntePayout()) // push
	assert.Equal(t, 100, fcp.GetPlayPayout()) // push
}

func TestFourCardPoker_AcesUp_Paytable(t *testing.T) {
	tests := []struct {
		name        string
		player      []*domain.Card
		expectedMul int // 0 means no payout
	}{
		{
			name: "FourOfAKind",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 9, false),
				domain.NewCard(domain.CardDesignClover, 9, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAcesUpFourOfAKind,
		},
		{
			name: "StraightFlush",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 6, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignSpade, 8, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAcesUpStraightFlush,
		},
		{
			name: "ThreeOfAKind",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 4, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAcesUpThreeOfAKind,
		},
		{
			name: "Flush",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 2, false),
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignSpade, 11, false),
				domain.NewCard(domain.CardDesignClover, 3, false),
			},
			expectedMul: domain.FourCardPokerAcesUpFlush,
		},
		{
			name: "Straight",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignClover, 6, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 8, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAcesUpStraight,
		},
		{
			name: "TwoPair",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAcesUpTwoPair,
		},
		{
			name: "PairOfAces",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignClover, 1, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAcesUpPairOfAces,
		},
		{
			name: "PairOfKings_NoPay",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 13, false),
				domain.NewCard(domain.CardDesignClover, 13, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: 0,
		},
		{
			name: "HighCard_NoPay",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 13, false),
				domain.NewCard(domain.CardDesignClover, 11, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fcp := setupActionPhase(t, tt.player, dealerLowSix(), 100, 50)
			require.NoError(t, fcp.Play(1))
			if tt.expectedMul == 0 {
				assert.Equal(t, 0, fcp.GetAcesUpPayout())
			} else {
				assert.Equal(t, 50+50*tt.expectedMul, fcp.GetAcesUpPayout())
			}
		})
	}
}

func TestFourCardPoker_AcesUp_NoBetNoPayout(t *testing.T) {
	player := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}
	fcp := setupActionPhase(t, player, dealerLowSix(), 100, 0)
	require.NoError(t, fcp.Play(1))
	assert.Equal(t, 0, fcp.GetAcesUpPayout())
}

func TestFourCardPoker_AnteBonus_Paytable(t *testing.T) {
	tests := []struct {
		name        string
		player      []*domain.Card
		expectedMul int
	}{
		{
			name: "ThreeOfAKind",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 9, false),
				domain.NewCard(domain.CardDesignClover, 9, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 5, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAnteBonusThreeOfAKind,
		},
		{
			name: "StraightFlush",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 6, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignSpade, 8, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAnteBonusStraightFlush,
		},
		{
			name: "FourOfAKind",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 9, false),
				domain.NewCard(domain.CardDesignClover, 9, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerAnteBonusFourOfAKind,
		},
		{
			name: "Pair_NoBonus",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fcp := setupActionPhase(t, tt.player, dealerLowSix(), 100, 0)
			require.NoError(t, fcp.Play(1))
			assert.Equal(t, tt.expectedMul*100, fcp.GetAnteBonusPayout())
		})
	}
}

func TestFourCardPoker_GetDealerUpCard_Nil(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	assert.Nil(t, fcp.GetDealerUpCard())
}

func TestFourCardPoker_GetTotalPayout(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	fcp.SetAntePayout(100)
	fcp.SetPlayPayout(200)
	fcp.SetAnteBonusPayout(50)
	fcp.SetAcesUpPayout(25)
	assert.Equal(t, 375, fcp.GetTotalPayout())
}

func TestFourCardPoker_ActionLog(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	require.NoError(t, fcp.Bet(100, 0))
	require.NoError(t, fcp.Play(1))
	log := fcp.GetActionLog()
	assert.NotEmpty(t, log)
}

func TestFourCardPoker_JSONRoundTrip(t *testing.T) {
	fcp := domain.NewDefaultFourCardPoker()
	require.NoError(t, fcp.Bet(100, 50))
	require.NoError(t, fcp.Play(2))

	data, err := json.Marshal(fcp)
	require.NoError(t, err)

	var got domain.FourCardPoker
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, fcp.GetPhase(), got.GetPhase())
	assert.Equal(t, fcp.GetAnteBet(), got.GetAnteBet())
	assert.Equal(t, fcp.GetAcesUpBet(), got.GetAcesUpBet())
	assert.Equal(t, fcp.GetPlayBet(), got.GetPlayBet())
	assert.Equal(t, fcp.GetPlayMultiplier(), got.GetPlayMultiplier())
	assert.Equal(t, fcp.GetResult(), got.GetResult())
	assert.Equal(t, fcp.GetTotalPayout(), got.GetTotalPayout())
}

func TestFourCardPoker_JSONUnmarshal_Reject_OversizedSlices(t *testing.T) {
	// Forge a payload with > maxSliceLen player hand entries.
	var big []domain.Card
	for i := 0; i < 1001; i++ {
		big = append(big, *domain.NewCard(domain.CardDesignSpade, 2, false))
	}
	payload := map[string]any{"ph": big}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	var fcp domain.FourCardPoker
	err = json.Unmarshal(data, &fcp)
	assert.Error(t, err)
}

func TestFourCardPoker_Setters(t *testing.T) {
	// Smoke test all setters and basic getters to exercise coverage paths.
	fcp := domain.NewDefaultFourCardPoker()
	fcp.SetAcesUpBet(50)
	fcp.SetPlayBet(200)
	fcp.SetResult(domain.GameResultWin)
	fcp.SetGameEndFlag(true)
	fcp.SetPlayerHandRank(domain.FourCardHandFlush)
	fcp.SetDealerHandRank(domain.FourCardHandPair)
	assert.Equal(t, 50, fcp.GetAcesUpBet())
	assert.Equal(t, 200, fcp.GetPlayBet())
	assert.Equal(t, domain.GameResultWin, fcp.GetResult())
	assert.True(t, fcp.GetGameEndFlag())
	assert.Equal(t, domain.FourCardHandFlush, fcp.GetPlayerHandRank())
	assert.Equal(t, domain.FourCardHandPair, fcp.GetDealerHandRank())
}

func TestFourCardPoker_RecommendPlayMultiplier(t *testing.T) {
	tests := []struct {
		name        string
		playerHand  []*domain.Card
		expectedMul int
	}{
		{
			name: "HighCard_Fold",
			playerHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 2, false),
				domain.NewCard(domain.CardDesignClover, 4, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 11, false),
			},
			expectedMul: 0,
		},
		{
			name: "Pair_Fold",
			playerHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 11, false),
			},
			expectedMul: 0,
		},
		{
			name: "TwoPair_Play1x",
			playerHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerMinPlayMul,
		},
		{
			name: "Straight_Play1x",
			playerHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignClover, 6, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 8, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerMinPlayMul,
		},
		{
			name: "Flush_Play1x",
			playerHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 2, false),
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignSpade, 11, false),
				domain.NewCard(domain.CardDesignClover, 3, false),
			},
			expectedMul: domain.FourCardPokerMinPlayMul,
		},
		{
			name: "ThreeOfAKind_Play3x",
			playerHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerMaxPlayMul,
		},
		{
			name: "StraightFlush_Play3x",
			playerHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 6, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignSpade, 8, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerMaxPlayMul,
		},
		{
			name: "FourOfAKind_Play3x",
			playerHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 9, false),
				domain.NewCard(domain.CardDesignClover, 9, false),
				domain.NewCard(domain.CardDesignHeart, 9, false),
				domain.NewCard(domain.CardDesignDiamond, 9, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			},
			expectedMul: domain.FourCardPokerMaxPlayMul,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fcp := setupActionPhase(t, tt.playerHand, dealerLowSix(), 100, 0)
			// playerHandRank is NOT set during the action phase
			assert.Equal(t, 0, fcp.GetPlayerHandRank())
			assert.Equal(t, tt.expectedMul, fcp.RecommendPlayMultiplier())
		})
	}

	t.Run("NonActionPhase_Bet", func(t *testing.T) {
		fcp := domain.NewDefaultFourCardPoker()
		assert.Equal(t, domain.FourCardPokerPhaseBet, fcp.GetPhase())
		assert.Equal(t, 0, fcp.RecommendPlayMultiplier())
	})

	t.Run("NonActionPhase_End", func(t *testing.T) {
		fcp := domain.NewDefaultFourCardPoker()
		fcp.SetPhase(domain.FourCardPokerPhaseEnd)
		assert.Equal(t, 0, fcp.RecommendPlayMultiplier())
	})

	t.Run("EmptyHandInActionPhase", func(t *testing.T) {
		fcp := domain.NewDefaultFourCardPoker()
		fcp.SetPhase(domain.FourCardPokerPhaseAction)
		fcp.SetPlayerHand(nil)
		assert.Equal(t, 0, fcp.RecommendPlayMultiplier())
	})
}
