package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func escobaBuildGame(t *testing.T) *domain.Escoba {
	t.Helper()
	e := domain.NewDefaultEscoba()
	e.Reset()
	return e
}

func TestEscobaWebPresenter_HintOutput(t *testing.T) {
	p := &presenter.EscobaWebPresenter{}
	e := escobaBuildGame(t)
	// HintOutput mirrors Output (the GUI computes its own hint client-side).
	var parsed map[string]any
	if err := json.Unmarshal([]byte(p.HintOutput(e)), &parsed); err != nil {
		t.Fatalf("hint output is not valid JSON: %v", err)
	}
	if _, ok := parsed["phase"]; !ok {
		t.Errorf("missing phase key in hint output")
	}
}

func TestEscobaWebPresenter_OutputJSON(t *testing.T) {
	p := &presenter.EscobaWebPresenter{}
	e := escobaBuildGame(t)
	out := p.Output(e, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	for _, k := range []string{
		"players", "tableCards", "phase", "config", "stockRemaining",
		"roundNumber", "currentTurn", "dealerIdx", "winnerIdx", "isHumanTurn",
		"lastCaptureIdx", "handCaptures", "gameEndFlag",
	} {
		if _, ok := parsed[k]; !ok {
			t.Errorf("missing key %q in output", k)
		}
	}
}

func TestEscobaWebPresenter_OutputError(t *testing.T) {
	p := &presenter.EscobaWebPresenter{}
	e := escobaBuildGame(t)
	out := p.Output(e, escobaAssertErrWeb{})
	if !strings.Contains(out, "kaboom") {
		t.Errorf("expected error message in JSON, got: %s", out)
	}
}

type escobaAssertErrWeb struct{}

func (escobaAssertErrWeb) Error() string { return "kaboom" }

func TestEscobaWebPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.EscobaWebPresenter{}
	e := escobaBuildGame(t)
	out := p.ActionLogOutput(e)
	var parsed any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Errorf("action log output not valid JSON: %v", err)
	}
}

// escobaPlayedOutGame returns an all-CPU game driven to game end, exercising the
// roundEnd / gameEnd / score-detail / result-message render branches.
func escobaPlayedOutGame(t *testing.T) *domain.Escoba {
	t.Helper()
	players := make([]*domain.ScopaPlayer, domain.EscobaPlayerCnt)
	for i := range players {
		players[i] = domain.NewScopaPlayer(false)
	}
	e := domain.NewEscoba(domain.NewTrumpCardsScopa(), players, domain.DefaultEscobaConfig())
	e.Reset()
	for guard := 0; !e.GetGameEndFlag() && guard < 200000; guard++ {
		switch e.GetPhase() {
		case domain.EscobaPhasePlayerTurn:
			e.CpuPlay()
		case domain.EscobaPhaseRoundEnd:
			e.NextRound()
		}
	}
	return e
}

func TestEscobaWebPresenter_OutputGameEnd(t *testing.T) {
	p := &presenter.EscobaWebPresenter{}
	e := escobaPlayedOutGame(t)
	out := p.Output(e, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["gameEndFlag"] != true {
		t.Errorf("expected gameEndFlag true at game end, got %v", parsed["gameEndFlag"])
	}
	if _, ok := parsed["lastRoundDetail"]; !ok {
		t.Errorf("expected lastRoundDetail key at game end")
	}
}
