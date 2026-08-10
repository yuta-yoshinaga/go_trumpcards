package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewDefaultMississippiStud(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	assert.Equal(t, domain.MississippiStudPhaseAnte, m.GetPhase())
	assert.Equal(t, domain.MississippiStudDefaultChips, m.GetChips())
	assert.False(t, m.GetGameEndFlag())
	assert.Nil(t, m.GetPlayerHand())
	assert.Nil(t, m.GetCommunityCards())
	assert.Equal(t, [domain.MississippiStudStreetCnt]int{}, m.GetStreetMultipliers())
}

func TestMississippiStud_Reset(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	require.NoError(t, m.Bet(100))
	require.NoError(t, m.Play(1))
	require.NoError(t, m.Play(1))
	require.NoError(t, m.Play(1))
	assert.Equal(t, domain.MississippiStudPhaseEnd, m.GetPhase())

	m.Reset()
	assert.Equal(t, domain.MississippiStudPhaseAnte, m.GetPhase())
	assert.False(t, m.GetGameEndFlag())
	assert.Nil(t, m.GetPlayerHand())
	assert.Nil(t, m.GetCommunityCards())
	assert.Equal(t, 0, m.GetAnteAmount())
	assert.Equal(t, [domain.MississippiStudStreetCnt]int{}, m.GetStreetMultipliers())
	assert.False(t, m.GetFolded())
}

func TestMississippiStud_Reset_RefillChips(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	m.SetChips(20) // ante=10 + 3*1x = 4*10 = 40 必要
	m.Reset()
	assert.Equal(t, domain.MississippiStudDefaultChips, m.GetChips())
}

func TestMississippiStud_Reset_NoRefillAboveThreshold(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	m.SetChips(500)
	m.Reset()
	assert.Equal(t, 500, m.GetChips())
}

func TestMississippiStud_Bet_WrongPhase(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	m.SetPhase(domain.MississippiStudPhaseThirdSt)
	err := m.Bet(100)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestMississippiStud_Bet_InvalidAmount(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		amount int
	}{
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
		{"Zero", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := domain.NewDefaultMississippiStud()
			err := m.Bet(tc.amount)
			require.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestMississippiStud_Bet_InsufficientChips(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	m.SetChips(50) // ante=100 不可
	err := m.Bet(100)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestMississippiStud_Bet_Success(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	err := m.Bet(100)
	require.NoError(t, err)
	assert.Equal(t, domain.MississippiStudPhaseThirdSt, m.GetPhase())
	assert.Equal(t, 100, m.GetAnteAmount())
	assert.Len(t, m.GetPlayerHand(), 2)
	assert.Len(t, m.GetCommunityCards(), 3)
	assert.Equal(t, domain.MississippiStudDefaultChips-100, m.GetChips())
	// すべてのコミュニティは伏せ
	revealed := m.GetCommunityRevealed()
	for _, r := range revealed {
		assert.False(t, r)
	}
}

func TestMississippiStud_Play_AdvancesPhase(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	require.NoError(t, m.Bet(100))
	require.NoError(t, m.Play(2))
	assert.Equal(t, domain.MississippiStudPhaseFourthSt, m.GetPhase())
	assert.Equal(t, [domain.MississippiStudStreetCnt]int{2, 0, 0}, m.GetStreetMultipliers())
	assert.True(t, m.GetCommunityRevealed()[0])
	assert.False(t, m.GetCommunityRevealed()[1])
	assert.Equal(t, domain.MississippiStudDefaultChips-100-200, m.GetChips())

	require.NoError(t, m.Play(1))
	assert.Equal(t, domain.MississippiStudPhaseFifthSt, m.GetPhase())
	assert.True(t, m.GetCommunityRevealed()[1])
	assert.False(t, m.GetCommunityRevealed()[2])

	require.NoError(t, m.Play(3))
	assert.Equal(t, domain.MississippiStudPhaseEnd, m.GetPhase())
	assert.True(t, m.GetGameEndFlag())
}

func TestMississippiStud_Play_WrongPhase(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	err := m.Play(1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestMississippiStud_Play_InvalidMultiplier(t *testing.T) {
	t.Parallel()
	cases := []int{0, 4, -1, 99}
	for _, mult := range cases {
		m := domain.NewDefaultMississippiStud()
		require.NoError(t, m.Bet(100))
		err := m.Play(mult)
		require.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	}
}

func TestMississippiStud_Play_InsufficientChips(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	m.SetChips(100)
	require.NoError(t, m.Bet(100))
	// chips = 0, attempting 1x play (cost 100) → 不足
	err := m.Play(1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestMississippiStud_Fold_WrongPhase(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	err := m.Fold()
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestMississippiStud_Fold_ThirdSt_LosesAllBets(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	require.NoError(t, m.Bet(100))
	chipsBeforeFold := m.GetChips()
	require.NoError(t, m.Fold())

	assert.True(t, m.GetGameEndFlag())
	assert.Equal(t, domain.MississippiStudPhaseEnd, m.GetPhase())
	assert.True(t, m.GetFolded())
	assert.Equal(t, domain.GameResultLose, m.GetResult())
	assert.Equal(t, 0, m.GetTotalPayout())
	assert.Equal(t, chipsBeforeFold, m.GetChips()) // チップ残高に変化なし
}

func TestMississippiStud_Fold_FifthSt_LosesAllBets(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	require.NoError(t, m.Bet(100))
	require.NoError(t, m.Play(3))
	require.NoError(t, m.Play(2))
	chipsBeforeFold := m.GetChips()
	require.NoError(t, m.Fold())

	assert.Equal(t, domain.GameResultLose, m.GetResult())
	assert.Equal(t, 0, m.GetTotalPayout())
	assert.Equal(t, chipsBeforeFold, m.GetChips())
	assert.Equal(t, 100+300+200, m.GetTotalBet())
}

// hand 構築用のヘルパ
func msCard(t *testing.T, design, value int) *domain.Card {
	t.Helper()
	return domain.NewCard(design, value, true)
}

// playToEnd は与えられたホール/コミュニティでフルロード解決を行う。
func msPlayToEnd(t *testing.T, hole, comm []*domain.Card) *domain.MississippiStud {
	t.Helper()
	m := domain.NewDefaultMississippiStud()
	require.NoError(t, m.Bet(100))
	m.SetPlayerHand(hole)
	m.SetCommunityCards(comm)
	require.NoError(t, m.Play(1))
	require.NoError(t, m.Play(1))
	require.NoError(t, m.Play(1))
	return m
}

func TestMississippiStud_Resolve_Paytable(t *testing.T) {
	t.Parallel()
	// design: 0=spades,1=hearts,2=diamonds,3=clubs (本リポジトリは internal/domain/Card.go の定数群)
	tests := []struct {
		name           string
		hole           []*domain.Card
		comm           []*domain.Card
		wantMultiplier int
		wantResult     domain.GameResult
		wantHandRank   int
	}{
		{
			name:           "RoyalFlush",
			hole:           []*domain.Card{msCard(t, 0, 1), msCard(t, 0, 13)},
			comm:           []*domain.Card{msCard(t, 0, 12), msCard(t, 0, 11), msCard(t, 0, 10)},
			wantMultiplier: domain.MississippiStudPayRoyalFlush,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandRoyalFlush,
		},
		{
			name:           "StraightFlush",
			hole:           []*domain.Card{msCard(t, 1, 9), msCard(t, 1, 8)},
			comm:           []*domain.Card{msCard(t, 1, 7), msCard(t, 1, 6), msCard(t, 1, 5)},
			wantMultiplier: domain.MississippiStudPayStraightFlush,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandStraightFlush,
		},
		{
			name:           "FourOfAKind",
			hole:           []*domain.Card{msCard(t, 0, 7), msCard(t, 1, 7)},
			comm:           []*domain.Card{msCard(t, 2, 7), msCard(t, 3, 7), msCard(t, 0, 2)},
			wantMultiplier: domain.MississippiStudPayFourOfAKind,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandFourOfAKind,
		},
		{
			name:           "FullHouse",
			hole:           []*domain.Card{msCard(t, 0, 8), msCard(t, 1, 8)},
			comm:           []*domain.Card{msCard(t, 2, 8), msCard(t, 0, 5), msCard(t, 1, 5)},
			wantMultiplier: domain.MississippiStudPayFullHouse,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandFullHouse,
		},
		{
			name:           "Flush",
			hole:           []*domain.Card{msCard(t, 2, 2), msCard(t, 2, 5)},
			comm:           []*domain.Card{msCard(t, 2, 8), msCard(t, 2, 11), msCard(t, 2, 13)},
			wantMultiplier: domain.MississippiStudPayFlush,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandFlush,
		},
		{
			name:           "Straight",
			hole:           []*domain.Card{msCard(t, 0, 5), msCard(t, 1, 6)},
			comm:           []*domain.Card{msCard(t, 2, 7), msCard(t, 3, 8), msCard(t, 0, 9)},
			wantMultiplier: domain.MississippiStudPayStraight,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandStraight,
		},
		{
			name:           "ThreeOfAKind",
			hole:           []*domain.Card{msCard(t, 0, 9), msCard(t, 1, 9)},
			comm:           []*domain.Card{msCard(t, 2, 9), msCard(t, 3, 2), msCard(t, 0, 5)},
			wantMultiplier: domain.MississippiStudPayThreeOfAKind,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandThreeOfAKind,
		},
		{
			name:           "TwoPair",
			hole:           []*domain.Card{msCard(t, 0, 4), msCard(t, 1, 4)},
			comm:           []*domain.Card{msCard(t, 2, 9), msCard(t, 3, 9), msCard(t, 0, 2)},
			wantMultiplier: domain.MississippiStudPayTwoPair,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandTwoPair,
		},
		{
			name:           "HighPair_Jacks",
			hole:           []*domain.Card{msCard(t, 0, 11), msCard(t, 1, 11)},
			comm:           []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 5), msCard(t, 0, 2)},
			wantMultiplier: domain.MississippiStudPayHighPair,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandOnePair,
		},
		{
			name:           "HighPair_Aces",
			hole:           []*domain.Card{msCard(t, 0, 1), msCard(t, 1, 1)},
			comm:           []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 5), msCard(t, 0, 2)},
			wantMultiplier: domain.MississippiStudPayHighPair,
			wantResult:     domain.GameResultWin,
			wantHandRank:   domain.PokerHandOnePair,
		},
		{
			name:           "Push_PairOfSixes",
			hole:           []*domain.Card{msCard(t, 0, 6), msCard(t, 1, 6)},
			comm:           []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 5), msCard(t, 0, 2)},
			wantMultiplier: domain.MississippiStudPayPush,
			wantResult:     domain.GameResultDraw,
			wantHandRank:   domain.PokerHandOnePair,
		},
		{
			name:           "Push_PairOfTens",
			hole:           []*domain.Card{msCard(t, 0, 10), msCard(t, 1, 10)},
			comm:           []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 5), msCard(t, 0, 2)},
			wantMultiplier: domain.MississippiStudPayPush,
			wantResult:     domain.GameResultDraw,
			wantHandRank:   domain.PokerHandOnePair,
		},
		{
			name:           "Loss_PairOfFives",
			hole:           []*domain.Card{msCard(t, 0, 5), msCard(t, 1, 5)},
			comm:           []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 3), msCard(t, 0, 2)},
			wantMultiplier: domain.MississippiStudPayLoss,
			wantResult:     domain.GameResultLose,
			wantHandRank:   domain.PokerHandOnePair,
		},
		{
			name:           "Loss_HighCard",
			hole:           []*domain.Card{msCard(t, 0, 2), msCard(t, 1, 5)},
			comm:           []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 11), msCard(t, 0, 9)},
			wantMultiplier: domain.MississippiStudPayLoss,
			wantResult:     domain.GameResultLose,
			wantHandRank:   domain.PokerHandHighCard,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := msPlayToEnd(t, tc.hole, tc.comm)
			assert.Equal(t, tc.wantHandRank, m.GetHandRank(), "hand rank")
			assert.Equal(t, tc.wantMultiplier, m.GetPayoutMultiplier(), "payout multiplier")
			assert.Equal(t, tc.wantResult, m.GetResult(), "result")
		})
	}
}

func TestMississippiStud_Resolve_TotalPayout_Win(t *testing.T) {
	t.Parallel()
	// ハイペア(1:1): ante=100, streets all 1x → total bet = 100 + 100*3 = 400
	// Win → returns 400 + 400*1 = 800
	hole := []*domain.Card{msCard(t, 0, 11), msCard(t, 1, 11)}
	comm := []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 5), msCard(t, 0, 2)}
	m := msPlayToEnd(t, hole, comm)
	assert.Equal(t, 800, m.GetTotalPayout())
	// chip: start=1000, paid 100+100+100+100=400, received 800 → 1000-400+800 = 1400
	assert.Equal(t, 1400, m.GetChips())
}

func TestMississippiStud_Resolve_TotalPayout_Push(t *testing.T) {
	t.Parallel()
	// ペア6(プッシュ): ante=100, streets all 1x → total bet=400
	// Push → returns 400 (元金のみ)
	hole := []*domain.Card{msCard(t, 0, 6), msCard(t, 1, 6)}
	comm := []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 5), msCard(t, 0, 2)}
	m := msPlayToEnd(t, hole, comm)
	assert.Equal(t, 400, m.GetTotalPayout())
	// チップは差し引きゼロ → 元の 1000 のまま
	assert.Equal(t, 1000, m.GetChips())
}

func TestMississippiStud_Resolve_TotalPayout_Loss(t *testing.T) {
	t.Parallel()
	hole := []*domain.Card{msCard(t, 0, 5), msCard(t, 1, 5)}
	comm := []*domain.Card{msCard(t, 2, 8), msCard(t, 3, 3), msCard(t, 0, 2)}
	m := msPlayToEnd(t, hole, comm)
	assert.Equal(t, 0, m.GetTotalPayout())
	// チップは元の 1000 - 400 = 600
	assert.Equal(t, 600, m.GetChips())
}

func TestMississippiStud_GetTotalBet_AllStreets3x(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	require.NoError(t, m.Bet(100))
	require.NoError(t, m.Play(3))
	require.NoError(t, m.Play(3))
	// during 5th street, GetTotalBet before last play
	assert.Equal(t, 100+300+300, m.GetTotalBet())
}

func TestMississippiStud_ActionLog_Recorded(t *testing.T) {
	t.Parallel()
	m := domain.NewDefaultMississippiStud()
	require.NoError(t, m.Bet(100))
	require.NoError(t, m.Play(1))
	require.NoError(t, m.Fold())
	log := m.GetActionLog()
	require.NotEmpty(t, log)
	// 期待エントリ: ante, deal, play, fold, result
	types := make([]string, 0, len(log))
	for _, e := range log {
		types = append(types, e.ActionType)
	}
	assert.Contains(t, types, "ante")
	assert.Contains(t, types, "deal")
	assert.Contains(t, types, "play")
	assert.Contains(t, types, "fold")
	assert.Contains(t, types, "result")
}

// **CUI には役評価も配当対象判定も推奨倍率も無かった (#4710)。**Web は
// ms-made-hand に役と配当対象かを常時出している。
func TestMississippiStud_CurrentMadeHandAndRecommendBet(t *testing.T) {
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }
	street := func(phase int, hole []*domain.Card, community []*domain.Card, revealed [3]bool) *domain.MississippiStud {
		m := domain.NewDefaultMississippiStud()
		m.SetPhase(phase)
		m.SetPlayerHand(hole)
		m.SetCommunityCards(community)
		m.SetCommunityRevealed(revealed)
		return m
	}
	noBoard := []*domain.Card{
		card(domain.CardDesignClover, 2), card(domain.CardDesignDiamond, 4), card(domain.CardDesignHeart, 9),
	}
	hidden := [3]bool{false, false, false}

	t.Run("no made hand from two unpaired cards", func(t *testing.T) {
		m := street(domain.MississippiStudPhaseThirdSt,
			[]*domain.Card{card(domain.CardDesignSpade, 7), card(domain.CardDesignHeart, 3)}, noBoard, hidden)
		assert.Nil(t, m.GetCurrentMadeHand())
	})

	// **2 のペアは「ワンペア」でも配当が付かない。**役名だけ出すと、
	// 配当の付かない手で 3 倍を置かせる。
	t.Run("a low pair is a made hand but does not pay", func(t *testing.T) {
		m := street(domain.MississippiStudPhaseThirdSt,
			[]*domain.Card{card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 2)}, noBoard, hidden)
		made := m.GetCurrentMadeHand()
		if assert.NotNil(t, made) {
			assert.Equal(t, domain.PokerHandOnePair, made.Rank)
			assert.False(t, made.PaytableEligible)
		}
	})

	t.Run("a pair of sixes pays", func(t *testing.T) {
		m := street(domain.MississippiStudPhaseThirdSt,
			[]*domain.Card{card(domain.CardDesignSpade, 6), card(domain.CardDesignHeart, 6)}, noBoard, hidden)
		made := m.GetCurrentMadeHand()
		if assert.NotNil(t, made) {
			assert.True(t, made.PaytableEligible)
		}
		assert.Equal(t, domain.MSRecommendPlay3x, m.RecommendBet())
	})

	// **エースのペアも配当対象。**value が 1 なので「6以上」だけ見ると外れる。
	t.Run("a pair of aces pays even though the value is 1", func(t *testing.T) {
		m := street(domain.MississippiStudPhaseThirdSt,
			[]*domain.Card{card(domain.CardDesignSpade, 1), card(domain.CardDesignHeart, 1)}, noBoard, hidden)
		made := m.GetCurrentMadeHand()
		if assert.NotNil(t, made) {
			assert.True(t, made.PaytableEligible)
		}
	})

	// **未公開のコミュニティカードは混ぜない。**プレイヤーがまだ見ていない
	// 札を根拠にした助言になる。
	t.Run("ignores community cards that are still face down", func(t *testing.T) {
		hole := []*domain.Card{card(domain.CardDesignSpade, 7), card(domain.CardDesignHeart, 3)}
		board := []*domain.Card{card(domain.CardDesignClover, 7), card(domain.CardDesignDiamond, 4), card(domain.CardDesignHeart, 9)}

		assert.Nil(t, street(domain.MississippiStudPhaseThirdSt, hole, board, hidden).GetCurrentMadeHand())

		shown := street(domain.MississippiStudPhaseFourthSt, hole, board, [3]bool{true, false, false})
		made := shown.GetCurrentMadeHand()
		if assert.NotNil(t, made) {
			assert.Equal(t, domain.PokerHandOnePair, made.Rank, "公開された 7 でペアになる")
		}
	})

	t.Run("bets 1x on a flush draw with nothing made", func(t *testing.T) {
		m := street(domain.MississippiStudPhaseFourthSt,
			[]*domain.Card{card(domain.CardDesignSpade, 2), card(domain.CardDesignSpade, 5)},
			[]*domain.Card{card(domain.CardDesignSpade, 9), card(domain.CardDesignDiamond, 4), card(domain.CardDesignHeart, 11)},
			[3]bool{true, false, false})
		assert.Equal(t, domain.MSRecommendPlay1x, m.RecommendBet())
	})

	t.Run("folds a hand with nothing made and no draw", func(t *testing.T) {
		m := street(domain.MississippiStudPhaseFourthSt,
			[]*domain.Card{card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 5)},
			[]*domain.Card{card(domain.CardDesignClover, 9), card(domain.CardDesignDiamond, 4), card(domain.CardDesignHeart, 11)},
			[3]bool{true, false, false})
		assert.Equal(t, domain.MSRecommendFold, m.RecommendBet())
	})

	// **3rd street だけはハイカード2枚でも賭ける。**あと3枚めくれるので
	// まだ伸びる。4th 以降で同じ手を賭けると、伸びない手に払い続ける。
	t.Run("two high cards are worth 1x only on third street", func(t *testing.T) {
		hole := []*domain.Card{card(domain.CardDesignSpade, 12), card(domain.CardDesignHeart, 11)}
		third := street(domain.MississippiStudPhaseThirdSt, hole, noBoard, hidden)
		assert.Equal(t, domain.MSRecommendPlay1x, third.RecommendBet())

		fourth := street(domain.MississippiStudPhaseFourthSt, hole,
			[]*domain.Card{card(domain.CardDesignClover, 2), card(domain.CardDesignDiamond, 4), card(domain.CardDesignHeart, 9)},
			[3]bool{true, false, false})
		assert.Equal(t, domain.MSRecommendFold, fourth.RecommendBet())
	})

	t.Run("no recommendation outside the betting streets", func(t *testing.T) {
		m := street(domain.MississippiStudPhaseAnte,
			[]*domain.Card{card(domain.CardDesignSpade, 6), card(domain.CardDesignHeart, 6)}, noBoard, hidden)
		assert.Equal(t, "", m.RecommendBet())
	})
}
