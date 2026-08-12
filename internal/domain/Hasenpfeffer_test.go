//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHasenpfeffer(t *testing.T) *Hasenpfeffer {
	t.Helper()
	h := NewDefaultHasenpfeffer()
	h.Reset()
	return h
}

// **issue の「5 枚ずつ配る / 5 トリック」は算術が合わない。** 4 × 5 = 20 で
// 25 枚のうち 5 枚も余る。6 枚ずつ = 24 + 伏せ札 1 = 25 枚ちょうど、6 トリック。
func TestHasenpfeffer_DeckIsTwentyFiveCards(t *testing.T) {
	deck := NewTrumpCardsHasenpfeffer()
	assert.Equal(t, HasenpfefferDeckSize, deck.GetTotalCount())
	assert.Equal(t, 25, HasenpfefferDeckSize)
	assert.Equal(t, HasenpfefferPlayerCnt*HasenpfefferHandSize+1, deck.GetTotalCount(),
		"4 人 × 6 枚 + 伏せ札 1 枚")
	assert.Equal(t, 6, HasenpfefferTricksPerRound, "ユーカーの 5 ではなく 6")

	seen := map[string]int{}
	jokers := 0
	for range HasenpfefferDeckSize {
		c := deck.DrawCard()
		require.NotNil(t, c)
		seen[cardStr(c)]++
		if IsJokerCard(c) {
			jokers++
			continue
		}
		assert.NotContains(t, []int{2, 3, 4, 5, 6, 7, 8}, c.GetValue(), "8 以下は入らない")
	}
	assert.Equal(t, 1, jokers, "ジョーカーは 1 枚だけ")
	assert.Len(t, seen, HasenpfefferDeckSize, "重複が無い")
}

func TestHasenpfeffer_ResetDealsSixEachPlusABlind(t *testing.T) {
	h := newTestHasenpfeffer(t)

	assert.Equal(t, HasenpfefferPhaseBid, h.GetPhase())
	assert.Equal(t, 0, h.GetTrumpSuit())
	assert.Equal(t, -1, h.GetDeclarerIdx())
	assert.Equal(t, 1, h.GetBlindSize(), "余った 1 枚は伏せ札")
	total := h.GetBlindSize()
	for i := range HasenpfefferPlayerCnt {
		assert.Equal(t, HasenpfefferHandSize, h.GetPlayer(i).GetCardsSize())
		total += h.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, HasenpfefferDeckSize, total, "25 枚すべてが行き先を持つ")
	assert.Empty(t, h.GetValidPlayIndices(0), "競り中は出せない")
}

// **ジョーカーが全カード中最強（Best Bower）。** これがこのゲームの顔。
func TestHasenpfeffer_JokerIsTheBestBower(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetTrumpSuitForTest(CardDesignHeart)

	joker := NewCard(CardDesignJoker, CardValueJoker, false)
	right := NewCard(CardDesignHeart, 11, false)   // Right Bower
	left := NewCard(CardDesignDiamond, 11, false)  // Left Bower（同色）
	trumpAce := NewCard(CardDesignHeart, 1, false) // 切り札の A

	assert.Greater(t, h.CardRank(joker), h.CardRank(right), "Joker > Right Bower")
	assert.Greater(t, h.CardRank(right), h.CardRank(left), "Right > Left Bower")
	assert.Greater(t, h.CardRank(left), h.CardRank(trumpAce), "Left Bower > 切り札の A")
	assert.Greater(t, h.CardRank(trumpAce), h.CardRank(NewCard(CardDesignSpade, 1, false)),
		"切り札の A > 非切り札の A")
}

// **ジョーカーと Left Bower は切り札スート扱い。** フォロー判定にも効く。
func TestHasenpfeffer_JokerAndLeftBowerCountAsTrump(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetTrumpSuitForTest(CardDesignHeart)

	assert.Equal(t, CardDesignHeart, h.EffectiveSuit(NewCard(CardDesignJoker, CardValueJoker, false)))
	assert.Equal(t, CardDesignHeart, h.EffectiveSuit(NewCard(CardDesignDiamond, 11, false)),
		"Left Bower は元のスートではない")
	assert.Equal(t, CardDesignSpade, h.EffectiveSuit(NewCard(CardDesignSpade, 11, false)),
		"別色の J はただの J")
}

// **切り札がリードされたら、ジョーカーと Left Bower もフォロー対象。**
func TestHasenpfeffer_MustFollowWithTheLeftBowerAndJoker(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetTrumpSuitForTest(CardDesignHeart)
	h.SetPhaseForTest(HasenpfefferPhasePlay)
	h.SetLeadPlayerIdxForTest(0)
	h.SetCurrentPlayerIdxForTest(0)

	hasenpfefferHandOf(h, 0, NewCard(CardDesignHeart, 9, false))
	hasenpfefferHandOf(h, 1,
		NewCard(CardDesignDiamond, 11, false), // Left Bower = 切り札
		NewCard(CardDesignSpade, 1, false))
	hasenpfefferHandOf(h, 2, NewCard(CardDesignJoker, CardValueJoker, false), NewCard(CardDesignSpade, 9, false))
	hasenpfefferHandOf(h, 3, NewCard(CardDesignSpade, 10, false))

	require.NoError(t, h.PlayForTest(0, 0))
	assert.Equal(t, []int{0}, h.GetValidPlayIndices(1), "Left Bower でフォローする義務がある")
	assert.Equal(t, []int{0}, h.GetValidPlayIndices(2), "ジョーカーでフォローする義務がある")
}

// **競りは全員参加が義務。** 3 人が降りたら親は降りられない。
func TestHasenpfeffer_TheDealerCannotPassWhenEveryoneElseHas(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetDealerIdxForTest(3)
	h.SetCurrentPlayerIdxForTest(0)

	require.NoError(t, h.BidForTest(0, 0))
	require.NoError(t, h.BidForTest(1, 0))
	require.NoError(t, h.BidForTest(2, 0))

	assert.True(t, h.MustBid(3), "親は降りられない")
	assert.Error(t, h.BidForTest(3, 0), "降りる宣言は弾かれる")
	require.NoError(t, h.BidForTest(3, HasenpfefferMinBid))
	assert.Equal(t, 3, h.GetDeclarerIdx())
	assert.Equal(t, HasenpfefferMinBid, h.GetContract())
}

// **負のコントロール: 誰かが落札していれば親も降りられる。**
func TestHasenpfeffer_TheDealerMayPassOnceSomebodyHasBid(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetDealerIdxForTest(3)
	h.SetCurrentPlayerIdxForTest(0)

	require.NoError(t, h.BidForTest(0, HasenpfefferMinBid))
	require.NoError(t, h.BidForTest(1, 0))
	require.NoError(t, h.BidForTest(2, 0))

	assert.False(t, h.MustBid(3))
	assert.NoError(t, h.BidForTest(3, 0), "落札者がいるので降りられる")
	assert.Equal(t, 0, h.GetDeclarerIdx())
}

// **競りは上回らないと通らない。**
func TestHasenpfeffer_BidsMustOutbidTheStandingOne(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetDealerIdxForTest(3)
	h.SetCurrentPlayerIdxForTest(0)

	assert.Equal(t, HasenpfefferMinBid, h.NextBid(), "最初は下限から")
	assert.Error(t, h.BidForTest(0, HasenpfefferMinBid-1), "下限未満は通らない")
	assert.Error(t, h.BidForTest(0, HasenpfefferMaxBid+1), "上限超えは通らない")

	require.NoError(t, h.BidForTest(0, 4))
	assert.Equal(t, 5, h.NextBid())
	assert.Error(t, h.BidForTest(1, 4), "同額では上回れない")
	assert.NoError(t, h.BidForTest(1, 5))

	// **上限が立ったら誰も上回れない。** 同額で横取りできてしまうのを防ぐ。
	require.NoError(t, h.BidForTest(2, HasenpfefferMaxBid))
	assert.Zero(t, h.NextBid(), "もう宣言できない")
	assert.Error(t, h.BidForTest(3, HasenpfefferMaxBid), "同額での横取りは通らない")
	assert.NoError(t, h.BidForTest(3, 0), "降りるしかない")
	assert.Equal(t, 2, h.GetDeclarerIdx(), "上限を出した席が落札する")
}

// **落札者は伏せ札を取り込み、1 枚捨てて切り札を決める。**
func TestHasenpfeffer_TheDeclarerTakesTheBlindAndDiscards(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetDealerIdxForTest(3)
	h.SetCurrentPlayerIdxForTest(0)

	require.NoError(t, h.BidForTest(0, 4))
	require.NoError(t, h.BidForTest(1, 0))
	require.NoError(t, h.BidForTest(2, 0))
	require.NoError(t, h.BidForTest(3, 0))

	assert.Equal(t, HasenpfefferPhaseDiscard, h.GetPhase())
	assert.Equal(t, 0, h.GetBlindSize(), "伏せ札は落札者の手に入る")
	assert.Equal(t, HasenpfefferHandSize+1, h.GetPlayer(0).GetCardsSize(), "7 枚になる")

	require.NoError(t, h.DiscardForTest(0, 0, CardDesignHeart))
	assert.Equal(t, HasenpfefferPhasePlay, h.GetPhase())
	assert.Equal(t, CardDesignHeart, h.GetTrumpSuit())
	assert.Equal(t, HasenpfefferHandSize, h.GetPlayer(0).GetCardsSize(), "6 枚に戻る")
}

func TestHasenpfeffer_DiscardRejectsBadInput(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetContractForTest(0, 4)
	h.SetPhaseForTest(HasenpfefferPhaseDiscard)

	assert.Error(t, h.DiscardForTest(0, 0, 0), "スート 0 は通らない")
	assert.Error(t, h.DiscardForTest(0, 0, 9), "範囲外のスートは通らない")
	assert.Error(t, h.DiscardForTest(0, 99, CardDesignHeart), "範囲外の札は通らない")
	assert.Error(t, h.DiscardForTest(1, 0, CardDesignHeart), "落札者以外は捨てられない")
}

// **1 ハンドはちょうど 6 トリック。**
func TestHasenpfeffer_AHandIsExactlySixTricks(t *testing.T) {
	h := newTestHasenpfeffer(t)
	playOutBidding(t, h)
	for h.GetPhase() == HasenpfefferPhasePlay {
		idx := h.GetCurrentPlayerIdx()
		require.NoError(t, h.PlayForTest(idx, h.CpuChoiceForTest(idx)))
	}
	assert.Equal(t, HasenpfefferTricksPerRound, h.GetTrickNumber())
	total := 0
	for i := range HasenpfefferPlayerCnt {
		total += h.GetPlayer(i).GetTrickCount()
		assert.Zero(t, h.GetPlayer(i).GetCardsSize(), "手札を打ち切る")
	}
	assert.Equal(t, HasenpfefferTricksPerRound, total)
}

// playOutBidding は競りと捨て札を CPU に任せてプレイフェーズまで進める。
func playOutBidding(t *testing.T, h *Hasenpfeffer) {
	t.Helper()
	for h.GetPhase() == HasenpfefferPhaseBid {
		idx := h.GetCurrentPlayerIdx()
		require.NoError(t, h.BidForTest(idx, h.CpuBidChoiceForTest(idx)))
	}
	if h.GetPhase() == HasenpfefferPhaseDiscard {
		d := h.GetDeclarerIdx()
		suit := h.CpuTrumpChoiceForTest(d)
		require.NoError(t, h.DiscardForTest(d, 0, suit))
	}
}

// **達成したら取ったトリック数が点、落としたら相手に落札額。**
func TestHasenpfeffer_ScoringMadeAndEuchred(t *testing.T) {
	for _, tc := range []struct {
		name       string
		contract   int
		declTricks int
		wantScores []int
		euchred    bool
	}{
		{"ちょうど達成", 3, 3, []int{3, 0}, false},
		{"超過して達成", 3, 5, []int{5, 0}, false},
		{"落とした", 4, 3, []int{0, 4}, true},
		{"全部落とした", 6, 0, []int{0, 6}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHasenpfeffer(t)
			h.SetContractForTest(0, tc.contract)
			h.SetPhaseForTest(HasenpfefferPhasePlay)
			h.GiveTricksForTest(0, tc.declTricks)
			h.GiveTricksForTest(1, HasenpfefferTricksPerRound-tc.declTricks)
			h.FinishHandForTest()

			assert.Equal(t, tc.wantScores[0], h.GetScore(0))
			assert.Equal(t, tc.wantScores[1], h.GetScore(1))
			assert.Equal(t, tc.euchred, h.GetLastHandEuchred())
			assert.Equal(t, tc.declTricks, h.GetLastHandTricks())
		})
	}
}

// **目標点に届いたら終わる。**
func TestHasenpfeffer_ReachingTheTargetEndsTheGame(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetContractForTest(0, 6)
	h.SetPhaseForTest(HasenpfefferPhasePlay)
	h.SetScoreForTestUse(0, h.GetConfig().Target-HasenpfefferTricksPerRound)
	h.GiveTricksForTest(0, HasenpfefferTricksPerRound)
	h.FinishHandForTest()

	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, HasenpfefferPhaseGameEnd, h.GetPhase())
	assert.Equal(t, 0, h.GetWinnerTeam())
}

func TestHasenpfeffer_NextHandRotatesTheDealer(t *testing.T) {
	h := newTestHasenpfeffer(t)
	before := h.GetDealerIdx()
	h.SetPhaseForTest(HasenpfefferPhaseHandEnd)
	h.NextHand()
	assert.Equal(t, (before+1)%HasenpfefferPlayerCnt, h.GetDealerIdx())
	assert.Equal(t, HasenpfefferPhaseBid, h.GetPhase())
	assert.Equal(t, 1, h.GetBlindSize(), "配り直すので伏せ札が戻る")

	h.FinishGameForTest()
	after := h.GetDealerIdx()
	h.NextHand()
	assert.Equal(t, after, h.GetDealerIdx(), "終局後は進まない")
}

// **切り札は非切り札より強い。** 別スートの A はトリックを取れない。
func TestHasenpfeffer_TrickWinnerOrdering(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetTrumpSuitForTest(CardDesignHeart)

	h.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 9, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 13, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignClover, 1, false)},
	})
	assert.Equal(t, 1, h.TrickWinnerForTest(), "切り札の 9 が ♠A に勝つ")

	h.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 11, false)}, // Right Bower
		{PlayerIdx: 1, Card: NewCard(CardDesignJoker, CardValueJoker, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 11, false)}, // Left Bower
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 1, false)},
	})
	assert.Equal(t, 1, h.TrickWinnerForTest(), "ジョーカーが Right Bower にも勝つ")

	h.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 10, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 1, false)}, // 別スートの A は取れない
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 9, false)},
	})
	assert.Equal(t, 2, h.TrickWinnerForTest())
}

func TestHasenpfeffer_FollowSuitIsCompulsory(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetTrumpSuitForTest(CardDesignHeart)
	h.SetPhaseForTest(HasenpfefferPhasePlay)
	h.SetLeadPlayerIdxForTest(0)
	h.SetCurrentPlayerIdxForTest(0)
	hasenpfefferHandOf(h, 0, NewCard(CardDesignSpade, 9, false))
	hasenpfefferHandOf(h, 1, NewCard(CardDesignSpade, 10, false), NewCard(CardDesignClover, 1, false))
	hasenpfefferHandOf(h, 2, NewCard(CardDesignSpade, 12, false))
	hasenpfefferHandOf(h, 3, NewCard(CardDesignSpade, 13, false))

	require.NoError(t, h.PlayForTest(0, 0))
	assert.Equal(t, []int{0}, h.GetValidPlayIndices(1))
	assert.Error(t, h.PlayForTest(1, 1))
}

func TestHasenpfeffer_RejectsOutOfTurnAndBadIndices(t *testing.T) {
	h := newTestHasenpfeffer(t)
	assert.Error(t, h.PlayForTest(0, 0), "競り中は打てない")

	playOutBidding(t, h)
	idx := h.GetCurrentPlayerIdx()
	assert.Error(t, h.PlayForTest((idx+1)%HasenpfefferPlayerCnt, 0), "手番でない席は打てない")
	assert.Error(t, h.PlayForTest(idx, -1))
	assert.Error(t, h.PlayForTest(idx, 999))

	h.FinishGameForTest()
	assert.Error(t, h.PlayForTest(idx, 0), "終局後は打てない")
	assert.Error(t, h.BidForTest(0, 3), "終局後は宣言できない")
}

// **CPU は必ず合法手を返す。**
func TestHasenpfeffer_CpuAlwaysChoosesLegally(t *testing.T) {
	for range 100 {
		h := NewDefaultHasenpfeffer()
		h.Reset()
		for range 40 {
			if h.GetGameEndFlag() {
				break
			}
			switch h.GetPhase() {
			case HasenpfefferPhaseBid:
				idx := h.GetCurrentPlayerIdx()
				bid := h.CpuBidChoiceForTest(idx)
				require.NoError(t, h.BidForTest(idx, bid), "CPU の宣言は必ず合法")
			case HasenpfefferPhaseDiscard:
				d := h.GetDeclarerIdx()
				require.NoError(t, h.DiscardForTest(d, 0, h.CpuTrumpChoiceForTest(d)))
			case HasenpfefferPhasePlay:
				idx := h.GetCurrentPlayerIdx()
				choice := h.CpuChoiceForTest(idx)
				require.Contains(t, h.GetValidPlayIndices(idx), choice)
				require.NoError(t, h.PlayForTest(idx, choice))
			case HasenpfefferPhaseHandEnd:
				h.NextHand()
			default:
			}
		}
	}
}

// **どの局も必ず終わる。**
func TestHasenpfeffer_GamesTerminate(t *testing.T) {
	for range 30 {
		h := NewDefaultHasenpfeffer()
		h.Reset()
		for turns := 0; !h.GetGameEndFlag(); turns++ {
			require.Less(t, turns, 3000, "終わらない")
			switch h.GetPhase() {
			case HasenpfefferPhaseBid:
				idx := h.GetCurrentPlayerIdx()
				require.NoError(t, h.BidForTest(idx, h.CpuBidChoiceForTest(idx)))
			case HasenpfefferPhaseDiscard:
				d := h.GetDeclarerIdx()
				require.NoError(t, h.DiscardForTest(d, 0, h.CpuTrumpChoiceForTest(d)))
			case HasenpfefferPhasePlay:
				idx := h.GetCurrentPlayerIdx()
				require.NoError(t, h.PlayForTest(idx, h.CpuChoiceForTest(idx)))
			case HasenpfefferPhaseHandEnd:
				h.NextHand()
			default:
			}
		}
		assert.GreaterOrEqual(t, max(h.GetScore(0), h.GetScore(1)), h.GetConfig().Target)
	}
}

// **落札者は必ず決まる。** 義務競りがあるので流局しない。
func TestHasenpfeffer_EveryHandFindsADeclarer(t *testing.T) {
	for range 200 {
		h := NewDefaultHasenpfeffer()
		h.Reset()
		for h.GetPhase() == HasenpfefferPhaseBid {
			idx := h.GetCurrentPlayerIdx()
			require.NoError(t, h.BidForTest(idx, h.CpuBidChoiceForTest(idx)))
		}
		require.GreaterOrEqual(t, h.GetDeclarerIdx(), 0, "流局しない")
		assert.GreaterOrEqual(t, h.GetContract(), HasenpfefferMinBid)
		assert.LessOrEqual(t, h.GetContract(), HasenpfefferMaxBid)
	}
}

func TestHasenpfeffer_GiveUp(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.GiveUp()
	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, 1, h.GetWinnerTeam())

	h.GiveUp()
	assert.Equal(t, 1, h.GetWinnerTeam())
}

// **公開の入口も踏む。** テストが `*ForTest` の私有経路ばかり通ると、
// 実際に画面から呼ばれる `Player*` / `Cpu*` のガードが未検証のまま残る。
func TestHasenpfeffer_PublicEntryPointsGuardTheTurn(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetDealerIdxForTest(3)
	h.SetCurrentPlayerIdxForTest(1)

	// 競り: 人間の番でなければ弾く。
	assert.False(t, h.IsHumanBidTurn())
	assert.Error(t, h.PlayerBid(4))

	h.SetCurrentPlayerIdxForTest(0)
	assert.True(t, h.IsHumanBidTurn())
	before := h.GetPlayer(1).GetBid()
	h.CpuBid()
	assert.Equal(t, before, h.GetPlayer(1).GetBid(), "人間の番では CPU は動かない")

	// **人間が先に上限で落札する。** 順番が逆だと配り依存で落ちます——先に CPU に
	// 宣言させると、その CPU が上限を宣言した配り (実測 1.4%) では人間が上回れず
	// `PlayerBid` が "bid 6 is already the maximum" で弾かれます。
	require.NoError(t, h.PlayerBid(HasenpfefferMaxBid))

	// 上限が立ったあとでも、CPU は自分の番に「降りる」という宣言をする。
	h.SetCurrentPlayerIdxForTest(1)
	h.CpuBid()
	assert.True(t, h.GetPlayer(1).HasBid(), "CPU は自分の番で宣言する")
	assert.Equal(t, HasenpfefferMaxBid, h.GetContract())

	// 残りを CPU に任せて捨て札フェーズへ。
	for h.GetPhase() == HasenpfefferPhaseBid {
		h.CpuBid()
	}
	require.Equal(t, HasenpfefferPhaseDiscard, h.GetPhase())
	require.Equal(t, 0, h.GetDeclarerIdx(), "落札したのは人間")

	// 捨て札: 落札者が人間なら CPU は動かない。
	assert.True(t, h.IsHumanDiscardTurn())
	sizeBefore := h.GetPlayer(0).GetCardsSize()
	h.CpuDiscard()
	assert.Equal(t, sizeBefore, h.GetPlayer(0).GetCardsSize(), "人間の捨て札を勝手にやらない")
	require.NoError(t, h.PlayerDiscard(0, CardDesignHeart))
	assert.Equal(t, HasenpfefferPhasePlay, h.GetPhase())

	// プレイ: 手番でなければ弾く。
	h.SetCurrentPlayerIdxForTest(1)
	assert.False(t, h.IsHumanTurn())
	assert.Error(t, h.PlayerPlay(0))
	cpuBefore := h.GetPlayer(1).GetCardsSize()
	h.CpuPlay()
	assert.Equal(t, cpuBefore-1, h.GetPlayer(1).GetCardsSize(), "CPU は自分の番で打つ")

	h.SetCurrentPlayerIdxForTest(0)
	humanBefore := h.GetPlayer(0).GetCardsSize()
	h.CpuPlay()
	assert.Equal(t, humanBefore, h.GetPlayer(0).GetCardsSize(), "人間の番では CPU は動かない")
	require.NoError(t, h.PlayerPlay(h.GetValidPlayIndices(0)[0]))
	assert.Equal(t, humanBefore-1, h.GetPlayer(0).GetCardsSize())
}

// **CPU の捨て札は落札者が CPU のときだけ動く。**
func TestHasenpfeffer_CpuDiscardRunsForACpuDeclarer(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetContractForTest(1, 4)
	h.SetPhaseForTest(HasenpfefferPhaseDiscard)
	before := h.GetPlayer(1).GetCardsSize()

	h.CpuDiscard()

	assert.Equal(t, HasenpfefferPhasePlay, h.GetPhase())
	assert.Equal(t, before-1, h.GetPlayer(1).GetCardsSize())
	assert.GreaterOrEqual(t, h.GetTrumpSuit(), CardDesignSpade, "切り札も決まる")
}

// **CPU は切り札を捨てない。**
func TestHasenpfeffer_CpuDiscardKeepsTrumps(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetContractForTest(1, 4)
	h.SetPhaseForTest(HasenpfefferPhaseDiscard)
	hasenpfefferHandOf(h, 1,
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignHeart, 11, false),   // Right Bower
		NewCard(CardDesignDiamond, 11, false), // Left Bower
		NewCard(CardDesignJoker, CardValueJoker, false),
		NewCard(CardDesignSpade, 9, false)) // これが捨てられるはず

	discard := h.GetPlayer(1).GetCard(h.CpuDiscardChoiceForTest(1, CardDesignHeart))
	assert.Equal(t, CardDesignSpade, discard.GetDesign(), "切り札でない最弱札を捨てる")
	assert.Equal(t, 9, discard.GetValue())
}

func TestHasenpfeffer_Hint(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetDealerIdxForTest(3)
	h.SetCurrentPlayerIdxForTest(0)

	// **競りの助言は札ではなく額を指す。**
	hint := h.GetHint()
	require.NotNil(t, hint)
	assert.Nil(t, hint.CardIndex)
	assert.Contains(t, []string{"hasenpfefferBid", "hasenpfefferPass", "hasenpfefferMustBid"}, hint.Reason)

	require.NoError(t, h.BidForTest(0, 4))
	require.NoError(t, h.BidForTest(1, 0))
	require.NoError(t, h.BidForTest(2, 0))
	require.NoError(t, h.BidForTest(3, 0))

	// **捨て札の助言はスートと札の両方を指す。**
	hint = h.GetHint()
	require.NotNil(t, hint)
	require.NotNil(t, hint.CardIndex)
	assert.Equal(t, "hasenpfefferDiscard", hint.Reason)
	assert.GreaterOrEqual(t, hint.Suit, CardDesignSpade)

	require.NoError(t, h.DiscardForTest(0, *hint.CardIndex, hint.Suit))
	h.SetCurrentPlayerIdxForTest(0)
	hint = h.GetHint()
	require.NotNil(t, hint)
	require.NotNil(t, hint.CardIndex)
	assert.Contains(t, h.GetValidPlayIndices(0), *hint.CardIndex, "勧める札は必ず合法")

	h.FinishGameForTest()
	assert.Nil(t, h.GetHint(), "終局後は助言しない")
}

// **親が降りられない場面では、そう言う助言になる。**
func TestHasenpfeffer_HintSaysWhenTheDealerCannotPass(t *testing.T) {
	h := newTestHasenpfeffer(t)
	h.SetDealerIdxForTest(0)
	h.SetCurrentPlayerIdxForTest(1)
	require.NoError(t, h.BidForTest(1, 0))
	require.NoError(t, h.BidForTest(2, 0))
	require.NoError(t, h.BidForTest(3, 0))

	require.True(t, h.MustBid(0))
	hint := h.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "hasenpfefferMustBid", hint.Reason)
	assert.GreaterOrEqual(t, hint.Value, HasenpfefferMinBid, "降りる助言はしない")
}

func TestHasenpfeffer_JSONRoundTrip(t *testing.T) {
	h := newTestHasenpfeffer(t)
	playOutBidding(t, h)
	for range 4 {
		idx := h.GetCurrentPlayerIdx()
		require.NoError(t, h.PlayForTest(idx, h.CpuChoiceForTest(idx)))
	}

	data, err := json.Marshal(h)
	require.NoError(t, err)

	var restored Hasenpfeffer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, h.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, h.GetDeclarerIdx(), restored.GetDeclarerIdx())
	assert.Equal(t, h.GetContract(), restored.GetContract(), "落札額が消えない")
	assert.Equal(t, h.GetTrickNumber(), restored.GetTrickNumber())
	for i := range HasenpfefferPlayerCnt {
		assert.Equal(t, h.GetPlayer(i).GetBid(), restored.GetPlayer(i).GetBid(), "宣言が消えない")
		assert.Equal(t, h.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
	}
}

// **壊れたスナップショットは弾く。**
func TestHasenpfeffer_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	base := func(t *testing.T) map[string]any {
		t.Helper()
		h := newTestHasenpfeffer(t)
		playOutBidding(t, h)
		data, err := json.Marshal(h)
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
		{"declarer out of range", func(m map[string]any) { m["di"] = 9 }},
		{"contract below the minimum", func(m map[string]any) { m["co"] = 1 }},
		{"contract above the maximum", func(m map[string]any) { m["co"] = 99 }},
		{"a blind that survived the deal", func(m map[string]any) {
			m["bl"] = []any{map[string]any{"d": 1, "v": 9, "j": false}}
		}},
		{"current player out of range", func(m map[string]any) { m["ci"] = 9 }},
		{"dealer out of range", func(m map[string]any) { m["dl"] = -1 }},
		{"winner before the game ended", func(m map[string]any) { m["wt"] = 1 }},
		{"hand number below one", func(m map[string]any) { m["hn"] = 0 }},
		{"negative trick number", func(m map[string]any) { m["tn"] = -1 }},
		{"scores of the wrong length", func(m map[string]any) { m["sc"] = []any{0} }},
		{"config out of range", func(m map[string]any) { m["cf"] = map[string]any{"t": 0} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base(t)
			tc.mutate(m)
			data, err := json.Marshal(m)
			require.NoError(t, err)
			var restored Hasenpfeffer
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通る。**
	data, err := json.Marshal(base(t))
	require.NoError(t, err)
	var ok Hasenpfeffer
	assert.NoError(t, json.Unmarshal(data, &ok))
}

// **場札の中身も検証する。** 枚数だけ見て素通しすると、壊れた KV から
// nil の Card や範囲外の席番号が入り、**復元したあとで panic する**
// （レビュー指摘 PR #5310）。
func TestHasenpfeffer_UnmarshalRejectsBrokenTrickEntries(t *testing.T) {
	base := func(t *testing.T) map[string]any {
		t.Helper()
		h := newTestHasenpfeffer(t)
		playOutBidding(t, h)
		require.NoError(t, h.PlayForTest(h.GetCurrentPlayerIdx(), 0))
		data, err := json.Marshal(h)
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(data, &m))
		require.Len(t, m["ct"], 1, "場に 1 枚出ている状態を作る")
		return m
	}

	for _, tc := range []struct {
		name  string
		entry any
	}{
		{"a nil entry", nil},
		{"an entry with no card", map[string]any{"playerIdx": 0}},
		{"a player index above the table", map[string]any{"playerIdx": 9, "card": map[string]any{"d": 1, "v": 9, "j": false}}},
		{"a negative player index", map[string]any{"playerIdx": -1, "card": map[string]any{"d": 1, "v": 9, "j": false}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base(t)
			m["ct"] = []any{tc.entry}
			data, err := json.Marshal(m)
			require.NoError(t, err)
			var restored Hasenpfeffer
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていない場札は通り、復元後も panic しない。**
	data, err := json.Marshal(base(t))
	require.NoError(t, err)
	var ok Hasenpfeffer
	require.NoError(t, json.Unmarshal(data, &ok))
	assert.NotPanics(t, func() {
		_ = ok.GetValidPlayIndices(ok.GetCurrentPlayerIdx())
		_ = ok.TrickWinnerForTest()
	})
}

// **競り中は落札者も落札額も空。** 対で検証する。
func TestHasenpfeffer_UnmarshalRejectsADeclarerDuringBidding(t *testing.T) {
	h := newTestHasenpfeffer(t)
	data, err := json.Marshal(h)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	m["di"] = 2 // 競り中なのに落札者がいる

	bad, err := json.Marshal(m)
	require.NoError(t, err)
	var restored Hasenpfeffer
	assert.Error(t, json.Unmarshal(bad, &restored))
}

// **宣言は -1 / 0 / 3..6 のいずれか。**
func TestHasenpfefferPlayer_UnmarshalRejectsBrokenBids(t *testing.T) {
	for _, body := range []string{`{"bd":1}`, `{"bd":2}`, `{"bd":7}`, `{"bd":-2}`} {
		var p HasenpfefferPlayer
		assert.Error(t, json.Unmarshal([]byte(body), &p), body)
	}
	// **負のコントロール: 正しい値は通る。**
	for _, body := range []string{`{"bd":-1}`, `{"bd":0}`, `{"bd":3}`, `{"bd":6}`} {
		var p HasenpfefferPlayer
		assert.NoError(t, json.Unmarshal([]byte(body), &p), body)
	}
}

func TestHasenpfefferConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultHasenpfefferConfig().Validate())
	assert.NoError(t, HasenpfefferConfig{Target: HasenpfefferTargetMin}.Validate())
	assert.NoError(t, HasenpfefferConfig{Target: HasenpfefferTargetMax}.Validate())
	assert.Error(t, HasenpfefferConfig{Target: HasenpfefferTargetMin - 1}.Validate())
	assert.Error(t, HasenpfefferConfig{Target: HasenpfefferTargetMax + 1}.Validate())
}

func TestHasenpfeffer_TeamsAndBounds(t *testing.T) {
	assert.Equal(t, 0, HasenpfefferTeamOf(0))
	assert.Equal(t, 1, HasenpfefferTeamOf(1))
	assert.Equal(t, 0, HasenpfefferTeamOf(2), "0 と 2 が味方")
	assert.Equal(t, 1, HasenpfefferTeamOf(3))

	h := newTestHasenpfeffer(t)
	assert.Nil(t, h.GetPlayer(-1))
	assert.Nil(t, h.GetPlayer(99))
	assert.Zero(t, h.GetScore(-1))
	assert.Zero(t, h.GetScore(99))
	assert.Empty(t, h.GetValidPlayIndices(-1))
	assert.Equal(t, HasenpfefferPlayerCnt, h.GetPlayerCnt())
	assert.NotEmpty(t, h.GetActionLog())
	assert.False(t, IsJokerCard(nil))
}
