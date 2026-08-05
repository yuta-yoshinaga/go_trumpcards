//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"
)

func pcCard(design, value int) *Card { return NewCard(design, value, true) }

// pcReady puts a fresh game into the given phase for the given seat.
func pcReady(t *testing.T, phase PochPhase, seat int) *Poch {
	t.Helper()
	p := NewDefaultPoch()
	p.Reset()
	p.SetPhaseForTest(phase)
	p.SetCurrentPlayerForTest(seat)
	return p
}

// setPcHand replaces a seat's hand outright.
func setPcHand(p *Poch, seat int, cards []*Card) {
	pl := p.GetPlayer(seat)
	pl.Reset()
	for _, c := range cards {
		pl.AddCard(c)
	}
}

func TestPochPoolNames(t *testing.T) {
	// 盤は 9 区画で固定。issue の「各ランク + ペア枠」では足りない。
	if PochPoolCount != 9 {
		t.Fatalf("PochPoolCount = %d, want 9", PochPoolCount)
	}
	want := []string{"ace", "king", "queen", "jack", "ten", "marriage", "sequence", "pocher", "centre"}
	for i, w := range want {
		if got := PochPool(i).String(); got != w {
			t.Errorf("pool %d = %q, want %q", i, got, w)
		}
	}
	if got := PochPool(-1).String(); got != "unknown" {
		t.Errorf("out-of-range pool = %q, want %q", got, "unknown")
	}
	if got := PochPool(PochPoolCount).String(); got != "unknown" {
		t.Errorf("out-of-range pool = %q, want %q", got, "unknown")
	}
}

func TestPochBoard(t *testing.T) {
	var b PochBoard
	b.Ante(4)
	for i := range PochPoolCount {
		if got := b.Get(PochPool(i)); got != 4 {
			t.Fatalf("pool %d = %d, want 4", i, got)
		}
	}
	// **取られなかったプールは持ち越す。**
	b.Ante(4)
	if got := b.Get(PochPoolMarriage); got != 8 {
		t.Errorf("marriage = %d, want it carried over to 8", got)
	}
	if got := b.Take(PochPoolMarriage); got != 8 {
		t.Errorf("Take returned %d, want 8", got)
	}
	if got := b.Get(PochPoolMarriage); got != 0 {
		t.Errorf("marriage = %d, want it emptied", got)
	}
	b.Add(PochPoolMarriage, 3)
	if got := b.Get(PochPoolMarriage); got != 3 {
		t.Errorf("marriage = %d, want 3", got)
	}

	// 範囲外はどれも無害に振る舞う。
	b.Add(PochPool(-1), 5)
	b.Add(PochPoolAce, -5)
	if got := b.Take(PochPool(99)); got != 0 {
		t.Errorf("Take of an unknown pool = %d, want 0", got)
	}
	if got := b.Get(PochPool(99)); got != 0 {
		t.Errorf("Get of an unknown pool = %d, want 0", got)
	}
}

func TestPochBestCombo(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  PochCombo
	}{
		{
			name:  "no pair is no combo",
			cards: []*Card{pcCard(CardDesignSpade, 7), pcCard(CardDesignHeart, 8)},
			want:  PochCombo{},
		},
		{
			name:  "a pair",
			cards: []*Card{pcCard(CardDesignSpade, 9), pcCard(CardDesignHeart, 9)},
			want:  PochCombo{Size: 2, Rank: 9},
		},
		{
			// **枚数が先。**三枚は、ランクがどれだけ高いペアにも勝つ。
			name: "three beats a higher pair",
			cards: []*Card{
				pcCard(CardDesignSpade, 7), pcCard(CardDesignHeart, 7), pcCard(CardDesignClover, 7),
				pcCard(CardDesignSpade, 1), pcCard(CardDesignHeart, 1),
			},
			want: PochCombo{Size: 3, Rank: 7},
		},
		{
			name: "four beats three",
			cards: []*Card{
				pcCard(CardDesignSpade, 8), pcCard(CardDesignHeart, 8),
				pcCard(CardDesignClover, 8), pcCard(CardDesignDiamond, 8),
				pcCard(CardDesignSpade, 13), pcCard(CardDesignHeart, 13), pcCard(CardDesignClover, 13),
			},
			want: PochCombo{Size: 4, Rank: 8},
		},
		{
			// A は最上位。同じ枚数ならランクで決まる。
			name: "the ace outranks the king at the same size",
			cards: []*Card{
				pcCard(CardDesignSpade, 1), pcCard(CardDesignHeart, 1),
				pcCard(CardDesignSpade, 13), pcCard(CardDesignHeart, 13),
			},
			want: PochCombo{Size: 2, Rank: 1},
		},
		{
			name:  "an empty card is ignored",
			cards: []*Card{pcCard(CardDesignSpade, 9), nil, pcCard(CardDesignHeart, 9)},
			want:  PochCombo{Size: 2, Rank: 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PochBestCombo(tt.cards); got != tt.want {
				t.Errorf("combo = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPochComboBeats(t *testing.T) {
	three := PochCombo{Size: 3, Rank: 7}
	pair := PochCombo{Size: 2, Rank: 14}
	if !three.Beats(pair) {
		t.Error("three of a kind must beat any pair")
	}
	if pair.Beats(three) {
		t.Error("a pair must not beat three of a kind")
	}
	if !(PochCombo{Size: 2, Rank: 14}).Beats(PochCombo{Size: 2, Rank: 13}) {
		t.Error("at the same size the higher rank wins")
	}
}

// TestPochDealLeavesOneCardUp pins the deal: cards go round one at a time until
// exactly one is left, and that card's suit is the pay suit.
func TestPochDealLeavesOneCardUp(t *testing.T) {
	p := NewDefaultPoch()
	p.Reset()

	total := 0
	for _, pl := range p.GetPlayers() {
		total += pl.GetCardsSize()
	}
	if total != PochDeckSize-1 {
		t.Fatalf("dealt %d cards, want %d with one left face up", total, PochDeckSize-1)
	}
	turnUp := p.GetTurnUp()
	if turnUp == nil {
		t.Fatal("there must be a card face up")
	}
	if p.GetPaySuit() != turnUp.GetDesign() {
		t.Errorf("pay suit = %d, want the turn-up's suit %d", p.GetPaySuit(), turnUp.GetDesign())
	}
}

// TestPochAnteFillsEveryPool は、区画に札ではなくチップを置くことを確かめる。
func TestPochAnteFillsEveryPool(t *testing.T) {
	p := NewDefaultPoch()
	p.Reset()
	board := p.GetBoard()
	// 第 1 段階が済んでいるので、取られたプールは空になっている。取られていない
	// ものは 4 (人数ぶん) のまま。
	filled := 0
	for i := range PochPoolCount {
		if board.Get(PochPool(i)) == PochPlayerCnt {
			filled++
		}
	}
	if filled == 0 {
		t.Fatal("no pool holds the ante; chips are not being placed")
	}
	// centre と pocher は第 1 段階では取られない。
	if got := board.Get(PochPoolCentre); got != PochPlayerCnt {
		t.Errorf("centre = %d, want %d", got, PochPlayerCnt)
	}
	if got := board.Get(PochPoolPocher); got != PochPlayerCnt {
		t.Errorf("pocher = %d, want %d", got, PochPlayerCnt)
	}
}

// TestPochStakingNeedsThePaySuit is the correction that matters most in stage
// one: the rank alone does not pay.
func TestPochStakingNeedsThePaySuit(t *testing.T) {
	p := NewDefaultPoch()
	p.Reset()
	p.SetPaySuitForTest(CardDesignSpade)
	var board PochBoard
	board.Ante(PochPlayerCnt)
	p.SetBoardForTest(board)
	for i := range PochPlayerCnt {
		p.GetPlayer(i).Reset()
	}
	// 席 0 は**別スート**の A を持つ。取れてはいけない。
	setPcHand(p, 0, []*Card{pcCard(CardDesignHeart, 1)})
	// 席 1 は pay suit の K。取れる。
	setPcHand(p, 1, []*Card{pcCard(CardDesignSpade, 13)})

	before0, before1 := p.GetPlayer(0).GetChips(), p.GetPlayer(1).GetChips()
	p.ResolveStakingForTest()

	if got := p.GetPlayer(0).GetChips(); got != before0 {
		t.Errorf("seat 0 gained %d; an off-suit ace pays nothing", got-before0)
	}
	if got := p.GetPlayer(1).GetChips(); got != before1+PochPlayerCnt {
		t.Errorf("seat 1 chips = %d, want %d from the king pool", got, before1+PochPlayerCnt)
	}
	if got := p.GetBoard().Get(PochPoolAce); got != PochPlayerCnt {
		t.Errorf("ace pool = %d, want it left standing to carry over", got)
	}
}

// TestPochMarriageAndSequenceNeedEveryCard は、Marriage と Sequence が
// 「同じ人が全部」でしか取れないことを確かめる。
func TestPochMarriageAndSequenceNeedEveryCard(t *testing.T) {
	// K と Q が別の人に割れていれば Marriage は取られない。
	split := NewDefaultPoch()
	split.Reset()
	split.SetPaySuitForTest(CardDesignSpade)
	var b PochBoard
	b.Ante(PochPlayerCnt)
	split.SetBoardForTest(b)
	for i := range PochPlayerCnt {
		split.GetPlayer(i).Reset()
	}
	setPcHand(split, 0, []*Card{pcCard(CardDesignSpade, 13)})
	setPcHand(split, 1, []*Card{pcCard(CardDesignSpade, 12)})
	split.ResolveStakingForTest()
	if got := split.GetBoard().Get(PochPoolMarriage); got != PochPlayerCnt {
		t.Errorf("marriage = %d, want it left standing when K and Q are split", got)
	}

	// 同じ人が両方持っていれば取れる。7-8-9 も三枚そろって初めて取れる。
	both := NewDefaultPoch()
	both.Reset()
	both.SetPaySuitForTest(CardDesignSpade)
	var b2 PochBoard
	b2.Ante(PochPlayerCnt)
	both.SetBoardForTest(b2)
	for i := range PochPlayerCnt {
		both.GetPlayer(i).Reset()
	}
	setPcHand(both, 0, []*Card{
		pcCard(CardDesignSpade, 13), pcCard(CardDesignSpade, 12),
		pcCard(CardDesignSpade, 7), pcCard(CardDesignSpade, 8),
	})
	both.ResolveStakingForTest()
	if got := both.GetBoard().Get(PochPoolMarriage); got != 0 {
		t.Errorf("marriage = %d, want it taken", got)
	}
	// 9 が無いので Sequence は取れない。
	if got := both.GetBoard().Get(PochPoolSequence); got != PochPlayerCnt {
		t.Errorf("sequence = %d, want it left standing without the nine", got)
	}
}

func TestPochStakingMovesToPochen(t *testing.T) {
	p := NewDefaultPoch()
	p.Reset()
	if p.GetPhase() != PochPhasePochen {
		t.Fatalf("phase = %v, want Pochen: stage one resolves itself", p.GetPhase())
	}
	// ディーラーの左隣が先手。
	if got := p.GetCurrentPlayerIdx(); got != 1 {
		t.Errorf("current = %d, want 1", got)
	}
}

// TestPochenIsAComparisonNotADeclaration は、pochen が宣言でもブラフでもなく
// 手札の組の比べ合いであることを確かめる。
func TestPochenIsAComparisonNotADeclaration(t *testing.T) {
	p := pcReady(t, PochPhasePochen, 0)
	for i := range PochPlayerCnt {
		p.GetPlayer(i).Reset()
		p.GetPlayer(i).ResetBetting()
	}
	// 席 2 だけがスリーカード。ほかはペア以下。
	setPcHand(p, 0, []*Card{pcCard(CardDesignSpade, 1), pcCard(CardDesignHeart, 1)})
	setPcHand(p, 1, []*Card{pcCard(CardDesignSpade, 8), pcCard(CardDesignHeart, 9)})
	setPcHand(p, 2, []*Card{
		pcCard(CardDesignSpade, 7), pcCard(CardDesignHeart, 7), pcCard(CardDesignClover, 7),
	})
	setPcHand(p, 3, []*Card{pcCard(CardDesignSpade, 10), pcCard(CardDesignHeart, 11)})

	var board PochBoard
	board.Add(PochPoolPocher, 12)
	p.SetBoardForTest(board)

	// 全員が 1 単位ずつ出して額が揃う。
	for i := range PochPlayerCnt {
		p.SetCurrentPlayerForTest(i)
		if err := p.Bet(i); err != nil {
			t.Fatalf("Bet seat %d: %v", i, err)
		}
	}

	// **A のペアではなく 7 のスリーカードが勝つ。**枚数が先。
	if got := p.GetPochenWinner(); got != 2 {
		t.Fatalf("pochen winner = %d, want 2 (three of a kind)", got)
	}
	// 賭け 4 + Pocher プール 12。
	if got := p.GetPochenPot(); got != PochPlayerCnt+12 {
		t.Errorf("pot = %d, want %d", got, PochPlayerCnt+12)
	}
	if got := p.GetBoard().Get(PochPoolPocher); got != 0 {
		t.Errorf("pocher pool = %d, want it emptied", got)
	}
	if p.GetPhase() != PochPhaseStops {
		t.Errorf("phase = %v, want Stops", p.GetPhase())
	}
	// **pochen を取った人が出し始める。**
	if got := p.GetCurrentPlayerIdx(); got != 2 {
		t.Errorf("current = %d, want the pochen winner 2", got)
	}
}

func TestPochenEndsWhenOnlyOneStays(t *testing.T) {
	p := pcReady(t, PochPhasePochen, 0)
	for i := range PochPlayerCnt {
		p.GetPlayer(i).ResetBetting()
	}
	for i := range PochPlayerCnt - 1 {
		p.SetCurrentPlayerForTest(i)
		if err := p.Fold(i); err != nil {
			t.Fatalf("Fold seat %d: %v", i, err)
		}
	}
	if got := p.GetPochenWinner(); got != PochPlayerCnt-1 {
		t.Errorf("pochen winner = %d, want the last player standing", got)
	}
	if p.GetPhase() != PochPhaseStops {
		t.Errorf("phase = %v, want Stops", p.GetPhase())
	}
}

func TestPochenRejections(t *testing.T) {
	p := pcReady(t, PochPhasePochen, 0)
	if err := p.Bet(1); err == nil {
		t.Error("expected an error betting out of turn")
	}
	if err := p.Fold(1); err == nil {
		t.Error("expected an error folding out of turn")
	}
	if err := p.Fold(0); err != nil {
		t.Fatalf("Fold: %v", err)
	}
	p.SetCurrentPlayerForTest(0)
	if err := p.Bet(0); err == nil {
		t.Error("expected an error betting after folding")
	}

	p.SetPhaseForTest(PochPhaseStops)
	if err := p.Bet(0); err == nil {
		t.Error("expected an error betting outside the betting stage")
	}
}

// TestPochStopsFollowsTheSuitUpwards は、ストップスが「同じスートの次に高い札」
// であることを確かめる。issue の「出せなくなったら手番終了」では足りない。
func TestPochStopsFollowsTheSuitUpwards(t *testing.T) {
	p := pcReady(t, PochPhaseStops, 0)
	for i := range PochPlayerCnt {
		p.GetPlayer(i).Reset()
	}
	setPcHand(p, 0, []*Card{pcCard(CardDesignSpade, 7), pcCard(CardDesignHeart, 13)})
	setPcHand(p, 1, []*Card{pcCard(CardDesignSpade, 8), pcCard(CardDesignHeart, 7)})
	setPcHand(p, 2, []*Card{pcCard(CardDesignClover, 9)})
	setPcHand(p, 3, []*Card{pcCard(CardDesignDiamond, 10)})

	// 並びが止まっているので何でも出せる。
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♠7: %v", err)
	}
	if got := p.GetStopsSuit(); got != CardDesignSpade {
		t.Errorf("stops suit = %d, want spades", got)
	}
	if got := p.GetCurrentPlayerIdx(); got != 1 {
		t.Fatalf("current = %d, want 1 (holds ♠8)", got)
	}
	// ♥7 は次の札ではない。
	if err := p.Play(1, 1); err == nil {
		t.Error("expected an error playing a card of another suit mid-run")
	}
	if err := p.Play(1, 0); err != nil {
		t.Fatalf("Play ♠8: %v", err)
	}
}

// TestPochStopRestartsWithTheLastHighestCard は stop の扱いを確かめる。
func TestPochStopRestartsWithTheLastHighestCard(t *testing.T) {
	p := pcReady(t, PochPhaseStops, 0)
	for i := range PochPlayerCnt {
		p.GetPlayer(i).Reset()
	}
	// ♠7 の次 (♠8) は誰も持っていない。
	setPcHand(p, 0, []*Card{pcCard(CardDesignSpade, 7), pcCard(CardDesignHeart, 13)})
	setPcHand(p, 1, []*Card{pcCard(CardDesignHeart, 7)})
	setPcHand(p, 2, []*Card{pcCard(CardDesignClover, 9)})
	setPcHand(p, 3, []*Card{pcCard(CardDesignDiamond, 10)})

	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play ♠7: %v", err)
	}
	if got := p.GetStopsSuit(); got != -1 {
		t.Errorf("stops suit = %d, want -1: the run is stopped", got)
	}
	// **最後に最高札を出した人**が再開する。
	if got := p.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want 0, who played the last card", got)
	}
}

func TestPochPlayRejections(t *testing.T) {
	p := pcReady(t, PochPhaseStops, 0)
	setPcHand(p, 0, []*Card{pcCard(CardDesignSpade, 7)})
	if err := p.Play(1, 0); err == nil {
		t.Error("expected an error playing out of turn")
	}
	if err := p.Play(0, 9); err == nil {
		t.Error("expected an error for an out-of-range hand index")
	}
	p.SetPhaseForTest(PochPhasePochen)
	if err := p.Play(0, 0); err == nil {
		t.Error("expected an error playing outside the play stage")
	}
}

// TestPochGoingOutCollectsTheCentreAndTheCardsLeft covers the payment the
// issue does not mention -- one chip per card still held.
func TestPochGoingOutCollectsTheCentreAndTheCardsLeft(t *testing.T) {
	p := pcReady(t, PochPhaseStops, 0)
	var board PochBoard
	board.Add(PochPoolCentre, 8)
	p.SetBoardForTest(board)
	for i := range PochPlayerCnt {
		p.GetPlayer(i).Reset()
	}
	setPcHand(p, 0, []*Card{pcCard(CardDesignSpade, 7)})
	setPcHand(p, 1, []*Card{pcCard(CardDesignHeart, 9), pcCard(CardDesignHeart, 10)})
	setPcHand(p, 2, []*Card{pcCard(CardDesignClover, 9)})
	setPcHand(p, 3, nil)

	before := p.GetPlayer(0).GetChips()
	beforeOne := p.GetPlayer(1).GetChips()
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play: %v", err)
	}

	if p.GetPhase() != PochPhaseDealEnd {
		t.Fatalf("phase = %v, want DealEnd", p.GetPhase())
	}
	if got := p.GetDealWinner(); got != 0 {
		t.Errorf("deal winner = %d, want 0", got)
	}
	// centre 8 + 残り札 3 枚ぶん。
	if got := p.GetPlayer(0).GetChips(); got != before+8+3 {
		t.Errorf("winner chips = %d, want %d", got, before+8+3)
	}
	// 席 1 は 2 枚残していたので 2 払う。
	if got := p.GetPlayer(1).GetChips(); got != beforeOne-2 {
		t.Errorf("seat 1 chips = %d, want %d", got, beforeOne-2)
	}
	if got := p.GetBoard().Get(PochPoolCentre); got != 0 {
		t.Errorf("centre = %d, want it emptied", got)
	}
}

func TestPochNextDealAndGameEnd(t *testing.T) {
	p := NewDefaultPoch()
	p.Reset()
	if err := p.NextDeal(); err == nil {
		t.Error("expected an error while the deal is still live")
	}

	p.SetPhaseForTest(PochPhaseDealEnd)
	if err := p.NextDeal(); err != nil {
		t.Fatalf("NextDeal: %v", err)
	}
	if p.GetPhase() != PochPhasePochen {
		t.Errorf("phase = %v, want Pochen", p.GetPhase())
	}
	// ディーラーが 1 つ進むので、先手も 1 つ進む。
	if got := p.GetCurrentPlayerIdx(); got != 2 {
		t.Errorf("current = %d, want 2", got)
	}

	p.SetDealNumberForTest(p.GetConfig().TargetDeals - 1)
	p.SetPhaseForTest(PochPhaseStops)
	p.SetStopsForTest(-1, 0)
	for i := range PochPlayerCnt {
		p.GetPlayer(i).Reset()
	}
	setPcHand(p, 0, []*Card{pcCard(CardDesignSpade, 7)})
	p.SetCurrentPlayerForTest(0)
	if err := p.Play(0, 0); err != nil {
		t.Fatalf("Play: %v", err)
	}
	if !p.GetGameEndFlag() {
		t.Fatal("the game should be over after the last deal")
	}
	if idx := p.GetWinnerIdx(); idx < 0 || idx >= PochPlayerCnt {
		t.Errorf("winnerIdx = %d, want a real seat", idx)
	}
	if err := p.NextDeal(); err == nil {
		t.Error("expected an error dealing after the game is over")
	}
	if err := p.Bet(0); err == nil {
		t.Error("expected an error betting after the game is over")
	}
	if err := p.Play(0, 0); err == nil {
		t.Error("expected an error playing after the game is over")
	}
}

// TestPochPoolsCarryOverBetweenDeals は、この game の中心的な動機づけを押さえる。
func TestPochPoolsCarryOverBetweenDeals(t *testing.T) {
	p := NewDefaultPoch()
	p.Reset()
	// 誰も取れないように pay suit を持たせない状況を作るのは難しいので、
	// 盤を直接置いて次のディールで積み増されることを見る。
	var board PochBoard
	board.Add(PochPoolMarriage, 20)
	p.SetBoardForTest(board)
	p.SetPhaseForTest(PochPhaseDealEnd)
	if err := p.NextDeal(); err != nil {
		t.Fatalf("NextDeal: %v", err)
	}
	if got := p.GetBoard().Get(PochPoolMarriage); got != 0 && got < 20 {
		t.Errorf("marriage = %d, want the 20 carried over (or taken whole)", got)
	}
}

func TestPochCpuDecide(t *testing.T) {
	p := pcReady(t, PochPhasePochen, 1)
	setPcHand(p, 1, []*Card{
		pcCard(CardDesignSpade, 7), pcCard(CardDesignHeart, 7), pcCard(CardDesignClover, 7),
	})
	if act := p.PochCpuDecide(1); act.Type != "bet" {
		t.Errorf("with three of a kind the CPU should bet, got %+v", act)
	}

	// 組がなく、追いつくのに払いが要るなら降りる。
	setPcHand(p, 1, []*Card{pcCard(CardDesignSpade, 7), pcCard(CardDesignHeart, 8)})
	p.SetCurrentPlayerForTest(0)
	if err := p.Bet(0); err != nil {
		t.Fatalf("Bet: %v", err)
	}
	p.SetCurrentPlayerForTest(1)
	if act := p.PochCpuDecide(1); act.Type != "fold" {
		t.Errorf("with nothing and a price to pay the CPU should fold, got %+v", act)
	}

	p.SetPhaseForTest(PochPhaseStops)
	p.SetStopsForTest(-1, 0)
	act := p.PochCpuDecide(1)
	if act.Type != "play" || act.HandIdx < 0 {
		t.Fatalf("in the play stage the CPU should play, got %+v", act)
	}
	// 出せる中で最も低い札を選ぶ。
	if got := p.GetPlayer(1).GetCard(act.HandIdx).GetValue(); got != 7 {
		t.Errorf("the CPU played a %d, want the lowest playable (7)", got)
	}

	if p.PochCpuDecide(9).HandIdx != -1 {
		t.Error("an unknown seat must not produce a move")
	}
	p.SetPhaseForTest(PochPhaseDealEnd)
	if act := p.PochCpuDecide(1); act.Type != "fold" {
		t.Errorf("outside the two live stages the CPU does nothing, got %+v", act)
	}
}

// TestPochCpuDrivesDealsToAnEnd checks that four CPUs always finish a deal.
func TestPochCpuDrivesDealsToAnEnd(t *testing.T) {
	for trial := range 30 {
		p := NewDefaultPoch()
		p.Reset()
		for range 2000 {
			ph := p.GetPhase()
			if ph == PochPhaseDealEnd || ph == PochPhaseGameEnd {
				break
			}
			cur := p.GetCurrentPlayerIdx()
			act := p.PochCpuDecide(cur)
			var err error
			switch act.Type {
			case "bet":
				err = p.Bet(cur)
			case "fold":
				err = p.Fold(cur)
			default:
				if act.HandIdx < 0 {
					t.Fatalf("trial %d: seat %d had nothing playable in the play stage", trial, cur)
				}
				err = p.Play(cur, act.HandIdx)
			}
			if err != nil {
				t.Fatalf("trial %d: %v", trial, err)
			}
		}
		if ph := p.GetPhase(); ph != PochPhaseDealEnd && ph != PochPhaseGameEnd {
			t.Fatalf("trial %d: the deal never ended (phase = %v)", trial, ph)
		}
		if p.GetDealWinner() < 0 {
			t.Fatalf("trial %d: nobody went out", trial)
		}
	}
}

func TestPochActionLog(t *testing.T) {
	p := NewDefaultPoch()
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

func TestPochConfigValidate(t *testing.T) {
	if err := DefaultPochConfig().Validate(); err != nil {
		t.Fatalf("the default config should validate: %v", err)
	}
	if err := (PochConfig{CpuDifficulty: 5, TargetDeals: 5}).Validate(); err == nil {
		t.Error("expected an error for an unknown difficulty")
	}
	if err := (PochConfig{TargetDeals: 0}).Validate(); err == nil {
		t.Error("expected an error for zero deals")
	}
	p := NewDefaultPoch()
	p.SetConfig(PochConfig{TargetDeals: 3})
	if p.GetConfig().TargetDeals != 3 {
		t.Error("SetConfig/GetConfig disagree")
	}
}

func TestPochAccessorBounds(t *testing.T) {
	p := NewDefaultPoch()
	p.Reset()
	if p.GetPlayer(-1) != nil || p.GetPlayer(PochPlayerCnt) != nil {
		t.Error("GetPlayer should return nil outside the table")
	}
	if got := p.GetDealNumber(); got != 0 {
		t.Errorf("GetDealNumber = %d, want 0", got)
	}
	if p.GetPlayedPile() != nil && len(p.GetPlayedPile()) != 0 {
		t.Error("nothing has been played yet")
	}
}

func TestPochJSONRoundTrip(t *testing.T) {
	p := NewDefaultPoch()
	p.Reset()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Poch
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// **盤のチップは持ち越しそのもの。**落とすとゲームの動機づけが消える。
	for i := range PochPoolCount {
		if got.GetBoard().Get(PochPool(i)) != p.GetBoard().Get(PochPool(i)) {
			t.Errorf("pool %d = %d, want %d", i, got.GetBoard().Get(PochPool(i)), p.GetBoard().Get(PochPool(i)))
		}
	}
	if got.GetPaySuit() != p.GetPaySuit() {
		t.Errorf("pay suit = %d, want %d", got.GetPaySuit(), p.GetPaySuit())
	}
	if got.GetCurrentPlayerIdx() != p.GetCurrentPlayerIdx() {
		t.Errorf("current = %d, want %d", got.GetCurrentPlayerIdx(), p.GetCurrentPlayerIdx())
	}
	if len(got.GetActionLog()) != len(p.GetActionLog()) {
		t.Errorf("action log = %d entries, want %d", len(got.GetActionLog()), len(p.GetActionLog()))
	}
}

func TestPochUnmarshalRejectsGarbage(t *testing.T) {
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
			var p Poch
			if err := json.Unmarshal([]byte(tt.data), &p); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestPochUnmarshalClampsIndices(t *testing.T) {
	var p Poch
	data := `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":1,"cur":99,"dl":-3,` +
		`"pw":42,"dw":42,"wi":42,"ss":99,` +
		`"sa":[null,{"Pool":99,"Player":0,"Chips":1},{"Pool":0,"Player":99,"Chips":1},` +
		`{"Pool":0,"Player":1,"Chips":4}]}`
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := p.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want it clamped to 0", got)
	}
	if got := p.GetPochenWinner(); got != -1 {
		t.Errorf("pochenWinner = %d, want -1", got)
	}
	if got := p.GetWinnerIdx(); got != -1 {
		t.Errorf("winnerIdx = %d, want -1", got)
	}
	// **-1 は「好きな札で開始できる」という意味を持つ。**
	if got := p.GetStopsSuit(); got != -1 {
		t.Errorf("stopsSuit = %d, want it reset to -1", got)
	}
	if got := len(p.GetStakingAwards()); got != 1 {
		t.Errorf("staking awards = %d, want the malformed ones dropped", got)
	}
}

// TestPochUnmarshalKeepsAStoppedRun は、-1 を潰さないことを直に確かめる。
func TestPochUnmarshalKeepsAStoppedRun(t *testing.T) {
	var p Poch
	data := `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":2,"ss":-1}`
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := p.GetStopsSuit(); got != -1 {
		t.Errorf("stopsSuit = %d, want -1 preserved", got)
	}

	var live Poch
	liveData := `{"pl":[{},{},{},{}],"cfg":{"cd":0,"td":5},"ph":2,"ss":1,"sr":8}`
	if err := json.Unmarshal([]byte(liveData), &live); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := live.GetStopsSuit(); got != 1 {
		t.Errorf("stopsSuit = %d, want 1 preserved", got)
	}
	if got := live.GetStopsRank(); got != 8 {
		t.Errorf("stopsRank = %d, want 8", got)
	}
}

// **並びに従う義務がある。**出せる札を先に示さないと、押して初めて弾かれる (#4933)。
func TestPoch_PochValidPlays(t *testing.T) {
	g := NewDefaultPoch()
	g.Reset()
	g.SetPhaseForTest(PochPhaseStops)
	g.SetCurrentPlayerForTest(0)

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 9, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	p.AddCard(NewCard(CardDesignSpade, 10, false))

	// 自由リード (stopsSuit < 0) では全部出せる。
	g.SetStopsForTest(-1, 0)
	if got := g.PochValidPlays(0); len(got) != 3 {
		t.Fatalf("free lead should allow every card, got %v", got)
	}

	// ♠8 の次は ♠9 だけ。同ランクの ♥9 も ♠10 も続けられない。
	g.SetStopsForTest(CardDesignSpade, pochRankOrder(8))
	if got := g.PochValidPlays(0); len(got) != 1 || got[0] != 0 {
		t.Fatalf("only the next higher card of the run suit is legal, got %v", got)
	}

	// 続けられる札が 1 枚も無ければ空。
	g.SetStopsForTest(CardDesignSpade, pochRankOrder(2))
	if got := g.PochValidPlays(0); len(got) != 0 {
		t.Fatalf("nothing should be playable, got %v", got)
	}

	// 手番でない/フェーズ違いは nil。
	if got := g.PochValidPlays(1); got != nil {
		t.Fatalf("off-turn must be nil, got %v", got)
	}
	g.SetPhaseForTest(PochPhasePochen)
	if got := g.PochValidPlays(0); got != nil {
		t.Fatalf("outside the stops phase must be nil, got %v", got)
	}
}
