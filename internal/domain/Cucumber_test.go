//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCucumberForTest(t *testing.T, n int) *Cucumber {
	t.Helper()
	cfg := DefaultCucumberConfig()
	cfg.PlayerCnt = n
	c := NewCucumber(newCucumberSeats(n), cfg)
	c.Reset()
	return c
}

// **7 枚固定。** 52 枚は 3/5/6 人で割り切れないので、人数で割りません。
func TestCucumberResetDealsSevenEach(t *testing.T) {
	for n := CucumberPlayerCntMin; n <= CucumberPlayerCntMax; n++ {
		c := newCucumberForTest(t, n)
		require.Equal(t, CucumberPhasePlay, c.GetPhase())
		for i := range n {
			assert.Equal(t, CucumberHandSize, c.GetPlayer(i).GetCardsSize(), "%d 人: 席 %d", n, i)
			assert.Zero(t, c.GetPlayer(i).GetPenalty())
		}
		// 使うのは 7×人数 枚だけ。残りは配りません。
		assert.LessOrEqual(t, CucumberHandSize*n, 52, "%d 人", n)
	}
}

// **A がいちばん強い。**
func TestCucumberRankPutsAceOnTop(t *testing.T) {
	assert.Equal(t, 14, CucumberRankForTest(NewCard(CardDesignSpade, 1, false)))
	assert.Equal(t, 13, CucumberRankForTest(NewCard(CardDesignSpade, 13, false)))
	assert.Equal(t, 2, CucumberRankForTest(NewCard(CardDesignSpade, 2, false)))
}

// **更新できるならその中から、できないならいちばん低い札 1 枚だけ。**
func TestCucumberValidPlaysFollowTheComparisonRule(t *testing.T) {
	c := newCucumberForTest(t, 4)
	c.SetCurrentPlayerIdxForTest(0)
	c.GiveHandForTest(0,
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignClover, 12, false))

	// リードは何でも出せる。
	c.SetCurrentTrickForTest(nil)
	assert.Equal(t, []int{0, 1, 2}, c.GetValidPlayIndices(0))

	// 8 が出ているので、9 と 12 だけが合法。**スートは無関係。**
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 8, false)},
	})
	assert.Equal(t, []int{1, 2}, c.GetValidPlayIndices(0))

	// 13 が出ていると更新できないので、**いちばん低い 3 の 1 枚だけ**。
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 13, false)},
	})
	assert.Equal(t, []int{0}, c.GetValidPlayIndices(0))
}

// **出す札が決まっている場面と、選べる場面を言い分けます。**
func TestCucumberPlayRejectsIllegalCards(t *testing.T) {
	c := newCucumberForTest(t, 4)
	c.SetCurrentPlayerIdxForTest(0)
	c.GiveHandForTest(0,
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 9, false))
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 8, false)},
	})

	err := c.PlayerPlay(0)
	require.Error(t, err, "更新できるのに低い札は出せない")
	assert.Contains(t, err.Error(), "超える")
	require.NoError(t, c.PlayerPlay(1))

	// 更新できない場面では、低い札以外が弾かれる。
	c.SetCurrentPlayerIdxForTest(0)
	c.GiveHandForTest(0,
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 5, false))
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 13, false)},
	})
	err = c.PlayerPlay(1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "いちばん低い")

	assert.Error(t, c.PlayerPlay(-1))
	assert.Error(t, c.PlayerPlay(99))
}

// **スート無関係で最高ランクが勝ち。同ランクは先着。**
func TestCucumberTrickWinnerIgnoresSuit(t *testing.T) {
	c := newCucumberForTest(t, 4)
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 9, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 11, false)},
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 11, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 4, false)},
	})
	assert.Equal(t, 3, c.trickWinner(), "同ランクなら先に出したほう")
}

// **失点は最終トリックだけ、取った札のランクぶん。**
func TestCucumberOnlyTheLastTrickScores(t *testing.T) {
	c := newCucumberForTest(t, 3)
	// 各自 1 枚だけにして 1 トリックで終わらせる。
	c.GiveHandForTest(0, NewCard(CardDesignSpade, 13, false))
	c.GiveHandForTest(1, NewCard(CardDesignHeart, 2, false))
	c.GiveHandForTest(2, NewCard(CardDesignClover, 3, false))
	c.SetLeadPlayerIdxForTest(0)
	c.SetCurrentPlayerIdxForTest(0)
	c.SetCurrentTrickForTest(nil)

	require.NoError(t, c.PlayForTest(0, 0))
	require.NoError(t, c.PlayForTest(1, 0))
	require.NoError(t, c.PlayForTest(2, 0))

	assert.Equal(t, CucumberPhaseRoundEnd, c.GetPhase())
	assert.Equal(t, 0, c.GetLastTrickWinnerIdx())
	// K を出して取ったので 13 点。
	assert.Equal(t, 13, c.GetLastPenalty())
	assert.Equal(t, 13, c.GetPlayer(0).GetPenalty())
	assert.Zero(t, c.GetPlayer(1).GetPenalty(), "取らなかった席は 0 点")
}

// **ラウンドの区切りは観測できます。** 読む前に配り直しません。
func TestCucumberRoundEndWaitsForTheNextDeal(t *testing.T) {
	c := newCucumberForTest(t, 3)
	c.GiveHandForTest(0, NewCard(CardDesignSpade, 5, false))
	c.GiveHandForTest(1, NewCard(CardDesignHeart, 2, false))
	c.GiveHandForTest(2, NewCard(CardDesignClover, 3, false))
	c.SetCurrentPlayerIdxForTest(0)
	c.SetCurrentTrickForTest(nil)
	for i := range 3 {
		require.NoError(t, c.PlayForTest(i, 0))
	}
	require.Equal(t, CucumberPhaseRoundEnd, c.GetPhase())
	require.Equal(t, 1, c.GetRoundNumber())

	require.NoError(t, c.NextRound())
	assert.Equal(t, CucumberPhasePlay, c.GetPhase())
	assert.Equal(t, 2, c.GetRoundNumber())
	assert.Equal(t, CucumberHandSize, c.GetPlayer(0).GetCardsSize())
	// **最終トリックを取った席が次のリード。**
	assert.Equal(t, 0, c.GetLeadPlayerIdx())
	// 失点は持ち越す。
	assert.Equal(t, 5, c.GetPlayer(0).GetPenalty())
}

// **失点上限に達したら終わり。** 少ない人の勝ちです。
func TestCucumberEndsAtTheTargetScore(t *testing.T) {
	c := newCucumberForTest(t, 3)
	c.GetPlayer(1).SetPenalty(c.GetConfig().TargetScore - 2)
	c.GiveHandForTest(0, NewCard(CardDesignSpade, 2, false))
	c.GiveHandForTest(1, NewCard(CardDesignHeart, 13, false))
	c.GiveHandForTest(2, NewCard(CardDesignClover, 3, false))
	c.SetCurrentPlayerIdxForTest(0)
	c.SetCurrentTrickForTest(nil)
	for i := range 3 {
		require.NoError(t, c.PlayForTest(i, 0))
	}

	assert.True(t, c.GetGameEndFlag())
	assert.Equal(t, CucumberPhaseGameEnd, c.GetPhase())
	assert.Equal(t, 0, c.GetWinnerIdx(), "失点 0 の席が勝つ")
	assert.Error(t, c.NextRound(), "終局後は配り直せない")
}

func TestCucumberGiveUp(t *testing.T) {
	c := newCucumberForTest(t, 4)
	c.GiveUp()
	assert.True(t, c.GetGameEndFlag())
	assert.NotEqual(t, 0, c.GetWinnerIdx(), "投了した人は勝たない")

	c.GiveUp()
	assert.True(t, c.GetGameEndFlag())
}

// **CPU は更新できるならいちばん低い更新札を出す。** 高い札を終盤に残さない。
func TestCucumberCpuPlaysTheCheapestBeat(t *testing.T) {
	c := newCucumberForTest(t, 4)
	c.SetCurrentPlayerIdxForTest(1)
	c.GiveHandForTest(1,
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignClover, 14, false))
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignDiamond, 9, false)},
	})
	assert.Equal(t, 1, c.CpuChoiceForTest(1), "10 で足りるので A は温存")
}

func TestCucumberHint(t *testing.T) {
	c := newCucumberForTest(t, 4)
	c.SetCurrentPlayerIdxForTest(0)
	c.SetCurrentTrickForTest(nil)
	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "cucumberLead", h.Reason)

	c.GiveHandForTest(0,
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 10, false))
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 9, false)},
	})
	h = c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "cucumberBeat", h.Reason)

	// 更新できない場面は「決まっている」と言う。
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 13, false)},
	})
	h = c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "cucumberForced", h.Reason)

	c.GiveUp()
	assert.Nil(t, c.GetHint(), "終局後は助言しない")
}

// **全人数を終局まで指して、1 手ごとに保存して読み直す。**
func TestCucumber_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for n := CucumberPlayerCntMin; n <= CucumberPlayerCntMax; n++ {
		for range 5 {
			c := newCucumberForTest(t, n)
			for turns := 0; ; turns++ {
				require.Less(t, turns, 5000, "%d 人: 終わらない", n)

				data, err := json.Marshal(c)
				require.NoError(t, err)
				var back Cucumber
				require.NoError(t, json.Unmarshal(data, &back),
					"%d 人 %d 手目: 書き込み側が codec の不変条件を破った", n, turns)

				if c.GetGameEndFlag() {
					break
				}
				switch c.GetPhase() {
				case CucumberPhasePlay:
					idx := c.GetCurrentPlayerIdx()
					require.NoError(t, c.PlayForTest(idx, c.CpuChoiceForTest(idx)))
				case CucumberPhaseRoundEnd:
					require.NoError(t, c.NextRound())
				default:
					t.Fatalf("%d 人: 進めないフェーズ %d", n, c.GetPhase())
				}
			}
		}
	}
}

// **判定はドメインに 1 か所だけ。**
//
// 「合法手が 1 つ = 更新できない」は偽なので、UI 側で数え直すと片方だけ直り
// 損ねます。#5320 のレビュー指摘を受けて公開したのがこのメソッドです。
func TestCucumberIsForcedLowest(t *testing.T) {
	c := newCucumberForTest(t, 4)
	c.SetCurrentPlayerIdxForTest(0)

	// リードは決まっていない。
	c.SetCurrentTrickForTest(nil)
	assert.False(t, c.IsForcedLowest(0))

	// **合法手が 1 つでも、それが更新するなら forced ではない。**
	c.GiveHandForTest(0,
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 10, false))
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 9, false)},
	})
	require.Len(t, c.GetValidPlayIndices(0), 1, "合法手はちょうど 1 つ")
	assert.False(t, c.IsForcedLowest(0), "その 1 枚は更新できるので forced ではない")

	// 更新できないときだけ真。
	c.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 13, false)},
	})
	assert.True(t, c.IsForcedLowest(0))

	// ラウンドの区切りや終局では判定しない。
	c.SetPhaseForTest(CucumberPhaseRoundEnd)
	assert.False(t, c.IsForcedLowest(0))
}
