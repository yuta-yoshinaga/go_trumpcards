package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- ヘルパー ---

func card(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func jokerCard() *domain.Card {
	return domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false)
}

// --- Constructor / Reset ---

func TestNewDefaultPaiGow(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	assert.Equal(t, domain.PaiGowPhaseBet, pg.GetPhase())
	assert.Equal(t, domain.PaiGowDefaultChips, pg.GetChips())
	assert.False(t, pg.GetGameEndFlag())
	assert.Nil(t, pg.GetPlayerCards())
	assert.Nil(t, pg.GetDealerCards())
}

func TestPaiGow_Reset(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	err := pg.Bet(100)
	require.NoError(t, err)
	// Force end by setting hands
	pg.SetDealerCards([]*domain.Card{
		card(1, 2), card(1, 3), card(1, 4), card(1, 5), card(1, 6), card(1, 7), card(1, 8),
	})
	_ = pg.SetHands(0, 1) // may or may not error depending on hand validity

	pg.Reset()
	assert.Equal(t, domain.PaiGowPhaseBet, pg.GetPhase())
	assert.False(t, pg.GetGameEndFlag())
	assert.Nil(t, pg.GetPlayerCards())
	assert.Nil(t, pg.GetDealerCards())
	assert.Equal(t, 0, pg.GetBet())
}

func TestPaiGow_Reset_RefillChips(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetChips(5) // Below minimum bet
	pg.Reset()
	assert.Equal(t, domain.PaiGowDefaultChips, pg.GetChips())
}

// --- Bet ---

func TestPaiGow_Bet_Success(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	err := pg.Bet(100)
	require.NoError(t, err)
	assert.Equal(t, domain.PaiGowPhaseSetHands, pg.GetPhase())
	assert.Equal(t, 100, pg.GetBet())
	assert.Equal(t, domain.PaiGowDefaultChips-100, pg.GetChips())
	assert.Len(t, pg.GetPlayerCards(), 7)
	assert.Len(t, pg.GetDealerCards(), 7)
}

func TestPaiGow_Bet_WrongPhase(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	err := pg.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestPaiGow_Bet_InvalidAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int
	}{
		{"too low", 5},
		{"not multiple", 15},
		{"too high", 20000},
		{"zero", 0},
		{"negative", -10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := domain.NewDefaultPaiGow()
			err := pg.Bet(tt.amount)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestPaiGow_Bet_InsufficientChips(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetChips(50)
	err := pg.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

// --- SetHands ---

func TestPaiGow_SetHands_WrongPhase(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	err := pg.SetHands(0, 1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestPaiGow_SetHands_SameIndex(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	pg.SetPlayerCards(make([]*domain.Card, 7))
	err := pg.SetHands(2, 2)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidCard))
}

func TestPaiGow_SetHands_OutOfRange(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	pg.SetPlayerCards(make([]*domain.Card, 7))

	tests := []struct {
		name string
		i, j int
	}{
		{"negative first", -1, 3},
		{"negative second", 3, -1},
		{"too large first", 7, 3},
		{"too large second", 3, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pg.SetHands(tt.i, tt.j)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidCard))
		})
	}
}

func TestPaiGow_SetHands_HighMustBeatLow(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	pg.SetBet(100)
	// カード: A,A が indices 0,1, 残りの5枚が弱い (2,3,7,9,J) ← ストレートにならない
	pg.SetPlayerCards([]*domain.Card{
		card(1, 1), card(2, 1), // AA (pair)
		card(1, 2), card(2, 3), card(3, 7), card(4, 9), card(1, 11),
	})
	pg.SetDealerCards([]*domain.Card{
		card(1, 2), card(1, 3), card(1, 4), card(1, 5), card(1, 7), card(1, 9), card(1, 11),
	})
	// indices 0,1 を low にすると low=AA(pair), high=2,3,7,9,J(high card) → 無効
	err := pg.SetHands(0, 1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))

	// #5526: 生の英文をそのまま画面に出していた。プレイヤーの言語で
	// ルールを説明できるよう、i18n キーを名乗る形にする。
	code, _ := domain.ErrorMessageCode(err)
	assert.Equal(t, "paigow.foulHighMustBeat", code)
	assert.NotContains(t, err.Error(), "High hand must be stronger")
}

func TestPaiGow_SetHands_ValidSplit(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	pg.SetBet(100)
	pg.SetChips(0) // already subtracted
	// Player: K,K,Q,J,10,3,2 → low=3,2 high=K,K,Q,J,10 (pair of K)
	pg.SetPlayerCards([]*domain.Card{
		card(1, 13), card(2, 13), card(1, 12), card(2, 11), card(3, 10), card(4, 3), card(1, 2),
	})
	// Dealer: weak hand
	pg.SetDealerCards([]*domain.Card{
		card(1, 2), card(2, 3), card(3, 4), card(4, 5), card(1, 7), card(2, 8), card(3, 9),
	})
	err := pg.SetHands(5, 6) // low=3,2
	require.NoError(t, err)
	assert.Equal(t, domain.PaiGowPhaseEnd, pg.GetPhase())
	assert.True(t, pg.GetGameEndFlag())
	assert.Len(t, pg.GetPlayerHighHand(), 5)
	assert.Len(t, pg.GetPlayerLowHand(), 2)
	assert.Len(t, pg.GetDealerHighHand(), 5)
	assert.Len(t, pg.GetDealerLowHand(), 2)
}

// --- Settlement ---

func TestPaiGow_BothWin(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	pg.SetBet(100)
	pg.SetChips(0)
	// Player: A,K,Q,J,10 (straight) high + A,K low
	pg.SetPlayerCards([]*domain.Card{
		card(1, 1), card(1, 13), card(1, 12), card(1, 11), card(2, 10),
		card(3, 1), card(4, 13),
	})
	// Dealer: weak
	pg.SetDealerCards([]*domain.Card{
		card(1, 2), card(2, 3), card(3, 4), card(4, 5), card(1, 7),
		card(2, 8), card(3, 6),
	})
	err := pg.SetHands(5, 6) // low=A♦,K♦
	require.NoError(t, err)
	assert.Equal(t, domain.GameResultWin, pg.GetResult())
	assert.Equal(t, domain.GameResultWin, pg.GetHighHandResult())
	assert.Equal(t, domain.GameResultWin, pg.GetLowHandResult())
	// Payout: bet*2 - 5% commission on winnings (bet amount)
	expectedCommission := 5 // 100 * 5 / 100
	expectedPayout := 200 - expectedCommission
	assert.Equal(t, expectedCommission, pg.GetCommission())
	assert.Equal(t, expectedPayout, pg.GetPayout())
	assert.Equal(t, expectedPayout, pg.GetChips())
}

func TestPaiGow_BothLose(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	pg.SetBet(100)
	pg.SetChips(0)
	// Player: weak hand high=J,9,7,4,3 low=3,2
	pg.SetPlayerCards([]*domain.Card{
		card(1, 11), card(2, 9), card(3, 7), card(4, 4), card(1, 3),
		card(2, 3), card(3, 2),
	})
	// Dealer: strong (A,K,Q,J,10 straight + A,K)
	pg.SetDealerCards([]*domain.Card{
		card(1, 1), card(1, 13), card(1, 12), card(1, 11), card(2, 10),
		card(3, 1), card(4, 13),
	})
	err := pg.SetHands(5, 6) // low=3♣,2♦
	require.NoError(t, err)
	assert.Equal(t, domain.GameResultLose, pg.GetResult())
	assert.Equal(t, 0, pg.GetPayout())
	assert.Equal(t, 0, pg.GetChips())
}

func TestPaiGow_Push_SplitResult(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	pg.SetBet(100)
	pg.SetChips(0)
	// Player: strong high (A,A,K,Q,J) but weak low (2,3)
	pg.SetPlayerCards([]*domain.Card{
		card(1, 1), card(2, 1), card(1, 13), card(2, 12), card(3, 11),
		card(4, 2), card(1, 3),
	})
	// Dealer: medium high (K,K,Q,J,10) but strong low (A,Q)
	pg.SetDealerCards([]*domain.Card{
		card(3, 13), card(4, 13), card(3, 12), card(4, 11), card(1, 10),
		card(2, 1), card(3, 12),
	})
	err := pg.SetHands(5, 6) // low=2,3
	require.NoError(t, err)
	assert.Equal(t, domain.GameResultDraw, pg.GetResult())
	assert.Equal(t, 100, pg.GetPayout()) // bet returned
	assert.Equal(t, 100, pg.GetChips())
}

func TestPaiGow_TieGoesToDealer_HighHand(t *testing.T) {
	// Test that ties in hand comparison go to dealer (result 0 from compare means tie → dealer wins)
	// We test this indirectly through comparePaiGowHighHands returning 0
	a := []*domain.Card{card(1, 10), card(2, 9), card(3, 8), card(4, 7), card(1, 2)}
	b := []*domain.Card{card(4, 10), card(3, 9), card(2, 8), card(1, 7), card(4, 2)}
	cmp := domain.ComparePaiGowHighHandsExported(a, b)
	assert.Equal(t, 0, cmp) // tie

	// Same for low hands
	la := []*domain.Card{card(1, 5), card(2, 3)}
	lb := []*domain.Card{card(3, 5), card(4, 3)}
	lcmp := domain.ComparePaiGowLowHandsExported(la, lb)
	assert.Equal(t, 0, lcmp) // tie
}

func TestPaiGow_TieGoesToDealer_FullGame(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPhase(domain.PaiGowPhaseSetHands)
	pg.SetBet(100)
	pg.SetChips(0)
	// Cards: A,K,J,9,7,4,2 (no pairs/straights/flushes)
	// House way maximizes low: best valid low = K,J → high = A,9,7,4,2
	// Player does the same: low=K(idx1),J(idx2)
	pg.SetPlayerCards([]*domain.Card{
		card(1, 1), card(2, 13), card(3, 11), card(4, 9), card(1, 7),
		card(2, 4), card(3, 2),
	})
	pg.SetDealerCards([]*domain.Card{
		card(4, 1), card(3, 13), card(2, 11), card(1, 9), card(4, 7),
		card(3, 4), card(2, 2),
	})
	err := pg.SetHands(1, 2) // low=K,J; high=A,9,7,4,2
	require.NoError(t, err)
	// Tie on both → dealer wins
	assert.Equal(t, domain.GameResultLose, pg.GetResult())
	assert.Equal(t, 0, pg.GetPayout())
}

// --- Hand Evaluation ---

func TestEvalPaiGowLowHand_Pair(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPlayerLowHand([]*domain.Card{card(1, 5), card(2, 5)})
	assert.Equal(t, domain.PaiGowLowHandPair, domain.EvalPaiGowLowHandExported(pg.GetPlayerLowHand()))
}

func TestEvalPaiGowLowHand_HighCard(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.SetPlayerLowHand([]*domain.Card{card(1, 10), card(2, 5)})
	assert.Equal(t, domain.PaiGowLowHandHighCard, domain.EvalPaiGowLowHandExported(pg.GetPlayerLowHand()))
}

func TestEvalPaiGowLowHand_JokerAsAce(t *testing.T) {
	// Joker + Ace = pair of aces
	low := []*domain.Card{jokerCard(), card(1, 1)}
	assert.Equal(t, domain.PaiGowLowHandPair, domain.EvalPaiGowLowHandExported(low))
}

func TestEvalPaiGowHighHand_WithJoker_CompletesFlush(t *testing.T) {
	// 4 hearts + joker = flush
	high := []*domain.Card{
		card(3, 2), card(3, 5), card(3, 8), card(3, 11), jokerCard(),
	}
	rank := domain.EvalPaiGowHighHandExported(high)
	assert.GreaterOrEqual(t, rank, domain.PokerHandFlush)
}

func TestEvalPaiGowHighHand_WithJoker_CompletesStraight(t *testing.T) {
	// 4,5,6,7,Joker → straight (joker=8 or 3)
	high := []*domain.Card{
		card(1, 4), card(2, 5), card(3, 6), card(4, 7), jokerCard(),
	}
	rank := domain.EvalPaiGowHighHandExported(high)
	assert.GreaterOrEqual(t, rank, domain.PokerHandStraight)
}

func TestEvalPaiGowHighHand_WithJoker_DefaultAce(t *testing.T) {
	// No straight/flush possible → joker = ace
	high := []*domain.Card{
		card(1, 2), card(2, 5), card(3, 8), card(4, 11), jokerCard(),
	}
	rank := domain.EvalPaiGowHighHandExported(high)
	// With joker as ace, it's just high card (A high)
	assert.Equal(t, domain.PokerHandHighCard, rank)
}

func TestEvalPaiGowHighHand_NoJoker(t *testing.T) {
	high := []*domain.Card{
		card(1, 1), card(1, 13), card(1, 12), card(1, 11), card(1, 10),
	}
	rank := domain.EvalPaiGowHighHandExported(high)
	assert.Equal(t, domain.PokerHandRoyalFlush, rank)
}

// --- House Way ---

func TestPaiGowHouseWay_SplitsCorrectly(t *testing.T) {
	// Full house: K,K,K,5,5,3,2 → high=K,K,K,3,2 (trips) low=5,5 (pair)
	cards := []*domain.Card{
		card(1, 13), card(2, 13), card(3, 13), card(1, 5), card(2, 5),
		card(3, 3), card(4, 2),
	}
	high, low := domain.PaiGowHouseWayExported(cards)
	require.Len(t, high, 5)
	require.Len(t, low, 2)

	// Low should be the pair of 5s (best low hand from house way)
	lowRank := domain.EvalPaiGowLowHandExported(low)
	assert.Equal(t, domain.PaiGowLowHandPair, lowRank)
}

func TestPaiGowHouseWay_NoPair(t *testing.T) {
	// All different: A,K,Q,J,9,7,3
	cards := []*domain.Card{
		card(1, 1), card(2, 13), card(3, 12), card(4, 11), card(1, 9),
		card(2, 7), card(3, 3),
	}
	high, low := domain.PaiGowHouseWayExported(cards)
	require.Len(t, high, 5)
	require.Len(t, low, 2)
}

// --- Compare Low Hands ---

func TestComparePaiGowLowHands(t *testing.T) {
	tests := []struct {
		name string
		a, b []*domain.Card
		want int
	}{
		{
			name: "pair beats high card",
			a:    []*domain.Card{card(1, 5), card(2, 5)},
			b:    []*domain.Card{card(1, 1), card(2, 13)},
			want: 1,
		},
		{
			name: "higher pair wins",
			a:    []*domain.Card{card(1, 10), card(2, 10)},
			b:    []*domain.Card{card(1, 5), card(2, 5)},
			want: 1,
		},
		{
			name: "higher kicker wins in high card",
			a:    []*domain.Card{card(1, 1), card(2, 13)},
			b:    []*domain.Card{card(1, 1), card(2, 12)},
			want: 1,
		},
		{
			name: "tie",
			a:    []*domain.Card{card(1, 10), card(2, 5)},
			b:    []*domain.Card{card(3, 10), card(4, 5)},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ComparePaiGowLowHandsExported(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Compare High Hands ---

func TestComparePaiGowHighHands(t *testing.T) {
	tests := []struct {
		name string
		a, b []*domain.Card
		want int
	}{
		{
			name: "flush beats straight",
			a:    []*domain.Card{card(1, 2), card(1, 5), card(1, 8), card(1, 11), card(1, 13)},
			b:    []*domain.Card{card(1, 4), card(2, 5), card(3, 6), card(4, 7), card(1, 8)},
			want: 1,
		},
		{
			name: "same rank higher cards win",
			a:    []*domain.Card{card(1, 1), card(2, 13), card(3, 12), card(4, 11), card(1, 9)},
			b:    []*domain.Card{card(1, 1), card(2, 13), card(3, 12), card(4, 11), card(1, 8)},
			want: 1,
		},
		{
			name: "tie",
			a:    []*domain.Card{card(1, 10), card(2, 9), card(3, 8), card(4, 7), card(1, 2)},
			b:    []*domain.Card{card(4, 10), card(3, 9), card(2, 8), card(1, 7), card(4, 2)},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ComparePaiGowHighHandsExported(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- JSON ---

func TestPaiGow_JSON_RoundTrip(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	err := pg.Bet(100)
	require.NoError(t, err)

	data, err := json.Marshal(pg)
	require.NoError(t, err)

	var restored domain.PaiGow
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	assert.Equal(t, pg.GetPhase(), restored.GetPhase())
	assert.Equal(t, pg.GetBet(), restored.GetBet())
	assert.Equal(t, pg.GetChips(), restored.GetChips())
	assert.Len(t, restored.GetPlayerCards(), 7)
	assert.Len(t, restored.GetDealerCards(), 7)
}

func TestPaiGow_UnmarshalJSON_SliceLimit(t *testing.T) {
	// Create a JSON with oversized arrays
	bigSlice := make([]*domain.Card, 1001)
	j := struct {
		PC []*domain.Card `json:"pc"`
	}{PC: bigSlice}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	var pg domain.PaiGow
	err = json.Unmarshal(data, &pg)
	assert.Error(t, err)
}

func TestPaiGow_UnmarshalJSON_NilFields(t *testing.T) {
	data := []byte(`{"ps":1}`)
	var pg domain.PaiGow
	err := json.Unmarshal(data, &pg)
	require.NoError(t, err)
	assert.NotNil(t, pg.GetPlayerCards())
	assert.NotNil(t, pg.GetDealerCards())
	assert.NotNil(t, pg.GetActionLog())
}

// --- Action Log ---

func TestPaiGow_ActionLog(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	err := pg.Bet(100)
	require.NoError(t, err)
	log := pg.GetActionLog()
	assert.GreaterOrEqual(t, len(log), 2) // bet + deal
}

// --- Getters with test helpers ---

func TestPaiGow_TestHelpers(t *testing.T) {
	pg := domain.NewDefaultPaiGow()

	pg.SetPhase(2)
	assert.Equal(t, 2, pg.GetPhase())

	pg.SetBet(200)
	assert.Equal(t, 200, pg.GetBet())

	pg.SetChips(500)
	assert.Equal(t, 500, pg.GetChips())

	pg.SetResult(domain.GameResultWin)
	assert.Equal(t, domain.GameResultWin, pg.GetResult())

	pg.SetGameEndFlag(true)
	assert.True(t, pg.GetGameEndFlag())

	pg.SetHighHandResult(domain.GameResultWin)
	assert.Equal(t, domain.GameResultWin, pg.GetHighHandResult())

	pg.SetLowHandResult(domain.GameResultLose)
	assert.Equal(t, domain.GameResultLose, pg.GetLowHandResult())

	pg.SetPayout(190)
	assert.Equal(t, 190, pg.GetPayout())

	pg.SetCommission(10)
	assert.Equal(t, 10, pg.GetCommission())

	pg.SetPlayerHighRank(5)
	assert.Equal(t, 5, pg.GetPlayerHighRank())

	pg.SetPlayerLowRank(1)
	assert.Equal(t, 1, pg.GetPlayerLowRank())

	pg.SetDealerHighRank(3)
	assert.Equal(t, 3, pg.GetDealerHighRank())

	pg.SetDealerLowRank(0)
	assert.Equal(t, 0, pg.GetDealerLowRank())

	cards := []*domain.Card{card(1, 1)}
	pg.SetPlayerHighHand(cards)
	assert.Equal(t, cards, pg.GetPlayerHighHand())

	pg.SetPlayerLowHand(cards)
	assert.Equal(t, cards, pg.GetPlayerLowHand())

	pg.SetDealerHighHand(cards)
	assert.Equal(t, cards, pg.GetDealerHighHand())

	pg.SetDealerLowHand(cards)
	assert.Equal(t, cards, pg.GetDealerLowHand())
}

func TestPaiGow_IsFoulSplit(t *testing.T) {
	t.Run("agreement with SetHands across all 21 pairs", func(t *testing.T) {
		// 手札は配らずに組む。Bet() から取ると、その配りに反則が 1 つも無い
		// (または全部が反則の) 局が出たとき、下の 2 分岐のどちらかが一度も
		// 走らないまま緑になる。この 7 枚は両方を必ず含む:
		//   低に A A を回すと low=ワンペア / high=7ハイ で反則、
		//   低に 2 3 を回すと low=3ハイ / high=Aのペア で反則にならない。
		cards := []*domain.Card{
			card(domain.CardDesignSpade, 1),
			card(domain.CardDesignHeart, 1),
			card(domain.CardDesignSpade, 2),
			card(domain.CardDesignHeart, 3),
			card(domain.CardDesignClover, 4),
			card(domain.CardDesignDiamond, 5),
			card(domain.CardDesignSpade, 7),
		}
		require.Len(t, cards, 7)

		// **符号の基準点。** SetHands は IsFoulSplit を呼ぶようになったので、
		// 下の一致ループだけでは判定を自分自身と比べているにすぎず、
		// 戻り値を反転させても一致したまま通る (実測済み)。この 2 行が
		// 「どちらが反則か」を外から固定する。
		//   (0,1): low = A のペア / high = 7 ハイ    -> low が勝つので反則
		//   (2,3): low = 3 ハイ    / high = A のペア -> high が勝つので反則でない
		anchor := domain.NewDefaultPaiGow()
		anchor.SetPhase(domain.PaiGowPhaseSetHands)
		anchor.SetPlayerCards(cards)
		assert.True(t, anchor.IsFoulSplit(0, 1), "ローに A A を回したら反則")
		assert.False(t, anchor.IsFoulSplit(2, 3), "ローに 2 3 を回したら反則ではない")

		agreementPairs := 0
		foulSeen, cleanSeen := 0, 0
		for i := range 7 {
			for j := i + 1; j < 7; j++ {
				// We need a fresh game for each SetHands call to avoid advancing phase if successful.
				pgClone := domain.NewDefaultPaiGow()
				pgClone.SetPhase(domain.PaiGowPhaseSetHands)
				pgClone.SetPlayerCards(cards)
				// We don't care about dealer cards since SetHands sets them by house way, but let's give some dummies.
				dummyDealer := make([]*domain.Card, 7)
				for k := range dummyDealer {
					dummyDealer[k] = card(domain.CardDesignSpade, 2)
				}
				pgClone.SetDealerCards(dummyDealer)

				isFoul := pgClone.IsFoulSplit(i, j)

				// Ensure IsFoulSplit did not mutate state
				assert.Nil(t, pgClone.GetPlayerHighHand())
				assert.Nil(t, pgClone.GetPlayerLowHand())

				err := pgClone.SetHands(i, j)
				if isFoul {
					require.Error(t, err)
					var de *domain.DomainError
					require.True(t, errors.As(err, &de))
					assert.ErrorIs(t, err, domain.ErrInvalidPlay)
					assert.Equal(t, "paigow.foulHighMustBeat", de.Code)
				} else {
					require.NoError(t, err)
					assert.Equal(t, domain.PaiGowPhaseEnd, pgClone.GetPhase()) // because SetHands resolves the game
				}
				if isFoul {
					foulSeen++
				} else {
					cleanSeen++
				}
				agreementPairs++
			}
		}
		assert.Equal(t, 21, agreementPairs)
		// 両分岐を実際に通ったことを見る。片方が 0 なら、上のループは
		// 一致を主張しているようで片側しか試していない。
		assert.Positive(t, foulSeen, "反則になる分割が 1 つも無い手札では一致を主張できない")
		assert.Positive(t, cleanSeen, "反則にならない分割が 1 つも無い手札では一致を主張できない")
	})

	t.Run("out of bounds and same index", func(t *testing.T) {
		pg := domain.NewDefaultPaiGow()
		pg.SetPhase(domain.PaiGowPhaseSetHands)
		dummyCards := make([]*domain.Card, 7)
		for i := range dummyCards {
			dummyCards[i] = card(domain.CardDesignSpade, i+2)
		}
		pg.SetPlayerCards(dummyCards)

		assert.True(t, pg.IsFoulSplit(0, 0)) // same index
		assert.True(t, pg.IsFoulSplit(-1, 2))
		assert.True(t, pg.IsFoulSplit(0, 7))
	})
}

// TestPaiGow_IsFoulSplitBeforeDeal は配る前に呼んでも panic しないことを見る。
//
// IsFoulSplit は interface 越しに公開されるので、SET HANDS 以外のフェーズからも
// 呼べてしまう。インデックスは PaiGowHandSize と突き合わせて通るのに playerCards
// はまだ空、という状態で添字を引くと panic する。
func TestPaiGow_IsFoulSplitBeforeDeal(t *testing.T) {
	pg := domain.NewDefaultPaiGow()
	pg.Reset()

	assert.Empty(t, pg.GetPlayerCards(), "前提: この時点で手札は配られていない")
	assert.NotPanics(t, func() {
		assert.True(t, pg.IsFoulSplit(0, 1))
	})
}
