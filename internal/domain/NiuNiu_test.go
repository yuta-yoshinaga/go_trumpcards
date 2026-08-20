//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

// newTestNiuNiu returns a freshly reset NiuNiu.
func newTestNiuNiu() *NiuNiu {
	n := NewDefaultNiuNiu()
	n.Reset()
	return n
}

// nnCard builds a card for tests.
func nnCard(design, value int) *Card {
	return NewCard(design, value, true)
}

// nnHandOf builds a five-card hand from ranks, all in spades unless a suit is
// needed for a tie-break.
func nnHandOf(values ...int) []*Card {
	out := make([]*Card, len(values))
	for i, v := range values {
		out[i] = nnCard(CardDesignSpade, v)
	}
	return out
}

func TestNiuNiu_Reset(t *testing.T) {
	n := newTestNiuNiu()

	if n.GetPhase() != NiuNiuPhaseBet {
		t.Errorf("phase = %d, want bet", n.GetPhase())
	}
	if n.GetChips() != NiuNiuDefaultChips {
		t.Errorf("chips = %d, want %d", n.GetChips(), NiuNiuDefaultChips)
	}
	if got := len(n.GetSeats()); got != NiuNiuSeatCnt {
		t.Errorf("seats = %d, want %d", got, NiuNiuSeatCnt)
	}
	if n.GetGameEndFlag() {
		t.Error("a fresh game should not be over")
	}
}

// A brand-new session must open on the betting phase with a CPU banking --
// `banker`'s zero value is the human seat, and starting there would end the
// round before the player ever placed a stake.
func TestNiuNiu_AFreshGameDoesNotOpenWithTheHumanBanking(t *testing.T) {
	n := NewDefaultNiuNiu()
	if n.GetBankerIdx() == 0 {
		t.Error("a brand-new game should not start with the human banking")
	}
	n.Reset()
	if n.GetBankerIdx() == 0 {
		t.Error("the first Reset should not hand the human the bank either")
	}
	if err := n.PlaceBet(100); err != nil {
		t.Errorf("PlaceBet on a fresh game: %v", err)
	}
}

func TestNiuNiu_CardPoints(t *testing.T) {
	tests := []struct {
		name string
		card *Card
		want int
	}{
		{"ace is one", nnCard(CardDesignSpade, 1), 1},
		{"seven is seven", nnCard(CardDesignSpade, 7), 7},
		{"ten is ten", nnCard(CardDesignSpade, 10), 10},
		{"jack is ten", nnCard(CardDesignSpade, 11), 10},
		{"queen is ten", nnCard(CardDesignSpade, 12), 10},
		{"king is ten", nnCard(CardDesignSpade, 13), 10},
		{"nil is nothing", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := niuNiuCardPoints(tt.card); got != tt.want {
				t.Errorf("points = %d, want %d", got, tt.want)
			}
		})
	}
}

// #4397 says face cards are 10; other sources say 0. Everything here is used
// mod 10, and 10 = 0 (mod 10), so the two conventions cannot disagree -- not on
// which triples form a bull, and not on the two-card digit. This pins that.
func TestNiuNiu_FaceCardConventionsAgree(t *testing.T) {
	// Points under the "face = 0" convention.
	zeroFace := func(c *Card) int {
		v := c.GetValue()
		if v >= 10 {
			return 0
		}
		return v
	}

	hands := [][]*Card{
		nnHandOf(11, 12, 13, 5, 5),
		nnHandOf(11, 5, 5, 3, 7),
		nnHandOf(10, 10, 10, 1, 2),
		nnHandOf(13, 13, 4, 6, 9),
		nnHandOf(1, 2, 3, 4, 5),
		nnHandOf(11, 12, 4, 6, 8),
	}
	for i, cards := range hands {
		_, rank := niuNiuEvaluate(cards)

		// Re-evaluate with face = 0 and compare.
		total := 0
		pts := make([]int, len(cards))
		for j, c := range cards {
			pts[j] = zeroFace(c)
			total += pts[j]
		}
		altRank := NiuNiuRankNone
		found := false
		for a := range NiuNiuHandSize {
			for b := a + 1; b < NiuNiuHandSize && !found; b++ {
				for c := b + 1; c < NiuNiuHandSize; c++ {
					if (pts[a]+pts[b]+pts[c])%10 != 0 {
						continue
					}
					rest := (total - pts[a] - pts[b] - pts[c]) % 10
					altRank = NiuNiuRank(rest)
					if rest == 0 {
						altRank = NiuNiuRankNiuNiu
					}
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if rank != altRank {
			t.Errorf("hand %d: face=10 gives %v but face=0 gives %v", i, rank, altRank)
		}
	}
}

func TestNiuNiu_Evaluate(t *testing.T) {
	tests := []struct {
		name     string
		values   []int
		wantRank NiuNiuRank
		wantBull bool
	}{
		// 3+7 = 10 with a face card... 3+7+10 = 20, rest 5+5 = 10 -> niu niu.
		{"three tens and a pair make niu niu", []int{10, 10, 10, 5, 5}, NiuNiuRankNiuNiu, true},
		// 1+2+7 = 10, rest 3+4 = 7 -> niu 7.
		{"a bull of ten leaves seven", []int{1, 2, 7, 3, 4}, 7, true},
		// 5+5+10 = 20, rest 1+2 = 3 -> niu 3.
		{"a bull of twenty leaves three", []int{5, 5, 10, 1, 2}, 3, true},
		// All face cards: 10+10+10 = 30, rest 10+10 = 20 -> digit 0 -> niu niu.
		{"five face cards are niu niu", []int{11, 12, 13, 11, 12}, NiuNiuRankNiuNiu, true},
		// Every triple here lands between 4 and 7, so nothing reaches a multiple
		// of ten. Picked by exhaustive check -- {1,2,3,4,6} looks like a no-bull
		// but 1+3+6 = 10, which is exactly the sort of arithmetic a fixture gets
		// wrong by eye.
		{"no combination is a no-bull", []int{1, 1, 2, 2, 3}, NiuNiuRankNone, false},
		// 9+9+2 = 20, rest 3+6 = 9 -> niu 9, the best non-niu-niu hand.
		{"niu nine", []int{9, 9, 2, 3, 6}, 9, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combo, rank := niuNiuEvaluate(nnHandOf(tt.values...))
			if rank != tt.wantRank {
				t.Errorf("rank = %v, want %v", rank, tt.wantRank)
			}
			if (combo != nil) != tt.wantBull {
				t.Errorf("combo = %v, want bull = %v", combo, tt.wantBull)
			}
			if tt.wantBull && len(combo) != NiuNiuComboSize {
				t.Errorf("combo holds %d cards, want %d", len(combo), NiuNiuComboSize)
			}
		})
	}
}

// The combo the search returns is genuinely a multiple of ten, and the rank is
// genuinely the remaining two cards' last digit. Checking the pair together
// stops a bug that picks the wrong three cards from hiding behind a right-
// looking rank.
func TestNiuNiu_EvaluateReturnsAConsistentCombo(t *testing.T) {
	for _, values := range [][]int{
		{10, 10, 10, 5, 5},
		{1, 2, 7, 3, 4},
		{5, 5, 10, 1, 2},
		{9, 9, 2, 3, 6},
		{11, 12, 13, 4, 8},
	} {
		cards := nnHandOf(values...)
		combo, rank := niuNiuEvaluate(cards)
		if combo == nil {
			t.Fatalf("%v: expected a bull", values)
		}
		sum := 0
		used := map[int]bool{}
		for _, i := range combo {
			sum += niuNiuCardPoints(cards[i])
			used[i] = true
		}
		if sum%10 != 0 {
			t.Errorf("%v: combo sums to %d, not a multiple of ten", values, sum)
		}
		rest := 0
		for i, c := range cards {
			if !used[i] {
				rest += niuNiuCardPoints(c)
			}
		}
		want := NiuNiuRank(rest % 10)
		if rest%10 == 0 {
			want = NiuNiuRankNiuNiu
		}
		if rank != want {
			t.Errorf("%v: rank = %v but the leftover pair is %d", values, rank, rest%10)
		}
	}
}

// The search returns the FIRST triple it finds. That is only safe because every
// triple summing to a multiple of ten leaves the same remainder -- the five-card
// total is fixed, so removing congruent triples leaves congruent pairs. If that
// were false, the hand's rank would depend on iteration order.
func TestNiuNiu_EveryValidComboGivesTheSameRank(t *testing.T) {
	// 5+5+10+10+10: several triples reach a multiple of ten.
	for _, values := range [][]int{
		{5, 5, 10, 10, 10},
		{10, 10, 10, 10, 10},
		{2, 8, 10, 4, 6},
		{1, 9, 10, 3, 7},
	} {
		cards := nnHandOf(values...)
		pts := make([]int, len(cards))
		total := 0
		for i, c := range cards {
			pts[i] = niuNiuCardPoints(c)
			total += pts[i]
		}
		var seen []int
		for a := range NiuNiuHandSize {
			for b := a + 1; b < NiuNiuHandSize; b++ {
				for c := b + 1; c < NiuNiuHandSize; c++ {
					if (pts[a]+pts[b]+pts[c])%10 == 0 {
						seen = append(seen, (total-pts[a]-pts[b]-pts[c])%10)
					}
				}
			}
		}
		if len(seen) < 2 {
			t.Fatalf("%v: expected several valid triples, found %d", values, len(seen))
		}
		for _, r := range seen[1:] {
			if r != seen[0] {
				t.Errorf("%v: triples disagree on the leftover digit: %v", values, seen)
			}
		}
	}
}

func TestNiuNiu_Multiplier(t *testing.T) {
	tests := []struct {
		rank NiuNiuRank
		want int
	}{
		{NiuNiuRankNiuNiu, 3},
		{9, 2},
		{8, 2},
		{7, 2},
		{6, 1},
		{1, 1},
		{NiuNiuRankNone, 1},
	}
	for _, tt := range tests {
		if got := niuNiuMultiplier(tt.rank); got != tt.want {
			t.Errorf("multiplier(%v) = %d, want %d", tt.rank, got, tt.want)
		}
	}
}

func TestNiuNiu_RankLabel(t *testing.T) {
	tests := []struct {
		rank NiuNiuRank
		want string
	}{
		{NiuNiuRankNone, "無牛"},
		{NiuNiuRankNiuNiu, "牛牛"},
		{1, "牛1"},
		{9, "牛9"},
	}
	for _, tt := range tests {
		if got := NiuNiuRankLabel(tt.rank); got != tt.want {
			t.Errorf("label(%v) = %q, want %q", tt.rank, got, tt.want)
		}
	}
}

// Ranks alone leave many ties -- there are only eleven of them -- so the
// tie-break has to be real.
func TestNiuNiu_TieBreakByHighestCardThenSuit(t *testing.T) {
	// Same rank, different high card.
	a := &NiuNiuHand{cards: nnHandOf(1, 2, 7, 3, 13), rank: 7}
	b := &NiuNiuHand{cards: nnHandOf(1, 2, 7, 3, 4), rank: 7}
	if !niuNiuBeats(a, b) {
		t.Error("a king should beat a four at equal rank")
	}
	if niuNiuBeats(b, a) {
		t.Error("the comparison must not be symmetric")
	}

	// Same rank and same high rank, different suit.
	spade := &NiuNiuHand{cards: []*Card{nnCard(CardDesignSpade, 13)}, rank: 5}
	heart := &NiuNiuHand{cards: []*Card{nnCard(CardDesignHeart, 13)}, rank: 5}
	club := &NiuNiuHand{cards: []*Card{nnCard(CardDesignClover, 13)}, rank: 5}
	diamond := &NiuNiuHand{cards: []*Card{nnCard(CardDesignDiamond, 13)}, rank: 5}
	for _, tc := range []struct{ hi, lo *NiuNiuHand }{
		{spade, heart}, {heart, club}, {club, diamond}, {spade, diamond},
	} {
		if !niuNiuBeats(tc.hi, tc.lo) {
			t.Errorf("suit order broken: %d should beat %d",
				tc.hi.cards[0].GetDesign(), tc.lo.cards[0].GetDesign())
		}
	}
}

// The ace is the WEAKEST card for the tie-break even though it is worth one
// point -- the order is K > Q > J > 10 > ... > 2 > A.
func TestNiuNiu_AceIsWeakestInTheTieBreak(t *testing.T) {
	ace := &NiuNiuHand{cards: []*Card{nnCard(CardDesignSpade, 1)}, rank: 3}
	two := &NiuNiuHand{cards: []*Card{nnCard(CardDesignDiamond, 2)}, rank: 3}
	if !niuNiuBeats(two, ace) {
		t.Error("a two should beat an ace at equal rank, even across suits")
	}
	king := &NiuNiuHand{cards: []*Card{nnCard(CardDesignDiamond, 13)}, rank: 3}
	if !niuNiuBeats(king, ace) {
		t.Error("a king should beat an ace")
	}
}

func TestNiuNiu_HigherRankBeatsLower(t *testing.T) {
	niuniu := &NiuNiuHand{cards: nnHandOf(1, 2, 3, 4, 5), rank: NiuNiuRankNiuNiu}
	niu9 := &NiuNiuHand{cards: nnHandOf(13, 13, 13, 13, 13), rank: 9}
	none := &NiuNiuHand{cards: nnHandOf(13, 13, 13, 13, 13), rank: NiuNiuRankNone}

	if !niuNiuBeats(niuniu, niu9) {
		t.Error("niu niu should beat niu 9 regardless of cards")
	}
	if !niuNiuBeats(niu9, none) {
		t.Error("niu 9 should beat a no-bull")
	}
	if niuNiuBeats(none, niu9) {
		t.Error("a no-bull must not beat a bull")
	}
}

// setupNiuNiuRound puts exact hands on the table.
func setupNiuNiuRound(n *NiuNiu, playerCards, bankerCards []*Card, bet int) *NiuNiuHand {
	n.banker = 3
	n.seats = []*NiuNiuSeat{
		{name: "あなた", isCPU: false},
		{name: "CPU1", isCPU: true},
		{name: "CPU2", isCPU: true},
		{name: "親", isCPU: true},
	}
	h := &NiuNiuHand{cards: playerCards, bet: bet}
	h.comboIdx, h.rank = niuNiuEvaluate(playerCards)
	n.seats[0].hand = h

	bh := &NiuNiuHand{cards: bankerCards}
	bh.comboIdx, bh.rank = niuNiuEvaluate(bankerCards)
	n.bankerHand = bh
	return h
}

func TestNiuNiu_SettleHand(t *testing.T) {
	const bet = 100

	tests := []struct {
		name    string
		player  []int
		banker  []int
		want    int
		comment string
	}{
		// 牛牛 (x3) vs 牛3.
		{"niu niu pays triple", []int{10, 10, 10, 5, 5}, []int{5, 5, 10, 1, 2}, bet * 3, ""},
		// 牛9 (x2) vs 牛3.
		{"niu nine pays double", []int{9, 9, 2, 3, 6}, []int{5, 5, 10, 1, 2}, bet * 2, ""},
		// 牛3 (x1) vs 無牛.
		{"a low bull pays once", []int{5, 5, 10, 1, 2}, []int{1, 1, 2, 2, 3}, bet, ""},
		// 無牛 vs 牛牛 -- the BANKER's multiplier applies when the banker wins.
		{"the banker's niu niu takes triple", []int{1, 1, 2, 2, 3}, []int{10, 10, 10, 5, 5}, -bet * 3, ""},
		// 無牛 vs 牛8 -- banker x2.
		{"the banker's niu eight takes double", []int{1, 1, 2, 2, 3}, []int{9, 9, 2, 3, 5}, -bet * 2, ""},
		// 牛3 vs 牛7 -- banker x2.
		{"losing to a niu seven costs double", []int{5, 5, 10, 1, 2}, []int{1, 2, 7, 3, 4}, -bet * 2, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newTestNiuNiu()
			h := setupNiuNiuRound(n, nnHandOf(tt.player...), nnHandOf(tt.banker...), bet)
			if got := n.settleHand(h); got != tt.want {
				t.Errorf("settleHand = %d, want %d (player %v vs banker %v)",
					got, tt.want, h.GetRank(), n.GetBankerHand().GetRank())
			}
		})
	}
}

func TestNiuNiu_SettleCreditsTheHumanStack(t *testing.T) {
	n := newTestNiuNiu()
	setupNiuNiuRound(n, nnHandOf(10, 10, 10, 5, 5), nnHandOf(1, 1, 2, 2, 3), 100)
	before := n.GetChips()

	n.settle()

	// 賭け金 100 が戻り、牛牛の 3 倍で 300 の勝ち。
	if n.GetChips() != before+400 {
		t.Errorf("chips = %d, want %d", n.GetChips(), before+400)
	}
	if n.GetPhase() != NiuNiuPhaseEnd {
		t.Errorf("phase = %d, want end", n.GetPhase())
	}
	if !n.GetGameEndFlag() {
		t.Error("GetGameEndFlag should be true after settling")
	}
	if n.GetLastResult() == "" {
		t.Error("the settlement should be summarised")
	}
}

func TestNiuNiu_PlaceBet(t *testing.T) {
	n := newTestNiuNiu()

	if err := n.PlaceBet(NiuNiuMinBet - 1); err == nil {
		t.Error("a bet below the minimum should be rejected")
	}
	if err := n.PlaceBet(NiuNiuMaxBet + 1); err == nil {
		t.Error("a bet above the maximum should be rejected")
	}
	if err := n.PlaceBet(NiuNiuDefaultChips * 10); err == nil {
		t.Error("betting more than the stack should be rejected")
	}

	if err := n.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	// 選択肢が無いゲームなので、ベットしたらその場で精算まで進む。
	if n.GetPhase() != NiuNiuPhaseEnd {
		t.Errorf("phase = %d, want end -- the round resolves at the bet", n.GetPhase())
	}
	if n.GetSeats()[0].GetHand() == nil {
		t.Fatal("the human should hold a hand")
	}
	if got := len(n.GetSeats()[0].GetHand().GetCards()); got != NiuNiuHandSize {
		t.Errorf("cards = %d, want %d", got, NiuNiuHandSize)
	}
}

// The stake is not what the player risks: a banker's Niu Niu takes THREE times
// it. Checking the stake alone against the stack let the balance go negative --
// 25 chips, a legal 10 stake, and a banker Niu Niu left -5 on screen.
func TestNiuNiu_ChipsNeverGoNegative(t *testing.T) {
	for range 500 {
		n := NewDefaultNiuNiu()
		n.Reset()
		// 最悪の負けをちょうど賄える最小の残高。
		n.chips.SetChips(NiuNiuMinBet * NiuNiuMaxMultiplier)
		if err := n.PlaceBet(NiuNiuMinBet); err != nil {
			t.Fatalf("PlaceBet: %v", err)
		}
		if n.GetChips() < 0 {
			t.Fatalf("chips = %d after a banker %s", n.GetChips(),
				NiuNiuRankLabel(n.GetBankerHand().GetRank()))
		}
	}
}

// A stake the stack cannot cover three times over is refused, because the loss
// can be three times the stake.
func TestNiuNiu_RejectsAStakeItCannotCover(t *testing.T) {
	n := NewDefaultNiuNiu()
	n.Reset()

	// 25 チップでは合法な賭け金が一つも無い。最低額 10 でも最悪 30 取られる。
	// **この行き止まりがあるからこそ** Reset の積み増し閾値は最低額ではなく
	// 最低額×最大倍率になっている。
	n.chips.SetChips(25)
	if err := n.PlaceBet(NiuNiuMinBet); err == nil {
		t.Error("a stake whose worst case exceeds the stack should be rejected")
	}

	// ちょうど賄える残高なら通る。
	n.chips.SetChips(NiuNiuMinBet * NiuNiuMaxMultiplier)
	if err := n.PlaceBet(NiuNiuMinBet); err != nil {
		t.Errorf("a coverable stake was rejected: %v", err)
	}
}

// Reset must top up below MIN*MAX, not below MIN. Topping up only below the
// minimum leaves balances like 25 that can be bet but not paid, and the player
// would be stuck refusing every legal stake.
func TestNiuNiu_ResetTopsUpWhenTheWorstCaseIsUncoverable(t *testing.T) {
	n := NewDefaultNiuNiu()
	n.Reset()
	n.chips.SetChips(25) // >= NiuNiuMinBet, but < NiuNiuMinBet*NiuNiuMaxMultiplier
	n.Reset()

	if n.GetChips() != NiuNiuDefaultChips {
		t.Errorf("chips = %d, want a fresh stack", n.GetChips())
	}
	if err := n.PlaceBet(NiuNiuMinBet); err != nil {
		t.Errorf("the minimum stake should be playable after a top-up: %v", err)
	}
}

func TestNiuNiu_MaxMultiplierMatchesTheTable(t *testing.T) {
	// 表と定数がずれると、賄えない賭けを通してしまう。
	best := 0
	for r := NiuNiuRankNone; r <= NiuNiuRankNiuNiu; r++ {
		if m := niuNiuMultiplier(r); m > best {
			best = m
		}
	}
	if best != NiuNiuMaxMultiplier {
		t.Errorf("the payout table peaks at x%d but NiuNiuMaxMultiplier is %d",
			best, NiuNiuMaxMultiplier)
	}
	if n := NewDefaultNiuNiu(); n.GetMaxMultiplier() != NiuNiuMaxMultiplier {
		t.Error("GetMaxMultiplier should mirror the constant")
	}
}

func TestNiuNiu_PlaceBetRejectedOutOfPhase(t *testing.T) {
	n := newTestNiuNiu()
	n.phase = NiuNiuPhaseEnd
	if err := n.PlaceBet(100); err == nil {
		t.Error("PlaceBet should fail outside the betting phase")
	}
}

// Every seat gets five cards and a settled rank; the banker plays no stake.
func TestNiuNiu_DealGivesEverySeatAHand(t *testing.T) {
	n := newTestNiuNiu()
	if err := n.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	for i, s := range n.GetSeats() {
		if i == n.GetBankerIdx() {
			if s.GetHand() != nil {
				t.Errorf("the banker seat should not hold a staked hand")
			}
			continue
		}
		h := s.GetHand()
		if h == nil {
			t.Fatalf("seat %d has no hand", i)
		}
		if len(h.GetCards()) != NiuNiuHandSize {
			t.Errorf("seat %d holds %d cards, want %d", i, len(h.GetCards()), NiuNiuHandSize)
		}
		if h.GetBet() <= 0 {
			t.Errorf("seat %d staked %d", i, h.GetBet())
		}
	}
	if n.GetBankerHand() == nil || len(n.GetBankerHand().GetCards()) != NiuNiuHandSize {
		t.Error("the banker should hold five cards")
	}
	if n.GetBankerHand().GetBet() != 0 {
		t.Errorf("the banker staked %d, want 0", n.GetBankerHand().GetBet())
	}
}

func TestNiuNiu_ActionLog(t *testing.T) {
	n := newTestNiuNiu()
	if err := n.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	log := n.GetActionLog()
	if len(log) < 2 {
		t.Fatalf("log = %d entries, want the deal and the result", len(log))
	}
	if log[0].ActionType != "deal" {
		t.Errorf("log[0].ActionType = %q, want deal", log[0].ActionType)
	}
	if log[len(log)-1].ActionType != "result" {
		t.Errorf("last entry = %q, want result", log[len(log)-1].ActionType)
	}
}

func TestNiuNiu_DisplayHelpers(t *testing.T) {
	n := newTestNiuNiu()
	if n.GetBankerRankKey() != "" {
		t.Error("GetBankerRankKey should be empty before the banker hand exists")
	}
	if n.GetMultiplier(NiuNiuRankNiuNiu) != 3 {
		t.Error("GetMultiplier should mirror niuNiuMultiplier")
	}
	if n.GetSeats()[0].GetName() != "あなた" {
		t.Errorf("name = %q", n.GetSeats()[0].GetName())
	}
	if n.GetSeats()[0].IsCPU() {
		t.Error("seat 0 is the human")
	}
	if !n.GetSeats()[1].IsCPU() {
		t.Error("seat 1 should be a CPU")
	}
}

func TestNiuNiu_JSONRoundTrip(t *testing.T) {
	n := newTestNiuNiu()
	if err := n.PlaceBet(100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}

	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	restored := NewDefaultNiuNiu()
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored.GetPhase() != n.GetPhase() {
		t.Errorf("phase = %d, want %d", restored.GetPhase(), n.GetPhase())
	}
	if restored.GetChips() != n.GetChips() {
		t.Errorf("chips = %d, want %d", restored.GetChips(), n.GetChips())
	}
	if restored.GetBankerIdx() != n.GetBankerIdx() {
		t.Errorf("banker = %d, want %d", restored.GetBankerIdx(), n.GetBankerIdx())
	}

	want := n.GetSeats()[0].GetHand()
	got := restored.GetSeats()[0].GetHand()
	if got == nil || want == nil {
		t.Fatal("the human's hand must survive the wire")
	}
	if len(got.GetCards()) != len(want.GetCards()) {
		t.Errorf("cards = %d, want %d", len(got.GetCards()), len(want.GetCards()))
	}
	// 役と、それを作った 3 枚の位置も戻ること。ここが落ちると画面で牛が消える。
	if got.GetRank() != want.GetRank() {
		t.Errorf("rank = %v, want %v", got.GetRank(), want.GetRank())
	}
	if len(got.GetComboIdx()) != len(want.GetComboIdx()) {
		t.Errorf("combo = %v, want %v", got.GetComboIdx(), want.GetComboIdx())
	}
	if got.GetPayout() != want.GetPayout() {
		t.Errorf("payout = %d, want %d", got.GetPayout(), want.GetPayout())
	}
}

func TestNiuNiu_UnmarshalJSONRejectsInvalidState(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"malformed", `{`},
		{"phase too small", `{"ph":0}`},
		{"phase too large", `{"ph":99}`},
		{"too many seats", `{"ph":1,"st":[{},{},{},{},{}]}`},
		{"negative banker", `{"ph":1,"bk":-1}`},
		{"banker out of range", `{"ph":1,"bk":9}`},
		{"negative chips", `{"ph":1,"ch":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewDefaultNiuNiu()
			if err := json.Unmarshal([]byte(tt.data), n); err == nil {
				t.Errorf("Unmarshal(%s) should fail", tt.data)
			}
		})
	}
}

func TestNiuNiu_UnmarshalJSONRejectsInvalidHands(t *testing.T) {
	// 棋譜の上限。
	log := `{"ph":1,"al":[`
	for i := range niuNiuMaxSliceLen + 1 {
		if i > 0 {
			log += ","
		}
		log += `{}`
	}
	log += `]}`
	n := NewDefaultNiuNiu()
	if err := json.Unmarshal([]byte(log), n); err == nil {
		t.Error("an oversized action log should be rejected")
	}

	// 手札の上限。
	cards := `{"cd":[`
	for i := range niuNiuMaxSliceLen + 1 {
		if i > 0 {
			cards += ","
		}
		cards += `{"d":1,"v":1,"f":true}`
	}
	cards += `]}`
	var h NiuNiuHand
	if err := json.Unmarshal([]byte(cards), &h); err == nil {
		t.Error("an oversized hand should be rejected")
	}

	for _, tt := range []struct{ name, data string }{
		{"combo too long", `{"ci":[0,1,2,3]}`},
		{"combo index negative", `{"ci":[-1]}`},
		{"combo index out of range", `{"ci":[9]}`},
		{"rank negative", `{"rk":-1}`},
		{"rank too large", `{"rk":99}`},
		{"malformed hand", `{`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var bh NiuNiuHand
			if err := json.Unmarshal([]byte(tt.data), &bh); err == nil {
				t.Errorf("Unmarshal(%s) should fail", tt.data)
			}
		})
	}

	var seat NiuNiuSeat
	if err := json.Unmarshal([]byte(`{`), &seat); err == nil {
		t.Error("a malformed seat should be rejected")
	}
}

func TestNiuNiu_ResetRestoresBrokeStack(t *testing.T) {
	n := newTestNiuNiu()
	n.chips.SetChips(NiuNiuMinBet - 1)
	n.Reset()
	if n.GetChips() != NiuNiuDefaultChips {
		t.Errorf("chips = %d, want a fresh stack", n.GetChips())
	}
}

// Deal after deal, every hand must evaluate to a legal rank with a combo that
// agrees with it. This is the property that a shuffled deck can break in ways
// hand-built fixtures never reach.
func TestNiuNiu_EveryDealtHandIsSelfConsistent(t *testing.T) {
	n := newTestNiuNiu()
	for range 300 {
		n.Reset()
		// 最低額で賭ける。大きく賭けると負けが込んだときに残高が最悪の負けを
		// 賄えなくなり、PlaceBet が（正しく）弾いてループが止まる。
		if err := n.PlaceBet(NiuNiuMinBet); err != nil {
			t.Fatalf("PlaceBet: %v", err)
		}
		hands := []*NiuNiuHand{n.GetBankerHand()}
		for i, s := range n.GetSeats() {
			if i == n.GetBankerIdx() {
				continue
			}
			hands = append(hands, s.GetHand())
		}
		for _, h := range hands {
			if h == nil {
				t.Fatal("a seat was dealt no hand")
			}
			if h.GetRank() < NiuNiuRankNone || h.GetRank() > NiuNiuRankNiuNiu {
				t.Fatalf("rank %v out of range", h.GetRank())
			}
			combo := h.GetComboIdx()
			if combo == nil {
				if h.GetRank() != NiuNiuRankNone {
					t.Fatalf("no combo but rank %v", h.GetRank())
				}
				continue
			}
			if h.GetRank() == NiuNiuRankNone {
				t.Fatal("a combo was found but the rank says no bull")
			}
			sum := 0
			for _, i := range combo {
				sum += niuNiuCardPoints(h.GetCards()[i])
			}
			if sum%10 != 0 {
				t.Fatalf("combo sums to %d, not a multiple of ten", sum)
			}
		}
	}
}

// 親の格は表示文字列ではなくキーで運ぶ。
//
// settle() が "親: 牛牛" という日本語を組み立てて presenter がそれをそのまま流していたため、
// 英語ロケールでも日本語が出ていた (#5567)。ロケールに依存しない識別子を返し、
// 文言の組み立ては各 presenter の i18n に任せる。
func TestNiuNiuRankKey_IsLocaleIndependent(t *testing.T) {
	cases := []struct {
		rank NiuNiuRank
		want string
	}{
		{NiuNiuRankNone, "none"},
		{NiuNiuRankNiuNiu, "niuniu"},
		{NiuNiuRank(1), "n1"},
		{NiuNiuRank(9), "n9"},
	}
	for _, c := range cases {
		got := NiuNiuRankKey(c.rank)
		if got != c.want {
			t.Errorf("NiuNiuRankKey(%v) = %q, want %q", c.rank, got, c.want)
		}
		// キーに日本語が混ざっていたら、それは表示文字列であってキーではない。
		for _, r := range got {
			if r >= 128 {
				t.Errorf("キーに非ASCIIが混ざっている: %q", got)
			}
		}
	}
}
