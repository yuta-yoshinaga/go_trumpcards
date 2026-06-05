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

func newBidWhistGame() *domain.BidWhist {
	players := []*domain.BidWhistPlayer{
		domain.NewBidWhistPlayer(true, 0),
		domain.NewBidWhistPlayer(false, 1),
		domain.NewBidWhistPlayer(false, 0),
		domain.NewBidWhistPlayer(false, 1),
	}
	return domain.NewBidWhist(domain.NewTrumpCards(2), players, domain.DefaultBidWhistConfig())
}

func TestBidWhistWebPresenter_Output(t *testing.T) {
	g := newBidWhistGame()
	g.Reset()
	p := &presenter.BidWhistWebPresenter{}
	out := p.Output(g, nil)

	var parsed controller.BidWhistWebOutput
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Players) != 4 {
		t.Errorf("players = %d, want 4", len(parsed.Players))
	}
	if parsed.Config.TargetScore != domain.BidWhistDefaultTargetScore {
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

func TestBidWhistWebPresenter_Error(t *testing.T) {
	g := newBidWhistGame()
	g.Reset()
	p := &presenter.BidWhistWebPresenter{}
	out := p.Output(g, errors.New("boom"))
	if !strings.Contains(out, "boom") {
		t.Errorf("error message not propagated: %s", out)
	}
}

func TestBidWhistWebPresenter_Hint(t *testing.T) {
	g := newBidWhistGame()
	g.SetContract(3, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	g.SetPhase(domain.BidWhistPhasePlay)
	g.SetCurrentPlayerIdx(0)

	p := &presenter.BidWhistWebPresenter{}
	out := p.HintOutput(g)
	if !strings.Contains(out, "hint") {
		t.Errorf("expected hint in output: %s", out)
	}
}

func TestBidWhistWebPresenter_ActionLog(t *testing.T) {
	g := newBidWhistGame()
	g.Reset()
	p := &presenter.BidWhistWebPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log output is empty")
	}
}

func TestBidWhistWebPresenter_PhaseMessages(t *testing.T) {
	p := &presenter.BidWhistWebPresenter{}
	for _, phase := range []domain.BidWhistPhase{
		domain.BidWhistPhaseTrumpDeclaration,
		domain.BidWhistPhaseKittyExchange,
		domain.BidWhistPhasePlay,
		domain.BidWhistPhaseTrickEnd,
		domain.BidWhistPhaseRoundEnd,
	} {
		g := newBidWhistGame()
		g.SetContract(3, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
		g.SetTrumpSuit(domain.CardDesignSpade)
		g.SetDeclarerIdx(0)
		g.SetPhase(phase)
		if phase == domain.BidWhistPhasePlay {
			g.SetCurrentTrick([]*domain.BidWhistTrickCard{
				{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
			})
		}
		if out := p.Output(g, nil); out == "" {
			t.Errorf("phase %d output empty", phase)
		}
	}

	// Game-end message branch.
	g := newBidWhistGame()
	g.SetTeamScore(0, 6)
	g.SetContract(1, domain.BidWhistDirectionUptown, domain.CardDesignSpade)
	g.SetDeclarerIdx(0)
	for i := 0; i < 12; i++ {
		g.GetPlayer(0).AddTrick([]*domain.Card{nil})
	}
	g.SetPhase(domain.BidWhistPhaseRoundEnd)
	g.ScoreRound()
	out := p.Output(g, nil)
	if !strings.Contains(out, "team0Win") && !strings.Contains(out, "ゲーム終了") {
		t.Errorf("game-end message missing: %s", out)
	}
}
