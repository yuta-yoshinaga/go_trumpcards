package domain

import (
	"encoding/json"
	"testing"
)

func TestBourreScoreHandResultsIncludePenalty(t *testing.T) {
	players := make([]*BourrePlayer, BourrePlayerCnt)
	for i := range players {
		players[i] = NewBourrePlayer(false)
		players[i].SetChips(95)
		players[i].SetDecided(true)
		players[i].SetIsFinished(false)
	}
	players[1].AddTrick(nil)
	players[1].AddTrick(nil)
	players[2].SetFolded(true)
	players[3].SetFolded(true)
	players[4].SetIsFinished(true)

	b := NewBourre(NewTrumpCards(0), players, DefaultBourreConfig())
	b.pot = 30
	before := players[0].GetChips()
	b.scoreHand()

	results := b.GetLastResults()
	if len(results) != 5 {
		t.Fatalf("results len = %d, want 5", len(results))
	}
	byPlayer := make(map[int]*BourreHandResult, len(results))
	for _, result := range results {
		byPlayer[result.PlayerIdx] = result
	}
	penalty := before - players[0].GetChips()
	if got := byPlayer[0].PaidAmount; got != penalty || got == 0 {
		t.Errorf("bourreed PaidAmount = %d, chip reduction = %d; want equal nonzero values", got, penalty)
	}
	for _, idx := range []int{1, 2, 3, 4} {
		if got := byPlayer[idx].PaidAmount; got != 0 {
			t.Errorf("player %d PaidAmount = %d, want 0", idx, got)
		}
	}
	if got := byPlayer[1].WonAmount; got != 30 {
		t.Errorf("winner WonAmount = %d, want pot 30", got)
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var restored Bourre
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	for i, result := range b.GetLastResults() {
		if got := restored.GetLastResults()[i].PaidAmount; got != result.PaidAmount {
			t.Errorf("restored result %d PaidAmount = %d, want %d", i, got, result.PaidAmount)
		}
	}
}

func TestBourreNoContestResultsHaveNoPayments(t *testing.T) {
	players := make([]*BourrePlayer, BourrePlayerCnt)
	for i := range players {
		players[i] = NewBourrePlayer(false)
		players[i].SetDecided(true)
		players[i].SetFolded(true)
	}
	players[0].SetFolded(false)
	b := NewBourre(NewTrumpCards(0), players, DefaultBourreConfig())
	b.pot = 25
	b.resolveNoContest()
	for _, result := range b.GetLastResults() {
		if result.PaidAmount != 0 {
			t.Errorf("player %d PaidAmount = %d, want 0", result.PlayerIdx, result.PaidAmount)
		}
		if result.PlayerIdx == 0 && result.WonAmount != 25 {
			t.Errorf("sole winner WonAmount = %d, want 25", result.WonAmount)
		}
	}
}
