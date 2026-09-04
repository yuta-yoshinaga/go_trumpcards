//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pontoonDealAttempts caps the reshuffles a test may take while looking for a
// deal it can observe. A banker pontoon settles the round inside the deal, so a
// test that wants to see the mid-round state has to deal past one.
const pontoonDealAttempts = 100

// newTestPontoon returns a freshly reset Pontoon.
func newTestPontoon() *Pontoon {
	p := NewDefaultPontoon()
	p.Reset()
	return p
}

// pontoonCard builds a card for tests.
func pontoonCard(design, value int) *Card {
	return NewCard(design, value, true)
}

// pontoonHandOf builds a hand from (design, value) pairs.
func pontoonHandOf(bet int, vals ...int) *PontoonHand {
	h := &PontoonHand{bet: bet}
	for i, v := range vals {
		h.cards = append(h.cards, pontoonCard(i%4+1, v))
	}
	return h
}

func TestPontoon_Reset(t *testing.T) {
	p := newTestPontoon()

	if p.GetPhase() != PontoonPhaseBet {
		t.Errorf("phase = %d, want bet", p.GetPhase())
	}
	if p.GetChips() != PontoonDefaultChips {
		t.Errorf("chips = %d, want %d", p.GetChips(), PontoonDefaultChips)
	}
	if got := len(p.GetSeats()); got != PontoonSeatCnt {
		t.Errorf("seats = %d, want %d", got, PontoonSeatCnt)
	}
	if p.GetBankerIdx() != 0 && p.GetBankerIdx() >= PontoonSeatCnt {
		t.Errorf("banker = %d, out of range", p.GetBankerIdx())
	}
	if p.GetGameEndFlag() {
		t.Error("a fresh game should not be over")
	}
}

// A brand-new session must open on the ordinary flow -- place a bet -- rather
// than on "you deal". Two zero values conspire here: `banker`'s is seat 0, and
// `nextBanker`'s is 0 too, which Reset applies with `>= 0`. Fixing only the
// first would be undone by the second.
func TestPontoon_AFreshGameDoesNotOpenWithTheHumanBanking(t *testing.T) {
	p := NewDefaultPontoon()
	if p.IsHumanBanker() {
		t.Error("a brand-new game should not start with the human banking")
	}
	if p.GetNextBanker() != -1 {
		t.Errorf("nextBanker = %d, want -1 -- a zero would hand seat 0 the bank at Reset",
			p.GetNextBanker())
	}
	p.Reset()
	if p.IsHumanBanker() {
		t.Error("the first Reset should not hand the human the bank either")
	}
	if err := p.PlaceBet(100); err != nil {
		t.Errorf("PlaceBet on a fresh game: %v", err)
	}
}

func TestPontoon_Total(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want int
	}{
		{"two aces count 12", []int{1, 1}, 12},
		{"ace plus ten is 21", []int{1, 10}, 21},
		{"ace plus king is 21", []int{1, 13}, 21},
		{"face cards are ten", []int{11, 12}, 20},
		{"ace drops to one when needed", []int{1, 10, 5}, 16},
		{"two aces and a nine", []int{1, 1, 9}, 21},
		{"plain numbers", []int{2, 3, 4}, 9},
		{"empty", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := pontoonHandOf(0, tt.vals...)
			if got := pontoonTotal(h.cards); got != tt.want {
				t.Errorf("total = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPontoon_TotalIgnoresNil(t *testing.T) {
	if got := pontoonTotal([]*Card{nil, pontoonCard(CardDesignSpade, 5)}); got != 5 {
		t.Errorf("total = %d, want 5", got)
	}
}

func TestPontoon_Rank(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want PontoonRank
	}{
		{"ace and ten is a pontoon", []int{1, 10}, PontoonRankPontoon},
		{"ace and jack is a pontoon", []int{1, 11}, PontoonRankPontoon},
		// 21 built from three cards is NOT a pontoon -- only the two-card hand is.
		{"three cards making 21 is not a pontoon", []int{7, 7, 7}, PontoonRankPoints},
		{"exactly five cards under 21", []int{2, 3, 4, 5, 2}, PontoonRankFiveCard},
		{"five cards making 21", []int{2, 3, 4, 5, 7}, PontoonRankFiveCard},
		{"bust beats nothing", []int{10, 10, 10}, PontoonRankBust},
		{"ordinary points", []int{10, 8}, PontoonRankPoints},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := pontoonHandOf(0, tt.vals...)
			if got := pontoonRankOf(h.cards); got != tt.want {
				t.Errorf("rank = %v, want %v", got, tt.want)
			}
		})
	}
}

// A five card trick is EXACTLY five cards -- reaching five ends the hand, so a
// six-card hand cannot arise. #4379 says "five or more".
func TestPontoon_FiveCardTrickIsExactlyFive(t *testing.T) {
	four := pontoonHandOf(0, 2, 3, 4, 5)
	if got := pontoonRankOf(four.cards); got != PontoonRankPoints {
		t.Errorf("four cards rank = %v, want points", got)
	}
	five := pontoonHandOf(0, 2, 3, 4, 5, 2)
	if got := pontoonRankOf(five.cards); got != PontoonRankFiveCard {
		t.Errorf("five cards rank = %v, want five card trick", got)
	}
}

func TestPontoon_PlaceBet(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1 // a CPU banks, so the human bets

	if err := p.PlaceBet(PontoonMinBet - 1); err == nil {
		t.Error("a bet below the minimum should be rejected")
	}
	if err := p.PlaceBet(PontoonMaxBet + 1); err == nil {
		t.Error("a bet above the maximum should be rejected")
	}
	if err := p.PlaceBet(PontoonDefaultChips * 10); err == nil {
		t.Error("betting more than the stack should be rejected")
	}

	// 親がポンツーンを引いた局は deal の時点で精算まで走るので、賭け金だけが
	// 引かれた状態を観測できない。手番に入る配札が出るまで配り直す。
	var before int
	for range pontoonDealAttempts {
		p.Reset()
		p.banker = 1
		before = p.GetChips()
		if err := p.PlaceBet(100); err != nil {
			t.Fatalf("PlaceBet: %v", err)
		}
		if p.GetPhase() == PontoonPhasePlayerTurn {
			break
		}
	}
	if p.GetPhase() != PontoonPhasePlayerTurn {
		t.Fatalf("no deal in %d left the banker without a pontoon", pontoonDealAttempts)
	}
	if p.GetChips() != before-100 {
		t.Errorf("chips = %d, want %d", p.GetChips(), before-100)
	}
}

func TestPontoon_PlaceBetRejectedOutOfPhase(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1
	p.phase = PontoonPhaseEnd
	if err := p.PlaceBet(100); err == nil {
		t.Error("PlaceBet should fail outside the betting phase")
	}
}

// The banker takes the other players' bets rather than making one.
func TestPontoon_BankerDoesNotBet(t *testing.T) {
	p := newTestPontoon()
	p.banker = 0
	if err := p.PlaceBet(100); err == nil {
		t.Error("the banker must not be asked for a bet")
	}
	// 自分がポンツーンを引いた局はその場で全員から取り立てるので残高が動く。
	// 動かないことを見たいのは「配っただけ」の局。
	// 基準はループ内で取り直す。Reset は残高を戻さないので、前の周回で取り立てた
	// ぶんがそのまま残る。
	var before int
	for range pontoonDealAttempts {
		p.Reset()
		p.banker = 0
		before = p.GetChips()
		if err := p.StartAsBanker(); err != nil {
			t.Fatalf("StartAsBanker: %v", err)
		}
		if p.GetPhase() != PontoonPhaseEnd {
			break
		}
	}
	if p.GetPhase() == PontoonPhaseEnd {
		t.Fatalf("no deal in %d left the banker without a pontoon", pontoonDealAttempts)
	}
	if p.GetChips() != before {
		t.Errorf("the banker's stack changed at the deal: %d, want %d", p.GetChips(), before)
	}
}

func TestPontoon_StartAsBankerRejectedWhenNotBanker(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1
	if err := p.StartAsBanker(); err == nil {
		t.Error("StartAsBanker should fail when a CPU banks")
	}
}

// setupHumanTurn puts the human on turn with an exact hand.
func setupHumanTurn(p *Pontoon, vals ...int) *PontoonHand {
	p.banker = 1
	p.seats = []*PontoonSeat{
		{name: "あなた", isCPU: false},
		{name: "CPU1", isCPU: true},
		{name: "CPU2", isCPU: true},
	}
	h := pontoonHandOf(100, vals...)
	p.seats[0].hands = []*PontoonHand{h}
	p.seats[2].hands = []*PontoonHand{pontoonHandOf(20, 10, 8)}
	p.bankerHand = pontoonHandOf(0, 10, 8)
	p.phase = PontoonPhasePlayerTurn
	p.activeSeat = 0
	p.activeHand = 0
	return h
}

// Sticking below 15 is not allowed -- the player must keep drawing.
func TestPontoon_CannotStickBelowFifteen(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 10, 4)

	if p.CanStick() {
		t.Error("CanStick should be false on 14")
	}
	if err := p.Stick(); err == nil {
		t.Error("sticking on 14 should be rejected")
	}

	setupHumanTurn(p, 10, 5)
	if !p.CanStick() {
		t.Error("CanStick should be true on 15")
	}
	if err := p.Stick(); err != nil {
		t.Errorf("sticking on 15: %v", err)
	}
}

func TestPontoon_Twist(t *testing.T) {
	p := newTestPontoon()
	h := setupHumanTurn(p, 5, 4)

	if !p.CanTwist() {
		t.Fatal("CanTwist should be true on 9")
	}
	if err := p.Twist(); err != nil {
		t.Fatalf("Twist: %v", err)
	}
	if len(h.cards) != 3 {
		t.Errorf("cards = %d, want 3", len(h.cards))
	}
	if !h.IsTwisted() {
		t.Error("the hand should be marked as twisted")
	}
	if h.GetBet() != 100 {
		t.Errorf("bet = %d, twisting must not change the stake", h.GetBet())
	}
}

func TestPontoon_TwistRejectedAtTwentyOneAndFiveCards(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 7, 7, 7)
	if p.CanTwist() {
		t.Error("CanTwist should be false on 21")
	}
	if err := p.Twist(); err == nil {
		t.Error("twisting on 21 should be rejected")
	}

	setupHumanTurn(p, 2, 3, 2, 3, 2)
	if p.CanTwist() {
		t.Error("CanTwist should be false on a five card trick")
	}
	if err := p.Twist(); err == nil {
		t.Error("twisting on five cards should be rejected")
	}
}

// Buying after twisting is forbidden: it would let a player draw cheaply and
// then raise, which is exactly what the rule exists to stop.
func TestPontoon_CannotBuyAfterTwisting(t *testing.T) {
	p := newTestPontoon()
	h := setupHumanTurn(p, 5, 4)

	if !p.CanBuy() {
		t.Fatal("CanBuy should be true before twisting")
	}
	if err := p.Twist(); err != nil {
		t.Fatalf("Twist: %v", err)
	}
	if p.CanBuy() {
		t.Error("CanBuy should be false once the hand has twisted")
	}
	if err := p.Buy(100); err == nil {
		t.Error("buying after a twist should be rejected")
	}
	_ = h
}

func TestPontoon_Buy(t *testing.T) {
	p := newTestPontoon()
	h := setupHumanTurn(p, 5, 4)
	before := p.GetChips()

	if err := p.Buy(PontoonMinBet - 1); err == nil {
		t.Error("an extra stake below the minimum should be rejected")
	}
	if err := p.Buy(h.GetBet()*2 + 1); err == nil {
		t.Error("an extra stake above twice the bet should be rejected")
	}
	if err := p.Buy(100); err != nil {
		t.Fatalf("Buy: %v", err)
	}
	if h.GetBet() != 200 {
		t.Errorf("bet = %d, want 200", h.GetBet())
	}
	if p.GetChips() != before-100 {
		t.Errorf("chips = %d, want %d", p.GetChips(), before-100)
	}
	if len(h.GetCards()) != 3 {
		t.Errorf("cards = %d, want 3", len(h.GetCards()))
	}
}

func TestPontoon_BuyRejectedWithoutChips(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 5, 4)
	p.chips.SetChips(0)
	if p.CanBuy() {
		t.Error("CanBuy should be false with no chips")
	}
	if err := p.Buy(100); err == nil {
		t.Error("buying without chips should be rejected")
	}
}

func TestPontoon_Split(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 8, 8)
	before := p.GetChips()

	if !p.CanSplit() {
		t.Fatal("CanSplit should be true on a pair")
	}
	if err := p.Split(); err != nil {
		t.Fatalf("Split: %v", err)
	}
	if got := len(p.GetSeats()[0].GetHands()); got != 2 {
		t.Errorf("hands = %d, want 2", got)
	}
	if p.GetChips() != before-100 {
		t.Errorf("chips = %d, want %d", p.GetChips(), before-100)
	}
	for i, h := range p.GetSeats()[0].GetHands() {
		if len(h.GetCards()) != 2 {
			t.Errorf("hand %d has %d cards, want 2", i, len(h.GetCards()))
		}
	}
}

// Two ten-point cards of different rank are not a pair: a queen and a jack are
// both worth ten but cannot be split.
func TestPontoon_SplitNeedsEqualRankNotEqualValue(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 12, 11)
	if p.CanSplit() {
		t.Error("a queen and a jack must not be splittable")
	}
	if err := p.Split(); err == nil {
		t.Error("splitting a queen and a jack should be rejected")
	}

	setupHumanTurn(p, 12, 12)
	if !p.CanSplit() {
		t.Error("two queens should be splittable")
	}
}

func TestPontoon_SplitRejectedAtTheHandLimit(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 8, 8)
	s := p.GetSeats()[0]
	for len(s.hands) < PontoonMaxHands {
		s.hands = append(s.hands, pontoonHandOf(100, 8, 8))
	}
	if p.CanSplit() {
		t.Error("CanSplit should be false at the hand limit")
	}
	if err := p.Split(); err == nil {
		t.Error("splitting past the limit should be rejected")
	}
}

func TestPontoon_ActionsRejectedOutOfTurn(t *testing.T) {
	p := newTestPontoon()
	p.phase = PontoonPhaseBet
	for name, fn := range map[string]func() error{
		"stick": p.Stick,
		"twist": p.Twist,
		"split": p.Split,
		"buy":   func() error { return p.Buy(100) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := fn(); err == nil {
				t.Errorf("%s should be rejected outside the player's turn", name)
			}
		})
	}
}

func TestPontoon_SettleHand(t *testing.T) {
	p := newTestPontoon()
	const bet = 100

	tests := []struct {
		name        string
		hand        []int
		bankerRank  PontoonRank
		bankerTotal int
		want        int
	}{
		{"bust loses the stake", []int{10, 10, 10}, PontoonRankPoints, 18, -bet},
		{"pontoon pays double", []int{1, 10}, PontoonRankPoints, 20, bet * 2},
		{"five card trick pays double", []int{2, 3, 4, 5, 2}, PontoonRankPoints, 20, bet * 2},
		// 親のポンツーンは全員から倍額を取る。プレイヤーのポンツーンでも勝てない。
		{"the banker's pontoon beats a player's", []int{1, 10}, PontoonRankPontoon, 21, -bet * 2},
		// 親のファイブカード・トリックはポンツーン以外に倍額で勝つ。
		{"the banker's five card trick beats 21", []int{7, 7, 7}, PontoonRankFiveCard, 19, -bet * 2},
		{"a pontoon still beats the banker's five card trick", []int{1, 10}, PontoonRankFiveCard, 19, bet * 2},
		{"the banker busting pays everyone", []int{10, 6}, PontoonRankBust, 24, bet},
		{"beating the banker pays once", []int{10, 9}, PontoonRankPoints, 18, bet},
		// 同点は親の勝ち。
		{"a tie goes to the banker", []int{10, 8}, PontoonRankPoints, 18, -bet},
		{"losing on points", []int{10, 7}, PontoonRankPoints, 18, -bet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := pontoonHandOf(bet, tt.hand...)
			if got := p.settleHand(h, tt.bankerRank, tt.bankerTotal); got != tt.want {
				t.Errorf("settleHand = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPontoon_SettleCreditsTheHumanStack(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 10, 9)
	p.bankerHand = pontoonHandOf(0, 10, 8)
	before := p.GetChips()

	p.settle()

	// 賭け金 100 が戻り、さらに 100 の勝ち。
	if p.GetChips() != before+200 {
		t.Errorf("chips = %d, want %d", p.GetChips(), before+200)
	}
	if p.GetPhase() != PontoonPhaseEnd {
		t.Errorf("phase = %d, want end", p.GetPhase())
	}
	if !p.GetGameEndFlag() {
		t.Error("GetGameEndFlag should be true after settling")
	}
	if p.GetLastResult() == "" {
		t.Error("the settlement should be summarised")
	}
}

// When the human banks, the players' winnings come out of the human's stack.
func TestPontoon_SettleDebitsTheHumanBanker(t *testing.T) {
	p := newTestPontoon()
	p.banker = 0
	p.seats = []*PontoonSeat{
		{name: "あなた", isCPU: false},
		{name: "CPU1", isCPU: true, hands: []*PontoonHand{pontoonHandOf(50, 10, 9)}},
		{name: "CPU2", isCPU: true, hands: []*PontoonHand{pontoonHandOf(50, 10, 4)}},
	}
	p.bankerHand = pontoonHandOf(0, 10, 8)
	before := p.GetChips()

	p.settle()

	// CPU1 は 19 で勝ち (-50)、CPU2 は 14 で負け (+50) → 差し引き 0。
	if p.GetChips() != before {
		t.Errorf("chips = %d, want %d", p.GetChips(), before)
	}
}

// The bank passes to a player who makes an unsplit pontoon, not to whoever has
// the best hand among the winners. #4379 has this the other way round.
func TestPontoon_PontoonTakesTheBank(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1
	p.seats = []*PontoonSeat{
		{name: "あなた", isCPU: false, hands: []*PontoonHand{pontoonHandOf(100, 1, 10)}},
		{name: "CPU1", isCPU: true},
		{name: "CPU2", isCPU: true, hands: []*PontoonHand{pontoonHandOf(20, 10, 9)}},
	}
	p.bankerHand = pontoonHandOf(0, 10, 8)
	p.nextBanker = -1

	p.settle()

	if p.GetNextBanker() != 0 {
		t.Errorf("nextBanker = %d, want 0", p.GetNextBanker())
	}
	p.Reset()
	if p.GetBankerIdx() != 0 {
		t.Errorf("banker = %d, want the pontoon holder", p.GetBankerIdx())
	}
	if !p.IsHumanBanker() {
		t.Error("IsHumanBanker should be true after taking the bank")
	}
}

// The banker's own pontoon keeps the bank where it is.
func TestPontoon_BankerPontoonKeepsTheBank(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1
	p.seats = []*PontoonSeat{
		{name: "あなた", isCPU: false, hands: []*PontoonHand{pontoonHandOf(100, 1, 10)}},
		{name: "CPU1", isCPU: true},
		{name: "CPU2", isCPU: true},
	}
	p.bankerHand = pontoonHandOf(0, 1, 13)
	p.nextBanker = -1

	p.settle()

	if p.GetNextBanker() != -1 {
		t.Errorf("nextBanker = %d, want -1", p.GetNextBanker())
	}
}

// A pontoon made across split hands does not take the bank.
func TestPontoon_SplitPontoonDoesNotTakeTheBank(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1
	p.seats = []*PontoonSeat{
		{name: "あなた", isCPU: false, hands: []*PontoonHand{
			pontoonHandOf(100, 1, 10),
			pontoonHandOf(100, 10, 5),
		}},
		{name: "CPU1", isCPU: true},
		{name: "CPU2", isCPU: true},
	}
	p.bankerHand = pontoonHandOf(0, 10, 8)
	p.nextBanker = -1

	p.settle()

	if p.GetNextBanker() != -1 {
		t.Errorf("nextBanker = %d, want -1 for a split hand", p.GetNextBanker())
	}
}

func TestPontoon_BankerTurnWhenHumanBanks(t *testing.T) {
	p := newTestPontoon()
	p.banker = 0
	p.seats = []*PontoonSeat{
		{name: "あなた", isCPU: false},
		{name: "CPU1", isCPU: true, hands: []*PontoonHand{pontoonHandOf(50, 10, 9)}},
		{name: "CPU2", isCPU: true, hands: []*PontoonHand{pontoonHandOf(50, 10, 8)}},
	}
	p.bankerHand = pontoonHandOf(0, 5, 4)
	p.phase = PontoonPhasePlayerTurn
	p.activeSeat = 0
	p.advanceToHuman()

	if p.GetPhase() != PontoonPhaseBankerTurn {
		t.Fatalf("phase = %d, want banker turn", p.GetPhase())
	}
	if err := p.BankerTwist(); err != nil {
		t.Fatalf("BankerTwist: %v", err)
	}
	if len(p.GetBankerHand().GetCards()) != 3 {
		t.Errorf("banker cards = %d, want 3", len(p.GetBankerHand().GetCards()))
	}
	if p.GetPhase() == PontoonPhaseBankerTurn {
		if err := p.BankerStay(); err != nil {
			t.Fatalf("BankerStay: %v", err)
		}
	}
	if p.GetPhase() != PontoonPhaseEnd {
		t.Errorf("phase = %d, want end", p.GetPhase())
	}
}

func TestPontoon_BankerActionsRejectedOutOfPhase(t *testing.T) {
	p := newTestPontoon()
	p.phase = PontoonPhaseBet
	if err := p.BankerTwist(); err == nil {
		t.Error("BankerTwist should be rejected outside the banker's turn")
	}
	if err := p.BankerStay(); err == nil {
		t.Error("BankerStay should be rejected outside the banker's turn")
	}
}

func TestPontoon_BankerTwistRejectedAtLimits(t *testing.T) {
	p := newTestPontoon()
	p.banker = 0
	p.phase = PontoonPhaseBankerTurn
	p.bankerHand = pontoonHandOf(0, 2, 3, 4, 5, 2)
	if err := p.BankerTwist(); err == nil {
		t.Error("the banker must not draw a sixth card")
	}
	p.bankerHand = pontoonHandOf(0, 10, 10, 10)
	if err := p.BankerTwist(); err == nil {
		t.Error("a bust banker must not draw again")
	}
}

func TestPontoon_CpuBanksAndSettlesAutomatically(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1
	if err := p.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	// 人間の手を打ち切ると、CPU 親が自動で引いて精算まで進む。
	for range 10 {
		if p.GetPhase() != PontoonPhasePlayerTurn {
			break
		}
		if p.CanStick() {
			if err := p.Stick(); err != nil {
				t.Fatalf("Stick: %v", err)
			}
			continue
		}
		if err := p.Twist(); err != nil {
			t.Fatalf("Twist: %v", err)
		}
	}
	if p.GetPhase() != PontoonPhaseEnd {
		t.Errorf("phase = %d, want end once the CPU banker has played", p.GetPhase())
	}
}

func TestPontoon_ActionLog(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1
	if err := p.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	log := p.GetActionLog()
	if len(log) == 0 {
		t.Fatal("the deal should be logged")
	}
	if log[0].ActionType != "deal" {
		t.Errorf("log[0].ActionType = %q, want deal", log[0].ActionType)
	}
}

func TestPontoon_JSONRoundTrip(t *testing.T) {
	p := newTestPontoon()
	p.banker = 1
	if err := p.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	restored := NewDefaultPontoon()
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored.GetPhase() != p.GetPhase() {
		t.Errorf("phase = %d, want %d", restored.GetPhase(), p.GetPhase())
	}
	if restored.GetChips() != p.GetChips() {
		t.Errorf("chips = %d, want %d", restored.GetChips(), p.GetChips())
	}
	if restored.GetBankerIdx() != p.GetBankerIdx() {
		t.Errorf("banker = %d, want %d", restored.GetBankerIdx(), p.GetBankerIdx())
	}
	if len(restored.GetSeats()) != len(p.GetSeats()) {
		t.Fatalf("seats = %d, want %d", len(restored.GetSeats()), len(p.GetSeats()))
	}
	// 手札そのものが戻ること。ここが空だと盤面が消える。
	want := p.GetSeats()[0].GetHands()
	got := restored.GetSeats()[0].GetHands()
	if len(got) != len(want) {
		t.Fatalf("hands = %d, want %d", len(got), len(want))
	}
	if len(want) > 0 && len(got[0].GetCards()) != len(want[0].GetCards()) {
		t.Errorf("cards = %d, want %d", len(got[0].GetCards()), len(want[0].GetCards()))
	}
	if len(want) > 0 && got[0].GetBet() != want[0].GetBet() {
		t.Errorf("bet = %d, want %d", got[0].GetBet(), want[0].GetBet())
	}
	if restored.GetBankerHand() == nil || len(restored.GetBankerHand().GetCards()) == 0 {
		t.Error("the banker's hand must survive the wire")
	}
}

func TestPontoon_UnmarshalJSONRejectsInvalidState(t *testing.T) {
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
		{"active hand out of range", `{"ph":1,"ah":99}`},
		{"negative chips", `{"ph":1,"ch":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewDefaultPontoon()
			if err := json.Unmarshal([]byte(tt.data), p); err == nil {
				t.Errorf("Unmarshal(%s) should fail", tt.data)
			}
		})
	}
}

func TestPontoon_UnmarshalJSONRejectsOversizedArrays(t *testing.T) {
	build := func(field, elem string, n int) string {
		out := `{"ph":1,"` + field + `":[`
		for i := range n {
			if i > 0 {
				out += ","
			}
			out += elem
		}
		return out + `]}`
	}
	p := NewDefaultPontoon()
	if err := json.Unmarshal([]byte(build("al", `{}`, pontoonMaxSliceLen+1)), p); err == nil {
		t.Error("an oversized action log should be rejected")
	}

	// 席あたりの手の上限。
	hands := `{"hd":[`
	for i := range PontoonMaxHands + 1 {
		if i > 0 {
			hands += ","
		}
		hands += `{}`
	}
	hands += `]}`
	var seat PontoonSeat
	if err := json.Unmarshal([]byte(hands), &seat); err == nil {
		t.Error("a seat holding too many hands should be rejected")
	}

	// 手あたりの札の上限。
	cards := `{"cd":[`
	for i := range pontoonMaxSliceLen + 1 {
		if i > 0 {
			cards += ","
		}
		cards += `{"d":1,"v":1,"f":true}`
	}
	cards += `]}`
	var hand PontoonHand
	if err := json.Unmarshal([]byte(cards), &hand); err == nil {
		t.Error("an oversized hand should be rejected")
	}
}

func TestPontoon_UnmarshalJSONRejectsMalformedNested(t *testing.T) {
	var hand PontoonHand
	if err := json.Unmarshal([]byte(`{`), &hand); err == nil {
		t.Error("a malformed hand should be rejected")
	}
	var seat PontoonSeat
	if err := json.Unmarshal([]byte(`{`), &seat); err == nil {
		t.Error("a malformed seat should be rejected")
	}
}

func TestPontoon_ResetRestoresBrokeStack(t *testing.T) {
	p := newTestPontoon()
	p.chips.SetChips(PontoonMinBet - 1)
	p.Reset()
	if p.GetChips() != PontoonDefaultChips {
		t.Errorf("chips = %d, want a fresh stack", p.GetChips())
	}
}

func TestPontoon_HandAndSeatAccessors(t *testing.T) {
	p := newTestPontoon()
	h := setupHumanTurn(p, 10, 5)

	if h.IsStuck() {
		t.Error("a fresh hand should not be stuck")
	}
	if h.GetPayout() != 0 {
		t.Errorf("payout = %d, want 0 before settling", h.GetPayout())
	}
	// Stick は席を進め、CPU 親が自動で打って精算まで走る。順序を入れ替えると
	// payout が既に埋まっている。
	if err := p.Stick(); err != nil {
		t.Fatalf("Stick: %v", err)
	}
	if !h.IsStuck() {
		t.Error("IsStuck should be true after sticking")
	}

	s := p.GetSeats()[0]
	if s.GetName() != "あなた" {
		t.Errorf("name = %q, want あなた", s.GetName())
	}
	if s.IsCPU() {
		t.Error("seat 0 is the human")
	}
	if !p.GetSeats()[1].IsCPU() {
		t.Error("seat 1 should be a CPU")
	}
}

func TestPontoon_PayoutRecordedOnSettle(t *testing.T) {
	p := newTestPontoon()
	h := setupHumanTurn(p, 10, 9)
	p.bankerHand = pontoonHandOf(0, 10, 8)
	p.settle()

	if h.GetPayout() != 100 {
		t.Errorf("payout = %d, want 100", h.GetPayout())
	}
}

func TestPontoon_TurnPositionAccessors(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 8, 8)

	if p.GetActiveSeat() != 0 {
		t.Errorf("activeSeat = %d, want 0", p.GetActiveSeat())
	}
	if p.GetActiveHand() != 0 {
		t.Errorf("activeHand = %d, want 0", p.GetActiveHand())
	}
	if err := p.Split(); err != nil {
		t.Fatalf("Split: %v", err)
	}
	if err := p.Stick(); err != nil && p.CanStick() {
		t.Fatalf("Stick: %v", err)
	}
}

func TestPontoon_DisplayHelpers(t *testing.T) {
	p := newTestPontoon()
	cards := pontoonHandOf(0, 1, 10).cards

	if got := p.GetHandTotal(cards); got != 21 {
		t.Errorf("GetHandTotal = %d, want 21", got)
	}
	if got := p.GetHandRank(cards); got != PontoonRankPontoon {
		t.Errorf("GetHandRank = %v, want pontoon", got)
	}
}

func TestPontoon_DescribeResult(t *testing.T) {
	p := newTestPontoon()
	tests := []struct {
		name  string
		rank  PontoonRank
		total int
		want  string
	}{
		{"pontoon", PontoonRankPontoon, 21, "親がポンツーン"},
		{"five card trick", PontoonRankFiveCard, 19, "親がファイブカード・トリック"},
		{"bust", PontoonRankBust, 24, "親がバースト（24）"},
		{"points", PontoonRankPoints, 18, "親は 18"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.describeResult(tt.rank, tt.total); got != tt.want {
				t.Errorf("describeResult = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPontoon_CurrentHandRejectsCpuAndMissingHand(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 10, 5)

	// 手番が CPU 席に進んでいるとき。
	p.activeSeat = 2
	if _, err := p.currentHand(); err == nil {
		t.Error("currentHand should fail while a CPU is on turn")
	}

	// 席は人間だが手の番号が範囲外のとき。
	p.activeSeat = 0
	p.activeHand = 9
	if _, err := p.currentHand(); err == nil {
		t.Error("currentHand should fail with no hand at the index")
	}
	if p.CanSplit() {
		t.Error("CanSplit should be false with no hand in play")
	}
	if p.CanStick() || p.CanTwist() || p.CanBuy() {
		t.Error("no action should be offered with no hand in play")
	}
}

// A CPU seat plays itself out: it must never stop below 15 and never exceed
// five cards.
func TestPontoon_CpuSeatRespectsTheStickMinimum(t *testing.T) {
	p := newTestPontoon()
	for range 200 {
		p.Reset()
		p.banker = 0
		if err := p.StartAsBanker(); err != nil {
			t.Fatalf("StartAsBanker: %v", err)
		}
		// 親がポンツーンなら局は配った時点で終わり、他の席は打たない。その局の
		// 2 枚の手は「止まった」のではなく「打つ前」なので対象外。
		if pontoonRankOf(p.GetBankerHand().GetCards()) == PontoonRankPontoon {
			continue
		}
		for i, s := range p.GetSeats() {
			if i == p.GetBankerIdx() {
				continue
			}
			for _, h := range s.GetHands() {
				total := pontoonTotal(h.GetCards())
				if len(h.GetCards()) > PontoonMaxCards {
					t.Fatalf("seat %d holds %d cards", i, len(h.GetCards()))
				}
				// 5 枚に達した手はファイブカード・トリックで、21 以下なら
				// 14 点でも成立する。止まってよいのはここと 15 以上だけ。
				if len(h.GetCards()) < PontoonMaxCards &&
					total <= PontoonTarget && total < PontoonStickMin {
					t.Fatalf("seat %d stopped on %d with %d cards, below the stick minimum",
						i, total, len(h.GetCards()))
				}
			}
		}
	}
}

func TestPontoon_SplitHandsArePlayedInOrder(t *testing.T) {
	p := newTestPontoon()
	setupHumanTurn(p, 8, 8)
	if err := p.Split(); err != nil {
		t.Fatalf("Split: %v", err)
	}
	hands := p.GetSeats()[0].GetHands()
	if len(hands) != 2 {
		t.Fatalf("hands = %d, want 2", len(hands))
	}
	// 分けた手は元の手のすぐ後ろに入る。手番は先頭のまま。
	if p.GetActiveHand() != 0 {
		t.Errorf("activeHand = %d, want 0", p.GetActiveHand())
	}
	if hands[0].GetBet() != hands[1].GetBet() {
		t.Errorf("split hands carry different bets: %d vs %d", hands[0].GetBet(), hands[1].GetBet())
	}
}

// #5565: 停止ラインは 2 箇所に 17 が直書きされていた。定数にしたので、実際に
// 打たせて確かめる。
//
// **issue の前提は誤りだった。**「CPU 席・CPU 親のどちらも 17 まで引き続ける」と
// 書かれているが、CPU 席は 15 で止まる (4 枚のときだけファイブカード・トリックを
// 狙って 17 未満でも引く)。17 まで引くのは親だけ。案内をそのまま書いていたら、
// 相手の停止ラインについて嘘を教えることになっていた。
func TestPontoonCpuStopsAtTheNamedThreshold(t *testing.T) {
	// 定数そのものは規則の一部。黙って変わると案内も一緒に変わってしまうので固定する。
	assert.Equal(t, 17, PontoonCpuStickMin)
	assert.Greater(t, PontoonCpuStickMin, PontoonStickMin,
		"the CPU cannot stop below the total a player is allowed to declare at")

	// 200 局を**最後まで**打たせて、CPU 席と親が閾値未満で止まっていないこと。
	// 局の途中で見ると、まだ打っていない手を「閾値未満で止まった」と誤読する。
	for range 200 {
		p := NewDefaultPontoon()
		p.Reset()
		if err := p.PlaceBet(10); err != nil {
			continue
		}
		for range 10 { // 人間の手番を終わらせる。5 枚上限があるので必ず抜ける
			if p.GetPhase() != PontoonPhasePlayerTurn {
				break
			}
			if p.CanStick() {
				_ = p.Stick()
				continue
			}
			if p.Twist() != nil {
				break
			}
		}
		if p.GetPhase() != PontoonPhaseEnd {
			continue
		}
		if bh := p.bankerHand; bh != nil {
			total := pontoonTotal(bh.cards)
			if total <= PontoonTarget && len(bh.cards) < PontoonMaxCards {
				assert.GreaterOrEqual(t, total, PontoonCpuStickMin, "the banker stopped below the threshold")
			}
		}
		for i, s := range p.GetSeats() {
			// 親の席は席として打たない (手は bankerHand の側)。ここを外すと、
			// 配ったまま止まっている手を「閾値未満で止まった」と誤読する。
			if !s.IsCPU() || i == p.banker {
				continue
			}
			for _, h := range s.GetHands() {
				// **止まった手だけを見る。**バースト・5 枚・局が先に終わって
				// 打たれなかった手は「閾値未満で止まった」ではない。
				if !h.stuck {
					continue
				}
				total := pontoonTotal(h.cards)
				// **CPU 席は 17 ではなく 15 で止まる。**4 枚のときだけ
				// ファイブカード・トリックを狙って 17 未満でも引く。
				assert.GreaterOrEqual(t, total, PontoonStickMin,
					"a CPU seat stopped below the total it is allowed to declare at")
			}
		}
	}
}

// 4 枚で 15〜16 のとき、CPU 席はファイブカード・トリックを狙ってもう 1 枚引く。
// 「15 で止まる」とだけ案内すると、この手を見た人が規則を疑うことになる。
func TestPontoonCpuSeatChasesTheFiveCardTrick(t *testing.T) {
	p := NewDefaultPontoon()
	p.Reset()
	s := p.GetSeats()[1]
	h := &PontoonHand{cards: []*Card{
		NewCard(CardDesignSpade, 2, true),
		NewCard(CardDesignHeart, 3, true),
		NewCard(CardDesignClover, 4, true),
		NewCard(CardDesignDiamond, 6, true),
	}}
	s.hands = []*PontoonHand{h}
	require.Equal(t, 15, pontoonTotal(h.cards))

	p.playCpuSeat(s)
	assert.Len(t, h.cards, PontoonMaxCards, "a four-card 15 draws one more")
}

// 3 枚で 15 なら止まる。上の手と違うのは枚数だけ。
func TestPontoonCpuSeatSticksOnFifteenWithThreeCards(t *testing.T) {
	p := NewDefaultPontoon()
	p.Reset()
	s := p.GetSeats()[1]
	h := &PontoonHand{cards: []*Card{
		NewCard(CardDesignSpade, 5, true),
		NewCard(CardDesignHeart, 4, true),
		NewCard(CardDesignClover, 6, true),
	}}
	s.hands = []*PontoonHand{h}
	require.Equal(t, 15, pontoonTotal(h.cards))

	p.playCpuSeat(s)
	assert.Len(t, h.cards, 3, "a three-card 15 sticks")
	assert.True(t, h.stuck)
}

// 4 枚でも 17 以上なら止まる。ファイブカード・トリック狙いは 17 未満のときだけで、
// ここを外すと 18 の手を壊してまで 5 枚目を引く。
func TestPontoonCpuSeatDoesNotChaseFromSeventeen(t *testing.T) {
	p := NewDefaultPontoon()
	p.Reset()
	s := p.GetSeats()[1]
	h := &PontoonHand{cards: []*Card{
		NewCard(CardDesignSpade, 5, true),
		NewCard(CardDesignHeart, 4, true),
		NewCard(CardDesignClover, 6, true),
		NewCard(CardDesignDiamond, 3, true),
	}}
	s.hands = []*PontoonHand{h}
	require.Equal(t, 18, pontoonTotal(h.cards))

	p.playCpuSeat(s)
	assert.Len(t, h.cards, 4, "a four-card 18 sticks rather than chasing the five-card trick")
	assert.True(t, h.stuck)
}

func TestPontoon_ErrorsCarryAnI18nCode(t *testing.T) {
	codeOf := func(t *testing.T, err error) string {
		t.Helper()
		if err == nil {
			return ""
		}
		code, _ := ErrorMessageCode(err)
		return code
	}

	t.Run("every refusal names a key instead of an English sentence", func(t *testing.T) {
		cases := []struct {
			name string
			run  func(y *Pontoon) error
		}{
			{"PlaceBet outside phase", func(y *Pontoon) error {
				y.phase = PontoonPhaseEnd
				return y.PlaceBet(100)
			}},
			{"StartAsBanker when not banker", func(y *Pontoon) error {
				y.phase = PontoonPhaseBet
				y.banker = 1
				return y.StartAsBanker()
			}},
			{"Stick below 15", func(y *Pontoon) error {
				setupHumanTurn(y, 10, 4)
				return y.Stick()
			}},
			{"Twist out of turn", func(y *Pontoon) error {
				y.phase = PontoonPhaseEnd
				return y.Twist()
			}},
			{"Buy out of turn", func(y *Pontoon) error {
				y.phase = PontoonPhaseEnd
				return y.Buy(10)
			}},
			{"Split non-matching cards", func(y *Pontoon) error {
				setupHumanTurn(y, 10, 9)
				return y.Split()
			}},
			{"BankerTwist out of phase", func(y *Pontoon) error {
				y.phase = PontoonPhaseEnd
				return y.BankerTwist()
			}},
			{"BankerStay out of phase", func(y *Pontoon) error {
				y.phase = PontoonPhaseEnd
				return y.BankerStay()
			}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				y := newTestPontoon()
				err := tc.run(y)
				require.Error(t, err, "この操作は拒否されるはずで、拒否されないと何も測れない")
				code := codeOf(t, err)
				assert.NotEmpty(t, code, "コードが無いと CUI は英語をそのまま出す")
				assert.Truef(t, strings.HasPrefix(code, "pontoon."),
					"キーは pontoon 名前空間に置く (got %q)", code)
			})
		}
	})
}
