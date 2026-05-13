//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPiquetWebPresenter_Output(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetWebPresenter{}
	out := p.Output(g, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\nbody=%s", err, out)
	}
	if parsed["dealNumber"].(float64) != 1 {
		t.Errorf("dealNumber = %v, want 1", parsed["dealNumber"])
	}
	if _, ok := parsed["players"]; !ok {
		t.Errorf("missing players field")
	}
	if _, ok := parsed["config"]; !ok {
		t.Errorf("missing config field")
	}
}

func TestPiquetWebPresenter_Output_WithError(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetWebPresenter{}
	out := p.Output(g, errors.New("boom"))
	var parsed map[string]any
	_ = json.Unmarshal([]byte(out), &parsed)
	if parsed["message"] != "boom" {
		t.Errorf("expected message=boom, got %v", parsed["message"])
	}
}

func TestPiquetWebPresenter_HintOutput(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetWebPresenter{}
	out := p.HintOutput(g)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestPiquetWebPresenter_ActionLogOutput(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetWebPresenter{}
	out := p.ActionLogOutput(g)
	if out == "" {
		t.Error("expected non-empty action log output")
	}
}

func TestPiquetWebPresenter_PlayersHandHiddenForCPU(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetWebPresenter{}
	out := p.Output(g, nil)
	var parsed map[string]any
	_ = json.Unmarshal([]byte(out), &parsed)
	players := parsed["players"].([]any)
	cpuPlayer := players[1].(map[string]any) // idx=1 is CPU
	cards := cpuPlayer["cards"].([]any)
	if len(cards) != 0 {
		t.Errorf("CPU cards should be hidden, got %d entries", len(cards))
	}
	// But cardCount must still reflect actual hand
	if cpuPlayer["cardCount"].(float64) != float64(domain.PiquetHandSize) {
		t.Errorf("CPU cardCount = %v, want %d", cpuPlayer["cardCount"], domain.PiquetHandSize)
	}
}
