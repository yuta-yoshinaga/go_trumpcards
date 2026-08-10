package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func makeHandCH(specs ...cd) []*domain.Card {
	cards := make([]*domain.Card, len(specs))
	for i, s := range specs {
		cards[i] = domain.NewCard(s.d, s.v, false)
	}
	return cards
}

func TestNewDefaultCasinoHoldem(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	assert.Equal(t, domain.CasinoHoldemPhaseBet, g.GetPhase())
	assert.Equal(t, domain.CasinoHoldemDefaultChips, g.GetChips())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetPlayerHand())
	assert.Nil(t, g.GetDealerHand())
	assert.Nil(t, g.GetCommunity())
}

func TestCasinoHoldem_Reset(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Fold())
	assert.Equal(t, domain.CasinoHoldemPhaseEnd, g.GetPhase())

	g.Reset()
	assert.Equal(t, domain.CasinoHoldemPhaseBet, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetPlayerHand())
	assert.Nil(t, g.GetDealerHand())
	assert.Nil(t, g.GetCommunity())
	assert.Equal(t, 0, g.GetAnteBet())
	assert.Equal(t, 0, g.GetBonusBet())
	assert.Equal(t, 0, g.GetCallBet())
	assert.False(t, g.GetDealerQualify())
}

func TestCasinoHoldem_Reset_RefillChips(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetChips(5)
	g.Reset()
	assert.Equal(t, domain.CasinoHoldemDefaultChips, g.GetChips())
}

func TestCasinoHoldem_Bet_Success(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	require.NoError(t, g.Bet(100, 50))
	assert.Equal(t, domain.CasinoHoldemPhaseFlop, g.GetPhase())
	assert.Equal(t, 100, g.GetAnteBet())
	assert.Equal(t, 50, g.GetBonusBet())
	assert.Len(t, g.GetPlayerHand(), 2)
	assert.Len(t, g.GetDealerHand(), 2)
	assert.Len(t, g.GetCommunity(), 3)
	assert.Equal(t, domain.CasinoHoldemDefaultChips-150, g.GetChips())
}

func TestCasinoHoldem_Bet_WrongPhase(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	err := g.Bet(100, 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCasinoHoldem_Bet_InvalidAnte(t *testing.T) {
	cases := []struct {
		name      string
		ante      int
		bonus     int
		setupChip int
	}{
		{name: "zero ante", ante: 0, bonus: 0, setupChip: 1000},
		{name: "below min", ante: 5, bonus: 0, setupChip: 1000},
		{name: "non multiple", ante: 105, bonus: 0, setupChip: 1000},
		{name: "above max", ante: domain.CasinoHoldemMaxBet + 10, bonus: 0, setupChip: 100000},
		{name: "negative bonus", ante: 100, bonus: -10, setupChip: 1000},
		{name: "bonus below min", ante: 100, bonus: 5, setupChip: 1000},
		{name: "bonus above max", ante: 100, bonus: domain.CasinoHoldemMaxBet + 10, setupChip: 100000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := domain.NewDefaultCasinoHoldem()
			g.SetChips(tc.setupChip)
			err := g.Bet(tc.ante, tc.bonus)
			require.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestCasinoHoldem_Bet_InsufficientChips(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetChips(50)
	err := g.Bet(100, 0)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestCasinoHoldem_Call_Success(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Call())
	assert.Equal(t, domain.CasinoHoldemPhaseEnd, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 200, g.GetCallBet())
	assert.Len(t, g.GetCommunity(), 5)
}

func TestCasinoHoldem_Call_WrongPhase(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	err := g.Call()
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestCasinoHoldem_Call_InsufficientChips(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	require.NoError(t, g.Bet(100, 0))
	g.SetChips(50)
	err := g.Call()
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestCasinoHoldem_Fold_Success(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Fold())
	assert.Equal(t, domain.CasinoHoldemPhaseEnd, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, g.GetResult())
}

func TestCasinoHoldem_Fold_WrongPhase(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	err := g.Fold()
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

// 配当ロジック：プレイヤーが Royal Flush で勝利、ディーラークオリファイ
func TestCasinoHoldem_Resolve_PlayerWinsWithRoyalFlush(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetChips(1000)
	g.SetAnteBet(100)
	g.SetCallBet(200)
	// プレイヤー: A♠ K♠
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignSpade, 1},
		cd{domain.CardDesignSpade, 13},
	))
	// ディーラー: A♥ A♦ (Pair of Aces — qualifies)
	g.SetDealerHand(makeHandCH(
		cd{domain.CardDesignHeart, 1},
		cd{domain.CardDesignDiamond, 1},
	))
	// コミュニティ: Q♠ J♠ 10♠ 2♣ 3♥ — プレイヤー Royal Flush, ディーラー A trips
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignSpade, 12},
		cd{domain.CardDesignSpade, 11},
		cd{domain.CardDesignSpade, 10},
		cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignHeart, 3},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	g.ForceResolve()

	assert.Equal(t, domain.GameResultWin, g.GetResult())
	assert.True(t, g.GetDealerQualify())
	// アンテ: 100 + 100*100 = 10100
	assert.Equal(t, 100+100*domain.CasinoHoldemAntePayRoyalFlush, g.GetAntePayout())
	// コール: 200*2 = 400
	assert.Equal(t, 400, g.GetCallPayout())
}

// 配当ロジック：ディーラーがクオリファイしない（High Card のみ）
// アンテはハンドランクで支払う、コールはプッシュ
func TestCasinoHoldem_Resolve_DealerDoesNotQualify(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetChips(1000)
	g.SetAnteBet(100)
	g.SetCallBet(200)
	// プレイヤー: A♠ K♠
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignSpade, 1},
		cd{domain.CardDesignSpade, 13},
	))
	// ディーラー: 2♥ 7♦
	g.SetDealerHand(makeHandCH(
		cd{domain.CardDesignHeart, 2},
		cd{domain.CardDesignDiamond, 7},
	))
	// コミュニティ: Q♠ J♠ 10♠ 4♣ 5♥ — プレイヤー Royal Flush, ディーラー High Card
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignSpade, 12},
		cd{domain.CardDesignSpade, 11},
		cd{domain.CardDesignSpade, 10},
		cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignHeart, 5},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	g.ForceResolve()

	assert.False(t, g.GetDealerQualify())
	// プレイヤー勝利
	assert.Equal(t, domain.GameResultWin, g.GetResult())
	// アンテはハンドランクに応じて支払う（Royal Flush）
	assert.Equal(t, 100+100*domain.CasinoHoldemAntePayRoyalFlush, g.GetAntePayout())
	// コールはプッシュ（元金返却のみ）
	assert.Equal(t, 200, g.GetCallPayout())
}

// ディーラーが Pair of Twos（クオリファイしない最低限）でも、
// プレイヤーのハンドランクに応じてアンテは支払われる
func TestCasinoHoldem_Resolve_DealerHasLowPairNoQualify(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetChips(1000)
	g.SetAnteBet(100)
	g.SetCallBet(200)
	// プレイヤー: K♠ K♥（ペア）
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignSpade, 13},
		cd{domain.CardDesignHeart, 13},
	))
	// ディーラー: 2♣ 7♦
	g.SetDealerHand(makeHandCH(
		cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignDiamond, 7},
	))
	// コミュニティ: 2♥ 5♣ 8♦ 9♠ J♣ — ディーラー Pair of Twos（< 4 でクオリファイしない）
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignHeart, 2},
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignClover, 11},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	g.ForceResolve()

	assert.False(t, g.GetDealerQualify())
	// アンテ 1:1（Other）, コールはプッシュ
	assert.Equal(t, 100+100*domain.CasinoHoldemAntePayOther, g.GetAntePayout())
	assert.Equal(t, 200, g.GetCallPayout())
}

// ディーラーが Pair of Fours でクオリファイ
func TestCasinoHoldem_Resolve_DealerQualifiesOnPairOfFours(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetChips(1000)
	g.SetAnteBet(100)
	g.SetCallBet(200)
	// プレイヤー: A♠ K♠
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignSpade, 1},
		cd{domain.CardDesignSpade, 13},
	))
	// ディーラー: 4♣ 7♦
	g.SetDealerHand(makeHandCH(
		cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignDiamond, 7},
	))
	// コミュニティ: 4♥ Q♠ J♠ 10♠ 9♣ — プレイヤー Royal-1（A high straight flush 不成立、高カード／ストレート要件不足。
	// ここではディーラー Pair of Fours クオリファイの検証だけが目的なので、結果は気にしない。
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignHeart, 4},
		cd{domain.CardDesignSpade, 12},
		cd{domain.CardDesignSpade, 11},
		cd{domain.CardDesignSpade, 10},
		cd{domain.CardDesignClover, 9},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	g.ForceResolve()

	assert.True(t, g.GetDealerQualify())
}

// プレイヤー敗北：ディーラークオリファイ＋ディーラー勝利
func TestCasinoHoldem_Resolve_PlayerLoses(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetChips(1000)
	g.SetAnteBet(100)
	g.SetCallBet(200)
	// プレイヤー: 2♣ 3♦
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignDiamond, 3},
	))
	// ディーラー: A♥ A♦（Pair of Aces qualify）
	g.SetDealerHand(makeHandCH(
		cd{domain.CardDesignHeart, 1},
		cd{domain.CardDesignDiamond, 1},
	))
	// コミュニティ: 5♣ 8♦ 10♠ J♣ Q♥
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 10},
		cd{domain.CardDesignClover, 11},
		cd{domain.CardDesignHeart, 12},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	g.ForceResolve()

	assert.True(t, g.GetDealerQualify())
	assert.Equal(t, domain.GameResultLose, g.GetResult())
	assert.Equal(t, 0, g.GetAntePayout())
	assert.Equal(t, 0, g.GetCallPayout())
}

// 引き分け：同じ最良5枚
func TestCasinoHoldem_Resolve_Draw(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetChips(1000)
	g.SetAnteBet(100)
	g.SetCallBet(200)
	// プレイヤーとディーラーがどちらも同じハンドになるよう、コミュニティ 5 枚で
	// ストレート 10-J-Q-K-A を完成させる。両者のホールカードは無関係（最良5枚はコミュニティのまま）。
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignDiamond, 3},
	))
	g.SetDealerHand(makeHandCH(
		cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignDiamond, 5},
	))
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignSpade, 10},
		cd{domain.CardDesignHeart, 11},
		cd{domain.CardDesignDiamond, 12},
		cd{domain.CardDesignClover, 13},
		cd{domain.CardDesignSpade, 1},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	g.ForceResolve()

	assert.True(t, g.GetDealerQualify())
	assert.Equal(t, domain.GameResultDraw, g.GetResult())
	// プッシュ：アンテ・コールとも元金返却
	assert.Equal(t, 100, g.GetAntePayout())
	assert.Equal(t, 200, g.GetCallPayout())
}

// AA ボーナス：プレイヤーの 2 枚 + フロップ 3 枚で Pair of Aces 以上
func TestCasinoHoldem_Bonus_PairOfAces(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetBonusBet(100)
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignSpade, 1},
		cd{domain.CardDesignHeart, 1},
	))
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 10},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	require.NoError(t, g.Fold())
	assert.Equal(t, 100+100*domain.CasinoHoldemBonusPayPairOfAces, g.GetBonusPayout())
}

// AA ボーナス：弱いペア（K のペア）はペイアウトされない
func TestCasinoHoldem_Bonus_WeakPairLoses(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetBonusBet(100)
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignSpade, 13},
		cd{domain.CardDesignHeart, 13},
	))
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 10},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	require.NoError(t, g.Fold())
	assert.Equal(t, 0, g.GetBonusPayout())
}

// AA ボーナス：Royal Flush（最高配当）
func TestCasinoHoldem_Bonus_RoyalFlush(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetBonusBet(100)
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignSpade, 1},
		cd{domain.CardDesignSpade, 13},
	))
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignSpade, 12},
		cd{domain.CardDesignSpade, 11},
		cd{domain.CardDesignSpade, 10},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	require.NoError(t, g.Fold())
	assert.Equal(t, 100+100*domain.CasinoHoldemBonusPayRoyalFlush, g.GetBonusPayout())
}

// AA ボーナス：StraightFlush, FourOfAKind, FullHouse, Flush, Straight, ThreeOfAKind, TwoPair
func TestCasinoHoldem_Bonus_AllRanks(t *testing.T) {
	cases := []struct {
		name      string
		player    []cd
		community []cd
		wantMult  int
	}{
		{
			name:      "straight flush",
			player:    []cd{{domain.CardDesignSpade, 9}, {domain.CardDesignSpade, 8}},
			community: []cd{{domain.CardDesignSpade, 7}, {domain.CardDesignSpade, 6}, {domain.CardDesignSpade, 5}},
			wantMult:  domain.CasinoHoldemBonusPayStraightFlush,
		},
		{
			name:      "four of a kind",
			player:    []cd{{domain.CardDesignSpade, 7}, {domain.CardDesignHeart, 7}},
			community: []cd{{domain.CardDesignClover, 7}, {domain.CardDesignDiamond, 7}, {domain.CardDesignSpade, 2}},
			wantMult:  domain.CasinoHoldemBonusPayFourOfAKind,
		},
		{
			name:      "full house",
			player:    []cd{{domain.CardDesignSpade, 7}, {domain.CardDesignHeart, 7}},
			community: []cd{{domain.CardDesignClover, 7}, {domain.CardDesignDiamond, 2}, {domain.CardDesignSpade, 2}},
			wantMult:  domain.CasinoHoldemBonusPayFullHouse,
		},
		{
			name:      "flush",
			player:    []cd{{domain.CardDesignSpade, 2}, {domain.CardDesignSpade, 4}},
			community: []cd{{domain.CardDesignSpade, 6}, {domain.CardDesignSpade, 8}, {domain.CardDesignSpade, 11}},
			wantMult:  domain.CasinoHoldemBonusPayFlush,
		},
		{
			name:      "straight",
			player:    []cd{{domain.CardDesignSpade, 9}, {domain.CardDesignHeart, 8}},
			community: []cd{{domain.CardDesignClover, 7}, {domain.CardDesignDiamond, 6}, {domain.CardDesignSpade, 5}},
			wantMult:  domain.CasinoHoldemBonusPayStraight,
		},
		{
			name:      "three of a kind",
			player:    []cd{{domain.CardDesignSpade, 7}, {domain.CardDesignHeart, 7}},
			community: []cd{{domain.CardDesignClover, 7}, {domain.CardDesignDiamond, 5}, {domain.CardDesignSpade, 2}},
			wantMult:  domain.CasinoHoldemBonusPayThreeOfAKind,
		},
		{
			name:      "two pair",
			player:    []cd{{domain.CardDesignSpade, 7}, {domain.CardDesignHeart, 7}},
			community: []cd{{domain.CardDesignClover, 5}, {domain.CardDesignDiamond, 5}, {domain.CardDesignSpade, 2}},
			wantMult:  domain.CasinoHoldemBonusPayTwoPair,
		},
		{
			name:      "high card no payout",
			player:    []cd{{domain.CardDesignSpade, 2}, {domain.CardDesignHeart, 4}},
			community: []cd{{domain.CardDesignClover, 6}, {domain.CardDesignDiamond, 9}, {domain.CardDesignSpade, 11}},
			wantMult:  0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := domain.NewDefaultCasinoHoldem()
			g.SetBonusBet(100)
			g.SetPlayerHand(makeHandCH(tc.player...))
			g.SetCommunity(makeHandCH(tc.community...))
			g.SetPhase(domain.CasinoHoldemPhaseFlop)
			require.NoError(t, g.Fold())
			if tc.wantMult == 0 {
				assert.Equal(t, 0, g.GetBonusPayout())
			} else {
				assert.Equal(t, 100+100*tc.wantMult, g.GetBonusPayout())
			}
		})
	}
}

// AA ボーナス：bonusBet=0 のときは何もしない
func TestCasinoHoldem_Bonus_NoBonusBet(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	g.SetBonusBet(0)
	g.SetPlayerHand(makeHandCH(
		cd{domain.CardDesignSpade, 1},
		cd{domain.CardDesignHeart, 1},
	))
	g.SetCommunity(makeHandCH(
		cd{domain.CardDesignClover, 5},
		cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignSpade, 10},
	))
	g.SetPhase(domain.CasinoHoldemPhaseFlop)
	require.NoError(t, g.Fold())
	assert.Equal(t, 0, g.GetBonusPayout())
}

// アンテ配当倍率の各役カバレッジ
func TestCasinoHoldem_AnteMultiplier_AllRanks(t *testing.T) {
	cases := []struct {
		name     string
		player   []cd
		dealer   []cd
		commBoth []cd // 両者で共有するコミュニティ 5 枚
		wantMult int
	}{
		{
			name:     "royal flush",
			player:   []cd{{domain.CardDesignSpade, 1}, {domain.CardDesignSpade, 13}},
			dealer:   []cd{{domain.CardDesignClover, 2}, {domain.CardDesignDiamond, 3}},
			commBoth: []cd{{domain.CardDesignSpade, 12}, {domain.CardDesignSpade, 11}, {domain.CardDesignSpade, 10}, {domain.CardDesignHeart, 4}, {domain.CardDesignClover, 5}},
			wantMult: domain.CasinoHoldemAntePayRoyalFlush,
		},
		{
			name:     "straight flush",
			player:   []cd{{domain.CardDesignSpade, 9}, {domain.CardDesignSpade, 8}},
			dealer:   []cd{{domain.CardDesignClover, 2}, {domain.CardDesignDiamond, 3}},
			commBoth: []cd{{domain.CardDesignSpade, 7}, {domain.CardDesignSpade, 6}, {domain.CardDesignSpade, 5}, {domain.CardDesignHeart, 4}, {domain.CardDesignClover, 13}},
			wantMult: domain.CasinoHoldemAntePayStraightFlush,
		},
		{
			name:     "four of a kind",
			player:   []cd{{domain.CardDesignSpade, 7}, {domain.CardDesignHeart, 7}},
			dealer:   []cd{{domain.CardDesignClover, 2}, {domain.CardDesignDiamond, 3}},
			commBoth: []cd{{domain.CardDesignClover, 7}, {domain.CardDesignDiamond, 7}, {domain.CardDesignSpade, 5}, {domain.CardDesignHeart, 9}, {domain.CardDesignClover, 11}},
			wantMult: domain.CasinoHoldemAntePayFourOfAKind,
		},
		{
			name:     "full house",
			player:   []cd{{domain.CardDesignSpade, 7}, {domain.CardDesignHeart, 7}},
			dealer:   []cd{{domain.CardDesignClover, 4}, {domain.CardDesignDiamond, 6}},
			commBoth: []cd{{domain.CardDesignClover, 7}, {domain.CardDesignDiamond, 2}, {domain.CardDesignSpade, 2}, {domain.CardDesignHeart, 9}, {domain.CardDesignClover, 11}},
			wantMult: domain.CasinoHoldemAntePayFullHouse,
		},
		{
			name:     "flush",
			player:   []cd{{domain.CardDesignSpade, 2}, {domain.CardDesignSpade, 4}},
			dealer:   []cd{{domain.CardDesignClover, 9}, {domain.CardDesignDiamond, 11}},
			commBoth: []cd{{domain.CardDesignSpade, 6}, {domain.CardDesignSpade, 8}, {domain.CardDesignSpade, 13}, {domain.CardDesignHeart, 3}, {domain.CardDesignClover, 5}},
			wantMult: domain.CasinoHoldemAntePayFlush,
		},
		{
			name:     "straight (other 1:1)",
			player:   []cd{{domain.CardDesignSpade, 9}, {domain.CardDesignHeart, 8}},
			dealer:   []cd{{domain.CardDesignClover, 2}, {domain.CardDesignDiamond, 3}},
			commBoth: []cd{{domain.CardDesignClover, 7}, {domain.CardDesignDiamond, 6}, {domain.CardDesignSpade, 5}, {domain.CardDesignHeart, 13}, {domain.CardDesignClover, 11}},
			wantMult: domain.CasinoHoldemAntePayOther,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := domain.NewDefaultCasinoHoldem()
			g.SetChips(1000)
			g.SetAnteBet(100)
			g.SetCallBet(200)
			g.SetPlayerHand(makeHandCH(tc.player...))
			g.SetDealerHand(makeHandCH(tc.dealer...))
			g.SetCommunity(makeHandCH(tc.commBoth...))
			g.SetPhase(domain.CasinoHoldemPhaseFlop)
			g.ForceResolve()
			expected := 100 + 100*tc.wantMult
			assert.Equal(t, expected, g.GetAntePayout(), "wrong ante payout for %s", tc.name)
		})
	}
}

// JSON 往復
func TestCasinoHoldem_JSON_RoundTrip(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	require.NoError(t, g.Bet(100, 50))

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.CasinoHoldem
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetAnteBet(), restored.GetAnteBet())
	assert.Equal(t, g.GetBonusBet(), restored.GetBonusBet())
	assert.Equal(t, g.GetChips(), restored.GetChips())
	assert.Len(t, restored.GetPlayerHand(), len(g.GetPlayerHand()))
}

// JSON サイズ上限
func TestCasinoHoldem_JSON_TooLarge(t *testing.T) {
	// PlayerHand を巨大にして上限を超過させる
	cards := make([]any, 1001)
	for i := range cards {
		cards[i] = map[string]any{"des": 1, "val": 2, "rev": false}
	}
	payload, err := json.Marshal(map[string]any{"ph": cards})
	require.NoError(t, err)
	var restored domain.CasinoHoldem
	err = json.Unmarshal(payload, &restored)
	require.Error(t, err)
}

// JSON 不正フォーマット
func TestCasinoHoldem_JSON_InvalidFormat(t *testing.T) {
	var restored domain.CasinoHoldem
	err := restored.UnmarshalJSON([]byte("not json"))
	require.Error(t, err)
}

// PlayerHandRank が フロップ後（中盤）に更新されることを検証
func TestCasinoHoldem_PlayerHandRank_PopulatedAtFlop(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	require.NoError(t, g.Bet(100, 0))
	// フロップ後にホール+フロップ 5 枚の評価が反映されている（>= 0）
	assert.Equal(t, domain.CasinoHoldemPhaseFlop, g.GetPhase())
	// 5 枚あれば必ず High Card 以上が評価される
	assert.GreaterOrEqual(t, g.GetPlayerHandRank(), domain.PokerHandHighCard)
}

// アクションログ取得
func TestCasinoHoldem_ActionLog(t *testing.T) {
	g := domain.NewDefaultCasinoHoldem()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Fold())
	log := g.GetActionLog()
	require.NotEmpty(t, log)
	// bet → deal → flop → fold → result
	assert.GreaterOrEqual(t, len(log), 5)
}

func TestCasinoHoldem_RecommendCall(t *testing.T) {
	newFlop := func(hole, community []*domain.Card) *domain.CasinoHoldem {
		g := domain.NewDefaultCasinoHoldem()
		g.SetPlayerHand(hole)
		g.SetCommunity(community)
		g.SetPhase(domain.CasinoHoldemPhaseFlop)
		return g
	}

	t.Run("call with a pair", func(t *testing.T) {
		g := newFlop(
			makeHandCH(cd{domain.CardDesignSpade, 8}, cd{domain.CardDesignHeart, 8}),
			makeHandCH(cd{domain.CardDesignClover, 3}, cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignSpade, 9}),
		)
		assert.True(t, g.RecommendCall())
	})

	t.Run("call with an ace high", func(t *testing.T) {
		g := newFlop(
			makeHandCH(cd{domain.CardDesignSpade, 1}, cd{domain.CardDesignHeart, 7}),
			makeHandCH(cd{domain.CardDesignClover, 3}, cd{domain.CardDesignDiamond, 5}, cd{domain.CardDesignSpade, 9}),
		)
		assert.True(t, g.RecommendCall())
	})

	// **ボードの A / K も数える。**hole だけ見る実装に狭めると、
	// フロント側の casinoholdemHint.ts と食い違う (#4712)。
	t.Run("call when the king is on the board rather than in the hole", func(t *testing.T) {
		g := newFlop(
			makeHandCH(cd{domain.CardDesignSpade, 3}, cd{domain.CardDesignHeart, 7}),
			makeHandCH(cd{domain.CardDesignClover, 13}, cd{domain.CardDesignDiamond, 11}, cd{domain.CardDesignSpade, 5}),
		)
		assert.True(t, g.RecommendCall())
	})

	t.Run("fold with junk", func(t *testing.T) {
		g := newFlop(
			makeHandCH(cd{domain.CardDesignSpade, 3}, cd{domain.CardDesignHeart, 7}),
			makeHandCH(cd{domain.CardDesignClover, 9}, cd{domain.CardDesignDiamond, 11}, cd{domain.CardDesignSpade, 5}),
		)
		assert.False(t, g.RecommendCall())
	})
}
