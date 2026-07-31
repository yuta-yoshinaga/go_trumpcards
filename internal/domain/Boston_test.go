//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"
)

func bsCard(suit, value int) *Card { return NewCard(suit, value, true) }

// bsPlaying puts a game into the play phase with a fixed contract.
func bsPlaying(t *testing.T, declarer, partner int, level BostonBidLevel, suit int) *Boston {
	t.Helper()
	b := NewDefaultBoston()
	b.Reset()
	b.SetPhaseForTest(BostonPhasePlay)
	b.SetContractForTest(declarer, partner, level, suit)
	b.SetCurrentPlayerForTest(0)
	b.SetTrickLeaderForTest(0)
	return b
}

// TestBostonBidLadderInterleavesMisereWithTrickBids は issue の最大の誤りを
// 押さえる。
//
// **Little Misère は 7 トリックより下、Grand Misère は 9 トリックより下。**
// issue のようにトリック宣言を並べたあとにミゼールを置くと競りが変わる。
func TestBostonBidLadderInterleavesMisereWithTrickBids(t *testing.T) {
	ascending := []BostonBidLevel{
		BostonBidFive,
		BostonBidSix,
		BostonBidLittleMisere,
		BostonBidSeven,
		BostonBidPiccolissimo,
		BostonBidEight,
		BostonBidGrandMisere,
		BostonBidNine,
		BostonBidLittleMisereTable,
		BostonBidTen,
		BostonBidGrandMisereTable,
		BostonBidEleven,
		BostonBidTwelve,
		BostonBidChelem,
		BostonBidChelemTable,
	}
	if len(ascending) != int(BostonBidLevelCount)-1 {
		t.Fatalf("the ladder lists %d bids, the table holds %d", len(ascending), int(BostonBidLevelCount)-1)
	}
	for i := 1; i < len(ascending); i++ {
		if ascending[i-1] >= ascending[i] {
			t.Errorf("%s must rank below %s", BostonBidName(ascending[i-1]), BostonBidName(ascending[i]))
		}
	}

	// **これが issue との決定的な差。**
	if BostonBidLittleMisere >= BostonBidSeven {
		t.Error("Little Misere ranks BELOW seven tricks")
	}
	if BostonBidLittleMisere <= BostonBidSix {
		t.Error("Little Misere ranks above six tricks")
	}
	if BostonBidGrandMisere >= BostonBidNine {
		t.Error("Grand Misere ranks BELOW nine tricks")
	}
	if BostonBidGrandMisere <= BostonBidEight {
		t.Error("Grand Misere ranks above eight tricks")
	}
	// 公開版はそれぞれの 1 つ上ではなく、さらに高い段に入る。
	if BostonBidLittleMisereTable <= BostonBidNine {
		t.Error("Little Misere on the Table ranks above nine tricks")
	}
	if BostonBidGrandMisereTable <= BostonBidTen {
		t.Error("Grand Misere on the Table ranks above ten tricks")
	}
}

// TestBostonPiccolissimoWantsExactlyOneTrick covers the bid the issue omits.
//
// **0 トリックでも失敗。**ミゼールともトリック宣言とも違う第 3 の型である。
func TestBostonPiccolissimoWantsExactlyOneTrick(t *testing.T) {
	if BostonBidKindOf(BostonBidPiccolissimo) != BostonKindPiccolissimo {
		t.Fatal("Piccolissimo is its own kind of bid")
	}
	if !BostonBidSucceeded(BostonBidPiccolissimo, 1) {
		t.Error("exactly one trick makes Piccolissimo")
	}
	if BostonBidSucceeded(BostonBidPiccolissimo, 0) {
		t.Error("zero tricks FAILS Piccolissimo — it is not a misere")
	}
	if BostonBidSucceeded(BostonBidPiccolissimo, 2) {
		t.Error("two tricks fails Piccolissimo")
	}
	// 序列は 7 と 8 の間。
	if BostonBidPiccolissimo <= BostonBidSeven || BostonBidPiccolissimo >= BostonBidEight {
		t.Error("Piccolissimo sits between seven and eight tricks")
	}
	// 切札なし。
	if BostonBidNeedsTrump(BostonBidPiccolissimo) {
		t.Error("Piccolissimo is played at no trump")
	}
}

func TestBostonBidSucceededPerKind(t *testing.T) {
	// トリック宣言は「以上」。
	if !BostonBidSucceeded(BostonBidFive, 5) || !BostonBidSucceeded(BostonBidFive, 9) {
		t.Error("a trick bid is made by taking AT LEAST that many")
	}
	if BostonBidSucceeded(BostonBidFive, 4) {
		t.Error("four tricks fails a bid of five")
	}
	// ミゼールは 0 ちょうど。
	if !BostonBidSucceeded(BostonBidGrandMisere, 0) {
		t.Error("no tricks makes a misere")
	}
	if BostonBidSucceeded(BostonBidGrandMisere, 1) {
		t.Error("a single trick breaks a misere")
	}
	// 範囲外はパス扱いで常に失敗。
	if BostonBidSucceeded(BostonBidPass, 0) {
		t.Error("a pass is not a contract")
	}
	if BostonBidSucceeded(BostonBidLevel(99), 0) {
		t.Error("an out-of-range level must not succeed")
	}
}

// TestBostonExposedIsItsOwnBid covers the third correction.
//
// **公開はオプションではなく別の宣言。**
func TestBostonExposedIsItsOwnBid(t *testing.T) {
	if BostonBidIsExposed(BostonBidLittleMisere) {
		t.Error("plain Little Misere does not expose the hand")
	}
	if !BostonBidIsExposed(BostonBidLittleMisereTable) {
		t.Error("Little Misere on the Table exposes the hand")
	}
	if !BostonBidIsExposed(BostonBidGrandMisereTable) {
		t.Error("Grand Misere on the Table exposes the hand")
	}
	if !BostonBidIsExposed(BostonBidChelemTable) {
		t.Error("Chelem on the Table exposes the hand")
	}
	if BostonBidIsExposed(BostonBidChelem) {
		t.Error("plain Chelem does not expose the hand")
	}
	// 同じミゼールでも別段。
	if BostonBidLittleMisere == BostonBidLittleMisereTable {
		t.Error("the two Little Misere bids are distinct levels")
	}
}

// TestBostonPartnerCallingIsOnlyForTrickBids covers the fourth correction.
//
// **「各自個人戦」ではない。**トリック数の宣言 (5〜10) だけパートナーを呼べる。
func TestBostonPartnerCallingIsOnlyForTrickBids(t *testing.T) {
	for _, level := range []BostonBidLevel{
		BostonBidFive, BostonBidSix, BostonBidSeven,
		BostonBidEight, BostonBidNine, BostonBidTen,
	} {
		if !BostonBidCanCallPartner(level) {
			t.Errorf("%s may call a partner", BostonBidName(level))
		}
	}
	for _, level := range []BostonBidLevel{
		BostonBidLittleMisere, BostonBidPiccolissimo, BostonBidGrandMisere,
		BostonBidLittleMisereTable, BostonBidGrandMisereTable,
		BostonBidEleven, BostonBidTwelve, BostonBidChelem, BostonBidChelemTable,
	} {
		if BostonBidCanCallPartner(level) {
			t.Errorf("%s is played alone against three", BostonBidName(level))
		}
	}
}

// TestBostonDealsInFourFourFive covers the deal pattern the issue omits.
func TestBostonDealsInFourFourFive(t *testing.T) {
	if got := bostonDealPattern; len(got) != 3 || got[0] != 4 || got[1] != 4 || got[2] != 5 {
		t.Fatalf("deal pattern = %v, want [4 4 5]", got)
	}
	sum := 0
	for _, n := range bostonDealPattern {
		sum += n
	}
	if sum != BostonHandSize {
		t.Fatalf("the pattern deals %d cards, want %d", sum, BostonHandSize)
	}

	b := NewDefaultBoston()
	b.Reset()
	for i := range BostonPlayerCnt {
		if got := b.GetPlayer(i).GetCardsSize(); got != BostonHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, BostonHandSize)
		}
	}
	// 52 枚を配り切る。
	if BostonPlayerCnt*BostonHandSize != 52 {
		t.Errorf("%d seats x %d cards must use the whole 52-card pack", BostonPlayerCnt, BostonHandSize)
	}
	if got := len(newBostonDeck()); got != 52 {
		t.Errorf("the deck holds %d cards, want 52", got)
	}
}

func TestBostonBidding(t *testing.T) {
	t.Run("the dealer's left bids first", func(t *testing.T) {
		b := NewDefaultBoston()
		b.Reset()
		if got := b.GetBidPlayerIdx(); got != 1 {
			t.Errorf("bid seat = %d, want 1", got)
		}
	})

	t.Run("a bid must beat the standing one", func(t *testing.T) {
		b := NewDefaultBoston()
		b.Reset()
		if err := b.Bid(1, BostonBidSeven, CardDesignSpade); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		// **Little Misère は 7 より下なので通らない。**
		if err := b.Bid(2, BostonBidLittleMisere, 0); err == nil {
			t.Error("Little Misere ranks below seven tricks and must be refused")
		}
		// Piccolissimo は 7 より上なので通る。
		if err := b.Bid(2, BostonBidPiccolissimo, 0); err != nil {
			t.Errorf("Piccolissimo ranks above seven tricks: %v", err)
		}
	})

	t.Run("trick bids need a suit, misere bids do not", func(t *testing.T) {
		b := NewDefaultBoston()
		b.Reset()
		if err := b.Bid(1, BostonBidFive, 99); err == nil {
			t.Error("a trick bid needs a real suit")
		}
		if err := b.Bid(1, BostonBidFive, CardDesignHeart); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		if got := b.GetHighBid().Suit; got != CardDesignHeart {
			t.Errorf("suit = %d, want hearts", got)
		}

		// ミゼールに渡したスートは捨てられる。
		b2 := NewDefaultBoston()
		b2.Reset()
		if err := b2.Bid(1, BostonBidGrandMisere, CardDesignHeart); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		if got := b2.GetHighBid().Suit; got != 0 {
			t.Errorf("a misere carries suit %d, want 0", got)
		}
	})

	t.Run("three passes settle the contract", func(t *testing.T) {
		b := NewDefaultBoston()
		b.Reset()
		if err := b.Bid(1, BostonBidGrandMisere, 0); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		for _, seat := range []int{2, 3, 0} {
			if err := b.PassBid(seat); err != nil {
				t.Fatalf("PassBid(%d): %v", seat, err)
			}
		}
		if got := b.GetDeclarerIdx(); got != 1 {
			t.Errorf("declarer = %d, want 1", got)
		}
		// **ミゼールはパートナーを呼べないので直接プレイへ。**
		if got := b.GetPhase(); got != BostonPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
	})

	t.Run("a trick bid stops to ask about a partner", func(t *testing.T) {
		b := NewDefaultBoston()
		b.Reset()
		if err := b.Bid(1, BostonBidSix, CardDesignSpade); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		for _, seat := range []int{2, 3, 0} {
			_ = b.PassBid(seat)
		}
		if got := b.GetPhase(); got != BostonPhaseCallPartner {
			t.Errorf("phase = %v, want the partner question", got)
		}
	})

	t.Run("four passes redeal without advancing the hand number", func(t *testing.T) {
		b := NewDefaultBoston()
		b.Reset()
		hand := b.GetHandNumber()
		for _, seat := range []int{1, 2, 3, 0} {
			_ = b.PassBid(seat)
		}
		if got := b.GetPhase(); got != BostonPhaseBid {
			t.Errorf("phase = %v, want a fresh auction", got)
		}
		if got := b.GetHandNumber(); got != hand {
			t.Errorf("hand number = %d, want %d", got, hand)
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		b := NewDefaultBoston()
		b.Reset()
		if err := b.Bid(0, BostonBidFive, CardDesignSpade); err == nil {
			t.Error("bidding out of turn must be refused")
		}
		if err := b.Bid(1, BostonBidPass, 0); err == nil {
			t.Error("passing via Bid must be refused")
		}
		if err := b.Bid(1, BostonBidLevel(99), 0); err == nil {
			t.Error("an out-of-range level must be refused")
		}
		b.SetPhaseForTest(BostonPhasePlay)
		if err := b.PassBid(1); err == nil {
			t.Error("bidding outside the auction must be refused")
		}
	})
}

// TestBostonCallPartner covers the 1-vs-3 / 2-vs-2 split.
func TestBostonCallPartner(t *testing.T) {
	setup := func(t *testing.T) *Boston {
		t.Helper()
		b := NewDefaultBoston()
		b.Reset()
		if err := b.Bid(1, BostonBidSix, CardDesignSpade); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		for _, seat := range []int{2, 3, 0} {
			_ = b.PassBid(seat)
		}
		return b
	}

	t.Run("calling a partner makes it two against two", func(t *testing.T) {
		b := setup(t)
		if err := b.CallPartner(1, 3); err != nil {
			t.Fatalf("CallPartner: %v", err)
		}
		if got := b.GetPartnerIdx(); got != 3 {
			t.Errorf("partner = %d, want 3", got)
		}
		if !b.BostonIsDeclarerSide(1) || !b.BostonIsDeclarerSide(3) {
			t.Error("both the declarer and the partner are on the declaring side")
		}
		if b.BostonIsDeclarerSide(0) || b.BostonIsDeclarerSide(2) {
			t.Error("the other two defend")
		}
		if got := b.GetPhase(); got != BostonPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
	})

	t.Run("going alone leaves one against three", func(t *testing.T) {
		b := setup(t)
		if err := b.CallPartner(1, -1); err != nil {
			t.Fatalf("CallPartner: %v", err)
		}
		if got := b.GetPartnerIdx(); got != -1 {
			t.Errorf("partner = %d, want -1", got)
		}
		for _, seat := range []int{0, 2, 3} {
			if b.BostonIsDeclarerSide(seat) {
				t.Errorf("seat %d defends when the declarer is alone", seat)
			}
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		b := setup(t)
		if err := b.CallPartner(2, 3); err == nil {
			t.Error("only the declarer calls a partner")
		}
		if err := b.CallPartner(1, 1); err == nil {
			t.Error("the declarer cannot partner itself")
		}
		if err := b.CallPartner(1, 9); err == nil {
			t.Error("an out-of-range partner must be refused")
		}
	})

	t.Run("there is no partner question outside that phase", func(t *testing.T) {
		b := NewDefaultBoston()
		b.Reset()
		if err := b.CallPartner(0, 1); err == nil {
			t.Error("calling a partner during the auction must be refused")
		}
	})
}

// **リードはディーラーの左隣。**落札者ではない。
func TestBostonLeadIsTheDealersLeft(t *testing.T) {
	b := NewDefaultBoston()
	b.Reset()
	// 宣言はディーラーの左隣 (席 1) から始まる。席 2 に落札させるため 1 は降りる。
	if err := b.PassBid(1); err != nil {
		t.Fatalf("PassBid: %v", err)
	}
	if err := b.Bid(2, BostonBidGrandMisere, 0); err != nil {
		t.Fatalf("Bid: %v", err)
	}
	for _, seat := range []int{3, 0, 1} {
		if err := b.PassBid(seat); err != nil {
			t.Fatalf("PassBid(%d): %v", seat, err)
		}
	}
	if got := b.GetDeclarerIdx(); got != 2 {
		t.Fatalf("declarer = %d, want 2", got)
	}
	if got := b.GetTrickLeaderIdx(); got != 1 {
		t.Errorf("leader = %d, want the dealer's left (1), not the declarer (2)", got)
	}
}

// 追随は強制。切札は強制ではない。
func TestBostonFollowingSuitIsCompulsory(t *testing.T) {
	b := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
	b.SetHandForTest(0, []*Card{bsCard(CardDesignHeart, 1)})
	b.SetHandForTest(1, []*Card{
		bsCard(CardDesignHeart, 13), bsCard(CardDesignSpade, 2), bsCard(CardDesignDiamond, 5),
	})
	if err := b.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	valid := b.BostonValidPlays(1)
	if len(valid) != 1 || valid[0] != 0 {
		t.Errorf("valid = %v, want only the heart", valid)
	}
	if err := b.PlayCard(1, 1); err == nil {
		t.Error("trumping while able to follow must be refused")
	}

	// フォローできなければ何でも出せる。切札は強制ではない。
	b2 := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
	b2.SetHandForTest(0, []*Card{bsCard(CardDesignHeart, 1)})
	b2.SetHandForTest(1, []*Card{bsCard(CardDesignSpade, 2), bsCard(CardDesignDiamond, 5)})
	if err := b2.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := len(b2.BostonValidPlays(1)); got != 2 {
		t.Errorf("with no card of the suit both are legal, got %d", got)
	}
}

func TestBostonPlayGuards(t *testing.T) {
	b := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
	b.SetHandForTest(0, []*Card{bsCard(CardDesignHeart, 1)})
	if err := b.PlayCard(1, 0); err == nil {
		t.Error("playing out of turn must be refused")
	}
	if err := b.PlayCard(0, 99); err == nil {
		t.Error("an out-of-range index must be refused")
	}
	b.SetPhaseForTest(BostonPhaseBid)
	if err := b.PlayCard(0, 0); err == nil {
		t.Error("playing outside the play phase must be refused")
	}
	if b.BostonValidPlays(99) != nil {
		t.Error("an unknown seat has no legal plays")
	}
}

// **切札なしの宣言では切札が効かない。**ミゼールはノートランプ。
func TestBostonTrumpOnlyAppliesWhenTheContractHasOne(t *testing.T) {
	withTrump := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
	withTrump.SetHandForTest(0, []*Card{bsCard(CardDesignHeart, 1)})
	withTrump.SetHandForTest(1, []*Card{bsCard(CardDesignSpade, 2)})
	withTrump.SetHandForTest(2, []*Card{bsCard(CardDesignHeart, 2)})
	withTrump.SetHandForTest(3, []*Card{bsCard(CardDesignHeart, 3)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := withTrump.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	if got := withTrump.GetTricksWon(1); got != 1 {
		t.Errorf("the low trump takes the trick, tricks for seat 1 = %d", got)
	}

	noTrump := bsPlaying(t, 0, -1, BostonBidGrandMisere, 0)
	noTrump.SetHandForTest(0, []*Card{bsCard(CardDesignHeart, 1)})
	noTrump.SetHandForTest(1, []*Card{bsCard(CardDesignSpade, 2)})
	noTrump.SetHandForTest(2, []*Card{bsCard(CardDesignHeart, 2)})
	noTrump.SetHandForTest(3, []*Card{bsCard(CardDesignHeart, 3)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := noTrump.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	// 切札が無いので ♥A のリードが勝つ。
	if got := noTrump.GetTricksWon(0); got != 1 {
		t.Errorf("at no trump the led ace wins, tricks for seat 0 = %d", got)
	}
}

// TestBostonSettlement covers the per-player payment.
func TestBostonSettlement(t *testing.T) {
	t.Run("making a solo bid collects from all three", func(t *testing.T) {
		b := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
		b.SetTricksWonForTest(0, 6)
		b.FinishHandForTest()

		if !b.IsBidMade() {
			t.Fatal("six tricks makes a bid of six")
		}
		pay := BostonBidPayout(BostonBidSix)
		if got := b.GetChips(0); got != pay*3 {
			t.Errorf("the declarer gains %d, want %d from three opponents", got, pay*3)
		}
		for _, seat := range []int{1, 2, 3} {
			if got := b.GetChips(seat); got != -pay {
				t.Errorf("seat %d pays %d, want %d", seat, got, -pay)
			}
		}
	})

	t.Run("failing pays each opponent", func(t *testing.T) {
		b := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
		b.SetTricksWonForTest(0, 5)
		b.FinishHandForTest()

		if b.IsBidMade() {
			t.Fatal("five tricks fails a bid of six")
		}
		pay := BostonBidPayout(BostonBidSix)
		if got := b.GetChips(0); got != -pay*3 {
			t.Errorf("the declarer loses %d, want %d", got, -pay*3)
		}
	})

	// **パートナーの取ったトリックも数える。**2 対 2 なので合算しないと判定が狂う。
	t.Run("a partner's tricks count toward the contract", func(t *testing.T) {
		b := bsPlaying(t, 0, 2, BostonBidSix, CardDesignSpade)
		b.SetTricksWonForTest(0, 3)
		b.SetTricksWonForTest(2, 3)
		if got := b.BostonDeclarerTricks(); got != 6 {
			t.Fatalf("the declaring side took %d, want 6", got)
		}
		b.FinishHandForTest()
		if !b.IsBidMade() {
			t.Error("3 + 3 makes a bid of six")
		}
		// **払うのは相手 2 人だけ。**パートナーは払わない。
		pay := BostonBidPayout(BostonBidSix)
		if got := b.GetChips(0); got != pay*2 {
			t.Errorf("the declarer gains %d, want %d from two opponents", got, pay*2)
		}
		if got := b.GetChips(2); got != 0 {
			t.Errorf("the partner's own chips move by %d, want 0", got)
		}
	})

	t.Run("a misere is broken by a single trick", func(t *testing.T) {
		b := bsPlaying(t, 0, -1, BostonBidGrandMisere, 0)
		b.SetTricksWonForTest(0, 1)
		b.FinishHandForTest()
		if b.IsBidMade() {
			t.Error("one trick breaks a misere")
		}
	})

	// **ピッコリッシモは 0 でも失敗。**
	t.Run("piccolissimo fails on zero as well as on two", func(t *testing.T) {
		zero := bsPlaying(t, 0, -1, BostonBidPiccolissimo, 0)
		zero.SetTricksWonForTest(0, 0)
		zero.FinishHandForTest()
		if zero.IsBidMade() {
			t.Error("zero tricks fails Piccolissimo")
		}

		one := bsPlaying(t, 0, -1, BostonBidPiccolissimo, 0)
		one.SetTricksWonForTest(0, 1)
		one.FinishHandForTest()
		if !one.IsBidMade() {
			t.Error("exactly one trick makes Piccolissimo")
		}
	})

	// 高い宣言ほど配当が大きい。
	t.Run("payouts rise with the ladder", func(t *testing.T) {
		prev := 0
		for l := BostonBidFive; l < BostonBidLevelCount; l++ {
			p := BostonBidPayout(l)
			if p <= prev {
				t.Errorf("%s pays %d, not more than the previous %d", BostonBidName(l), p, prev)
			}
			prev = p
		}
	})
}

func TestBostonNextHandAndGameEnd(t *testing.T) {
	b := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
	b.SetTricksWonForTest(0, 6)
	b.FinishHandForTest()

	dealer := b.GetDealerIdx()
	if err := b.NextHand(); err != nil {
		t.Fatalf("NextHand: %v", err)
	}
	if got := b.GetDealerIdx(); got == dealer {
		t.Errorf("the dealer stayed at %d; it must rotate", got)
	}

	// 規定局数に達すると決着する。
	end := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
	end.SetHandNumberForTest(end.GetTargetHands())
	end.SetTricksWonForTest(0, 6)
	end.FinishHandForTest()
	if !end.GetGameEndFlag() {
		t.Fatal("reaching the target hand count ends the game")
	}
	if got := end.GetWinnerIdx(); got != 0 {
		t.Errorf("winner = %d, want the seat with the most chips", got)
	}
	if err := end.NextHand(); err == nil {
		t.Error("dealing after the game is over must be refused")
	}
}

func TestBostonNextHandGuards(t *testing.T) {
	b := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
	if err := b.NextHand(); err == nil {
		t.Error("dealing again mid-hand must be refused")
	}
}

func TestBostonIsHumanTurnAndCpuPlay(t *testing.T) {
	b := NewDefaultBoston()
	b.Reset()
	if b.IsHumanTurn() {
		t.Error("the dealer's left bids first and it is a CPU")
	}
	b.CpuPlay()
	if b.GetBidPlayerIdx() == 1 && b.GetPhase() == BostonPhaseBid {
		t.Error("CpuPlay must move the auction along")
	}

	b2 := NewDefaultBoston()
	b2.Reset()
	b2.SetPhaseForTest(BostonPhaseGameEnd)
	b2.gameEndFlag = true
	if b2.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	b2.CpuPlay()
}

// **CPU だけで 1 局を回し切れること。**途中で止まると詰む。
func TestBostonCpuDrivesAFullHand(t *testing.T) {
	for attempt := range 30 {
		b := NewDefaultBoston()
		b.Reset()
		for step := 0; step < 800; step++ {
			if b.GetPhase() == BostonPhaseHandEnd || b.GetGameEndFlag() {
				break
			}
			if !b.IsHumanTurn() {
				b.CpuPlay()
				continue
			}
			switch b.GetPhase() {
			case BostonPhaseBid:
				idx := b.GetBidPlayerIdx()
				level, suit := b.BostonCpuBid(idx)
				if level == BostonBidPass || b.Bid(idx, level, suit) != nil {
					_ = b.PassBid(idx)
				}
			case BostonPhaseCallPartner:
				_ = b.CallPartner(b.GetDeclarerIdx(), -1)
			case BostonPhasePlay:
				idx := b.GetCurrentPlayerIdx()
				if i := b.BostonCpuPlay(idx); i >= 0 {
					_ = b.PlayCard(idx, i)
				}
			}
		}
		if b.GetPhase() != BostonPhaseHandEnd && !b.GetGameEndFlag() {
			t.Fatalf("attempt %d: the hand never finished (phase %v)", attempt, b.GetPhase())
		}
		if got := b.GetTrickNumber(); got != BostonHandSize {
			t.Fatalf("attempt %d: %d tricks played, want %d", attempt, got, BostonHandSize)
		}
		// **13 トリックが 4 席に分配されている。**
		total := 0
		for i := range BostonPlayerCnt {
			total += b.GetTricksWon(i)
		}
		if total != BostonHandSize {
			t.Fatalf("attempt %d: %d tricks accounted for, want %d", attempt, total, BostonHandSize)
		}
	}
}

func TestBostonCpuEdges(t *testing.T) {
	b := bsPlaying(t, 0, -1, BostonBidSix, CardDesignSpade)
	b.SetHandForTest(0, []*Card{})
	if got := b.BostonCpuPlay(0); got != -1 {
		t.Errorf("an empty hand has no play, got %d", got)
	}
	if got := b.BostonCpuPlay(99); got != -1 {
		t.Errorf("an unknown seat has no play, got %d", got)
	}
	if level, _ := b.BostonCpuBid(99); level != BostonBidPass {
		t.Errorf("an unknown seat bids %v, want pass", level)
	}

	// **ミゼール側は取らないことが目的。**常に一番安い札を出す。
	mis := bsPlaying(t, 0, -1, BostonBidGrandMisere, 0)
	mis.SetHandForTest(0, []*Card{bsCard(CardDesignHeart, 1), bsCard(CardDesignHeart, 2)})
	if got := mis.BostonCpuPlay(0); got != 1 {
		t.Errorf("a misere declarer leads its lowest card, got index %d", got)
	}
}

func TestBostonAccessors(t *testing.T) {
	b := NewDefaultBoston()
	b.Reset()
	if got := b.GetHandNumber(); got != 1 {
		t.Errorf("hand number = %d, want 1", got)
	}
	if got := b.GetWinnerIdx(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if got := b.GetDeclarerIdx(); got != -1 {
		t.Errorf("declarer = %d, want -1 before the auction settles", got)
	}
	if got := b.GetPartnerIdx(); got != -1 {
		t.Errorf("partner = %d, want -1", got)
	}
	if b.GetHighBid() != nil {
		t.Error("no bid stands at the start")
	}
	if b.IsExposed() {
		t.Error("nothing is exposed before a contract is set")
	}
	if got := b.GetTargetHands(); got != BostonTargetHandsDefault {
		t.Errorf("target = %d, want %d", got, BostonTargetHandsDefault)
	}
	if got := len(b.GetPlayers()); got != BostonPlayerCnt {
		t.Errorf("%d seats, want %d", got, BostonPlayerCnt)
	}
	if b.GetPlayer(-1) != nil || b.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
	}
	if b.GetTricksWon(-1) != 0 || b.GetChips(99) != 0 {
		t.Error("out-of-range values must be 0, not a panic")
	}
	if got := len(b.GetTrick()); got != 0 {
		t.Errorf("the trick starts empty, got %d", got)
	}
	if len(b.GetActionLog()) == 0 {
		t.Error("dealing writes to the action log")
	}
	if got := len(b.GetBids()); got != 0 {
		t.Errorf("the bid history starts empty, got %d", got)
	}
	// パス扱いの既定値。
	if BostonBidName(BostonBidLevel(99)) != "pass" {
		t.Error("an out-of-range level reads as a pass")
	}
	if BostonBidTricks(BostonBidLevel(-5)) != 0 {
		t.Error("an out-of-range level asks for no tricks")
	}
}

func TestBostonConfigValidate(t *testing.T) {
	if err := DefaultBostonConfig().Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	for _, bad := range []BostonConfig{
		{CpuDifficulty: 9, TargetHands: 8},
		{CpuDifficulty: BostonCpuDifficultyNormal, TargetHands: 0},
		{CpuDifficulty: BostonCpuDifficultyNormal, TargetHands: 999},
	} {
		if err := bad.Validate(); err == nil {
			t.Errorf("%+v must not validate", bad)
		}
	}
}

func TestBostonRoundTripsThroughJSON(t *testing.T) {
	b := NewDefaultBoston()
	b.Reset()
	_ = b.Bid(1, BostonBidGrandMisere, 0)

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Boston
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetHighBid() == nil || restored.GetHighBid().Level != BostonBidGrandMisere {
		t.Error("the standing bid did not survive the round trip")
	}
	if got := restored.GetTargetHands(); got != BostonTargetHandsDefault {
		t.Errorf("target = %d, want %d", got, BostonTargetHandsDefault)
	}
	if got := restored.GetPlayer(0).GetCardsSize(); got != BostonHandSize {
		t.Errorf("the restored hand holds %d, want %d", got, BostonHandSize)
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestBostonRejectsBadJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[],"ph":0,"di":0,"ci":0,"bi":0}`},
		{"bad phase", `{"pl":[{},{},{},{}],"ph":99,"di":0,"ci":0,"bi":0}`},
		{"bad dealer", `{"pl":[{},{},{},{}],"ph":0,"di":9,"ci":0,"bi":0}`},
		{"bad declarer", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":9,"pa":-1,"tl":-1,"wi":-1}`},
		{"bad trump suit", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"pa":-1,"tl":-1,"wi":-1,"ts":9}`},
		{"oversized trick", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"pa":-1,"tl":-1,"wi":-1,"tk":[{},{},{},{},{}]}`},
		{"bad bid level", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"pa":-1,"tl":-1,"wi":-1,"hb":{"Level":99}}`},
		// **落札者が自分をパートナーにすると 1 対 3 も 2 対 2 も成り立たない。**
		{"declarer partners itself", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":1,"pa":1,"tl":-1,"wi":-1}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b Boston
			if err := json.Unmarshal([]byte(tc.body), &b); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}

func TestBostonTrumpAndConfigAccessors(t *testing.T) {
	b := NewDefaultBoston()
	b.Reset()
	// 落札前は切札なし。
	if got := b.GetTrumpSuit(); got != 0 {
		t.Errorf("trump = %d, want 0 before a contract", got)
	}
	if err := b.Bid(1, BostonBidFive, CardDesignDiamond); err != nil {
		t.Fatalf("Bid: %v", err)
	}
	for _, seat := range []int{2, 3, 0} {
		_ = b.PassBid(seat)
	}
	if got := b.GetTrumpSuit(); got != CardDesignDiamond {
		t.Errorf("trump = %d, want diamonds", got)
	}

	// **ミゼールでは切札が残らない。**前局の切札を引きずるとノートランプが壊れる。
	b2 := NewDefaultBoston()
	b2.Reset()
	if err := b2.Bid(1, BostonBidGrandMisere, CardDesignSpade); err != nil {
		t.Fatalf("Bid: %v", err)
	}
	for _, seat := range []int{2, 3, 0} {
		_ = b2.PassBid(seat)
	}
	if got := b2.GetTrumpSuit(); got != 0 {
		t.Errorf("trump = %d, want 0 for a misere", got)
	}

	cfg := b.GetConfig()
	if cfg.TargetHands != BostonTargetHandsDefault {
		t.Errorf("target = %d, want %d", cfg.TargetHands, BostonTargetHandsDefault)
	}
	cfg.TargetHands = 3
	b.SetConfig(cfg)
	if got := b.GetConfig().TargetHands; got != 3 {
		t.Errorf("SetConfig did not take effect, got %d", got)
	}
}
