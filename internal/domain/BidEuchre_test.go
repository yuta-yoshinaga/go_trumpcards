//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"testing"
)

func beCard(suit, value int) *Card { return NewCard(suit, value, true) }

// bePlaying puts a game into the play phase with a fixed contract.
func bePlaying(t *testing.T, declarer, value int, trump BidEuchreTrump) *BidEuchre {
	t.Helper()
	b := NewDefaultBidEuchre()
	b.Reset()
	b.SetPhaseForTest(BidEuchrePhasePlay)
	b.SetContractForTest(declarer, value, trump)
	b.SetCurrentPlayerForTest(0)
	b.SetTrickLeaderForTest(0)
	return b
}

// TestBidEuchreDealsEverythingWithNoKitty は issue の「残りはキティ」が誤りで
// あることを押さえる。
//
// **24 ÷ 4 = 6 でちょうど配り切る。**そもそも余りが出ない。
func TestBidEuchreDealsEverythingWithNoKitty(t *testing.T) {
	if got := len(newBidEuchreDeck()); got != BidEuchreDeckSize {
		t.Fatalf("the deck holds %d cards, want %d", got, BidEuchreDeckSize)
	}
	// **算術で確かめる。**キティが残る余地が無い。
	if BidEuchrePlayerCnt*BidEuchreHandSize != BidEuchreDeckSize {
		t.Fatalf("%d seats x %d cards != %d — a kitty would have to come from somewhere",
			BidEuchrePlayerCnt, BidEuchreHandSize, BidEuchreDeckSize)
	}

	b := NewDefaultBidEuchre()
	b.Reset()
	total := 0
	for i := range BidEuchrePlayerCnt {
		got := b.GetPlayer(i).GetCardsSize()
		if got != BidEuchreHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, BidEuchreHandSize)
		}
		total += got
	}
	if total != BidEuchreDeckSize {
		t.Errorf("%d cards are in hands, want the whole pack of %d", total, BidEuchreDeckSize)
	}

	// デッキは A-K-Q-J-10-9 だけ。
	for _, c := range newBidEuchreDeck() {
		switch c.GetValue() {
		case 1, 13, 12, 11, 10, 9:
		default:
			t.Errorf("an unexpected rank is in the deck: %d", c.GetValue())
		}
	}
}

// TestBidEuchreMinimumBidIsThree covers the floor the issue omits.
func TestBidEuchreMinimumBidIsThree(t *testing.T) {
	if BidEuchreMinBid != 3 {
		t.Fatalf("BidEuchreMinBid = %d, want 3", BidEuchreMinBid)
	}
	b := NewDefaultBidEuchre()
	b.Reset()
	if err := b.Bid(b.GetBidPlayerIdx(), 2); err == nil {
		t.Error("a bid of 2 must be refused")
	}
	if err := b.Bid(b.GetBidPlayerIdx(), BidEuchreMaxBid+1); err == nil {
		t.Error("a bid above the hand size must be refused")
	}
	if err := b.Bid(b.GetBidPlayerIdx(), BidEuchreMinBid); err != nil {
		t.Errorf("a bid of %d must be accepted: %v", BidEuchreMinBid, err)
	}
}

// TestBidEuchreDealerMayEqualTheHighestBid covers the privilege the issue omits.
//
// **ディーラーだけは同額で奪える。**他の席は上回らないと通らない。
func TestBidEuchreDealerMayEqualTheHighestBid(t *testing.T) {
	b := NewDefaultBidEuchre()
	b.Reset()
	b.SetDealerForTest(0)
	b.SetBidPlayerForTest(1)
	if err := b.Bid(1, 4); err != nil {
		t.Fatalf("Bid: %v", err)
	}

	// 非ディーラーは同額で被せられない。
	if b.BidEuchreCanBid(2, 4) {
		t.Error("a non-dealer must beat the standing bid, not equal it")
	}
	if !b.BidEuchreCanBid(2, 5) {
		t.Error("a non-dealer may bid higher")
	}
	// **ディーラーは同額で奪える。**
	if !b.BidEuchreCanBid(0, 4) {
		t.Error("the dealer may EQUAL the standing bid")
	}
	if b.BidEuchreCanBid(0, 3) {
		t.Error("even the dealer may not bid lower")
	}
}

// **ディーラーが同額で落札すると、ディーラーが宣言側になる。**
func TestBidEuchreDealerEqualBidTakesTheContract(t *testing.T) {
	b := NewDefaultBidEuchre()
	b.Reset()
	b.SetDealerForTest(3)
	b.SetBidPlayerForTest(0)
	if err := b.Bid(0, 4); err != nil {
		t.Fatalf("Bid: %v", err)
	}
	_ = b.PassBid(1)
	_ = b.PassBid(2)
	// ディーラー (席 3) が同額で奪う。
	if err := b.Bid(3, 4); err != nil {
		t.Fatalf("the dealer may equal the bid: %v", err)
	}
	if got := b.GetDeclarerIdx(); got != 3 {
		t.Errorf("declarer = %d, want the dealer (3)", got)
	}
}

// TestBidEuchreBowerRanking covers the ordering the issue does describe.
func TestBidEuchreBowerRanking(t *testing.T) {
	const trump = BidEuchreTrumpSpade
	suit := BidEuchreTrumpSuit(trump)

	right := beCard(CardDesignSpade, 11)
	left := beCard(CardDesignClover, 11) // 同色
	ace := beCard(CardDesignSpade, 1)
	king := beCard(CardDesignSpade, 13)

	if !IsBidEuchreRightBower(right, suit) {
		t.Error("the trump jack is the right bower")
	}
	if !IsBidEuchreLeftBower(left, suit) {
		t.Error("the same-colour jack is the left bower")
	}
	// 別色の J はボワーではない。
	if IsBidEuchreLeftBower(beCard(CardDesignHeart, 11), suit) {
		t.Error("an off-colour jack is not a bower")
	}

	descending := []*Card{right, left, ace, king,
		beCard(CardDesignSpade, 12), beCard(CardDesignSpade, 10), beCard(CardDesignSpade, 9)}
	for i := 1; i < len(descending); i++ {
		hi := BidEuchreCardRank(descending[i-1], trump)
		lo := BidEuchreCardRank(descending[i], trump)
		if hi <= lo {
			t.Errorf("card %d must outrank card %d in the trump suit", i-1, i)
		}
	}
	// 切札は平の札より強い。
	if BidEuchreCardRank(beCard(CardDesignSpade, 9), trump) <= BidEuchreCardRank(beCard(CardDesignHeart, 1), trump) {
		t.Error("the lowest trump beats the highest plain card")
	}
}

// **左ボワーは全ての目的で切札スートに属する。**追随の判定でも切札として扱う。
func TestBidEuchreLeftBowerCountsAsTrump(t *testing.T) {
	const trump = BidEuchreTrumpSpade
	suit := BidEuchreTrumpSuit(trump)
	left := beCard(CardDesignClover, 11)

	if got := BidEuchreEffectiveSuit(left, suit); got != suit {
		t.Errorf("the left bower's effective suit is %d, want the trump suit %d", got, suit)
	}
	// 素のスート (♣) には属さない。
	if BidEuchreEffectiveSuit(left, suit) == CardDesignClover {
		t.Error("the left bower must not follow its printed suit")
	}
	if got := BidEuchreEffectiveSuit(nil, suit); got != 0 {
		t.Errorf("a nil card has effective suit %d, want 0", got)
	}

	// 切札がリードされたら左ボワーは追随を強制される。
	b := bePlaying(t, 0, 3, trump)
	b.SetHandForTest(0, []*Card{beCard(CardDesignSpade, 1)})
	b.SetHandForTest(1, []*Card{left, beCard(CardDesignHeart, 9)})
	if err := b.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	valid := b.BidEuchreValidPlays(1)
	if len(valid) != 1 || valid[0] != 0 {
		t.Errorf("valid = %v, want only the left bower — it belongs to the trump suit", valid)
	}

	// **♣ がリードされたら左ボワーは追随に使えない。**
	b2 := bePlaying(t, 0, 3, trump)
	b2.SetHandForTest(0, []*Card{beCard(CardDesignClover, 1)})
	b2.SetHandForTest(1, []*Card{left, beCard(CardDesignClover, 9)})
	if err := b2.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	valid2 := b2.BidEuchreValidPlays(1)
	if len(valid2) != 1 || valid2[0] != 1 {
		t.Errorf("valid = %v, want only the plain club — the left bower is a trump now", valid2)
	}
}

// TestBidEuchreNoTrumpLowReversesTheRanking covers the declaration the issue omits.
func TestBidEuchreNoTrumpLowReversesTheRanking(t *testing.T) {
	nine := beCard(CardDesignHeart, 9)
	ace := beCard(CardDesignHeart, 1)

	// ノートランプ・ハイでは A が最強。
	if BidEuchreCardRank(ace, BidEuchreTrumpNoHigh) <= BidEuchreCardRank(nine, BidEuchreTrumpNoHigh) {
		t.Error("at no trump high the ace beats the nine")
	}
	// **ノートランプ・ローでは 9 が最強。**
	if BidEuchreCardRank(nine, BidEuchreTrumpNoLow) <= BidEuchreCardRank(ace, BidEuchreTrumpNoLow) {
		t.Error("at no trump LOW the nine beats the ace")
	}
	// 全体として逆になる。
	descendingLow := []int{9, 10, 11, 12, 13, 1}
	for i := 1; i < len(descendingLow); i++ {
		hi := BidEuchreCardRank(beCard(CardDesignHeart, descendingLow[i-1]), BidEuchreTrumpNoLow)
		lo := BidEuchreCardRank(beCard(CardDesignHeart, descendingLow[i]), BidEuchreTrumpNoLow)
		if hi <= lo {
			t.Errorf("at no trump low %d must beat %d", descendingLow[i-1], descendingLow[i])
		}
	}
	// **ノートランプではボワーが無い。**J はただの J。
	if BidEuchreTrumpSuit(BidEuchreTrumpNoHigh) != 0 || BidEuchreTrumpSuit(BidEuchreTrumpNoLow) != 0 {
		t.Error("a no-trump declaration has no trump suit")
	}
	if !BidEuchreIsNoTrump(BidEuchreTrumpNoHigh) || !BidEuchreIsNoTrump(BidEuchreTrumpNoLow) {
		t.Error("both no-trump declarations must report as no trump")
	}
	if BidEuchreIsNoTrump(BidEuchreTrumpSpade) {
		t.Error("a suit declaration is not no trump")
	}
	if BidEuchreCardRank(nil, BidEuchreTrumpNoHigh) != 0 {
		t.Error("a nil card has no rank")
	}
}

// TestBidEuchreScoring covers the settlement.
func TestBidEuchreScoring(t *testing.T) {
	// **達成なら両チームが取ったトリック数ぶん得点する。**
	t.Run("making the bid scores both sides their tricks", func(t *testing.T) {
		b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
		b.SetTricksWonForTest(0, 2)
		b.SetTricksWonForTest(2, 2) // 宣言側 4
		b.SetTricksWonForTest(1, 1)
		b.SetTricksWonForTest(3, 1) // 守備側 2
		b.FinishHandForTest()

		res := b.GetLastResult()
		if !res.Made {
			t.Fatal("4 tricks makes a bid of 3")
		}
		if got := res.Points[0]; got != 4 {
			t.Errorf("the declaring side scores %d, want its 4 tricks", got)
		}
		// **守備側も自分のトリックを得点する。**
		if got := res.Points[1]; got != 2 {
			t.Errorf("the defenders score %d, want their 2 tricks", got)
		}
	})

	// **未達なら引かれるのは宣言額。**取ったトリック数ではない。
	t.Run("being set costs the bid, not the tricks", func(t *testing.T) {
		b := bePlaying(t, 0, 5, BidEuchreTrumpSpade)
		b.SetTricksWonForTest(0, 1)
		b.SetTricksWonForTest(2, 1) // 宣言側 2
		b.SetTricksWonForTest(1, 2)
		b.SetTricksWonForTest(3, 2) // 守備側 4
		b.FinishHandForTest()

		res := b.GetLastResult()
		if res.Made {
			t.Fatal("2 tricks fails a bid of 5")
		}
		// **-5 であって -2 ではない。**
		if got := res.Points[0]; got != -5 {
			t.Errorf("the set side scores %d, want -5 (the BID, not the tricks taken)", got)
		}
		// **未達でも守備側は自分のトリックを得点する。**
		if got := res.Points[1]; got != 4 {
			t.Errorf("the defenders score %d, want their 4 tricks even on a set", got)
		}
	})

	// 宣言額ちょうどは達成。
	t.Run("exactly the bid makes it", func(t *testing.T) {
		b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
		b.SetTricksWonForTest(0, 3)
		b.SetTricksWonForTest(1, 3)
		b.FinishHandForTest()
		if !b.GetLastResult().Made {
			t.Error("taking exactly the bid makes it")
		}
	})
}

// TestBidEuchreGameTargetIsThirtyTwo covers the target the issue omits.
func TestBidEuchreGameTargetIsThirtyTwo(t *testing.T) {
	if BidEuchreGameTarget != 32 {
		t.Fatalf("BidEuchreGameTarget = %d, want 32", BidEuchreGameTarget)
	}
	b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
	b.SetScoreForTest(0, BidEuchreGameTarget-3)
	b.SetTricksWonForTest(0, 4)
	b.SetTricksWonForTest(1, 2)
	b.FinishHandForTest()

	if !b.GetGameEndFlag() {
		t.Fatal("reaching 32 ends the game")
	}
	if got := b.GetWinnerTeam(); got != 0 {
		t.Errorf("winner = %d, want 0", got)
	}
	if err := b.NextHand(); err == nil {
		t.Error("dealing after the game is over must be refused")
	}
}

// 両チームが同時に超えたら落札側の勝ち。
func TestBidEuchreGameEndTieBreak(t *testing.T) {
	b := bePlaying(t, 1, 3, BidEuchreTrumpSpade)
	b.SetScoreForTest(0, BidEuchreGameTarget)
	b.SetScoreForTest(1, BidEuchreGameTarget)
	b.checkGameEnd()
	if got := b.GetWinnerTeam(); got != BidEuchreTeamOf(1) {
		t.Errorf("winner = %d, want the declaring team", got)
	}
}

func TestBidEuchreBidding(t *testing.T) {
	t.Run("the dealer's left bids first", func(t *testing.T) {
		b := NewDefaultBidEuchre()
		b.Reset()
		if got := b.GetBidPlayerIdx(); got != 1 {
			t.Errorf("bid seat = %d, want 1", got)
		}
	})

	t.Run("winning the auction asks for a declaration", func(t *testing.T) {
		b := NewDefaultBidEuchre()
		b.Reset()
		if err := b.Bid(1, 4); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		for _, seat := range []int{2, 3, 0} {
			_ = b.PassBid(seat)
		}
		if got := b.GetPhase(); got != BidEuchrePhaseChooseTrump {
			t.Errorf("phase = %v, want the trump question", got)
		}
		if got := b.GetDeclarerIdx(); got != 1 {
			t.Errorf("declarer = %d, want 1", got)
		}
		if b.IsTrumpChosen() {
			t.Error("nothing is chosen until the declarer names it")
		}
	})

	t.Run("four passes redeal without advancing the hand number", func(t *testing.T) {
		b := NewDefaultBidEuchre()
		b.Reset()
		hand := b.GetHandNumber()
		for _, seat := range []int{1, 2, 3, 0} {
			_ = b.PassBid(seat)
		}
		if got := b.GetPhase(); got != BidEuchrePhaseBid {
			t.Errorf("phase = %v, want a fresh auction", got)
		}
		if got := b.GetHandNumber(); got != hand {
			t.Errorf("hand number = %d, want %d", got, hand)
		}
		// 配り直しても全札が配られる。
		for i := range BidEuchrePlayerCnt {
			if got := b.GetPlayer(i).GetCardsSize(); got != BidEuchreHandSize {
				t.Errorf("seat %d holds %d after the redeal, want %d", i, got, BidEuchreHandSize)
			}
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		b := NewDefaultBidEuchre()
		b.Reset()
		if err := b.Bid(0, 4); err == nil {
			t.Error("bidding out of turn must be refused")
		}
		b.SetPhaseForTest(BidEuchrePhasePlay)
		if err := b.PassBid(1); err == nil {
			t.Error("bidding outside the auction must be refused")
		}
	})
}

func TestBidEuchreChooseTrump(t *testing.T) {
	setup := func(t *testing.T) *BidEuchre {
		t.Helper()
		b := NewDefaultBidEuchre()
		b.Reset()
		if err := b.Bid(1, 4); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		for _, seat := range []int{2, 3, 0} {
			_ = b.PassBid(seat)
		}
		return b
	}

	t.Run("naming a suit starts the play", func(t *testing.T) {
		b := setup(t)
		if err := b.ChooseTrump(1, BidEuchreTrumpHeart); err != nil {
			t.Fatalf("ChooseTrump: %v", err)
		}
		if got := b.GetTrumpSuit(); got != CardDesignHeart {
			t.Errorf("trump = %d, want hearts", got)
		}
		if got := b.GetPhase(); got != BidEuchrePhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
		// **リードは落札者。**
		if got := b.GetTrickLeaderIdx(); got != 1 {
			t.Errorf("leader = %d, want the declarer", got)
		}
	})

	t.Run("a no-trump declaration leaves no trump suit", func(t *testing.T) {
		b := setup(t)
		if err := b.ChooseTrump(1, BidEuchreTrumpNoLow); err != nil {
			t.Fatalf("ChooseTrump: %v", err)
		}
		if got := b.GetTrumpSuit(); got != 0 {
			t.Errorf("trump = %d, want 0 at no trump", got)
		}
		if got := b.GetTrump(); got != BidEuchreTrumpNoLow {
			t.Errorf("declaration = %v, want no trump low", got)
		}
	})

	// **設定でノートランプを切れる。**読まなければ config が黙って効かない。
	t.Run("no trump can be switched off", func(t *testing.T) {
		b := setup(t)
		cfg := b.GetConfig()
		cfg.AllowNoTrump = false
		b.SetConfig(cfg)

		for _, t2 := range []BidEuchreTrump{BidEuchreTrumpNoHigh, BidEuchreTrumpNoLow} {
			if err := b.ChooseTrump(1, t2); err == nil {
				t.Errorf("%v must be refused when allowNoTrump is off", t2)
			}
		}
		// スート宣言はそのまま通る。
		if err := b.ChooseTrump(1, BidEuchreTrumpSpade); err != nil {
			t.Errorf("a suit declaration must still be accepted: %v", err)
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		b := setup(t)
		if err := b.ChooseTrump(2, BidEuchreTrumpHeart); err == nil {
			t.Error("only the declarer names trump")
		}
		if err := b.ChooseTrump(1, BidEuchreTrump(99)); err == nil {
			t.Error("a bad declaration must be refused")
		}
		if err := b.ChooseTrump(1, BidEuchreTrumpHeart); err != nil {
			t.Fatalf("ChooseTrump: %v", err)
		}
		// プレイに入ったら変えられない。
		if err := b.ChooseTrump(1, BidEuchreTrumpSpade); err == nil {
			t.Error("trump cannot be renamed once play starts")
		}
	})
}

func TestBidEuchrePlayGuards(t *testing.T) {
	b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
	b.SetHandForTest(0, []*Card{beCard(CardDesignHeart, 1)})
	if err := b.PlayCard(1, 0); err == nil {
		t.Error("playing out of turn must be refused")
	}
	if err := b.PlayCard(0, 99); err == nil {
		t.Error("an out-of-range index must be refused")
	}
	b.SetPhaseForTest(BidEuchrePhaseBid)
	if err := b.PlayCard(0, 0); err == nil {
		t.Error("playing outside the play phase must be refused")
	}
	if b.BidEuchreValidPlays(99) != nil {
		t.Error("an unknown seat has no legal plays")
	}
}

// 切札が勝ち、ノートランプでは効かない。
func TestBidEuchreTrickResolution(t *testing.T) {
	withTrump := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
	withTrump.SetHandForTest(0, []*Card{beCard(CardDesignHeart, 1)})
	withTrump.SetHandForTest(1, []*Card{beCard(CardDesignSpade, 9)})
	withTrump.SetHandForTest(2, []*Card{beCard(CardDesignHeart, 13)})
	withTrump.SetHandForTest(3, []*Card{beCard(CardDesignHeart, 12)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := withTrump.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	if got := withTrump.GetTricksWon(1); got != 1 {
		t.Errorf("the low trump takes the trick, seat 1 has %d", got)
	}

	// **ノートランプ・ローでは 9 が勝つ。**
	noLow := bePlaying(t, 0, 3, BidEuchreTrumpNoLow)
	noLow.SetHandForTest(0, []*Card{beCard(CardDesignHeart, 1)})
	noLow.SetHandForTest(1, []*Card{beCard(CardDesignHeart, 9)})
	noLow.SetHandForTest(2, []*Card{beCard(CardDesignHeart, 13)})
	noLow.SetHandForTest(3, []*Card{beCard(CardDesignHeart, 12)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := noLow.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	if got := noLow.GetTricksWon(1); got != 1 {
		t.Errorf("at no trump low the nine wins, seat 1 has %d", got)
	}
}

func TestBidEuchreTeamTricks(t *testing.T) {
	b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
	b.SetTricksWonForTest(0, 2)
	b.SetTricksWonForTest(2, 1)
	b.SetTricksWonForTest(1, 2)
	b.SetTricksWonForTest(3, 1)
	if got := b.BidEuchreTeamTricks(0); got != 3 {
		t.Errorf("team 0 took %d, want 3", got)
	}
	if got := b.BidEuchreTeamTricks(1); got != 3 {
		t.Errorf("team 1 took %d, want 3", got)
	}
	if b.BidEuchreTeamTricks(99) != 0 {
		t.Error("an out-of-range team took nothing")
	}
	// パートナーは向かい合わせ。
	if BidEuchreTeamOf(0) != BidEuchreTeamOf(2) || BidEuchreTeamOf(1) != BidEuchreTeamOf(3) {
		t.Error("seats 0/2 and 1/3 are partners")
	}
	// **範囲外は -1。**Go の剰余は -1 % 2 = -1 なので素通しは危険。
	if got := BidEuchreTeamOf(-1); got != -1 {
		t.Errorf("BidEuchreTeamOf(-1) = %d, want -1", got)
	}
	if got := BidEuchreTeamOf(99); got != -1 {
		t.Errorf("BidEuchreTeamOf(99) = %d, want -1", got)
	}
}

// 落札前に公開アクセサを呼んでも落ちない。
func TestBidEuchreAccessorsBeforeTheAuctionSettles(t *testing.T) {
	b := NewDefaultBidEuchre()
	b.Reset()
	if got := b.GetDeclarerIdx(); got != -1 {
		t.Fatalf("declarer = %d, want -1 during the auction", got)
	}
	for _, get := range []func() int{
		func() int { return b.BidEuchreTeamTricks(BidEuchreTeamOf(b.GetDeclarerIdx())) },
		func() int { return b.GetScore(BidEuchreTeamOf(b.GetDeclarerIdx())) },
		func() int { return b.GetTricksWon(b.GetDeclarerIdx()) },
	} {
		if got := get(); got != 0 {
			t.Errorf("an accessor returned %d before the auction settles, want 0", got)
		}
	}
}

func TestBidEuchreNextHandRotatesTheDealer(t *testing.T) {
	b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
	b.SetTricksWonForTest(0, 4)
	b.FinishHandForTest()

	dealer := b.GetDealerIdx()
	if err := b.NextHand(); err != nil {
		t.Fatalf("NextHand: %v", err)
	}
	if got := b.GetDealerIdx(); got == dealer {
		t.Errorf("the dealer stayed at %d; it must rotate", got)
	}
	for i := range BidEuchrePlayerCnt {
		if got := b.GetPlayer(i).GetCardsSize(); got != BidEuchreHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, BidEuchreHandSize)
		}
	}
}

func TestBidEuchreNextHandGuards(t *testing.T) {
	b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
	if err := b.NextHand(); err == nil {
		t.Error("dealing again mid-hand must be refused")
	}
}

func TestBidEuchreIsHumanTurnAndCpuPlay(t *testing.T) {
	b := NewDefaultBidEuchre()
	b.Reset()
	if b.IsHumanTurn() {
		t.Error("the dealer's left bids first and it is a CPU")
	}
	b.CpuPlay()
	if b.GetBidPlayerIdx() == 1 && b.GetPhase() == BidEuchrePhaseBid {
		t.Error("CpuPlay must move the auction along")
	}

	b2 := NewDefaultBidEuchre()
	b2.Reset()
	b2.SetPhaseForTest(BidEuchrePhaseGameEnd)
	b2.gameEndFlag = true
	if b2.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	b2.CpuPlay()
}

// **CPU だけで 1 局を回し切れること。**途中で止まると詰む。
func TestBidEuchreCpuDrivesAFullHand(t *testing.T) {
	for attempt := range 30 {
		b := NewDefaultBidEuchre()
		b.Reset()
		for step := 0; step < 400; step++ {
			if b.GetPhase() == BidEuchrePhaseHandEnd || b.GetGameEndFlag() {
				break
			}
			if !b.IsHumanTurn() {
				b.CpuPlay()
				continue
			}
			switch b.GetPhase() {
			case BidEuchrePhaseBid:
				idx := b.GetBidPlayerIdx()
				value := b.BidEuchreCpuBid(idx)
				if value < BidEuchreMinBid || b.Bid(idx, value) != nil {
					_ = b.PassBid(idx)
				}
			case BidEuchrePhaseChooseTrump:
				idx := b.GetDeclarerIdx()
				_ = b.ChooseTrump(idx, b.BidEuchreCpuTrump(idx))
			case BidEuchrePhasePlay:
				idx := b.GetCurrentPlayerIdx()
				if i := b.BidEuchreCpuPlay(idx); i >= 0 {
					_ = b.PlayCard(idx, i)
				}
			}
		}
		if b.GetPhase() != BidEuchrePhaseHandEnd && !b.GetGameEndFlag() {
			t.Fatalf("attempt %d: the hand never finished (phase %v)", attempt, b.GetPhase())
		}
		if got := b.GetTrickNumber(); got != BidEuchreHandSize {
			t.Fatalf("attempt %d: %d tricks played, want %d", attempt, got, BidEuchreHandSize)
		}
		// **6 トリックが 4 席に分配されている。**
		total := 0
		for i := range BidEuchrePlayerCnt {
			total += b.GetTricksWon(i)
		}
		if total != BidEuchreHandSize {
			t.Fatalf("attempt %d: %d tricks accounted for, want %d", attempt, total, BidEuchreHandSize)
		}
	}
}

func TestBidEuchreCpuEdges(t *testing.T) {
	b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
	b.SetHandForTest(0, []*Card{})
	if got := b.BidEuchreCpuPlay(0); got != -1 {
		t.Errorf("an empty hand has no play, got %d", got)
	}
	if got := b.BidEuchreCpuPlay(99); got != -1 {
		t.Errorf("an unknown seat has no play, got %d", got)
	}
	if got := b.BidEuchreCpuBid(99); got != 0 {
		t.Errorf("an unknown seat bids %d, want 0", got)
	}
	if got := b.BidEuchreCpuTrump(99); got != BidEuchreTrumpSpade {
		t.Errorf("an unknown seat falls back to %v", got)
	}
}

func TestBidEuchreAccessors(t *testing.T) {
	b := NewDefaultBidEuchre()
	b.Reset()
	if got := b.GetHandNumber(); got != 1 {
		t.Errorf("hand number = %d, want 1", got)
	}
	if got := b.GetWinnerTeam(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if b.GetHighBid() != nil {
		t.Error("no bid stands at the start")
	}
	if b.GetLastResult() != nil {
		t.Error("there is no result before the first settlement")
	}
	if got := len(b.GetPlayers()); got != BidEuchrePlayerCnt {
		t.Errorf("%d seats, want %d", got, BidEuchrePlayerCnt)
	}
	if b.GetPlayer(-1) != nil || b.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
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
	cfg := b.GetConfig()
	if !cfg.AllowNoTrump {
		t.Error("no-trump declarations are allowed by default")
	}
	cfg.AllowNoTrump = false
	b.SetConfig(cfg)
	if b.GetConfig().AllowNoTrump {
		t.Error("SetConfig must take effect")
	}
	// プレイヤーのチーム判定。
	p := b.GetPlayer(0)
	if p.GetTeam(0) != p.GetTeam(2) || p.GetTeam(0) == p.GetTeam(1) {
		t.Error("seats 0/2 are partners and 0/1 are opponents")
	}
}

func TestBidEuchreConfigValidate(t *testing.T) {
	if err := DefaultBidEuchreConfig().Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	if err := (BidEuchreConfig{CpuDifficulty: 9}).Validate(); err == nil {
		t.Error("a bad difficulty must not validate")
	}
}

func TestBidEuchreRoundTripsThroughJSON(t *testing.T) {
	b := NewDefaultBidEuchre()
	b.Reset()
	_ = b.Bid(1, 4)

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored BidEuchre
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetHighBid() == nil || restored.GetHighBid().Value != 4 {
		t.Error("the standing bid did not survive the round trip")
	}
	if got := restored.GetPlayer(0).GetCardsSize(); got != BidEuchreHandSize {
		t.Errorf("the restored hand holds %d, want %d", got, BidEuchreHandSize)
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestBidEuchreRejectsBadJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[],"ph":0,"di":0,"ci":0,"bi":0}`},
		{"bad phase", `{"pl":[{},{},{},{}],"ph":99,"di":0,"ci":0,"bi":0}`},
		{"bad dealer", `{"pl":[{},{},{},{}],"ph":0,"di":9,"ci":0,"bi":0}`},
		{"bad declarer", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":9,"tl":-1,"wt":-1}`},
		{"bad winner team", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":9}`},
		{"bad trump declaration", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":-1,"tp":99}`},
		{"bad trump suit", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":-1,"ts":9}`},
		{"oversized trick", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":-1,"tk":[{},{},{},{},{}]}`},
		{"impossible high bid", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":-1,"hb":{"Value":99}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b BidEuchre
			if err := json.Unmarshal([]byte(tc.body), &b); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}

// 24 枚デッキに無い札は序列を持たない。
func TestBidEuchreRankOfACardOutsideTheDeck(t *testing.T) {
	for _, tr := range []BidEuchreTrump{BidEuchreTrumpSpade, BidEuchreTrumpNoHigh, BidEuchreTrumpNoLow} {
		if got := BidEuchreCardRank(beCard(CardDesignHeart, 2), tr); got != 0 {
			t.Errorf("a two ranks %d under %v, want 0 — it is not in the 24-card pack", got, tr)
		}
	}
}

// 出せない札は弾かれる。
func TestBidEuchreRejectsAnIllegalPlay(t *testing.T) {
	b := bePlaying(t, 0, 3, BidEuchreTrumpSpade)
	b.SetHandForTest(0, []*Card{beCard(CardDesignSpade, 1)})
	b.SetHandForTest(1, []*Card{beCard(CardDesignSpade, 9), beCard(CardDesignHeart, 9)})
	if err := b.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	// **♠ を持っているので ♥ は出せない。**
	if err := b.PlayCard(1, 1); err == nil {
		t.Error("a card outside the valid plays must be refused")
	}
	if err := b.PlayCard(1, 0); err != nil {
		t.Errorf("following suit must be accepted: %v", err)
	}
}

func TestBidEuchreGetScoreOutOfRange(t *testing.T) {
	b := NewDefaultBidEuchre()
	b.Reset()
	for _, team := range []int{-1, BidEuchreTeamCnt} {
		if got := b.GetScore(team); got != 0 {
			t.Errorf("GetScore(%d) = %d, want 0", team, got)
		}
	}
	b.SetScoreForTest(1, 17)
	if got := b.GetScore(1); got != 17 {
		t.Errorf("GetScore(1) = %d, want 17", got)
	}
}
