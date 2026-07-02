package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestTrenteEtQuaranteCuiPresenter_OutputBetPhase(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "チップ")
}

func TestTrenteEtQuaranteCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestTrenteEtQuaranteCuiPresenter_OutputResult(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Noir")
}

func TestTrenteEtQuaranteCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestTrenteEtQuaranteCuiPresenter_HintOutputNone(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	// After the round resolves (result phase) GetHint returns nil.
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestTrenteEtQuaranteCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
