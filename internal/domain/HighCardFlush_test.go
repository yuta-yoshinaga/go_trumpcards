package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// makeHCFCards is a tiny helper for building card slices in tests.
func makeHCFCards(specs ...[2]int) []*domain.Card {
	cards := make([]*domain.Card, 0, len(specs))
	for _, s := range specs {
		cards = append(cards, domain.NewCard(s[0], s[1], true))
	}
	return cards
}

func TestNewDefaultHighCardFlush(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	assert.Equal(t, domain.HighCardFlushPhaseBet, hcf.GetPhase())
	assert.Equal(t, domain.HighCardFlushDefaultChips, hcf.GetChips())
	assert.False(t, hcf.GetGameEndFlag())
	assert.Nil(t, hcf.GetPlayerHand())
	assert.Nil(t, hcf.GetDealerHand())
}

func TestHighCardFlush_Reset(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	require.NoError(t, hcf.Bet(100, 50, 20))
	require.NoError(t, hcf.Raise(1))
	assert.Equal(t, domain.HighCardFlushPhaseEnd, hcf.GetPhase())

	hcf.Reset()
	assert.Equal(t, domain.HighCardFlushPhaseBet, hcf.GetPhase())
	assert.False(t, hcf.GetGameEndFlag())
	assert.Nil(t, hcf.GetPlayerHand())
	assert.Nil(t, hcf.GetDealerHand())
	assert.Equal(t, 0, hcf.GetAnteBet())
	assert.Equal(t, 0, hcf.GetFlushBonusBet())
	assert.Equal(t, 0, hcf.GetStraightFlushBet())
	assert.Equal(t, 0, hcf.GetRaiseBet())
	assert.Equal(t, 0, hcf.GetAntePayout())
	assert.Equal(t, 0, hcf.GetRaisePayout())
	assert.Equal(t, 0, hcf.GetFlushBonusPayout())
	assert.Equal(t, 0, hcf.GetStraightFlushPayout())
	assert.False(t, hcf.GetDealerQualified())
}

func TestHighCardFlush_Reset_RefillChips(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetChips(5)
	hcf.Reset()
	assert.Equal(t, domain.HighCardFlushDefaultChips, hcf.GetChips())
}

func TestHighCardFlush_Bet_WrongPhase(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	err := hcf.Bet(100, 0, 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHighCardFlush_Bet_InvalidAmounts(t *testing.T) {
	cases := []struct {
		name             string
		ante, fb, sf     int
		wantInsufficient bool
	}{
		{"negative ante", -10, 0, 0, false},
		{"zero ante", 0, 0, 0, false},
		{"below min ante", 5, 0, 0, false},
		{"unit-violation ante", 15, 0, 0, false},
		{"above max ante", domain.HighCardFlushMaxBet + 10, 0, 0, false},
		{"negative flush bonus", 100, -1, 0, false},
		{"below min flush bonus", 100, 5, 0, false},
		{"unit-violation flush bonus", 100, 15, 0, false},
		{"negative straight flush", 100, 0, -5, false},
		{"unit-violation straight flush", 100, 0, 12, false},
		{"above max straight flush", 100, 0, domain.HighCardFlushMaxBet + 10, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hcf := domain.NewDefaultHighCardFlush()
			err := hcf.Bet(tc.ante, tc.fb, tc.sf)
			require.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestHighCardFlush_Bet_InsufficientChips(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetChips(50)
	err := hcf.Bet(100, 0, 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestHighCardFlush_Bet_DealsCards(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	require.NoError(t, hcf.Bet(100, 0, 0))
	assert.Equal(t, domain.HighCardFlushPhaseAction, hcf.GetPhase())
	assert.Len(t, hcf.GetPlayerHand(), domain.HighCardFlushHandSize)
	assert.Len(t, hcf.GetDealerHand(), domain.HighCardFlushHandSize)
	// Player flush length must be ≥ 2 (pigeonhole on 7 cards / 4 suits).
	assert.GreaterOrEqual(t, hcf.GetPlayerFlushLen(), 2)
}

func TestHighCardFlush_Raise_WrongPhase(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	err := hcf.Raise(1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHighCardFlush_Raise_InvalidMultiplier(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	// Two-card flush: max multiplier is 1.
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	hcf.SetAnteBet(100)
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignHeart, 7}, [2]int{domain.CardDesignHeart, 9},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	assert.Equal(t, 1, hcf.MaxRaiseMultiplier())
	err := hcf.Raise(0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	err = hcf.Raise(2)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
}

func TestHighCardFlush_Raise_InsufficientChips(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	hcf.SetAnteBet(100)
	hcf.SetChips(50)
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignHeart, 7}, [2]int{domain.CardDesignHeart, 9},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	err := hcf.Raise(1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestHighCardFlush_MaxRaiseMultiplier(t *testing.T) {
	cases := []struct {
		name      string
		hand      []*domain.Card
		wantMax   int
		wantFLLen int
	}{
		{
			name: "2 cards same suit -> 1x",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignHeart, 7}, [2]int{domain.CardDesignHeart, 9},
				[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
				[2]int{domain.CardDesignDiamond, 8},
			),
			wantMax:   1,
			wantFLLen: 2,
		},
		{
			name: "4 cards same suit -> 1x",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 9},
				[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
				[2]int{domain.CardDesignDiamond, 8},
			),
			wantMax:   1,
			wantFLLen: 4,
		},
		{
			name: "5 cards same suit -> 2x",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 9},
				[2]int{domain.CardDesignSpade, 11}, [2]int{domain.CardDesignClover, 6},
				[2]int{domain.CardDesignDiamond, 8},
			),
			wantMax:   2,
			wantFLLen: 5,
		},
		{
			name: "6 cards same suit -> 3x",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 9},
				[2]int{domain.CardDesignSpade, 11}, [2]int{domain.CardDesignSpade, 13},
				[2]int{domain.CardDesignDiamond, 8},
			),
			wantMax:   3,
			wantFLLen: 6,
		},
		{
			name: "7 cards same suit -> 3x",
			hand: makeHCFCards(
				[2]int{domain.CardDesignHeart, 1}, [2]int{domain.CardDesignHeart, 5},
				[2]int{domain.CardDesignHeart, 7}, [2]int{domain.CardDesignHeart, 9},
				[2]int{domain.CardDesignHeart, 11}, [2]int{domain.CardDesignHeart, 13},
				[2]int{domain.CardDesignHeart, 3},
			),
			wantMax:   3,
			wantFLLen: 7,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hcf := domain.NewDefaultHighCardFlush()
			hcf.SetPlayerHand(tc.hand)
			assert.Equal(t, tc.wantFLLen, hcf.GetPlayerFlushLen())
			assert.Equal(t, tc.wantMax, hcf.MaxRaiseMultiplier())
		})
	}
}

func TestHighCardFlush_Fold_WrongPhase(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	err := hcf.Fold()
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHighCardFlush_Fold_ForfeitsAnte(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	require.NoError(t, hcf.Bet(100, 0, 0))
	chipsBeforeFold := hcf.GetChips()

	require.NoError(t, hcf.Fold())
	assert.Equal(t, domain.HighCardFlushPhaseEnd, hcf.GetPhase())
	assert.True(t, hcf.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, hcf.GetResult())
	// No side bets placed, so no chip change on fold.
	assert.Equal(t, chipsBeforeFold, hcf.GetChips())
}

func TestHighCardFlush_Fold_PaysFlushBonus(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	hcf.SetFlushBonusBet(100)
	// Force a 4-card flush on player hand
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 9},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignHeart, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	chipsBefore := hcf.GetChips()
	require.NoError(t, hcf.Fold())
	// 4-card flush pays 1:1: 100 stake returned + 100 winnings = 200
	expected := chipsBefore + 100*(1+domain.HighCardFlushBonusFour)
	assert.Equal(t, expected, hcf.GetChips())
	assert.Equal(t, 100*(1+domain.HighCardFlushBonusFour), hcf.GetFlushBonusPayout())
}

func TestHighCardFlush_Resolve_DealerNotQualified(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	hcf.SetAnteBet(100)
	hcf.SetRaiseBet(100)
	// Dealer: best flush is 2 cards (max). Doesn't qualify (length < 3).
	hcf.SetDealerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignHeart, 7}, [2]int{domain.CardDesignHeart, 9},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	// Player: 3-card flush (good)
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 10}, [2]int{domain.CardDesignSpade, 11},
		[2]int{domain.CardDesignSpade, 13}, [2]int{domain.CardDesignHeart, 9},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	chipsBefore := hcf.GetChips()
	hcf.SetChips(chipsBefore) // anchor
	// Use Raise to drive resolve(); ante and raise are already set, so we pre-deduct chips
	// by re-setting state and invoking Raise(1) instead. But Raise() re-deducts: simulate the
	// already-paid state by clearing chips of bets first.
	hcf.SetChips(chipsBefore - 200) // simulate ante+raise paid
	hcf.SetRaiseBet(0)              // Raise(1) will set raiseBet=ante
	chipsBeforeResolve := hcf.GetChips()
	require.NoError(t, hcf.Raise(1))

	assert.False(t, hcf.GetDealerQualified())
	// Dealer not qualified: ante pays 1:1 (200 returned), raise pushes (100 returned)
	expected := chipsBeforeResolve - 100 /*raise paid*/ + 200 /*ante payout*/ + 100 /*raise push*/
	assert.Equal(t, expected, hcf.GetChips())
}

func TestHighCardFlush_Resolve_PlayerWins(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	hcf.SetAnteBet(100)
	// Dealer: 3-card flush A-K-9 (qualifies, but smaller than player)
	hcf.SetDealerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 13}, [2]int{domain.CardDesignSpade, 9},
		[2]int{domain.CardDesignSpade, 1}, [2]int{domain.CardDesignHeart, 7},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	// Player: 4-card flush — beats 3-card flush on length alone
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignHeart, 10}, [2]int{domain.CardDesignHeart, 11},
		[2]int{domain.CardDesignHeart, 12}, [2]int{domain.CardDesignHeart, 13},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	hcf.SetChips(500) // post-bet baseline
	require.NoError(t, hcf.Raise(1))

	assert.True(t, hcf.GetDealerQualified())
	assert.Equal(t, domain.GameResultWin, hcf.GetResult())
	// Win: ante 1:1 (200) + raise 1:1 (200) = 400 added; raise of 100 was deducted.
	assert.Equal(t, 500-100+400, hcf.GetChips())
}

func TestHighCardFlush_Resolve_PlayerLoses(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	hcf.SetAnteBet(100)
	// Player: 3-card flush 8-high (qualifies bar for player, dealer must also qualify)
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignHeart, 2}, [2]int{domain.CardDesignHeart, 5},
		[2]int{domain.CardDesignHeart, 8}, [2]int{domain.CardDesignDiamond, 4},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignSpade, 1},
	))
	// Dealer: 4-card flush 10-high (qualifies, beats player)
	hcf.SetDealerHand(makeHCFCards(
		[2]int{domain.CardDesignClover, 10}, [2]int{domain.CardDesignClover, 9},
		[2]int{domain.CardDesignClover, 7}, [2]int{domain.CardDesignClover, 3},
		[2]int{domain.CardDesignHeart, 4}, [2]int{domain.CardDesignSpade, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	hcf.SetChips(500)
	require.NoError(t, hcf.Raise(1))
	assert.True(t, hcf.GetDealerQualified())
	assert.Equal(t, domain.GameResultLose, hcf.GetResult())
	// Lose: ante and raise lost. Raise of 100 was deducted; no return.
	assert.Equal(t, 400, hcf.GetChips())
}

func TestHighCardFlush_Resolve_Push(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	hcf.SetAnteBet(100)
	// Identical 3-card flushes (different suits but same lengths and ranks): tie
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignHeart, 11}, [2]int{domain.CardDesignHeart, 9},
		[2]int{domain.CardDesignHeart, 7}, [2]int{domain.CardDesignSpade, 4},
		[2]int{domain.CardDesignClover, 5}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	hcf.SetDealerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 11}, [2]int{domain.CardDesignSpade, 9},
		[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignHeart, 4},
		[2]int{domain.CardDesignClover, 5}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	hcf.SetChips(500)
	require.NoError(t, hcf.Raise(1))
	assert.True(t, hcf.GetDealerQualified())
	assert.Equal(t, domain.GameResultDraw, hcf.GetResult())
	// Push: ante and raise returned at 1×. Raise of 100 deducted, 100+100=200 returned.
	assert.Equal(t, 500-100+100+100, hcf.GetChips())
}

func TestHighCardFlush_DealerQualify_HighCardThreshold(t *testing.T) {
	// Dealer has 3-card flush but with 8-high → does not qualify.
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	hcf.SetAnteBet(100)
	hcf.SetDealerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 8}, [2]int{domain.CardDesignHeart, 7},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 9},
	))
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignHeart, 10}, [2]int{domain.CardDesignHeart, 11},
		[2]int{domain.CardDesignHeart, 12}, [2]int{domain.CardDesignClover, 9},
		[2]int{domain.CardDesignSpade, 4}, [2]int{domain.CardDesignSpade, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	hcf.SetChips(500)
	require.NoError(t, hcf.Raise(1))
	assert.False(t, hcf.GetDealerQualified())
}

func TestHighCardFlush_FlushBonusPaytable(t *testing.T) {
	cases := []struct {
		name     string
		hand     []*domain.Card
		wantMult int
	}{
		{
			name: "no flush ≥ 4 → no payout",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 8}, [2]int{domain.CardDesignHeart, 7},
				[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: 0,
		},
		{
			name: "4-card flush → 1:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 8}, [2]int{domain.CardDesignSpade, 13},
				[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: domain.HighCardFlushBonusFour,
		},
		{
			name: "5-card flush → 10:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 8}, [2]int{domain.CardDesignSpade, 13},
				[2]int{domain.CardDesignSpade, 11}, [2]int{domain.CardDesignClover, 6},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: domain.HighCardFlushBonusFive,
		},
		{
			name: "6-card flush → 50:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 8}, [2]int{domain.CardDesignSpade, 13},
				[2]int{domain.CardDesignSpade, 11}, [2]int{domain.CardDesignSpade, 12},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: domain.HighCardFlushBonusSix,
		},
		{
			name: "7-card flush → 100:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 8}, [2]int{domain.CardDesignSpade, 13},
				[2]int{domain.CardDesignSpade, 11}, [2]int{domain.CardDesignSpade, 12},
				[2]int{domain.CardDesignSpade, 9},
			),
			wantMult: domain.HighCardFlushBonusSeven,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hcf := domain.NewDefaultHighCardFlush()
			hcf.SetPhase(domain.HighCardFlushPhaseAction)
			hcf.SetFlushBonusBet(100)
			hcf.SetPlayerHand(tc.hand)
			require.NoError(t, hcf.Fold())
			if tc.wantMult == 0 {
				assert.Equal(t, 0, hcf.GetFlushBonusPayout())
			} else {
				assert.Equal(t, 100+100*tc.wantMult, hcf.GetFlushBonusPayout())
			}
		})
	}
}

func TestHighCardFlush_StraightFlushBonusPaytable(t *testing.T) {
	cases := []struct {
		name     string
		hand     []*domain.Card
		wantMult int
	}{
		{
			// 4-card flush in Spades with gapped ranks {2,5,8,11} — no run of 3+.
			name: "no straight flush → no payout",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
				[2]int{domain.CardDesignSpade, 8}, [2]int{domain.CardDesignSpade, 11},
				[2]int{domain.CardDesignHeart, 7}, [2]int{domain.CardDesignClover, 4},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: 0,
		},
		{
			name: "3-card straight flush → 8:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 5}, [2]int{domain.CardDesignSpade, 6},
				[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignHeart, 13},
				[2]int{domain.CardDesignClover, 11}, [2]int{domain.CardDesignClover, 12},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: domain.HighCardFlushSFBonusThree,
		},
		{
			name: "4-card straight flush → 60:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 5}, [2]int{domain.CardDesignSpade, 6},
				[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 8},
				[2]int{domain.CardDesignClover, 11}, [2]int{domain.CardDesignClover, 12},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: domain.HighCardFlushSFBonusFour,
		},
		{
			name: "5-card straight flush → 100:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 5}, [2]int{domain.CardDesignSpade, 6},
				[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 8},
				[2]int{domain.CardDesignSpade, 9}, [2]int{domain.CardDesignClover, 12},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: domain.HighCardFlushSFBonusFive,
		},
		{
			name: "6-card straight flush → 1000:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 5}, [2]int{domain.CardDesignSpade, 6},
				[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 8},
				[2]int{domain.CardDesignSpade, 9}, [2]int{domain.CardDesignSpade, 10},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: domain.HighCardFlushSFBonusSix,
		},
		{
			name: "7-card straight flush → 8000:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignSpade, 5}, [2]int{domain.CardDesignSpade, 6},
				[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 8},
				[2]int{domain.CardDesignSpade, 9}, [2]int{domain.CardDesignSpade, 10},
				[2]int{domain.CardDesignSpade, 11},
			),
			wantMult: domain.HighCardFlushSFBonusSeven,
		},
		{
			name: "wheel A-2-3 straight flush → 8:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignClover, 1}, [2]int{domain.CardDesignClover, 2},
				[2]int{domain.CardDesignClover, 3}, [2]int{domain.CardDesignHeart, 13},
				[2]int{domain.CardDesignSpade, 11}, [2]int{domain.CardDesignSpade, 12},
				[2]int{domain.CardDesignDiamond, 9},
			),
			wantMult: domain.HighCardFlushSFBonusThree,
		},
		{
			name: "Q-K-A straight flush → 8:1",
			hand: makeHCFCards(
				[2]int{domain.CardDesignClover, 12}, [2]int{domain.CardDesignClover, 13},
				[2]int{domain.CardDesignClover, 1}, [2]int{domain.CardDesignHeart, 9},
				[2]int{domain.CardDesignSpade, 11}, [2]int{domain.CardDesignSpade, 6},
				[2]int{domain.CardDesignDiamond, 4},
			),
			wantMult: domain.HighCardFlushSFBonusThree,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hcf := domain.NewDefaultHighCardFlush()
			hcf.SetPhase(domain.HighCardFlushPhaseAction)
			hcf.SetStraightFlushBet(100)
			hcf.SetPlayerHand(tc.hand)
			require.NoError(t, hcf.Fold())
			if tc.wantMult == 0 {
				assert.Equal(t, 0, hcf.GetStraightFlushPayout())
			} else {
				assert.Equal(t, 100+100*tc.wantMult, hcf.GetStraightFlushPayout())
			}
		})
	}
}

func TestHighCardFlush_TotalPayout(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetAnteBet(50)
	hcf.SetRaiseBet(50)
	hcf.SetFlushBonusBet(20)
	hcf.SetStraightFlushBet(10)
	hcf.SetDealerQualified(true)
	hcf.SetResult(domain.GameResultWin)
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 5}, [2]int{domain.CardDesignSpade, 6},
		[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 8},
		[2]int{domain.CardDesignSpade, 9}, [2]int{domain.CardDesignSpade, 10},
		[2]int{domain.CardDesignDiamond, 9},
	))
	// Drive resolve by calling Raise(1) — but state is already in Action phase via fields.
	// Just ensure GetTotalPayout sums correctly with manual values.
	hcf.SetPhase(domain.HighCardFlushPhaseAction)
	require.NoError(t, hcf.Raise(1)) // ante was 50; raise(1)=50 chips deducted again
	total := hcf.GetAntePayout() + hcf.GetRaisePayout() + hcf.GetFlushBonusPayout() + hcf.GetStraightFlushPayout()
	assert.Equal(t, total, hcf.GetTotalPayout())
}

func TestHighCardFlush_FullRound_RaiseAndDealResults(t *testing.T) {
	// Drive a full round through Bet → Raise to ensure deterministic chip accounting.
	hcf := domain.NewDefaultHighCardFlush()
	startChips := hcf.GetChips()
	require.NoError(t, hcf.Bet(100, 0, 0))
	assert.Equal(t, startChips-100, hcf.GetChips())
	// Force player win path
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignHeart, 1}, [2]int{domain.CardDesignHeart, 13},
		[2]int{domain.CardDesignHeart, 12}, [2]int{domain.CardDesignHeart, 11},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	hcf.SetDealerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 2}, [2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignHeart, 9},
		[2]int{domain.CardDesignClover, 4}, [2]int{domain.CardDesignClover, 6},
		[2]int{domain.CardDesignDiamond, 8},
	))
	require.NoError(t, hcf.Raise(1))
	assert.True(t, hcf.GetGameEndFlag())
}

func TestHighCardFlush_ActionLog(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	require.NoError(t, hcf.Bet(100, 0, 0))
	require.NoError(t, hcf.Fold())
	log := hcf.GetActionLog()
	assert.GreaterOrEqual(t, len(log), 3) // bet, deal, fold, result
	assert.Equal(t, "bet", log[0].ActionType)
}

func TestHighCardFlush_JSONRoundTrip(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	require.NoError(t, hcf.Bet(100, 50, 20))
	data, err := json.Marshal(hcf)
	require.NoError(t, err)

	var loaded domain.HighCardFlush
	require.NoError(t, json.Unmarshal(data, &loaded))
	assert.Equal(t, hcf.GetPhase(), loaded.GetPhase())
	assert.Equal(t, hcf.GetAnteBet(), loaded.GetAnteBet())
	assert.Equal(t, hcf.GetFlushBonusBet(), loaded.GetFlushBonusBet())
	assert.Equal(t, hcf.GetStraightFlushBet(), loaded.GetStraightFlushBet())
	assert.Equal(t, hcf.GetChips(), loaded.GetChips())
	assert.Equal(t, len(hcf.GetPlayerHand()), len(loaded.GetPlayerHand()))
}

func TestHighCardFlush_JSONUnmarshal_OverflowRejected(t *testing.T) {
	// Construct a payload with a player-hand array that exceeds the cap.
	// Build via direct JSON to avoid touching the in-memory model.
	const longHandJSON = `{"ph":[`
	const maxLen = 1001
	payload := longHandJSON
	for i := 0; i < maxLen; i++ {
		if i > 0 {
			payload += ","
		}
		payload += `{"design":1,"value":2,"draw":true}`
	}
	payload += `]}`
	var hcf domain.HighCardFlush
	err := json.Unmarshal([]byte(payload), &hcf)
	require.Error(t, err)
}

func TestHighCardFlush_JSONUnmarshal_GarbageRejected(t *testing.T) {
	var hcf domain.HighCardFlush
	err := json.Unmarshal([]byte("not json"), &hcf)
	require.Error(t, err)
}

func TestHighCardFlush_EvalHandSkipsNilCard(t *testing.T) {
	// evalHighCardFlushHand and evalLongestStraightFlushLen both have defensive
	// nil-card guards — exercise them via SetPlayerHand so the package coverage
	// records the branch.
	hcf := domain.NewDefaultHighCardFlush()
	hand := []*domain.Card{
		nil, // skip
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
	}
	hcf.SetPlayerHand(hand)
	assert.Equal(t, 3, hcf.GetPlayerFlushLen())
	assert.Equal(t, 3, hcf.GetPlayerStraightFlushLen())
}

func TestHighCardFlush_DealerAndStraightFlushGetters(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.SetDealerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 5}, [2]int{domain.CardDesignSpade, 6},
		[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 8},
		[2]int{domain.CardDesignHeart, 9}, [2]int{domain.CardDesignClover, 4},
		[2]int{domain.CardDesignDiamond, 10},
	))
	assert.Equal(t, 4, hcf.GetDealerFlushLen())

	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 5}, [2]int{domain.CardDesignSpade, 6},
		[2]int{domain.CardDesignSpade, 7}, [2]int{domain.CardDesignSpade, 8},
		[2]int{domain.CardDesignHeart, 9}, [2]int{domain.CardDesignClover, 4},
		[2]int{domain.CardDesignDiamond, 10},
	))
	assert.Equal(t, 4, hcf.GetPlayerStraightFlushLen())

	hcf.SetGameEndFlag(true)
	assert.True(t, hcf.GetGameEndFlag())
}

// #5607: 「4枚フラッシュ」と長さだけ言われても、7枚のうちどれがその4枚なのかは
// 自分で数えるしかなかった。どのスートで数えた長さなのかはドメインが既に
// 決めている (evalHighCardFlushHand の Suit) ので、それを名前を付けて出す。
func TestHighCardFlushExposesTheSuitItCountedTheFlushIn(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.Reset()
	require.NoError(t, hcf.Bet(100, 0, 0))

	suit := hcf.GetPlayerFlushSuit()
	// 数えた長さと、そのスートの実際の枚数が一致すること。
	count := 0
	for _, c := range hcf.GetPlayerHand() {
		if c.GetDesign() == suit {
			count++
		}
	}
	assert.Equal(t, hcf.GetPlayerFlushLen(), count,
		"公開したスートの枚数が、公開した長さと一致する")
}

// 同着のときにどちらを採るかを固定する。**「どちらでもよい」ではない** --
// 画面の印と長さの行が別のスートを指すと、数えても合わなくなる。
func TestHighCardFlushBreaksASuitTieByTheHighCards(t *testing.T) {
	hcf := domain.NewDefaultHighCardFlush()
	hcf.Reset()
	// ♠ と ♥ がどちらも 3 枚。♥ の方が高い札を持つので ♥ が選ばれる。
	// SetPlayerHand は長さもスートも計算し直す (テスト用のセッタ)。
	hcf.SetPlayerHand(makeHCFCards(
		[2]int{domain.CardDesignSpade, 2},
		[2]int{domain.CardDesignSpade, 3},
		[2]int{domain.CardDesignSpade, 4},
		[2]int{domain.CardDesignHeart, 13},
		[2]int{domain.CardDesignHeart, 12},
		[2]int{domain.CardDesignHeart, 11},
		[2]int{domain.CardDesignClover, 5},
	))

	assert.Equal(t, domain.CardDesignHeart, hcf.GetPlayerFlushSuit())
	assert.Equal(t, 3, hcf.GetPlayerFlushLen())
}
