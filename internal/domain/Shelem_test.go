//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestShelem(t *testing.T) *Shelem {
	t.Helper()
	s := NewDefaultShelem()
	s.Reset()
	return s
}

// shelemHandOf は指定プレイヤーの手札を固定の並びに差し替える。
func shelemHandOf(s *Shelem, idx int, cards ...*Card) {
	p := s.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// settleBidding は競りと捨て札を最後まで進める。
func settleBidding(t *testing.T, s *Shelem) {
	t.Helper()
	for range ShelemPlayerCnt * 4 {
		if s.GetPhase() != ShelemPhaseBid {
			break
		}
		if s.IsHumanBidTurn() {
			if err := s.PlayerPass(); err != nil {
				require.NoError(t, s.PlayerBid(ShelemMinBid))
			}
			continue
		}
		s.CpuBid()
	}
	if s.GetPhase() == ShelemPhaseDiscard && s.IsHumanDiscardTurn() {
		require.NoError(t, s.PlayerDiscard([]int{0, 1, 2, 3}, CardDesignSpade))
	}
	require.Equal(t, ShelemPhasePlay, s.GetPhase())
}

// --- 配り ---

// **12 枚 × 4 人 + ウィドウ 4 枚 = 52 枚。** issue の「52 枚 + ウィドウ 4 枚」は
// 56 枚になり、標準デッキでは配れない。
func TestShelem_DealIsTwelveEachPlusAFourCardWidow(t *testing.T) {
	s := newTestShelem(t)

	total := 0
	for i := range ShelemPlayerCnt {
		assert.Equal(t, ShelemHandSize, s.GetPlayer(i).GetCardsSize(), "player %d", i)
		total += s.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, ShelemWidowSize, s.GetWidowSize())
	assert.Equal(t, 52, total+s.GetWidowSize())
	assert.Equal(t, ShelemPhaseBid, s.GetPhase())
}

// --- カード点 ---

// **点になるのは A / 10 / 5 だけ。** 1 ラウンドの合計はちょうど 100 点。
func TestShelem_CardPoints(t *testing.T) {
	for _, tc := range []struct{ value, want int }{
		{1, 10}, {10, 10}, {5, 5},
		{2, 0}, {3, 0}, {4, 0}, {6, 0}, {7, 0}, {8, 0}, {9, 0}, {11, 0}, {12, 0}, {13, 0},
	} {
		assert.Equal(t, tc.want, ShelemCardPoints(NewCard(CardDesignSpade, tc.value, false)), "rank %d", tc.value)
	}
	assert.Equal(t, 0, ShelemCardPoints(nil))
}

func TestShelem_DeckHoldsExactlyOneHundredPoints(t *testing.T) {
	total := 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for v := 1; v <= 13; v++ {
			total += ShelemCardPoints(NewCard(suit, v, false))
		}
	}
	assert.Equal(t, ShelemHandPoints, total)
}

// --- 競り ---

// 入札は 100 から 5 刻み。範囲外・刻み違い・上回らない額は拒否する。
func TestShelem_BidValidation(t *testing.T) {
	s := newTestShelem(t)
	s.SetBidPlayerIdxForTest(0)

	assert.Error(t, s.PlayerBid(ShelemMinBid-ShelemBidStep))
	assert.Error(t, s.PlayerBid(ShelemMaxBid+ShelemBidStep))
	assert.Error(t, s.PlayerBid(102), "5 刻みでない")
	require.NoError(t, s.PlayerBid(ShelemMinBid))

	s.SetBidPlayerIdxForTest(0)
	s.GetPlayer(0).SetPassed(false)
	assert.Error(t, s.PlayerBid(ShelemMinBid), "同額は上回らない")
}

// 降りたら戻れない。
func TestShelem_PassedPlayerCannotBidAgain(t *testing.T) {
	s := newTestShelem(t)
	s.SetContractForTest(2, 110, false)
	s.SetBidPlayerIdxForTest(0)
	require.NoError(t, s.PlayerPass())
	assert.True(t, s.GetPlayer(0).GetPassed())

	s.SetBidPlayerIdxForTest(0)
	assert.Error(t, s.PlayerBid(120))
}

// **誰も入札しないまま最後の 1 人になったら降りられない。**
func TestShelem_LastBidderStandingCannotPass(t *testing.T) {
	s := newTestShelem(t)
	for _, i := range []int{1, 2, 3} {
		s.GetPlayer(i).SetPassed(true)
	}
	s.SetBidPlayerIdxForTest(0)

	err := s.PlayerPass()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must bid")
	assert.False(t, s.GetPlayer(0).GetPassed())
}

// **Shelem はどんな点数入札にも勝ち、その場で競りが決着する。**
func TestShelem_ShelemBidEndsTheAuction(t *testing.T) {
	s := newTestShelem(t)
	s.SetContractForTest(2, 150, false)
	s.SetBidPlayerIdxForTest(0)

	require.NoError(t, s.PlayerBidShelem())
	assert.True(t, s.GetShelemBid())
	assert.Equal(t, 0, s.GetDeclarerIdx())
	assert.NotEqual(t, ShelemPhaseBid, s.GetPhase(), "競りは即決着する")
}

// 競りは必ず決着し、落札者にウィドウが渡る。
func TestShelem_BiddingAlwaysSettlesAndHandsOverTheWidow(t *testing.T) {
	for range 20 {
		s := newTestShelem(t)
		s.SetDealerIdxForTest(1)
		for range ShelemPlayerCnt * 4 {
			if s.GetPhase() != ShelemPhaseBid {
				break
			}
			if s.IsHumanBidTurn() {
				if err := s.PlayerPass(); err != nil {
					require.NoError(t, s.PlayerBid(ShelemMinBid))
				}
				continue
			}
			s.CpuBid()
		}
		require.NotEqual(t, ShelemPhaseBid, s.GetPhase())
		assert.GreaterOrEqual(t, s.GetDeclarerIdx(), 0)
		assert.GreaterOrEqual(t, s.GetContract(), ShelemMinBid)
		// **ウィドウは落札者の手に入る。** 伏せ札は残らない。
		assert.Equal(t, 0, s.GetWidowSize())
		if s.GetPhase() == ShelemPhaseDiscard {
			assert.Equal(t, ShelemHandSize+ShelemWidowSize, s.GetPlayer(s.GetDeclarerIdx()).GetCardsSize())
		}
	}
}

// --- ウィドウ交換 ---

// **落札者は 16 枚から 4 枚捨てて 12 枚に戻す。**
func TestShelem_DiscardReturnsToTwelve(t *testing.T) {
	s := newTestShelem(t)
	s.SetContractForTest(0, 120, false)
	s.CloseBiddingForTest()
	require.Equal(t, ShelemHandSize+ShelemWidowSize, s.GetPlayer(0).GetCardsSize())

	require.NoError(t, s.PlayerDiscard([]int{0, 1, 2, 3}, CardDesignHeart))
	assert.Equal(t, ShelemHandSize, s.GetPlayer(0).GetCardsSize())
	assert.Equal(t, CardDesignHeart, s.GetTrumpSuit())
	assert.Equal(t, ShelemPhasePlay, s.GetPhase())
}

// **捨てるのはちょうど 4 枚。** 重複も範囲外も拒否する。
func TestShelem_DiscardValidation(t *testing.T) {
	s := newTestShelem(t)
	s.SetContractForTest(0, 120, false)
	s.CloseBiddingForTest()

	assert.Error(t, s.PlayerDiscard([]int{0, 1, 2}, CardDesignHeart), "3 枚は不足")
	assert.Error(t, s.PlayerDiscard([]int{0, 1, 2, 3, 4}, CardDesignHeart), "5 枚は多い")
	assert.Error(t, s.PlayerDiscard([]int{0, 1, 2, 2}, CardDesignHeart), "重複")
	assert.Error(t, s.PlayerDiscard([]int{0, 1, 2, 99}, CardDesignHeart), "範囲外")
	assert.Error(t, s.PlayerDiscard([]int{0, 1, 2, 3}, 0), "スートが不正")
	assert.Equal(t, ShelemPhaseDiscard, s.GetPhase(), "どれも通っていない")
}

// **後ろから消さないと残りのインデックスがずれる。** 大きい番号を含めて踏む。
func TestShelem_DiscardRemovesTheRequestedCards(t *testing.T) {
	s := newTestShelem(t)
	s.SetContractForTest(0, 120, false)
	s.CloseBiddingForTest()
	p := s.GetPlayer(0)
	kept := p.GetCard(4)

	require.NoError(t, s.PlayerDiscard([]int{0, 15, 7, 3}, CardDesignHeart))
	assert.Equal(t, ShelemHandSize, p.GetCardsSize())
	found := false
	for i := range p.GetCardsSize() {
		if p.GetCard(i) == kept {
			found = true
		}
	}
	assert.True(t, found, "捨てていない札が残っている")
}

func TestShelem_DiscardGuards(t *testing.T) {
	s := newTestShelem(t)
	assert.Error(t, s.PlayerDiscard([]int{0, 1, 2, 3}, CardDesignHeart), "競り中は捨てられない")

	s.SetContractForTest(2, 120, false)
	s.CloseBiddingForTest()
	s.SetPhaseForTest(ShelemPhaseDiscard)
	assert.Error(t, s.PlayerDiscard([]int{0, 1, 2, 3}, CardDesignHeart), "落札者でなければ捨てられない")
}

// --- 得点 ---

// **契約を取り切れば加点、届かなければ同額を失点する。**
func TestShelem_ContractMadeOrSet(t *testing.T) {
	t.Run("made", func(t *testing.T) {
		s := newTestShelem(t)
		s.SetContractForTest(0, 120, false)
		s.SetRoundPointsForTest(0, 120)
		s.SetRoundPointsForTest(1, 0)
		s.FinishRoundForTest()
		assert.Equal(t, 120, s.GetScore(0))
	})

	t.Run("set", func(t *testing.T) {
		s := newTestShelem(t)
		s.SetContractForTest(0, 120, false)
		s.SetRoundPointsForTest(0, 95)
		s.SetRoundPointsForTest(1, 5)
		s.FinishRoundForTest()
		assert.Equal(t, -120, s.GetScore(0))
		// **相手は取ったカード点をそのまま得る。**
		assert.Equal(t, 5, s.GetScore(1))
	})

	t.Run("exactly on the contract counts as made", func(t *testing.T) {
		s := newTestShelem(t)
		s.SetContractForTest(0, 100, false)
		s.SetRoundPointsForTest(0, 100)
		s.FinishRoundForTest()
		assert.Equal(t, 100, s.GetScore(0))
	})
}

// **Shelem は全トリック取れたかどうかだけ。** カード点は見ない。
func TestShelem_ShelemScoring(t *testing.T) {
	t.Run("all tricks taken", func(t *testing.T) {
		s := newTestShelem(t)
		s.SetContractForTest(0, ShelemMaxBid, true)
		s.GiveTricksForTest(0, ShelemTricksPerRound)
		s.FinishRoundForTest()
		assert.Equal(t, ShelemValue, s.GetScore(0))
	})

	t.Run("one trick lost", func(t *testing.T) {
		s := newTestShelem(t)
		s.SetContractForTest(0, ShelemMaxBid, true)
		s.GiveTricksForTest(0, ShelemTricksPerRound-1)
		s.GiveTricksForTest(1, 1)
		s.FinishRoundForTest()
		assert.Equal(t, -ShelemValue, s.GetScore(0))
	})

	// **カード点を全部取っていても、トリックを1つ落とせば失敗。**
	t.Run("card points do not rescue a failed shelem", func(t *testing.T) {
		s := newTestShelem(t)
		s.SetContractForTest(0, ShelemMaxBid, true)
		s.GiveTricksForTest(0, ShelemTricksPerRound-1)
		s.GiveTricksForTest(1, 1)
		s.SetRoundPointsForTest(0, ShelemHandPoints)
		s.FinishRoundForTest()
		assert.Equal(t, -ShelemValue, s.GetScore(0))
	})
}

// **Shelem は 1 トリック落とした時点で終わる。** 残りは打たせない。
func TestShelem_ShelemEndsTheRoundOnTheFirstLostTrick(t *testing.T) {
	s := newTestShelem(t)
	s.SetContractForTest(0, ShelemMaxBid, true)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPhaseForTest(ShelemPhasePlay)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetLeadPlayerIdxForTest(0)

	shelemHandOf(s, 0, NewCard(CardDesignSpade, 2, false), NewCard(CardDesignSpade, 3, false))
	shelemHandOf(s, 1, NewCard(CardDesignSpade, 13, false), NewCard(CardDesignSpade, 4, false))
	shelemHandOf(s, 2, NewCard(CardDesignSpade, 6, false), NewCard(CardDesignSpade, 7, false))
	shelemHandOf(s, 3, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignSpade, 9, false))

	for i := range ShelemPlayerCnt {
		require.NoError(t, s.PlayForTest(i, 0))
	}

	// 相手 (1) が取ったので、その場で失敗が確定する。
	assert.Equal(t, 1, s.GetTrickNumber(), "残りは打たない")
	assert.Equal(t, -ShelemValue, s.GetScore(0))
	assert.NotEqual(t, ShelemPhasePlay, s.GetPhase())
}

// 1 ラウンドのカード点は必ず 100 点が 2 チームに分配される。
func TestShelem_RoundPointsAlwaysSumToOneHundred(t *testing.T) {
	s := newTestShelem(t)
	settleBidding(t, s)
	if s.GetShelemBid() {
		t.Skip("Shelem 宣言のラウンドはカード点で採点しない")
	}

	for s.GetPhase() == ShelemPhasePlay {
		if s.IsHumanTurn() {
			valid := s.GetValidPlayIndices(0)
			require.NotEmpty(t, valid)
			require.NoError(t, s.PlayerPlay(valid[0]))
			continue
		}
		s.CpuPlay()
	}

	assert.Equal(t, ShelemTricksPerRound, s.GetTrickNumber())
	assert.Equal(t, ShelemHandPoints, s.GetRoundPoints(0)+s.GetRoundPoints(1))
}

// 規定点に届かなければ次ラウンドへ、届けば終局。
func TestShelem_NextRoundAndGameEnd(t *testing.T) {
	s := newTestShelem(t)
	s.SetConfig(ShelemConfig{Target: ShelemTargetMax})
	s.SetContractForTest(0, 120, false)
	s.SetRoundPointsForTest(0, 120)
	s.FinishRoundForTest()
	require.Equal(t, ShelemPhaseRoundEnd, s.GetPhase())

	s.NextRound()
	assert.Equal(t, 2, s.GetRoundNumber())
	assert.Equal(t, 1, s.GetDealerIdx(), "親は時計回りに動く")
	assert.Equal(t, ShelemPhaseBid, s.GetPhase())
	assert.Equal(t, 0, s.GetTrumpSuit(), "切り札はラウンドごとに決め直す")
	assert.Equal(t, ShelemWidowSize, s.GetWidowSize(), "ウィドウも配り直す")
	assert.Equal(t, -1, s.GetDeclarerIdx())
}

func TestShelem_GameEndsAtTarget(t *testing.T) {
	s := newTestShelem(t)
	s.SetConfig(ShelemConfig{Target: ShelemTargetMin})
	s.SetContractForTest(0, 100, false)
	s.SetRoundPointsForTest(0, 100)
	s.FinishRoundForTest()

	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, ShelemPhaseGameEnd, s.GetPhase())
	assert.Equal(t, 0, s.GetWinnerTeam())

	before := s.GetRoundNumber()
	s.NextRound()
	assert.Equal(t, before, s.GetRoundNumber())
}

func TestShelem_TieHasNoWinner(t *testing.T) {
	s := newTestShelem(t)
	s.SetScoreForTestUse(0, 300)
	s.SetScoreForTestUse(1, 300)
	s.FinishGameForTest()
	assert.Equal(t, -1, s.GetWinnerTeam())
}

// --- プレイ ---

func TestShelem_MustFollowSuit(t *testing.T) {
	s := newTestShelem(t)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPhaseForTest(ShelemPhasePlay)
	s.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}})
	s.SetCurrentPlayerIdxForTest(0)
	shelemHandOf(s, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignSpade, 7, false))

	require.Error(t, s.PlayerPlay(0))
	require.NoError(t, s.PlayerPlay(1))
}

func TestShelem_TrumpBeatsTheLeadSuit(t *testing.T) {
	s := newTestShelem(t)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPhaseForTest(ShelemPhasePlay)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetLeadPlayerIdxForTest(0)

	shelemHandOf(s, 0, NewCard(CardDesignSpade, 1, false))
	shelemHandOf(s, 1, NewCard(CardDesignSpade, 13, false))
	shelemHandOf(s, 2, NewCard(CardDesignHeart, 2, false))
	shelemHandOf(s, 3, NewCard(CardDesignSpade, 12, false))

	for i := range ShelemPlayerCnt {
		require.NoError(t, s.PlayForTest(i, 0))
	}
	assert.Equal(t, 1, s.GetPlayer(2).GetTrickCount(), "切り札の 2 が A に勝つ")
	// A は 10 点、他は 0 点。チーム0 のカード点になる。
	assert.Equal(t, 10, s.GetRoundPoints(ShelemTeamOf(2)))
}

func TestShelem_PlayRejectsInvalidIndex(t *testing.T) {
	s := newTestShelem(t)
	settleBidding(t, s)
	s.SetCurrentPlayerIdxForTest(0)

	assert.Error(t, s.PlayerPlay(-1))
	assert.Error(t, s.PlayerPlay(99))
}

func TestShelem_PlayGuards(t *testing.T) {
	s := newTestShelem(t)
	assert.Error(t, s.PlayerPlay(0), "競り中は出せない")

	settleBidding(t, s)
	s.SetCurrentPlayerIdxForTest(1)
	assert.Error(t, s.PlayerPlay(0))

	s.SetCurrentPlayerIdxForTest(0)
	s.GiveUp()
	assert.Error(t, s.PlayerPlay(0))
}

func TestShelem_ValidIndicesOutOfRange(t *testing.T) {
	s := newTestShelem(t)
	assert.Nil(t, s.GetValidPlayIndices(-1))
	assert.Nil(t, s.GetValidPlayIndices(ShelemPlayerCnt))
}

// **味方が勝っているなら点札を乗せる。** カード点がそのまま契約の達否になる。
func TestShelem_CpuFeedsWinningPartner(t *testing.T) {
	s := newTestShelem(t)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPhaseForTest(ShelemPhasePlay)
	s.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
	})
	s.SetCurrentPlayerIdxForTest(2)
	shelemHandOf(s, 2,
		NewCard(CardDesignSpade, 3, false),  // 0 点
		NewCard(CardDesignSpade, 10, false)) // 10 点

	assert.True(t, s.partnerIsWinning(2))
	assert.Equal(t, 1, s.CpuChoiceForTest(2), "点札を乗せる")
}

func TestShelem_CpuTakesTrickCheaply(t *testing.T) {
	s := newTestShelem(t)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPhaseForTest(ShelemPhasePlay)
	s.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 9, false)}})
	s.SetCurrentPlayerIdxForTest(2)
	shelemHandOf(s, 2,
		NewCard(CardDesignSpade, 13, false), // 勝てるが強い
		NewCard(CardDesignSpade, 11, false), // 勝てて一番弱い
		NewCard(CardDesignSpade, 3, false))  // 勝てない

	assert.Equal(t, 1, s.CpuChoiceForTest(2))
}

// --- ヒント ---

func TestShelem_HintDuringBidding(t *testing.T) {
	s := newTestShelem(t)
	s.SetBidPlayerIdxForTest(0)

	h := s.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex)
	assert.Contains(t, []string{"shelemBid", "shelemPass"}, h.Reason)
}

func TestShelem_HintDuringDiscard(t *testing.T) {
	s := newTestShelem(t)
	s.SetContractForTest(0, 120, false)
	s.CloseBiddingForTest()

	h := s.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "shelemDiscard", h.Reason)
	assert.GreaterOrEqual(t, h.Suit, CardDesignSpade)
}

func TestShelem_HintDuringPlay(t *testing.T) {
	s := newTestShelem(t)
	settleBidding(t, s)
	s.SetCurrentPlayerIdxForTest(0)

	h := s.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Contains(t, s.GetValidPlayIndices(0), *h.CardIndex)
}

func TestShelem_HintNilWhenNotHumanTurn(t *testing.T) {
	s := newTestShelem(t)
	settleBidding(t, s)
	s.SetCurrentPlayerIdxForTest(2)
	assert.Nil(t, s.GetHint())

	s.GiveUp()
	assert.Nil(t, s.GetHint())
}

// --- その他 ---

func TestShelem_GiveUp(t *testing.T) {
	s := newTestShelem(t)
	s.GiveUp()
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 1, s.GetWinnerTeam())

	s.SetWinnerTeamForTest(0)
	s.GiveUp()
	assert.Equal(t, 0, s.GetWinnerTeam(), "二度目は何も起きない")
}

func TestShelem_AccessorsOutOfRange(t *testing.T) {
	s := newTestShelem(t)
	assert.Nil(t, s.GetPlayer(-1))
	assert.Nil(t, s.GetPlayer(ShelemPlayerCnt))
	assert.Equal(t, ShelemPlayerCnt, s.GetPlayerCnt())
	assert.Equal(t, 0, s.GetScore(-1))
	assert.Equal(t, 0, s.GetScore(ShelemTeamCnt))
	assert.Equal(t, 0, s.GetRoundPoints(-1))
	assert.Equal(t, 0, s.TeamTricks(-1))
	assert.Equal(t, 0, s.TeamTricks(ShelemTeamCnt))
	assert.Equal(t, 1, s.GetRoundNumber())
	assert.Empty(t, s.GetCurrentTrick())
}

func TestShelem_TeamAssignment(t *testing.T) {
	assert.Equal(t, ShelemTeamOf(0), ShelemTeamOf(2))
	assert.Equal(t, ShelemTeamOf(1), ShelemTeamOf(3))
	assert.NotEqual(t, ShelemTeamOf(0), ShelemTeamOf(1))
}

func TestShelemConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultShelemConfig().Validate())
	assert.NoError(t, ShelemConfig{Target: ShelemTargetMin}.Validate())
	assert.NoError(t, ShelemConfig{Target: ShelemTargetMax}.Validate())
	assert.Error(t, ShelemConfig{Target: ShelemTargetMin - 1}.Validate())
	assert.Error(t, ShelemConfig{Target: ShelemTargetMax + 1}.Validate())
}

// --- JSON 往復 ---

func TestShelem_JSONRoundTrip(t *testing.T) {
	s := newTestShelem(t)
	s.SetContractForTest(0, 130, false)
	s.CloseBiddingForTest()
	require.NoError(t, s.PlayerDiscard([]int{0, 1, 2, 3}, CardDesignDiamond))
	s.SetScoreForTestUse(0, 240)

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var got Shelem
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, s.GetPhase(), got.GetPhase())
	assert.Equal(t, CardDesignDiamond, got.GetTrumpSuit())
	assert.Equal(t, 130, got.GetContract())
	assert.Equal(t, 0, got.GetDeclarerIdx())
	assert.False(t, got.GetShelemBid())
	assert.Equal(t, 240, got.GetScore(0))
	assert.Equal(t, 0, got.GetWidowSize(), "ウィドウは落札者に渡っている")
	for i := range ShelemPlayerCnt {
		assert.Equal(t, s.GetPlayer(i).GetCardsSize(), got.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
}

func TestShelem_UnmarshalRejectsInvalid(t *testing.T) {
	valid := func() shelemJSON {
		return shelemJSON{
			Config:      DefaultShelemConfig(),
			Phase:       ShelemPhasePlay,
			RoundNumber: 1,
			TrumpSuit:   CardDesignSpade,
			Contract:    120,
			DeclarerIdx: 0,
			WinnerTeam:  -1,
		}
	}
	cases := map[string]func(*shelemJSON){
		"bad config":   func(j *shelemJSON) { j.Config.Target = 0 },
		"bad phase":    func(j *shelemJSON) { j.Phase = ShelemPhase(99) },
		"bad trick":    func(j *shelemJSON) { j.TrickNumber = ShelemTricksPerRound + 1 },
		"bad round":    func(j *shelemJSON) { j.RoundNumber = 0 },
		"bad contract": func(j *shelemJSON) { j.Contract = ShelemMaxBid + 1 },
		"low contract": func(j *shelemJSON) { j.Contract = ShelemMinBid - 1 },
		"bad current":  func(j *shelemJSON) { j.CurrentPlayerIdx = ShelemPlayerCnt },
		"bad bidder":   func(j *shelemJSON) { j.BidPlayerIdx = -1 },
		"bad lead":     func(j *shelemJSON) { j.LeadPlayerIdx = -1 },
		"bad dealer":   func(j *shelemJSON) { j.DealerIdx = ShelemPlayerCnt },
		"bad declarer": func(j *shelemJSON) { j.DeclarerIdx = ShelemPlayerCnt },
		"bad winner":   func(j *shelemJSON) { j.WinnerTeam = ShelemTeamCnt },
		"big widow":    func(j *shelemJSON) { j.Widow = make([]*Card, ShelemWidowSize+1) },
		// **切り札はフェーズと整合していなければならない。** 両方向を踏む。
		"trump before it was chosen": func(j *shelemJSON) {
			j.Phase, j.TrumpSuit = ShelemPhaseBid, CardDesignHeart
		},
		"no trump after it was chosen": func(j *shelemJSON) {
			j.Phase, j.TrumpSuit = ShelemPhasePlay, 0
		},
		"bogus trump": func(j *shelemJSON) { j.TrumpSuit = 99 },
		"long trick": func(j *shelemJSON) {
			j.CurrentTrick = make([]*TrickCard, ShelemPlayerCnt+1)
		},
		"long log": func(j *shelemJSON) {
			j.ActionLog = make([]*ActionLogEntry, shelemMaxSliceLen+1)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			j := valid()
			mutate(&j)
			data, err := json.Marshal(j)
			require.NoError(t, err)
			var got Shelem
			assert.Error(t, json.Unmarshal(data, &got))
		})
	}

	var got Shelem
	assert.Error(t, got.UnmarshalJSON([]byte("{")))

	// 正のコントロール: 正しいスナップショットは通る。
	data, err := json.Marshal(valid())
	require.NoError(t, err)
	assert.NoError(t, json.Unmarshal(data, &got))

	// 競り中で切り札 0・契約 0 も通る。
	j := valid()
	j.Phase, j.TrumpSuit, j.Contract = ShelemPhaseBid, 0, 0
	pre, err := json.Marshal(j)
	require.NoError(t, err)
	var okPre Shelem
	assert.NoError(t, json.Unmarshal(pre, &okPre))
}

func TestShelem_ActionLog(t *testing.T) {
	s := newTestShelem(t)
	require.NotEmpty(t, s.actionLog, "配りが記録される")
	s.SetBidPlayerIdxForTest(0)
	require.NoError(t, s.PlayerBid(ShelemMinBid))

	kinds := map[string]bool{}
	for _, e := range s.actionLog {
		kinds[e.ActionType] = true
	}
	assert.True(t, kinds["bid"])
}
