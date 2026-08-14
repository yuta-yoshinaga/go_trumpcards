//go:build test

package presenter

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newHoneymoonBridgeForCui(t *testing.T) *domain.HoneymoonBridge {
	t.Helper()
	h := domain.NewDefaultHoneymoonBridge()
	h.Reset()
	return h
}

func TestHoneymoonBridgeCuiPresenterOutput(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)
	out := p.Output(newHoneymoonBridgeForCui(t), nil)

	assert.Contains(t, out, i18n.T("honeymoonbridge.helpTitle"))
	assert.Contains(t, out, fixedPart("honeymoonbridge.header"))
	// **前半と後半で意味が変わる。** 規則そのものを毎回書く。
	assert.Contains(t, out, i18n.T("honeymoonbridge.rule"))
	// 「獲得」はルール行にも出るので、席行の並び（獲得N 累計N）で数える。
	assert.Len(t, regexp.MustCompile(`獲得\d+ 累計\d+`).FindAllString(out, -1),
		domain.HoneymoonBridgePlayerCnt, "両者の席行に獲得数と累計が出る")
}

// **引き合いのあいだは山札の残りを出す。** 何枚引けるかが読めないと打てない。
func TestHoneymoonBridgeCuiPresenterStockLine(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)

	drawing := newHoneymoonBridgeForCui(t)
	out := p.Output(drawing, nil)
	assert.Contains(t, out, fixedPart("honeymoonbridge.stockLine"))
	assert.Contains(t, out, strconv.Itoa(domain.HoneymoonBridgeStockSize))

	bidding := newHoneymoonBridgeForCui(t)
	bidding.SetPhaseForTest(domain.HoneymoonBridgePhaseBid)
	assert.NotContains(t, p.Output(bidding, nil), fixedPart("honeymoonbridge.stockLine"),
		"山札を使い切ってから競るので出さない")
}

// 契約は未決定と確定の両側を踏む。
func TestHoneymoonBridgeCuiPresenterContractLine(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)

	undecided := newHoneymoonBridgeForCui(t)
	out := p.Output(undecided, nil)
	assert.Contains(t, out, i18n.T("honeymoonbridge.contractUndecided"))

	decided := newHoneymoonBridgeForCui(t)
	decided.SetPhaseForTest(domain.HoneymoonBridgePhasePlay)
	decided.SetContractForTest(0, 4, domain.CardDesignHeart)
	out = p.Output(decided, nil)
	assert.Contains(t, out, fixedPart("honeymoonbridge.contractLine"))
	assert.Contains(t, out, i18n.T("honeymoonbridge.suitHeart"))
	// **必要トリック数を出す。** 「4♥」だけでは何トリック要るか分からない。
	assert.Contains(t, out, strconv.Itoa(domain.HoneymoonBridgeBookTricks+4))
	assert.Contains(t, out, i18n.T("honeymoonbridge.roleDeclarer"))
	assert.NotContains(t, out, i18n.T("honeymoonbridge.contractUndecided"))
}

// **4 スートとノートランプの 5 つすべてに名前がある。**
func TestHoneymoonBridgeCuiPresenterNamesEverySuit(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)
	for suit, key := range map[int]string{
		domain.CardDesignSpade:   "honeymoonbridge.suitSpade",
		domain.CardDesignClover:  "honeymoonbridge.suitClover",
		domain.CardDesignHeart:   "honeymoonbridge.suitHeart",
		domain.CardDesignDiamond: "honeymoonbridge.suitDiamond",
		0:                        "honeymoonbridge.suitNoTrump",
	} {
		h := newHoneymoonBridgeForCui(t)
		h.SetPhaseForTest(domain.HoneymoonBridgePhasePlay)
		h.SetContractForTest(0, 1, suit)
		assert.Contains(t, p.Output(h, nil), i18n.T(key), "suit=%d", suit)
	}
}

// **通る最小の宣言を出す。** これが無いと拒否される値を打ち込むことになる。
func TestHoneymoonBridgeCuiPresenterBidPrompt(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)

	h := newHoneymoonBridgeForCui(t)
	h.SetPhaseForTest(domain.HoneymoonBridgePhaseBid)
	h.SetCurrentPlayerIdxForTest(0)
	out := p.Output(h, nil)
	assert.Contains(t, out, fixedPart("honeymoonbridge.promptBid"))
	assert.NotContains(t, out, i18n.T("honeymoonbridge.promptBidWait"))

	// **上限に張り付いたら pass しかない。**
	h.SetContractForTest(1, domain.HoneymoonBridgeMaxLevel, 0)
	h.SetCurrentPlayerIdxForTest(0)
	out = p.Output(h, nil)
	assert.Contains(t, out, i18n.T("honeymoonbridge.promptBidCapped"))
	assert.NotContains(t, out, fixedPart("honeymoonbridge.promptBid"),
		"通る宣言が無いのに最小値を案内してはいけない")

	h.SetCurrentPlayerIdxForTest(1)
	assert.Contains(t, p.Output(h, nil), i18n.T("honeymoonbridge.promptBidWait"))
}

func TestHoneymoonBridgeCuiPresenterPrompts(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)

	playing := newHoneymoonBridgeForCui(t)
	playing.SetPhaseForTest(domain.HoneymoonBridgePhasePlay)
	assert.Contains(t, p.Output(playing, nil), i18n.T("honeymoonbridge.promptPlay"))

	roundEnd := newHoneymoonBridgeForCui(t)
	roundEnd.SetPhaseForTest(domain.HoneymoonBridgePhaseRoundEnd)
	assert.Contains(t, p.Output(roundEnd, nil), i18n.T("honeymoonbridge.promptNext"))
}

func TestHoneymoonBridgeCuiPresenterGameEndBanners(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)
	for _, tc := range []struct {
		name   string
		winner int
		key    string
	}{
		{"あなたの勝ち", 0, "honeymoonbridge.gameEndYou"},
		{"相手の勝ち", 1, "honeymoonbridge.gameEndCpu"},
		{"同点", -1, "honeymoonbridge.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHoneymoonBridgeForCui(t)
			if tc.winner >= 0 {
				h.GetPlayer(tc.winner).SetScore(1)
			}
			h.FinishGameForTest()
			out := p.Output(h, nil)
			assert.Contains(t, out, fixedPart(tc.key))
			assert.NotContains(t, out, i18n.T("honeymoonbridge.promptPlay"), "終局後は促さない")
		})
	}
}

func TestHoneymoonBridgeCuiPresenterShowsErrors(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)
	assert.Contains(t, p.Output(newHoneymoonBridgeForCui(t), assert.AnError), assert.AnError.Error())
}

func TestHoneymoonBridgeCuiPresenterHintOutput(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)

	// 引き合いは札を指す。
	h := newHoneymoonBridgeForCui(t)
	h.SetCurrentPlayerIdxForTest(0)
	out := p.HintOutput(h)
	assert.Contains(t, out, fixedPart("honeymoonbridge.hintCard"))
	assert.Contains(t, out, i18n.T("honeymoonbridge.hintReasonDraw"))

	// **競りの助言は契約を指す。**
	h.SetPhaseForTest(domain.HoneymoonBridgePhaseBid)
	out = p.HintOutput(h)
	assert.NotContains(t, out, fixedPart("honeymoonbridge.hintCard"))

	// pass を勧める側も踏む。
	h.SetContractForTest(1, domain.HoneymoonBridgeMaxLevel, 0)
	h.SetCurrentPlayerIdxForTest(0)
	out = p.HintOutput(h)
	assert.Contains(t, out, i18n.T("honeymoonbridge.hintReasonPass"))

	// 本番の助言は取りにいく。
	play := newHoneymoonBridgeForCui(t)
	play.SetPhaseForTest(domain.HoneymoonBridgePhasePlay)
	play.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(play), i18n.T("honeymoonbridge.hintReasonWinTrick"))

	play.FinishGameForTest()
	assert.Contains(t, p.HintOutput(play), i18n.T("honeymoonbridge.hintNone"))
}

func TestHoneymoonBridgeCuiPresenterActionLogOutput(t *testing.T) {
	p := new(HoneymoonBridgeCuiPresenter)
	h := newHoneymoonBridgeForCui(t)
	require.NotNil(t, h)
	h.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(h))
}
