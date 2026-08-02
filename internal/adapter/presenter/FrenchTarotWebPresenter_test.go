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

func newFrenchTarotGame() *domain.FrenchTarot {
	return domain.NewDefaultFrenchTarot()
}

func TestFrenchTarotWebPresenter_Output(t *testing.T) {
	g := newFrenchTarotGame()
	g.Reset()
	p := &presenter.FrenchTarotWebPresenter{}

	var parsed controller.FrenchTarotWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Players) != 4 {
		t.Errorf("players = %d, want 4", len(parsed.Players))
	}
	if parsed.Config.TargetDeals != domain.FrenchTarotDefaultDeals {
		t.Errorf("targetDeals = %d", parsed.Config.TargetDeals)
	}
	// Human cards revealed, CPU cards hidden.
	if len(parsed.Players[0].Cards) == 0 {
		t.Errorf("human cards should be revealed")
	}
	if len(parsed.Players[1].Cards) != 0 {
		t.Errorf("CPU cards should be hidden")
	}
	if parsed.ChienCount != domain.FrenchTarotChienSize {
		t.Errorf("chienCount = %d", parsed.ChienCount)
	}
}

// TestFrenchTarotWebPresenter_ProceduralFaces asserts a trump serializes with
// deck:"tarot" + purple, the Excuse with label "Excuse" + gold, and a suit card
// with the suit colour.
func TestFrenchTarotWebPresenter_ProceduralFaces(t *testing.T) {
	g := newFrenchTarotGame()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetPhase(domain.FrenchTarotPhasePlay)
	// Human hand: the Excuse + a heart king.
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.FrenchTarotExcuseDesign, domain.FrenchTarotExcuseValue, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 14, false))
	// Trick: a trump and a spade.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.FrenchTarotTrumpDesign, 21, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
	})

	p := &presenter.FrenchTarotWebPresenter{}
	var parsed controller.FrenchTarotWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Human's Excuse: label "Excuse", color gold, deck tarot; heart king red.
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
	// Trick: trump serialized with deck tarot + purple; spade black.
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

func TestFrenchTarotWebPresenter_ChienRevealedToHumanDeclarer(t *testing.T) {
	g := newFrenchTarotGame()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetChien([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.FrenchTarotTrumpDesign, 2, false),
	})
	g.SetPhase(domain.FrenchTarotPhaseChien)

	p := &presenter.FrenchTarotWebPresenter{}
	var parsed controller.FrenchTarotWebOutput
	_ = json.Unmarshal([]byte(p.Output(g, nil)), &parsed)
	if len(parsed.Chien) != 2 {
		t.Errorf("chien should be revealed to human declarer, got %d", len(parsed.Chien))
	}
	if parsed.Chien[0].Deck != "tarot" {
		t.Errorf("chien cards should carry deck:tarot")
	}
}

func TestFrenchTarotWebPresenter_Error(t *testing.T) {
	g := newFrenchTarotGame()
	g.Reset()
	p := &presenter.FrenchTarotWebPresenter{}
	out := p.Output(g, errors.New("boom"))
	if !strings.Contains(out, "boom") {
		t.Errorf("error message not propagated: %s", out)
	}
}

func TestFrenchTarotWebPresenter_Hint(t *testing.T) {
	g := newFrenchTarotGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	p := &presenter.FrenchTarotWebPresenter{}
	out := p.HintOutput(g)
	if !strings.Contains(out, "hint") && !strings.Contains(out, "reason") {
		t.Errorf("expected hint in output: %s", out)
	}
}

func TestFrenchTarotWebPresenter_ActionLog(t *testing.T) {
	g := newFrenchTarotGame()
	g.Reset()
	p := &presenter.FrenchTarotWebPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log output is empty")
	}
}

func TestFrenchTarotWebPresenter_PhaseMessages(t *testing.T) {
	p := &presenter.FrenchTarotWebPresenter{}
	for _, phase := range []domain.FrenchTarotPhase{
		domain.FrenchTarotPhaseBid,
		domain.FrenchTarotPhaseChien,
		domain.FrenchTarotPhasePlay,
		domain.FrenchTarotPhaseTrickEnd,
		domain.FrenchTarotPhaseRoundEnd,
	} {
		g := newFrenchTarotGame()
		g.Reset()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.FrenchTarotBidGarde)
		g.SetPhase(phase)
		if phase == domain.FrenchTarotPhasePlay {
			g.SetCurrentTrick([]*domain.TrickCard{
				{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			})
		}
		if out := p.Output(g, nil); out == "" {
			t.Errorf("phase %d output empty", phase)
		}
	}
}

// frenchTarotForceEnd builds a match that ends after one deal, pre-loading the
// player scores so the post-deal totals resolve to a specific leader (or a tie).
// The declarer captures nothing (a heavy Petite loss: declarer −243, each
// defender +81), so the final standings are fully determined by the presets.
func frenchTarotForceEnd(declarer int, preset [domain.FrenchTarotPlayerCnt]int) *domain.FrenchTarot {
	g := domain.NewDefaultFrenchTarot()
	g.Reset()
	cfg := domain.DefaultFrenchTarotConfig()
	cfg.TargetDeals = 1
	g.SetConfig(cfg)
	g.SetRoundNumber(1)
	g.SetDeclarerIdx(declarer)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetPlayerScores(preset)
	g.SetPhase(domain.FrenchTarotPhaseRoundEnd)
	g.ScoreRound()
	return g
}

func TestFrenchTarotWebPresenter_WinnerMessages(t *testing.T) {
	p := &presenter.FrenchTarotWebPresenter{}
	cases := []struct {
		name     string
		declarer int
		preset   [domain.FrenchTarotPlayerCnt]int
		wantCode string
	}{
		{"humanWin", 1, [domain.FrenchTarotPlayerCnt]int{300, 0, 0, 0}, "frenchtarot.result.humanWin"},
		{"cpuWin", 0, [domain.FrenchTarotPlayerCnt]int{0, 10, 0, 0}, "frenchtarot.result.cpuWin"},
		{"draw", 0, [domain.FrenchTarotPlayerCnt]int{324, 0, 0, 0}, "frenchtarot.result.draw"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := frenchTarotForceEnd(c.declarer, c.preset)
			if !g.GetGameEndFlag() {
				t.Fatalf("expected game-end state")
			}
			var parsed controller.FrenchTarotWebOutput
			if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if parsed.MessageCode != c.wantCode {
				t.Errorf("messageCode = %q, want %q", parsed.MessageCode, c.wantCode)
			}
		})
	}
}

func TestFrenchTarotWebPresenter_GameEnd(t *testing.T) {
	g := newFrenchTarotGame()
	g.Reset()
	// Force a game-end state via a full deal at the target-deals boundary.
	cfg := domain.DefaultFrenchTarotConfig()
	cfg.TargetDeals = 1
	g.SetConfig(cfg)
	g.SetRoundNumber(1)
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidGarde)
	g.SetTrickNumber(domain.FrenchTarotTrickCount)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.FrenchTarotPhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 3, false)},
		{PlayerIdx: 0, Card: domain.NewCard(domain.FrenchTarotTrumpDesign, 1, false)},
	})
	g.ResolveTrick()
	p := &presenter.FrenchTarotWebPresenter{}
	out := p.Output(g, nil)
	if !strings.Contains(out, "result") && !strings.Contains(out, "ゲーム終了") {
		t.Errorf("game-end message missing: %s", out)
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestFrenchTarotWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ状態。
	g := newFrenchTarotGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	out := new(presenter.FrenchTarotWebPresenter).Output(g, nil)
	if !strings.Contains(out, `"hint"`) {
		t.Error("Output must carry the hint -- the frontend reads state.hint")
	}
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	if strings.Contains(out, "frenchtarot.hintRequested") {
		t.Error("Output must not mark the response as a requested hint")
	}
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestFrenchTarotWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := newFrenchTarotGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	if !strings.Contains(new(presenter.FrenchTarotWebPresenter).HintOutput(g), "frenchtarot.hintRequested") {
		t.Error("HintOutput must mark the response as a requested hint")
	}
}
