//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHoneymoonBridge(t *testing.T) *HoneymoonBridge {
	t.Helper()
	h := NewDefaultHoneymoonBridge()
	h.Reset()
	return h
}

// **2 人 × 13 枚 = 26、残りもちょうど 26。** 引き合いは 13 トリック × 2 枚引き
// = 26 で山札を使い切り、両者とも再び 13 枚に戻る。
func TestHoneymoonBridge_DealAndStockLineUp(t *testing.T) {
	assert.Equal(t, 26, HoneymoonBridgeStockSize, "52 - 2×13 = 26")
	assert.Equal(t, HoneymoonBridgeTricksPerPhase*HoneymoonBridgePlayerCnt, HoneymoonBridgeStockSize,
		"13 トリック × 2 枚引き = 山札ちょうど")

	h := newTestHoneymoonBridge(t)
	total := h.GetStockSize()
	for i := range HoneymoonBridgePlayerCnt {
		assert.Equal(t, HoneymoonBridgeHandSize, h.GetPlayer(i).GetCardsSize())
		total += h.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total, "52 枚すべてが行き先を持つ")
}

// **席数が合わない呼び出しでも遊べる盤面を返す（レビュー指摘 PR #5312）。**
func TestHoneymoonBridge_ConstructorFixesTheSeatCount(t *testing.T) {
	for _, given := range [][]*HoneymoonBridgePlayer{
		nil,
		{NewHoneymoonBridgePlayer(true)},
		{NewHoneymoonBridgePlayer(true), NewHoneymoonBridgePlayer(false), NewHoneymoonBridgePlayer(false)},
	} {
		h := NewHoneymoonBridge(given, DefaultHoneymoonBridgeConfig())
		assert.Equal(t, HoneymoonBridgePlayerCnt, h.GetPlayerCnt())
		assert.NotPanics(t, h.Reset)
		assert.Equal(t, HoneymoonBridgeHandSize, h.GetPlayer(0).GetCardsSize())
	}

	// **負のコントロール: 2 人ちょうどなら渡したものをそのまま使う。**
	mine := []*HoneymoonBridgePlayer{NewHoneymoonBridgePlayer(false), NewHoneymoonBridgePlayer(true)}
	h := NewHoneymoonBridge(mine, DefaultHoneymoonBridgeConfig())
	assert.False(t, h.GetPlayer(0).GetIsHuman(), "席の並びを勝手に入れ替えない")
	assert.True(t, h.GetPlayer(1).GetIsHuman())
}

func TestHoneymoonBridge_ResetStartsInTheDrawPhase(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	assert.Equal(t, HoneymoonBridgePhaseDraw, h.GetPhase())
	assert.Equal(t, 0, h.GetTrumpSuit(), "引き合いは切り札なし")
	assert.Equal(t, -1, h.GetDeclarerIdx())
	assert.Equal(t, 0, h.GetContractLevel())
	assert.Equal(t, 1, h.GetRoundNumber())
	assert.NotEmpty(t, h.GetValidPlayIndices(0), "引き合いでも札は出せる")
}

// **引き合いを打ち切ると山札が尽き、両者 13 枚で競りに入る。**
func TestHoneymoonBridge_DrawPhaseExhaustsTheStock(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	for h.GetPhase() == HoneymoonBridgePhaseDraw {
		idx := h.GetCurrentPlayerIdx()
		require.NoError(t, h.PlayForTest(idx, h.CpuChoiceForTest(idx)))
	}
	assert.Equal(t, HoneymoonBridgePhaseBid, h.GetPhase())
	assert.Equal(t, 0, h.GetStockSize(), "山札を使い切る")
	for i := range HoneymoonBridgePlayerCnt {
		assert.Equal(t, HoneymoonBridgeHandSize, h.GetPlayer(i).GetCardsSize(), "両者 13 枚に戻る")
	}
}

// **引き合いのトリックは得点にならない。** 数えるのは本番だけ。
func TestHoneymoonBridge_DrawPhaseTricksDoNotScore(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	for h.GetPhase() == HoneymoonBridgePhaseDraw {
		idx := h.GetCurrentPlayerIdx()
		require.NoError(t, h.PlayForTest(idx, h.CpuChoiceForTest(idx)))
	}
	for i := range HoneymoonBridgePlayerCnt {
		assert.Zero(t, h.GetPlayer(i).GetTrickCount(), "席 %d の獲得は 0 のまま", i)
	}
}

// **勝った人が先に引く。** 山札は必ず 2 枚ずつ減る。
func TestHoneymoonBridge_TheTrickWinnerDrawsFirst(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetLeadPlayerIdxForTest(0)
	h.SetCurrentPlayerIdxForTest(0)
	honeymoonBridgeHandOf(h, 0, NewCard(CardDesignSpade, 1, false))
	honeymoonBridgeHandOf(h, 1, NewCard(CardDesignSpade, 2, false))
	// 山札の先頭が勝者に、次が敗者に行く。
	h.SetStockForTest([]*Card{
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 2, false),
	})

	require.NoError(t, h.PlayForTest(0, 0))
	require.NoError(t, h.PlayForTest(1, 0))

	assert.Zero(t, h.GetStockSize(), "2 枚とも配られる")
	assert.Equal(t, 13, h.GetPlayer(0).GetCard(0).GetValue(), "♠A で勝った 0 が K を引く")
	assert.Equal(t, 2, h.GetPlayer(1).GetCard(0).GetValue())
	assert.Equal(t, 0, h.GetLeadPlayerIdx(), "勝者が次のリード")
}

// **引き合いは切り札なし。** 別スートの A は取れない。
func TestHoneymoonBridge_DrawPhaseHasNoTrump(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 3, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
	})
	assert.Equal(t, 0, h.TrickWinnerForTest(), "♥A は ♠3 に勝てない")
}

// **競りは上回らないと通らない。** 同レベルならスート序列、NT が最強。
func TestHoneymoonBridge_BidsMustOutbid(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetPhaseForTest(HoneymoonBridgePhaseBid)
	h.SetCurrentPlayerIdxForTest(0)

	assert.True(t, h.OutbidsForTest(1, CardDesignSpade), "最初は何でも通る")
	require.NoError(t, h.BidForTest(0, 2, CardDesignHeart))

	assert.False(t, h.OutbidsForTest(2, CardDesignSpade), "同レベルで下のスートは通らない")
	assert.False(t, h.OutbidsForTest(2, CardDesignHeart), "同じ宣言は通らない")
	assert.True(t, h.OutbidsForTest(2, CardDesignDiamond), "同レベルで上のスートは通る")
	assert.True(t, h.OutbidsForTest(2, 0), "ノートランプが最強")
	assert.True(t, h.OutbidsForTest(3, CardDesignSpade), "レベルが上なら何でも通る")

	assert.Error(t, h.BidForTest(1, 2, CardDesignSpade))
	assert.NoError(t, h.BidForTest(1, 2, 0), "NT で上回れる")
	assert.Equal(t, 1, h.GetDeclarerIdx())
	assert.Equal(t, 0, h.GetTrumpSuit(), "NT は切り札なし")
}

// **スート序列は CardDesign の並びとは違う。** NT を最強に置くための独自序列。
func TestHoneymoonBridge_SuitRankPutsNoTrumpOnTop(t *testing.T) {
	assert.Greater(t, honeymoonBridgeSuitRank(0), honeymoonBridgeSuitRank(CardDesignDiamond), "NT > ♦")
	assert.Greater(t, honeymoonBridgeSuitRank(CardDesignDiamond), honeymoonBridgeSuitRank(CardDesignHeart))
	assert.Greater(t, honeymoonBridgeSuitRank(CardDesignHeart), honeymoonBridgeSuitRank(CardDesignClover))
	assert.Greater(t, honeymoonBridgeSuitRank(CardDesignClover), honeymoonBridgeSuitRank(CardDesignSpade))
}

// **2 回続けてパスしたら競りが締まる。**
func TestHoneymoonBridge_TwoPassesCloseTheAuction(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetPhaseForTest(HoneymoonBridgePhaseBid)
	h.SetCurrentPlayerIdxForTest(0)

	require.NoError(t, h.BidForTest(0, 3, CardDesignHeart))
	require.NoError(t, h.BidForTest(1, 0, 0))
	require.NoError(t, h.BidForTest(0, 0, 0))

	assert.Equal(t, HoneymoonBridgePhasePlay, h.GetPhase())
	assert.Equal(t, 0, h.GetDeclarerIdx())
	assert.Equal(t, 3, h.GetContractLevel())
	// **リードは落札者の相手から。**
	assert.Equal(t, 1, h.GetLeadPlayerIdx())
}

// **両者パスならディールは流れる。**
func TestHoneymoonBridge_BothPassingThrowsTheDealIn(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetPhaseForTest(HoneymoonBridgePhaseBid)
	h.SetCurrentPlayerIdxForTest(0)

	require.NoError(t, h.BidForTest(0, 0, 0))
	require.NoError(t, h.BidForTest(1, 0, 0))

	assert.Equal(t, HoneymoonBridgePhaseRoundEnd, h.GetPhase())
	assert.Equal(t, -1, h.GetDeclarerIdx(), "落札者はいない")
	assert.Zero(t, h.GetContractLevel())
}

// **契約は 6 + レベル トリック。** ブリッジと同じ数え方。
func TestHoneymoonBridge_RequiredTricksAddsTheBook(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	assert.Zero(t, h.RequiredTricks(), "契約が無ければ 0")
	for level := 1; level <= HoneymoonBridgeMaxLevel; level++ {
		h.SetContractForTest(0, level, CardDesignHeart)
		assert.Equal(t, HoneymoonBridgeBookTricks+level, h.RequiredTricks())
	}
	h.SetContractForTest(0, HoneymoonBridgeMaxLevel, CardDesignHeart)
	assert.Equal(t, HoneymoonBridgeTricksPerPhase, h.RequiredTricks(), "レベル 7 は全トリック")
}

func TestHoneymoonBridge_BidRejectsBadInput(t *testing.T) {
	for _, tc := range []struct {
		name        string
		level, suit int
	}{
		{"level below one", 0, CardDesignHeart},
		{"level above the maximum", HoneymoonBridgeMaxLevel + 1, CardDesignHeart},
		{"suit out of range", 2, 9},
		{"negative suit", 2, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHoneymoonBridge(t)
			h.SetPhaseForTest(HoneymoonBridgePhaseBid)
			h.SetCurrentPlayerIdxForTest(0)
			if tc.level == 0 {
				// level 0 は「降りる」なので suit 付きでも通る。別途 suit を検査する。
				require.NoError(t, h.BidForTest(0, 0, tc.suit))
				assert.Zero(t, h.GetContractLevel())
				return
			}
			assert.Error(t, h.BidForTest(0, tc.level, tc.suit))
		})
	}
}

// **切り札は本番のトリックに効く。**
func TestHoneymoonBridge_TrumpBeatsTheLedSuitInPlay(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetTrumpSuitForTest(CardDesignDiamond)
	h.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 2, false)},
	})
	assert.Equal(t, 1, h.TrickWinnerForTest(), "切り札の 2 が ♠A に勝つ")
}

func TestHoneymoonBridge_FollowSuitIsCompulsory(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetPhaseForTest(HoneymoonBridgePhasePlay)
	h.SetLeadPlayerIdxForTest(0)
	h.SetCurrentPlayerIdxForTest(0)
	honeymoonBridgeHandOf(h, 0, NewCard(CardDesignSpade, 8, false))
	honeymoonBridgeHandOf(h, 1, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 8, false))

	require.NoError(t, h.PlayForTest(0, 0))
	assert.Equal(t, []int{0}, h.GetValidPlayIndices(1))
	assert.Error(t, h.PlayForTest(1, 1))
}

// **成立すればレベル×10、オーバートリックも点。失敗は相手に不足×10。**
func TestHoneymoonBridge_Scoring(t *testing.T) {
	for _, tc := range []struct {
		name      string
		level     int
		took      int
		wantDecl  int
		wantOther int
		made      bool
	}{
		{"ちょうど成立", 2, 8, 20, 0, true},
		{"オーバートリック", 2, 10, 30, 0, true},
		{"1 トリック不足", 3, 8, 0, 10, false},
		{"全部落とす", 1, 0, 0, 70, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHoneymoonBridge(t)
			h.SetContractForTest(0, tc.level, CardDesignHeart)
			h.SetPhaseForTest(HoneymoonBridgePhasePlay)
			h.GiveTricksForTest(0, tc.took)
			h.GiveTricksForTest(1, HoneymoonBridgeTricksPerPhase-tc.took)
			h.FinishRoundForTest()

			assert.Equal(t, tc.wantDecl, h.GetPlayer(0).GetScore())
			assert.Equal(t, tc.wantOther, h.GetPlayer(1).GetScore())
			assert.Equal(t, tc.made, h.GetLastMade())
			assert.Equal(t, tc.took, h.GetLastTricks())
		})
	}
}

func TestHoneymoonBridge_ReachingTheTargetEndsTheGame(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetContractForTest(0, 7, CardDesignHeart)
	h.SetPhaseForTest(HoneymoonBridgePhasePlay)
	h.GetPlayer(0).SetScore(h.GetConfig().Target - 70)
	h.GiveTricksForTest(0, HoneymoonBridgeTricksPerPhase)
	h.FinishRoundForTest()

	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, 0, h.GetWinnerIdx())
}

func TestHoneymoonBridge_NextRoundRotatesTheDealer(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	before := h.GetDealerIdx()
	h.SetPhaseForTest(HoneymoonBridgePhaseRoundEnd)
	h.NextRound()
	assert.Equal(t, (before+1)%HoneymoonBridgePlayerCnt, h.GetDealerIdx())
	assert.Equal(t, HoneymoonBridgePhaseDraw, h.GetPhase())
	assert.Equal(t, HoneymoonBridgeStockSize, h.GetStockSize(), "配り直すので山札が戻る")

	h.FinishGameForTest()
	after := h.GetDealerIdx()
	h.NextRound()
	assert.Equal(t, after, h.GetDealerIdx(), "終局後は進まない")
}

func TestHoneymoonBridge_RejectsOutOfTurnAndBadIndices(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	idx := h.GetCurrentPlayerIdx()
	assert.Error(t, h.PlayForTest((idx+1)%HoneymoonBridgePlayerCnt, 0), "手番でない席は打てない")
	assert.Error(t, h.PlayForTest(idx, -1))
	assert.Error(t, h.PlayForTest(idx, 999))

	h.SetPhaseForTest(HoneymoonBridgePhaseBid)
	assert.Error(t, h.PlayForTest(idx, 0), "競り中は打てない")

	h.FinishGameForTest()
	assert.Error(t, h.PlayForTest(idx, 0), "終局後は打てない")
	assert.Error(t, h.BidForTest(0, 1, CardDesignHeart), "終局後は宣言できない")
}

// **公開の入口も踏む。** `Player*` / `Cpu*` のガードが未検証のまま残らないように。
func TestHoneymoonBridge_PublicEntryPointsGuardTheTurn(t *testing.T) {
	h := newTestHoneymoonBridge(t)
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

	// 競り: 人間の番でなければ弾く。
	h.SetPhaseForTest(HoneymoonBridgePhaseBid)
	h.SetCurrentPlayerIdxForTest(1)
	assert.False(t, h.IsHumanBidTurn())
	assert.Error(t, h.PlayerBid(1, CardDesignHeart))
	assert.Error(t, h.PlayerPass())
	h.CpuBid()
	assert.NotEqual(t, 1, h.GetCurrentPlayerIdx(), "CPU が宣言して手番が移る")

	h.SetCurrentPlayerIdxForTest(0)
	assert.True(t, h.IsHumanBidTurn())
	require.NoError(t, h.PlayerBid(HoneymoonBridgeMaxLevel, 0))
	assert.Equal(t, 0, h.GetDeclarerIdx())
}

// **CPU は必ず合法手を返す。**
func TestHoneymoonBridge_CpuAlwaysChoosesLegally(t *testing.T) {
	for range 60 {
		h := NewDefaultHoneymoonBridge()
		h.Reset()
		for turns := 0; !h.GetGameEndFlag() && turns < 400; turns++ {
			switch h.GetPhase() {
			case HoneymoonBridgePhaseDraw, HoneymoonBridgePhasePlay:
				idx := h.GetCurrentPlayerIdx()
				choice := h.CpuChoiceForTest(idx)
				require.Contains(t, h.GetValidPlayIndices(idx), choice)
				require.NoError(t, h.PlayForTest(idx, choice))
			case HoneymoonBridgePhaseBid:
				idx := h.GetCurrentPlayerIdx()
				level, suit := h.CpuBidChoiceForTest(idx)
				require.NoError(t, h.BidForTest(idx, level, suit), "CPU の宣言は必ず合法")
			case HoneymoonBridgePhaseRoundEnd:
				h.NextRound()
			default:
			}
		}
	}
}

// **どの局も必ず終わる。**
func TestHoneymoonBridge_GamesTerminate(t *testing.T) {
	for range 20 {
		h := NewDefaultHoneymoonBridge()
		h.Reset()
		for turns := 0; !h.GetGameEndFlag(); turns++ {
			require.Less(t, turns, 20000, "終わらない")
			switch h.GetPhase() {
			case HoneymoonBridgePhaseDraw, HoneymoonBridgePhasePlay:
				idx := h.GetCurrentPlayerIdx()
				require.NoError(t, h.PlayForTest(idx, h.CpuChoiceForTest(idx)))
			case HoneymoonBridgePhaseBid:
				idx := h.GetCurrentPlayerIdx()
				level, suit := h.CpuBidChoiceForTest(idx)
				require.NoError(t, h.BidForTest(idx, level, suit))
			case HoneymoonBridgePhaseRoundEnd:
				h.NextRound()
			default:
			}
		}
		assert.GreaterOrEqual(t, max(h.GetPlayer(0).GetScore(), h.GetPlayer(1).GetScore()),
			h.GetConfig().Target)
	}
}

func TestHoneymoonBridge_GiveUp(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.GiveUp()
	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, 1, h.GetWinnerIdx())
	h.GiveUp()
	assert.Equal(t, 1, h.GetWinnerIdx())
}

// **助言は引き合い／競り／本番で形が違う。**
func TestHoneymoonBridge_Hint(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetCurrentPlayerIdxForTest(0)

	drawHint := h.GetHint()
	require.NotNil(t, drawHint)
	require.NotNil(t, drawHint.CardIndex)
	assert.Equal(t, "honeymoonbridgeDraw", drawHint.Reason, "引き合いは得点にならない")

	h.SetPhaseForTest(HoneymoonBridgePhaseBid)
	bidHint := h.GetHint()
	require.NotNil(t, bidHint)
	assert.Nil(t, bidHint.CardIndex)
	assert.Contains(t, []string{"honeymoonbridgeBid", "honeymoonbridgePass"}, bidHint.Reason)

	h.SetPhaseForTest(HoneymoonBridgePhasePlay)
	h.SetCurrentPlayerIdxForTest(0)
	playHint := h.GetHint()
	require.NotNil(t, playHint)
	require.NotNil(t, playHint.CardIndex)
	assert.Equal(t, "honeymoonbridgeWinTrick", playHint.Reason)
	assert.Contains(t, h.GetValidPlayIndices(0), *playHint.CardIndex, "勧める札は必ず合法")

	h.SetCurrentPlayerIdxForTest(1)
	assert.Nil(t, h.GetHint(), "人間の手番でなければ助言しない")

	h.SetCurrentPlayerIdxForTest(0)
	h.FinishGameForTest()
	assert.Nil(t, h.GetHint(), "終局後は助言しない")
}

func TestHoneymoonBridge_JSONRoundTrip(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	for range 6 {
		idx := h.GetCurrentPlayerIdx()
		require.NoError(t, h.PlayForTest(idx, h.CpuChoiceForTest(idx)))
	}

	data, err := json.Marshal(h)
	require.NoError(t, err)

	var restored HoneymoonBridge
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, h.GetStockSize(), restored.GetStockSize(), "山札の残りが消えない")
	assert.Equal(t, h.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, h.GetTrickNumber(), restored.GetTrickNumber())
	for i := range HoneymoonBridgePlayerCnt {
		assert.Equal(t, h.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
		assert.Equal(t, h.GetPlayer(i).GetScore(), restored.GetPlayer(i).GetScore())
	}
}

// **壊れたスナップショットは弾く。**
func TestHoneymoonBridge_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	base := func(t *testing.T) map[string]any {
		t.Helper()
		h := newTestHoneymoonBridge(t)
		require.NoError(t, h.PlayForTest(h.GetCurrentPlayerIdx(), 0))
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
		{"trump suit without a contract", func(m map[string]any) { m["ts"] = 3 }},
		{"declarer during the draw phase", func(m map[string]any) { m["di"] = 1 }},
		{"contract level out of range", func(m map[string]any) { m["cl"] = 99 }},
		{"current player out of range", func(m map[string]any) { m["ci"] = 9 }},
		{"dealer out of range", func(m map[string]any) { m["dl"] = -1 }},
		{"winner before the game ended", func(m map[string]any) { m["wi"] = 1 }},
		{"round number below one", func(m map[string]any) { m["rn"] = 0 }},
		{"negative trick number", func(m map[string]any) { m["tn"] = -1 }},
		{"pass count above the table", func(m map[string]any) { m["pc"] = 9 }},
		// **場札は枚数だけでなく中身も見る（#5310 の再発防止）。**
		{"a trick entry with no card", func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 0}}
		}},
		{"a trick entry with a bad seat", func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 9, "card": map[string]any{"d": 1, "v": 9, "j": false}}}
		}},
		{"config out of range", func(m map[string]any) { m["cf"] = map[string]any{"t": 0} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base(t)
			tc.mutate(m)
			data, err := json.Marshal(m)
			require.NoError(t, err)
			var restored HoneymoonBridge
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通り、使っても落ちない。**
	data, err := json.Marshal(base(t))
	require.NoError(t, err)
	var ok HoneymoonBridge
	require.NoError(t, json.Unmarshal(data, &ok))
	assert.NotPanics(t, func() {
		_ = ok.GetValidPlayIndices(ok.GetCurrentPlayerIdx())
		_ = ok.TrickWinnerForTest()
	})
}

// **本番フェーズには必ず落札者がいる。** レベルと対で 0 のまま一致していても、
// 復元して 13 トリック目を打つと `finishRound` が `h.players[-1]` で落ちる
// （レビュー指摘 PR #5312、ガードを外して実際に panic を踏んで確認した）。
func TestHoneymoonBridge_UnmarshalRejectsPlayPhaseWithoutDeclarer(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetPhaseForTest(HoneymoonBridgePhasePlay)
	h.SetStockForTest(nil)
	data, err := json.Marshal(h)
	require.NoError(t, err)
	require.Equal(t, -1, h.GetDeclarerIdx(), "落札者がいない状態を作れている")

	var restored HoneymoonBridge
	assert.Error(t, json.Unmarshal(data, &restored))

	// **負のコントロール: 落札者がいれば同じフェーズでも通り、打ち切れる。**
	h.SetContractForTest(0, 2, CardDesignHeart)
	ok, err := json.Marshal(h)
	require.NoError(t, err)
	var good HoneymoonBridge
	require.NoError(t, json.Unmarshal(ok, &good))
	good.SetTrickNumberForTest(HoneymoonBridgeTricksPerPhase - 1)
	good.SetLeadPlayerIdxForTest(0)
	good.SetCurrentPlayerIdxForTest(0)
	honeymoonBridgeHandOf(&good, 0, NewCard(CardDesignSpade, 5, false))
	honeymoonBridgeHandOf(&good, 1, NewCard(CardDesignSpade, 6, false))
	assert.NotPanics(t, func() {
		require.NoError(t, good.PlayForTest(0, 0))
		require.NoError(t, good.PlayForTest(1, 0))
	})
	assert.Equal(t, HoneymoonBridgePhaseRoundEnd, good.GetPhase(), "最後のトリックで精算まで進む")
}

// **落札者と契約レベルは対。** 片方だけ立っていたら壊れている。
func TestHoneymoonBridge_UnmarshalRejectsAHalfContract(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	h.SetPhaseForTest(HoneymoonBridgePhasePlay)
	h.SetContractForTest(0, 3, CardDesignHeart)
	h.SetStockForTest(nil)
	data, err := json.Marshal(h)
	require.NoError(t, err)

	for _, mutate := range []func(map[string]any){
		func(m map[string]any) { m["cl"] = 0; m["ts"] = 0 }, // 落札者だけ残る
		func(m map[string]any) { m["di"] = -1 },             // レベルだけ残る
	} {
		var m map[string]any
		require.NoError(t, json.Unmarshal(data, &m))
		mutate(m)
		bad, err := json.Marshal(m)
		require.NoError(t, err)
		var restored HoneymoonBridge
		assert.Error(t, json.Unmarshal(bad, &restored))
	}
}

// **競りは 0 / 1..7 のいずれか、レベル 0 ならスートも 0。**
func TestHoneymoonBridgePlayer_UnmarshalRejectsBrokenBids(t *testing.T) {
	for _, body := range []string{`{"bl":8}`, `{"bl":-1}`, `{"bl":2,"bs":9}`, `{"bl":0,"bs":3}`} {
		var p HoneymoonBridgePlayer
		assert.Error(t, json.Unmarshal([]byte(body), &p), body)
	}
	for _, body := range []string{`{"bl":0,"bs":0}`, `{"bl":1,"bs":0}`, `{"bl":7,"bs":4}`} {
		var p HoneymoonBridgePlayer
		assert.NoError(t, json.Unmarshal([]byte(body), &p), body)
	}
}

func TestHoneymoonBridgeConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultHoneymoonBridgeConfig().Validate())
	assert.NoError(t, HoneymoonBridgeConfig{Target: HoneymoonBridgeTargetMin}.Validate())
	assert.NoError(t, HoneymoonBridgeConfig{Target: HoneymoonBridgeTargetMax}.Validate())
	assert.Error(t, HoneymoonBridgeConfig{Target: HoneymoonBridgeTargetMin - 1}.Validate())
	assert.Error(t, HoneymoonBridgeConfig{Target: HoneymoonBridgeTargetMax + 1}.Validate())
}

func TestHoneymoonBridge_AccessorsAndBounds(t *testing.T) {
	h := newTestHoneymoonBridge(t)
	assert.Nil(t, h.GetPlayer(-1))
	assert.Nil(t, h.GetPlayer(99))
	assert.Empty(t, h.GetValidPlayIndices(-1))
	assert.Equal(t, HoneymoonBridgePlayerCnt, h.GetPlayerCnt())
	assert.Equal(t, HoneymoonBridgeDefaultTarget, h.GetConfig().Target)
	assert.NotEmpty(t, h.GetActionLog())
	assert.Empty(t, h.GetCurrentTrick())
	assert.Equal(t, "NT", honeymoonBridgeContractSuitStr(0))
	assert.NotEqual(t, "NT", honeymoonBridgeContractSuitStr(CardDesignHeart))
}

// **得点式は細かい** (契約レベル×10 + オーバートリック×5 / 失敗は不足×10)。
// 画面に出せるよう、実際に動いた点を残す (#5760)。
func TestHoneymoonBridgeRecordsTheRoundPoints(t *testing.T) {
	t.Run("made with overtricks", func(t *testing.T) {
		h := newTestHoneymoonBridge(t)
		h.SetContractForTest(0, 3, CardDesignSpade)
		need := h.RequiredTricks()
		// 必要数 +2 取る。
		hbGiveTricks(h, 0, need+2)
		before := h.GetPlayer(0).GetScore()

		h.FinishRoundForTest()

		want := 3*10 + 2*5
		if h.GetLastPoints() != want {
			t.Errorf("lastPoints = %d, want %d", h.GetLastPoints(), want)
		}
		// **累計スコアの動きと一致する** (受け入れ条件4)。
		if got := h.GetPlayer(0).GetScore() - before; got != want {
			t.Errorf("the declarer's score moved by %d, want %d", got, want)
		}
	})

	t.Run("down", func(t *testing.T) {
		h := newTestHoneymoonBridge(t)
		h.SetContractForTest(0, 4, CardDesignHeart)
		need := h.RequiredTricks()
		hbGiveTricks(h, 0, need-3)
		before := h.GetPlayer(1).GetScore()

		h.FinishRoundForTest()

		want := 3 * 10
		if h.GetLastPoints() != want {
			t.Errorf("lastPoints = %d, want %d", h.GetLastPoints(), want)
		}
		if got := h.GetPlayer(1).GetScore() - before; got != want {
			t.Errorf("the opponent's score moved by %d, want %d", got, want)
		}
	})
}

// hbGiveTricks は指定席に n トリック持たせる。
func hbGiveTricks(h *HoneymoonBridge, idx, n int) {
	p := h.GetPlayer(idx)
	p.ResetTricks()
	for range n {
		p.AddTrick([]*Card{})
	}
}
