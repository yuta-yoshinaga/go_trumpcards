//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTarabishForWeb(t *testing.T) *domain.Tarabish {
	t.Helper()
	tb := domain.NewDefaultTarabish()
	tb.Reset()
	return tb
}

func decodeTarabish(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestTarabishWebPresenterOutput(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)

	m := decodeTarabish(t, p.Output(tb, nil))

	assert.Equal(t, float64(domain.TarabishPhaseBid), m["phase"], "配り直後は入札フェーズ")
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(-1), m["trumpTakerIdx"], "まだ誰も引き受けていない")
	require.NotNil(t, m["upCard"], "切り札候補が出る")

	players := m["players"].([]any)
	require.Len(t, players, domain.TarabishPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	// **チーム番号を出す。** 向かい合う席が味方であることは盤面から読めない。
	assert.Equal(t, float64(0), human["team"])
	assert.Equal(t, float64(1), players[1].(map[string]any)["team"])
	assert.Equal(t, float64(0), players[2].(map[string]any)["team"], "0 と 2 が味方")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")

	// チーム得点は 2 要素の配列で出る。
	assert.Len(t, m["scores"].([]any), domain.TarabishTeamCnt)
	assert.Len(t, m["roundPoints"].([]any), domain.TarabishTeamCnt)
}

// メルドは配り札から自動判定され、内訳ごと出る。
func TestTarabishWebPresenterMeldSurfaces(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)
	tb.GetPlayer(0).SetMeldPoints(70)
	tb.GetPlayer(0).SetRunLen(4)
	tb.GetPlayer(0).SetHasBella(true)

	human := decodeTarabish(t, p.Output(tb, nil))["players"].([]any)[0].(map[string]any)
	assert.Equal(t, float64(70), human["meldPoints"])
	assert.Equal(t, float64(4), human["runLen"])
	assert.True(t, human["hasBella"].(bool))
}

// **親は見送れないので案内が変わる。** 選べない選択肢を出さない。両側を踏む。
func TestTarabishWebPresenterBidMessages(t *testing.T) {
	p := new(TarabishWebPresenter)

	t.Run("the human is the dealer", func(t *testing.T) {
		tb := newTarabishForWeb(t)
		tb.SetDealerIdxForTest(0)
		tb.SetCurrentPlayerIdxForTest(0)
		assert.Equal(t, "tarabish.bid.dealerStuck", decodeTarabish(t, p.Output(tb, nil))["messageCode"])
	})

	t.Run("someone else is the dealer", func(t *testing.T) {
		tb := newTarabishForWeb(t)
		tb.SetDealerIdxForTest(2)
		tb.SetCurrentPlayerIdxForTest(0)
		assert.Equal(t, "tarabish.bid.choose", decodeTarabish(t, p.Output(tb, nil))["messageCode"])
	})
}

func TestTarabishWebPresenterRoundEndMessage(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)
	tb.SetPhaseForTest(domain.TarabishPhaseRoundEnd)
	tb.SetRoundPointsForTest(0, 90)
	tb.SetRoundPointsForTest(1, 72)

	m := decodeTarabish(t, p.Output(tb, nil))
	assert.Equal(t, "tarabish.roundEnd", m["messageCode"])
	params := m["messageParams"].(map[string]any)
	assert.Equal(t, "90", params["t0"])
	assert.Equal(t, "72", params["t1"])
}

func TestTarabishWebPresenterResultMessage(t *testing.T) {
	p := new(TarabishWebPresenter)

	for _, tc := range []struct {
		name     string
		t0, t1   int
		wantCode string
	}{
		{"your team wins", 520, 300, "tarabish.result.team0"},
		{"they win", 300, 520, "tarabish.result.team1"},
		{"a tie", 500, 500, "tarabish.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb := newTarabishForWeb(t)
			tb.SetScoreForTestUse(0, tc.t0)
			tb.SetScoreForTestUse(1, tc.t1)
			tb.FinishGameForTest()

			m := decodeTarabish(t, p.Output(tb, nil))
			assert.Equal(t, tc.wantCode, m["messageCode"])
		})
	}
}

func TestTarabishWebPresenterError(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)

	m := decodeTarabish(t, p.Output(tb, assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestTarabishWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)
	tb.GiveUp()

	m := decodeTarabish(t, p.Output(tb, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

// 入札フェーズのヒントは札を指さない。
func TestTarabishWebPresenterHintInBidPhase(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)
	tb.SetCurrentPlayerIdxForTest(0)

	hint, ok := decodeTarabish(t, p.HintOutput(tb))["hint"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, hint["cardIndex"], "入札では札を指さない")
	assert.Contains(t, []any{"tarabishTakeTrump", "tarabishPassTrump"}, hint["reason"])
}

func TestTarabishWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)
	tb.GiveUp()
	assert.Nil(t, decodeTarabish(t, p.HintOutput(tb))["hint"])
}

func TestTarabishWebPresenterConfigSurfaces(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)
	tb.SetConfig(domain.TarabishConfig{Target: 300})

	assert.Equal(t, float64(300), decodeTarabish(t, p.Output(tb, nil))["config"].(map[string]any)["target"])
}

func TestTarabishWebPresenterActionLogOutput(t *testing.T) {
	p := new(TarabishWebPresenter)
	tb := newTarabishForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(tb)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	tb.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(tb)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
