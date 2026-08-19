//go:build test

package presenter

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newHasenpfefferForCui(t *testing.T) *domain.Hasenpfeffer {
	t.Helper()
	h := domain.NewDefaultHasenpfeffer()
	h.Reset()
	return h
}

func TestHasenpfefferCuiPresenterOutput(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)
	out := p.Output(newHasenpfefferForCui(t), nil)

	assert.Contains(t, out, i18n.T("hasenpfeffer.helpTitle"))
	assert.Contains(t, out, fixedPart("hasenpfeffer.header"))
	// **ジョーカーが最強という序列を毎回書く。** 知らないと打ち方が変わる。
	assert.Contains(t, out, i18n.T("hasenpfeffer.rule"))
	assert.Contains(t, out, fixedPart("hasenpfeffer.scoreLine"))
	assert.Len(t, regexp.MustCompile(`\[T[01]\]`).FindAllString(out, -1),
		domain.HasenpfefferPlayerCnt, "全員の席行が出る")
}

// 伏せ札・未宣言・確定の 3 状態を踏む。
func TestHasenpfefferCuiPresenterTrumpLine(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)

	bidding := newHasenpfefferForCui(t)
	assert.Contains(t, p.Output(bidding, nil), fixedPart("hasenpfeffer.blindLine"),
		"競り中は伏せ札があることを言う")

	// **伏せ札を実際に取り込ませる。** SetContractForTest だけでは伏せ札が
	// 残ったままになり、表示が blindLine のほうに落ちる。
	taken := newHasenpfefferForCui(t)
	taken.SetDealerIdxForTest(3)
	taken.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, taken.BidForTest(0, 4))
	require.NoError(t, taken.BidForTest(1, 0))
	require.NoError(t, taken.BidForTest(2, 0))
	require.NoError(t, taken.BidForTest(3, 0))
	require.Zero(t, taken.GetBlindSize(), "落札者が取り込んだ")
	assert.Contains(t, p.Output(taken, nil), i18n.T("hasenpfeffer.trumpUndecided"))

	declared := newHasenpfefferForCui(t)
	declared.SetContractForTest(1, 4)
	declared.SetTrumpSuitForTest(domain.CardDesignHeart)
	declared.SetPhaseForTest(domain.HasenpfefferPhasePlay)
	out := p.Output(declared, nil)
	assert.Contains(t, out, fixedPart("hasenpfeffer.trumpLine"))
	assert.Contains(t, out, i18n.T("hasenpfeffer.suitHeart"))
}

// **4 スートすべてに名前がある。**
func TestHasenpfefferCuiPresenterNamesEverySuit(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)
	for suit, key := range map[int]string{
		domain.CardDesignSpade:   "hasenpfeffer.suitSpade",
		domain.CardDesignClover:  "hasenpfeffer.suitClover",
		domain.CardDesignHeart:   "hasenpfeffer.suitHeart",
		domain.CardDesignDiamond: "hasenpfeffer.suitDiamond",
	} {
		h := newHasenpfefferForCui(t)
		h.SetContractForTest(1, 4)
		h.SetTrumpSuitForTest(suit)
		h.SetPhaseForTest(domain.HasenpfefferPhasePlay)
		assert.Contains(t, p.Output(h, nil), i18n.T(key))
	}
	assert.Equal(t, "?", hasenpfefferSuitName(0))
}

// **宣言の状態は 3 通り。** 未宣言 / 降り / 数字を取り違えない。
func TestHasenpfefferCuiPresenterShowsEveryBidState(t *testing.T) {
	assert.Equal(t, i18n.T("hasenpfeffer.bidNone"), hasenpfefferBidStr(-1))
	assert.Equal(t, i18n.T("hasenpfeffer.bidPassed"), hasenpfefferBidStr(0))
	assert.Equal(t, i18n.Tf("hasenpfeffer.bidValue", "n", "4"), hasenpfefferBidStr(4))

	p := new(HasenpfefferCuiPresenter)
	h := newHasenpfefferForCui(t)
	h.GetPlayer(1).SetBid(0)
	h.GetPlayer(2).SetBid(4)
	out := p.Output(h, nil)
	assert.Contains(t, out, i18n.T("hasenpfeffer.bidPassed"))
	assert.Contains(t, out, i18n.T("hasenpfeffer.bidNone"))
}

// **降りられるかどうかが場面で変わる。** 3 通りすべてを踏む。
func TestHasenpfefferCuiPresenterBidPrompts(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)

	normal := newHasenpfefferForCui(t)
	normal.SetDealerIdxForTest(3)
	normal.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.Output(normal, nil), fixedPart("hasenpfeffer.promptBid"))

	// 親が降りられない場面。
	must := newHasenpfefferForCui(t)
	must.SetDealerIdxForTest(0)
	must.SetCurrentPlayerIdxForTest(1)
	require.NoError(t, must.BidForTest(1, 0))
	require.NoError(t, must.BidForTest(2, 0))
	require.NoError(t, must.BidForTest(3, 0))
	assert.Contains(t, p.Output(must, nil), fixedPart("hasenpfeffer.promptMustBid"))

	// 上限が立っていて降りるしかない場面。
	capped := newHasenpfefferForCui(t)
	capped.SetDealerIdxForTest(3)
	capped.SetCurrentPlayerIdxForTest(1)
	require.NoError(t, capped.BidForTest(1, domain.HasenpfefferMaxBid))
	capped.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.Output(capped, nil), fixedPart("hasenpfeffer.promptBidCapped"))

	// 相手の番。
	waiting := newHasenpfefferForCui(t)
	waiting.SetCurrentPlayerIdxForTest(2)
	assert.Contains(t, p.Output(waiting, nil), i18n.T("hasenpfeffer.promptBidWait"))
}

func TestHasenpfefferCuiPresenterDiscardPrompts(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)

	mine := newHasenpfefferForCui(t)
	mine.SetContractForTest(0, 4)
	mine.SetPhaseForTest(domain.HasenpfefferPhaseDiscard)
	assert.Contains(t, p.Output(mine, nil), fixedPart("hasenpfeffer.promptDiscard"))

	theirs := newHasenpfefferForCui(t)
	theirs.SetContractForTest(1, 4)
	theirs.SetPhaseForTest(domain.HasenpfefferPhaseDiscard)
	assert.Contains(t, p.Output(theirs, nil), i18n.T("hasenpfeffer.promptDiscardWait"))
}

// **落としたのか達成したのかは盤面から読めない。** 両側を踏む。
func TestHasenpfefferCuiPresenterHandEnd(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)
	for _, tc := range []struct {
		name   string
		tricks int
		key    string
	}{
		{"達成", 5, "hasenpfeffer.handEndMade"},
		{"失敗", 2, "hasenpfeffer.handEndEuchred"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHasenpfefferForCui(t)
			h.SetContractForTest(0, 4)
			h.SetPhaseForTest(domain.HasenpfefferPhasePlay)
			h.GiveTricksForTest(0, tc.tricks)
			h.GiveTricksForTest(1, domain.HasenpfefferTricksPerRound-tc.tricks)
			h.FinishHandForTest()

			out := p.Output(h, nil)
			assert.Contains(t, out, fixedPart(tc.key))
			assert.Contains(t, out, i18n.T("hasenpfeffer.promptNext"))
		})
	}
}

func TestHasenpfefferCuiPresenterGameEndBanners(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)
	for _, tc := range []struct {
		name   string
		t0, t1 int
		key    string
	}{
		{"team0", 12, 3, "hasenpfeffer.gameEndTeam0"},
		{"team1", 3, 12, "hasenpfeffer.gameEndTeam1"},
		{"tie", 12, 12, "hasenpfeffer.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHasenpfefferForCui(t)
			h.SetScoreForTestUse(0, tc.t0)
			h.SetScoreForTestUse(1, tc.t1)
			h.FinishGameForTest()
			assert.Contains(t, p.Output(h, nil), fixedPart(tc.key))
		})
	}
}

func TestHasenpfefferCuiPresenterShowsErrors(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)
	h := newHasenpfefferForCui(t)
	err := h.PlayerPlay(0)
	require.Error(t, err)
	assert.Contains(t, p.Output(h, err), err.Error())
}

// **助言は競り／捨て札／プレイで形が違う。** 3 通りすべてを踏む。
func TestHasenpfefferCuiPresenterHint(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)
	h := newHasenpfefferForCui(t)
	h.SetDealerIdxForTest(3)
	h.SetCurrentPlayerIdxForTest(0)

	bidHint := p.HintOutput(h)
	assert.Contains(t, bidHint, fixedPart("hasenpfeffer.hintBid"))

	require.NoError(t, h.BidForTest(0, 4))
	require.NoError(t, h.BidForTest(1, 0))
	require.NoError(t, h.BidForTest(2, 0))
	require.NoError(t, h.BidForTest(3, 0))
	discardHint := p.HintOutput(h)
	assert.Contains(t, discardHint, fixedPart("hasenpfeffer.hintDiscard"))

	require.NoError(t, h.DiscardForTest(0, 0, domain.CardDesignHeart))
	h.SetCurrentPlayerIdxForTest(0)
	cardHint := p.HintOutput(h)
	assert.Contains(t, cardHint, "HINT")
	for id := range hasenpfefferHintReasonKeys {
		assert.NotContains(t, cardHint, id, "識別子がそのまま漏れていない")
	}

	h.FinishGameForTest()
	assert.Contains(t, p.HintOutput(h), i18n.T("hasenpfeffer.hintNone"))
}

// **理由の識別子はすべて訳文を持つ。**
func TestHasenpfefferCuiPresenterHintReasonsAllTranslate(t *testing.T) {
	assert.NotEmpty(t, hasenpfefferHintReasonKeys)
	for id, key := range hasenpfefferHintReasonKeys {
		assert.NotEqual(t, key, i18n.T(key), "訳が無い: "+id)
	}
}

func TestHasenpfefferCuiPresenterActionLog(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(newHasenpfefferForCui(t)))
}

// **Web の打ち止めバナーと同じ条件・同じ意味であること** (#5758)。
// CUI は promptBidCapped、Web は minBid === 0 && !mustBid で出す。
func TestHasenpfefferCuiPresenterCappedPromptMatchesTheWebCondition(t *testing.T) {
	p := new(HasenpfefferCuiPresenter)
	h := newHasenpfefferForCui(t)
	h.SetPhaseForTest(domain.HasenpfefferPhaseBid)
	h.SetCurrentPlayerIdxForTest(0)
	// 上限が立っている = 次に打てる額が無い (NextBid が 0)。
	h.SetContractForTest(1, domain.HasenpfefferMaxBid)

	// **Web が見る値と同じ条件であることを、その値ごと確認する。**
	assert.Equal(t, 0, h.NextBid(), "minBid = 0 が Web の打ち止め条件")
	assert.False(t, h.MustBid(0))

	out := p.Output(h, nil)
	assert.Contains(t, out, fixedPart("hasenpfeffer.promptBidCapped"))
	assert.NotContains(t, out, fixedPart("hasenpfeffer.promptMustBid"))
}
