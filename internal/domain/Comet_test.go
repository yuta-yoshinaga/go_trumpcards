//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCometGame(t *testing.T) *Comet {
	t.Helper()
	c := NewDefaultComet()
	c.Reset()
	return c
}

// **開幕は人間の手番。** 親の左隣が先に打つ規則なので親を最後の席にしてある ──
// 親を 0 にすると人間は最初の連なりの先頭を選べない。
func TestComet_ResetDealsAndStartsWithTheHuman(t *testing.T) {
	c := newCometGame(t)
	assert.Equal(t, CometPhasePlay, c.GetPhase())
	assert.Equal(t, 1, c.GetRoundNumber())
	assert.Equal(t, 0, c.GetCurrentPlayerIdx(), "人間が先に打てない")
	assert.True(t, c.IsHumanTurn())
	assert.Equal(t, c.GetPlayerCnt()-1, c.GetDealerIdx())
	// **先頭は何でも出せる。**
	assert.Equal(t, 0, c.GetNeed())
	assert.Len(t, c.PlayableIdxs(0), c.GetPlayer(0).GetCardsSize())
}

// **51 枚を配り切り、余りは伏せる。** 均等に配れないぶんを 1 人に足すと
// その席だけ上がりが遠くなる。
func TestComet_DealsEvenlyAndBuriesTheRemainder(t *testing.T) {
	for _, seats := range []int{2, 3, 4, 5} {
		c := NewDefaultComet()
		cfg := DefaultCometConfig()
		cfg.Players = seats
		c.SetConfig(cfg)
		c.Reset()

		require.Equal(t, seats, c.GetPlayerCnt(), "席数 %d", seats)
		per := c.GetPlayer(0).GetCardsSize()
		total := 0
		for i := 0; i < seats; i++ {
			assert.Equal(t, per, c.GetPlayer(i).GetCardsSize(),
				"%d 人卓で席 %d だけ枚数が違う", seats, i)
			total += c.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, CometDeckSize, total+c.GetDeadCount(), "%d 人卓で札が合わない", seats)
		assert.Equal(t, CometDeckSize%seats, c.GetDeadCount(), "%d 人卓の死に手", seats)
	}
}

// **連なりはスートを無視してランクだけで昇る。**
func TestComet_SequenceIgnoresSuit(t *testing.T) {
	c := newCometGame(t)
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cmtCard(CardDesignSpade, 5))
	c.GetPlayer(0).AddCard(cmtCard(CardDesignHeart, 6))
	// **3 枚目を持たせる。** 2 枚で打ち切ると上がってしまい、局が終わって
	// need が更新されないまま検査に入る。
	c.GetPlayer(0).AddCard(cmtCard(CardDesignClover, 2))
	c.SetNeedForTest(0)
	c.SetCurrentForTest(0)

	require.NoError(t, c.applyPlay(0, 0))
	assert.Equal(t, 6, c.GetNeed(), "5 の次に 6 が要ることになっていない")
	// スートの違う 6 が出せる。
	c.SetCurrentForTest(0)
	assert.Equal(t, []int{0}, c.PlayableIdxs(0))
	require.NoError(t, c.applyPlay(0, 0))
	assert.Equal(t, 7, c.GetNeed())
}

// **K とコメットは連なりを切り、出した本人が次を始める。**
func TestComet_KingAndCometStopTheSequence(t *testing.T) {
	for _, tc := range []struct {
		name string
		card *Card
	}{
		{"king", cmtCard(CardDesignSpade, 13)},
		{"comet", cmtCard(CardDesignDiamond, 9)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCometGame(t)
			c.GetPlayer(0).Reset()
			c.GetPlayer(0).AddCard(tc.card)
			c.GetPlayer(0).AddCard(cmtCard(CardDesignHeart, 2))
			c.SetNeedForTest(0)
			c.SetCurrentForTest(0)

			require.NoError(t, c.applyPlay(0, 0))
			assert.Equal(t, 0, c.GetNeed(), "連なりが切れていない")
			assert.Equal(t, 0, c.GetCurrentPlayerIdx(), "出した本人に手番が戻っていない")
		})
	}
}

// **コメットはどのランクの代わりにもなる。**
func TestComet_WildSubstitutes(t *testing.T) {
	c := newCometGame(t)
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cmtCard(CardDesignDiamond, 9))
	c.GetPlayer(0).AddCard(cmtCard(CardDesignSpade, 2))
	c.SetNeedForTest(11) // 誰も J を持っていない想定
	c.SetCurrentForTest(0)
	assert.Equal(t, []int{0}, c.PlayableIdxs(0), "コメットが J の代わりに出せない")
	require.NoError(t, c.applyPlay(0, 0))
	assert.Equal(t, 0, c.GetNeed())
}

// **出せる札があるならパスできない。** 見送れると、コメットを抱えたまま
// 局を止められてしまう。
func TestComet_CannotPassWhenAPlayExists(t *testing.T) {
	c := newCometGame(t)
	c.SetNeedForTest(0)
	c.SetCurrentForTest(0)
	require.NotEmpty(t, c.PlayableIdxs(0))
	assert.Error(t, c.PlayerPass(), "出せるのにパスできてしまう")

	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cmtCard(CardDesignSpade, 2))
	c.SetNeedForTest(5)
	require.Empty(t, c.PlayableIdxs(0))
	assert.NoError(t, c.PlayerPass())
}

// **全員が出せなければストップ。** 最後に出した席が好きな札で再開する。
func TestComet_AllPassRestartsFromTheLastPlayer(t *testing.T) {
	c := newCometGame(t)
	n := c.GetPlayerCnt()
	// 席 1 が最後に出した状態を作り、以降 全員が出せない盤面にする。
	for i := 0; i < n; i++ {
		c.GetPlayer(i).Reset()
		c.GetPlayer(i).AddCard(cmtCard(CardDesignSpade, 2))
		c.GetPlayer(i).AddCard(cmtCard(CardDesignHeart, 3))
	}
	c.lastPlayer = 1
	c.SetNeedForTest(12) // Q は誰も持っていない
	c.SetCurrentForTest(0)
	for i := 0; i < n; i++ {
		c.applyPass(c.GetCurrentPlayerIdx())
	}
	assert.Equal(t, 0, c.GetNeed(), "ストップで連なりが切れていない")
	assert.Equal(t, 1, c.GetCurrentPlayerIdx(), "最後に出した席が再開していない")
}

// **集計は上がり 1 + 相手の残り 1 枚 1 + 出なかった K 1 枚 2、
// コメットを抱えていたら 1 失点。**
func TestComet_ScoreRound(t *testing.T) {
	c := newCometGame(t)
	n := c.GetPlayerCnt()
	for i := 0; i < n; i++ {
		c.GetPlayer(i).Reset()
	}
	// 席 1 は 3 枚残し、うち 1 枚がコメット。他の席は 2 枚ずつ。
	c.GetPlayer(1).AddCard(cmtCard(CardDesignDiamond, 9))
	c.GetPlayer(1).AddCard(cmtCard(CardDesignSpade, 2))
	c.GetPlayer(1).AddCard(cmtCard(CardDesignHeart, 3))
	left := 3
	for i := 2; i < n; i++ {
		c.GetPlayer(i).AddCard(cmtCard(CardDesignClover, 4))
		c.GetPlayer(i).AddCard(cmtCard(CardDesignSpade, 5))
		left += 2
	}
	// K が 1 枚だけ出ている → 出なかった K は 3 枚。
	c.SetPileForTest([]*Card{cmtCard(CardDesignSpade, 13), cmtCard(CardDesignHeart, 2)})

	c.finishRound(0)
	res := c.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, 0, res.WinnerIdx)
	assert.Equal(t, 3, res.UnplayedKings, "出なかった K の数が合わない")
	assert.Equal(t, 1, res.HeldWildIdx, "コメットを抱えた席が拾えていない")

	want := CometGoOutPoints + left + 3*CometUnplayedKingPoints
	assert.Equal(t, want, res.Gained[0])
	assert.Equal(t, want, c.GetPlayer(0).GetScore())
	assert.Equal(t, -CometHoldingWildPenalty, res.Gained[1], "コメットの罰点が引かれていない")
	assert.Equal(t, -CometHoldingWildPenalty, c.GetPlayer(1).GetScore())
	assert.Equal(t, CometPhaseRoundEnd, c.GetPhase())
}

// **上がった瞬間に局が終わる。**
func TestComet_GoingOutEndsTheRound(t *testing.T) {
	c := newCometGame(t)
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cmtCard(CardDesignSpade, 7))
	c.SetNeedForTest(0)
	c.SetCurrentForTest(0)
	require.NoError(t, c.applyPlay(0, 0))
	assert.Equal(t, CometPhaseRoundEnd, c.GetPhase())
	require.NotNil(t, c.GetLastResult())
	assert.Equal(t, 0, c.GetLastResult().WinnerIdx)
}

// **同点では終局しない。**
func TestComet_ATieDoesNotEndTheMatch(t *testing.T) {
	c := newCometGame(t)
	for i := 0; i < c.GetPlayerCnt(); i++ {
		c.GetPlayer(i).ResetScore()
		c.GetPlayer(i).AddScore(c.GetConfig().TargetScore)
	}
	c.checkGameEnd()
	assert.False(t, c.GetGameEndFlag(), "同点で終局している")

	c.GetPlayer(0).AddScore(1)
	c.checkGameEnd()
	assert.True(t, c.GetGameEndFlag())
	assert.Equal(t, 0, c.GetWinnerIdx())
}

// **1 局を最後まで打てる。**
func TestComet_PlaysARoundThrough(t *testing.T) {
	c := newCometGame(t)
	for step := 0; step < 500 && c.GetPhase() == CometPhasePlay; step++ {
		if c.IsHumanTurn() {
			h := c.GetHint()
			if h.HandIdx < 0 {
				require.NoError(t, c.PlayerPass())
				continue
			}
			require.NoError(t, c.PlayerPlay(h.HandIdx))
			continue
		}
		c.CpuPlay()
	}
	require.Equal(t, CometPhaseRoundEnd, c.GetPhase(), "局が終わらない")
	res := c.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, 0, res.CardsLeft[res.WinnerIdx], "上がった席に札が残っている")
}

// **試合は終局まで届く。**
func TestComet_ReachesTheTarget(t *testing.T) {
	c := NewDefaultComet()
	cfg := DefaultCometConfig()
	cfg.TargetScore = CometMinTarget
	cfg.CpuDifficulty = CometCpuDifficultyEasy
	c.SetConfig(cfg)
	c.Reset()
	for round := 0; round < 100 && !c.GetGameEndFlag(); round++ {
		for step := 0; step < 500 && c.GetPhase() == CometPhasePlay; step++ {
			if c.IsHumanTurn() {
				h := c.GetHint()
				if h.HandIdx < 0 {
					require.NoError(t, c.PlayerPass())
					continue
				}
				require.NoError(t, c.PlayerPlay(h.HandIdx))
				continue
			}
			c.CpuPlay()
		}
		require.NotEqual(t, CometPhasePlay, c.GetPhase(), "局が終わらない")
		c.NextRound()
	}
	require.True(t, c.GetGameEndFlag(), "20 点勝負でも終局に届かない")
	assert.GreaterOrEqual(t, c.GetWinnerIdx(), 0)
}

// **ヒントは CPU の難易度で鈍らない。**
func TestComet_HintIgnoresCpuDifficulty(t *testing.T) {
	want := -2
	for _, diff := range []CometCpuDifficulty{
		CometCpuDifficultyEasy, CometCpuDifficultyNormal, CometCpuDifficultyHard,
	} {
		c := NewDefaultComet()
		cfg := DefaultCometConfig()
		cfg.CpuDifficulty = diff
		c.SetConfig(cfg)
		c.Reset()
		c.GetPlayer(0).Reset()
		c.GetPlayer(0).AddCard(cmtCard(CardDesignDiamond, 9)) // コメット
		c.GetPlayer(0).AddCard(cmtCard(CardDesignSpade, 13))  // K
		c.GetPlayer(0).AddCard(cmtCard(CardDesignHeart, 4))
		c.SetNeedForTest(0)
		c.SetCurrentForTest(0)

		for i := 0; i < 20; i++ {
			got := c.GetHint().HandIdx
			if want == -2 {
				want = got
			}
			assert.Equal(t, want, got, "難易度 %d でヒントが変わった", diff)
		}
	}
	// **コメットと K は最後に取っておく。** 平札があるならそちらを勧める。
	assert.Equal(t, 2, want, "コメットや K を先に切るよう勧めている")
}

func TestComet_RejectsBadInput(t *testing.T) {
	c := newCometGame(t)
	assert.Error(t, c.PlayerPlay(-1))
	assert.Error(t, c.PlayerPlay(99))

	c.SetNeedForTest(13)
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cmtCard(CardDesignSpade, 2))
	assert.Error(t, c.PlayerPlay(0), "要らないランクの札が出せてしまう")

	c.phase = CometPhaseRoundEnd
	assert.Error(t, c.PlayerPlay(0))
	assert.Error(t, c.PlayerPass())
	c.phase = CometPhasePlay
	c.gameEndFlag = true
	assert.Error(t, c.PlayerPlay(0))
	c.gameEndFlag = false
	c.SetCurrentForTest(1)
	assert.Error(t, c.PlayerPlay(0))
	assert.Error(t, c.PlayerPass())
}

// **保存した盤で打ち続けられる。**
func TestComet_SaveRestoreKeepsPlaying(t *testing.T) {
	c := newCometGame(t)
	h := c.GetHint()
	require.GreaterOrEqual(t, h.HandIdx, 0)
	require.NoError(t, c.PlayerPlay(h.HandIdx))

	data, err := json.Marshal(c)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	var r Comet
	require.NoError(t, json.Unmarshal(data, &r))
	assert.Equal(t, c.GetPhase(), r.GetPhase())
	assert.Equal(t, c.GetNeed(), r.GetNeed(), "次に要るランクが消えている")
	assert.Equal(t, c.GetDeadCount(), r.GetDeadCount(), "死に手が消えている")
	assert.Equal(t, len(c.GetPile()), len(r.GetPile()), "連なりが消えている")
	assert.Equal(t, c.GetLastPlayerIdx(), r.GetLastPlayerIdx())
	assert.Equal(t, c.GetCurrentPlayerIdx(), r.GetCurrentPlayerIdx())
	for i := 0; i < c.GetPlayerCnt(); i++ {
		assert.Equal(t, c.GetPlayer(i).GetCardsSize(), r.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
		assert.Equal(t, c.GetPlayer(i).GetScore(), r.GetPlayer(i).GetScore(), "席 %d の得点", i)
	}

	for step := 0; step < 500 && r.GetPhase() == CometPhasePlay; step++ {
		if r.IsHumanTurn() {
			hh := r.GetHint()
			if hh.HandIdx < 0 {
				require.NoError(t, r.PlayerPass())
				continue
			}
			require.NoError(t, r.PlayerPlay(hh.HandIdx))
			continue
		}
		r.CpuPlay()
	}
	assert.Equal(t, CometPhaseRoundEnd, r.GetPhase(), "復元した盤で局が終わらない")

	var bad Comet
	assert.Error(t, json.Unmarshal([]byte("{"), &bad))
}

func TestCometConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultCometConfig().Validate())
	for _, mutate := range []func(*CometConfig){
		func(c *CometConfig) { c.TargetScore = 1 },
		func(c *CometConfig) { c.TargetScore = 999 },
		func(c *CometConfig) { c.Players = 1 },
		func(c *CometConfig) { c.Players = 9 },
		func(c *CometConfig) { c.CpuDifficulty = 9 },
	} {
		cfg := DefaultCometConfig()
		mutate(&cfg)
		assert.Error(t, cfg.Validate())
	}
	for _, v := range CometTargetOptions {
		cfg := DefaultCometConfig()
		cfg.TargetScore = v
		assert.NoError(t, cfg.Validate(), "選べる目標点 %d が弾かれる", v)
	}
}
