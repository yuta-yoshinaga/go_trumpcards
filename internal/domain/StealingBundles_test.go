//go:build test

package domain

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStealingBundlesForTest(t *testing.T, n int) *StealingBundles {
	t.Helper()
	cfg := DefaultStealingBundlesConfig()
	cfg.PlayerCnt = n
	s := NewStealingBundles(newStealingBundlesSeats(n), cfg)
	s.Reset()
	return s
}

// **場に 4 枚、各自に 4 枚。** 残りが山札。
func TestStealingBundlesResetDeals(t *testing.T) {
	for n := StealingBundlesPlayerCntMin; n <= StealingBundlesPlayerCntMax; n++ {
		s := newStealingBundlesForTest(t, n)
		require.Equal(t, StealingBundlesPhasePlay, s.GetPhase())
		assert.Len(t, s.GetTableCards(), StealingBundlesTableSize, "%d 人", n)
		for i := range n {
			assert.Equal(t, StealingBundlesHandSize, s.GetPlayer(i).GetCardsSize(), "%d 人: 席 %d", n, i)
			assert.Zero(t, s.GetPlayer(i).GetBundleSize(), "%d 人: 席 %d の束は空", n, i)
		}
		assert.Equal(t, StealingBundlesDeckSize-StealingBundlesTableSize-n*StealingBundlesHandSize,
			s.GetDeckRemaining(), "%d 人", n)
		assert.Equal(t, -1, s.GetLastCaptureIdx())
	}
}

// **配りは割り切れます。** 2/3/4 人とも山札をちょうど使い切ります。
func TestStealingBundlesDealDividesEvenly(t *testing.T) {
	for n := StealingBundlesPlayerCntMin; n <= StealingBundlesPlayerCntMax; n++ {
		rest := StealingBundlesDeckSize - StealingBundlesTableSize
		assert.Zero(t, rest%(n*StealingBundlesHandSize),
			"%d 人: 残り %d 枚が %d 枚ずつで割り切れる", n, rest, n*StealingBundlesHandSize)
	}
}

// **場から同じランクを取る。** 出した札が束の一番上になります。
func TestStealingBundlesTakeFromTable(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*Card{
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignDiamond, 3, false),
	})
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 7, false))

	require.NoError(t, s.PlayerTake(0))
	// 7 が 2 枚 + 出した 1 枚。
	assert.Equal(t, 3, s.GetPlayer(0).GetBundleSize())
	assert.Len(t, s.GetTableCards(), 1, "3 は場に残る")
	// **一番上は出した札。** 次に狙われるランクはこれです。
	top := s.GetPlayer(0).GetBundleTop()
	require.NotNil(t, top)
	assert.Equal(t, CardDesignSpade, top.GetDesign())
	assert.Equal(t, 7, top.GetValue())
	assert.Equal(t, 0, s.GetLastCaptureIdx())
}

// **束は一番上のランクだけが弱点。** 中身は関係ありません。
func TestStealingBundlesStealWholeBundle(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest(nil)
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 9, false))
	// 席 1 の束: 中身は 2 と 5 だが、一番上が 9。
	s.GetPlayer(1).SetBundle([]*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignDiamond, 9, false),
	})

	assert.Equal(t, []int{1}, s.GetStealTargets(0, 0))
	require.NoError(t, s.PlayerSteal(0, 1))

	// 3 枚 + 出した 1 枚。
	assert.Equal(t, 4, s.GetPlayer(0).GetBundleSize())
	assert.Zero(t, s.GetPlayer(1).GetBundleSize(), "奪われた側は空になる")
	assert.Equal(t, 0, s.GetLastCaptureIdx())
}

// **一番上でないランクでは奪えません。**
func TestStealingBundlesCannotStealByBuriedRank(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest(nil)
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 2, false))
	s.GetPlayer(1).SetBundle([]*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignDiamond, 9, false),
	})

	assert.Empty(t, s.GetStealTargets(0, 0), "埋まっている 2 は狙えない")
	assert.Error(t, s.PlayerSteal(0, 1))
	assert.Equal(t, 2, s.GetPlayer(1).GetBundleSize(), "束は動かない")
}

// **取れる手があるときは場に置けません。**
func TestStealingBundlesTrailIsBlockedWhileACaptureExists(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*Card{NewCard(CardDesignHeart, 7, false)})
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 7, false), NewCard(CardDesignClover, 4, false))

	assert.True(t, s.CanCapture(0))
	assert.Error(t, s.PlayerTrail(1), "取れる手があるので置けない")

	// **負のコントロール: 取れなくなれば置ける。**
	s.SetTableCardsForTest([]*Card{NewCard(CardDesignHeart, 9, false)})
	s.GiveHandForTest(0, NewCard(CardDesignClover, 4, false))
	assert.False(t, s.CanCapture(0))
	require.NoError(t, s.PlayerTrail(0))
	assert.Len(t, s.GetTableCards(), 2)
}

func TestStealingBundlesRejectsBadInput(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	s.SetCurrentPlayerIdxForTest(0)

	assert.Error(t, s.PlayerTake(-1))
	assert.Error(t, s.PlayerTake(99))
	assert.Error(t, s.PlayerSteal(0, 0), "自分の束は奪えない")
	assert.Error(t, s.PlayerSteal(0, 99))

	// 手番でなければ何もできない。
	s.SetCurrentPlayerIdxForTest(1)
	assert.Error(t, s.PlayerTake(0))
	assert.Error(t, s.PlayerTrail(0))
	assert.Error(t, s.PlayerSteal(0, 1))
}

// **山札が尽きたら終わり。** 場に残った札は最後に取った人のものです。
func TestStealingBundlesLastCaptureTakesTheTable(t *testing.T) {
	s := newStealingBundlesForTest(t, 2)
	s.DrainDeckForTest()
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*Card{
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 3, false),
	})
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 7, false))
	s.GiveHandForTest(1)

	require.NoError(t, s.PlayerTake(0))
	require.True(t, s.GetGameEndFlag())
	// 7 が 2 枚 + 場に残った 3。
	assert.Equal(t, 3, s.GetPlayer(0).GetBundleSize())
	assert.Empty(t, s.GetTableCards())
	assert.Equal(t, 0, s.GetWinnerIdx())
}

// **誰も取らずに終わったら場札は残ります。**
func TestStealingBundlesTableStaysWhenNobodyEverCaptured(t *testing.T) {
	s := newStealingBundlesForTest(t, 2)
	s.DrainDeckForTest()
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*Card{NewCard(CardDesignHeart, 9, false)})
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 4, false))
	s.GiveHandForTest(1)

	require.NoError(t, s.PlayerTrail(0))
	require.True(t, s.GetGameEndFlag())
	assert.Len(t, s.GetTableCards(), 2, "受け取る人が居ないので場に残る")
	assert.Zero(t, s.GetPlayer(0).GetBundleSize())
}

// **全員の手札が尽きたら配り直します。**
func TestStealingBundlesRedealsWhenHandsRunOut(t *testing.T) {
	s := newStealingBundlesForTest(t, 2)
	require.Equal(t, 1, s.GetPacksDealt())
	// **ランクは全員ばらばらにする。** 席 0 が置いた札を席 1 が取れてしまうと、
	// 「取れるときは置けない」に弾かれて配り直しまで進めません。
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 4, false))
	s.GiveHandForTest(1, NewCard(CardDesignClover, 6, false))
	s.SetTableCardsForTest([]*Card{NewCard(CardDesignHeart, 9, false)})
	s.SetCurrentPlayerIdxForTest(0)

	require.NoError(t, s.TrailForTest(0, 0))
	require.NoError(t, s.TrailForTest(1, 0))

	assert.Equal(t, 2, s.GetPacksDealt())
	assert.Equal(t, StealingBundlesHandSize, s.GetPlayer(0).GetCardsSize())
	assert.False(t, s.GetGameEndFlag())
}

func TestStealingBundlesGiveUp(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	s.GiveUp()
	assert.True(t, s.GetGameEndFlag())
	assert.NotEqual(t, 0, s.GetWinnerIdx(), "投了した人は勝たない")

	s.GiveUp()
	assert.True(t, s.GetGameEndFlag())
}

// **CPU はいちばん大きい束を狙います。**
func TestStealingBundlesCpuPrefersTheBiggestBundle(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	s.SetCurrentPlayerIdxForTest(1)
	s.SetTableCardsForTest([]*Card{NewCard(CardDesignHeart, 5, false)})
	s.GiveHandForTest(1, NewCard(CardDesignSpade, 5, false), NewCard(CardDesignClover, 8, false))
	// 席 2 の束は小さく、席 3 の束が大きい。どちらも一番上は 8。
	s.GetPlayer(2).SetBundle([]*Card{NewCard(CardDesignHeart, 8, false)})
	s.GetPlayer(3).SetBundle([]*Card{
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignDiamond, 8, false),
	})

	card, victim := s.CpuChoiceForTest(1)
	assert.Equal(t, 1, card, "8 を出す")
	assert.Equal(t, 3, victim, "大きいほうの束を狙う")
}

func TestStealingBundlesHint(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*Card{NewCard(CardDesignHeart, 5, false)})
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 5, false))

	h := s.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stealingbundlesTake", h.Reason)
	assert.Equal(t, -1, h.VictimIdx)

	// 束を奪えるなら、そちらを勧める。
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 9, false))
	s.GetPlayer(1).SetBundle([]*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignDiamond, 9, false),
	})
	h = s.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stealingbundlesSteal", h.Reason)
	assert.Equal(t, 1, h.VictimIdx)

	s.GiveUp()
	assert.Nil(t, s.GetHint(), "終局後は助言しない")
}

// **全人数を終局まで指して、1 手ごとに保存して読み直す。**
//
// #5316 のレビュー指摘から標準化したテスト。範囲検査では出ない「フィールド間の
// 食い違い」を、書き込み側が実際に作れるかどうかで確かめます。
func TestStealingBundles_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for n := StealingBundlesPlayerCntMin; n <= StealingBundlesPlayerCntMax; n++ {
		for range 10 {
			s := newStealingBundlesForTest(t, n)
			for turns := 0; ; turns++ {
				require.Less(t, turns, 500, "%d 人: 終わらない", n)

				data, err := json.Marshal(s)
				require.NoError(t, err)
				var back StealingBundles
				require.NoError(t, json.Unmarshal(data, &back),
					"%d 人 %d 手目: 書き込み側が codec の不変条件を破った", n, turns)

				if s.GetGameEndFlag() {
					break
				}
				idx := s.GetCurrentPlayerIdx()
				card, victim := s.CpuChoiceForTest(idx)
				switch {
				case victim >= 0:
					require.NoError(t, s.StealForTest(idx, card, victim))
				case len(s.GetTableMatches(idx, card)) > 0:
					require.NoError(t, s.TakeForTest(idx, card))
				default:
					require.NoError(t, s.TrailForTest(idx, card))
				}
			}
		}
	}
}

// **盗みは相手の束を直接減らす。** 場から取るのと同じ印では区別が付かないので、
// 種別と被害者を状態として持つ (#5767)。
func TestStealingBundlesRecordsWhichKindOfCaptureItWas(t *testing.T) {
	s := newStealingBundlesForTest(t, 4)
	assert.Empty(t, s.GetLastCaptureKind(), "まだ誰も取っていない")
	assert.Equal(t, -1, s.GetLastCaptureVictimIdx())

	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*Card{NewCard(CardDesignHeart, 7, false)})
	s.GiveHandForTest(0, NewCard(CardDesignSpade, 7, false))
	require.NoError(t, s.PlayerTake(0))
	assert.Equal(t, StealingBundlesCaptureTake, s.GetLastCaptureKind())
	assert.Equal(t, -1, s.GetLastCaptureVictimIdx(), "取りに被害者はいない")

	// 盗みは種別も被害者も差し替える。
	s.SetCurrentPlayerIdxForTest(2)
	s.SetTableCardsForTest(nil)
	s.GiveHandForTest(2, NewCard(CardDesignClover, 7, false))
	s.GetPlayer(1).SetBundle([]*Card{NewCard(CardDesignDiamond, 7, false)})
	require.NoError(t, s.StealForTest(2, 0, 1))
	assert.Equal(t, StealingBundlesCaptureSteal, s.GetLastCaptureKind())
	assert.Equal(t, 1, s.GetLastCaptureVictimIdx())
	assert.Equal(t, 2, s.GetLastCaptureIdx())
}

// **種別と被害者は一緒に決まる。** 片方だけの保存は印を誤らせる。
func TestStealingBundlesRejectsAnInconsistentCaptureRecord(t *testing.T) {
	// 実際に盗みが起きるまで打つ。手で束を差し替えると 52 枚の不変条件が壊れ、
	// 弾かれた理由が種別なのか枚数なのか分からなくなる。
	s := newStealingBundlesForTest(t, 4)
	for turns := 0; s.GetLastCaptureKind() != StealingBundlesCaptureSteal; turns++ {
		require.Less(t, turns, 500, "盗みが 1 度も起きなかった")
		require.False(t, s.GetGameEndFlag(), "盗みが 1 度も起きずに終わった")
		idx := s.GetCurrentPlayerIdx()
		card, victim := s.CpuChoiceForTest(idx)
		switch {
		case victim >= 0:
			require.NoError(t, s.StealForTest(idx, card, victim))
		case len(s.GetTableMatches(idx, card)) > 0:
			require.NoError(t, s.TakeForTest(idx, card))
		default:
			require.NoError(t, s.TrailForTest(idx, card))
		}
	}
	victim := s.GetLastCaptureVictimIdx()
	require.GreaterOrEqual(t, victim, 0)

	data, err := json.Marshal(s)
	require.NoError(t, err)
	var back StealingBundles
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, StealingBundlesCaptureSteal, back.GetLastCaptureKind())
	assert.Equal(t, victim, back.GetLastCaptureVictimIdx())

	// 盗みなのに被害者が居ない保存は弾く。
	broken := bytes.Replace(data, []byte(`"lv":`+strconv.Itoa(victim)), []byte(`"lv":-1`), 1)
	require.NotEqual(t, string(data), string(broken), "置換が当たっていない")
	assert.Error(t, json.Unmarshal(broken, new(StealingBundles)))

	// **取った席が居ないのに種別だけある**保存も弾く（レビュー指摘）。
	// 束を持っている盤面だと「誰も取っていないのに束がある」の側で落ちてしまい、
	// この検査を通ったことにならない——まだ誰も取っていない局面で試す。
	fresh, err := json.Marshal(newStealingBundlesForTest(t, 4))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(fresh, new(StealingBundles)), "元の保存は通る")
	orphan := bytes.Replace(fresh, []byte(`"lk":""`), []byte(`"lk":"take"`), 1)
	require.NotEqual(t, string(fresh), string(orphan), "置換が当たっていない")
	assert.Error(t, json.Unmarshal(orphan, new(StealingBundles)))

	// 知らない種別も弾く。
	unknown := bytes.Replace(data, []byte(`"lk":"steal"`), []byte(`"lk":"rob"`), 1)
	require.NotEqual(t, string(data), string(unknown), "置換が当たっていない")
	assert.Error(t, json.Unmarshal(unknown, new(StealingBundles)))
}
