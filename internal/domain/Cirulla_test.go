//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCirullaForTest(t *testing.T) *Cirulla {
	t.Helper()
	c := NewDefaultCirulla()
	c.Reset()
	return c
}

// cirullaCard は札を作る薄い別名。
func cirullaCard(suit, value int) *Card { return NewCard(suit, value, false) }

// **値は A=1, 2-7 はそのまま, J=8, Q=9, K=10。**
func TestCirullaCardValue(t *testing.T) {
	for _, tt := range []struct{ value, want int }{
		{1, 1}, {2, 2}, {7, 7}, {11, 8}, {12, 9}, {13, 10},
	} {
		assert.Equal(t, tt.want, CirullaCardValue(cirullaCard(CardDesignSpade, tt.value)), "値 %d", tt.value)
	}
	assert.Equal(t, 0, CirullaCardValue(nil))
}

// **15 は Scopa の捕獲に「足される」。置き換えではない。** #5457 は捕獲基準が
// 15 に固定されると書いているが、それだと 7 で 7 を取る当たり前の手が消える。
func TestCirulla_FifteenIsAddedToTheOrdinaryCaptures(t *testing.T) {
	table := []*Card{
		cirullaCard(CardDesignSpade, 7), // 同値 (7)
		cirullaCard(CardDesignHeart, 3), // 3 + 5 + 出した 7 = 15
		cirullaCard(CardDesignClover, 5),
	}
	// **同値の 1 枚は他の選択肢を締め出さない。** Scopa の優先規則を持ち込むと、
	// 3+5+7=15 というこの派生の看板ルールのほうが黙って消える。
	got := EnumerateCirullaCaptures(cirullaCard(CardDesignDiamond, 7), table)
	require.Len(t, got, 2, "同値の 1 枚と 15 の組の両方が出るはず")
	assert.Contains(t, got, []int{0}, "同値の 7 を取る手が無い")
	assert.Contains(t, got, []int{1, 2}, "3+5 で 15 にする手が無い")
	// 選んだほうがどちらでも通る。
	played := cirullaCard(CardDesignDiamond, 7)
	assert.True(t, IsValidCirullaCapture(played, table, []int{0}))
	assert.True(t, IsValidCirullaCapture(played, table, []int{1, 2}),
		"同値があると 15 の組が弾かれている")
	// 規則のどれにも合わない選択は通らない。
	assert.False(t, IsValidCirullaCapture(played, table, []int{1}))
	assert.False(t, IsValidCirullaCapture(played, table, []int{0, 1}))

	// 同値が無い盤面: 合計一致と 15 の両方が出る。
	table2 := []*Card{
		cirullaCard(CardDesignSpade, 3),
		cirullaCard(CardDesignHeart, 5),
		cirullaCard(CardDesignClover, 6), // 6 + 2 = 8 → 出した 8 と一致
	}
	got2 := EnumerateCirullaCaptures(cirullaCard(CardDesignDiamond, 11), table2) // J = 8
	// 合計 8 の組 (3+5) と、合計 7 の組 (無い) — 15-8=7 の組も探す。
	assert.NotEmpty(t, got2)
	found8 := false
	for _, g := range got2 {
		sum := 0
		for _, i := range g {
			sum += CirullaCardValue(table2[i])
		}
		if sum == 8 {
			found8 = true
		}
	}
	assert.True(t, found8, "合計一致の捕獲が消えている")
}

func TestCirulla_FifteenCapture(t *testing.T) {
	// 出した 6 と場の 9 (Q) で 15。合計一致 (6) の組は無い。
	table := []*Card{cirullaCard(CardDesignSpade, 12)} // Q = 9
	got := EnumerateCirullaCaptures(cirullaCard(CardDesignHeart, 6), table)
	require.Len(t, got, 1)
	assert.Equal(t, []int{0}, got[0])
	assert.True(t, IsValidCirullaCapture(cirullaCard(CardDesignHeart, 6), table, []int{0}))
}

// **アッソ・ピリアトゥット。** 場にアッソが無ければ総取り、あればただの 1。
func TestCirulla_AceTakesTheTable(t *testing.T) {
	table := []*Card{
		cirullaCard(CardDesignSpade, 5),
		cirullaCard(CardDesignHeart, 12),
		cirullaCard(CardDesignClover, 3),
	}
	ace := cirullaCard(CardDesignDiamond, 1)
	assert.True(t, CirullaAceTakesAll(ace, table))
	got := EnumerateCirullaCaptures(ace, table)
	require.Len(t, got, 1)
	assert.Equal(t, []int{0, 1, 2}, got[0], "総取りになっていない")
	// 半端に選ぶことはできない。
	assert.False(t, IsValidCirullaCapture(ace, table, []int{0}))
	assert.True(t, IsValidCirullaCapture(ace, table, []int{0, 1, 2}))

	// **場にアッソがあると総取りが消え、ただの 1 に戻る。** そこから先は
	// 他の札と同じで、同値の 1 枚と 15 の組を自由に選べる
	// (場は ♠A ♠5 ♥Q(9) ♣3 なので 5+9+出した 1 = 15)。
	withAce := append([]*Card{cirullaCard(CardDesignSpade, 1)}, table...)
	assert.False(t, CirullaAceTakesAll(ace, withAce))
	got2 := EnumerateCirullaCaptures(ace, withAce)
	assert.Contains(t, got2, []int{0}, "同値のアッソを取る手が無い")
	assert.Contains(t, got2, []int{1, 2}, "5+9 で 15 にする手が無い")
	assert.False(t, IsValidCirullaCapture(ace, withAce, []int{0, 1, 2, 3}),
		"場にアッソがあるのに総取りが通っている")
}

func TestCirulla_NoCaptureOnAnEmptyTable(t *testing.T) {
	assert.Empty(t, EnumerateCirullaCaptures(cirullaCard(CardDesignSpade, 1), nil))
	assert.False(t, CirullaAceTakesAll(cirullaCard(CardDesignSpade, 1), nil))
}

// **バルセゴンは 10 点、バルセガは 3 点。** 桁が違う。
func TestCirulla_DealBonuses(t *testing.T) {
	// 合計 9 以下 → バルセガ。
	name, pts := CirullaDealBonus([]*Card{
		cirullaCard(CardDesignSpade, 1), cirullaCard(CardDesignHeart, 3), cirullaCard(CardDesignClover, 5),
	})
	assert.Equal(t, CirullaBonusBarsega, name)
	// **定数どうしで比べない。** 両辺が同じ定数だと、定数を書き換えた瞬間に
	// 期待値も一緒に動いて何も確かめられなくなる。
	assert.Equal(t, 3, pts, "バルセガは 3 点")

	// 合計 10 → ボーナス無し。
	name, pts = CirullaDealBonus([]*Card{
		cirullaCard(CardDesignSpade, 2), cirullaCard(CardDesignHeart, 3), cirullaCard(CardDesignClover, 5),
	})
	assert.Equal(t, CirullaBonusNone, name)
	assert.Zero(t, pts)

	// 同位 3 枚 → バルセゴン。合計 9 以下でもこちらが優先。
	name, pts = CirullaDealBonus([]*Card{
		cirullaCard(CardDesignSpade, 2), cirullaCard(CardDesignHeart, 2), cirullaCard(CardDesignClover, 2),
	})
	assert.Equal(t, CirullaBonusBarsegon, name)
	assert.Equal(t, 10, pts, "バルセゴンは 10 点")
	// **桁が違うのが要点。** 同じ点にすると、同位 3 枚を狙う意味が消える。
	assert.Greater(t, CirullaBarsegonPoints, CirullaBarsegaPoints)

	// **7♥ はワイルド。** 同位 2 枚 + 7♥ でバルセゴン。
	name, pts = CirullaDealBonus([]*Card{
		cirullaCard(CardDesignSpade, 12), cirullaCard(CardDesignClover, 12), cirullaCard(CardDesignHeart, 7),
	})
	assert.Equal(t, CirullaBonusBarsegon, name, "7♥ がワイルドになっていない")
	assert.Equal(t, 10, pts)

	// 別のスートの 7 はワイルドではない。
	name, _ = CirullaDealBonus([]*Card{
		cirullaCard(CardDesignSpade, 12), cirullaCard(CardDesignClover, 12), cirullaCard(CardDesignDiamond, 7),
	})
	assert.Equal(t, CirullaBonusNone, name, "7♦ までワイルドになっている")

	// 3 枚でなければ判定しない。
	name, _ = CirullaDealBonus([]*Card{cirullaCard(CardDesignSpade, 1)})
	assert.Equal(t, CirullaBonusNone, name)
}

// **7♥ のワイルドは捕獲では効かない。** ただの 7 として扱う。
func TestCirulla_WildIsBonusOnly(t *testing.T) {
	seven := cirullaCard(CardDesignHeart, 7)
	assert.True(t, CirullaIsWildForBonus(seven))
	assert.Equal(t, 7, CirullaCardValue(seven))
	table := []*Card{cirullaCard(CardDesignSpade, 3)}
	// ワイルドなら 3 を取れてしまうが、捕獲では 7 なので取れない。
	assert.Empty(t, EnumerateCirullaCaptures(seven, table))
}

func TestCirullaPrimiera(t *testing.T) {
	// 4 スート揃わなければ 0。
	assert.Zero(t, CirullaPrimiera([]*Card{cirullaCard(CardDesignSpade, 7)}))
	full := []*Card{
		cirullaCard(CardDesignSpade, 7), cirullaCard(CardDesignHeart, 7),
		cirullaCard(CardDesignClover, 7), cirullaCard(CardDesignDiamond, 7),
	}
	assert.Equal(t, 21*4, CirullaPrimiera(full))
	// 各スートの最高 1 枚だけを数える。
	assert.Equal(t, 21*4, CirullaPrimiera(append(full, cirullaCard(CardDesignSpade, 6))))
}

// **ピッコラは A♦ から続いた枚数。** 途切れたところで止まる。
func TestCirullaPiccola(t *testing.T) {
	d := func(v int) *Card { return cirullaCard(CardDesignDiamond, v) }
	assert.Zero(t, CirullaPiccola([]*Card{d(2), d(3)}), "A♦ が無い")
	assert.Zero(t, CirullaPiccola([]*Card{d(1), d(3)}), "2♦ が無い")
	assert.Equal(t, 2, CirullaPiccola([]*Card{d(1), d(2)}))
	assert.Equal(t, 3, CirullaPiccola([]*Card{d(1), d(2), d(3)}))
	assert.Equal(t, 3, CirullaPiccola([]*Card{d(1), d(2), d(3), d(5)}), "4♦ で途切れている")
	assert.Equal(t, 7, CirullaPiccola([]*Card{d(1), d(2), d(3), d(4), d(5), d(6), d(7)}))
	// 別スートは数えない。
	assert.Zero(t, CirullaPiccola([]*Card{cirullaCard(CardDesignSpade, 1), cirullaCard(CardDesignSpade, 2)}))
}

func TestCirullaGrande(t *testing.T) {
	d := func(v int) *Card { return cirullaCard(CardDesignDiamond, v) }
	assert.True(t, CirullaHasGrande([]*Card{d(11), d(12), d(13)}))
	assert.False(t, CirullaHasGrande([]*Card{d(11), d(12)}))
	assert.False(t, CirullaHasGrande([]*Card{
		cirullaCard(CardDesignSpade, 11), cirullaCard(CardDesignSpade, 12), cirullaCard(CardDesignSpade, 13),
	}), "別スートでグランデになっている")
}

// **配りは 3 枚ずつ + 場に 4 枚。**
func TestCirulla_DealsThreeEachAndFourToTheTable(t *testing.T) {
	c := newCirullaForTest(t)
	assert.Equal(t, CirullaPhasePlay, c.GetPhase())
	for i, p := range c.GetPlayers() {
		assert.Equal(t, CirullaHandSize, p.GetCardsSize(), "席 %d の手札", i)
	}
	assert.Len(t, c.GetTable(), CirullaTableSize)
	assert.Equal(t, CirullaDeckSize-CirullaHandSize*CirullaPlayerCnt-CirullaTableSize, c.GetDeckRemaining())
	// **開幕は人間の手番。** 親を席 1 にしてあるので席 0 から打つ。
	assert.Equal(t, 0, c.GetCurrentPlayerIdx())
	assert.True(t, c.IsHumanTurn())
}

// **取れるのに置くことはできない。**
func TestCirulla_MustCaptureWhenPossible(t *testing.T) {
	c := newCirullaForTest(t)
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(cirullaCard(CardDesignSpade, 5))
	c.table = []*Card{cirullaCard(CardDesignHeart, 5)}
	assert.Error(t, c.PlayerPlay(0, nil), "取れるのに置けてしまう")
	require.NoError(t, c.PlayerPlay(0, []int{0}))
	assert.Empty(t, c.GetTable())
	assert.Len(t, c.GetPlayer(0).GetCaptured(), 2)
}

// **場を空にしたらスコパ。** ただし最後の手は数えない。
func TestCirulla_ScopaOnClearingTheTable(t *testing.T) {
	c := newCirullaForTest(t)
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(cirullaCard(CardDesignSpade, 5))
	c.GetPlayer(1).Reset()
	c.GetPlayer(1).AddCard(cirullaCard(CardDesignClover, 2))
	c.table = []*Card{cirullaCard(CardDesignHeart, 5)}
	require.NoError(t, c.PlayerPlay(0, []int{0}))
	assert.Equal(t, 1, c.GetPlayer(0).GetScope())
}

func TestCirulla_NoScopaOnTheFinalPlay(t *testing.T) {
	c := newCirullaForTest(t)
	c.drawIdx = len(c.deck) // 山を尽きさせる
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(cirullaCard(CardDesignSpade, 5))
	c.GetPlayer(1).Reset()
	c.table = []*Card{cirullaCard(CardDesignHeart, 5)}
	require.NoError(t, c.PlayerPlay(0, []int{0}))
	assert.Zero(t, c.GetPlayer(0).GetScope(), "最後の手でスコパが付いている")
}

// **場に残った札は最後に取った人へ。** 拾わないと 40 枚の勘定が合わない。
func TestCirulla_LeftoverTableGoesToTheLastCapturer(t *testing.T) {
	c := newCirullaForTest(t)
	cirullaPlayRound(t, c)
	// **フェーズを RoundEnd に決め打たない。** 1 ラウンドで目標点に届く配りがあり、
	// そのとき GetPhase は GameEnd になる。実測で 3000 配り中 20 回 (0.67%)。
	// 場の札が最後に取った側へ行くことは、試合が続こうと終わろうと同じ。
	require.Contains(t, []string{CirullaPhaseRoundEnd, CirullaPhaseGameEnd}, c.GetPhase())
	assert.Empty(t, c.GetTable(), "場に札が残っている")
	total := 0
	for _, p := range c.GetPlayers() {
		total += len(p.GetCaptured())
	}
	assert.Equal(t, CirullaDeckSize, total, "40 枚の勘定が合わない")
}

// **最多枚数・最多デナリは同数なら誰にも入らない。**
func TestCirullaAwardMaxIsEmptyOnATie(t *testing.T) {
	assert.Equal(t, [CirullaPlayerCnt]int{1, 0}, cirullaAwardMax([CirullaPlayerCnt]int{21, 19}))
	assert.Equal(t, [CirullaPlayerCnt]int{0, 1}, cirullaAwardMax([CirullaPlayerCnt]int{19, 21}))
	assert.Equal(t, [CirullaPlayerCnt]int{0, 0}, cirullaAwardMax([CirullaPlayerCnt]int{20, 20}))
}

// ラウンドの集計には 8 つの項目が並ぶ。
func TestCirulla_RoundResultHasEveryCategory(t *testing.T) {
	c := newCirullaForTest(t)
	cirullaPlayRound(t, c)
	res := c.GetLastResult()
	require.NotNil(t, res)
	keys := make([]string, 0, len(res.Lines))
	for _, l := range res.Lines {
		keys = append(keys, l.Key)
	}
	assert.ElementsMatch(t,
		[]string{"cards", "denari", "settebello", "primiera", "piccola", "grande", "scope", "bonus"},
		keys)
}

// **全デナリを取ると即勝ち。** 点が足りていなくても決まる。
func TestCirulla_SweepingTheDenariWinsOutright(t *testing.T) {
	c := newCirullaForTest(t)
	c.GetPlayer(0).ResetRound()
	c.GetPlayer(1).ResetRound()
	all := make([]*Card, 0, 10)
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 11, 12, 13} {
		all = append(all, cirullaCard(CardDesignDiamond, v))
	}
	c.GetPlayer(0).AddCaptured(all)
	c.table = make([]*Card, 0)
	c.finishRound()
	assert.True(t, c.GetGameEndFlag(), "全デナリでも終わらない")
	assert.Equal(t, 0, c.GetWinnerIdx())
	assert.Equal(t, 0, c.GetLastResult().SweptDenari)
}

// **1 マッチを通しで打てる。**
func TestCirulla_PlaysAMatchThrough(t *testing.T) {
	c := newCirullaForTest(t)
	for round := 0; round < 200 && !c.GetGameEndFlag(); round++ {
		cirullaPlayRound(t, c)
		c.NextRound()
	}
	require.True(t, c.GetGameEndFlag(), "マッチが終わらない")
	assert.Equal(t, CirullaPhaseGameEnd, c.GetPhase())
	assert.GreaterOrEqual(t, c.GetWinnerIdx(), 0)
}

// **同点では終わらない。** 目標点に届いても並んでいれば続く。
func TestCirulla_ATieDoesNotEndTheMatch(t *testing.T) {
	c := newCirullaForTest(t)
	c.GetPlayer(0).ResetScore()
	c.GetPlayer(1).ResetScore()
	c.GetPlayer(0).AddScore(c.GetConfig().TargetScore)
	c.GetPlayer(1).AddScore(c.GetConfig().TargetScore)
	c.checkGameEnd(&CirullaRoundResult{SweptDenari: -1})
	assert.False(t, c.GetGameEndFlag(), "同点で終わっている")
}

func TestCirulla_HintFollowsThePhase(t *testing.T) {
	c := newCirullaForTest(t)
	hint := c.GetHint()
	require.NotNil(t, hint)
	assert.Contains(t, []string{"capture", "sweep", "lay_off"}, hint.Reason)
	assert.GreaterOrEqual(t, hint.HandIdx, 0)

	cirullaPlayRound(t, c)
	if !c.GetGameEndFlag() {
		assert.Equal(t, "next_round", c.GetHint().Reason)
	}
}

// **助言は CPU の難易度に引きずられない。**
func TestCirulla_HintIgnoresCpuDifficulty(t *testing.T) {
	players := []*CirullaPlayer{NewCirullaPlayer(true), NewCirullaPlayer(false)}
	c := NewCirulla(players, CirullaConfig{
		CpuDifficulty: CirullaCpuDifficultyEasy,
		TargetScore:   CirullaDefaultTarget,
	})
	c.Reset()
	require.True(t, c.IsHumanTurn())
	// **選択肢が 2 つ以上ある局面に固定する。**
	c.GetPlayer(0).Reset()
	c.GetPlayer(0).AddCard(cirullaCard(CardDesignSpade, 5))
	c.GetPlayer(0).AddCard(cirullaCard(CardDesignHeart, 3))
	c.GetPlayer(0).AddCard(cirullaCard(CardDesignClover, 7))
	c.table = []*Card{cirullaCard(CardDesignDiamond, 5), cirullaCard(CardDesignSpade, 3)}
	require.Greater(t, len(c.enumerateChoices(0)), 1, "候補が 1 つしかない")

	want := c.GetHint()
	for i := 0; i < 20; i++ {
		got := c.GetHint()
		assert.Equal(t, want.HandIdx, got.HandIdx, "%d 回目で札がぶれた", i+1)
		assert.Equal(t, want.CaptureIdxs, got.CaptureIdxs, "%d 回目で捕獲がぶれた", i+1)
	}
}

func TestCirulla_RejectsBadInput(t *testing.T) {
	c := newCirullaForTest(t)
	assert.Error(t, c.PlayerPlay(99, nil))
	assert.Error(t, c.PlayerPlay(-1, nil))
	// 存在しない場札は選べない。
	assert.Error(t, c.PlayerPlay(0, []int{99}))
}

// **保存した盤で打ち続けられる。**
func TestCirulla_SaveRestoreKeepsPlaying(t *testing.T) {
	c := newCirullaForTest(t)
	for i := 0; i < 3 && c.GetPhase() == CirullaPhasePlay; i++ {
		if c.IsHumanTurn() {
			handIdx, capture := c.smartChoose(0)
			require.NoError(t, c.PlayerPlay(handIdx, capture))
			continue
		}
		c.CpuPlay()
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored := new(Cirulla)
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, c.GetDeckRemaining(), restored.GetDeckRemaining())
	assert.Equal(t, c.GetLastCapturer(), restored.GetLastCapturer(), "最後の捕獲者が消えている")
	assert.Len(t, restored.GetTable(), len(c.GetTable()))
	for i := range c.GetPlayers() {
		assert.Equal(t, c.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
		assert.Len(t, restored.GetPlayer(i).GetCaptured(), len(c.GetPlayer(i).GetCaptured()), "席 %d の取り札", i)
	}
	for round := 0; round < 200 && !restored.GetGameEndFlag(); round++ {
		cirullaPlayRound(t, restored)
		restored.NextRound()
	}
	assert.True(t, restored.GetGameEndFlag())
}

func TestCirulla_RejectsTamperedSnapshot(t *testing.T) {
	restored := new(Cirulla)
	assert.Error(t, restored.UnmarshalJSON([]byte("{")))
	assert.Error(t, restored.UnmarshalJSON([]byte(`{"pl":[]}`)))
}

func TestCirullaConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultCirullaConfig().Validate())
	assert.Error(t, CirullaConfig{CpuDifficulty: -1, TargetScore: 51}.Validate())
	assert.Error(t, CirullaConfig{CpuDifficulty: 1, TargetScore: 1}.Validate())
	assert.Error(t, CirullaConfig{CpuDifficulty: 1, TargetScore: 99}.Validate())
}

// cirullaPlayRound は現在のラウンドを最後まで打つ。
func cirullaPlayRound(t *testing.T, c *Cirulla) {
	t.Helper()
	for step := 0; step < 500 && c.GetPhase() == CirullaPhasePlay; step++ {
		if c.IsHumanTurn() {
			handIdx, capture := c.smartChoose(0)
			require.NoError(t, c.PlayerPlay(handIdx, capture))
			continue
		}
		c.CpuPlay()
	}
	require.NotEqual(t, CirullaPhasePlay, c.GetPhase(), "ラウンドが終わらない")
}
