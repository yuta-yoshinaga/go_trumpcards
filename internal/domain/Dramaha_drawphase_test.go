//go:build test && (!js || !wasm || casino)

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dramahaChipTotal は卓上の総額 (ポット + 全員のチップ)。
func dramahaChipTotal(g *Dramaha) int {
	n := g.pot
	for _, p := range g.players {
		n += p.GetChips()
	}
	return n
}

// TestDramaha_TheDrawRoundIsNotSkippedByTheAllInShortCircuit pins the early
// return at the end of advancePhase's Flop case.
//
// **落とすと人間が引かないままショーダウンに行く。** autoDrawForCPUs は最後の
// CPU が引いた時点で Draw() 経由で advancePhase を入れ子に呼ぶ。return しないと
// 外側のフレームが switch を抜けて `activeCnt <= 1` の短絡に落ち、残りのボードを
// 配って決着させてしまう —— CPU は引き終えているのに、人間だけ 5 枚を引けない。
//
// 卓上の総額では捕まらないことを実測した: この経路では finalizeShowdown が先に
// ポットを 0 にしているので、二度目の resolveShowdown は空のポットを配って
// 総額を変えない。**見るのは金額ではなく「人間の手番が残っているか」。**
func TestDramaha_TheDrawRoundIsNotSkippedByTheAllInShortCircuit(t *testing.T) {
	g := NewDefaultDramaha()
	require.NoError(t, g.Reset())

	// 席 0 (人間) を含む 3 席をオールインにして、活動席を席 3 の 1 つに絞る。
	// オールインは**実際にチップをポットへ移して**作る。SetAllIn だけ立てると
	// サイドポットの持ち分が嘘になる。
	for _, i := range []int{0, 1, 2} {
		stack := g.players[i].GetChips()
		g.players[i].SetChips(0)
		g.players[i].SetCurrentBet(stack)
		g.players[i].SetAllIn(true)
		g.pot += stack
	}
	activeCnt := 0
	for _, p := range g.players {
		if !p.GetFolded() && !p.GetAllIn() {
			activeCnt++
		}
	}
	require.LessOrEqual(t, activeCnt, 1,
		"この短絡に落ちる盤でしか return の有無が効かない (activeCnt=%d)", activeCnt)

	// **フロップから入る。** プリフロップ時点で activeCnt <= 1 だと、
	// PreFlop の case が先に短絡してボードを走り切ってしまい、Flop の case に
	// 一度も入らない。この return が効くのは「フロップのベッティングで
	// 活動席が 1 つに落ちた」ときだけなので、その盤を直接作る。
	g.phase = DramahaPhaseFlop
	g.communityCards = dramahaTestFlop()

	before := dramahaChipTotal(g)

	g.advancePhase() // Flop -> Draw (CPU が引き切り、入れ子で先へ)

	assert.Equal(t, DramahaPhaseDraw, g.GetPhase(),
		"人間がまだ引いていないのだからドローで待つ (短絡に落ちるとショーダウンへ飛ぶ)")
	assert.False(t, g.drawnFlags[0], "人間の交換はこれから")
	for i := 1; i < len(g.drawnFlags); i++ {
		assert.True(t, g.drawnFlags[i], "CPU %d は自動で引き終えている", i)
	}
	assert.Equal(t, before, dramahaChipTotal(g),
		"ドローを待っているあいだにポットは動かない")
}

// TestDramaha_TheDrawRoundWaitsForTheHuman は、人間が残っているあいだは
// ドローフェーズで止まることを固定する。
func TestDramaha_TheDrawRoundWaitsForTheHuman(t *testing.T) {
	g := NewDefaultDramaha()
	require.NoError(t, g.Reset())

	g.advancePhase() // -> Flop
	g.advancePhase() // -> Draw

	assert.Equal(t, DramahaPhaseDraw, g.GetPhase(), "人間のドロー待ちで止まる")

	drawn := 0
	for _, done := range g.drawnFlags {
		if done {
			drawn++
		}
	}
	assert.Equal(t, len(g.players)-1, drawn, "CPU は自動で引き終えている")
	assert.False(t, g.drawnFlags[0], "人間だけ残っている")

	board := len(g.GetCommunityCards())
	require.NoError(t, g.Draw(0, []int{0, 1}))
	assert.Equal(t, DramahaPhaseTurn, g.GetPhase(), "人間が引いたらターンへ進む")
	assert.Equal(t, board+1, len(g.GetCommunityCards()), "ターンが 1 枚配られる")
}

// TestDramaha_DrawnFlagsSurviveTheWire pins the KV round trip of the draw state.
//
// **落とすと panic する。** ドロー中に保存した卓を戻すと drawnFlags が空に
// なり、Draw() が添字で落ちる。Worker はリクエストごとに卓を戻すので、これは
// 「起きうる」ではなく「必ず通る」経路。
func TestDramaha_DrawnFlagsSurviveTheWire(t *testing.T) {
	g := NewDefaultDramaha()
	require.NoError(t, g.Reset())
	g.advancePhase() // -> Flop
	g.advancePhase() // -> Draw (CPU は引き終え、人間だけ残る)
	require.Equal(t, DramahaPhaseDraw, g.GetPhase())
	require.False(t, g.drawnFlags[0], "人間はまだ引いていない")

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var got Dramaha
	require.NoError(t, json.Unmarshal(data, &got))
	require.Len(t, got.drawnFlags, len(got.players), "席数ぶん揃っていること")
	assert.False(t, got.drawnFlags[0], "人間の未ドローが往復で保たれる")
	// **引き終えた席が「まだ」に戻らないこと。** ここが本命の検査 ——
	// Unmarshal は長さが合わないとフラグを組み直すので、「人間が未ドロー」だけ
	// を見ていると、フィールドを丸ごと落としても組み直しが同じ答えを出して
	// 通ってしまう。CPU の「済」は組み直しでは復元できない。
	for i := 1; i < len(got.players); i++ {
		assert.True(t, got.drawnFlags[i],
			"座席 %d は保存時点で引き終えている。往復で「未ドロー」に戻ると二度引ける", i)
	}

	// 復元した卓でそのまま引ける (落ちない)。
	require.NotPanics(t, func() { _ = got.Draw(0, []int{0}) })
}

// TestDramaha_TheHandKeepsMovingAfterTheHumanDraws pins that the CPUs are driven.
//
// Draw は advancePhase までしかやっていなかった。ターン直後の手番は CPU なので、
// 誰も動かさないまま人間の入力を待ち「あなたの番ではありません」から抜けられ
// なくなる —— ハンドがそこで死ぬ。
func TestDramaha_TheHandKeepsMovingAfterTheHumanDraws(t *testing.T) {
	g := NewDefaultDramaha()
	require.NoError(t, g.Reset())
	g.advancePhase() // -> Flop
	g.advancePhase() // -> Draw
	require.Equal(t, DramahaPhaseDraw, g.GetPhase())

	require.NoError(t, g.Draw(0, nil))

	// ターン以降へ進み、かつ手番が人間に戻っている (CPU が動いた証拠)。
	assert.NotEqual(t, DramahaPhaseDraw, g.GetPhase(), "ドローフェーズを抜ける")
	if g.GetPhase() < DramahaPhaseShowdown {
		assert.True(t, g.players[g.currentTurn].GetIsHuman(),
			"CPU が動き切って人間に手番が戻ること (止まると CPU のまま固まる)")
	}
}
