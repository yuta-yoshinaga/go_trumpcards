//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func pjCard(design, value int) *Card { return NewCard(design, value, true) }

// pjReady puts a fresh game into the play phase for the given seat.
func pjReady(t *testing.T, seat int) *PopeJoan {
	t.Helper()
	p := NewDefaultPopeJoan()
	p.Reset()
	p.SetPhaseForTest(PopeJoanPhasePlay)
	p.SetCurrentPlayerForTest(seat)
	return p
}

// setPjHand replaces a seat's hand outright.
func setPjHand(p *PopeJoan, seat int, cards []*Card) {
	pl := p.GetPlayer(seat)
	pl.Reset()
	for _, c := range cards {
		pl.AddCard(c)
	}
}

// pjBoard returns a dressed board.
func pjBoard() PopeJoanBoard {
	var b PopeJoanBoard
	b.Dress()
	return b
}

func TestPopeJoanCompartmentNames(t *testing.T) {
	// 盤は 8 区画。issue の「8 種の絵札 + ポープ」ではない。
	if PopeJoanCompartmentCount != 8 {
		t.Fatalf("PopeJoanCompartmentCount = %d, want 8", PopeJoanCompartmentCount)
	}
	want := []string{"ace", "king", "queen", "jack", "game", "pope", "matrimony", "intrigue"}
	for i, w := range want {
		if got := PopeJoanCompartment(i).String(); got != w {
			t.Errorf("compartment %d = %q, want %q", i, got, w)
		}
	}
	if got := PopeJoanCompartment(-1).String(); got != "unknown" {
		t.Errorf("out-of-range = %q, want %q", got, "unknown")
	}
	if got := PopeJoanCompartment(PopeJoanCompartmentCount).String(); got != "unknown" {
		t.Errorf("out-of-range = %q, want %q", got, "unknown")
	}
}

// TestPopeJoanDressIsFixedAtFifteen pins the ante the issue describes as
// something players distribute: it is the dealer's, and the split is fixed.
func TestPopeJoanDressIsFixedAtFifteen(t *testing.T) {
	var b PopeJoanBoard
	if got := b.Dress(); got != PopeJoanDressTotal {
		t.Fatalf("Dress paid %d, want %d", got, PopeJoanDressTotal)
	}
	if PopeJoanDressTotal != 15 {
		t.Fatalf("PopeJoanDressTotal = %d, want 15", PopeJoanDressTotal)
	}
	if got := b.Get(PopeJoanPope); got != 6 {
		t.Errorf("pope = %d, want 6", got)
	}
	for _, c := range []PopeJoanCompartment{PopeJoanMatrimony, PopeJoanIntrigue} {
		if got := b.Get(c); got != 2 {
			t.Errorf("%s = %d, want 2", c, got)
		}
	}
	for _, c := range []PopeJoanCompartment{PopeJoanAce, PopeJoanKing, PopeJoanQueen, PopeJoanJack, PopeJoanGame} {
		if got := b.Get(c); got != 1 {
			t.Errorf("%s = %d, want 1", c, got)
		}
	}

	// **取られなかった区画は持ち越す。**
	b.Dress()
	if got := b.Get(PopeJoanPope); got != 12 {
		t.Errorf("pope = %d, want it carried over to 12", got)
	}
	if got := b.Take(PopeJoanPope); got != 12 {
		t.Errorf("Take returned %d, want 12", got)
	}
	if got := b.Get(PopeJoanPope); got != 0 {
		t.Errorf("pope = %d, want it emptied", got)
	}

	// 範囲外はどれも無害に振る舞う。
	if got := b.Take(PopeJoanCompartment(99)); got != 0 {
		t.Errorf("Take of an unknown compartment = %d, want 0", got)
	}
	if got := b.Get(PopeJoanCompartment(99)); got != 0 {
		t.Errorf("Get of an unknown compartment = %d, want 0", got)
	}
}

func TestPopeJoanCompartmentForRank(t *testing.T) {
	for rank, want := range map[int]PopeJoanCompartment{
		1: PopeJoanAce, 13: PopeJoanKing, 12: PopeJoanQueen, 11: PopeJoanJack,
	} {
		got, ok := PopeJoanCompartmentForRank(rank)
		if !ok || got != want {
			t.Errorf("rank %d = %v/%v, want %v/true", rank, got, ok, want)
		}
	}
	// 10 以下は単札区画を持たない。9 は Pope だが、それはスートで決まる。
	for _, rank := range []int{2, 9, 10} {
		if _, ok := PopeJoanCompartmentForRank(rank); ok {
			t.Errorf("rank %d should have no compartment", rank)
		}
	}
}

// TestPopeJoanDeckDropsTheEightOfDiamonds is the whole mechanism of the game:
// nothing follows the seven of diamonds, so a run always stops there.
func TestPopeJoanDeckDropsTheEightOfDiamonds(t *testing.T) {
	deck := newPopeJoanDeck()
	if len(deck) != PopeJoanDeckSize {
		t.Fatalf("deck holds %d cards, want %d", len(deck), PopeJoanDeckSize)
	}
	if PopeJoanDeckSize != 51 {
		t.Fatalf("PopeJoanDeckSize = %d, want 51", PopeJoanDeckSize)
	}
	for _, c := range deck {
		if c.GetDesign() == CardDesignDiamond && c.GetValue() == 8 {
			t.Fatal("the eight of diamonds must not be in the pack")
		}
	}
	// ♦9 (Pope) は残っている。
	found := false
	for _, c := range deck {
		if popeJoanIsPope(c) {
			found = true
		}
	}
	if !found {
		t.Error("the Pope (nine of diamonds) must be in the pack")
	}
}

// TestPopeJoanDealsADeadHandAndTurnsTrump covers the rule the issue omits.
func TestPopeJoanDealsADeadHandAndTurnsTrump(t *testing.T) {
	p := NewDefaultPopeJoan()
	p.Reset()

	dealt := 0
	for _, pl := range p.GetPlayers() {
		dealt += pl.GetCardsSize()
	}
	// **プレイヤー数より 1 つ多く配る**ので、1 人分は誰の手にも入らない。
	if dealt >= PopeJoanDeckSize {
		t.Fatalf("players hold %d of %d cards; a dead hand must be set aside", dealt, PopeJoanDeckSize)
	}
	turnUp := p.GetTurnUp()
	if turnUp == nil {
		t.Fatal("a card must be turned for trump")
	}
	if p.GetTrumpSuit() != turnUp.GetDesign() {
		t.Errorf("trump = %d, want the turn-up's suit %d", p.GetTrumpSuit(), turnUp.GetDesign())
	}
	// ディーラーが 15 払っているので、盤には少なくとも 15 が乗っている
	// (めくり札で即取りされたぶんを除く)。
	total := 0
	for i := range PopeJoanCompartmentCount {
		total += p.GetBoard().Get(PopeJoanCompartment(i))
	}
	taken := 0
	for _, a := range p.GetAwards() {
		taken += a.Chips
	}
	if total+taken != PopeJoanDressTotal {
		t.Errorf("board holds %d and %d was awarded, want %d in total", total, taken, PopeJoanDressTotal)
	}
}

// TestPopeJoanTurnUpPaysTheDealer covers the other half of that rule: a
// turn-up of Pope/A/K/Q/J hands the dealer that compartment outright.
func TestPopeJoanTurnUpPaysTheDealer(t *testing.T) {
	for name, tc := range map[string]struct {
		turnUp *Card
		trump  int
		want   PopeJoanCompartment
		paid   bool
	}{
		"pope pays six":  {pjCard(CardDesignDiamond, PopeJoanPopeRank), CardDesignDiamond, PopeJoanPope, true},
		"trump king":     {pjCard(CardDesignSpade, 13), CardDesignSpade, PopeJoanKing, true},
		"trump ace":      {pjCard(CardDesignSpade, 1), CardDesignSpade, PopeJoanAce, true},
		"a plain number": {pjCard(CardDesignSpade, 5), CardDesignSpade, PopeJoanAce, false},
	} {
		t.Run(name, func(t *testing.T) {
			p := NewDefaultPopeJoan()
			p.Reset()
			p.SetBoardForTest(pjBoard())
			p.SetTrumpSuitForTest(tc.trump)
			p.SetTurnUpForTest(tc.turnUp)
			before := p.GetPlayer(0).GetChips()

			p.ResolveTurnUpForTest()

			gained := p.GetPlayer(0).GetChips() - before
			if tc.paid {
				if gained == 0 {
					t.Fatalf("the dealer should have taken the %s compartment", tc.want)
				}
				if got := p.GetBoard().Get(tc.want); got != 0 {
					t.Errorf("%s = %d, want it emptied", tc.want, got)
				}
			} else if gained != 0 {
				t.Errorf("a plain number card pays nothing, but the dealer gained %d", gained)
			}
		})
	}
}

// TestPopeJoanLeadMustBeYourLowestCard は、issue の「最小連番」ではなく
// 「自分の最も低い札ならスート自由」であることを確かめる。
func TestPopeJoanLeadMustBeYourLowestCard(t *testing.T) {
	p := pjReady(t, 0)
	p.SetRunForTest(-1, 0)
	setPjHand(p, 0, []*Card{
		pjCard(CardDesignSpade, 5), pjCard(CardDesignHeart, 3), pjCard(CardDesignClover, 9),
	})
	// 席 1 に ♥4 を持たせておく。これが無いと出した瞬間に並びが止まり、
	// runSuit が -1 に戻ってしまって「何のスートで始まったか」を確認できない。
	setPjHand(p, 1, []*Card{pjCard(CardDesignHeart, 4)})
	for i := 2; i < PopeJoanPlayerCnt; i++ {
		setPjHand(p, i, []*Card{pjCard(CardDesignDiamond, 2)})
	}

	// ♠5 は最も低くない。
	if err := p.Play(0, 0); err == nil {
		t.Error("expected an error leading a card that is not your lowest")
	}
	// ♥3 は最も低い。**スートは問われない。**
	if err := p.Play(0, 1); err != nil {
		t.Fatalf("Play the lowest card: %v", err)
	}
	if got := p.GetRunSuit(); got != CardDesignHeart {
		t.Errorf("run suit = %d, want hearts", got)
	}
	if got := p.GetCurrentPlayerIdx(); got != 1 {
		t.Errorf("current = %d, want 1 (holds the four of hearts)", got)
	}
}

func TestPopeJoanRunFollowsTheSuitUpwards(t *testing.T) {
	p := pjReady(t, 0)
	p.SetRunForTest(-1, 0)
	setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, 5), pjCard(CardDesignHeart, 9)})
	setPjHand(p, 1, []*Card{pjCard(CardDesignSpade, 6), pjCard(CardDesignHeart, 2)})
	setPjHand(p, 2, []*Card{pjCard(CardDesignClover, 4)})
	setPjHand(p, 3, []*Card{pjCard(CardDesignDiamond, 3)})

	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♠5: %v", err)
	}
	if got := p.GetCurrentPlayerIdx(); got != 1 {
		t.Fatalf("current = %d, want 1 (holds ♠6)", got)
	}
	// ♥2 は続きではない。
	if err := p.Play(1, 1); err == nil {
		t.Error("expected an error playing another suit mid-run")
	}
	if err := p.Play(1, 0); err != nil {
		t.Fatalf("Play ♠6: %v", err)
	}
}

// TestPopeJoanStopsAtTheSevenOfDiamonds is the ♦8 mechanism in action.
func TestPopeJoanStopsAtTheSevenOfDiamonds(t *testing.T) {
	p := pjReady(t, 0)
	p.SetRunForTest(CardDesignDiamond, 6)
	setPjHand(p, 0, []*Card{pjCard(CardDesignDiamond, 7), pjCard(CardDesignHeart, 2)})
	for i := 1; i < PopeJoanPlayerCnt; i++ {
		setPjHand(p, i, []*Card{pjCard(CardDesignDiamond, PopeJoanPopeRank)})
	}

	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♦7: %v", err)
	}
	// ♦8 が無いので誰も続けられない。**最後に出した人が再開する。**
	if got := p.GetRunSuit(); got != -1 {
		t.Errorf("run suit = %d, want -1: nothing follows the seven of diamonds", got)
	}
	if got := p.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want 0, who played the last card", got)
	}
}

// TestPopeJoanStopsAtAKing covers the other stop the issue does not mention:
// a king is the highest card, so nothing can follow it either.
func TestPopeJoanStopsAtAKing(t *testing.T) {
	p := pjReady(t, 0)
	p.SetRunForTest(CardDesignSpade, 12)
	setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, 13), pjCard(CardDesignHeart, 2)})
	for i := 1; i < PopeJoanPlayerCnt; i++ {
		setPjHand(p, i, []*Card{pjCard(CardDesignClover, 4)})
	}

	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♠K: %v", err)
	}
	if got := p.GetRunSuit(); got != -1 {
		t.Errorf("run suit = %d, want -1: nothing outranks a king", got)
	}
	if got := p.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want 0", got)
	}
}

// TestPopeJoanCompartmentsNeedTheTrumpSuit は、issue の「該当する札を出した際は」
// では足りず、トランプでなければ払われないことを確かめる。
func TestPopeJoanCompartmentsNeedTheTrumpSuit(t *testing.T) {
	p := pjReady(t, 0)
	p.SetTrumpSuitForTest(CardDesignSpade)
	p.SetBoardForTest(pjBoard())
	p.SetRunForTest(CardDesignHeart, 12)
	setPjHand(p, 0, []*Card{pjCard(CardDesignHeart, 13), pjCard(CardDesignClover, 2)})
	for i := 1; i < PopeJoanPlayerCnt; i++ {
		setPjHand(p, i, []*Card{pjCard(CardDesignClover, 4)})
	}

	before := p.GetPlayer(0).GetChips()
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♥K: %v", err)
	}
	if got := p.GetPlayer(0).GetChips(); got != before {
		t.Errorf("an off-trump king paid %d; it should pay nothing", got-before)
	}
	if got := p.GetBoard().Get(PopeJoanKing); got != 1 {
		t.Errorf("king compartment = %d, want it left standing", got)
	}
}

// TestPopeJoanPopePaysRegardlessOfTrump は、Pope が ♦9 という札そのもので
// あってトランプに依らないことを確かめる。
func TestPopeJoanPopePaysRegardlessOfTrump(t *testing.T) {
	p := pjReady(t, 0)
	p.SetTrumpSuitForTest(CardDesignSpade)
	p.SetBoardForTest(pjBoard())
	p.SetRunForTest(CardDesignDiamond, 8)
	setPjHand(p, 0, []*Card{pjCard(CardDesignDiamond, PopeJoanPopeRank), pjCard(CardDesignClover, 2)})
	for i := 1; i < PopeJoanPlayerCnt; i++ {
		setPjHand(p, i, []*Card{pjCard(CardDesignClover, 4)})
	}

	before := p.GetPlayer(0).GetChips()
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play the Pope: %v", err)
	}
	if got := p.GetPlayer(0).GetChips() - before; got != 6 {
		t.Errorf("the Pope paid %d, want 6", got)
	}
}

// TestPopeJoanMatrimonyNeedsBothFromOneHand は、2 枚組の区画が「同じ人が両方」
// でしか払われないことを確かめる。
func TestPopeJoanMatrimonyNeedsBothFromOneHand(t *testing.T) {
	p := pjReady(t, 0)
	p.SetTrumpSuitForTest(CardDesignSpade)
	p.SetBoardForTest(pjBoard())

	// 席 0 がトランプの Q を出す。まだ Matrimony は動かない。
	p.SetRunForTest(CardDesignSpade, 11)
	setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, 12), pjCard(CardDesignClover, 2)})
	setPjHand(p, 1, []*Card{pjCard(CardDesignSpade, 13), pjCard(CardDesignClover, 3)})
	for i := 2; i < PopeJoanPlayerCnt; i++ {
		setPjHand(p, i, []*Card{pjCard(CardDesignHeart, 4)})
	}
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♠Q: %v", err)
	}
	if got := p.GetBoard().Get(PopeJoanMatrimony); got != 2 {
		t.Errorf("matrimony = %d, want it left standing after only the queen", got)
	}

	// **別の人**が K を出しても Matrimony は動かない。
	if got := p.GetCurrentPlayerIdx(); got != 1 {
		t.Fatalf("current = %d, want 1", got)
	}
	if err := p.Play(1, 0); err != nil {
		t.Fatalf("Play ♠K: %v", err)
	}
	if got := p.GetBoard().Get(PopeJoanMatrimony); got != 2 {
		t.Errorf("matrimony = %d, want it left standing when K and Q come from different hands", got)
	}
}

func TestPopeJoanMatrimonyPaysWhenOneHandPlaysBoth(t *testing.T) {
	p := pjReady(t, 0)
	p.SetTrumpSuitForTest(CardDesignSpade)
	p.SetBoardForTest(pjBoard())
	p.SetRunForTest(CardDesignSpade, 11)
	setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, 12), pjCard(CardDesignSpade, 13), pjCard(CardDesignClover, 2)})
	for i := 1; i < PopeJoanPlayerCnt; i++ {
		setPjHand(p, i, []*Card{pjCard(CardDesignHeart, 4)})
	}

	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♠Q: %v", err)
	}
	// 誰も ♠K を持っていないので手番は 0 に戻る。
	if got := p.GetCurrentPlayerIdx(); got != 0 {
		t.Fatalf("current = %d, want 0", got)
	}
	p.SetRunForTest(CardDesignSpade, 12)
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♠K: %v", err)
	}
	if got := p.GetBoard().Get(PopeJoanMatrimony); got != 0 {
		t.Errorf("matrimony = %d, want it taken by the hand that played both", got)
	}
}

// TestPopeJoanGoingOutExcusesThePopeHolder covers the payment rule the issue
// states without its exception.
func TestPopeJoanGoingOutExcusesThePopeHolder(t *testing.T) {
	p := pjReady(t, 0)
	p.SetBoardForTest(pjBoard())
	p.SetRunForTest(-1, 0)
	setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, 2)})
	// 席 1 は 2 枚残す。Pope は持っていないので払う。
	setPjHand(p, 1, []*Card{pjCard(CardDesignHeart, 5), pjCard(CardDesignHeart, 6)})
	// 席 2 は Pope を抱えている。**免除。**
	setPjHand(p, 2, []*Card{pjCard(CardDesignDiamond, PopeJoanPopeRank), pjCard(CardDesignClover, 4)})
	setPjHand(p, 3, []*Card{pjCard(CardDesignClover, 7)})

	before1 := p.GetPlayer(1).GetChips()
	before2 := p.GetPlayer(2).GetChips()
	before3 := p.GetPlayer(3).GetChips()
	beforeWin := p.GetPlayer(0).GetChips()

	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play: %v", err)
	}
	if p.GetPhase() != PopeJoanPhaseDealEnd {
		t.Fatalf("phase = %v, want DealEnd", p.GetPhase())
	}
	if got := p.GetPlayer(1).GetChips(); got != before1-2 {
		t.Errorf("seat 1 chips = %d, want %d (two cards left)", got, before1-2)
	}
	if got := p.GetPlayer(2).GetChips(); got != before2 {
		t.Errorf("seat 2 paid %d; the Pope holder is excused", before2-got)
	}
	if got := p.GetPlayer(3).GetChips(); got != before3-1 {
		t.Errorf("seat 3 chips = %d, want %d", got, before3-1)
	}
	// Game 区画 1 + 席 1 の 2 + 席 3 の 1 = 4。
	if got := p.GetPlayer(0).GetChips(); got != beforeWin+1+3 {
		t.Errorf("winner chips = %d, want %d", got, beforeWin+1+3)
	}
}

func TestPopeJoanPlayRejections(t *testing.T) {
	p := pjReady(t, 0)
	setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, 2)})
	if err := p.Play(1, 0); err == nil {
		t.Error("expected an error playing out of turn")
	}
	if err := p.Play(0, 9); err == nil {
		t.Error("expected an error for an out-of-range hand index")
	}
	p.SetPhaseForTest(PopeJoanPhaseDealEnd)
	if err := p.Play(0, 0); err == nil {
		t.Error("expected an error playing outside the play phase")
	}
}

func TestPopeJoanNextDealAndGameEnd(t *testing.T) {
	p := NewDefaultPopeJoan()
	p.Reset()
	if err := p.NextDeal(); err == nil {
		t.Error("expected an error while the deal is still live")
	}

	p.SetPhaseForTest(PopeJoanPhaseDealEnd)
	if err := p.NextDeal(); err != nil {
		t.Fatalf("NextDeal: %v", err)
	}
	if p.GetPhase() != PopeJoanPhasePlay {
		t.Errorf("phase = %v, want Play", p.GetPhase())
	}
	if got := p.GetCurrentPlayerIdx(); got != 2 {
		t.Errorf("current = %d, want 2", got)
	}

	p.SetDealNumberForTest(p.GetConfig().TargetDeals - 1)
	p.SetRunForTest(-1, 0)
	for i := range PopeJoanPlayerCnt {
		p.GetPlayer(i).Reset()
	}
	setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, 2)})
	p.SetCurrentPlayerForTest(0)
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play: %v", err)
	}
	if !p.GetGameEndFlag() {
		t.Fatal("the game should be over after the last deal")
	}
	if idx := p.GetWinnerIdx(); idx < 0 || idx >= PopeJoanPlayerCnt {
		t.Errorf("winnerIdx = %d, want a real seat", idx)
	}
	if err := p.NextDeal(); err == nil {
		t.Error("expected an error dealing after the game is over")
	}
	if err := p.Play(0, 0); err == nil {
		t.Error("expected an error playing after the game is over")
	}
}

func TestPopeJoanCpuDecide(t *testing.T) {
	p := pjReady(t, 1)
	p.SetRunForTest(CardDesignSpade, 6)
	setPjHand(p, 1, []*Card{pjCard(CardDesignHeart, 2), pjCard(CardDesignSpade, 7)})
	idx := p.PopeJoanCpuDecide(1)
	if idx != 1 {
		t.Fatalf("the CPU should play the only legal continuation, got %d", idx)
	}
	if err := p.Play(1, idx); err != nil {
		t.Fatalf("the CPU's play was rejected: %v", err)
	}

	// 並びが止まっていれば最も低い札しか出せない。
	p.SetRunForTest(-1, 0)
	p.SetCurrentPlayerForTest(1)
	setPjHand(p, 1, []*Card{pjCard(CardDesignHeart, 9), pjCard(CardDesignClover, 3)})
	idx = p.PopeJoanCpuDecide(1)
	if got := p.GetPlayer(1).GetCard(idx).GetValue(); got != 3 {
		t.Errorf("the CPU led a %d, want its lowest (3)", got)
	}

	if p.PopeJoanCpuDecide(9) != -1 {
		t.Error("an unknown seat must not produce a move")
	}
}

// TestPopeJoanCpuDrivesDealsToAnEnd checks that four CPUs always finish a deal.
func TestPopeJoanCpuDrivesDealsToAnEnd(t *testing.T) {
	for trial := range 30 {
		p := NewDefaultPopeJoan()
		p.Reset()
		for range 2000 {
			if p.GetPhase() != PopeJoanPhasePlay {
				break
			}
			cur := p.GetCurrentPlayerIdx()
			idx := p.PopeJoanCpuDecide(cur)
			if idx < 0 {
				t.Fatalf("trial %d: seat %d had nothing playable", trial, cur)
			}
			if err := p.Play(cur, idx); err != nil {
				t.Fatalf("trial %d: %v", trial, err)
			}
		}
		if ph := p.GetPhase(); ph != PopeJoanPhaseDealEnd && ph != PopeJoanPhaseGameEnd {
			t.Fatalf("trial %d: the deal never ended (phase = %v)", trial, ph)
		}
		if p.GetDealWinner() < 0 {
			t.Fatalf("trial %d: nobody went out", trial)
		}
	}
}

func TestPopeJoanActionLog(t *testing.T) {
	p := NewDefaultPopeJoan()
	p.Reset()
	log := p.GetActionLog()
	if len(log) == 0 {
		t.Fatal("the deal should be logged")
	}
	if log[0].ActionType != "deal" {
		t.Errorf("first entry = %q, want a deal", log[0].ActionType)
	}
	last := log[len(log)-1]
	if last.TurnNumber != len(log) {
		t.Errorf("TurnNumber = %d, want %d", last.TurnNumber, len(log))
	}
}

func TestPopeJoanConfigValidate(t *testing.T) {
	if err := DefaultPopeJoanConfig().Validate(); err != nil {
		t.Fatalf("the default config should validate: %v", err)
	}
	if err := (PopeJoanConfig{CpuDifficulty: 5, TargetDeals: 5}).Validate(); err == nil {
		t.Error("expected an error for an unknown difficulty")
	}
	if err := (PopeJoanConfig{TargetDeals: 0}).Validate(); err == nil {
		t.Error("expected an error for zero deals")
	}
	p := NewDefaultPopeJoan()
	p.SetConfig(PopeJoanConfig{TargetDeals: 3})
	if p.GetConfig().TargetDeals != 3 {
		t.Error("SetConfig/GetConfig disagree")
	}
}

func TestPopeJoanAccessorBounds(t *testing.T) {
	p := NewDefaultPopeJoan()
	p.Reset()
	if p.GetPlayer(-1) != nil || p.GetPlayer(PopeJoanPlayerCnt) != nil {
		t.Error("GetPlayer should return nil outside the table")
	}
	if got := p.GetDealNumber(); got != 0 {
		t.Errorf("GetDealNumber = %d, want 0", got)
	}
}

func TestPopeJoanJSONRoundTrip(t *testing.T) {
	p := NewDefaultPopeJoan()
	p.Reset()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got PopeJoan
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// **盤のチップは持ち越しそのもの。**
	for i := range PopeJoanCompartmentCount {
		c := PopeJoanCompartment(i)
		if got.GetBoard().Get(c) != p.GetBoard().Get(c) {
			t.Errorf("%s = %d, want %d", c, got.GetBoard().Get(c), p.GetBoard().Get(c))
		}
	}
	if got.GetTrumpSuit() != p.GetTrumpSuit() {
		t.Errorf("trump = %d, want %d", got.GetTrumpSuit(), p.GetTrumpSuit())
	}
	if len(got.GetActionLog()) != len(p.GetActionLog()) {
		t.Errorf("action log = %d entries, want %d", len(got.GetActionLog()), len(p.GetActionLog()))
	}
}

func TestPopeJoanUnmarshalRejectsGarbage(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[null,null],"cfg":{"cd":0,"td":5},"ph":0}`},
		{"bad config", `{"pl":[{},{},{},{}],"cfg":{"cd":9,"td":5},"ph":0}`},
		{"unknown phase", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":9}`},
		{"negative phase", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p PopeJoan
			if err := json.Unmarshal([]byte(tt.data), &p); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestPopeJoanUnmarshalClampsIndices(t *testing.T) {
	var p PopeJoan
	data := `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":0,"cur":99,"dl":-3,` +
		`"kb":42,"qb":42,"dw":42,"wi":42,"rs":99,` +
		`"aw":[null,{"Compartment":99,"Player":0,"Chips":1},{"Compartment":0,"Player":99,"Chips":1},` +
		`{"Compartment":5,"Player":1,"Chips":6}]}`
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := p.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want it clamped to 0", got)
	}
	if got := p.GetWinnerIdx(); got != -1 {
		t.Errorf("winnerIdx = %d, want -1", got)
	}
	if got := p.GetDealWinner(); got != -1 {
		t.Errorf("dealWinner = %d, want -1", got)
	}
	// **-1 は「好きな札で始められる」という意味を持つ。**
	if got := p.GetRunSuit(); got != -1 {
		t.Errorf("runSuit = %d, want it reset to -1", got)
	}
	if got := len(p.GetAwards()); got != 1 {
		t.Errorf("awards = %d, want the malformed ones dropped", got)
	}
}

// TestPopeJoanUnmarshalKeepsALiveRun は、-1 を潰さないことと生きた並びを
// 保つことの両方を確かめる。
func TestPopeJoanUnmarshalKeepsALiveRun(t *testing.T) {
	var live PopeJoan
	data := `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":0,"rs":1,"rr":8}`
	if err := json.Unmarshal([]byte(data), &live); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := live.GetRunSuit(); got != 1 {
		t.Errorf("runSuit = %d, want 1 preserved", got)
	}
	if got := live.GetRunRank(); got != 8 {
		t.Errorf("runRank = %d, want 8", got)
	}
}

// TestPopeJoanDealRotatesTheOddCard covers the fairness bug the review caught:
// 51 cards over 5 hands leaves one seat with an extra card, and dealing always
// from seat 0 would park that on the human every single deal.
//
// **Holding more cards is a pure disadvantage here** -- going out first
// collects from everyone else, and every card left costs a chip.
func TestPopeJoanDealRotatesTheOddCard(t *testing.T) {
	p := NewDefaultPopeJoan()
	p.Reset()

	longest := func() int {
		seat, best := -1, -1
		for i, pl := range p.GetPlayers() {
			if pl.GetCardsSize() > best {
				seat, best = i, pl.GetCardsSize()
			}
		}
		return seat
	}

	seen := map[int]bool{}
	for range PopeJoanPlayerCnt {
		seen[longest()] = true
		p.SetPhaseForTest(PopeJoanPhaseDealEnd)
		if err := p.NextDeal(); err != nil {
			t.Fatalf("NextDeal: %v", err)
		}
	}
	if len(seen) < 2 {
		t.Fatalf("the extra card landed on seat %v every deal; it must rotate", seen)
	}
	if len(seen) == 1 && seen[0] {
		t.Fatal("the human seat took the extra card every deal")
	}
}

// TestPopeJoanIntrigueScoresInEitherOrder covers the second review finding.
//
// **A run only ever climbs**, so J (11) is played before Q (12) in the natural
// case. Checking only "queen already played" meant the common order never paid.
func TestPopeJoanIntrigueScoresInEitherOrder(t *testing.T) {
	for name, jackFirst := range map[string]bool{
		"jack then queen": true,
		"queen then jack": false,
	} {
		t.Run(name, func(t *testing.T) {
			p := pjReady(t, 0)
			p.SetTrumpSuitForTest(CardDesignSpade)
			p.SetBoardForTest(pjBoard())
			for i := 1; i < PopeJoanPlayerCnt; i++ {
				setPjHand(p, i, []*Card{pjCard(CardDesignHeart, 4)})
			}

			first, second := 11, 12
			if !jackFirst {
				first, second = 12, 11
			}
			setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, first), pjCard(CardDesignSpade, second)})

			p.SetRunForTest(CardDesignSpade, first-1)
			if err := p.Play(0, 0); err != nil {
				t.Fatalf("play the first court card: %v", err)
			}
			p.SetCurrentPlayerForTest(0)
			p.SetRunForTest(CardDesignSpade, second-1)
			if err := p.Play(0, 0); err != nil {
				t.Fatalf("play the second court card: %v", err)
			}

			if got := p.GetBoard().Get(PopeJoanIntrigue); got != 0 {
				t.Errorf("intrigue = %d, want it taken in the %s order", got, name)
			}
		})
	}
}

// TestPopeJoanIntrigueNeedsBothFromOneHand は逆方向 -- 別の人が出したら払わない。
func TestPopeJoanIntrigueNeedsBothFromOneHand(t *testing.T) {
	p := pjReady(t, 0)
	p.SetTrumpSuitForTest(CardDesignSpade)
	p.SetBoardForTest(pjBoard())
	// 席 0 に 2 枚持たせる。1 枚だと出した瞬間に上がってディールが終わる。
	setPjHand(p, 0, []*Card{pjCard(CardDesignSpade, 11), pjCard(CardDesignHeart, 2)})
	setPjHand(p, 1, []*Card{pjCard(CardDesignSpade, 12), pjCard(CardDesignHeart, 3)})
	for i := 2; i < PopeJoanPlayerCnt; i++ {
		setPjHand(p, i, []*Card{pjCard(CardDesignHeart, 4)})
	}

	p.SetRunForTest(CardDesignSpade, 10)
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♠J: %v", err)
	}
	p.SetCurrentPlayerForTest(1)
	p.SetRunForTest(CardDesignSpade, 11)
	if err := p.Play(1, 0); err != nil {
		t.Fatalf("Play ♠Q: %v", err)
	}
	if got := p.GetBoard().Get(PopeJoanIntrigue); got != 2 {
		t.Errorf("intrigue = %d, want it left standing when J and Q come from different hands", got)
	}
}

// **並びに従う義務がある。**さらに自由リードでも「自分の最も低い札」に限られる
// ので、全部出せるわけではない (#4934)。
func TestPopeJoan_PopeJoanValidPlays(t *testing.T) {
	g := NewDefaultPopeJoan()
	g.Reset()
	g.SetCurrentPlayerForTest(0)

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(pjCard(CardDesignSpade, 5))
	p.AddCard(pjCard(CardDesignHeart, 9))
	p.AddCard(pjCard(CardDesignSpade, 6))

	// **自由リードでも最も低い札だけ。**issue は「全部選択可能」と書いているが、
	// ドメインは新しい並びを最低札で始めることを要求する。
	g.SetRunForTest(-1, 0)
	if got := g.PopeJoanValidPlays(0); len(got) != 1 || got[0] != 0 {
		t.Fatalf("a new run must be led with the lowest card, got %v", got)
	}

	// ♠5 の次は ♠6 だけ。♥9 はスート違い。
	g.SetRunForTest(CardDesignSpade, popeJoanRankOrder(5))
	if got := g.PopeJoanValidPlays(0); len(got) != 1 || got[0] != 2 {
		t.Fatalf("only the next higher card of the run suit is legal, got %v", got)
	}

	// 続けられる札が無ければ空。
	g.SetRunForTest(CardDesignClover, popeJoanRankOrder(2))
	if got := g.PopeJoanValidPlays(0); len(got) != 0 {
		t.Fatalf("nothing should be playable, got %v", got)
	}

	// 手番でなければ nil。
	if got := g.PopeJoanValidPlays(1); got != nil {
		t.Fatalf("off-turn must be nil, got %v", got)
	}
}

// #5723: Pope (♦9) を抱えている人はその区画への支払いを免除される。表示側 (CUI/Web)
// がこの判定を共有するので、判定そのものをここで固定する。
func TestPopeJoanHoldsPope(t *testing.T) {
	card := func(design, value int) *Card { return NewCard(design, value, false) }

	holder := NewPopeJoanPlayer(false)
	holder.AddCard(card(CardDesignSpade, 3))
	holder.AddCard(card(CardDesignDiamond, PopeJoanPopeRank))
	assert.True(t, PopeJoanHoldsPope(holder))

	// **同じランクでもスートが違えば Pope ではない。**
	sameRank := NewPopeJoanPlayer(false)
	sameRank.AddCard(card(CardDesignSpade, PopeJoanPopeRank))
	assert.False(t, PopeJoanHoldsPope(sameRank))

	// 同じスートでもランクが違えば Pope ではない。
	sameSuit := NewPopeJoanPlayer(false)
	sameSuit.AddCard(card(CardDesignDiamond, PopeJoanPopeRank+1))
	assert.False(t, PopeJoanHoldsPope(sameSuit))

	assert.False(t, PopeJoanHoldsPope(NewPopeJoanPlayer(true)), "an empty hand holds nothing")
	assert.False(t, PopeJoanHoldsPope(nil))
}
