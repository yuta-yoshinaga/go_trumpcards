package domain

import (
	"encoding/json"
	"testing"
)

// --- cariocaIsSet exhaustive branches ---

func TestCarioca_IsSet_Branches(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  bool
	}{
		{"too small", []*Card{cariocaCard(1, 5), cariocaCard(2, 5)}, false},
		{"plain set", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}, true},
		{"set of 4", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5), cariocaCard(4, 5)}, true},
		{"duplicate suit allowed", []*Card{cariocaCard(1, 5), cariocaCard(1, 5), cariocaCard(2, 5)}, true},
		{"one joker ok", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaJoker(1)}, true},
		{"two jokers rejected", []*Card{cariocaCard(1, 5), cariocaJoker(1), cariocaJoker(2)}, false},
		{"all jokers rejected", []*Card{cariocaJoker(1), cariocaJoker(2), cariocaJoker(3)}, false},
		{"mixed rank rejected", []*Card{cariocaCard(1, 5), cariocaCard(2, 6), cariocaCard(3, 5)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cariocaIsSet(tt.cards); got != tt.want {
				t.Errorf("cariocaIsSet(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// --- cariocaIsRun exhaustive branches ---

func TestCarioca_IsRun_Branches(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  bool
	}{
		{"too short", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7)}, false},
		{"ace-low run", []*Card{cariocaCard(1, 1), cariocaCard(1, 2), cariocaCard(1, 3), cariocaCard(1, 4)}, true},
		{"ace-high run", []*Card{cariocaCard(1, 11), cariocaCard(1, 12), cariocaCard(1, 13), cariocaCard(1, 1)}, true},
		{"middle run", []*Card{cariocaCard(2, 5), cariocaCard(2, 6), cariocaCard(2, 7), cariocaCard(2, 8)}, true},
		{"joker fills gap", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaJoker(1), cariocaCard(1, 8)}, true},
		{"joker extends end", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7), cariocaJoker(1)}, true},
		{"two jokers rejected", []*Card{cariocaCard(1, 5), cariocaJoker(1), cariocaJoker(2), cariocaCard(1, 8)}, false},
		{"mixed suit rejected", []*Card{cariocaCard(1, 5), cariocaCard(2, 6), cariocaCard(1, 7), cariocaCard(1, 8)}, false},
		{"duplicate value rejected", []*Card{cariocaCard(1, 5), cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7)}, false},
		{"span too wide rejected", []*Card{cariocaCard(1, 2), cariocaCard(1, 4), cariocaCard(1, 6), cariocaCard(1, 8)}, false},
		{"wraparound rejected", []*Card{cariocaCard(1, 12), cariocaCard(1, 13), cariocaCard(1, 1), cariocaCard(1, 2)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cariocaIsRun(tt.cards); got != tt.want {
				t.Errorf("cariocaIsRun(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// --- All 7 contracts: validate each slot success + failure ---

func TestCarioca_AllContracts_SlotValidation(t *testing.T) {
	goodSet := []*Card{cariocaCard(1, 9), cariocaCard(2, 9), cariocaCard(3, 9)}
	badSet := []*Card{cariocaCard(1, 9), cariocaCard(2, 8), cariocaCard(3, 7)}
	goodRun := []*Card{cariocaCard(1, 4), cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7)}
	badRun := []*Card{cariocaCard(1, 4), cariocaCard(2, 5), cariocaCard(1, 6), cariocaCard(1, 9)}

	for round := 1; round <= CariocaTotalRounds; round++ {
		contract := CariocaContractForRound(round)
		if len(contract.Slots) == 0 {
			t.Fatalf("round %d has empty contract", round)
		}
		for si, slot := range contract.Slots {
			var good, bad []*Card
			if slot.Kind == ContractSlotSet {
				good, bad = goodSet, badSet
			} else {
				good, bad = goodRun, badRun
			}
			if !cariocaValidateContractSlot(slot, good) {
				t.Errorf("round %d slot %d: good cards should validate", round, si)
			}
			if cariocaValidateContractSlot(slot, bad) {
				t.Errorf("round %d slot %d: bad cards should NOT validate", round, si)
			}
			// wrong size always fails
			if cariocaValidateContractSlot(slot, good[:1]) {
				t.Errorf("round %d slot %d: wrong size should fail", round, si)
			}
		}
	}
}

func TestCarioca_ValidateContractSlot_UnknownKind(t *testing.T) {
	slot := ContractSlot{Kind: ContractSlotKind(42), Size: 3}
	if cariocaValidateContractSlot(slot, []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}) {
		t.Error("unknown slot kind should not validate")
	}
}

// --- canAddToCariocaMeld exhaustive ---

func TestCarioca_CanAddToMeld_Extended(t *testing.T) {
	setMeld := []*Card{cariocaCard(1, 7), cariocaCard(2, 7), cariocaCard(3, 7)}
	setWithJoker := []*Card{cariocaCard(1, 7), cariocaCard(2, 7), cariocaJoker(1)}
	runMeld := []*Card{cariocaCard(1, 4), cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7)}
	runWithJoker := []*Card{cariocaCard(1, 4), cariocaJoker(1), cariocaCard(1, 6), cariocaCard(1, 7)}

	tests := []struct {
		name string
		meld []*Card
		card *Card
		want bool
	}{
		{"nil meld", nil, cariocaCard(1, 7), false},
		{"empty meld", []*Card{}, cariocaCard(1, 7), false},
		{"nil card", setMeld, nil, false},
		{"set matching rank", setMeld, cariocaCard(4, 7), true},
		{"set joker ok", setMeld, cariocaJoker(1), true},
		{"set wrong rank", setMeld, cariocaCard(4, 8), false},
		{"set already has joker + joker rejected", setWithJoker, cariocaJoker(2), false},
		{"set with joker + matching rank ok", setWithJoker, cariocaCard(4, 7), true},
		{"run low extend", runMeld, cariocaCard(1, 3), true},
		{"run high extend", runMeld, cariocaCard(1, 8), true},
		{"run wrong suit", runMeld, cariocaCard(2, 8), false},
		{"run gap", runMeld, cariocaCard(1, 10), false},
		{"run with joker + extend ok", runWithJoker, cariocaCard(1, 8), true},
		{"run with joker + second joker rejected", runWithJoker, cariocaJoker(2), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canAddToCariocaMeld(tt.meld, tt.card); got != tt.want {
				t.Errorf("canAddToCariocaMeld(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// --- penalty for every card kind ---

func TestCarioca_CardPenalty_AllKinds(t *testing.T) {
	cases := []struct {
		card *Card
		want int
	}{
		{cariocaJoker(1), CariocaJokerPenalty},
		{cariocaCard(1, 1), 15},
		{cariocaCard(1, 10), 10},
		{cariocaCard(1, 11), 10},
		{cariocaCard(1, 12), 10},
		{cariocaCard(1, 13), 10},
		{cariocaCard(1, 2), 2},
		{cariocaCard(1, 9), 9},
	}
	for _, c := range cases {
		if got := cariocaCardPenalty(c.card); got != c.want {
			t.Errorf("penalty(design=%d,val=%d) = %d, want %d", c.card.GetDesign(), c.card.GetValue(), got, c.want)
		}
	}
}

// --- contract meld going-out (empty hand) ---

func TestCarioca_ContractMeld_GoesOutWhenHandEmpties(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	p := g.GetPlayer(0)
	// Exactly the 6 contract cards → melding empties the hand → go out.
	cariocaSetHand(p, []*Card{
		cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5),
		cariocaCard(1, 9), cariocaCard(2, 9), cariocaCard(3, 9),
	})
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err != nil {
		t.Fatalf("meld error: %v", err)
	}
	if g.GetPhase() != CariocaPhaseRoundEnd {
		t.Errorf("expected round end after going out, got %d", g.GetPhase())
	}
	if g.GetRoundWinnerIdx() != 0 {
		t.Errorf("round winner = %d, want 0", g.GetRoundWinnerIdx())
	}
}

func TestCarioca_ContractMeld_RejectsWrongSlotSize(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	p := g.GetPlayer(0)
	cariocaSetHand(p, []*Card{
		cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5), cariocaCard(4, 5),
		cariocaCard(1, 9), cariocaCard(2, 9), cariocaCard(3, 9),
	})
	// First slot given 2 indices, but a set needs 3 → wrong size.
	if err := g.PlayerMeldContract([][]int{{0, 1}, {4, 5, 6}}); err == nil {
		t.Error("expected wrong-slot-size error")
	}
}

func TestCarioca_ContractMeld_RejectsOOBIndex(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	cariocaSetHand(g.GetPlayer(0), []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {50, 51, 52}}); err == nil {
		t.Error("expected out-of-range index error")
	}
}

func TestCarioca_ContractMeld_WrongPhaseAndGameEndAndNotHuman(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)

	g.SetPhase(CariocaPhaseDraw)
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err == nil {
		t.Error("expected wrong-phase error")
	}
	g.SetPhase(CariocaPhasePlay)
	g.SetCurrentPlayerIdx(1) // CPU
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err == nil {
		t.Error("expected not-human error")
	}
	g.SetCurrentPlayerIdx(0)
	g.gameEndFlag = true
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err == nil {
		t.Error("expected game-ended error")
	}
}

// --- extra meld error branches ---

func TestCarioca_MeldExtra_Branches(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	p := g.GetPlayer(0)
	p.SetContractMet(true)

	// too few cards
	cariocaSetHand(p, []*Card{cariocaCard(1, 5), cariocaCard(2, 5)})
	if err := g.PlayerMeldExtra([]int{0, 1}); err == nil {
		t.Error("expected too-few error")
	}
	// out-of-range index
	cariocaSetHand(p, []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	if err := g.PlayerMeldExtra([]int{0, 1, 99}); err == nil {
		t.Error("expected OOB index error")
	}
	// wrong phase
	g.SetPhase(CariocaPhaseDraw)
	if err := g.PlayerMeldExtra([]int{0, 1, 2}); err == nil {
		t.Error("expected wrong-phase error")
	}
	// game ended
	g.SetPhase(CariocaPhasePlay)
	g.gameEndFlag = true
	if err := g.PlayerMeldExtra([]int{0, 1, 2}); err == nil {
		t.Error("expected game-ended error")
	}
}

func TestCarioca_MeldExtra_GoesOut(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	p := g.GetPlayer(0)
	p.SetContractMet(true)
	cariocaSetHand(p, []*Card{cariocaCard(1, 9), cariocaCard(2, 9), cariocaCard(3, 9)})
	if err := g.PlayerMeldExtra([]int{0, 1, 2}); err != nil {
		t.Fatalf("meld extra error: %v", err)
	}
	if g.GetPhase() != CariocaPhaseRoundEnd {
		t.Errorf("expected round end after extra meld empties hand, got %d", g.GetPhase())
	}
}

// --- layoff error branches ---

func TestCarioca_Layoff_Branches(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	p0 := g.GetPlayer(0)
	p0.SetContractMet(true)
	cariocaSetHand(p0, []*Card{cariocaCard(4, 7), cariocaCard(1, 2)})
	// target player 1 met with a set meld
	g.GetPlayer(1).SetContractMet(true)
	g.GetPlayer(1).AppendMeld([]*Card{cariocaCard(1, 7), cariocaCard(2, 7), cariocaCard(3, 7)})

	// bad target index
	if err := g.PlayerLayoff(99, 0, 0); err == nil {
		t.Error("expected bad target index error")
	}
	// target not met (player 2)
	if err := g.PlayerLayoff(2, 0, 0); err == nil {
		t.Error("expected target-not-met error")
	}
	// bad meld index
	if err := g.PlayerLayoff(1, 99, 0); err == nil {
		t.Error("expected bad meld index error")
	}
	// bad card index
	if err := g.PlayerLayoff(1, 0, 99); err == nil {
		t.Error("expected bad card index error")
	}
	// card that cannot be added (index 1 = a 2, meld is 7s)
	if err := g.PlayerLayoff(1, 0, 1); err == nil {
		t.Error("expected cannot-add error")
	}
	// valid layoff (index 0 = 7)
	if err := g.PlayerLayoff(1, 0, 0); err != nil {
		t.Fatalf("valid layoff error: %v", err)
	}
	if len(g.GetPlayer(1).GetMeld(0)) != 4 {
		t.Errorf("meld should grow to 4")
	}
}

func TestCarioca_Layoff_WrongPhaseAndGameEnd(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhaseDraw)
	if err := g.PlayerLayoff(0, 0, 0); err == nil {
		t.Error("expected wrong-phase error")
	}
	g.SetPhase(CariocaPhasePlay)
	g.gameEndFlag = true
	if err := g.PlayerLayoff(0, 0, 0); err == nil {
		t.Error("expected game-ended error")
	}
}

func TestCarioca_Layoff_GoesOut(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	p0 := g.GetPlayer(0)
	p0.SetContractMet(true)
	cariocaSetHand(p0, []*Card{cariocaCard(4, 7)}) // single card
	g.GetPlayer(1).SetContractMet(true)
	g.GetPlayer(1).AppendMeld([]*Card{cariocaCard(1, 7), cariocaCard(2, 7), cariocaCard(3, 7)})
	if err := g.PlayerLayoff(1, 0, 0); err != nil {
		t.Fatalf("layoff error: %v", err)
	}
	if g.GetPhase() != CariocaPhaseRoundEnd {
		t.Errorf("expected round end after layoff empties hand, got %d", g.GetPhase())
	}
}

// --- discard branches ---

func TestCarioca_Discard_Branches(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)

	// out-of-range index
	if err := g.PlayerDiscard(99); err == nil {
		t.Error("expected OOB discard error")
	}
	// wrong phase
	g.SetPhase(CariocaPhaseDraw)
	if err := g.PlayerDiscard(0); err == nil {
		t.Error("expected wrong-phase discard error")
	}
	// not human
	g.SetPhase(CariocaPhasePlay)
	g.SetCurrentPlayerIdx(1)
	if err := g.PlayerDiscard(0); err == nil {
		t.Error("expected not-human discard error")
	}
	// game ended
	g.SetCurrentPlayerIdx(0)
	g.gameEndFlag = true
	if err := g.PlayerDiscard(0); err == nil {
		t.Error("expected game-ended discard error")
	}
}

// --- draw wrong-turn (CPU) rejection for discard-draw ---

func TestCarioca_DrawFromDiscard_NotHumanAndGameEnd(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(1)
	if err := g.PlayerDrawFromDiscard(); err == nil {
		t.Error("expected not-human error")
	}
	g.SetCurrentPlayerIdx(0)
	g.gameEndFlag = true
	if err := g.PlayerDrawFromDiscard(); err == nil {
		t.Error("expected game-ended error")
	}
}

// --- finishRound guard when already ended ---

func TestCarioca_FinishRound_GuardWhenAlreadyEnded(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetPhase(CariocaPhaseRoundEnd)
	g.roundWinnerIdx = 3
	g.finishRound(0) // guarded → no change
	if g.GetRoundWinnerIdx() != 3 {
		t.Errorf("finishRound should be a no-op when already at round end, got %d", g.GetRoundWinnerIdx())
	}
}

// --- finalizeGameEnd picks the minimum cumulative ---

func TestCarioca_FinalizeGameEnd_PicksMinimum(t *testing.T) {
	g := helperCariocaHand(t)
	scores := []int{30, 10, 50, 20}
	for i, s := range scores {
		g.GetPlayer(i).SetCumulativeScore(s)
	}
	g.finalizeGameEnd()
	if !g.GetGameEndFlag() {
		t.Error("game should be ended")
	}
	if g.GetPhase() != CariocaPhaseGameEnd {
		t.Errorf("phase = %d, want GameEnd", g.GetPhase())
	}
	if g.GetWinnerIdx() != 1 {
		t.Errorf("winner = %d, want 1 (lowest score)", g.GetWinnerIdx())
	}
}

// --- NextRound passes the deal to the player after the round winner ---

func TestCarioca_NextRound_StartingPlayerFollowsWinner(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	// finish round with winner 2 → phase RoundEnd
	for i := 0; i < g.GetPlayerCnt(); i++ {
		cariocaSetHand(g.GetPlayer(i), []*Card{cariocaCard(1, 5)})
	}
	g.GetPlayer(2).SetContractMet(true)
	g.finishRound(2)
	if g.GetPhase() != CariocaPhaseRoundEnd {
		t.Fatalf("expected round end, got %d", g.GetPhase())
	}
	g.NextRound()
	if g.GetCurrentPlayerIdx() != 3 {
		t.Errorf("next dealer = %d, want 3 (after winner 2)", g.GetCurrentPlayerIdx())
	}
	if g.GetRoundNumber() != 3 {
		t.Errorf("round = %d, want 3", g.GetRoundNumber())
	}
}

// --- multi-round cumulative accumulation ---

func TestCarioca_MultiRoundAccumulation(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)

	// Round 1: player 1 holds a 5 (penalty 5), everyone met (no fail penalty).
	for i := 0; i < g.GetPlayerCnt(); i++ {
		cariocaSetHand(g.GetPlayer(i), []*Card{cariocaCard(1, 5)})
		g.GetPlayer(i).SetContractMet(true)
	}
	g.finishRound(0)
	firstScore := g.GetPlayer(1).GetCumulativeScore()
	if firstScore != 5 {
		t.Fatalf("round1 cumulative = %d, want 5", firstScore)
	}

	// Advance and run a second round.
	g.NextRound()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		cariocaSetHand(g.GetPlayer(i), []*Card{cariocaCard(1, 13)}) // King = 10
		g.GetPlayer(i).SetContractMet(true)
	}
	g.finishRound(0)
	if got := g.GetPlayer(1).GetCumulativeScore(); got != firstScore+10 {
		t.Errorf("cumulative after round2 = %d, want %d", got, firstScore+10)
	}
}

// --- UnmarshalJSON validation branches ---

func cariocaUnmarshalErr(t *testing.T, doc map[string]any) error {
	t.Helper()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal doc: %v", err)
	}
	var g Carioca
	return json.Unmarshal(data, &g)
}

func TestCarioca_Unmarshal_ValidationBranches(t *testing.T) {
	fourPlayers := []map[string]any{{}, {}, {}, {}}

	t.Run("bad phase", func(t *testing.T) {
		if err := cariocaUnmarshalErr(t, map[string]any{"pl": fourPlayers, "ps": 99}); err == nil {
			t.Error("expected bad-phase error")
		}
	})
	t.Run("bad round", func(t *testing.T) {
		if err := cariocaUnmarshalErr(t, map[string]any{"pl": fourPlayers, "ps": 0, "rn": 99}); err == nil {
			t.Error("expected bad-round error")
		}
	})
	t.Run("OOB currentPlayerIdx", func(t *testing.T) {
		if err := cariocaUnmarshalErr(t, map[string]any{"pl": fourPlayers, "ps": 0, "rn": 1, "ci": 9}); err == nil {
			t.Error("expected OOB ci error")
		}
	})
	t.Run("OOB startingPlayer", func(t *testing.T) {
		if err := cariocaUnmarshalErr(t, map[string]any{"pl": fourPlayers, "ps": 0, "rn": 1, "ci": 0, "sp": 9}); err == nil {
			t.Error("expected OOB sp error")
		}
	})
	t.Run("OOB winnerIdx", func(t *testing.T) {
		if err := cariocaUnmarshalErr(t, map[string]any{"pl": fourPlayers, "ps": 0, "rn": 1, "ci": 0, "sp": 0, "wi": 9}); err == nil {
			t.Error("expected OOB wi error")
		}
	})
	t.Run("OOB roundWinnerIdx", func(t *testing.T) {
		if err := cariocaUnmarshalErr(t, map[string]any{"pl": fourPlayers, "ps": 0, "rn": 1, "ci": 0, "sp": 0, "wi": -1, "rw": 9}); err == nil {
			t.Error("expected OOB rw error")
		}
	})
	t.Run("sentinel winnerIdx -1 ok", func(t *testing.T) {
		if err := cariocaUnmarshalErr(t, map[string]any{"pl": fourPlayers, "ps": 0, "rn": 1, "ci": 0, "sp": 0, "wi": -1, "rw": -1}); err != nil {
			t.Errorf("sentinel -1 should be accepted: %v", err)
		}
	})
	t.Run("oversized players", func(t *testing.T) {
		big := make([]map[string]any, cariocaMaxSliceLen+1)
		for i := range big {
			big[i] = map[string]any{}
		}
		if err := cariocaUnmarshalErr(t, map[string]any{"pl": big}); err == nil {
			t.Error("expected oversized error")
		}
	})
	t.Run("bad json", func(t *testing.T) {
		var g Carioca
		if err := json.Unmarshal([]byte("not json"), &g); err == nil {
			t.Error("expected json parse error")
		}
	})
}

func TestCarioca_Unmarshal_DefaultsNilFields(t *testing.T) {
	// Minimal valid doc: no tc/dp/wp/al/cf → defaults filled in, no error.
	data, _ := json.Marshal(map[string]any{
		"pl": []map[string]any{{}, {}, {}, {}},
		"ps": 0, "rn": 1, "ci": 0, "sp": 0, "wi": -1, "rw": -1,
	})
	var g Carioca
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("minimal doc should unmarshal: %v", err)
	}
	if g.GetPlayerCnt() != 4 {
		t.Errorf("player count = %d, want 4", g.GetPlayerCnt())
	}
	// config PlayerCount was absent (0) → backfilled to len(players).
	if g.GetConfig().PlayerCount != 4 {
		t.Errorf("config playerCount = %d, want 4", g.GetConfig().PlayerCount)
	}
	// nil trumpCards defaulted → drawing/round machinery must not panic.
	if g.GetDrawPileCount() != 0 {
		t.Errorf("draw pile should default empty, got %d", g.GetDrawPileCount())
	}
}

func TestCarioca_Unmarshal_EmptyPlayersAllowsAnyIndex(t *testing.T) {
	// No players (n==0): index checks are skipped, sentinel branch returns nil.
	data, _ := json.Marshal(map[string]any{"pl": []map[string]any{}, "ps": 0, "rn": 0, "wi": 5, "rw": -1})
	var g Carioca
	if err := json.Unmarshal(data, &g); err != nil {
		t.Errorf("empty-player doc should unmarshal: %v", err)
	}
}

// --- full marshal → unmarshal round-trip preserving melds/scores ---

func TestCarioca_JSON_FullRoundTripState(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	g.SetRoundNumber(3)
	p := g.GetPlayer(0)
	p.SetContractMet(true)
	p.AppendMeld([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaJoker(1)})
	p.SetCumulativeScore(42)

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Carioca
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetRoundNumber() != 3 {
		t.Errorf("round = %d, want 3", restored.GetRoundNumber())
	}
	rp := restored.GetPlayer(0)
	if !rp.IsContractMet() {
		t.Error("contractMet should survive")
	}
	if rp.GetMeldCount() != 1 || len(rp.GetMeld(0)) != 3 {
		t.Errorf("meld should survive round-trip")
	}
	if rp.GetCumulativeScore() != 42 {
		t.Errorf("cumulative score = %d, want 42", rp.GetCumulativeScore())
	}
}

// --- CPU-driven full game (Easy difficulty) to hit AI/round/scoring paths ---

func TestCarioca_FullGameDriven_Easy(t *testing.T) {
	g := NewDefaultCarioca()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = CariocaCpuDifficultyEasy
	g.SetConfig(cfg)
	g.Reset()

	drawFromDiscardToggle := false
	for i := 0; i < 6000 && !g.GetGameEndFlag(); i++ {
		switch g.GetPhase() {
		case CariocaPhaseRoundEnd:
			g.NextRound()
		case CariocaPhaseDraw:
			if g.IsHumanTurn() {
				// Alternate the human's draw source to exercise both paths.
				if drawFromDiscardToggle {
					if err := g.PlayerDrawFromDiscard(); err != nil {
						_ = g.PlayerDrawFromStock()
					}
				} else {
					_ = g.PlayerDrawFromStock()
				}
				drawFromDiscardToggle = !drawFromDiscardToggle
			} else {
				g.CpuPlay()
			}
		case CariocaPhasePlay:
			if g.IsHumanTurn() {
				// Try to meet the contract; otherwise discard the first card.
				if !g.GetPlayer(0).IsContractMet() {
					if groups, ok := FindContractMeld(CariocaContractForRound(g.GetRoundNumber()), cariocaCollectCards(g.GetPlayer(0))); ok {
						handIdx := cariocaMapCardsToHandIndices(g.GetPlayer(0), groups)
						if handIdx != nil {
							_ = g.PlayerMeldContract(handIdx)
						}
					}
				}
				if g.GetPhase() == CariocaPhasePlay && g.IsHumanTurn() && g.GetPlayer(0).GetCardsSize() > 0 {
					_ = g.PlayerDiscard(0)
				}
			} else {
				g.CpuPlay()
			}
		default:
			// GameEnd — loop guard exits.
		}
	}
	// Coverage-driven: random Easy play is not guaranteed to complete a
	// contract within the budget (a turn draws 1 and discards 1, and an
	// exhausted stock recycles the discard), so assert a consistent state
	// rather than a hard game end. Round-end/scoring paths are covered
	// deterministically by the going-out and finalize tests above.
	if p := g.GetPhase(); p < CariocaPhaseDraw || p > CariocaPhaseGameEnd {
		t.Errorf("game left in invalid phase %d", p)
	}
	if g.GetGameEndFlag() && g.GetWinnerIdx() < 0 {
		t.Error("winner should be set at game end")
	}
}

func TestCarioca_FullGameDriven_Hard(t *testing.T) {
	g := NewDefaultCarioca()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = CariocaCpuDifficultyHard
	cfg.PlayerCount = 3
	g.SetConfig(cfg)
	g.Reset()
	if g.GetPlayerCnt() != 3 {
		t.Fatalf("expected 3 players, got %d", g.GetPlayerCnt())
	}
	for i := 0; i < 6000 && !g.GetGameEndFlag(); i++ {
		switch g.GetPhase() {
		case CariocaPhaseRoundEnd:
			g.NextRound()
		case CariocaPhaseDraw:
			if g.IsHumanTurn() {
				_ = g.PlayerDrawFromStock()
			} else {
				g.CpuPlay()
			}
		case CariocaPhasePlay:
			if g.IsHumanTurn() {
				_ = g.PlayerDiscard(0)
			} else {
				g.CpuPlay()
			}
		}
	}
	// Coverage-driven (see the Easy variant): convergence to a game end is not
	// guaranteed, so assert a consistent state instead.
	if p := g.GetPhase(); p < CariocaPhaseDraw || p > CariocaPhaseGameEnd {
		t.Errorf("game left in invalid phase %d", p)
	}
	if g.GetGameEndFlag() && g.GetWinnerIdx() < 0 {
		t.Error("winner should be set at game end")
	}
}

// --- player meld getters/setters coverage ---

func TestCarioca_Player_MeldGetters(t *testing.T) {
	p := NewCariocaPlayer(true)
	if p.GetMelds() != nil {
		t.Error("fresh player melds should be nil")
	}
	p.SetMelds([][]*Card{{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}})
	if p.GetMeldCount() != 1 {
		t.Errorf("meld count = %d, want 1", p.GetMeldCount())
	}
	if p.GetMeld(0) == nil {
		t.Error("meld 0 should exist")
	}
	if !p.AddCardToMeld(0, cariocaCard(4, 5)) {
		t.Error("AddCardToMeld should succeed for valid index")
	}
	if len(p.GetMeld(0)) != 4 {
		t.Errorf("meld should have grown to 4")
	}
	p.SetContractIndex([]int{0})
	if len(p.GetContractIndex()) != 1 {
		t.Error("contract index should be set")
	}
	p.ClearMelds()
	if p.GetMeldCount() != 0 || p.IsContractMet() || p.GetContractIndex() != nil {
		t.Error("ClearMelds should reset melds, contract-met and index")
	}
}
