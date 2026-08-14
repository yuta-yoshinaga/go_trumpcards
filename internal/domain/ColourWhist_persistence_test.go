//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// colourWhistTampered は本物の盤面を JSON にしてから 1 か所だけ壊し、復元を試みる。
func colourWhistTampered(t *testing.T, base *ColourWhist, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back ColourWhist
	return json.Unmarshal(broken, &back)
}

// colourWhistMidPlay は Samen の契約でプレイ中の盤面を返す。
func colourWhistMidPlay(t *testing.T) *ColourWhist {
	t.Helper()
	g := newColourWhistWithHuman(t, 21)
	g.phase = ColourWhistPhaseCall
	g.contract = ColourWhistContractSamen
	g.declarerIdx = 0
	g.currentTurn = 0
	g.troelForced = false
	g.passed = make([]bool, ColourWhistPlayerCnt)
	require.NoError(t, g.call(0, CardDesignSpade))
	return g
}

// **到達できる盤面はすべて往復する。** これが負のコントロールです。
func TestColourWhist_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for seed := range 10 {
		g := NewDefaultColourWhist()
		g.SetRand(rand.New(rand.NewSource(int64(seed) + 1)))
		g.Reset()

		for turns := 0; turns < 400 && !g.GetGameEndFlag(); turns++ {
			data, err := json.Marshal(g)
			require.NoError(t, err)
			var back ColourWhist
			require.NoError(t, json.Unmarshal(data, &back),
				"seed %d %d 手目 (phase=%d, contract=%d, troel=%v): 書き込み側が破った",
				seed, turns, g.GetPhase(), g.GetContract(), g.IsTroelForced())
			assert.Equal(t, g.GetContract(), back.GetContract())
			assert.Equal(t, g.IsTroelForced(), back.IsTroelForced())

			switch {
			case g.GetPhase() == ColourWhistPhaseRoundEnd:
				require.NoError(t, g.NextRound())
			case !g.IsHumanTurn():
				g.CpuPlay()
			case g.GetPhase() == ColourWhistPhaseBid:
				require.NoError(t, g.Bid(ColourWhistContractNone))
			case g.GetPhase() == ColourWhistPhaseCall:
				require.NoError(t, g.Call(CardDesignHeart))
			case g.GetPhase() == ColourWhistPhasePlay:
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
func TestColourWhist_UnmarshalRejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"ラウンド数が範囲外", func(m map[string]any) { m["cf"] = map[string]any{"rd": 9999} }, "rounds must be"},
		{"席数が 4 でない", func(m map[string]any) { m["pl"] = m["pl"].([]any)[:3] }, "seat count 3 does not match"},
		{"席が欠けている", func(m map[string]any) { m["pl"].([]any)[1] = nil }, "seat 1 is missing"},
		{"フェーズが範囲外", func(m map[string]any) { m["ph"] = 99 }, "phase out of range"},
		{"終了フラグとフェーズが矛盾", func(m map[string]any) { m["ge"] = true }, "the game-end flag and the phase disagree"},
		{"契約が範囲外", func(m map[string]any) { m["ct"] = 99 }, "contract out of range"},
		{"親の席が範囲外", func(m map[string]any) { m["di"] = 9 }, "dealer index out of range"},
		{"手番が範囲外", func(m map[string]any) { m["cu"] = -1 }, "current turn index out of range"},
		{"契約者が範囲外", func(m map[string]any) { m["dc"] = 9 }, "declarer index out of range"},
		{"パスの枠が席数と合わない", func(m map[string]any) { m["ps"] = []any{true} }, "pass flags hold 1 slots"},
		{
			// **相方が付くのは Samen と Troel だけ。**
			name:   "Alleen なのに相方が居る",
			mutate: func(m map[string]any) { m["ct"] = ColourWhistContractAlleen; m["cc"] = nil },
			want:   "has no partner but seat",
		},
		{
			// **Troel は配りで決まるので札を指名しません。**
			name: "Troel なのに札を指名している",
			mutate: func(m map[string]any) {
				m["ct"] = ColourWhistContractTroel
			},
			want: "does not call a card",
		},
		{
			name: "Miserie なのに切り札がある",
			mutate: func(m map[string]any) {
				m["ct"] = ColourWhistContractMiserie
				m["pi"] = -1
				m["cc"] = nil
				m["pr"] = false
			},
			want: "played without trump but the suit is",
		},
		{
			// **Troel は競りをしません。** 降りた記録があれば矛盾です。
			name: "Troel なのに降りた記録がある",
			mutate: func(m map[string]any) {
				m["ct"] = ColourWhistContractTroel
				m["cc"] = nil
				m["tf"] = true
				m["ps"] = []any{false, true, false, false}
			},
			want: "troel skips the auction but seat 1",
		},
		{
			name: "Troel フラグだけ立っている",
			mutate: func(m map[string]any) {
				m["tf"] = true
			},
			want: "the troel flag is set but the contract is",
		},
		{"切り札のスートが範囲外", func(m map[string]any) { m["ts"] = 9 }, "trump suit out of range"},
		{"トリック数が範囲外", func(m map[string]any) { m["tn"] = ColourWhistTrickCnt + 1 }, "trick count out of range"},
		{"契約側のトリックが全体より多い", func(m map[string]any) { m["tn"] = 2; m["dt"] = 5 }, "took 5 of 2 tricks"},
		{
			name: "進行中のトリックに 4 枚以上ある",
			mutate: func(m map[string]any) {
				c := map[string]any{"pi": 0, "c": map[string]any{"d": 1, "v": 3, "w": false}}
				m["tk"] = []any{c, c, c, c}
			},
			want: "resolves at 4",
		},
		{
			name: "直前のトリックの枚数が 4 でない",
			mutate: func(m map[string]any) {
				c := map[string]any{"pi": 0, "c": map[string]any{"d": 1, "v": 3, "w": false}}
				m["lt"] = []any{c, c}
			},
			want: "the last trick holds 2 cards",
		},
		{"相方が公開済みなのに席が無い", func(m map[string]any) { m["pr"] = true; m["pi"] = -1 }, "revealed but no seat is named"},
		{"ラウンド数が上限超え", func(m map[string]any) { m["rn"] = 999 }, "round number out of range"},
		{
			name: "棋譜が長すぎる",
			mutate: func(m map[string]any) {
				log := make([]any, colourWhistMaxSliceLen+1)
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
			err := colourWhistTampered(t, colourWhistMidPlay(t), tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want, "別のガードが先に落としている可能性がある")
		})
	}
}

func TestColourWhist_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var g ColourWhist
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`), &g))

	var c ColourWhistConfig
	assert.Error(t, json.Unmarshal([]byte(`{"rd":`), &c))

	var p ColourWhistPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"gp":`), &p))
}

func TestColourWhistPlayer_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var p ColourWhistPlayer
	err := json.Unmarshal([]byte(`{"gp":null,"th":null,"sc":0}`), &p)
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing its base player")

	// **得点は負にもなります。** ゼロサムなので当然そうなります。
	var ok ColourWhistPlayer
	require.NoError(t, json.Unmarshal(
		[]byte(`{"gp":{"p":{"c":[]},"h":true,"f":false},"th":{"tt":[]},"sc":-5}`), &ok))
	assert.Equal(t, -5, ok.GetScore())
	ok.AddScore(8)
	assert.Equal(t, 3, ok.GetScore())
	ok.ResetGame()
	assert.Zero(t, ok.GetScore())
	assert.Zero(t, ok.CountAces())
}

func TestColourWhistConfig_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var c ColourWhistConfig
	require.Error(t, json.Unmarshal([]byte(`{"rd":1}`), &c))

	var ok ColourWhistConfig
	require.NoError(t, json.Unmarshal([]byte(`{"rd":16}`), &ok))
	assert.Equal(t, 16, ok.Rounds)
}

// **山札や棋譜が欠けていても落ちない。**
func TestColourWhist_UnmarshalFillsMissingSlices(t *testing.T) {
	base := colourWhistMidPlay(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	delete(m, "tc")
	delete(m, "al")
	delete(m, "ps")
	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back ColourWhist
	require.NoError(t, json.Unmarshal(broken, &back))
	assert.False(t, back.HasPassed(0))
	assert.False(t, back.HasPassed(3))
}
