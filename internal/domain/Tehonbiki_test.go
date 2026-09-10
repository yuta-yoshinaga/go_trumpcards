//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func TestTehonbikiPayoutsAndLosses(t *testing.T) {
	tests := []struct {
		name    string
		kind    TehonbikiBetType
		numbers []int
		payout  int
	}{
		{"single", TehonbikiBetSingle, []int{4}, 225},
		{"double", TehonbikiBetDouble, []int{2, 4}, 90},
		{"triple", TehonbikiBetTriple, []int{1, 4, 6}, 45},
		{"half", TehonbikiBetHalf, []int{4, 5, 6}, 45},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDefaultTehonbiki()
			if err := g.SetParentCard(4); err != nil {
				t.Fatal(err)
			}
			if err := g.PlaceBet(tt.numbers, tt.kind, 50); err != nil {
				t.Fatal(err)
			}
			if g.GetResult() != TehonbikiResultWin || g.GetPayout() != tt.payout {
				t.Fatalf("result=%v payout=%d", g.GetResult(), g.GetPayout())
			}
		})
	}
	g := NewDefaultTehonbiki()
	if err := g.SetParentCard(6); err != nil {
		t.Fatal(err)
	}
	if err := g.PlaceBet([]int{1, 2}, TehonbikiBetDouble, 50); err != nil {
		t.Fatal(err)
	}
	if g.GetResult() != TehonbikiResultLose || g.GetPayout() != 0 || g.GetChips() != 950 {
		t.Fatalf("loss state: result=%v payout=%d chips=%d", g.GetResult(), g.GetPayout(), g.GetChips())
	}
}

func TestTehonbikiParentCardRange(t *testing.T) {
	g := NewDefaultTehonbiki()
	for _, n := range []int{0, 7} {
		if err := g.SetParentCard(n); err == nil {
			t.Fatalf("accepted parent card %d", n)
		}
	}
	for range 20 {
		g.Reset()
		if g.GetParentCard() < 1 || g.GetParentCard() > 6 {
			t.Fatalf("parent card=%d", g.GetParentCard())
		}
	}
}

func TestTehonbikiHalfGroupsAndTripleSelection(t *testing.T) {
	for _, numbers := range [][]int{{1, 2, 3}, {4, 5, 6}} {
		g := NewDefaultTehonbiki()
		if err := g.PlaceBet(numbers, TehonbikiBetHalf, 50); err != nil {
			t.Fatalf("half group %v rejected: %v", numbers, err)
		}
	}
	half := NewDefaultTehonbiki()
	if err := half.PlaceBet([]int{1, 2, 4}, TehonbikiBetHalf, 50); err == nil {
		t.Fatal("half accepted an arbitrary three-number selection")
	}
	triple := NewDefaultTehonbiki()
	if err := triple.PlaceBet([]int{1, 2, 4}, TehonbikiBetTriple, 50); err != nil {
		t.Fatalf("triple rejected an arbitrary three-number selection: %v", err)
	}
}

func TestTehonbikiRemainingCardsUsesDeck(t *testing.T) {
	g := NewDefaultTehonbiki()
	if got := g.GetRemainingCards(); got != len(g.deck) {
		t.Fatalf("remaining cards=%d, deck length=%d", got, len(g.deck))
	}
}

func TestTehonbikiPlaceBetRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		kind    TehonbikiBetType
		bet     int
	}{
		{"unknown kind", []int{1}, "unknown", 50},
		{"wrong count", []int{1}, TehonbikiBetDouble, 50},
		{"out of range", []int{0}, TehonbikiBetSingle, 50},
		{"duplicate", []int{1, 1}, TehonbikiBetDouble, 50},
		{"invalid half group", []int{1, 2, 4}, TehonbikiBetHalf, 50},
		{"bet below minimum", []int{1}, TehonbikiBetSingle, TehonbikiMinBet - 1},
		{"bet above maximum", []int{1}, TehonbikiBetSingle, TehonbikiMaxBet + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewTehonbiki(nil, NewTehonbikiPlayer(1000), DefaultTehonbikiConfig())
			if err := g.PlaceBet(tt.numbers, tt.kind, tt.bet); err == nil {
				t.Fatal("invalid bet was accepted")
			}
			if g.GetPhase() != TehonbikiPhaseBet || g.GetChips() != 1000 {
				t.Fatalf("state changed after rejection: phase=%v chips=%d", g.GetPhase(), g.GetChips())
			}
		})
	}
	g := NewDefaultTehonbikiWithPlayer(NewTehonbikiPlayer(10), DefaultTehonbikiConfig())
	if err := g.PlaceBet([]int{1}, TehonbikiBetSingle, 10); err != nil {
		t.Fatal(err)
	}
	if err := g.PlaceBet([]int{1}, TehonbikiBetSingle, 10); err == nil {
		t.Fatal("bet accepted after result phase")
	}
}

func TestTehonbikiPlaceBetInsufficientChips(t *testing.T) {
	g := NewDefaultTehonbikiWithPlayer(NewTehonbikiPlayer(10), DefaultTehonbikiConfig())
	if err := g.SetParentCard(6); err != nil {
		t.Fatal(err)
	}
	if err := g.PlaceBet([]int{1}, TehonbikiBetSingle, 10); err != nil {
		t.Fatal(err)
	}
	if err := g.NextRound(); err != nil {
		t.Fatal(err)
	}
	if g.GetGameEndFlag() != true || g.GetPhase() != TehonbikiPhaseGameEnd {
		t.Fatalf("game did not end: phase=%v end=%v", g.GetPhase(), g.GetGameEndFlag())
	}
}

func TestTehonbikiNextRoundAndGetters(t *testing.T) {
	g := NewTehonbiki(nil, NewTehonbikiPlayer(1000), DefaultTehonbikiConfig())
	if g.GetHint() == nil || g.GetConfig() != DefaultTehonbikiConfig() || g.GetRoundNumber() != 1 {
		t.Fatal("initial getters returned unexpected state")
	}
	if err := g.NextRound(); err == nil {
		t.Fatal("next round accepted during betting")
	}
	g.SetConfig(TehonbikiConfig{InitialChips: 500, DefaultBet: 20})
	g.SetChips(1000)
	if g.GetConfig().DefaultBet != 20 || g.GetChips() != 1000 || g.GetPlayer().GetChips() != 1000 {
		t.Fatal("config/player getters or setters failed")
	}
	if err := g.SetParentCard(2); err != nil {
		t.Fatal(err)
	}
	if err := g.PlaceBet([]int{2}, TehonbikiBetSingle, 50); err != nil {
		t.Fatal(err)
	}
	if err := g.NextRound(); err != nil || g.GetRoundNumber() != 2 || g.GetPhase() != TehonbikiPhaseBet {
		t.Fatalf("next round failed: err=%v round=%d phase=%v", err, g.GetRoundNumber(), g.GetPhase())
	}
	if len(g.GetNumbers()) != 0 || g.GetBet() != 0 || g.GetPayout() != 0 || g.GetResult() != TehonbikiResultNone || g.GetHint() == nil {
		t.Fatal("next round did not clear wager state")
	}
	g.Reset()
	if g.GetRoundNumber() != 1 || g.GetActionLog() != nil || g.GetGameEndFlag() || g.GetHint() == nil {
		t.Fatal("reset did not restore initial state")
	}
	if err := g.SetParentCard(1); err != nil {
		t.Fatal(err)
	}
	if err := g.PlaceBet([]int{1}, TehonbikiBetSingle, 50); err != nil {
		t.Fatal(err)
	}
	if g.GetHint() != nil || g.GetNumbers()[0] != 1 || g.GetBetType() != TehonbikiBetSingle || g.GetResult() != TehonbikiResultWin {
		t.Fatal("result getters returned unexpected state")
	}
}

func TestTehonbikiHalfGroupsAndPayoutTable(t *testing.T) {
	for _, numbers := range [][]int{{1, 2, 3}, {4, 5, 6}} {
		g := NewDefaultTehonbiki()
		if err := g.PlaceBet(numbers, TehonbikiBetHalf, 50); err != nil {
			t.Fatalf("half group %v rejected: %v", numbers, err)
		}
	}
	for _, numbers := range [][]int{{1, 2, 4}, {1, 3, 5}, {2, 4, 6}} {
		g := NewDefaultTehonbiki()
		if err := g.PlaceBet(numbers, TehonbikiBetHalf, 50); err == nil {
			t.Fatalf("invalid half group %v accepted", numbers)
		}
	}
	for kind, payout := range map[TehonbikiBetType]int{
		TehonbikiBetSingle: 225, TehonbikiBetDouble: 90, TehonbikiBetTriple: 45, TehonbikiBetHalf: 45,
	} {
		g := NewDefaultTehonbiki()
		_ = g.SetParentCard(6)
		numbers := map[TehonbikiBetType][]int{TehonbikiBetSingle: {6}, TehonbikiBetDouble: {1, 6}, TehonbikiBetTriple: {1, 2, 6}, TehonbikiBetHalf: {4, 5, 6}}[kind]
		if err := g.PlaceBet(numbers, kind, 50); err != nil || g.GetPayout() != payout || g.GetResult() != TehonbikiResultWin {
			t.Fatalf("%s: err=%v payout=%d result=%v", kind, err, g.GetPayout(), g.GetResult())
		}
	}
}

func TestTehonbikiPersistenceRestoresState(t *testing.T) {
	g := NewDefaultTehonbiki()
	_ = g.SetParentCard(4)
	if err := g.PlaceBet([]int{4}, TehonbikiBetSingle, 50); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	var restored Tehonbiki
	if err := json.Unmarshal(b, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.GetParentCard() != 4 || restored.GetPhase() != TehonbikiPhaseResult || restored.GetResult() != TehonbikiResultWin || restored.GetPayout() != 225 {
		t.Fatalf("restored state mismatch: parent=%d phase=%v result=%v payout=%d", restored.GetParentCard(), restored.GetPhase(), restored.GetResult(), restored.GetPayout())
	}
}

func TestTehonbikiConfigValidate(t *testing.T) {
	valid := DefaultTehonbikiConfig()
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, cfg := range []TehonbikiConfig{
		{InitialChips: TehonbikiMinChips - 1, DefaultBet: TehonbikiDefaultBet},
		{InitialChips: TehonbikiMaxChips + 1, DefaultBet: TehonbikiDefaultBet},
		{InitialChips: TehonbikiDefaultChips, DefaultBet: TehonbikiMinBet - 1},
		{InitialChips: TehonbikiDefaultChips, DefaultBet: TehonbikiMaxBet + 1},
	} {
		if err := cfg.Validate(); err == nil {
			t.Fatalf("invalid config accepted: %+v", cfg)
		}
	}
}
