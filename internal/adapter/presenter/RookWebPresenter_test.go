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

func newRookGame() *domain.Rook {
	players := []*domain.RookPlayer{
		domain.NewRookPlayer(true, 0),
		domain.NewRookPlayer(false, 1),
		domain.NewRookPlayer(false, 0),
		domain.NewRookPlayer(false, 1),
	}
	return domain.NewRook(players, domain.DefaultRookConfig())
}

func TestRookWebPresenter_Output(t *testing.T) {
	g := newRookGame()
	g.Reset()
	p := &presenter.RookWebPresenter{}
	out := p.Output(g, nil)

	var parsed controller.RookWebOutput
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Players) != 4 {
		t.Errorf("players = %d, want 4", len(parsed.Players))
	}
	if parsed.Config.TargetScore != domain.RookDefaultTargetScore {
		t.Errorf("targetScore = %d", parsed.Config.TargetScore)
	}
	// Human cards revealed, CPU cards hidden.
	if len(parsed.Players[0].Cards) == 0 {
		t.Errorf("human cards should be revealed")
	}
	if len(parsed.Players[1].Cards) != 0 {
		t.Errorf("CPU cards should be hidden")
	}
}

// TestRookWebPresenter_ProceduralFaces verifies that Rook cards serialize with
// the procedural render descriptor (deck:"rook"), colour tokens for the four
// colours, and the "Rook" label for the bird.
func TestRookWebPresenter_ProceduralFaces(t *testing.T) {
	g := newRookGame()
	// Give the human a color card and the Rook bird, then reveal the trick too.
	g.GetPlayer(0).AddCard(domain.NewCard(1, 7, false)) // red 7
	g.SetTrumpColor(1)
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.RookPhasePlay)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(3, 9, false)},                                        // green 9
		{PlayerIdx: 2, Card: domain.NewCard(domain.RookBirdDesign, domain.RookBirdValue, false)}, // rook bird
	})

	p := &presenter.RookWebPresenter{}
	var parsed controller.RookWebOutput
	if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Human's red 7 card carries deck:"rook" + color "red".
	found := false
	for _, c := range parsed.Players[0].Cards {
		if c.Deck == "rook" && c.Color == "red" && c.Label == "7" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a red 7 card with deck:rook, got %+v", parsed.Players[0].Cards)
	}
	// Trick contains a green card and the Rook bird.
	var sawGreen, sawBird bool
	for _, tc := range parsed.CurrentTrick {
		if tc.Card.Deck == "rook" && tc.Card.Color == "green" {
			sawGreen = true
		}
		if tc.Card.Deck == "rook" && tc.Card.Label == "Rook" {
			sawBird = true
		}
	}
	if !sawGreen {
		t.Errorf("expected a green card in the trick")
	}
	if !sawBird {
		t.Errorf("expected the Rook bird (label Rook) in the trick")
	}
}

func TestRookWebPresenter_NestRevealedToHumanDeclarer(t *testing.T) {
	g := newRookGame()
	g.SetDeclarerIdx(0)
	g.GetPlayer(0).SetIsDeclarer(true)
	g.SetNest([]*domain.Card{domain.NewCard(2, 5, false), domain.NewCard(4, 10, false)})
	g.SetPhase(domain.RookPhaseNestExchange)

	p := &presenter.RookWebPresenter{}
	var parsed controller.RookWebOutput
	_ = json.Unmarshal([]byte(p.Output(g, nil)), &parsed)
	if len(parsed.Nest) != 2 {
		t.Errorf("nest should be revealed to human declarer, got %d", len(parsed.Nest))
	}
	if parsed.Nest[0].Deck != "rook" {
		t.Errorf("nest cards should carry deck:rook")
	}
}

func TestRookWebPresenter_Error(t *testing.T) {
	g := newRookGame()
	g.Reset()
	p := &presenter.RookWebPresenter{}
	out := p.Output(g, errors.New("boom"))
	if !strings.Contains(out, "boom") {
		t.Errorf("error message not propagated: %s", out)
	}
}

func TestRookWebPresenter_Hint(t *testing.T) {
	g := newRookGame()
	g.SetTrumpColor(1)
	g.SetDeclarerIdx(0)
	g.GetPlayer(0).AddCard(domain.NewCard(1, 1, false))
	g.SetPhase(domain.RookPhasePlay)
	g.SetCurrentPlayerIdx(0)

	p := &presenter.RookWebPresenter{}
	out := p.HintOutput(g)
	if !strings.Contains(out, "hint") {
		t.Errorf("expected hint in output: %s", out)
	}
}

func TestRookWebPresenter_ActionLog(t *testing.T) {
	g := newRookGame()
	g.Reset()
	p := &presenter.RookWebPresenter{}
	if out := p.ActionLogOutput(g); out == "" {
		t.Errorf("action log output is empty")
	}
}

func TestRookWebPresenter_PhaseMessages(t *testing.T) {
	p := &presenter.RookWebPresenter{}
	for _, phase := range []domain.RookPhase{
		domain.RookPhaseBid,
		domain.RookPhaseNestExchange,
		domain.RookPhasePlay,
		domain.RookPhaseTrickEnd,
		domain.RookPhaseRoundEnd,
	} {
		g := newRookGame()
		g.SetContractBid(80)
		g.SetDeclarerIdx(0)
		g.SetPhase(phase)
		if phase == domain.RookPhasePlay {
			g.SetCurrentTrick([]*domain.TrickCard{
				{PlayerIdx: 0, Card: domain.NewCard(1, 1, false)},
			})
		}
		if out := p.Output(g, nil); out == "" {
			t.Errorf("phase %d output empty", phase)
		}
	}

	// Game-end message branch.
	g := newRookGame()
	g.SetTeamScore(0, 450)
	g.SetDeclarerIdx(0)
	g.SetContractBid(70)
	g.GetPlayer(0).SetPoints(100)
	g.SetPhase(domain.RookPhaseRoundEnd)
	g.ScoreRound()
	out := p.Output(g, nil)
	if !strings.Contains(out, "team0Win") && !strings.Contains(out, "ゲーム終了") {
		t.Errorf("game-end message missing: %s", out)
	}
}
