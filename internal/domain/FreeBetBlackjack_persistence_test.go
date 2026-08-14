//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fbTampered(t *testing.T, base *FreeBetBlackjack, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back FreeBetBlackjack
	return json.Unmarshal(broken, &back)
}

// fbMidRound は無料スプリット済みでプレイ中の盤面を返す。
func fbMidRound(t *testing.T) *FreeBetBlackjack {
	t.Helper()
	g := fbStaged(t, 100,
		[]*Card{fbCard(CardDesignSpade, 8), fbCard(CardDesignHeart, 8)},
		[]*Card{fbCard(CardDesignClover, 13), fbCard(CardDesignDiamond, 7)})
	require.NoError(t, g.FreeSplit())
	return g
}

// **到達できる盤面はすべて往復する。** これが負のコントロール。
func TestFreeBet_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	stages := []struct {
		name  string
		build func(t *testing.T) *FreeBetBlackjack
	}{
		{"賭ける前", func(t *testing.T) *FreeBetBlackjack { return newFreeBetForTest(t) }},
		{"配った直後", func(t *testing.T) *FreeBetBlackjack {
			g := newFreeBetForTest(t)
			require.NoError(t, g.PlaceBet(50))
			return g
		}},
		{"無料スプリット後", fbMidRound},
		{"決着後", func(t *testing.T) *FreeBetBlackjack {
			g := fbMidRound(t)
			for g.GetPhase() == FreeBetPhasePlay {
				require.NoError(t, g.Stand())
			}
			return g
		}},
	}

	for _, st := range stages {
		t.Run(st.name, func(t *testing.T) {
			g := st.build(t)
			data, err := json.Marshal(g)
			require.NoError(t, err)

			var back FreeBetBlackjack
			require.NoError(t, json.Unmarshal(data, &back),
				"書き込み側が codec の不変条件を破った (phase=%d)", g.GetPhase())

			assert.Equal(t, g.GetPhase(), back.GetPhase())
			assert.Equal(t, g.GetAnteBet(), back.GetAnteBet())
			assert.Equal(t, g.GetChips(), back.GetChips())
			assert.Equal(t, g.GetHandCount(), back.GetHandCount())
			assert.Equal(t, g.GetFreeBets(), back.GetFreeBets())
			assert.Equal(t, g.IsDealerPushed22(), back.IsDealerPushed22())
		})
	}
}

// **改竄した保存データは、壊した場所を名指しして弾く。**
func TestFreeBet_UnmarshalRejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"プレイヤーが欠けている", func(m map[string]any) { m["pl"] = nil }, "the player is missing"},
		{"設定が範囲外", func(m map[string]any) { m["cf"] = map[string]any{"ic": 5, "da": 50} }, "chips must be"},
		{"フェーズが範囲外", func(m map[string]any) { m["ph"] = 99 }, "phase out of range"},
		{"アンティが負", func(m map[string]any) { m["an"] = -10 }, "ante must not be negative"},
		{"ラウンド番号が 0", func(m map[string]any) { m["rn"] = 0 }, "round number out of range"},
		{
			// **手札とハウス出資は必ず同数。** ずれると別の手札の金を動かす。
			name:   "ハウス出資の数が手札と合わない",
			mutate: func(m map[string]any) { m["fb"] = []any{0} },
			want:   "free bets for 2 hands",
		},
		{
			name:   "決着の数が手札と合わない",
			mutate: func(m map[string]any) { m["rs"] = []any{0} },
			want:   "results for 2 hands",
		},
		{
			name:   "ハウス出資が負",
			mutate: func(m map[string]any) { m["fb"] = []any{-10, 0} },
			want:   "free bet must not be negative",
		},
		{
			name:   "決着の値が範囲外",
			mutate: func(m map[string]any) { m["rs"] = []any{99, 0} },
			want:   "result out of range",
		},
		{
			name:   "操作中の手札が範囲外",
			mutate: func(m map[string]any) { m["ah"] = 9 },
			want:   "active hand out of range",
		},
		{
			name: "スプリット上限を超えている",
			mutate: func(m map[string]any) {
				h := m["hd"].([]any)[0]
				m["hd"] = []any{h, h, h, h, h}
				m["fb"] = []any{0, 0, 0, 0, 0}
				m["rs"] = []any{0, 0, 0, 0, 0}
			},
			want: "exceeds the split limit",
		},
		{
			name:   "賭ける前なのに配られている",
			mutate: func(m map[string]any) { m["ph"] = int(FreeBetPhaseBet) },
			want:   "cards are dealt before the ante is placed",
		},
		{
			name:   "アンティ無しで配られている",
			mutate: func(m map[string]any) { m["an"] = 0; m["ph"] = int(FreeBetPhasePlay) },
			want:   "cards are dealt without an ante",
		},
		{
			name:   "棋譜が長すぎる",
			mutate: func(m map[string]any) { m["al"] = make([]any, freeBetMaxSliceLen+1) },
			want:   "action log too long",
		},
		{
			name: "チップが負",
			mutate: func(m map[string]any) {
				m["pl"].(map[string]any)["ch"] = -5
			},
			want: "chips must not be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fbTampered(t, fbMidRound(t), tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want, "別のガードが先に落としている可能性がある")
		})
	}
}

func TestFreeBet_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var g FreeBetBlackjack
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`), &g))

	var c FreeBetBlackjackConfig
	assert.Error(t, json.Unmarshal([]byte(`{"ic":`), &c))

	var p FreeBetBlackjackPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":`), &p))
}

func TestFreeBetConfig_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var bad FreeBetBlackjackConfig
	assert.Error(t, json.Unmarshal([]byte(`{"ic":1,"da":50}`), &bad))

	var ok FreeBetBlackjackConfig
	require.NoError(t, json.Unmarshal([]byte(`{"ic":500,"da":20}`), &ok))
	assert.Equal(t, 500, ok.InitialChips)
}

func TestFreeBetPlayer_RoundTrip(t *testing.T) {
	t.Parallel()

	p := NewFreeBetBlackjackPlayer(250)
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var back FreeBetBlackjackPlayer
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, 250, back.GetChips())
	assert.True(t, back.SubtractChips(50))
	assert.False(t, back.SubtractChips(1000), "足りない額は引けない")
	assert.Equal(t, 200, back.GetChips())
	back.AddChips(25)
	assert.Equal(t, 225, back.GetChips())
}

// **シューが欠けていても落ちない。**
func TestFreeBet_UnmarshalFillsAMissingShoe(t *testing.T) {
	base := fbMidRound(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	delete(m, "sh")
	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back FreeBetBlackjack
	require.NoError(t, json.Unmarshal(broken, &back))
	assert.Equal(t, 52*FreeBetDeckCount, back.GetRemainingCards())
}

// 復元した盤面はそのまま続けられる。
func TestFreeBet_RestoredGameKeepsPlaying(t *testing.T) {
	base := fbMidRound(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var back FreeBetBlackjack
	require.NoError(t, json.Unmarshal(data, &back))

	for step := 0; step < 50 && back.GetPhase() == FreeBetPhasePlay; step++ {
		require.NoError(t, back.Stand())
	}
	assert.Equal(t, FreeBetPhaseResult, back.GetPhase())
	require.NoError(t, back.NextRound())
	assert.Equal(t, FreeBetPhaseBet, back.GetPhase())
}
