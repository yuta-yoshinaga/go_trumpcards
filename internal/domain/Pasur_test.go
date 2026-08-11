//go:build test

package domain

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPasur(t *testing.T) *Pasur {
	t.Helper()
	p := NewDefaultPasur()
	p.Reset()
	return p
}

// **場に 4 枚置いた残り 48 枚は 2/3/4 人のどれでも割り切れる。**
func TestPasur_DealDividesForEveryPlayerCount(t *testing.T) {
	for n := PasurPlayerCntMin; n <= PasurPlayerCntMax; n++ {
		rest := 52 - PasurInitialTableSize
		assert.Zero(t, rest%(n*PasurHandSize), "%d 人で割り切れる", n)

		p := NewPasur(nil, PasurConfig{PlayerCnt: n})
		p.Reset()
		assert.Len(t, p.GetTableCards(), PasurInitialTableSize)
		total := len(p.GetTableCards()) + p.GetDeckRemaining()
		for i := range n {
			assert.Equal(t, PasurHandSize, p.GetPlayer(i).GetCardsSize())
			total += p.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, 52, total, "%d 人でも 52 枚すべてに行き先がある", n)
	}
}

// **1 デッキから出る得点はちょうど 21。** issue に得点表が無いので自分で決めた。
func TestPasur_TotalScoreIsFixed(t *testing.T) {
	total := 0
	for design := CardDesignSpade; design <= CardDesignDiamond; design++ {
		for value := 1; value <= 13; value++ {
			total += pasurCardScore(NewCard(design, value, false))
		}
	}
	assert.Equal(t, PasurTotalScore, total)
	assert.Equal(t, 21, total, "クラブ13 + 2♣の上乗せ1 + 10♦の3 + エース4")
}

func TestPasur_CardScores(t *testing.T) {
	for _, tc := range []struct {
		name          string
		design, value int
		want          int
	}{
		{"ふつうのクラブ", CardDesignClover, 7, 1},
		{"2♣ は 2 点", CardDesignClover, 2, 2},
		{"A♣ はクラブ+エース", CardDesignClover, 1, 2},
		{"10♦", CardDesignDiamond, 10, 3},
		{"ふつうのダイヤ", CardDesignDiamond, 7, 0},
		{"A♠", CardDesignSpade, 1, 1},
		{"K♠ は 0 点", CardDesignSpade, 13, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, pasurCardScore(NewCard(tc.design, tc.value, false)))
		})
	}
}

// **数札は合計 11、手札 1 枚だけでは絶対に取れない。**
func TestPasur_NumeralCapturesSumToEleven(t *testing.T) {
	p := newTestPasur(t)
	p.SetTableForTest([]*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 7, false),
	})
	pasurHandOf(p, 0, NewCard(CardDesignDiamond, 4, false))

	opts := p.GetCaptureOptions(0, 0)
	// 4 を出すと 7 が要る: 単独の 7、または 4+3。
	require.Len(t, opts, 2)
	sums := map[int]bool{}
	for _, o := range opts {
		s := 0
		for _, i := range o {
			s += p.GetTableCards()[i].GetValue()
		}
		sums[s] = true
		assert.NotEmpty(t, o, "場の札を使わない捕獲は無い")
	}
	assert.Equal(t, map[int]bool{7: true}, sums, "どの組み合わせも 11-4 = 7")
}

// **同じ値の札が場に複数あっても、候補は札ごとに 1 通りずつ。**
//
// 位置で列挙するので、値が同じでも**別の札は別の選択肢**になり、同じ組み合わせが
// 二重に出ることはありません。捕獲の検証も値ではなく位置の集合で比べます。
func TestPasur_DuplicateTableValuesGiveDistinctOptions(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(0)
	p.SetTableForTest([]*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
	})

	// ♠7 は 4 が要る。4 が 3 枚あるので単独 3 通り。
	pasurHandOf(p, 0, NewCard(CardDesignSpade, 7, false))
	single := p.GetCaptureOptions(0, 0)
	assert.ElementsMatch(t, [][]int{{0}, {1}, {3}}, single)

	// ♠4 は 7 が要る。4+3 の組が 3 通り。
	pasurHandOf(p, 0, NewCard(CardDesignSpade, 4, false))
	pairs := p.GetCaptureOptions(0, 0)
	assert.ElementsMatch(t, [][]int{{0, 2}, {1, 2}, {2, 3}}, pairs)

	seen := map[string]bool{}
	for _, o := range append(append([][]int{}, single...), pairs...) {
		key := fmt.Sprint(o)
		assert.False(t, seen[key], "同じ組み合わせが二度出ない: %s", key)
		seen[key] = true
	}

	// **どの候補も実際に受理される。**
	for _, o := range pairs {
		q := newTestPasur(t)
		q.SetCurrentPlayerIdxForTest(0)
		q.SetTableForTest([]*Card{
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 3, false),
			NewCard(CardDesignDiamond, 4, false),
		})
		pasurHandOf(q, 0, NewCard(CardDesignSpade, 4, false))
		assert.NoError(t, q.PlayForTest(0, 0, o), "候補 %v", o)
	}
}

// **絵札は数値に混ぜず、同ランクだけ取れる。**
func TestPasur_FaceCardsCaptureOnlyTheSameRank(t *testing.T) {
	p := newTestPasur(t)
	p.SetTableForTest([]*Card{
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 11, false),
		NewCard(CardDesignDiamond, 5, false),
	})
	pasurHandOf(p, 0,
		NewCard(CardDesignDiamond, 12, false), // Q: 場の Q 2 枚を取る
		NewCard(CardDesignSpade, 13, false),   // K: 場に K が無いので取れない
	)

	opts := p.GetCaptureOptions(0, 0)
	require.Len(t, opts, 1)
	assert.ElementsMatch(t, []int{0, 1}, opts[0], "同ランクをまとめて取る")
	assert.Empty(t, p.GetCaptureOptions(0, 1), "同ランクが無ければ取れない")

	// **絵札は数札の合計に混ざらない。** 場の J(11) を 11 として使えない。
	pasurHandOf(p, 0, NewCard(CardDesignDiamond, 6, false))
	for _, o := range p.GetCaptureOptions(0, 0) {
		for _, i := range o {
			assert.False(t, pasurIsFace(p.GetTableCards()[i]), "絵札は合計に使わない")
		}
	}
}

// **場を空にしたらスール。** その捕獲だけが倍になる。
func TestPasur_SoorDoublesOnlyThatCapture(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(0)
	// 場は 2♣ 1 枚だけ。9 を出すと 2 を取って場が空になる。
	p.SetTableForTest([]*Card{NewCard(CardDesignClover, 2, false)})
	pasurHandOf(p, 0, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignSpade, 3, false))

	require.NoError(t, p.PlayForTest(0, 0, []int{0}))
	assert.Empty(t, p.GetTableCards())
	assert.Equal(t, 1, p.GetPlayer(0).GetSoors())
	assert.Len(t, p.GetPlayer(0).GetSoorCaptured(), 2, "出した札も一緒に取る")
	assert.Empty(t, p.GetPlayer(0).GetCaptured(), "通常の山には入らない")
	// 2♣ は 2 点、♠9 は 0 点。スールなので倍で 4 点。
	assert.Equal(t, 4, p.ScoreOfForTest(0))
}

// **場が残るなら通常の捕獲。** 倍にはならない。
func TestPasur_NormalCaptureIsNotDoubled(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(0)
	p.SetTableForTest([]*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignSpade, 5, false),
	})
	pasurHandOf(p, 0, NewCard(CardDesignSpade, 9, false))

	require.NoError(t, p.PlayForTest(0, 0, []int{0}))
	assert.Len(t, p.GetTableCards(), 1, "場が残る")
	assert.Zero(t, p.GetPlayer(0).GetSoors())
	assert.Equal(t, 2, p.ScoreOfForTest(0), "2♣ の 2 点だけ")
}

// **取れるときにトレールはできない。** 点札を場に流す戦術を成立させない。
func TestPasur_MustCaptureWhenPossible(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(0)
	p.SetTableForTest([]*Card{NewCard(CardDesignSpade, 7, false)})
	pasurHandOf(p, 0, NewCard(CardDesignDiamond, 4, false))

	assert.Error(t, p.PlayForTest(0, 0, nil), "取れるのに置くのは違法")
	assert.Error(t, p.PlayForTest(0, 0, []int{}), "空スライスでも同じ")
	assert.NoError(t, p.PlayForTest(0, 0, []int{0}))
}

func TestPasur_TrailWhenNothingCanBeCaptured(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(0)
	p.SetTableForTest([]*Card{NewCard(CardDesignSpade, 13, false)})
	pasurHandOf(p, 0, NewCard(CardDesignDiamond, 4, false))

	require.NoError(t, p.PlayForTest(0, 0, nil))
	assert.Len(t, p.GetTableCards(), 2, "場に置かれる")
	assert.Zero(t, p.GetPlayer(0).GetCapturedCount())
}

func TestPasur_RejectsIllegalCaptures(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(0)
	p.SetTableForTest([]*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 9, false),
	})
	pasurHandOf(p, 0, NewCard(CardDesignDiamond, 4, false))

	assert.Error(t, p.PlayForTest(0, 0, []int{1}), "合計が 11 にならない")
	assert.Error(t, p.PlayForTest(0, 0, []int{0, 1}), "多すぎる")
	assert.Error(t, p.PlayForTest(0, 0, []int{9}), "場に無いインデックス")
	assert.Error(t, p.PlayForTest(0, -1, nil))
	assert.Error(t, p.PlayForTest(0, 99, nil))
	assert.Error(t, p.PlayForTest(1, 0, nil), "手番でない席は打てない")
}

// **場に残った札は最後に取った人のもの。**
func TestPasur_LeftoversGoToTheLastCapturer(t *testing.T) {
	p := newTestPasur(t)
	p.SetTableForTest([]*Card{NewCard(CardDesignClover, 5, false)})
	p.SetLastCaptureIdxForTest(1)
	p.EmptyHandsForTest()
	p.DrainDeckForTest()
	p.FinishGameForTest()

	assert.Empty(t, p.GetTableCards())
	assert.Equal(t, 1, p.GetPlayer(1).GetCapturedCount())
	assert.Equal(t, 1, p.GetScore(1), "♣5 の 1 点")

	// **誰も取っていなければ場に残す。** 存在しない席へ入れない。
	q := newTestPasur(t)
	q.SetTableForTest([]*Card{NewCard(CardDesignClover, 5, false)})
	q.SetLastCaptureIdxForTest(-1)
	q.EmptyHandsForTest()
	q.DrainDeckForTest()
	q.FinishGameForTest()
	assert.Len(t, q.GetTableCards(), 1)
	for i := range q.GetPlayerCnt() {
		assert.Zero(t, q.GetPlayer(i).GetCapturedCount())
	}
}

// **手札が尽きたら配り足し、山札が尽きたら終わる。**
func TestPasur_DealsAnotherPackUntilTheDeckRunsOut(t *testing.T) {
	p := newTestPasur(t)
	assert.Equal(t, 1, p.GetPacksDealt())
	assert.Equal(t, 52-PasurInitialTableSize-PasurDefaultPlayerCnt*PasurHandSize, p.GetDeckRemaining())

	for !p.GetGameEndFlag() {
		idx, table := p.CpuChoiceForTest(p.GetCurrentPlayerIdx())
		require.NoError(t, p.PlayForTest(p.GetCurrentPlayerIdx(), idx, table))
	}
	assert.Zero(t, p.GetDeckRemaining())
	assert.Equal(t, 3, p.GetPacksDealt(), "4 人なら 48 / 16 = 3 パック")
	assert.Equal(t, PasurPhaseGameEnd, p.GetPhase())
}

// **どの人数でも必ず終わり、52 枚が行き先を持つ。**
func TestPasur_GamesTerminateForEveryPlayerCount(t *testing.T) {
	for n := PasurPlayerCntMin; n <= PasurPlayerCntMax; n++ {
		for range 10 {
			p := NewPasur(nil, PasurConfig{PlayerCnt: n})
			p.Reset()
			for turns := 0; !p.GetGameEndFlag(); turns++ {
				require.Less(t, turns, 200, "終わらない (%d 人)", n)
				idx, table := p.CpuChoiceForTest(p.GetCurrentPlayerIdx())
				require.NoError(t, p.PlayForTest(p.GetCurrentPlayerIdx(), idx, table))
			}
			held := len(p.GetTableCards())
			for i := range n {
				held += p.GetPlayer(i).GetCapturedCount()
			}
			assert.Equal(t, 52, held, "%d 人: 52 枚すべてがどこかにある", n)
			assert.NotEmpty(t, p.GetWinners())
		}
	}
}

// **スールが無ければ、配られた得点はちょうど 21 に収まる。**
func TestPasur_ScoresAddUpToTheDeckTotal(t *testing.T) {
	for range 20 {
		p := NewDefaultPasur()
		p.Reset()
		for !p.GetGameEndFlag() {
			idx, table := p.CpuChoiceForTest(p.GetCurrentPlayerIdx())
			require.NoError(t, p.PlayForTest(p.GetCurrentPlayerIdx(), idx, table))
		}
		total, soors := 0, 0
		for i := range p.GetPlayerCnt() {
			total += p.GetScore(i)
			soors += p.GetPlayer(i).GetSoors()
		}
		if soors == 0 {
			// 場に残った札が誰にも入らないケースがあるので上限で見る。
			assert.LessOrEqual(t, total, PasurTotalScore)
		}
		assert.GreaterOrEqual(t, total, 0)
	}
}

// **CPU は必ず合法手を返す。**
func TestPasur_CpuAlwaysChoosesLegally(t *testing.T) {
	for range 20 {
		p := NewDefaultPasur()
		p.Reset()
		for !p.GetGameEndFlag() {
			idx, table := p.CpuChoiceForTest(p.GetCurrentPlayerIdx())
			require.NoError(t, p.PlayForTest(p.GetCurrentPlayerIdx(), idx, table))
		}
	}
}

// **CPU は場を空にできるならそうする。** 倍になるので。
func TestPasur_CpuPrefersASoor(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(1)
	p.SetTableForTest([]*Card{NewCard(CardDesignSpade, 4, false)})
	pasurHandOf(p, 1,
		NewCard(CardDesignHeart, 7, false), // ♠4 を取って場が空 → スール
		NewCard(CardDesignClover, 3, false),
	)
	idx, table := p.CpuChoiceForTest(1)
	assert.Equal(t, 0, idx)
	assert.Equal(t, []int{0}, table)
}

// **取れないときは、いちばん点にならない札を置く。**
func TestPasur_CpuTrailsItsCheapestCard(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(1)
	p.SetTableForTest([]*Card{NewCard(CardDesignSpade, 13, false)})
	pasurHandOf(p, 1,
		NewCard(CardDesignClover, 2, false), // 2 点 — 置きたくない
		NewCard(CardDesignHeart, 5, false),  // 0 点
	)
	idx, table := p.CpuChoiceForTest(1)
	assert.Equal(t, 1, idx, "点にならない札を置く")
	assert.Empty(t, table)
}

func TestPasur_PublicEntryPointsGuardTheTurn(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(1)
	assert.False(t, p.IsHumanTurn())
	assert.Error(t, p.PlayerPlay(0, nil))
	before := p.GetPlayer(1).GetCardsSize()
	p.CpuPlay()
	assert.Equal(t, before-1, p.GetPlayer(1).GetCardsSize())

	p.SetCurrentPlayerIdxForTest(0)
	assert.True(t, p.IsHumanTurn())
	humanBefore := p.GetPlayer(0).GetCardsSize()
	p.CpuPlay()
	assert.Equal(t, humanBefore, p.GetPlayer(0).GetCardsSize(), "人間の番では CPU は動かない")

	idx, table := p.CpuChoiceForTest(0)
	require.NoError(t, p.PlayerPlay(idx, table))
}

func TestPasur_GiveUp(t *testing.T) {
	p := newTestPasur(t)
	p.GiveUp()
	assert.True(t, p.GetGameEndFlag())
	assert.NotContains(t, p.GetWinners(), 0, "投了した席は勝者にならない")
	assert.Len(t, p.GetWinners(), PasurDefaultPlayerCnt-1)
	before := p.GetWinners()
	p.GiveUp()
	assert.Equal(t, before, p.GetWinners())
	assert.Error(t, p.PlayForTest(0, 0, nil), "終局後は打てない")
}

func TestPasur_Hint(t *testing.T) {
	p := newTestPasur(t)
	p.SetCurrentPlayerIdxForTest(0)
	p.SetTableForTest([]*Card{NewCard(CardDesignSpade, 7, false)})
	pasurHandOf(p, 0, NewCard(CardDesignDiamond, 4, false))

	h := p.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Equal(t, "pasurSoor", h.Reason, "場を空にできる")
	assert.Equal(t, []int{0}, h.TableIndices)

	// 場が残るなら通常の捕獲。
	p.SetTableForTest([]*Card{NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 9, false)})
	pasurHandOf(p, 0, NewCard(CardDesignDiamond, 4, false))
	assert.Equal(t, "pasurCapture", p.GetHint().Reason)

	// 取れないならトレール。
	p.SetTableForTest([]*Card{NewCard(CardDesignSpade, 13, false)})
	pasurHandOf(p, 0, NewCard(CardDesignDiamond, 4, false))
	h = p.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "pasurTrail", h.Reason)
	assert.Empty(t, h.TableIndices)

	p.SetCurrentPlayerIdxForTest(1)
	assert.Nil(t, p.GetHint(), "相手の手番では助言しない")

	p.SetCurrentPlayerIdxForTest(0)
	p.EmptyHandsForTest()
	assert.Nil(t, p.GetHint(), "手札が無ければ助言しない")

	p.GiveUp()
	assert.Nil(t, p.GetHint(), "終局後は助言しない")
}

func TestPasur_AccessorsAndBounds(t *testing.T) {
	p := newTestPasur(t)
	assert.Nil(t, p.GetPlayer(-1))
	assert.Nil(t, p.GetPlayer(99))
	assert.Zero(t, p.GetScore(-1))
	assert.Zero(t, p.GetScore(99))
	assert.Nil(t, p.GetCaptureOptions(-1, 0))
	assert.Nil(t, p.GetCaptureOptions(0, -1))
	assert.Nil(t, p.GetCaptureOptions(0, 99))
	assert.Equal(t, PasurDefaultPlayerCnt, p.GetPlayerCnt())
	assert.Equal(t, -1, p.GetLastCaptureIdx())
	assert.NotEmpty(t, p.GetActionLog())

	assert.Equal(t, PasurDefaultPlayerCnt, p.GetConfig().PlayerCnt)

	// **人数を変えたら席も作り直す。**
	p.SetConfig(PasurConfig{PlayerCnt: 2})
	assert.Equal(t, 2, p.GetConfig().PlayerCnt)
	assert.Equal(t, 2, p.GetPlayerCnt())
	assert.NotPanics(t, p.Reset)
	assert.Equal(t, PasurHandSize, p.GetPlayer(1).GetCardsSize())
}

func TestPasurConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultPasurConfig().Validate())
	for n := PasurPlayerCntMin; n <= PasurPlayerCntMax; n++ {
		assert.NoError(t, PasurConfig{PlayerCnt: n}.Validate())
	}
	assert.Error(t, PasurConfig{PlayerCnt: PasurPlayerCntMin - 1}.Validate())
	assert.Error(t, PasurConfig{PlayerCnt: PasurPlayerCntMax + 1}.Validate())

	// **不正な設定で作っても遊べる盤面になる。**
	p := NewPasur(nil, PasurConfig{PlayerCnt: 99})
	assert.Equal(t, PasurDefaultPlayerCnt, p.GetPlayerCnt())
}

func TestPasur_JSONRoundTrip(t *testing.T) {
	p := newTestPasur(t)
	for range 6 {
		idx, table := p.CpuChoiceForTest(p.GetCurrentPlayerIdx())
		require.NoError(t, p.PlayForTest(p.GetCurrentPlayerIdx(), idx, table))
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored Pasur
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, p.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Len(t, restored.GetTableCards(), len(p.GetTableCards()), "場の札が消えない")
	assert.Equal(t, p.GetDeckRemaining(), restored.GetDeckRemaining())
	assert.Equal(t, p.GetLastCaptureIdx(), restored.GetLastCaptureIdx())
	for i := range p.GetPlayerCnt() {
		assert.Equal(t, p.GetPlayer(i).GetCapturedCount(), restored.GetPlayer(i).GetCapturedCount())
		assert.Equal(t, p.GetPlayer(i).GetSoors(), restored.GetPlayer(i).GetSoors(), "スール回数が消えない")
	}
	assert.Equal(t, p.GetScore(0), restored.GetScore(0))
}

// **壊れたスナップショットは弾く。**
func TestPasur_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	snapshot := func(t *testing.T, ended bool) map[string]any {
		t.Helper()
		p := newTestPasur(t)
		if ended {
			for !p.GetGameEndFlag() {
				idx, table := p.CpuChoiceForTest(p.GetCurrentPlayerIdx())
				require.NoError(t, p.PlayForTest(p.GetCurrentPlayerIdx(), idx, table))
			}
		}
		data, err := json.Marshal(p)
		require.NoError(t, err)
		var out map[string]any
		require.NoError(t, json.Unmarshal(data, &out))
		return out
	}

	for _, tc := range []struct {
		name   string
		ended  bool
		mutate func(map[string]any)
	}{
		{"phase out of range", false, func(m map[string]any) { m["ph"] = 9 }},
		// **終了フラグとフェーズは対（#5313 で踏んだ形）。**
		{"game end flag without the phase", false, func(m map[string]any) { m["ge"] = true }},
		{"game end phase without the flag", true, func(m map[string]any) { m["ge"] = false }},
		{"current player out of range", false, func(m map[string]any) { m["ci"] = 9 }},
		{"last capture out of range", false, func(m map[string]any) { m["lc"] = 9 }},
		{"negative packs dealt", false, func(m map[string]any) { m["pd"] = -1 }},
		{"config out of range", false, func(m map[string]any) { m["cf"] = map[string]any{"p": 9} }},
		{"player count disagrees with the seats", false, func(m map[string]any) { m["cf"] = map[string]any{"p": 3} }},
		// **得点と勝者は終局とセット。**
		{"scores before the game ended", false, func(m map[string]any) { m["sr"] = []any{1, 2, 3, 4} }},
		{"winners before the game ended", false, func(m map[string]any) { m["wn"] = []any{0} }},
		{"no winners after it ended", true, func(m map[string]any) { m["wn"] = []any{} }},
		{"winner out of range", true, func(m map[string]any) { m["wn"] = []any{9} }},
		{"scores has the wrong length", true, func(m map[string]any) { m["sr"] = []any{1} }},
		{"a nil table card", false, func(m map[string]any) { m["tb"] = []any{nil} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := snapshot(t, tc.ended)
			tc.mutate(s)
			data, err := json.Marshal(s)
			require.NoError(t, err)
			var restored Pasur
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通り、使っても落ちない。**
	for _, ended := range []bool{false, true} {
		data, err := json.Marshal(snapshot(t, ended))
		require.NoError(t, err)
		var ok Pasur
		require.NoError(t, json.Unmarshal(data, &ok))
		assert.NotPanics(t, func() {
			_ = ok.GetCaptureOptions(0, 0)
			_ = ok.GetScore(0)
			_ = ok.GetHint()
		})
	}
}

// **スールの回数と札の山は対。** 回数だけ立つと根拠なく倍付けされる。
func TestPasurPlayer_UnmarshalRejectsAHalfSoor(t *testing.T) {
	for _, body := range []string{
		`{"so":1}`,
		`{"so":0,"sc":[{"d":1,"v":5,"j":false}]}`,
		`{"so":-1}`,
	} {
		var p PasurPlayer
		assert.Error(t, json.Unmarshal([]byte(body), &p), body)
	}
	for _, body := range []string{
		`{"so":0}`,
		`{"so":1,"sc":[{"d":1,"v":5,"j":false}]}`,
	} {
		var p PasurPlayer
		assert.NoError(t, json.Unmarshal([]byte(body), &p), body)
	}
}
