//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func quinzeTestCard(value int) *Card {
	return NewCard(CardDesignSpade, value, true)
}

func quinzeTestHand(bet int, values ...int) *QuinzeHand {
	cards := make([]*Card, 0, len(values))
	for _, value := range values {
		cards = append(cards, quinzeTestCard(value))
	}
	return &QuinzeHand{cards: cards, bet: bet}
}

func quinzeTestGame() *Quinze {
	game := NewDefaultQuinze()
	game.Reset()
	return game
}

func TestQuinzeResetUsesTheExpectedDefaults(t *testing.T) {
	game := quinzeTestGame()
	if game.GetPhase() != QuinzePhaseBet {
		t.Fatalf("phase = %d, want %d", game.GetPhase(), QuinzePhaseBet)
	}
	if game.GetChips() != QuinzeDefaultChips {
		t.Fatalf("chips = %d, want %d", game.GetChips(), QuinzeDefaultChips)
	}
	if len(game.GetSeats()) != QuinzeSeatCnt {
		t.Fatalf("seats = %d, want %d", len(game.GetSeats()), QuinzeSeatCnt)
	}
}

func TestQuinzeDeckHas52Cards(t *testing.T) {
	deck := NewTrumpCards(0)
	total := 0
	for deck.DrawCard() != nil {
		total++
	}
	if total != QuinzeDeckSize {
		t.Fatalf("deck = %d cards, want %d", total, QuinzeDeckSize)
	}
}

func TestQuinzeCardValues(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{name: "ace is one", value: 1, want: 1},
		{name: "number cards keep their value", value: 7, want: 7},
		{name: "ten is ten", value: 10, want: 10},
		{name: "jack is ten", value: 11, want: 10},
		{name: "queen is ten", value: 12, want: 10},
		{name: "king is ten", value: 13, want: 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := quinzeCardPoints(quinzeTestCard(test.value)); got != test.want {
				t.Fatalf("card value = %d, want %d", got, test.want)
			}
		})
	}
}

func TestQuinzeTargetAndBust(t *testing.T) {
	game := quinzeTestGame()
	exact := quinzeTestHand(10, 10, 5)
	if got := game.handPoints(exact); got != QuinzeTarget {
		t.Fatalf("exact total = %d, want %d", got, QuinzeTarget)
	}
	if got := game.settleHand(exact, 14, false); got != exact.bet {
		t.Fatalf("exact target payout = %d, want %d", got, exact.bet)
	}

	bust := quinzeTestHand(10, 10, 6)
	if got := game.handPoints(bust); got != 16 {
		t.Fatalf("bust total = %d, want 16", got)
	}
	if got := game.settleHand(bust, 14, false); got != -bust.bet {
		t.Fatalf("bust payout = %d, want %d", got, -bust.bet)
	}
}

func TestQuinzeTieGoesToTheBanker(t *testing.T) {
	game := quinzeTestGame()
	hand := quinzeTestHand(25, 10, 5)
	if got := game.settleHand(hand, QuinzeTarget, false); got != -hand.bet {
		t.Fatalf("tie payout = %d, want %d", got, -hand.bet)
	}
}

func TestQuinzeHumanCanStandWithAnExplicitHand(t *testing.T) {
	game := quinzeTestGame()
	game.banker = 1
	game.seats[0].hand = quinzeTestHand(100, 9)
	game.seats[2].hand = quinzeTestHand(20, 2)
	game.bankerHand = quinzeTestHand(0, 2)
	game.phase = QuinzePhasePlayerTurn
	game.activeSeat = 0

	if !game.CanStand() {
		t.Fatal("the human should be able to stand")
	}
	if err := game.Stand(); err != nil {
		t.Fatalf("Stand: %v", err)
	}
	if !game.seats[0].hand.stood {
		t.Fatal("the human hand should be stood")
	}
}

func TestQuinzePlaceBetDealsAndCpuBankerSettles(t *testing.T) {
	game := quinzeTestGame()
	game.banker = 1
	before := game.GetChips()
	if err := game.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	if game.GetChips() != before-100 {
		t.Fatalf("chips after bet = %d, want %d", game.GetChips(), before-100)
	}
	if game.GetPhase() != QuinzePhasePlayerTurn || game.GetActiveSeat() != 0 {
		t.Fatalf("after deal: phase=%d active=%d", game.GetPhase(), game.GetActiveSeat())
	}
	if game.GetSeats()[0].GetHand() == nil || game.GetSeats()[0].GetHand().GetBet() != 100 {
		t.Fatal("the human hand should contain the placed bet")
	}
	if game.GetSeats()[1].GetHand() != nil {
		t.Fatal("the banker seat should not contain a player hand")
	}
	if err := game.Stand(); err != nil {
		t.Fatalf("Stand: %v", err)
	}
	if game.GetPhase() != QuinzePhaseEnd {
		t.Fatalf("phase after CPU banker settles = %d, want end", game.GetPhase())
	}
	if !game.GetGameEndFlag() || game.GetLastResult() == "" || len(game.GetActionLog()) == 0 {
		t.Fatal("settlement should expose an ended game, result, and action log")
	}
}

func TestQuinzeHitCanHitAndStopsOnBurst(t *testing.T) {
	game := quinzeTestGame()
	game.seats[0].hand = quinzeTestHand(100, 10, 4)
	game.seats[1].hand = quinzeTestHand(20, 12)
	game.seats[2].hand = quinzeTestHand(20, 12)
	game.bankerHand = quinzeTestHand(0, 12)
	game.banker = 1
	game.phase = QuinzePhasePlayerTurn
	game.activeSeat = 0
	game.trumpCards.ReplenishForTest()
	if n := game.trumpCards.StackTopForTest(quinzeTestCard(2)); n != 1 {
		t.Fatalf("stacked cards = %d, want 1", n)
	}
	if !game.CanHit() || !game.CanStand() {
		t.Fatal("a human hand below the target should allow hit and stand")
	}
	if err := game.Hit(); err != nil {
		t.Fatalf("Hit: %v", err)
	}
	if !game.seats[0].hand.IsStood() {
		t.Fatal("a burst must stand the hand")
	}
	if game.GetPhase() == QuinzePhasePlayerTurn {
		t.Fatal("a burst must end the hand's player turn")
	}
	hitLogged := false
	for _, entry := range game.GetActionLog() {
		if entry.ActionType == "hit" {
			hitLogged = true
			break
		}
	}
	if !hitLogged {
		t.Fatal("the hit should be logged")
	}
}

func TestQuinzeHumanBankerFlowSettlesAndUsesBankerActions(t *testing.T) {
	game := quinzeTestGame()
	game.banker = 0
	game.seats[1].hand = quinzeTestHand(50, 6)
	game.seats[2].hand = quinzeTestHand(50, 5)
	game.trumpCards.ReplenishForTest()
	before := game.GetChips()
	if err := game.StartAsBanker(); err != nil {
		t.Fatalf("StartAsBanker: %v", err)
	}
	if game.GetPhase() != QuinzePhaseBankerTurn || game.GetChips() != before {
		t.Fatalf("banker start: phase=%d chips=%d", game.GetPhase(), game.GetChips())
	}
	if err := game.BankerHit(); err != nil {
		t.Fatalf("BankerHit: %v", err)
	}
	if game.GetPhase() == QuinzePhaseBankerTurn {
		if err := game.BankerStand(); err != nil {
			t.Fatalf("BankerStand: %v", err)
		}
	}
	if game.GetPhase() != QuinzePhaseEnd || !game.GetBankerHand().IsStood() {
		t.Fatal("banker action should settle the round")
	}
	if game.GetSeats()[1].GetHand().GetPayout() == 0 && game.GetSeats()[2].GetHand().GetPayout() == 0 {
		t.Fatal("at least one player result should be recorded")
	}
}

func TestQuinzeAccessorsExposeStateAndPredicatesRejectInvalidTurns(t *testing.T) {
	game := quinzeTestGame()
	hand := quinzeTestHand(42, 4)
	hand.stood = true
	hand.payout = 17
	game.seats[0].hand = hand
	game.bankerHand = quinzeTestHand(0, 10)
	game.activeSeat = 0
	game.phase = QuinzePhasePlayerTurn
	if game.GetHandPoints(hand) != 4 || game.FormatPoints(15) != "15" {
		t.Fatal("point accessors returned the wrong values")
	}
	seat := game.GetSeats()[0]
	if seat.GetName() != "あなた" || seat.IsCPU() || seat.GetHand() != hand {
		t.Fatal("human seat accessors returned the wrong values")
	}
	if hand.GetBet() != 42 || !hand.IsStood() || hand.GetPayout() != 17 || len(hand.GetCards()) != 1 {
		t.Fatal("hand accessors returned the wrong values")
	}
	if !game.CanHit() || !game.CanStand() || game.IsHumanBanker() || game.GetBankerIdx() != quinzeOpeningBanker {
		t.Fatal("turn and banker predicates returned the wrong values")
	}
	game.phase = QuinzePhaseEnd
	if game.CanHit() || game.CanStand() {
		t.Fatal("actions should be unavailable after settlement")
	}
	if game.GetBankerHand() == nil || game.GetActiveSeat() != 0 || game.GetNextBanker() != -1 || game.GetLastResult() != "" {
		t.Fatal("game accessors returned the wrong state")
	}
}

func TestQuinzeJSONRoundTripRestoresPlayableState(t *testing.T) {
	game := quinzeTestGame()
	game.banker = 1
	game.seats[0].hand = quinzeTestHand(75, 4)
	game.seats[1].hand = quinzeTestHand(20, 12)
	game.seats[2].hand = quinzeTestHand(20, 13)
	game.bankerHand = quinzeTestHand(0, 10)
	game.phase = QuinzePhasePlayerTurn
	game.activeSeat = 0
	game.nextBanker = 2
	game.lastResult = "保存済み"
	game.actionLog = []*ActionLogEntry{{TurnNumber: 0, PlayerIdx: 0, ActionType: "deal", Detail: "保存"}}
	data, err := json.Marshal(game)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	restored := NewDefaultQuinze()
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if restored.GetChips() != game.GetChips() || restored.GetBankerIdx() != 1 || restored.GetNextBanker() != 2 || restored.GetLastResult() != "保存済み" {
		t.Fatal("game state did not survive the round trip")
	}
	if restored.GetSeats()[0].GetHand().GetBet() != 75 || restored.GetHandPoints(restored.GetSeats()[0].GetHand()) != 4 {
		t.Fatal("the playable human hand did not survive the round trip")
	}
	restored.trumpCards.ReplenishForTest()
	if n := restored.trumpCards.StackTopForTest(quinzeTestCard(2)); n != 1 {
		t.Fatalf("stacked cards = %d, want 1", n)
	}
	if err := restored.Hit(); err != nil {
		t.Fatalf("Hit after restore: %v", err)
	}
	if restored.GetHandPoints(restored.GetSeats()[0].GetHand()) != 6 {
		t.Fatal("restored game should remain playable")
	}
}

func TestQuinzeJSONUnmarshalHandlesMissingFieldsAndRejectsInvalidData(t *testing.T) {
	var hand QuinzeHand
	if err := json.Unmarshal([]byte(`{"cd":[{"d":1,"v":4,"f":true}],"bt":9,"sd":true,"po":3}`), &hand); err != nil {
		t.Fatalf("hand Unmarshal: %v", err)
	}
	if hand.GetBet() != 9 || !hand.IsStood() || hand.GetPayout() != 3 || len(hand.GetCards()) != 1 {
		t.Fatal("hand JSON fields were not restored")
	}
	var seat QuinzeSeat
	if err := json.Unmarshal([]byte(`{"nm":"検証席","cp":true}`), &seat); err != nil {
		t.Fatalf("seat Unmarshal: %v", err)
	}
	if seat.GetName() != "検証席" || !seat.IsCPU() || seat.GetHand() != nil {
		t.Fatal("seat JSON fields were not restored")
	}
	for _, data := range []string{`{`, `{"ph":0}`, `{"ph":1,"bk":-1}`, `{"ph":1,"ch":-1}`, `{"ph":1,"as":-1}`} {
		if err := json.Unmarshal([]byte(data), NewDefaultQuinze()); err == nil {
			t.Errorf("Unmarshal(%s) should fail", data)
		}
	}
	oversizedHand := `{"cd":[` + strings.Repeat(`{},`, quinzeMaxSliceLen) + `{ }]}`
	if err := json.Unmarshal([]byte(oversizedHand), &hand); err == nil {
		t.Error("an oversized hand should be rejected")
	}
	oversizedLog := `{"ph":1,"al":[` + strings.Repeat(`{},`, quinzeMaxSliceLen) + `{}]}`
	if err := json.Unmarshal([]byte(oversizedLog), NewDefaultQuinze()); err == nil {
		t.Error("an oversized action log should be rejected")
	}
}

func TestQuinzeRejectsInvalidActionsAndResetsState(t *testing.T) {
	game := quinzeTestGame()
	game.chips.SetChips(QuinzeMinBet - 1)
	game.nextBanker = 2
	game.Reset()
	if game.GetChips() != QuinzeDefaultChips || game.GetBankerIdx() != 2 {
		t.Fatal("Reset should restore a broke stack and apply the next banker")
	}
	if err := game.PlaceBet(QuinzeMinBet - 1); err == nil {
		t.Error("a bet below the minimum should fail")
	}
	if err := game.PlaceBet(QuinzeMaxBet + 1); err == nil {
		t.Error("a bet above the maximum should fail")
	}
	if err := game.PlaceBet(QuinzeDefaultChips + 1); err == nil {
		t.Error("a bet above the stack should fail")
	}
	game.phase = QuinzePhaseEnd
	if err := game.PlaceBet(100); err == nil {
		t.Error("betting after settlement should fail")
	}
	if err := game.StartAsBanker(); err == nil {
		t.Error("a CPU banker should not allow human banker start")
	}
	game.banker = 0
	game.phase = QuinzePhaseBet
	if err := game.StartAsBanker(); err != nil {
		t.Fatalf("StartAsBanker: %v", err)
	}
	game.phase = QuinzePhasePlayerTurn
	game.activeSeat = 0
	game.seats[0].hand = quinzeTestHand(10, 10, 5)
	if game.CanHit() {
		t.Error("CanHit should be false at the target")
	}
	if err := game.Hit(); err == nil {
		t.Error("Hit at the target should fail")
	}
	game.phase = QuinzePhaseEnd
	if err := game.Stand(); err == nil {
		t.Error("Stand after settlement should fail")
	}
	if err := game.BankerHit(); err == nil {
		t.Error("BankerHit outside the banker phase should fail")
	}
	if err := game.BankerStand(); err == nil {
		t.Error("BankerStand outside the banker phase should fail")
	}
	for game.trumpCards.DrawCard() != nil {
	}
	if game.drawOne() != nil {
		t.Error("drawOne should return nil for an empty deck")
	}
}
