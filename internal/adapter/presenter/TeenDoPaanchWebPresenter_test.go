//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTeenDoPaanchForWeb(t *testing.T) *domain.TeenDoPaanch {
	t.Helper()
	g := domain.NewDefaultTeenDoPaanch()
	g.Reset()
	return g
}

func decodeTeenDoPaanch(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestTeenDoPaanchWebPresenterOutput(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)
	m := decodeTeenDoPaanch(t, p.Output(newTeenDoPaanchForWeb(t), nil))

	assert.Equal(t, float64(domain.TeenDoPaanchPhaseTrump), m["phase"], "配り直後は切り札の宣言")
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(0), m["trumpSuit"], "まだ宣言されていない")
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(0), m["lastExchange"])

	players := m["players"].([]any)
	require.Len(t, players, domain.TeenDoPaanchPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	// **手札が揃う前に切り札を決めるのが賭けどころ。**
	assert.Equal(t, float64(domain.TeenDoPaanchFirstDeal), human["cardCount"])
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **ノルマは宣言ではなく割り当て。** 3・2・5 が必ず 1 つずつワイヤに載る。
func TestTeenDoPaanchWebPresenterCarriesTheAssignedTargets(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)
	m := decodeTeenDoPaanch(t, p.Output(newTeenDoPaanchForWeb(t), nil))

	got := map[float64]int{}
	for _, pl := range m["players"].([]any) {
		got[pl.(map[string]any)["target"].(float64)]++
	}
	assert.Equal(t, map[float64]int{3: 1, 2: 1, 5: 1}, got)

	five := int(m["fivePlayerIdx"].(float64))
	assert.Equal(t, float64(5), m["players"].([]any)[five].(map[string]any)["target"],
		"切り札を決める席がノルマ 5")
}

// **獲得数と達成回数の両方が要る。** ノルマにあと何トリックかが読めないと打てない。
func TestTeenDoPaanchWebPresenterCarriesProgress(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)
	g := newTeenDoPaanchForWeb(t)
	g.GiveTricksForTest(0, 2)
	g.GetPlayer(0).SetMet(1)

	human := decodeTeenDoPaanch(t, p.Output(g, nil))["players"].([]any)[0].(map[string]any)
	assert.Equal(t, float64(2), human["trickCount"])
	assert.Equal(t, float64(1), human["met"])
}

// 切り札は未宣言と確定の両側を踏む。
func TestTeenDoPaanchWebPresenterMessageTracksTheTrumpCall(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)

	mine := newTeenDoPaanchForWeb(t)
	mine.SetFivePlayerIdxForTest(0)
	out := decodeTeenDoPaanch(t, p.Output(mine, nil))
	assert.Equal(t, "teendopaanch.trump.choose", out["messageCode"])
	assert.Equal(t, "5", out["messageParams"].(map[string]any)["seen"], "5 枚しか見えていない")

	theirs := newTeenDoPaanchForWeb(t)
	theirs.SetFivePlayerIdxForTest(1)
	assert.Equal(t, "teendopaanch.trump.wait", decodeTeenDoPaanch(t, p.Output(theirs, nil))["messageCode"])

	declared := newTeenDoPaanchForWeb(t)
	declared.SetFivePlayerIdxForTest(0)
	require.NoError(t, declared.DeclareTrump(domain.CardDesignHeart))
	out = decodeTeenDoPaanch(t, p.Output(declared, nil))
	assert.Equal(t, float64(domain.CardDesignHeart), out["trumpSuit"])
	assert.Equal(t, "teendopaanch.play", out["messageCode"])
	// **多く取ってもうれしくない。** ノルマと獲得数を毎回添える。
	assert.Contains(t, out["messageParams"].(map[string]any), "target")
	assert.Contains(t, out["messageParams"].(map[string]any), "took")
}

// **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
func TestTeenDoPaanchWebPresenterReportsTheExchange(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)
	g := newTeenDoPaanchForWeb(t)
	g.SetFivePlayerIdxForTest(0)
	require.NoError(t, g.DeclareTrump(domain.CardDesignSpade))
	g.GiveTricksForTest(0, 7)
	g.GiveTricksForTest(1, 1)
	g.GiveTricksForTest(2, 2)
	g.FinishRoundForTest()
	g.NextRound()
	require.NoError(t, g.DeclareTrump(domain.CardDesignSpade))

	assert.Equal(t, float64(2), decodeTeenDoPaanch(t, p.Output(g, nil))["lastExchange"])
}

func TestTeenDoPaanchWebPresenterRoundEnd(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)
	g := newTeenDoPaanchForWeb(t)
	require.NoError(t, g.DeclareTrump(domain.CardDesignSpade))
	g.FinishRoundForTest()

	out := decodeTeenDoPaanch(t, p.Output(g, nil))
	assert.Equal(t, float64(domain.TeenDoPaanchPhaseRoundEnd), out["phase"])
	assert.Equal(t, "teendopaanch.roundEnd", out["messageCode"])
}

func TestTeenDoPaanchWebPresenterGameEnd(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)
	for _, tc := range []struct {
		name string
		met  [domain.TeenDoPaanchPlayerCnt]int
		code string
		want float64
	}{
		{"you", [3]int{3, 1, 0}, "teendopaanch.result.you", 0},
		{"cpu", [3]int{0, 3, 1}, "teendopaanch.result.cpu", 1},
		{"tie", [3]int{2, 2, 2}, "teendopaanch.result.tie", -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTeenDoPaanchForWeb(t)
			for i, n := range tc.met {
				g.GetPlayer(i).SetMet(n)
			}
			g.FinishGameForTest()

			out := decodeTeenDoPaanch(t, p.Output(g, nil))
			assert.True(t, out["gameEndFlag"].(bool))
			assert.Equal(t, tc.want, out["winnerIdx"])
			assert.Equal(t, tc.code, out["messageCode"])
		})
	}
}

func TestTeenDoPaanchWebPresenterSurfacesErrors(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)
	g := newTeenDoPaanchForWeb(t)
	err := g.DeclareTrump(99)
	require.Error(t, err)

	out := decodeTeenDoPaanch(t, p.Output(g, err))
	assert.Equal(t, err.Error(), out["message"])
	assert.Empty(t, out["messageCode"], "エラー時はコードを立てない")
}

func TestTeenDoPaanchWebPresenterHintAndLog(t *testing.T) {
	p := new(TeenDoPaanchWebPresenter)
	g := newTeenDoPaanchForWeb(t)
	g.SetFivePlayerIdxForTest(0)

	// **宣言フェーズの助言はスートを指し、札は指さない。**
	hint := decodeTeenDoPaanch(t, p.HintOutput(g))["hint"].(map[string]any)
	assert.NotContains(t, hint, "cardIndex")
	assert.Positive(t, hint["suit"])

	require.NoError(t, g.DeclareTrump(domain.CardDesignHeart))
	g.SetCurrentPlayerIdxForTest(0)
	hint = decodeTeenDoPaanch(t, p.HintOutput(g))["hint"].(map[string]any)
	idx := int(hint["cardIndex"].(float64))
	assert.Contains(t, g.GetValidPlayIndices(0), idx, "勧める札は必ず合法")

	var logOut map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(g)), &logOut))
	assert.Contains(t, logOut, "entries")
}
