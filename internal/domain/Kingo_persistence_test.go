//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kingoPlayedBoard は決着まで進めた卓を返す (往復と改竄の土台)。
func kingoPlayedBoard(t *testing.T) *Kingo {
	t.Helper()
	g := newKingoForTest(t)
	kingoPlayRound(t, g)
	require.Equal(t, KingoPhaseResult, g.GetPhase())
	return g
}

// kingoTamper は保存 JSON を 1 か所だけ書き換えて復元させる。
func kingoTamper(t *testing.T, g *Kingo, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	return json.Unmarshal(tampered, new(Kingo))
}

// --- 往復 ---

func TestKingo_RoundTrip(t *testing.T) {
	g := kingoPlayedBoard(t)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(Kingo)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetBankerSeat(), restored.GetBankerSeat())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetRemainingCards(), restored.GetRemainingCards())
	require.Len(t, restored.GetPlayers(), len(g.GetPlayers()))
	for i, p := range g.GetPlayers() {
		r := restored.GetPlayers()[i]
		assert.Equal(t, p.GetChips(), r.GetChips(), "席 %d のチップ", i)
		assert.Equal(t, p.GetBet(), r.GetBet(), "席 %d の張り", i)
		assert.Equal(t, p.GetRank(), r.GetRank(), "席 %d の役", i)
	}
}

// **1 手ごとに往復させる。** 手書きの局面では出ない書き込み側の違反を拾う。
func TestKingo_RoundTripEveryMove(t *testing.T) {
	g := newKingoForTest(t)
	for steps := 0; !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 200)

		data, err := json.Marshal(g)
		require.NoError(t, err, "%d 手目の保存", steps)
		require.NoError(t, json.Unmarshal(data, new(Kingo)),
			"%d 手目の保存が自分で復元できない: %s", steps, data)

		kingoPlayRound(t, g)
		if g.GetGameEndFlag() {
			break
		}
		require.NoError(t, g.NextRound())
	}
}

// --- 改竄 ---

func TestKingo_RejectsTamperedSaves(t *testing.T) {
	g := kingoPlayedBoard(t)

	for _, tc := range []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"phase out of range", func(m map[string]any) { m["ph"] = 99 }, "phase out of range"},
		{"negative phase", func(m map[string]any) { m["ph"] = -1 }, "phase out of range"},
		{"banker out of range", func(m map[string]any) { m["bk"] = 99 }, "banker seat out of range"},
		{"negative banker", func(m map[string]any) { m["bk"] = -1 }, "banker seat out of range"},
		{"round number zero", func(m map[string]any) { m["rn"] = 0 }, "round number out of range"},
		{
			// **設定を超えたラウンド番号は範囲検査を通ってしまう。**
			// 「終わっているのに続く卓」で、親の回りも設定の意味も壊れる。
			"round number beyond the configured rounds",
			func(m map[string]any) { m["rn"] = KingoMaxRounds + 1 },
			"round number exceeds",
		},
		{
			"seat count does not match the config",
			func(m map[string]any) {
				players, _ := m["pl"].([]any)
				m["pl"] = players[:len(players)-1]
			},
			"seat count does not match the config",
		},
		{
			"deck larger than a kabufuda deck",
			func(m map[string]any) {
				deck, _ := m["dk"].([]any)
				for len(deck) <= KingoDeckSize {
					deck = append(deck, map[string]any{"d": 1, "v": 3, "w": false})
				}
				m["dk"] = deck
			},
			"more cards than a kabufuda deck",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := kingoTamper(t, g, tc.mutate)
			require.Error(t, err, "改竄した保存が復元できてしまった")
			assert.ErrorContains(t, err, tc.want)
		})
	}
}

// **控えの役が手札と合わない保存は通さない。**
//
// 役は `GetRank` が手札から毎回数えるので、控えがずれていても勝敗は正しく
// 出る ── つまり**画面にだけ嘘の役が出る**保存が作れてしまう。範囲検査では
// 絶対に見つからない。
func TestKingoPlayer_RejectsARankThatDoesNotMatchTheHand(t *testing.T) {
	g := kingoPlayedBoard(t)

	err := kingoTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		seat, _ := players[0].(map[string]any)
		// いまの役と必ず違う値にする。
		cur := int(seat["rk"].(float64))
		seat["rk"] = (cur + 1) % (int(KingoRankMax) + 1)
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "does not match the hand")

	// 範囲外も弾く。
	err = kingoTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		players[0].(map[string]any)["rk"] = 99
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "does not match the hand")
}

func TestKingoPlayer_RejectsAPartialHand(t *testing.T) {
	g := kingoPlayedBoard(t)
	err := kingoTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		seat, _ := players[0].(map[string]any)
		cards, _ := seat["cd"].([]any)
		seat["cd"] = cards[:len(cards)-1]
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "must hold three cards")
}

func TestKingoPlayer_RejectsNegativeSeatValues(t *testing.T) {
	g := kingoPlayedBoard(t)
	for _, tc := range []struct{ key, want string }{
		{"c", "chips must not be negative"},
		{"b", "bet must not be negative"},
	} {
		err := kingoTamper(t, g, func(m map[string]any) {
			players, _ := m["pl"].([]any)
			players[0].(map[string]any)[tc.key] = -1
		})
		require.Error(t, err, "%s を負にした保存が通った", tc.key)
		assert.ErrorContains(t, err, tc.want)
	}
}

// **設定も復元時に検証する。** 親が回らない卓を保存から作れない。
func TestKingoConfig_Codec(t *testing.T) {
	cfg := DefaultKingoConfig()
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var restored KingoConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)

	for _, bad := range []string{
		`{"s":5,"c":1000,"b":10,"r":4}`,
		`{"s":1,"c":1000,"b":10,"r":10}`,
		`{"s":4,"c":10,"b":10,"r":10}`,
		`{"s":4,"c":1000,"b":15,"r":10}`,
	} {
		assert.Error(t, json.Unmarshal([]byte(bad), new(KingoConfig)),
			"範囲外の設定が復元できてしまった: %s", bad)
	}
}

func TestKingo_RejectsOversizedSlices(t *testing.T) {
	g := kingoPlayedBoard(t)
	err := kingoTamper(t, g, func(m map[string]any) {
		log := make([]any, kingoMaxSliceLen+1)
		for i := range log {
			log[i] = map[string]any{"t": i, "p": 0, "a": "x", "d": "", "c": nil}
		}
		m["al"] = log
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "slice too long")
}
