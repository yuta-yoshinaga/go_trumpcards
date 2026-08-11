//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newHokmForWeb(t *testing.T) *domain.Hokm {
	t.Helper()
	h := domain.NewDefaultHokm()
	h.Reset()
	return h
}

func decodeHokm(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestHokmWebPresenterOutput(t *testing.T) {
	p := new(HokmWebPresenter)
	h := newHokmForWeb(t)

	m := decodeHokm(t, p.Output(h, nil))

	assert.Equal(t, float64(domain.HokmPhaseTrump), m["phase"], "配り直後は宣言フェーズ")
	assert.Equal(t, float64(1), m["handNumber"])
	assert.Equal(t, float64(0), m["trumpSuit"], "切り札はまだ宣言されていない")
	assert.Equal(t, float64(-1), m["lastHandWinner"])

	players := m["players"].([]any)
	require.Len(t, players, domain.HokmPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	// **親だけが先に5枚持っている。**
	assert.Equal(t, float64(domain.HokmPeekSize), human["cardCount"])
	assert.True(t, human["isHakem"].(bool))
	assert.False(t, players[1].(map[string]any)["isHakem"].(bool))
	assert.Equal(t, float64(0), human["team"])
	assert.Equal(t, float64(0), players[2].(map[string]any)["team"], "0 と 2 が味方")
}

// **7 先取の進捗はトリック数のほうに出る。** 13 まで打たないので。
func TestHokmWebPresenterSurfacesTheRace(t *testing.T) {
	p := new(HokmWebPresenter)
	h := newHokmForWeb(t)
	h.GiveTricksForTest(0, 4)
	h.GiveTricksForTest(1, 2)

	m := decodeHokm(t, p.Output(h, nil))
	tricks := m["teamTricks"].([]any)
	assert.Equal(t, float64(4), tricks[0])
	assert.Equal(t, float64(2), tricks[1])
	assert.Equal(t, float64(domain.HokmTricksToWin), m["tricksToWin"])
}

// **Kot かどうかがワイヤに載る。** 2 点入った理由が盤面から読めない。
func TestHokmWebPresenterSurfacesKot(t *testing.T) {
	p := new(HokmWebPresenter)

	kot := newHokmForWeb(t)
	kot.GiveTricksForTest(0, domain.HokmTricksToWin)
	kot.FinishHandForTest(0)
	m := decodeHokm(t, p.Output(kot, nil))
	assert.True(t, m["lastHandKot"].(bool))
	assert.Equal(t, float64(0), m["lastHandWinner"])
	assert.Equal(t, float64(domain.HokmKotPoints), m["scores"].([]any)[0])
	assert.Equal(t, "hokm.handEndKot", m["messageCode"])

	// 負のコントロール: 1 トリックでも取られていれば Kot ではない。
	normal := newHokmForWeb(t)
	normal.GiveTricksForTest(0, domain.HokmTricksToWin)
	normal.GiveTricksForTest(1, 1)
	normal.FinishHandForTest(0)
	m2 := decodeHokm(t, p.Output(normal, nil))
	assert.False(t, m2["lastHandKot"].(bool))
	assert.Equal(t, float64(domain.HokmHandPoints), m2["scores"].([]any)[0])
	assert.Equal(t, "hokm.handEnd", m2["messageCode"])
}

// **親のときは宣言させ、そうでなければ待たせる。** 両側を踏む。
func TestHokmWebPresenterTrumpMessages(t *testing.T) {
	p := new(HokmWebPresenter)

	hakem := newHokmForWeb(t)
	hakem.SetHakemIdxForTest(0)
	assert.Equal(t, "hokm.trump.choose", decodeHokm(t, p.Output(hakem, nil))["messageCode"])

	other := newHokmForWeb(t)
	other.SetHakemIdxForTest(2)
	assert.Equal(t, "hokm.trump.wait", decodeHokm(t, p.Output(other, nil))["messageCode"])
}

func TestHokmWebPresenterResultMessage(t *testing.T) {
	p := new(HokmWebPresenter)

	for _, tc := range []struct {
		name     string
		t0, t1   int
		wantCode string
	}{
		{"your team wins", 7, 3, "hokm.result.team0"},
		{"they win", 3, 7, "hokm.result.team1"},
		{"a tie", 5, 5, "hokm.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHokmForWeb(t)
			h.SetScoreForTestUse(0, tc.t0)
			h.SetScoreForTestUse(1, tc.t1)
			h.FinishGameForTest()

			assert.Equal(t, tc.wantCode, decodeHokm(t, p.Output(h, nil))["messageCode"])
		})
	}
}

func TestHokmWebPresenterError(t *testing.T) {
	p := new(HokmWebPresenter)
	m := decodeHokm(t, p.Output(newHokmForWeb(t), assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestHokmWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(HokmWebPresenter)
	h := newHokmForWeb(t)
	h.GiveUp()

	m := decodeHokm(t, p.Output(h, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

// 宣言フェーズのヒントは札を指さず、スートを運ぶ。
func TestHokmWebPresenterHintCarriesTheSuit(t *testing.T) {
	p := new(HokmWebPresenter)
	h := newHokmForWeb(t)
	h.SetHakemIdxForTest(0)

	hint, ok := decodeHokm(t, p.HintOutput(h))["hint"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, hint["cardIndex"], "宣言では札を指さない")
	assert.Equal(t, "hokmDeclareTrump", hint["reason"])
	assert.GreaterOrEqual(t, hint["suit"], float64(domain.CardDesignSpade))
}

func TestHokmWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(HokmWebPresenter)
	h := newHokmForWeb(t)
	h.GiveUp()
	assert.Nil(t, decodeHokm(t, p.HintOutput(h))["hint"])
}

func TestHokmWebPresenterConfigSurfaces(t *testing.T) {
	p := new(HokmWebPresenter)
	h := newHokmForWeb(t)
	h.SetConfig(domain.HokmConfig{Target: 9})

	assert.Equal(t, float64(9), decodeHokm(t, p.Output(h, nil))["config"].(map[string]any)["target"])
}

func TestHokmWebPresenterActionLogOutput(t *testing.T) {
	p := new(HokmWebPresenter)
	h := newHokmForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(h)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	h.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(h)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
