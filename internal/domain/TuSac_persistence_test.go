//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tuSacPlayedBoard は何手か進めた局面を返す (往復と改竄の土台)。
func tuSacPlayedBoard(t *testing.T) *TuSac {
	t.Helper()
	g := newTuSacForTest(t)
	require.True(t, g.IsHumanTurn())
	require.NoError(t, g.Draw(false))
	// 出せる組み合わせがあれば 1 つ出しておく (メルドの検証を効かせるため)。
	if h := g.GetHint(); h != nil && h.Action == "meld" {
		require.NoError(t, g.Meld(h.Indexes))
	}
	return g
}

// tuSacTamper は保存 JSON を 1 か所だけ書き換えて復元させる。
func tuSacTamper(t *testing.T, g *TuSac, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	return json.Unmarshal(tampered, new(TuSac))
}

// --- 往復 ---

func TestTuSac_RoundTrip(t *testing.T) {
	g := tuSacPlayedBoard(t)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(TuSac)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetTurnSeat(), restored.GetTurnSeat())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, g.GetDiscardCount(), restored.GetDiscardCount())
	assert.Equal(t, g.GetWentOutSeat(), restored.GetWentOutSeat())
	require.Len(t, restored.GetPlayers(), len(g.GetPlayers()))
	for i, p := range g.GetPlayers() {
		r := restored.GetPlayers()[i]
		assert.Len(t, r.GetCards(), len(p.GetCards()), "席 %d の枚数", i)
		assert.Len(t, r.GetMelds(), len(p.GetMelds()), "席 %d のメルド数", i)
		assert.Equal(t, p.MeldPoints(), r.MeldPoints(), "席 %d のメルド点", i)
	}
}

// **1 手ごとに往復させる。** 手書きの局面では出ない書き込み側の違反を拾う。
func TestTuSac_RoundTripEveryMove(t *testing.T) {
	g := newTuSacForTest(t)
	for steps := 0; g.GetPhase() != TuSacPhaseRoundEnd && !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 600)

		data, err := json.Marshal(g)
		require.NoError(t, err, "%d 手目の保存", steps)
		require.NoError(t, json.Unmarshal(data, new(TuSac)),
			"%d 手目の保存が自分で復元できない", steps)

		if !g.IsHumanTurn() {
			break
		}
		switch g.GetPhase() {
		case TuSacPhaseDraw:
			require.NoError(t, g.Draw(false))
		case TuSacPhaseDiscard:
			require.NoError(t, g.Discard(0))
		}
	}
}

// --- 改竄 ---

func TestTuSac_RejectsTamperedSaves(t *testing.T) {
	g := tuSacPlayedBoard(t)

	for _, tc := range []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"phase out of range", func(m map[string]any) { m["ph"] = 99 }, "phase out of range"},
		{"negative phase", func(m map[string]any) { m["ph"] = -1 }, "phase out of range"},
		{"turn out of range", func(m map[string]any) { m["tu"] = 99 }, "turn seat out of range"},
		{"negative turn", func(m map[string]any) { m["tu"] = -1 }, "turn seat out of range"},
		{"round number zero", func(m map[string]any) { m["rn"] = 0 }, "round number out of range"},
		{
			"round number beyond the configured rounds",
			func(m map[string]any) { m["rn"] = TuSacMaxRounds + 1 },
			"round number out of range",
		},
		{
			// **上がった席は実在の席か -1。** 範囲外だと結果が別の席を指す。
			"went-out seat out of range",
			func(m map[string]any) { m["wo"] = 99 },
			"turn seat out of range",
		},
		{"went-out below -1", func(m map[string]any) { m["wo"] = -2 }, "turn seat out of range"},
		{
			"seat count does not match the config",
			func(m map[string]any) {
				players, _ := m["pl"].([]any)
				m["pl"] = players[:len(players)-1]
			},
			"seat count does not match the config",
		},
		{
			"stock larger than the deck",
			func(m map[string]any) {
				stock, _ := m["st"].([]any)
				for len(stock) <= TuSacDeckSize {
					stock = append(stock, map[string]any{"d": 1, "v": 3, "w": false})
				}
				m["st"] = stock
			},
			"more cards than the deck",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tuSacTamper(t, g, tc.mutate)
			require.Error(t, err, "改竄した保存が復元できてしまった")
			assert.ErrorContains(t, err, tc.want)
		})
	}
}

// **出した組み合わせは復元時に組み直して検証する。**
//
// メルドの種別は保存に入っているが、札のほうを書き換えれば「卒 2 枚で 5 点」の
// ような保存が作れる ── 種別の範囲検査だけでは通ってしまい、得点だけが静かに
// 増える。
func TestTuSacPlayer_RejectsAMeldThatIsNotOne(t *testing.T) {
	g := tuSacPlayedBoard(t)
	// メルドを持っている席を探す。
	seat := -1
	for i, p := range g.GetPlayers() {
		if len(p.GetMelds()) > 0 {
			seat = i
			break
		}
	}
	if seat < 0 {
		t.Skip("この配りではメルドが出ていない")
	}

	// 1) 札を 1 枚落として、種別と合わなくする。
	err := tuSacTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		p, _ := players[seat].(map[string]any)
		melds, _ := p["ml"].([]any)
		meld, _ := melds[0].(map[string]any)
		cards, _ := meld["c"].([]any)
		meld["c"] = cards[:len(cards)-1]
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "not a valid combination")

	// 2) 種別だけを高いほうに書き換える (得点だけが増える改竄)。
	err = tuSacTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		p, _ := players[seat].(map[string]any)
		melds, _ := p["ml"].([]any)
		meld, _ := melds[0].(map[string]any)
		cur := int(meld["k"].(float64))
		meld["k"] = cur%int(TuSacMeldKindMax) + 1
		if meld["k"] == cur {
			meld["k"] = int(TuSacMeldKindMax)
		}
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "not a valid combination")

	// 3) 範囲外の種別。
	err = tuSacTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		p, _ := players[seat].(map[string]any)
		melds, _ := p["ml"].([]any)
		melds[0].(map[string]any)["k"] = 99
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "not a valid combination")
}

func TestTuSacPlayer_RejectsAnOversizedHand(t *testing.T) {
	g := tuSacPlayedBoard(t)
	err := tuSacTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		p, _ := players[0].(map[string]any)
		cards, _ := p["cd"].([]any)
		for len(cards) <= TuSacHandSize+1 {
			cards = append(cards, map[string]any{"d": 1, "v": 3, "w": false})
		}
		p["cd"] = cards
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "more cards than the deal allows")
}

func TestTuSacPlayer_RejectsANegativeScore(t *testing.T) {
	g := tuSacPlayedBoard(t)
	err := tuSacTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		players[0].(map[string]any)["s"] = -1
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "score must not be negative")
}

func TestTuSacConfig_Codec(t *testing.T) {
	cfg := DefaultTuSacConfig()
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var restored TuSacConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)

	for _, bad := range []string{
		`{"s":1,"r":5}`,
		`{"s":5,"r":5}`,
		`{"s":4,"r":0}`,
		`{"s":4,"r":99}`,
	} {
		assert.Error(t, json.Unmarshal([]byte(bad), new(TuSacConfig)),
			"範囲外の設定が復元できてしまった: %s", bad)
	}
}

func TestTuSac_RejectsOversizedSlices(t *testing.T) {
	g := tuSacPlayedBoard(t)
	err := tuSacTamper(t, g, func(m map[string]any) {
		log := make([]any, tuSacMaxSliceLen+1)
		for i := range log {
			log[i] = map[string]any{"t": i, "p": 0, "a": "x", "d": "", "c": nil}
		}
		m["al"] = log
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "slice too long")
}
