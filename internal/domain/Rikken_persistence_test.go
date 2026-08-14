//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rikkenTampered は本物の盤面を JSON にしてから 1 か所だけ壊し、復元を試みる。
//
// **土台は手書きせず、実際に指し進めた局面から作ります。**
func rikkenTampered(t *testing.T, base *Rikken, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back Rikken
	return json.Unmarshal(broken, &back)
}

// rikkenMidPlay は Rik の契約でプレイ中の盤面を返す。
func rikkenMidPlay(t *testing.T) *Rikken {
	t.Helper()
	g := newRikkenWithHuman(t, 21)
	g.phase = RikkenPhaseCall
	g.contract = RikkenContractRik
	g.declarerIdx = 0
	g.currentTurn = 0
	require.NoError(t, g.call(0, CardDesignSpade))
	return g
}

// **到達できる盤面はすべて往復する。** これが負のコントロールです。
func TestRikken_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for seed := range 10 {
		g := NewDefaultRikken()
		g.SetRand(rand.New(rand.NewSource(int64(seed) + 1)))
		g.Reset()

		for turns := 0; turns < 400 && !g.GetGameEndFlag(); turns++ {
			data, err := json.Marshal(g)
			require.NoError(t, err)
			var back Rikken
			require.NoError(t, json.Unmarshal(data, &back),
				"seed %d %d 手目 (phase=%d, contract=%d): 書き込み側が codec の不変条件を破った",
				seed, turns, g.GetPhase(), g.GetContract())
			assert.Equal(t, g.GetPhase(), back.GetPhase())
			assert.Equal(t, g.GetContract(), back.GetContract())
			assert.Equal(t, g.GetTrumpSuit(), back.GetTrumpSuit())

			switch {
			case g.GetPhase() == RikkenPhaseRoundEnd:
				require.NoError(t, g.NextRound())
			case !g.IsHumanTurn():
				g.CpuPlay()
			case g.GetPhase() == RikkenPhaseBid:
				require.NoError(t, g.Bid(RikkenContractNone))
			case g.GetPhase() == RikkenPhaseCall:
				require.NoError(t, g.Call(CardDesignHeart))
			case g.GetPhase() == RikkenPhasePlay:
				valid := g.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayCard(valid[0]))
			default:
				t.Fatalf("seed %d: 進めないフェーズ %d", seed, g.GetPhase())
			}
		}
	}
}

// **改竄した保存データは、壊した場所を名指しして弾く。**
func TestRikken_UnmarshalRejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{
			name:   "ラウンド数が範囲外",
			mutate: func(m map[string]any) { m["cf"] = map[string]any{"rd": 9999} },
			want:   "rounds must be",
		},
		{
			name:   "席数が 4 でない",
			mutate: func(m map[string]any) { m["pl"] = m["pl"].([]any)[:2] },
			want:   "seat count 2 does not match",
		},
		{
			name:   "席が欠けている",
			mutate: func(m map[string]any) { m["pl"].([]any)[2] = nil },
			want:   "seat 2 is missing",
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
			name:   "契約が範囲外",
			mutate: func(m map[string]any) { m["ct"] = 99 },
			want:   "contract out of range",
		},
		{
			name:   "親の席が範囲外",
			mutate: func(m map[string]any) { m["di"] = 9 },
			want:   "dealer index out of range",
		},
		{
			name:   "手番が範囲外",
			mutate: func(m map[string]any) { m["cu"] = -1 },
			want:   "current turn index out of range",
		},
		{
			name:   "落札者が範囲外",
			mutate: func(m map[string]any) { m["dc"] = 9 },
			want:   "declarer index out of range",
		},
		{
			name:   "パスの枠が席数と合わない",
			mutate: func(m map[string]any) { m["ps"] = []any{true, false} },
			want:   "pass flags hold 2 slots",
		},
		{
			// **相方が付くのは Rik だけ。** 範囲チェックは通ってしまいます。
			name: "Rik ではないのに相方が居る",
			mutate: func(m map[string]any) {
				m["ct"] = RikkenContractSolo
				m["ts"] = CardDesignSpade
			},
			want: "has no partner but seat",
		},
		{
			// **Misere に切り札はありません。**
			name: "Misere なのに切り札がある",
			mutate: func(m map[string]any) {
				m["ct"] = RikkenContractMisere
				m["pi"] = -1
				m["cc"] = nil
				m["ts"] = CardDesignHeart
			},
			want: "played without trump but the suit is",
		},
		{
			name: "Rik ではないのに札を指名している",
			mutate: func(m map[string]any) {
				m["ct"] = RikkenContractSolo
				m["pi"] = -1
				m["ts"] = CardDesignSpade
			},
			want: "does not call a card",
		},
		{
			name: "相方が公開済みなのに席が無い",
			mutate: func(m map[string]any) {
				m["pr"] = true
				m["pi"] = -1
			},
			want: "revealed but no seat is named",
		},
		{
			name:   "切り札のスートが範囲外",
			mutate: func(m map[string]any) { m["ts"] = 9 },
			want:   "trump suit out of range",
		},
		{
			name:   "トリック数が範囲外",
			mutate: func(m map[string]any) { m["tn"] = RikkenTrickCnt + 1 },
			want:   "trick count out of range",
		},
		{
			// **宣言側が取ったトリックが全体を超えることはありません。**
			name:   "宣言側のトリックが全体より多い",
			mutate: func(m map[string]any) { m["tn"] = 2; m["dt"] = 5 },
			want:   "took 5 of 2 tricks",
		},
		{
			name: "進行中のトリックに 4 枚以上ある",
			mutate: func(m map[string]any) {
				card := map[string]any{"pi": 0, "c": map[string]any{"d": 1, "v": 3, "w": false}}
				m["tk"] = []any{card, card, card, card}
			},
			want: "resolves at 4",
		},
		{
			name: "直前のトリックの枚数が 4 でない",
			mutate: func(m map[string]any) {
				card := map[string]any{"pi": 0, "c": map[string]any{"d": 1, "v": 3, "w": false}}
				m["lt"] = []any{card, card}
			},
			want: "the last trick holds 2 cards",
		},
		{
			name:   "ラウンド数が上限を超えている",
			mutate: func(m map[string]any) { m["rn"] = 999 },
			want:   "round number out of range",
		},
		{
			name: "棋譜が長すぎる",
			mutate: func(m map[string]any) {
				log := make([]any, rikkenMaxSliceLen+1)
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
			err := rikkenTampered(t, rikkenMidPlay(t), tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want, "別のガードが先に落としている可能性がある")
		})
	}
}

func TestRikken_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var g Rikken
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`), &g))

	var c RikkenConfig
	assert.Error(t, json.Unmarshal([]byte(`{"rd":`), &c))

	var p RikkenPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"gp":`), &p))
}

func TestRikkenPlayer_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var p RikkenPlayer
	err := json.Unmarshal([]byte(`{"gp":null,"th":null,"sc":0}`), &p)
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing its base player")

	// **得点は負にもなります。** ゼロサムなので当然そうなります。
	var ok RikkenPlayer
	require.NoError(t, json.Unmarshal(
		[]byte(`{"gp":{"p":{"c":[]},"h":true,"f":false},"th":{"tt":[]},"sc":-7}`), &ok))
	assert.Equal(t, -7, ok.GetScore())
	assert.True(t, ok.GetIsHuman())

	ok.AddScore(10)
	assert.Equal(t, 3, ok.GetScore())
	ok.ResetGame()
	assert.Zero(t, ok.GetScore())
}

func TestRikkenConfig_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var c RikkenConfig
	require.Error(t, json.Unmarshal([]byte(`{"rd":1}`), &c))

	var ok RikkenConfig
	require.NoError(t, json.Unmarshal([]byte(`{"rd":12}`), &ok))
	assert.Equal(t, 12, ok.Rounds)
}

// **山札や棋譜が欠けていても落ちない。** 空で補います。
func TestRikken_UnmarshalFillsMissingSlices(t *testing.T) {
	base := rikkenMidPlay(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	delete(m, "tc")
	delete(m, "al")
	delete(m, "ps")
	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back Rikken
	require.NoError(t, json.Unmarshal(broken, &back))
	assert.False(t, back.HasPassed(0), "パスの枠は席数ぶん補う")
	assert.False(t, back.HasPassed(3))
}
