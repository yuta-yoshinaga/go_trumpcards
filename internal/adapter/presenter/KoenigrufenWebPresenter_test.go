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

func newKoenigrufenGame() *domain.Koenigrufen {
	return domain.NewDefaultKoenigrufen()
}

func TestKoenigrufenWebPresenter_Output(t *testing.T) {
	g := newKoenigrufenGame()
	g.Reset()
	p := &presenter.KoenigrufenWebPresenter{}

	var parsed controller.KoenigrufenWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Players) != 4 {
		t.Errorf("players = %d, want 4", len(parsed.Players))
	}
	if parsed.Config.TargetDeals != domain.KoenigrufenDefaultDeals {
		t.Errorf("targetDeals = %d", parsed.Config.TargetDeals)
	}
	if len(parsed.Players[0].Cards) == 0 {
		t.Errorf("human cards should be revealed")
	}
	if len(parsed.Players[1].Cards) != 0 {
		t.Errorf("CPU cards should be hidden")
	}
	if parsed.TalonCount != domain.KoenigrufenTalonSize {
		t.Errorf("talonCount = %d", parsed.TalonCount)
	}
}

// TestKoenigrufenWebPresenter_ProceduralFaces asserts a trump serializes with
// deck:"tarot" + purple, the Sküs with label "Sküs" + gold, and suit cards with
// the suit colour.
func TestKoenigrufenWebPresenter_ProceduralFaces(t *testing.T) {
	g := newKoenigrufenGame()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetPhase(domain.KoenigrufenPhasePlay)
	// Human hand: the Sküs + a heart king.
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.KoenigrufenSkusDesign, domain.KoenigrufenSkusValue, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, domain.KoenigrufenKingValue, false))
	// Trick: a trump and a spade.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.KoenigrufenTrumpDesign, 21, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
	})

	p := &presenter.KoenigrufenWebPresenter{}
	var parsed controller.KoenigrufenWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var sawSkus, sawRedKing bool
	for _, c := range parsed.Players[0].Cards {
		if c.Deck == "tarot" && c.Label == "Sküs" && c.Color == "gold" {
			sawSkus = true
		}
		if c.Deck == "tarot" && c.Color == "red" && c.Label == "K" {
			sawRedKing = true
		}
	}
	if !sawSkus {
		t.Errorf("expected Sküs card (label Sküs, gold), got %+v", parsed.Players[0].Cards)
	}
	if !sawRedKing {
		t.Errorf("expected a red king (label K) in the hand")
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

// TestKoenigrufenWebPresenter_PartnerHiddenUntilRevealed asserts the secret
// partner is NOT leaked: while partnerRevealed is false, partnerIdx must be -1
// and no player may be flagged IsPartner. Once revealed, both surface.
func TestKoenigrufenWebPresenter_PartnerHiddenUntilRevealed(t *testing.T) {
	g := newKoenigrufenGame()
	g.Reset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetCalledKing(domain.CardDesignHeart)
	g.SetPartnerIdx(2) // server-side partner
	g.SetPartnerRevealed(false)
	g.SetPhase(domain.KoenigrufenPhasePlay)

	p := &presenter.KoenigrufenWebPresenter{}
	var parsed controller.KoenigrufenWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Unrevealed: partnerIdx MUST be -1 and no IsPartner flag set.
	if parsed.PartnerIdx != -1 {
		t.Errorf("partnerIdx leaked while unrevealed: got %d, want -1", parsed.PartnerIdx)
	}
	if parsed.PartnerRevealed {
		t.Errorf("partnerRevealed should be false")
	}
	for _, pl := range parsed.Players {
		if pl.IsPartner {
			t.Errorf("player %d IsPartner leaked while partnership unrevealed", pl.ID)
		}
	}

	// Now reveal → partnerIdx and IsPartner surface.
	g.SetPartnerRevealed(true)
	var parsed2 controller.KoenigrufenWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed2.PartnerIdx != 2 {
		t.Errorf("partnerIdx = %d, want 2 after reveal", parsed2.PartnerIdx)
	}
	if !parsed2.Players[2].IsPartner {
		t.Errorf("player 2 should be flagged IsPartner after reveal")
	}
}

func TestKoenigrufenWebPresenter_TalonRevealedToHumanDeclarer(t *testing.T) {
	g := newKoenigrufenGame()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetTalon([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.KoenigrufenTrumpDesign, 2, false),
	})
	g.SetPhase(domain.KoenigrufenPhaseTalon)

	p := &presenter.KoenigrufenWebPresenter{}
	var parsed controller.KoenigrufenWebOutput
	_ = json.Unmarshal([]byte(p.Output(g, nil)), &parsed)
	if len(parsed.Talon) != 2 {
		t.Errorf("talon should be revealed to human declarer, got %d", len(parsed.Talon))
	}
	if parsed.Talon[0].Deck != "tarot" {
		t.Errorf("talon cards should carry deck:tarot")
	}
}

func TestKoenigrufenWebPresenter_Error(t *testing.T) {
	g := newKoenigrufenGame()
	g.Reset()
	p := &presenter.KoenigrufenWebPresenter{}
	out := p.Output(g, errors.New("boom"))
	if !strings.Contains(out, "boom") {
		t.Errorf("error message not propagated: %s", out)
	}
}

func TestKoenigrufenWebPresenter_Hint(t *testing.T) {
	g := newKoenigrufenGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	p := &presenter.KoenigrufenWebPresenter{}
	out := p.HintOutput(g)
	if !strings.Contains(out, "hint") && !strings.Contains(out, "reason") {
		t.Errorf("expected hint in output: %s", out)
	}
}

func TestKoenigrufenWebPresenter_ActionLog(t *testing.T) {
	g := newKoenigrufenGame()
	g.Reset()
	p := &presenter.KoenigrufenWebPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log output is empty")
	}
}

func TestKoenigrufenWebPresenter_PhaseMessages(t *testing.T) {
	p := &presenter.KoenigrufenWebPresenter{}
	for _, phase := range []domain.KoenigrufenPhase{
		domain.KoenigrufenPhaseBid,
		domain.KoenigrufenPhaseCall,
		domain.KoenigrufenPhaseTalon,
		domain.KoenigrufenPhasePlay,
		domain.KoenigrufenPhaseTrickEnd,
		domain.KoenigrufenPhaseRoundEnd,
	} {
		g := newKoenigrufenGame()
		g.Reset()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.KoenigrufenBidRufer)
		g.SetPhase(phase)
		if phase == domain.KoenigrufenPhasePlay {
			g.SetCurrentTrick([]*domain.TrickCard{
				{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			})
		}
		if out := p.Output(g, nil); out == "" {
			t.Errorf("phase %d output empty", phase)
		}
	}
}

// koenigrufenForceEnd builds a match that ends after one deal (declarer plays
// solo and captures nothing → a heavy loss: declarer −189, each opponent +63),
// pre-loading player scores so the final standings resolve to a set leader/tie.
func koenigrufenForceEnd(declarer int, preset [domain.KoenigrufenPlayerCnt]int) *domain.Koenigrufen {
	g := domain.NewDefaultKoenigrufen()
	g.Reset()
	cfg := domain.DefaultKoenigrufenConfig()
	cfg.TargetDeals = 1
	g.SetConfig(cfg)
	g.SetRoundNumber(1)
	g.SetDeclarerIdx(declarer)
	g.SetPartnerIdx(-1) // solo
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetPlayerScores(preset)
	g.SetPhase(domain.KoenigrufenPhaseRoundEnd)
	g.ScoreRound()
	return g
}

func TestKoenigrufenWebPresenter_WinnerMessages(t *testing.T) {
	p := &presenter.KoenigrufenWebPresenter{}
	cases := []struct {
		name     string
		declarer int
		preset   [domain.KoenigrufenPlayerCnt]int
		wantCode string
	}{
		{"humanWin", 1, [domain.KoenigrufenPlayerCnt]int{300, 0, 0, 0}, "koenigrufen.result.humanWin"},
		{"cpuWin", 0, [domain.KoenigrufenPlayerCnt]int{0, 10, 0, 0}, "koenigrufen.result.cpuWin"},
		{"draw", 0, [domain.KoenigrufenPlayerCnt]int{252, 0, 0, 0}, "koenigrufen.result.draw"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := koenigrufenForceEnd(c.declarer, c.preset)
			if !g.GetGameEndFlag() {
				t.Fatalf("expected game-end state")
			}
			var parsed controller.KoenigrufenWebOutput
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
func TestKoenigrufenWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ状態。
	g := newKoenigrufenGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	out := new(presenter.KoenigrufenWebPresenter).Output(g, nil)
	if !strings.Contains(out, `"hint"`) {
		t.Error("Output must carry the hint -- the frontend reads state.hint")
	}
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	if strings.Contains(out, "koenigrufen.hintRequested") {
		t.Error("Output must not mark the response as a requested hint")
	}
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestKoenigrufenWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := newKoenigrufenGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	if !strings.Contains(new(presenter.KoenigrufenWebPresenter).HintOutput(g), "koenigrufen.hintRequested") {
		t.Error("HintOutput must mark the response as a requested hint")
	}
}
