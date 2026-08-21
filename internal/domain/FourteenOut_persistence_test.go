//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// foPersistCard is a local helper; the exported-package tests have their own.
func foPersistCard(design, value int) *Card { return NewCard(design, value, true) }

// foPersistBoard builds a 12-column board with the given leading columns.
func foPersistBoard(cols ...[]*Card) [][]*Card {
	board := make([][]*Card, FourteenOutColumnCnt)
	for i := range board {
		if i < len(cols) {
			board[i] = cols[i]
		}
	}
	return board
}

// **Worker はリクエストごとに KV から作り直す (#1860)。**アンドゥ履歴が往復
// しないと、復元したゲームは `cannot undo: no history` になり、Undo が黙って
// 効かなくなる。
func TestFourteenOut_PersistsUndoHistory(t *testing.T) {
	t.Parallel()
	g := NewDefaultFourteenOut()
	g.Reset()
	g.SetColumns(foPersistBoard(
		[]*Card{foPersistCard(CardDesignSpade, 3), foPersistCard(CardDesignSpade, 9)},
		[]*Card{foPersistCard(CardDesignHeart, 5)},
	))
	require.NoError(t, g.Remove(0, 1))

	data, err := json.Marshal(g)
	require.NoError(t, err)

	restored := NewDefaultFourteenOut()
	require.NoError(t, json.Unmarshal(data, restored))
	require.True(t, restored.CanUndo(), "history must survive the KV round trip")
	require.NoError(t, restored.Undo())
	assert.Equal(t, 0, restored.GetRemovedCount())
	require.Len(t, restored.GetColumns()[0], 2, "both cards come back")
	assert.Equal(t, 9, restored.GetColumns()[0][1].GetValue())
}

// スナップショットは中身まで復元される。深さだけ残って盤が空では、Undo が
// 盤を消す動作になってしまう。
func TestFourteenOut_SnapshotRestoresTheExactBoard(t *testing.T) {
	t.Parallel()
	g := NewDefaultFourteenOut()
	g.Reset()
	before := foPersistBoard(
		[]*Card{foPersistCard(CardDesignClover, 6), foPersistCard(CardDesignSpade, 8)},
		[]*Card{foPersistCard(CardDesignHeart, 6)},
	)
	g.SetColumns(before)
	require.NoError(t, g.Remove(0, 1))

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := NewDefaultFourteenOut()
	require.NoError(t, json.Unmarshal(data, restored))
	require.NoError(t, restored.Undo())

	got := restored.GetColumns()
	require.Len(t, got[0], 2)
	assert.Equal(t, CardDesignClover, got[0][0].GetDesign())
	assert.Equal(t, 6, got[0][0].GetValue())
	assert.Equal(t, CardDesignSpade, got[0][1].GetDesign())
	assert.Equal(t, 8, got[0][1].GetValue())
}

// **takeSnapshot は列の深いコピーを取らなければならない。**Remove は再スライス
// するだけなので、スナップショットが同じ配列を共有していると、Undo しても
// 取り除いた札が戻らないことがある。
func TestFourteenOut_SnapshotDoesNotShareTheColumnArray(t *testing.T) {
	t.Parallel()
	g := NewDefaultFourteenOut()
	g.Reset()
	g.SetColumns(foPersistBoard(
		[]*Card{foPersistCard(CardDesignSpade, 9)},
		[]*Card{foPersistCard(CardDesignHeart, 5)},
		[]*Card{foPersistCard(CardDesignClover, 8)},
		[]*Card{foPersistCard(CardDesignDiamond, 6)},
	))
	require.NoError(t, g.Remove(0, 1))
	require.NoError(t, g.Remove(2, 3))
	require.NoError(t, g.Undo())
	require.NoError(t, g.Undo())

	for i, want := range []int{9, 5, 8, 6} {
		require.Len(t, g.GetColumns()[i], 1, "column %d", i)
		assert.Equal(t, want, g.GetColumns()[i][0].GetValue(), "column %d", i)
	}
}

// --- trust-boundary rejections ---

func TestFourteenOut_RejectsOversizedHistory(t *testing.T) {
	t.Parallel()
	snaps := make([]string, 0, fourteenOutMaxSliceLen+1)
	for range fourteenOutMaxSliceLen + 1 {
		snaps = append(snaps, `{"rc":0}`)
	}
	payload := `{"hi":[` + joinStrs(snaps, ",") + `]}`
	assert.Error(t, json.Unmarshal([]byte(payload), NewDefaultFourteenOut()))
}

func TestFourteenOut_RejectsSnapshotActionLogLnOutOfRange(t *testing.T) {
	t.Parallel()
	for _, payload := range []string{`{"ll":-1}`, `{"ll":100000}`} {
		var s fourteenOutSnapshot
		assert.Error(t, json.Unmarshal([]byte(payload), &s), "payload %s", payload)
	}
}

// 列数は配り切った 12 を超えられない。Undo が範囲外を触って panic するより、
// 信頼境界で拒む。
func TestFourteenOut_RejectsSnapshotColumnCountOutOfRange(t *testing.T) {
	t.Parallel()
	cols := make([]string, 0, FourteenOutColumnCnt+1)
	for range FourteenOutColumnCnt + 1 {
		cols = append(cols, "[]")
	}
	payload := `{"cl":[` + joinStrs(cols, ",") + `]}`

	var s fourteenOutSnapshot
	assert.Error(t, json.Unmarshal([]byte(payload), &s))
	assert.Error(t, json.Unmarshal([]byte(payload), NewDefaultFourteenOut()))
}

func joinStrs(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}
