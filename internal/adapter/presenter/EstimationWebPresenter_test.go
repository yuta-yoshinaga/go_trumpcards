//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newEstimationForWeb(t *testing.T) *domain.Estimation {
	t.Helper()
	e := domain.NewDefaultEstimation()
	e.Reset()
	return e
}

func decodeEstimation(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestEstimationWebPresenterOutput(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)

	m := decodeEstimation(t, p.Output(e, nil))

	assert.Equal(t, float64(domain.EstimationPhaseTrump), m["phase"], "配り直後は切り札選択")
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(0), m["trumpSuit"], "切り札はまだ決まっていない")
	assert.Equal(t, float64(-1), m["restrictedBid"], "宣言はまだ始まっていない")

	players := m["players"].([]any)
	require.Len(t, players, domain.EstimationPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.EstimationHandSize), human["cardCount"])
	assert.Equal(t, float64(-1), human["bid"], "未宣言は -1")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **禁止値はワイヤに載る。** 載せないとクライアントが押せない宣言を出す。
func TestEstimationWebPresenterSurfacesTheRestrictedBid(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.SetDealerIdxForTest(1)
	e.CpuSelectTrump()
	e.SetBidsForTest(map[int]int{1: 4, 2: 4, 3: 4})
	e.SetBidPlayerIdxForTest(0)

	m := decodeEstimation(t, p.Output(e, nil))
	assert.Equal(t, float64(1), m["restrictedBid"], "4+4+4=12 なので 1 が禁止")
	assert.Equal(t, "estimation.bid.restricted", m["messageCode"])
	assert.Equal(t, "1", m["messageParams"].(map[string]any)["n"])
}

// 宣言の種類がワイヤに載る。Dash と Risk の両方を踏む。
func TestEstimationWebPresenterSurfacesCallTypes(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(domain.CardDesignSpade))
	e.SetBidsForTest(map[int]int{0: 0, 1: 6})
	e.GetPlayer(1).SetCallType(domain.EstimationCallRisk)

	players := decodeEstimation(t, p.Output(e, nil))["players"].([]any)
	assert.Equal(t, float64(domain.EstimationCallDash), players[0].(map[string]any)["callType"])
	assert.Equal(t, float64(domain.EstimationCallRisk), players[1].(map[string]any)["callType"])
	assert.Equal(t, float64(0), players[0].(map[string]any)["bid"])
	assert.Equal(t, float64(6), players[1].(map[string]any)["bid"])
}

// 得点はラウンド増減と累計の両方が出る。
func TestEstimationWebPresenterSurfacesScores(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.GetPlayer(0).SetRoundScore(-14)
	e.GetPlayer(0).SetTotalScore(31)

	human := decodeEstimation(t, p.Output(e, nil))["players"].([]any)[0].(map[string]any)
	assert.Equal(t, float64(-14), human["roundScore"])
	assert.Equal(t, float64(31), human["totalScore"])
}

// **親のときは選ばせ、そうでなければ待たせる。** 両側を踏む。
func TestEstimationWebPresenterTrumpMessages(t *testing.T) {
	p := new(EstimationWebPresenter)

	dealer := newEstimationForWeb(t)
	dealer.SetDealerIdxForTest(0)
	assert.Equal(t, "estimation.trump.choose", decodeEstimation(t, p.Output(dealer, nil))["messageCode"])

	other := newEstimationForWeb(t)
	other.SetDealerIdxForTest(2)
	assert.Equal(t, "estimation.trump.wait", decodeEstimation(t, p.Output(other, nil))["messageCode"])
}

func TestEstimationWebPresenterBidMessage(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(domain.CardDesignSpade))

	assert.Equal(t, "estimation.bid.choose", decodeEstimation(t, p.Output(e, nil))["messageCode"])
}

func TestEstimationWebPresenterRoundEndMessage(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.SetPhaseForTest(domain.EstimationPhaseRoundEnd)
	e.GetPlayer(0).SetRoundScore(14)

	m := decodeEstimation(t, p.Output(e, nil))
	assert.Equal(t, "estimation.roundEnd", m["messageCode"])
	assert.Equal(t, "14", m["messageParams"].(map[string]any)["score"])
}

func TestEstimationWebPresenterResultMessage(t *testing.T) {
	p := new(EstimationWebPresenter)

	for _, tc := range []struct {
		name     string
		scores   [4]int
		wantCode string
	}{
		{"you win", [4]int{50, 10, 10, 10}, "estimation.result.win"},
		{"a cpu wins", [4]int{10, 50, 10, 10}, "estimation.result.lose"},
		{"a tie", [4]int{20, 20, 20, 20}, "estimation.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEstimationForWeb(t)
			for i, s := range tc.scores {
				e.GetPlayer(i).SetTotalScore(s)
			}
			e.FinishGameForTest()

			assert.Equal(t, tc.wantCode, decodeEstimation(t, p.Output(e, nil))["messageCode"])
		})
	}
}

func TestEstimationWebPresenterError(t *testing.T) {
	p := new(EstimationWebPresenter)
	m := decodeEstimation(t, p.Output(newEstimationForWeb(t), assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestEstimationWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.GiveUp()

	m := decodeEstimation(t, p.Output(e, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

// 切り札選択・宣言のヒントは札を指さず、値を運ぶ。
func TestEstimationWebPresenterHintCarriesAValue(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.SetDealerIdxForTest(0)

	hint, ok := decodeEstimation(t, p.HintOutput(e))["hint"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, hint["cardIndex"], "切り札選択では札を指さない")
	assert.Equal(t, "estimationSelectTrump", hint["reason"])
	assert.GreaterOrEqual(t, hint["value"], float64(domain.CardDesignSpade))
}

func TestEstimationWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.GiveUp()
	assert.Nil(t, decodeEstimation(t, p.HintOutput(e))["hint"])
}

func TestEstimationWebPresenterConfigSurfaces(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)
	e.SetConfig(domain.EstimationConfig{Rounds: 9})

	assert.Equal(t, float64(9), decodeEstimation(t, p.Output(e, nil))["config"].(map[string]any)["rounds"])
}

func TestEstimationWebPresenterActionLogOutput(t *testing.T) {
	p := new(EstimationWebPresenter)
	e := newEstimationForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(e)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	e.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(e)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
