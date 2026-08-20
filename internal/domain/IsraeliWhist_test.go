//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestIsraeliWhist(t *testing.T) *IsraeliWhist {
	t.Helper()
	w := NewDefaultIsraeliWhist()
	w.Reset()
	return w
}

// israeliWhistHandOf は指定プレイヤーの手札を固定の並びに差し替える。
func israeliWhistHandOf(w *IsraeliWhist, idx int, cards ...*Card) {
	p := w.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// settleAuctionAndBids は 2 段階の入札を最後まで進める。
func settleAuctionAndBids(t *testing.T, w *IsraeliWhist) {
	t.Helper()
	for range IsraeliWhistPlayerCnt * 3 {
		if w.GetPhase() != IsraeliWhistPhaseAuction {
			break
		}
		if w.IsHumanAuctionTurn() {
			if err := w.PlayerAuctionPass(); err != nil {
				require.NoError(t, w.PlayerAuctionBid(IsraeliWhistMinAuctionBid, CardDesignSpade))
			}
			continue
		}
		w.CpuAuction()
	}
	require.Equal(t, IsraeliWhistPhaseBid, w.GetPhase(), "auction never settled")

	for range IsraeliWhistPlayerCnt * 2 {
		if w.GetPhase() != IsraeliWhistPhaseBid {
			break
		}
		if w.IsHumanBidTurn() {
			bid := w.MinimumBidFor(0)
			if r := w.GetRestrictedBid(); r == bid {
				bid++
			}
			require.NoError(t, w.PlayerBid(bid))
			continue
		}
		w.CpuBid()
	}
	require.Equal(t, IsraeliWhistPhasePlay, w.GetPhase(), "bidding never settled")
}

// --- 配り ---

func TestIsraeliWhist_DealsThirteenEach(t *testing.T) {
	w := newTestIsraeliWhist(t)

	assert.Equal(t, IsraeliWhistPhaseAuction, w.GetPhase())
	total := 0
	for i := range IsraeliWhistPlayerCnt {
		assert.Equal(t, IsraeliWhistHandSize, w.GetPlayer(i).GetCardsSize(), "player %d", i)
		total += w.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)
	assert.Equal(t, 0, w.GetTrumpSuit(), "切り札はオークションで決まる")
}

// --- 1 段階目: オークション ---

// **入札のスート序列は ♣ < ♦ < ♥ < ♠。** CardDesign の並びとは違う。
func TestIsraeliWhist_SuitRankIsNotTheConstantOrder(t *testing.T) {
	order := []int{CardDesignClover, CardDesignDiamond, CardDesignHeart, CardDesignSpade}
	for i := 0; i < len(order)-1; i++ {
		assert.Less(t, israeliWhistSuitRank(order[i]), israeliWhistSuitRank(order[i+1]),
			"%d must rank below %d", order[i], order[i+1])
	}
	// 定数の並びをそのまま使うと ♦(4) が ♠(1) より強くなってしまう。
	assert.Greater(t, israeliWhistSuitRank(CardDesignSpade), israeliWhistSuitRank(CardDesignDiamond))
	assert.Equal(t, 0, israeliWhistSuitRank(CardDesignJoker))
}

// 競り上げは「数が上」または「同数でスートが上」。
func TestIsraeliWhist_OutbidRules(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetDeclarerForTest(1, 6, CardDesignHeart)

	assert.True(t, w.OutbidsForTest(7, CardDesignClover), "数が上なら勝つ")
	assert.True(t, w.OutbidsForTest(6, CardDesignSpade), "同数ならスートが上で勝つ")
	assert.False(t, w.OutbidsForTest(6, CardDesignDiamond), "同数でスートが下は負け")
	assert.False(t, w.OutbidsForTest(6, CardDesignHeart), "同じ入札は上回らない")
	assert.False(t, w.OutbidsForTest(5, CardDesignSpade), "数が下は負け")
}

// 最低入札を下回る／範囲外は拒否する。
func TestIsraeliWhist_AuctionRejectsOutOfRange(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetAuctionPlayerIdxForTest(0)

	assert.Error(t, w.PlayerAuctionBid(IsraeliWhistMinAuctionBid-1, CardDesignSpade))
	assert.Error(t, w.PlayerAuctionBid(IsraeliWhistHandSize+1, CardDesignSpade))
	assert.Error(t, w.PlayerAuctionBid(IsraeliWhistMinAuctionBid, 0))
	assert.Error(t, w.PlayerAuctionBid(IsraeliWhistMinAuctionBid, 99))
	assert.NoError(t, w.PlayerAuctionBid(IsraeliWhistMinAuctionBid, CardDesignSpade))
}

// 現在の最高入札を上回らない入札は拒否する。
func TestIsraeliWhist_AuctionRejectsWeakBid(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetAuctionPlayerIdxForTest(0)
	w.SetDeclarerForTest(2, 8, CardDesignSpade)

	assert.Error(t, w.PlayerAuctionBid(8, CardDesignHeart))
	assert.NoError(t, w.PlayerAuctionBid(9, CardDesignClover))
}

// 降りたら戻れない。
func TestIsraeliWhist_PassedPlayerCannotBidAgain(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetAuctionPlayerIdxForTest(0)
	w.SetDeclarerForTest(2, 6, CardDesignSpade)
	require.NoError(t, w.PlayerAuctionPass())
	assert.True(t, w.GetPlayer(0).GetPassed())

	w.SetAuctionPlayerIdxForTest(0)
	assert.Error(t, w.PlayerAuctionBid(9, CardDesignSpade))
}

// **誰も入札していないまま最後の 1 人になったら、降りられない。**
// 全員が降りると切り札が決まらずラウンドを始められない。
func TestIsraeliWhist_LastBidderStandingCannotPass(t *testing.T) {
	w := newTestIsraeliWhist(t)
	for _, i := range []int{1, 2, 3} {
		w.GetPlayer(i).SetPassed(true)
	}
	w.SetAuctionPlayerIdxForTest(0)

	err := w.PlayerAuctionPass()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must bid")
	assert.False(t, w.GetPlayer(0).GetPassed())
}

// 入札のある状態で生存者が 1 人になれば決着し、そのスートが切り札になる。
func TestIsraeliWhist_AuctionClosesOnLastSurvivor(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetAuctionPlayerIdxForTest(0)
	require.NoError(t, w.PlayerAuctionBid(7, CardDesignHeart))
	for _, i := range []int{1, 2, 3} {
		w.SetAuctionPlayerIdxForTest(i)
		w.CpuAuction()
		w.GetPlayer(i).SetPassed(true)
	}
	if w.GetPhase() == IsraeliWhistPhaseAuction {
		w.CloseAuctionForTest()
	}

	assert.Equal(t, IsraeliWhistPhaseBid, w.GetPhase())
	assert.Equal(t, CardDesignHeart, w.GetTrumpSuit())
	assert.Equal(t, 0, w.GetDeclarerIdx())
	assert.Equal(t, 7, w.GetHighBid())
}

// オークションは必ず決着する（誰も入札しなくても最後の席が引き受ける）。
func TestIsraeliWhist_AuctionAlwaysSettles(t *testing.T) {
	for range 30 {
		w := newTestIsraeliWhist(t)
		w.SetDealerIdxForTest(1)
		for range IsraeliWhistPlayerCnt * 3 {
			if w.GetPhase() != IsraeliWhistPhaseAuction {
				break
			}
			if w.IsHumanAuctionTurn() {
				if err := w.PlayerAuctionPass(); err != nil {
					require.NoError(t, w.PlayerAuctionBid(IsraeliWhistMinAuctionBid, CardDesignSpade))
				}
				continue
			}
			w.CpuAuction()
		}
		require.Equal(t, IsraeliWhistPhaseBid, w.GetPhase())
		assert.GreaterOrEqual(t, w.GetTrumpSuit(), CardDesignSpade)
		assert.GreaterOrEqual(t, w.GetHighBid(), IsraeliWhistMinAuctionBid)
		assert.GreaterOrEqual(t, w.GetDeclarerIdx(), 0)
	}
}

// --- 2 段階目: 宣言 ---

// **落札者はノルマ以上を宣言しなければならない。** 他の席には掛からない。
func TestIsraeliWhist_DeclarerMustMeetTheQuota(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetDeclarerForTest(0, 8, CardDesignSpade)
	w.CloseAuctionForTest()
	w.SetBidPlayerIdxForTest(0)

	assert.Equal(t, 8, w.MinimumBidFor(0))
	assert.Equal(t, 0, w.MinimumBidFor(1), "落札者以外に下限は無い")

	err := w.PlayerBid(7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 8")
	assert.Equal(t, -1, w.GetPlayer(0).GetBid())
	require.NoError(t, w.PlayerBid(8))
}

// **合計 13 は最後の宣言者が避ける。**
func TestIsraeliWhist_LastBidderCannotMakeTheTotalThirteen(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetDeclarerForTest(1, 5, CardDesignSpade)
	w.CloseAuctionForTest()
	w.SetBidsForTest(map[int]int{1: 5, 2: 4, 3: 3})
	w.SetBidPlayerIdxForTest(0)

	assert.Equal(t, 1, w.GetRestrictedBid(), "5+4+3=12 なので 1 が禁止")
	require.Error(t, w.PlayerBid(1))
	assert.Equal(t, -1, w.GetPlayer(0).GetBid())
	require.NoError(t, w.PlayerBid(2))
}

// 制限は最後の 1 人にだけ掛かる。
func TestIsraeliWhist_RestrictedBidOnlyAppliesToTheLastBidder(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetDeclarerForTest(1, 5, CardDesignSpade)
	w.CloseAuctionForTest()
	assert.Equal(t, -1, w.GetRestrictedBid())

	w.SetBidsForTest(map[int]int{1: 5, 2: 4})
	assert.Equal(t, -1, w.GetRestrictedBid())

	w.SetBidsForTest(map[int]int{1: 5, 2: 4, 3: 3})
	assert.Equal(t, 1, w.GetRestrictedBid())
}

// **4 人全員が宣言してからプレイに入る。** 3 人で締め切らない。
func TestIsraeliWhist_EveryoneBidsBeforePlay(t *testing.T) {
	w := newTestIsraeliWhist(t)
	settleAuctionAndBids(t, w)

	for i := range IsraeliWhistPlayerCnt {
		assert.GreaterOrEqual(t, w.GetPlayer(i).GetBid(), 0, "player %d never bid", i)
	}
	total := 0
	for i := range IsraeliWhistPlayerCnt {
		total += w.GetPlayer(i).GetBid()
	}
	assert.NotEqual(t, IsraeliWhistHandSize, total, "宣言の合計が 13 になってはいけない")
}

// 範囲外の宣言は拒否する。
func TestIsraeliWhist_BidRejectsOutOfRange(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetDeclarerForTest(1, 5, CardDesignSpade)
	w.CloseAuctionForTest()
	w.SetBidPlayerIdxForTest(0)

	assert.Error(t, w.PlayerBid(-1))
	assert.Error(t, w.PlayerBid(IsraeliWhistHandSize+1))
	assert.NoError(t, w.PlayerBid(0), "落札者でなければ 0 も宣言できる")
}

// 自分の番でない／フェーズ違い／終局後は操作できない。
func TestIsraeliWhist_Guards(t *testing.T) {
	w := newTestIsraeliWhist(t)
	assert.Error(t, w.PlayerBid(3), "オークション中は宣言できない")
	assert.Error(t, w.PlayerPlay(0), "オークション中は出せない")

	w.SetDeclarerForTest(1, 5, CardDesignSpade)
	w.CloseAuctionForTest()
	assert.Error(t, w.PlayerAuctionBid(9, CardDesignSpade), "宣言フェーズでは入札できない")
	w.SetBidPlayerIdxForTest(2)
	assert.Error(t, w.PlayerBid(3))

	w.SetBidPlayerIdxForTest(0)
	w.GiveUp()
	assert.Error(t, w.PlayerBid(3))
	assert.Error(t, w.PlayerAuctionPass())
	assert.Error(t, w.PlayerPlay(0))
}

// CPU は人間の番に動かない。
func TestIsraeliWhist_CpuIgnoresHumanTurn(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetAuctionPlayerIdxForTest(0)
	w.CpuAuction()
	assert.Equal(t, -1, w.GetPlayer(0).GetAuctionBid())

	w.SetDeclarerForTest(1, 5, CardDesignSpade)
	w.CloseAuctionForTest()
	w.SetBidPlayerIdxForTest(0)
	w.CpuBid()
	assert.Equal(t, -1, w.GetPlayer(0).GetBid())
}

// --- 得点 ---

// **的中は宣言の 2 乗、外しは過不足に比例。** 0 だけ別枠。
func TestIsraeliWhist_ScoreTable(t *testing.T) {
	for _, tc := range []struct {
		name     string
		bid, got int
		doubled  bool
		want     int
	}{
		{"exact 1", 1, 1, false, 11},
		{"exact 5", 5, 5, false, 35},
		{"exact 13", 13, 13, false, 179},
		{"exact zero", 0, 0, false, IsraeliWhistZeroScore},
		{"one short", 5, 4, false, -10},
		{"three over", 5, 8, false, -30},
		{"zero missed", 0, 2, false, -20},
		{"exact doubled", 5, 5, true, 70},
		{"miss doubled", 5, 4, true, -20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsraeliWhistScoreFor(tc.bid, tc.got, tc.doubled))
		})
	}
}

// **大きく宣言して当てるほど跳ねる。** 2 乗なので単調に増える。
func TestIsraeliWhist_BiggerExactCallsScoreMore(t *testing.T) {
	for n := 1; n < IsraeliWhistHandSize; n++ {
		assert.Greater(t,
			IsraeliWhistScoreFor(n+1, n+1, false),
			IsraeliWhistScoreFor(n, n, false), "bid %d", n)
	}
}

// **全員的中と全員外しはどちらも 2 倍。** 3 人だけなら等倍。
func TestIsraeliWhist_AllExactAndAllMissDouble(t *testing.T) {
	t.Run("all four exact", func(t *testing.T) {
		w := newTestIsraeliWhist(t)
		w.SetDeclarerForTest(0, 5, CardDesignSpade)
		w.CloseAuctionForTest()
		for i := range IsraeliWhistPlayerCnt {
			w.GetPlayer(i).SetBid(3)
			w.GetPlayer(i).AddTrick([]*Card{})
			w.GetPlayer(i).AddTrick([]*Card{})
			w.GetPlayer(i).AddTrick([]*Card{})
		}
		w.SetTrickNumberForTest(IsraeliWhistTricksPerRound - 1)
		w.SetCurrentTrickForTest(nil)
		w.FinishRoundForTest()

		for i := range IsraeliWhistPlayerCnt {
			assert.Equal(t, IsraeliWhistScoreFor(3, 3, true), w.GetPlayer(i).GetRoundScore(), "player %d", i)
		}
	})

	t.Run("all four miss", func(t *testing.T) {
		w := newTestIsraeliWhist(t)
		w.SetDeclarerForTest(0, 5, CardDesignSpade)
		w.CloseAuctionForTest()
		for i := range IsraeliWhistPlayerCnt {
			w.GetPlayer(i).SetBid(5)
		}
		w.FinishRoundForTest()

		for i := range IsraeliWhistPlayerCnt {
			assert.Equal(t, IsraeliWhistScoreFor(5, 0, true), w.GetPlayer(i).GetRoundScore(), "player %d", i)
		}
	})

	t.Run("a mixed round is not doubled", func(t *testing.T) {
		w := newTestIsraeliWhist(t)
		w.SetDeclarerForTest(0, 5, CardDesignSpade)
		w.CloseAuctionForTest()
		for i := range IsraeliWhistPlayerCnt {
			w.GetPlayer(i).SetBid(0)
		}
		w.GetPlayer(1).SetBid(4) // この 1 人だけ外す
		w.FinishRoundForTest()

		assert.Equal(t, IsraeliWhistScoreFor(0, 0, false), w.GetPlayer(0).GetRoundScore())
		assert.Equal(t, IsraeliWhistScoreFor(4, 0, false), w.GetPlayer(1).GetRoundScore())
	})
}

// 1 ラウンド通して、各自の増減が得点表どおりに乗る。
func TestIsraeliWhist_RoundScoringMatchesTheTable(t *testing.T) {
	w := newTestIsraeliWhist(t)
	settleAuctionAndBids(t, w)

	for w.GetPhase() == IsraeliWhistPhasePlay {
		if w.IsHumanTurn() {
			valid := w.GetValidPlayIndices(0)
			require.NotEmpty(t, valid)
			require.NoError(t, w.PlayerPlay(valid[0]))
			continue
		}
		w.CpuPlay()
	}

	assert.Equal(t, IsraeliWhistTricksPerRound, w.GetTrickNumber())
	exact, tricks := 0, 0
	for i := range IsraeliWhistPlayerCnt {
		p := w.GetPlayer(i)
		tricks += p.GetTrickCount()
		if p.GetTrickCount() == p.GetBid() {
			exact++
		}
	}
	// **トリックは 13 個ちょうどが 4 人に分配される。** どこにも消えない。
	assert.Equal(t, IsraeliWhistTricksPerRound, tricks)

	doubled := exact == IsraeliWhistPlayerCnt || exact == 0
	for i := range IsraeliWhistPlayerCnt {
		p := w.GetPlayer(i)
		assert.Equal(t, IsraeliWhistScoreFor(p.GetBid(), p.GetTrickCount(), doubled),
			p.GetRoundScore(), "player %d", i)
	}
}

// 規定ラウンドに届かなければ次へ、届けば終局。
func TestIsraeliWhist_NextRoundAndGameEnd(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetConfig(IsraeliWhistConfig{Rounds: 2})
	settleAuctionAndBids(t, w)
	for w.GetPhase() == IsraeliWhistPhasePlay {
		if w.IsHumanTurn() {
			require.NoError(t, w.PlayerPlay(w.GetValidPlayIndices(0)[0]))
			continue
		}
		w.CpuPlay()
	}
	require.Equal(t, IsraeliWhistPhaseRoundEnd, w.GetPhase())

	w.NextRound()
	assert.Equal(t, 2, w.GetRoundNumber())
	assert.Equal(t, 1, w.GetDealerIdx(), "親は時計回りに動く")
	assert.Equal(t, IsraeliWhistPhaseAuction, w.GetPhase())
	assert.Equal(t, 0, w.GetTrumpSuit(), "切り札はラウンドごとに競り直す")
	for i := range IsraeliWhistPlayerCnt {
		assert.Equal(t, -1, w.GetPlayer(i).GetBid(), "宣言もラウンドごとに白紙")
		assert.False(t, w.GetPlayer(i).GetPassed(), "降りた記録も白紙")
	}
}

func TestIsraeliWhist_GameEndsAfterTheConfiguredRounds(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetConfig(IsraeliWhistConfig{Rounds: 1})
	settleAuctionAndBids(t, w)
	for w.GetPhase() == IsraeliWhistPhasePlay {
		if w.IsHumanTurn() {
			require.NoError(t, w.PlayerPlay(w.GetValidPlayIndices(0)[0]))
			continue
		}
		w.CpuPlay()
	}

	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, IsraeliWhistPhaseGameEnd, w.GetPhase())
	if idx := w.GetWinnerIdx(); idx >= 0 {
		for i := range IsraeliWhistPlayerCnt {
			assert.GreaterOrEqual(t, w.GetPlayer(idx).GetTotalScore(), w.GetPlayer(i).GetTotalScore())
		}
	}

	before := w.GetRoundNumber()
	w.NextRound()
	assert.Equal(t, before, w.GetRoundNumber())
}

func TestIsraeliWhist_TieHasNoWinner(t *testing.T) {
	w := newTestIsraeliWhist(t)
	for i := range IsraeliWhistPlayerCnt {
		w.GetPlayer(i).SetTotalScore(30)
	}
	w.FinishGameForTest()
	assert.Equal(t, -1, w.GetWinnerIdx())
}

// --- プレイ ---

func TestIsraeliWhist_MustFollowSuit(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetTrumpSuitForTest(CardDesignHeart)
	w.SetPhaseForTest(IsraeliWhistPhasePlay)
	w.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}})
	w.SetCurrentPlayerIdxForTest(0)
	israeliWhistHandOf(w, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignSpade, 7, false))

	require.Error(t, w.PlayerPlay(0))
	require.NoError(t, w.PlayerPlay(1))
}

func TestIsraeliWhist_TrumpBeatsTheLeadSuit(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetTrumpSuitForTest(CardDesignHeart)
	w.SetPhaseForTest(IsraeliWhistPhasePlay)
	w.SetCurrentPlayerIdxForTest(0)
	w.SetLeadPlayerIdxForTest(0)

	israeliWhistHandOf(w, 0, NewCard(CardDesignSpade, 1, false))
	israeliWhistHandOf(w, 1, NewCard(CardDesignSpade, 13, false))
	israeliWhistHandOf(w, 2, NewCard(CardDesignHeart, 2, false))
	israeliWhistHandOf(w, 3, NewCard(CardDesignSpade, 12, false))

	for i := range IsraeliWhistPlayerCnt {
		require.NoError(t, w.PlayForTest(i, 0))
	}
	assert.Equal(t, 1, w.GetPlayer(2).GetTrickCount(), "切り札の 2 が A に勝つ")
}

func TestIsraeliWhist_PlayRejectsInvalidIndex(t *testing.T) {
	w := newTestIsraeliWhist(t)
	settleAuctionAndBids(t, w)
	w.SetCurrentPlayerIdxForTest(0)

	assert.Error(t, w.PlayerPlay(-1))
	assert.Error(t, w.PlayerPlay(99))
}

func TestIsraeliWhist_ValidIndicesOutOfRange(t *testing.T) {
	w := newTestIsraeliWhist(t)
	assert.Nil(t, w.GetValidPlayIndices(-1))
	assert.Nil(t, w.GetValidPlayIndices(IsraeliWhistPlayerCnt))
}

// **宣言に足りなければ取りに行き、足りていれば逃げる。** 両側を踏む。
func TestIsraeliWhist_CpuChasesThenDucks(t *testing.T) {
	build := func(bid int) *IsraeliWhist {
		w := newTestIsraeliWhist(t)
		w.SetTrumpSuitForTest(CardDesignHeart)
		w.SetPhaseForTest(IsraeliWhistPhasePlay)
		w.GetPlayer(2).SetBid(bid)
		w.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 12, false)}})
		w.SetCurrentPlayerIdxForTest(2)
		israeliWhistHandOf(w, 2,
			NewCard(CardDesignSpade, 1, false),  // 勝てるが強すぎる
			NewCard(CardDesignSpade, 13, false), // 勝てて一番弱い
			NewCard(CardDesignSpade, 7, false))  // 勝てない
		return w
	}

	assert.Equal(t, 1, build(3).CpuChoiceForTest(2), "足りないので K で取る")
	assert.Equal(t, 2, build(0).CpuChoiceForTest(2), "足りているので勝たない札を出す")
}

// --- ヒント ---

func TestIsraeliWhist_HintDuringAuction(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetAuctionPlayerIdxForTest(0)

	h := w.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex)
	assert.Contains(t, []string{"israeliwhistAuctionBid", "israeliwhistAuctionPass"}, h.Reason)
}

// **落札者にはノルマを守るヒントを出す。** 押せない宣言を勧めない。
func TestIsraeliWhist_HintMeetsTheQuota(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetDeclarerForTest(0, IsraeliWhistHandSize, CardDesignSpade)
	w.CloseAuctionForTest()
	w.SetBidPlayerIdxForTest(0)

	h := w.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "israeliwhistMeetQuota", h.Reason)
	assert.Equal(t, IsraeliWhistHandSize, h.Value)
}

func TestIsraeliWhist_HintDuringPlay(t *testing.T) {
	w := newTestIsraeliWhist(t)
	settleAuctionAndBids(t, w)
	w.SetCurrentPlayerIdxForTest(0)

	h := w.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Contains(t, w.GetValidPlayIndices(0), *h.CardIndex)
}

func TestIsraeliWhist_HintNilWhenNotHumanTurn(t *testing.T) {
	w := newTestIsraeliWhist(t)
	settleAuctionAndBids(t, w)
	w.SetCurrentPlayerIdxForTest(2)
	assert.Nil(t, w.GetHint())

	w.GiveUp()
	assert.Nil(t, w.GetHint())
}

// --- その他 ---

func TestIsraeliWhist_GiveUp(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.GiveUp()
	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, 1, w.GetWinnerIdx())

	w.SetWinnerIdxForTest(0)
	w.GiveUp()
	assert.Equal(t, 0, w.GetWinnerIdx(), "二度目は何も起きない")
}

func TestIsraeliWhist_AccessorsOutOfRange(t *testing.T) {
	w := newTestIsraeliWhist(t)
	assert.Nil(t, w.GetPlayer(-1))
	assert.Nil(t, w.GetPlayer(IsraeliWhistPlayerCnt))
	assert.Equal(t, IsraeliWhistPlayerCnt, w.GetPlayerCnt())
	assert.Equal(t, 0, w.GetTrickNumber())
	assert.Equal(t, 1, w.GetRoundNumber())
	assert.Empty(t, w.GetCurrentTrick())
	assert.Equal(t, 0, w.GetHighSuit())
	assert.GreaterOrEqual(t, w.GetAuctionPlayerIdx(), 0)
	assert.GreaterOrEqual(t, w.GetBidPlayerIdx(), 0)
	assert.GreaterOrEqual(t, w.GetLeadPlayerIdx(), 0)
}

func TestIsraeliWhistConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultIsraeliWhistConfig().Validate())
	assert.NoError(t, IsraeliWhistConfig{Rounds: IsraeliWhistRoundsMin}.Validate())
	assert.NoError(t, IsraeliWhistConfig{Rounds: IsraeliWhistRoundsMax}.Validate())
	assert.Error(t, IsraeliWhistConfig{Rounds: IsraeliWhistRoundsMin - 1}.Validate())
	assert.Error(t, IsraeliWhistConfig{Rounds: IsraeliWhistRoundsMax + 1}.Validate())
}

// --- JSON 往復 ---

// **2 段階ぶんの入札が往復しないと、入札がやり直しになる。**
func TestIsraeliWhist_JSONRoundTrip(t *testing.T) {
	w := newTestIsraeliWhist(t)
	w.SetAuctionPlayerIdxForTest(0)
	require.NoError(t, w.PlayerAuctionBid(7, CardDesignDiamond))
	w.SetDeclarerForTest(0, 7, CardDesignDiamond)
	w.CloseAuctionForTest()
	w.SetBidPlayerIdxForTest(0)
	require.NoError(t, w.PlayerBid(7))
	w.GetPlayer(0).SetTotalScore(59)
	w.GetPlayer(2).SetPassed(true)

	data, err := json.Marshal(w)
	require.NoError(t, err)

	var got IsraeliWhist
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, w.GetPhase(), got.GetPhase())
	assert.Equal(t, CardDesignDiamond, got.GetTrumpSuit())
	assert.Equal(t, 7, got.GetHighBid())
	assert.Equal(t, 0, got.GetDeclarerIdx())
	assert.Equal(t, 7, got.GetPlayer(0).GetBid())
	assert.Equal(t, 7, got.GetPlayer(0).GetAuctionBid())
	assert.Equal(t, CardDesignDiamond, got.GetPlayer(0).GetAuctionSuit())
	assert.Equal(t, 59, got.GetPlayer(0).GetTotalScore())
	assert.True(t, got.GetPlayer(2).GetPassed())
	assert.Equal(t, w.GetAuctionPlayerIdx(), got.GetAuctionPlayerIdx())
	assert.Equal(t, w.GetBidPlayerIdx(), got.GetBidPlayerIdx())
}

// 壊れたスナップショットは復元しない。
func TestIsraeliWhist_UnmarshalRejectsInvalid(t *testing.T) {
	valid := func() israeliWhistJSON {
		return israeliWhistJSON{
			Config:      DefaultIsraeliWhistConfig(),
			Phase:       IsraeliWhistPhasePlay,
			RoundNumber: 1,
			TrumpSuit:   CardDesignSpade,
			DeclarerIdx: 0,
			HighBid:     5,
			WinnerIdx:   -1,
		}
	}
	cases := map[string]func(*israeliWhistJSON){
		"bad config":   func(j *israeliWhistJSON) { j.Config.Rounds = 0 },
		"bad phase":    func(j *israeliWhistJSON) { j.Phase = IsraeliWhistPhase(99) },
		"bad trick":    func(j *israeliWhistJSON) { j.TrickNumber = IsraeliWhistTricksPerRound + 1 },
		"bad round":    func(j *israeliWhistJSON) { j.RoundNumber = 0 },
		"bad high bid": func(j *israeliWhistJSON) { j.HighBid = IsraeliWhistHandSize + 1 },
		"bad current":  func(j *israeliWhistJSON) { j.CurrentPlayerIdx = IsraeliWhistPlayerCnt },
		"bad auction":  func(j *israeliWhistJSON) { j.AuctionPlayerIdx = -1 },
		"bad bidder":   func(j *israeliWhistJSON) { j.BidPlayerIdx = -1 },
		"bad lead":     func(j *israeliWhistJSON) { j.LeadPlayerIdx = -1 },
		"bad dealer":   func(j *israeliWhistJSON) { j.DealerIdx = IsraeliWhistPlayerCnt },
		"bad declarer": func(j *israeliWhistJSON) { j.DeclarerIdx = IsraeliWhistPlayerCnt },
		"bad winner":   func(j *israeliWhistJSON) { j.WinnerIdx = IsraeliWhistPlayerCnt },
		// **切り札はフェーズと整合していなければならない。** 両方向を踏む。
		"trump before the auction closed": func(j *israeliWhistJSON) {
			j.Phase, j.TrumpSuit = IsraeliWhistPhaseAuction, CardDesignHeart
		},
		"no trump after the auction closed": func(j *israeliWhistJSON) {
			j.Phase, j.TrumpSuit = IsraeliWhistPhasePlay, 0
		},
		"bogus trump": func(j *israeliWhistJSON) { j.TrumpSuit = 99 },
		"long trick": func(j *israeliWhistJSON) {
			j.CurrentTrick = make([]*TrickCard, IsraeliWhistPlayerCnt+1)
		},
		"long log": func(j *israeliWhistJSON) {
			j.ActionLog = make([]*ActionLogEntry, israeliWhistMaxSliceLen+1)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			j := valid()
			mutate(&j)
			data, err := json.Marshal(j)
			require.NoError(t, err)
			var got IsraeliWhist
			assert.Error(t, json.Unmarshal(data, &got))
		})
	}

	var got IsraeliWhist
	assert.Error(t, got.UnmarshalJSON([]byte("{")))

	// 正のコントロール: 正しいスナップショットは通る。
	data, err := json.Marshal(valid())
	require.NoError(t, err)
	assert.NoError(t, json.Unmarshal(data, &got))

	// オークション中で切り札 0 も通る（ガードが一律に弾いていないこと）。
	j := valid()
	j.Phase, j.TrumpSuit = IsraeliWhistPhaseAuction, 0
	auction, err := json.Marshal(j)
	require.NoError(t, err)
	var okAuction IsraeliWhist
	assert.NoError(t, json.Unmarshal(auction, &okAuction))
}

func TestIsraeliWhist_ActionLog(t *testing.T) {
	w := newTestIsraeliWhist(t)
	require.NotEmpty(t, w.actionLog, "配りが記録される")
	w.SetAuctionPlayerIdxForTest(0)
	require.NoError(t, w.PlayerAuctionBid(6, CardDesignSpade))

	kinds := map[string]bool{}
	for _, e := range w.actionLog {
		kinds[e.ActionType] = true
	}
	assert.True(t, kinds["auction"])
}

// **全員的中と全員外しはどちらも 2 倍** (#5752)。発動したかどうかを
// アクションログの外からも読めるようにした値を、実際の精算で確かめる。
func TestIsraeliWhistRoundDoubledFlags(t *testing.T) {
	setBids := func(w *IsraeliWhist, bids, tricks [IsraeliWhistPlayerCnt]int) {
		for i := range IsraeliWhistPlayerCnt {
			p := w.GetPlayer(i)
			p.SetBid(bids[i])
			p.ResetTricks()
			for range tricks[i] {
				p.AddTrick([]*Card{})
			}
		}
	}

	t.Run("every seat hits", func(t *testing.T) {
		w := newTestIsraeliWhist(t)
		setBids(w, [IsraeliWhistPlayerCnt]int{3, 4, 3, 3}, [IsraeliWhistPlayerCnt]int{3, 4, 3, 3})
		w.finishRound()
		if !w.GetRoundDoubled() || !w.GetRoundAllExact() {
			t.Errorf("doubled=%v allExact=%v, want true/true", w.GetRoundDoubled(), w.GetRoundAllExact())
		}
	})

	t.Run("every seat misses", func(t *testing.T) {
		w := newTestIsraeliWhist(t)
		setBids(w, [IsraeliWhistPlayerCnt]int{3, 4, 3, 3}, [IsraeliWhistPlayerCnt]int{0, 1, 5, 7})
		w.finishRound()
		if !w.GetRoundDoubled() {
			t.Error("all four missing must double the round")
		}
		if w.GetRoundAllExact() {
			t.Error("the reason must read as everyone missing, not everyone hitting")
		}
	})

	// **1 人でも違えば通常ラウンド** (負のコントロール)。
	t.Run("a mixed round is not doubled", func(t *testing.T) {
		w := newTestIsraeliWhist(t)
		setBids(w, [IsraeliWhistPlayerCnt]int{3, 4, 3, 3}, [IsraeliWhistPlayerCnt]int{3, 0, 5, 5})
		w.finishRound()
		if w.GetRoundDoubled() {
			t.Error("a round where only some seats hit must not be doubled")
		}
	})
}
