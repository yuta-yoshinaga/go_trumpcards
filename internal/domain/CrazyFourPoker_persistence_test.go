//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// crazyFourPokerTampered は本物の盤面を JSON にしてから 1 か所だけ壊し、復元を試みる。
func crazyFourPokerTampered(t *testing.T, base *CrazyFourPoker, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back CrazyFourPoker
	return json.Unmarshal(broken, &back)
}

// crazyFourPokerMidRound は賭け済み・配札済みで判断待ちの盤面を返す。
func crazyFourPokerMidRound(t *testing.T) *CrazyFourPoker {
	t.Helper()
	return c4pStaged(t, 50, 20, c4pAcePair(), c4pKingHigh())
}

// **到達できる盤面はすべて往復する。** これが負のコントロール。
func TestCrazyFourPoker_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	stages := []struct {
		name  string
		build func(t *testing.T) *CrazyFourPoker
	}{
		{"賭ける前", func(t *testing.T) *CrazyFourPoker { return newCrazyFourPokerForTest(t) }},
		{"判断待ち", crazyFourPokerMidRound},
		{"3 倍で決着", func(t *testing.T) *CrazyFourPoker {
			g := crazyFourPokerMidRound(t)
			require.NoError(t, g.Play(3))
			return g
		}},
		{"同額で決着", func(t *testing.T) *CrazyFourPoker {
			g := c4pStaged(t, 50, 0, c4pKingHigh(), c4pJackHigh())
			require.NoError(t, g.Play(1))
			return g
		}},
		{"降りた後", func(t *testing.T) *CrazyFourPoker {
			g := crazyFourPokerMidRound(t)
			require.NoError(t, g.Fold())
			return g
		}},
		{"次ラウンド", func(t *testing.T) *CrazyFourPoker {
			g := crazyFourPokerMidRound(t)
			require.NoError(t, g.Fold())
			require.NoError(t, g.NextRound())
			return g
		}},
	}

	for _, st := range stages {
		t.Run(st.name, func(t *testing.T) {
			g := st.build(t)
			data, err := json.Marshal(g)
			require.NoError(t, err)

			var back CrazyFourPoker
			require.NoError(t, json.Unmarshal(data, &back),
				"書き込み側が codec の不変条件を破った (phase=%d, result=%d)", g.GetPhase(), g.GetResult())

			assert.Equal(t, g.GetPhase(), back.GetPhase())
			assert.Equal(t, g.GetResult(), back.GetResult())
			assert.Equal(t, g.GetAnteBet(), back.GetAnteBet())
			assert.Equal(t, g.GetSuperBet(), back.GetSuperBet())
			assert.Equal(t, g.GetQueensUpBet(), back.GetQueensUpBet())
			assert.Equal(t, g.GetPlayBet(), back.GetPlayBet())
			assert.Equal(t, g.GetPlayMultiplier(), back.GetPlayMultiplier())
			assert.Equal(t, g.GetChips(), back.GetChips())
			assert.Equal(t, g.GetPlayerHandRank(), back.GetPlayerHandRank())
			assert.Equal(t, g.GetDealerHandRank(), back.GetDealerHandRank())
			assert.Equal(t, g.GetRoundNumber(), back.GetRoundNumber())
		})
	}
}

// **改竄した保存データは、壊した場所を名指しして弾く。**
func TestCrazyFourPoker_UnmarshalRejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{
			name:   "プレイヤーが欠けている",
			mutate: func(m map[string]any) { m["pl"] = nil },
			want:   "the player is missing",
		},
		{
			name:   "設定が範囲外",
			mutate: func(m map[string]any) { m["cf"] = map[string]any{"ic": 5, "da": 50} },
			want:   "chips must be",
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
			name:   "アンティが負",
			mutate: func(m map[string]any) { m["an"] = -10; m["sb"] = -10 },
			want:   "ante must not be negative",
		},
		{
			name:   "ラウンド番号が 0",
			mutate: func(m map[string]any) { m["rn"] = 0 },
			want:   "round number out of range",
		},
		{
			name: "手札が 6 枚ある",
			mutate: func(m map[string]any) {
				card := map[string]any{"d": 1, "v": 3, "w": false}
				m["ph5"] = []any{card, card, card, card, card, card}
			},
			want: "a hand holds more than 5 cards",
		},
		{
			name: "最良手が 5 枚ある",
			mutate: func(m map[string]any) {
				card := map[string]any{"d": 1, "v": 3, "w": false}
				m["pb"] = []any{card, card, card, card, card}
			},
			want: "a best hand holds more than 4 cards",
		},
		{
			name: "棋譜が長すぎる",
			mutate: func(m map[string]any) {
				log := make([]any, crazyFourPokerMaxSliceLen+1)
				for i := range log {
					log[i] = map[string]any{}
				}
				m["al"] = log
			},
			want: "action log too long",
		},
		{
			// **Super Bonus は必ずアンティと同額。** 賭けていない額に配当が付く。
			name:   "Super Bonus がアンティと違う",
			mutate: func(m map[string]any) { m["sb"] = 999 },
			want:   "does not match the ante",
		},
		{
			name: "賭ける前なのに配られている",
			mutate: func(m map[string]any) {
				m["ph"] = int(CrazyFourPokerPhaseBet)
				m["rs"] = int(CrazyFourPokerResultNone)
			},
			want: "cards are dealt before the ante is placed",
		},
		{
			name:   "アンティ無しのプレイベット",
			mutate: func(m map[string]any) { m["an"] = 0; m["sb"] = 0; m["pbt"] = 100; m["pm"] = 0 },
			want:   "a play bet without an ante",
		},
		{
			name:   "倍率とプレイベット額が合わない",
			mutate: func(m map[string]any) { m["pm"] = 2; m["pbt"] = 999 },
			want:   "is not 2 x the ante",
		},
		{
			name:   "倍率が範囲外",
			mutate: func(m map[string]any) { m["pm"] = 9; m["pbt"] = 450 },
			want:   "play multiplier out of range",
		},
		{
			// **決着はラウンド終了フェーズでしか持てない。**
			name:   "決着済みなのに判断フェーズ",
			mutate: func(m map[string]any) { m["rs"] = int(CrazyFourPokerResultWin) },
			want:   "the round is decided but the phase is",
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
			err := crazyFourPokerTampered(t, crazyFourPokerMidRound(t), tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want, "別のガードが先に落としている可能性がある")
		})
	}
}

// **エース未満なのに 3 倍が置かれた保存データを弾く。**
//
// 倍率と金額の辻褄は合っているので、範囲チェックだけでは通ってしまう。手役まで
// 見ないと「弱い手で 3 倍」という有利すぎる盤面が復元できる。
func TestCrazyFourPoker_UnmarshalRejectsUnearnedMultiplier(t *testing.T) {
	weak := c4pStaged(t, 50, 0, c4pKingHigh(), c4pJackHigh())
	require.NoError(t, weak.Play(1))

	err := crazyFourPokerTampered(t, weak, func(m map[string]any) {
		m["pm"] = 3
		m["pbt"] = 150
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "needs a pair of aces or better")

	// エースのペアなら同じ改竄が通る (負のコントロール)。
	strong := c4pStaged(t, 50, 0, c4pAcePair(), c4pKingHigh())
	require.NoError(t, strong.Play(1))
	assert.NoError(t, crazyFourPokerTampered(t, strong, func(m map[string]any) {
		m["pm"] = 3
		m["pbt"] = 150
	}), "エース以上なら 3 倍は正当なはず")
}

func TestCrazyFourPoker_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var g CrazyFourPoker
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`), &g))

	var c CrazyFourPokerConfig
	assert.Error(t, json.Unmarshal([]byte(`{"ic":`), &c))

	var p CrazyFourPokerPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":`), &p))
}

func TestCrazyFourPokerConfig_UnmarshalValidates(t *testing.T) {
	t.Parallel()

	var bad CrazyFourPokerConfig
	assert.Error(t, json.Unmarshal([]byte(`{"ic":1,"da":50}`), &bad))

	var ok CrazyFourPokerConfig
	require.NoError(t, json.Unmarshal([]byte(`{"ic":500,"da":20}`), &ok))
	assert.Equal(t, 500, ok.InitialChips)
	assert.Equal(t, 20, ok.DefaultAnte)
}

func TestCrazyFourPokerPlayer_RoundTrip(t *testing.T) {
	t.Parallel()

	p := NewCrazyFourPokerPlayer(250)
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var back CrazyFourPokerPlayer
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, 250, back.GetChips())

	assert.True(t, back.SubtractChips(50))
	assert.Equal(t, 200, back.GetChips())
	assert.False(t, back.SubtractChips(1000), "足りない額は引けない")
	assert.Equal(t, 200, back.GetChips(), "失敗した減算が額を変えた")
	back.AddChips(25)
	assert.Equal(t, 225, back.GetChips())
}

// **デッキが欠けていても落ちない。** 新しいデッキで補う。
func TestCrazyFourPoker_UnmarshalFillsAMissingDeck(t *testing.T) {
	base := crazyFourPokerMidRound(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	delete(m, "tc")
	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back CrazyFourPoker
	require.NoError(t, json.Unmarshal(broken, &back))
	assert.Equal(t, 52, back.GetRemainingCards())
}

// 復元した盤面はそのまま続けられる。**復元できることと遊べることは別。**
func TestCrazyFourPoker_RestoredGameKeepsPlaying(t *testing.T) {
	base := crazyFourPokerMidRound(t)
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var back CrazyFourPoker
	require.NoError(t, json.Unmarshal(data, &back))

	require.NoError(t, back.Play(back.MaxPlayMultiplier()))
	assert.Equal(t, CrazyFourPokerPhaseResult, back.GetPhase())
	require.NoError(t, back.NextRound())
	assert.Equal(t, CrazyFourPokerPhaseBet, back.GetPhase())
}

// **同じ壊れた保存は、いつ読んでも同じ理由で落ちる。**
//
// 額の検査を map の range で回していたため、2 つ以上が同時に負の保存では
// 走査順が実行ごとに変わり、返るエラーが毎回違っていた。利用者から見れば
// 「同じ保存を読み込むたびに別の理由を言われる」で、テストから見れば数回に
// 1 回だけ落ちるフレーク (CI で実際に落ちた)。
//
// 1 回の実行では偶然そろうことがあるので、**何度も読み直して全部同じ**で
// あることを見る。
func TestCrazyFourPoker_ReportsTheSameReasonEveryTime(t *testing.T) {
	first := ""
	for i := range 50 {
		err := crazyFourPokerTampered(t, crazyFourPokerMidRound(t), func(m map[string]any) {
			// アンティとスーパーボーナスは一致していなければならないので、
			// 両方を負にするしかない ── ここで順序が揺れていた。
			m["an"] = -10
			m["sb"] = -10
		})
		require.Error(t, err, "改竄した保存データが素通しした")
		if i == 0 {
			first = err.Error()
			assert.Contains(t, first, "ante must not be negative",
				"最初に報告されるのは宣言順のいちばん上であるべき")
			continue
		}
		require.Equal(t, first, err.Error(),
			"%d 回目で違う理由が返った (走査順が固定されていない)", i+1)
	}
}
