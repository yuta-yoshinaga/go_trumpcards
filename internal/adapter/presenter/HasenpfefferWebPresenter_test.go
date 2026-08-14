//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newHasenpfefferForWeb(t *testing.T) *domain.Hasenpfeffer {
	t.Helper()
	h := domain.NewDefaultHasenpfeffer()
	h.Reset()
	return h
}

func decodeHasenpfeffer(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestHasenpfefferWebPresenterOutput(t *testing.T) {
	p := new(HasenpfefferWebPresenter)
	m := decodeHasenpfeffer(t, p.Output(newHasenpfefferForWeb(t), nil))

	assert.Equal(t, float64(domain.HasenpfefferPhaseBid), m["phase"])
	assert.Equal(t, float64(1), m["handNumber"])
	assert.Equal(t, float64(0), m["trumpSuit"], "まだ宣言されていない")
	assert.Equal(t, float64(-1), m["declarerIdx"])
	assert.Equal(t, float64(1), m["blindSize"], "伏せ札 1 枚")
	assert.Equal(t, float64(-1), m["winnerTeam"])
	// **押せない宣言額を出さないための下限。**
	assert.Equal(t, float64(domain.HasenpfefferMinBid), m["minBid"])

	players := m["players"].([]any)
	require.Len(t, players, domain.HasenpfefferPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.HasenpfefferHandSize), human["cardCount"])
	assert.Equal(t, float64(-1), human["bid"], "未宣言は -1")
	assert.Equal(t, float64(0), human["team"])
	assert.Equal(t, float64(0), players[2].(map[string]any)["team"], "0 と 2 が味方")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **上限が立ったら 0 を返す。** 同額での横取りをクライアントに出させない。
func TestHasenpfefferWebPresenterCapsTheBid(t *testing.T) {
	p := new(HasenpfefferWebPresenter)

	fresh := newHasenpfefferForWeb(t)
	assert.Equal(t, float64(domain.HasenpfefferMinBid),
		decodeHasenpfeffer(t, p.Output(fresh, nil))["minBid"])

	raised := newHasenpfefferForWeb(t)
	raised.SetContractForTest(2, 4)
	assert.Equal(t, float64(5), decodeHasenpfeffer(t, p.Output(raised, nil))["minBid"])

	capped := newHasenpfefferForWeb(t)
	capped.SetContractForTest(2, domain.HasenpfefferMaxBid)
	out := decodeHasenpfeffer(t, p.Output(capped, nil))
	assert.Equal(t, float64(0), out["minBid"], "もう宣言できない")
}

// **義務競りはワイヤに載せる。** 降りるボタンを出してはいけない場面がある。
func TestHasenpfefferWebPresenterCarriesTheCompulsoryBid(t *testing.T) {
	p := new(HasenpfefferWebPresenter)
	h := newHasenpfefferForWeb(t)
	h.SetDealerIdxForTest(0)
	h.SetCurrentPlayerIdxForTest(1)
	require.NoError(t, h.BidForTest(1, 0))
	require.NoError(t, h.BidForTest(2, 0))
	require.NoError(t, h.BidForTest(3, 0))

	out := decodeHasenpfeffer(t, p.Output(h, nil))
	assert.True(t, out["mustBid"].(bool), "親は降りられない")
	assert.Equal(t, "hasenpfeffer.bid.must", out["messageCode"])

	// 負のコントロール: 誰かが落札していれば降りられる。
	other := newHasenpfefferForWeb(t)
	other.SetDealerIdxForTest(0)
	other.SetCurrentPlayerIdxForTest(1)
	require.NoError(t, other.BidForTest(1, 4))
	require.NoError(t, other.BidForTest(2, 0))
	require.NoError(t, other.BidForTest(3, 0))
	assert.False(t, decodeHasenpfeffer(t, p.Output(other, nil))["mustBid"].(bool))
}

// **落札額と落札者はワイヤに載る。**
func TestHasenpfefferWebPresenterCarriesTheContract(t *testing.T) {
	p := new(HasenpfefferWebPresenter)
	h := newHasenpfefferForWeb(t)
	h.SetContractForTest(2, 5)
	h.SetPhaseForTest(domain.HasenpfefferPhaseDiscard)

	out := decodeHasenpfeffer(t, p.Output(h, nil))
	assert.Equal(t, float64(2), out["declarerIdx"])
	assert.Equal(t, float64(5), out["contract"])
	assert.Equal(t, "hasenpfeffer.discard.wait", out["messageCode"], "落札者が CPU なら待ち")
}

// **達成と失敗で別のコードを返す。** 盤面からは読めない。
func TestHasenpfefferWebPresenterReportsTheHandOutcome(t *testing.T) {
	for _, tc := range []struct {
		name    string
		tricks  int
		code    string
		euchred bool
	}{
		{"達成", 5, "hasenpfeffer.handEnd.made", false},
		{"失敗", 2, "hasenpfeffer.handEnd.euchred", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := new(HasenpfefferWebPresenter)
			h := newHasenpfefferForWeb(t)
			h.SetContractForTest(0, 4)
			h.SetPhaseForTest(domain.HasenpfefferPhasePlay)
			h.GiveTricksForTest(0, tc.tricks)
			h.GiveTricksForTest(1, domain.HasenpfefferTricksPerRound-tc.tricks)
			h.FinishHandForTest()

			out := decodeHasenpfeffer(t, p.Output(h, nil))
			assert.Equal(t, tc.code, out["messageCode"])
			assert.Equal(t, tc.euchred, out["lastHandEuchred"].(bool))
			assert.Equal(t, float64(tc.tricks), out["lastHandTricks"])
		})
	}
}

func TestHasenpfefferWebPresenterGameEnd(t *testing.T) {
	p := new(HasenpfefferWebPresenter)
	for _, tc := range []struct {
		name   string
		t0, t1 int
		code   string
		want   float64
	}{
		{"team0", 12, 3, "hasenpfeffer.result.team0", 0},
		{"team1", 3, 12, "hasenpfeffer.result.team1", 1},
		{"tie", 12, 12, "hasenpfeffer.result.tie", -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHasenpfefferForWeb(t)
			h.SetScoreForTestUse(0, tc.t0)
			h.SetScoreForTestUse(1, tc.t1)
			h.FinishGameForTest()

			out := decodeHasenpfeffer(t, p.Output(h, nil))
			assert.True(t, out["gameEndFlag"].(bool))
			assert.Equal(t, tc.want, out["winnerTeam"])
			assert.Equal(t, tc.code, out["messageCode"])
		})
	}
}

func TestHasenpfefferWebPresenterSurfacesErrors(t *testing.T) {
	p := new(HasenpfefferWebPresenter)
	h := newHasenpfefferForWeb(t)
	err := h.PlayerPlay(0)
	require.Error(t, err)

	out := decodeHasenpfeffer(t, p.Output(h, err))
	assert.Equal(t, err.Error(), out["message"])
	assert.Empty(t, out["messageCode"])
}

func TestHasenpfefferWebPresenterHintAndLog(t *testing.T) {
	p := new(HasenpfefferWebPresenter)
	h := newHasenpfefferForWeb(t)
	h.SetDealerIdxForTest(3)
	h.SetCurrentPlayerIdxForTest(0)

	// **競りの助言は札ではなく額を指す。**
	hint := decodeHasenpfeffer(t, p.HintOutput(h))["hint"].(map[string]any)
	assert.NotContains(t, hint, "cardIndex")
	assert.NotEmpty(t, hint["reason"])

	var logOut map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(h)), &logOut))
	assert.Contains(t, logOut, "entries")
}
