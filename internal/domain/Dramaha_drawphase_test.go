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

// TestDramaha_TheDrawRoundDoesNotSettleTwice pins the re-entrancy fix.
//
// **ドローの完了は advancePhase を入れ子に呼ぶ。** フロップの case が
// autoDrawForCPUs を呼び、最後の CPU が引いた時点で Draw() が advancePhase を
// もう一度走らせる。内側のフレームがターンを配って activeCnt ブロックまで
// 実行し終えた後、外側のフレームが switch を抜けて**同じブロックを再実行**
// すると resolveShowdown が二度走り、ポットが二重に配られる。
//
// 実測: 卓上の総額が 4000 → 4015 に増えた (本来 +5 は Omaha 由来の既存挙動で、
// 差分の +10 がこのバグぶん)。
func TestDramaha_TheDrawRoundDoesNotSettleTwice(t *testing.T) {
	g := NewDefaultDramaha()
	require.NoError(t, g.Reset())

	// **人間が降りれば残りは CPU だけ** —— autoDrawForCPUs がその場で引き切り、
	// 最後の 1 人の Draw() が advancePhase を入れ子に呼ぶ。これが再入経路。
	// オールインは作らない: 賭けを伴わない SetAllIn はサイドポットの計算を
	// 歪め、クローン元の Omaha でも同じ差分が出る (テスト設定の作り物)。
	g.players[0].SetFolded(true)

	g.advancePhase() // -> Flop
	before := dramahaChipTotal(g)

	g.advancePhase() // -> Draw (CPU が引き切り、入れ子で先へ)

	assert.Equal(t, before, dramahaChipTotal(g),
		"ドローラウンドを跨いでも卓上の総額は変わらない (二重決着ならポットぶん増える)")

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
