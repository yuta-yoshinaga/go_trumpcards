//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newScartoGame() *domain.Scarto {
	return domain.NewDefaultScarto()
}

func TestScartoWebPresenter_Output(t *testing.T) {
	g := newScartoGame()
	g.Reset()
	p := &presenter.ScartoWebPresenter{}

	var parsed controller.ScartoWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Players) != 3 {
		t.Errorf("players = %d, want 3", len(parsed.Players))
	}
	if parsed.Config.TargetDeals != domain.ScartoDefaultDeals {
		t.Errorf("targetDeals = %d", parsed.Config.TargetDeals)
	}
	// Human (dealer) cards revealed, CPU cards hidden.
	if len(parsed.Players[0].Cards) == 0 {
		t.Errorf("human cards should be revealed")
	}
	if len(parsed.Players[1].Cards) != 0 {
		t.Errorf("CPU cards should be hidden")
	}
	if !parsed.IsHumanScarto {
		t.Errorf("human dealer should be in scarto turn after reset")
	}
}

// TestScartoWebPresenter_ProceduralFaces asserts a trump serializes with
// deck:"tarot" + purple, the Excuse with label "Excuse" + gold, and suit cards
// with the suit colour.
func TestScartoWebPresenter_ProceduralFaces(t *testing.T) {
	g := newScartoGame()
	g.SetPhase(domain.ScartoPhasePlay)
	// Human hand: the Excuse + a heart king.
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.ScartoExcuseDesign, domain.ScartoExcuseValue, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 14, false))
	// Trick: a trump 21 and a spade.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.ScartoTrumpDesign, 21, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
	})

	p := &presenter.ScartoWebPresenter{}
	var parsed controller.ScartoWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var sawExcuse, sawRedKing bool
	for _, c := range parsed.Players[0].Cards {
		if c.Deck == "tarot" && c.Label == "Excuse" && c.Color == "gold" {
			sawExcuse = true
		}
		if c.Deck == "tarot" && c.Color == "red" && c.Label == "R" {
			sawRedKing = true
		}
	}
	if !sawExcuse {
		t.Errorf("expected Excuse card (label Excuse, gold), got %+v", parsed.Players[0].Cards)
	}
	if !sawRedKing {
		t.Errorf("expected a red king (label R) in the hand")
	}
	var sawTrump, sawSpade bool
	for _, tc := range parsed.CurrentTrick {
		if tc.Card.Deck == "tarot" && tc.Card.Color == "purple" && tc.Card.Label == "21" {
			sawTrump = true
		}
		if tc.Card.Deck == "tarot" && tc.Card.Color == "black" {
			sawSpade = true
		}
	}
	if !sawTrump {
		t.Errorf("expected a trump card with deck:tarot + purple, got %+v", parsed.CurrentTrick)
	}
	if !sawSpade {
		t.Errorf("expected a black spade card in the trick")
	}
}

func TestScartoWebPresenter_Error(t *testing.T) {
	g := newScartoGame()
	g.Reset()
	p := &presenter.ScartoWebPresenter{}
	out := p.Output(g, errors.New("boom"))
	if !strings.Contains(out, "boom") {
		t.Errorf("error message not propagated: %s", out)
	}
}

func TestScartoWebPresenter_Hint(t *testing.T) {
	g := newScartoGame()
	g.Reset() // human dealer in scarto phase
	p := &presenter.ScartoWebPresenter{}
	out := p.HintOutput(g)
	if !strings.Contains(out, "hint") && !strings.Contains(out, "reason") {
		t.Errorf("expected hint in output: %s", out)
	}
}

func TestScartoWebPresenter_ActionLog(t *testing.T) {
	g := newScartoGame()
	g.Reset()
	p := &presenter.ScartoWebPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log output is empty")
	}
}

func TestScartoWebPresenter_PhaseMessages(t *testing.T) {
	p := &presenter.ScartoWebPresenter{}
	for _, phase := range []domain.ScartoPhase{
		domain.ScartoPhaseScarto,
		domain.ScartoPhasePlay,
		domain.ScartoPhaseTrickEnd,
		domain.ScartoPhaseRoundEnd,
	} {
		g := newScartoGame()
		g.Reset()
		g.SetPhase(phase)
		if phase == domain.ScartoPhasePlay {
			g.SetCurrentTrick([]*domain.TrickCard{
				{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			})
		}
		if out := p.Output(g, nil); out == "" {
			t.Errorf("phase %d output empty", phase)
		}
	}
}

// scartoForceEnd ends a one-deal match with preset scores. Since no tricks are
// captured the deal scores are all zero, so the final standings equal the presets.
func scartoForceEnd(preset [domain.ScartoPlayerCnt]int) *domain.Scarto {
	g := domain.NewDefaultScarto()
	g.Reset()
	cfg := domain.DefaultScartoConfig()
	cfg.TargetDeals = 1
	g.SetConfig(cfg)
	g.SetRoundNumber(1)
	g.SetPlayerScores(preset)
	g.SetPhase(domain.ScartoPhaseRoundEnd)
	g.ScoreRound()
	return g
}

func TestScartoWebPresenter_WinnerMessages(t *testing.T) {
	p := &presenter.ScartoWebPresenter{}
	cases := []struct {
		name     string
		preset   [domain.ScartoPlayerCnt]int
		wantCode string
	}{
		{"humanWin", [domain.ScartoPlayerCnt]int{300, 0, 0}, "scarto.result.humanWin"},
		{"cpuWin", [domain.ScartoPlayerCnt]int{0, 10, 0}, "scarto.result.cpuWin"},
		{"draw", [domain.ScartoPlayerCnt]int{5, 5, 5}, "scarto.result.draw"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := scartoForceEnd(c.preset)
			if !g.GetGameEndFlag() {
				t.Fatalf("expected game-end state")
			}
			var parsed controller.ScartoWebOutput
			if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if parsed.MessageCode != c.wantCode {
				t.Errorf("messageCode = %q, want %q", parsed.MessageCode, c.wantCode)
			}
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestScartoWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ状態。300 回試して nil 0 件で確認済み。
	g := newScartoGame()
	g.Reset() // human dealer in scarto phase
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	out := new(presenter.ScartoWebPresenter).Output(g, nil)
	if !strings.Contains(out, `"hint"`) {
		t.Error("Output must carry the hint -- the frontend reads state.hint")
	}
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	if strings.Contains(out, "scarto.hintRequested") {
		t.Error("Output must not mark the response as a requested hint")
	}
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestScartoWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := newScartoGame()
	g.Reset()
	if !strings.Contains(new(presenter.ScartoWebPresenter).HintOutput(g), "scarto.hintRequested") {
		t.Error("HintOutput must mark the response as a requested hint")
	}
}
