//go:build test

package presenter

import (
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
