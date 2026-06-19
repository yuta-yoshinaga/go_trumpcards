package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func buildScoredCuarenta(t *testing.T) *domain.Cuarenta {
	t.Helper()
	g := domain.NewDefaultCuarenta()
	g.Reset()
	return g
}

func TestCuarentaWebPresenter_OutputJSON(t *testing.T) {
	p := &presenter.CuarentaWebPresenter{}
	g := buildScoredCuarenta(t)
	out := p.Output(g, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	for _, k := range []string{"players", "tableCards", "phase", "config", "teamScores"} {
		if _, ok := parsed[k]; !ok {
			t.Errorf("missing key %q in output", k)
		}
	}
}

func TestCuarentaWebPresenter_OutputError(t *testing.T) {
	p := &presenter.CuarentaWebPresenter{}
	g := buildScoredCuarenta(t)
	out := p.Output(g, cuarentaAssertErrWeb{})
	if !strings.Contains(out, "kaboom") {
		t.Errorf("expected error message in JSON, got: %s", out)
	}
}

type cuarentaAssertErrWeb struct{}

func (cuarentaAssertErrWeb) Error() string { return "kaboom" }

func TestCuarentaWebPresenter_GameEndMessage(t *testing.T) {
	p := &presenter.CuarentaWebPresenter{}
	cfg := domain.DefaultCuarentaConfig()
	cfg.TargetScore = 6
	players := make([]*domain.CuarentaPlayer, domain.CuarentaPlayerCnt)
	for i := 0; i < domain.CuarentaPlayerCnt; i++ {
		players[i] = domain.NewCuarentaPlayer(false)
	}
	g := domain.NewCuarenta(domain.NewTrumpCardsScopa(), players, cfg)
	g.Reset()
	for i := 0; i < 100000 && !g.GetGameEndFlag(); i++ {
		g.CpuPlay()
		if !g.GetGameEndFlag() && g.GetPhase() != int(domain.CuarentaPhasePlay) {
			g.NextRound()
		}
	}
	out := p.Output(g, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if parsed["messageCode"] != "cuarenta.result.scores" {
		t.Errorf("expected result message code, got: %v", parsed["messageCode"])
	}
}

func TestCuarentaWebPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.CuarentaWebPresenter{}
	g := buildScoredCuarenta(t)
	out := p.ActionLogOutput(g)
	var parsed any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Errorf("action log output not valid JSON: %v", err)
	}
}
