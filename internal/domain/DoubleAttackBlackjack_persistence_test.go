//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func daTampered(t *testing.T, base *DoubleAttackBlackjack, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back DoubleAttackBlackjack
	return json.Unmarshal(broken, &back)
}

// daMidRound は追加ベット済みでプレイ中の盤面を返す。
func daMidRound(t *testing.T) *DoubleAttackBlackjack {
	t.Helper()
	g := newDoubleAttackForTest(t)
	require.NoError(t, g.PlaceBet(50, 20))
	require.NoError(t, g.Attack(50))
	return g
}

// **到達できる盤面はすべて往復する。** これが負のコントロール。
func TestDoubleAttack_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	stages := []struct {
		name  string
		build func(t *testing.T) *DoubleAttackBlackjack
	}{
		{"賭ける前", func(t *testing.T) *DoubleAttackBlackjack { return newDoubleAttackForTest(t) }},
		{"アップカードのみ", func(t *testing.T) *DoubleAttackBlackjack {
			g := newDoubleAttackForTest(t)
			require.NoError(t, g.PlaceBet(50, 20))
			return g
		}},
		{"追加ベット後", daMidRound},
		{"決着後", func(t *testing.T) *DoubleAttackBlackjack {
			g := daMidRound(t)
			for g.GetPhase() == DoubleAttackPhasePlay {
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

			var back DoubleAttackBlackjack
			require.NoError(t, json.Unmarshal(data, &back),
				"書き込み側が codec の不変条件を破った (phase=%d)", g.GetPhase())

			assert.Equal(t, g.GetPhase(), back.GetPhase())
			assert.Equal(t, g.GetAnteBet(), back.GetAnteBet())
			assert.Equal(t, g.GetAttackBet(), back.GetAttackBet())
			assert.Equal(t, g.GetBustItBet(), back.GetBustItBet())
			assert.Equal(t, g.GetChips(), back.GetChips())
			assert.Equal(t, g.GetHandCount(), back.GetHandCount())
			assert.Equal(t, g.IsDealerHoleDealt(), back.IsDealerHoleDealt())
			assert.Len(t, back.GetDealerCards(), len(g.GetDealerCards()))
		})
	}
}

// **改竄した保存データは、壊した場所を名指しして弾く。**
func TestDoubleAttack_UnmarshalRejectsTamperedState(t *testing.T) {
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
			name: "手札がスプリット上限を超えている",
			mutate: func(m map[string]any) {
				h := m["hd"].([]any)[0]
				m["hd"] = []any{h, h, h, h, h}
				m["rs"] = []any{0, 0, 0, 0, 0}
			},
			want: "exceeds the split limit",
		},
		{
			name:   "決着の数が手札と合わない",
			mutate: func(m map[string]any) { m["rs"] = []any{0, 0} },
			want:   "results for 1 hands",
		},
		{
			name:   "決着の値が範囲外",
			mutate: func(m map[string]any) { m["rs"] = []any{99} },
			want:   "result out of range",
		},
		{
			name:   "操作中の手札が範囲外",
			mutate: func(m map[string]any) { m["ah"] = 9 },
			want:   "active hand out of range",
		},
		{
			// **追加ベットはアンティを超えない。** 賭けていない額に配当が付く。
			name:   "追加ベットがアンティより大きい",
			mutate: func(m map[string]any) { m["at"] = 9999 },
			want:   "exceeds the ante",
		},
		{
			// **アップカードだけの間に 2 枚目があってはいけない。**
			name: "追加ベット前にホールカードがある",
			mutate: func(m map[string]any) {
				m["dd"] = false
			},
			want: "hole card is dealt before the double attack",
		},
		{
			name:   "棋譜が長すぎる",
			mutate: func(m map[string]any) { m["al"] = make([]any, doubleAttackMaxSliceLen+1) },
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
			err := daTampered(t, daMidRound(t), tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want, "別のガードが先に落としている可能性がある")
		})
	}
}

// 賭ける前に配られている保存データを弾く。
func TestDoubleAttack_UnmarshalRejectsCardsBeforeTheAnte(t *testing.T) {
	err := daTampered(t, daMidRound(t), func(m map[string]any) {
		m["ph"] = int(DoubleAttackPhaseBet)
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "cards are dealt before the ante is placed")
}

func TestDoubleAttack_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var g DoubleAttackBlackjack
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`), &g))

	var c DoubleAttackBlackjackConfig
	assert.Error(t, json.Unmarshal([]byte(`{"ic":`), &c))

	var p DoubleAttackBlackjackPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":`), &p))
}

func TestDoubleAttackConfig_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var bad DoubleAttackBlackjackConfig
	assert.Error(t, json.Unmarshal([]byte(`{"ic":1,"da":50}`), &bad))

	var ok DoubleAttackBlackjackConfig
	require.NoError(t, json.Unmarshal([]byte(`{"ic":500,"da":20}`), &ok))
	assert.Equal(t, 500, ok.InitialChips)
}

func TestDoubleAttackPlayer_RoundTrip(t *testing.T) {
	t.Parallel()

	p := NewDoubleAttackBlackjackPlayer(250)
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var back DoubleAttackBlackjackPlayer
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, 250, back.GetChips())
	assert.True(t, back.SubtractChips(50))
	assert.False(t, back.SubtractChips(1000), "足りない額は引けない")
	assert.Equal(t, 200, back.GetChips())
	back.AddChips(25)
	assert.Equal(t, 225, back.GetChips())
}

// **シューが欠けていても落ちない。**
func TestDoubleAttack_UnmarshalFillsAMissingShoe(t *testing.T) {
	base := daMidRound(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	delete(m, "sh")
	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back DoubleAttackBlackjack
	require.NoError(t, json.Unmarshal(broken, &back))
	assert.Equal(t, DoubleAttackDeckSize*DoubleAttackDeckCount, back.GetRemainingCards())
}

// 復元した盤面はそのまま続けられる。
func TestDoubleAttack_RestoredGameKeepsPlaying(t *testing.T) {
	base := daMidRound(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var back DoubleAttackBlackjack
	require.NoError(t, json.Unmarshal(data, &back))

	for step := 0; step < 50 && back.GetPhase() == DoubleAttackPhasePlay; step++ {
		require.NoError(t, back.Stand())
	}
	assert.Equal(t, DoubleAttackPhaseResult, back.GetPhase())
	require.NoError(t, back.NextRound())
	assert.Equal(t, DoubleAttackPhaseBet, back.GetPhase())
}
