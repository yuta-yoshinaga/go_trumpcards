//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestEstimation(t *testing.T) *Estimation {
	t.Helper()
	e := NewDefaultEstimation()
	e.Reset()
	return e
}

// estimationHandOf は指定プレイヤーの手札を固定の並びに差し替える。
func estimationHandOf(e *Estimation, idx int, cards ...*Card) {
	p := e.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// settleTrumpAndBids は切り札選択と宣言を最後まで進める。
func settleTrumpAndBids(t *testing.T, e *Estimation) {
	t.Helper()
	if e.IsHumanTrumpTurn() {
		require.NoError(t, e.SelectTrump(CardDesignSpade))
	} else {
		e.CpuSelectTrump()
	}
	for range EstimationPlayerCnt * 2 {
		if e.GetPhase() != EstimationPhaseBid {
			break
		}
		if e.IsHumanBidTurn() {
			bid := 3
			if r := e.GetRestrictedBid(); r == bid {
				bid = 2
			}
			require.NoError(t, e.PlayerBid(bid))
			continue
		}
		e.CpuBid()
	}
	require.Equal(t, EstimationPhasePlay, e.GetPhase())
}

// --- 配り ---

// 52 枚を 13 枚ずつ、ちょうど配り切る。
func TestEstimation_DealsThirteenEach(t *testing.T) {
	e := newTestEstimation(t)

	assert.Equal(t, EstimationPhaseTrump, e.GetPhase())
	total := 0
	for i := range EstimationPlayerCnt {
		assert.Equal(t, EstimationHandSize, e.GetPlayer(i).GetCardsSize(), "player %d", i)
		total += e.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)
}

// ジョーカーは使わない。
func TestEstimation_NoJokers(t *testing.T) {
	e := newTestEstimation(t)
	for i := range EstimationPlayerCnt {
		p := e.GetPlayer(i)
		for j := range p.GetCardsSize() {
			assert.NotEqual(t, CardDesignJoker, p.GetCard(j).GetDesign())
		}
	}
}

// --- 切り札の選択 ---

// 親が人間なら選ばせ、CPU なら自動で決まる。両側を踏む。
func TestEstimation_TrumpSelection(t *testing.T) {
	human := newTestEstimation(t)
	human.SetDealerIdxForTest(0)
	require.True(t, human.IsHumanTrumpTurn())
	require.NoError(t, human.SelectTrump(CardDesignHeart))
	assert.Equal(t, CardDesignHeart, human.GetTrumpSuit())
	assert.Equal(t, EstimationPhaseBid, human.GetPhase())

	cpu := newTestEstimation(t)
	cpu.SetDealerIdxForTest(2)
	assert.False(t, cpu.IsHumanTrumpTurn())
	cpu.CpuSelectTrump()
	assert.Equal(t, EstimationPhaseBid, cpu.GetPhase())
	assert.GreaterOrEqual(t, cpu.GetTrumpSuit(), CardDesignSpade)
}

// 存在しないスートは拒否する。
func TestEstimation_SelectTrumpRejectsInvalidSuit(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	assert.Error(t, e.SelectTrump(0))
	assert.Error(t, e.SelectTrump(99))
	assert.Equal(t, EstimationPhaseTrump, e.GetPhase())
}

// 親でない／フェーズ違い／終局後は選べない。
func TestEstimation_SelectTrumpGuards(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(1)
	assert.Error(t, e.SelectTrump(CardDesignSpade), "親でなければ選べない")

	e.SetDealerIdxForTest(0)
	e.SetPhaseForTest(EstimationPhaseBid)
	assert.Error(t, e.SelectTrump(CardDesignSpade))

	e.SetPhaseForTest(EstimationPhaseTrump)
	e.GiveUp()
	assert.Error(t, e.SelectTrump(CardDesignSpade))
}

// CPU は人間が親のときに勝手に決めない。
func TestEstimation_CpuSelectTrumpIgnoresHumanDealer(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	e.CpuSelectTrump()
	assert.Equal(t, EstimationPhaseTrump, e.GetPhase())
	assert.Equal(t, 0, e.GetTrumpSuit())
}

// --- 宣言 ---

// 宣言は親から時計回りに 4 人分。
func TestEstimation_BiddingGoesRoundOnce(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(1)
	e.CpuSelectTrump()

	for range EstimationPlayerCnt {
		if e.IsHumanBidTurn() {
			require.NoError(t, e.PlayerBid(2))
			continue
		}
		e.CpuBid()
	}
	assert.Equal(t, EstimationPhasePlay, e.GetPhase())
	for i := range EstimationPlayerCnt {
		assert.GreaterOrEqual(t, e.GetPlayer(i).GetBid(), 0, "player %d", i)
	}
}

// **最後の宣言者は合計を 13 ちょうどにできない。**
func TestEstimation_LastBidderCannotMakeTheTotalThirteen(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(1)
	e.CpuSelectTrump()
	// 人間 (0) が最後になるよう、他の 3 人を先に埋める。
	e.SetBidsForTest(map[int]int{1: 4, 2: 4, 3: 4})
	e.SetBidPlayerIdxForTest(0)

	assert.Equal(t, 1, e.GetRestrictedBid(), "4+4+4=12 なので 1 が禁止")
	require.Error(t, e.PlayerBid(1))
	assert.Equal(t, -1, e.GetPlayer(0).GetBid(), "拒否されたので未宣言のまま")
	require.NoError(t, e.PlayerBid(2))
	assert.Equal(t, 2, e.GetPlayer(0).GetBid())
}

// 制限は最後の 1 人にだけ掛かる。
func TestEstimation_RestrictedBidOnlyAppliesToTheLastBidder(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(1)
	e.CpuSelectTrump()
	assert.Equal(t, -1, e.GetRestrictedBid(), "まだ 4 人残っている")

	e.SetBidsForTest(map[int]int{1: 4, 2: 4})
	assert.Equal(t, -1, e.GetRestrictedBid(), "まだ 2 人残っている")

	e.SetBidsForTest(map[int]int{1: 4, 2: 4, 3: 4})
	assert.Equal(t, 1, e.GetRestrictedBid())
}

// **CPU も制限を守る。** 見積もりが禁止値なら 1 つずらす。
func TestEstimation_CpuAvoidsTheRestrictedBid(t *testing.T) {
	for range 30 {
		e := newTestEstimation(t)
		e.SetDealerIdxForTest(1)
		e.CpuSelectTrump()
		for e.GetPhase() == EstimationPhaseBid {
			if e.IsHumanBidTurn() {
				bid := 3
				if r := e.GetRestrictedBid(); r == bid {
					bid = 2
				}
				require.NoError(t, e.PlayerBid(bid))
				continue
			}
			e.CpuBid()
		}
		total := 0
		for i := range EstimationPlayerCnt {
			total += e.GetPlayer(i).GetBid()
		}
		assert.NotEqual(t, EstimationHandSize, total, "宣言の合計が 13 になってはいけない")
	}
}

// 範囲外の宣言は拒否する。
func TestEstimation_BidRejectsOutOfRange(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignSpade))

	assert.Error(t, e.PlayerBid(-1))
	assert.Error(t, e.PlayerBid(EstimationHandSize+1))
	assert.NoError(t, e.PlayerBid(EstimationHandSize), "13 そのものは合法")
}

// 自分の番でない／フェーズ違い／終局後は宣言できない。
func TestEstimation_BidGuards(t *testing.T) {
	e := newTestEstimation(t)
	assert.Error(t, e.PlayerBid(3), "切り札選択フェーズでは宣言できない")

	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignSpade))
	e.SetBidPlayerIdxForTest(2)
	assert.Error(t, e.PlayerBid(3))

	e.SetBidPlayerIdxForTest(0)
	e.GiveUp()
	assert.Error(t, e.PlayerBid(3))
}

// CPU は人間の番に宣言しない。
func TestEstimation_CpuBidIgnoresHumanTurn(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignSpade))
	e.SetBidPlayerIdxForTest(0)

	e.CpuBid()
	assert.Equal(t, -1, e.GetPlayer(0).GetBid())
}

// --- 宣言の種類 ---

// 0 を宣言すると Dash Call になる。
func TestEstimation_ZeroBidIsADashCall(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignSpade))
	require.NoError(t, e.PlayerBid(0))

	assert.Equal(t, EstimationCallDash, e.GetPlayer(0).GetCallType())
}

// **Risk は最高宣言者に付く。** 同値なら親から時計回りで先の人。
func TestEstimation_RiskGoesToTheHighestBidder(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignSpade))
	e.SetBidsForTest(map[int]int{0: 2, 1: 5, 2: 3})
	e.SetBidPlayerIdxForTest(3)
	e.CloseBiddingForTest()

	assert.Equal(t, EstimationCallRisk, e.GetPlayer(1).GetCallType())
	for _, i := range []int{0, 2, 3} {
		assert.NotEqual(t, EstimationCallRisk, e.GetPlayer(i).GetCallType(), "player %d", i)
	}
}

// 同値なら親に近いほうが Risk を取る。
func TestEstimation_RiskTieGoesToTheEarlierBidder(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(2)
	e.CpuSelectTrump()
	e.SetBidsForTest(map[int]int{0: 5, 1: 1, 2: 5, 3: 1})
	e.CloseBiddingForTest()

	// 親は 2 なので、2 → 3 → 0 → 1 の順。同値 5 の先着は 2。
	assert.Equal(t, EstimationCallRisk, e.GetPlayer(2).GetCallType())
	assert.NotEqual(t, EstimationCallRisk, e.GetPlayer(0).GetCallType())
}

// **全員が Dash なら Risk は誰にも付かない。** 0 を「最高宣言」にしない。
func TestEstimation_NoRiskWhenEveryoneDashes(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignSpade))
	e.SetBidsForTest(map[int]int{0: 0, 1: 0, 2: 0, 3: 0})
	e.CloseBiddingForTest()

	for i := range EstimationPlayerCnt {
		assert.NotEqual(t, EstimationCallRisk, e.GetPlayer(i).GetCallType(), "player %d", i)
	}
}

// --- 得点 ---

// **過不足の量では変わらない。** 1 つ足りなくても 5 つ多くても同じ減点。
func TestEstimation_ScoreIsExactOrNothing(t *testing.T) {
	for _, tc := range []struct {
		name     string
		bid, got int
		call     EstimationCallType
		want     int
	}{
		{"normal exact", 4, 4, EstimationCallNormal, 14},
		{"normal one short", 4, 3, EstimationCallNormal, -14},
		{"normal five over", 4, 9, EstimationCallNormal, -14},
		{"dash exact", 0, 0, EstimationCallDash, EstimationDashScore},
		{"dash missed", 0, 1, EstimationCallDash, -EstimationDashScore},
		{"risk exact", 6, 6, EstimationCallRisk, 32},
		{"risk missed", 6, 5, EstimationCallRisk, -32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, EstimationScoreFor(tc.bid, tc.got, tc.call))
		})
	}
}

// **Dash と Risk は通常より大きく振れる。** 同じ宣言数どうしで比べる。
func TestEstimation_SpecialCallsSwingWider(t *testing.T) {
	normal := EstimationScoreFor(0, 0, EstimationCallNormal)
	dash := EstimationScoreFor(0, 0, EstimationCallDash)
	assert.Greater(t, dash, normal, "0 宣言の的中は Dash のほうが大きい")

	plain := EstimationScoreFor(5, 5, EstimationCallNormal)
	risk := EstimationScoreFor(5, 5, EstimationCallRisk)
	assert.Equal(t, plain*2, risk)
}

// 1 ラウンド通して、各自の増減が得点表どおりに乗る。
func TestEstimation_RoundScoringMatchesTheTable(t *testing.T) {
	e := newTestEstimation(t)
	settleTrumpAndBids(t, e)

	for e.GetPhase() == EstimationPhasePlay {
		if e.IsHumanTurn() {
			valid := e.GetValidPlayIndices(0)
			require.NotEmpty(t, valid)
			require.NoError(t, e.PlayerPlay(valid[0]))
			continue
		}
		e.CpuPlay()
	}

	assert.Equal(t, EstimationTricksPerRound, e.GetTrickNumber())
	tricks := 0
	for i := range EstimationPlayerCnt {
		p := e.GetPlayer(i)
		tricks += p.GetTrickCount()
		assert.Equal(t,
			EstimationScoreFor(p.GetBid(), p.GetTrickCount(), p.GetCallType()),
			p.GetRoundScore(), "player %d", i)
		assert.Equal(t, p.GetRoundScore(), p.GetTotalScore(), "1 ラウンド目は累計＝増減")
	}
	// **トリックは 13 個ちょうどが 4 人に分配される。** どこにも消えない。
	assert.Equal(t, EstimationTricksPerRound, tricks)
}

// 規定ラウンドに届かなければ次へ、届けば終局。
func TestEstimation_NextRoundAndGameEnd(t *testing.T) {
	e := newTestEstimation(t)
	e.SetConfig(EstimationConfig{Rounds: 2})
	settleTrumpAndBids(t, e)
	for e.GetPhase() == EstimationPhasePlay {
		if e.IsHumanTurn() {
			require.NoError(t, e.PlayerPlay(e.GetValidPlayIndices(0)[0]))
			continue
		}
		e.CpuPlay()
	}
	require.Equal(t, EstimationPhaseRoundEnd, e.GetPhase())

	e.NextRound()
	assert.Equal(t, 2, e.GetRoundNumber())
	assert.Equal(t, 1, e.GetDealerIdx(), "親は時計回りに動く")
	assert.Equal(t, EstimationPhaseTrump, e.GetPhase())
	assert.Equal(t, 0, e.GetTrumpSuit(), "切り札はラウンドごとに決め直す")
	for i := range EstimationPlayerCnt {
		assert.Equal(t, -1, e.GetPlayer(i).GetBid(), "宣言もラウンドごとに白紙")
	}
}

// 最終ラウンドを終えると終局し、最高得点者が勝つ。
func TestEstimation_GameEndsAfterTheConfiguredRounds(t *testing.T) {
	e := newTestEstimation(t)
	e.SetConfig(EstimationConfig{Rounds: 1})
	settleTrumpAndBids(t, e)
	for e.GetPhase() == EstimationPhasePlay {
		if e.IsHumanTurn() {
			require.NoError(t, e.PlayerPlay(e.GetValidPlayIndices(0)[0]))
			continue
		}
		e.CpuPlay()
	}

	assert.True(t, e.GetGameEndFlag())
	assert.Equal(t, EstimationPhaseGameEnd, e.GetPhase())
	if w := e.GetWinnerIdx(); w >= 0 {
		for i := range EstimationPlayerCnt {
			assert.GreaterOrEqual(t, e.GetPlayer(w).GetTotalScore(), e.GetPlayer(i).GetTotalScore())
		}
	}

	before := e.GetRoundNumber()
	e.NextRound()
	assert.Equal(t, before, e.GetRoundNumber(), "終局後は進まない")
}

// 同点なら勝者なし。
func TestEstimation_TieHasNoWinner(t *testing.T) {
	e := newTestEstimation(t)
	for i := range EstimationPlayerCnt {
		e.GetPlayer(i).SetTotalScore(20)
	}
	e.FinishGameForTest()
	assert.Equal(t, -1, e.GetWinnerIdx())
}

// --- プレイ ---

// リードのスートを持っていれば必ずフォローする。
func TestEstimation_MustFollowSuit(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignHeart))
	e.SetPhaseForTest(EstimationPhasePlay)
	e.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}})
	e.SetCurrentPlayerIdxForTest(0)
	estimationHandOf(e, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignSpade, 7, false))

	require.Error(t, e.PlayerPlay(0))
	require.NoError(t, e.PlayerPlay(1))
}

// リードのスートが無ければ何を出してもよい。
func TestEstimation_MayDiscardWhenVoid(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignHeart))
	e.SetPhaseForTest(EstimationPhasePlay)
	e.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}})
	e.SetCurrentPlayerIdxForTest(0)
	estimationHandOf(e, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignClover, 7, false))

	assert.Equal(t, []int{0, 1}, e.GetValidPlayIndices(0))
}

// 切り札はリードのスートに勝つ。
func TestEstimation_TrumpBeatsTheLeadSuit(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignHeart))
	e.SetPhaseForTest(EstimationPhasePlay)
	e.SetCurrentPlayerIdxForTest(0)
	e.SetLeadPlayerIdxForTest(0)

	estimationHandOf(e, 0, NewCard(CardDesignSpade, 1, false))
	estimationHandOf(e, 1, NewCard(CardDesignSpade, 13, false))
	estimationHandOf(e, 2, NewCard(CardDesignHeart, 2, false))
	estimationHandOf(e, 3, NewCard(CardDesignSpade, 12, false))

	for i := range EstimationPlayerCnt {
		require.NoError(t, e.PlayForTest(i, 0))
	}
	assert.Equal(t, 1, e.GetPlayer(2).GetTrickCount(), "切り札の 2 が A に勝つ")
}

// 範囲外のインデックスは拒否する。
func TestEstimation_PlayRejectsInvalidIndex(t *testing.T) {
	e := newTestEstimation(t)
	settleTrumpAndBids(t, e)
	e.SetCurrentPlayerIdxForTest(0)

	assert.Error(t, e.PlayerPlay(-1))
	assert.Error(t, e.PlayerPlay(99))
}

// 自分の番でない／フェーズ違い／終局後は出せない。
func TestEstimation_PlayGuards(t *testing.T) {
	e := newTestEstimation(t)
	assert.Error(t, e.PlayerPlay(0), "切り札選択フェーズでは出せない")

	settleTrumpAndBids(t, e)
	e.SetCurrentPlayerIdxForTest(1)
	assert.Error(t, e.PlayerPlay(0))

	e.SetCurrentPlayerIdxForTest(0)
	e.GiveUp()
	assert.Error(t, e.PlayerPlay(0))
}

// 範囲外プレイヤーの合法手は nil。
func TestEstimation_ValidIndicesOutOfRange(t *testing.T) {
	e := newTestEstimation(t)
	assert.Nil(t, e.GetValidPlayIndices(-1))
	assert.Nil(t, e.GetValidPlayIndices(EstimationPlayerCnt))
}

// CPU は人間の番に打たない。
func TestEstimation_CpuPlayIgnoresHumanTurn(t *testing.T) {
	e := newTestEstimation(t)
	settleTrumpAndBids(t, e)
	e.SetCurrentPlayerIdxForTest(0)
	size := e.GetPlayer(0).GetCardsSize()

	e.CpuPlay()
	assert.Equal(t, size, e.GetPlayer(0).GetCardsSize())
}

// --- CPU ---

// **宣言に足りなければ取りに行く。** 取れる札のうち一番弱いもの。
func TestEstimation_CpuTakesTrickCheaplyWhenShort(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignHeart))
	e.SetPhaseForTest(EstimationPhasePlay)
	e.GetPlayer(2).SetBid(3)
	e.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 12, false)}})
	e.SetCurrentPlayerIdxForTest(2)
	estimationHandOf(e, 2,
		NewCard(CardDesignSpade, 1, false),  // 勝てるが強すぎる
		NewCard(CardDesignSpade, 13, false), // 勝てて一番弱い
		NewCard(CardDesignSpade, 7, false))  // 勝てない

	assert.Equal(t, 1, e.CpuChoiceForTest(2))
}

// **宣言に足りていれば取らない。** 勝たない札のうち一番強いものを吐く。
func TestEstimation_CpuDucksWhenItHasEnough(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignHeart))
	e.SetPhaseForTest(EstimationPhasePlay)
	e.GetPlayer(2).SetBid(0)
	e.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 1, false)}})
	e.SetCurrentPlayerIdxForTest(2)
	estimationHandOf(e, 2,
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 13, false)) // 勝たないなかで一番強い

	assert.Equal(t, 1, e.CpuChoiceForTest(2))
}

// 取りたいときのリードは最強札から。
func TestEstimation_CpuLeadsHighWhenShort(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignHeart))
	e.SetPhaseForTest(EstimationPhasePlay)
	e.GetPlayer(1).SetBid(5)
	e.SetCurrentTrickForTest(nil)
	e.SetCurrentPlayerIdxForTest(1)
	estimationHandOf(e, 1,
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 11, false))

	assert.Equal(t, 1, e.CpuChoiceForTest(1))
}

// --- ヒント ---

// 切り札選択中のヒントはスートを勧める。
func TestEstimation_HintDuringTrumpSelection(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)

	h := e.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex)
	assert.Equal(t, "estimationSelectTrump", h.Reason)
	assert.GreaterOrEqual(t, h.Value, CardDesignSpade)
}

// 宣言中のヒントは数を勧める。
func TestEstimation_HintDuringBidding(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignSpade))

	h := e.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex)
	assert.Contains(t, []string{"estimationBid", "estimationDashCall"}, h.Reason)
	assert.GreaterOrEqual(t, h.Value, 0)
}

// **禁止値に当たったときは別の数を勧める。** 押せない宣言を勧めない。
func TestEstimation_HintAvoidsTheRestrictedBid(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(1)
	e.CpuSelectTrump()
	e.SetBidPlayerIdxForTest(0)
	// 人間の見積もりを取り、それを禁止値にする配置を作る。
	want := e.EstimateForTest(0)
	e.SetBidsForTest(map[int]int{1: EstimationHandSize - want, 2: 0, 3: 0})

	h := e.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "estimationAvoidRestricted", h.Reason)
	assert.NotEqual(t, e.GetRestrictedBid(), h.Value)
}

// プレイ中のヒントは合法な札を指す。
func TestEstimation_HintDuringPlay(t *testing.T) {
	e := newTestEstimation(t)
	settleTrumpAndBids(t, e)
	e.SetCurrentPlayerIdxForTest(0)

	h := e.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Contains(t, e.GetValidPlayIndices(0), *h.CardIndex)
	assert.Contains(t, []string{"estimationWinTrick", "estimationDuck"}, h.Reason)
}

// 自分の番でなければヒントは無い。
func TestEstimation_HintNilWhenNotHumanTurn(t *testing.T) {
	e := newTestEstimation(t)
	settleTrumpAndBids(t, e)
	e.SetCurrentPlayerIdxForTest(2)
	assert.Nil(t, e.GetHint())

	e.GiveUp()
	assert.Nil(t, e.GetHint())
}

// --- その他 ---

func TestEstimation_GiveUp(t *testing.T) {
	e := newTestEstimation(t)
	e.GiveUp()
	assert.True(t, e.GetGameEndFlag())
	assert.Equal(t, 1, e.GetWinnerIdx())

	e.SetWinnerIdxForTest(0)
	e.GiveUp()
	assert.Equal(t, 0, e.GetWinnerIdx(), "二度目は何も起きない")
}

func TestEstimation_AccessorsOutOfRange(t *testing.T) {
	e := newTestEstimation(t)
	assert.Nil(t, e.GetPlayer(-1))
	assert.Nil(t, e.GetPlayer(EstimationPlayerCnt))
	assert.Equal(t, EstimationPlayerCnt, e.GetPlayerCnt())
	assert.Equal(t, 0, e.GetTrickNumber())
	assert.Equal(t, 1, e.GetRoundNumber())
	assert.Empty(t, e.GetCurrentTrick())
	assert.GreaterOrEqual(t, e.GetLeadPlayerIdx(), 0)
	assert.GreaterOrEqual(t, e.GetBidPlayerIdx(), 0)
}

func TestEstimationConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultEstimationConfig().Validate())
	assert.NoError(t, EstimationConfig{Rounds: EstimationRoundsMin}.Validate())
	assert.NoError(t, EstimationConfig{Rounds: EstimationRoundsMax}.Validate())
	assert.Error(t, EstimationConfig{Rounds: EstimationRoundsMin - 1}.Validate())
	assert.Error(t, EstimationConfig{Rounds: EstimationRoundsMax + 1}.Validate())
}

// --- JSON 往復 ---

// **Worker はリクエストごとに KV から作り直す。** 宣言・種別・累計が往復しないと
// 得点が消える。
func TestEstimation_JSONRoundTrip(t *testing.T) {
	e := newTestEstimation(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignDiamond))
	require.NoError(t, e.PlayerBid(4))
	e.GetPlayer(0).SetTotalScore(37)
	e.GetPlayer(1).SetCallType(EstimationCallRisk)

	data, err := json.Marshal(e)
	require.NoError(t, err)

	var got Estimation
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, e.GetPhase(), got.GetPhase())
	assert.Equal(t, CardDesignDiamond, got.GetTrumpSuit())
	assert.Equal(t, 4, got.GetPlayer(0).GetBid())
	assert.Equal(t, 37, got.GetPlayer(0).GetTotalScore())
	assert.Equal(t, EstimationCallRisk, got.GetPlayer(1).GetCallType())
	assert.Equal(t, e.GetBidPlayerIdx(), got.GetBidPlayerIdx())
	assert.Equal(t, e.GetDealerIdx(), got.GetDealerIdx())
	for i := range EstimationPlayerCnt {
		assert.Equal(t, e.GetPlayer(i).GetCardsSize(), got.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
}

// 壊れたスナップショットは復元しない。
func TestEstimation_UnmarshalRejectsInvalid(t *testing.T) {
	valid := func() estimationJSON {
		return estimationJSON{
			Config:      DefaultEstimationConfig(),
			Phase:       EstimationPhasePlay,
			RoundNumber: 1,
			WinnerIdx:   -1,
		}
	}
	cases := map[string]func(*estimationJSON){
		"bad config":  func(j *estimationJSON) { j.Config.Rounds = 0 },
		"bad phase":   func(j *estimationJSON) { j.Phase = EstimationPhase(99) },
		"bad trick":   func(j *estimationJSON) { j.TrickNumber = EstimationTricksPerRound + 1 },
		"bad round":   func(j *estimationJSON) { j.RoundNumber = 0 },
		"bad current": func(j *estimationJSON) { j.CurrentPlayerIdx = EstimationPlayerCnt },
		"bad bidder":  func(j *estimationJSON) { j.BidPlayerIdx = -1 },
		"bad lead":    func(j *estimationJSON) { j.LeadPlayerIdx = -1 },
		"bad dealer":  func(j *estimationJSON) { j.DealerIdx = EstimationPlayerCnt },
		"bad winner":  func(j *estimationJSON) { j.WinnerIdx = EstimationPlayerCnt },
		"long trick": func(j *estimationJSON) {
			j.CurrentTrick = make([]*TrickCard, EstimationPlayerCnt+1)
		},
		"long log": func(j *estimationJSON) {
			j.ActionLog = make([]*ActionLogEntry, estimationMaxSliceLen+1)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			j := valid()
			mutate(&j)
			data, err := json.Marshal(j)
			require.NoError(t, err)
			var got Estimation
			assert.Error(t, json.Unmarshal(data, &got))
		})
	}

	var got Estimation
	assert.Error(t, got.UnmarshalJSON([]byte("{")))

	// 正しいスナップショットは通る（ガードの負のコントロール）。
	data, err := json.Marshal(valid())
	require.NoError(t, err)
	assert.NoError(t, json.Unmarshal(data, &got))
}

func TestEstimation_ActionLog(t *testing.T) {
	e := newTestEstimation(t)
	require.NotEmpty(t, e.actionLog, "配りが記録される")
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(CardDesignSpade))
	require.NoError(t, e.PlayerBid(3))

	kinds := map[string]bool{}
	for _, entry := range e.actionLog {
		kinds[entry.ActionType] = true
	}
	assert.True(t, kinds["trump"])
	assert.True(t, kinds["bid"])
}
