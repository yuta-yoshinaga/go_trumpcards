package domain

import (
	"encoding/json"
	"testing"
)

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

func TestBasset_ConstructorsAndReset(t *testing.T) {
	if got := NewDefaultBasset().GetConfig(); got != DefaultBassetConfig() {
		t.Fatalf("default config = %+v, want %+v", got, DefaultBassetConfig())
	}
	custom := BassetConfig{StartChips: 250, MinBet: 5, MaxBet: 50}
	b := NewBassetWithConfig(nil, custom)
	if b.GetChips() != custom.StartChips || b.GetPhase() != BassetPhaseBetting {
		t.Fatalf("custom constructor state = chips %d phase %d", b.GetChips(), b.GetPhase())
	}
	invalid := NewBassetWithConfig(NewTrumpCards(0), BassetConfig{StartChips: 0, MinBet: 0, MaxBet: 0})
	if invalid.GetConfig() != DefaultBassetConfig() {
		t.Fatalf("invalid config was not replaced: %+v", invalid.GetConfig())
	}

	b.SetChips(0)
	b.SetPhase(BassetPhaseGameEnd)
	b.Reset()
	if b.GetChips() != custom.StartChips || b.GetPhase() != BassetPhaseBetting || b.GetGameEndFlag() {
		t.Fatalf("reset state = chips %d phase %d gameEnd %t", b.GetChips(), b.GetPhase(), b.GetGameEndFlag())
	}
	if b.GetRemainingCount() != BassetDeckSize {
		t.Fatalf("reset remaining = %d, want %d", b.GetRemainingCount(), BassetDeckSize)
	}
	b.SetChips(custom.StartChips)
	b.Reset()
	if b.GetChips() != custom.StartChips {
		t.Fatalf("reset changed sufficient bankroll to %d", b.GetChips())
	}
}

func TestBasset_PlaceBetValidationAndReplacement(t *testing.T) {
	b := newBassetDeterministic(4, 3)
	for _, tc := range []struct {
		name      string
		rank, amt int
	}{
		{"low rank", BassetMinRank - 1, 10},
		{"high rank", BassetMaxRank + 1, 10},
		{"below minimum", 7, 5},
		{"not multiple", 7, 15},
		{"over maximum", 7, BassetDefaultMaxBet + 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := b.PlayerPlaceBet(tc.rank, tc.amt); err == nil {
				t.Fatalf("PlayerPlaceBet(%d, %d) unexpectedly succeeded", tc.rank, tc.amt)
			}
		})
	}
	if err := b.PlayerPlaceBet(7, 100); err != nil {
		t.Fatal(err)
	}
	if err := b.PlayerPlaceBet(7, 40); err != nil {
		t.Fatal(err)
	}
	if b.GetChips() != BassetDefaultStartChips-40 {
		t.Fatalf("lowered bet chips = %d", b.GetChips())
	}
	if err := b.PlayerPlaceBet(7, 100); err != nil {
		t.Fatal(err)
	}
	if b.GetChips() != BassetDefaultStartChips-100 {
		t.Fatalf("raised bet chips = %d", b.GetChips())
	}
	b.SetChips(0)
	if err := b.PlayerPlaceBet(8, 10); err == nil {
		t.Fatal("insufficient chips unexpectedly succeeded")
	}
	b.SetPhase(BassetPhaseDecision)
	if err := b.PlayerPlaceBet(7, 10); err == nil {
		t.Fatal("bet in decision phase unexpectedly succeeded")
	}
}

func TestBasset_DealAndActionErrors(t *testing.T) {
	b := newBassetDeterministic(4, 3)
	if err := b.PlayerDealTurn(); err == nil {
		t.Fatal("deal without bet unexpectedly succeeded")
	}
	b.SetPhase(BassetPhaseGameEnd)
	if err := b.PlayerDealTurn(); err == nil {
		t.Fatal("deal in game-end phase unexpectedly succeeded")
	}
	b.SetPhase(BassetPhaseBetting)
	placeBassetBet(t, b, 7)
	b.SetDeckForTest(bassetTestCards(4))
	if err := b.PlayerDealTurn(); err == nil {
		t.Fatal("deal with exhausted deck unexpectedly succeeded")
	}
	if err := b.PlayerTakeWinnings(); err == nil {
		t.Fatal("take without decision unexpectedly succeeded")
	}
	if err := b.PlayerPressParoli(); err == nil {
		t.Fatal("paroli without decision unexpectedly succeeded")
	}
}

func TestBasset_GettersAndRemainingByRank(t *testing.T) {
	b := newBassetDeterministic(4, 7)
	cards := []*Card{
		nil,
		NewCard(CardDesignSpade, 0, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 1, true),
		NewCard(CardDesignDiamond, 14, false),
		NewCard(CardDesignSpade, 13, false),
	}
	b.trumpCards = NewTrumpCards(0)
	b.trumpCards.deck = cards
	b.trumpCards.deckCnt = len(cards)
	remaining := b.GetRemainingByRank()
	if remaining[1] != 1 || remaining[13] != 1 || remaining[0] != 0 {
		t.Fatalf("remaining by rank = %v", remaining)
	}
	if b.GetRemainingCount() != len(cards) || b.GetTurnsPlayed() != 0 || b.GetTurnsTotal() != BassetTurnsPerDeal {
		t.Fatalf("deck getters = remaining %d turns %d/%d", b.GetRemainingCount(), b.GetTurnsPlayed(), b.GetTurnsTotal())
	}
	if b.GetLastTurn() != nil || b.GetTotalPayout() != 0 || b.GetConfig() != DefaultBassetConfig() {
		t.Fatalf("initial getters returned unexpected state")
	}
	b.SetPhase(BassetPhaseTurn)
	b.SetChips(321)
	if b.GetPhase() != BassetPhaseTurn || b.GetChips() != 321 {
		t.Fatalf("setters did not update state")
	}
}

func TestBasset_NextRoundAndGameEnd(t *testing.T) {
	b := NewDefaultBasset()
	b.SetChips(BassetDefaultMinBet)
	b.SetPhase(BassetPhaseGameEnd)
	b.NextRound()
	if b.GetPhase() != BassetPhaseBetting || b.GetGameEndFlag() {
		t.Fatalf("next round state = phase %d gameEnd %t", b.GetPhase(), b.GetGameEndFlag())
	}
	b.SetChips(BassetDefaultMinBet - 1)
	b.NextRound()
	if !b.GetGameEndFlag() || b.GetPhase() != BassetPhaseGameEnd {
		t.Fatalf("empty bankroll state = phase %d gameEnd %t", b.GetPhase(), b.GetGameEndFlag())
	}
}

func TestBasset_JSONRoundTripAndHardening(t *testing.T) {
	b := newBassetDeterministic(4, 7)
	placeBassetBet(t, b, 7)
	if err := b.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var restored Basset
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.GetChips() != b.GetChips() || restored.GetPhase() != b.GetPhase() || restored.GetTurnsPlayed() != b.GetTurnsPlayed() {
		t.Fatalf("round-trip state differs: got chips %d phase %d turns %d", restored.GetChips(), restored.GetPhase(), restored.GetTurnsPlayed())
	}
	if restored.GetLastTurn() == nil {
		t.Fatal("round-trip lost turn or bet")
	}
	if bet, rank := restored.GetBet(); bet == nil || rank != 7 {
		t.Fatalf("round-trip lost bet: %#v rank %d", bet, rank)
	}
	for _, raw := range []string{
		"{invalid",
		`{"cf":{"sc":-1,"mn":10,"mx":100},"tc":{},"ch":{},"ps":1}`,
		`{"cf":{"sc":1000,"mn":10,"mx":100},"ps":1,"ch":{}}`,
		`{"cf":{"sc":1000,"mn":10,"mx":100},"tc":{},"ch":{},"ps":0}`,
		`{"cf":{"sc":1000,"mn":10,"mx":100},"tc":{},"ch":{},"ps":1,"br":14}`,
	} {
		var invalid Basset
		if err := json.Unmarshal([]byte(raw), &invalid); err == nil {
			t.Fatalf("invalid JSON state accepted: %s", raw)
		}
	}
}

func TestBasset_ParoliFinalStageCashOut(t *testing.T) {
	b := newBassetDeterministic(4, 7, 4, 7, 4, 7, 4, 7, 4, 7)
	placeBassetBet(t, b, 7)
	if err := b.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(BassetPayoutMultipliers)-1; i++ {
		if err := b.PlayerPressParoli(); err != nil {
			t.Fatal(err)
		}
		if i < len(BassetPayoutMultipliers)-2 {
			if err := b.PlayerDealTurn(); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := b.PlayerDealTurn(); err != nil {
		t.Fatal(err)
	}
	if err := b.PlayerPressParoli(); err != nil {
		t.Fatal(err)
	}
	if bet, rank := b.GetBet(); bet != nil || rank != 0 {
		t.Fatalf("final paroli retained bet: %#v rank %d", bet, rank)
	}
	if b.GetChips() != BassetDefaultStartChips+10*BassetPayoutMultipliers[len(BassetPayoutMultipliers)-1] {
		t.Fatalf("final paroli cash-out chips = %d", b.GetChips())
	}
}

func TestBassetConfig_ValidateAndJSON(t *testing.T) {
	if err := DefaultBassetConfig().Validate(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		cfg  BassetConfig
	}{
		{"start below minimum", BassetConfig{StartChips: 0, MinBet: 10, MaxBet: 100}},
		{"start above maximum", BassetConfig{StartChips: BassetChipsUpperBound + 1, MinBet: 10, MaxBet: 100}},
		{"minimum below minimum", BassetConfig{StartChips: 1000, MinBet: 0, MaxBet: 100}},
		{"minimum above maximum", BassetConfig{StartChips: 1000, MinBet: BassetChipsUpperBound + 1, MaxBet: BassetChipsUpperBound + 1}},
		{"maximum below minimum", BassetConfig{StartChips: 1000, MinBet: 200, MaxBet: 100}},
		{"maximum above maximum", BassetConfig{StartChips: 1000, MinBet: 10, MaxBet: BassetChipsUpperBound + 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err == nil {
				t.Fatalf("invalid config accepted: %+v", tc.cfg)
			}
		})
	}
	want := BassetConfig{StartChips: 250, MinBet: 5, MaxBet: 50}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got BassetConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("config round-trip = %+v, want %+v", got, want)
	}
	if err := json.Unmarshal([]byte("{invalid"), &got); err == nil {
		t.Fatal("invalid config JSON accepted")
	}
}
