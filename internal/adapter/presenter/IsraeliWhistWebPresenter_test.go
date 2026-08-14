//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newIsraeliWhistForWeb(t *testing.T) *domain.IsraeliWhist {
	t.Helper()
	w := domain.NewDefaultIsraeliWhist()
	w.Reset()
	return w
}

func decodeIsraeliWhist(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &m))
	return m
}

func TestIsraeliWhistWebPresenterOutput(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)

	m := decodeIsraeliWhist(t, p.Output(w, nil))

	assert.Equal(t, float64(domain.IsraeliWhistPhaseAuction), m["phase"], "配り直後はオークション")
	assert.Equal(t, float64(1), m["roundNumber"])
	assert.Equal(t, float64(0), m["trumpSuit"], "切り札はオークションで決まる")
	assert.Equal(t, float64(-1), m["declarerIdx"])
	assert.Equal(t, float64(-1), m["restrictedBid"])

	players := m["players"].([]any)
	require.Len(t, players, domain.IsraeliWhistPlayerCnt)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.IsraeliWhistHandSize), human["cardCount"])
	assert.Equal(t, float64(-1), human["bid"], "未宣言は -1")
	assert.Equal(t, float64(-1), human["auctionBid"], "未入札は -1")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **2 段階ぶんの状態がワイヤに載る。** 片方でも欠けるとクライアントが
// どちらの段にいるのか判断できない。
func TestIsraeliWhistWebPresenterBothStagesSurface(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)
	w.GetPlayer(0).SetAuction(8, domain.CardDesignClover)
	w.GetPlayer(1).SetPassed(true)
	w.SetDeclarerForTest(0, 8, domain.CardDesignClover)
	w.CloseAuctionForTest()
	w.SetBidsForTest(map[int]int{0: 9})

	m := decodeIsraeliWhist(t, p.Output(w, nil))
	assert.Equal(t, float64(domain.CardDesignClover), m["trumpSuit"])
	assert.Equal(t, float64(8), m["highBid"])
	assert.Equal(t, float64(domain.CardDesignClover), m["highSuit"])
	assert.Equal(t, float64(0), m["declarerIdx"])

	players := m["players"].([]any)
	assert.Equal(t, float64(8), players[0].(map[string]any)["auctionBid"])
	assert.Equal(t, float64(domain.CardDesignClover), players[0].(map[string]any)["auctionSuit"])
	assert.Equal(t, float64(9), players[0].(map[string]any)["bid"])
	assert.True(t, players[1].(map[string]any)["passed"].(bool))
}

// **落札者のノルマはワイヤに載る。** 載せないとクライアントが押せない宣言を出す。
func TestIsraeliWhistWebPresenterSurfacesTheQuota(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)

	declarer := newIsraeliWhistForWeb(t)
	declarer.SetDeclarerForTest(0, 9, domain.CardDesignSpade)
	declarer.CloseAuctionForTest()
	declarer.SetBidPlayerIdxForTest(0)
	m := decodeIsraeliWhist(t, p.Output(declarer, nil))
	assert.Equal(t, float64(9), m["minimumBid"])
	assert.Equal(t, "israeliwhist.bid.quota", m["messageCode"])

	// 負のコントロール: 落札者でなければノルマは 0。
	other := newIsraeliWhistForWeb(t)
	other.SetDeclarerForTest(2, 9, domain.CardDesignSpade)
	other.CloseAuctionForTest()
	other.SetBidPlayerIdxForTest(0)
	m2 := decodeIsraeliWhist(t, p.Output(other, nil))
	assert.Equal(t, float64(0), m2["minimumBid"])
	assert.Equal(t, "israeliwhist.bid.choose", m2["messageCode"])
}

func TestIsraeliWhistWebPresenterSurfacesTheRestrictedBid(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)
	w.SetDeclarerForTest(1, 5, domain.CardDesignSpade)
	w.CloseAuctionForTest()
	w.SetBidsForTest(map[int]int{1: 5, 2: 4, 3: 3})
	w.SetBidPlayerIdxForTest(0)

	m := decodeIsraeliWhist(t, p.Output(w, nil))
	assert.Equal(t, float64(1), m["restrictedBid"])
	assert.Equal(t, "israeliwhist.bid.restricted", m["messageCode"])
}

// **入札できる番と待つ番で案内が変わる。** 両側を踏む。
func TestIsraeliWhistWebPresenterAuctionMessages(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)

	mine := newIsraeliWhistForWeb(t)
	mine.SetAuctionPlayerIdxForTest(0)
	assert.Equal(t, "israeliwhist.auction.choose", decodeIsraeliWhist(t, p.Output(mine, nil))["messageCode"])

	theirs := newIsraeliWhistForWeb(t)
	theirs.SetAuctionPlayerIdxForTest(2)
	assert.Equal(t, "israeliwhist.auction.wait", decodeIsraeliWhist(t, p.Output(theirs, nil))["messageCode"])
}

func TestIsraeliWhistWebPresenterRoundEndMessage(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)
	w.SetPhaseForTest(domain.IsraeliWhistPhaseRoundEnd)
	w.GetPlayer(0).SetRoundScore(35)

	m := decodeIsraeliWhist(t, p.Output(w, nil))
	assert.Equal(t, "israeliwhist.roundEnd", m["messageCode"])
	assert.Equal(t, "35", m["messageParams"].(map[string]any)["score"])
}

func TestIsraeliWhistWebPresenterResultMessage(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)

	for _, tc := range []struct {
		name     string
		scores   [4]int
		wantCode string
	}{
		{"you win", [4]int{80, 10, 10, 10}, "israeliwhist.result.win"},
		{"a cpu wins", [4]int{10, 80, 10, 10}, "israeliwhist.result.lose"},
		{"a tie", [4]int{30, 30, 30, 30}, "israeliwhist.result.tie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newIsraeliWhistForWeb(t)
			for i, s := range tc.scores {
				w.GetPlayer(i).SetTotalScore(s)
			}
			w.FinishGameForTest()

			assert.Equal(t, tc.wantCode, decodeIsraeliWhist(t, p.Output(w, nil))["messageCode"])
		})
	}
}

func TestIsraeliWhistWebPresenterError(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	m := decodeIsraeliWhist(t, p.Output(newIsraeliWhistForWeb(t), assert.AnError))
	assert.Equal(t, assert.AnError.Error(), m["message"])
	assert.Empty(t, m["messageCode"])
}

func TestIsraeliWhistWebPresenterValidPlaysNeverNull(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)
	w.GiveUp()

	m := decodeIsraeliWhist(t, p.Output(w, nil))
	require.NotNil(t, m["validPlays"])
	assert.IsType(t, []any{}, m["validPlays"])
}

// オークションのヒントは札を指さず、数とスートを運ぶ。
func TestIsraeliWhistWebPresenterHintCarriesBidAndSuit(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)
	w.SetAuctionPlayerIdxForTest(0)

	hint, ok := decodeIsraeliWhist(t, p.HintOutput(w))["hint"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, hint["cardIndex"], "オークションでは札を指さない")
	assert.Contains(t, []any{"israeliwhistAuctionBid", "israeliwhistAuctionPass"}, hint["reason"])
}

func TestIsraeliWhistWebPresenterHintOmittedAfterGameEnd(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)
	w.GiveUp()
	assert.Nil(t, decodeIsraeliWhist(t, p.HintOutput(w))["hint"])
}

func TestIsraeliWhistWebPresenterConfigSurfaces(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)
	w.SetConfig(domain.IsraeliWhistConfig{Rounds: 6})

	assert.Equal(t, float64(6), decodeIsraeliWhist(t, p.Output(w, nil))["config"].(map[string]any)["rounds"])
}

func TestIsraeliWhistWebPresenterActionLogOutput(t *testing.T) {
	p := new(IsraeliWhistWebPresenter)
	w := newIsraeliWhistForWeb(t)

	var during map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(w)), &during))
	assert.Empty(t, during["entries"], "進行中は空")

	w.GiveUp()
	var after map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(w)), &after))
	assert.NotEmpty(t, after["entries"], "終局後は残る")
}
