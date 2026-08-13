//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMonteBankForTest(t *testing.T) *MonteBank {
	t.Helper()
	g := NewDefaultMonteBank()
	g.Reset()
	return g
}

func mbCard(design, value int) *Card { return NewCard(design, value, false) }

// mbStackNext は次に引かれる札を指定する。
//
// **場札もゲートも引く札で決まる。** 固定しないと同じ assert が配りで通ったり
// 落ちたりする。
func mbStackNext(g *MonteBank, cards ...*Card) {
	for i, c := range cards {
		g.deck.deck[g.deck.deckDrawCnt+i] = c
	}
}

// mbStaged は指定の場札を並べた卓を返す。
func mbStaged(t *testing.T, layout ...*Card) *MonteBank {
	t.Helper()
	require.Len(t, layout, MonteBankLayoutSize)
	g := NewDefaultMonteBank()
	g.deck.Replenish()
	g.deck.Shuffle()
	mbStackNext(g, layout...)
	g.Reset()
	// Reset は自分で配り直すので、積み直してから場札だけ差し替える。
	g.layout = append([]*Card(nil), layout...)
	return g
}

// --- 配当の裏取り ---

// **配当は写したものではなく数えた結果。** ここが崩れるとゲームが成立しない。
//
// 場札 4 枚が見えている時点で、ゲートは残り 36 枚の一様な 1 枚。賭けた札の
// スートが場札に何枚出ているかで勝率が決まる。3:1 は「1 枚だけのスートが
// ちょうど互角」になる唯一の倍率で、控除率はすべてプレイヤーの選択から出る。
func TestMonteBank_PayoutIsDerivedNotCopied(t *testing.T) {
	t.Parallel()
	require.Equal(t, 3, MonteBankPayout)

	// 賭けた札のスートが場札に dup 枚あるときの、期待収支 (賭け金 1 単位あたり)。
	// 分数のまま比べる: win = 3 * (10-dup), lose = 1 * (36-(10-dup))
	for _, tt := range []struct {
		dup                     int
		wantWinNum, wantLoseNum int
	}{
		{dup: 1, wantWinNum: 3 * 9, wantLoseNum: 36 - 9}, // 27 対 27 = 互角
		{dup: 2, wantWinNum: 3 * 8, wantLoseNum: 36 - 8}, // 24 対 28
		{dup: 3, wantWinNum: 3 * 7, wantLoseNum: 36 - 7}, // 21 対 29
		{dup: 4, wantWinNum: 3 * 6, wantLoseNum: 36 - 6}, // 18 対 30
	} {
		remaining := MonteBankSuitSize - tt.dup
		gains := MonteBankPayout * remaining
		losses := (MonteBankDeckSize - MonteBankLayoutSize) - remaining
		assert.Equal(t, tt.wantWinNum, gains, "dup=%d の取り分", tt.dup)
		assert.Equal(t, tt.wantLoseNum, losses, "dup=%d の払い", tt.dup)
		if tt.dup == 1 {
			assert.Equal(t, gains, losses, "1 枚だけのスートが互角になっていない")
		} else {
			assert.Less(t, gains, losses, "dup=%d が損になっていない", tt.dup)
		}
	}

	// **4:1 は使えない。** プレイヤー側に優位が付く (issue の上限値)。
	assert.Greater(t, 4*9, 36-9, "4:1 だとプレイヤー有利になることを確認")
}

// **場札を見ずに賭けた場合の総合勝率は 9/39。** 実測で確かめる。
func TestMonteBank_UnconditionalMatchRate(t *testing.T) {
	t.Parallel()
	const rounds = 20000
	hits := 0
	for range rounds {
		g := newMonteBankForTest(t)
		for steps := 0; !g.GetGameEndFlag(); steps++ {
			require.Less(t, steps, 100)
			require.NoError(t, g.PlaceBet(0, MonteBankMinBet)) // 常に先頭に賭ける
			if g.GetResult() == MonteBankResultWin {
				hits++
			}
			require.NoError(t, g.NextRound())
			break // 1 局 1 ラウンドだけ数える (山の減りで偏らせない)
		}
	}
	rate := float64(hits) / float64(rounds)
	// 9/39 = 0.2308。20000 回なら ±0.02 に十分収まる。
	assert.InDelta(t, 9.0/39.0, rate, 0.02, "総合勝率が 9/39 から離れている: %.4f", rate)
}

// --- 進行 ---

func TestMonteBank_Reset(t *testing.T) {
	t.Parallel()
	g := newMonteBankForTest(t)
	assert.Equal(t, MonteBankPhaseBet, g.GetPhase())
	assert.Len(t, g.GetLayout(), MonteBankLayoutSize)
	assert.Nil(t, g.GetGate(), "賭ける前にゲートが開いている")
	assert.Equal(t, -1, g.GetPick())
	assert.Equal(t, MonteBankDefaultChips, g.GetChips())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.NotEmpty(t, g.GetActionLog())
}

// **スートが一致すれば 3:1。** 賭け金の返却を含めて 4 倍が戻る。
func TestMonteBank_WinPaysThreeToOne(t *testing.T) {
	t.Parallel()
	g := mbStaged(t,
		mbCard(CardDesignSpade, 1), mbCard(CardDesignHeart, 2),
		mbCard(CardDesignClover, 3), mbCard(CardDesignDiamond, 4))
	mbStackNext(g, mbCard(CardDesignSpade, 7)) // 場札 0 と同じスート
	before := g.GetChips()

	require.NoError(t, g.PlaceBet(0, 100))
	assert.Equal(t, MonteBankResultWin, g.GetResult())
	// 賭け金 100 の返却 + 300 の配当。
	assert.Equal(t, 400, g.GetPayout())
	assert.Equal(t, before+300, g.GetChips())
}

func TestMonteBank_LoseTakesTheStake(t *testing.T) {
	t.Parallel()
	g := mbStaged(t,
		mbCard(CardDesignSpade, 1), mbCard(CardDesignHeart, 2),
		mbCard(CardDesignClover, 3), mbCard(CardDesignDiamond, 4))
	mbStackNext(g, mbCard(CardDesignHeart, 7)) // 場札 0 とは別のスート
	before := g.GetChips()

	require.NoError(t, g.PlaceBet(0, 100))
	assert.Equal(t, MonteBankResultLose, g.GetResult())
	assert.Zero(t, g.GetPayout())
	assert.Equal(t, before-100, g.GetChips())
}

// **一致の判定はスートだけ。** ランクは見ない。
func TestMonteBank_OnlyTheSuitMatters(t *testing.T) {
	t.Parallel()
	g := mbStaged(t,
		mbCard(CardDesignSpade, 1), mbCard(CardDesignHeart, 2),
		mbCard(CardDesignClover, 3), mbCard(CardDesignDiamond, 4))
	// ランクは場札 0 と同じだがスートが違う → 外れ。
	mbStackNext(g, mbCard(CardDesignClover, 1))
	require.NoError(t, g.PlaceBet(0, 100))
	assert.Equal(t, MonteBankResultLose, g.GetResult(), "ランク一致で当たりになっている")
}

// --- 場札のスート数 ---

func TestMonteBank_SuitCountInLayout(t *testing.T) {
	t.Parallel()
	g := mbStaged(t,
		mbCard(CardDesignSpade, 1), mbCard(CardDesignSpade, 2),
		mbCard(CardDesignHeart, 3), mbCard(CardDesignClover, 4))

	assert.Equal(t, 2, g.SuitCountInLayout(CardDesignSpade))
	assert.Equal(t, 1, g.SuitCountInLayout(CardDesignHeart))
	assert.Zero(t, g.SuitCountInLayout(CardDesignDiamond))
	assert.Equal(t, MonteBankSuitSize-2, g.RemainingOfSuit(CardDesignSpade))
	assert.Equal(t, MonteBankSuitSize, g.RemainingOfSuit(CardDesignDiamond))
}

// --- 助言 ---

// **1 枚しか出ていないスートを薦める。** それが唯一の互角の賭け。
func TestMonteBank_HintPrefersTheLoneSuit(t *testing.T) {
	t.Parallel()
	g := mbStaged(t,
		mbCard(CardDesignSpade, 1), mbCard(CardDesignSpade, 2),
		mbCard(CardDesignSpade, 3), mbCard(CardDesignHeart, 4))

	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, 3, h.PickIdx, "1 枚だけのハートを選んでいない")
	assert.Equal(t, "loneSuit", h.Reason)
}

// **どれも重複しているなら、そう言う。** 「得な手がある」と誤解させないため。
func TestMonteBank_HintSaysWhenNothingIsEven(t *testing.T) {
	t.Parallel()
	g := mbStaged(t,
		mbCard(CardDesignSpade, 1), mbCard(CardDesignSpade, 2),
		mbCard(CardDesignHeart, 3), mbCard(CardDesignHeart, 4))

	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "allDuplicated", h.Reason)
	assert.Equal(t, 2, g.SuitCountInLayout(g.GetLayout()[h.PickIdx].GetDesign()))
}

func TestMonteBank_HintOnlyWhileBetting(t *testing.T) {
	t.Parallel()
	g := newMonteBankForTest(t)
	assert.NotNil(t, g.GetHint())

	require.NoError(t, g.PlaceBet(0, MonteBankDefaultBet))
	assert.Nil(t, g.GetHint(), "決着後に助言が出ている")

	g.gameEndFlag = true
	assert.Nil(t, g.GetHint())
}

// --- 入力の検証 ---

func TestMonteBank_PlaceBetValidation(t *testing.T) {
	t.Parallel()
	g := newMonteBankForTest(t)

	assert.ErrorIs(t, g.PlaceBet(-1, MonteBankDefaultBet), errMonteBankPickRange)
	assert.ErrorIs(t, g.PlaceBet(MonteBankLayoutSize, MonteBankDefaultBet), errMonteBankPickRange)
	assert.ErrorIs(t, g.PlaceBet(0, MonteBankMinBet-MonteBankBetUnit), errMonteBankBetRange)
	assert.ErrorIs(t, g.PlaceBet(0, MonteBankMaxBet+MonteBankBetUnit), errMonteBankBetRange)
	assert.ErrorIs(t, g.PlaceBet(0, 55), errMonteBankBetUnit)

	g.SetChips(20)
	assert.ErrorIs(t, g.PlaceBet(0, 50), errMonteBankNotEnough)
	assert.Equal(t, 20, g.GetChips(), "拒否したのにチップが減っている")
}

func TestMonteBank_PhaseGuards(t *testing.T) {
	t.Parallel()
	g := newMonteBankForTest(t)
	assert.ErrorIs(t, g.NextRound(), errMonteBankWrongPhase)

	require.NoError(t, g.PlaceBet(0, MonteBankDefaultBet))
	assert.ErrorIs(t, g.PlaceBet(1, MonteBankDefaultBet), errMonteBankWrongPhase)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlaceBet(0, MonteBankDefaultBet), errMonteBankFinished)
	assert.ErrorIs(t, g.NextRound(), errMonteBankFinished)
}

// --- 終局 ---

// **山が 1 ラウンドぶんに足りなくなったら終わる。** 場札が 4 枚に満たないまま
// 賭けさせると、選択肢の数が黙って変わる。
func TestMonteBank_EndsWhenTheDeckRunsOut(t *testing.T) {
	t.Parallel()
	g := newMonteBankForTest(t)
	rounds := 0
	for !g.GetGameEndFlag() {
		rounds++
		require.Less(t, rounds, 100, "局が終わらない")
		require.NoError(t, g.PlaceBet(0, MonteBankMinBet))
		require.Len(t, g.GetLayout(), MonteBankLayoutSize, "場札が 4 枚に満たない")
		require.NoError(t, g.NextRound())
	}
	assert.Equal(t, MonteBankPhaseGameEnd, g.GetPhase())
	// 40 枚 / 1 ラウンド 5 枚 = 8 ラウンド。
	assert.Equal(t, MonteBankDeckSize/MonteBankCardsPerRound, rounds)
}

func TestMonteBank_EndsWhenTheChipsRunOut(t *testing.T) {
	t.Parallel()
	g := newMonteBankForTest(t)
	require.NoError(t, g.PlaceBet(0, MonteBankMinBet))
	g.SetChips(MonteBankMinBet - 1)
	require.NoError(t, g.NextRound())
	assert.True(t, g.GetGameEndFlag(), "賭けられないのに局が続いている")
}

func TestMonteBank_Accessors(t *testing.T) {
	t.Parallel()
	g := newMonteBankForTest(t)
	assert.NotNil(t, g.GetPlayer())
	assert.Equal(t, MonteBankDeckSize-MonteBankLayoutSize, g.GetRemainingCards())
	assert.Zero(t, g.GetBet())
	assert.Equal(t, MonteBankResultNone, g.GetResult())
	assert.Zero(t, g.GetPayout())

	g.SetConfig(MonteBankConfig{InitialChips: 500, DefaultBet: 20})
	assert.Equal(t, 500, g.GetConfig().InitialChips)
}

func TestMonteBank_ResultNames(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "win", MonteBankResultName(MonteBankResultWin))
	assert.Equal(t, "lose", MonteBankResultName(MonteBankResultLose))
	assert.Equal(t, "none", MonteBankResultName(MonteBankResultNone))
	assert.Equal(t, "none", MonteBankResultName(MonteBankResult(99)))
}

func TestMonteBank_ConfigValidate(t *testing.T) {
	t.Parallel()
	assert.NoError(t, DefaultMonteBankConfig().Validate())
	assert.ErrorIs(t, MonteBankConfig{InitialChips: 1, DefaultBet: 50}.Validate(), errMonteBankChipsRange)
	assert.ErrorIs(t, MonteBankConfig{InitialChips: 1000, DefaultBet: 5}.Validate(), errMonteBankBetRangeCfg)
	assert.ErrorIs(t, MonteBankConfig{InitialChips: 1000, DefaultBet: 55}.Validate(), errMonteBankBetUnitCfg)
}
