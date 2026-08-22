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
type cdrCard struct {
	d, v int
}

func makeCdrHand(specs ...cdrCard) []*domain.Card {
	cards := make([]*domain.Card, len(specs))
	for i, s := range specs {
		cards[i] = domain.NewCard(s.d, s.v, false)
	}
	return cards
}

func TestNewDefaultCaribbeanDraw(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	assert.Equal(t, domain.CaribbeanDrawPhaseBet, cs.GetPhase())
	assert.Equal(t, domain.CaribbeanDrawDefaultChips, cs.GetChips())
	assert.False(t, cs.GetGameEndFlag())
	assert.Nil(t, cs.GetPlayerHand())
	assert.Nil(t, cs.GetDealerHand())
}

func TestCaribbeanDraw_Reset(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
	require.NoError(t, cs.Play())
	assert.Equal(t, domain.CaribbeanDrawPhaseEnd, cs.GetPhase())

	cs.Reset()
	assert.Equal(t, domain.CaribbeanDrawPhaseBet, cs.GetPhase())
	assert.False(t, cs.GetGameEndFlag())
	assert.Nil(t, cs.GetPlayerHand())
	assert.Nil(t, cs.GetDealerHand())
	assert.Equal(t, 0, cs.GetAnteBet())
	assert.Equal(t, 0, cs.GetJackpotBet())
	assert.Equal(t, 0, cs.GetPlayBet())
}

func TestCaribbeanDraw_Reset_RefillChips(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	cs.SetChips(5)
	cs.Reset()
	assert.Equal(t, domain.CaribbeanDrawDefaultChips, cs.GetChips())
}

func TestCaribbeanDraw_Bet_WrongPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	cs.SetPhase(domain.CaribbeanDrawPhaseAction)
	err := cs.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCaribbeanDraw_Bet_InvalidAnteAmount(t *testing.T) {
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
			cs := domain.NewDefaultCaribbeanDraw()
			err := cs.Bet(tt.ante, 0)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestCaribbeanDraw_Bet_InvalidJackpotAmount(t *testing.T) {
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
			cs := domain.NewDefaultCaribbeanDraw()
			err := cs.Bet(100, tt.jackpot)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestCaribbeanDraw_Bet_InsufficientChips(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	cs.SetChips(50)
	err := cs.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestCaribbeanDraw_Bet_Success(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	err := cs.Bet(100, 50)
	assert.NoError(t, err)
	// **ベットの次はドロー。** クローン元はここで直接アクションに入る。
	assert.Equal(t, domain.CaribbeanDrawPhaseDraw, cs.GetPhase())
	assert.Equal(t, 100, cs.GetAnteBet())
	assert.Equal(t, 50, cs.GetJackpotBet())
	assert.Len(t, cs.GetPlayerHand(), 5)
	assert.Len(t, cs.GetDealerHand(), 5)
	assert.Equal(t, domain.CaribbeanDrawDefaultChips-150, cs.GetChips())
}

func TestCaribbeanDraw_Bet_AnteOnlyNoJackpot(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	err := cs.Bet(100, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, cs.GetJackpotBet())
	assert.Equal(t, domain.CaribbeanDrawDefaultChips-100, cs.GetChips())
}

func TestCaribbeanDraw_Play_WrongPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	err := cs.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCaribbeanDraw_Play_InsufficientChips(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Draw(nil))
	cs.SetChips(0)
	err := cs.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestCaribbeanDraw_Play_PlacesDoubleAnte(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
	require.NoError(t, cs.Play())
	assert.Equal(t, 200, cs.GetPlayBet())
	assert.Equal(t, domain.CaribbeanDrawPhaseEnd, cs.GetPhase())
	assert.True(t, cs.GetGameEndFlag())
}

func TestCaribbeanDraw_Fold_WrongPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	err := cs.Fold()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCaribbeanDraw_Fold_Success(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Draw(nil))
	chipsBefore := cs.GetChips()
	require.NoError(t, cs.Fold())
	assert.Equal(t, domain.CaribbeanDrawPhaseEnd, cs.GetPhase())
	assert.True(t, cs.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, cs.GetResult())
	assert.Equal(t, chipsBefore, cs.GetChips())
}

func TestCaribbeanDraw_Fold_JackpotStillEvaluated(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 10))

	// Player hand: a flush (jackpot pays 50:1)
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 2},
		cdrCard{domain.CardDesignSpade, 5},
		cdrCard{domain.CardDesignSpade, 7},
		cdrCard{domain.CardDesignSpade, 9},
		cdrCard{domain.CardDesignSpade, 11},
	))
	require.NoError(t, cs.Draw(nil))
	chipsBefore := cs.GetChips()
	require.NoError(t, cs.Fold())
	assert.Equal(t, 10*domain.CaribbeanDrawJackpotFlush, cs.GetJackpotPayout())
	assert.Equal(t, chipsBefore+10*domain.CaribbeanDrawJackpotFlush, cs.GetChips())
}

func TestCaribbeanDraw_DealerQualification(t *testing.T) {
	tests := []struct {
		name      string
		dealer    []*domain.Card
		qualified bool
	}{
		// **8 のペア以上。** クローン元 (Caribbean Stud) の「ペア以上、または
		// A-K ハイ」ではない。境界の両側と、A-K ハイが落ちることを押さえる。
		{
			name: "PairOfTwosDoesNotQualify",
			dealer: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignClover, 2},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 7},
				cdrCard{domain.CardDesignSpade, 9}),
			qualified: false,
		},
		{
			name: "PairOfSevensDoesNotQualify",
			dealer: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignClover, 7},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 3},
				cdrCard{domain.CardDesignSpade, 9}),
			qualified: false,
		},
		{
			name: "PairOfEightsQualifies",
			dealer: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 8}, cdrCard{domain.CardDesignClover, 8},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 3},
				cdrCard{domain.CardDesignSpade, 9}),
			qualified: true,
		},
		{
			// **A は 14 として比べる。** 1 のまま比べると最強のペアが 8 の
			// ペアに負け、A のペアがクオリファイしないという逆転が起きる。
			name: "PairOfAcesQualifies",
			dealer: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 1}, cdrCard{domain.CardDesignClover, 1},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 3},
				cdrCard{domain.CardDesignSpade, 9}),
			qualified: true,
		},
		{
			// **クローン元なら通る手。** A-K ハイはスタッドではクオリファイ
			// するが、こちらはペアが要る。
			name: "AceKingHighDoesNotQualify",
			dealer: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 1}, cdrCard{domain.CardDesignClover, 13},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 7},
				cdrCard{domain.CardDesignSpade, 9}),
			qualified: false,
		},
		{
			name: "AceQueenHighDoesNotQualify",
			dealer: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 1}, cdrCard{domain.CardDesignClover, 12},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 7},
				cdrCard{domain.CardDesignSpade, 9}),
			qualified: false,
		},
		{
			// ツーペア以上はペアの大小を見るまでもなく足りる。
			name: "TwoPairOfLowCardsQualifies",
			dealer: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignClover, 2},
				cdrCard{domain.CardDesignHeart, 3}, cdrCard{domain.CardDesignDiamond, 3},
				cdrCard{domain.CardDesignSpade, 9}),
			qualified: true,
		},
		{
			name: "KingHighDoesNotQualify",
			dealer: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 13}, cdrCard{domain.CardDesignClover, 12},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 7},
				cdrCard{domain.CardDesignSpade, 9}),
			qualified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := domain.NewDefaultCaribbeanDraw()
			require.NoError(t, cs.Bet(100, 0))

			// Weak player hand (Ace high) so dealer wins via kicker if needed
			cs.SetPlayerHand(makeCdrHand(
				cdrCard{domain.CardDesignSpade, 1}, cdrCard{domain.CardDesignClover, 4},
				cdrCard{domain.CardDesignHeart, 6}, cdrCard{domain.CardDesignDiamond, 8},
				cdrCard{domain.CardDesignSpade, 10}))
			cs.SetDealerHand(tt.dealer)
			require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
			require.NoError(t, cs.Play())
			assert.Equal(t, tt.qualified, cs.GetDealerQualified())
		})
	}
}

func TestCaribbeanDraw_Payouts_DealerNotQualified(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))

	// Player: high card; Dealer: Queen high (does not qualify)
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 1}, cdrCard{domain.CardDesignClover, 4},
		cdrCard{domain.CardDesignHeart, 6}, cdrCard{domain.CardDesignDiamond, 8},
		cdrCard{domain.CardDesignSpade, 10}))
	cs.SetDealerHand(makeCdrHand(
		cdrCard{domain.CardDesignDiamond, 12}, cdrCard{domain.CardDesignHeart, 5},
		cdrCard{domain.CardDesignClover, 3}, cdrCard{domain.CardDesignSpade, 7},
		cdrCard{domain.CardDesignDiamond, 9}))

	require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
	require.NoError(t, cs.Play())
	assert.False(t, cs.GetDealerQualified())
	assert.Equal(t, 200, cs.GetAntePayout()) // ante 1:1
	assert.Equal(t, 200, cs.GetPlayPayout()) // play push (returned, equal to playBet=200)
}

func TestCaribbeanDraw_Payouts_PlayerWinsFlush(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))

	// Player: flush; Dealer: pair (qualifies but loses)
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignSpade, 5},
		cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignSpade, 9},
		cdrCard{domain.CardDesignSpade, 11}))
	cs.SetDealerHand(makeCdrHand(
		cdrCard{domain.CardDesignDiamond, 9}, cdrCard{domain.CardDesignHeart, 9},
		cdrCard{domain.CardDesignClover, 6}, cdrCard{domain.CardDesignSpade, 3},
		cdrCard{domain.CardDesignDiamond, 10}))

	require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
	require.NoError(t, cs.Play())
	assert.True(t, cs.GetDealerQualified())
	assert.Equal(t, domain.GameResultWin, cs.GetResult())
	assert.Equal(t, 200, cs.GetAntePayout())
	// playBet=200、フラッシュの倍率は 4 (クローン元は 5) → 200 + 200*4 = 1000
	assert.Equal(t, 200+200*domain.CaribbeanDrawPayFlush, cs.GetPlayPayout())
	assert.Equal(t, 1000, cs.GetPlayPayout(), "倍率を定数から作ると値そのものは固定されない")
}

func TestCaribbeanDraw_Payouts_PlayerLoses(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))

	// プレイヤーはハイカード、ディーラーは **9 のペア** —— クオリファイし、かつ
	// 勝つ手。8 未満のペアではクオリファイせず、負けたはずのアンテが払い戻される。
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 1}, cdrCard{domain.CardDesignClover, 4},
		cdrCard{domain.CardDesignHeart, 6}, cdrCard{domain.CardDesignDiamond, 8},
		cdrCard{domain.CardDesignSpade, 10}))
	cs.SetDealerHand(makeCdrHand(
		cdrCard{domain.CardDesignDiamond, 9}, cdrCard{domain.CardDesignHeart, 9},
		cdrCard{domain.CardDesignClover, 6}, cdrCard{domain.CardDesignSpade, 3},
		cdrCard{domain.CardDesignDiamond, 11}))

	require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
	require.NoError(t, cs.Play())
	assert.True(t, cs.GetDealerQualified())
	assert.Equal(t, domain.GameResultLose, cs.GetResult())
	assert.Equal(t, 0, cs.GetAntePayout())
	assert.Equal(t, 0, cs.GetPlayPayout())
}

func TestCaribbeanDraw_Payouts_Push(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))

	// 役位もキッカーも同じ → 引き分け。**ペアは 10 にする** —— 8 未満だと
	// ディーラーがクオリファイせず、引き分けではなく未クオリファイの経路に入る。
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 10}, cdrCard{domain.CardDesignClover, 10},
		cdrCard{domain.CardDesignHeart, 7}, cdrCard{domain.CardDesignDiamond, 9},
		cdrCard{domain.CardDesignSpade, 11}))
	cs.SetDealerHand(makeCdrHand(
		cdrCard{domain.CardDesignDiamond, 10}, cdrCard{domain.CardDesignHeart, 10},
		cdrCard{domain.CardDesignClover, 7}, cdrCard{domain.CardDesignSpade, 9},
		cdrCard{domain.CardDesignDiamond, 11}))

	require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
	require.NoError(t, cs.Play())
	assert.True(t, cs.GetDealerQualified())
	assert.Equal(t, domain.GameResultDraw, cs.GetResult())
	assert.Equal(t, 100, cs.GetAntePayout()) // push (return)
	assert.Equal(t, 200, cs.GetPlayPayout()) // push (return)
}

func TestCaribbeanDraw_PlayMultipliers(t *testing.T) {
	tests := []struct {
		name       string
		hand       []*domain.Card
		multiplier int
	}{
		{
			name: "RoyalFlush",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 1}, cdrCard{domain.CardDesignSpade, 10},
				cdrCard{domain.CardDesignSpade, 11}, cdrCard{domain.CardDesignSpade, 12},
				cdrCard{domain.CardDesignSpade, 13}),
			multiplier: domain.CaribbeanDrawPayRoyalFlush,
		},
		{
			name: "StraightFlush",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignClover, 5}, cdrCard{domain.CardDesignClover, 6},
				cdrCard{domain.CardDesignClover, 7}, cdrCard{domain.CardDesignClover, 8},
				cdrCard{domain.CardDesignClover, 9}),
			multiplier: domain.CaribbeanDrawPayStraightFlush,
		},
		{
			name: "FourOfAKind",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignClover, 7},
				cdrCard{domain.CardDesignHeart, 7}, cdrCard{domain.CardDesignDiamond, 7},
				cdrCard{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanDrawPayFourOfAKind,
		},
		{
			name: "FullHouse",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignClover, 7},
				cdrCard{domain.CardDesignHeart, 7}, cdrCard{domain.CardDesignDiamond, 2},
				cdrCard{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanDrawPayFullHouse,
		},
		{
			name: "Straight",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 5}, cdrCard{domain.CardDesignClover, 6},
				cdrCard{domain.CardDesignHeart, 7}, cdrCard{domain.CardDesignDiamond, 8},
				cdrCard{domain.CardDesignClover, 9}),
			multiplier: domain.CaribbeanDrawPayStraight,
		},
		{
			name: "ThreeOfAKind",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignClover, 7},
				cdrCard{domain.CardDesignHeart, 7}, cdrCard{domain.CardDesignDiamond, 4},
				cdrCard{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanDrawPayThreeOfAKind,
		},
		{
			name: "TwoPair",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignClover, 7},
				cdrCard{domain.CardDesignHeart, 4}, cdrCard{domain.CardDesignDiamond, 4},
				cdrCard{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanDrawPayTwoPair,
		},
		{
			// ディーラーが 8 のペアでクオリファイするので、**それに勝つ**
			// ペアでないとプレイ配当の経路に入らない。
			name: "OnePair",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 11}, cdrCard{domain.CardDesignClover, 11},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 4},
				cdrCard{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanDrawPayPair,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := domain.NewDefaultCaribbeanDraw()
			cs.SetChips(100000)
			require.NoError(t, cs.Bet(100, 0))
			cs.SetPlayerHand(tt.hand)
			// Dealer: Queen high (does not qualify) → force play multiplier path manually
			// Use a weaker dealer hand BUT qualifying so player wins.
			// **8 のペア。** クローン元では 3 のペアで足りたが、この卓は
			// 8 以上でないとクオリファイしない。
			cs.SetDealerHand(makeCdrHand(
				cdrCard{domain.CardDesignDiamond, 8}, cdrCard{domain.CardDesignHeart, 8},
				cdrCard{domain.CardDesignClover, 6}, cdrCard{domain.CardDesignSpade, 4},
				cdrCard{domain.CardDesignDiamond, 10}))
			require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
			require.NoError(t, cs.Play())
			assert.Equal(t, domain.GameResultWin, cs.GetResult())
			expected := 200 + 200*tt.multiplier
			assert.Equal(t, expected, cs.GetPlayPayout())
		})
	}
}

func TestCaribbeanDraw_JackpotPayouts(t *testing.T) {
	tests := []struct {
		name       string
		hand       []*domain.Card
		multiplier int
	}{
		{
			name: "RoyalFlushJackpot",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 1}, cdrCard{domain.CardDesignSpade, 10},
				cdrCard{domain.CardDesignSpade, 11}, cdrCard{domain.CardDesignSpade, 12},
				cdrCard{domain.CardDesignSpade, 13}),
			multiplier: domain.CaribbeanDrawJackpotRoyalFlush,
		},
		{
			name: "StraightFlushJackpot",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignClover, 5}, cdrCard{domain.CardDesignClover, 6},
				cdrCard{domain.CardDesignClover, 7}, cdrCard{domain.CardDesignClover, 8},
				cdrCard{domain.CardDesignClover, 9}),
			multiplier: domain.CaribbeanDrawJackpotStraightFlush,
		},
		{
			name: "FourOfAKindJackpot",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignClover, 7},
				cdrCard{domain.CardDesignHeart, 7}, cdrCard{domain.CardDesignDiamond, 7},
				cdrCard{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanDrawJackpotFourOfAKind,
		},
		{
			name: "FullHouseJackpot",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignClover, 7},
				cdrCard{domain.CardDesignHeart, 7}, cdrCard{domain.CardDesignDiamond, 2},
				cdrCard{domain.CardDesignSpade, 2}),
			multiplier: domain.CaribbeanDrawJackpotFullHouse,
		},
		{
			name: "FlushJackpot",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignSpade, 5},
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignSpade, 9},
				cdrCard{domain.CardDesignSpade, 11}),
			multiplier: domain.CaribbeanDrawJackpotFlush,
		},
		{
			name: "NoJackpotPair",
			hand: makeCdrHand(
				cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignClover, 7},
				cdrCard{domain.CardDesignHeart, 5}, cdrCard{domain.CardDesignDiamond, 4},
				cdrCard{domain.CardDesignSpade, 2}),
			multiplier: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := domain.NewDefaultCaribbeanDraw()
			cs.SetChips(100000)
			require.NoError(t, cs.Bet(100, 10))
			cs.SetPlayerHand(tt.hand)
			// **8 のペア。** クローン元では 3 のペアで足りたが、この卓は
			// 8 以上でないとクオリファイしない。
			cs.SetDealerHand(makeCdrHand(
				cdrCard{domain.CardDesignDiamond, 8}, cdrCard{domain.CardDesignHeart, 8},
				cdrCard{domain.CardDesignClover, 6}, cdrCard{domain.CardDesignSpade, 4},
				cdrCard{domain.CardDesignDiamond, 10}))
			require.NoError(t, cs.Draw(nil)) // ドローフェーズを素通りする（交換しない）
			require.NoError(t, cs.Play())
			assert.Equal(t, 10*tt.multiplier, cs.GetJackpotPayout())
		})
	}
}

func TestCaribbeanDraw_TotalPayoutZero(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	assert.Equal(t, 0, cs.GetTotalPayout())
}

func TestCaribbeanDraw_GetActionLog(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Draw(nil))
	require.NoError(t, cs.Play())
	assert.NotEmpty(t, cs.GetActionLog())
}

func TestCaribbeanDraw_JSONRoundTrip(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 10))
	require.NoError(t, cs.Draw([]int{0, 1}))
	require.NoError(t, cs.Play())

	data, err := json.Marshal(cs)
	require.NoError(t, err)

	var restored domain.CaribbeanDraw
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

func TestCaribbeanDraw_UnmarshalJSON_InvalidData(t *testing.T) {
	var cs domain.CaribbeanDraw
	err := cs.UnmarshalJSON([]byte("not json"))
	assert.Error(t, err)
}

func TestCaribbeanDraw_SetSettersExposeFields(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	cs.SetPlayerHand(makeCdrHand(cdrCard{domain.CardDesignSpade, 1}))
	cs.SetDealerHand(makeCdrHand(cdrCard{domain.CardDesignClover, 13}))
	cs.SetAnteBet(100)
	cs.SetJackpotBet(20)
	cs.SetPlayBet(200)
	assert.Len(t, cs.GetPlayerHand(), 1)
	assert.Len(t, cs.GetDealerHand(), 1)
	assert.Equal(t, 100, cs.GetAnteBet())
	assert.Equal(t, 20, cs.GetJackpotBet())
	assert.Equal(t, 200, cs.GetPlayBet())
}

// --- ドローフェーズ: クローン元 (Caribbean Stud) に無い、このゲームの本体 ---

func TestCaribbeanDraw_DrawReplacesExactlyTheNamedCards(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	before := append([]*domain.Card(nil), cs.GetPlayerHand()...)
	require.Len(t, before, 5)

	require.NoError(t, cs.Draw([]int{0, 3}))
	after := cs.GetPlayerHand()
	require.Len(t, after, 5, "枚数は変わらない")

	// 指名した札は差し替わり、残りはそのまま。**添字がずれると「別の札が
	// 消える」形で壊れる**ので、触っていない 3 枚の同一性まで見る。
	assert.NotSame(t, before[0], after[0], "0 番は引き直される")
	assert.NotSame(t, before[3], after[3], "3 番は引き直される")
	assert.Same(t, before[1], after[1])
	assert.Same(t, before[2], after[2])
	assert.Same(t, before[4], after[4])
}

func TestCaribbeanDraw_StandingPatCostsNothingAndKeepsTheHand(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	before := append([]*domain.Card(nil), cs.GetPlayerHand()...)
	chips := cs.GetChips()

	require.NoError(t, cs.Draw(nil))
	assert.Equal(t, chips, cs.GetChips(), "引かなければ手数料は取らない")
	assert.Equal(t, 0, cs.GetDrawCost())
	assert.Equal(t, before, cs.GetPlayerHand())
	assert.Equal(t, domain.CaribbeanDrawPhaseAction, cs.GetPhase())
}

func TestCaribbeanDraw_DrawingChargesTheAnte(t *testing.T) {
	// **交換はタダではない。** 無料なら常に引くのが最適になり、このゲームが
	// クローン元に足している唯一の判断が消える。
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	chips := cs.GetChips()

	require.NoError(t, cs.Draw([]int{2}))
	assert.Equal(t, chips-100, cs.GetChips(), "手数料はアンテと同額")
	assert.Equal(t, 100, cs.GetDrawCost())
}

func TestCaribbeanDraw_DrawRejectsMoreThanTwoCards(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	before := append([]*domain.Card(nil), cs.GetPlayerHand()...)
	chips := cs.GetChips()

	err := cs.Draw([]int{0, 1, 2})
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	assert.Equal(t, before, cs.GetPlayerHand(), "弾いた交換で手札は動かない")
	assert.Equal(t, chips, cs.GetChips(), "弾いた交換で手数料は取らない")
	assert.Equal(t, domain.CaribbeanDrawPhaseDraw, cs.GetPhase(), "フェーズも進めない")
}

func TestCaribbeanDraw_DrawRejectsBadIndices(t *testing.T) {
	for _, tt := range []struct {
		name    string
		indices []int
	}{
		{"negative", []int{-1}},
		{"past the end", []int{5}},
		{"the same card twice", []int{2, 2}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cs := domain.NewDefaultCaribbeanDraw()
			require.NoError(t, cs.Bet(100, 0))
			before := append([]*domain.Card(nil), cs.GetPlayerHand()...)
			chips := cs.GetChips()

			err := cs.Draw(tt.indices)
			require.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
			assert.Equal(t, before, cs.GetPlayerHand())
			assert.Equal(t, chips, cs.GetChips())
		})
	}
}

func TestCaribbeanDraw_DrawIsRejectedOutsideItsPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	err := cs.Draw([]int{0})
	require.Error(t, err, "ベット前には引けない")
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))

	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Draw(nil))
	err = cs.Draw([]int{0})
	require.Error(t, err, "アクションフェーズに入ったらもう引けない")
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCaribbeanDraw_DrawRefusedWhenTheFeeIsUnaffordable(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	before := append([]*domain.Card(nil), cs.GetPlayerHand()...)
	cs.SetChips(99) // 手数料 100 に 1 足りない

	err := cs.Draw([]int{0})
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
	assert.Equal(t, before, cs.GetPlayerHand(), "払えなかった交換で手札は動かない")
	assert.Equal(t, 99, cs.GetChips())
	assert.Equal(t, domain.CaribbeanDrawPhaseDraw, cs.GetPhase(),
		"払えなければドローフェーズに留まる（引かずに進む道が残る）")
}

func TestCaribbeanDraw_TheDrawnHandIsWhatGetsSettled(t *testing.T) {
	// **交換後の手で勝負する。** 配られた手のまま判定していると、引いて
	// 完成させた役が無視される。
	cs := domain.NewDefaultCaribbeanDraw()
	cs.SetChips(100000)
	require.NoError(t, cs.Bet(100, 10))
	require.NoError(t, cs.Draw(nil))

	// ドロー後の手として役を置き、精算がその手を読むことを見る。
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignSpade, 5},
		cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignSpade, 9},
		cdrCard{domain.CardDesignSpade, 11}))
	cs.SetDealerHand(makeCdrHand(
		cdrCard{domain.CardDesignDiamond, 9}, cdrCard{domain.CardDesignHeart, 9},
		cdrCard{domain.CardDesignClover, 6}, cdrCard{domain.CardDesignSpade, 3},
		cdrCard{domain.CardDesignDiamond, 10}))
	require.NoError(t, cs.Play())

	assert.Equal(t, domain.GameResultWin, cs.GetResult())
	assert.Equal(t, 10*domain.CaribbeanDrawJackpotFlush, cs.GetJackpotPayout(),
		"サイドベットも交換後の手で払う")
}

func TestCaribbeanDraw_ResetClearsTheDrawCost(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	require.NoError(t, cs.Draw([]int{0}))
	require.Equal(t, 100, cs.GetDrawCost())

	cs.Reset()
	assert.Equal(t, 0, cs.GetDrawCost())
	assert.Equal(t, domain.CaribbeanDrawPhaseBet, cs.GetPhase())
}

func TestCaribbeanDraw_ThePlayerRankIsLiveFromTheDeal(t *testing.T) {
	// **配った時点で自分の役は確定する。** ここが 0 のままだと、フラッシュを
	// 配られても画面には "High Card" (rank 0) と出る —— しかもこのゲームでは
	// それが「どれを捨てるか」を決める唯一の材料。
	//
	// 素の配りが偶然ハイカードだと「設定し忘れ」と区別が付かないので、役位が
	// 0 でない手が出るまで配り直す。
	cs := domain.NewDefaultCaribbeanDraw()
	rank := 0
	for range 200 {
		cs.Reset()
		require.NoError(t, cs.Bet(100, 0))
		if rank = cs.GetPlayerHandRank(); rank > domain.PokerHandHighCard {
			break
		}
	}
	require.Greater(t, rank, domain.PokerHandHighCard,
		"200 回配ってハイカード以外が一度も出なかった")
	assert.Equal(t, 0, cs.GetDealerHandRank(), "ディーラーの役は勝負するまで出さない")
}

func TestCaribbeanDraw_DrawingReevaluatesTheHand(t *testing.T) {
	// 引いて役が変わったら、役位もその場で追いつく。**引く札を仕込んで**
	// 結果を固定する —— 山任せだと「役位を持ち越していても偶然一致する」
	// 配りが混ざり、何も証明しない。
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))

	// ♠2 ♠5 ♠7 ♠9 ♥J — スペード4枚 + ハートのJ。今はハイカード。
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignSpade, 5},
		cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignSpade, 9},
		cdrCard{domain.CardDesignHeart, 11}))

	// 次に引かれる札を ♠K に固定 → 交換でフラッシュが完成する。
	//
	// **仕込めるのは山に残っている札だけ。** ♠K が最初の配りで誰かの手に
	// 行っていると StackTopForTest は 0 を返し、交換で引かれるのは山任せの
	// 別の札になる —— 52 枚中 10 枚が配られるので、およそ 5 回に 1 回。
	// 置けるまで配り直す。
	placed := 0
	for range 100 {
		if placed = cs.TrumpCardsForTest().StackTopForTest(
			domain.NewCard(domain.CardDesignSpade, 13, false)); placed == 1 {
			break
		}
		cs.Reset()
		require.NoError(t, cs.Bet(100, 0))
		cs.SetPlayerHand(makeCdrHand(
			cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignSpade, 5},
			cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignSpade, 9},
			cdrCard{domain.CardDesignHeart, 11}))
	}
	require.Equal(t, 1, placed, "100 回配り直しても ♠K が山に残らなかった")

	require.NoError(t, cs.Draw([]int{4}))

	assert.Equal(t, domain.CardDesignSpade, cs.GetPlayerHand()[4].GetDesign())
	assert.Equal(t, 13, cs.GetPlayerHand()[4].GetValue())
	assert.Equal(t, domain.PokerHandFlush, cs.GetPlayerHandRank(),
		"交換で完成したフラッシュが役位に反映されること")
}

func TestCaribbeanDraw_AnUnqualifiedDealerIsNeverAnnouncedAsALoss(t *testing.T) {
	// **儲かった局面を負けと呼ばない。** クオリファイしなければアンテが 1:1、
	// プレイは返却 —— 3×ante 賭けて 4×ante 戻るので、役がどれだけ悪くても
	// 取り分は必ず増える。素の比較で result を決めていた頃は、+100 儲けた手に
	// 「ディーラーの勝ちです」と赤字で出していた。
	cs := domain.NewDefaultCaribbeanDraw()
	cs.SetPhase(domain.CaribbeanDrawPhaseAction)
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignHeart, 5},
		cdrCard{domain.CardDesignClover, 7}, cdrCard{domain.CardDesignDiamond, 9},
		cdrCard{domain.CardDesignSpade, 11}))
	// 3 のペア: 役ではプレイヤーに勝つが、8 未満なのでクオリファイしない。
	cs.SetDealerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 3}, cdrCard{domain.CardDesignHeart, 3},
		cdrCard{domain.CardDesignClover, 8}, cdrCard{domain.CardDesignDiamond, 10},
		cdrCard{domain.CardDesignSpade, 12}))
	cs.SetAnteBet(100)
	cs.SetChips(200) // プレイベット 200 ちょうど
	require.NoError(t, cs.Play())

	require.False(t, cs.GetDealerQualified())
	require.Greater(t, cs.GetDealerHandRank(), cs.GetPlayerHandRank(), "役では負けている局面であること")
	assert.Equal(t, domain.GameResultWin, cs.GetResult(), "取り分が増えたのだから負けではない")
	assert.Equal(t, 400, cs.GetChips(), "300 賭けて 400 戻る = 差し引き +100")
}

func TestCaribbeanDraw_DrawSurvivesAnExhaustedDeck(t *testing.T) {
	// **山が尽きたら札は替えない。** DrawCard は nil を返すので、そのまま
	// 手札に入れると役の評価が nil を踏んで落ちる。通常のラウンドでは
	// 起きないが、KV から戻した卓の deckDrawCnt 次第では到達しうる。
	cs := domain.NewDefaultCaribbeanDraw()
	require.NoError(t, cs.Bet(100, 0))
	before := append([]*domain.Card(nil), cs.GetPlayerHand()...)

	// 山を空にする。
	deck := cs.TrumpCardsForTest()
	for deck.DrawCard() != nil { //nolint:revive // 意図的に引き切る
	}

	require.NotPanics(t, func() { _ = cs.Draw([]int{0, 1}) })
	for i, c := range cs.GetPlayerHand() {
		require.NotNil(t, c, "手札に nil が混ざらないこと (index %d)", i)
	}
	assert.Equal(t, before, cs.GetPlayerHand(), "引けなければ手札はそのまま")

	// 続けて勝負しても落ちない。
	require.NotPanics(t, func() { _ = cs.Play() })
}

func TestCaribbeanDraw_FoldingStillRecordsWhetherTheDealerQualified(t *testing.T) {
	// 降りても結果画面はディーラーの手を開いて資格を書く。計算しないと
	// dealerQualified が false のまま残り、K のペアに「クオリファイせず」と出る。
	cs := domain.NewDefaultCaribbeanDraw()
	cs.SetPhase(domain.CaribbeanDrawPhaseAction)
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignHeart, 5},
		cdrCard{domain.CardDesignClover, 7}, cdrCard{domain.CardDesignDiamond, 9},
		cdrCard{domain.CardDesignSpade, 11}))
	cs.SetDealerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 13}, cdrCard{domain.CardDesignHeart, 13},
		cdrCard{domain.CardDesignClover, 6}, cdrCard{domain.CardDesignDiamond, 3},
		cdrCard{domain.CardDesignSpade, 9}))
	cs.SetAnteBet(100)
	require.NoError(t, cs.Fold())
	assert.True(t, cs.GetDealerQualified(), "K のペアは 8 のペア以上")

	// 境界の反対側も見る。7 のペアは資格に届かない。
	cs = domain.NewDefaultCaribbeanDraw()
	cs.SetPhase(domain.CaribbeanDrawPhaseAction)
	cs.SetPlayerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 2}, cdrCard{domain.CardDesignHeart, 5},
		cdrCard{domain.CardDesignClover, 7}, cdrCard{domain.CardDesignDiamond, 9},
		cdrCard{domain.CardDesignSpade, 11}))
	cs.SetDealerHand(makeCdrHand(
		cdrCard{domain.CardDesignSpade, 7}, cdrCard{domain.CardDesignHeart, 7},
		cdrCard{domain.CardDesignClover, 6}, cdrCard{domain.CardDesignDiamond, 3},
		cdrCard{domain.CardDesignSpade, 9}))
	cs.SetAnteBet(100)
	require.NoError(t, cs.Fold())
	assert.False(t, cs.GetDealerQualified(), "7 のペアは 8 に届かない")
}

// TestCaribbeanDraw_PaytableValuesArePinnedToNumbers fixes the actual payout
// table to literals.
//
// **定数を両辺に使うと配当表は固定されない。** `expected := 200 + 200*tt.multiplier`
// は被検査コードと同じ定数を読むので、`CaribbeanDrawPayRoyalFlush` を 50 から 5 に
// 落としても両辺が一緒に動いて緑のまま通る —— 実測済み。倍率そのものを数字で
// 書いておかないと、配当表を静かに 10 倍間違えても誰も気付かない。
func TestCaribbeanDraw_PaytableValuesArePinnedToNumbers(t *testing.T) {
	t.Run("call payouts", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			got  int
			want int
		}{
			{"royal flush", domain.CaribbeanDrawPayRoyalFlush, 50},
			{"straight flush", domain.CaribbeanDrawPayStraightFlush, 20},
			{"four of a kind", domain.CaribbeanDrawPayFourOfAKind, 10},
			{"full house", domain.CaribbeanDrawPayFullHouse, 5},
			{"flush", domain.CaribbeanDrawPayFlush, 4},
			{"straight", domain.CaribbeanDrawPayStraight, 3},
			{"three of a kind", domain.CaribbeanDrawPayThreeOfAKind, 2},
			{"two pair", domain.CaribbeanDrawPayTwoPair, 1},
			{"pair or less", domain.CaribbeanDrawPayPair, 1},
		} {
			assert.Equal(t, tt.want, tt.got, "%s の倍率", tt.name)
		}
	})

	t.Run("jackpot payouts", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			got  int
			want int
		}{
			{"royal flush", domain.CaribbeanDrawJackpotRoyalFlush, 10000},
			{"straight flush", domain.CaribbeanDrawJackpotStraightFlush, 2000},
			{"four of a kind", domain.CaribbeanDrawJackpotFourOfAKind, 200},
			{"full house", domain.CaribbeanDrawJackpotFullHouse, 50},
			{"flush", domain.CaribbeanDrawJackpotFlush, 25},
		} {
			assert.Equal(t, tt.want, tt.got, "%s のジャックポット倍率", tt.name)
		}
	})

	t.Run("a better hand never pays less", func(t *testing.T) {
		// 表を書き換えるときに順序が崩れないよう、単調性そのものも押さえる。
		ladder := []int{
			domain.CaribbeanDrawPayPair, domain.CaribbeanDrawPayTwoPair,
			domain.CaribbeanDrawPayThreeOfAKind, domain.CaribbeanDrawPayStraight,
			domain.CaribbeanDrawPayFlush, domain.CaribbeanDrawPayFullHouse,
			domain.CaribbeanDrawPayFourOfAKind, domain.CaribbeanDrawPayStraightFlush,
			domain.CaribbeanDrawPayRoyalFlush,
		}
		for i := 1; i < len(ladder); i++ {
			assert.GreaterOrEqual(t, ladder[i], ladder[i-1], "役位 %d で配当が下がっている", i)
		}
	})

	t.Run("thinner than the clone source", func(t *testing.T) {
		// **クローン元より薄いこと自体が仕様。** Caribbean Stud の表をそのまま
		// 戻すと、引ける卓ではプレイヤー有利に振り切れる。
		assert.Less(t, domain.CaribbeanDrawPayRoyalFlush, 100,
			"Caribbean Stud の 100:1 をそのまま持ってきていないこと")
		assert.Less(t, domain.CaribbeanDrawJackpotRoyalFlush, 20000,
			"Caribbean Stud の 20000:1 をそのまま持ってきていないこと")
	})
}
