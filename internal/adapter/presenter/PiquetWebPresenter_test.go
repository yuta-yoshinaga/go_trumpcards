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

// TestPiquetWebPresenter_PhaseMessageCodes exercises piquetPhaseMessageCode across all phases.
func TestPiquetWebPresenter_PhaseMessageCodes(t *testing.T) {
	cases := []struct {
		phase domain.PiquetPhase
		want  string
	}{
		{domain.PiquetPhaseExchange, "piquet.exchange"},
		{domain.PiquetPhaseDeclaration, "piquet.declaration"},
		{domain.PiquetPhasePlay, "piquet.play"},
		{domain.PiquetPhaseScore, "piquet.score"},
		{domain.PiquetPhaseGameEnd, "piquet.gameEnd"},
	}
	for _, tc := range cases {
		g := newPiquetForPresenter(t)
		// JSON round-trip to set phase
		data, _ := json.Marshal(g)
		var raw map[string]any
		_ = json.Unmarshal(data, &raw)
		raw["ph"] = int(tc.phase)
		mod, _ := json.Marshal(raw)
		g2 := &domain.Piquet{}
		_ = json.Unmarshal(mod, g2)

		p := &PiquetWebPresenter{}
		out := p.Output(g2, nil)
		var parsed map[string]any
		_ = json.Unmarshal([]byte(out), &parsed)
		if parsed["messageCode"] != tc.want {
			t.Errorf("phase %d → messageCode=%v, want %s", tc.phase, parsed["messageCode"], tc.want)
		}
	}
}

// TestPiquetWebPresenter_DeclResultsWithClaims exercises piquetBuildClaim + piquetBuildClaims.
func TestPiquetWebPresenter_DeclResultsWithClaims(t *testing.T) {
	g := newPiquetForPresenter(t)
	data, _ := json.Marshal(g)
	var raw map[string]any
	_ = json.Unmarshal(data, &raw)
	raw["ph"] = int(domain.PiquetPhaseDeclaration)
	// Seed declaration results with claims + sets
	raw["dr"] = []map[string]any{
		{
			"k":  0,
			"ec": map[string]any{"l": 5, "p": 41, "s": 1, "c": []map[string]any{}},
			"yc": map[string]any{"l": 4, "p": 34, "s": 3, "c": []map[string]any{}},
			"w":  0, "sc": 5, "sb": 0,
		},
		{
			"k":  2,
			"ec": map[string]any{"l": 4, "r": 8, "c": []map[string]any{}},
			"w":  0, "sc": 14, "sb": 0,
			"sets": []map[string]any{
				{"l": 4, "r": 8, "c": []map[string]any{}},
				{"l": 3, "r": 7, "c": []map[string]any{}},
			},
		},
	}
	mod, _ := json.Marshal(raw)
	g2 := &domain.Piquet{}
	if err := json.Unmarshal(mod, g2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	p := &PiquetWebPresenter{}
	out := p.Output(g2, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	decls, ok := parsed["declResults"].([]any)
	if !ok || len(decls) != 2 {
		t.Fatalf("expected 2 declResults, got %v", parsed["declResults"])
	}
	first := decls[0].(map[string]any)
	if first["elderClaim"] == nil {
		t.Error("expected elderClaim to be populated")
	}
	second := decls[1].(map[string]any)
	if second["sets"] == nil {
		t.Error("expected sets to be populated for Set declaration")
	}
}

// TestPiquetWebPresenter_HintOutput_WithHint exercises the hint-populated path.
func TestPiquetWebPresenter_HintOutput_WithHint(t *testing.T) {
	// In play phase the domain's GetHint returns a hint when current player is the elder.
	// Drive forward to a play state where CPU is current player.
	players := []*domain.PiquetPlayer{
		domain.NewPiquetPlayer(false),
		domain.NewPiquetPlayer(false),
	}
	g := domain.NewPiquet(domain.NewTrumpCardsBelote(), players,
		domain.PiquetConfig{DealsPerPartie: 1, CpuDifficulty: domain.PiquetCpuDifficultyNormal})
	g.Reset()
	for g.GetPhase() == domain.PiquetPhaseExchange {
		g.CpuPlay()
	}
	for g.GetPhase() == domain.PiquetPhaseDeclaration {
		_, _ = g.ResolveDeclaration()
	}
	p := &PiquetWebPresenter{}
	out := p.HintOutput(g)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["messageCode"] != "piquet.hintAvailable" && parsed["messageCode"] != "piquet.noHint" {
		t.Errorf("messageCode unexpected: %v", parsed["messageCode"])
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// 交換と申告を CPU に進めさせた時点でヒントが必ず出ます（100 回中 100 回で確認）。
// Piquet の GetHint は現在手番を引数に取るので、呼び出し式ごと HintOutput から
// 写しています。
func TestPiquetWebPresenterOutputCarriesTheHint(t *testing.T) {
	// **人間席のあるフィクスチャで組む。**前のテストは両席とも CPU で作っていて、
	// 「CPU の手番でも CPU の手札のヒントが出る」という本題を検出できなかった
	// (#4554 のレビュー指摘)。
	//
	// フェーズは進めない。人間席があると `for phase == Exchange { CpuPlay() }` は
	// 止まらない（CpuPlay は人間席では何もしない）。配り終えた交換フェーズで
	// エルダーに捨て札ヒントが出るので、席の是非だけをここで確かめられる。
	newGame := func(humanSeat int) *domain.Piquet {
		players := []*domain.PiquetPlayer{
			domain.NewPiquetPlayer(humanSeat == 0),
			domain.NewPiquetPlayer(humanSeat == 1),
		}
		g := domain.NewPiquet(domain.NewTrumpCardsBelote(), players,
			domain.PiquetConfig{DealsPerPartie: 1, CpuDifficulty: domain.PiquetCpuDifficultyNormal})
		g.Reset()
		return g
	}

	t.Run("carries the hint on the human's turn", func(t *testing.T) {
		g := newGame(newGame(0).GetCurrentPlayerIdx())
		if g.GetHint(g.GetCurrentPlayerIdx()) == nil {
			t.Fatal("fixture must actually produce a hint")
		}

		var parsed map[string]any
		if err := json.Unmarshal([]byte(new(PiquetWebPresenter).Output(g, nil)), &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if parsed["hint"] == nil {
			t.Error("Output must carry the hint -- the frontend reads state.hint")
		}
	})

	// **相手の席のヒントは出さない。**出すと CPU の手札を説明する行が人間に見える。
	t.Run("stays silent for a CPU seat", func(t *testing.T) {
		g := newGame(1 - newGame(0).GetCurrentPlayerIdx())
		if g.GetHint(g.GetCurrentPlayerIdx()) != nil {
			t.Error("GetHint must not describe a CPU seat's hand")
		}

		var parsed map[string]any
		if err := json.Unmarshal([]byte(new(PiquetWebPresenter).Output(g, nil)), &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if parsed["hint"] != nil {
			t.Error("Output must not leak a hint for the CPU's hand")
		}
	})
}
