//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAndarBaharForTest(t *testing.T) *AndarBahar {
	t.Helper()
	ab := NewDefaultAndarBahar()
	require.Equal(t, AndarBaharPhaseBet, ab.GetPhase())
	require.NotNil(t, ab.GetJoker(), "ベット前に基準札が公開されている")
	return ab
}

// **先に配る列は基準札の色で決まる。** 黒ならアンダー、赤ならバハール。
func TestAndarBaharFirstColumnFollowsTheJokerColour(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		design int
		want   int
		name   string
	}{
		{CardDesignSpade, AndarBaharBetAndar, "スペードは黒"},
		{CardDesignClover, AndarBaharBetAndar, "クラブは黒"},
		{CardDesignHeart, AndarBaharBetBahar, "ハートは赤"},
		{CardDesignDiamond, AndarBaharBetBahar, "ダイヤは赤"},
	} {
		assert.Equal(t, tt.want, AndarBaharFirstColumnFor(NewCard(tt.design, 7, false)), tt.name)
	}
	// 札が無ければ既定でアンダー。
	assert.Equal(t, AndarBaharBetAndar, AndarBaharFirstColumnFor(nil))
}

// **交互配布の不変条件を、実際に回した全ラウンドで検査する。**
//
// 勝った列は「配った枚数の偶奇」で決まります——奇数枚で終われば先に配る列、偶数枚なら
// 後の列。ここが崩れるのは配る順を取り違えたときだけなので、統計に頼らず確実に落とせます。
func TestAndarBaharDealAlternatesAndStopsOnTheFirstMatch(t *testing.T) {
	ab := newAndarBaharForTest(t)

	for round := range 400 {
		ab.SetChips(AndarBaharDefaultChips)
		require.NoError(t, ab.Bet(AndarBaharMinBet, AndarBaharBetAndar, 0, AndarBaharSideNone))

		total := ab.DealtCount()
		require.Positive(t, total, "ラウンド %d: 1 枚も配られていない", round)
		require.LessOrEqual(t, total, AndarBaharMaxCards, "ラウンド %d: 配りすぎ", round)

		first, second := ab.GetAndarCards(), ab.GetBaharCards()
		if ab.GetFirstColumn() == AndarBaharBetBahar {
			first, second = second, first
		}
		// **先の列は後の列と同数か 1 枚多い。**
		assert.Equal(t, (total+1)/2, len(first), "ラウンド %d: 先の列の枚数", round)
		assert.Equal(t, total/2, len(second), "ラウンド %d: 後の列の枚数", round)

		// **枚数の偶奇が勝った列を決める。**
		wantWinner := ab.GetFirstColumn()
		if total%2 == 0 {
			wantWinner = andarBaharOtherColumn(ab.GetFirstColumn())
		}
		assert.Equal(t, wantWinner, ab.GetWinner(), "ラウンド %d: %d 枚で決着", round, total)

		// **同ランクは決着の 1 枚だけ。**
		rank := andarBaharRank(ab.GetJoker())
		matches := 0
		for _, col := range [][]*Card{ab.GetAndarCards(), ab.GetBaharCards()} {
			for _, c := range col {
				if andarBaharRank(c) == rank {
					matches++
				}
			}
		}
		assert.Equal(t, 1, matches, "ラウンド %d: 途中で止まり損ねている", round)

		ab.Reset()
	}
}

// **チップは賭けた額だけ減り、払戻額だけ増える。**
func TestAndarBaharChipsAreConserved(t *testing.T) {
	ab := newAndarBaharForTest(t)

	for round := range 200 {
		ab.SetChips(AndarBaharDefaultChips)
		before := ab.GetChips()
		stake, side := 100, 50
		require.NoError(t, ab.Bet(stake, AndarBaharBetAndar, side, AndarBaharSide2To5))
		assert.Equal(t, before-stake-side+ab.GetPayout(), ab.GetChips(),
			"ラウンド %d: チップの出入りが合わない", round)
		ab.Reset()
	}
}

// **先に配る列は 0.9:1、後の列は 1:1。** 同じ配当にするとプレイヤー有利になります。
//
// 乱数に依存する分岐なので、当たり／外れの両方を踏むまで回します。
func TestAndarBaharPaysTheFirstColumnLess(t *testing.T) {
	ab := newAndarBaharForTest(t)

	sawFirstWin, sawSecondWin, sawLoss := false, false, false
	const stake = 100
	for range 500 {
		ab.SetChips(AndarBaharDefaultChips)
		target := ab.GetFirstColumn()
		if sawFirstWin && !sawSecondWin {
			target = andarBaharOtherColumn(ab.GetFirstColumn())
		}
		require.NoError(t, ab.Bet(stake, target, 0, AndarBaharSideNone))

		switch {
		case ab.GetResult() != GameResultWin:
			assert.Zero(t, ab.GetPayout(), "外れたら払い戻しは無い")
			sawLoss = true
		case target == ab.GetFirstColumn():
			// 0.9:1 → 賭け金 100 が 190 になって返る。
			assert.Equal(t, stake*AndarBaharFirstColumnPayout/AndarBaharPayoutScale, ab.GetPayout())
			assert.Equal(t, 190, ab.GetPayout(), "0.9:1 は端数を出さない")
			sawFirstWin = true
		default:
			assert.Equal(t, stake*2, ab.GetPayout(), "後の列は 1:1")
			sawSecondWin = true
		}
		ab.Reset()
		if sawFirstWin && sawSecondWin && sawLoss {
			break
		}
	}
	assert.True(t, sawFirstWin, "先の列で当たる回を踏めなかった")
	assert.True(t, sawSecondWin, "後の列で当たる回を踏めなかった")
	assert.True(t, sawLoss, "外れる回を踏めなかった")
}

// **10 の倍数の賭け金なら 0.9:1 でも端数が出ない。**
func TestAndarBaharFirstColumnPayoutIsExact(t *testing.T) {
	t.Parallel()

	for _, stake := range []int{10, 20, 50, 130, 1000, AndarBaharMaxBet} {
		got := stake * AndarBaharFirstColumnPayout / AndarBaharPayoutScale
		assert.Equal(t, stake*19, got*10, "賭け金 %d で丸めが出た", stake)
		assert.Equal(t, stake+stake*9/10, got, "賭け金 %d は 0.9:1", stake)
	}
}

// **サイドベットは決着枚数の帯が一致したときだけ払う。**
func TestAndarBaharSideBetPaysOnlyItsBand(t *testing.T) {
	ab := newAndarBaharForTest(t)

	sawHit, sawMiss := false, false
	const side = 100
	for range 500 {
		ab.SetChips(AndarBaharDefaultChips)
		// メインベットは外れても構わないので、サイドベットだけを見る。
		require.NoError(t, ab.Bet(AndarBaharMinBet, AndarBaharBetAndar, side, AndarBaharSide6To10))

		wantMain := 0
		if ab.GetResult() == GameResultWin {
			rate := AndarBaharSecondColumnPayout
			if ab.GetBetTarget() == ab.GetFirstColumn() {
				rate = AndarBaharFirstColumnPayout
			}
			wantMain = AndarBaharMinBet * rate / AndarBaharPayoutScale
		}
		// **内訳は引き算で作らない** (#5770)。ドメインが持つ値そのものを見る。
		assert.Equal(t, wantMain, ab.GetMainPayout())
		sidePayout := ab.GetSidePayout()
		assert.Equal(t, ab.GetPayout(), ab.GetMainPayout()+ab.GetSidePayout(), "内訳の和が合計")

		n := ab.DealtCount()
		if n >= 6 && n <= 10 {
			want, ok := AndarBaharSidePayout(AndarBaharSide6To10)
			require.True(t, ok)
			assert.Equal(t, side*want/AndarBaharPayoutScale, sidePayout, "%d 枚は 6-10 の帯", n)
			sawHit = true
		} else {
			assert.Zero(t, sidePayout, "%d 枚は 6-10 の帯ではない", n)
			sawMiss = true
		}
		ab.Reset()
		if sawHit && sawMiss {
			break
		}
	}
	assert.True(t, sawHit, "帯に当たる回を踏めなかった")
	assert.True(t, sawMiss, "帯を外す回を踏めなかった")
}

// **帯は 1..51 を隙間なく覆う。**
func TestAndarBaharSideBandsCoverEveryCount(t *testing.T) {
	t.Parallel()

	ab := newAndarBaharForTest(t)
	for n := 1; n <= AndarBaharMaxCards; n++ {
		band := ab.SideBandOf(n)
		require.NotEqual(t, AndarBaharSideNone, band, "%d 枚が帯に入らない", n)
		lo, hi, ok := AndarBaharSideBand(band)
		require.True(t, ok)
		assert.True(t, n >= lo && n <= hi, "%d 枚が帯 %d (%d-%d) に入らない", n, band, lo, hi)
	}
	assert.Equal(t, AndarBaharSideNone, ab.SideBandOf(0))

	// 範囲外の帯は引けない。
	_, _, ok := AndarBaharSideBand(-1)
	assert.False(t, ok)
	_, _, ok = AndarBaharSideBand(AndarBaharSide36Plus + 1)
	assert.False(t, ok)
	_, ok = AndarBaharSidePayout(-1)
	assert.False(t, ok)
	_, ok = AndarBaharSidePayout(AndarBaharSide36Plus + 1)
	assert.False(t, ok)
}

func TestAndarBaharBetRejectsBadInput(t *testing.T) {
	ab := newAndarBaharForTest(t)

	assert.Error(t, ab.Bet(AndarBaharMinBet, 9, 0, AndarBaharSideNone), "ベット先が範囲外")
	assert.Error(t, ab.Bet(AndarBaharMinBet-1, AndarBaharBetAndar, 0, AndarBaharSideNone), "最低額未満")
	assert.Error(t, ab.Bet(AndarBaharMaxBet+10, AndarBaharBetAndar, 0, AndarBaharSideNone), "上限超え")
	assert.Error(t, ab.Bet(15, AndarBaharBetAndar, 0, AndarBaharSideNone), "10 の倍数でない")
	assert.Error(t, ab.Bet(AndarBaharMinBet, AndarBaharBetAndar, 0, 99), "サイドベットの帯が範囲外")
	assert.Error(t, ab.Bet(AndarBaharMinBet, AndarBaharBetAndar, 15, AndarBaharSide2To5),
		"サイドベット額が 10 の倍数でない")

	ab.SetChips(10)
	assert.Error(t, ab.Bet(100, AndarBaharBetAndar, 0, AndarBaharSideNone), "チップ不足")
	// **メインとサイドの合計で足りるかを見る。**
	ab.SetChips(100)
	assert.Error(t, ab.Bet(100, AndarBaharBetAndar, 50, AndarBaharSide2To5), "合計でチップ不足")
	assert.Equal(t, 100, ab.GetChips(), "弾いたベットでチップは減らない")

	ab.SetChips(AndarBaharDefaultChips)
	require.NoError(t, ab.Bet(AndarBaharMinBet, AndarBaharBetAndar, 0, AndarBaharSideNone))
	assert.Error(t, ab.Bet(AndarBaharMinBet, AndarBaharBetAndar, 0, AndarBaharSideNone),
		"決着後は賭けられない")
}

// **サイドベットなしなら金額は 0 に落とす。**
func TestAndarBaharSideNoneIgnoresTheAmount(t *testing.T) {
	ab := newAndarBaharForTest(t)
	before := ab.GetChips()
	require.NoError(t, ab.Bet(100, AndarBaharBetAndar, 500, AndarBaharSideNone))
	assert.Zero(t, ab.GetSideAmount())
	assert.Equal(t, AndarBaharSideNone, ab.GetSideBand())
	assert.Equal(t, before-100+ab.GetPayout(), ab.GetChips(), "サイド分は引かれない")
}

// **チップが尽きたら再スタートさせる。**
func TestAndarBaharResetRefillsWhenBroke(t *testing.T) {
	ab := newAndarBaharForTest(t)
	ab.SetChips(5)
	ab.Reset()
	assert.Equal(t, AndarBaharDefaultChips, ab.GetChips())

	ab.SetChips(500)
	ab.Reset()
	assert.Equal(t, 500, ab.GetChips(), "足りているぶんには足さない")
}

// **罫線は勝った列を積む。**
func TestAndarBaharHistory(t *testing.T) {
	ab := newAndarBaharForTest(t)
	for range 3 {
		require.NoError(t, ab.Bet(AndarBaharMinBet, AndarBaharBetAndar, 0, AndarBaharSideNone))
		ab.Reset()
	}
	h := ab.GetHistory()
	assert.Len(t, h, 3, "Reset では罫線を消さない")
	for _, v := range h {
		assert.Contains(t, []int{AndarBaharBetAndar, AndarBaharBetBahar}, v)
	}

	ab.ClearHistory()
	assert.Empty(t, ab.GetHistory())

	ab.SetHistory([]int{AndarBaharBetBahar})
	assert.Equal(t, []int{AndarBaharBetBahar}, ab.GetHistory())
}

func TestAndarBaharHint(t *testing.T) {
	ab := newAndarBaharForTest(t)

	want := "andarBaharHintAndarFirst"
	if ab.GetFirstColumn() == AndarBaharBetBahar {
		want = "andarBaharHintBaharFirst"
	}
	assert.Equal(t, want, ab.GetHint())

	require.NoError(t, ab.Bet(AndarBaharMinBet, AndarBaharBetAndar, 0, AndarBaharSideNone))
	assert.Equal(t, "andarBaharHintWaitNextRound", ab.GetHint(), "決着後は次ラウンドを促す")
}

func TestAndarBaharGetters(t *testing.T) {
	ab := newAndarBaharForTest(t)
	assert.Equal(t, -1, ab.GetWinner(), "決着前は勝者なし")
	assert.Zero(t, ab.DealtCount())
	assert.Zero(t, ab.GetBetAmount())
	assert.Zero(t, ab.GetPayout())
	assert.False(t, ab.GetGameEndFlag())

	require.NoError(t, ab.Bet(20, AndarBaharBetBahar, 10, AndarBaharSideFirst))
	assert.Equal(t, 20, ab.GetBetAmount())
	assert.Equal(t, AndarBaharBetBahar, ab.GetBetTarget())
	assert.Equal(t, 10, ab.GetSideAmount())
	assert.Equal(t, AndarBaharSideFirst, ab.GetSideBand())
	assert.True(t, ab.GetGameEndFlag())
	assert.Equal(t, AndarBaharPhaseEnd, ab.GetPhase())

	ab.SetPhase(AndarBaharPhaseBet)
	assert.Equal(t, AndarBaharPhaseBet, ab.GetPhase())

	assert.Equal(t, "andar", andarBaharColumnName(AndarBaharBetAndar))
	assert.Equal(t, "bahar", andarBaharColumnName(AndarBaharBetBahar))
	assert.Equal(t, "unknown", andarBaharColumnName(99))
	assert.Zero(t, andarBaharRank(nil))
}
