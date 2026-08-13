//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bbPlayedBoard は何手か進めた局面を返す (往復と改竄の土台)。
func bbPlayedBoard(t *testing.T) *BaseballPoker {
	t.Helper()
	for range 50 {
		g := newBaseballForTest(t)
		for steps := 0; g.GetStreet() < 2 && g.GetPhase() != BaseballPhaseShowdown &&
			!g.GetGameEndFlag(); steps++ {
			require.Less(t, steps, 200)
			switch {
			case g.IsHumanBuying():
				require.NoError(t, g.AnswerBuyIn(BaseballBuyPay))
			case g.IsHumanTurn():
				if err := g.PlayerAction(BaseballActionCheck, 0); err != nil {
					require.NoError(t, g.PlayerAction(BaseballActionCall, 0))
				}
			default:
				g.CpuPlay()
			}
		}
		if g.GetStreet() >= 2 && g.GetPhase() == BaseballPhaseBetting {
			return g
		}
	}
	t.Fatalf("50 回配っても 2 ストリート進んだ局面が出なかった")
	return nil
}

// bbTamper は保存 JSON を 1 か所だけ書き換えて復元させる。
//
// **本物の局面を 1 か所だけ壊す。** 手で組んだ最小の JSON を投げると、
// 検査の順序次第で狙った枝の手前で落ちて、通ったつもりになる。
func bbTamper(t *testing.T, g *BaseballPoker, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)
	tampered, err := json.Marshal(m)
	require.NoError(t, err)
	return json.Unmarshal(tampered, new(BaseballPoker))
}

// --- 往復 ---

func TestBaseballPoker_RoundTrip(t *testing.T) {
	g := bbPlayedBoard(t)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(BaseballPoker)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetStreet(), restored.GetStreet())
	assert.Equal(t, g.GetPot(), restored.GetPot())
	assert.Equal(t, g.GetTurnSeat(), restored.GetTurnSeat())
	assert.Equal(t, g.GetBuyerSeat(), restored.GetBuyerSeat())
	assert.Equal(t, g.GetHandNumber(), restored.GetHandNumber())
	assert.Equal(t, g.GetRemainingCards(), restored.GetRemainingCards())

	// **向きの列とボーナス枚数まで戻る。** ここが落ちると伏せ札が公開される。
	require.Len(t, restored.GetPlayers(), len(g.GetPlayers()))
	for i, p := range g.GetPlayers() {
		r := restored.GetPlayers()[i]
		assert.Equal(t, p.GetChips(), r.GetChips(), "席 %d のチップ", i)
		assert.Equal(t, len(p.GetCards()), len(r.GetCards()), "席 %d の枚数", i)
		assert.Equal(t, p.GetFaceUp(), r.GetFaceUp(), "席 %d の向き", i)
		assert.Equal(t, p.GetBonusCards(), r.GetBonusCards(), "席 %d のボーナス枚数", i)
	}
}

// **1 手ごとに往復させる。** 手書きの局面では出ない書き込み側の違反を拾う。
func TestBaseballPoker_RoundTripEveryMove(t *testing.T) {
	g := newBaseballForTest(t)
	for steps := 0; g.GetPhase() != BaseballPhaseShowdown && !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 400)

		data, err := json.Marshal(g)
		require.NoError(t, err, "%d 手目の保存", steps)
		require.NoError(t, json.Unmarshal(data, new(BaseballPoker)),
			"%d 手目の保存が自分で復元できない: %s", steps, data)

		switch {
		case g.IsHumanBuying():
			require.NoError(t, g.AnswerBuyIn(BaseballBuyPay))
		case g.IsHumanTurn():
			if err := g.PlayerAction(BaseballActionCheck, 0); err != nil {
				require.NoError(t, g.PlayerAction(BaseballActionCall, 0))
			}
		default:
			g.CpuPlay()
		}
	}
}

// --- 改竄 ---

func TestBaseballPoker_RejectsTamperedSaves(t *testing.T) {
	g := bbPlayedBoard(t)

	for _, tc := range []struct {
		name   string
		mutate func(m map[string]any)
		want   string
	}{
		{"phase out of range", func(m map[string]any) { m["ph"] = 99 }, "phase out of range"},
		{"negative phase", func(m map[string]any) { m["ph"] = -1 }, "phase out of range"},
		{"street out of range", func(m map[string]any) { m["st"] = 99 }, "street out of range"},
		{"negative street", func(m map[string]any) { m["st"] = -1 }, "street out of range"},
		{"turn out of range", func(m map[string]any) { m["tu"] = 99 }, "turn seat out of range"},
		{"negative turn", func(m map[string]any) { m["tu"] = -1 }, "turn seat out of range"},
		{"negative pot", func(m map[string]any) { m["po"] = -1 }, "must not be negative"},
		{"negative buy cost", func(m map[string]any) { m["bc"] = -1 }, "must not be negative"},
		{"buyer out of range", func(m map[string]any) { m["by"] = 99 }, "buyer seat out of range"},
		{"buyer below -1", func(m map[string]any) { m["by"] = -2 }, "buyer seat out of range"},
		{
			// **範囲検査では通ってしまう組合せ。** ベット中に買い手が立っている
			// 保存は席番号としては正しいが、その席を誰も動かさないまま止まる。
			"buyer set outside the buy-in phase",
			func(m map[string]any) { m["by"] = 0 },
			"buyer is set outside the buy-in phase",
		},
		{
			// 逆向き。買い増しフェーズなのに買い手がいない。
			"buy-in phase with no buyer",
			func(m map[string]any) { m["ph"] = int(BaseballPhaseBuyIn) },
			"buyer is set outside the buy-in phase",
		},
		{
			"seat count does not match the config",
			func(m map[string]any) {
				players, _ := m["pl"].([]any)
				m["pl"] = players[:len(players)-1]
			},
			"seat count does not match the config",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := bbTamper(t, g, tc.mutate)
			require.Error(t, err, "改竄した保存が復元できてしまった")
			assert.ErrorContains(t, err, tc.want)
		})
	}
}

// **向きの列が手札とずれた保存は通さない。** 枚数の範囲検査だけでは
// 素通りし、伏せ札が公開されるか表札が隠れる ── どちらも症状が出ないまま
// 勝負だけが静かに変わる。
func TestBaseballPokerPlayer_RejectsAMismatchedFaceUpList(t *testing.T) {
	g := bbPlayedBoard(t)
	err := bbTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		seat, _ := players[0].(map[string]any)
		faceUp, _ := seat["fu"].([]any)
		seat["fu"] = faceUp[:len(faceUp)-1]
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "face-up flags do not match")
}

// **ボーナス枚数は手札の枚数とも整合していなければならない。**
func TestBaseballPokerPlayer_RejectsAnImpossibleBonusCount(t *testing.T) {
	g := bbPlayedBoard(t)

	for _, tc := range []struct {
		name  string
		value any
		want  string
	}{
		{"negative", -1, "bonus card count is out of range"},
		{"above the cap", BaseballMaxBonusCards + 1, "bonus card count is out of range"},
		{"more bonuses than cards allow", BaseballMaxBonusCards, "bonus card count is out of range"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := bbTamper(t, g, func(m map[string]any) {
				players, _ := m["pl"].([]any)
				seat, _ := players[0].(map[string]any)
				seat["bn"] = tc.value
				if tc.name == "more bonuses than cards allow" {
					// 手札を最小にして、ボーナスだけ上限まで立てる。
					cards, _ := seat["cd"].([]any)
					faceUp, _ := seat["fu"].([]any)
					seat["cd"] = cards[:3]
					seat["fu"] = faceUp[:3]
				}
			})
			require.Error(t, err, "%s: 改竄した保存が復元できてしまった", tc.name)
			assert.ErrorContains(t, err, tc.want)
		})
	}
}

func TestBaseballPokerPlayer_RejectsAnOversizedHand(t *testing.T) {
	g := bbPlayedBoard(t)
	err := bbTamper(t, g, func(m map[string]any) {
		players, _ := m["pl"].([]any)
		seat, _ := players[0].(map[string]any)
		cards, _ := seat["cd"].([]any)
		faceUp, _ := seat["fu"].([]any)
		for len(cards) <= BaseballBaseCards+BaseballMaxBonusCards {
			cards = append(cards, cards[0])
			faceUp = append(faceUp, false)
		}
		seat["cd"], seat["fu"] = cards, faceUp
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "more cards than the deal allows")
}

func TestBaseballPokerPlayer_RejectsNegativeSeatValues(t *testing.T) {
	g := bbPlayedBoard(t)

	for _, tc := range []struct{ key, want string }{
		{"c", "chips must not be negative"},
		{"b", "bet must not be negative"},
	} {
		err := bbTamper(t, g, func(m map[string]any) {
			players, _ := m["pl"].([]any)
			seat, _ := players[0].(map[string]any)
			seat[tc.key] = -1
		})
		require.Error(t, err, "%s を負にした保存が通った", tc.key)
		assert.ErrorContains(t, err, tc.want)
	}
}

// **設定も復元時に検証する。** 7 席の保存は配札の前に落とす。
func TestBaseballPokerConfig_Codec(t *testing.T) {
	cfg := DefaultBaseballPokerConfig()
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var restored BaseballPokerConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)

	for _, bad := range []string{
		`{"s":7,"c":1000,"a":10}`,
		`{"s":1,"c":1000,"a":10}`,
		`{"s":4,"c":10,"a":10}`,
		`{"s":4,"c":1000,"a":1}`,
	} {
		assert.Error(t, json.Unmarshal([]byte(bad), new(BaseballPokerConfig)),
			"範囲外の設定が復元できてしまった: %s", bad)
	}
}

// **配列の長さには上限を置く。** 巨大な保存でメモリを食わせない。
func TestBaseballPoker_RejectsOversizedSlices(t *testing.T) {
	g := bbPlayedBoard(t)
	err := bbTamper(t, g, func(m map[string]any) {
		log := make([]any, baseballMaxSliceLen+1)
		for i := range log {
			log[i] = map[string]any{"t": i, "p": 0, "a": "x", "d": "", "c": nil}
		}
		m["al"] = log
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "slice too long")
}
