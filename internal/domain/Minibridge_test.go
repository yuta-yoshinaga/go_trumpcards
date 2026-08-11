//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMinibridge(t *testing.T) *Minibridge {
	t.Helper()
	m := NewDefaultMinibridge()
	m.Reset()
	return m
}

// **4 × 13 = 52 ちょうど。** 着手前に検算した唯一の配り。
func TestMinibridge_DealUsesTheWholeDeck(t *testing.T) {
	assert.Equal(t, 52, MinibridgePlayerCnt*MinibridgeHandSize, "4 人 × 13 枚 = 52")

	m := newTestMinibridge(t)
	total := 0
	for i := range MinibridgePlayerCnt {
		assert.Equal(t, MinibridgeHandSize, m.GetPlayer(i).GetCardsSize())
		total += m.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)
}

// **HCP の総和は必ず 40。** A4 + K3 + Q2 + J1 = 10 が 4 スートぶん。
// 配り方に依らないので、何度配っても崩れない。
func TestMinibridge_HcpAlwaysTotalsForty(t *testing.T) {
	for range 50 {
		m := NewDefaultMinibridge()
		m.Reset()
		total := 0
		for i := range MinibridgePlayerCnt {
			hcp := m.GetPlayer(i).GetHcp()
			assert.GreaterOrEqual(t, hcp, 0)
			assert.LessOrEqual(t, hcp, MinibridgeTotalHcp)
			total += hcp
		}
		require.Equal(t, MinibridgeTotalHcp, total)
	}
}

func TestMinibridge_HcpValues(t *testing.T) {
	for value, want := range map[int]int{1: 4, 13: 3, 12: 2, 11: 1, 10: 0, 2: 0} {
		assert.Equal(t, want, minibridgeHcpValue(NewCard(CardDesignSpade, value, false)), "value=%d", value)
	}
	// 1 スートぶんの合計が 10 になることまで確かめる（40 の根拠）。
	suitTotal := 0
	for v := 1; v <= 13; v++ {
		suitTotal += minibridgeHcpValue(NewCard(CardDesignHeart, v, false))
	}
	assert.Equal(t, MinibridgeTotalHcp/4, suitTotal)
}

func TestMinibridge_ResetStartsInTheContractPhase(t *testing.T) {
	m := newTestMinibridge(t)
	assert.Equal(t, MinibridgePhaseContract, m.GetPhase())
	assert.Equal(t, 0, m.GetContractLevel(), "契約はまだ選ばれていない")
	assert.Equal(t, 1, m.GetRoundNumber())
	assert.GreaterOrEqual(t, m.GetDeclarerIdx(), 0, "配った時点で落札者は決まっている")
	assert.Equal(t, (m.GetDeclarerIdx()+MinibridgeTeamCnt)%MinibridgePlayerCnt, m.GetDummyIdx())
	assert.Nil(t, m.GetDummyHand(), "契約が決まるまでダミーは伏せる")
}

// **落札者は HCP の合計が多いペアの、多い方の席。**
func TestMinibridge_DeclarerIsTheStrongestSeatOfTheStrongerPair(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetDealerIdxForTest(0)
	m.SetHcpForTest([MinibridgePlayerCnt]int{5, 12, 11, 12})
	m.DecideDeclarerForTest()
	// team0 = 席0,2 = 16 / team1 = 席1,3 = 24 → team1、多いのは同点なので親から先の席 1
	assert.Equal(t, 1, m.GetDeclarerIdx())
	assert.Equal(t, 3, m.GetDummyIdx())

	m.SetHcpForTest([MinibridgePlayerCnt]int{9, 5, 21, 5})
	m.DecideDeclarerForTest()
	assert.Equal(t, 2, m.GetDeclarerIdx(), "team0 が 30 で強く、多いのは席 2")
	assert.Equal(t, 0, m.GetDummyIdx())
}

// **20-20 は起こる（実測 8.1%）。** issue はこの場合を書いていないので、
// 決定的に「親の側」と決めている。書かないと 12 回に 1 回止まる。
func TestMinibridge_PairTieGoesToTheDealersSide(t *testing.T) {
	for dealer := range MinibridgePlayerCnt {
		m := newTestMinibridge(t)
		m.SetDealerIdxForTest(dealer)
		m.SetHcpForTest([MinibridgePlayerCnt]int{10, 10, 10, 10})
		m.DecideDeclarerForTest()

		declTeam := m.GetPlayer(m.GetDeclarerIdx()).GetTeam()
		assert.Equal(t, m.GetPlayer(dealer).GetTeam(), declTeam, "dealer=%d", dealer)
		// 席も同点なので、親から見て先の席（＝親自身）が取る。
		assert.Equal(t, dealer, m.GetDeclarerIdx())
	}
}

// **ペアが決まっても席が同点のことがある（実測 4.4%）。**
func TestMinibridge_SeatTieGoesToTheSeatNearestTheDealer(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetDealerIdxForTest(2)
	// team0 = 席0,2 = 24 で強い。席 0 と 2 が同点なので、親（席 2）から先の席が取る。
	m.SetHcpForTest([MinibridgePlayerCnt]int{12, 8, 12, 8})
	m.DecideDeclarerForTest()
	assert.Equal(t, 2, m.GetDeclarerIdx())

	m.SetDealerIdxForTest(0)
	m.DecideDeclarerForTest()
	assert.Equal(t, 0, m.GetDeclarerIdx(), "親が変われば取る席も変わる")
}

// **ダミーは落札者の正面。**
func TestMinibridge_DummyIsAlwaysThePartner(t *testing.T) {
	for range 30 {
		m := NewDefaultMinibridge()
		m.Reset()
		decl, dummy := m.GetDeclarerIdx(), m.GetDummyIdx()
		require.NotEqual(t, decl, dummy)
		assert.Equal(t, m.GetPlayer(decl).GetTeam(), m.GetPlayer(dummy).GetTeam())
	}
}

// **契約は 6 + レベル。**
func TestMinibridge_RequiredTricksAddsTheBook(t *testing.T) {
	m := newTestMinibridge(t)
	assert.Zero(t, m.RequiredTricks(), "契約が無ければ 0")
	for level := 1; level <= MinibridgeMaxLevel; level++ {
		m.SetContractForTest(0, level, CardDesignHeart)
		assert.Equal(t, MinibridgeBookTricks+level, m.RequiredTricks())
	}
	assert.Equal(t, MinibridgeTotalTricks, m.RequiredTricks(), "レベル 7 は全トリック")
}

func TestMinibridge_ContractSelection(t *testing.T) {
	m := newTestMinibridge(t)
	decl := m.GetDeclarerIdx()

	// **落札者以外は契約を選べない。**
	other := (decl + 1) % MinibridgePlayerCnt
	assert.Error(t, m.SelectContractForTest(other, 3, 0))
	assert.Equal(t, MinibridgePhaseContract, m.GetPhase())

	require.NoError(t, m.SelectContractForTest(decl, 3, 0))
	assert.Equal(t, MinibridgePhasePlay, m.GetPhase())
	assert.Equal(t, 3, m.GetContractLevel())
	assert.Equal(t, 0, m.GetContractSuit(), "0 はノートランプ")
	assert.Equal(t, MinibridgeBookTricks+3, m.RequiredTricks())
	// **リードは落札者の左隣から。**
	assert.Equal(t, (decl+1)%MinibridgePlayerCnt, m.GetLeadPlayerIdx())
	assert.NotEmpty(t, m.GetDummyHand(), "契約が決まるとダミーが公開される")

	// 二度は選べない。
	assert.Error(t, m.SelectContractForTest(decl, 4, CardDesignHeart))
}

func TestMinibridge_ContractRejectsBadInput(t *testing.T) {
	for _, tc := range []struct {
		name        string
		level, suit int
	}{
		{"level below one", 0, CardDesignHeart},
		{"level above the maximum", MinibridgeMaxLevel + 1, CardDesignHeart},
		{"suit out of range", 2, 9},
		{"negative suit", 2, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestMinibridge(t)
			assert.Error(t, m.SelectContractForTest(m.GetDeclarerIdx(), tc.level, tc.suit))
			assert.Equal(t, MinibridgePhaseContract, m.GetPhase())
		})
	}
}

// **切り札は契約のスート。ノートランプでは効かない。**
func TestMinibridge_TrickWinner(t *testing.T) {
	trick := []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 2, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 13, false)},
	}

	m := newTestMinibridge(t)
	m.SetCurrentTrickForTest(trick)
	m.SetContractForTest(0, 1, 0)
	assert.Equal(t, 1, m.TrickWinnerForTest(), "NT なので ♠A")

	m.SetContractForTest(0, 1, CardDesignDiamond)
	assert.Equal(t, 2, m.TrickWinnerForTest(), "切り札の ♦2 が ♠A に勝つ")

	m.SetContractForTest(0, 1, CardDesignHeart)
	assert.Equal(t, 3, m.TrickWinnerForTest(), "切り札の ♥K")
}

func TestMinibridge_FollowSuitIsCompulsory(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetContractForTest(0, 1, CardDesignHeart)
	m.SetPhaseForTest(MinibridgePhasePlay)
	m.SetLeadPlayerIdxForTest(0)
	m.SetCurrentPlayerIdxForTest(0)
	minibridgeHandOf(m, 0, NewCard(CardDesignSpade, 8, false))
	minibridgeHandOf(m, 1, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 8, false))

	require.NoError(t, m.PlayForTest(0, 0))
	assert.Equal(t, []int{0}, m.GetValidPlayIndices(1))
	assert.Error(t, m.PlayForTest(1, 1))
}

// **得点表を全ケース踏む。** 契約点 100 が game / partscore の境目。
func TestMinibridge_ContractPoints(t *testing.T) {
	m := newTestMinibridge(t)
	for _, tc := range []struct {
		level, suit, want int
	}{
		{1, CardDesignClover, 20}, {5, CardDesignClover, 100}, {7, CardDesignDiamond, 140},
		{1, CardDesignHeart, 30}, {4, CardDesignSpade, 120}, {7, CardDesignHeart, 210},
		{1, 0, 40}, {2, 0, 70}, {3, 0, 100}, {7, 0, 220},
	} {
		m.SetContractForTest(0, tc.level, tc.suit)
		assert.Equal(t, tc.want, m.ContractPointsForTest(), "level=%d suit=%d", tc.level, tc.suit)
	}

	// **ゲーム成立はちょうど 3NT / 4M / 5m。** 実際のブリッジと一致する。
	for suit, wantLevel := range map[int]int{
		0: 3, CardDesignHeart: 4, CardDesignSpade: 4, CardDesignClover: 5, CardDesignDiamond: 5,
	} {
		lowest := 0
		for level := 1; level <= MinibridgeMaxLevel; level++ {
			m.SetContractForTest(0, level, suit)
			if m.ContractPointsForTest() >= MinibridgeGameThreshold {
				lowest = level
				break
			}
		}
		assert.Equal(t, wantLevel, lowest, "suit=%d", suit)
	}
}

func TestMinibridge_Scoring(t *testing.T) {
	for _, tc := range []struct {
		name        string
		level, suit int
		took        int
		wantDecl    int
		wantOther   int
		made        bool
	}{
		{"partscore ちょうど", 2, CardDesignHeart, 8, 60 + 50, 0, true},
		{"partscore オーバートリック", 2, CardDesignHeart, 10, 60 + 50 + 60, 0, true},
		{"game", 4, CardDesignHeart, 10, 120 + 300, 0, true},
		{"NT の game", 3, 0, 9, 100 + 300, 0, true},
		{"スモールスラム", 6, CardDesignSpade, 12, 180 + 300 + 500, 0, true},
		{"グランドスラム", 7, 0, 13, 220 + 300 + 1000, 0, true},
		{"1 トリック不足", 3, CardDesignHeart, 8, 0, 50, false},
		{"全部落とす", 1, 0, 0, 0, 350, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestMinibridge(t)
			m.SetContractForTest(0, tc.level, tc.suit)
			m.SetPhaseForTest(MinibridgePhasePlay)
			m.GiveTricksForTest(0, tc.took)
			m.GiveTricksForTest(1, MinibridgeTotalTricks-tc.took)
			m.FinishRoundForTest()

			assert.Equal(t, tc.wantDecl, m.GetTeamScore(0))
			assert.Equal(t, tc.wantOther, m.GetTeamScore(1))
			assert.Equal(t, tc.made, m.GetLastMade())
			assert.Equal(t, tc.took, m.GetLastTricks())
		})
	}
}

// **ペアの 2 席ぶんを合算する。** 落札者ひとりの取得数で判定すると必ず落ちる。
func TestMinibridge_ScoringCountsBothSeatsOfTheDeclaringSide(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetContractForTest(0, 1, CardDesignHeart)
	m.SetPhaseForTest(MinibridgePhasePlay)
	m.GiveTricksForTest(0, 3)
	m.GiveTricksForTest(2, 4) // ダミーが取ったぶん
	m.GiveTricksForTest(1, 6)
	m.FinishRoundForTest()

	assert.True(t, m.GetLastMade(), "3 + 4 = 7 で 1♥ の契約は成立する")
	assert.Equal(t, 7, m.GetLastTricks())
}

func TestMinibridge_NextRoundRotatesTheDealer(t *testing.T) {
	m := newTestMinibridge(t)
	before := m.GetDealerIdx()
	m.SetPhaseForTest(MinibridgePhaseRoundEnd)
	m.NextRound()
	assert.Equal(t, (before+1)%MinibridgePlayerCnt, m.GetDealerIdx())
	assert.Equal(t, MinibridgePhaseContract, m.GetPhase())
	assert.Equal(t, 2, m.GetRoundNumber())

	m.FinishGameForTest()
	after := m.GetDealerIdx()
	m.NextRound()
	assert.Equal(t, after, m.GetDealerIdx(), "終局後は進まない")
}

func TestMinibridge_ReachingTheLastRoundEndsTheGame(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetRoundNumberForTest(m.GetConfig().Rounds)
	m.SetContractForTest(0, 1, CardDesignHeart)
	m.SetPhaseForTest(MinibridgePhasePlay)
	m.GiveTricksForTest(0, MinibridgeTotalTricks)
	m.FinishRoundForTest()

	assert.True(t, m.GetGameEndFlag())
	assert.Equal(t, 0, m.GetWinnerTeam())
}

func TestMinibridge_RejectsOutOfTurnAndBadIndices(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetContractForTest(0, 1, CardDesignHeart)
	m.SetPhaseForTest(MinibridgePhasePlay)
	idx := m.GetCurrentPlayerIdx()
	assert.Error(t, m.PlayForTest((idx+1)%MinibridgePlayerCnt, 0), "手番でない席は打てない")
	assert.Error(t, m.PlayForTest(idx, -1))
	assert.Error(t, m.PlayForTest(idx, 999))

	m.SetPhaseForTest(MinibridgePhaseContract)
	assert.Error(t, m.PlayForTest(idx, 0), "契約フェーズでは打てない")

	m.FinishGameForTest()
	assert.Error(t, m.PlayForTest(idx, 0), "終局後は打てない")
	assert.Error(t, m.SelectContractForTest(m.GetDeclarerIdx(), 1, 0), "終局後は契約を選べない")
}

// **ダミーの手番は人間デクレアラーが操作する。** ここを落とすと盤面が止まる。
func TestMinibridge_HumanDeclarerPlaysTheDummy(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetContractForTest(0, 1, CardDesignHeart) // 席 0 = 人間が落札者、ダミーは席 2
	m.SetPhaseForTest(MinibridgePhasePlay)
	m.SetCurrentPlayerIdxForTest(2)

	assert.True(t, m.IsHumanTurn(), "ダミーの手番も人間の出番")
	before := m.GetPlayer(2).GetCardsSize()
	require.NoError(t, m.PlayerPlay(m.GetValidPlayIndices(2)[0]))
	assert.Equal(t, before-1, m.GetPlayer(2).GetCardsSize())

	// **CPU はダミーを触らない。**
	m.SetCurrentPlayerIdxForTest(2)
	dummyBefore := m.GetPlayer(2).GetCardsSize()
	m.CpuPlay()
	assert.Equal(t, dummyBefore, m.GetPlayer(2).GetCardsSize())
}

// **落札者が CPU なら、ダミーも CPU が動かす。**
//
// **ダミーが人間の席と重なる組み合わせを必ず踏む（レビュー指摘 PR #5313）。**
// ダミーは落札者の相方なので、落札者が席 2 ならダミーは席 0 ——人間自身の席。
// 「その席が人間か」で判定していると、ここだけ人間の番と誤判定して盤面が止まる。
func TestMinibridge_CpuDeclarerPlaysItsOwnDummy(t *testing.T) {
	for _, tc := range []struct {
		name            string
		declarer, dummy int
	}{
		{"ダミーも CPU の席", 1, 3},
		{"ダミーが人間の席", 2, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestMinibridge(t)
			m.SetContractForTest(tc.declarer, 1, CardDesignHeart)
			require.Equal(t, tc.dummy, m.GetDummyIdx())
			m.SetPhaseForTest(MinibridgePhasePlay)
			m.SetCurrentPlayerIdxForTest(tc.dummy)

			assert.False(t, m.IsHumanTurn(), "ダミーの手番を握るのは席の持ち主でなく落札者")
			assert.Error(t, m.PlayerPlay(0), "人間は CPU 落札者のダミーを操作できない")

			before := m.GetPlayer(tc.dummy).GetCardsSize()
			m.CpuPlay()
			assert.Equal(t, before-1, m.GetPlayer(tc.dummy).GetCardsSize(), "CPU が進めるので盤面は止まらない")
		})
	}
}

// **公開の入口も踏む。**
func TestMinibridge_PublicEntryPointsGuardTheTurn(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetContractForTest(1, 1, CardDesignHeart)
	m.SetPhaseForTest(MinibridgePhasePlay)
	m.SetCurrentPlayerIdxForTest(1)
	assert.False(t, m.IsHumanTurn())
	assert.Error(t, m.PlayerPlay(0))

	m.SetPhaseForTest(MinibridgePhaseContract)
	m.SetContractForTest(1, 0, 0)
	assert.False(t, m.IsHumanContractTurn(), "CPU が落札者なら人間は選ばない")
	assert.Error(t, m.PlayerSelectContract(3, 0))
	m.CpuSelectContract()
	assert.Equal(t, MinibridgePhasePlay, m.GetPhase(), "CPU が契約を選んで進む")

	m2 := newTestMinibridge(t)
	m2.SetContractForTest(0, 0, 0)
	m2.SetPhaseForTest(MinibridgePhaseContract)
	assert.True(t, m2.IsHumanContractTurn())
	before := m2.GetPhase()
	m2.CpuSelectContract()
	assert.Equal(t, before, m2.GetPhase(), "人間が落札者なら CPU は選ばない")
	require.NoError(t, m2.PlayerSelectContract(2, CardDesignSpade))
	assert.Equal(t, 2, m2.GetContractLevel())
}

// **CPU は必ず合法手・合法な契約を返す。**
func TestMinibridge_CpuAlwaysChoosesLegally(t *testing.T) {
	for range 40 {
		m := NewDefaultMinibridge()
		m.Reset()
		for turns := 0; !m.GetGameEndFlag() && turns < 500; turns++ {
			switch m.GetPhase() {
			case MinibridgePhaseContract:
				level, suit := m.CpuContractChoiceForTest(m.GetDeclarerIdx())
				require.NoError(t, m.SelectContractForTest(m.GetDeclarerIdx(), level, suit))
			case MinibridgePhasePlay:
				idx := m.GetCurrentPlayerIdx()
				choice := m.CpuChoiceForTest(idx)
				require.Contains(t, m.GetValidPlayIndices(idx), choice)
				require.NoError(t, m.PlayForTest(idx, choice))
			case MinibridgePhaseRoundEnd:
				m.NextRound()
			default:
			}
		}
	}
}

// **どの局も必ず終わる。**
func TestMinibridge_GamesTerminate(t *testing.T) {
	for range 20 {
		m := NewDefaultMinibridge()
		m.Reset()
		for turns := 0; !m.GetGameEndFlag(); turns++ {
			require.Less(t, turns, 5000, "終わらない")
			switch m.GetPhase() {
			case MinibridgePhaseContract:
				level, suit := m.CpuContractChoiceForTest(m.GetDeclarerIdx())
				require.NoError(t, m.SelectContractForTest(m.GetDeclarerIdx(), level, suit))
			case MinibridgePhasePlay:
				idx := m.GetCurrentPlayerIdx()
				require.NoError(t, m.PlayForTest(idx, m.CpuChoiceForTest(idx)))
			case MinibridgePhaseRoundEnd:
				m.NextRound()
			default:
			}
		}
		assert.Equal(t, m.GetConfig().Rounds, m.GetRoundNumber())
		// 13 トリックが必ず全部どこかの席に入る。
		took := 0
		for i := range MinibridgePlayerCnt {
			took += m.GetPlayer(i).GetTrickCount()
		}
		assert.Equal(t, MinibridgeTotalTricks, took)
	}
}

// **CPU の契約は配りに依らず決まる。** map の反復順に乗ると同じ手で違う契約が出る。
func TestMinibridge_CpuContractIsDeterministic(t *testing.T) {
	m := newTestMinibridge(t)
	minibridgeHandOf(m, 0,
		NewCard(CardDesignSpade, 1, false), NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 1, false), NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 12, false), NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignHeart, 2, false),
	)
	m.SetContractForTest(0, 0, 0)
	first, firstSuit := m.CpuContractChoiceForTest(0)
	for range 20 {
		level, suit := m.CpuContractChoiceForTest(0)
		require.Equal(t, first, level)
		require.Equal(t, firstSuit, suit)
	}
	assert.Contains(t, []int{CardDesignSpade, CardDesignHeart}, firstSuit, "5 枚あるスートを選ぶ")
}

func TestMinibridge_GiveUp(t *testing.T) {
	m := newTestMinibridge(t)
	m.GiveUp()
	assert.True(t, m.GetGameEndFlag())
	assert.Equal(t, 1, m.GetWinnerTeam())
	m.GiveUp()
	assert.Equal(t, 1, m.GetWinnerTeam())
}

func TestMinibridge_Hint(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetContractForTest(0, 0, 0)
	m.SetPhaseForTest(MinibridgePhaseContract)

	contractHint := m.GetHint()
	require.NotNil(t, contractHint)
	assert.Nil(t, contractHint.CardIndex)
	assert.Equal(t, "minibridgeContract", contractHint.Reason)
	assert.GreaterOrEqual(t, contractHint.Level, 1)

	m.SetContractForTest(0, 2, CardDesignHeart)
	m.SetPhaseForTest(MinibridgePhasePlay)
	m.SetCurrentPlayerIdxForTest(0)
	playHint := m.GetHint()
	require.NotNil(t, playHint)
	require.NotNil(t, playHint.CardIndex)
	assert.Equal(t, "minibridgeWinTrick", playHint.Reason)
	assert.Contains(t, m.GetValidPlayIndices(0), *playHint.CardIndex)

	// **ダミーを動かしているときは、そう言う。**
	m.SetCurrentPlayerIdxForTest(2)
	dummyHint := m.GetHint()
	require.NotNil(t, dummyHint)
	assert.Equal(t, "minibridgeDummy", dummyHint.Reason)

	m.SetCurrentPlayerIdxForTest(1)
	assert.Nil(t, m.GetHint(), "相手の手番では助言しない")

	m.SetCurrentPlayerIdxForTest(0)
	m.FinishGameForTest()
	assert.Nil(t, m.GetHint(), "終局後は助言しない")
}

func TestMinibridge_ConstructorFixesTheSeatCount(t *testing.T) {
	for _, given := range [][]*MinibridgePlayer{nil, {NewMinibridgePlayer(true, 0)}} {
		m := NewMinibridge(given, DefaultMinibridgeConfig())
		assert.Equal(t, MinibridgePlayerCnt, m.GetPlayerCnt())
		assert.NotPanics(t, m.Reset)
		assert.Equal(t, MinibridgeHandSize, m.GetPlayer(0).GetCardsSize())
	}

	// **負のコントロール: 4 人ちょうどなら渡したものをそのまま使う。**
	mine := []*MinibridgePlayer{
		NewMinibridgePlayer(false, 0), NewMinibridgePlayer(true, 1),
		NewMinibridgePlayer(false, 0), NewMinibridgePlayer(false, 1),
	}
	m := NewMinibridge(mine, DefaultMinibridgeConfig())
	assert.False(t, m.GetPlayer(0).GetIsHuman(), "席の並びを勝手に入れ替えない")
	assert.True(t, m.GetPlayer(1).GetIsHuman())
}

func TestMinibridge_AccessorsAndBounds(t *testing.T) {
	m := newTestMinibridge(t)
	assert.Nil(t, m.GetPlayer(-1))
	assert.Nil(t, m.GetPlayer(99))
	assert.Empty(t, m.GetValidPlayIndices(-1))
	assert.Empty(t, m.GetValidPlayIndices(99))
	assert.Equal(t, MinibridgePlayerCnt, m.GetPlayerCnt())
	assert.Zero(t, m.GetTeamScore(-1))
	assert.Zero(t, m.GetTeamScore(99))
	m.SetTeamScore(99, 10)
	m.SetTeamScore(-1, 10)
	assert.Zero(t, m.GetTeamScore(0), "範囲外の設定は無視する")
	assert.NotEmpty(t, m.GetActionLog())
	assert.Empty(t, m.GetCurrentTrick())
	assert.Zero(t, m.GetTrickNumber())
	cfg := MinibridgeConfig{Rounds: 8}
	m.SetConfig(cfg)
	assert.Equal(t, cfg, m.GetConfig())
	assert.Equal(t, "NT", minibridgeContractSuitStr(0))
	assert.Equal(t, "H", minibridgeContractSuitStr(CardDesignHeart))
	assert.True(t, minibridgeIsMinor(CardDesignClover))
	assert.True(t, minibridgeIsMinor(CardDesignDiamond))
	assert.False(t, minibridgeIsMinor(CardDesignSpade))
	assert.False(t, minibridgeIsMinor(CardDesignHeart))
}

func TestMinibridgeConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultMinibridgeConfig().Validate())
	assert.NoError(t, MinibridgeConfig{Rounds: 8}.Validate())
	assert.Error(t, MinibridgeConfig{Rounds: MinibridgeRoundsMin - 1}.Validate())
	assert.Error(t, MinibridgeConfig{Rounds: MinibridgeRoundsMax + 1}.Validate())
	// **4 の倍数でないと親が一巡しない。**
	assert.Error(t, MinibridgeConfig{Rounds: 6}.Validate())
}

func TestMinibridge_JSONRoundTrip(t *testing.T) {
	m := newTestMinibridge(t)
	decl := m.GetDeclarerIdx()
	require.NoError(t, m.SelectContractForTest(decl, 3, CardDesignHeart))
	for range 3 {
		idx := m.GetCurrentPlayerIdx()
		require.NoError(t, m.PlayForTest(idx, m.CpuChoiceForTest(idx)))
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	var restored Minibridge
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, m.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, m.GetContractLevel(), restored.GetContractLevel())
	assert.Equal(t, m.GetContractSuit(), restored.GetContractSuit())
	assert.Equal(t, m.GetDeclarerIdx(), restored.GetDeclarerIdx())
	assert.Equal(t, m.GetDummyIdx(), restored.GetDummyIdx())
	for i := range MinibridgePlayerCnt {
		assert.Equal(t, m.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
		assert.Equal(t, m.GetPlayer(i).GetHcp(), restored.GetPlayer(i).GetHcp(), "HCP が消えない")
		assert.Equal(t, m.GetPlayer(i).GetTeam(), restored.GetPlayer(i).GetTeam())
	}
	assert.NotEmpty(t, restored.GetDummyHand(), "契約後はダミーが公開されたまま")
}

// **壊れたスナップショットは弾く。**
//
// このコーデックは 6 PR 連続で「個々のフィールドは範囲内だが組み合わせが
// あり得ない」を通していたので、フェーズ別のスナップショットを作って変異させる
// （#5312 は Draw のスナップショットしか作らず、Play 固有の穴に当たらなかった）。
func TestMinibridge_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	snapshot := func(t *testing.T, playing bool) map[string]any {
		t.Helper()
		m := newTestMinibridge(t)
		if playing {
			require.NoError(t, m.SelectContractForTest(m.GetDeclarerIdx(), 3, CardDesignHeart))
			require.NoError(t, m.PlayForTest(m.GetCurrentPlayerIdx(), 0))
		}
		data, err := json.Marshal(m)
		require.NoError(t, err)
		var out map[string]any
		require.NoError(t, json.Unmarshal(data, &out))
		return out
	}

	for _, tc := range []struct {
		name    string
		playing bool
		mutate  func(map[string]any)
	}{
		{"phase out of range", false, func(m map[string]any) { m["ph"] = 9 }},
		{"contract suit out of range", true, func(m map[string]any) { m["cs"] = 9 }},
		{"contract suit without a level", false, func(m map[string]any) { m["cs"] = 3 }},
		{"contract level out of range", true, func(m map[string]any) { m["cl"] = 99 }},
		// **契約フェーズは未決定、それ以降は決定済み（#5312 で踏んだ形）。**
		{"contract already chosen in the contract phase", false, func(m map[string]any) { m["cl"] = 3; m["cs"] = 3 }},
		{"play phase without a contract", true, func(m map[string]any) { m["cl"] = 0; m["cs"] = 0 }},
		{"declarer out of range", true, func(m map[string]any) { m["di"] = 9 }},
		{"no declarer at all", true, func(m map[string]any) { m["di"] = -1; m["dm"] = -1 }},
		{"dummy is not the partner", true, func(m map[string]any) { m["dm"] = (int(m["di"].(float64)) + 1) % 4 }},
		{"declarer and dummy disagree", true, func(m map[string]any) { m["dm"] = -1 }},
		{"current player out of range", true, func(m map[string]any) { m["ci"] = 9 }},
		{"dealer out of range", true, func(m map[string]any) { m["dl"] = -1 }},
		{"round number below one", true, func(m map[string]any) { m["rn"] = 0 }},
		{"round number above the configured rounds", true, func(m map[string]any) { m["rn"] = 99 }},
		{"negative trick number", true, func(m map[string]any) { m["tn"] = -1 }},
		{"winner before the game ended", true, func(m map[string]any) { m["wt"] = 1 }},
		// **終了フラグとフェーズは対。** 片方だけ立つと投了でも復旧できない
		// 恒久デッドロックになる（レビュー指摘 PR #5313）。
		{"game end flag without the game end phase", true, func(m map[string]any) { m["ge"] = true }},
		{"game end phase without the flag", true, func(m map[string]any) { m["ph"] = 3 }},
		{"winner team out of range", true, func(m map[string]any) { m["ge"] = true; m["wt"] = 9 }},
		{"config out of range", true, func(m map[string]any) { m["cf"] = map[string]any{"r": 6} }},
		// **場札は枚数だけでなく中身も見る（#5310 の再発防止）。**
		{"a trick entry with no card", true, func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 0}}
		}},
		{"a trick entry with a bad seat", true, func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 9, "card": map[string]any{"d": 1, "v": 9, "j": false}}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := snapshot(t, tc.playing)
			tc.mutate(s)
			data, err := json.Marshal(s)
			require.NoError(t, err)
			var restored Minibridge
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通り、使っても落ちない。**
	for _, playing := range []bool{false, true} {
		data, err := json.Marshal(snapshot(t, playing))
		require.NoError(t, err)
		var ok Minibridge
		require.NoError(t, json.Unmarshal(data, &ok))
		assert.NotPanics(t, func() {
			_ = ok.GetValidPlayIndices(ok.GetCurrentPlayerIdx())
			_ = ok.TrickWinnerForTest()
			_ = ok.GetDummyHand()
		})
	}
}

// **HCP の総和は席ごとの範囲検査を素通りする。** 崩れていると落札者の決定が別物になる。
func TestMinibridge_UnmarshalRejectsAnImpossibleHcpTotal(t *testing.T) {
	m := newTestMinibridge(t)
	data, err := json.Marshal(m)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	players := raw["pl"].([]any)
	// 席ごとには 0..40 の正当な値でも、合計が 40 でなければあり得ない。
	for _, p := range players {
		p.(map[string]any)["h"] = 1
	}
	bad, err := json.Marshal(raw)
	require.NoError(t, err)
	var restored Minibridge
	assert.Error(t, json.Unmarshal(bad, &restored))

	// 負のコントロール: 合計 40 なら通る。
	players[0].(map[string]any)["h"] = 37
	good, err := json.Marshal(raw)
	require.NoError(t, err)
	var ok Minibridge
	assert.NoError(t, json.Unmarshal(good, &ok))
}

// **席単体の HCP も検証する。**
func TestMinibridgePlayer_UnmarshalRejectsBrokenFields(t *testing.T) {
	for _, body := range []string{`{"h":-1}`, `{"h":41}`, `{"t":-1}`, `{"t":2}`} {
		var p MinibridgePlayer
		assert.Error(t, json.Unmarshal([]byte(body), &p), body)
	}
	for _, body := range []string{`{"h":0,"t":0}`, `{"h":40,"t":1}`, `{"h":10,"t":0}`} {
		var p MinibridgePlayer
		assert.NoError(t, json.Unmarshal([]byte(body), &p), body)
	}
}

// **終了フラグとフェーズが割れた盤面は、通すと二度と動かせない。**
// すべての入口が `gameEndFlag` で早期 return する一方、フェーズは終局ではないので
// 何も進まず、`GiveUp` でも復旧できない（レビュー指摘 PR #5313）。
func TestMinibridge_UnmarshalRejectsADeadlockedSnapshot(t *testing.T) {
	m := newTestMinibridge(t)
	data, err := json.Marshal(m)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["ge"] = true // フェーズは Contract のまま

	bad, err := json.Marshal(raw)
	require.NoError(t, err)
	var restored Minibridge
	assert.Error(t, json.Unmarshal(bad, &restored))

	// **負のコントロール: 対で立っていれば通り、終局として扱える。**
	m.FinishGameForTest()
	good, err := json.Marshal(m)
	require.NoError(t, err)
	var ok Minibridge
	require.NoError(t, json.Unmarshal(good, &ok))
	assert.True(t, ok.GetGameEndFlag())
	assert.Equal(t, MinibridgePhaseGameEnd, ok.GetPhase())
}

// **勝てないときに最強札を投げない。** リードのスートが無いと合法手は手札全部に
// なるので、素朴に最強を選ぶと切り札でもないエースを捨て札にする
// （レビュー指摘 PR #5313）。
func TestMinibridge_CpuDiscardsLowWhenItCannotWin(t *testing.T) {
	m := newTestMinibridge(t)
	m.SetContractForTest(0, 1, CardDesignHeart)
	m.SetPhaseForTest(MinibridgePhasePlay)
	m.SetLeadPlayerIdxForTest(0)
	m.SetCurrentPlayerIdxForTest(1)
	// ♠ がリードされていて、席 1 は ♠ も切り札 ♥ も持っていない。
	m.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)},
	})
	minibridgeHandOf(m, 1,
		NewCard(CardDesignDiamond, 1, false), // ♦A（最強だが勝てない）
		NewCard(CardDesignClover, 2, false),  // ♣2（いちばん安い）
	)
	assert.Equal(t, 1, m.CpuChoiceForTest(1), "勝てないので安い札を捨てる")

	// **負のコントロール: 取れるなら取りにいく。**
	minibridgeHandOf(m, 1,
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 1, false), // ♠A で ♠K に勝てる
	)
	assert.Equal(t, 1, m.CpuChoiceForTest(1))

	// 切り札で取れるならそれを使う。
	minibridgeHandOf(m, 1,
		NewCard(CardDesignDiamond, 1, false),
		NewCard(CardDesignHeart, 2, false), // 切り札の ♥2
	)
	assert.Equal(t, 1, m.CpuChoiceForTest(1), "切り札なら 2 でも取れる")

	// リードする番なら「勝てる」扱いで強く出る。
	m.SetCurrentTrickForTest(nil)
	minibridgeHandOf(m, 1, NewCard(CardDesignClover, 3, false), NewCard(CardDesignClover, 1, false))
	assert.Equal(t, 1, m.CpuChoiceForTest(1), "リードは強い札から")
}
