//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCostlyGame(t *testing.T) *CostlyColours {
	t.Helper()
	c := NewDefaultCostlyColours()
	c.Reset()
	return c
}

// **配るのは 3 枚だけ、次の 1 枚は表に返す。** #5461 は枚数に触れていないが、
// Cribbage の 6 枚+クリブとはまったく別の形。
func TestCostlyColours_DealsThreeAndTurnsOneUp(t *testing.T) {
	c := newCostlyGame(t)
	assert.Equal(t, 1, c.GetDealNumber())
	for i := 0; i < CostlyColoursPlayerCnt; i++ {
		assert.Equal(t, CostlyColoursHandSize, c.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
	}
	assert.NotNil(t, c.GetTurnUp(), "トランプが表に返っていない")
	// **交換フェーズから始まる。** ここを飛ばすと mog が無いゲームになる。
	assert.Equal(t, CostlyColoursPhaseMog, c.GetPhase())
	assert.True(t, c.IsHumanTurn())
	assert.Equal(t, CostlyColoursPlayerCnt-1, c.GetDealerIdx(), "親が最後の席でない")
}

// **表に返った J は親の 4 点。** 打ち始める前に入る ("for his heels")。
func TestCostlyColours_HeelsPaysTheDealer(t *testing.T) {
	seen := false
	for try := 0; try < 300 && !seen; try++ {
		c := NewDefaultCostlyColours()
		c.Reset()
		if c.GetTurnUp() == nil || c.GetTurnUp().GetValue() != 11 {
			continue
		}
		seen = true
		assert.Equal(t, CostlyHeelsPoints, c.GetPlayer(c.GetDealerIdx()).GetScore(),
			"J が表に返ったのに親へ 4 点入っていない")
		assert.Equal(t, 0, c.GetPlayer((c.GetDealerIdx()+1)%CostlyColoursPlayerCnt).GetScore(),
			"非親にも点が入っている")
	}
	require.True(t, seen, "300 回配っても J が表に返らない — 判定が死んでいる")
}

// **交換を断ると相手に 1 点。** 断った側ではない。
func TestCostlyColours_RefusingTheMogPaysTheOpponent(t *testing.T) {
	c := newCostlyGame(t)
	before0 := c.GetPlayer(0).GetScore()
	before1 := c.GetPlayer(1).GetScore()
	require.NoError(t, c.PlayerMog(false))
	assert.Equal(t, before0, c.GetPlayer(0).GetScore(), "断った側に点が入っている")
	assert.Equal(t, before1+CostlyMogRefusalPoints, c.GetPlayer(1).GetScore(),
		"断られた側に 1 点入っていない")
	assert.Equal(t, CostlyColoursPhasePlay, c.GetPhase(), "数え上げへ進んでいない")
}

// **交換しても手札は 3 枚のまま。** 1 枚ずつ取り替えるだけ。
func TestCostlyColours_MogKeepsThreeCardsEach(t *testing.T) {
	c := newCostlyGame(t)
	require.NoError(t, c.PlayerMog(true))
	for i := 0; i < CostlyColoursPlayerCnt; i++ {
		assert.Equal(t, CostlyColoursHandSize, c.GetPlayer(i).GetCardsSize(),
			"交換後に席 %d の手札が %d 枚でない", i, CostlyColoursHandSize)
		assert.True(t, c.GetPlayer(i).IsMoggedIn(), "席 %d が受け取っていない", i)
	}
	assert.Equal(t, CostlyColoursPhasePlay, c.GetPhase())
}

// **31 を超える札は出せない。**
func TestCostlyColours_CannotExceedThirtyOne(t *testing.T) {
	c := newCostlyGame(t)
	require.NoError(t, c.PlayerMog(false))
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(ccCard(CardDesignSpade, 10))
	c.GetPlayer(0).AddCard(ccCard(CardDesignHeart, 3))
	// 28 + 10 = 38 で超える / 28 + 3 = 31 ちょうど。
	c.SetTotalForTest(28)
	c.SetCurrentForTest(0)

	assert.Equal(t, []int{1}, c.PlayableIdxs(0), "29 に 10 が足せてしまっている")
	assert.Error(t, c.PlayerPlay(0), "31 を超える札が通っている")
	require.NoError(t, c.PlayerPlay(1))
	assert.Equal(t, 0, c.GetTotal(), "31 ちょうどで数え上げが畳まれていない")
}

// **15・25・31 はどれも作った枚数ぶん点になる。** 25 が Cribbage との
// 分かれ目。
func TestCostlyColours_PegsAtEveryMark(t *testing.T) {
	for _, tc := range []struct {
		name  string
		start int
		card  int
		want  int
	}{
		{"fifteen", 8, 7, 2},
		{"twenty-five", 18, 7, 2},
		{"thirty-one", 24, 7, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCostlyGame(t)
			require.NoError(t, c.PlayerMog(false))
			c.GetPlayer(0).Reset()
			c.GetPlayer(0).AddCard(ccCard(CardDesignSpade, tc.card))
			c.GetPlayer(0).AddCard(ccCard(CardDesignHeart, 4))
			c.GetPlayer(1).Reset()
			c.GetPlayer(1).AddCard(ccCard(CardDesignClover, 4))
			// 累計を作るため、既に 1 枚出ている体にする。
			c.SetTotalForTest(tc.start)
			c.pile = []*Card{ccCard(CardDesignDiamond, 10)}
			c.SetCurrentForTest(0)

			before := c.GetPlayer(0).GetScore()
			require.NoError(t, c.PlayerPlay(0))
			assert.Equal(t, before+tc.want, c.GetPlayer(0).GetScore(),
				"%s で作った枚数ぶん入っていない", tc.name)
		})
	}
}

// **ゴーは出した側の 1 点。**
func TestCostlyColours_GoPaysThePlayerWhoCanContinue(t *testing.T) {
	c := newCostlyGame(t)
	require.NoError(t, c.PlayerMog(false))
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(ccCard(CardDesignSpade, 2))
	c.GetPlayer(0).AddCard(ccCard(CardDesignHeart, 2))
	c.GetPlayer(1).Reset()
	c.GetPlayer(1).AddCard(ccCard(CardDesignClover, 10))
	c.SetTotalForTest(20)
	c.pile = []*Card{ccCard(CardDesignDiamond, 10), ccCard(CardDesignSpade, 10)}
	c.SetCurrentForTest(0)

	before := c.GetPlayer(0).GetScore()
	require.NoError(t, c.PlayerPlay(0)) // 22。相手は 10 が出せない。
	assert.Greater(t, c.GetPlayer(0).GetScore(), before, "ゴーの 1 点が入っていない")
	assert.Equal(t, 0, c.GetCurrentPlayerIdx(), "出せる側に手番が戻っていない")
}

// **ショーは手札 3 枚 + 表の 1 枚で数える。** 4 枚同スートが成立しないと
// ゲーム名の役が永遠に出ない。
func TestCostlyColours_ShowCountsTheTurnUp(t *testing.T) {
	c := newCostlyGame(t)
	c.SetPhaseForTest(CostlyColoursPhasePlay)
	c.SetTurnUpForTest(ccCard(CardDesignSpade, 12))
	for i := 0; i < CostlyColoursPlayerCnt; i++ {
		c.GetPlayer(i).ResetDeal()
	}
	// 席 0 は ♠ を 3 枚出し切った → 表の ♠Q と合わせて 4 枚同スート。
	for _, v := range []int{3, 5, 9} {
		c.GetPlayer(0).AddPlayed(ccCard(CardDesignSpade, v))
	}
	// 席 1 はばらばら。
	c.GetPlayer(1).AddPlayed(ccCard(CardDesignHeart, 4))
	c.GetPlayer(1).AddPlayed(ccCard(CardDesignClover, 6))
	c.GetPlayer(1).AddPlayed(ccCard(CardDesignDiamond, 8))

	c.finishDeal()
	res := c.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, CostlyComboCostlyColours, res.Combos[0],
		"表の 1 枚を数えず 4 枚同スートが成立していない")
	require.Len(t, res.Lines, 3)
	byKey := map[string][]int{}
	for _, l := range res.Lines {
		byKey[l.Key] = l.Points
	}
	assert.Equal(t, CostlyColoursPoints, byKey["colour"][0])
	assert.Equal(t, 0, byKey["colour"][1], "ばらばらの手に色役が付いている")
	assert.Equal(t, CostlyColoursPhaseShow, c.GetPhase())
}

// **表に返った J は手札の J として二重に数えない。** 既に親の 4 点になっている。
func TestCostlyColours_TurnUpJackIsNotCountedTwice(t *testing.T) {
	c := newCostlyGame(t)
	c.SetPhaseForTest(CostlyColoursPhasePlay)
	c.SetTurnUpForTest(ccCard(CardDesignSpade, 11))
	for i := 0; i < CostlyColoursPlayerCnt; i++ {
		c.GetPlayer(i).ResetDeal()
		for _, v := range []int{3, 5, 9} {
			c.GetPlayer(i).AddPlayed(ccCard(CardDesignHeart, v))
		}
	}
	c.finishDeal()
	byKey := map[string][]int{}
	for _, l := range c.GetLastResult().Lines {
		byKey[l.Key] = l.Points
	}
	assert.Equal(t, 0, byKey["jackDeuce"][0], "表の J が手札の J として数えられている")
	assert.Equal(t, 0, byKey["jackDeuce"][1])
}

// **同点では終局しない。**
func TestCostlyColours_ATieDoesNotEndTheMatch(t *testing.T) {
	c := newCostlyGame(t)
	for i := 0; i < CostlyColoursPlayerCnt; i++ {
		c.GetPlayer(i).ResetScore()
		c.GetPlayer(i).AddScore(c.GetConfig().TargetScore)
	}
	c.gameEndFlag = false
	c.checkGameEnd()
	assert.False(t, c.GetGameEndFlag(), "同点で終局している")

	c.GetPlayer(0).AddScore(1)
	c.checkGameEnd()
	assert.True(t, c.GetGameEndFlag())
	assert.Equal(t, 0, c.GetWinnerIdx())
}

// **点は打っている最中にも入るので、途中で終局しうる。**
func TestCostlyColours_CanWinMidDeal(t *testing.T) {
	c := NewDefaultCostlyColours()
	cfg := DefaultCostlyColoursConfig()
	cfg.TargetScore = CostlyColoursMinTarget
	c.SetConfig(cfg)
	c.Reset()
	require.NoError(t, c.PlayerMog(false))
	c.GetPlayer(0).ResetScore()
	c.GetPlayer(0).AddScore(CostlyColoursMinTarget - 2)
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(ccCard(CardDesignSpade, 7))
	c.GetPlayer(0).AddCard(ccCard(CardDesignHeart, 4))
	c.SetTotalForTest(8)
	c.pile = []*Card{ccCard(CardDesignDiamond, 8)}
	c.SetCurrentForTest(0)

	require.NoError(t, c.PlayerPlay(0)) // 15 で 2 点。
	assert.True(t, c.GetGameEndFlag(), "打っている最中の点で終局していない")
	assert.Equal(t, 0, c.GetWinnerIdx())
}

// **1 ディールを最後まで打てる。**
func TestCostlyColours_PlaysADealThrough(t *testing.T) {
	c := newCostlyGame(t)
	costlyRunDeal(t, c)
	require.Equal(t, CostlyColoursPhaseShow, c.GetPhase())
	res := c.GetLastResult()
	require.NotNil(t, res)
	assert.Len(t, res.Lines, 3)
}

// **試合は終局まで届く。**
func TestCostlyColours_ReachesTheTarget(t *testing.T) {
	c := NewDefaultCostlyColours()
	cfg := DefaultCostlyColoursConfig()
	cfg.TargetScore = CostlyColoursMinTarget
	cfg.CpuDifficulty = CostlyColoursCpuDifficultyEasy
	c.SetConfig(cfg)
	c.Reset()
	for deal := 0; deal < 200 && !c.GetGameEndFlag(); deal++ {
		costlyRunDeal(t, c)
		c.NextDeal()
	}
	require.True(t, c.GetGameEndFlag(), "31 点勝負でも終局に届かない")
	assert.GreaterOrEqual(t, c.GetWinnerIdx(), 0)
}

// **ヒントは CPU の難易度で鈍らない。**
func TestCostlyColours_HintIgnoresCpuDifficulty(t *testing.T) {
	want := -2
	for _, diff := range []CostlyColoursCpuDifficulty{
		CostlyColoursCpuDifficultyEasy, CostlyColoursCpuDifficultyNormal, CostlyColoursCpuDifficultyHard,
	} {
		c := NewDefaultCostlyColours()
		cfg := DefaultCostlyColoursConfig()
		cfg.CpuDifficulty = diff
		c.SetConfig(cfg)
		c.Reset()
		require.NoError(t, c.PlayerMog(false))
		c.GetPlayer(0).Reset()
		c.GetPlayer(0).AddCard(ccCard(CardDesignSpade, 4))
		c.GetPlayer(0).AddCard(ccCard(CardDesignHeart, 7)) // 8 + 7 = 15
		c.SetTotalForTest(8)
		c.pile = []*Card{ccCard(CardDesignDiamond, 8)}
		c.SetCurrentForTest(0)

		for i := 0; i < 20; i++ {
			h := c.GetHint()
			if want == -2 {
				want = h.HandIdx
			}
			assert.Equal(t, want, h.HandIdx, "難易度 %d でヒントが変わった", diff)
			assert.Equal(t, "fifteen", h.Reason)
		}
	}
	assert.Equal(t, 1, want, "15 を作る札を勧めていない")
}

func TestCostlyColours_RejectsBadInput(t *testing.T) {
	c := newCostlyGame(t)
	// 交換フェーズでは札を出せない。
	assert.Error(t, c.PlayerPlay(0))
	require.NoError(t, c.PlayerMog(false))
	// 数え上げでは交換できない。
	assert.Error(t, c.PlayerMog(true))

	assert.Error(t, c.PlayerPlay(-1))
	assert.Error(t, c.PlayerPlay(99))

	c.SetCurrentForTest(1)
	assert.Error(t, c.PlayerPlay(0))
	c.SetCurrentForTest(0)
	c.gameEndFlag = true
	assert.Error(t, c.PlayerPlay(0))
	assert.Error(t, c.PlayerMog(true))
}

// **保存した盤で打ち続けられる。**
func TestCostlyColours_SaveRestoreKeepsPlaying(t *testing.T) {
	c := newCostlyGame(t)
	require.NoError(t, c.PlayerMog(true))

	data, err := json.Marshal(c)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	var r CostlyColours
	require.NoError(t, json.Unmarshal(data, &r))
	assert.Equal(t, c.GetPhase(), r.GetPhase())
	assert.Equal(t, c.GetTotal(), r.GetTotal())
	require.NotNil(t, r.GetTurnUp(), "表の 1 枚が消えている")
	assert.Equal(t, c.GetTurnUp().GetDesign(), r.GetTurnUp().GetDesign())
	assert.Equal(t, c.GetTurnUp().GetValue(), r.GetTurnUp().GetValue())
	for i := 0; i < CostlyColoursPlayerCnt; i++ {
		assert.Equal(t, c.GetPlayer(i).GetCardsSize(), r.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
		assert.Equal(t, c.GetPlayer(i).GetScore(), r.GetPlayer(i).GetScore(), "席 %d の得点", i)
		assert.Equal(t, c.GetPlayer(i).IsMoggedIn(), r.GetPlayer(i).IsMoggedIn())
	}

	costlyRunDeal(t, &r)
	assert.Equal(t, CostlyColoursPhaseShow, r.GetPhase(), "復元した盤でショーに届かない")

	var bad CostlyColours
	assert.Error(t, json.Unmarshal([]byte("{"), &bad))
}

func TestCostlyColoursConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultCostlyColoursConfig().Validate())
	// **既定は Cotton の 61 点。** Parlett の 121 も選べる。
	assert.Equal(t, 61, DefaultCostlyColoursConfig().TargetScore)
	for _, mutate := range []func(*CostlyColoursConfig){
		func(c *CostlyColoursConfig) { c.TargetScore = 1 },
		func(c *CostlyColoursConfig) { c.TargetScore = 999 },
		func(c *CostlyColoursConfig) { c.CpuDifficulty = 9 },
	} {
		cfg := DefaultCostlyColoursConfig()
		mutate(&cfg)
		assert.Error(t, cfg.Validate())
	}
	for _, v := range CostlyColoursTargetOptions {
		cfg := DefaultCostlyColoursConfig()
		cfg.TargetScore = v
		assert.NoError(t, cfg.Validate(), "選べる目標点 %d が弾かれる", v)
	}
	assert.Contains(t, CostlyColoursTargetOptions, 121, "Parlett の 121 が選べない")
}

// costlyRunDeal は 1 ディールをショーまで打つ。
func costlyRunDeal(t *testing.T, c *CostlyColours) {
	t.Helper()
	for step := 0; step < 200; step++ {
		if c.GetGameEndFlag() || c.GetPhase() == CostlyColoursPhaseShow {
			return
		}
		if !c.IsHumanTurn() {
			c.CpuAct()
			continue
		}
		if c.GetPhase() == CostlyColoursPhaseMog {
			require.NoError(t, c.PlayerMog(c.GetHint().AcceptMog))
			continue
		}
		h := c.GetHint()
		require.GreaterOrEqual(t, h.HandIdx, 0, "手番なのに出せる札が無い")
		require.NoError(t, c.PlayerPlay(h.HandIdx))
	}
	require.Contains(t, []string{CostlyColoursPhaseShow, CostlyColoursPhaseGameEnd}, c.GetPhase(),
		"ディールが終わらない")
}

// **「ゴー」の 1 点は 1 回の行き詰まりにつき 1 度だけ。** 相手が詰まったあとも
// 出し続けられるが、1 枚ごとに再び 1 点が入るわけではない。
func TestCostlyColours_GoPegsOncePerStall(t *testing.T) {
	c := newCostlyGame(t)
	require.NoError(t, c.PlayerMog(false))
	// **節目もペアも階段も踏まない札を選ぶ。** どれかに当たると、そちらの
	// 点と「2 度目のゴー」が見分けられなくなる。
	// 累計 25 → 3 を出して 28 → 2 を出して 30。**A を残しておく**ので、
	// 2 手目のあとも席 0 は出せる状態が続く ── 手札が尽きると「ラター」の
	// 1 点が入って、2 度目のゴーと見分けられなくなる。
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(ccCard(CardDesignSpade, 3))
	c.GetPlayer(0).AddCard(ccCard(CardDesignHeart, 2))
	c.GetPlayer(0).AddCard(ccCard(CardDesignClover, 1))
	c.GetPlayer(1).Reset()
	c.GetPlayer(1).AddCard(ccCard(CardDesignDiamond, 13)) // 10 なので最初から出せない
	c.SetTotalForTest(25)
	c.pile = []*Card{ccCard(CardDesignDiamond, 10), ccCard(CardDesignSpade, 10),
		ccCard(CardDesignClover, 5)}
	c.SetCurrentForTest(0)

	before := c.GetPlayer(0).GetScore()
	require.NoError(t, c.applyPlay(0, 0)) // 28
	afterFirst := c.GetPlayer(0).GetScore()
	assert.Equal(t, before+CostlyGoPoints, afterFirst, "最初のゴーで 1 点入っていない")
	require.Equal(t, 28, c.GetTotal())

	require.NoError(t, c.applyPlay(0, 0)) // 30
	require.Equal(t, 30, c.GetTotal(), "節目を踏んでいる — 検査が別の点を見ている")
	assert.Equal(t, afterFirst, c.GetPlayer(0).GetScore(),
		"同じ行き詰まりで 2 度目のゴーが入っている")
}

// **ショーは非親から数える。** 非親の分だけで目標点に届いたら、親の手は
// 数えずにそこで終わる ── 打っている最中と同じで「先に届いたほうが勝ち」。
func TestCostlyColours_ShowCountsElderFirst(t *testing.T) {
	c := NewDefaultCostlyColours()
	cfg := DefaultCostlyColoursConfig()
	cfg.TargetScore = CostlyColoursMinTarget
	c.SetConfig(cfg)
	c.Reset()
	c.SetPhaseForTest(CostlyColoursPhasePlay)
	c.SetTurnUpForTest(ccCard(CardDesignSpade, 12))
	// 親を席 0 にして、非親 (席 1) が先に数える形にする。
	c.dealerIdx = 0
	for i := 0; i < CostlyColoursPlayerCnt; i++ {
		c.GetPlayer(i).ResetDeal()
		c.GetPlayer(i).ResetScore()
		// 両者 4 枚同スート = 6 点。
		for _, v := range []int{3, 5, 9} {
			c.GetPlayer(i).AddPlayed(ccCard(CardDesignSpade, v))
		}
	}
	// どちらもあと 1 点で目標点。**先に数える非親が勝つ。**
	c.GetPlayer(0).AddScore(cfg.TargetScore - 1)
	c.GetPlayer(1).AddScore(cfg.TargetScore - 1)

	c.FinishDealForTest()
	require.True(t, c.GetGameEndFlag(), "ショーで終局していない")
	assert.Equal(t, 1, c.GetWinnerIdx(), "親のほうが先に数えられている")
	// 親の分はまだ足されていない。
	assert.Equal(t, cfg.TargetScore-1, c.GetPlayer(0).GetScore(),
		"終局後も親の手が数えられている")
}
