package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func buildPlayedScopa(t *testing.T) *domain.Scopa {
	t.Helper()
	s := domain.NewDefaultScopa()
	s.Reset()
	return s
}

func TestScopaCuiPresenter_Output(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	if out := p.Output(s, nil); out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestScopaCuiPresenter_Error(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	out := p.Output(s, scopaAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type scopaAssertErr struct{}

func (scopaAssertErr) Error() string { return "boom" }

func TestScopaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	if out := p.ActionLogOutput(s); out == "" {
		t.Error("expected non-empty action log output")
	}
}
