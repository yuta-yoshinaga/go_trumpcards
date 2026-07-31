//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"testing"
)

func sbsCard(suit, value int) *Card { return NewCard(suit, value, true) }

// sbsPlaying puts a game into the play phase with a fixed contract.
func sbsPlaying(t *testing.T, declarer int, kind SixBidSoloBidKind, trumpSuit int) *SixBidSolo {
	t.Helper()
	s := NewDefaultSixBidSolo()
	s.Reset()
	s.SetPhaseForTest(SixBidSoloPhasePlay)
	s.SetContractForTest(declarer, kind, trumpSuit)
	s.SetCurrentPlayerForTest(0)
	s.SetTrickLeaderForTest(0)
	return s
}

// TestSixBidSoloDealsElevenEachPlusAWidow は issue の「12枚ずつ・スキャットなし」
// が誤りであることを押さえる。
//
// **11 × 3 + 3 = 36。**12 枚ずつだとウィドウの 3 枚が入らない。
func TestSixBidSoloDealsElevenEachPlusAWidow(t *testing.T) {
	if got := len(newSixBidSoloDeck()); got != SixBidSoloDeckSize {
		t.Fatalf("the deck holds %d cards, want %d", got, SixBidSoloDeckSize)
	}
	// **算術で確かめる。**issue の 12 枚ずつでは widow が置けない。
	if SixBidSoloPlayerCnt*SixBidSoloHandSize+SixBidSoloWidowSize != SixBidSoloDeckSize {
		t.Fatalf("%d seats x %d cards + a widow of %d != %d",
			SixBidSoloPlayerCnt, SixBidSoloHandSize, SixBidSoloWidowSize, SixBidSoloDeckSize)
	}
	if SixBidSoloHandSize != 11 {
		t.Fatalf("SixBidSoloHandSize = %d, want 11 — the issue's 12 is wrong", SixBidSoloHandSize)
	}

	s := NewDefaultSixBidSolo()
	s.Reset()
	total := 0
	for i := range SixBidSoloPlayerCnt {
		got := s.GetPlayer(i).GetCardsSize()
		if got != SixBidSoloHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, SixBidSoloHandSize)
		}
		total += got
	}
	if got := len(s.GetWidow()); got != SixBidSoloWidowSize {
		t.Errorf("the widow holds %d, want %d — the issue says there is none", got, SixBidSoloWidowSize)
	}
	if total+len(s.GetWidow()) != SixBidSoloDeckSize {
		t.Errorf("%d cards accounted for, want the whole pack of %d", total+len(s.GetWidow()), SixBidSoloDeckSize)
	}

	// デッキは A-10-K-Q-J-9-8-7-6 だけ。
	for _, c := range newSixBidSoloDeck() {
		switch c.GetValue() {
		case 1, 10, 13, 12, 11, 9, 8, 7, 6:
		default:
			t.Errorf("an unexpected rank is in the deck: %d", c.GetValue())
		}
	}
	// **札はすべて別物。**ウィドウを抜いても重複しない。
	seen := map[[2]int]bool{}
	for i := range SixBidSoloPlayerCnt {
		p := s.GetPlayer(i)
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			k := [2]int{c.GetDesign(), c.GetValue()}
			if seen[k] {
				t.Fatalf("card %v was dealt twice", k)
			}
			seen[k] = true
		}
	}
	for _, c := range s.GetWidow() {
		k := [2]int{c.GetDesign(), c.GetValue()}
		if seen[k] {
			t.Fatalf("widow card %v is also in a hand", k)
		}
		seen[k] = true
	}
}

// TestSixBidSoloCardValues covers the point schedule the issue omits entirely.
func TestSixBidSoloCardValues(t *testing.T) {
	for _, tc := range []struct {
		value, want int
	}{{1, 11}, {10, 10}, {13, 4}, {12, 3}, {11, 2}, {9, 0}, {8, 0}, {7, 0}, {6, 0}} {
		if got := SixBidSoloCardPoint(sbsCard(CardDesignSpade, tc.value)); got != tc.want {
			t.Errorf("card %d is worth %d, want %d", tc.value, got, tc.want)
		}
	}
	if got := SixBidSoloCardPoint(nil); got != 0 {
		t.Errorf("a nil card is worth %d, want 0", got)
	}

	// **合計はちょうど 120 点。**
	total := 0
	for _, c := range newSixBidSoloDeck() {
		total += SixBidSoloCardPoint(c)
	}
	if total != SixBidSoloTotalPoints {
		t.Errorf("the pack holds %d points, want %d", total, SixBidSoloTotalPoints)
	}
}

// **10 が K より強い。**スカート系の序列。
func TestSixBidSoloRanking(t *testing.T) {
	descending := []int{1, 10, 13, 12, 11, 9, 8, 7, 6}
	for i := 1; i < len(descending); i++ {
		hi := sixBidSoloRank(sbsCard(CardDesignSpade, descending[i-1]))
		lo := sixBidSoloRank(sbsCard(CardDesignSpade, descending[i]))
		if hi <= lo {
			t.Errorf("%d must outrank %d", descending[i-1], descending[i])
		}
	}
	// 10 が K を上回ることを名指しで押さえる。
	if sixBidSoloRank(sbsCard(CardDesignSpade, 10)) <= sixBidSoloRank(sbsCard(CardDesignSpade, 13)) {
		t.Error("the ten outranks the king in this family")
	}
	if sixBidSoloRank(nil) != 0 {
		t.Error("a nil card has no rank")
	}
}

// TestSixBidSoloBidLadder covers the six bids in ascending order.
func TestSixBidSoloBidLadder(t *testing.T) {
	ladder := []SixBidSoloBidKind{
		SixBidSoloBidSolo, SixBidSoloBidHeartSolo, SixBidSoloBidMisere,
		SixBidSoloBidGuarantee, SixBidSoloBidSpreadMisere, SixBidSoloBidCall,
	}
	for i := 1; i < len(ladder); i++ {
		if ladder[i] <= ladder[i-1] {
			t.Errorf("step %d must sit above step %d", i, i-1)
		}
	}
	if got := int(SixBidSoloBidCount) - 1; got != len(ladder) {
		t.Errorf("%d bids besides pass, want %d", got, len(ladder))
	}
}

// TestSixBidSoloTargets covers the targets, two of which the issue omits.
func TestSixBidSoloTargets(t *testing.T) {
	// **通常ビッドは 60 ちょうどでは足りない。61 点以上。**
	for _, kind := range []SixBidSoloBidKind{SixBidSoloBidSolo, SixBidSoloBidHeartSolo} {
		if got := SixBidSoloTargetPoints(kind, CardDesignSpade); got != 61 {
			t.Errorf("%v needs %d, want 61 — more than half of 120", kind, got)
		}
	}
	// **ギャランティーはスートで変わる。**issue はこの数字を出していない。
	if got := SixBidSoloTargetPoints(SixBidSoloBidGuarantee, CardDesignHeart); got != SixBidSoloGuaranteeHeart {
		t.Errorf("guarantee at hearts needs %d, want %d", got, SixBidSoloGuaranteeHeart)
	}
	if got := SixBidSoloTargetPoints(SixBidSoloBidGuarantee, CardDesignHeart); got != 74 {
		t.Errorf("guarantee at hearts needs %d, want 74", got)
	}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignDiamond} {
		if got := SixBidSoloTargetPoints(SixBidSoloBidGuarantee, suit); got != 80 {
			t.Errorf("guarantee at suit %d needs %d, want 80", suit, got)
		}
	}
	if got := SixBidSoloTargetPoints(SixBidSoloBidCall, CardDesignSpade); got != SixBidSoloTotalPoints {
		t.Errorf("call solo needs %d, want all %d", got, SixBidSoloTotalPoints)
	}
	for _, kind := range []SixBidSoloBidKind{SixBidSoloBidMisere, SixBidSoloBidSpreadMisere} {
		if got := SixBidSoloTargetPoints(kind, 0); got != 0 {
			t.Errorf("%v needs %d, want 0", kind, got)
		}
	}
	if got := SixBidSoloTargetPoints(SixBidSoloBidPass, 0); got != 0 {
		t.Errorf("a pass has target %d, want 0", got)
	}
}

// TestSixBidSoloMisereIsZeroPointsNotZeroTricks は issue の
// 「1トリックも取らなければ勝利」が誤りであることを押さえる。
//
// **9・8・7・6 だけのトリックは取っても構わない。**
func TestSixBidSoloMisereIsZeroPointsNotZeroTricks(t *testing.T) {
	s := sbsPlaying(t, 0, SixBidSoloBidMisere, 0)
	// 宣言者はトリックを 4 つ取っているが、どれも 0 点だった。
	s.SetTricksWonForTest(0, 4)
	s.SetPointsForTest(0, 0)
	s.SetPointsForTest(1, 60)
	s.SetPointsForTest(2, 60)
	s.FinishHandForTest()

	res := s.GetLastResult()
	if res == nil {
		t.Fatal("the hand must settle")
	}
	if !res.Made {
		t.Error("a misère with zero CARD POINTS is made, even after taking tricks")
	}

	// 1 点でも取れば失敗する。
	s2 := sbsPlaying(t, 0, SixBidSoloBidMisere, 0)
	s2.SetTricksWonForTest(0, 1)
	s2.SetPointsForTest(0, 2) // J を 1 枚含むトリック
	s2.FinishHandForTest()
	if s2.GetLastResult().Made {
		t.Error("a single card point breaks a misère")
	}
}

// TestSixBidSoloWidowCountsToTheDeclarerExceptAtMisere covers the widow rule the
// issue omits entirely.
func TestSixBidSoloWidowCountsToTheDeclarerExceptAtMisere(t *testing.T) {
	widow := []*Card{sbsCard(CardDesignSpade, 1), sbsCard(CardDesignHeart, 10), sbsCard(CardDesignClover, 13)}
	widowPts := 11 + 10 + 4 // 25

	t.Run("a suit contract picks it up", func(t *testing.T) {
		s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
		s.SetWidowForTest(widow)
		if got := s.SixBidSoloWidowPoints(); got != widowPts {
			t.Fatalf("the widow is worth %d, want %d", got, widowPts)
		}
		// 場で 40 点しか取れなくても、ウィドウ 25 を足せば 65 で足りる。
		s.SetPointsForTest(0, 40)
		s.FinishHandForTest()

		res := s.GetLastResult()
		if got := res.WidowPoints; got != widowPts {
			t.Errorf("the settlement records %d widow points, want %d", got, widowPts)
		}
		if got := res.DeclarerPoints; got != 65 {
			t.Errorf("the declarer holds %d, want 40 + %d", got, widowPts)
		}
		if !res.Made {
			t.Error("65 beats 60 — the widow is what carries it")
		}
	})

	// **ミゼール系だけは加算しない。**
	t.Run("a misère does not", func(t *testing.T) {
		for _, kind := range []SixBidSoloBidKind{SixBidSoloBidMisere, SixBidSoloBidSpreadMisere} {
			s := sbsPlaying(t, 0, kind, 0)
			s.SetWidowForTest(widow)
			s.SetPointsForTest(0, 0)
			s.FinishHandForTest()

			res := s.GetLastResult()
			if got := res.WidowPoints; got != 0 {
				t.Errorf("%v recorded %d widow points, want 0", kind, got)
			}
			if !res.Made {
				t.Errorf("%v must survive a point-heavy widow — it is excluded", kind)
			}
		}
	})

	if SixBidSoloUsesWidow(SixBidSoloBidPass) {
		t.Error("a pass never picks up the widow")
	}
}

// TestSixBidSoloSettlement covers the values, which the issue omits.
func TestSixBidSoloSettlement(t *testing.T) {
	// **通常ビッドは固定額ではない。**60 点との差 × 倍率。
	t.Run("a simple solo pays twice the difference from 60", func(t *testing.T) {
		s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
		s.SetWidowForTest([]*Card{})
		s.SetPointsForTest(0, 75)
		s.FinishHandForTest()

		res := s.GetLastResult()
		if !res.Made {
			t.Fatal("75 beats 60")
		}
		if got := res.Value; got != (75-60)*SixBidSoloSoloMultiplier {
			t.Errorf("value = %d, want (75 - 60) x 2 = 30", got)
		}
		// **対戦者 2 人ぶん受け取る。**
		if got := res.Deltas[0]; got != 60 {
			t.Errorf("the declarer gains %d, want 30 from each of two opponents", got)
		}
		if got := res.Deltas[1]; got != -30 {
			t.Errorf("an opponent loses %d, want 30", got)
		}
	})

	t.Run("a heart solo pays three times the difference", func(t *testing.T) {
		s := sbsPlaying(t, 0, SixBidSoloBidHeartSolo, CardDesignHeart)
		s.SetWidowForTest([]*Card{})
		s.SetPointsForTest(0, 75)
		s.FinishHandForTest()
		if got := s.GetLastResult().Value; got != (75-60)*SixBidSoloHeartMultiplier {
			t.Errorf("value = %d, want (75 - 60) x 3 = 45", got)
		}
	})

	// 未達なら向きが反転する。
	t.Run("a failed contract reverses the payment", func(t *testing.T) {
		s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
		s.SetWidowForTest([]*Card{})
		s.SetPointsForTest(0, 40)
		s.FinishHandForTest()

		res := s.GetLastResult()
		if res.Made {
			t.Fatal("40 does not beat 60")
		}
		// 差は |40 - 60| = 20、× 2 = 40。
		if got := res.Value; got != 40 {
			t.Errorf("value = %d, want |40 - 60| x 2 = 40", got)
		}
		if got := res.Deltas[0]; got != -80 {
			t.Errorf("the declarer pays %d, want 40 to each of two opponents", got)
		}
		if got := res.Deltas[1]; got != 40 {
			t.Errorf("an opponent gains %d, want 40", got)
		}
	})

	// 固定額のビッド。
	t.Run("the fixed values", func(t *testing.T) {
		for _, tc := range []struct {
			kind  SixBidSoloBidKind
			trump int
			want  int
		}{
			{SixBidSoloBidMisere, 0, SixBidSoloMisereValue},
			{SixBidSoloBidGuarantee, CardDesignSpade, SixBidSoloGuaranteeValue},
			{SixBidSoloBidSpreadMisere, 0, SixBidSoloSpreadValue},
			{SixBidSoloBidCall, CardDesignSpade, SixBidSoloCallValue},
			// **♥ のコール・ソロだけ 150。**
			{SixBidSoloBidCall, CardDesignHeart, SixBidSoloCallHeartValue},
		} {
			s := sbsPlaying(t, 0, tc.kind, tc.trump)
			s.SetWidowForTest([]*Card{})
			s.SetPointsForTest(0, 0)
			s.FinishHandForTest()
			if got := s.GetLastResult().Value; got != tc.want {
				t.Errorf("%v at suit %d pays %d, want %d", tc.kind, tc.trump, got, tc.want)
			}
		}
	})

	// **ミゼール 30 < ギャランティー 40。**額の上でも序列が保たれる。
	t.Run("the fixed values keep the ladder monotone", func(t *testing.T) {
		if SixBidSoloMisereValue >= SixBidSoloGuaranteeValue {
			t.Errorf("misère %d must sit below guarantee %d — that is why pagat's 30 is used, not Wikipedia's 40",
				SixBidSoloMisereValue, SixBidSoloGuaranteeValue)
		}
		if SixBidSoloGuaranteeValue >= SixBidSoloSpreadValue {
			t.Error("guarantee must sit below spread misère")
		}
		if SixBidSoloSpreadValue >= SixBidSoloCallValue {
			t.Error("spread misère must sit below call solo")
		}
	})
}

// TestSixBidSoloBidding covers the auction.
func TestSixBidSoloBidding(t *testing.T) {
	t.Run("the dealer's left bids first", func(t *testing.T) {
		s := NewDefaultSixBidSolo()
		s.Reset()
		if got := s.GetBidPlayerIdx(); got != 1 {
			t.Errorf("bid seat = %d, want 1", got)
		}
	})

	// **上回る宣言だけが通る。**同額で奪える席は無い。
	t.Run("a bid must beat the standing one", func(t *testing.T) {
		s := NewDefaultSixBidSolo()
		s.Reset()
		s.SetBidPlayerForTest(1)
		if err := s.Bid(1, SixBidSoloBidMisere); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		if s.SixBidSoloCanBid(2, SixBidSoloBidMisere) {
			t.Error("an equal bid must not stand")
		}
		if s.SixBidSoloCanBid(2, SixBidSoloBidSolo) {
			t.Error("a lower bid must not stand")
		}
		if !s.SixBidSoloCanBid(2, SixBidSoloBidGuarantee) {
			t.Error("a higher bid may stand")
		}
		if s.SixBidSoloCanBid(2, SixBidSoloBidPass) {
			t.Error("pass is not a bid you place")
		}
		if s.SixBidSoloCanBid(2, SixBidSoloBidCount) {
			t.Error("an out-of-range bid must be refused")
		}
		if s.SixBidSoloCanBid(99, SixBidSoloBidCall) {
			t.Error("an unknown seat cannot bid")
		}
	})

	// 落札すると切札の指定へ進む。
	t.Run("winning the auction asks for a trump", func(t *testing.T) {
		s := NewDefaultSixBidSolo()
		s.Reset()
		s.SetBidPlayerForTest(1)
		if err := s.Bid(1, SixBidSoloBidSolo); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		_ = s.PassBid(2)
		_ = s.PassBid(0)
		if got := s.GetPhase(); got != SixBidSoloPhaseDeclare {
			t.Errorf("phase = %v, want the declaration", got)
		}
		if got := s.GetDeclarerIdx(); got != 1 {
			t.Errorf("declarer = %d, want 1", got)
		}
		if s.IsDeclared() {
			t.Error("nothing is settled until the declarer names it")
		}
	})

	// **ハート・ソロは切札が固定なのでそのままプレイへ。**
	t.Run("a heart solo needs no declaration", func(t *testing.T) {
		s := NewDefaultSixBidSolo()
		s.Reset()
		s.SetBidPlayerForTest(1)
		if err := s.Bid(1, SixBidSoloBidHeartSolo); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		_ = s.PassBid(2)
		_ = s.PassBid(0)
		if got := s.GetPhase(); got != SixBidSoloPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
		if got := s.GetTrumpSuit(); got != CardDesignHeart {
			t.Errorf("trump = %d, want hearts", got)
		}
		if SixBidSoloNeedsTrump(SixBidSoloBidHeartSolo) {
			t.Error("a heart solo does not choose its trump")
		}
	})

	// ミゼール系は切札なしでプレイへ。
	t.Run("a misère plays at no trump", func(t *testing.T) {
		s := NewDefaultSixBidSolo()
		s.Reset()
		s.SetBidPlayerForTest(1)
		if err := s.Bid(1, SixBidSoloBidMisere); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		_ = s.PassBid(2)
		_ = s.PassBid(0)
		if got := s.GetPhase(); got != SixBidSoloPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
		if got := s.GetTrumpSuit(); got != 0 {
			t.Errorf("trump = %d, want none", got)
		}
	})

	t.Run("three passes redeal without advancing the hand number", func(t *testing.T) {
		s := NewDefaultSixBidSolo()
		s.Reset()
		hand := s.GetHandNumber()
		for _, seat := range []int{1, 2, 0} {
			_ = s.PassBid(seat)
		}
		if got := s.GetPhase(); got != SixBidSoloPhaseBid {
			t.Errorf("phase = %v, want a fresh auction", got)
		}
		if got := s.GetHandNumber(); got != hand {
			t.Errorf("hand number = %d, want %d", got, hand)
		}
		// 配り直しても 11 + 3 の形は変わらない。
		for i := range SixBidSoloPlayerCnt {
			if got := s.GetPlayer(i).GetCardsSize(); got != SixBidSoloHandSize {
				t.Errorf("seat %d holds %d after the redeal, want %d", i, got, SixBidSoloHandSize)
			}
		}
		if got := len(s.GetWidow()); got != SixBidSoloWidowSize {
			t.Errorf("the widow holds %d after the redeal, want %d", got, SixBidSoloWidowSize)
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		s := NewDefaultSixBidSolo()
		s.Reset()
		if err := s.Bid(0, SixBidSoloBidSolo); err == nil {
			t.Error("bidding out of turn must be refused")
		}
		if err := s.Bid(s.GetBidPlayerIdx(), SixBidSoloBidPass); err == nil {
			t.Error("pass is not a bid you place")
		}
		s.SetPhaseForTest(SixBidSoloPhasePlay)
		if err := s.PassBid(1); err == nil {
			t.Error("bidding outside the auction must be refused")
		}
	})
}

// TestSixBidSoloDeclare covers naming the trump.
func TestSixBidSoloDeclare(t *testing.T) {
	setup := func(t *testing.T, kind SixBidSoloBidKind) *SixBidSolo {
		t.Helper()
		s := NewDefaultSixBidSolo()
		s.Reset()
		s.SetBidPlayerForTest(1)
		if err := s.Bid(1, kind); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		_ = s.PassBid(2)
		_ = s.PassBid(0)
		return s
	}

	t.Run("naming a suit starts the play", func(t *testing.T) {
		s := setup(t, SixBidSoloBidSolo)
		if err := s.Declare(1, CardDesignDiamond, nil); err != nil {
			t.Fatalf("Declare: %v", err)
		}
		if got := s.GetTrumpSuit(); got != CardDesignDiamond {
			t.Errorf("trump = %d, want diamonds", got)
		}
		if got := s.GetPhase(); got != SixBidSoloPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
		// **リードは親の左であって落札者ではない。**
		if got := s.GetTrickLeaderIdx(); got != (s.GetDealerIdx()+1)%SixBidSoloPlayerCnt {
			t.Errorf("leader = %d, want the seat to the dealer's left", got)
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		s := setup(t, SixBidSoloBidSolo)
		if err := s.Declare(2, CardDesignSpade, nil); err == nil {
			t.Error("only the declarer names the trump")
		}
		if err := s.Declare(1, 0, nil); err == nil {
			t.Error("a bad suit must be refused")
		}
		if err := s.Declare(1, 9, nil); err == nil {
			t.Error("a bad suit must be refused")
		}
		if err := s.Declare(1, CardDesignSpade, nil); err != nil {
			t.Fatalf("Declare: %v", err)
		}
		if err := s.Declare(1, CardDesignHeart, nil); err == nil {
			t.Error("the trump cannot be renamed once play starts")
		}
	})

	// **コール・ソロは札を指名しなければ成立しない。**
	t.Run("a call solo must name a card", func(t *testing.T) {
		s := setup(t, SixBidSoloBidCall)
		if err := s.Declare(1, CardDesignSpade, nil); err == nil {
			t.Error("a call solo without a named card must be refused")
		}
	})
}

// TestSixBidSoloCallSoloExchangesTheNamedCard covers the rule the issue omits.
//
// **持っている者は交換に応じる義務がある。**
func TestSixBidSoloCallSoloExchangesTheNamedCard(t *testing.T) {
	s := NewDefaultSixBidSolo()
	s.Reset()
	s.SetPhaseForTest(SixBidSoloPhaseDeclare)
	s.SetContractForTest(0, SixBidSoloBidCall, 0)

	want := sbsCard(CardDesignSpade, 1)
	// 宣言者は ♠A を持たず、席 1 が持っている。
	s.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 6), sbsCard(CardDesignHeart, 7)})
	s.SetHandForTest(1, []*Card{want, sbsCard(CardDesignClover, 8)})
	s.SetHandForTest(2, []*Card{sbsCard(CardDesignDiamond, 9)})

	if err := s.Declare(0, CardDesignSpade, want); err != nil {
		t.Fatalf("Declare: %v", err)
	}

	if got := s.GetCalledCard(); got == nil || got.GetDesign() != CardDesignSpade || got.GetValue() != 1 {
		t.Fatalf("the called card was not recorded: %v", got)
	}
	if sixBidSoloIndexOf(s.GetPlayer(0), want) < 0 {
		t.Error("the declarer must receive the named card")
	}
	if sixBidSoloIndexOf(s.GetPlayer(1), want) >= 0 {
		t.Error("the holder must give the named card up")
	}
	// **交換なので枚数は動かない。**
	if got := s.GetPlayer(0).GetCardsSize(); got != 2 {
		t.Errorf("the declarer holds %d, want 2 — it is an exchange", got)
	}
	if got := s.GetPlayer(1).GetCardsSize(); got != 2 {
		t.Errorf("the holder holds %d, want 2 — it is an exchange", got)
	}
}

// **ウィドウにあったときは交換が起こらない。**
func TestSixBidSoloCallSoloNoExchangeWhenTheCardIsInTheWidow(t *testing.T) {
	s := NewDefaultSixBidSolo()
	s.Reset()
	s.SetPhaseForTest(SixBidSoloPhaseDeclare)
	s.SetContractForTest(0, SixBidSoloBidCall, 0)

	want := sbsCard(CardDesignSpade, 1)
	s.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 6)})
	s.SetHandForTest(1, []*Card{sbsCard(CardDesignClover, 8)})
	s.SetHandForTest(2, []*Card{sbsCard(CardDesignDiamond, 9)})
	s.SetWidowForTest([]*Card{want})

	if err := s.Declare(0, CardDesignSpade, want); err != nil {
		t.Fatalf("Declare: %v", err)
	}
	for i := range SixBidSoloPlayerCnt {
		if got := s.GetPlayer(i).GetCardsSize(); got != 1 {
			t.Errorf("seat %d holds %d, want 1 — no exchange happens", i, got)
		}
	}
}

// 自分が既に持っている札は指名できない。
func TestSixBidSoloCallSoloCannotNameAHeldCard(t *testing.T) {
	s := NewDefaultSixBidSolo()
	s.Reset()
	s.SetPhaseForTest(SixBidSoloPhaseDeclare)
	s.SetContractForTest(0, SixBidSoloBidCall, 0)
	held := sbsCard(CardDesignSpade, 1)
	s.SetHandForTest(0, []*Card{held})
	if err := s.Declare(0, CardDesignSpade, held); err == nil {
		t.Error("naming a card you already hold must be refused")
	}
}

// TestSixBidSoloFollowSuitIsCompulsory covers the play restriction.
func TestSixBidSoloFollowSuitIsCompulsory(t *testing.T) {
	s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	s.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 1)})
	s.SetHandForTest(1, []*Card{sbsCard(CardDesignHeart, 6), sbsCard(CardDesignClover, 1)})
	if err := s.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	valid := s.SixBidSoloValidPlays(1)
	if len(valid) != 1 || valid[0] != 0 {
		t.Errorf("valid = %v, want only the heart", valid)
	}
	if err := s.PlayCard(1, 1); err == nil {
		t.Error("discarding while holding the led suit must be refused")
	}

	// 持っていなければ何でも出せる。
	s2 := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	s2.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 1)})
	s2.SetHandForTest(1, []*Card{sbsCard(CardDesignClover, 6), sbsCard(CardDesignSpade, 6)})
	if err := s2.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := len(s2.SixBidSoloValidPlays(1)); got != 2 {
		t.Errorf("%d plays are legal, want both — the led suit is void", got)
	}
	if s.SixBidSoloValidPlays(99) != nil {
		t.Error("an unknown seat has no legal plays")
	}
}

// 切札が勝ち、点はトリックの取り手に入る。
func TestSixBidSoloTrickResolution(t *testing.T) {
	s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	s.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 1)})  // 11 点
	s.SetHandForTest(1, []*Card{sbsCard(CardDesignSpade, 6)})  // 0 点、切札
	s.SetHandForTest(2, []*Card{sbsCard(CardDesignHeart, 13)}) // 4 点
	for _, seat := range []int{0, 1, 2} {
		if err := s.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	if got := s.GetTricksWon(1); got != 1 {
		t.Errorf("the low trump takes the trick, seat 1 has %d", got)
	}
	if got := s.GetPoints(1); got != 15 {
		t.Errorf("seat 1 took %d points, want 11 + 4", got)
	}
	if got := s.GetTrickNumber(); got != 1 {
		t.Errorf("%d tricks played, want 1", got)
	}
	// 次のリードは取り手。
	if got := s.GetTrickLeaderIdx(); got != 1 {
		t.Errorf("leader = %d, want the winner", got)
	}

	// ノートランプでは平の序列だけで決まる。
	nt := sbsPlaying(t, 0, SixBidSoloBidMisere, 0)
	nt.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 13)})
	nt.SetHandForTest(1, []*Card{sbsCard(CardDesignSpade, 1)})
	nt.SetHandForTest(2, []*Card{sbsCard(CardDesignHeart, 10)})
	for _, seat := range []int{0, 1, 2} {
		if err := nt.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	// **♠A は追随していないので効かない。**♥10 が ♥K を上回る。
	if got := nt.GetTricksWon(2); got != 1 {
		t.Errorf("the ten of the led suit takes it, seat 2 has %d", got)
	}
}

func TestSixBidSoloPlayGuards(t *testing.T) {
	s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	s.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 1)})
	if err := s.PlayCard(1, 0); err == nil {
		t.Error("playing out of turn must be refused")
	}
	if err := s.PlayCard(0, 99); err == nil {
		t.Error("an out-of-range index must be refused")
	}
	s.SetPhaseForTest(SixBidSoloPhaseBid)
	if err := s.PlayCard(0, 0); err == nil {
		t.Error("playing outside the play phase must be refused")
	}
}

// **スプレッド・ミゼールは他の 2 人が 1 枚ずつ出したら公開する。**
func TestSixBidSoloSpreadMisereOpensAfterTwoCards(t *testing.T) {
	s := sbsPlaying(t, 1, SixBidSoloBidSpreadMisere, 0)
	s.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 6)})
	s.SetHandForTest(1, []*Card{sbsCard(CardDesignHeart, 7)})
	s.SetHandForTest(2, []*Card{sbsCard(CardDesignHeart, 8)})

	if s.IsSpreadOpen() {
		t.Fatal("the hand starts concealed")
	}
	if err := s.PlayCard(0, 0); err != nil {
		t.Fatalf("PlayCard: %v", err)
	}
	if s.IsSpreadOpen() {
		t.Error("one card is not enough to open the hand")
	}
	if err := s.PlayCard(1, 0); err != nil {
		t.Fatalf("PlayCard: %v", err)
	}
	if !s.IsSpreadOpen() {
		t.Error("the declarer's hand must be exposed after the second card")
	}

	// 通常ビッドでは公開されない。
	plain := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	plain.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 6)})
	plain.SetHandForTest(1, []*Card{sbsCard(CardDesignHeart, 7)})
	_ = plain.PlayCard(0, 0)
	_ = plain.PlayCard(1, 0)
	if plain.IsSpreadOpen() {
		t.Error("only a spread misère exposes the hand")
	}
}

func TestSixBidSoloNextHandRotatesTheDealer(t *testing.T) {
	s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	s.SetPointsForTest(0, 70)
	s.FinishHandForTest()

	dealer := s.GetDealerIdx()
	if err := s.NextHand(); err != nil {
		t.Fatalf("NextHand: %v", err)
	}
	if got := s.GetDealerIdx(); got == dealer {
		t.Errorf("the dealer stayed at %d; it must rotate", got)
	}
	for i := range SixBidSoloPlayerCnt {
		if got := s.GetPlayer(i).GetCardsSize(); got != SixBidSoloHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, SixBidSoloHandSize)
		}
	}
}

func TestSixBidSoloNextHandGuards(t *testing.T) {
	s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	if err := s.NextHand(); err == nil {
		t.Error("dealing again mid-hand must be refused")
	}
}

// **規定局数で終わり、首位が勝つ。**
func TestSixBidSoloGameEndsAfterTheTargetHands(t *testing.T) {
	s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	s.SetHandNumberForTest(s.GetConfig().TargetHands)
	s.SetWidowForTest([]*Card{})
	s.SetPointsForTest(0, 90)
	s.FinishHandForTest()

	if !s.GetGameEndFlag() {
		t.Fatal("the game ends once the target hand is settled")
	}
	if got := s.GetWinnerIdx(); got != 0 {
		t.Errorf("winner = %d, want the seat in front", got)
	}
	if err := s.NextHand(); err == nil {
		t.Error("dealing after the game is over must be refused")
	}
	if got := s.GetPhase(); got != SixBidSoloPhaseGameEnd {
		t.Errorf("phase = %v, want game end", got)
	}
}

// 精算が付かない局 (落札者なし) でも落ちない。
func TestSixBidSoloFinishWithoutADeclarer(t *testing.T) {
	s := NewDefaultSixBidSolo()
	s.Reset()
	s.SetPhaseForTest(SixBidSoloPhasePlay)
	s.FinishHandForTest()
	if got := s.GetPhase(); got != SixBidSoloPhaseHandEnd {
		t.Errorf("phase = %v, want hand end", got)
	}
	if s.GetLastResult() != nil {
		t.Error("no contract means no settlement")
	}
}

func TestSixBidSoloIsHumanTurnAndCpuPlay(t *testing.T) {
	s := NewDefaultSixBidSolo()
	s.Reset()
	if s.IsHumanTurn() {
		t.Error("the dealer's left bids first and it is a CPU")
	}
	s.CpuPlay()
	if s.GetBidPlayerIdx() == 1 && s.GetPhase() == SixBidSoloPhaseBid {
		t.Error("CpuPlay must move the auction along")
	}

	over := NewDefaultSixBidSolo()
	over.Reset()
	over.SetPhaseForTest(SixBidSoloPhaseGameEnd)
	over.gameEndFlag = true
	if over.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	over.CpuPlay()

	// 精算画面は誰の手番でもない。
	settled := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	settled.SetPhaseForTest(SixBidSoloPhaseHandEnd)
	if settled.IsHumanTurn() {
		t.Error("the settlement is nobody's turn")
	}
}

// **CPU だけで 1 局を回し切れること。**途中で止まると詰む。
func TestSixBidSoloCpuDrivesAFullHand(t *testing.T) {
	for attempt := range 30 {
		s := NewDefaultSixBidSolo()
		s.Reset()
		for step := 0; step < 400; step++ {
			if s.GetPhase() == SixBidSoloPhaseHandEnd || s.GetGameEndFlag() {
				break
			}
			if !s.IsHumanTurn() {
				s.CpuPlay()
				continue
			}
			switch s.GetPhase() {
			case SixBidSoloPhaseBid:
				idx := s.GetBidPlayerIdx()
				kind := s.SixBidSoloCpuBid(idx)
				if kind == SixBidSoloBidPass || s.Bid(idx, kind) != nil {
					_ = s.PassBid(idx)
				}
			case SixBidSoloPhaseDeclare:
				idx := s.GetDeclarerIdx()
				suit := s.SixBidSoloCpuTrump(idx)
				var called *Card
				if s.GetHighBid() != nil && s.GetHighBid().Kind == SixBidSoloBidCall {
					called = s.SixBidSoloCpuCall(idx, suit)
				}
				_ = s.Declare(idx, suit, called)
			case SixBidSoloPhasePlay:
				idx := s.GetCurrentPlayerIdx()
				if i := s.SixBidSoloCpuPlay(idx); i >= 0 {
					_ = s.PlayCard(idx, i)
				}
			}
		}
		if s.GetPhase() != SixBidSoloPhaseHandEnd && !s.GetGameEndFlag() {
			t.Fatalf("attempt %d: the hand never finished (phase %v)", attempt, s.GetPhase())
		}
		if got := s.GetTrickNumber(); got != SixBidSoloTricks {
			t.Fatalf("attempt %d: %d tricks played, want %d", attempt, got, SixBidSoloTricks)
		}
		// **場に出た点は 120 からウィドウのぶんを引いた額。**
		played := 0
		for i := range SixBidSoloPlayerCnt {
			played += s.GetPoints(i)
		}
		if want := SixBidSoloTotalPoints - s.SixBidSoloWidowPoints(); played != want {
			t.Fatalf("attempt %d: %d points were taken, want %d", attempt, played, want)
		}
	}
}

func TestSixBidSoloCpuEdges(t *testing.T) {
	s := sbsPlaying(t, 0, SixBidSoloBidSolo, CardDesignSpade)
	s.SetHandForTest(0, []*Card{})
	if got := s.SixBidSoloCpuPlay(0); got != -1 {
		t.Errorf("an empty hand has no play, got %d", got)
	}
	if got := s.SixBidSoloCpuPlay(99); got != -1 {
		t.Errorf("an unknown seat has no play, got %d", got)
	}
	if got := s.SixBidSoloCpuBid(99); got != SixBidSoloBidPass {
		t.Errorf("an unknown seat bids %v, want pass", got)
	}
	if got := s.SixBidSoloCpuTrump(99); got != CardDesignSpade {
		t.Errorf("an unknown seat falls back to %d", got)
	}
	if s.SixBidSoloCpuCall(99, CardDesignSpade) != nil {
		t.Error("an unknown seat names nothing")
	}
	// **指名するのは持っていない札。**
	s.SetHandForTest(1, []*Card{sbsCard(CardDesignSpade, 1), sbsCard(CardDesignSpade, 10)})
	got := s.SixBidSoloCpuCall(1, CardDesignSpade)
	if got == nil || sixBidSoloIndexOf(s.GetPlayer(1), got) >= 0 {
		t.Errorf("the CPU named %v, which it already holds", got)
	}
}

// ミゼールの宣言者はいちばん弱い札を捨てる。
func TestSixBidSoloCpuDucksOnMisere(t *testing.T) {
	s := sbsPlaying(t, 1, SixBidSoloBidMisere, 0)
	s.SetHandForTest(0, []*Card{sbsCard(CardDesignHeart, 12)})
	s.SetHandForTest(1, []*Card{sbsCard(CardDesignHeart, 1), sbsCard(CardDesignHeart, 6)})
	if err := s.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := s.SixBidSoloCpuPlay(1); got != 1 {
		t.Errorf("the misère declarer played index %d, want the low card", got)
	}
}

func TestSixBidSoloAccessors(t *testing.T) {
	s := NewDefaultSixBidSolo()
	s.Reset()
	if got := s.GetHandNumber(); got != 1 {
		t.Errorf("hand number = %d, want 1", got)
	}
	if got := s.GetWinnerIdx(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if got := s.GetDeclarerIdx(); got != -1 {
		t.Errorf("declarer = %d, want -1 during the auction", got)
	}
	if s.GetHighBid() != nil {
		t.Error("no bid stands at the start")
	}
	if s.GetLastResult() != nil {
		t.Error("there is no result before the first settlement")
	}
	if s.GetCalledCard() != nil {
		t.Error("no card is called until a call solo declares")
	}
	if got := len(s.GetPlayers()); got != SixBidSoloPlayerCnt {
		t.Errorf("%d seats, want %d", got, SixBidSoloPlayerCnt)
	}
	if s.GetPlayer(-1) != nil || s.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
	}
	if got := len(s.GetTrick()); got != 0 {
		t.Errorf("the trick starts empty, got %d", got)
	}
	if len(s.GetActionLog()) == 0 {
		t.Error("dealing writes to the action log")
	}
	if got := len(s.GetBids()); got != 0 {
		t.Errorf("the bid history starts empty, got %d", got)
	}
	// 範囲外の席は 0。
	for _, idx := range []int{-1, SixBidSoloPlayerCnt} {
		if s.GetPoints(idx) != 0 || s.GetTricksWon(idx) != 0 || s.GetScore(idx) != 0 {
			t.Errorf("seat %d must read as zero", idx)
		}
	}
	s.SetPointsForTest(1, 30)
	s.SetTricksWonForTest(1, 2)
	if s.GetPoints(1) != 30 || s.GetTricksWon(1) != 2 {
		t.Error("the test setters must take effect")
	}
	s.SetPointsForTest(99, 5)
	s.SetTricksWonForTest(99, 5)
	s.SetHandForTest(99, nil)
	s.SetWidowForTest(nil)

	cfg := s.GetConfig()
	if cfg.TargetHands != SixBidSoloDefaultHands {
		t.Errorf("target hands = %d, want %d", cfg.TargetHands, SixBidSoloDefaultHands)
	}
	cfg.TargetHands = 8
	s.SetConfig(cfg)
	if s.GetConfig().TargetHands != 8 {
		t.Error("SetConfig must take effect")
	}
	// 棋譜用の内部名は 6 段階ぶんそろっている。
	for kind, want := range map[SixBidSoloBidKind]string{
		SixBidSoloBidPass:         "pass",
		SixBidSoloBidSolo:         "solo",
		SixBidSoloBidHeartSolo:    "heartSolo",
		SixBidSoloBidMisere:       "misere",
		SixBidSoloBidGuarantee:    "guarantee",
		SixBidSoloBidSpreadMisere: "spreadMisere",
		SixBidSoloBidCall:         "callSolo",
	} {
		if got := sixBidSoloBidName(kind); got != want {
			t.Errorf("bid name for %v = %q, want %q", kind, got, want)
		}
	}
	for suit, want := range map[int]string{
		CardDesignSpade: "S", CardDesignClover: "C", CardDesignHeart: "H", CardDesignDiamond: "D", 0: "-",
	} {
		if got := sixBidSoloSuitName(suit); got != want {
			t.Errorf("suit name for %d = %q, want %q", suit, got, want)
		}
	}

	// 通算得点は範囲内なら読める。
	s.SetScoreForTest(1, 42)
	if got := s.GetScore(1); got != 42 {
		t.Errorf("GetScore(1) = %d, want 42", got)
	}
	// 親も差し替えられる。
	s.SetDealerForTest(2)
	if got := s.GetDealerIdx(); got != 2 {
		t.Errorf("dealer = %d, want 2", got)
	}
}

func TestSixBidSoloConfigValidate(t *testing.T) {
	if err := DefaultSixBidSoloConfig().Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	if err := (SixBidSoloConfig{CpuDifficulty: 9, TargetHands: SixBidSoloDefaultHands}).Validate(); err == nil {
		t.Error("a bad difficulty must not validate")
	}
	for _, n := range []int{SixBidSoloMinHands - 1, SixBidSoloMaxHands + 1} {
		if err := (SixBidSoloConfig{TargetHands: n}).Validate(); err == nil {
			t.Errorf("%d hands must not validate", n)
		}
	}
}

func TestSixBidSoloRoundTripsThroughJSON(t *testing.T) {
	s := NewDefaultSixBidSolo()
	s.Reset()
	s.SetBidPlayerForTest(1)
	_ = s.Bid(1, SixBidSoloBidSolo)

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored SixBidSolo
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetHighBid() == nil || restored.GetHighBid().Kind != SixBidSoloBidSolo {
		t.Error("the standing bid did not survive the round trip")
	}
	if got := restored.GetPlayer(0).GetCardsSize(); got != SixBidSoloHandSize {
		t.Errorf("the restored hand holds %d, want %d", got, SixBidSoloHandSize)
	}
	// **ウィドウも往復する。**落とすと精算が狂う。
	if got := len(restored.GetWidow()); got != SixBidSoloWidowSize {
		t.Errorf("the restored widow holds %d, want %d", got, SixBidSoloWidowSize)
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestSixBidSoloRejectsBadJSON(t *testing.T) {
	base := `"pl":[{},{},{}],"cf":{"cd":0,"th":6},"ph":0,"di":0,"bi":0,"ci":0,"tl":0,"de":-1,"wi":-1,"ts":0`
	for _, tc := range []struct {
		name string
		body string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[],"cf":{"cd":0,"th":6},"ph":0}`},
		{"bad phase", `{` + base + `,"ph":99}`},
		{"bad dealer", `{` + base + `,"di":9}`},
		{"bad bid seat", `{` + base + `,"bi":9}`},
		{"bad current seat", `{` + base + `,"ci":9}`},
		{"bad trick leader", `{` + base + `,"tl":9}`},
		{"bad declarer", `{` + base + `,"de":9}`},
		{"bad winner", `{` + base + `,"wi":9}`},
		{"bad trump suit", `{` + base + `,"ts":9}`},
		{"oversized trick", `{` + base + `,"tk":[{},{},{},{}]}`},
		{"oversized widow", `{` + base + `,"wd":[{},{},{},{}]}`},
		{"unknown bid", `{` + base + `,"hb":{"Player":0,"Kind":99}}`},
		{"bad trick number", `{` + base + `,"tn":99}`},
		{"bad config", `{"pl":[{},{},{}],"cf":{"cd":0,"th":99},"ph":0,"di":0,"bi":0,"ci":0,"tl":0,"de":-1,"wi":-1,"ts":0}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var s SixBidSolo
			if err := json.Unmarshal([]byte(tc.body), &s); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}
