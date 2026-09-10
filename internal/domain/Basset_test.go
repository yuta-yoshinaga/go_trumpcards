package domain

import "testing"

func bassetTestCards(values ...int) []*Card {
	cards := make([]*Card, 0, len(values))
	for i, value := range values {
		cards = append(cards, NewCard((i%4)+1, value, false))
	}
	return cards
}
func newBassetDeterministic(values ...int) *Basset {
	b := NewBasset(NewTrumpCards(0))
	b.SetDeckForTest(bassetTestCards(values...))
	return b
}
func placeBassetBet(t *testing.T, b *Basset, rank int) {
	t.Helper()
	if err := b.PlayerPlaceBet(rank, 10); err != nil {
		t.Fatalf("place bet: %v", err)
	}
}

func TestBasset_FirstCardMatchLosesAndSecondCardMatchWins(t *testing.T) {
	losing := newBassetDeterministic(7, 4, 3, 2)
	placeBassetBet(t, losing, 7)
	if err := losing.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	if losing.GetPhase() != BassetPhaseTurn {
		t.Fatalf("first-card match should lose bet, phase=%d", losing.GetPhase())
	}
	if bet, rank := losing.GetBet(); bet != nil || rank != 0 {
		t.Fatalf("first-card match retained bet: %#v/%d", bet, rank)
	}
	winning := newBassetDeterministic(4, 7, 3, 2)
	placeBassetBet(t, winning, 7)
	if err := winning.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	if winning.GetPhase() != BassetPhaseDecision {
		t.Fatalf("second-card match should open decision phase: %d", winning.GetPhase())
	}
	if bet, rank := winning.GetBet(); bet == nil || rank != 7 {
		t.Fatalf("winning bet not retained: %#v/%d", bet, rank)
	}
}

func TestBasset_ParoliAdvancesNamedMultipliers(t *testing.T) {
	b := newBassetDeterministic(4, 7, 4, 7, 4, 7, 4, 7, 4, 7)
	placeBassetBet(t, b, 7)
	for stage := range BassetPayoutMultipliers[:len(BassetPayoutMultipliers)-1] {
		if err := b.PlayerDealTurn(); err != nil {
			t.Fatalf("stage %d deal: %v", stage, err)
		}
		bet, _ := b.GetBet()
		if bet == nil || bet.Stage != stage {
			t.Fatalf("stage %d: bet=%v", stage, bet)
		}
		if stage < len(BassetPayoutMultipliers)-1 {
			if err := b.PlayerPressParoli(); err != nil {
				t.Fatalf("stage %d paroli: %v", stage, err)
			}
		}
	}
	if err := b.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	if err := b.PlayerTakeWinnings(); err != nil {
		t.Fatal(err)
	}
	if b.GetPhase() != BassetPhaseRoundEnd {
		t.Fatalf("final stage must cash out: phase=%d", b.GetPhase())
	}
	if bet, rank := b.GetBet(); bet != nil || rank != 0 {
		t.Fatalf("final stage retained bet: %#v/%d", bet, rank)
	}
}

func TestBasset_FinalStageCannotIncreaseFurther(t *testing.T) {
	b := newBassetDeterministic(4, 7, 4, 7, 4, 7, 4, 7, 4, 7)
	placeBassetBet(t, b, 7)
	for range BassetPayoutMultipliers[:len(BassetPayoutMultipliers)-1] {
		if err := b.PlayerDealTurn(); err != nil {
			t.Fatal(err)
		}
		if err := b.PlayerPressParoli(); err != nil {
			t.Fatal(err)
		}
	}
	if bet, _ := b.GetBet(); bet == nil || bet.Stage != len(BassetPayoutMultipliers)-1 {
		t.Fatalf("not at final stage: %#v", bet)
	}
}

func TestBasset_ParoliLossLosesAccumulatedWager(t *testing.T) {
	b := newBassetDeterministic(4, 7, 7, 4)
	placeBassetBet(t, b, 7)
	if err := b.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	if err := b.PlayerPressParoli(); err != nil {
		t.Fatal(err)
	}
	if err := b.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	if bet, rank := b.GetBet(); bet != nil || rank != 0 {
		t.Fatalf("paroli loss retained wager: %#v/%d", bet, rank)
	}
	if b.GetChips() != BassetDefaultStartChips-10 {
		t.Fatalf("paroli loss chips=%d", b.GetChips())
	}
}

func TestBasset_TakeWinningsResetsStage(t *testing.T) {
	b := newBassetDeterministic(4, 7, 3, 2)
	placeBassetBet(t, b, 7)
	if err := b.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	if err := b.PlayerTakeWinnings(); err != nil {
		t.Fatal(err)
	}
	if bet, rank := b.GetBet(); bet != nil || rank != 0 {
		t.Fatalf("cash-out retained bet: %#v/%d", bet, rank)
	}
	if b.GetChips() != BassetDefaultStartChips+10 {
		t.Fatalf("cash-out chips=%d", b.GetChips())
	}
	placeBassetBet(t, b, 7)
	if bet, _ := b.GetBet(); bet.Stage != 0 {
		t.Fatalf("new bet stage=%d, want 0", bet.Stage)
	}
}

func TestBasset_DeckExhaustionEndsRound(t *testing.T) {
	b := newBassetDeterministic(4, 7)
	placeBassetBet(t, b, 7)
	if err := b.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	if err := b.PlayerTakeWinnings(); err != nil {
		t.Fatal(err)
	}
	if b.GetPhase() != BassetPhaseRoundEnd {
		t.Fatalf("phase=%d, want round end", b.GetPhase())
	}
}
