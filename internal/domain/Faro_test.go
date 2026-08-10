package domain

import (
	"encoding/json"
	"testing"
)

// faroCard はテスト用のカード生成ヘルパー。
func faroCard(d, v int) *Card {
	return NewCard(d, v, false)
}

// newFaroForTest は決定的テスト用に、ベットフェーズで開始したファロを返す。
func newFaroForTest() *Faro {
	f := NewDefaultFaro()
	f.Reset()
	return f
}

func TestFaro_NewDefault(t *testing.T) {
	f := NewDefaultFaro()
	if f.GetChips() != FaroDefaultStartChips {
		t.Errorf("chips = %d, want %d", f.GetChips(), FaroDefaultStartChips)
	}
	if f.GetPhase() != FaroPhaseBetting {
		t.Errorf("phase = %d, want betting", f.GetPhase())
	}
}

func TestFaro_NewWithInvalidConfigFallsBack(t *testing.T) {
	f := NewFaroWithConfig(NewTrumpCards(0), FaroConfig{StartChips: -5, MinBet: 0, MaxBet: 0})
	if f.GetConfig().StartChips != FaroDefaultStartChips {
		t.Errorf("invalid config should fall back to default, got %+v", f.GetConfig())
	}
}

func TestFaroConfig_Validate(t *testing.T) {
	if err := DefaultFaroConfig().Validate(); err != nil {
		t.Errorf("default config invalid: %v", err)
	}
	cases := []FaroConfig{
		{StartChips: 0, MinBet: 10, MaxBet: 100},
		{StartChips: 1000, MinBet: 0, MaxBet: 100},
		{StartChips: 1000, MinBet: 200, MaxBet: 100}, // max < min
	}
	for i, c := range cases {
		if err := c.Validate(); err == nil {
			t.Errorf("case %d: expected validation error", i)
		}
	}
}

func TestFaro_PlaceBet_Validation(t *testing.T) {
	f := newFaroForTest()
	if err := f.PlayerPlaceBet(0, 100, false); err == nil {
		t.Error("rank 0 should be invalid")
	}
	if err := f.PlayerPlaceBet(14, 100, false); err == nil {
		t.Error("rank 14 should be invalid")
	}
	if err := f.PlayerPlaceBet(7, 5, false); err == nil {
		t.Error("amount below min should be invalid")
	}
	if err := f.PlayerPlaceBet(7, 15, false); err == nil {
		t.Error("amount not multiple of min should be invalid")
	}
	if err := f.PlayerPlaceBet(7, 100, false); err != nil {
		t.Errorf("valid bet rejected: %v", err)
	}
	if f.GetChips() != FaroDefaultStartChips-100 {
		t.Errorf("chips not deducted: %d", f.GetChips())
	}
}

func TestFaro_PlaceBet_OverwriteRefundsDelta(t *testing.T) {
	f := newFaroForTest()
	_ = f.PlayerPlaceBet(7, 100, false)
	// Lower the bet — should refund 60.
	_ = f.PlayerPlaceBet(7, 40, false)
	if f.GetChips() != FaroDefaultStartChips-40 {
		t.Errorf("chips after lowering = %d, want %d", f.GetChips(), FaroDefaultStartChips-40)
	}
	// Raise it — should deduct 60 more.
	_ = f.PlayerPlaceBet(7, 100, true)
	if f.GetChips() != FaroDefaultStartChips-100 {
		t.Errorf("chips after raising = %d, want %d", f.GetChips(), FaroDefaultStartChips-100)
	}
	if !f.GetBets()[7].Copper {
		t.Error("copper flag should be set after overwrite")
	}
}

func TestFaro_PlaceBet_InsufficientChips(t *testing.T) {
	f := newFaroForTest()
	f.SetChips(50)
	if err := f.PlayerPlaceBet(7, 100, false); err == nil {
		t.Error("expected insufficient chips error")
	}
}

func TestFaro_PlaceBet_WrongPhase(t *testing.T) {
	f := newFaroForTest()
	f.SetPhase(FaroPhaseCall)
	if err := f.PlayerPlaceBet(7, 100, false); err == nil {
		t.Error("bet in call phase should error")
	}
}

func TestFaro_ClearBetAndAll(t *testing.T) {
	f := newFaroForTest()
	_ = f.PlayerPlaceBet(3, 100, false)
	_ = f.PlayerPlaceBet(9, 50, false)
	if err := f.PlayerClearBet(3); err != nil {
		t.Fatalf("clear bet: %v", err)
	}
	if _, ok := f.GetBets()[3]; ok {
		t.Error("bet 3 should be cleared")
	}
	if f.GetChips() != FaroDefaultStartChips-50 {
		t.Errorf("chips after clear = %d", f.GetChips())
	}
	if err := f.PlayerClearAll(); err != nil {
		t.Fatalf("clear all: %v", err)
	}
	if len(f.GetBets()) != 0 {
		t.Error("all bets should be cleared")
	}
	if f.GetChips() != FaroDefaultStartChips {
		t.Errorf("chips fully refunded = %d", f.GetChips())
	}
}

func TestFaro_ClearBet_WrongPhase(t *testing.T) {
	f := newFaroForTest()
	f.SetPhase(FaroPhaseCall)
	if err := f.PlayerClearBet(3); err == nil {
		t.Error("clear in call phase should error")
	}
	if err := f.PlayerClearAll(); err == nil {
		t.Error("clear all in call phase should error")
	}
}

// dealOneTurn forces a deterministic turn by injecting the losing/winning cards
// at the front of the deck via SetBet+manual draw is not possible, so we test
// resolveTurn directly through PlayerDealTurn after stacking deck order is not
// feasible; instead we test resolution logic via the exported resolution path.

func TestFaro_ResolveTurn_PlainWinLose(t *testing.T) {
	f := newFaroForTest()
	// Plain bet on winning rank wins 1:1; plain bet on losing rank is collected.
	f.SetBet(5, 100, false)  // will be the winning rank
	f.SetBet(10, 100, false) // will be the losing rank
	start := f.GetChips()
	net := f.resolveTurn(10, 5, false)
	// Winning plain bet pays 2x (stake+win); losing plain bet collected (no payout).
	if f.GetChips() != start+200 {
		t.Errorf("chips = %d, want %d (win payout)", f.GetChips(), start+200)
	}
	if net != 0 { // +100 (win) -100 (loss)
		t.Errorf("net = %d, want 0", net)
	}
	if _, ok := f.GetBets()[5]; ok {
		t.Error("resolved winning bet should be removed")
	}
	if _, ok := f.GetBets()[10]; ok {
		t.Error("resolved losing bet should be removed")
	}
}

func TestFaro_ResolveTurn_CopperWinLose(t *testing.T) {
	f := newFaroForTest()
	// Copper on losing rank WINS; copper on winning rank LOSES.
	f.SetBet(10, 100, true) // losing rank → copper wins
	f.SetBet(5, 100, true)  // winning rank → copper loses
	start := f.GetChips()
	net := f.resolveTurn(10, 5, false)
	if f.GetChips() != start+200 { // copper-on-losing pays 2x
		t.Errorf("chips = %d, want %d", f.GetChips(), start+200)
	}
	if net != 0 { // +100 copper win, -100 copper loss
		t.Errorf("net = %d, want 0", net)
	}
}

func TestFaro_ResolveTurn_Split(t *testing.T) {
	f := newFaroForTest()
	f.SetBet(8, 100, false)
	start := f.GetChips()
	net := f.resolveTurn(8, 8, true)
	if net != -50 {
		t.Errorf("split net = %d, want -50", net)
	}
	if f.GetBets()[8].Amount != 50 {
		t.Errorf("remaining bet after split = %d, want 50", f.GetBets()[8].Amount)
	}
	if f.GetChips() != start { // chips untouched; only the staked bet shrinks
		t.Errorf("chips = %d, want %d", f.GetChips(), start)
	}
}

func TestFaro_ResolveTurn_SplitOddAmount(t *testing.T) {
	f := newFaroForTest()
	f.SetBet(8, 5, false) // odd: half = 2 (floor), remaining = 3
	net := f.resolveTurn(8, 8, true)
	if net != -2 {
		t.Errorf("odd split net = %d, want -2", net)
	}
	if f.GetBets()[8].Amount != 3 {
		t.Errorf("remaining = %d, want 3", f.GetBets()[8].Amount)
	}
}

func TestFaro_ResolveTurn_SplitConsumesWholeBet(t *testing.T) {
	f := newFaroForTest()
	f.SetBet(8, 0, false) // edge: amount 0 → half 0 → removed
	f.resolveTurn(8, 8, true)
	if _, ok := f.GetBets()[8]; ok {
		// amount stays 0 → b.Amount == 0 → deleted
		t.Error("zero-amount bet should be removed after split")
	}
}

func TestFaro_DealTurn_WrongPhase(t *testing.T) {
	f := newFaroForTest()
	f.SetPhase(FaroPhaseCall)
	if err := f.PlayerDealTurn(); err == nil {
		t.Error("deal in call phase should error")
	}
}

func TestFaro_Call_WrongPhase(t *testing.T) {
	f := newFaroForTest()
	if err := f.PlayerCall([]int{1, 2, 3}); err == nil {
		t.Error("call in betting phase should error")
	}
}

func TestFaro_Call_InvalidOrder(t *testing.T) {
	f := newFaroForTest()
	f.SetPhase(FaroPhaseCall)
	f.callCards = []*Card{faroCard(CardDesignSpade, 3), faroCard(CardDesignHeart, 9), faroCard(CardDesignClover, 12)}
	if err := f.PlayerCall([]int{3, 9}); err == nil {
		t.Error("wrong-length call should error")
	}
	if err := f.PlayerCall([]int{3, 9, 14}); err == nil {
		t.Error("out-of-range rank should error")
	}
}

func TestFaro_Call_CorrectAndWrong(t *testing.T) {
	// Correct call.
	f := newFaroForTest()
	f.SetPhase(FaroPhaseCall)
	f.SetBet(3, 100, false)
	f.callCards = []*Card{faroCard(CardDesignSpade, 3), faroCard(CardDesignHeart, 9), faroCard(CardDesignClover, 12)}
	start := f.GetChips()
	if err := f.PlayerCall([]int{3, 9, 12}); err != nil {
		t.Fatalf("call: %v", err)
	}
	if !f.GetCallWon() {
		t.Error("correct call should win")
	}
	// stake=100, payout=100+400=500; chips: start-100+500 = start+400
	if f.GetChips() != start+400 {
		t.Errorf("chips after winning call = %d, want %d", f.GetChips(), start+400)
	}
	if f.GetPhase() != FaroPhaseRoundEnd {
		t.Errorf("phase after call = %d, want round end", f.GetPhase())
	}

	// Wrong call.
	f2 := newFaroForTest()
	f2.SetPhase(FaroPhaseCall)
	f2.SetBet(3, 100, false)
	f2.callCards = []*Card{faroCard(CardDesignSpade, 3), faroCard(CardDesignHeart, 9), faroCard(CardDesignClover, 12)}
	start2 := f2.GetChips()
	if err := f2.PlayerCall([]int{9, 3, 12}); err != nil {
		t.Fatalf("call: %v", err)
	}
	if f2.GetCallWon() {
		t.Error("wrong call should lose")
	}
	if f2.GetChips() != start2-100 {
		t.Errorf("chips after losing call = %d, want %d", f2.GetChips(), start2-100)
	}
}

func TestFaro_Call_Skip(t *testing.T) {
	f := newFaroForTest()
	f.SetPhase(FaroPhaseCall)
	f.callCards = []*Card{faroCard(CardDesignSpade, 3), faroCard(CardDesignHeart, 9), faroCard(CardDesignClover, 12)}
	if err := f.PlayerCall(nil); err != nil {
		t.Fatalf("skip call: %v", err)
	}
	if f.GetCallOrder() != nil {
		t.Error("skipped call should leave order nil")
	}
	if f.GetPhase() != FaroPhaseRoundEnd {
		t.Error("skip should end the round")
	}
}

func TestFaro_Call_InsufficientChips(t *testing.T) {
	f := newFaroForTest()
	f.SetPhase(FaroPhaseCall)
	f.SetChips(5)
	f.callCards = []*Card{faroCard(CardDesignSpade, 3), faroCard(CardDesignHeart, 9), faroCard(CardDesignClover, 12)}
	if err := f.PlayerCall([]int{3, 9, 12}); err == nil {
		t.Error("call without chips should error")
	}
}

// TestFaro_FullDealAutoPlay places bets and deals turns to exhaustion, then calls.
// It MUST terminate (finite deck) and exercise a split + the call.
func TestFaro_FullDealAutoPlay(t *testing.T) {
	f := newFaroForTest()
	// Spread bets across all ranks so every turn resolves something and a split
	// is overwhelmingly likely across 24 turns. Bets persist across turns, so we
	// re-place any cleared rank each loop.
	guard := 0
	sawSplit := false
	for f.GetPhase() == FaroPhaseBetting || f.GetPhase() == FaroPhaseTurn {
		guard++
		if guard > 1000 {
			t.Fatal("auto-play did not terminate")
		}
		// Re-place a bet on every rank that has chips available.
		for r := FaroMinRank; r <= FaroMaxRank; r++ {
			if f.GetChips() < f.GetConfig().MinBet {
				break
			}
			if _, ok := f.GetBets()[r]; !ok {
				_ = f.PlayerPlaceBet(r, f.GetConfig().MinBet, r%2 == 0)
			}
		}
		if err := f.PlayerDealTurn(); err != nil {
			t.Fatalf("deal turn error: %v", err)
		}
		if lt := f.GetLastTurn(); lt != nil && lt.Split {
			sawSplit = true
		}
	}
	if f.GetTurnsPlayed() != FaroTurnsPerDeal {
		t.Errorf("turns played = %d, want %d", f.GetTurnsPlayed(), FaroTurnsPerDeal)
	}
	if f.GetPhase() != FaroPhaseCall {
		t.Fatalf("phase after all turns = %d, want call", f.GetPhase())
	}
	if len(f.GetCallCards()) != FaroCallCards {
		t.Errorf("call cards = %d, want %d", len(f.GetCallCards()), FaroCallCards)
	}
	// 52-card accounting: 1 soda + 24*2 turns + 3 call = 52, deck empty.
	if f.GetRemainingCount() != 0 {
		t.Errorf("remaining after call cards drawn = %d, want 0", f.GetRemainingCount())
	}
	if !sawSplit {
		t.Log("note: no split occurred in this shuffle (acceptable, branch covered elsewhere)")
	}
	// Make the call using the actual order (correct call branch).
	order := make([]int, len(f.GetCallCards()))
	for i, c := range f.GetCallCards() {
		order[i] = c.GetValue()
	}
	if err := f.PlayerCall(order); err != nil {
		t.Fatalf("final call: %v", err)
	}
	if !f.GetCallWon() {
		t.Error("calling the actual order should win")
	}
	if f.GetPhase() != FaroPhaseRoundEnd {
		t.Error("round should end after call")
	}
}

func TestFaro_NextRoundAndGameEnd(t *testing.T) {
	f := newFaroForTest()
	f.SetChips(500)
	f.NextRound()
	if f.GetPhase() != FaroPhaseBetting {
		t.Errorf("phase after next round = %d, want betting", f.GetPhase())
	}
	// Out of chips → game end.
	f.SetChips(0)
	f.NextRound()
	if !f.GetGameEndFlag() || f.GetPhase() != FaroPhaseGameEnd {
		t.Error("zero chips should end the game")
	}
}

func TestFaro_Reset_RefillsChips(t *testing.T) {
	f := NewDefaultFaro()
	f.SetChips(0)
	f.Reset()
	if f.GetChips() != FaroDefaultStartChips {
		t.Errorf("reset should refill chips, got %d", f.GetChips())
	}
	if f.GetSoda() == nil {
		t.Error("reset should burn a soda card")
	}
}

func TestFaro_GetBetRanksSorted(t *testing.T) {
	f := newFaroForTest()
	f.SetBet(9, 10, false)
	f.SetBet(2, 10, false)
	f.SetBet(13, 10, false)
	ranks := f.GetBetRanks()
	want := []int{2, 9, 13}
	for i := range want {
		if ranks[i] != want[i] {
			t.Errorf("ranks = %v, want %v", ranks, want)
		}
	}
}

func TestFaro_JSONRoundTrip(t *testing.T) {
	f := newFaroForTest()
	_ = f.PlayerPlaceBet(7, 100, true)
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var g Faro
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g.GetChips() != f.GetChips() {
		t.Errorf("chips mismatch: %d vs %d", g.GetChips(), f.GetChips())
	}
	if g.GetBets()[7] == nil || g.GetBets()[7].Amount != 100 || !g.GetBets()[7].Copper {
		t.Errorf("bet not preserved: %+v", g.GetBets()[7])
	}
}

func TestFaro_UnmarshalJSON_Hardening(t *testing.T) {
	cases := []string{
		`{"ps":99}`,         // phase out of range
		`{"ps":1,"tn":-1}`,  // negative turns
		`{"ps":1,"tn":100}`, // turns over max
		`{"ps":1,"bt":{"99":{"am":10,"cp":false}}}`, // rank out of range
		`{"ps":1,"bt":{"7":{"am":-5,"cp":false}}}`,  // negative amount
		`{"ps":1,"co":[99,1,2]}`,                    // call order rank out of range
	}
	for i, c := range cases {
		var g Faro
		if err := json.Unmarshal([]byte(c), &g); err == nil {
			t.Errorf("case %d: expected error for %s", i, c)
		}
	}
	// Invalid config falls back rather than erroring.
	var g Faro
	if err := json.Unmarshal([]byte(`{"ps":1,"cf":{"sc":-1,"mn":0,"mx":0}}`), &g); err != nil {
		t.Errorf("invalid config should fall back, got error: %v", err)
	}
	if g.GetConfig().StartChips != FaroDefaultStartChips {
		t.Error("config should fall back to default")
	}
}

func TestFaro_DealTurn_DeckGuard(t *testing.T) {
	f := newFaroForTest()
	// Force turnsPlayed near the limit; deal the rest and confirm it stops.
	f.SetPhase(FaroPhaseBetting)
	for f.GetPhase() == FaroPhaseBetting || f.GetPhase() == FaroPhaseTurn {
		if err := f.PlayerDealTurn(); err != nil {
			t.Fatalf("deal: %v", err)
		}
	}
	// Now in call phase; further deal must error.
	if err := f.PlayerDealTurn(); err == nil {
		t.Error("dealing past the last turn should error")
	}
}

// **ケースキーパーはこのゲームの中核。**未配の山札から直接数えるので、
// 公開済みカードの蓄積漏れや二重計上で嘘をつくことがない (#4894)。
func TestFaro_RemainingByRank(t *testing.T) {
	f := newFaroForTest()

	// リセット直後: ソーダ 1 枚だけが抜けている。
	start := f.GetRemainingByRank()
	total := 0
	for r := FaroMinRank; r <= FaroMaxRank; r++ {
		if start[r] < 0 || start[r] > 4 {
			t.Fatalf("rank %d = %d, want 0..4", r, start[r])
		}
		total += start[r]
	}
	if total != f.GetRemainingCount() {
		t.Fatalf("per-rank total = %d, deck remaining = %d", total, f.GetRemainingCount())
	}
	// ソーダの分だけ 51。
	if total != 51 {
		t.Fatalf("after the soda burn the deck should hold 51, got %d", total)
	}

	// **ターンを重ねると減る。**2 枚めくるので合計は 2 減る。
	if err := f.PlayerPlaceBet(1, 10, false); err != nil {
		t.Fatalf("bet: %v", err)
	}
	if err := f.PlayerDealTurn(); err != nil {
		t.Fatalf("deal: %v", err)
	}
	after := f.GetRemainingByRank()
	sum := 0
	for r := FaroMinRank; r <= FaroMaxRank; r++ {
		if after[r] > start[r] {
			t.Fatalf("rank %d went up: %d -> %d", r, start[r], after[r])
		}
		sum += after[r]
	}
	if sum != total-2 {
		t.Fatalf("after one turn the deck should be 2 smaller: %d -> %d", total, sum)
	}
	if sum != f.GetRemainingCount() {
		t.Fatalf("per-rank total = %d, deck remaining = %d", sum, f.GetRemainingCount())
	}

	// **リセットで元に戻る。**蓄積を持たないので取りこぼしようがない。
	f.Reset()
	again := f.GetRemainingByRank()
	back := 0
	for r := FaroMinRank; r <= FaroMaxRank; r++ {
		back += again[r]
	}
	if back != 51 {
		t.Fatalf("after a reset the deck should hold 51 again, got %d", back)
	}
}
