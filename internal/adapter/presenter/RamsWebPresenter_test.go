//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newRamsForWeb(t *testing.T) *domain.Rams {
	t.Helper()
	r := domain.NewDefaultRams()
	r.Reset()
	return r
}

func decodeRams(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestRamsWebPresenterOutput(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)

	m := decodeRams(t, p.Output(r, nil))

	assert.Equal(t, float64(domain.RamsPhaseDecide), m["phase"], "配り直後は選択フェーズ")
	assert.Equal(t, float64(1), m["roundNumber"])
	// **ポットと切り札は参加判断の材料。** 常に出す。
	assert.Equal(t, float64(domain.RamsAnte*domain.RamsPlayerCntDefault), m["pot"])
	require.NotNil(t, m["upCard"], "切り札を決めた 1 枚が出る")
	assert.Equal(t, float64(r.GetTrumpSuit()), m["trumpSuit"])

	players := m["players"].([]any)
	require.Len(t, players, domain.RamsPlayerCntDefault)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Len(t, human["cards"].([]any), domain.RamsHandSize, "人間の手札は見える")
	assert.Equal(t, float64(domain.RamsStartingChips-domain.RamsAnte), human["chips"])
	assert.False(t, human["decided"].(bool), "まだ選んでいない")
	assert.False(t, human["inRound"].(bool))

	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **人数は可変。** 5 人でも席が 5 つ出る。
func TestRamsWebPresenterFivePlayers(t *testing.T) {
	p := new(RamsWebPresenter)
	r := domain.NewDefaultRams()
	r.SetConfig(domain.RamsConfig{PlayerCnt: 5, Rounds: 4})
	r.Reset()

	m := decodeRams(t, p.Output(r, nil))
	assert.Len(t, m["players"].([]any), 5)
	assert.Equal(t, float64(5), m["config"].(map[string]any)["playerCnt"])
	assert.Equal(t, float64(domain.RamsAnte*5), m["pot"], "5 人分のアンティ")
}

// 参加状況が出る。降りた人と参加した人の両方を踏む。
func TestRamsWebPresenterDecisionsSurface(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)
	require.NoError(t, r.Decide(false)) // 人間は降りる

	players := decodeRams(t, p.Output(r, nil))["players"].([]any)
	human := players[0].(map[string]any)
	assert.True(t, human["decided"].(bool))
	assert.False(t, human["inRound"].(bool), "降りた")
	// CPU も全員決めている。
	for i := 1; i < len(players); i++ {
		assert.True(t, players[i].(map[string]any)["decided"].(bool), "player %d", i)
	}
}

func TestRamsWebPresenterDecidePhaseMessage(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)

	m := decodeRams(t, p.Output(r, nil))
	assert.Equal(t, "rams.decidePhase", m["messageCode"])
	assert.Equal(t, "12", m["messageParams"].(map[string]any)["pot"])
}

// **降りたラウンドは「見ている」と伝える。** 操作待ちに見えてはいけない。
func TestRamsWebPresenterWatchingMessage(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)
	r.GetPlayer(0).SetDecided(true)
	r.GetPlayer(0).SetInRound(false)
	r.SetPhaseForTest(domain.RamsPhasePlay)

	assert.Equal(t, "rams.watching", decodeRams(t, p.Output(r, nil))["messageCode"])
}

func TestRamsWebPresenterPlayMessage(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)
	r.GetPlayer(0).SetDecided(true)
	r.GetPlayer(0).SetInRound(true)
	r.SetPhaseForTest(domain.RamsPhasePlay)

	assert.Equal(t, "rams.play", decodeRams(t, p.Output(r, nil))["messageCode"])
}

func TestRamsWebPresenterRoundEndMessage(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)
	r.SetPhaseForTest(domain.RamsPhaseRoundEnd)

	m := decodeRams(t, p.Output(r, nil))
	assert.Equal(t, "rams.roundEnd", m["messageCode"])
	assert.Equal(t, "1", m["messageParams"].(map[string]any)["round"])
}

func TestRamsWebPresenterResultMessage(t *testing.T) {
	p := new(RamsWebPresenter)

	t.Run("a winner", func(t *testing.T) {
		r := newRamsForWeb(t)
		r.GetPlayer(2).SetChips(150)
		r.FinishGameForTest()

		m := decodeRams(t, p.Output(r, nil))
		assert.Equal(t, "rams.result.winner", m["messageCode"])
		params := m["messageParams"].(map[string]any)
		assert.Equal(t, "2", params["idx"])
		assert.Equal(t, "150", params["chips"])
	})

	t.Run("a tie", func(t *testing.T) {
		r := newRamsForWeb(t)
		r.FinishGameForTest()
		assert.Equal(t, "rams.result.tie", decodeRams(t, p.Output(r, nil))["messageCode"])
	})
}

func TestRamsWebPresenterError(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)

	m := decodeRams(t, p.Output(r, assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestRamsWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)
	r.GiveUp()

	m := decodeRams(t, p.Output(r, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

// 選択フェーズのヒントは札を指さない。
func TestRamsWebPresenterHintInDecidePhase(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)

	hint, ok := decodeRams(t, p.HintOutput(r))["hint"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, hint["cardIndex"], "参加可否の助言なので札は指さない")
	assert.Contains(t, []any{"ramsPlayIn", "ramsPassOut"}, hint["reason"])
}

func TestRamsWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)
	r.GiveUp()
	assert.Nil(t, decodeRams(t, p.HintOutput(r))["hint"])
}

func TestRamsWebPresenterActionLogOutput(t *testing.T) {
	p := new(RamsWebPresenter)
	r := newRamsForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(r)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	r.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(r)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
