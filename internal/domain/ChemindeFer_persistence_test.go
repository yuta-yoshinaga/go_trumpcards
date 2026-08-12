//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chemindeFerTampered は本物の盤面を JSON にしてから 1 か所だけ壊し、復元を試みる。
//
// **土台は手書きせず、実際に進めた局面から作る。** 手書きの土台だと、そもそも
// 到達できない値を「壊す前」から持ってしまい、何を検査しているのか分からなくなる。
func chemindeFerTampered(t *testing.T, base *ChemindeFer, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back ChemindeFer
	return json.Unmarshal(broken, &back)
}

// chemindeFerBetOpen は席 0 が親で 100 を張り、まだ誰も賭けていない盤面を返す。
func chemindeFerBetOpen(t *testing.T) *ChemindeFer {
	t.Helper()
	g := newChemindeFerAllCpu(t, 77)
	require.NoError(t, g.SetStake(100))
	require.Equal(t, ChemindeFerPhaseBet, g.GetPhase())
	require.Equal(t, 0, g.GetBankerIdx())
	return g
}

// **到達できる盤面はすべて往復する。** これが負のコントロール。
//
// 改竄テストだけを書くと「何もかも弾く UnmarshalJSON」でも全部通ってしまう。
// 正しい盤面が確かに戻ることを、局面を進めながら毎手検査する。
func TestChemindeFer_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for seed := range 10 {
		g := newChemindeFerAllCpu(t, int64(seed)+1)

		for step := 0; step < 2000 && !g.GetGameEndFlag(); step++ {
			data, err := json.Marshal(g)
			require.NoError(t, err)

			var back ChemindeFer
			require.NoError(t, json.Unmarshal(data, &back),
				"seed %d の %d 手目 (phase=%d, result=%d): 書き込み側が codec の不変条件を破った",
				seed, step, g.GetPhase(), g.GetResult())

			assert.Equal(t, g.GetPhase(), back.GetPhase())
			assert.Equal(t, g.GetBankerIdx(), back.GetBankerIdx())
			assert.Equal(t, g.GetStake(), back.GetStake())
			assert.Equal(t, g.GetRepresentativeIdx(), back.GetRepresentativeIdx())
			assert.Equal(t, g.GetTotalChips(), back.GetTotalChips())
			assert.Equal(t, g.GetBetTurn(), back.GetBetTurn())
			assert.Equal(t, g.GetPunterTotal(), back.GetPunterTotal())
			assert.Equal(t, g.GetBankerTotal(), back.GetBankerTotal())

			if g.GetPhase() == ChemindeFerPhaseRoundEnd {
				require.NoError(t, g.NextRound())
				continue
			}
			g.CpuPlay()
		}
	}
}

// **改竄した保存データは、壊した場所を名指しして弾く。**
func TestChemindeFer_UnmarshalRejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{
			name:   "ラウンド数の設定が範囲外",
			mutate: func(m map[string]any) { m["cf"] = map[string]any{"rd": 9999, "ic": 1000} },
			want:   "rounds must be",
		},
		{
			name:   "席数が足りない",
			mutate: func(m map[string]any) { m["pl"] = m["pl"].([]any)[:2] },
			want:   "seat count 2 does not match",
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
			name:   "決着の値が範囲外",
			mutate: func(m map[string]any) { m["rs"] = 99 },
			want:   "result out of range",
		},
		{
			name:   "親の席が範囲外",
			mutate: func(m map[string]any) { m["bi"] = 9 },
			want:   "banker index out of range",
		},
		{
			name:   "代表の席が範囲外",
			mutate: func(m map[string]any) { m["ri"] = 9 },
			want:   "representative index out of range",
		},
		{
			name:   "バンク額が負",
			mutate: func(m map[string]any) { m["st"] = -1 },
			want:   "stake must not be negative",
		},
		{
			name:   "ラウンド番号が設定を超えている",
			mutate: func(m map[string]any) { m["rn"] = 999 },
			want:   "round number out of range",
		},
		{
			name: "手札が 4 枚ある",
			mutate: func(m map[string]any) {
				card := map[string]any{"d": 1, "v": 3, "w": false}
				m["bh"] = []any{card, card, card, card}
			},
			want: "a hand holds more than 3 cards",
		},
		{
			name: "棋譜が長すぎる",
			mutate: func(m map[string]any) {
				log := make([]any, chemindeFerMaxSliceLen+1)
				for i := range log {
					log[i] = map[string]any{}
				}
				m["al"] = log
			},
			want: "action log too long",
		},
		{
			// **親は子ではない。** 席番号としてはどちらも範囲内なので、
			// 範囲チェックだけでは通ってしまう。
			name:   "親が賭け順に混ざっている",
			mutate: func(m map[string]any) { m["bo"] = []any{1, 0, 2} },
			want:   "the banker (seat 0) is in the bet order",
		},
		{
			name:   "賭け順に同じ席が 2 度出てくる",
			mutate: func(m map[string]any) { m["bo"] = []any{1, 2, 1} },
			want:   "seat 1 appears twice in the bet order",
		},
		{
			name:   "賭け順に居ない席がある",
			mutate: func(m map[string]any) { m["bo"] = []any{9} },
			want:   "bet order holds seat 9",
		},
		{
			name:   "賭けの現在位置が範囲外",
			mutate: func(m map[string]any) { m["bp"] = 99 },
			want:   "bet position out of range",
		},
		{
			// **親が自分の子になることはない。**
			name:   "代表が親と同じ席",
			mutate: func(m map[string]any) { m["ri"] = 0 },
			want:   "is both the banker and the punter representative",
		},
		{
			// **子の賭け総額はバンク額を超えない。** 超えると親が払えない額を晒す。
			name: "賭けの総額がバンク額を超えている",
			mutate: func(m map[string]any) {
				m["st"] = 10
				m["pl"].([]any)[1].(map[string]any)["bt"] = 50
			},
			want: "punters staked 50 against a bank of 10",
		},
		{
			name:   "賭けが始まっているのにバンク額が 0",
			mutate: func(m map[string]any) { m["st"] = 0 },
			want:   "betting is open but the bank is 0",
		},
		{
			// **張る前に配ることはできない。**
			name: "張りフェーズなのに札が配られている",
			mutate: func(m map[string]any) {
				m["ph"] = int(ChemindeFerPhaseStake)
				m["bh"] = []any{map[string]any{"d": 1, "v": 3, "w": false}}
			},
			want: "cards are dealt before the bank is staked",
		},
		{
			// **3 枚目を引いたなら手札は 3 枚。**
			name: "引いたことになっているのに手札が 2 枚",
			mutate: func(m map[string]any) {
				card := map[string]any{"d": 1, "v": 3, "w": false}
				m["pd"] = true
				m["pn"] = []any{card, card}
			},
			want: "the punter drew but holds only 2 cards",
		},
		{
			// **決着はラウンド終了フェーズでしか持てない。**
			name:   "決着済みなのに進行中のフェーズ",
			mutate: func(m map[string]any) { m["rs"] = int(ChemindeFerResultBanker) },
			want:   "the coup is decided but the phase is",
		},
		{
			name:   "終了フラグとフェーズが矛盾する",
			mutate: func(m map[string]any) { m["ge"] = true },
			want:   "the game-end flag and the phase disagree",
		},
		{
			name: "席のチップが負",
			mutate: func(m map[string]any) {
				m["pl"].([]any)[1].(map[string]any)["ch"] = -5
			},
			want: "chips must not be negative",
		},
		{
			name: "席の賭け金が負",
			mutate: func(m map[string]any) {
				m["pl"].([]any)[1].(map[string]any)["bt"] = -5
			},
			want: "bet must not be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := chemindeFerTampered(t, chemindeFerBetOpen(t), tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want, "別のガードが先に落としている可能性がある")
		})
	}
}

func TestChemindeFer_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var g ChemindeFer
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`), &g))

	var c ChemindeFerConfig
	assert.Error(t, json.Unmarshal([]byte(`{"rd":`), &c))

	var p ChemindeFerPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":`), &p))
}

func TestChemindeFerConfig_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var bad ChemindeFerConfig
	assert.Error(t, json.Unmarshal([]byte(`{"rd":1,"ic":1000}`), &bad))

	var ok ChemindeFerConfig
	require.NoError(t, json.Unmarshal([]byte(`{"rd":12,"ic":500}`), &ok))
	assert.Equal(t, 12, ok.Rounds)
	assert.Equal(t, 500, ok.InitialChips)
}

func TestChemindeFerPlayer_RoundTrip(t *testing.T) {
	t.Parallel()

	p := NewChemindeFerPlayer("Alice", 250, true)
	p.SetBet(40)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var back ChemindeFerPlayer
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, "Alice", back.GetName())
	assert.Equal(t, 250, back.GetChips())
	assert.Equal(t, 40, back.GetBet())
	assert.True(t, back.GetIsHuman())

	assert.True(t, back.SubtractChips(50))
	assert.Equal(t, 200, back.GetChips())
	assert.False(t, back.SubtractChips(1000), "足りない額は引けない")
	assert.Equal(t, 200, back.GetChips(), "失敗した減算が額を変えた")
	back.AddChips(25)
	assert.Equal(t, 225, back.GetChips())
}

// **シューが欠けていても落ちない。** 新しいシューで補う。
func TestChemindeFer_UnmarshalFillsAMissingShoe(t *testing.T) {
	base := chemindeFerBetOpen(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	delete(m, "sh")
	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back ChemindeFer
	require.NoError(t, json.Unmarshal(broken, &back))
	assert.Equal(t, ChemindeFerDeckCnt*52, back.GetRemainingCards())
}

// 復元した盤面はそのまま指し続けられる。**復元できることと遊べることは別。**
func TestChemindeFer_RestoredGameKeepsPlaying(t *testing.T) {
	base := chemindeFerBetOpen(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var back ChemindeFer
	require.NoError(t, json.Unmarshal(data, &back))

	want := back.GetTotalChips()
	for step := 0; step < 2000 && !back.GetGameEndFlag(); step++ {
		if back.GetPhase() == ChemindeFerPhaseRoundEnd {
			require.NoError(t, back.NextRound())
			continue
		}
		back.CpuPlay()
		require.Equal(t, want, back.GetTotalChips(), "復元後にチップが湧いた")
	}
	assert.True(t, back.GetGameEndFlag(), "復元した盤面が終わらなかった")
}
