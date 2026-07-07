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

func newCegoGame() *domain.Cego {
	return domain.NewDefaultCego()
}

func TestCegoWebPresenter_Output(t *testing.T) {
	g := newCegoGame()
	g.Reset()
	p := &presenter.CegoWebPresenter{}

	var parsed controller.CegoWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Players) != 4 {
		t.Errorf("players = %d, want 4", len(parsed.Players))
	}
	if parsed.Config.TargetDeals != domain.CegoDefaultDeals {
		t.Errorf("targetDeals = %d", parsed.Config.TargetDeals)
	}
	if len(parsed.Players[0].Cards) == 0 {
		t.Errorf("human cards should be revealed")
	}
	if len(parsed.Players[1].Cards) != 0 {
		t.Errorf("CPU cards should be hidden")
	}
	if parsed.BlindCount != domain.CegoBlindSize {
		t.Errorf("blindCount = %d, want %d", parsed.BlindCount, domain.CegoBlindSize)
	}
}

// TestCegoWebPresenter_BlindNeverExposed asserts the face-down blind cards are
// NOT serialised, both before and during the exchange (only the count leaks).
func TestCegoWebPresenter_BlindNeverExposed(t *testing.T) {
	g := newCegoGame()
	g.Reset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	p := &presenter.CegoWebPresenter{}

	// Before the exchange (bid phase).
	var before controller.CegoWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &before); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(before.Blind) != 0 {
		t.Errorf("blind cards leaked before exchange: got %d", len(before.Blind))
	}
	if before.BlindCount != domain.CegoBlindSize {
		t.Errorf("blindCount = %d, want %d", before.BlindCount, domain.CegoBlindSize)
	}

	// During the exchange, as the human declarer.
	g.SetContractType(domain.CegoContractCego)
	g.SetPhase(domain.CegoPhaseExchange)
	var during controller.CegoWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &during); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(during.Blind) != 0 {
		t.Errorf("blind cards leaked during exchange: got %d", len(during.Blind))
	}
}

// TestCegoWebPresenter_ProceduralFaces asserts a trump serializes with
// deck:"tarot" + purple, the Sküs with label "Sküs" + gold, and suit cards with
// the suit colour.
func TestCegoWebPresenter_ProceduralFaces(t *testing.T) {
	g := newCegoGame()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhasePlay)
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CegoSkusDesign, domain.CegoSkusValue, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, domain.CegoKingValue, false))
	g.SetCurrentTrick([]*domain.CegoTrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CegoTrumpDesign, 21, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
	})

	p := &presenter.CegoWebPresenter{}
	var parsed controller.CegoWebOutput
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

func TestCegoWebPresenter_Error(t *testing.T) {
	g := newCegoGame()
	g.Reset()
	p := &presenter.CegoWebPresenter{}
	out := p.Output(g, errors.New("boom"))
	if !strings.Contains(out, "boom") {
		t.Errorf("error message not propagated: %s", out)
	}
}

func TestCegoWebPresenter_Hint(t *testing.T) {
	g := newCegoGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	p := &presenter.CegoWebPresenter{}
	out := p.HintOutput(g)
	if !strings.Contains(out, "hint") && !strings.Contains(out, "reason") {
		t.Errorf("expected hint in output: %s", out)
	}
}

func TestCegoWebPresenter_ActionLog(t *testing.T) {
	g := newCegoGame()
	g.Reset()
	p := &presenter.CegoWebPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log output is empty")
	}
}

func TestCegoWebPresenter_PhaseMessages(t *testing.T) {
	p := &presenter.CegoWebPresenter{}
	for _, phase := range []domain.CegoPhase{
		domain.CegoPhaseBid,
		domain.CegoPhaseContract,
		domain.CegoPhaseExchange,
		domain.CegoPhasePlay,
		domain.CegoPhaseTrickEnd,
		domain.CegoPhaseRoundEnd,
	} {
		g := newCegoGame()
		g.Reset()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(phase)
		out := p.Output(g, nil)
		if out == "" {
			t.Errorf("empty output for phase %d", phase)
		}
	}
}
