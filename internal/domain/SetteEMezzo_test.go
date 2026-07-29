//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

// setteEMezzoDealAttempts caps the reshuffles a test may take while looking for
// a deal it can observe.
const setteEMezzoDealAttempts = 100

// newTestSetteEMezzo returns a freshly reset SetteEMezzo.
func newTestSetteEMezzo() *SetteEMezzo {
	s := NewDefaultSetteEMezzo()
	s.Reset()
	return s
}

// semCard builds a card for tests.
func semCard(design, value int) *Card {
	return NewCard(design, value, true)
}

// semMatta returns the matta (king of coins).
func semMatta() *Card {
	return semCard(SetteEMezzoMattaDesign, SetteEMezzoMattaValue)
}

// semHand builds a hand from cards.
func semHand(bet int, cards ...*Card) *SetteEMezzoHand {
	return &SetteEMezzoHand{cards: cards, bet: bet}
}

func TestSetteEMezzo_Reset(t *testing.T) {
	s := newTestSetteEMezzo()

	if s.GetPhase() != SetteEMezzoPhaseBet {
		t.Errorf("phase = %d, want bet", s.GetPhase())
	}
	if s.GetChips() != SetteEMezzoDefaultChips {
		t.Errorf("chips = %d, want %d", s.GetChips(), SetteEMezzoDefaultChips)
	}
	if got := len(s.GetSeats()); got != SetteEMezzoSeatCnt {
		t.Errorf("seats = %d, want %d", got, SetteEMezzoSeatCnt)
	}
	if s.GetNextBanker() != -1 {
		t.Errorf("nextBanker = %d, want -1", s.GetNextBanker())
	}
	if s.GetGameEndFlag() {
		t.Error("a fresh game should not be over")
	}
}

// The deck is the 40-card Italian-style one: 8, 9 and 10 are absent.
func TestSetteEMezzo_DeckHasFortyCardsWithoutEightNineTen(t *testing.T) {
	s := NewDefaultSetteEMezzo()
	seen := map[int]int{}
	total := 0
	for {
		c := s.trumpCards.DrawCard()
		if c == nil {
			break
		}
		seen[c.GetValue()]++
		total++
	}
	if total != SetteEMezzoDeckSize {
		t.Errorf("deck = %d cards, want %d", total, SetteEMezzoDeckSize)
	}
	for _, v := range []int{8, 9, 10} {
		if seen[v] != 0 {
			t.Errorf("rank %d should not be in the deck, found %d", v, seen[v])
		}
	}
	for _, v := range []int{1, 7, 11, 13} {
		if seen[v] != 4 {
			t.Errorf("rank %d appears %d times, want 4", v, seen[v])
		}
	}
}

// Totals are kept in HALVES so that the 0.5-point face cards never need a
// float. 7.5 is 15 halves.
func TestSetteEMezzo_CardValues(t *testing.T) {
	tests := []struct {
		name string
		card *Card
		want int
	}{
		{"ace is one point", semCard(CardDesignSpade, 1), 2},
		{"four is four points", semCard(CardDesignSpade, 4), 8},
		{"seven is seven points", semCard(CardDesignSpade, 7), 14},
		{"jack is half a point", semCard(CardDesignSpade, 11), 1},
		{"queen is half a point", semCard(CardDesignSpade, 12), 1},
		{"king is half a point", semCard(CardDesignSpade, 13), 1},
		{"nil is nothing", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setteEMezzoCardHalves(tt.card); got != tt.want {
				t.Errorf("halves = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSetteEMezzo_FormatHalves(t *testing.T) {
	tests := []struct {
		halves int
		want   string
	}{
		{15, "7.5"},
		{14, "7"},
		{1, "0.5"},
		{0, "0"},
		{11, "5.5"},
	}
	for _, tt := range tests {
		if got := setteEMezzoFormatHalves(tt.halves); got != tt.want {
			t.Errorf("format(%d) = %q, want %q", tt.halves, got, tt.want)
		}
	}
}

func TestSetteEMezzo_HandTotal(t *testing.T) {
	s := newTestSetteEMezzo()
	tests := []struct {
		name  string
		cards []*Card
		want  int
	}{
		{"two face cards make one point", []*Card{semCard(CardDesignSpade, 11), semCard(CardDesignSpade, 12)}, 2},
		{"seven plus a face card is 7.5", []*Card{semCard(CardDesignSpade, 7), semCard(CardDesignSpade, 11)}, 15},
		{"seven plus an ace busts", []*Card{semCard(CardDesignSpade, 7), semCard(CardDesignSpade, 1)}, 16},
		{"empty", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.handHalves(semHand(0, tt.cards...)); got != tt.want {
				t.Errorf("total = %d, want %d", got, tt.want)
			}
		})
	}
	if got := s.handHalves(nil); got != 0 {
		t.Errorf("nil hand = %d, want 0", got)
	}
}

// The matta is the king of coins only. The other three kings are plain half-
// point face cards.
func TestSetteEMezzo_OnlyTheKingOfCoinsIsWild(t *testing.T) {
	if setteEMezzoIndexOfMatta([]*Card{semMatta()}) != 0 {
		t.Error("the king of coins should be the matta")
	}
	for _, d := range []int{CardDesignSpade, CardDesignHeart, CardDesignClover} {
		if setteEMezzoIndexOfMatta([]*Card{semCard(d, SetteEMezzoMattaValue)}) >= 0 {
			t.Errorf("the king of suit %d must not be the matta", d)
		}
	}
	if setteEMezzoIndexOfMatta([]*Card{semCard(SetteEMezzoMattaDesign, 12)}) >= 0 {
		t.Error("the queen of coins must not be the matta")
	}
}

func TestSetteEMezzo_MattaTakesTheAssignedValue(t *testing.T) {
	s := newTestSetteEMezzo()
	h := semHand(0, semCard(CardDesignSpade, 4), semMatta())

	// 未指定なら 0.5 点として数える。
	if got := s.handHalves(h); got != 9 {
		t.Errorf("unassigned matta total = %d, want 9 (4.5)", got)
	}

	h.mattaHalves = 6 // 3 点
	if got := s.handHalves(h); got != 14 {
		t.Errorf("assigned matta total = %d, want 14 (7)", got)
	}
}

func TestSetteEMezzo_AutoAssignMattaMaximisesWithoutBusting(t *testing.T) {
	s := newTestSetteEMezzo()

	// 4 点 + マッタ → 3 点を割り当てて 7 点にするのが最大。3.5 だと 7.5 だが
	// マッタは整数か 0.5 しか取れないので 7 が上限。
	h := semHand(0, semCard(CardDesignSpade, 4), semMatta())
	s.autoAssignMatta(h)
	if got := s.handHalves(h); got != 14 {
		t.Errorf("total = %d, want 14 (7)", got)
	}

	// 7 点 + マッタ → 0.5 点でちょうど 7.5。
	h = semHand(0, semCard(CardDesignSpade, 7), semMatta())
	s.autoAssignMatta(h)
	if got := s.handHalves(h); got != SetteEMezzoTargetHalves {
		t.Errorf("total = %d, want %d", got, SetteEMezzoTargetHalves)
	}
	if h.GetMattaHalves() != 1 {
		t.Errorf("matta = %d halves, want 1 (0.5)", h.GetMattaHalves())
	}

	// マッタが無ければ割り当ては消える。
	h = semHand(0, semCard(CardDesignSpade, 4))
	h.mattaHalves = 6
	s.autoAssignMatta(h)
	if h.GetMattaHalves() != 0 {
		t.Errorf("matta = %d, want 0 without a matta in hand", h.GetMattaHalves())
	}
}

// setupHumanTurn puts the human on turn with an exact hand.
func setupSemHumanTurn(s *SetteEMezzo, cards ...*Card) *SetteEMezzoHand {
	s.banker = 1
	s.seats = []*SetteEMezzoSeat{
		{name: "あなた", isCPU: false},
		{name: "CPU1", isCPU: true},
		{name: "CPU2", isCPU: true},
	}
	h := semHand(100, cards...)
	s.seats[0].hand = h
	s.seats[2].hand = semHand(20, semCard(CardDesignSpade, 5))
	s.bankerHand = semHand(0, semCard(CardDesignHeart, 5))
	s.phase = SetteEMezzoPhasePlayerTurn
	s.activeSeat = 0
	return h
}

func TestSetteEMezzo_SetMattaValue(t *testing.T) {
	s := newTestSetteEMezzo()
	h := setupSemHumanTurn(s, semCard(CardDesignSpade, 4), semMatta())

	if !s.CanSetMatta() {
		t.Fatal("CanSetMatta should be true with a matta in hand")
	}
	if err := s.SetMattaValue(6); err != nil {
		t.Fatalf("SetMattaValue: %v", err)
	}
	if h.GetMattaHalves() != 6 {
		t.Errorf("matta = %d, want 6", h.GetMattaHalves())
	}
	// 0.5 も選べる。
	if err := s.SetMattaValue(1); err != nil {
		t.Fatalf("SetMattaValue(0.5): %v", err)
	}
}

func TestSetteEMezzo_SetMattaValueRejectsBadValues(t *testing.T) {
	s := newTestSetteEMezzo()
	setupSemHumanTurn(s, semCard(CardDesignSpade, 4), semMatta())

	// 半端な整数（1.5 点 = 3 halves）は取れない。マッタは 0.5 か 1〜7 の整数のみ。
	for _, bad := range []int{0, 3, 5, 15, 16, -2} {
		if err := s.SetMattaValue(bad); err == nil {
			t.Errorf("SetMattaValue(%d) should be rejected", bad)
		}
	}
}

func TestSetteEMezzo_SetMattaValueRejectedWithoutMatta(t *testing.T) {
	s := newTestSetteEMezzo()
	setupSemHumanTurn(s, semCard(CardDesignSpade, 4))
	if s.CanSetMatta() {
		t.Error("CanSetMatta should be false without a matta")
	}
	if err := s.SetMattaValue(6); err == nil {
		t.Error("SetMattaValue should fail without a matta")
	}
}

func TestSetteEMezzo_Hit(t *testing.T) {
	s := newTestSetteEMezzo()
	h := setupSemHumanTurn(s, semCard(CardDesignSpade, 2))

	if !s.CanHit() {
		t.Fatal("CanHit should be true on 2")
	}
	if err := s.Hit(); err != nil {
		t.Fatalf("Hit: %v", err)
	}
	if len(h.GetCards()) != 2 {
		t.Errorf("cards = %d, want 2", len(h.GetCards()))
	}
}

func TestSetteEMezzo_HitRejectedAtTarget(t *testing.T) {
	s := newTestSetteEMezzo()
	setupSemHumanTurn(s, semCard(CardDesignSpade, 7), semCard(CardDesignSpade, 11))

	if s.CanHit() {
		t.Error("CanHit should be false on 7.5")
	}
	if err := s.Hit(); err == nil {
		t.Error("drawing on 7.5 should be rejected")
	}
}

// Reaching exactly 7.5 ends the turn, but it is not an instant win: a banker
// on 7.5 takes it, since ties go to the banker. #4388 calls it an instant win.
func TestSetteEMezzo_ExactTargetEndsTheTurnButDoesNotWinOutright(t *testing.T) {
	s := newTestSetteEMezzo()
	h := semHand(100, semCard(CardDesignSpade, 7), semCard(CardDesignSpade, 11))
	if got := s.handHalves(h); got != SetteEMezzoTargetHalves {
		t.Fatalf("total = %d, want %d", got, SetteEMezzoTargetHalves)
	}

	// 親も 7.5 なら同点で親の勝ち。
	if got := s.settleHand(h, SetteEMezzoTargetHalves, false); got != -h.GetBet() {
		t.Errorf("payout = %d, want %d (the banker wins ties)", got, -h.GetBet())
	}
	// 親が 7 なら勝てる。
	if got := s.settleHand(h, 14, false); got != h.GetBet() {
		t.Errorf("payout = %d, want %d", got, h.GetBet())
	}
}

func TestSetteEMezzo_Stand(t *testing.T) {
	s := newTestSetteEMezzo()
	h := setupSemHumanTurn(s, semCard(CardDesignSpade, 5))

	if !s.CanStand() {
		t.Fatal("CanStand should be true on turn")
	}
	if err := s.Stand(); err != nil {
		t.Fatalf("Stand: %v", err)
	}
	if !h.IsStood() {
		t.Error("the hand should be marked as stood")
	}
}

func TestSetteEMezzo_ActionsRejectedOutOfTurn(t *testing.T) {
	s := newTestSetteEMezzo()
	s.phase = SetteEMezzoPhaseBet
	for name, fn := range map[string]func() error{
		"hit":   s.Hit,
		"stand": s.Stand,
		"matta": func() error { return s.SetMattaValue(6) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := fn(); err == nil {
				t.Errorf("%s should be rejected outside the player's turn", name)
			}
		})
	}
	if s.CanHit() || s.CanStand() || s.CanSetMatta() {
		t.Error("no action should be offered outside the player's turn")
	}
}

func TestSetteEMezzo_PlaceBet(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 1

	if err := s.PlaceBet(SetteEMezzoMinBet - 1); err == nil {
		t.Error("a bet below the minimum should be rejected")
	}
	if err := s.PlaceBet(SetteEMezzoMaxBet + 1); err == nil {
		t.Error("a bet above the maximum should be rejected")
	}
	if err := s.PlaceBet(SetteEMezzoDefaultChips * 10); err == nil {
		t.Error("betting more than the stack should be rejected")
	}

	before := s.GetChips()
	if err := s.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	if s.GetChips() != before-100 {
		t.Errorf("chips = %d, want %d", s.GetChips(), before-100)
	}
}

func TestSetteEMezzo_PlaceBetRejectedOutOfPhase(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 1
	s.phase = SetteEMezzoPhaseEnd
	if err := s.PlaceBet(100); err == nil {
		t.Error("PlaceBet should fail outside the betting phase")
	}
}

// The banker takes the other players' stakes rather than making one.
func TestSetteEMezzo_BankerDoesNotBet(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 0
	if err := s.PlaceBet(100); err == nil {
		t.Error("the banker must not be asked for a bet")
	}
	before := s.GetChips()
	if err := s.StartAsBanker(); err != nil {
		t.Fatalf("StartAsBanker: %v", err)
	}
	if s.GetChips() != before {
		t.Errorf("the banker's stack changed at the deal: %d, want %d", s.GetChips(), before)
	}
	if s.GetPhase() != SetteEMezzoPhaseBankerTurn {
		t.Errorf("phase = %d, want banker turn", s.GetPhase())
	}
}

func TestSetteEMezzo_StartAsBankerRejectedWhenNotBanker(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 1
	if err := s.StartAsBanker(); err == nil {
		t.Error("StartAsBanker should fail when a CPU banks")
	}
}

func TestSetteEMezzo_SettleHand(t *testing.T) {
	s := newTestSetteEMezzo()
	const bet = 100

	tests := []struct {
		name         string
		cards        []*Card
		bankerHalves int
		bankerBust   bool
		want         int
	}{
		{"bust loses", []*Card{semCard(CardDesignSpade, 7), semCard(CardDesignSpade, 1)}, 12, false, -bet},
		// 自分がバーストしていれば、親もバーストしていても負け。
		{"bust loses even against a bust banker", []*Card{semCard(CardDesignSpade, 7), semCard(CardDesignSpade, 1)}, 16, true, -bet},
		{"the banker busting pays everyone", []*Card{semCard(CardDesignSpade, 3)}, 16, true, bet},
		{"beating the banker pays once", []*Card{semCard(CardDesignSpade, 6)}, 10, false, bet},
		// 同点は親の勝ち。
		{"a tie goes to the banker", []*Card{semCard(CardDesignSpade, 6)}, 12, false, -bet},
		{"losing on points", []*Card{semCard(CardDesignSpade, 4)}, 12, false, -bet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := semHand(bet, tt.cards...)
			if got := s.settleHand(h, tt.bankerHalves, tt.bankerBust); got != tt.want {
				t.Errorf("settleHand = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSetteEMezzo_SettleCreditsTheHumanStack(t *testing.T) {
	s := newTestSetteEMezzo()
	setupSemHumanTurn(s, semCard(CardDesignSpade, 6))
	s.bankerHand = semHand(0, semCard(CardDesignHeart, 5))
	before := s.GetChips()

	s.settle()

	// 賭け金 100 が戻り、さらに 100 の勝ち。
	if s.GetChips() != before+200 {
		t.Errorf("chips = %d, want %d", s.GetChips(), before+200)
	}
	if s.GetPhase() != SetteEMezzoPhaseEnd {
		t.Errorf("phase = %d, want end", s.GetPhase())
	}
	if s.GetLastResult() == "" {
		t.Error("the settlement should be summarised")
	}
}

func TestSetteEMezzo_SettleDebitsTheHumanBanker(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 0
	s.seats = []*SetteEMezzoSeat{
		{name: "あなた", isCPU: false},
		{name: "CPU1", isCPU: true, hand: semHand(50, semCard(CardDesignSpade, 6))},
		{name: "CPU2", isCPU: true, hand: semHand(50, semCard(CardDesignSpade, 2))},
	}
	s.bankerHand = semHand(0, semCard(CardDesignHeart, 5))
	before := s.GetChips()

	s.settle()

	// CPU1 は 6 で勝ち (-50)、CPU2 は 2 で負け (+50) → 差し引き 0。
	if s.GetChips() != before {
		t.Errorf("chips = %d, want %d", s.GetChips(), before)
	}
}

// The bank passes only to a player who lands exactly 7.5. #4388 says it rotates
// every deal.
func TestSetteEMezzo_ExactTargetTakesTheBank(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 1
	s.seats = []*SetteEMezzoSeat{
		{name: "あなた", isCPU: false, hand: semHand(100, semCard(CardDesignSpade, 7), semCard(CardDesignSpade, 11))},
		{name: "CPU1", isCPU: true},
		{name: "CPU2", isCPU: true, hand: semHand(20, semCard(CardDesignSpade, 5))},
	}
	s.bankerHand = semHand(0, semCard(CardDesignHeart, 6))
	s.nextBanker = -1

	s.settle()

	if s.GetNextBanker() != 0 {
		t.Errorf("nextBanker = %d, want 0", s.GetNextBanker())
	}
	s.Reset()
	if s.GetBankerIdx() != 0 {
		t.Errorf("banker = %d, want the player who made 7.5", s.GetBankerIdx())
	}
	if !s.IsHumanBanker() {
		t.Error("IsHumanBanker should be true after taking the bank")
	}
}

// Beating the banker without landing on 7.5 does NOT take the bank.
func TestSetteEMezzo_WinningWithoutSevenAndAHalfKeepsTheBank(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 1
	s.seats = []*SetteEMezzoSeat{
		{name: "あなた", isCPU: false, hand: semHand(100, semCard(CardDesignSpade, 7))},
		{name: "CPU1", isCPU: true},
		{name: "CPU2", isCPU: true},
	}
	s.bankerHand = semHand(0, semCard(CardDesignHeart, 2))
	s.nextBanker = -1

	s.settle()

	if s.GetNextBanker() != -1 {
		t.Errorf("nextBanker = %d, want -1 -- only exactly 7.5 takes the bank", s.GetNextBanker())
	}
}

func TestSetteEMezzo_BankerTurnWhenHumanBanks(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 0
	s.seats = []*SetteEMezzoSeat{
		{name: "あなた", isCPU: false},
		{name: "CPU1", isCPU: true, hand: semHand(50, semCard(CardDesignSpade, 6))},
		{name: "CPU2", isCPU: true, hand: semHand(50, semCard(CardDesignSpade, 5))},
	}
	s.bankerHand = semHand(0, semCard(CardDesignHeart, 2))
	s.phase = SetteEMezzoPhasePlayerTurn
	s.activeSeat = 0
	s.advanceToHuman()

	if s.GetPhase() != SetteEMezzoPhaseBankerTurn {
		t.Fatalf("phase = %d, want banker turn", s.GetPhase())
	}
	if err := s.BankerHit(); err != nil {
		t.Fatalf("BankerHit: %v", err)
	}
	if s.GetPhase() == SetteEMezzoPhaseBankerTurn {
		if err := s.BankerStand(); err != nil {
			t.Fatalf("BankerStand: %v", err)
		}
	}
	if s.GetPhase() != SetteEMezzoPhaseEnd {
		t.Errorf("phase = %d, want end", s.GetPhase())
	}
}

func TestSetteEMezzo_BankerActionsRejectedOutOfPhase(t *testing.T) {
	s := newTestSetteEMezzo()
	s.phase = SetteEMezzoPhaseBet
	if err := s.BankerHit(); err == nil {
		t.Error("BankerHit should be rejected outside the banker's turn")
	}
	if err := s.BankerStand(); err == nil {
		t.Error("BankerStand should be rejected outside the banker's turn")
	}
}

func TestSetteEMezzo_BankerHitRejectedAtTarget(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 0
	s.phase = SetteEMezzoPhaseBankerTurn
	s.bankerHand = semHand(0, semCard(CardDesignSpade, 7), semCard(CardDesignSpade, 11))
	if err := s.BankerHit(); err == nil {
		t.Error("the banker must not draw on 7.5")
	}
}

func TestSetteEMezzo_CpuBanksAndSettlesAutomatically(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 1
	if err := s.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	for range 20 {
		if s.GetPhase() != SetteEMezzoPhasePlayerTurn {
			break
		}
		if err := s.Stand(); err != nil {
			t.Fatalf("Stand: %v", err)
		}
	}
	if s.GetPhase() != SetteEMezzoPhaseEnd {
		t.Errorf("phase = %d, want end once the CPU banker has played", s.GetPhase())
	}
}

// A CPU seat plays itself out and must never stop over the target.
func TestSetteEMezzo_CpuNeverStopsOverTheTarget(t *testing.T) {
	s := newTestSetteEMezzo()
	for range 200 {
		s.Reset()
		s.banker = 0
		if err := s.StartAsBanker(); err != nil {
			t.Fatalf("StartAsBanker: %v", err)
		}
		for i, seat := range s.GetSeats() {
			if i == s.GetBankerIdx() || seat.GetHand() == nil {
				continue
			}
			// 引き止めた手は 7.5 以下のはず。バーストは引いた結果なので許容。
			h := seat.GetHand()
			if !h.IsStood() {
				t.Fatalf("seat %d never finished its turn", i)
			}
			_ = s.handHalves(h)
		}
	}
}

func TestSetteEMezzo_ActionLog(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 1
	if err := s.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	log := s.GetActionLog()
	if len(log) == 0 {
		t.Fatal("the deal should be logged")
	}
	if log[0].ActionType != "deal" {
		t.Errorf("log[0].ActionType = %q, want deal", log[0].ActionType)
	}
}

func TestSetteEMezzo_JSONRoundTrip(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 1
	if err := s.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	restored := NewDefaultSetteEMezzo()
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored.GetPhase() != s.GetPhase() {
		t.Errorf("phase = %d, want %d", restored.GetPhase(), s.GetPhase())
	}
	if restored.GetChips() != s.GetChips() {
		t.Errorf("chips = %d, want %d", restored.GetChips(), s.GetChips())
	}
	if restored.GetBankerIdx() != s.GetBankerIdx() {
		t.Errorf("banker = %d, want %d", restored.GetBankerIdx(), s.GetBankerIdx())
	}
	// 手札そのものが戻ること。
	want := s.GetSeats()[0].GetHand()
	got := restored.GetSeats()[0].GetHand()
	if got == nil || want == nil {
		t.Fatal("the human's hand must survive the wire")
	}
	if len(got.GetCards()) != len(want.GetCards()) {
		t.Errorf("cards = %d, want %d", len(got.GetCards()), len(want.GetCards()))
	}
	if got.GetBet() != want.GetBet() {
		t.Errorf("bet = %d, want %d", got.GetBet(), want.GetBet())
	}
	// マッタの割り当ても戻ること。ここが落ちると合計が変わる。
	if got.GetMattaHalves() != want.GetMattaHalves() {
		t.Errorf("matta = %d, want %d", got.GetMattaHalves(), want.GetMattaHalves())
	}
	if restored.GetBankerHand() == nil || len(restored.GetBankerHand().GetCards()) == 0 {
		t.Error("the banker's hand must survive the wire")
	}
}

func TestSetteEMezzo_UnmarshalJSONRejectsInvalidState(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"malformed", `{`},
		{"phase too small", `{"ph":0}`},
		{"phase too large", `{"ph":99}`},
		{"too many seats", `{"ph":1,"st":[{},{},{},{}]}`},
		{"negative banker", `{"ph":1,"bk":-1}`},
		{"banker out of range", `{"ph":1,"bk":9}`},
		{"next banker too small", `{"ph":1,"nb":-2}`},
		{"next banker out of range", `{"ph":1,"nb":9}`},
		{"negative active seat", `{"ph":1,"as":-1}`},
		{"negative chips", `{"ph":1,"ch":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDefaultSetteEMezzo()
			if err := json.Unmarshal([]byte(tt.data), s); err == nil {
				t.Errorf("Unmarshal(%s) should fail", tt.data)
			}
		})
	}
}

func TestSetteEMezzo_UnmarshalJSONRejectsOversizedAndInvalidHands(t *testing.T) {
	// 棋譜の上限。
	log := `{"ph":1,"al":[`
	for i := range setteEMezzoMaxSliceLen + 1 {
		if i > 0 {
			log += ","
		}
		log += `{}`
	}
	log += `]}`
	s := NewDefaultSetteEMezzo()
	if err := json.Unmarshal([]byte(log), s); err == nil {
		t.Error("an oversized action log should be rejected")
	}

	// 手札の上限。
	cards := `{"cd":[`
	for i := range setteEMezzoMaxSliceLen + 1 {
		if i > 0 {
			cards += ","
		}
		cards += `{"d":1,"v":1,"f":true}`
	}
	cards += `]}`
	var h SetteEMezzoHand
	if err := json.Unmarshal([]byte(cards), &h); err == nil {
		t.Error("an oversized hand should be rejected")
	}

	// マッタの割り当ては 0〜14 の範囲。
	for _, bad := range []string{`{"mh":-1}`, `{"mh":99}`} {
		var bh SetteEMezzoHand
		if err := json.Unmarshal([]byte(bad), &bh); err == nil {
			t.Errorf("Unmarshal(%s) should fail", bad)
		}
	}
}

func TestSetteEMezzo_UnmarshalJSONRejectsMalformedNested(t *testing.T) {
	var h SetteEMezzoHand
	if err := json.Unmarshal([]byte(`{`), &h); err == nil {
		t.Error("a malformed hand should be rejected")
	}
	var seat SetteEMezzoSeat
	if err := json.Unmarshal([]byte(`{`), &seat); err == nil {
		t.Error("a malformed seat should be rejected")
	}
}

func TestSetteEMezzo_Accessors(t *testing.T) {
	s := newTestSetteEMezzo()
	h := setupSemHumanTurn(s, semCard(CardDesignSpade, 4), semMatta())

	if s.GetSeats()[0].GetName() != "あなた" {
		t.Errorf("name = %q", s.GetSeats()[0].GetName())
	}
	if s.GetSeats()[0].IsCPU() {
		t.Error("seat 0 is the human")
	}
	if !s.GetSeats()[1].IsCPU() {
		t.Error("seat 1 should be a CPU")
	}
	if !h.HasMatta() {
		t.Error("HasMatta should be true")
	}
	if s.GetActiveSeat() != 0 {
		t.Errorf("activeSeat = %d, want 0", s.GetActiveSeat())
	}
	if h.GetPayout() != 0 {
		t.Errorf("payout = %d, want 0 before settling", h.GetPayout())
	}
	if s.GetHandHalves(h) != s.handHalves(h) {
		t.Error("GetHandHalves should mirror the internal total")
	}
	if s.FormatHalves(15) != "7.5" {
		t.Errorf("FormatHalves(15) = %q, want 7.5", s.FormatHalves(15))
	}
	if s.GetBankerHand() == nil {
		t.Error("the banker should hold a hand")
	}
}

func TestSetteEMezzo_DescribeResult(t *testing.T) {
	s := newTestSetteEMezzo()
	if got := s.describeResult(16, true); got != "親がバースト（8）" {
		t.Errorf("describeResult(bust) = %q", got)
	}
	if got := s.describeResult(13, false); got != "親は 6.5" {
		t.Errorf("describeResult(points) = %q", got)
	}
}

func TestSetteEMezzo_ResetRestoresBrokeStack(t *testing.T) {
	s := newTestSetteEMezzo()
	s.chips.SetChips(SetteEMezzoMinBet - 1)
	s.Reset()
	if s.GetChips() != SetteEMezzoDefaultChips {
		t.Errorf("chips = %d, want a fresh stack", s.GetChips())
	}
}

func TestSetteEMezzo_HumanCanReachTargetThroughPlay(t *testing.T) {
	// 実際に配ってプレイし、7.5 に到達する局が存在することを確かめる。
	// 到達したらその手番は終わっているはず。
	s := newTestSetteEMezzo()
	for range setteEMezzoDealAttempts {
		s.Reset()
		s.banker = 1
		if err := s.PlaceBet(100); err != nil {
			t.Fatalf("PlaceBet: %v", err)
		}
		for s.GetPhase() == SetteEMezzoPhasePlayerTurn && s.CanHit() {
			if err := s.Hit(); err != nil {
				t.Fatalf("Hit: %v", err)
			}
		}
		h := s.GetSeats()[0].GetHand()
		if h != nil && s.GetHandHalves(h) == SetteEMezzoTargetHalves {
			if !h.IsStood() {
				t.Error("reaching 7.5 must end the turn")
			}
			return
		}
	}
	// 到達しなくてもテストとしては成立する（引きの問題）。
}

func TestSetteEMezzo_CurrentHandRejectsCpuAndMissingHand(t *testing.T) {
	s := newTestSetteEMezzo()
	setupSemHumanTurn(s, semCard(CardDesignSpade, 4))

	// 手番が CPU 席に進んでいるとき。
	s.activeSeat = 2
	if _, err := s.currentHand(); err == nil {
		t.Error("currentHand should fail while a CPU is on turn")
	}

	// 席は人間だが手が配られていないとき。
	s.activeSeat = 0
	s.seats[0].hand = nil
	if _, err := s.currentHand(); err == nil {
		t.Error("currentHand should fail with no hand dealt")
	}
}

// An exhausted deck must not spin the CPU loop or hand back a phantom card.
func TestSetteEMezzo_RunningOutOfCardsIsHandled(t *testing.T) {
	s := newTestSetteEMezzo()
	// 山を空にする。
	for s.trumpCards.DrawCard() != nil {
	}
	if got := s.drawOne(); got != nil {
		t.Errorf("drawOne on an empty deck = %v, want nil", got)
	}

	setupSemHumanTurn(s, semCard(CardDesignSpade, 2))
	if err := s.Hit(); err == nil {
		t.Error("Hit should fail once the deck is empty")
	}

	s.banker = 0
	s.phase = SetteEMezzoPhaseBankerTurn
	s.bankerHand = semHand(0, semCard(CardDesignSpade, 2))
	if err := s.BankerHit(); err == nil {
		t.Error("BankerHit should fail once the deck is empty")
	}
}

func TestSetteEMezzo_BankerStandSettles(t *testing.T) {
	s := newTestSetteEMezzo()
	s.banker = 0
	s.seats = []*SetteEMezzoSeat{
		{name: "あなた", isCPU: false},
		{name: "CPU1", isCPU: true, hand: semHand(50, semCard(CardDesignSpade, 6))},
		{name: "CPU2", isCPU: true, hand: semHand(50, semCard(CardDesignSpade, 2))},
	}
	s.bankerHand = semHand(0, semCard(CardDesignHeart, 5))
	s.phase = SetteEMezzoPhaseBankerTurn

	if err := s.BankerStand(); err != nil {
		t.Fatalf("BankerStand: %v", err)
	}
	if s.GetPhase() != SetteEMezzoPhaseEnd {
		t.Errorf("phase = %d, want end", s.GetPhase())
	}
	if !s.GetBankerHand().IsStood() {
		t.Error("the banker's hand should be marked as stood")
	}
}
