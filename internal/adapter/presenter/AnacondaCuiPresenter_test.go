package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestAnacondaCuiPresenter_OutputPassPhase(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	p := new(presenter.AnacondaCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "アナコンダ")
	assert.Contains(t, out, "ラウンド")
	assert.Contains(t, out, "パスフェーズ")
	// The pass prompt names the recipient (the next player to the left).
	assert.Contains(t, out, strings.Split(i18n.T("anaconda.promptPassTo"), "{{")[0])
}

func TestAnacondaCuiPresenter_PassNoRecipientWhenAlone(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	// Eliminate every non-human seat → the human is the only participant, so
	// there is no one to pass to and the direction line is omitted.
	for i := 1; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetOut(true)
	}
	p := new(presenter.AnacondaCuiPresenter)
	out := p.Output(g, nil)
	assert.NotContains(t, out, strings.Split(i18n.T("anaconda.promptPassTo"), "{{")[0])
}

func TestAnacondaCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	p := new(presenter.AnacondaCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestAnacondaCuiPresenter_OutputSetAndRoll(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	p := new(presenter.AnacondaCuiPresenter)

	g.SetPhase(domain.AnacondaPhaseSet)
	assert.Contains(t, p.Output(g, nil), "セットフェーズ")

	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetRollIndex(1)
	assert.Contains(t, p.Output(g, nil), "ロールフェーズ")
}

func TestAnacondaCuiPresenter_OutputResult(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaWebSetKept(g.GetPlayer(i), anacondaWebWeak()...)
	}
	// human wins with four of a kind.
	anacondaWebSetKept(g.GetPlayer(0),
		anacondaWebCard(domain.CardDesignSpade, 8), anacondaWebCard(domain.CardDesignHeart, 8),
		anacondaWebCard(domain.CardDesignClover, 8), anacondaWebCard(domain.CardDesignDiamond, 8),
		anacondaWebCard(domain.CardDesignSpade, 2))
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetPot(60)
	g.ResolveShowdownForTest()

	p := new(presenter.AnacondaCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "フォーカード")
}

func TestAnacondaCuiPresenter_OutputGameEnd(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaWebSetKept(g.GetPlayer(i), anacondaWebWeak()...)
	}
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.ResolveShowdownForTest()
	p := new(presenter.AnacondaCuiPresenter)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestAnacondaCuiPresenter_HintOutputs(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	p := new(presenter.AnacondaCuiPresenter)
	// Pass phase → hint present.
	assert.NotEmpty(t, p.HintOutput(g))
	// Result phase → GetHint returns nil, still non-empty text.
	g.SetPhase(domain.AnacondaPhaseResult)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestAnacondaCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	_ = g.Pass([]int{0, 1, 2})
	p := new(presenter.AnacondaCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
