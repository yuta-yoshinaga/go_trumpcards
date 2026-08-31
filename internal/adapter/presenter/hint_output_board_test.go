//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// **ヒントを押すと盤面が消えるゲームがあった。**`HintOutput` は
// `buildBaseOutput` で盤面を埋めた直後に空配列で上書きする作りで、
// 「ヒント応答はページの state にマージされない」という前提の下では安全だった。
// だが 9 本のページは `useGameApi` の `exec('hint')` を呼んでおり、
// `useGameApi` は `setState(res)` で**状態を丸ごと差し替える** ── 空になった
// 盤面がそのまま画面に流れ込む (#6800)。
//
// 前提が破れている 9 本については上書きをやめた。ここではその不変条件を
// **presenter 自身の Output と突き合わせて**確かめる: 盤面の中身を game ごとに
// 知る必要がなく、ヒント応答が盤面について嘘をつかないことだけを見る。
func TestHintOutputKeepsTheBoardForPagesThatMergeIt(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		output func(t *testing.T) (full string, hint string)
	}{
		{
			name:   "osmosis",
			fields: []string{"waste", "reserve", "foundation"},
			output: func(t *testing.T) (string, string) {
				p := &presenter.OsmosisWebPresenter{}
				g := domain.NewDefaultOsmosis()
				g.Reset()
				// **配り直後はウェイストが空。**空同士を比べても何も証明しない
				// ので、1 枚引いて 3 つとも中身のある盤面にしてから測る。
				require.NoError(t, g.Draw())
				return p.Output(g, nil), p.HintOutput(g)
			},
		},
	}

	decode := func(t *testing.T, raw string) map[string]json.RawMessage {
		t.Helper()
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(raw), &m))
		return m
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullRaw, hintRaw := tt.output(t)
			full, hint := decode(t, fullRaw), decode(t, hintRaw)

			for _, f := range tt.fields {
				require.Contains(t, full, f, "the field list is stale: %q is not in this game's output", f)
				// **配り直後の盤面が空でないことを先に確かめる。**ここを見ないと、
				// 「両方空」でも一致して通ってしまう。
				assert.NotEqual(t, "[]", string(full[f]), "the full board is already empty; the comparison would prove nothing")
				assert.JSONEq(t, string(full[f]), string(hint[f]),
					"HintOutput disagrees with Output about %q — the page merges this response and the board would vanish", f)
			}
		})
	}
}
