//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **打ちかけのハンドごと復元する。** 正本だけを保存すると、復元のたびに
// 進行中のハンドが消えて配り直しになる。
func TestHorseRoundTripJSON_MidHand(t *testing.T) {
	g := newHorseForTest(t)
	require.Equal(t, HorsePhaseHand, g.GetPhase())
	beforeChips := horseTotalChips(g)
	beforeTurn := g.GetCurrentTurn()

	b, err := json.Marshal(g)
	require.NoError(t, err)
	var back Horse
	require.NoError(t, json.Unmarshal(b, &back))

	assert.Equal(t, g.GetPhase(), back.GetPhase())
	assert.Equal(t, g.GetDiscipline(), back.GetDiscipline())
	assert.Equal(t, g.GetHandNumber(), back.GetHandNumber())
	assert.Equal(t, g.GetHandInDiscipline(), back.GetHandInDiscipline())
	assert.Equal(t, beforeChips, horseTotalChips(&back), "チップの総量が変わっている")
	assert.Equal(t, beforeTurn, back.GetCurrentTurn(), "手番が変わっている")
	// **復元した卓で打ち続けられ、結果が正本に戻る。**
	//
	// フェーズが進むことだけを見ても足りない ── 回収の配線を繋ぎ直さなくても
	// `settleIfHandOver` は nil を素通りして HandEnd にするので、**チップが
	// 動いたか**まで見る (1 ハンド打てばブラインド/アンティは必ず動く)。
	// **打つ直前の残高と比べる。** 以前は InitialChips と比べており、
	// 打った結果たまたま全席が初期値へ戻る配りだと「1 席も動いていない」と
	// 誤検知した (#5812 で 2 回観測、どちらも Horse と無関係な PR)。
	// 見たいのは「このハンドでチップが動いたか」なので、基準は初期値ではなく
	// 打つ直前の残高。
	beforeFold := make([]int, back.GetSeatCount())
	for i := range beforeFold {
		beforeFold[i] = back.GetSeatChips(i)
	}

	// **1 ハンドでは足りない。**降りたぶんの掛け金がそのまま自分に戻る配りだと、
	// 全席が打つ前と同じ残高で終わる。それを「回収が繋がっていない」と読むと
	// 配りに依存して落ちる (#5869 で 3 回観測、いずれも Horse と無関係な PR)。
	// 回収が本当に切れていれば**何ハンド打っても**動かないので、動くまで数ハンド
	// 続ける形にする。
	moved := false
	for hand := 0; hand < 3 && !moved; hand++ {
		if hand > 0 {
			require.NoError(t, back.NextHand())
		}
		horseFoldOutHand(t, &back)
		require.Equal(t, HorsePhaseHandEnd, back.GetPhase())
		assert.Equal(t, beforeChips, horseTotalChips(&back), "総量が変わっている")
		for i := range beforeFold {
			if back.GetSeatChips(i) != beforeFold[i] {
				moved = true
				break
			}
		}
	}
	assert.True(t, moved, "3 ハンド打っても残高が 1 席も動いていない (回収が繋がっていない)")
}

func TestHorseRoundTripJSON_BetweenHands(t *testing.T) {
	g := newHorseForTest(t)
	horseFoldOutHand(t, g)
	require.Equal(t, HorsePhaseHandEnd, g.GetPhase())

	b, err := json.Marshal(g)
	require.NoError(t, err)
	var back Horse
	require.NoError(t, json.Unmarshal(b, &back))

	assert.Equal(t, HorsePhaseHandEnd, back.GetPhase())
	assert.Equal(t, horseTotalChips(g), horseTotalChips(&back))
	require.NoError(t, back.NextHand())
	assert.Equal(t, HorsePhaseHand, back.GetPhase())
}

// 改竄した保存データは、本物の局面を 1 か所だけ壊して作る。
func horseTamper(t *testing.T, mutate func(m map[string]any)) error {
	t.Helper()
	g := newHorseForTest(t)
	b, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	var back Horse
	return json.Unmarshal(tampered, &back)
}

func TestHorseUnmarshal_RejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"フェーズが範囲外", func(m map[string]any) { m["ph"] = 9 }, "invalid phase"},
		{"種目が範囲外", func(m map[string]any) { m["dc"] = 9 }, "invalid discipline"},
		{"種目が負", func(m map[string]any) { m["dc"] = -1 }, "invalid discipline"},
		{"ハンド番号が 0", func(m map[string]any) { m["hn"] = 0 }, "hand number"},
		{"種目内ハンドが上限超え", func(m map[string]any) {
			m["hd"] = HorseDefaultHandsPerDiscipline + 5
		}, "hand-in-discipline"},
		{"席数が設定と食い違う", func(m map[string]any) {
			m["st"] = m["st"].([]any)[:2]
		}, "does not match config"},
		{"席のチップが負", func(m map[string]any) {
			st := m["st"].([]any)
			st[0].(map[string]any)["c"] = -1
			m["st"] = st
		}, "must not be negative"},
		{"席マップが範囲外", func(m map[string]any) { m["sm"] = []any{99} }, "seat map"},
		{"設定が不正 (席数)", func(m map[string]any) {
			cf := m["cf"].(map[string]any)
			cf["s"] = 3 // 種目が受け付けない卓サイズ
			m["cf"] = cf
		}, "seats out of range"},
		{"卓の JSON が壊れている", func(m map[string]any) { m["tb"] = map[string]any{"ph": "x"} }, "restore table"},
		// **卓の人数が席数と食い違う保存は受け取らない。** 種目側の
		// UnmarshalJSON はプレイヤー列を差し替えるだけで人数を検めないので、
		// 通すと次の 1 手で範囲外を触って落ちる。
		{"卓の人数が席数より少ない", func(m map[string]any) {
			tb := m["tb"].(map[string]any)
			tb["pl"] = tb["pl"].([]any)[:2]
			m["tb"] = tb
		}, "seats 2 players but 4 seats are funded"},
		// ハンド中なのに卓が無い保存は、落ちない代わりに二度と進まない。
		{"ハンド中なのに卓が無い", func(m map[string]any) { delete(m, "tb") }, "no table was saved"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := horseTamper(t, tt.mutate)
			require.Error(t, err, "改竄が素通りしている")
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestHorseUnmarshal_RejectsOversizedArrays(t *testing.T) {
	big := make([]map[string]int, horseMaxSliceLen+1)
	for i := range big {
		big[i] = map[string]int{"c": 1}
	}
	payload, err := json.Marshal(map[string]any{"st": big})
	require.NoError(t, err)
	var g Horse
	err = json.Unmarshal(payload, &g)
	require.Error(t, err)
	assert.ErrorContains(t, err, "maximum allowed size")
}

func TestHorseUnmarshal_RejectsMalformedJSON(t *testing.T) {
	var g Horse
	assert.Error(t, json.Unmarshal([]byte(`{"ph":"x"}`), &g))
}
