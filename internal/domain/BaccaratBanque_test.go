//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBanqueGame(t *testing.T) *BaccaratBanque {
	t.Helper()
	b := NewDefaultBaccaratBanque()
	b.Reset()
	return b
}

// bbStack は先頭から順に配られる決め打ちのシューを返す。
// 配りは 右・左・親 を 2 周なので、その順に並べる。
func bbStack(vals ...int) []*Card {
	out := make([]*Card, 0, len(vals)+20)
	for _, v := range vals {
		out = append(out, bbCard(CardDesignSpade, v))
	}
	// 3 枚目以降に使う予備。
	for i := 0; i < 20; i++ {
		out = append(out, bbCard(CardDesignHeart, 10)) // 0 点
	}
	return out
}

// **席は 3 つ。** バンカーと左右 2 つのタブロー。
func TestBaccaratBanque_DealsBankerAndTwoTableaux(t *testing.T) {
	b := newBanqueGame(t)
	assert.Equal(t, 3, b.GetPlayerCnt())
	assert.Equal(t, 1, b.GetCoupNumber())
	for i := 0; i < 3; i++ {
		assert.GreaterOrEqual(t, b.GetPlayer(i).GetCardsSize(), 2, "席 %d に 2 枚配られていない", i)
	}
	// **シューは 3 組 156 枚。**
	assert.Equal(t, BaccaratBanqueDeckSize-b.GetPlayer(0).GetCardsSize()-
		b.GetPlayer(1).GetCardsSize()-b.GetPlayer(2).GetCardsSize(), b.GetShoeRemaining())
	assert.True(t, b.GetPlayer(BaccaratBanqueBankerIdx).GetIsHuman(), "席 0 が人間でない")
}

// **左右は別勘定。** 片方に勝ってもう片方に負けることがある。
func TestBaccaratBanque_SettlesEachSideSeparately(t *testing.T) {
	b := newBanqueGame(t)
	// 右 = 9、左 = 2、親 = 5 になるよう積む (配り順: 右左親 ×2)。
	b.SetShoeForTest(bbStack(4, 1, 2, 5, 1, 3))
	b.SetPhaseForTest(BaccaratBanquePhaseResult)
	b.NextCoup()
	// 親は 5 でナチュラルではないので、判断を求められる。
	require.Equal(t, BaccaratBanquePhaseBanker, b.GetPhase(), "親の判断待ちになっていない")
	require.NoError(t, b.BankerDraw(false))

	require.Equal(t, BaccaratBanquePhaseResult, b.GetPhase(), "決着していない")
	res := b.GetLastResult()
	require.NotNil(t, res)
	require.Len(t, res.Sides, 2)

	outcomes := map[int]string{}
	for _, s := range res.Sides {
		outcomes[s.SeatIdx] = s.Outcome
	}
	assert.NotEqual(t, outcomes[BaccaratBanqueRightIdx], outcomes[BaccaratBanqueLeftIdx],
		"左右が同じ勘定になっている ── 別々に決着していない")
}

// **チップは湧かない。** バンカーの増減と子の増減は打ち消し合う。
func TestBaccaratBanque_ChipsAreConserved(t *testing.T) {
	b := newBanqueGame(t)
	total := func() int {
		n := 0
		for i := 0; i < b.GetPlayerCnt(); i++ {
			n += b.GetPlayer(i).GetChips()
		}
		return n
	}
	want := total()
	for coup := 0; coup < 30 && !b.GetGameEndFlag(); coup++ {
		if b.GetPhase() == BaccaratBanquePhaseBanker {
			require.NoError(t, b.BankerDraw(b.GetHint().Draw))
		}
		assert.Equal(t, want, total(), "クー %d でチップが合わない", coup+1)
		b.NextCoup()
	}
}

// **1 回負けてもバンクは動かない。** #5462 の要件 6 はシュマン・ド・フェールの
// 規則で、バカラ・バンクではない ── バンカーはシューを配り切るか、自分から
// 退くか、資金が尽きるまで席を持つ。
func TestBaccaratBanque_BankSurvivesALoss(t *testing.T) {
	b := newBanqueGame(t)
	// 右 = 9、左 = 9、親 = 1 ── 親は両方に負ける。
	b.SetShoeForTest(bbStack(4, 4, 1, 5, 5, 10))
	b.SetPhaseForTest(BaccaratBanquePhaseResult)
	b.NextCoup()
	if b.GetPhase() == BaccaratBanquePhaseBanker {
		require.NoError(t, b.BankerDraw(false))
	}

	res := b.GetLastResult()
	require.NotNil(t, res)
	assert.Negative(t, res.BankerDelta, "親が両方に負けていない")
	assert.False(t, b.GetGameEndFlag(), "1 回負けただけでバンクが終わっている")
	assert.False(t, b.IsRetired())

	held := b.GetBankHeld()
	b.NextCoup()
	assert.Equal(t, held+1, b.GetBankHeld(), "同じバンクで続いていない")
	assert.False(t, b.GetGameEndFlag())
}

// **資金が尽きたらバンクは終わる。**
func TestBaccaratBanque_BankEndsWhenBankrupt(t *testing.T) {
	b := newBanqueGame(t)
	b.GetPlayer(BaccaratBanqueBankerIdx).SetChips(1)
	// 右 = 9、左 = 9、親 = 1。
	b.SetShoeForTest(bbStack(4, 4, 1, 5, 5, 10))
	b.SetPhaseForTest(BaccaratBanquePhaseResult)
	b.NextCoup()
	if b.GetPhase() == BaccaratBanquePhaseBanker {
		require.NoError(t, b.BankerDraw(false))
	}
	assert.True(t, b.GetGameEndFlag(), "資金が尽きてもバンクが続いている")
	assert.LessOrEqual(t, b.GetPlayer(BaccaratBanqueBankerIdx).GetChips(), 0)
}

// **自分から降りられる。**
func TestBaccaratBanque_BankerMayRetire(t *testing.T) {
	b := newBanqueGame(t)
	if b.GetPhase() == BaccaratBanquePhaseBanker {
		require.NoError(t, b.BankerDraw(false))
	}
	require.Equal(t, BaccaratBanquePhaseResult, b.GetPhase())
	require.NoError(t, b.Retire())
	assert.True(t, b.IsRetired())
	assert.True(t, b.GetGameEndFlag())
	// 決着の外では降りられない。
	assert.Error(t, b.Retire())
}

// **ナチュラルなら 3 枚目は無い。**
func TestBaccaratBanque_NaturalSkipsTheDraw(t *testing.T) {
	b := newBanqueGame(t)
	// 親 = 9 (4+5)、右 = 1、左 = 2。
	b.SetShoeForTest(bbStack(1, 2, 4, 10, 10, 5))
	b.SetPhaseForTest(BaccaratBanquePhaseResult)
	b.NextCoup()

	banker := b.GetPlayer(BaccaratBanqueBankerIdx)
	require.Equal(t, 9, banker.GetTotal(), "親がナチュラルになっていない")
	// ナチュラルなので親の判断を待たずに決着している。
	assert.Equal(t, BaccaratBanquePhaseResult, b.GetPhase(), "ナチュラルで決着していない")
	assert.False(t, banker.HasDrawn(), "ナチュラルなのに引いている")
	require.NotNil(t, b.GetLastResult())
	assert.True(t, b.GetLastResult().BankerNatural)
}

// **ナチュラルの親は引けない。**
func TestBaccaratBanque_NaturalBankerCannotDraw(t *testing.T) {
	b := newBanqueGame(t)
	b.GetPlayer(BaccaratBanqueBankerIdx).Reset()
	b.GetPlayer(BaccaratBanqueBankerIdx).AddCard(bbCard(CardDesignSpade, 4))
	b.GetPlayer(BaccaratBanqueBankerIdx).AddCard(bbCard(CardDesignHeart, 5))
	b.SetPhaseForTest(BaccaratBanquePhaseBanker)
	assert.Error(t, b.BankerDraw(true), "ナチュラルの親が引けてしまっている")
}

// **子の 0-4 は必ず引き、6-7 は必ず止まる。** 難易度に関わらず。
func TestBaccaratBanque_PunterDrawIsNotDiscretionaryOutsideFive(t *testing.T) {
	for _, diff := range []BaccaratBanqueCpuDifficulty{
		BaccaratBanqueCpuDifficultyEasy, BaccaratBanqueCpuDifficultyNormal,
		BaccaratBanqueCpuDifficultyHard,
	} {
		b := NewDefaultBaccaratBanque()
		cfg := DefaultBaccaratBanqueConfig()
		cfg.CpuDifficulty = diff
		b.SetConfig(cfg)
		b.Reset()
		// 右 = 3 (必ず引く)、左 = 7 (必ず止まる)、親 = 6。
		b.SetShoeForTest(bbStack(1, 3, 3, 2, 4, 3))
		b.SetPhaseForTest(BaccaratBanquePhaseResult)
		b.NextCoup()

		assert.True(t, b.GetPlayer(BaccaratBanqueRightIdx).HasDrawn(),
			"難易度 %d: 合計 3 の子が引いていない", diff)
		assert.False(t, b.GetPlayer(BaccaratBanqueLeftIdx).HasDrawn(),
			"難易度 %d: 合計 7 の子が引いている", diff)
	}
}

// **ヒントは難易度で鈍らない。**
func TestBaccaratBanque_HintIgnoresCpuDifficulty(t *testing.T) {
	want := ""
	for _, diff := range []BaccaratBanqueCpuDifficulty{
		BaccaratBanqueCpuDifficultyEasy, BaccaratBanqueCpuDifficultyNormal,
		BaccaratBanqueCpuDifficultyHard,
	} {
		b := NewDefaultBaccaratBanque()
		cfg := DefaultBaccaratBanqueConfig()
		cfg.CpuDifficulty = diff
		b.SetConfig(cfg)
		b.Reset()
		b.GetPlayer(BaccaratBanqueBankerIdx).Reset()
		b.GetPlayer(BaccaratBanqueBankerIdx).AddCard(bbCard(CardDesignSpade, 1))
		b.GetPlayer(BaccaratBanqueBankerIdx).AddCard(bbCard(CardDesignHeart, 2))
		for _, idx := range []int{BaccaratBanqueRightIdx, BaccaratBanqueLeftIdx} {
			b.GetPlayer(idx).Reset()
			b.GetPlayer(idx).AddCard(bbCard(CardDesignClover, 3))
			b.GetPlayer(idx).AddCard(bbCard(CardDesignDiamond, 3))
		}
		b.SetPhaseForTest(BaccaratBanquePhaseBanker)

		for i := 0; i < 20; i++ {
			h := b.GetHint()
			got := h.Reason
			if want == "" {
				want = got
			}
			assert.Equal(t, want, got, "難易度 %d でヒントが変わった", diff)
			assert.True(t, h.Draw, "合計 3 なのに引けと言っていない")
		}
	}
	assert.Equal(t, "behind_both", want, "両方に負けている状況を見ていない")
}

func TestBaccaratBanque_RejectsBadInput(t *testing.T) {
	b := newBanqueGame(t)
	if b.GetPhase() == BaccaratBanquePhaseBanker {
		require.NoError(t, b.BankerDraw(false))
	}
	// 決着中は引けない。
	assert.Error(t, b.BankerDraw(true))
	b.gameEndFlag = true
	assert.Error(t, b.BankerDraw(false))
	assert.Error(t, b.Retire())
}

// **保存した盤で続けられる。**
func TestBaccaratBanque_SaveRestoreKeepsPlaying(t *testing.T) {
	b := newBanqueGame(t)
	if b.GetPhase() == BaccaratBanquePhaseBanker {
		require.NoError(t, b.BankerDraw(b.GetHint().Draw))
	}

	data, err := json.Marshal(b)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	var r BaccaratBanque
	require.NoError(t, json.Unmarshal(data, &r))
	assert.Equal(t, b.GetPhase(), r.GetPhase())
	assert.Equal(t, b.GetCoupNumber(), r.GetCoupNumber())
	assert.Equal(t, b.GetBankHeld(), r.GetBankHeld(), "バンクの継続数が消えている")
	assert.Equal(t, b.GetShoeRemaining(), r.GetShoeRemaining(), "シューの位置が消えている")
	for i := 0; i < b.GetPlayerCnt(); i++ {
		assert.Equal(t, b.GetPlayer(i).GetChips(), r.GetPlayer(i).GetChips(), "席 %d のチップ", i)
	}

	for coup := 0; coup < 30 && !r.GetGameEndFlag(); coup++ {
		r.NextCoup()
		if r.GetPhase() == BaccaratBanquePhaseBanker {
			require.NoError(t, r.BankerDraw(r.GetHint().Draw))
		}
	}
	assert.NotEqual(t, 0, r.GetCoupNumber(), "復元した盤で 1 クーも進まない")

	var bad BaccaratBanque
	assert.Error(t, json.Unmarshal([]byte("{"), &bad))
}

func TestBaccaratBanqueConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultBaccaratBanqueConfig().Validate())
	for _, mutate := range []func(*BaccaratBanqueConfig){
		func(c *BaccaratBanqueConfig) { c.StartChips = 1 },
		func(c *BaccaratBanqueConfig) { c.StartChips = 999999 },
		func(c *BaccaratBanqueConfig) { c.BetAmount = 1 },
		func(c *BaccaratBanqueConfig) { c.BetAmount = 99999 },
		func(c *BaccaratBanqueConfig) { c.CpuDifficulty = 9 },
	} {
		cfg := DefaultBaccaratBanqueConfig()
		mutate(&cfg)
		assert.Error(t, cfg.Validate())
	}
	for _, v := range BaccaratBanqueChipsOptions {
		cfg := DefaultBaccaratBanqueConfig()
		cfg.StartChips = v
		assert.NoError(t, cfg.Validate(), "選べる元手 %d が弾かれる", v)
	}
	for _, v := range BaccaratBanqueBetOptions {
		cfg := DefaultBaccaratBanqueConfig()
		cfg.BetAmount = v
		assert.NoError(t, cfg.Validate(), "選べる張り額 %d が弾かれる", v)
	}
}
