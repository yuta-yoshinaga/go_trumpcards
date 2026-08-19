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

func newReversisForCui(t *testing.T) *domain.Reversis {
	t.Helper()
	r := domain.NewDefaultReversis()
	r.Reset()
	return r
}

func TestReversisCuiPresenterOutput(t *testing.T) {
	p := new(ReversisCuiPresenter)
	r := newReversisForCui(t)
	r.SetCurrentPlayerIdxForTest(0)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("reversis.helpTitle"))
	// **失点の配分と印付きの 2 枚は盤面から読めない。** 常時出す。
	assert.Contains(t, out, i18n.T("reversis.penaltyLine"))
	// プールも常時出す。
	assert.Contains(t, out, fixedPart("reversis.poolLine"))
	assert.Contains(t, out, i18n.T("reversis.promptPlay"))
}

// 印を取っているかどうかで表示が変わる。両側を踏む。
func TestReversisCuiPresenterMarks(t *testing.T) {
	p := new(ReversisCuiPresenter)

	clean := newReversisForCui(t)
	assert.Contains(t, p.Output(clean, nil), i18n.T("reversis.markNone"))

	hit := newReversisForCui(t)
	hit.GetPlayer(0).SetTookQuinola(true)
	hit.GetPlayer(1).SetTookDiamondAce(true)
	out := p.Output(hit, nil)
	assert.Contains(t, out, i18n.T("reversis.markQuinola"))
	assert.Contains(t, out, i18n.T("reversis.markDiamondAce"))
}

func TestReversisCuiPresenterRoundEnd(t *testing.T) {
	p := new(ReversisCuiPresenter)
	r := newReversisForCui(t)
	r.SetPhaseForTest(domain.ReversisPhaseRoundEnd)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("reversis.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("reversis.promptNext"))
	assert.NotContains(t, out, i18n.T("reversis.promptPlay"))
}

func TestReversisCuiPresenterError(t *testing.T) {
	p := new(ReversisCuiPresenter)
	assert.Contains(t, p.Output(newReversisForCui(t), assert.AnError), assert.AnError.Error())
}

func TestReversisCuiPresenterGameEnd(t *testing.T) {
	p := new(ReversisCuiPresenter)

	t.Run("a winner", func(t *testing.T) {
		r := newReversisForCui(t)
		r.GetPlayer(1).SetChips(200)
		r.FinishGameForTest()
		out := p.Output(r, nil)
		assert.Contains(t, out, fixedPart("reversis.gameEndWinner"))
		assert.NotContains(t, out, i18n.T("reversis.promptPlay"))
	})

	t.Run("a tie", func(t *testing.T) {
		r := newReversisForCui(t)
		r.FinishGameForTest()
		assert.Contains(t, p.Output(r, nil), i18n.T("reversis.gameEndTie"))
	})
}

func TestReversisCuiPresenterHintOutput(t *testing.T) {
	p := new(ReversisCuiPresenter)
	r := newReversisForCui(t)
	r.SetCurrentPlayerIdxForTest(0)

	out := p.HintOutput(r)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "reversisLeadSafe", "生のキーが出ていたら未登録")
}

// 4 つの理由キーがすべて日本語文言に解決される。
func TestReversisCuiPresenterHintReasons(t *testing.T) {
	p := new(ReversisCuiPresenter)

	lead := newReversisForCui(t)
	lead.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(lead), i18n.T("reversis.hintReasonLeadSafe"))

	marked := newReversisForCui(t)
	marked.SetCurrentPlayerIdxForTest(0)
	marked.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 11, false)},
	})
	assert.Contains(t, p.HintOutput(marked), i18n.T("reversis.hintReasonAvoidMarked"))

	points := newReversisForCui(t)
	points.SetCurrentPlayerIdxForTest(0)
	points.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
	})
	assert.Contains(t, p.HintOutput(points), i18n.T("reversis.hintReasonAvoidPoints"))

	safe := newReversisForCui(t)
	safe.SetCurrentPlayerIdxForTest(0)
	safe.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
	})
	assert.Contains(t, p.HintOutput(safe), i18n.T("reversis.hintReasonDumpHigh"))
}

func TestReversisCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(ReversisCuiPresenter)
	r := newReversisForCui(t)
	r.GiveUp()
	assert.Contains(t, p.HintOutput(r), i18n.T("reversis.hintNone"))
}

func TestReversisCuiPresenterActionLogOutput(t *testing.T) {
	p := new(ReversisCuiPresenter)
	r := newReversisForCui(t)
	r.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(r))
}

// **どの札が何点かを手札に併記する** (#5747)。A=4 / K=3 / Q=2 / J=1 を
// 暗算し続けるゲームではない。
func TestReversisCuiPresenterShowsCardPoints(t *testing.T) {
	p := new(ReversisCuiPresenter)
	r := newReversisForCui(t)
	r.SetCurrentPlayerIdxForTest(0)

	human := r.GetPlayer(0)
	human.ResetRound()
	for _, c := range []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, true),  // A = 4
		domain.NewCard(domain.CardDesignHeart, 13, true), // K = 3
		domain.NewCard(domain.CardDesignSpade, 7, true),  // 平札 = 0
	} {
		human.AddCard(c)
	}

	out := reversisPlain(p.Output(r, nil))

	// 札ごとに突き合わせる。どこかに 4 がある、では隣の札の点でも通る。
	assert.Contains(t, out, i18n.Tf("reversis.handCard", "idx", "0", "card", "SPADE 1", "points", "4"))
	assert.Contains(t, out, i18n.Tf("reversis.handCard", "idx", "1", "card", "HEART 13", "points", "3"))
	// **0 点の札も明示する** (受け入れ条件2)。
	assert.Contains(t, out, i18n.Tf("reversis.handCard", "idx", "2", "card", "SPADE 7", "points", "0"))
	assert.NotContains(t, out, "{{")
}

// reversisPlain は色付けのエスケープを落とす。
var reversisAnsi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func reversisPlain(s string) string { return reversisAnsi.ReplaceAllString(s, "") }
