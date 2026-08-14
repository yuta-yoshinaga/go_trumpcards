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

func newShelemForWeb(t *testing.T) *domain.Shelem {
	t.Helper()
	s := domain.NewDefaultShelem()
	s.Reset()
	return s
}

func decodeShelem(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestShelemWebPresenterOutput(t *testing.T) {
	p := new(ShelemWebPresenter)
	s := newShelemForWeb(t)

	m := decodeShelem(t, p.Output(s, nil))

	assert.Equal(t, float64(domain.ShelemPhaseBid), m["phase"], "配り直後は競り")
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(0), m["trumpSuit"], "切り札はまだ決まっていない")
	assert.Equal(t, float64(-1), m["declarerIdx"])
	assert.Equal(t, float64(domain.ShelemWidowSize), m["widowSize"])
	assert.Equal(t, float64(domain.ShelemWidowSize), m["discardCount"])

	players := m["players"].([]any)
	require.Len(t, players, domain.ShelemPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.ShelemHandSize), human["cardCount"])
	assert.Equal(t, float64(-1), human["bid"], "未入札は -1")
	assert.Equal(t, float64(0), human["team"])
	assert.Equal(t, float64(0), players[2].(map[string]any)["team"], "0 と 2 が味方")
}

// **次に出せる最小額をワイヤに載せる。** 載せないと上回らない入札を出してしまう。
func TestShelemWebPresenterSurfacesTheNextBid(t *testing.T) {
	p := new(ShelemWebPresenter)

	fresh := newShelemForWeb(t)
	assert.Equal(t, float64(domain.ShelemMinBid), decodeShelem(t, p.Output(fresh, nil))["minBid"])

	raised := newShelemForWeb(t)
	raised.SetContractForTest(2, 90, false)
	assert.Equal(t, float64(90+domain.ShelemBidStep), decodeShelem(t, p.Output(raised, nil))["minBid"])

	// **上限を超えない。** 165 が立っていたら 165 のまま。
	capped := newShelemForWeb(t)
	capped.SetContractForTest(2, domain.ShelemMaxBid, false)
	assert.Equal(t, float64(domain.ShelemMaxBid), decodeShelem(t, p.Output(capped, nil))["minBid"])
}

// 競りの状態が席ごとにワイヤに載る。
func TestShelemWebPresenterSurfacesBiddingState(t *testing.T) {
	p := new(ShelemWebPresenter)
	s := newShelemForWeb(t)
	s.SetContractForTest(0, 95, false)
	s.GetPlayer(0).SetBid(95)
	s.GetPlayer(1).SetPassed(true)

	m := decodeShelem(t, p.Output(s, nil))
	assert.Equal(t, float64(95), m["contract"])
	assert.Equal(t, float64(0), m["declarerIdx"])
	assert.False(t, m["shelemBid"].(bool))

	players := m["players"].([]any)
	assert.Equal(t, float64(95), players[0].(map[string]any)["bid"])
	assert.True(t, players[1].(map[string]any)["passed"].(bool))
}

// **Shelem 宣言はワイヤに載る。** 精算の意味がまるごと変わるため。
func TestShelemWebPresenterSurfacesShelemBid(t *testing.T) {
	p := new(ShelemWebPresenter)
	s := newShelemForWeb(t)
	s.SetContractForTest(0, domain.ShelemMaxBid, true)
	s.GetPlayer(0).SetDeclaredShelem(true)

	m := decodeShelem(t, p.Output(s, nil))
	assert.True(t, m["shelemBid"].(bool))
	assert.True(t, m["players"].([]any)[0].(map[string]any)["declaredShelem"].(bool))
}

// **入札できる番と待つ番で案内が変わる。** 両側を踏む。
func TestShelemWebPresenterBidMessages(t *testing.T) {
	p := new(ShelemWebPresenter)

	mine := newShelemForWeb(t)
	mine.SetBidPlayerIdxForTest(0)
	m := decodeShelem(t, p.Output(mine, nil))
	assert.Equal(t, "shelem.bid.choose", m["messageCode"])
	assert.Equal(t, strconv.Itoa(domain.ShelemMinBid), m["messageParams"].(map[string]any)["min"])

	theirs := newShelemForWeb(t)
	theirs.SetBidPlayerIdxForTest(2)
	assert.Equal(t, "shelem.bid.wait", decodeShelem(t, p.Output(theirs, nil))["messageCode"])
}

// **捨て札は落札者だけ。** 両側を踏む。
func TestShelemWebPresenterDiscardMessages(t *testing.T) {
	p := new(ShelemWebPresenter)

	mine := newShelemForWeb(t)
	mine.SetContractForTest(0, 90, false)
	mine.SetPhaseForTest(domain.ShelemPhaseDiscard)
	assert.Equal(t, "shelem.discard.choose", decodeShelem(t, p.Output(mine, nil))["messageCode"])

	theirs := newShelemForWeb(t)
	theirs.SetContractForTest(2, 90, false)
	theirs.SetPhaseForTest(domain.ShelemPhaseDiscard)
	assert.Equal(t, "shelem.discard.wait", decodeShelem(t, p.Output(theirs, nil))["messageCode"])
}

// **Shelem と通常契約で結末の意味が違う。** 別のメッセージにする。
func TestShelemWebPresenterRoundEndMessages(t *testing.T) {
	p := new(ShelemWebPresenter)

	normal := newShelemForWeb(t)
	normal.SetContractForTest(0, 90, false)
	normal.SetRoundPointsForTest(0, 95)
	normal.SetPhaseForTest(domain.ShelemPhaseRoundEnd)
	m := decodeShelem(t, p.Output(normal, nil))
	assert.Equal(t, "shelem.roundEnd", m["messageCode"])
	assert.Equal(t, "90", m["messageParams"].(map[string]any)["contract"])
	assert.Equal(t, "95", m["messageParams"].(map[string]any)["got"])

	shelem := newShelemForWeb(t)
	shelem.SetContractForTest(0, domain.ShelemMaxBid, true)
	shelem.SetPhaseForTest(domain.ShelemPhaseRoundEnd)
	assert.Equal(t, "shelem.roundEnd.shelem", decodeShelem(t, p.Output(shelem, nil))["messageCode"])
}

func TestShelemWebPresenterResultMessage(t *testing.T) {
	p := new(ShelemWebPresenter)

	for _, tc := range []struct {
		name     string
		t0, t1   int
		wantCode string
	}{
		{"your team wins", 520, 300, "shelem.result.team0"},
		{"they win", 300, 520, "shelem.result.team1"},
		{"a tie", 500, 500, "shelem.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newShelemForWeb(t)
			s.SetScoreForTestUse(0, tc.t0)
			s.SetScoreForTestUse(1, tc.t1)
			s.FinishGameForTest()

			assert.Equal(t, tc.wantCode, decodeShelem(t, p.Output(s, nil))["messageCode"])
		})
	}
}

func TestShelemWebPresenterError(t *testing.T) {
	p := new(ShelemWebPresenter)
	m := decodeShelem(t, p.Output(newShelemForWeb(t), assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestShelemWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(ShelemWebPresenter)
	s := newShelemForWeb(t)
	s.GiveUp()

	m := decodeShelem(t, p.Output(s, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

// 競りのヒントは札を指さず、点数を運ぶ。
func TestShelemWebPresenterHintCarriesAValue(t *testing.T) {
	p := new(ShelemWebPresenter)
	s := newShelemForWeb(t)
	s.SetBidPlayerIdxForTest(0)

	hint, ok := decodeShelem(t, p.HintOutput(s))["hint"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, hint["cardIndex"], "競りでは札を指さない")
	assert.Contains(t, []any{"shelemBid", "shelemPass"}, hint["reason"])
}

func TestShelemWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(ShelemWebPresenter)
	s := newShelemForWeb(t)
	s.GiveUp()
	assert.Nil(t, decodeShelem(t, p.HintOutput(s))["hint"])
}

func TestShelemWebPresenterConfigSurfaces(t *testing.T) {
	p := new(ShelemWebPresenter)
	s := newShelemForWeb(t)
	s.SetConfig(domain.ShelemConfig{Target: 700})

	assert.Equal(t, float64(700), decodeShelem(t, p.Output(s, nil))["config"].(map[string]any)["target"])
}

func TestShelemWebPresenterActionLogOutput(t *testing.T) {
	p := new(ShelemWebPresenter)
	s := newShelemForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(s)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	s.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(s)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
