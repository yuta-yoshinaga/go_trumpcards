//go:build test && (!js || !wasm || casino)

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDramaha_ShortBoardDoesNotPanic は、山が尽きてボードが 5 枚に届かなかった
// ときに advancePhase が固定の添字でボードを切らないことを見る。
//
// クローン元の Omaha/Hold'em はこの状態に到達できない (1 席 4 枚以下なので
// 9-max でも 41 枚) ため、`o.communityCards[4:]` を無条件に書いてよかった。
// ドラマハは 1 席 5 枚 + 交換 5 枚でその前提を壊している。
//
// cap == len で作るのが肝心。append の伸長で cap に余裕が残ると
// `[4:]` は空スライスとして通ってしまい、バグが再現しない —— そして
// KV から復元した盤は必ず cap == len なので、Worker 経路だけが落ちる。
func TestDramaha_ShortBoardDoesNotPanic(t *testing.T) {
	for _, boardLen := range []int{0, 3, 4} {
		t.Run("board of "+string(rune('0'+boardLen)), func(t *testing.T) {
			o := newTestDramaha()
			require.NoError(t, o.Reset())

			// cap == len でなければ意味がない。append で伸ばすと cap に余裕が
			// 残り、`[4:]` が空スライスとして通ってバグが再現しない。
			// make の 3 引数形は staticcheck が畳もうとするので、
			// 明示的に確保して cap を固定する。
			board := make([]*Card, boardLen)
			require.Equal(t, boardLen, cap(board), "cap == len でないと再現しない")
			copy(board, o.communityCards)
			o.communityCards = board
			for o.trumpCards.DrawCard() != nil { // 山を空にする
			}
			o.phase = DramahaPhaseDraw
			o.drawnFlags = []bool{true, true, true, true}

			assert.NotPanics(t, func() {
				o.advancePhase() // Draw -> Turn
				o.advancePhase() // Turn -> River
			}, "a board of %d cards must not be sliced at a fixed index", boardLen)
		})
	}
}

// TestDramaha_ResetClearsDrawnFlags は、ドロー済みフラグがハンドをまたいで
// 残らないことを見る。
//
// 残っていると、次のハンドの runOutDraw が「全員引き終えている」と読んで
// 1 枚も交換せず、オールインの走り切りがドロー側のポットを誰も引いていない
// 手で決めてしまう。長さでしか判定していないので、1 ハンド目だけは
// (nil から作り直されるため) 正しく動き、単発のテストでは見えない。
func TestDramaha_ResetClearsDrawnFlags(t *testing.T) {
	o := newTestDramaha()
	require.NoError(t, o.Reset())
	require.Len(t, o.drawnFlags, len(o.players))

	// 1 ハンド目がドローラウンドを終えた状態を作る。
	for i := range o.drawnFlags {
		o.drawnFlags[i] = true
	}

	require.NoError(t, o.Reset()) // 2 ハンド目

	for i, done := range o.drawnFlags {
		assert.False(t, done,
			"seat %d still marked as having drawn at the start of a new hand", i)
	}
}

// TestDramaha_RunOutDrawReplacesCardsOnEveryHand は上の一段上、
// 「実際に札が入れ替わるか」を runOutDraw 越しに見る。フラグを直接読む
// テストは、runOutDraw 側の判定が変わると素通りしてしまう。
func TestDramaha_RunOutDrawReplacesCardsOnEveryHand(t *testing.T) {
	o := newTestDramaha()

	replacedOnHand := func() int {
		before := make([][]*Card, len(o.players))
		for i, p := range o.players {
			before[i] = p.HoleCardsCopy()
		}
		o.runOutDraw()
		changed := 0
		for i, p := range o.players {
			after := p.HoleCardsCopy()
			for j := range after {
				if j < len(before[i]) && after[j] != before[i][j] {
					changed++
				}
			}
		}
		return changed
	}

	require.NoError(t, o.Reset())
	first := replacedOnHand()
	require.Positive(t, first, "the run-out draw must replace cards on the first hand")

	// ドローラウンドを終えたハンドを一度挟む。
	for i := range o.drawnFlags {
		o.drawnFlags[i] = true
	}

	require.NoError(t, o.Reset())
	second := replacedOnHand()
	assert.Positive(t, second,
		"the run-out draw replaced %d cards on hand 1 but %d on hand 2 — stale drawnFlags",
		first, second)
}
