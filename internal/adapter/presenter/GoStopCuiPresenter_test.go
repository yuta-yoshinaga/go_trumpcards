package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestGoStopCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.GoStopCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]")
}

func TestGoStopCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.GoStopCuiPresenter)

	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.GoStopPhaseGoDecision)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.GoStopPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.GoStopPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	g.SetPhase(domain.GoStopPhasePlay)
	g.SetFieldCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestGoStopCuiPresenter_CategoryCounts(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	g.GetPlayer(0).AddCaptured(gostopGwang3Cards()) // 3 光 (Gwang)
	p := new(presenter.GoStopCuiPresenter)
	out := p.Output(g, nil)
	assert.Contains(t, out, strings.Split(i18n.T("gostop.categoryCounts"), "{{")[0])
	// The three captured Gwang are counted in the category line.
	assert.Contains(t, out, "光3")
}

func TestGoStopCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.GoStopCuiPresenter)

	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.HintOutput(g))

	g.SetPhase(domain.GoStopPhaseGoDecision)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestGoStopCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	p := new(presenter.GoStopCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(g))
}

func gostopGwang3Cards() []*domain.Card {
	return []*domain.Card{domain.NewCard(1, 1, false), domain.NewCard(3, 1, false), domain.NewCard(12, 1, false)}
}

func TestGoStopCuiPresenter_RoundResult_HumanWinner(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(gostopGwang3Cards())
	require.NoError(t, g.PlayerDecide(false))
	require.Equal(t, domain.GoStopPhaseRoundEnd, g.GetPhase())

	p := new(presenter.GoStopCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestGoStopCuiPresenter_RoundResult_CpuWinner(t *testing.T) {
	g := domain.NewDefaultGoStop()
	g.Reset()
	cfg := domain.DefaultGoStopConfig()
	cfg.CpuDifficulty = domain.GoStopCpuDifficultyEasy // Easy always stops
	g.SetConfig(cfg)
	g.SetCurrentTurn(1)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(1).AddCaptured(gostopGwang3Cards())
	g.CpuDecide()
	require.Equal(t, domain.GoStopPhaseRoundEnd, g.GetPhase())

	// **CPU 名も i18n と太字を通す。**人間側だけ i18n を通し、CPU は英語
	// リテラルのままだった (#4855)。色を有効にして太字の有無まで見る。
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.GoStopCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// **ラウンド勝利行そのものを見る。**プレイヤー一覧は元から太字の CPU 名を
	// 出しているので、出力全体で探すと修正前でも当たってしまう。
	bold1 := color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))
	var winLine string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "ラウンド勝利") {
			winLine = line
		}
	}
	require.NotEmpty(t, winLine, "round-win line missing from output: %q", out)
	assert.Contains(t, winLine, bold1)
}

// #5710: 「あと何枚でどの役が揃うか」は Go/Stop 判断の材料そのもので、Web は
// gostop-yaku-preview として出しているのに、CUI は確定点しか出していなかった。
func TestGoStopCuiPresenter_NearYakuPreview(t *testing.T) {
	p := new(presenter.GoStopCuiPresenter)

	atDecision := func(bd *domain.GoStopBreakdown) *domain.GoStop {
		g := domain.NewDefaultGoStop()
		g.Reset()
		g.SetCurrentTurn(0)
		g.SetPhase(domain.GoStopPhaseGoDecision)
		g.SetPendingBreakdown(bd)
		return g
	}

	t.Run("lists every yaku within reach", func(t *testing.T) {
		out := p.Output(atDecision(&domain.GoStopBreakdown{
			BrightCount: 2, RibbonCount: 4, AnimalCount: 3, PiCount: 9,
		}), nil)

		assert.Contains(t, out, i18n.Tf("gostop.previewItem",
			"name", i18n.T("gostop.previewSamgwang"), "remaining", "1"))
		assert.Contains(t, out, i18n.Tf("gostop.previewItem",
			"name", i18n.T("gostop.previewYeol"), "remaining", "2"))
		assert.NotContains(t, out, "{{")
	})

	t.Run("says nothing when no yaku is close", func(t *testing.T) {
		out := p.Output(atDecision(&domain.GoStopBreakdown{
			BrightCount: 0, RibbonCount: 0, AnimalCount: 0, PiCount: 0,
		}), nil)

		assert.NotContains(t, out, strings.Split(i18n.T("gostop.previewTitle"), "{{")[0])
	})

	// 内訳がまだ無い局面 (決断前) でも落ちない。
	t.Run("survives a missing breakdown", func(t *testing.T) {
		out := p.Output(atDecision(nil), nil)

		assert.NotContains(t, out, strings.Split(i18n.T("gostop.previewTitle"), "{{")[0])
	})
}
