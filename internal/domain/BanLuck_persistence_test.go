//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// blPlayedBoard は実際に打った局面を返す (往復と改竄の土台)。
func blPlayedBoard(t *testing.T) *BanLuck {
	t.Helper()
	g := newBanLuckForTest(t)
	blDealPlain(g)
	require.NoError(t, g.PlaceBet(50))
	require.Equal(t, BanLuckPhasePlay, g.GetPhase())
	return g
}

// blTamper は保存 JSON を 1 か所だけ書き換えて復元させる。
//
// **本物の局面を 1 か所だけ壊す。** 手書きの JSON を渡すと、壊した所より前で
// 落ちて「検証が効いている」ように見えてしまう。
func blTamper(t *testing.T, g *BanLuck, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	return json.Unmarshal(tampered, new(BanLuck))
}

// --- 往復 ---

func TestBanLuck_RoundTrip(t *testing.T) {
	t.Parallel()
	g := blPlayedBoard(t)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(BanLuck)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetBankerSeat(), restored.GetBankerSeat())
	assert.Equal(t, g.GetTurnSeat(), restored.GetTurnSeat())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetConfig(), restored.GetConfig())
	assert.Len(t, restored.GetPlayers(), len(g.GetPlayers()))
	assert.Len(t, restored.GetHands(), len(g.GetHands()))
	for i, p := range g.GetPlayers() {
		assert.Equal(t, p.GetChips(), restored.GetPlayers()[i].GetChips(), "席 %d のチップ", i)
		assert.Equal(t, p.GetBet(), restored.GetPlayers()[i].GetBet(), "席 %d の賭け金", i)
		assert.Equal(t, p.GetIsHuman(), restored.GetPlayers()[i].GetIsHuman(), "席 %d の人間フラグ", i)
		assert.Equal(t, g.GetHands()[i].GetScore(), restored.GetHands()[i].GetScore(), "席 %d の点数", i)
	}
	assert.Equal(t, g.GetResults(), restored.GetResults())
}

// **毎手ごとに往復させる。** 書き込み側の違反は、局面を進めた後でしか出ない。
func TestBanLuck_RoundTripEveryMove(t *testing.T) {
	t.Parallel()
	for range 30 {
		g := newBanLuckForTest(t)
		for steps := 0; !g.GetGameEndFlag(); steps++ {
			require.Less(t, steps, 3000)
			data, err := json.Marshal(g)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(data, new(BanLuck)),
				"phase=%d banker=%d turn=%d round=%d で保存が自分の検証に落ちた",
				g.GetPhase(), g.GetBankerSeat(), g.GetTurnSeat(), g.GetRoundNumber())
			blStep(t, g)
		}
	}
}

// --- 改竄 ---

func TestBanLuck_RejectsTamperedSaves(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		mutate  func(m map[string]any)
		wantMsg string
	}{
		{
			name:    "親の席が範囲外",
			mutate:  func(m map[string]any) { m["bk"] = 99 },
			wantMsg: "banker seat out of range",
		},
		{
			name:    "手番の席が範囲外",
			mutate:  func(m map[string]any) { m["tu"] = -1 },
			wantMsg: "turn seat out of range",
		},
		{
			name:    "席が少なすぎる",
			mutate:  func(m map[string]any) { m["pl"] = []any{} },
			wantMsg: "seats out of range",
		},
		{
			name:    "フェーズが範囲外",
			mutate:  func(m map[string]any) { m["ph"] = 9 },
			wantMsg: "phase out of range",
		},
		{
			name:    "ラウンド数が設定を超える",
			mutate:  func(m map[string]any) { m["rn"] = 999 },
			wantMsg: "round number out of range",
		},
		{
			name:    "役が範囲外",
			mutate:  func(m map[string]any) { m["rk"] = []any{9, 9, 9, 9} },
			wantMsg: "rank out of range",
		},
		{
			name:    "決着が範囲外",
			mutate:  func(m map[string]any) { m["oc"] = []any{7, 7, 7, 7} },
			wantMsg: "outcome out of range",
		},
		{
			// **席と同数でないスライスは範囲チェックを通る。** 精算が別の席の
			// 金を動かしても添字は正当なままなので、長さで捕まえるしかない。
			name:    "役の数が席と合わない",
			mutate:  func(m map[string]any) { m["rk"] = []any{1, 1} },
			wantMsg: "ranks for",
		},
		{
			name:    "増減の数が席と合わない",
			mutate:  func(m map[string]any) { m["dl"] = []any{0} },
			wantMsg: "deltas for",
		},
		{
			// **範囲としては全部正当。** フェーズと盤面の食い違いは別に見る。
			name:    "配る前なのに手札がある",
			mutate:  func(m map[string]any) { m["ph"] = int(BanLuckPhaseBet) },
			wantMsg: "dealt before the bets",
		},
		{
			name: "配った後なのに手札が無い",
			mutate: func(m map[string]any) {
				m["hd"] = []any{}
				m["rk"] = []any{}
				m["oc"] = []any{}
				m["dl"] = []any{}
			},
			wantMsg: "no cards dealt",
		},
		{
			name:    "プレイ中なのに精算済み",
			mutate:  func(m map[string]any) { m["st"] = true },
			wantMsg: "settled but still in play",
		},
		{
			name:    "棋譜が長すぎる",
			mutate:  func(m map[string]any) { m["al"] = make([]any, banLuckMaxSliceLen+1) },
			wantMsg: "action log too long",
		},
		{
			name:    "設定が範囲外",
			mutate:  func(m map[string]any) { m["cf"] = map[string]any{"s": 1, "c": 1000, "r": 10, "b": 50} },
			wantMsg: "seats out of range",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := blTamper(t, blPlayedBoard(t), tt.mutate)
			require.Error(t, err, "改竄した保存が通ってしまった")
			assert.ErrorContains(t, err, tt.wantMsg)
		})
	}
}

// **親に賭け金が乗った保存を拒む。** 添字としては全部正当なので、範囲検査では
// 捕まらない ── 通すと精算で親が自分から取ることになる。
func TestBanLuck_RejectsBankerWithABet(t *testing.T) {
	t.Parallel()
	g := blPlayedBoard(t)
	err := blTamper(t, g, func(m map[string]any) {
		players, ok := m["pl"].([]any)
		require.True(t, ok)
		seat, ok := players[g.GetBankerSeat()].(map[string]any)
		require.True(t, ok)
		seat["b"] = 50
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "carries a bet")
}

// **席の負の残高を拒む。**
func TestBanLuck_RejectsNegativeChips(t *testing.T) {
	t.Parallel()
	g := blPlayedBoard(t)
	err := blTamper(t, g, func(m map[string]any) {
		players, ok := m["pl"].([]any)
		require.True(t, ok)
		seat, ok := players[1].(map[string]any)
		require.True(t, ok)
		seat["c"] = -1
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "chips must not be negative")

	err = blTamper(t, g, func(m map[string]any) {
		players, ok := m["pl"].([]any)
		require.True(t, ok)
		seat, ok := players[1].(map[string]any)
		require.True(t, ok)
		seat["b"] = -1
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "bet must not be negative")
}

// **山が無い保存でも落ちない。** 復元後に引けなくなるとそこで固まる。
func TestBanLuck_MissingDeckIsRebuilt(t *testing.T) {
	t.Parallel()
	g := blPlayedBoard(t)
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	m["dk"] = nil
	tampered, err := json.Marshal(m)
	require.NoError(t, err)

	restored := new(BanLuck)
	require.NoError(t, json.Unmarshal(tampered, restored))
	assert.Positive(t, restored.GetRemainingCards(), "山が空のまま復元された")
}

func TestBanLuck_ConfigCodec(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(DefaultBanLuckConfig())
	require.NoError(t, err)
	var restored BanLuckConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, DefaultBanLuckConfig(), restored)

	// **復元した設定も Validate に通す。**
	assert.Error(t, json.Unmarshal([]byte(`{"s":1,"c":1000,"r":10,"b":50}`), new(BanLuckConfig)))
	assert.Error(t, json.Unmarshal([]byte(`{"s":4,"c":1000,"r":10,"b":55}`), new(BanLuckConfig)))
}

// **手番が「もう動けない席」を指した保存を拒む。**
//
// 席の添字としては正当なので範囲検査は通る。しかし復元後は誰も動かせず、
// 精算も走らないので**盤面が固まる** ── 配りの時に潰したのと同じ穴が、
// 復元経路から開く。
func TestBanLuck_RejectsTurnOnASeatThatCannotAct(t *testing.T) {
	t.Parallel()
	g := blPlayedBoard(t)
	turn := g.GetTurnSeat()

	for _, tt := range []struct {
		name   string
		break_ func(m map[string]any)
	}{
		{
			name: "手番の席が打ち止め済み",
			break_: func(m map[string]any) {
				hands, ok := m["hd"].([]any)
				require.True(t, ok)
				h, ok := hands[turn].(map[string]any)
				require.True(t, ok)
				h["st"] = true
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := blTamper(t, g, tt.break_)
			require.Error(t, err, "動けない席に手番がある保存が通ってしまった")
			assert.ErrorContains(t, err, "cannot act")
		})
	}
}
