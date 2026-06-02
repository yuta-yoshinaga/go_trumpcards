package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func buildScoredScopa(t *testing.T) *domain.Scopa {
	t.Helper()
	s := domain.NewDefaultScopa()
	s.Reset()
	return s
}

func TestScopaWebPresenter_OutputJSON(t *testing.T) {
	p := &presenter.ScopaWebPresenter{}
	s := buildScoredScopa(t)
	out := p.Output(s, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	for _, k := range []string{"players", "tableCards", "phase", "config"} {
		if _, ok := parsed[k]; !ok {
			t.Errorf("missing key %q in output", k)
		}
	}
}

func TestScopaWebPresenter_OutputError(t *testing.T) {
	p := &presenter.ScopaWebPresenter{}
	s := buildScoredScopa(t)
	out := p.Output(s, scopaAssertErrWeb{})
	if !strings.Contains(out, "kaboom") {
		t.Errorf("expected error message in JSON, got: %s", out)
	}
}

type scopaAssertErrWeb struct{}

func (scopaAssertErrWeb) Error() string { return "kaboom" }

func TestScopaWebPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScopaWebPresenter{}
	s := buildScoredScopa(t)
	out := p.ActionLogOutput(s)
	var parsed any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Errorf("action log output not valid JSON: %v", err)
	}
}
