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

func TestBidWhistWebPresenter_KittyIndices(t *testing.T) {
	kitty := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
	}
	orig := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	}

	parseOutput := func(t *testing.T, g *domain.BidWhist) controller.BidWhistWebOutput {
		t.Helper()
		p := &presenter.BidWhistWebPresenter{}
		var parsed controller.BidWhistWebOutput
		if err := json.Unmarshal([]byte(p.Output(g, nil)), &parsed); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return parsed
	}

	// Entitled: the human is the declarer during kitty exchange, so the six
	// kitty-origin cards in their merged hand are surfaced as indices.
	t.Run("human declarer during exchange -> populated", func(t *testing.T) {
		g := newBidWhistGame()
		human := g.GetPlayer(0)
		for _, c := range orig {
			human.AddCard(c)
		}
		for _, c := range kitty {
			human.AddCard(c) // kitty cards appended after the originals -> indices 2,3
		}
		g.SetDeclarerKitty(kitty)
		g.SetDeclarerIdx(0)
		g.SetPhase(domain.BidWhistPhaseKittyExchange)

		parsed := parseOutput(t, g)
		if got := parsed.KittyIndices; len(got) != 2 || got[0] != 2 || got[1] != 3 {
			t.Errorf("kittyIndices = %v, want [2 3]", got)
		}
	})

	// Gated (presenter): a CPU declarer must not leak its kitty to the human.
	t.Run("cpu declarer during exchange -> empty", func(t *testing.T) {
		g := newBidWhistGame()
		cpu := g.GetPlayer(1)
		for _, c := range kitty {
			cpu.AddCard(c)
		}
		g.SetDeclarerKitty(kitty)
		g.SetDeclarerIdx(1)
		g.SetPhase(domain.BidWhistPhaseKittyExchange)

		if got := parseOutput(t, g).KittyIndices; len(got) != 0 {
			t.Errorf("kittyIndices = %v, want empty for CPU declarer", got)
		}
	})

	// Gated (domain): outside the exchange phase there is nothing to highlight.
	t.Run("human declarer after exchange -> empty", func(t *testing.T) {
		g := newBidWhistGame()
		human := g.GetPlayer(0)
		for _, c := range kitty {
			human.AddCard(c)
		}
		g.SetDeclarerKitty(kitty)
		g.SetDeclarerIdx(0)
		g.SetPhase(domain.BidWhistPhasePlay)

		if got := parseOutput(t, g).KittyIndices; len(got) != 0 {
			t.Errorf("kittyIndices = %v, want empty outside exchange phase", got)
		}
	})
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
			g.SetCurrentTrick([]*domain.TrickCard{
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

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Reset 直後は入札フェーズに入っておらず、GetHint は 200 回中 200 回 nil を
// 返す。フェーズと入札手番を明示して初めてヒントが出る（この状態でも 200/200 で
// 確認済み）。
func TestBidWhistWebPresenterOutputCarriesTheHint(t *testing.T) {
	g := newBidWhistGame()
	g.Reset()
	g.SetPhase(domain.BidWhistPhaseBid)
	g.SetBidPlayerIdx(0)
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	var parsed controller.BidWhistWebOutput
	if err := json.Unmarshal([]byte(new(presenter.BidWhistWebPresenter).Output(g, nil)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Hint == nil {
		t.Error("Output must carry the hint -- the frontend reads state.hint")
	}
}
