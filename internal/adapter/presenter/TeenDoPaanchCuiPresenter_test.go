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

func newTeenDoPaanchForCui(t *testing.T) *domain.TeenDoPaanch {
	t.Helper()
	g := domain.NewDefaultTeenDoPaanch()
	g.Reset()
	return g
}

func TestTeenDoPaanchCuiPresenterOutput(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	out := p.Output(newTeenDoPaanchForCui(t), nil)

	assert.Contains(t, out, i18n.T("teendopaanch.helpTitle"))
	assert.Contains(t, out, fixedPart("teendopaanch.header"))
	// **ノルマは宣言ではなく割り当て。** 規則そのものを毎回書く。
	assert.Contains(t, out, i18n.T("teendopaanch.rule"))
	// **「ノルマ」はルール行・切り札行・役割印にも出る。** 席行だけを数えるため
	// playerLine の並び（ノルマN 獲得N）で拾う。
	assert.Len(t, regexp.MustCompile(`ノルマ\d 獲得\d`).FindAllString(out, -1),
		domain.TeenDoPaanchPlayerCnt, "全員の席行にノルマと獲得数が出る")
}

// 切り札は未宣言と確定の両側を踏む。
func TestTeenDoPaanchCuiPresenterTrumpLine(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)

	undeclared := newTeenDoPaanchForCui(t)
	assert.Contains(t, p.Output(undeclared, nil), fixedPart("teendopaanch.trumpUndecided"))

	declared := newTeenDoPaanchForCui(t)
	declared.SetFivePlayerIdxForTest(0)
	require.NoError(t, declared.DeclareTrump(domain.CardDesignHeart))
	out := p.Output(declared, nil)
	assert.Contains(t, out, fixedPart("teendopaanch.trumpLine"))
	assert.Contains(t, out, i18n.T("teendopaanch.suitHeart"))
	assert.NotContains(t, out, fixedPart("teendopaanch.trumpUndecided"))
}

// **4 スートすべてに名前がある。** 既定の "?" に落ちない。
func TestTeenDoPaanchCuiPresenterNamesEverySuit(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	for suit, key := range map[int]string{
		domain.CardDesignSpade:   "teendopaanch.suitSpade",
		domain.CardDesignClover:  "teendopaanch.suitClover",
		domain.CardDesignHeart:   "teendopaanch.suitHeart",
		domain.CardDesignDiamond: "teendopaanch.suitDiamond",
	} {
		g := newTeenDoPaanchForCui(t)
		g.SetTrumpSuitForTest(suit)
		assert.Contains(t, p.Output(g, nil), i18n.T(key))
	}
	assert.Equal(t, "?", teenDoPaanchSuitName(0))
}

// **ノルマ 5 の席には印が付く。** 切り札を決めるのがその席だから。
func TestTeenDoPaanchCuiPresenterMarksTheFiveSeat(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	g := newTeenDoPaanchForCui(t)
	g.SetFivePlayerIdxForTest(1)

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("teendopaanch.roleFive"))
	assert.Equal(t, 1, len(regexp.MustCompile(regexp.QuoteMeta(i18n.T("teendopaanch.roleFive"))).FindAllString(out, -1)),
		"印が付くのは 1 席だけ")
}

// **宣言するのが自分か相手かで案内が違う。** 両側を踏む。
func TestTeenDoPaanchCuiPresenterTrumpPrompts(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)

	mine := newTeenDoPaanchForCui(t)
	mine.SetFivePlayerIdxForTest(0)
	assert.Contains(t, p.Output(mine, nil), fixedPart("teendopaanch.promptTrump"))

	theirs := newTeenDoPaanchForCui(t)
	theirs.SetFivePlayerIdxForTest(1)
	out := p.Output(theirs, nil)
	assert.Contains(t, out, i18n.T("teendopaanch.promptTrumpWait"))
	assert.NotContains(t, out, fixedPart("teendopaanch.promptTrump"))
}

// **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
func TestTeenDoPaanchCuiPresenterReportsTheExchange(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	fresh := newTeenDoPaanchForCui(t)
	assert.NotContains(t, p.Output(fresh, nil), fixedPart("teendopaanch.exchangeLine"))

	g := newTeenDoPaanchForCui(t)
	g.SetFivePlayerIdxForTest(0)
	require.NoError(t, g.DeclareTrump(domain.CardDesignSpade))
	g.GiveTricksForTest(0, 7)
	g.GiveTricksForTest(1, 1)
	g.GiveTricksForTest(2, 2)
	g.FinishRoundForTest()
	g.NextRound()
	require.NoError(t, g.DeclareTrump(domain.CardDesignSpade))

	assert.Contains(t, p.Output(g, nil), fixedPart("teendopaanch.exchangeLine"))
}

func TestTeenDoPaanchCuiPresenterRoundEndPrompt(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	g := newTeenDoPaanchForCui(t)
	require.NoError(t, g.DeclareTrump(domain.CardDesignSpade))
	g.FinishRoundForTest()
	assert.Contains(t, p.Output(g, nil), i18n.T("teendopaanch.promptNext"))
}

func TestTeenDoPaanchCuiPresenterGameEndBanners(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	for _, tc := range []struct {
		name string
		met  [domain.TeenDoPaanchPlayerCnt]int
		key  string
	}{
		{"you", [3]int{3, 1, 0}, "teendopaanch.gameEndYou"},
		{"cpu", [3]int{0, 3, 1}, "teendopaanch.gameEndCpu"},
		{"tie", [3]int{2, 2, 2}, "teendopaanch.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTeenDoPaanchForCui(t)
			for i, n := range tc.met {
				g.GetPlayer(i).SetMet(n)
			}
			g.FinishGameForTest()
			assert.Contains(t, p.Output(g, nil), fixedPart(tc.key))
		})
	}
}

func TestTeenDoPaanchCuiPresenterShowsErrors(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	g := newTeenDoPaanchForCui(t)
	err := g.DeclareTrump(99)
	require.Error(t, err)
	assert.Contains(t, p.Output(g, err), err.Error())
}

// **宣言フェーズの助言はスートを指し、プレイ中は札を指す。**
func TestTeenDoPaanchCuiPresenterHint(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	g := newTeenDoPaanchForCui(t)
	g.SetFivePlayerIdxForTest(0)

	trumpHint := p.HintOutput(g)
	assert.Contains(t, trumpHint, "HINT")
	assert.Contains(t, trumpHint, fixedPart("teendopaanch.hintTrump"))

	require.NoError(t, g.DeclareTrump(domain.CardDesignHeart))
	g.SetCurrentPlayerIdxForTest(0)
	cardHint := p.HintOutput(g)
	// **勧める札は配りで変わる。** 固定の添字ではなく「合法な札を指している」を見る。
	idx := regexp.MustCompile(`\[HINT: \[(\d+)\]`).FindStringSubmatch(cardHint)
	require.Len(t, idx, 2, "札を指した助言になっている")
	n, err := strconv.Atoi(idx[1])
	require.NoError(t, err)
	assert.Contains(t, g.GetValidPlayIndices(0), n, "勧める札は必ず合法")
	for id := range teenDoPaanchHintReasonKeys {
		assert.NotContains(t, cardHint, id, "識別子がそのまま漏れていない")
	}

	g.FinishGameForTest()
	assert.Contains(t, p.HintOutput(g), i18n.T("teendopaanch.hintNone"))
}

// **理由の識別子はすべて訳文を持つ。**
func TestTeenDoPaanchCuiPresenterHintReasonsAllTranslate(t *testing.T) {
	assert.NotEmpty(t, teenDoPaanchHintReasonKeys)
	for id, key := range teenDoPaanchHintReasonKeys {
		assert.NotEqual(t, key, i18n.T(key), "訳が無い: "+id)
	}
}

func TestTeenDoPaanchCuiPresenterActionLog(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(newTeenDoPaanchForCui(t)))
}

// **合計だけでは、誰の最強札が誰に渡ったのか分からない** (#5757)。
func TestTeenDoPaanchCuiPresenterShowsTheExchangeBreakdown(t *testing.T) {
	p := new(TeenDoPaanchCuiPresenter)
	g := newTeenDoPaanchForCui(t)
	g.SetPhaseForTest(domain.TeenDoPaanchPhasePlay)
	teenDoPaanchCuiHandOf(g, 0, domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignClover, 9, false))
	teenDoPaanchCuiHandOf(g, 1, domain.NewCard(domain.CardDesignHeart, 1, false))
	teenDoPaanchCuiHandOf(g, 2, domain.NewCard(domain.CardDesignSpade, 13, false))
	g.SetSurplusForTest([]int{2, -1, -1})
	g.ExchangeForTest()

	out := teenDoPaanchPlain(p.Output(g, nil))

	// 誰から誰へ何枚か、ペアごとに出る (受け入れ条件1・2)。
	for _, giver := range []int{1, 2} {
		assert.Contains(t, out, i18n.Tf("teendopaanch.exchangePair",
			"giver", teenDoPaanchPlain(cuiPlayerName(g.GetPlayer(giver), giver)),
			"taker", teenDoPaanchPlain(cuiPlayerName(g.GetPlayer(0), 0)),
			"n", "1"))
	}
	assert.NotContains(t, out, "{{")

	// 交換が起きなければ何も出ない (受け入れ条件3)。
	quiet := newTeenDoPaanchForCui(t)
	quiet.SetPhaseForTest(domain.TeenDoPaanchPhasePlay)
	teenDoPaanchCuiHandOf(quiet, 0, domain.NewCard(domain.CardDesignClover, 8, false))
	quiet.SetSurplusForTest([]int{0, 0, 0})
	quiet.ExchangeForTest()
	assert.NotContains(t, teenDoPaanchPlain(p.Output(quiet, nil)), fixedPart("teendopaanch.exchangeLine"))
}

// teenDoPaanchCuiHandOf は指定席の手札を差し替える。
func teenDoPaanchCuiHandOf(g *domain.TeenDoPaanch, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.ResetRound()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// teenDoPaanchPlain は色付けのエスケープを落とす。
var teenDoPaanchAnsi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func teenDoPaanchPlain(s string) string { return teenDoPaanchAnsi.ReplaceAllString(s, "") }
