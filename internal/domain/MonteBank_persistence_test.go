//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mbPlayedBoard は実際に賭けて決着させた局面を返す (往復と改竄の土台)。
func mbPlayedBoard(t *testing.T) *MonteBank {
	t.Helper()
	g := newMonteBankForTest(t)
	require.NoError(t, g.PlaceBet(1, 50))
	require.Equal(t, MonteBankPhaseResult, g.GetPhase())
	return g
}

// mbTamper は保存 JSON を 1 か所だけ書き換えて復元させる。
//
// **本物の局面を 1 か所だけ壊す。** 手書きの JSON を渡すと、壊した所より前で
// 落ちて「検証が効いている」ように見えてしまう。
func mbTamper(t *testing.T, g *MonteBank, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	return json.Unmarshal(tampered, new(MonteBank))
}

// --- 往復 ---

func TestMonteBank_RoundTrip(t *testing.T) {
	t.Parallel()
	g := mbPlayedBoard(t)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(MonteBank)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetPick(), restored.GetPick())
	assert.Equal(t, g.GetBet(), restored.GetBet())
	assert.Equal(t, g.GetResult(), restored.GetResult())
	assert.Equal(t, g.GetPayout(), restored.GetPayout())
	assert.Equal(t, g.GetChips(), restored.GetChips())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetConfig(), restored.GetConfig())
	require.Len(t, restored.GetLayout(), MonteBankLayoutSize)
	for i, c := range g.GetLayout() {
		assert.Equal(t, c.GetDesign(), restored.GetLayout()[i].GetDesign(), "場札 %d のスート", i)
		assert.Equal(t, c.GetValue(), restored.GetLayout()[i].GetValue(), "場札 %d の数字", i)
	}
	require.NotNil(t, restored.GetGate())
	assert.Equal(t, g.GetGate().GetDesign(), restored.GetGate().GetDesign())
}

// **毎手ごとに往復させる。** 書き込み側の違反は、局面を進めた後でしか出ない。
func TestMonteBank_RoundTripEveryMove(t *testing.T) {
	t.Parallel()
	for range 30 {
		g := newMonteBankForTest(t)
		for steps := 0; !g.GetGameEndFlag(); steps++ {
			require.Less(t, steps, 100)
			data, err := json.Marshal(g)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(data, new(MonteBank)),
				"phase=%d pick=%d round=%d で保存が自分の検証に落ちた",
				g.GetPhase(), g.GetPick(), g.GetRoundNumber())

			if g.GetPhase() == MonteBankPhaseBet {
				require.NoError(t, g.PlaceBet(0, MonteBankMinBet))
			} else {
				require.NoError(t, g.NextRound())
			}
		}
		// 終局後も保存できる。
		data, err := json.Marshal(g)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, new(MonteBank)))
	}
}

// --- 改竄 ---

func TestMonteBank_RejectsTamperedSaves(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		mutate  func(m map[string]any)
		wantMsg string
	}{
		{
			name:    "フェーズが範囲外",
			mutate:  func(m map[string]any) { m["ph"] = 9 },
			wantMsg: "phase out of range",
		},
		{
			name:    "決着が範囲外",
			mutate:  func(m map[string]any) { m["rs"] = 7 },
			wantMsg: "result out of range",
		},
		{
			name:    "場札が多すぎる",
			mutate:  func(m map[string]any) { m["ly"] = make([]any, MonteBankLayoutSize+1) },
			wantMsg: "layout cards exceeds",
		},
		{
			name:    "賭け金が負",
			mutate:  func(m map[string]any) { m["bt"] = -1 },
			wantMsg: "bet must not be negative",
		},
		{
			name:    "払い戻しが負",
			mutate:  func(m map[string]any) { m["po"] = -1 },
			wantMsg: "payout must not be negative",
		},
		{
			name:    "ラウンド数が範囲外",
			mutate:  func(m map[string]any) { m["rn"] = 0 },
			wantMsg: "round number out of range",
		},
		{
			name:    "棋譜が長すぎる",
			mutate:  func(m map[string]any) { m["al"] = make([]any, monteBankMaxSliceLen+1) },
			wantMsg: "action log too long",
		},
		{
			name:    "設定が範囲外",
			mutate:  func(m map[string]any) { m["cf"] = map[string]any{"c": 1, "b": 50} },
			wantMsg: "initial chips out of range",
		},
		{
			// **添字としては正当でも、場札の外を指していれば駄目。**
			name:    "賭けた位置が場札の外",
			mutate:  func(m map[string]any) { m["pk"] = MonteBankLayoutSize },
			wantMsg: "pick out of range",
		},
		{
			name:    "賭けた位置が -1 未満",
			mutate:  func(m map[string]any) { m["pk"] = -2 },
			wantMsg: "pick out of range",
		},
		{
			// **勝ったのに払い戻しが 0 の保存を弾く。** 片方だけ書き換えた形。
			name:    "勝ちなのに払い戻しが無い",
			mutate:  func(m map[string]any) { m["po"] = 0 },
			wantMsg: "winning round paid nothing",
		},
		{
			name: "負けなのに払い戻しがある",
			mutate: func(m map[string]any) {
				m["rs"] = int(MonteBankResultLose)
				m["po"] = 100
			},
			wantMsg: "losing round paid",
		},
		{
			name:    "決着したのに決着値が無い",
			mutate:  func(m map[string]any) { m["rs"] = int(MonteBankResultNone) },
			wantMsg: "settled with no result",
		},
		{
			name:    "決着したのに賭けた札が無い",
			mutate:  func(m map[string]any) { m["pk"] = -1 },
			wantMsg: "settled with no bet placed",
		},
		{
			name:    "決着したのに賭け金が 0",
			mutate:  func(m map[string]any) { m["bt"] = 0 },
			wantMsg: "settled with a zero bet",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// 勝ちの局面を土台にする (払い戻しの検査に要る)。
			g := mbPlayedBoard(t)
			for g.GetResult() != MonteBankResultWin {
				g = mbStaged(t,
					mbCard(CardDesignSpade, 1), mbCard(CardDesignHeart, 2),
					mbCard(CardDesignClover, 3), mbCard(CardDesignDiamond, 4))
				mbStackNext(g, mbCard(CardDesignHeart, 7))
				require.NoError(t, g.PlaceBet(1, 50))
			}
			err := mbTamper(t, g, tt.mutate)
			require.Error(t, err, "改竄した保存が通ってしまった")
			assert.ErrorContains(t, err, tt.wantMsg)
		})
	}
}

// **賭ける前の盤面にも整合の規則がある。**
func TestMonteBank_RejectsInconsistentBettingPhase(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		mutate  func(m map[string]any)
		wantMsg string
	}{
		{
			name:    "賭ける前にゲートが開いている",
			mutate:  func(m map[string]any) { m["gt"] = map[string]any{"d": 0, "v": 1} },
			wantMsg: "gate is open before the bet",
		},
		{
			name:    "賭ける前に賭け金が乗っている",
			mutate:  func(m map[string]any) { m["bt"] = 50 },
			wantMsg: "bet is recorded before",
		},
		{
			name:    "賭ける前に場札が足りない",
			mutate:  func(m map[string]any) { m["ly"] = []any{} },
			wantMsg: "layout cards in the betting phase",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := mbTamper(t, newMonteBankForTest(t), tt.mutate)
			require.Error(t, err, "改竄した保存が通ってしまった")
			assert.ErrorContains(t, err, tt.wantMsg)
		})
	}
}

// **プレイヤーが欠けた保存を弾く。**
func TestMonteBank_RejectsMissingPlayer(t *testing.T) {
	t.Parallel()
	err := mbTamper(t, newMonteBankForTest(t), func(m map[string]any) { m["pl"] = nil })
	require.Error(t, err)
	assert.ErrorContains(t, err, "player is missing")
}

// **負の残高を弾く。**
func TestMonteBank_RejectsNegativeChips(t *testing.T) {
	t.Parallel()
	err := mbTamper(t, newMonteBankForTest(t), func(m map[string]any) {
		m["pl"] = map[string]any{"c": -1}
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "chips must not be negative")
}

// **山が無い保存でも落ちない。** 復元後に引けなくなるとそこで固まる。
func TestMonteBank_MissingDeckIsRebuilt(t *testing.T) {
	t.Parallel()
	g := newMonteBankForTest(t)
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	m["dk"] = nil
	tampered, err := json.Marshal(m)
	require.NoError(t, err)

	restored := new(MonteBank)
	require.NoError(t, json.Unmarshal(tampered, restored))
	assert.Positive(t, restored.GetRemainingCards(), "山が空のまま復元された")
}

// **壊れた JSON はそのまま返す。**
func TestMonteBank_RejectsMalformedJSON(t *testing.T) {
	t.Parallel()
	assert.Error(t, json.Unmarshal([]byte(`{`), new(MonteBank)))
	assert.Error(t, json.Unmarshal([]byte(`{"cf":"nope"}`), new(MonteBankConfig)))
	assert.Error(t, json.Unmarshal([]byte(`{"c":"nope"}`), new(MonteBankPlayer)))
}
