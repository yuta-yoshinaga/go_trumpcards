//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedSnapClock は決まった時刻を返す時計。**反射ゲームの決定性はここ次第。**
func fixedSnapClock(ms int64) func() time.Time {
	return func() time.Time { return time.UnixMilli(ms) }
}

func newTestSnap(t *testing.T) *Snap {
	t.Helper()
	g := NewDefaultSnap()
	g.SetClock(fixedSnapClock(0))
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

// **52 枚を配り切る。** 3 人では割り切れないので、1 枚多い席が出る。
func TestSnap_DealsEveryCardForEveryPlayerCount(t *testing.T) {
	for n := SnapPlayerCntMin; n <= SnapPlayerCntMax; n++ {
		g := NewSnap(nil, SnapConfig{PlayerCnt: n, CpuDifficulty: SnapCpuNormal})
		g.SetClock(fixedSnapClock(0))
		g.Reset()

		total, sizes := 0, make([]int, n)
		for i := range n {
			sizes[i] = g.GetPlayer(i).GetStockSize()
			total += sizes[i]
		}
		assert.Equal(t, SnapDeckSize, total, "%d 人でも 52 枚すべてが配られる", n)

		// 端数は 1 枚差までに収まる。
		lo, hi := sizes[0], sizes[0]
		for _, s := range sizes {
			lo, hi = min(lo, s), max(hi, s)
		}
		assert.LessOrEqual(t, hi-lo, 1, "%d 人: 差は 1 枚まで（%v）", n, sizes)
		if SnapDeckSize%n == 0 {
			assert.Equal(t, lo, hi, "%d 人なら均等", n)
		}
	}
}

// **トリガーは固定ではない。** 「直前の札と同ランク」なので 1 枚では成立しない。
func TestSnap_SnapNeedsTwoMatchingCards(t *testing.T) {
	g := newTestSnap(t)

	g.SetCenterPileForTest(nil)
	assert.False(t, g.IsSnapAvailable(), "場が空なら成立しない")

	g.SetCenterPileForTest([]*Card{NewCard(CardDesignSpade, 7, false)})
	assert.False(t, g.IsSnapAvailable(), "**1 枚では直前の札が無い**")

	g.SetCenterPileForTest([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 9, false),
	})
	assert.False(t, g.IsSnapAvailable(), "ランクが違う")

	g.SetCenterPileForTest([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
	})
	assert.True(t, g.IsSnapAvailable(), "スートが違ってもランクが同じなら成立")

	// **見るのは上 2 枚だけ。** 下に同ランクがあっても関係ない。
	g.SetCenterPileForTest([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 2, false),
	})
	assert.False(t, g.IsSnapAvailable(), "上 2 枚が違えば、下に揃っていても成立しない")
}

// **正しい宣言は場札を総取りし、取った人が次にめくる。**
func TestSnap_CorrectSnapTakesThePile(t *testing.T) {
	g := newTestSnap(t)
	pile := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
	}
	g.SetCenterPileForTest(pile)
	g.SetCurrentTurnIdxForTest(1)
	before := g.GetPlayer(0).GetStockSize()

	require.NoError(t, g.PlayerSnap())
	assert.Empty(t, g.GetCenterPile())
	assert.Equal(t, before+len(pile), g.GetPlayer(0).GetStockSize())
	assert.Equal(t, SnapEventSnapCorrect, g.GetLastEvent().Kind)
	assert.Equal(t, 0, g.GetCurrentTurnIdx(), "取った人が次にめくる")
}

// **誤宣言はストックから 1 枚を場に差し出す。**
func TestSnap_WrongSnapPaysACard(t *testing.T) {
	g := newTestSnap(t)
	g.SetCenterPileForTest([]*Card{NewCard(CardDesignSpade, 7, false)})
	before := g.GetPlayer(0).GetStockSize()

	require.NoError(t, g.PlayerSnap())
	assert.Equal(t, before-1, g.GetPlayer(0).GetStockSize())
	assert.Len(t, g.GetCenterPile(), 2, "差し出した札が場に乗る")
	assert.Equal(t, SnapEventSnapWrong, g.GetLastEvent().Kind)

	// **差し出す札が無くても落ちない。**
	g.GiveStockForTest(0)
	g.SetCenterPileForTest([]*Card{NewCard(CardDesignSpade, 7, false)})
	assert.NotPanics(t, func() { _ = g.PlayerSnap() })
}

// **札は増えも減りもしない。** 誤宣言も総取りも、52 枚のどこかにある。
func TestSnap_CardsAreConserved(t *testing.T) {
	g := newTestSnap(t)
	count := func() int {
		total := g.GetCenterPileSize()
		for i := range g.GetPlayerCnt() {
			total += g.GetPlayer(i).GetStockSize()
		}
		return total
	}
	require.Equal(t, SnapDeckSize, count())

	for range 40 {
		if g.GetGameEndFlag() {
			break
		}
		if g.IsSnapAvailable() {
			require.NoError(t, g.PlayerSnap())
		} else if g.GetCurrentTurnIdx() == 0 {
			require.NoError(t, g.PlayerStep())
		} else {
			g.StepForTest(g.GetCurrentTurnIdx())
		}
		require.Equal(t, SnapDeckSize, count(), "52 枚が保たれる")
	}
}

// **ストックが尽きた席は飛ばす。**
func TestSnap_EmptyStockIsSkipped(t *testing.T) {
	g := newTestSnap(t)
	g.GiveStockForTest(1)
	g.SetCurrentTurnIdxForTest(1)

	g.StepForTest(1)
	assert.Equal(t, SnapEventEliminated, g.GetLastEvent().Kind)
	assert.Equal(t, 0, g.GetCurrentTurnIdx(), "ストックのある席へ回る")
}

// **1 人が全札を集めたら終わり。**
func TestSnap_GameEndsWhenOnePlayerHoldsEverything(t *testing.T) {
	g := newTestSnap(t)
	all := make([]*Card, 0, SnapDeckSize)
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= 13; v++ {
			all = append(all, NewCard(d, v, false))
		}
	}
	// 場に同ランクの 2 枚、残りは席 0 が全部持っている。
	pile := all[:2]
	pile[1] = NewCard(CardDesignHeart, pile[0].GetValue(), false)
	g.GiveStockForTest(0, all[2:]...)
	g.GiveStockForTest(1)
	g.SetCenterPileForTest(pile)
	g.SetCurrentTurnIdxForTest(0)
	require.True(t, g.IsSnapAvailable())

	// **総取りした結果、場が空で 1 人が全札を持つと終わる。**
	require.NoError(t, g.PlayerSnap())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, SnapDeckSize, g.GetPlayer(0).GetStockSize())

	// **負のコントロール: 相手に札が残っていれば終わらない。**
	h := newTestSnap(t)
	h.GiveStockForTest(0, all[2:10]...)
	h.GiveStockForTest(1, all[10:14]...)
	h.SetCenterPileForTest(pile)
	require.NoError(t, h.PlayerSnap())
	assert.False(t, h.GetGameEndFlag())
}

func TestSnap_GiveUp(t *testing.T) {
	g := newTestSnap(t)
	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx(), "投了した席以外の最多保持者")
	assert.Equal(t, SnapPendingNone, g.GetPending().Kind, "保留は消える")

	before := g.GetWinnerIdx()
	g.GiveUp()
	assert.Equal(t, before, g.GetWinnerIdx())
	assert.Error(t, g.PlayerStep(), "終局後はめくれない")
	assert.Error(t, g.PlayerSnap(), "終局後は宣言できない")
}

// **宣言が成立しているなら、CPU は「めくる」でなく「宣言」を予約する。**
func TestSnap_SchedulesASnapWhenOneIsAvailable(t *testing.T) {
	g := newTestSnap(t)
	g.SetCenterPileForTest([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
	})
	g.SetCurrentTurnIdxForTest(0)
	g.ScheduleNextForTest()

	p := g.GetPending()
	assert.Equal(t, SnapPendingSnap, p.Kind)
	assert.Positive(t, p.PlayerIdx, "予約は CPU のもの")
	assert.Greater(t, p.DeadlineMs, int64(0))
}

// **人間の手番で宣言が成立していなければ、CPU は何も予約しない。**
func TestSnap_SchedulesNothingWhileWaitingForTheHuman(t *testing.T) {
	g := newTestSnap(t)
	g.SetCenterPileForTest([]*Card{NewCard(CardDesignSpade, 7, false)})
	g.SetCurrentTurnIdxForTest(0)
	g.ScheduleNextForTest()
	assert.Equal(t, SnapPendingNone, g.GetPending().Kind)
}

// **期限が来るまで Tick は何もしない。** これが人間の勝ち目。
func TestSnap_TickWaitsForTheDeadline(t *testing.T) {
	g := newTestSnap(t)
	g.SetClock(fixedSnapClock(1000))
	g.SetCenterPileForTest([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
	})
	g.SetPendingForTest(SnapPending{Kind: SnapPendingSnap, PlayerIdx: 1, DeadlineMs: 2000})

	assert.Equal(t, SnapPendingNone, g.Tick(), "まだ期限前")
	assert.Equal(t, SnapPendingSnap, g.GetPending().Kind, "予約は残る")

	g.SetClock(fixedSnapClock(2000))
	assert.Equal(t, SnapPendingSnap, g.Tick(), "期限で発火")
	assert.Equal(t, 1, g.GetLastEvent().PlayerIdx, "CPU が取った")
	assert.Equal(t, SnapEventSnapCorrect, g.GetLastEvent().Kind)
}

// **人間が先に宣言すれば CPU の予約より速く取れる。**
func TestSnap_HumanCanBeatThePendingCpuSnap(t *testing.T) {
	g := newTestSnap(t)
	g.SetClock(fixedSnapClock(1000))
	g.SetCenterPileForTest([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
	})
	g.SetPendingForTest(SnapPending{Kind: SnapPendingSnap, PlayerIdx: 1, DeadlineMs: 2000})

	require.NoError(t, g.PlayerSnap())
	assert.Equal(t, 0, g.GetLastEvent().PlayerIdx)
	assert.Equal(t, SnapEventSnapCorrect, g.GetLastEvent().Kind)
	// 場が空になったので、CPU の宣言予約は残らない。
	assert.NotEqual(t, SnapPendingSnap, g.GetPending().Kind)
}

func TestSnap_TickDoesNothingWithoutAPending(t *testing.T) {
	g := newTestSnap(t)
	g.SetPendingForTest(SnapPending{Kind: SnapPendingNone})
	assert.Equal(t, SnapPendingNone, g.Tick())

	g.GiveUp()
	assert.Equal(t, SnapPendingNone, g.Tick(), "終局後は動かない")
}

// **反応時間は難易度で変わり、下限で止まる。**
func TestSnap_ReactionTimeRespectsTheDifficultyAndFloor(t *testing.T) {
	means := map[SnapCpuDifficulty]int{}
	for _, d := range []SnapCpuDifficulty{SnapCpuEasy, SnapCpuNormal, SnapCpuHard} {
		g := NewSnap(nil, SnapConfig{PlayerCnt: 2, CpuDifficulty: d})
		g.SetRand(rand.New(rand.NewSource(7)))
		total := 0
		for range 200 {
			ms := g.ReactionMsForTest()
			assert.GreaterOrEqual(t, ms, SnapMinReactionMs, "下限でクランプする")
			total += ms
		}
		means[d] = total / 200
	}
	assert.Greater(t, means[SnapCpuEasy], means[SnapCpuNormal], "易しいほど遅い")
	assert.Greater(t, means[SnapCpuNormal], means[SnapCpuHard], "難しいほど速い")
}

func TestSnap_Hint(t *testing.T) {
	g := newTestSnap(t)

	g.SetCenterPileForTest([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
	})
	h := g.GetHint()
	require.NotNil(t, h)
	assert.True(t, h.Snap)
	assert.Equal(t, "snapDeclare", h.Reason)

	g.SetCenterPileForTest([]*Card{NewCard(CardDesignSpade, 7, false)})
	g.SetCurrentTurnIdxForTest(0)
	h = g.GetHint()
	require.NotNil(t, h)
	assert.False(t, h.Snap)
	assert.Equal(t, "snapStep", h.Reason)

	g.SetCurrentTurnIdxForTest(1)
	assert.Equal(t, "snapWait", g.GetHint().Reason)

	g.GiveUp()
	assert.Nil(t, g.GetHint(), "終局後は助言しない")
}

func TestSnap_AccessorsAndBounds(t *testing.T) {
	g := newTestSnap(t)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.Equal(t, SnapDefaultPlayerCnt, g.GetPlayerCnt())
	assert.Equal(t, SnapPhasePlay, g.GetPhase())
	assert.NotEmpty(t, g.GetActionLog())
	assert.Equal(t, DefaultSnapConfig(), g.GetConfig())

	g.SetCenterPileForTest(nil)
	assert.Nil(t, g.GetTopCard())
	g.SetCenterPileForTest([]*Card{NewCard(CardDesignSpade, 3, false)})
	require.NotNil(t, g.GetTopCard())
	assert.Equal(t, 3, g.GetTopCard().GetValue())

	// **人数を変えたら席も作り直す。**
	g.SetConfig(SnapConfig{PlayerCnt: 4, CpuDifficulty: SnapCpuHard})
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.NotPanics(t, g.Reset)
	assert.NotNil(t, g.GetPlayer(3))

	// nil を渡しても差し替えない。
	assert.NotPanics(t, func() { g.SetClock(nil); g.SetRand(nil) })
}

func TestSnapConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultSnapConfig().Validate())
	for n := SnapPlayerCntMin; n <= SnapPlayerCntMax; n++ {
		assert.NoError(t, SnapConfig{PlayerCnt: n, CpuDifficulty: SnapCpuNormal}.Validate())
	}
	assert.Error(t, SnapConfig{PlayerCnt: SnapPlayerCntMin - 1}.Validate())
	assert.Error(t, SnapConfig{PlayerCnt: SnapPlayerCntMax + 1}.Validate())
	assert.Error(t, SnapConfig{PlayerCnt: 2, CpuDifficulty: -1}.Validate())
	assert.Error(t, SnapConfig{PlayerCnt: 2, CpuDifficulty: SnapCpuHard + 1}.Validate())

	// **不正な設定で作っても遊べる盤面になる。**
	g := NewSnap(nil, SnapConfig{PlayerCnt: 99})
	assert.Equal(t, SnapDefaultPlayerCnt, g.GetPlayerCnt())
}

func TestSnap_JSONRoundTrip(t *testing.T) {
	g := newTestSnap(t)
	g.StepForTest(0)
	g.StepForTest(g.GetCurrentTurnIdx())

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored Snap
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetCenterPileSize(), restored.GetCenterPileSize())
	assert.Equal(t, g.GetCurrentTurnIdx(), restored.GetCurrentTurnIdx())
	assert.Equal(t, g.GetPending(), restored.GetPending())
	for i := range g.GetPlayerCnt() {
		assert.Equal(t, g.GetPlayer(i).GetStockSize(), restored.GetPlayer(i).GetStockSize())
	}
	// **復元しても動く。** 時計と乱数が nil のままだと Tick で落ちる。
	assert.NotPanics(t, func() { restored.Tick() })
}

// **壊れたスナップショットは弾く。**
func TestSnap_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	snapshot := func(t *testing.T) map[string]any {
		t.Helper()
		g := newTestSnap(t)
		g.StepForTest(0)
		data, err := json.Marshal(g)
		require.NoError(t, err)
		var out map[string]any
		require.NoError(t, json.Unmarshal(data, &out))
		return out
	}

	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"phase out of range", func(m map[string]any) { m["ph"] = 9 }},
		{"game end flag without the phase", func(m map[string]any) { m["ge"] = true }},
		{"current turn out of range", func(m map[string]any) { m["ct"] = 9 }},
		{"winner before the game ended", func(m map[string]any) { m["wi"] = 1 }},
		{"config out of range", func(m map[string]any) { m["cf"] = map[string]any{"p": 9, "d": 1} }},
		{"pending kind out of range", func(m map[string]any) {
			m["pd"] = map[string]any{"kind": 9, "playerIdx": 1, "deadlineMs": 1}
		}},
		// **保留の種類と席は対。**
		{"pending payload with no action", func(m map[string]any) {
			m["pd"] = map[string]any{"kind": 0, "playerIdx": 1, "deadlineMs": 500}
		}},
		{"pending action for the human", func(m map[string]any) {
			m["pd"] = map[string]any{"kind": 1, "playerIdx": 0, "deadlineMs": 500}
		}},
		{"last event kind out of range", func(m map[string]any) {
			m["le"] = map[string]any{"kind": 9, "playerIdx": 0}
		}},
		{"last event player out of range", func(m map[string]any) {
			m["le"] = map[string]any{"kind": 1, "playerIdx": 9}
		}},
		{"a nil card in the pile", func(m map[string]any) { m["cp"] = []any{nil} }},
		// **札は 52 枚しかない（#5314 で踏んだ「数え元と突き合わせる」形）。**
		{"a card appears from nowhere", func(m map[string]any) {
			pile := m["cp"].([]any)
			m["cp"] = append(pile, map[string]any{"d": 1, "v": 5, "j": false})
		}},
		{"the pile loses a card", func(m map[string]any) { m["cp"] = []any{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := snapshot(t)
			tc.mutate(s)
			data, err := json.Marshal(s)
			require.NoError(t, err)
			var restored Snap
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通り、使っても落ちない。**
	data, err := json.Marshal(snapshot(t))
	require.NoError(t, err)
	var ok Snap
	require.NoError(t, json.Unmarshal(data, &ok))
	assert.NotPanics(t, func() {
		_ = ok.IsSnapAvailable()
		_ = ok.GetHint()
		ok.Tick()
	})
}

// **ストックの中身も見る。**
func TestSnapPlayer_UnmarshalRejectsANilCard(t *testing.T) {
	var p SnapPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"st":[null]}`), &p))

	var ok SnapPlayer
	require.NoError(t, json.Unmarshal([]byte(`{"st":[]}`), &ok))
	assert.Zero(t, ok.GetStockSize())
	assert.False(t, ok.HasStock())
}

// **誰も動かせない盤面を「まだ続いている」と言わない。**
//
// 最後の 1 枚を出すと全員のストックが空になり得ます。めくった直後に終局を
// 見ないと、人間が空のストックをめくって初めて終わりが分かる形になります。
func TestSnap_EndsWhenTheLastCardLeavesEveryStock(t *testing.T) {
	g := newTestSnap(t)
	all := make([]*Card, 0, SnapDeckSize)
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= 13; v++ {
			all = append(all, NewCard(d, v, false))
		}
	}
	g.GiveStockForTest(0, all[0])
	g.GiveStockForTest(1)
	g.SetCenterPileForTest(all[1:])
	g.SetCurrentTurnIdxForTest(0)

	require.NoError(t, g.PlayerStep())
	assert.True(t, g.GetGameEndFlag(), "誰も動かせないのに続いている状態を返さない")
	assert.Equal(t, -1, g.GetWinnerIdx(), "全札が場に出たままなので勝者はいない")
	assert.Equal(t, SnapPendingNone, g.GetPending().Kind, "予約も残さない")

	// **負のコントロール: 相手に札が残っていれば続く。**
	h := newTestSnap(t)
	h.GiveStockForTest(0, all[0])
	h.GiveStockForTest(1, all[1], all[2])
	h.SetCenterPileForTest(all[3:])
	h.SetCurrentTurnIdxForTest(0)
	require.NoError(t, h.PlayerStep())
	assert.False(t, h.GetGameEndFlag())
}

// **どの局も必ず終わる。** 時計を進めながら最後まで回す。
func TestSnap_GamesTerminate(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		g := NewDefaultSnap()
		ms := int64(0)
		g.SetClock(func() time.Time { return time.UnixMilli(ms) })
		g.SetRand(rand.New(rand.NewSource(seed)))
		g.Reset()

		for turns := 0; !g.GetGameEndFlag(); turns++ {
			require.Less(t, turns, 20000, "seed %d: 終わらない", seed)
			ms += 5000 // どんな反応時間でも必ず期限を越える
			if g.GetPending().Kind != SnapPendingNone {
				g.Tick()
				continue
			}
			if g.IsSnapAvailable() {
				require.NoError(t, g.PlayerSnap())
				continue
			}
			if g.GetCurrentTurnIdx() == 0 {
				require.NoError(t, g.PlayerStep())
				continue
			}
			g.StepForTest(g.GetCurrentTurnIdx())
		}

		// **札は増えも減りもしない。**
		total := g.GetCenterPileSize()
		for i := range g.GetPlayerCnt() {
			total += g.GetPlayer(i).GetStockSize()
		}
		assert.Equal(t, SnapDeckSize, total, "seed %d", seed)
	}
}
