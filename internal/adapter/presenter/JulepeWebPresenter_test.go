//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newJulepeForWeb(t *testing.T) *domain.Julepe {
	t.Helper()
	r := domain.NewDefaultJulepe()
	r.Reset()
	return r
}

func decodeJulepe(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestJulepeWebPresenterOutput(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)

	m := decodeJulepe(t, p.Output(r, nil))

	assert.Equal(t, float64(domain.JulepePhaseDecide), m["phase"], "配り直後は選択フェーズ")
	assert.Equal(t, float64(1), m["roundNumber"])
	// **ポットと切り札は参加判断の材料。** 常に出す。
	assert.Equal(t, float64(domain.JulepeAnte*domain.JulepePlayerCntDefault), m["pot"])
	require.NotNil(t, m["upCard"], "切り札を決めた 1 枚が出る")
	assert.Equal(t, float64(r.GetTrumpSuit()), m["trumpSuit"])

	players := m["players"].([]any)
	require.Len(t, players, domain.JulepePlayerCntDefault)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Len(t, human["cards"].([]any), domain.JulepeHandSize, "人間の手札は見える")
	assert.Equal(t, float64(domain.JulepeStartingChips-domain.JulepeAnte), human["chips"])
	assert.False(t, human["decided"].(bool), "まだ選んでいない")
	assert.False(t, human["inRound"].(bool))

	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **人数は可変。** 5 人でも席が 5 つ出る。
func TestJulepeWebPresenterFivePlayers(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := domain.NewDefaultJulepe()
	r.SetConfig(domain.JulepeConfig{PlayerCnt: 5, Rounds: 4})
	r.Reset()

	m := decodeJulepe(t, p.Output(r, nil))
	assert.Len(t, m["players"].([]any), 5)
	assert.Equal(t, float64(5), m["config"].(map[string]any)["playerCnt"])
	assert.Equal(t, float64(domain.JulepeAnte*5), m["pot"], "5 人分のアンティ")
}

// 参加状況が出る。降りた人と参加した人の両方を踏む。
func TestJulepeWebPresenterDecisionsSurface(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)
	require.NoError(t, r.Decide(false)) // 人間は降りる

	players := decodeJulepe(t, p.Output(r, nil))["players"].([]any)
	human := players[0].(map[string]any)
	assert.True(t, human["decided"].(bool))
	assert.False(t, human["inRound"].(bool), "降りた")
	// CPU も全員決めている。
	for i := 1; i < len(players); i++ {
		assert.True(t, players[i].(map[string]any)["decided"].(bool), "player %d", i)
	}
}

func TestJulepeWebPresenterDecidePhaseMessage(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)

	m := decodeJulepe(t, p.Output(r, nil))
	assert.Equal(t, "julepe.decidePhase", m["messageCode"])
	assert.Equal(t, "12", m["messageParams"].(map[string]any)["pot"])
}

// **降りたラウンドは「見ている」と伝える。** 操作待ちに見えてはいけない。
func TestJulepeWebPresenterWatchingMessage(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)
	r.GetPlayer(0).SetDecided(true)
	r.GetPlayer(0).SetInRound(false)
	r.SetPhaseForTest(domain.JulepePhasePlay)

	assert.Equal(t, "julepe.watching", decodeJulepe(t, p.Output(r, nil))["messageCode"])
}

func TestJulepeWebPresenterPlayMessage(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)
	r.GetPlayer(0).SetDecided(true)
	r.GetPlayer(0).SetInRound(true)
	r.SetPhaseForTest(domain.JulepePhasePlay)

	assert.Equal(t, "julepe.play", decodeJulepe(t, p.Output(r, nil))["messageCode"])
}

func TestJulepeWebPresenterRoundEndMessage(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)
	r.SetPhaseForTest(domain.JulepePhaseRoundEnd)

	m := decodeJulepe(t, p.Output(r, nil))
	assert.Equal(t, "julepe.roundEnd", m["messageCode"])
	assert.Equal(t, "1", m["messageParams"].(map[string]any)["round"])
}

func TestJulepeWebPresenterResultMessage(t *testing.T) {
	p := new(JulepeWebPresenter)

	t.Run("a winner", func(t *testing.T) {
		r := newJulepeForWeb(t)
		r.GetPlayer(2).SetChips(150)
		r.FinishGameForTest()

		m := decodeJulepe(t, p.Output(r, nil))
		assert.Equal(t, "julepe.result.winner", m["messageCode"])
		params := m["messageParams"].(map[string]any)
		assert.Equal(t, "2", params["idx"])
		assert.Equal(t, "150", params["chips"])
	})

	t.Run("a tie", func(t *testing.T) {
		r := newJulepeForWeb(t)
		r.FinishGameForTest()
		assert.Equal(t, "julepe.result.tie", decodeJulepe(t, p.Output(r, nil))["messageCode"])
	})
}

func TestJulepeWebPresenterError(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)

	m := decodeJulepe(t, p.Output(r, assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestJulepeWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)
	r.GiveUp()

	m := decodeJulepe(t, p.Output(r, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

// 選択フェーズのヒントは札を指さない。
func TestJulepeWebPresenterHintInDecidePhase(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)

	hint, ok := decodeJulepe(t, p.HintOutput(r))["hint"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, hint["cardIndex"], "参加可否の助言なので札は指さない")
	assert.Contains(t, []any{"julepePlayIn", "julepePassOut"}, hint["reason"])
}

func TestJulepeWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)
	r.GiveUp()
	assert.Nil(t, decodeJulepe(t, p.HintOutput(r))["hint"])
}

func TestJulepeWebPresenterActionLogOutput(t *testing.T) {
	p := new(JulepeWebPresenter)
	r := newJulepeForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(r)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	r.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(r)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
