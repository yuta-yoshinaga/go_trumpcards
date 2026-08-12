//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// botifarraTampered は本物の盤面を JSON にしてから 1 か所だけ壊し、復元を試みる。
//
// **土台は手書きせず、実際に指し進めた局面から作ります。** 手で組んだ JSON は
// 書き込み側の形を外していても気付けず、検証したいガードの手前で落ちます。
func botifarraTampered(t *testing.T, base *Botifarra, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back Botifarra
	return json.Unmarshal(broken, &back)
}

// botifarraMidPlay は札が何枚か出たところで止まった盤面を返す。
func botifarraMidPlay(t *testing.T) *Botifarra {
	t.Helper()
	g := newBotifarraWithHuman(t, 11)
	require.Equal(t, BotifarraPhasePlay, g.GetPhase())
	return g
}

// **到達できる盤面はすべて往復する。** これが負のコントロールです。
//
// 1 手ごとに保存して読み直すので、**書き込み側が codec の不変条件を破っていれば
// ここで落ちます**（読み込み側のガードが厳しすぎる場合も同じ）。
func TestBotifarra_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for seed := range 12 {
		g := NewDefaultBotifarra()
		g.SetRand(rand.New(rand.NewSource(int64(seed) + 1)))
		g.Reset()

		for turns := 0; turns < 200 && !g.GetGameEndFlag(); turns++ {
			data, err := json.Marshal(g)
			require.NoError(t, err)
			var back Botifarra
			require.NoError(t, json.Unmarshal(data, &back),
				"seed %d %d 手目 (phase=%d): 書き込み側が codec の不変条件を破った",
				seed, turns, g.GetPhase())
			assert.Equal(t, g.GetPhase(), back.GetPhase())
			assert.Equal(t, g.GetTrumpSuit(), back.GetTrumpSuit())
			assert.Equal(t, g.GetTrickCount(), back.GetTrickCount())

			switch g.GetPhase() {
			case BotifarraPhaseDeclare, BotifarraPhaseDelegated:
				require.NoError(t, g.Declare(g.longestSuitOf(0)))
			case BotifarraPhaseDouble:
				require.NoError(t, g.PassDouble())
			case BotifarraPhasePlay:
				if !g.IsHumanTurn() {
					g.CpuPlay()
					continue
				}
				valid := g.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayCard(valid[0]))
			case BotifarraPhaseRoundEnd:
				require.NoError(t, g.NextRound())
			default:
				t.Fatalf("seed %d: 進めないフェーズ %d", seed, g.GetPhase())
			}
		}
	}
}

// **改竄した保存データは、壊した場所を名指しして弾く。**
//
// エラー本文まで検査するのは、**手前の別のガードが落としたのを「検出できた」と
// 数えないため**です。
func TestBotifarra_UnmarshalRejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{
			name:   "上がり点が範囲外",
			mutate: func(m map[string]any) { m["cf"] = map[string]any{"ts": 9999, "ad": true} },
			want:   "target score must be",
		},
		{
			name:   "席数が 4 でない",
			mutate: func(m map[string]any) { m["pl"] = m["pl"].([]any)[:3] },
			want:   "seat count 3 does not match",
		},
		{
			name:   "席が欠けている",
			mutate: func(m map[string]any) { m["pl"].([]any)[1] = nil },
			want:   "seat 1 is missing",
		},
		{
			name:   "フェーズが範囲外",
			mutate: func(m map[string]any) { m["ph"] = 99 },
			want:   "phase out of range",
		},
		{
			name:   "終了フラグとフェーズが矛盾する",
			mutate: func(m map[string]any) { m["ge"] = true },
			want:   "the game-end flag and the phase disagree",
		},
		{
			name:   "親の席が範囲外",
			mutate: func(m map[string]any) { m["di"] = 9 },
			want:   "dealer index out of range",
		},
		{
			name:   "手番が範囲外",
			mutate: func(m map[string]any) { m["ct"] = -1 },
			want:   "current turn index out of range",
		},
		{
			name:   "宣言者が範囲外",
			mutate: func(m map[string]any) { m["dc"] = 9 },
			want:   "declarer index out of range",
		},
		{
			name:   "直前のトリックの勝者が範囲外",
			mutate: func(m map[string]any) { m["lw"] = 9 },
			want:   "last trick winner out of range",
		},
		{
			name:   "勝ちチームが範囲外",
			mutate: func(m map[string]any) { m["wt"] = 5 },
			want:   "winner team out of range",
		},
		{
			name:   "切り札のスートが範囲外",
			mutate: func(m map[string]any) { m["ts"] = 9 },
			want:   "trump suit out of range",
		},
		{
			name:   "倍率が範囲外",
			mutate: func(m map[string]any) { m["ml"] = 3 },
			want:   "multiplier out of range",
		},
		{
			// **宣言前に倍率は上がらない。**
			name: "誰も宣言していないのに倍付けされている",
			mutate: func(m map[string]any) {
				m["dc"] = -1
				m["ml"] = BotifarraMultiplierContrar
			},
			want: "doubled before anyone declared",
		},
		{
			name:   "トリック数が範囲外",
			mutate: func(m map[string]any) { m["tn"] = BotifarraTrickCnt + 1 },
			want:   "trick count out of range",
		},
		{
			// **1 トリックは 4 枚で決着する。** 5 枚目は積まれません。
			name: "進行中のトリックに 4 枚以上ある",
			mutate: func(m map[string]any) {
				card := map[string]any{"pi": 0, "c": map[string]any{"d": 0, "v": 3, "w": false}}
				m["tk"] = []any{card, card, card, card}
			},
			want: "resolves at 4",
		},
		{
			name: "直前のトリックの枚数が 4 でない",
			mutate: func(m map[string]any) {
				card := map[string]any{"pi": 0, "c": map[string]any{"d": 0, "v": 3, "w": false}}
				m["lt"] = []any{card, card}
			},
			want: "the last trick holds 2 cards",
		},
		{
			name: "トリックの札が席を名指ししていない",
			mutate: func(m map[string]any) {
				m["tk"] = []any{map[string]any{"pi": 9, "c": map[string]any{"d": 0, "v": 3, "w": false}}}
			},
			want: "names seat 9",
		},
		{
			name: "トリックに空の枠がある",
			mutate: func(m map[string]any) {
				m["tk"] = []any{map[string]any{"pi": 0, "c": nil}}
			},
			want: "empty slot",
		},
		{
			name:   "ラウンドの点が範囲外",
			mutate: func(m map[string]any) { m["rp"] = []any{99, 0} },
			want:   "round points for team 0 out of range",
		},
		{
			name:   "通算得点が負",
			mutate: func(m map[string]any) { m["sc"] = []any{-1, 0} },
			want:   "score for team 0 cannot be negative",
		},
		{
			// **配り切ったラウンドの点は 72 ちょうど。**
			name: "決着したラウンドの点が 72 でない",
			mutate: func(m map[string]any) {
				m["tn"] = BotifarraTrickCnt
				m["rp"] = []any{10, 10}
			},
			want: "a finished round holds 20 points",
		},
		{
			name:   "場に 72 点より多く出ている",
			mutate: func(m map[string]any) { m["rp"] = []any{40, 40} },
			want:   "points are in play",
		},
		{
			name: "棋譜が長すぎる",
			mutate: func(m map[string]any) {
				log := make([]any, botifarraMaxSliceLen+1)
				for i := range log {
					log[i] = map[string]any{}
				}
				m["al"] = log
			},
			want: "action log too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := botifarraTampered(t, botifarraMidPlay(t), tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want, "別のガードが先に落としている可能性がある")
		})
	}
}

func TestBotifarra_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var g Botifarra
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`), &g))

	var c BotifarraConfig
	assert.Error(t, json.Unmarshal([]byte(`{"ts":`), &c))

	var p BotifarraPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"gp":`), &p))
}

// **土台のプレイヤーが欠けた保存データは弾く。**
func TestBotifarraPlayer_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var p BotifarraPlayer
	err := json.Unmarshal([]byte(`{"gp":null,"th":null}`), &p)
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing its base player")

	// 正しい形は通る (負のコントロール)。
	var ok BotifarraPlayer
	require.NoError(t, json.Unmarshal(
		[]byte(`{"gp":{"p":{"c":[]},"h":true,"f":false},"th":{"tt":[]}}`), &ok))
	assert.True(t, ok.GetIsHuman())
	assert.Zero(t, ok.GetTrickCount())
}

func TestBotifarraConfig_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var c BotifarraConfig
	require.Error(t, json.Unmarshal([]byte(`{"ts":1,"ad":true}`), &c))

	var ok BotifarraConfig
	require.NoError(t, json.Unmarshal([]byte(`{"ts":151,"ad":false}`), &ok))
	assert.Equal(t, 151, ok.TargetScore)
	assert.False(t, ok.AllowDoubling)
}
