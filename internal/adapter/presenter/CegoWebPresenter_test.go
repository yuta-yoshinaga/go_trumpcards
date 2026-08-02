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
	g.SetCurrentTrick([]*domain.TrickCard{
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

// cegoDriveToGameEnd drives a single terminal deal (TargetDeals=1) where the
// human declarer (seat 0) captures exactly 53 points (a loss: -30 for the
// declarer, +10 for each opponent) starting from the supplied cumulative scores,
// then returns the finished game so the winner message can be asserted.
func cegoDriveToGameEnd(pre [domain.CegoPlayerCnt]int) *domain.Cego {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetConfig(domain.CegoConfig{CpuDifficulty: domain.CegoCpuDifficultyNormal, TargetDeals: 1})
	g.SetRoundNumber(1)
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	g.SetContractType(domain.CegoContractHandspiel)
	stash := make([]*domain.Card, 0, 11)
	for i := 0; i < 10; i++ {
		stash = append(stash, domain.NewCard(domain.CardDesignSpade, domain.CegoKingValue, false)) // 5 pts each
	}
	stash = append(stash, domain.NewCard(domain.CardDesignSpade, 6, false)) // Cavalier -> 3 pts
	g.SetStash(stash)
	g.SetStashOwner(0)
	g.SetPlayerScores(pre)
	g.SetPhase(domain.CegoPhaseRoundEnd)
	g.ScoreRound()
	return g
}

func TestCegoWebPresenter_WinnerMessages(t *testing.T) {
	p := &presenter.CegoWebPresenter{}
	cases := []struct {
		name string
		pre  [domain.CegoPlayerCnt]int
		want string
	}{
		{"human win", [domain.CegoPlayerCnt]int{100, 0, 0, 0}, "cego.result.humanWin"},
		{"cpu win", [domain.CegoPlayerCnt]int{0, 100, 0, 0}, "cego.result.cpuWin"},
		{"draw", [domain.CegoPlayerCnt]int{40, 0, 0, 0}, "cego.result.draw"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := cegoDriveToGameEnd(tc.pre)
			if !g.GetGameEndFlag() {
				t.Fatalf("expected game end")
			}
			out := p.Output(g, nil)
			if !strings.Contains(out, tc.want) {
				t.Errorf("message code %q not found in %s", tc.want, out)
			}
		})
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

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCegoWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ状態。300 回試して nil 0 件で確認済み。
	g := newCegoGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	out := new(presenter.CegoWebPresenter).Output(g, nil)
	if !strings.Contains(out, `"hint"`) {
		t.Error("Output must carry the hint -- the frontend reads state.hint")
	}
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	if strings.Contains(out, "cego.hintRequested") {
		t.Error("Output must not mark the response as a requested hint")
	}
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestCegoWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := newCegoGame()
	g.Reset()
	g.SetBidPlayerIdx(0)
	if !strings.Contains(new(presenter.CegoWebPresenter).HintOutput(g), "cego.hintRequested") {
		t.Error("HintOutput must mark the response as a requested hint")
	}
}
