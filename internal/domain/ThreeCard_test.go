package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewDefaultThreeCard(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	assert.Equal(t, domain.ThreeCardPhaseBet, tc.GetPhase())
	assert.Equal(t, domain.ThreeCardDefaultChips, tc.GetChips())
	assert.False(t, tc.GetGameEndFlag())
	assert.Nil(t, tc.GetPlayerHand())
	assert.Nil(t, tc.GetDealerHand())
}

func TestThreeCard_Reset(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	// Play a round
	err := tc.Bet(100, 0)
	require.NoError(t, err)
	err = tc.Play()
	require.NoError(t, err)
	assert.Equal(t, domain.ThreeCardPhaseEnd, tc.GetPhase())

	// Reset
	tc.Reset()
	assert.Equal(t, domain.ThreeCardPhaseBet, tc.GetPhase())
	assert.False(t, tc.GetGameEndFlag())
	assert.Nil(t, tc.GetPlayerHand())
	assert.Nil(t, tc.GetDealerHand())
	assert.Equal(t, 0, tc.GetAnteBet())
	assert.Equal(t, 0, tc.GetPairPlusBet())
	assert.Equal(t, 0, tc.GetPlayBet())
}

func TestThreeCard_Reset_RefillChips(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	tc.SetChips(5) // Below minimum bet
	tc.Reset()
	assert.Equal(t, domain.ThreeCardDefaultChips, tc.GetChips())
}

func TestThreeCard_Bet_WrongPhase(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	tc.SetPhase(domain.ThreeCardPhaseAction)
	err := tc.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestThreeCard_Bet_InvalidAnteAmount(t *testing.T) {
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
			tc := domain.NewDefaultThreeCard()
			err := tc.Bet(tt.ante, 0)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestThreeCard_Bet_InvalidPairPlusAmount(t *testing.T) {
	tests := []struct {
		name     string
		pairPlus int
	}{
		{"Negative", -10},
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := domain.NewDefaultThreeCard()
			err := tc.Bet(100, tt.pairPlus)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestThreeCard_Bet_InsufficientChips(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	tc.SetChips(50)
	err := tc.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestThreeCard_Bet_Success(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 50)
	assert.NoError(t, err)
	assert.Equal(t, domain.ThreeCardPhaseAction, tc.GetPhase())
	assert.Equal(t, 100, tc.GetAnteBet())
	assert.Equal(t, 50, tc.GetPairPlusBet())
	assert.Len(t, tc.GetPlayerHand(), 3)
	assert.Len(t, tc.GetDealerHand(), 3)
	assert.Equal(t, domain.ThreeCardDefaultChips-150, tc.GetChips())
}

func TestThreeCard_Bet_AnteOnlyNoPairPlus(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, tc.GetPairPlusBet())
	assert.Equal(t, domain.ThreeCardDefaultChips-100, tc.GetChips())
}

func TestThreeCard_Play_WrongPhase(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestThreeCard_Play_InsufficientChips(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0)
	require.NoError(t, err)
	tc.SetChips(0) // Force insufficient for play bet
	err = tc.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestThreeCard_Play_Success(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0)
	require.NoError(t, err)
	err = tc.Play()
	assert.NoError(t, err)
	assert.Equal(t, domain.ThreeCardPhaseEnd, tc.GetPhase())
	assert.True(t, tc.GetGameEndFlag())
	assert.Equal(t, 100, tc.GetPlayBet())
}

func TestThreeCard_Fold_WrongPhase(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Fold()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestThreeCard_Fold_Success(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0)
	require.NoError(t, err)
	chipsBefore := tc.GetChips()
	err = tc.Fold()
	assert.NoError(t, err)
	assert.Equal(t, domain.ThreeCardPhaseEnd, tc.GetPhase())
	assert.True(t, tc.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, tc.GetResult())
	// Ante is already deducted; no further deduction on fold
	assert.Equal(t, chipsBefore, tc.GetChips())
}

func TestThreeCard_Fold_PairPlusStillEvaluated(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 50)
	require.NoError(t, err)

	// Set player hand to a pair (Pair Plus pays 1:1)
	tc.SetPlayerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	})

	err = tc.Fold()
	assert.NoError(t, err)
	assert.Equal(t, domain.GameResultLose, tc.GetResult())
	// Pair Plus should pay: 50 + 50*1 = 100
	assert.Equal(t, 100, tc.GetPairPlusPayout())
}

func TestThreeCard_DealerQualification(t *testing.T) {
	tests := []struct {
		name      string
		dealer    []*domain.Card
		qualified bool
	}{
		{
			name: "QualifiesWithQueen",
			dealer: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 12, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
			},
			qualified: true,
		},
		{
			name: "QualifiesWithKing",
			dealer: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 13, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
			},
			qualified: true,
		},
		{
			name: "QualifiesWithAce",
			dealer: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
			},
			qualified: true,
		},
		{
			name: "QualifiesWithPairOfTwos",
			dealer: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 2, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
			},
			qualified: true,
		},
		{
			name: "NotQualified_JackHigh",
			dealer: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 11, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
			},
			qualified: false,
		},
		{
			name: "NotQualified_TenHigh",
			dealer: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 10, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
			},
			qualified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := domain.NewDefaultThreeCard()
			err := tc.Bet(100, 0)
			require.NoError(t, err)

			// Set player hand to a high card (weak)
			tc.SetPlayerHand([]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
				domain.NewCard(domain.CardDesignHeart, 2, false),
			})
			tc.SetDealerHand(tt.dealer)

			err = tc.Play()
			require.NoError(t, err)
			assert.Equal(t, tt.qualified, tc.GetDealerQualified())
		})
	}
}

func TestThreeCard_Payouts_DealerNotQualified(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0)
	require.NoError(t, err)

	// Player: Ace high, Dealer: Jack high (doesn't qualify)
	tc.SetPlayerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
	})
	tc.SetDealerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 11, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
	})

	err = tc.Play()
	require.NoError(t, err)
	assert.False(t, tc.GetDealerQualified())
	assert.Equal(t, 200, tc.GetAntePayout()) // ante 1:1 (100 + 100)
	assert.Equal(t, 100, tc.GetPlayPayout()) // play push (returned)
}

func TestThreeCard_Payouts_PlayerWins(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0)
	require.NoError(t, err)

	// Player: Ace high, Dealer: Queen-low (qualifies, but loses)
	tc.SetPlayerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	})
	tc.SetDealerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	})

	err = tc.Play()
	require.NoError(t, err)
	assert.True(t, tc.GetDealerQualified())
	assert.Equal(t, domain.GameResultWin, tc.GetResult())
	assert.Equal(t, 200, tc.GetAntePayout()) // 100 + 100
	assert.Equal(t, 200, tc.GetPlayPayout()) // 100 + 100
}

func TestThreeCard_Payouts_DealerWins(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0)
	require.NoError(t, err)

	// Player: low, Dealer: Ace high (qualifies and wins)
	tc.SetPlayerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	})
	tc.SetDealerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 1, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	err = tc.Play()
	require.NoError(t, err)
	assert.True(t, tc.GetDealerQualified())
	assert.Equal(t, domain.GameResultLose, tc.GetResult())
	assert.Equal(t, 0, tc.GetAntePayout())
	assert.Equal(t, 0, tc.GetPlayPayout())
}

func TestThreeCard_Payouts_Push(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0)
	require.NoError(t, err)

	// Same strength hands (tie)
	tc.SetPlayerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	})
	tc.SetDealerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 1, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
	})

	err = tc.Play()
	require.NoError(t, err)
	assert.Equal(t, domain.GameResultDraw, tc.GetResult())
	assert.Equal(t, 100, tc.GetAntePayout()) // push
	assert.Equal(t, 100, tc.GetPlayPayout()) // push
}

func TestThreeCard_AnteBonus(t *testing.T) {
	tests := []struct {
		name     string
		player   []*domain.Card
		expected int
	}{
		{
			name: "StraightFlush_5to1",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 6, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
			},
			expected: 500, // 100*5 = 500
		},
		{
			name: "ThreeOfAKind_4to1",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 8, false),
				domain.NewCard(domain.CardDesignClover, 8, false),
				domain.NewCard(domain.CardDesignHeart, 8, false),
			},
			expected: 400, // 100*4 = 400
		},
		{
			name: "Straight_1to1",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 4, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 6, false),
			},
			expected: 100, // 100*1 = 100
		},
		{
			name: "Flush_NoAnteBonus",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 2, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignSpade, 11, false),
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := domain.NewDefaultThreeCard()
			err := tc.Bet(100, 0)
			require.NoError(t, err)

			tc.SetPlayerHand(tt.player)
			// Dealer qualifies with Queen high
			tc.SetDealerHand([]*domain.Card{
				domain.NewCard(domain.CardDesignDiamond, 12, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			})

			err = tc.Play()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tc.GetAnteBonusPayout())
		})
	}
}

func TestThreeCard_PairPlus(t *testing.T) {
	tests := []struct {
		name     string
		player   []*domain.Card
		expected int
	}{
		{
			name: "StraightFlush_40to1",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignSpade, 6, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
			},
			expected: 2050, // 50 + 50*40 = 2050
		},
		{
			name: "ThreeOfAKind_30to1",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 8, false),
				domain.NewCard(domain.CardDesignClover, 8, false),
				domain.NewCard(domain.CardDesignHeart, 8, false),
			},
			expected: 1550, // 50 + 50*30 = 1550
		},
		{
			name: "Straight_6to1",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 4, false),
				domain.NewCard(domain.CardDesignClover, 5, false),
				domain.NewCard(domain.CardDesignHeart, 6, false),
			},
			expected: 350, // 50 + 50*6 = 350
		},
		{
			name: "Flush_3to1",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 2, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignSpade, 11, false),
			},
			expected: 200, // 50 + 50*3 = 200
		},
		{
			name: "Pair_1to1",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 10, false),
				domain.NewCard(domain.CardDesignClover, 10, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
			},
			expected: 100, // 50 + 50*1 = 100
		},
		{
			name: "HighCard_NoPayment",
			player: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 2, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
				domain.NewCard(domain.CardDesignHeart, 11, false),
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := domain.NewDefaultThreeCard()
			err := tc.Bet(100, 50)
			require.NoError(t, err)

			tc.SetPlayerHand(tt.player)
			// Dealer qualifies
			tc.SetDealerHand([]*domain.Card{
				domain.NewCard(domain.CardDesignDiamond, 12, false),
				domain.NewCard(domain.CardDesignHeart, 3, false),
				domain.NewCard(domain.CardDesignClover, 2, false),
			})

			err = tc.Play()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tc.GetPairPlusPayout())
		})
	}
}

func TestThreeCard_PairPlus_NoBet(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 0) // No pair plus
	require.NoError(t, err)

	tc.SetPlayerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
	})
	tc.SetDealerHand([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	})

	err = tc.Play()
	require.NoError(t, err)
	assert.Equal(t, 0, tc.GetPairPlusPayout())
}

func TestThreeCard_GetTotalPayout(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	tc.SetAntePayout(200)
	tc.SetPlayPayout(200)
	tc.SetAnteBonusPayout(100)
	tc.SetPairPlusPayout(50)
	assert.Equal(t, 550, tc.GetTotalPayout())
}

func TestThreeCard_ActionLog(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 50)
	require.NoError(t, err)
	err = tc.Play()
	require.NoError(t, err)
	log := tc.GetActionLog()
	assert.True(t, len(log) >= 3) // bet, deal, play, result (at least)
}

func TestThreeCard_JSON_RoundTrip(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	err := tc.Bet(100, 50)
	require.NoError(t, err)

	data, err := json.Marshal(tc)
	require.NoError(t, err)

	var restored domain.ThreeCard
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, tc.GetPhase(), restored.GetPhase())
	assert.Equal(t, tc.GetAnteBet(), restored.GetAnteBet())
	assert.Equal(t, tc.GetPairPlusBet(), restored.GetPairPlusBet())
	assert.Equal(t, tc.GetChips(), restored.GetChips())
	assert.Len(t, restored.GetPlayerHand(), 3)
	assert.Len(t, restored.GetDealerHand(), 3)
}

func TestThreeCard_JSON_Unmarshal_NilFields(t *testing.T) {
	data := []byte(`{"tc":null,"ph":null,"dh":null,"ch":null,"al":null}`)
	var tc domain.ThreeCard
	err := json.Unmarshal(data, &tc)
	require.NoError(t, err)
	assert.NotNil(t, tc.GetPlayerHand())
	assert.NotNil(t, tc.GetDealerHand())
	assert.NotNil(t, tc.GetActionLog())
}

func TestThreeCard_JSON_Unmarshal_OversizedArray(t *testing.T) {
	// Create a JSON with oversized action log
	huge := make([]*domain.ActionLogEntry, 1001)
	data, _ := json.Marshal(map[string]any{"al": huge})
	var tc domain.ThreeCard
	err := json.Unmarshal(data, &tc)
	assert.Error(t, err)
}

// #5513: Web はラウンド終了時の ante/pairPlus を React 側に覚えてワンクリック
// 再ベットできるのに、CLI/CUI には同等のコマンドが無く毎回手打ちしていた。
// **Reset が anteBet/pairPlusBet を消す**ので、覚える場所がドメインに要る。
func TestThreeCard_Rebet(t *testing.T) {
	newBetPhase := func() *domain.ThreeCard {
		tc := domain.NewDefaultThreeCard()
		tc.Reset()
		return tc
	}

	t.Run("repeats the previous round's amounts", func(t *testing.T) {
		tc := newBetPhase()
		require.NoError(t, tc.Bet(20, 10))
		tc.Reset() // 次のラウンドへ

		require.NoError(t, tc.Rebet())
		assert.Equal(t, 20, tc.GetAnteBet())
		assert.Equal(t, 10, tc.GetPairPlusBet())
	})

	// **一度も賭けていなければ、その理由で断る。** ガードが無くても Bet が
	// 「アンテ額が不正」で弾くのでエラーにはなるが、それは的外れな説明になる
	// (プレイヤーは額を打っていない)。理由まで見て初めてガードが検証できる。
	t.Run("refuses before any bet has been placed, and says why", func(t *testing.T) {
		err := newBetPhase().Rebet()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "まだ賭けていない")
	})

	// **チップ不足は明確に断る。** Bet と同じ検査を通すので理由も同じ。
	t.Run("refuses when the chips no longer cover it", func(t *testing.T) {
		tc := newBetPhase()
		require.NoError(t, tc.Bet(20, 10))
		tc.Reset()
		tc.SetChips(5)
		assert.Error(t, tc.Rebet())
	})

	// **保存に載っていなければ Worker では毎回消える。** KV は毎リクエスト
	// 状態を往復させるので、スナップショットに入っていない値は 0 で戻る。
	t.Run("survives a save/load round trip", func(t *testing.T) {
		tc := newBetPhase()
		require.NoError(t, tc.Bet(30, 20))
		tc.Reset()

		data, err := json.Marshal(tc)
		require.NoError(t, err)
		restored := domain.NewDefaultThreeCard()
		require.NoError(t, json.Unmarshal(data, restored))

		require.NoError(t, restored.Rebet())
		assert.Equal(t, 30, restored.GetAnteBet())
		assert.Equal(t, 20, restored.GetPairPlusBet())
	})
}
