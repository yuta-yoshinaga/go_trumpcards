//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newHorseForTest(t *testing.T) *Horse {
	t.Helper()
	g := NewDefaultHorse()
	g.Reset()
	return g
}

// horseTotalChips は卓のチップ総量を返す。
func horseTotalChips(g *Horse) int {
	total := 0
	for i := range g.GetSeatCount() {
		total += g.GetSeatChips(i)
	}
	return total
}

// horseFoldOutHand は人間が降りてハンドを閉じる。
//
// **降りるのが唯一いつでも合法な手。** 種目ごとにベット額の刻みもレイズ上限も
// 違うので、テストから合法手を組み立てようとすると 5 種目ぶんの規則を書き直す
// ことになる。
func horseFoldOutHand(t *testing.T, g *Horse) {
	t.Helper()
	for steps := 0; g.GetPhase() == HorsePhaseHand; steps++ {
		require.Less(t, steps, 50, "ハンドが終わらない")
		if !g.IsHumanTurn() {
			// CPU の手番のまま止まっているなら、種目側が進めていない。
			require.FailNow(t, "人間の手番が来ないままハンドが止まっている")
		}
		require.NoError(t, g.PlayerAction(HoldemActionFold, 0, 0))
	}
}

// --- 種目のローテーション ---

// **並びが頭文字そのもの。** 入れ替わるとゲーム名と実際の順序が食い違う。
func TestHorse_DisciplineOrderSpellsTheName(t *testing.T) {
	t.Parallel()
	letters := ""
	for d := range HorseDiscipline(HorseDisciplineCount) {
		letters += HorseDisciplineLetter(d)
	}
	assert.Equal(t, "HORSE", letters)
}

func TestHorse_DisciplineNames(t *testing.T) {
	t.Parallel()
	for d, want := range map[HorseDiscipline]string{
		HorseHoldem:         "holdem",
		HorseOmahaHiLo:      "omahaHiLo",
		HorseRazz:           "razz",
		HorseStud:           "stud",
		HorseStudHiLo:       "studHiLo",
		HorseDiscipline(99): "unknown",
	} {
		assert.Equal(t, want, HorseDisciplineName(d))
	}
	assert.Equal(t, "?", HorseDisciplineLetter(HorseDiscipline(99)))
}

// **設定したハンド数ごとに切り替わり、E の次は H に戻る。**
func TestHorse_RotatesEveryNHands(t *testing.T) {
	t.Parallel()
	g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 2})
	g.Reset()

	seen := make([]HorseDiscipline, 0, 12)
	for hand := 0; hand < 12 && !g.GetGameEndFlag(); hand++ {
		seen = append(seen, g.GetDiscipline())
		horseFoldOutHand(t, g)
		if g.GetGameEndFlag() {
			break
		}
		require.NoError(t, g.NextHand())
	}

	require.GreaterOrEqual(t, len(seen), 11, "12 ハンド回らなかった")
	// 2 ハンドずつ H,H,O,O,R,R,S,S,E,E,H,H
	want := []HorseDiscipline{
		HorseHoldem, HorseHoldem, HorseOmahaHiLo, HorseOmahaHiLo,
		HorseRazz, HorseRazz, HorseStud, HorseStud,
		HorseStudHiLo, HorseStudHiLo, HorseHoldem,
	}
	assert.Equal(t, want, seen[:len(want)], "種目のローテーションが H-O-R-S-E で回っていない")
}

// **1 ハンドごとの設定でも回る。**
func TestHorse_RotatesEveryHandWhenConfiguredSo(t *testing.T) {
	t.Parallel()
	g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
	g.Reset()

	seen := make([]HorseDiscipline, 0, 6)
	for hand := 0; hand < 6 && !g.GetGameEndFlag(); hand++ {
		seen = append(seen, g.GetDiscipline())
		horseFoldOutHand(t, g)
		if g.GetGameEndFlag() {
			break
		}
		require.NoError(t, g.NextHand())
	}
	require.GreaterOrEqual(t, len(seen), 6)
	assert.Equal(t, []HorseDiscipline{
		HorseHoldem, HorseOmahaHiLo, HorseRazz, HorseStud, HorseStudHiLo, HorseHoldem,
	}, seen[:6])
}

// **ハンドの途中で種目は変わらない。**
func TestHorse_DisciplineIsStableWithinAHand(t *testing.T) {
	t.Parallel()
	g := newHorseForTest(t)
	before := g.GetDiscipline()
	require.NoError(t, g.PlayerAction(HoldemActionFold, 0, 0))
	assert.Equal(t, before, g.GetDiscipline())
}

// --- チップの持ち回し ---

// **これがこのゲームの唯一の難所。** 種目ごとにプレイヤー型が違うので、
// ハンドの開始時に正本を配り、終了時に回収する。どこかで落とすと、種目が
// 変わった瞬間に全員の残高が初期値に戻る。
//
// **スタッド系で確かめる。** Holdem/Omaha は End の時点でポットが未配分なので
// 持ち回りが安定しない (`TestHorse_HoldemFamilyPotIsUndistributedAtEnd` 参照)。
func TestHorse_ChipsCarryAcrossDisciplines(t *testing.T) {
	t.Parallel()
	g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
	g.Reset()
	// R まで進める。
	for g.GetDiscipline() != HorseRazz && !g.GetGameEndFlag() {
		horseFoldOutHand(t, g)
		require.NoError(t, g.NextHand())
	}
	require.False(t, g.GetGameEndFlag())

	horseFoldOutHand(t, g)
	afterRazz := make([]int, g.GetSeatCount())
	for i := range afterRazz {
		afterRazz[i] = g.GetSeatChips(i)
	}
	require.NoError(t, g.NextHand())

	// 種目が変わった直後も、残高は初期値ではなく前ハンドの結果。
	require.Equal(t, HorseStud, g.GetDiscipline())
	for i := range afterRazz {
		assert.Equal(t, afterRazz[i], g.GetSeatChips(i),
			"席 %d の残高が種目の切り替えで戻っている", i)
	}
	// **必ず誰かは動いている。** アンティは毎ハンド取られる。
	moved := false
	for i := range g.GetSeatCount() {
		if g.GetSeatChips(i) != HorseDefaultChips {
			moved = true
			break
		}
	}
	assert.True(t, moved, "何ハンドも打ったのに全席が初期値のまま — 回収されていない")
}

// **飛んだ席は種目に座らせない。** 各エンジンの `Reset` は残高 0 の席を
// `InitChips` まで黙って積み直すので、座らせるとチップが湧く。
// **席が 1 つでも飛んだらマッチを終える。**
//
// 種目側の卓は 4 / 6 / 9 人のいずれかしか作れないので、欠けた人数で座らせることが
// できない (3 人を渡すと黙って 4 人に落とされ、こちらが残高を配ったプレイヤーとは
// 別の 4 人が打つ)。積み直しを避けるには飛んだ席を座らせないしかなく、席を欠いた
// 卓が作れない以上、そこで区切るのが誤魔化しの無い扱いになる。
func TestHorse_ABustEndsTheMatch(t *testing.T) {
	t.Parallel()
	g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
	g.Reset()
	horseFoldOutHand(t, g)
	g.SetSeatChips(2, 0)
	require.NoError(t, g.NextHand())

	assert.True(t, g.GetGameEndFlag(), "席が飛んだのにマッチが続いている")
	assert.Equal(t, HorsePhaseGameEnd, g.GetPhase())
	assert.Zero(t, g.GetSeatChips(2), "飛んだ席が積み直されている")
}

// **席番号は正本に直して返す。**
//
// 卓は常に全席ぶんなので番号は 1 対 1 だが、変換を通さずに種目側の番号を出すと、
// 将来席の並びが変わったときに黙って別の席を指す。
func TestHorse_TurnIsReportedInCanonicalSeats(t *testing.T) {
	t.Parallel()
	g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
	g.Reset()

	require.Equal(t, []int{0, 1, 2, 3}, g.seatMap)
	for i := range 4 {
		assert.Equal(t, i, g.toCanonicalSeat(i))
	}
	assert.Equal(t, -1, g.toCanonicalSeat(99), "範囲外が -1 で返らない")
	assert.Contains(t, []int{0, 1, 2, 3}, g.GetCurrentTurn())
}

// スタッド系だけを見ていた古い検査は
// `TestHorse_EveryDisciplineConservesChips` に統合した。**歩き回る途中で席が
// 飛ぶと NextHand がマッチを終える**ので、終局の確認を挟まないループは 6 回に
// 1 回ほど落ちていた (実測)。統合先は毎ステップで終局を見る。

func TestHorse_EveryDisciplineConservesChips(t *testing.T) {
	t.Parallel()
	seen := map[HorseDiscipline]bool{}
	for range 30 {
		g := NewHorse(HorseConfig{Seats: 4, InitialChips: HorseDefaultChips, HandsPerDiscipline: 1})
		g.Reset()
		for range HorseDisciplineCount {
			if g.GetGameEndFlag() {
				break
			}
			d := g.GetDiscipline()
			before := horseTotalChips(g)
			horseFoldOutHand(t, g)
			assert.Equal(t, before, horseTotalChips(g),
				"%s のハンドで総量が変わった", HorseDisciplineName(d))
			seen[d] = true
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextHand())
		}
	}
	// **5 種目すべてを通ったことを確かめる。** 1 種目でも通らないと、
	// 「保存している」の根拠がその種目には無い。
	assert.Len(t, seen, HorseDisciplineCount, "通っていない種目がある")
}

// --- 進行 ---

func TestHorse_Reset(t *testing.T) {
	t.Parallel()
	g := newHorseForTest(t)
	assert.Equal(t, HorsePhaseHand, g.GetPhase())
	assert.Equal(t, HorseHoldem, g.GetDiscipline(), "1 ハンド目は H から始まる")
	assert.Equal(t, "H", g.GetDisciplineLetter())
	assert.Equal(t, 1, g.GetHandNumber())
	assert.Equal(t, 1, g.GetHandInDiscipline())
	assert.Equal(t, HorseDefaultSeats, g.GetSeatCount())
	assert.Zero(t, g.GetHumanSeat())
	assert.True(t, g.GetSeatIsHuman(0))
	assert.False(t, g.GetSeatIsHuman(1))
	assert.Equal(t, "YOU", g.GetSeatName(0))
	assert.False(t, g.GetGameEndFlag())
	assert.NotEmpty(t, g.GetActionLog())
}

func TestHorse_PhaseGuards(t *testing.T) {
	t.Parallel()
	g := newHorseForTest(t)
	assert.ErrorIs(t, g.NextHand(), errHorseWrongPhase, "ハンド中に次へ進めてしまう")

	horseFoldOutHand(t, g)
	assert.Equal(t, HorsePhaseHandEnd, g.GetPhase())
	assert.ErrorIs(t, g.PlayerAction(HoldemActionFold, 0, 0), errHorseWrongPhase,
		"決着後に手を受け付けてしまう")

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerAction(HoldemActionFold, 0, 0), errHorseFinished)
	assert.ErrorIs(t, g.NextHand(), errHorseFinished)
}

// **規則の判定はオーケストレータでは行わない。** 種目が拒んだエラーがそのまま返る。
func TestHorse_PassesTheDisciplinesErrorThrough(t *testing.T) {
	t.Parallel()
	g := newHorseForTest(t)
	// 不正なアクション値は種目側が弾く。
	err := g.PlayerAction(9999, 0, 0)
	assert.Error(t, err, "種目が拒むはずの手が通っている")
	assert.NotErrorIs(t, err, errHorseWrongPhase, "オーケストレータが独自に弾いている")
}

func TestHorse_Accessors(t *testing.T) {
	t.Parallel()
	g := newHorseForTest(t)
	assert.GreaterOrEqual(t, g.GetCurrentTurn(), 0)
	assert.GreaterOrEqual(t, g.GetPot(), 0)
	assert.GreaterOrEqual(t, g.GetTablePhase(), 0)
	assert.Equal(t, HorseDefaultChips, g.GetSeatChips(0))

	// 範囲外は静かに既定値。
	assert.Zero(t, g.GetSeatChips(-1))
	assert.Zero(t, g.GetSeatChips(999))
	assert.Equal(t, "?", g.GetSeatName(999))
	assert.False(t, g.GetSeatIsHuman(999))
	g.SetSeatChips(999, 1) // 何も起きない
	g.SetSeatChips(1, 42)
	assert.Equal(t, 42, g.GetSeatChips(1))

	g.SetConfig(HorseConfig{Seats: 4, InitialChips: 500, HandsPerDiscipline: 1})
	assert.Equal(t, 500, g.GetConfig().InitialChips)
}

func TestHorse_WinnerSeat(t *testing.T) {
	t.Parallel()
	g := newHorseForTest(t)
	g.SetSeatChips(2, 99999)
	assert.Equal(t, 2, g.WinnerSeat())

	for i := range g.GetSeatCount() {
		g.SetSeatChips(i, 100)
	}
	assert.Zero(t, g.WinnerSeat(), "同点なら若い席")
}

// **1 人を残して全員が飛んだら終わる。**
func TestHorse_EndsWhenOnlyOneSeatHasChips(t *testing.T) {
	t.Parallel()
	g := newHorseForTest(t)
	horseFoldOutHand(t, g)
	for i := 1; i < g.GetSeatCount(); i++ {
		g.SetSeatChips(i, 0)
	}
	require.NoError(t, g.NextHand())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, HorsePhaseGameEnd, g.GetPhase())
	assert.Zero(t, g.WinnerSeat())
}

func TestHorse_ConfigValidate(t *testing.T) {
	t.Parallel()
	assert.NoError(t, DefaultHorseConfig().Validate())
	assert.ErrorIs(t, HorseConfig{Seats: 1, InitialChips: 1000, HandsPerDiscipline: 2}.Validate(), errHorseSeatsRange)
	assert.ErrorIs(t, HorseConfig{Seats: 7, InitialChips: 1000, HandsPerDiscipline: 2}.Validate(), errHorseSeatsRange)
	assert.ErrorIs(t, HorseConfig{Seats: 6, InitialChips: 1, HandsPerDiscipline: 2}.Validate(), errHorseChipsRange)
	assert.ErrorIs(t, HorseConfig{Seats: 6, InitialChips: 1000, HandsPerDiscipline: 0}.Validate(), errHorseHandsRange)
	assert.ErrorIs(t, HorseConfig{Seats: 6, InitialChips: 1000, HandsPerDiscipline: 99}.Validate(), errHorseHandsRange)
}
