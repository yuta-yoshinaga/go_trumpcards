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

func newSlobberhannesForWeb(t *testing.T) *domain.Slobberhannes {
	t.Helper()
	s := domain.NewDefaultSlobberhannes()
	s.Reset()
	return s
}

func decodeSlobberhannes(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestSlobberhannesWebPresenterOutput(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)

	m := decodeSlobberhannes(t, p.Output(s, nil))

	assert.Equal(t, float64(domain.SlobberhannesPhasePlay), m["phase"])
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.False(t, m["gameEndFlag"].(bool))
	assert.Equal(t, float64(-1), m["winnerIdx"])

	players := m["players"].([]any)
	require.Len(t, players, domain.SlobberhannesPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.SlobberhannesHandSize), human["cardCount"])
	assert.Len(t, human["cards"].([]any), domain.SlobberhannesHandSize, "人間の手札は見える")
	assert.Equal(t, float64(0), human["score"])
	// 罰の内訳が最初から出ている（画面で「無傷」を表示するため）。
	assert.False(t, human["tookFirstTrick"].(bool))
	assert.False(t, human["tookLastTrick"].(bool))
	assert.False(t, human["tookQueen"].(bool))

	cpu := players[1].(map[string]any)
	assert.Empty(t, cpu["cards"], "CPU の手札は伏せる")
	assert.Equal(t, float64(domain.SlobberhannesHandSize), cpu["cardCount"])
}

func TestSlobberhannesWebPresenterPenaltiesSurface(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)
	s.SetPenaltiesForTest(0, true, false, true)

	players := decodeSlobberhannes(t, p.Output(s, nil))["players"].([]any)
	human := players[0].(map[string]any)
	assert.True(t, human["tookFirstTrick"].(bool))
	assert.False(t, human["tookLastTrick"].(bool))
	assert.True(t, human["tookQueen"].(bool))
}

func TestSlobberhannesWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)
	s.GiveUp()

	m := decodeSlobberhannes(t, p.Output(s, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

func TestSlobberhannesWebPresenterError(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)

	m := decodeSlobberhannes(t, p.Output(s, assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

// **最初と最後のトリックだけ案内が変わる。** 位置そのものが罰の対象という
// 規則が画面に出ていなければ、プレイヤーは気付けない。
func TestSlobberhannesWebPresenterTrickPositionMessages(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	for _, tc := range []struct {
		name     string
		trick    int
		wantCode string
	}{
		{"first trick", 0, "slobberhannes.play.firstTrick"},
		{"middle trick", 3, "slobberhannes.play.normal"},
		{"last trick", domain.SlobberhannesTricksPerRound - 1, "slobberhannes.play.lastTrick"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSlobberhannesForWeb(t)
			s.SetTrickNumberForTest(tc.trick)
			assert.Equal(t, tc.wantCode, decodeSlobberhannes(t, p.Output(s, nil))["messageCode"])
		})
	}
}

func TestSlobberhannesWebPresenterRoundEndMessage(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)
	s.SetPhase(domain.SlobberhannesPhaseRoundEnd)

	m := decodeSlobberhannes(t, p.Output(s, nil))
	assert.Equal(t, "slobberhannes.roundEnd", m["messageCode"])
	assert.Equal(t, "1", m["messageParams"].(map[string]any)["round"])
}

func TestSlobberhannesWebPresenterResultMessage(t *testing.T) {
	p := new(SlobberhannesWebPresenter)

	t.Run("a winner", func(t *testing.T) {
		s := newSlobberhannesForWeb(t)
		s.GetPlayer(2).SetScore(3)
		s.Finish()

		m := decodeSlobberhannes(t, p.Output(s, nil))
		assert.Equal(t, "slobberhannes.result.winner", m["messageCode"])
		params := m["messageParams"].(map[string]any)
		assert.Equal(t, "2", params["idx"])
		assert.Equal(t, strconv.Itoa(3), params["score"])
	})

	t.Run("a tie", func(t *testing.T) {
		s := newSlobberhannesForWeb(t)
		s.Finish() // 全員 0 点
		assert.Equal(t, "slobberhannes.result.tie", decodeSlobberhannes(t, p.Output(s, nil))["messageCode"])
	})
}

func TestSlobberhannesWebPresenterHintOutput(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)
	s.SetCurrentPlayerIdxForTest(0)

	m := decodeSlobberhannes(t, p.HintOutput(s))
	hint, ok := m["hint"].(map[string]any)
	require.True(t, ok, "人間の手番ならヒントが出る")
	assert.NotEmpty(t, hint["reason"])
	assert.NotNil(t, hint["cardIndex"])
}

func TestSlobberhannesWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)
	s.GiveUp()
	assert.Nil(t, decodeSlobberhannes(t, p.HintOutput(s))["hint"])
}

func TestSlobberhannesWebPresenterConfigSurfaces(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)
	s.SetConfig(domain.SlobberhannesConfig{Rounds: 6})

	cfg := decodeSlobberhannes(t, p.Output(s, nil))["config"].(map[string]any)
	assert.Equal(t, float64(6), cfg["rounds"])
}

func TestSlobberhannesWebPresenterActionLogOutput(t *testing.T) {
	p := new(SlobberhannesWebPresenter)
	s := newSlobberhannesForWeb(t)

	// 棋譜は終局するまで出さない（進行中に手を明かさないための共通仕様）。
	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(s)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	s.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(s)), &after))
	entries, ok := after["entries"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, entries, "終局後は配りと投了が残る")
}
