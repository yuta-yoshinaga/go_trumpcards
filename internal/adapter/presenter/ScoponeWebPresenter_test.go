package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func spBuildScoredScopone(t *testing.T) *domain.Scopone {
	t.Helper()
	s := domain.NewDefaultScopone()
	s.Reset()
	return s
}

func TestScoponeWebPresenter_OutputJSON(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spBuildScoredScopone(t)
	out := p.Output(s, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	for _, k := range []string{
		"players", "tableCards", "phase", "config", "teamScores",
		"roundNumber", "currentTurn", "dealerIdx", "winnerTeam", "isHumanTurn",
	} {
		if _, ok := parsed[k]; !ok {
			t.Errorf("missing key %q in output", k)
		}
	}
}

func TestScoponeWebPresenter_OutputError(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spBuildScoredScopone(t)
	out := p.Output(s, scoponeAssertErrWeb{})
	if !strings.Contains(out, "kaboom") {
		t.Errorf("expected error message in JSON, got: %s", out)
	}
}

type scoponeAssertErrWeb struct{}

func (scoponeAssertErrWeb) Error() string { return "kaboom" }

func TestScoponeWebPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spBuildScoredScopone(t)
	out := p.ActionLogOutput(s)
	var parsed any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Errorf("action log output not valid JSON: %v", err)
	}
}
