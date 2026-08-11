//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newReversisForWeb(t *testing.T) *domain.Reversis {
	t.Helper()
	r := domain.NewDefaultReversis()
	r.Reset()
	return r
}

func decodeReversis(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestReversisWebPresenterOutput(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)

	m := decodeReversis(t, p.Output(r, nil))

	assert.Equal(t, float64(domain.ReversisPhasePlay), m["phase"])
	assert.Equal(t, float64(1), m["roundNumber"])
	// **プールは常に出す。** 何を取り合っているのかが盤面から読めない。
	assert.Equal(t, float64(domain.ReversisAnte*domain.ReversisPlayerCnt), m["pool"])

	players := m["players"].([]any)
	require.Len(t, players, domain.ReversisPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Len(t, human["cards"].([]any), domain.ReversisHandSize, "人間の手札は見える")
	assert.Equal(t, float64(domain.ReversisStartingChips-domain.ReversisAnte), human["chips"])
	assert.Equal(t, float64(0), human["roundPenalty"])
	assert.False(t, human["tookQuinola"].(bool))
	assert.False(t, human["tookDiamondAce"].(bool))

	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// 印付きの札を取ったことが画面に出る。両側を踏む。
func TestReversisWebPresenterMarkedFlagsSurface(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)
	r.GetPlayer(0).SetTookQuinola(true)
	r.GetPlayer(2).SetTookDiamondAce(true)

	players := decodeReversis(t, p.Output(r, nil))["players"].([]any)
	assert.True(t, players[0].(map[string]any)["tookQuinola"].(bool))
	assert.False(t, players[0].(map[string]any)["tookDiamondAce"].(bool))
	assert.True(t, players[2].(map[string]any)["tookDiamondAce"].(bool))
	assert.False(t, players[1].(map[string]any)["tookQuinola"].(bool))
}

func TestReversisWebPresenterPlayMessageCarriesThePool(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)

	m := decodeReversis(t, p.Output(r, nil))
	assert.Equal(t, "reversis.play", m["messageCode"])
	assert.Equal(t, "20", m["messageParams"].(map[string]any)["pool"])
}

func TestReversisWebPresenterRoundEndMessage(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)
	r.SetPhaseForTest(domain.ReversisPhaseRoundEnd)

	m := decodeReversis(t, p.Output(r, nil))
	assert.Equal(t, "reversis.roundEnd", m["messageCode"])
	assert.Equal(t, "1", m["messageParams"].(map[string]any)["round"])
}

func TestReversisWebPresenterResultMessage(t *testing.T) {
	p := new(ReversisWebPresenter)

	t.Run("a winner", func(t *testing.T) {
		r := newReversisForWeb(t)
		r.GetPlayer(3).SetChips(120)
		r.FinishGameForTest()

		m := decodeReversis(t, p.Output(r, nil))
		assert.Equal(t, "reversis.result.winner", m["messageCode"])
		params := m["messageParams"].(map[string]any)
		assert.Equal(t, "3", params["idx"])
		assert.Equal(t, "120", params["chips"])
	})

	t.Run("a tie", func(t *testing.T) {
		r := newReversisForWeb(t)
		r.FinishGameForTest() // 全員同じチップ
		assert.Equal(t, "reversis.result.tie", decodeReversis(t, p.Output(r, nil))["messageCode"])
	})
}

func TestReversisWebPresenterError(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)

	m := decodeReversis(t, p.Output(r, assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestReversisWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)
	r.GiveUp()

	m := decodeReversis(t, p.Output(r, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

func TestReversisWebPresenterHintOutput(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)
	r.SetCurrentPlayerIdxForTest(0)

	hint, ok := decodeReversis(t, p.HintOutput(r))["hint"].(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, hint["reason"])
	assert.NotNil(t, hint["cardIndex"])
}

func TestReversisWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)
	r.GiveUp()
	assert.Nil(t, decodeReversis(t, p.HintOutput(r))["hint"])
}

func TestReversisWebPresenterConfigSurfaces(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)
	r.SetConfig(domain.ReversisConfig{Rounds: 6})

	assert.Equal(t, float64(6), decodeReversis(t, p.Output(r, nil))["config"].(map[string]any)["rounds"])
}

func TestReversisWebPresenterActionLogOutput(t *testing.T) {
	p := new(ReversisWebPresenter)
	r := newReversisForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(r)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	r.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(r)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
