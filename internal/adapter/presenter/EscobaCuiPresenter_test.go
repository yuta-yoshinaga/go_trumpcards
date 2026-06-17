package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func escobaBuildCuiGame(t *testing.T) *domain.Escoba {
	t.Helper()
	e := domain.NewDefaultEscoba()
	e.Reset()
	return e
}

func TestEscobaCuiPresenter_Output(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	if out := p.Output(e, nil); out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestEscobaCuiPresenter_Error(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	out := p.Output(e, escobaAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type escobaAssertErr struct{}

func (escobaAssertErr) Error() string { return "boom" }

func TestEscobaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	if out := p.ActionLogOutput(e); out == "" {
		t.Error("expected non-empty action log output")
	}
}

func TestEscobaCuiPresenter_OutputGameEnd(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaPlayedOutGame(t) // all-CPU game driven to game end (defined in the web presenter test)
	out := p.Output(e, nil)
	if out == "" {
		t.Fatal("expected non-empty output at game end")
	}
}
