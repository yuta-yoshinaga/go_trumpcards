//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newSergeantMajorForWeb(t *testing.T) *domain.SergeantMajor {
	t.Helper()
	s := domain.NewDefaultSergeantMajor()
	s.Reset()
	return s
}

func decodeSergeantMajor(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestSergeantMajorWebPresenterOutput(t *testing.T) {
	p := new(SergeantMajorWebPresenter)
	m := decodeSergeantMajor(t, p.Output(newSergeantMajorForWeb(t), nil))

	assert.Equal(t, float64(domain.SergeantMajorPhaseTrump), m["phase"])
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(0), m["trumpSuit"], "まだ宣言されていない")
	assert.Equal(t, float64(domain.SergeantMajorKittySize), m["kittySize"], "キティは 4 枚")
	assert.Equal(t, float64(domain.SergeantMajorKittySize), m["discardCount"])
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(0), m["lastExchange"])

	players := m["players"].([]any)
	require.Len(t, players, domain.SergeantMajorPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.SergeantMajorHandSize), human["cardCount"], "16 枚配る")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **ノルマは席順で決まる。** 8/5/3 が必ず 1 つずつワイヤに載る。
func TestSergeantMajorWebPresenterCarriesTheSeatTargets(t *testing.T) {
	p := new(SergeantMajorWebPresenter)
	m := decodeSergeantMajor(t, p.Output(newSergeantMajorForWeb(t), nil))

	got := map[float64]int{}
	for _, pl := range m["players"].([]any) {
		got[pl.(map[string]any)["target"].(float64)]++
	}
	assert.Equal(t, map[float64]int{8: 1, 5: 1, 3: 1}, got)

	dealer := int(m["dealerIdx"].(float64))
	assert.Equal(t, float64(8), m["players"].([]any)[dealer].(map[string]any)["target"],
		"親がノルマ 8")
}

// **獲得数と累計得点の両方が要る。**
func TestSergeantMajorWebPresenterCarriesProgress(t *testing.T) {
	p := new(SergeantMajorWebPresenter)
	s := newSergeantMajorForWeb(t)
	s.GiveTricksForTest(0, 3)
	s.GetPlayer(0).SetScore(2)

	human := decodeSergeantMajor(t, p.Output(s, nil))["players"].([]any)[0].(map[string]any)
	assert.Equal(t, float64(3), human["trickCount"])
	assert.Equal(t, float64(2), human["score"])
}

// 切り札は未宣言と確定の両側を踏む。
func TestSergeantMajorWebPresenterMessageTracksTheTrumpCall(t *testing.T) {
	p := new(SergeantMajorWebPresenter)

	mine := newSergeantMajorForWeb(t)
	mine.SetDealerIdxForTest(0)
	assert.Equal(t, "sergeantmajor.trump.choose", decodeSergeantMajor(t, p.Output(mine, nil))["messageCode"])

	theirs := newSergeantMajorForWeb(t)
	theirs.SetDealerIdxForTest(1)
	assert.Equal(t, "sergeantmajor.trump.wait", decodeSergeantMajor(t, p.Output(theirs, nil))["messageCode"])

	declared := newSergeantMajorForWeb(t)
	declared.SetDealerIdxForTest(0)
	require.NoError(t, declared.DeclareTrump(domain.CardDesignHeart))
	out := decodeSergeantMajor(t, p.Output(declared, nil))
	assert.Equal(t, float64(domain.CardDesignHeart), out["trumpSuit"])
	assert.Equal(t, float64(0), out["kittySize"], "親が取り込んだ")
	assert.Equal(t, "sergeantmajor.discard.choose", out["messageCode"])
}

// **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
func TestSergeantMajorWebPresenterReportsTheExchange(t *testing.T) {
	p := new(SergeantMajorWebPresenter)
	s := newSergeantMajorForWeb(t)
	s.SetDealerIdxForTest(0)
	require.NoError(t, s.DeclareTrump(domain.CardDesignSpade))
	require.NoError(t, s.DiscardForTest(0, []int{0, 1, 2, 3}))
	s.GiveTricksForTest(0, 10)
	s.GiveTricksForTest(1, 4)
	s.GiveTricksForTest(2, 2)
	s.FinishRoundForTest()
	s.NextRound()
	require.NoError(t, s.DeclareTrump(domain.CardDesignSpade))
	require.NoError(t, s.DiscardForTest(s.GetDealerIdx(), []int{0, 1, 2, 3}))

	assert.Positive(t, decodeSergeantMajor(t, p.Output(s, nil))["lastExchange"])
}

func TestSergeantMajorWebPresenterRoundEnd(t *testing.T) {
	p := new(SergeantMajorWebPresenter)
	s := newSergeantMajorForWeb(t)
	s.SetDealerIdxForTest(0)
	require.NoError(t, s.DeclareTrump(domain.CardDesignSpade))
	require.NoError(t, s.DiscardForTest(0, []int{0, 1, 2, 3}))
	s.FinishRoundForTest()

	out := decodeSergeantMajor(t, p.Output(s, nil))
	assert.Equal(t, float64(domain.SergeantMajorPhaseRoundEnd), out["phase"])
	assert.Equal(t, "sergeantmajor.roundEnd", out["messageCode"])
}

func TestSergeantMajorWebPresenterGameEnd(t *testing.T) {
	p := new(SergeantMajorWebPresenter)
	for _, tc := range []struct {
		name   string
		scores [domain.SergeantMajorPlayerCnt]int
		code   string
		want   float64
	}{
		{"you", [3]int{4, -2, -2}, "sergeantmajor.result.you", 0},
		{"cpu", [3]int{-2, 4, -2}, "sergeantmajor.result.cpu", 1},
		{"tie", [3]int{0, 0, 0}, "sergeantmajor.result.tie", -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSergeantMajorForWeb(t)
			for i, n := range tc.scores {
				s.GetPlayer(i).SetScore(n)
			}
			s.FinishGameForTest()

			out := decodeSergeantMajor(t, p.Output(s, nil))
			assert.True(t, out["gameEndFlag"].(bool))
			assert.Equal(t, tc.want, out["winnerIdx"])
			assert.Equal(t, tc.code, out["messageCode"])
		})
	}
}

func TestSergeantMajorWebPresenterSurfacesErrors(t *testing.T) {
	p := new(SergeantMajorWebPresenter)
	s := newSergeantMajorForWeb(t)
	err := s.DeclareTrump(99)
	require.Error(t, err)

	out := decodeSergeantMajor(t, p.Output(s, err))
	assert.Equal(t, err.Error(), out["message"])
	assert.Empty(t, out["messageCode"])
}

func TestSergeantMajorWebPresenterHintAndLog(t *testing.T) {
	p := new(SergeantMajorWebPresenter)
	s := newSergeantMajorForWeb(t)
	s.SetDealerIdxForTest(0)

	// **宣言の助言はスートを指し、札は指さない。**
	hint := decodeSergeantMajor(t, p.HintOutput(s))["hint"].(map[string]any)
	assert.NotContains(t, hint, "cardIndex")
	assert.Positive(t, hint["suit"])

	require.NoError(t, s.DeclareTrump(domain.CardDesignHeart))
	// **捨て札の助言は複数の札を指す。**
	hint = decodeSergeantMajor(t, p.HintOutput(s))["hint"].(map[string]any)
	assert.Len(t, hint["indices"], domain.SergeantMajorKittySize)

	var logOut map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(s)), &logOut))
	assert.Contains(t, logOut, "entries")
}

// **取り込むと手札に紛れて見分けが付かなくなる** (#5759)。位置は presenter が
// 解決して送る (手札は取り込み時に並べ替わるので、画面では追えない)。
func TestSergeantMajorWebPresenterMarksTheKittyCards(t *testing.T) {
	s := newSergeantMajorForWeb(t)
	s.SetPhaseForTest(domain.SergeantMajorPhaseTrump)
	s.SetDealerIdxForTest(0)
	require.NoError(t, s.DeclareTrump(domain.CardDesignSpade))

	out := decodeSergeantMajor(t, new(SergeantMajorWebPresenter).Output(s, nil))
	raw, ok := out["kittyIndices"].([]any)
	require.True(t, ok, "kittyIndices must be sent")
	assert.Len(t, raw, domain.SergeantMajorKittySize)

	// 送った位置の札が、実際にキティ由来であること。
	human := s.GetPlayer(0)
	for _, v := range raw {
		idx := int(v.(float64))
		assert.True(t, s.IsAbsorbedKittyCard(human.GetCard(idx)),
			"index %d is not a kitty card", idx)
	}

	// 捨て終われば空になる (受け入れ条件3)。
	require.NoError(t, s.DiscardForTest(0, []int{0, 1, 2, 3}))
	after := decodeSergeantMajor(t, new(SergeantMajorWebPresenter).Output(s, nil))
	assert.Empty(t, after["kittyIndices"])
}
