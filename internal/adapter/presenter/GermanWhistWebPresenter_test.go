//go:build test

package presenter

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newGermanWhistForWeb は実物のドメインを使う。モックだと「本当にその値が
// 出ているか」ではなく「モックに何を教えたか」を検証してしまう。
func newGermanWhistForWeb(t *testing.T) *domain.GermanWhist {
	t.Helper()
	g := domain.NewDefaultGermanWhist()
	g.Reset()
	return g
}

func decodeGermanWhist(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestGermanWhistWebPresenterOutput(t *testing.T) {
	p := new(GermanWhistWebPresenter)
	g := newGermanWhistForWeb(t)

	m := decodeGermanWhist(t, p.Output(g, nil))

	assert.Equal(t, float64(domain.GermanWhistPhaseDraw), m["phase"])
	assert.False(t, m["gameEndFlag"].(bool))
	assert.Equal(t, float64(-1), m["winnerIdx"])
	// 配り終えた直後: 各 13 枚、表向き 1 枚、残りが山札。
	assert.Equal(t, float64(52-13*2-1), m["stockCount"])
	require.NotNil(t, m["upCard"], "前半は表向きの札が必ずある")

	players := m["players"].([]any)
	require.Len(t, players, 2)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(13), human["cardCount"])
	assert.Len(t, human["cards"].([]any), 13, "人間の手札は見える")
	cpu := players[1].(map[string]any)
	assert.Empty(t, cpu["cards"], "CPU の手札は伏せる")
	assert.Equal(t, float64(13), cpu["cardCount"], "枚数だけは分かる")
}

// validPlays は nil ではなく空配列で出る。JSON の null はフロントで
// `.includes` を落とす。
func TestGermanWhistWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(GermanWhistWebPresenter)
	g := newGermanWhistForWeb(t)
	g.GiveUp()

	m := decodeGermanWhist(t, p.Output(g, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

func TestGermanWhistWebPresenterError(t *testing.T) {
	p := new(GermanWhistWebPresenter)
	g := newGermanWhistForWeb(t)

	m := decodeGermanWhist(t, p.Output(g, assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"], "エラー時は messageCode を出さない")
}

// 前半と後半で案内メッセージが入れ替わる。
func TestGermanWhistWebPresenterPhaseMessage(t *testing.T) {
	p := new(GermanWhistWebPresenter)
	g := newGermanWhistForWeb(t)

	assert.Equal(t, "germanwhist.phase.draw", decodeGermanWhist(t, p.Output(g, nil))["messageCode"])

	g.SetPhase(domain.GermanWhistPhaseScoring)
	assert.Equal(t, "germanwhist.phase.scoring", decodeGermanWhist(t, p.Output(g, nil))["messageCode"])
}

func TestGermanWhistWebPresenterResultMessage(t *testing.T) {
	for _, tc := range []struct {
		name     string
		s0, s1   int
		wantCode string
	}{
		{"human wins", 7, 6, "germanwhist.result.p0Win"},
		{"cpu wins", 6, 7, "germanwhist.result.p1Win"},
		{"tie", 6, 6, "germanwhist.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := new(GermanWhistWebPresenter)
			g := newGermanWhistForWeb(t)
			g.GetPlayer(0).SetScoringTricks(tc.s0)
			g.GetPlayer(1).SetScoringTricks(tc.s1)
			g.Finish()

			m := decodeGermanWhist(t, p.Output(g, nil))
			assert.Equal(t, tc.wantCode, m["messageCode"])
			params := m["messageParams"].(map[string]any)
			assert.Equal(t, strconv.Itoa(tc.s0), params["p0"])
			assert.Equal(t, strconv.Itoa(tc.s1), params["p1"])
		})
	}
}

func TestGermanWhistWebPresenterHintOutput(t *testing.T) {
	p := new(GermanWhistWebPresenter)
	g := newGermanWhistForWeb(t)

	m := decodeGermanWhist(t, p.HintOutput(g))
	hint, ok := m["hint"].(map[string]any)
	require.True(t, ok, "人間の手番ならヒントが出る")
	assert.NotEmpty(t, hint["reason"])
	assert.NotNil(t, hint["cardIndex"])
}

// 終局後はヒントを出さない。
func TestGermanWhistWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(GermanWhistWebPresenter)
	g := newGermanWhistForWeb(t)
	g.GiveUp()

	assert.Nil(t, decodeGermanWhist(t, p.HintOutput(g))["hint"])
}

func TestGermanWhistWebPresenterActionLogOutput(t *testing.T) {
	p := new(GermanWhistWebPresenter)
	g := newGermanWhistForWeb(t)
	g.GiveUp()

	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(g)), &m))
	entries, ok := m["entries"].([]any)
	require.True(t, ok, "棋譜は entries キーで出る")
	assert.NotEmpty(t, entries, "投了は棋譜に残る")
}
