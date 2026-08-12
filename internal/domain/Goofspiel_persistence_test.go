//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// goofspielStateAt は指定フェーズまで実際に指し進めた盤面を返す。
//
// **改竄テストの土台は手書きせず、本物の局面から作ります。** 手で組んだ JSON は
// 書き込み側の形を外しても気付けず、検証したいガードの手前で落ちてしまいます。
func goofspielStateAt(t *testing.T, phase GoofspielPhase) *Goofspiel {
	t.Helper()
	g := newGoofspielForTest(t, 2)
	switch phase {
	case GoofspielPhaseBid:
		return g
	case GoofspielPhaseReveal:
		require.NoError(t, g.PlayerBid(0))
		require.Equal(t, GoofspielPhaseReveal, g.GetPhase())
		return g
	case GoofspielPhaseGameEnd:
		for !g.GetGameEndFlag() {
			require.NoError(t, g.PlayerBid(0))
			if g.GetPhase() == GoofspielPhaseReveal && !g.GetGameEndFlag() {
				require.NoError(t, g.NextRound())
			}
		}
		return g
	default:
		t.Fatalf("未知のフェーズ %d", phase)
		return nil
	}
}

// goofspielTampered は本物の盤面を JSON にしてから 1 か所だけ壊し、復元を試みる。
func goofspielTampered(t *testing.T, base *Goofspiel, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back Goofspiel
	return json.Unmarshal(broken, &back)
}

// **壊していない盤面は素通しできなければならない。**
//
// これが負のコントロールです。これが落ちるなら、以下の改竄テストは「壊したから
// 落ちた」のではなく元から落ちていただけになります。
func TestGoofspiel_UntamperedStateRoundTrips(t *testing.T) {
	t.Parallel()

	for _, phase := range []GoofspielPhase{
		GoofspielPhaseBid, GoofspielPhaseReveal, GoofspielPhaseGameEnd,
	} {
		g := goofspielStateAt(t, phase)
		assert.NoError(t, goofspielTampered(t, g, func(map[string]any) {}),
			"フェーズ %d の無傷の盤面", phase)
	}
}

// **改竄した保存データは、壊した場所を名指しして弾く。**
//
// エラーの本文まで検査するのは、**手前の別のガードが落としたのを「検出できた」と
// 数えないため**です。範囲チェックは順番に並んでいるので、1 か所だけ壊したつもりが
// 別の不変条件を先に破っていることがあります。
func TestGoofspiel_UnmarshalRejectsTamperedState(t *testing.T) {
	t.Parallel()

	bid := goofspielStateAt(t, GoofspielPhaseBid)
	reveal := goofspielStateAt(t, GoofspielPhaseReveal)
	end := goofspielStateAt(t, GoofspielPhaseGameEnd)

	tests := []struct {
		name   string
		base   *Goofspiel
		mutate func(m map[string]any)
		want   string
	}{
		{
			name:   "人数が範囲外",
			base:   bid,
			mutate: func(m map[string]any) { m["cf"] = map[string]any{"p": 9, "tr": 0} },
			want:   "player count must be between",
		},
		{
			name:   "同点処理が範囲外",
			base:   bid,
			mutate: func(m map[string]any) { m["cf"] = map[string]any{"p": 2, "tr": 7} },
			want:   "tie rule must be",
		},
		{
			name:   "席数と設定人数が食い違う",
			base:   bid,
			mutate: func(m map[string]any) { m["pl"] = m["pl"].([]any)[:1] },
			want:   "does not match the configured",
		},
		{
			name:   "フェーズが範囲外",
			base:   bid,
			mutate: func(m map[string]any) { m["ph"] = 9 },
			want:   "phase out of range",
		},
		{
			name: "棋譜が長すぎる",
			base: bid,
			mutate: func(m map[string]any) {
				log := make([]any, goofspielMaxSliceLen+1)
				for i := range log {
					log[i] = map[string]any{}
				}
				m["al"] = log
			},
			want: "action log too long",
		},
		{
			name:   "伏せ札の枠が席数と合わない",
			base:   bid,
			mutate: func(m map[string]any) { m["bd"] = []any{nil} },
			want:   "bids has 1 slots for 2 seats",
		},
		{
			name:   "公開された入札の数が席数と合わない",
			base:   reveal,
			mutate: func(m map[string]any) { m["rb"] = m["rb"].([]any)[:1] },
			want:   "revealed bids has 1 entries for 2 seats",
		},
		{
			name:   "直前の勝者が範囲外",
			base:   bid,
			mutate: func(m map[string]any) { m["lw"] = 5 },
			want:   "last winner index out of range",
		},
		{
			name:   "勝者の席が範囲外",
			base:   bid,
			mutate: func(m map[string]any) { m["wi"] = 5 },
			want:   "winner index out of range",
		},
		{
			name:   "終局フラグとフェーズが矛盾する",
			base:   bid,
			mutate: func(m map[string]any) { m["ge"] = true },
			want:   "the game-end flag and the phase disagree",
		},
		{
			name:   "終局しているのに勝者が居ない",
			base:   end,
			mutate: func(m map[string]any) { m["wi"] = -1 },
			want:   "a finished game has a winner",
		},
		{
			name:   "ラウンド数が範囲外",
			base:   bid,
			mutate: func(m map[string]any) { m["rn"] = GoofspielRounds + 1 },
			want:   "round number out of range",
		},
		{
			name:   "得点の増分が負",
			base:   bid,
			mutate: func(m map[string]any) { m["lg"] = -1 },
			want:   "the last gain cannot be negative",
		},
		{
			// **同点なら誰も取らない。** 勝者が居ないのに点が動いた記録は矛盾です。
			name:   "勝者が居ないのに得点が動いている",
			base:   bid,
			mutate: func(m map[string]any) { m["lg"] = 5 },
			want:   "nobody won the round but 5 points moved",
		},
		{
			// **賞札 13 枚は増えも減りもしません。**
			name: "賞札の総数が合わない",
			base: bid,
			mutate: func(m map[string]any) {
				pile := m["pp"].([]any)
				m["pp"] = append(append([]any{}, pile...), pile...)
			},
			want: "the prizes do not add up",
		},
		{
			name:   "入札の場面なのに賞札が出ていない",
			base:   bid,
			mutate: func(m map[string]any) { m["cp"] = nil },
			want:   "the bidding phase needs a prize on the table",
		},
		{
			name: "公開後なのに賞札が残っている",
			base: reveal,
			mutate: func(m map[string]any) {
				m["cp"] = map[string]any{"d": GoofspielPrizeSuit(), "v": 4, "w": false}
			},
			want: "the prize is settled once the bids are revealed",
		},
		{
			name:   "席が欠けている",
			base:   bid,
			mutate: func(m map[string]any) { m["pl"].([]any)[0] = nil },
			want:   "seat 0 is missing",
		},
		{
			// **入札札は 13 枚から使ったぶんだけ減ります。**
			name: "入札札の残り枚数がラウンド数と合わない",
			base: bid,
			mutate: func(m map[string]any) {
				seat := m["pl"].([]any)[0].(map[string]any)
				gp := seat["gp"].(map[string]any)
				p := gp["p"].(map[string]any)
				p["c"] = p["c"].([]any)[:GoofspielRounds-2]
			},
			want: "seat 0 holds 11 bid cards in round 1, want 13",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := goofspielTampered(t, tt.base, tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want,
				"別のガードが先に落としている可能性がある")
		})
	}
}

func TestGoofspiel_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var g Goofspiel
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`), &g))

	var c GoofspielConfig
	assert.Error(t, json.Unmarshal([]byte(`{"p":`), &c))

	var p GoofspielPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"sc":`), &p))
}

// **設定は復元のたびに検証する。**
func TestGoofspielConfig_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		want string
	}{
		{name: "人数が少なすぎる", data: `{"p":1,"tr":0}`, want: "player count must be between"},
		{name: "人数が多すぎる", data: `{"p":4,"tr":0}`, want: "player count must be between"},
		{name: "同点処理が負", data: `{"p":2,"tr":-1}`, want: "tie rule must be"},
		{name: "同点処理が範囲外", data: `{"p":2,"tr":2}`, want: "tie rule must be"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var c GoofspielConfig
			err := json.Unmarshal([]byte(tt.data), &c)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.want)
		})
	}

	// 正しい設定は通る (負のコントロール)。
	var ok GoofspielConfig
	require.NoError(t, json.Unmarshal([]byte(`{"p":3,"tr":1}`), &ok))
	assert.Equal(t, 3, ok.PlayerCnt)
	assert.Equal(t, GoofspielTieCarryOver, ok.TieRule)
}

// **得点は賞札のランクの合計。** 13 枚ぶんを超えることも、負になることもありません。
func TestGoofspielPlayer_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	const maxScore = GoofspielRounds * (GoofspielRounds + 1) / 2

	tests := []struct {
		name string
		data string
		want string
	}{
		{name: "土台のプレイヤーが無い", data: `{"gp":null,"sc":0}`, want: "missing its base player"},
		{name: "得点が負", data: `{"gp":{"p":{"c":[]},"h":true,"f":false},"sc":-1}`, want: "score must be between"},
		{
			name: "得点が上限を超える",
			data: `{"gp":{"p":{"c":[]},"h":true,"f":false},"sc":92}`,
			want: "score must be between",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p GoofspielPlayer
			err := json.Unmarshal([]byte(tt.data), &p)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.want)
		})
	}

	// 上限ちょうどは通る (境界の負のコントロール)。
	var ok GoofspielPlayer
	require.NoError(t, json.Unmarshal(
		[]byte(`{"gp":{"p":{"c":[]},"h":true,"f":false},"sc":91}`), &ok))
	assert.Equal(t, maxScore, ok.GetScore())
}

// **持ち越し中の盤面も往復する。** 持ち越しは賞札の総数の数え方を変えます。
func TestGoofspiel_CarriedPrizesSurviveARoundTrip(t *testing.T) {
	t.Parallel()

	cfg := GoofspielConfig{PlayerCnt: 2, TieRule: GoofspielTieCarryOver}
	g := NewGoofspiel(newGoofspielSeats(2), cfg)
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	g.SetCurrentPrizeForTest(NewCard(GoofspielPrizeSuit(), 9, false))
	require.NoError(t, g.BidForTest(0, 6))
	require.NoError(t, g.BidForTest(1, 6))
	g.ResolveForTest()
	require.Len(t, g.GetCarriedPrizes(), 1)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	var back Goofspiel
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Len(t, back.GetCarriedPrizes(), 1)
	assert.Equal(t, g.PrizeValue(), back.PrizeValue())
}
