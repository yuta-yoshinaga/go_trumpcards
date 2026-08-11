//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newPolignacForCui(t *testing.T) *domain.Polignac {
	t.Helper()
	p := domain.NewDefaultPolignac()
	p.Reset()
	require.NoError(t, p.PassDeclaration())
	return p
}

func TestPolignacCuiPresenterOutput(t *testing.T) {
	p := new(PolignacCuiPresenter)
	g := newPolignacForCui(t)
	g.SetCurrentPlayerIdxForTest(0)

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("polignac.helpTitle"))
	// **失点するのはジャックだけで ♠J は2倍、という規則を常時出す。**
	assert.Contains(t, out, i18n.T("polignac.penaltyLine"))
	assert.Contains(t, out, i18n.T("polignac.promptPlay"))
}

// 宣言フェーズでは capot/pass を促し、打ち手は促さない。
func TestPolignacCuiPresenterDeclarePhase(t *testing.T) {
	p := new(PolignacCuiPresenter)
	g := domain.NewDefaultPolignac()
	g.Reset()

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("polignac.promptDeclare"))
	assert.Contains(t, out, i18n.T("polignac.promptDeclareHelp"))
	assert.NotContains(t, out, i18n.T("polignac.promptPlay"))
}

// capot 宣言中だけ警告帯が出る。負のコントロール付き。
func TestPolignacCuiPresenterCapotBanner(t *testing.T) {
	p := new(PolignacCuiPresenter)

	quiet := newPolignacForCui(t)
	assert.NotContains(t, p.Output(quiet, nil), i18n.T("polignac.capotMark"))

	loud := domain.NewDefaultPolignac()
	loud.Reset()
	require.NoError(t, loud.DeclareCapot())
	out := p.Output(loud, nil)
	assert.Contains(t, out, i18n.T("polignac.capotMark"), "宣言者に印が付く")
}

func TestPolignacCuiPresenterRoundEnd(t *testing.T) {
	p := new(PolignacCuiPresenter)
	g := newPolignacForCui(t)
	g.SetPhaseForTest(domain.PolignacPhaseRoundEnd)

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("polignac.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("polignac.promptNext"))
	assert.NotContains(t, out, i18n.T("polignac.promptPlay"))
}

func TestPolignacCuiPresenterError(t *testing.T) {
	p := new(PolignacCuiPresenter)
	assert.Contains(t, p.Output(newPolignacForCui(t), assert.AnError), assert.AnError.Error())
}

func TestPolignacCuiPresenterGameEnd(t *testing.T) {
	p := new(PolignacCuiPresenter)

	t.Run("a winner", func(t *testing.T) {
		g := newPolignacForCui(t)
		for i := range domain.PolignacPlayerCnt {
			g.GetPlayer(i).SetScore(5)
		}
		g.GetPlayer(1).SetScore(0)
		g.FinishGameForTest()
		out := p.Output(g, nil)
		assert.Contains(t, out, fixedPart("polignac.gameEndWinner"))
		assert.NotContains(t, out, i18n.T("polignac.promptPlay"))
	})

	t.Run("a tie", func(t *testing.T) {
		g := newPolignacForCui(t)
		g.FinishGameForTest()
		assert.Contains(t, p.Output(g, nil), i18n.T("polignac.gameEndTie"))
	})
}

func TestPolignacCuiPresenterHintOutput(t *testing.T) {
	p := new(PolignacCuiPresenter)
	g := newPolignacForCui(t)
	g.SetCurrentPlayerIdxForTest(0)

	out := p.HintOutput(g)
	assert.Contains(t, out, "HINT")
	// 生の理由キーが出ていたら i18n 未登録。
	assert.NotContains(t, out, "polignacLeadSafe")
	assert.NotContains(t, out, "polignacAvoidJack")
}

// 4 つの理由キーがすべて日本語文言に解決される。
func TestPolignacCuiPresenterHintReasons(t *testing.T) {
	p := new(PolignacCuiPresenter)
	spadeTen := []*domain.TrickCard{{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 10, false)}}

	lead := newPolignacForCui(t)
	lead.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(lead), i18n.T("polignac.hintReasonLeadSafe"))

	jack := newPolignacForCui(t)
	jack.SetCurrentPlayerIdxForTest(0)
	jack.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 11, false)},
	})
	assert.Contains(t, p.HintOutput(jack), i18n.T("polignac.hintReasonAvoidJack"))

	safe := newPolignacForCui(t)
	safe.SetCurrentPlayerIdxForTest(0)
	safe.SetCurrentTrickForTest(spadeTen)
	assert.Contains(t, p.HintOutput(safe), i18n.T("polignac.hintReasonDumpJack"))

	block := newPolignacForCui(t)
	block.SetCurrentPlayerIdxForTest(0)
	block.SetCapotIdxForTest(2)
	block.SetCurrentTrickForTest(spadeTen)
	assert.Contains(t, p.HintOutput(block), i18n.T("polignac.hintReasonBlockCapot"))

	// 自分の capot は狙いが反転する。別の文言になっていること。
	own := newPolignacForCui(t)
	own.SetCurrentPlayerIdxForTest(0)
	own.SetCapotIdxForTest(0)
	own.SetCurrentTrickForTest(spadeTen)
	out := p.HintOutput(own)
	assert.Contains(t, out, i18n.T("polignac.hintReasonWinCapot"))
	assert.NotContains(t, out, i18n.T("polignac.hintReasonBlockCapot"))
}

func TestPolignacCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(PolignacCuiPresenter)
	g := newPolignacForCui(t)
	g.GiveUp()
	assert.Contains(t, p.HintOutput(g), i18n.T("polignac.hintNone"))
}

func TestPolignacCuiPresenterActionLogOutput(t *testing.T) {
	p := new(PolignacCuiPresenter)
	g := newPolignacForCui(t)
	g.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(g))
}
