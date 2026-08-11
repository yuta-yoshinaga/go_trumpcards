//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTeenDoPaanch(t *testing.T) *TeenDoPaanch {
	t.Helper()
	g := NewDefaultTeenDoPaanch()
	g.Reset()
	return g
}

// **issue の「30枚（8〜A）」は算術が合わない。** 8〜A は 7 ランク × 4 = 28 枚
// しかなく、3 人 × 10 枚に 2 枚足りない。7♠ と 7♥ を足して 30 枚にする。
func TestTeenDoPaanch_DeckIsThirtyCards(t *testing.T) {
	deck := NewTrumpCardsTeenDoPaanch()
	assert.Equal(t, TeenDoPaanchDeckSize, deck.GetTotalCount())
	assert.Equal(t, 30, TeenDoPaanchDeckSize, "3 人 × 10 枚")
	assert.Equal(t, TeenDoPaanchPlayerCnt*TeenDoPaanchHandSize, deck.GetTotalCount(), "配り切れる")

	seen := map[string]int{}
	sevens := 0
	for range TeenDoPaanchDeckSize {
		c := deck.DrawCard()
		require.NotNil(t, c)
		seen[cardStr(c)]++
		if c.GetValue() == 7 {
			sevens++
		}
		assert.NotContains(t, []int{2, 3, 4, 5, 6}, c.GetValue(), "6 以下は入らない")
	}
	assert.Len(t, seen, TeenDoPaanchDeckSize, "重複が無い")
	assert.Equal(t, 2, sevens, "7 は ♠ と ♥ の 2 枚だけ")
}

// **3 + 2 + 5 = 10 でトリックが余りも不足もしない。**
func TestTeenDoPaanch_TargetsSumToTheTrickCount(t *testing.T) {
	sum := 0
	for _, v := range TeenDoPaanchTargets {
		sum += v
	}
	assert.Equal(t, TeenDoPaanchTricksPerRound, sum)
	assert.Equal(t, TeenDoPaanchHandSize, TeenDoPaanchTricksPerRound)
}

func TestTeenDoPaanch_ResetDealsFiveThenWaitsForTrump(t *testing.T) {
	g := newTestTeenDoPaanch(t)

	assert.Equal(t, TeenDoPaanchPhaseTrump, g.GetPhase(), "切り札を決めるまで打たない")
	assert.Equal(t, 0, g.GetTrumpSuit())
	assert.Equal(t, 1, g.GetRoundNumber())
	// **手札が揃う前に切り札を決めるのが賭けどころ。**
	for i := range TeenDoPaanchPlayerCnt {
		assert.Equal(t, TeenDoPaanchFirstDeal, g.GetPlayer(i).GetCardsSize(), "まず 5 枚だけ")
	}
	assert.Empty(t, g.GetValidPlayIndices(0), "宣言前は出せない")
}

// **ノルマは宣言ではなく割り当て。** 3 人に 3・2・5 が必ず 1 つずつ。
func TestTeenDoPaanch_TargetsAreAssignedNotBid(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	got := map[int]int{}
	for i := range TeenDoPaanchPlayerCnt {
		got[g.GetPlayer(i).GetTarget()]++
	}
	assert.Equal(t, map[int]int{3: 1, 2: 1, 5: 1}, got)
	assert.Equal(t, 5, g.GetPlayer(g.GetFivePlayerIdx()).GetTarget(), "5 の席が切り札を決める")
}

// **ノルマ 5 は席を移る。** 3 ラウンドで全員が 3・2・5 を一度ずつ引き受ける。
func TestTeenDoPaanch_TheFiveRoleRotates(t *testing.T) {
	g := NewDefaultTeenDoPaanch()
	g.SetConfig(TeenDoPaanchConfig{Rounds: TeenDoPaanchRoundsMax})
	g.Reset()

	seenFive := map[int]bool{}
	targetsSeen := map[int]map[int]bool{0: {}, 1: {}, 2: {}}
	for range TeenDoPaanchPlayerCnt {
		seenFive[g.GetFivePlayerIdx()] = true
		for i := range TeenDoPaanchPlayerCnt {
			targetsSeen[i][g.GetPlayer(i).GetTarget()] = true
		}
		g.SetPhaseForTest(TeenDoPaanchPhaseRoundEnd)
		g.NextRound()
	}
	assert.Len(t, seenFive, TeenDoPaanchPlayerCnt, "3 ラウンドで全員が 5 を引き受ける")
	for i := range TeenDoPaanchPlayerCnt {
		assert.Len(t, targetsSeen[i], TeenDoPaanchPlayerCnt, "席 %d が 3 種類すべてを経験する", i)
	}
}

// **宣言したら残りを配り切る。** 5 枚のままでは 10 トリック打てない。
func TestTeenDoPaanch_DeclaringTrumpCompletesTheDeal(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetFivePlayerIdxForTest(0)
	require.NoError(t, g.DeclareTrump(CardDesignHeart))

	assert.Equal(t, CardDesignHeart, g.GetTrumpSuit())
	assert.Equal(t, TeenDoPaanchPhasePlay, g.GetPhase())
	total := 0
	for i := range TeenDoPaanchPlayerCnt {
		assert.Equal(t, TeenDoPaanchHandSize, g.GetPlayer(i).GetCardsSize())
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, TeenDoPaanchDeckSize, total, "山札は残らない")
}

func TestTeenDoPaanch_DeclareTrumpRejectsBadInput(t *testing.T) {
	for _, suit := range []int{0, -1, 5, 99} {
		g := newTestTeenDoPaanch(t)
		assert.Error(t, g.DeclareTrump(suit))
		assert.Equal(t, TeenDoPaanchPhaseTrump, g.GetPhase())
	}

	g := newTestTeenDoPaanch(t)
	require.NoError(t, g.DeclareTrump(CardDesignSpade))
	assert.Error(t, g.DeclareTrump(CardDesignHeart), "二度は宣言できない")

	g.FinishGameForTest()
	assert.Error(t, g.DeclareTrump(CardDesignHeart), "終局後は宣言できない")
}

// **宣言できるのは 5 の席だけ。**
func TestTeenDoPaanch_OnlyTheFiveSeatDeclares(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetFivePlayerIdxForTest(1)
	assert.False(t, g.IsHumanTrumpTurn())
	assert.Error(t, g.PlayerDeclareTrump(CardDesignHeart))

	g.SetFivePlayerIdxForTest(0)
	assert.True(t, g.IsHumanTrumpTurn())
	assert.NoError(t, g.PlayerDeclareTrump(CardDesignHeart))
}

func TestTeenDoPaanch_CpuDeclaresItsLongestSuit(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetFivePlayerIdxForTest(1)
	teenDoPaanchHandOf(g, 1,
		NewCard(CardDesignHeart, 8, false), NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignHeart, 10, false), NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 13, false))
	assert.Equal(t, CardDesignHeart, g.CpuTrumpChoiceForTest(1))

	g.CpuDeclareTrump()
	assert.Equal(t, CardDesignHeart, g.GetTrumpSuit())
	assert.Equal(t, TeenDoPaanchPhasePlay, g.GetPhase())
}

// **フォロー義務あり。**
func TestTeenDoPaanch_FollowSuitIsCompulsory(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	require.NoError(t, g.DeclareTrump(CardDesignDiamond))
	g.SetLeadPlayerIdxForTest(0)
	g.SetCurrentPlayerIdxForTest(0)
	teenDoPaanchHandOf(g, 0, NewCard(CardDesignSpade, 8, false))
	teenDoPaanchHandOf(g, 1, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 8, false))
	teenDoPaanchHandOf(g, 2, NewCard(CardDesignSpade, 10, false))

	require.NoError(t, g.PlayForTest(0, 0))
	assert.Equal(t, []int{0}, g.GetValidPlayIndices(1))
	assert.Error(t, g.PlayForTest(1, 1), "フォローできるのに別スートは通さない")
}

// **切り札はリードスートより強い。**
func TestTeenDoPaanch_TrumpBeatsTheLedSuit(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetTrumpSuitForTest(CardDesignDiamond)
	g.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},   // ♠A
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 8, false)}, // 切り札の 8
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 13, false)},
	})
	assert.Equal(t, 1, g.TrickWinnerForTest(), "切り札の最弱札が ♠A に勝つ")
}

// **切り札同士なら強いほう、切り札が無ければリードスートの最強札。**
func TestTeenDoPaanch_TrickWinnerOrdering(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetTrumpSuitForTest(CardDesignDiamond)

	g.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignDiamond, 9, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 12, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 8, false)},
	})
	assert.Equal(t, 1, g.TrickWinnerForTest(), "切り札同士は強いほう")

	g.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 10, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)}, // 別スートの A は取れない
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 13, false)},
	})
	assert.Equal(t, 2, g.TrickWinnerForTest())

	// **A は最強、7 は最弱。** 序列を端から端まで踏む。
	g.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 13, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 1, false)},
	})
	assert.Equal(t, 2, g.TrickWinnerForTest())
}

// **ノルマちょうど以上で達成。多く取っても達成は達成。**
func TestTeenDoPaanch_MeetingTheTargetCountsOnceRegardlessOfSurplus(t *testing.T) {
	for _, tc := range []struct {
		name        string
		tricks      [TeenDoPaanchPlayerCnt]int
		wantMet     [TeenDoPaanchPlayerCnt]bool
		wantSurplus [TeenDoPaanchPlayerCnt]int
	}{
		{"全員ちょうど", [3]int{5, 3, 2}, [3]bool{true, true, true}, [3]int{0, 0, 0}},
		{"5 が超過し 3 が不足", [3]int{7, 1, 2}, [3]bool{true, false, true}, [3]int{2, -2, 0}},
		{"2 が超過し 5 が不足", [3]int{3, 3, 4}, [3]bool{false, true, true}, [3]int{-2, 0, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestTeenDoPaanch(t)
			g.SetFivePlayerIdxForTest(0)
			require.NoError(t, g.DeclareTrump(CardDesignSpade))
			for i, n := range tc.tricks {
				g.GiveTricksForTest(i, n)
			}
			g.FinishRoundForTest()

			for i := range TeenDoPaanchPlayerCnt {
				assert.Equal(t, tc.wantMet[i], g.GetPlayer(i).GetMet() == 1, "席 %d の達成", i)
				assert.Equal(t, tc.wantSurplus[i], g.GetSurplusForTest()[i], "席 %d の過不足", i)
			}
			// **過不足は必ず打ち消し合う。** 3+2+5 = 10 = 全トリック数だから。
			sum := 0
			for _, v := range g.GetSurplusForTest() {
				sum += v
			}
			assert.Zero(t, sum)
		})
	}
}

// **超過したぶんだけ相手の良い札を召し上げる。** 多く取る意味はここだけ。
func TestTeenDoPaanch_SurplusTakesTheBestCardsNextRound(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetFivePlayerIdxForTest(0)
	require.NoError(t, g.DeclareTrump(CardDesignSpade))
	g.GiveTricksForTest(0, 7) // ノルマ 5 に +2
	g.GiveTricksForTest(1, 1) // ノルマ 3 に -2
	g.GiveTricksForTest(2, 2)
	g.FinishRoundForTest()
	require.Equal(t, []int{2, -2, 0}, g.GetSurplusForTest())

	g.NextRound()
	require.NoError(t, g.DeclareTrump(CardDesignSpade))

	assert.Equal(t, 2, g.GetLastExchange(), "2 枚動く")
	for i := range TeenDoPaanchPlayerCnt {
		assert.Equal(t, TeenDoPaanchHandSize, g.GetPlayer(i).GetCardsSize(), "枚数は変わらない")
	}
	// **やり取りは使い切ったら消える。**
	assert.Equal(t, []int{0, 0, 0}, g.GetSurplusForTest())
}

// **やり取りは強い札が上へ、弱い札が下へ動く。**
func TestTeenDoPaanch_ExchangeMovesTheBestCardUp(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetPhaseForTest(TeenDoPaanchPhasePlay)
	teenDoPaanchHandOf(g, 0, NewCard(CardDesignClover, 8, false))
	teenDoPaanchHandOf(g, 1, NewCard(CardDesignHeart, 1, false)) // ♥A が最強
	teenDoPaanchHandOf(g, 2, NewCard(CardDesignSpade, 9, false))
	g.SetSurplusForTest([]int{1, -1, 0})

	g.ExchangeForTest()

	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, 14, teenDoPaanchRank(g.GetPlayer(0).GetCard(0)), "超過した席に ♥A が移る")
	assert.Equal(t, 8, teenDoPaanchRank(g.GetPlayer(1).GetCard(0)), "代わりに最弱札が戻る")
	assert.Equal(t, 1, g.GetLastExchange())
}

// **負のコントロール: 過不足が無ければ 1 枚も動かない。**
func TestTeenDoPaanch_NoSurplusMovesNothing(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetPhaseForTest(TeenDoPaanchPhasePlay)
	teenDoPaanchHandOf(g, 0, NewCard(CardDesignClover, 8, false))
	teenDoPaanchHandOf(g, 1, NewCard(CardDesignHeart, 1, false))
	teenDoPaanchHandOf(g, 2, NewCard(CardDesignSpade, 9, false))
	g.SetSurplusForTest([]int{0, 0, 0})

	g.ExchangeForTest()

	assert.Equal(t, 0, g.GetLastExchange())
	assert.Equal(t, 14, teenDoPaanchRank(g.GetPlayer(1).GetCard(0)), "♥A は動かない")
}

// **規定ラウンドで終わり、達成回数の多い人が勝ち。**
func TestTeenDoPaanch_MostTargetsMetWins(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.GetPlayer(0).SetMet(3)
	g.GetPlayer(1).SetMet(1)
	g.GetPlayer(2).SetMet(2)
	g.FinishGameForTest()

	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, TeenDoPaanchPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

// **同点は -1（勝者なし）。**
func TestTeenDoPaanch_TieHasNoWinner(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	for i := range TeenDoPaanchPlayerCnt {
		g.GetPlayer(i).SetMet(2)
	}
	g.FinishGameForTest()
	assert.Equal(t, -1, g.GetWinnerIdx())
}

// **規定ラウンドに達したらそこで終わる。**
func TestTeenDoPaanch_GameEndsAtTheConfiguredRound(t *testing.T) {
	g := NewDefaultTeenDoPaanch()
	g.SetConfig(TeenDoPaanchConfig{Rounds: TeenDoPaanchRoundsMin})
	g.Reset()

	for round := 1; round <= TeenDoPaanchRoundsMin; round++ {
		require.Equal(t, round, g.GetRoundNumber())
		require.NoError(t, g.DeclareTrump(CardDesignSpade))
		for !g.GetGameEndFlag() && g.GetPhase() == TeenDoPaanchPhasePlay {
			idx := g.GetCurrentPlayerIdx()
			require.NoError(t, g.PlayForTest(idx, g.CpuChoiceForTest(idx)))
		}
		if round < TeenDoPaanchRoundsMin {
			require.False(t, g.GetGameEndFlag(), "まだ終わらない")
			g.NextRound()
		}
	}
	assert.True(t, g.GetGameEndFlag())
	// **達成回数の合計はラウンド数を超えない席ごとの数え方。**
	for i := range TeenDoPaanchPlayerCnt {
		assert.LessOrEqual(t, g.GetPlayer(i).GetMet(), TeenDoPaanchRoundsMin)
	}
}

// **1 ラウンドはちょうど 10 トリック。**
func TestTeenDoPaanch_ARoundIsExactlyTenTricks(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	require.NoError(t, g.DeclareTrump(CardDesignSpade))
	for g.GetPhase() == TeenDoPaanchPhasePlay {
		idx := g.GetCurrentPlayerIdx()
		require.NoError(t, g.PlayForTest(idx, g.CpuChoiceForTest(idx)))
	}
	assert.Equal(t, TeenDoPaanchTricksPerRound, g.GetTrickNumber())
	total := 0
	for i := range TeenDoPaanchPlayerCnt {
		total += g.GetPlayer(i).GetTrickCount()
		assert.Zero(t, g.GetPlayer(i).GetCardsSize(), "手札を打ち切る")
	}
	assert.Equal(t, TeenDoPaanchTricksPerRound, total, "トリックは余りも不足もしない")
}

// **CPU は必ず合法手を返す。**
func TestTeenDoPaanch_CpuAlwaysChoosesALegalCard(t *testing.T) {
	for range 100 {
		g := NewDefaultTeenDoPaanch()
		g.SetConfig(TeenDoPaanchConfig{Rounds: TeenDoPaanchRoundsMin})
		g.Reset()
		for !g.GetGameEndFlag() {
			switch g.GetPhase() {
			case TeenDoPaanchPhaseTrump:
				require.NoError(t, g.DeclareTrump(g.CpuTrumpChoiceForTest(g.GetFivePlayerIdx())))
			case TeenDoPaanchPhasePlay:
				idx := g.GetCurrentPlayerIdx()
				choice := g.CpuChoiceForTest(idx)
				require.Contains(t, g.GetValidPlayIndices(idx), choice)
				require.NoError(t, g.PlayForTest(idx, choice))
			case TeenDoPaanchPhaseRoundEnd:
				g.NextRound()
			default:
			}
		}
	}
}

// **ノルマに届いていなければ取りにいき、届いたら降りる。**
func TestTeenDoPaanch_CpuChasesThenDucks(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetPhaseForTest(TeenDoPaanchPhasePlay)
	g.SetTrumpSuitForTest(CardDesignDiamond)
	g.SetCurrentPlayerIdxForTest(1)
	teenDoPaanchHandOf(g, 1,
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignSpade, 1, false), // ♠A
		NewCard(CardDesignSpade, 10, false))

	g.GetPlayer(1).SetTarget(3)
	assert.Equal(t, 1, g.CpuChoiceForTest(1), "届いていないので最強札")

	g.GiveTricksForTest(1, 3)
	assert.Equal(t, 0, g.CpuChoiceForTest(1), "届いたら最弱札")
}

func TestTeenDoPaanch_RejectsOutOfTurnAndBadIndices(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	assert.Error(t, g.PlayForTest(0, 0), "宣言前は打てない")

	require.NoError(t, g.DeclareTrump(CardDesignSpade))
	g.SetCurrentPlayerIdxForTest(0)
	assert.Error(t, g.PlayForTest(1, 0), "手番でない席は打てない")
	assert.Error(t, g.PlayForTest(0, -1))
	assert.Error(t, g.PlayForTest(0, 999))

	g.FinishGameForTest()
	assert.Error(t, g.PlayForTest(0, 0), "終局後は打てない")
}

func TestTeenDoPaanch_PlayerPlayGuardsOnTheHumanTurn(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	require.NoError(t, g.DeclareTrump(CardDesignSpade))
	g.SetCurrentPlayerIdxForTest(1)
	assert.False(t, g.IsHumanTurn())
	assert.Error(t, g.PlayerPlay(0))

	g.SetCurrentPlayerIdxForTest(0)
	assert.True(t, g.IsHumanTurn())
	assert.NoError(t, g.PlayerPlay(0))
}

func TestTeenDoPaanch_CpuPlayOnlyMovesOnACpuTurn(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	require.NoError(t, g.DeclareTrump(CardDesignSpade))
	g.SetCurrentPlayerIdxForTest(0)
	before := g.GetPlayer(0).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize(), "人間の手番では動かない")

	g.SetCurrentPlayerIdxForTest(1)
	beforeCpu := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, beforeCpu-1, g.GetPlayer(1).GetCardsSize())
}

func TestTeenDoPaanch_NextRoundOnlyFromRoundEnd(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	before := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, before, g.GetRoundNumber(), "宣言フェーズからは進まない")

	g.SetPhaseForTest(TeenDoPaanchPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, before+1, g.GetRoundNumber())

	g.SetPhaseForTest(TeenDoPaanchPhaseRoundEnd)
	g.FinishGameForTest()
	after := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, after, g.GetRoundNumber(), "終局後は進まない")
}

func TestTeenDoPaanch_GiveUp(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, TeenDoPaanchPhaseGameEnd, g.GetPhase())
	assert.Equal(t, -1, g.GetWinnerIdx())

	g.GiveUp() // 二度目は何も起きない
	assert.Equal(t, -1, g.GetWinnerIdx())
}

func TestTeenDoPaanch_Hint(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	g.SetFivePlayerIdxForTest(0)

	// **宣言フェーズの助言は札ではなくスートを指す。**
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex)
	assert.Equal(t, "teendopaanchSelectTrump", h.Reason)
	assert.GreaterOrEqual(t, h.Suit, CardDesignSpade)
	assert.LessOrEqual(t, h.Suit, CardDesignDiamond)

	require.NoError(t, g.DeclareTrump(CardDesignHeart))
	g.SetCurrentPlayerIdxForTest(0)
	h = g.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Contains(t, g.GetValidPlayIndices(0), *h.CardIndex, "勧める札は必ず合法")

	g.SetCurrentPlayerIdxForTest(1)
	assert.Nil(t, g.GetHint(), "人間の手番でなければ助言しない")

	g.SetCurrentPlayerIdxForTest(0)
	g.FinishGameForTest()
	assert.Nil(t, g.GetHint(), "終局後は助言しない")
}

// **助言の理由はノルマに届いたかで変わる。**
func TestTeenDoPaanch_HintReasonTracksTheTarget(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	require.NoError(t, g.DeclareTrump(CardDesignHeart))
	g.SetCurrentPlayerIdxForTest(0)
	g.GetPlayer(0).SetTarget(2)

	assert.Equal(t, "teendopaanchWinTrick", g.GetHint().Reason)
	g.GiveTricksForTest(0, 2)
	assert.Equal(t, "teendopaanchDuck", g.GetHint().Reason)
}

func TestTeenDoPaanch_JSONRoundTrip(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	require.NoError(t, g.DeclareTrump(CardDesignHeart))
	for range 6 {
		idx := g.GetCurrentPlayerIdx()
		require.NoError(t, g.PlayForTest(idx, g.CpuChoiceForTest(idx)))
	}

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored TeenDoPaanch
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetTrickNumber(), restored.GetTrickNumber())
	assert.Equal(t, g.GetFivePlayerIdx(), restored.GetFivePlayerIdx())
	for i := range TeenDoPaanchPlayerCnt {
		assert.Equal(t, g.GetPlayer(i).GetTarget(), restored.GetPlayer(i).GetTarget(), "ノルマが消えない")
		assert.Equal(t, g.GetPlayer(i).GetMet(), restored.GetPlayer(i).GetMet(), "達成数が消えない")
		assert.Equal(t, g.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
	}
}

// **壊れたスナップショットは弾く。**
func TestTeenDoPaanch_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	base := func(t *testing.T) map[string]any {
		t.Helper()
		g := newTestTeenDoPaanch(t)
		require.NoError(t, g.DeclareTrump(CardDesignHeart))
		data, err := json.Marshal(g)
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
		{"trump suit before it was declared", func(m map[string]any) {
			m["ph"] = int(TeenDoPaanchPhaseTrump)
		}},
		{"no trump suit after the declaration", func(m map[string]any) { m["ts"] = 0 }},
		{"current player out of range", func(m map[string]any) { m["ci"] = 9 }},
		{"five player out of range", func(m map[string]any) { m["fp"] = -1 }},
		{"winner before the game ended", func(m map[string]any) { m["wi"] = 1 }},
		{"round number below one", func(m map[string]any) { m["rn"] = 0 }},
		{"negative trick number", func(m map[string]any) { m["tn"] = -1 }},
		// **過不足は合計 0。** 誰かの超過は必ず誰かの不足。
		{"surplus does not cancel out", func(m map[string]any) { m["sp"] = []any{2, 0, 0} }},
		{"surplus has the wrong length", func(m map[string]any) { m["sp"] = []any{0, 0} }},
		{"config out of range", func(m map[string]any) { m["cf"] = map[string]any{"r": 0} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base(t)
			tc.mutate(m)
			data, err := json.Marshal(m)
			require.NoError(t, err)
			var restored TeenDoPaanch
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通る。**
	data, err := json.Marshal(base(t))
	require.NoError(t, err)
	var ok TeenDoPaanch
	assert.NoError(t, json.Unmarshal(data, &ok))
}

func TestTeenDoPaanchConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultTeenDoPaanchConfig().Validate())
	assert.NoError(t, TeenDoPaanchConfig{Rounds: TeenDoPaanchRoundsMin}.Validate())
	assert.NoError(t, TeenDoPaanchConfig{Rounds: TeenDoPaanchRoundsMax}.Validate())
	assert.Error(t, TeenDoPaanchConfig{Rounds: TeenDoPaanchRoundsMin - 1}.Validate())
	assert.Error(t, TeenDoPaanchConfig{Rounds: TeenDoPaanchRoundsMax + 1}.Validate())
}

// アクセサがそのまま盤面を返すこと（プレゼンタが読む値の入口）。
func TestTeenDoPaanch_AccessorsExposeTheBoard(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	assert.Equal(t, TeenDoPaanchDefaultRounds, g.GetConfig().Rounds)
	assert.Equal(t, g.GetFivePlayerIdx(), g.GetLeadPlayerIdx(), "配り直後のリードは 5 の席")
	assert.Empty(t, g.GetCurrentTrick())
	assert.NotEmpty(t, g.GetActionLog(), "配りが記録される")
	assert.Zero(t, g.GetLastExchange(), "初回はやり取り無し")

	require.NoError(t, g.DeclareTrump(CardDesignHeart))
	idx := g.GetCurrentPlayerIdx()
	require.NoError(t, g.PlayForTest(idx, g.CpuChoiceForTest(idx)))
	assert.Len(t, g.GetCurrentTrick(), 1)
	assert.Equal(t, idx, g.GetCurrentTrick()[0].PlayerIdx)
}

func TestTeenDoPaanch_GetPlayerBounds(t *testing.T) {
	g := newTestTeenDoPaanch(t)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetPlayer(0))
	assert.Empty(t, g.GetValidPlayIndices(-1))
	assert.Empty(t, g.GetValidPlayIndices(99))
	assert.Equal(t, TeenDoPaanchPlayerCnt, g.GetPlayerCnt())
}
