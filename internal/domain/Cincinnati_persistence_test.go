//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cinPlayedBoard は実際に打った局面を返す (往復と改竄の土台)。
func cinPlayedBoard(t *testing.T) *Cincinnati {
	t.Helper()
	g := newCincinnatiForTest(t)
	// 何枚かめくれた局面まで進める。
	for steps := 0; g.GetRevealedCount() < 2 && g.GetPhase() == CincinnatiPhaseBetting; steps++ {
		require.Less(t, steps, 200)
		if err := g.PlayerAction(CincinnatiActionCheck, 0); err != nil {
			require.NoError(t, g.PlayerAction(CincinnatiActionCall, 0))
		}
	}
	return g
}

// cinTamper は保存 JSON を 1 か所だけ書き換えて復元させる。
func cinTamper(t *testing.T, g *Cincinnati, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	return json.Unmarshal(tampered, new(Cincinnati))
}

// --- 往復 ---

func TestCincinnati_RoundTrip(t *testing.T) {
	t.Parallel()
	g := cinPlayedBoard(t)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(Cincinnati)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetRevealedCount(), restored.GetRevealedCount())
	assert.Len(t, restored.GetCommunityCards(), len(g.GetCommunityCards()))
	assert.Equal(t, g.GetPot(), restored.GetPot())
	assert.Equal(t, g.GetCurrentBet(), restored.GetCurrentBet())
	assert.Equal(t, g.GetTurnSeat(), restored.GetTurnSeat())
	assert.Equal(t, g.GetHandNumber(), restored.GetHandNumber())
	assert.Equal(t, g.GetConfig(), restored.GetConfig())
	require.Len(t, restored.GetPlayers(), len(g.GetPlayers()))
	for i, p := range g.GetPlayers() {
		assert.Equal(t, p.GetChips(), restored.GetPlayers()[i].GetChips(), "席 %d のチップ", i)
		assert.Equal(t, p.GetCardsSize(), restored.GetPlayers()[i].GetCardsSize(), "席 %d の手札枚数", i)
		assert.Equal(t, p.GetName(), restored.GetPlayers()[i].GetName(), "席 %d の名前", i)
	}
}

// **毎手ごとに往復させる。** 書き込み側の違反は局面を進めた後でしか出ない。
func TestCincinnati_RoundTripEveryMove(t *testing.T) {
	t.Parallel()
	for range 20 {
		g := newCincinnatiForTest(t)
		for hand := 0; hand < 3 && !g.GetGameEndFlag(); hand++ {
			for steps := 0; g.GetPhase() == CincinnatiPhaseBetting; steps++ {
				require.Less(t, steps, 200)
				data, err := json.Marshal(g)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(data, new(Cincinnati)),
					"phase=%d revealed=%d turn=%d で保存が自分の検証に落ちた",
					g.GetPhase(), g.GetRevealedCount(), g.GetTurnSeat())
				if err := g.PlayerAction(CincinnatiActionCheck, 0); err != nil {
					require.NoError(t, g.PlayerAction(CincinnatiActionCall, 0))
				}
			}
			data, err := json.Marshal(g)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(data, new(Cincinnati)), "決着後の保存が落ちた")
			if g.GetGameEndFlag() {
				break
			}
			require.NoError(t, g.NextHand())
		}
	}
}

// --- 改竄 ---

func TestCincinnati_RejectsTamperedSaves(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		mutate  func(m map[string]any)
		wantMsg string
	}{
		{
			// **これが本命。** 添字としては正当なので範囲検査では捕まらないが、
			// 通すと伏せたままの札を役の判定に使うことになる。
			name:    "公開枚数とコミュニティの実枚数が食い違う",
			mutate:  func(m map[string]any) { m["rv"] = CincinnatiCommunityCards },
			wantMsg: "community cards for",
		},
		{
			name:    "公開枚数が範囲外",
			mutate:  func(m map[string]any) { m["rv"] = 99 },
			wantMsg: "revealed count out of range",
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
			mutate:  func(m map[string]any) { m["rc"] = cincinnatiMaxRaisesPerRound + 1 },
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
			mutate:  func(m map[string]any) { m["al"] = make([]any, cincinnatiMaxSliceLen+1) },
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
			err := cinTamper(t, cinPlayedBoard(t), tt.mutate)
			require.Error(t, err, "改竄した保存が通ってしまった")
			assert.ErrorContains(t, err, tt.wantMsg)
		})
	}
}

// **手札は 0 枚か 5 枚だけ。** 途中の枚数は「配っている最中」を意味する。
func TestCincinnati_RejectsAPartialHand(t *testing.T) {
	t.Parallel()
	g := cinPlayedBoard(t)
	err := cinTamper(t, g, func(m map[string]any) {
		players, ok := m["pl"].([]any)
		require.True(t, ok)
		seat, ok := players[0].(map[string]any)
		require.True(t, ok)
		cards, ok := seat["cd"].([]any)
		require.True(t, ok)
		seat["cd"] = cards[:3] // 3 枚だけ残す
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "must hold five cards")
}

// **負の残高・負の賭け金を弾く。**
func TestCincinnati_RejectsNegativeSeatValues(t *testing.T) {
	t.Parallel()
	g := cinPlayedBoard(t)
	for _, tt := range []struct{ key, want string }{
		{"c", "chips must not be negative"},
		{"b", "bet must not be negative"},
	} {
		err := cinTamper(t, g, func(m map[string]any) {
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
func TestCincinnati_MissingDeckIsRebuilt(t *testing.T) {
	t.Parallel()
	g := cinPlayedBoard(t)
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	m["dk"] = nil
	tampered, err := json.Marshal(m)
	require.NoError(t, err)

	restored := new(Cincinnati)
	require.NoError(t, json.Unmarshal(tampered, restored))
	assert.Positive(t, restored.GetRemainingCards(), "山が空のまま復元された")
}

func TestCincinnati_ConfigCodec(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(DefaultCincinnatiConfig())
	require.NoError(t, err)
	var restored CincinnatiConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, DefaultCincinnatiConfig(), restored)

	// **復元した設定も Validate に通す。**
	assert.Error(t, json.Unmarshal([]byte(`{"s":9,"c":1000,"a":10}`), new(CincinnatiConfig)))
	assert.Error(t, json.Unmarshal([]byte(`{"s":"x"}`), new(CincinnatiConfig)))
	assert.Error(t, json.Unmarshal([]byte(`{`), new(Cincinnati)))
	assert.Error(t, json.Unmarshal([]byte(`{"c":"x"}`), new(CincinnatiPlayer)))
}
