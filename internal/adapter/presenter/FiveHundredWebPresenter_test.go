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

func newFiveHundredGame() *domain.FiveHundred {
	players := []*domain.FiveHundredPlayer{
		domain.NewFiveHundredPlayer(true, 0),
		domain.NewFiveHundredPlayer(false, 1),
		domain.NewFiveHundredPlayer(false, 0),
		domain.NewFiveHundredPlayer(false, 1),
	}
	return domain.NewFiveHundred(domain.NewTrumpCardsFiveHundred(), players, domain.DefaultFiveHundredConfig())
}

func TestFiveHundredWebPresenter_Output(t *testing.T) {
	g := newFiveHundredGame()
	g.Reset()
	p := &presenter.FiveHundredWebPresenter{}
	out := p.Output(g, nil)

	var parsed controller.FiveHundredWebOutput
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Players) != 4 {
		t.Errorf("players = %d, want 4", len(parsed.Players))
	}
	if parsed.Config.TargetScore != domain.FiveHundredDefaultTargetScore {
		t.Errorf("targetScore = %d", parsed.Config.TargetScore)
	}
	// Human player's cards are revealed; CPU cards are hidden.
	if len(parsed.Players[0].Cards) == 0 {
		t.Errorf("human cards should be revealed")
	}
	if len(parsed.Players[1].Cards) != 0 {
		t.Errorf("CPU cards should be hidden")
	}
}

func TestFiveHundredWebPresenter_Error(t *testing.T) {
	g := newFiveHundredGame()
	g.Reset()
	p := &presenter.FiveHundredWebPresenter{}
	out := p.Output(g, errors.New("boom"))
	if !strings.Contains(out, "boom") {
		t.Errorf("error message not propagated: %s", out)
	}
}

func TestFiveHundredWebPresenter_OpenMisereRevealsDeclarer(t *testing.T) {
	g := newFiveHundredGame()
	g.SetContract(domain.FiveHundredContractOpenMisere, 0, -1)
	g.SetDeclarerIdx(1)
	g.GetPlayer(1).SetIsDeclarer(true)
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	g.SetPhase(domain.FiveHundredPhasePlay)

	p := &presenter.FiveHundredWebPresenter{}
	var parsed controller.FiveHundredWebOutput
	_ = json.Unmarshal([]byte(p.Output(g, nil)), &parsed)
	if len(parsed.Players[1].Cards) == 0 {
		t.Errorf("open misere declarer cards should be revealed")
	}
}

func TestFiveHundredWebPresenter_Hint(t *testing.T) {
	g := newFiveHundredGame()
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	g.SetPhase(domain.FiveHundredPhasePlay)
	g.SetCurrentPlayerIdx(0)

	p := &presenter.FiveHundredWebPresenter{}
	out := p.HintOutput(g)
	if !strings.Contains(out, "hint") {
		t.Errorf("expected hint in output: %s", out)
	}
}

func TestFiveHundredWebPresenter_ActionLog(t *testing.T) {
	g := newFiveHundredGame()
	g.Reset()
	p := &presenter.FiveHundredWebPresenter{}
	out := p.ActionLogOutput(g)
	if out == "" {
		t.Errorf("action log output is empty")
	}
}
