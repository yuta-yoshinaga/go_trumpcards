//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newPolignacForWeb(t *testing.T) *domain.Polignac {
	t.Helper()
	p := domain.NewDefaultPolignac()
	p.Reset()
	require.NoError(t, p.PassDeclaration())
	return p
}

func decodePolignac(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestPolignacWebPresenterOutput(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := newPolignacForWeb(t)

	m := decodePolignac(t, p.Output(g, nil))

	assert.Equal(t, float64(domain.PolignacPhasePlay), m["phase"])
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(-1), m["capotIdx"], "宣言なしは -1")
	assert.False(t, m["gameEndFlag"].(bool))

	players := m["players"].([]any)
	require.Len(t, players, domain.PolignacPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Len(t, human["cards"].([]any), domain.PolignacHandSize, "人間の手札は見える")
	assert.Equal(t, float64(0), human["score"])
	assert.Equal(t, float64(0), human["roundPenalty"])
	assert.False(t, human["declaredCapot"].(bool))

	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **宣言フェーズは専用の案内を出す。** ここで打ち手を促すと、
// capot を選ぶ機会があることが伝わらない。
func TestPolignacWebPresenterDeclarePhaseMessage(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := domain.NewDefaultPolignac()
	g.Reset()

	m := decodePolignac(t, p.Output(g, nil))
	assert.Equal(t, "polignac.declarePhase", m["messageCode"])
	assert.Equal(t, float64(domain.PolignacPhaseDeclare), m["phase"])
}

// capot 宣言中は狙いが変わるので案内も変わる。両側を踏む。
func TestPolignacWebPresenterCapotMessages(t *testing.T) {
	p := new(PolignacWebPresenter)

	t.Run("no declaration", func(t *testing.T) {
		g := newPolignacForWeb(t)
		assert.Equal(t, "polignac.play.normal", decodePolignac(t, p.Output(g, nil))["messageCode"])
	})

	t.Run("capot declared", func(t *testing.T) {
		g := domain.NewDefaultPolignac()
		g.Reset()
		require.NoError(t, g.DeclareCapot())

		m := decodePolignac(t, p.Output(g, nil))
		assert.Equal(t, "polignac.play.capotActive", m["messageCode"])
		assert.Equal(t, "0", m["messageParams"].(map[string]any)["idx"])
		assert.Equal(t, float64(0), m["capotIdx"])
		assert.True(t, m["players"].([]any)[0].(map[string]any)["declaredCapot"].(bool))
	})
}

func TestPolignacWebPresenterRoundEndMessage(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := newPolignacForWeb(t)
	g.SetPhaseForTest(domain.PolignacPhaseRoundEnd)

	m := decodePolignac(t, p.Output(g, nil))
	assert.Equal(t, "polignac.roundEnd", m["messageCode"])
	assert.Equal(t, "1", m["messageParams"].(map[string]any)["round"])
}

func TestPolignacWebPresenterResultMessage(t *testing.T) {
	p := new(PolignacWebPresenter)

	t.Run("a winner", func(t *testing.T) {
		g := newPolignacForWeb(t)
		// 失点が最も少ない席が勝つ。
		for i := range domain.PolignacPlayerCnt {
			g.GetPlayer(i).SetScore(5)
		}
		g.GetPlayer(3).SetScore(1)
		g.FinishGameForTest()

		m := decodePolignac(t, p.Output(g, nil))
		assert.Equal(t, "polignac.result.winner", m["messageCode"])
		params := m["messageParams"].(map[string]any)
		assert.Equal(t, "3", params["idx"])
		assert.Equal(t, "1", params["score"])
	})

	t.Run("a tie", func(t *testing.T) {
		g := newPolignacForWeb(t)
		g.FinishGameForTest()
		assert.Equal(t, "polignac.result.tie", decodePolignac(t, p.Output(g, nil))["messageCode"])
	})
}

func TestPolignacWebPresenterError(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := newPolignacForWeb(t)

	m := decodePolignac(t, p.Output(g, assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestPolignacWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := newPolignacForWeb(t)
	g.GiveUp()

	m := decodePolignac(t, p.Output(g, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

func TestPolignacWebPresenterHintOutput(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := newPolignacForWeb(t)
	g.SetCurrentPlayerIdxForTest(0)

	hint, ok := decodePolignac(t, p.HintOutput(g))["hint"].(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, hint["reason"])
	assert.NotNil(t, hint["cardIndex"])
}

func TestPolignacWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := newPolignacForWeb(t)
	g.GiveUp()
	assert.Nil(t, decodePolignac(t, p.HintOutput(g))["hint"])
}

func TestPolignacWebPresenterConfigSurfaces(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := newPolignacForWeb(t)
	g.SetConfig(domain.PolignacConfig{Rounds: 6})

	assert.Equal(t, float64(6), decodePolignac(t, p.Output(g, nil))["config"].(map[string]any)["rounds"])
}

func TestPolignacWebPresenterActionLogOutput(t *testing.T) {
	p := new(PolignacWebPresenter)
	g := newPolignacForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(g)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	g.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(g)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
