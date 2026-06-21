package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func spBuildPlayedScopone(t *testing.T) *domain.Scopone {
	t.Helper()
	s := domain.NewDefaultScopone()
	s.Reset()
	return s
}

func TestScoponeCuiPresenter_Output(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	s := spBuildPlayedScopone(t)
	if out := p.Output(s, nil); out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestScoponeCuiPresenter_Error(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	s := spBuildPlayedScopone(t)
	out := p.Output(s, scoponeAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type scoponeAssertErr struct{}

func (scoponeAssertErr) Error() string { return "boom" }

func TestScoponeCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	s := spBuildPlayedScopone(t)
	if out := p.ActionLogOutput(s); out == "" {
		t.Error("expected non-empty action log output")
	}
}

func TestScoponeCuiPresenter_OutputGameEnd(t *testing.T) {
	p := &presenter.ScoponeCuiPresenter{}
	s := spPlayedOutScopone(t) // all-CPU game driven to game end (defined in the web presenter test)
	out := p.Output(s, nil)
	if out == "" {
		t.Fatal("expected non-empty output at game end")
	}
}
