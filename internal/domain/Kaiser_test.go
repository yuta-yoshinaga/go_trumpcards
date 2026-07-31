//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"
)

func kzCard(suit, value int) *Card { return NewCard(suit, value, true) }

// kzPlaying puts a game into the play phase with a fixed contract.
func kzPlaying(t *testing.T, trump int, contract KaiserContract) *Kaiser {
	t.Helper()
	k := NewDefaultKaiser()
	k.Reset()
	k.SetPhaseForTest(KaiserPhasePlay)
	k.SetContractForTest(0, KaiserMinBid, trump, contract)
	k.SetCurrentPlayerForTest(0)
	k.SetTrickLeaderForTest(0)
	return k
}

// TestKaiserDeckHasThirtyFourCardsIncludingBothSpecials は issue の自己矛盾を
// 押さえる。
//
// **「ランク 2〜6 を除外」すると ♥5 も ♠3 も消える。**このゲームの名前の由来で
// あり得点の要である 2 枚が、指定されたデッキに存在しないことになる。
func TestKaiserDeckHasThirtyFourCardsIncludingBothSpecials(t *testing.T) {
	deck := newKaiserDeck()
	if len(deck) != KaiserDeckSize {
		t.Fatalf("the deck holds %d cards, want %d", len(deck), KaiserDeckSize)
	}
	if KaiserDeckSize != 34 {
		t.Fatalf("KaiserDeckSize = %d, want 34 (the issue says 32)", KaiserDeckSize)
	}

	// **算術で確かめる。**4 人 × 8 枚 = 32、これにキティ 2 枚で 34。
	if KaiserPlayerCnt*KaiserHandSize+KaiserKittySize != KaiserDeckSize {
		t.Errorf("%d seats x %d cards + %d kitty != %d",
			KaiserPlayerCnt, KaiserHandSize, KaiserKittySize, KaiserDeckSize)
	}

	heartFive, spadeThree := 0, 0
	for _, c := range deck {
		if IsKaiserHeartFive(c) {
			heartFive++
		}
		if IsKaiserSpadeThree(c) {
			spadeThree++
		}
		// 2〜6 のうちデッキに居てよいのは ♥5 と ♠3 だけ。
		v := c.GetValue()
		if v >= 2 && v <= 6 && !IsKaiserHeartFive(c) && !IsKaiserSpadeThree(c) {
			t.Errorf("an unexpected low card is in the deck: suit %d value %d", c.GetDesign(), v)
		}
	}
	if heartFive != 1 {
		t.Errorf("the five of hearts appears %d times, want 1 — the issue's deck excludes it", heartFive)
	}
	if spadeThree != 1 {
		t.Errorf("the three of spades appears %d times, want 1 — the issue's deck excludes it", spadeThree)
	}
}

// TestKaiserHandTotalIsTen confirms the one figure the issue gets right.
func TestKaiserHandTotalIsTen(t *testing.T) {
	if KaiserHandTotal != 10 {
		t.Fatalf("KaiserHandTotal = %d, want 10", KaiserHandTotal)
	}
	if KaiserHandSize+KaiserHeartFiveBonus+KaiserSpadeThreePenalty != 10 {
		t.Errorf("8 tricks + 5 - 3 must be 10")
	}
}

// TestKaiserDealLeavesAKitty covers the structure the issue omits.
func TestKaiserDealLeavesAKitty(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	for i := range KaiserPlayerCnt {
		if got := k.GetPlayer(i).GetCardsSize(); got != KaiserHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, KaiserHandSize)
		}
	}
	// **配り切りではない。**32 枚デッキではここが 0 になってしまう。
	if got := k.GetKittySize(); got != KaiserKittySize {
		t.Fatalf("kitty holds %d, want %d", got, KaiserKittySize)
	}
	if got := k.GetBidPlayerIdx(); got != 1 {
		t.Errorf("bidding starts at %d, want the dealer's left", got)
	}
	if got := k.GetPhase(); got != KaiserPhaseBid {
		t.Errorf("phase = %v, want bidding", got)
	}
}

// TestKaiserMinimumBidIsSeven covers the value the issue does not state.
func TestKaiserMinimumBidIsSeven(t *testing.T) {
	if KaiserMinBid != 7 {
		t.Fatalf("KaiserMinBid = %d, want 7 — the kitty advantage raises the floor from 6", KaiserMinBid)
	}
	k := NewDefaultKaiser()
	k.Reset()
	if err := k.Bid(1, 6, KaiserContractTrump); err == nil {
		t.Error("a bid of 6 must be refused")
	}
	if err := k.Bid(1, KaiserMaxBid+1, KaiserContractTrump); err == nil {
		t.Error("a bid above the maximum must be refused")
	}
	if err := k.Bid(1, KaiserMinBid, KaiserContractTrump); err != nil {
		t.Errorf("a bid of %d must be accepted: %v", KaiserMinBid, err)
	}
}

// TestKaiserBidOrdering は同じ数字でもノートランプが上位に来ることを確かめる。
func TestKaiserBidOrdering(t *testing.T) {
	if KaiserBidRank(7, KaiserContractNoTrump) <= KaiserBidRank(7, KaiserContractTrump) {
		t.Error("no trump must outrank a trump bid of the same number")
	}
	if KaiserBidRank(7, KaiserContractLowNoTrump) <= KaiserBidRank(7, KaiserContractNoTrump) {
		t.Error("low no trump ranks just above the corresponding no trump")
	}
	if KaiserBidRank(8, KaiserContractTrump) <= KaiserBidRank(7, KaiserContractLowNoTrump) {
		t.Error("a higher number always outranks a lower one whatever the contract")
	}

	k := NewDefaultKaiser()
	k.Reset()
	if err := k.Bid(1, 8, KaiserContractTrump); err != nil {
		t.Fatalf("Bid: %v", err)
	}
	// 同額の切札ビッドは通らない。
	if err := k.Bid(2, 8, KaiserContractTrump); err == nil {
		t.Error("an equal trump bid must be refused")
	}
	// 同額でもノートランプなら通る。
	if err := k.Bid(2, 8, KaiserContractNoTrump); err != nil {
		t.Errorf("an equal no-trump bid must be accepted: %v", err)
	}
}

// TestKaiserLowNoTrumpReversesTheRanking covers the contract the issue omits.
func TestKaiserLowNoTrumpReversesTheRanking(t *testing.T) {
	seven := kzCard(CardDesignHeart, 7)
	ace := kzCard(CardDesignHeart, 1)

	if KaiserCardRank(ace, KaiserContractTrump) <= KaiserCardRank(seven, KaiserContractTrump) {
		t.Error("normally the ace beats the seven")
	}
	// **ロー・ノートランプでは 7 が最強。**
	if KaiserCardRank(seven, KaiserContractLowNoTrump) <= KaiserCardRank(ace, KaiserContractLowNoTrump) {
		t.Error("in low no trump the seven beats the ace")
	}
	// 順序が全体として逆になる。
	descendingLow := []int{7, 8, 9, 10, 11, 12, 13, 1}
	for i := 1; i < len(descendingLow); i++ {
		hi := KaiserCardRank(kzCard(CardDesignHeart, descendingLow[i-1]), KaiserContractLowNoTrump)
		lo := KaiserCardRank(kzCard(CardDesignHeart, descendingLow[i]), KaiserContractLowNoTrump)
		if hi <= lo {
			t.Errorf("in low no trump %d must beat %d", descendingLow[i-1], descendingLow[i])
		}
	}
	if KaiserCardRank(nil, KaiserContractTrump) != 0 {
		t.Error("a nil card has no rank")
	}
}

// TestKaiserSpecialCardsSitJustBelowTheAce pins where the 5 and 3 rank.
func TestKaiserSpecialCardsSitJustBelowTheAce(t *testing.T) {
	five := kzCard(CardDesignHeart, 5)
	ace := kzCard(CardDesignHeart, 1)
	king := kzCard(CardDesignHeart, 13)

	if KaiserCardRank(ace, KaiserContractTrump) <= KaiserCardRank(five, KaiserContractTrump) {
		t.Error("the ace beats the five of hearts")
	}
	if KaiserCardRank(five, KaiserContractTrump) <= KaiserCardRank(king, KaiserContractTrump) {
		t.Error("the five of hearts beats the king")
	}
	// **特殊札は低ノートランプでも逆転しない。**
	if KaiserCardRank(five, KaiserContractLowNoTrump) != KaiserCardRank(five, KaiserContractTrump) {
		t.Error("the five of hearts keeps its place in low no trump")
	}
}

// TestKaiserDeclarerTakesTheKitty covers the flow the issue omits.
func TestKaiserDeclarerTakesTheKitty(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	if err := k.Bid(1, 8, KaiserContractTrump); err != nil {
		t.Fatalf("Bid: %v", err)
	}
	for _, seat := range []int{2, 3, 0} {
		if err := k.PassBid(seat); err != nil {
			t.Fatalf("PassBid(%d): %v", seat, err)
		}
	}

	if got := k.GetPhase(); got != KaiserPhaseDiscard {
		t.Fatalf("phase = %v, want discard", got)
	}
	if got := k.GetDeclarerIdx(); got != 1 {
		t.Fatalf("declarer = %d, want 1", got)
	}
	// **落札者はキティを手札に加える。**8 + 2 = 10 枚になる。
	if got := k.GetPlayer(1).GetCardsSize(); got != KaiserHandSize+KaiserKittySize {
		t.Errorf("the declarer holds %d, want %d", got, KaiserHandSize+KaiserKittySize)
	}
	if got := k.GetKittySize(); got != 0 {
		t.Errorf("the kitty holds %d after being taken, want 0", got)
	}
}

// TestKaiserSpecialCardsMayNotBeDiscarded covers the restriction the issue omits.
//
// これを許すと、落札者は ♠3 を無条件で処分できてしまう。
func TestKaiserSpecialCardsMayNotBeDiscarded(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	k.SetPhaseForTest(KaiserPhaseDiscard)
	k.SetContractForTest(0, 8, CardDesignHeart, KaiserContractTrump)
	k.SetHandForTest(0, []*Card{
		kzCard(CardDesignSpade, 3), kzCard(CardDesignHeart, 5),
		kzCard(CardDesignClover, 7), kzCard(CardDesignClover, 8),
	})

	if err := k.Discard(0, []int{0, 2}); err == nil {
		t.Error("the three of spades may not be discarded")
	}
	if err := k.Discard(0, []int{1, 2}); err == nil {
		t.Error("the five of hearts may not be discarded")
	}
	if err := k.Discard(0, []int{2, 3}); err != nil {
		t.Errorf("ordinary cards may be discarded: %v", err)
	}
	if got := k.GetPhase(); got != KaiserPhasePlay {
		t.Errorf("phase = %v, want play after the discard", got)
	}
	// **リードは落札者。**
	if got := k.GetTrickLeaderIdx(); got != 0 {
		t.Errorf("leader = %d, want the declarer", got)
	}
}

func TestKaiserDiscardGuards(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	k.SetPhaseForTest(KaiserPhaseDiscard)
	k.SetContractForTest(0, 8, CardDesignHeart, KaiserContractTrump)
	k.SetHandForTest(0, []*Card{
		kzCard(CardDesignClover, 7), kzCard(CardDesignClover, 8), kzCard(CardDesignClover, 9),
	})

	if err := k.Discard(1, []int{0, 1}); err == nil {
		t.Error("only the declarer discards")
	}
	if err := k.Discard(0, []int{0}); err == nil {
		t.Error("exactly two cards must go")
	}
	if err := k.Discard(0, []int{0, 0}); err == nil {
		t.Error("the same card cannot go twice")
	}
	if err := k.Discard(0, []int{0, 99}); err == nil {
		t.Error("an out-of-range index must be refused")
	}
}

// **切札を指定する前には捨てられない。**捨てる札の判断が切札に依存する。
func TestKaiserTrumpMustBeNamedBeforeDiscarding(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	k.SetPhaseForTest(KaiserPhaseDiscard)
	k.SetContractForTest(0, 8, 0, KaiserContractTrump)
	k.SetHandForTest(0, []*Card{
		kzCard(CardDesignClover, 7), kzCard(CardDesignClover, 8), kzCard(CardDesignClover, 9),
	})
	if err := k.Discard(0, []int{0, 1}); err == nil {
		t.Error("the trump suit must be named first")
	}
	if err := k.SetTrump(0, CardDesignClover); err != nil {
		t.Fatalf("SetTrump: %v", err)
	}
	if err := k.Discard(0, []int{0, 1}); err != nil {
		t.Errorf("Discard: %v", err)
	}
}

func TestKaiserSetTrumpGuards(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	k.SetPhaseForTest(KaiserPhaseDiscard)
	k.SetContractForTest(0, 8, 0, KaiserContractTrump)

	if err := k.SetTrump(1, CardDesignSpade); err == nil {
		t.Error("only the declarer names trump")
	}
	if err := k.SetTrump(0, 99); err == nil {
		t.Error("a bad suit must be refused")
	}
	// ノートランプ契約では切札を指定できない。
	k.SetContractForTest(0, 8, 0, KaiserContractNoTrump)
	if err := k.SetTrump(0, CardDesignSpade); err == nil {
		t.Error("a no-trump contract has no trump suit")
	}
	// プレイに入ったあとも指定できない。
	k.SetPhaseForTest(KaiserPhasePlay)
	if err := k.SetTrump(0, CardDesignSpade); err == nil {
		t.Error("trump cannot be renamed once play starts")
	}
}

// TestKaiserSpecialCardsScore covers the +5 / -3 that give the game its name.
func TestKaiserSpecialCardsScore(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	// 席 0 が ♥ をリードして ♥A で取る。♥5 と ♠3 が同じトリックに落ちる。
	k.SetHandForTest(0, []*Card{kzCard(CardDesignHeart, 1)})
	k.SetHandForTest(1, []*Card{kzCard(CardDesignHeart, 5)})
	k.SetHandForTest(2, []*Card{kzCard(CardDesignHeart, 7)})
	k.SetHandForTest(3, []*Card{kzCard(CardDesignSpade, 3)})
	k.SetTrickNumberForTest(KaiserHandSize - 1)

	for _, seat := range []int{0, 1, 2, 3} {
		if err := k.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}

	if got := k.GetHeartFiveBy(); got != 0 {
		t.Errorf("the five of hearts went to seat %d, want 0", got)
	}
	if got := k.GetSpadeThreeBy(); got != 0 {
		t.Errorf("the three of spades went to seat %d, want 0", got)
	}
	// トリック 1 + ♥5 の 5 − ♠3 の 3 = 3。
	if got := k.GetHandPoints(0); got != 1+KaiserHeartFiveBonus+KaiserSpadeThreePenalty {
		t.Errorf("team 0 scored %d, want %d", got, 1+KaiserHeartFiveBonus+KaiserSpadeThreePenalty)
	}
}

// TestKaiserTrumpTakesTheTrick covers plain trick resolution.
func TestKaiserTrumpTakesTheTrick(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k.SetHandForTest(0, []*Card{kzCard(CardDesignHeart, 1)})
	k.SetHandForTest(1, []*Card{kzCard(CardDesignHeart, 13)})
	k.SetHandForTest(2, []*Card{kzCard(CardDesignClover, 7)}) // 切札
	k.SetHandForTest(3, []*Card{kzCard(CardDesignHeart, 12)})

	for _, seat := range []int{0, 1, 2, 3} {
		if err := k.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	// 席 2 のチーム (team 0) が取る。
	if got := k.GetHandPoints(0); got != 1 {
		t.Errorf("team 0 took %d tricks, want 1", got)
	}
	if got := k.GetTrickLeaderIdx(); got != 2 {
		t.Errorf("the trick winner leads next, got %d", got)
	}
}

// 追随は強制。切札は強制ではない。
func TestKaiserFollowingSuitIsCompulsoryButTrumpingIsNot(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k.SetHandForTest(0, []*Card{kzCard(CardDesignHeart, 1)})
	k.SetHandForTest(1, []*Card{
		kzCard(CardDesignHeart, 13), kzCard(CardDesignClover, 7), kzCard(CardDesignDiamond, 8),
	})
	if err := k.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	valid := k.KaiserValidPlays(1)
	if len(valid) != 1 || valid[0] != 0 {
		t.Errorf("valid = %v, want only the heart", valid)
	}
	if err := k.PlayCard(1, 1); err == nil {
		t.Error("trumping while able to follow must be refused")
	}

	// フォローできなければ何でも出せる。**切札は強制ではない。**
	k2 := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k2.SetHandForTest(0, []*Card{kzCard(CardDesignHeart, 1)})
	k2.SetHandForTest(1, []*Card{kzCard(CardDesignClover, 7), kzCard(CardDesignDiamond, 8)})
	if err := k2.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := len(k2.KaiserValidPlays(1)); got != 2 {
		t.Errorf("with no card of the suit both are legal, got %d", got)
	}
}

func TestKaiserPlayGuards(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k.SetHandForTest(0, []*Card{kzCard(CardDesignHeart, 1)})
	if err := k.PlayCard(1, 0); err == nil {
		t.Error("playing out of turn must be refused")
	}
	if err := k.PlayCard(0, 99); err == nil {
		t.Error("an out-of-range index must be refused")
	}
	k.SetPhaseForTest(KaiserPhaseBid)
	if err := k.PlayCard(0, 0); err == nil {
		t.Error("playing outside the play phase must be refused")
	}
}

// TestKaiserSettlement covers the making / being-set rule.
func TestKaiserSettlement(t *testing.T) {
	t.Run("making the bid scores what was taken", func(t *testing.T) {
		k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
		k.SetContractForTest(0, 8, CardDesignClover, KaiserContractTrump)
		k.SetHandPointsForTest(0, 9)
		k.SetHandPointsForTest(1, 1)
		k.FinishHandForTest()

		if !k.IsBidMade() {
			t.Fatal("9 makes a bid of 8")
		}
		if got := k.GetScore(0); got != 9 {
			t.Errorf("the declaring team scores %d, want 9 (what it took, not the bid)", got)
		}
		if got := k.GetScore(1); got != 1 {
			t.Errorf("the defenders score %d, want 1", got)
		}
	})

	// **未達なら宣言額をそのままマイナス。**取った点は入らない。
	t.Run("being set costs the bid", func(t *testing.T) {
		k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
		k.SetContractForTest(0, 9, CardDesignClover, KaiserContractTrump)
		k.SetHandPointsForTest(0, 6)
		k.SetHandPointsForTest(1, 4)
		k.FinishHandForTest()

		if k.IsBidMade() {
			t.Fatal("6 does not make a bid of 9")
		}
		if got := k.GetScore(0); got != -9 {
			t.Errorf("the set team scores %d, want -9", got)
		}
		if got := k.GetScore(1); got != 4 {
			t.Errorf("the defenders still score their %d", got)
		}
	})

	// 宣言額ちょうどは達成。
	t.Run("exactly the bid makes it", func(t *testing.T) {
		k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
		k.SetContractForTest(0, 8, CardDesignClover, KaiserContractTrump)
		k.SetHandPointsForTest(0, 8)
		k.FinishHandForTest()
		if !k.IsBidMade() {
			t.Error("scoring exactly the bid makes it")
		}
	})
}

// TestKaiserMustBidToScoreFromFortyFive covers the rule the issue omits.
func TestKaiserMustBidToScoreFromFortyFive(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k.SetContractForTest(0, 8, CardDesignClover, KaiserContractTrump)
	k.SetScoreForTest(1, KaiserMustBidThreshold)
	k.SetHandPointsForTest(0, 9)
	k.SetHandPointsForTest(1, 1)
	k.FinishHandForTest()

	// **45 点以上の側は自分がビッドしない限り加点できない。**
	if got := k.GetScore(1); got != KaiserMustBidThreshold {
		t.Errorf("the defenders scored %d, want %d — they must bid to get out",
			got, KaiserMustBidThreshold)
	}
	if got := k.GetScore(0); got != 9 {
		t.Errorf("the declarers still score %d, want 9", got)
	}
}

// TestKaiserNoTrumpRaisesTheTarget covers the target the issue omits.
func TestKaiserNoTrumpRaisesTheTarget(t *testing.T) {
	k := kzPlaying(t, 0, KaiserContractNoTrump)
	if got := k.GetTargetScore(); got != KaiserTargetScore {
		t.Fatalf("target starts at %d, want %d", got, KaiserTargetScore)
	}
	k.SetContractForTest(0, 8, 0, KaiserContractNoTrump)
	k.SetHandPointsForTest(0, 9)
	k.FinishHandForTest()

	if got := k.GetTargetScore(); got != KaiserNoTrumpTargetScore {
		t.Errorf("target = %d, want %d after a successful no-trump bid", got, KaiserNoTrumpTargetScore)
	}

	// 切札の契約では上がらない。
	k2 := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k2.SetContractForTest(0, 8, CardDesignClover, KaiserContractTrump)
	k2.SetHandPointsForTest(0, 9)
	k2.FinishHandForTest()
	if got := k2.GetTargetScore(); got != KaiserTargetScore {
		t.Errorf("target = %d, want it unchanged after a trump contract", got)
	}
}

func TestKaiserGameEnd(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k.SetContractForTest(0, 8, CardDesignClover, KaiserContractTrump)
	k.SetScoreForTest(0, KaiserTargetScore-5)
	k.SetHandPointsForTest(0, 9)
	k.FinishHandForTest()

	if !k.GetGameEndFlag() {
		t.Fatal("passing the target ends the game")
	}
	if got := k.GetWinnerTeam(); got != 0 {
		t.Errorf("winner = %d, want 0", got)
	}
	if err := k.NextHand(); err == nil {
		t.Error("dealing after the game is over must be refused")
	}
}

// 両チームが同時に超えたら落札側の勝ち。
func TestKaiserGameEndTieBreak(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k.SetContractForTest(1, 8, CardDesignClover, KaiserContractTrump)
	k.SetScoreForTest(0, KaiserTargetScore)
	k.SetScoreForTest(1, KaiserTargetScore)
	k.checkGameEnd()
	if got := k.GetWinnerTeam(); got != KaiserTeamOf(1) {
		t.Errorf("winner = %d, want the declaring team", got)
	}
}

func TestKaiserAllPassRedeals(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	hand := k.GetHandNumber()
	for _, seat := range []int{1, 2, 3, 0} {
		if err := k.PassBid(seat); err != nil {
			t.Fatalf("PassBid(%d): %v", seat, err)
		}
	}
	if got := k.GetPhase(); got != KaiserPhaseBid {
		t.Errorf("phase = %v, want a fresh bidding round", got)
	}
	// **流局は局数に数えない。**
	if got := k.GetHandNumber(); got != hand {
		t.Errorf("hand number = %d, want %d", got, hand)
	}
	if got := k.GetKittySize(); got != KaiserKittySize {
		t.Errorf("the redeal must produce a kitty again, got %d", got)
	}
}

func TestKaiserNextHandRotatesTheDealer(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k.SetContractForTest(0, 8, CardDesignClover, KaiserContractTrump)
	k.SetHandPointsForTest(0, 9)
	k.FinishHandForTest()

	dealer := k.GetDealerIdx()
	if err := k.NextHand(); err != nil {
		t.Fatalf("NextHand: %v", err)
	}
	if got := k.GetDealerIdx(); got == dealer {
		t.Errorf("the dealer stayed at %d; it must rotate", got)
	}
	if got := k.GetKittySize(); got != KaiserKittySize {
		t.Errorf("the next hand must produce a kitty, got %d", got)
	}
}

func TestKaiserNextHandGuards(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	if err := k.NextHand(); err == nil {
		t.Error("dealing again mid-hand must be refused")
	}
}

func TestKaiserBidGuards(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	if err := k.Bid(0, 8, KaiserContractTrump); err == nil {
		t.Error("bidding out of turn must be refused")
	}
	if err := k.Bid(1, 8, KaiserContract(99)); err == nil {
		t.Error("a bad contract must be refused")
	}
	k.SetPhaseForTest(KaiserPhasePlay)
	if err := k.PassBid(1); err == nil {
		t.Error("bidding outside the bid phase must be refused")
	}
}

func TestKaiserIsHumanTurnAndCpuPlay(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	// 席 1 (CPU) がビッドの先手。
	if k.IsHumanTurn() {
		t.Error("the dealer's left bids first and it is a CPU")
	}
	k.CpuPlay()
	if k.GetBidPlayerIdx() == 1 && k.GetPhase() == KaiserPhaseBid {
		t.Error("CpuPlay must move the bidding along")
	}

	k2 := NewDefaultKaiser()
	k2.Reset()
	k2.SetPhaseForTest(KaiserPhaseGameEnd)
	k2.gameEndFlag = true
	if k2.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	k2.CpuPlay()
}

// **CPU だけで 1 局を回し切れること。**途中で止まると詰む。
func TestKaiserCpuDrivesAFullHand(t *testing.T) {
	for attempt := range 30 {
		k := NewDefaultKaiser()
		k.Reset()
		for step := 0; step < 600; step++ {
			if k.GetPhase() == KaiserPhaseHandEnd || k.GetGameEndFlag() {
				break
			}
			if !k.IsHumanTurn() {
				k.CpuPlay()
				continue
			}
			// 人間席も CPU の判断で埋める。
			switch k.GetPhase() {
			case KaiserPhaseBid:
				idx := k.GetBidPlayerIdx()
				value, contract, _ := k.KaiserCpuBid(idx)
				if value < KaiserMinBid || k.Bid(idx, value, contract) != nil {
					_ = k.PassBid(idx)
				}
			case KaiserPhaseDiscard:
				idx := k.GetDeclarerIdx()
				if k.GetContract() == KaiserContractTrump && k.GetTrumpSuit() == 0 {
					_, _, suit := k.KaiserCpuBid(idx)
					if suit == 0 {
						suit = CardDesignSpade
					}
					_ = k.SetTrump(idx, suit)
				}
				_ = k.Discard(idx, k.KaiserCpuDiscard(idx))
			case KaiserPhasePlay:
				idx := k.GetCurrentPlayerIdx()
				if i := k.KaiserCpuPlay(idx); i >= 0 {
					_ = k.PlayCard(idx, i)
				}
			}
		}
		if k.GetPhase() != KaiserPhaseHandEnd && !k.GetGameEndFlag() {
			t.Fatalf("attempt %d: the hand never finished (phase %v)", attempt, k.GetPhase())
		}
		// **8 トリックすべてが解決している。**
		if got := k.GetTrickNumber(); got != KaiserHandSize {
			t.Fatalf("attempt %d: %d tricks played, want %d", attempt, got, KaiserHandSize)
		}
		// 両チームの点の合計は必ず 10。
		if got := k.GetHandPoints(0) + k.GetHandPoints(1); got != KaiserHandTotal {
			t.Fatalf("attempt %d: the hand scored %d in total, want %d", attempt, got, KaiserHandTotal)
		}
	}
}

// **CPU は ♥5 と ♠3 を捨てない。**捨てられると ♠3 が無条件で処分できる。
func TestKaiserCpuDiscardKeepsTheSpecialCards(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	k.SetPhaseForTest(KaiserPhaseDiscard)
	k.SetContractForTest(0, 8, CardDesignClover, KaiserContractTrump)
	k.SetHandForTest(0, []*Card{
		kzCard(CardDesignSpade, 3), kzCard(CardDesignHeart, 5),
		kzCard(CardDesignDiamond, 7), kzCard(CardDesignDiamond, 8),
	})
	got := k.KaiserCpuDiscard(0)
	if len(got) != KaiserKittySize {
		t.Fatalf("the CPU chose %d cards, want %d", len(got), KaiserKittySize)
	}
	for _, i := range got {
		c := k.GetPlayer(0).GetCard(i)
		if IsKaiserHeartFive(c) || IsKaiserSpadeThree(c) {
			t.Errorf("the CPU tried to discard a scoring card at index %d", i)
		}
	}
	if k.KaiserCpuDiscard(99) != nil {
		t.Error("an unknown seat has nothing to discard")
	}
}

func TestKaiserCpuPlayAndBidEdges(t *testing.T) {
	k := kzPlaying(t, CardDesignClover, KaiserContractTrump)
	k.SetHandForTest(0, []*Card{})
	if got := k.KaiserCpuPlay(0); got != -1 {
		t.Errorf("an empty hand has no play, got %d", got)
	}
	if got := k.KaiserCpuPlay(99); got != -1 {
		t.Errorf("an unknown seat has no play, got %d", got)
	}
	if v, _, _ := k.KaiserCpuBid(99); v != 0 {
		t.Errorf("an unknown seat bids %d, want 0", v)
	}
}

func TestKaiserAccessors(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	if got := k.GetHandNumber(); got != 1 {
		t.Errorf("hand number = %d, want 1", got)
	}
	if got := k.GetWinnerTeam(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if got := k.GetDeclarerIdx(); got != -1 {
		t.Errorf("declarer = %d, want -1 before anyone bids", got)
	}
	if k.GetHighBid() != nil {
		t.Error("no bid stands at the start")
	}
	if got := k.GetTargetScore(); got != KaiserTargetScore {
		t.Errorf("target = %d, want %d", got, KaiserTargetScore)
	}
	if got := len(k.GetPlayers()); got != KaiserPlayerCnt {
		t.Errorf("%d seats, want %d", got, KaiserPlayerCnt)
	}
	if k.GetPlayer(-1) != nil || k.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
	}
	if k.GetHandPoints(-1) != 0 || k.GetScore(99) != 0 {
		t.Error("out-of-range scores must be 0, not a panic")
	}
	if got := len(k.GetTrick()); got != 0 {
		t.Errorf("the trick starts empty, got %d", got)
	}
	if len(k.GetActionLog()) == 0 {
		t.Error("dealing writes to the action log")
	}
	// チーム分けは席の偶奇。
	if KaiserTeamOf(0) != KaiserTeamOf(2) || KaiserTeamOf(1) != KaiserTeamOf(3) {
		t.Error("seats 0/2 and 1/3 are partners")
	}
	if KaiserTeamOf(0) == KaiserTeamOf(1) {
		t.Error("neighbours are opponents")
	}
	if k.KaiserValidPlays(99) != nil {
		t.Error("an unknown seat has no legal plays")
	}
}

func TestKaiserConfigValidate(t *testing.T) {
	if err := DefaultKaiserConfig().Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	if err := (KaiserConfig{CpuDifficulty: 9}).Validate(); err == nil {
		t.Error("a bad difficulty must not validate")
	}
}

func TestKaiserRoundTripsThroughJSON(t *testing.T) {
	k := NewDefaultKaiser()
	k.Reset()
	_ = k.Bid(1, 8, KaiserContractTrump)

	data, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Kaiser
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetHighBid() == nil || restored.GetHighBid().Value != 8 {
		t.Errorf("the standing bid did not survive the round trip")
	}
	if got := restored.GetKittySize(); got != KaiserKittySize {
		t.Errorf("the kitty holds %d after a restore, want %d", got, KaiserKittySize)
	}
	if got := restored.GetTargetScore(); got != KaiserTargetScore {
		t.Errorf("target = %d, want %d", got, KaiserTargetScore)
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestKaiserRejectsBadJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[],"ph":0,"di":0,"ci":0,"bi":0}`},
		{"bad phase", `{"pl":[{},{},{},{}],"ph":99,"di":0,"ci":0,"bi":0}`},
		{"bad dealer", `{"pl":[{},{},{},{}],"ph":0,"di":9,"ci":0,"bi":0}`},
		{"bad declarer", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":9}`},
		{"bad winner team", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"wt":5}`},
		{"bad contract", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"wt":-1,"co":9}`},
		{"bad trump suit", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"wt":-1,"ts":9}`},
		{"oversized trick", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"wt":-1,"tk":[{},{},{},{},{}]}`},
		{"impossible high bid", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"wt":-1,"hb":{"Value":99}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var k Kaiser
			if err := json.Unmarshal([]byte(tc.body), &k); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}
