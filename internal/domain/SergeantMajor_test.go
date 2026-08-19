//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSergeantMajor(t *testing.T) *SergeantMajor {
	t.Helper()
	s := NewDefaultSergeantMajor()
	s.Reset()
	return s
}

// **issue の「52 枚を 3 人に均等に配り」は成立しない。** 52 は 3 で割り切れず、
// 実際は 16 枚ずつ = 48 枚 + 余り 4 枚（キティ）を親が取り込む。
func TestSergeantMajor_DealIsSixteenEachPlusAKitty(t *testing.T) {
	assert.Equal(t, 4, SergeantMajorKittySize, "52 - 3×16 = 4")
	assert.Equal(t, 52, SergeantMajorPlayerCnt*SergeantMajorHandSize+SergeantMajorKittySize,
		"3 人 × 16 枚 + キティ 4 枚 = 52 枚ちょうど")

	s := newTestSergeantMajor(t)
	total := s.GetKittySize()
	for i := range SergeantMajorPlayerCnt {
		assert.Equal(t, SergeantMajorHandSize, s.GetPlayer(i).GetCardsSize())
		total += s.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total, "52 枚すべてが行き先を持つ")
	assert.Equal(t, SergeantMajorKittySize, s.GetKittySize())
}

// **8 + 5 + 3 = 16 でトリックが余りも不足もしない。**
func TestSergeantMajor_TargetsSumToTheTrickCount(t *testing.T) {
	sum := 0
	for _, v := range SergeantMajorTargets {
		sum += v
	}
	assert.Equal(t, SergeantMajorTricksPerRound, sum)
	assert.Equal(t, SergeantMajorHandSize, SergeantMajorTricksPerRound)
	assert.Equal(t, 16, SergeantMajorTricksPerRound)
}

func TestSergeantMajor_ResetWaitsForTrump(t *testing.T) {
	s := newTestSergeantMajor(t)
	assert.Equal(t, SergeantMajorPhaseTrump, s.GetPhase())
	assert.Equal(t, 0, s.GetTrumpSuit())
	assert.Equal(t, 1, s.GetRoundNumber())
	assert.Empty(t, s.GetValidPlayIndices(0), "宣言前は出せない")
}

// **ノルマは席順で決まる。** 親が 8、左隣が 5、右隣が 3。
func TestSergeantMajor_TargetsFollowTheSeats(t *testing.T) {
	s := newTestSergeantMajor(t)
	for dealer := range SergeantMajorPlayerCnt {
		s.SetDealerIdxForTest(dealer)
		s.SetPhaseForTest(SergeantMajorPhaseRoundEnd)
		s.NextRound()
		d := s.GetDealerIdx()
		assert.Equal(t, 8, s.GetPlayer(d).GetTarget(), "親はノルマ 8")
		assert.Equal(t, 5, s.GetPlayer((d+1)%SergeantMajorPlayerCnt).GetTarget(), "左隣は 5")
		assert.Equal(t, 3, s.GetPlayer((d+2)%SergeantMajorPlayerCnt).GetTarget(), "右隣は 3")
	}
}

// **3 ラウンドで全員が 8・5・3 を一度ずつ引き受ける。**
func TestSergeantMajor_TheDealerRoleRotates(t *testing.T) {
	s := NewDefaultSergeantMajor()
	s.SetConfig(SergeantMajorConfig{Rounds: SergeantMajorRoundsMax})
	s.Reset()

	seen := map[int]map[int]bool{0: {}, 1: {}, 2: {}}
	for range SergeantMajorPlayerCnt {
		for i := range SergeantMajorPlayerCnt {
			seen[i][s.GetPlayer(i).GetTarget()] = true
		}
		s.SetPhaseForTest(SergeantMajorPhaseRoundEnd)
		s.NextRound()
	}
	for i := range SergeantMajorPlayerCnt {
		assert.Len(t, seen[i], SergeantMajorPlayerCnt, "席 %d が 3 種類すべてを経験する", i)
	}
}

// **親はキティを取り込んで 4 枚捨てる。**
func TestSergeantMajor_TheDealerTakesTheKittyAndDiscards(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetDealerIdxForTest(0)
	require.NoError(t, s.DeclareTrump(CardDesignHeart))

	assert.Equal(t, SergeantMajorPhaseDiscard, s.GetPhase())
	assert.Equal(t, 0, s.GetKittySize(), "キティは親の手に入る")
	assert.Equal(t, SergeantMajorHandSize+SergeantMajorKittySize, s.GetPlayer(0).GetCardsSize(), "20 枚になる")

	require.NoError(t, s.DiscardForTest(0, []int{0, 1, 2, 3}))
	assert.Equal(t, SergeantMajorPhasePlay, s.GetPhase())
	assert.Equal(t, SergeantMajorHandSize, s.GetPlayer(0).GetCardsSize(), "16 枚に戻る")
	// **リードは親の左隣。**
	assert.Equal(t, 1, s.GetLeadPlayerIdx())
}

// **捨て札はちょうど 4 枚、重複と範囲外を弾く。**
func TestSergeantMajor_DiscardRejectsBadInput(t *testing.T) {
	for _, tc := range []struct {
		name    string
		indices []int
	}{
		{"too few", []int{0, 1, 2}},
		{"too many", []int{0, 1, 2, 3, 4}},
		{"duplicate", []int{0, 1, 1, 2}},
		{"negative", []int{-1, 1, 2, 3}},
		{"out of range", []int{0, 1, 2, 999}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestSergeantMajor(t)
			s.SetDealerIdxForTest(0)
			require.NoError(t, s.DeclareTrump(CardDesignHeart))
			assert.Error(t, s.DiscardForTest(0, tc.indices))
			assert.Equal(t, SergeantMajorPhaseDiscard, s.GetPhase(), "フェーズは進まない")
		})
	}

	s := newTestSergeantMajor(t)
	s.SetDealerIdxForTest(0)
	require.NoError(t, s.DeclareTrump(CardDesignHeart))
	assert.Error(t, s.DiscardForTest(1, []int{0, 1, 2, 3}), "親以外は捨てられない")
}

// **後ろから消す。** 前から消すと残りのインデックスがずれて別の札が消える。
func TestSergeantMajor_DiscardRemovesTheRequestedCards(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetDealerIdxForTest(0)
	s.SetPhaseForTest(SergeantMajorPhaseDiscard)
	sergeantMajorHandOf(s, 0,
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 6, false))

	require.NoError(t, s.DiscardForTest(0, []int{0, 3, 1, 2}))

	require.Equal(t, 1, s.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 6, s.GetPlayer(0).GetCard(0).GetValue(), "触れていない札だけが残る")
}

func TestSergeantMajor_DeclareTrumpRejectsBadInput(t *testing.T) {
	for _, suit := range []int{0, -1, 5, 99} {
		s := newTestSergeantMajor(t)
		assert.Error(t, s.DeclareTrump(suit))
		assert.Equal(t, SergeantMajorPhaseTrump, s.GetPhase())
	}

	s := newTestSergeantMajor(t)
	require.NoError(t, s.DeclareTrump(CardDesignSpade))
	assert.Error(t, s.DeclareTrump(CardDesignHeart), "二度は宣言できない")

	s.FinishGameForTest()
	assert.Error(t, s.DeclareTrump(CardDesignHeart), "終局後は宣言できない")
}

// **宣言できるのは親だけ。**
func TestSergeantMajor_OnlyTheDealerDeclares(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetDealerIdxForTest(1)
	assert.False(t, s.IsHumanTrumpTurn())
	assert.Error(t, s.PlayerDeclareTrump(CardDesignHeart))

	s.SetDealerIdxForTest(0)
	assert.True(t, s.IsHumanTrumpTurn())
	assert.NoError(t, s.PlayerDeclareTrump(CardDesignHeart))
}

func TestSergeantMajor_CpuDeclaresAndDiscards(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetDealerIdxForTest(1)

	s.CpuDeclareTrump()
	assert.Positive(t, s.GetTrumpSuit())
	assert.Equal(t, SergeantMajorPhaseDiscard, s.GetPhase())

	s.CpuDiscard()
	assert.Equal(t, SergeantMajorPhasePlay, s.GetPhase())
	assert.Equal(t, SergeantMajorHandSize, s.GetPlayer(1).GetCardsSize())
}

// **CPU は切り札を捨てにくくする。**
func TestSergeantMajor_CpuDiscardKeepsTrumps(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetDealerIdxForTest(0)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPhaseForTest(SergeantMajorPhaseDiscard)
	sergeantMajorHandOf(s, 0,
		NewCard(CardDesignHeart, 2, false), NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignSpade, 2, false), NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false), NewCard(CardDesignSpade, 5, false))

	for _, i := range s.CpuDiscardChoiceForTest(0) {
		assert.NotEqual(t, CardDesignHeart, s.GetPlayer(0).GetCard(i).GetDesign(),
			"切り札は捨てない")
	}
}

func TestSergeantMajor_FollowSuitIsCompulsory(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetTrumpSuitForTest(CardDesignDiamond)
	s.SetPhaseForTest(SergeantMajorPhasePlay)
	s.SetLeadPlayerIdxForTest(0)
	s.SetCurrentPlayerIdxForTest(0)
	sergeantMajorHandOf(s, 0, NewCard(CardDesignSpade, 8, false))
	sergeantMajorHandOf(s, 1, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 8, false))
	sergeantMajorHandOf(s, 2, NewCard(CardDesignSpade, 10, false))

	require.NoError(t, s.PlayForTest(0, 0))
	assert.Equal(t, []int{0}, s.GetValidPlayIndices(1))
	assert.Error(t, s.PlayForTest(1, 1))
}

// **切り札 > リードスート > その他。** A が最強。
func TestSergeantMajor_TrickWinnerOrdering(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetTrumpSuitForTest(CardDesignDiamond)

	s.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 13, false)},
	})
	assert.Equal(t, 1, s.TrickWinnerForTest(), "切り札の 2 が ♠A に勝つ")

	s.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 10, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 1, false)},
	})
	assert.Equal(t, 2, s.TrickWinnerForTest(), "別スートの A は取れない")

	s.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignDiamond, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 13, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 1, false)},
	})
	assert.Equal(t, 2, s.TrickWinnerForTest(), "切り札同士は A が最強")
}

// **1 ラウンドはちょうど 16 トリック。**
func TestSergeantMajor_ARoundIsExactlySixteenTricks(t *testing.T) {
	s := newTestSergeantMajor(t)
	playOutSergeantMajorSetup(t, s)
	for s.GetPhase() == SergeantMajorPhasePlay {
		idx := s.GetCurrentPlayerIdx()
		require.NoError(t, s.PlayForTest(idx, s.CpuChoiceForTest(idx)))
	}
	assert.Equal(t, SergeantMajorTricksPerRound, s.GetTrickNumber())
	total := 0
	for i := range SergeantMajorPlayerCnt {
		total += s.GetPlayer(i).GetTrickCount()
		assert.Zero(t, s.GetPlayer(i).GetCardsSize(), "手札を打ち切る")
	}
	assert.Equal(t, SergeantMajorTricksPerRound, total)
}

// playOutSergeantMajorSetup は宣言と捨て札を CPU に任せてプレイまで進める。
func playOutSergeantMajorSetup(t *testing.T, s *SergeantMajor) {
	t.Helper()
	if s.GetPhase() == SergeantMajorPhaseTrump {
		require.NoError(t, s.DeclareTrump(s.CpuTrumpChoiceForTest(s.GetDealerIdx())))
	}
	if s.GetPhase() == SergeantMajorPhaseDiscard {
		d := s.GetDealerIdx()
		require.NoError(t, s.DiscardForTest(d, s.CpuDiscardChoiceForTest(d)))
	}
}

// **ノルマとの差がそのまま得点。** 過不足は必ず打ち消し合う。
func TestSergeantMajor_ScoringIsTheDifferenceFromTheTarget(t *testing.T) {
	for _, tc := range []struct {
		name        string
		tricks      [SergeantMajorPlayerCnt]int
		wantScores  [SergeantMajorPlayerCnt]int
		wantSurplus [SergeantMajorPlayerCnt]int
	}{
		{"全員ちょうど", [3]int{8, 5, 3}, [3]int{0, 0, 0}, [3]int{0, 0, 0}},
		{"親が超過", [3]int{10, 4, 2}, [3]int{2, -1, -1}, [3]int{2, -1, -1}},
		{"親が不足", [3]int{5, 8, 3}, [3]int{-3, 3, 0}, [3]int{-3, 3, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestSergeantMajor(t)
			s.SetDealerIdxForTest(0)
			playOutSergeantMajorSetup(t, s)
			for i, n := range tc.tricks {
				s.GiveTricksForTest(i, n)
			}
			s.FinishRoundForTest()

			for i := range SergeantMajorPlayerCnt {
				assert.Equal(t, tc.wantScores[i], s.GetPlayer(i).GetScore(), "席 %d の得点", i)
				assert.Equal(t, tc.wantSurplus[i], s.GetSurplusForTest()[i], "席 %d の過不足", i)
			}
			sum := 0
			for _, v := range s.GetSurplusForTest() {
				sum += v
			}
			assert.Zero(t, sum, "過不足は打ち消し合う")
		})
	}
}

// **不足した人が良い札を差し出す。** 枚数は変わらない。
func TestSergeantMajor_ExchangeMovesTheBestCardUp(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetPhaseForTest(SergeantMajorPhasePlay)
	sergeantMajorHandOf(s, 0, NewCard(CardDesignClover, 2, false))
	sergeantMajorHandOf(s, 1, NewCard(CardDesignHeart, 1, false)) // ♥A が最強
	sergeantMajorHandOf(s, 2, NewCard(CardDesignSpade, 9, false))
	s.SetSurplusForTest([]int{1, -1, 0})

	s.ExchangeForTest()

	assert.Equal(t, 14, sergeantMajorRank(s.GetPlayer(0).GetCard(0)), "超過した席に ♥A が移る")
	assert.Equal(t, 2, sergeantMajorRank(s.GetPlayer(1).GetCard(0)), "代わりに最弱札が戻る")
	assert.Equal(t, 1, s.GetLastExchange())
	assert.Equal(t, []int{0, 0, 0}, s.GetSurplusForTest(), "過不足は使い切る")
}

// **超過も不足も複数人いることがある。**
func TestSergeantMajor_ExchangeHandlesManyToMany(t *testing.T) {
	for _, tc := range []struct {
		name    string
		surplus []int
		moved   int
	}{
		{"1 人超過 2 人不足", []int{2, -1, -1}, 2},
		{"2 人超過 1 人不足", []int{1, -2, 1}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestSergeantMajor(t)
			s.SetPhaseForTest(SergeantMajorPhasePlay)
			for i := range SergeantMajorPlayerCnt {
				sergeantMajorHandOf(s, i,
					NewCard(CardDesignSpade, 2+i, false),
					NewCard(CardDesignHeart, 10+i, false),
					NewCard(CardDesignClover, 1, false))
			}
			s.SetSurplusForTest(append([]int(nil), tc.surplus...))

			s.ExchangeForTest()

			assert.Equal(t, tc.moved, s.GetLastExchange())
			for i := range SergeantMajorPlayerCnt {
				assert.Equal(t, 3, s.GetPlayer(i).GetCardsSize(), "席 %d の枚数は変わらない", i)
			}
			assert.Equal(t, []int{0, 0, 0}, s.GetSurplusForTest())
		})
	}
}

// **負のコントロール: 過不足が無ければ 1 枚も動かない。**
func TestSergeantMajor_NoSurplusMovesNothing(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetPhaseForTest(SergeantMajorPhasePlay)
	sergeantMajorHandOf(s, 0, NewCard(CardDesignClover, 2, false))
	sergeantMajorHandOf(s, 1, NewCard(CardDesignHeart, 1, false))
	sergeantMajorHandOf(s, 2, NewCard(CardDesignSpade, 9, false))
	s.SetSurplusForTest([]int{0, 0, 0})

	s.ExchangeForTest()
	assert.Equal(t, 0, s.GetLastExchange())
	assert.Equal(t, 14, sergeantMajorRank(s.GetPlayer(1).GetCard(0)), "♥A は動かない")
}

func TestSergeantMajor_MostPointsWins(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.GetPlayer(0).SetScore(4)
	s.GetPlayer(1).SetScore(-2)
	s.GetPlayer(2).SetScore(-2)
	s.FinishGameForTest()

	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 0, s.GetWinnerIdx())
}

func TestSergeantMajor_TieHasNoWinner(t *testing.T) {
	s := newTestSergeantMajor(t)
	for i := range SergeantMajorPlayerCnt {
		s.GetPlayer(i).SetScore(0)
	}
	s.FinishGameForTest()
	assert.Equal(t, -1, s.GetWinnerIdx())
}

func TestSergeantMajor_GameEndsAtTheConfiguredRound(t *testing.T) {
	s := NewDefaultSergeantMajor()
	s.SetConfig(SergeantMajorConfig{Rounds: SergeantMajorRoundsMin})
	s.Reset()

	for round := 1; round <= SergeantMajorRoundsMin; round++ {
		require.Equal(t, round, s.GetRoundNumber())
		playOutSergeantMajorSetup(t, s)
		for s.GetPhase() == SergeantMajorPhasePlay {
			idx := s.GetCurrentPlayerIdx()
			require.NoError(t, s.PlayForTest(idx, s.CpuChoiceForTest(idx)))
		}
		if round < SergeantMajorRoundsMin {
			require.False(t, s.GetGameEndFlag())
			s.NextRound()
		}
	}
	assert.True(t, s.GetGameEndFlag())
}

// **CPU は必ず合法手を返す。**
func TestSergeantMajor_CpuAlwaysChoosesLegally(t *testing.T) {
	for range 50 {
		s := NewDefaultSergeantMajor()
		s.SetConfig(SergeantMajorConfig{Rounds: SergeantMajorRoundsMin})
		s.Reset()
		for !s.GetGameEndFlag() {
			switch s.GetPhase() {
			case SergeantMajorPhaseTrump:
				require.NoError(t, s.DeclareTrump(s.CpuTrumpChoiceForTest(s.GetDealerIdx())))
			case SergeantMajorPhaseDiscard:
				d := s.GetDealerIdx()
				require.NoError(t, s.DiscardForTest(d, s.CpuDiscardChoiceForTest(d)))
			case SergeantMajorPhasePlay:
				idx := s.GetCurrentPlayerIdx()
				choice := s.CpuChoiceForTest(idx)
				require.Contains(t, s.GetValidPlayIndices(idx), choice)
				require.NoError(t, s.PlayForTest(idx, choice))
			case SergeantMajorPhaseRoundEnd:
				s.NextRound()
			default:
			}
		}
	}
}

func TestSergeantMajor_RejectsOutOfTurnAndBadIndices(t *testing.T) {
	s := newTestSergeantMajor(t)
	assert.Error(t, s.PlayForTest(0, 0), "宣言前は打てない")

	playOutSergeantMajorSetup(t, s)
	idx := s.GetCurrentPlayerIdx()
	assert.Error(t, s.PlayForTest((idx+1)%SergeantMajorPlayerCnt, 0), "手番でない席は打てない")
	assert.Error(t, s.PlayForTest(idx, -1))
	assert.Error(t, s.PlayForTest(idx, 999))

	s.FinishGameForTest()
	assert.Error(t, s.PlayForTest(idx, 0), "終局後は打てない")
}

// **公開の入口も踏む。** `Player*` / `Cpu*` のガードが未検証のまま残らないように。
func TestSergeantMajor_PublicEntryPointsGuardTheTurn(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetDealerIdxForTest(0)
	require.NoError(t, s.PlayerDeclareTrump(CardDesignHeart))

	assert.True(t, s.IsHumanDiscardTurn())
	before := s.GetPlayer(0).GetCardsSize()
	s.CpuDiscard()
	assert.Equal(t, before, s.GetPlayer(0).GetCardsSize(), "人間の捨て札を勝手にやらない")
	require.NoError(t, s.PlayerDiscard([]int{0, 1, 2, 3}))

	s.SetCurrentPlayerIdxForTest(1)
	assert.False(t, s.IsHumanTurn())
	assert.Error(t, s.PlayerPlay(0))
	cpuBefore := s.GetPlayer(1).GetCardsSize()
	s.CpuPlay()
	assert.Equal(t, cpuBefore-1, s.GetPlayer(1).GetCardsSize())

	s.SetCurrentPlayerIdxForTest(0)
	humanBefore := s.GetPlayer(0).GetCardsSize()
	s.CpuPlay()
	assert.Equal(t, humanBefore, s.GetPlayer(0).GetCardsSize(), "人間の番では CPU は動かない")
	require.NoError(t, s.PlayerPlay(s.GetValidPlayIndices(0)[0]))
}

func TestSergeantMajor_NextRoundOnlyFromRoundEnd(t *testing.T) {
	s := newTestSergeantMajor(t)
	before := s.GetRoundNumber()
	s.NextRound()
	assert.Equal(t, before, s.GetRoundNumber(), "宣言フェーズからは進まない")

	s.SetPhaseForTest(SergeantMajorPhaseRoundEnd)
	s.NextRound()
	assert.Equal(t, before+1, s.GetRoundNumber())
	assert.Equal(t, SergeantMajorKittySize, s.GetKittySize(), "配り直すのでキティが戻る")

	s.SetPhaseForTest(SergeantMajorPhaseRoundEnd)
	s.FinishGameForTest()
	after := s.GetRoundNumber()
	s.NextRound()
	assert.Equal(t, after, s.GetRoundNumber(), "終局後は進まない")
}

func TestSergeantMajor_GiveUp(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.GiveUp()
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, -1, s.GetWinnerIdx())
	s.GiveUp()
	assert.Equal(t, -1, s.GetWinnerIdx())
}

func TestSergeantMajor_Hint(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.SetDealerIdxForTest(0)

	h := s.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex)
	assert.Equal(t, "sergeantmajorSelectTrump", h.Reason)
	assert.GreaterOrEqual(t, h.Suit, CardDesignSpade)

	require.NoError(t, s.DeclareTrump(CardDesignHeart))
	h = s.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "sergeantmajorDiscard", h.Reason)
	assert.Len(t, h.Indices, SergeantMajorKittySize, "捨てる枚数ぶん勧める")

	require.NoError(t, s.PlayerDiscard(h.Indices))
	s.SetCurrentPlayerIdxForTest(0)
	h = s.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Contains(t, s.GetValidPlayIndices(0), *h.CardIndex, "勧める札は必ず合法")

	s.FinishGameForTest()
	assert.Nil(t, s.GetHint(), "終局後は助言しない")
}

func TestSergeantMajor_JSONRoundTrip(t *testing.T) {
	s := newTestSergeantMajor(t)
	playOutSergeantMajorSetup(t, s)
	for range 5 {
		idx := s.GetCurrentPlayerIdx()
		require.NoError(t, s.PlayForTest(idx, s.CpuChoiceForTest(idx)))
	}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var restored SergeantMajor
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, s.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, s.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, s.GetDealerIdx(), restored.GetDealerIdx())
	for i := range SergeantMajorPlayerCnt {
		assert.Equal(t, s.GetPlayer(i).GetTarget(), restored.GetPlayer(i).GetTarget(), "ノルマが消えない")
		assert.Equal(t, s.GetPlayer(i).GetScore(), restored.GetPlayer(i).GetScore(), "得点が消えない")
		assert.Equal(t, s.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
	}
}

// **壊れたスナップショットは弾く。**
func TestSergeantMajor_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	base := func(t *testing.T) map[string]any {
		t.Helper()
		s := newTestSergeantMajor(t)
		playOutSergeantMajorSetup(t, s)
		require.NoError(t, s.PlayForTest(s.GetCurrentPlayerIdx(), 0))
		data, err := json.Marshal(s)
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(data, &m))
		return m
	}

	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"phase out of range", func(m map[string]any) { m["ph"] = 9 }},
		{"trump suit out of range", func(m map[string]any) { m["ts"] = 9 }},
		{"no trump suit during play", func(m map[string]any) { m["ts"] = 0 }},
		{"a kitty that survived the deal", func(m map[string]any) {
			m["ki"] = []any{map[string]any{"d": 1, "v": 9, "j": false}}
		}},
		{"current player out of range", func(m map[string]any) { m["ci"] = 9 }},
		{"dealer out of range", func(m map[string]any) { m["dl"] = -1 }},
		{"winner before the game ended", func(m map[string]any) { m["wi"] = 1 }},
		{"round number below one", func(m map[string]any) { m["rn"] = 0 }},
		{"negative trick number", func(m map[string]any) { m["tn"] = -1 }},
		{"surplus does not cancel out", func(m map[string]any) { m["sp"] = []any{2, 0, 0} }},
		{"surplus of the wrong length", func(m map[string]any) { m["sp"] = []any{0, 0} }},
		// **場札は枚数だけでなく中身も見る（#5310 で踏んだ panic の再発防止）。**
		{"a trick entry with no card", func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 0}}
		}},
		{"a trick entry with a bad seat", func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 9, "card": map[string]any{"d": 1, "v": 9, "j": false}}}
		}},
		{"config out of range", func(m map[string]any) { m["cf"] = map[string]any{"r": 0} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base(t)
			tc.mutate(m)
			data, err := json.Marshal(m)
			require.NoError(t, err)
			var restored SergeantMajor
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通り、使っても落ちない。**
	data, err := json.Marshal(base(t))
	require.NoError(t, err)
	var ok SergeantMajor
	require.NoError(t, json.Unmarshal(data, &ok))
	assert.NotPanics(t, func() {
		_ = ok.GetValidPlayIndices(ok.GetCurrentPlayerIdx())
		_ = ok.TrickWinnerForTest()
	})
}

// **ノルマは 8/5/3 のいずれか。**
func TestSergeantMajorPlayer_UnmarshalRejectsBrokenTargets(t *testing.T) {
	for _, body := range []string{`{"tg":4}`, `{"tg":-1}`, `{"tg":99}`} {
		var p SergeantMajorPlayer
		assert.Error(t, json.Unmarshal([]byte(body), &p), body)
	}
	for _, body := range []string{`{"tg":0}`, `{"tg":8}`, `{"tg":5}`, `{"tg":3}`} {
		var p SergeantMajorPlayer
		assert.NoError(t, json.Unmarshal([]byte(body), &p), body)
	}
}

func TestSergeantMajorConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultSergeantMajorConfig().Validate())
	assert.NoError(t, SergeantMajorConfig{Rounds: SergeantMajorRoundsMin}.Validate())
	assert.NoError(t, SergeantMajorConfig{Rounds: SergeantMajorRoundsMax}.Validate())
	assert.Error(t, SergeantMajorConfig{Rounds: SergeantMajorRoundsMin - 1}.Validate())
	assert.Error(t, SergeantMajorConfig{Rounds: SergeantMajorRoundsMax + 1}.Validate())
}

// **3 の倍数でないラウンド数は不公平になる（レビュー指摘 PR #5311）。**
// 親は毎ラウンド 1 つ回るだけなので、4 ラウンドだと 1 人だけノルマ 8 を
// 2 回引き受ける——届きにくいノルマを多く負う側がそのまま不利になる。
func TestSergeantMajorConfig_RoundsMustCoverEverySeat(t *testing.T) {
	for _, n := range []int{4, 5, 7, 8, 29} {
		assert.Error(t, SergeantMajorConfig{Rounds: n}.Validate(), "rounds=%d", n)
	}
	// **負のコントロール: 3 の倍数は通る。**
	for _, n := range []int{3, 6, 9, 30} {
		assert.NoError(t, SergeantMajorConfig{Rounds: n}.Validate(), "rounds=%d", n)
	}

	// **偏りが実際に起きることを踏む。** 4 ラウンドなら 1 人が 8 を 2 回。
	counts := map[int]int{}
	for round := range 4 {
		dealer := round % SergeantMajorPlayerCnt
		counts[dealer]++
	}
	assert.Equal(t, 2, counts[0], "4 ラウンドだと席 0 だけが親を 2 回引く")
}

func TestSergeantMajor_AccessorsAndBounds(t *testing.T) {
	s := newTestSergeantMajor(t)
	assert.Nil(t, s.GetPlayer(-1))
	assert.Nil(t, s.GetPlayer(99))
	assert.Empty(t, s.GetValidPlayIndices(-1))
	assert.Equal(t, SergeantMajorPlayerCnt, s.GetPlayerCnt())
	assert.Equal(t, SergeantMajorKittySize, s.GetDiscardCount())
	assert.Equal(t, SergeantMajorDefaultRounds, s.GetConfig().Rounds)
	assert.NotEmpty(t, s.GetActionLog())
	assert.Empty(t, s.GetCurrentTrick())
	assert.Zero(t, s.GetLastExchange())
	assert.Equal(t, s.GetDealerIdx(), s.GetLeadPlayerIdx())
}

// **取り込むと手札に紛れて見分けが付かなくなる** (#5759)。捨てる 4 枚を選ぶ
// あいだだけ、どれがキティ由来かを覚えておく。
func TestSergeantMajorRemembersTheAbsorbedKitty(t *testing.T) {
	s := newTestSergeantMajor(t)
	s.Reset()
	s.SetPhaseForTest(SergeantMajorPhaseTrump)
	s.SetDealerIdxForTest(0)

	kitty := append([]*Card(nil), s.GetKittyForTest()...)
	if len(kitty) != SergeantMajorKittySize {
		t.Fatalf("the kitty holds %d cards, want %d", len(kitty), SergeantMajorKittySize)
	}

	if err := s.DeclareTrump(CardDesignSpade); err != nil {
		t.Fatalf("declaring the trump: %v", err)
	}

	// 取り込んだ 4 枚はすべて印が付く。
	for _, c := range kitty {
		if !s.IsAbsorbedKittyCard(c) {
			t.Errorf("%v came from the kitty but is not marked", c)
		}
	}
	// **元から持っていた札には付かない。**印の意味が無くなる。
	marked := 0
	dealer := s.GetPlayer(0)
	for i := range dealer.GetCardsSize() {
		if s.IsAbsorbedKittyCard(dealer.GetCard(i)) {
			marked++
		}
	}
	if marked != SergeantMajorKittySize {
		t.Errorf("%d cards in hand are marked, want %d", marked, SergeantMajorKittySize)
	}

	// 捨て終われば印は消える (受け入れ条件3)。
	if err := s.DiscardForTest(0, []int{0, 1, 2, 3}); err != nil {
		t.Fatalf("discarding: %v", err)
	}
	for _, c := range kitty {
		if s.IsAbsorbedKittyCard(c) {
			t.Errorf("%v is still marked after the discard", c)
		}
	}
}
