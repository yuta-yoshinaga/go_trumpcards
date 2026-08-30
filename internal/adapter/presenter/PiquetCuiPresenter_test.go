//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newPiquetForPresenter builds a fresh game in a known state for presenter tests.
func newPiquetForPresenter(t *testing.T) *domain.Piquet {
	t.Helper()
	players := []*domain.PiquetPlayer{
		domain.NewPiquetPlayer(true),
		domain.NewPiquetPlayer(false),
	}
	g := domain.NewPiquet(domain.NewTrumpCardsBelote(), players, domain.DefaultPiquetConfig())
	g.Reset()
	return g
}

// driveToPhase advances both CPUs forward until phase is reached or game ends.
// Players are swapped to both-CPU first so we don't need human input.
func driveToPhase(t *testing.T, phase domain.PiquetPhase) *domain.Piquet {
	t.Helper()
	players := []*domain.PiquetPlayer{
		domain.NewPiquetPlayer(false),
		domain.NewPiquetPlayer(false),
	}
	g := domain.NewPiquet(domain.NewTrumpCardsBelote(), players,
		domain.PiquetConfig{DealsPerPartie: 1, CpuDifficulty: domain.PiquetCpuDifficultyNormal})
	g.Reset()
	// Exchange autoruns to Declaration via CpuPlay
	for g.GetPhase() == domain.PiquetPhaseExchange {
		g.CpuPlay()
	}
	if phase == domain.PiquetPhaseDeclaration {
		return g
	}
	// Resolve 3 declarations
	for g.GetPhase() == domain.PiquetPhaseDeclaration {
		_, _ = g.ResolveDeclaration()
	}
	if phase == domain.PiquetPhasePlay {
		return g
	}
	// Drive plays to end
	for g.GetPhase() == domain.PiquetPhasePlay {
		g.CpuPlay()
	}
	return g
}

func TestPiquetCuiPresenter_Output_ExchangePhase(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetCuiPresenter{}
	out := p.Output(g, nil)
	if !strings.Contains(out, "Piquet") && !strings.Contains(out, "ピケ") {
		t.Errorf("expected title in output, got: %s", out)
	}
	if !strings.Contains(out, "1") {
		t.Errorf("expected deal number in output, got: %s", out)
	}
}

func TestPiquetCuiPresenter_Output_WithError(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetCuiPresenter{}
	out := p.Output(g, errors.New("test error"))
	if !strings.Contains(out, "test error") {
		t.Errorf("expected error in output, got: %s", out)
	}
}

func TestPiquetCuiPresenter_HintOutput(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetCuiPresenter{}
	out := p.HintOutput(g)
	if out == "" {
		t.Error("expected non-empty hint output")
	}
}

func TestPiquetCuiPresenter_HintOutput_PlayPhase(t *testing.T) {
	g := driveToPhase(t, domain.PiquetPhasePlay)
	p := &PiquetCuiPresenter{}
	out := p.HintOutput(g)
	if out == "" {
		t.Error("expected non-empty hint output in play phase")
	}
}

func TestPiquetCuiPresenter_ActionLogOutput(t *testing.T) {
	g := newPiquetForPresenter(t)
	p := &PiquetCuiPresenter{}
	out := p.ActionLogOutput(g)
	if out == "" {
		t.Error("expected non-empty log output")
	}
}

func TestPiquetCuiPresenter_Output_DeclarationPhase(t *testing.T) {
	g := driveToPhase(t, domain.PiquetPhaseDeclaration)
	p := &PiquetCuiPresenter{}
	out := p.Output(g, nil)
	if out == "" {
		t.Error("expected non-empty output in declaration phase")
	}
}

func TestPiquetCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := driveToPhase(t, domain.PiquetPhasePlay)
	p := &PiquetCuiPresenter{}
	out := p.Output(g, nil)
	if out == "" {
		t.Error("expected non-empty output in play phase")
	}
}

func TestPiquetCuiPresenter_Output_ScoreOrGameEnd(t *testing.T) {
	// After driving plays to end with DealsPerPartie=1, phase is GameEnd or Score
	g := driveToPhase(t, domain.PiquetPhasePlay)
	for g.GetPhase() == domain.PiquetPhasePlay {
		g.CpuPlay()
	}
	p := &PiquetCuiPresenter{}
	out := p.Output(g, nil)
	if out == "" {
		t.Error("expected non-empty output in score/game-end phase")
	}
}

// TestPiquetCuiPresenter_DeclResultsRendering uses JSON round-trip to set up
// declaration results and exercise piquetFormatDeclResult/piquetKindLabel paths.
func TestPiquetCuiPresenter_DeclResultsRendering(t *testing.T) {
	g := newPiquetForPresenter(t)
	// Force phase to Declaration with seeded results via JSON round-trip
	data, _ := json.Marshal(g)
	var raw map[string]any
	_ = json.Unmarshal(data, &raw)
	raw["ph"] = int(domain.PiquetPhaseDeclaration)
	raw["dr"] = []map[string]any{
		{"k": 0, "w": 0, "sc": 5, "sb": 0},   // Point — Elder scores
		{"k": 1, "w": -1, "sc": 0, "sb": -1}, // Sequence — tied
		{"k": 2, "w": 1, "sc": 14, "sb": 1},  // Set — Younger scores
	}
	mod, _ := json.Marshal(raw)
	g2 := &domain.Piquet{}
	if err := json.Unmarshal(mod, g2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	p := &PiquetCuiPresenter{}
	out := p.Output(g2, nil)
	if !strings.Contains(out, "Point") && !strings.Contains(out, "Sequence") && !strings.Contains(out, "Set") {
		t.Errorf("expected at least one declaration label, got: %s", out)
	}
}

// TestPiquetCuiPresenter_GameEndWithWinner exercises piquetWriteGameEndView's winner branch.
func TestPiquetCuiPresenter_GameEndWithWinner(t *testing.T) {
	g := newPiquetForPresenter(t)
	data, _ := json.Marshal(g)
	var raw map[string]any
	_ = json.Unmarshal(data, &raw)
	raw["ph"] = int(domain.PiquetPhaseGameEnd)
	raw["wi"] = 0
	raw["ge"] = true
	mod, _ := json.Marshal(raw)
	g2 := &domain.Piquet{}
	_ = json.Unmarshal(mod, g2)
	p := &PiquetCuiPresenter{}
	out := p.Output(g2, nil)
	if !strings.Contains(out, "0") {
		t.Errorf("expected winner idx in output: %s", out)
	}
}

// TestPiquetCuiPresenter_GameEndDraw exercises the draw branch.
func TestPiquetCuiPresenter_GameEndDraw(t *testing.T) {
	g := newPiquetForPresenter(t)
	data, _ := json.Marshal(g)
	var raw map[string]any
	_ = json.Unmarshal(data, &raw)
	raw["ph"] = int(domain.PiquetPhaseGameEnd)
	raw["wi"] = -1
	raw["ge"] = true
	mod, _ := json.Marshal(raw)
	g2 := &domain.Piquet{}
	_ = json.Unmarshal(mod, g2)
	p := &PiquetCuiPresenter{}
	out := p.Output(g2, nil)
	if out == "" {
		t.Error("expected non-empty draw output")
	}
}

func TestPiquetCuiPresenter_PlayerNames_ElderAndYounger(t *testing.T) {
	origLang := i18n.Lang()
	defer i18n.SetLang(origLang)
	origColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origColor)

	tests := []struct {
		name        string
		lang        string
		humanElder  bool
		wantElder   string
		wantYounger string
	}{
		{
			name:        "ja: human Elder, CPU Younger",
			lang:        "ja",
			humanElder:  true,
			wantElder:   "Elder (あなた)",
			wantYounger: "Younger (CPU 1)",
		},
		{
			name:        "ja: CPU Elder, human Younger",
			lang:        "ja",
			humanElder:  false,
			wantElder:   "Elder (CPU 0)",
			wantYounger: "Younger (あなた)",
		},
		{
			name:        "en: human Elder, CPU Younger",
			lang:        "en",
			humanElder:  true,
			wantElder:   "Elder (You)",
			wantYounger: "Younger (CPU 1)",
		},
		{
			name:        "en: CPU Elder, human Younger",
			lang:        "en",
			humanElder:  false,
			wantElder:   "Elder (CPU 0)",
			wantYounger: "Younger (You)",
		},
	}

	p := &PiquetCuiPresenter{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i18n.SetLang(tc.lang)
			players := []*domain.PiquetPlayer{
				domain.NewPiquetPlayer(tc.humanElder),
				domain.NewPiquetPlayer(!tc.humanElder),
			}
			g := domain.NewPiquet(domain.NewTrumpCardsBelote(), players, domain.DefaultPiquetConfig())
			g.Reset()

			out := p.Output(g, nil)

			// Assert actual output strings for Elder and Younger
			if !strings.Contains(out, tc.wantElder) {
				t.Errorf("expected %q in output, got:\n%s", tc.wantElder, out)
			}
			if !strings.Contains(out, tc.wantYounger) {
				t.Errorf("expected %q in output, got:\n%s", tc.wantYounger, out)
			}

			// Negative control: raw "(CPU)" or " (CPU)" must not appear
			if strings.Contains(out, "(CPU)") {
				t.Errorf("output must not contain raw (CPU), got:\n%s", out)
			}
			if strings.Contains(out, " (CPU)") {
				t.Errorf("output must not contain raw \" (CPU)\", got:\n%s", out)
			}
		})
	}
}
