//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newContGame(t *testing.T) *ContinentalRummy {
	t.Helper()
	c := NewDefaultContinentalRummy()
	c.Reset()
	return c
}

// **2〜5 人卓は 2 組 + ジョーカー 2 枚 = 106 枚。** #5464 の「人数 − 1 組」ではない。
func TestContinentalRummy_DealUsesTwoDecksAndTwoJokers(t *testing.T) {
	c := newContGame(t)
	require.Equal(t, 4, c.GetPlayerCnt())

	dealt := 0
	for i := 0; i < c.GetPlayerCnt(); i++ {
		n := c.GetPlayer(i).GetCardsSize()
		dealt += n
		// 人間以外は Reset のうちに手番を打っているので 15 か 16 枚。
		assert.GreaterOrEqual(t, n, ContinentalRummyHandSize)
		assert.LessOrEqual(t, n, ContinentalRummyHandSize+1)
	}
	assert.Equal(t, ContinentalRummyHandSize*ContinentalRummyPlayerCnt, dealt-continentalExtraHeld(c),
		"1 人 15 枚配れていない")

	// 山 + 手札 + 捨て札 を足すと 106 枚。
	assert.Equal(t, CardCnt*ContinentalRummyDeckCnt+ContinentalRummyJokerCnt,
		dealt+c.GetStockCount()+c.discardCountForTest(), "106 枚になっていない")
}

// continentalExtraHeld は「引いてまだ捨てていない」ぶんの余分な札を数える。
func continentalExtraHeld(c *ContinentalRummy) int {
	n := 0
	for i := 0; i < c.GetPlayerCnt(); i++ {
		if held := c.GetPlayer(i).GetCardsSize(); held > ContinentalRummyHandSize {
			n += held - ContinentalRummyHandSize
		}
	}
	return n
}

func TestContinentalRummy_OpensOnTheHumanDraw(t *testing.T) {
	c := newContGame(t)
	require.False(t, c.GetGameEndFlag())
	// **Reset は人間の番まで進める。** 進めないと最初の手番で固まる。
	if c.GetPhase() == ContinentalRummyPhaseRoundEnd {
		t.Skip("CPU が Reset のうちに上がった配り")
	}
	assert.Equal(t, ContinentalRummyHumanIdx, c.GetCurrentPlayerIdx())
	assert.Equal(t, ContinentalRummyPhaseDraw, c.GetPhase())
	assert.True(t, c.IsHumanTurn())
}

func TestContinentalRummy_DrawThenDiscard(t *testing.T) {
	c := newContGame(t)
	if c.GetPhase() != ContinentalRummyPhaseDraw {
		t.Skip("CPU が Reset のうちに上がった配り")
	}
	before := c.GetPlayer(ContinentalRummyHumanIdx).GetCardsSize()
	require.NoError(t, c.DrawStock())
	assert.Equal(t, before+1, c.GetPlayer(ContinentalRummyHumanIdx).GetCardsSize())
	assert.Equal(t, ContinentalRummyPhaseDiscard, c.GetPhase())

	// 引く前に捨てられない / 捨てる前に二度引けない。
	assert.Error(t, c.DrawStock(), "二度引けてしまっている")
	require.NoError(t, c.Discard(0))
	assert.Equal(t, before, c.GetPlayer(ContinentalRummyHumanIdx).GetCardsSize())
}

func TestContinentalRummy_DiscardRejectsAnIndexOutOfRange(t *testing.T) {
	c := newContGame(t)
	if c.GetPhase() != ContinentalRummyPhaseDraw {
		t.Skip("CPU が Reset のうちに上がった配り")
	}
	require.NoError(t, c.DrawStock())
	assert.Error(t, c.Discard(-1))
	assert.Error(t, c.Discard(999))
}

// **上がりは 15 枚を一度に並べるときだけ。** 形にならない手は断る。
func TestContinentalRummy_GoOutRefusesAHandThatIsNotALayout(t *testing.T) {
	c := newContGame(t)
	c.SetPhaseForTest(ContinentalRummyPhaseDiscard)
	c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
	p := c.GetPlayer(ContinentalRummyHumanIdx)
	p.ResetRound()
	// セット 5 組 + 余り 1 枚 = 16 枚。ランが 1 本も無い。
	for _, v := range []int{2, 5, 8, 11, 13} {
		for _, d := range []int{CardDesignSpade, CardDesignHeart, CardDesignClover} {
			p.AddCard(contCard(d, v))
		}
	}
	p.AddCard(contCard(CardDesignDiamond, 4))
	require.Equal(t, ContinentalRummyHandSize+1, p.GetCardsSize())

	_, ok := c.CanGoOut()
	assert.False(t, ok, "セットだけの手で上がれると案内している")
	assert.Error(t, c.GoOut(0), "セットだけの手で上がれてしまっている")
}

func TestContinentalRummy_GoOutLaysAllFifteenAndSettles(t *testing.T) {
	c := newContGame(t)
	c.SetPhaseForTest(ContinentalRummyPhaseDiscard)
	c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
	p := c.GetPlayer(ContinentalRummyHumanIdx)
	p.ResetRound()
	for _, run := range [][]*Card{
		contRun(CardDesignSpade, 2, 3), contRun(CardDesignSpade, 7, 3),
		contRun(CardDesignHeart, 4, 3), contRun(CardDesignClover, 9, 3),
		contRun(CardDesignDiamond, 5, 3),
	} {
		for _, card := range run {
			p.AddCard(card)
		}
	}
	p.AddCard(contCard(CardDesignDiamond, 13)) // 捨てる 1 枚
	require.Equal(t, ContinentalRummyHandSize+1, p.GetCardsSize())

	idx, ok := c.CanGoOut()
	require.True(t, ok, "上がれる手を上がれないと言っている")
	assert.Equal(t, ContinentalRummyHandSize, idx, "捨てるべき 1 枚が違う")
	require.NoError(t, c.GoOut(idx))

	assert.Equal(t, ContinentalRummyPhaseRoundEnd, c.GetPhase())
	assert.Len(t, p.GetMelds(), 5)
	assert.Equal(t, 0, p.GetCardsSize(), "手札が残っている")

	res := c.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, ContinentalRummyHumanIdx, res.WinnerIdx)
	// **取り立てるのは相手 1 人あたり。** 3 人から集めるので 3 倍。
	assert.Equal(t, res.PerOpponent*(ContinentalRummyPlayerCnt-1), res.Total)
	assert.Equal(t, res.Total, p.GetScore())
}

// **加点は「どう上がったか」で決まる。残り札は数えない。**
func TestContinentalRummy_BonusesRewardHowYouWentOut(t *testing.T) {
	setup := func(t *testing.T, hand []*Card, drew bool) *ContinentalRummyRoundResult {
		t.Helper()
		c := newContGame(t)
		c.SetPhaseForTest(ContinentalRummyPhaseDiscard)
		c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
		c.drewThisRound = make([]bool, ContinentalRummyPlayerCnt)
		c.drewThisRound[ContinentalRummyHumanIdx] = drew
		c.actionLog = nil
		p := c.GetPlayer(ContinentalRummyHumanIdx)
		p.ResetRound()
		for _, card := range hand {
			p.AddCard(card)
		}
		p.AddCard(contCard(CardDesignDiamond, 13))
		idx, ok := c.CanGoOut()
		require.True(t, ok, "仕込んだ手で上がれない -- 前提が崩れている")
		require.NoError(t, c.GoOut(idx))
		return c.GetLastResult()
	}
	pointsFor := func(res *ContinentalRummyRoundResult, key string) int {
		for _, b := range res.Bonuses {
			if b.Key == key {
				return b.Points
			}
		}
		return 0
	}

	plain := []*Card{}
	for _, run := range [][]*Card{
		contRun(CardDesignSpade, 2, 3), contRun(CardDesignSpade, 7, 3),
		contRun(CardDesignHeart, 4, 3), contRun(CardDesignClover, 9, 3),
		contRun(CardDesignDiamond, 5, 3),
	} {
		plain = append(plain, run...)
	}

	t.Run("no joker pays ten, a joker pays two each", func(t *testing.T) {
		res := setup(t, plain, true)
		assert.Equal(t, ContinentalRummyNoJokerPoints, pointsFor(res, "noJoker"))
		assert.Equal(t, 0, pointsFor(res, "joker"))

		withJoker := append([]*Card(nil), plain...)
		withJoker[0] = contJoker()
		res2 := setup(t, withJoker, true)
		assert.Equal(t, ContinentalRummyJokerPoints, pointsFor(res2, "joker"))
		assert.Equal(t, 0, pointsFor(res2, "noJoker"), "ジョーカーを使ったのに無使用の加点が付いている")
	})

	// **「配られた 15 枚のまま」と「最初の手番で」は別の加点。**
	t.Run("going out on the dealt fifteen pays more than going out on turn one", func(t *testing.T) {
		dealt := setup(t, plain, false)
		assert.Equal(t, ContinentalRummyDealtPoints, pointsFor(dealt, "dealt"))
		assert.Equal(t, 0, pointsFor(dealt, "firstTurn"))

		firstTurn := setup(t, plain, true)
		assert.Equal(t, ContinentalRummyFirstTurnPoints, pointsFor(firstTurn, "firstTurn"))
		assert.Equal(t, 0, pointsFor(firstTurn, "dealt"))
		assert.Greater(t, pointsFor(dealt, "dealt"), pointsFor(firstTurn, "firstTurn"))
	})

	t.Run("an all-one-suit hand pays ten", func(t *testing.T) {
		oneSuit := []*Card{}
		for _, run := range [][]*Card{
			contRun(CardDesignSpade, 1, 3), contRun(CardDesignSpade, 4, 3),
			contRun(CardDesignSpade, 7, 3), contRun(CardDesignSpade, 10, 3),
			{contCard(CardDesignSpade, 2), contCard(CardDesignSpade, 3), contCard(CardDesignSpade, 4)},
		} {
			oneSuit = append(oneSuit, run...)
		}
		res := setup(t, oneSuit, true)
		assert.Equal(t, ContinentalRummyOneSuitPoints, pointsFor(res, "oneSuit"))
		// 負のコントロール: 混ざっていれば付かない。
		assert.Equal(t, 0, pointsFor(setup(t, plain, true), "oneSuit"))
	})

	t.Run("winning itself always pays one", func(t *testing.T) {
		assert.Equal(t, ContinentalRummyWinPoints, pointsFor(setup(t, plain, true), "win"))
	})
}

func TestContinentalRummy_RoundsAndGameEnd(t *testing.T) {
	cfg := DefaultContinentalRummyConfig()
	cfg.TotalRounds = 1
	c := NewContinentalRummy(cfg)
	c.Reset()
	c.SetPhaseForTest(ContinentalRummyPhaseDiscard)
	c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
	p := c.GetPlayer(ContinentalRummyHumanIdx)
	p.ResetRound()
	for _, run := range [][]*Card{
		contRun(CardDesignSpade, 2, 3), contRun(CardDesignSpade, 7, 3),
		contRun(CardDesignHeart, 4, 3), contRun(CardDesignClover, 9, 3),
		contRun(CardDesignDiamond, 5, 3),
	} {
		for _, card := range run {
			p.AddCard(card)
		}
	}
	p.AddCard(contCard(CardDesignDiamond, 13))
	idx, ok := c.CanGoOut()
	require.True(t, ok)
	require.NoError(t, c.GoOut(idx))

	// 1 ラウンド設定なので、そのまま終局。**一番多く集めた席の勝ち。**
	assert.True(t, c.GetGameEndFlag())
	assert.Equal(t, ContinentalRummyPhaseGameEnd, c.GetPhase())
	assert.Equal(t, ContinentalRummyHumanIdx, c.GetWinnerIdx())

	// 終局後は何も動かない。
	before := p.GetCardsSize()
	assert.Error(t, c.DrawStock())
	c.NextRound()
	assert.Equal(t, before, p.GetCardsSize())
	assert.Equal(t, 1, c.GetRoundNumber())
}

// **山が尽きたら捨て札を裏返して積み直す。**
//
// 原典は山が尽きたときのことを書いていないが、そこで流局にすると実測で
// 200 局中 161 局が誰も上がれずに流れた。上がりに届かない盤を配り続ける
// くらいなら、ラミー系で普通に行われる積み直しを採る。
func TestContinentalRummy_StockExhaustedRecyclesTheDiscards(t *testing.T) {
	c := newContGame(t)
	c.SetPhaseForTest(ContinentalRummyPhaseDraw)
	c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
	c.SetStockForTest(nil)
	c.SetDiscardForTest([]*Card{
		contCard(CardDesignSpade, 2), contCard(CardDesignHeart, 5), contCard(CardDesignClover, 9)})
	require.Equal(t, 0, c.GetRecycleCount())

	require.NoError(t, c.DrawStock())
	assert.Equal(t, ContinentalRummyPhaseDiscard, c.GetPhase(), "流局してしまっている")
	assert.Equal(t, 1, c.GetRecycleCount())
	// **頭の 1 枚は場に残す。** 積み直したことで場が空になってはいけない。
	assert.NotNil(t, c.GetDiscardTop())
	assert.Equal(t, CardDesignClover, c.GetDiscardTop().GetDesign())
}

// **積み直しは無制限ではない。** 上限の無いループを Worker に持ち込まない。
func TestContinentalRummy_RecyclingIsBoundedThenTheRoundWashesOut(t *testing.T) {
	c := newContGame(t)
	c.SetPhaseForTest(ContinentalRummyPhaseDraw)
	c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
	for i := 0; i < continentalRummyMaxRecycles; i++ {
		c.SetStockForTest(nil)
		c.SetDiscardForTest([]*Card{contCard(CardDesignSpade, 2), contCard(CardDesignHeart, 5)})
		c.SetPhaseForTest(ContinentalRummyPhaseDraw)
		require.NoError(t, c.DrawStock())
	}
	require.Equal(t, continentalRummyMaxRecycles, c.GetRecycleCount())

	// 上限に達したら、次に尽きたところで流局する。
	c.SetStockForTest(nil)
	c.SetDiscardForTest([]*Card{contCard(CardDesignSpade, 2), contCard(CardDesignHeart, 5)})
	c.SetPhaseForTest(ContinentalRummyPhaseDraw)
	require.NoError(t, c.DrawStock())
	assert.Equal(t, ContinentalRummyPhaseRoundEnd, c.GetPhase())
	res := c.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, -1, res.WinnerIdx, "誰も上がっていないのに勝者が付いている")
	assert.Equal(t, 0, res.Total)
}

func TestContinentalRummy_ConfigValidation(t *testing.T) {
	cfg := DefaultContinentalRummyConfig()
	assert.NoError(t, cfg.Validate())
	bad := cfg
	bad.TotalRounds = 0
	assert.Error(t, bad.Validate())
	bad2 := cfg
	bad2.TotalRounds = ContinentalRummyMaxRounds + 1
	assert.Error(t, bad2.Validate())
	bad3 := cfg
	bad3.CpuDifficulty = 9
	assert.Error(t, bad3.Validate())
}

// **非公開フィールドしか無い型は MarshalJSON が無いと `{}` になる。**
func TestContinentalRummy_JSONRoundTrip(t *testing.T) {
	c := newContGame(t)
	if c.GetPhase() == ContinentalRummyPhaseDraw {
		require.NoError(t, c.DrawStock())
	}
	data, err := json.Marshal(c)
	require.NoError(t, err)
	assert.Greater(t, len(data), 2, "snapshot が `{}` -- MarshalJSON が無い")

	restored := new(ContinentalRummy)
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, c.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	for i := 0; i < c.GetPlayerCnt(); i++ {
		assert.Equal(t, c.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
		assert.Equal(t, c.GetPlayer(i).GetScore(), restored.GetPlayer(i).GetScore())
	}
	// 復元した盤で指し続けられる。
	if restored.GetPhase() == ContinentalRummyPhaseDiscard {
		require.NoError(t, restored.Discard(0))
	}
}

func TestContinentalRummy_HintSpeaksOnlyOnTheHumanTurn(t *testing.T) {
	c := newContGame(t)
	if c.GetPhase() == ContinentalRummyPhaseRoundEnd {
		t.Skip("CPU が Reset のうちに上がった配り")
	}
	hint := c.GetHint()
	require.NotNil(t, hint)
	assert.Contains(t, []string{"draw_stock", "take_discard"}, hint.Reason)

	require.NoError(t, c.DrawStock())
	hint2 := c.GetHint()
	require.NotNil(t, hint2)
	assert.Contains(t, []string{"go_out", "discard_loose"}, hint2.Reason)
	assert.GreaterOrEqual(t, hint2.DiscardIdx, 0)

	// 負のコントロール: 相手の番では黙る。
	c.SetCurrentIdxForTest(1)
	assert.Nil(t, c.GetHint())
}

// contWinningFifteen は上がれる 15 枚を返す。
func contWinningFifteen() []*Card {
	return contHand(
		contRun(CardDesignSpade, 2, 3), contRun(CardDesignSpade, 7, 3),
		contRun(CardDesignHeart, 4, 3), contRun(CardDesignClover, 9, 3),
		contRun(CardDesignDiamond, 5, 3))
}

// **配られた 15 枚のまま上がれること (レビュー指摘)。**
//
// 以前は「引く → 捨てる」しか入口が無く、draw() が必ず drewThisRound を
// 立ててから discard フェーズにしていたので、10 点の "dealt" 加点は**公開
// API からは一度も出せなかった** ── 私有フィールドを直接触るテストだけが
// 通っていて、到達可能性を誰も見ていなかった。
func TestContinentalRummy_GoesOutOnTheDealtFifteen(t *testing.T) {
	c := newContGame(t)
	c.SetPhaseForTest(ContinentalRummyPhaseDraw)
	c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
	p := c.GetPlayer(ContinentalRummyHumanIdx)
	p.ResetRound()
	for _, card := range contWinningFifteen() {
		p.AddCard(card)
	}
	require.Equal(t, ContinentalRummyHandSize, p.GetCardsSize())
	require.True(t, c.CanGoOutOnTheDeal(), "配られたままで上がれると案内していない")

	stock, discards := c.GetStockCount(), c.discardCountForTest()
	require.NoError(t, c.GoOut(-1))

	assert.Equal(t, ContinentalRummyPhaseRoundEnd, c.GetPhase())
	assert.Len(t, p.GetMelds(), 5)
	assert.Equal(t, 0, p.GetCardsSize())
	// **引かず、捨てない。** 山も捨て札も動いていないこと。
	assert.Equal(t, stock, c.GetStockCount(), "引かずに上がったのに山が減っている")
	assert.Equal(t, discards, c.discardCountForTest(), "捨てずに上がったのに捨て札が増えている")

	res := c.GetLastResult()
	require.NotNil(t, res)
	var dealt, firstTurn int
	for _, b := range res.Bonuses {
		switch b.Key {
		case "dealt":
			dealt = b.Points
		case "firstTurn":
			firstTurn = b.Points
		}
	}
	assert.Equal(t, ContinentalRummyDealtPoints, dealt, "「配られた 15 枚のまま」の加点が出ていない")
	assert.Equal(t, 0, firstTurn, "引かずに上がったのに「最初の手番で」の加点も付いている")
	assert.Greater(t, dealt, ContinentalRummyFirstTurnPoints, "引かない上がりのほうが軽い")
}

// 負のコントロール: 形になっていない 15 枚では引かずに上がれない。
func TestContinentalRummy_CannotGoOutOnTheDealWithABrokenHand(t *testing.T) {
	c := newContGame(t)
	c.SetPhaseForTest(ContinentalRummyPhaseDraw)
	c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
	p := c.GetPlayer(ContinentalRummyHumanIdx)
	p.ResetRound()
	hand := contWinningFifteen()
	hand[14] = contCard(CardDesignDiamond, 13) // 1 組を壊す
	for _, card := range hand {
		p.AddCard(card)
	}
	assert.False(t, c.CanGoOutOnTheDeal())
	assert.Error(t, c.GoOut(-1))
	assert.Equal(t, ContinentalRummyPhaseDraw, c.GetPhase(), "断ったのにフェーズが進んでいる")
}

// **「最初の手番で」の加点はラウンドをまたいでも出ること (レビュー指摘)。**
//
// 以前は棋譜の discard を数えていて、棋譜は startRound で消えないので、
// 2 ラウンド目以降は前のラウンドの捨て札まで数え、7 点が事実上 1 ラウンド目
// にしか出なかった。
func TestContinentalRummy_FirstTurnBonusSurvivesIntoLaterRounds(t *testing.T) {
	cfg := DefaultContinentalRummyConfig()
	cfg.TotalRounds = 5
	c := NewContinentalRummy(cfg)
	c.Reset()

	// 1 ラウンド目に人間が何度か捨てて、棋譜に discard を積む。
	for i := 0; i < 3; i++ {
		c.SetPhaseForTest(ContinentalRummyPhaseDraw)
		c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
		require.NoError(t, c.DrawStock())
		require.NoError(t, c.Discard(0))
	}
	require.NotEmpty(t, c.GetActionLog())

	// 2 ラウンド目に入り、1 手番目に引いて上がる。
	c.SetPhaseForTest(ContinentalRummyPhaseRoundEnd)
	c.NextRound()
	require.Equal(t, 2, c.GetRoundNumber())
	c.SetPhaseForTest(ContinentalRummyPhaseDiscard)
	c.SetCurrentIdxForTest(ContinentalRummyHumanIdx)
	c.drewThisRound = make([]bool, ContinentalRummyPlayerCnt)
	c.drewThisRound[ContinentalRummyHumanIdx] = true
	p := c.GetPlayer(ContinentalRummyHumanIdx)
	p.ResetRound()
	for _, card := range contWinningFifteen() {
		p.AddCard(card)
	}
	p.AddCard(contCard(CardDesignDiamond, 13))
	idx, ok := c.CanGoOut()
	require.True(t, ok)
	require.NoError(t, c.GoOut(idx))

	firstTurn := 0
	for _, b := range c.GetLastResult().Bonuses {
		if b.Key == "firstTurn" {
			firstTurn = b.Points
		}
	}
	assert.Equal(t, ContinentalRummyFirstTurnPoints, firstTurn,
		"2 ラウンド目で「最初の手番で」の加点が消えている -- 棋譜を数えている")
}
