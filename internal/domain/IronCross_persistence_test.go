//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// icPlayedBoard は何枚か開いた局面を返す (往復と改竄の土台)。
func icPlayedBoard(t *testing.T) *IronCross {
	t.Helper()
	for range 50 {
		g := newIronCrossForTest(t)
		for steps := 0; g.GetRevealedCount() < 2 && g.GetPhase() == IronCrossPhaseBetting; steps++ {
			require.Less(t, steps, 200)
			if err := g.PlayerAction(IronCrossActionCheck, 0); err != nil {
				require.NoError(t, g.PlayerAction(IronCrossActionCall, 0))
			}
		}
		if g.GetRevealedCount() >= 2 && g.GetPhase() == IronCrossPhaseBetting {
			return g
		}
	}
	t.Fatalf("50 回配っても 2 枚開いた局面が出なかった")
	return nil
}

// icTamper は保存 JSON を 1 か所だけ書き換えて復元させる。
func icTamper(t *testing.T, g *IronCross, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	return json.Unmarshal(tampered, new(IronCross))
}

// --- 往復 ---

func TestIronCross_RoundTrip(t *testing.T) {
	t.Parallel()
	g := icPlayedBoard(t)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(IronCross)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetRevealedCount(), restored.GetRevealedCount())
	assert.Equal(t, g.GetPot(), restored.GetPot())
	assert.Equal(t, g.GetTurnSeat(), restored.GetTurnSeat())
	assert.Equal(t, g.GetHandNumber(), restored.GetHandNumber())
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	// **十字は位置ごと復元される。** 伏せている位置は nil のまま。
	require.Len(t, restored.GetCross(), IronCrossCommunityCards)
	for i, c := range g.GetCross() {
		if c == nil {
			assert.Nil(t, restored.GetCross()[i], "位置 %d が伏せたまま復元されていない", i)
			continue
		}
		require.NotNil(t, restored.GetCross()[i], "位置 %d の札が失われている", i)
		assert.Equal(t, c.GetDesign(), restored.GetCross()[i].GetDesign(), "位置 %d のスート", i)
		assert.Equal(t, c.GetValue(), restored.GetCross()[i].GetValue(), "位置 %d の数字", i)
	}
	for i, p := range g.GetPlayers() {
		assert.Equal(t, p.GetChips(), restored.GetPlayers()[i].GetChips(), "席 %d のチップ", i)
		assert.Equal(t, p.GetCardsSize(), restored.GetPlayers()[i].GetCardsSize(), "席 %d の手札枚数", i)
		assert.Equal(t, p.GetLine(), restored.GetPlayers()[i].GetLine(), "席 %d の選んだ列", i)
	}
}

// **毎手ごとに往復させる。** 書き込み側の違反は局面を進めた後でしか出ない。
func TestIronCross_RoundTripEveryMove(t *testing.T) {
	t.Parallel()
	for range 20 {
		g := newIronCrossForTest(t)
		for hand := 0; hand < 2 && !g.GetGameEndFlag(); hand++ {
			for steps := 0; g.GetPhase() == IronCrossPhaseBetting; steps++ {
				require.Less(t, steps, 200)
				data, err := json.Marshal(g)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(data, new(IronCross)),
					"phase=%d revealed=%d turn=%d で保存が自分の検証に落ちた",
					g.GetPhase(), g.GetRevealedCount(), g.GetTurnSeat())
				if err := g.PlayerAction(IronCrossActionCheck, 0); err != nil {
					require.NoError(t, g.PlayerAction(IronCrossActionCall, 0))
				}
			}
			// **選択の場面も保存できる。** ここを飛ばすと一番新しい状態が未検査になる。
			if g.IsChoosing() {
				data, err := json.Marshal(g)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(data, new(IronCross)), "選択の場面の保存が落ちた")
				require.NoError(t, g.ChooseLine(IronCrossLineVertical))
			}
			data, err := json.Marshal(g)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(data, new(IronCross)), "決着後の保存が落ちた")
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextHand())
		}
	}
}

// --- 改竄 ---

func TestIronCross_RejectsTamperedSaves(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		mutate  func(m map[string]any)
		wantMsg string
	}{
		{
			// **これが本命。** 添字としては正当なので範囲検査では捕まらないが、
			// 通すと伏せたままの位置を役の判定に使うことになる。
			name:    "開いた枚数と実際の札の数が食い違う",
			mutate:  func(m map[string]any) { m["rv"] = IronCrossCommunityCards },
			wantMsg: "cards placed for",
		},
		{
			name:    "開いた枚数が範囲外",
			mutate:  func(m map[string]any) { m["rv"] = 99 },
			wantMsg: "revealed count out of range",
		},
		{
			name:    "十字の枠が 5 か所でない",
			mutate:  func(m map[string]any) { m["cr"] = []any{nil, nil} },
			wantMsg: "the cross holds",
		},
		{
			name:    "フェーズが範囲外",
			mutate:  func(m map[string]any) { m["ph"] = 9 },
			wantMsg: "phase out of range",
		},
		{
			name:    "手番が範囲外",
			mutate:  func(m map[string]any) { m["tu"] = 99 },
			wantMsg: "turn seat out of range",
		},
		{
			name:    "ハンド数が範囲外",
			mutate:  func(m map[string]any) { m["hn"] = 0 },
			wantMsg: "hand number out of range",
		},
		{
			name:    "ポットが負",
			mutate:  func(m map[string]any) { m["po"] = -1 },
			wantMsg: "pot must not be negative",
		},
		{
			name:    "レイズ回数が上限超え",
			mutate:  func(m map[string]any) { m["rc"] = ironCrossMaxRaisesPerRound + 1 },
			wantMsg: "raise count exceeds the cap",
		},
		{
			name:    "席数が設定と合わない",
			mutate:  func(m map[string]any) { m["pl"] = []any{} },
			wantMsg: "players for a",
		},
		{
			name:    "行動フラグの数が席と合わない",
			mutate:  func(m map[string]any) { m["af"] = []any{true} },
			wantMsg: "acted flags for",
		},
		{
			name:    "結果の数が席と合わない",
			mutate:  func(m map[string]any) { m["rs"] = []any{map[string]any{}} },
			wantMsg: "results for",
		},
		{
			name:    "棋譜が長すぎる",
			mutate:  func(m map[string]any) { m["al"] = make([]any, ironCrossMaxSliceLen+1) },
			wantMsg: "action log too long",
		},
		{
			name:    "設定が範囲外",
			mutate:  func(m map[string]any) { m["cf"] = map[string]any{"s": 9, "c": 1000, "a": 10} },
			wantMsg: "seats out of range",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := icTamper(t, icPlayedBoard(t), tt.mutate)
			require.Error(t, err, "改竄した保存が通ってしまった")
			assert.ErrorContains(t, err, tt.wantMsg)
		})
	}
}

// **選んだ列が範囲外の保存を拒む。**
func TestIronCross_RejectsAnInvalidLine(t *testing.T) {
	t.Parallel()
	err := icTamper(t, icPlayedBoard(t), func(m map[string]any) {
		players, ok := m["pl"].([]any)
		require.True(t, ok)
		seat, ok := players[0].(map[string]any)
		require.True(t, ok)
		seat["ln"] = 99
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "line out of range")
}

// **手札は 0 枚か 4 枚だけ。**
func TestIronCross_RejectsAPartialHand(t *testing.T) {
	t.Parallel()
	err := icTamper(t, icPlayedBoard(t), func(m map[string]any) {
		players, ok := m["pl"].([]any)
		require.True(t, ok)
		seat, ok := players[0].(map[string]any)
		require.True(t, ok)
		cards, ok := seat["cd"].([]any)
		require.True(t, ok)
		seat["cd"] = cards[:2]
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "must hold four cards")
}

// **負の残高・負の賭け金を弾く。**
func TestIronCross_RejectsNegativeSeatValues(t *testing.T) {
	t.Parallel()
	g := icPlayedBoard(t)
	for _, tt := range []struct{ key, want string }{
		{"c", "chips must not be negative"},
		{"b", "bet must not be negative"},
	} {
		err := icTamper(t, g, func(m map[string]any) {
			players, ok := m["pl"].([]any)
			require.True(t, ok)
			seat, ok := players[0].(map[string]any)
			require.True(t, ok)
			seat[tt.key] = -1
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, tt.want)
	}
}

// **山が無い保存でも落ちない。**
func TestIronCross_MissingDeckIsRebuilt(t *testing.T) {
	t.Parallel()
	g := icPlayedBoard(t)
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	m["dk"] = nil
	tampered, err := json.Marshal(m)
	require.NoError(t, err)

	restored := new(IronCross)
	require.NoError(t, json.Unmarshal(tampered, restored))
	assert.Positive(t, restored.GetRemainingCards(), "山が空のまま復元された")
}

func TestIronCross_ConfigCodec(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(DefaultIronCrossConfig())
	require.NoError(t, err)
	var restored IronCrossConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, DefaultIronCrossConfig(), restored)

	assert.Error(t, json.Unmarshal([]byte(`{"s":9,"c":1000,"a":10}`), new(IronCrossConfig)))
	assert.Error(t, json.Unmarshal([]byte(`{"s":"x"}`), new(IronCrossConfig)))
	assert.Error(t, json.Unmarshal([]byte(`{`), new(IronCross)))
	assert.Error(t, json.Unmarshal([]byte(`{"c":"x"}`), new(IronCrossPlayer)))
}
