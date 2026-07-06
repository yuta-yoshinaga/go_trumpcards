package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestKoiKoiCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.KoiKoiCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand + indexed field
}

func TestKoiKoiCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.KoiKoiCuiPresenter)

	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.KoiKoiPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.KoiKoiPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	// 空フィールド表示。
	g.SetPhase(domain.KoiKoiPhasePlay)
	g.SetFieldCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestKoiKoiCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.KoiKoiCuiPresenter)

	g := domain.NewDefaultKoiKoi()
	g.Reset()
	g.SetCurrentTurn(0)
	// play-phase hint
	assert.NotEmpty(t, p.HintOutput(g))

	// decision-phase hint
	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestKoiKoiCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultKoiKoi()
	g.Reset()
	p := new(presenter.KoiKoiCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(g))
}
