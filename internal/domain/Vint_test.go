//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"
)

func vtCard(suit, value int) *Card { return NewCard(suit, value, true) }

// vtPlaying puts a game into the play phase with a fixed contract.
func vtPlaying(t *testing.T, declarer, level, denom int) *Vint {
	t.Helper()
	v := NewDefaultVint()
	v.Reset()
	v.SetPhaseForTest(VintPhasePlay)
	v.SetContractForTest(declarer, level, denom)
	v.SetCurrentPlayerForTest(0)
	v.SetTrickLeaderForTest(0)
	return v
}

// TestVintBidSuitOrderIsNotBridge は issue が触れていない序列を押さえる。
//
// **♠ < ♣ < ♦ < ♥ < NT。**ブリッジ (♣<♦<♥<♠<NT) とは違い ♠ が最弱である。
func TestVintBidSuitOrderIsNotBridge(t *testing.T) {
	ascending := []int{VintDenomSpade, VintDenomClub, VintDenomDiamond, VintDenomHeart, VintDenomNoTrump}
	if len(ascending) != VintDenomCount {
		t.Fatalf("the ladder lists %d denominations, want %d", len(ascending), VintDenomCount)
	}
	for i := 1; i < len(ascending); i++ {
		if ascending[i-1] >= ascending[i] {
			t.Errorf("denomination %d must rank below %d", ascending[i-1], ascending[i])
		}
	}
	// **♠ が最弱で ♣ より下。**ブリッジなら逆。
	if VintDenomSpade >= VintDenomClub {
		t.Error("spades are the LOWEST denomination in Vint, below clubs")
	}
	// 同レベルなら NT が最強。
	if VintBidRank(VintDenomNoTrump, 3) <= VintBidRank(VintDenomHeart, 3) {
		t.Error("no trump outranks hearts at the same level")
	}
	// レベルが上なら常に強い。
	if VintBidRank(VintDenomSpade, 4) <= VintBidRank(VintDenomNoTrump, 3) {
		t.Error("a higher level always outranks a lower one")
	}
	if VintBidRank(99, 3) != 0 || VintBidRank(VintDenomSpade, 99) != 0 {
		t.Error("an out-of-range bid has no rank")
	}
}

// TestVintTrickValueDependsOnSuitAndLevel は issue が触れていない単価表を押さえる。
func TestVintTrickValueDependsOnSuitAndLevel(t *testing.T) {
	// レベル 1 の基準値。
	for _, tc := range []struct {
		denom, want int
	}{
		{VintDenomSpade, 4},
		{VintDenomClub, 6},
		{VintDenomDiamond, 8},
		{VintDenomHeart, 10},
		{VintDenomNoTrump, 12},
	} {
		if got := VintTrickValue(tc.denom, 1); got != tc.want {
			t.Errorf("denomination %d at level 1 = %d, want %d", tc.denom, got, tc.want)
		}
	}
	// **レベルが 1 上がるごとに +10。**
	if got := VintTrickValue(VintDenomSpade, 2); got != 14 {
		t.Errorf("2 spades = %d, want 14", got)
	}
	if got := VintTrickValue(VintDenomSpade, 3); got != 24 {
		t.Errorf("3 spades = %d, want 24", got)
	}
	if got := VintTrickValue(VintDenomNoTrump, 7); got != 72 {
		t.Errorf("7 no trump = %d, want 72", got)
	}
	if VintTrickValue(99, 1) != 0 || VintTrickValue(VintDenomSpade, 0) != 0 {
		t.Error("an out-of-range bid has no trick value")
	}
}

// TestVintBothSidesScoreTheirTricks は issue の最大の誤りを押さえる。
//
// **守備側も自分のトリックを得点する。**達成/失敗に関係ない。
func TestVintBothSidesScoreTheirTricks(t *testing.T) {
	v := vtPlaying(t, 0, 1, VintDenomSpade) // 単価 4
	// 宣言側 (team 0) が 7、守備側 (team 1) が 6。
	v.SetTricksWonForTest(0, 4)
	v.SetTricksWonForTest(2, 3)
	v.SetTricksWonForTest(1, 3)
	v.SetTricksWonForTest(3, 3)
	v.FinishHandForTest()

	res := v.GetLastResult()
	if res == nil {
		t.Fatal("the settlement must produce a result")
	}
	if got := res.TrickPoints[0]; got != 7*4 {
		t.Errorf("the declaring side scores %d, want %d", got, 7*4)
	}
	// **守備側も 6 トリックぶん得点する。**issue の「宣言側だけ」は誤り。
	if got := res.TrickPoints[1]; got != 6*4 {
		t.Errorf("the DEFENDERS score %d, want %d — both sides score their tricks", got, 6*4)
	}
	if got := v.GetBelow(1); got != 6*4 {
		t.Errorf("the defenders' below-the-line total is %d, want %d", got, 6*4)
	}
}

// **失敗しても守備側は自分のトリックを得点する。**
func TestVintDefendersScoreEvenWhenTheContractFails(t *testing.T) {
	v := vtPlaying(t, 0, 3, VintDenomSpade) // 9 トリック必要、単価 24
	v.SetTricksWonForTest(0, 3)
	v.SetTricksWonForTest(2, 2) // 宣言側 5
	v.SetTricksWonForTest(1, 4)
	v.SetTricksWonForTest(3, 4) // 守備側 8
	v.FinishHandForTest()

	res := v.GetLastResult()
	if res.Made {
		t.Fatal("5 tricks fails a bid of 3 (9 tricks)")
	}
	// 失敗しても両者が線下に得点する。
	if got := res.TrickPoints[0]; got != 5*24 {
		t.Errorf("the failing declarer still scores %d, want %d", got, 5*24)
	}
	if got := res.TrickPoints[1]; got != 8*24 {
		t.Errorf("the defenders score %d, want %d", got, 8*24)
	}
	// **ペナルティは不足数 × レベル × 500。**4 × 3 × 500 = 6000。
	if got := res.Penalty[1]; got != 4*3*VintUndertrickUnit {
		t.Errorf("the penalty is %d, want %d", got, 4*3*VintUndertrickUnit)
	}
}

// **宣言レベル N は「6 + N トリック」。**
func TestVintContractTargetIsSixPlusLevel(t *testing.T) {
	made := vtPlaying(t, 0, 1, VintDenomSpade)
	made.SetTricksWonForTest(0, 4)
	made.SetTricksWonForTest(2, 3) // 7 トリック
	made.SetTricksWonForTest(1, 3)
	made.SetTricksWonForTest(3, 3)
	made.FinishHandForTest()
	if !made.GetLastResult().Made {
		t.Error("7 tricks makes a bid of 1")
	}

	failed := vtPlaying(t, 0, 1, VintDenomSpade)
	failed.SetTricksWonForTest(0, 3)
	failed.SetTricksWonForTest(2, 3) // 6 トリック
	failed.SetTricksWonForTest(1, 4)
	failed.SetTricksWonForTest(3, 3)
	failed.FinishHandForTest()
	if failed.GetLastResult().Made {
		t.Error("6 tricks fails a bid of 1")
	}
}

// TestVintHonoursNeedThreeOrMore covers the floor the issue omits.
func TestVintHonoursNeedThreeOrMore(t *testing.T) {
	const value = 10
	for _, tc := range []struct {
		count, want int
	}{
		{0, 0}, {1, 0}, {2, 0}, // **2 枚以下は 0 点。**
		{3, value * 20},
		{4, value * 30},
		{5, value * 40},
	} {
		if got := VintHonourBonus(tc.count, value); got != tc.want {
			t.Errorf("%d honours = %d, want %d", tc.count, got, tc.want)
		}
	}
}

func TestVintHonourIdentification(t *testing.T) {
	const trump = CardDesignHeart
	for _, val := range []int{1, 13, 12, 11, 10} {
		if !IsVintHonour(vtCard(trump, val), trump) {
			t.Errorf("trump %d is an honour", val)
		}
	}
	if IsVintHonour(vtCard(trump, 9), trump) {
		t.Error("the trump nine is not an honour")
	}
	// **切札以外のスートは対象外。**
	if IsVintHonour(vtCard(CardDesignSpade, 1), trump) {
		t.Error("an ace outside the trump suit is not a trump honour")
	}
	// **ノートランプではオナーが無い。**
	if IsVintHonour(vtCard(trump, 1), 0) {
		t.Error("there are no trump honours at no trump")
	}
	if IsVintHonour(nil, trump) {
		t.Error("a nil card is not an honour")
	}
}

// TestVintAcesAreScoredSeparately covers the rule the issue omits entirely.
func TestVintAcesAreScoredSeparately(t *testing.T) {
	const value = 10
	// 多く持つ側が**全部**取る。
	ours, theirs := VintAceBonus(3, 1, value, false)
	if ours != 4*VintAceMultiplier*value || theirs != 0 {
		t.Errorf("3-1 split gave %d/%d, want %d/0", ours, theirs, 4*VintAceMultiplier*value)
	}
	ours, theirs = VintAceBonus(1, 3, value, true)
	if ours != 0 || theirs != 4*VintAceMultiplier*value {
		t.Errorf("1-3 split gave %d/%d, want 0/%d", ours, theirs, 4*VintAceMultiplier*value)
	}
	// **2 対 2 はトリックの多い側が総取り。**
	ours, theirs = VintAceBonus(2, 2, value, true)
	if ours != 4*VintAceMultiplier*value || theirs != 0 {
		t.Errorf("a 2-2 split with more tricks gave %d/%d, want %d/0", ours, theirs, 4*VintAceMultiplier*value)
	}
	ours, theirs = VintAceBonus(2, 2, value, false)
	if ours != 0 || theirs != 4*VintAceMultiplier*value {
		t.Errorf("a 2-2 split with fewer tricks gave %d/%d, want 0/%d", ours, theirs, 4*VintAceMultiplier*value)
	}
}

// 局の精算でオナーとエースが線上に入ること。
func TestVintSettlementAddsHonoursAndAces(t *testing.T) {
	const trump = CardDesignHeart
	v := vtPlaying(t, 0, 1, VintDenomHeart) // 単価 10
	v.SetTricksWonForTest(0, 4)
	v.SetTricksWonForTest(2, 3)
	v.SetTricksWonForTest(1, 3)
	v.SetTricksWonForTest(3, 3)
	// team 0 が切札オナー 4 枚とエース 3 枚を取った。
	v.SetTakenForTest(0, []*Card{
		vtCard(trump, 1), vtCard(trump, 13), vtCard(trump, 12), vtCard(trump, 11),
		vtCard(CardDesignSpade, 1), vtCard(CardDesignClover, 1),
	})
	v.SetTakenForTest(1, []*Card{vtCard(CardDesignDiamond, 1)})
	v.FinishHandForTest()

	res := v.GetLastResult()
	// オナー 4 枚 → 単価 × 30。
	if got := res.HonourPoints[0]; got != 10*30 {
		t.Errorf("four honours scored %d, want %d", got, 10*30)
	}
	if got := res.HonourPoints[1]; got != 0 {
		t.Errorf("the other side scored %d for honours, want 0", got)
	}
	// エース 3 対 1 → 4 枚ぶん、team 0 が総取り。
	if got := res.AcePoints[0]; got != 4*VintAceMultiplier*10 {
		t.Errorf("aces scored %d, want %d", got, 4*VintAceMultiplier*10)
	}
	if got := res.AcePoints[1]; got != 0 {
		t.Errorf("the other side scored %d for aces, want 0", got)
	}
	// 線上に積まれている。
	if got := v.GetAbove(0); got != res.HonourPoints[0]+res.AcePoints[0] {
		t.Errorf("above the line = %d, want %d", got, res.HonourPoints[0]+res.AcePoints[0])
	}
}

// **ノートランプではオナーが付かない。**エースだけが数えられる。
func TestVintNoTrumpHasNoHonoursOnlyAces(t *testing.T) {
	v := vtPlaying(t, 0, 1, VintDenomNoTrump)
	v.SetTricksWonForTest(0, 4)
	v.SetTricksWonForTest(2, 3)
	v.SetTricksWonForTest(1, 3)
	v.SetTricksWonForTest(3, 3)
	v.SetTakenForTest(0, []*Card{
		vtCard(CardDesignHeart, 1), vtCard(CardDesignHeart, 13),
		vtCard(CardDesignHeart, 12), vtCard(CardDesignSpade, 1),
	})
	v.FinishHandForTest()

	res := v.GetLastResult()
	if got := res.HonourPoints[0]; got != 0 {
		t.Errorf("no trump scored %d for honours, want 0", got)
	}
	// エースは数える。
	if res.AcePoints[0] <= 0 {
		t.Error("aces still score at no trump")
	}
}

// TestVintGameAndRubber covers the target and the bonuses.
func TestVintGameAndRubber(t *testing.T) {
	t.Run("500 below the line takes a game", func(t *testing.T) {
		v := vtPlaying(t, 0, 7, VintDenomNoTrump) // 単価 72
		v.SetBelowForTest(0, VintGameTarget-100)
		v.SetTricksWonForTest(0, 7)
		v.SetTricksWonForTest(2, 6)
		v.FinishHandForTest()

		if got := v.GetGamesWon(0); got != 1 {
			t.Fatalf("games won = %d, want 1", got)
		}
		// **ゲームを取ったら線下はリセットされる。**
		if got := v.GetBelow(0); got != 0 {
			t.Errorf("below the line = %d, want 0 after taking a game", got)
		}
		if v.GetAbove(0) < VintFirstGameBonus {
			t.Errorf("the first game bonus of %d must be added above the line", VintFirstGameBonus)
		}
		if v.GetGameEndFlag() {
			t.Error("one game is not the rubber")
		}
	})

	t.Run("the second game takes the rubber", func(t *testing.T) {
		v := vtPlaying(t, 0, 7, VintDenomNoTrump)
		v.SetGamesWonForTest(0, 1)
		v.SetBelowForTest(0, VintGameTarget-100)
		v.SetTricksWonForTest(0, 7)
		v.SetTricksWonForTest(2, 6)
		v.FinishHandForTest()

		if !v.GetGameEndFlag() {
			t.Fatal("the second game ends the rubber")
		}
		if got := v.GetWinnerTeam(); got != 0 {
			t.Errorf("winner = %d, want 0", got)
		}
		if got := v.GetPhase(); got != VintPhaseGameEnd {
			t.Errorf("phase = %v, want game end", got)
		}
		if err := v.NextHand(); err == nil {
			t.Error("dealing after the rubber must be refused")
		}
	})
}

func TestVintBidding(t *testing.T) {
	t.Run("the dealer's left bids first", func(t *testing.T) {
		v := NewDefaultVint()
		v.Reset()
		if got := v.GetBidPlayerIdx(); got != 1 {
			t.Errorf("bid seat = %d, want 1", got)
		}
	})

	t.Run("a bid must beat the standing one", func(t *testing.T) {
		v := NewDefaultVint()
		v.Reset()
		if err := v.Bid(1, 3, VintDenomHeart); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		// 同レベルの ♦ は ♥ より下なので通らない。
		if err := v.Bid(2, 3, VintDenomDiamond); err == nil {
			t.Error("diamonds rank below hearts and must be refused")
		}
		// 同レベルの NT は通る。
		if err := v.Bid(2, 3, VintDenomNoTrump); err != nil {
			t.Errorf("no trump outranks hearts: %v", err)
		}
	})

	t.Run("three passes settle the contract", func(t *testing.T) {
		v := NewDefaultVint()
		v.Reset()
		if err := v.Bid(1, 2, VintDenomHeart); err != nil {
			t.Fatalf("Bid: %v", err)
		}
		for _, seat := range []int{2, 3, 0} {
			if err := v.PassBid(seat); err != nil {
				t.Fatalf("PassBid(%d): %v", seat, err)
			}
		}
		if got := v.GetDeclarerIdx(); got != 1 {
			t.Errorf("declarer = %d, want 1", got)
		}
		if got := v.GetPhase(); got != VintPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
		if got := v.GetTrumpSuit(); got != CardDesignHeart {
			t.Errorf("trump = %d, want hearts", got)
		}
		// **リードはディーラーの左隣。**
		if got := v.GetTrickLeaderIdx(); got != 1 {
			t.Errorf("leader = %d, want the dealer's left", got)
		}
	})

	// ノートランプ契約では切札スートが 0 になる。
	t.Run("a no-trump contract leaves no trump suit", func(t *testing.T) {
		v := NewDefaultVint()
		v.Reset()
		_ = v.Bid(1, 2, VintDenomNoTrump)
		for _, seat := range []int{2, 3, 0} {
			_ = v.PassBid(seat)
		}
		if got := v.GetTrumpSuit(); got != 0 {
			t.Errorf("trump = %d, want 0 at no trump", got)
		}
	})

	t.Run("four passes redeal without advancing the hand number", func(t *testing.T) {
		v := NewDefaultVint()
		v.Reset()
		hand := v.GetHandNumber()
		for _, seat := range []int{1, 2, 3, 0} {
			_ = v.PassBid(seat)
		}
		if got := v.GetPhase(); got != VintPhaseBid {
			t.Errorf("phase = %v, want a fresh auction", got)
		}
		if got := v.GetHandNumber(); got != hand {
			t.Errorf("hand number = %d, want %d", got, hand)
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		v := NewDefaultVint()
		v.Reset()
		if err := v.Bid(0, 2, VintDenomHeart); err == nil {
			t.Error("bidding out of turn must be refused")
		}
		if err := v.Bid(1, 0, VintDenomHeart); err == nil {
			t.Error("a level below the minimum must be refused")
		}
		if err := v.Bid(1, 8, VintDenomHeart); err == nil {
			t.Error("a level above the maximum must be refused")
		}
		if err := v.Bid(1, 2, 99); err == nil {
			t.Error("a bad denomination must be refused")
		}
		v.SetPhaseForTest(VintPhasePlay)
		if err := v.PassBid(1); err == nil {
			t.Error("bidding outside the auction must be refused")
		}
	})
}

// 追随は強制。
func TestVintFollowingSuitIsCompulsory(t *testing.T) {
	v := vtPlaying(t, 0, 2, VintDenomHeart)
	v.SetHandForTest(0, []*Card{vtCard(CardDesignSpade, 1)})
	v.SetHandForTest(1, []*Card{vtCard(CardDesignSpade, 13), vtCard(CardDesignHeart, 2)})
	if err := v.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	valid := v.VintValidPlays(1)
	if len(valid) != 1 || valid[0] != 0 {
		t.Errorf("valid = %v, want only the spade", valid)
	}
	if err := v.PlayCard(1, 1); err == nil {
		t.Error("trumping while able to follow must be refused")
	}
}

func TestVintPlayGuards(t *testing.T) {
	v := vtPlaying(t, 0, 2, VintDenomHeart)
	v.SetHandForTest(0, []*Card{vtCard(CardDesignSpade, 1)})
	if err := v.PlayCard(1, 0); err == nil {
		t.Error("playing out of turn must be refused")
	}
	if err := v.PlayCard(0, 99); err == nil {
		t.Error("an out-of-range index must be refused")
	}
	v.SetPhaseForTest(VintPhaseBid)
	if err := v.PlayCard(0, 0); err == nil {
		t.Error("playing outside the play phase must be refused")
	}
	if v.VintValidPlays(99) != nil {
		t.Error("an unknown seat has no legal plays")
	}
}

// 切札が勝ち、ノートランプでは効かない。
func TestVintTrumpResolution(t *testing.T) {
	withTrump := vtPlaying(t, 0, 2, VintDenomHeart)
	withTrump.SetHandForTest(0, []*Card{vtCard(CardDesignSpade, 1)})
	withTrump.SetHandForTest(1, []*Card{vtCard(CardDesignHeart, 2)})
	withTrump.SetHandForTest(2, []*Card{vtCard(CardDesignSpade, 13)})
	withTrump.SetHandForTest(3, []*Card{vtCard(CardDesignSpade, 12)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := withTrump.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	if got := withTrump.GetTricksWon(1); got != 1 {
		t.Errorf("the low trump takes the trick, seat 1 has %d", got)
	}

	noTrump := vtPlaying(t, 0, 2, VintDenomNoTrump)
	noTrump.SetHandForTest(0, []*Card{vtCard(CardDesignSpade, 1)})
	noTrump.SetHandForTest(1, []*Card{vtCard(CardDesignHeart, 2)})
	noTrump.SetHandForTest(2, []*Card{vtCard(CardDesignSpade, 13)})
	noTrump.SetHandForTest(3, []*Card{vtCard(CardDesignSpade, 12)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := noTrump.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	if got := noTrump.GetTricksWon(0); got != 1 {
		t.Errorf("at no trump the led ace wins, seat 0 has %d", got)
	}
}

// **取った札はチームごとに残す。**オナーとエースの集計に要る。
func TestVintKeepsTakenCardsPerTeam(t *testing.T) {
	v := vtPlaying(t, 0, 2, VintDenomHeart)
	v.SetHandForTest(0, []*Card{vtCard(CardDesignSpade, 1)})
	v.SetHandForTest(1, []*Card{vtCard(CardDesignSpade, 2)})
	v.SetHandForTest(2, []*Card{vtCard(CardDesignSpade, 3)})
	v.SetHandForTest(3, []*Card{vtCard(CardDesignSpade, 4)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := v.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	// 席 0 (team 0) が取ったので 4 枚とも team 0 に付く。
	if got := len(v.takenCards[0]); got != VintPlayerCnt {
		t.Errorf("team 0 kept %d cards, want %d", got, VintPlayerCnt)
	}
	if got := len(v.takenCards[1]); got != 0 {
		t.Errorf("team 1 kept %d cards, want 0", got)
	}
}

func TestVintTeamTricks(t *testing.T) {
	v := vtPlaying(t, 0, 2, VintDenomHeart)
	v.SetTricksWonForTest(0, 3)
	v.SetTricksWonForTest(2, 2)
	v.SetTricksWonForTest(1, 4)
	v.SetTricksWonForTest(3, 4)
	if got := v.VintTeamTricks(0); got != 5 {
		t.Errorf("team 0 took %d, want 5", got)
	}
	if got := v.VintTeamTricks(1); got != 8 {
		t.Errorf("team 1 took %d, want 8", got)
	}
	if v.VintTeamTricks(99) != 0 {
		t.Error("an out-of-range team took nothing")
	}
	// パートナーは向かい合わせ。
	if VintTeamOf(0) != VintTeamOf(2) || VintTeamOf(1) != VintTeamOf(3) {
		t.Error("seats 0/2 and 1/3 are partners")
	}
}

func TestVintNextHandRotatesTheDealer(t *testing.T) {
	v := vtPlaying(t, 0, 1, VintDenomSpade)
	v.SetTricksWonForTest(0, 7)
	v.FinishHandForTest()

	dealer := v.GetDealerIdx()
	if err := v.NextHand(); err != nil {
		t.Fatalf("NextHand: %v", err)
	}
	if got := v.GetDealerIdx(); got == dealer {
		t.Errorf("the dealer stayed at %d; it must rotate", got)
	}
	for i := range VintPlayerCnt {
		if got := v.GetPlayer(i).GetCardsSize(); got != VintHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, VintHandSize)
		}
	}
}

func TestVintNextHandGuards(t *testing.T) {
	v := vtPlaying(t, 0, 1, VintDenomSpade)
	if err := v.NextHand(); err == nil {
		t.Error("dealing again mid-hand must be refused")
	}
}

func TestVintIsHumanTurnAndCpuPlay(t *testing.T) {
	v := NewDefaultVint()
	v.Reset()
	if v.IsHumanTurn() {
		t.Error("the dealer's left bids first and it is a CPU")
	}
	v.CpuPlay()
	if v.GetBidPlayerIdx() == 1 && v.GetPhase() == VintPhaseBid {
		t.Error("CpuPlay must move the auction along")
	}

	v2 := NewDefaultVint()
	v2.Reset()
	v2.SetPhaseForTest(VintPhaseGameEnd)
	v2.gameEndFlag = true
	if v2.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	v2.CpuPlay()
}

// **CPU だけで 1 局を回し切れること。**途中で止まると詰む。
func TestVintCpuDrivesAFullHand(t *testing.T) {
	for attempt := range 30 {
		v := NewDefaultVint()
		v.Reset()
		for step := 0; step < 800; step++ {
			if v.GetPhase() == VintPhaseHandEnd || v.GetGameEndFlag() {
				break
			}
			if !v.IsHumanTurn() {
				v.CpuPlay()
				continue
			}
			switch v.GetPhase() {
			case VintPhaseBid:
				idx := v.GetBidPlayerIdx()
				level, denom := v.VintCpuBid(idx)
				if level < VintMinLevel || v.Bid(idx, level, denom) != nil {
					_ = v.PassBid(idx)
				}
			case VintPhasePlay:
				idx := v.GetCurrentPlayerIdx()
				if i := v.VintCpuPlay(idx); i >= 0 {
					_ = v.PlayCard(idx, i)
				}
			}
		}
		if v.GetPhase() != VintPhaseHandEnd && !v.GetGameEndFlag() {
			t.Fatalf("attempt %d: the hand never finished (phase %v)", attempt, v.GetPhase())
		}
		if got := v.GetTrickNumber(); got != VintHandSize {
			t.Fatalf("attempt %d: %d tricks played, want %d", attempt, got, VintHandSize)
		}
		// **13 トリックが 4 席に分配されている。**
		total := 0
		for i := range VintPlayerCnt {
			total += v.GetTricksWon(i)
		}
		if total != VintHandSize {
			t.Fatalf("attempt %d: %d tricks accounted for, want %d", attempt, total, VintHandSize)
		}
		// 両チームとも線下に得点している (どちらかが 0 トリックでない限り)。
		res := v.GetLastResult()
		if res == nil {
			t.Fatalf("attempt %d: the settlement produced no result", attempt)
		}
		if res.TrickPoints[0]+res.TrickPoints[1] != VintHandSize*res.TrickValue {
			t.Fatalf("attempt %d: the two sides scored %d+%d, want %d in total",
				attempt, res.TrickPoints[0], res.TrickPoints[1], VintHandSize*res.TrickValue)
		}
	}
}

func TestVintCpuEdges(t *testing.T) {
	v := vtPlaying(t, 0, 2, VintDenomHeart)
	v.SetHandForTest(0, []*Card{})
	if got := v.VintCpuPlay(0); got != -1 {
		t.Errorf("an empty hand has no play, got %d", got)
	}
	if got := v.VintCpuPlay(99); got != -1 {
		t.Errorf("an unknown seat has no play, got %d", got)
	}
	if level, _ := v.VintCpuBid(99); level != 0 {
		t.Errorf("an unknown seat bids %d, want 0", level)
	}
}

func TestVintDenomToSuit(t *testing.T) {
	for _, tc := range []struct {
		denom, want int
	}{
		{VintDenomSpade, CardDesignSpade},
		{VintDenomClub, CardDesignClover},
		{VintDenomDiamond, CardDesignDiamond},
		{VintDenomHeart, CardDesignHeart},
		// **ノートランプは 0。**切札スートが存在しない。
		{VintDenomNoTrump, 0},
		{99, 0},
	} {
		if got := VintDenomToSuit(tc.denom); got != tc.want {
			t.Errorf("denomination %d maps to %d, want %d", tc.denom, got, tc.want)
		}
	}
}

func TestVintAccessors(t *testing.T) {
	v := NewDefaultVint()
	v.Reset()
	if got := v.GetHandNumber(); got != 1 {
		t.Errorf("hand number = %d, want 1", got)
	}
	if got := v.GetWinnerTeam(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if got := v.GetDeclarerIdx(); got != -1 {
		t.Errorf("declarer = %d, want -1 before the auction settles", got)
	}
	if v.GetHighBid() != nil {
		t.Error("no bid stands at the start")
	}
	if v.GetLastResult() != nil {
		t.Error("there is no result before the first settlement")
	}
	if got := len(v.GetPlayers()); got != VintPlayerCnt {
		t.Errorf("%d seats, want %d", got, VintPlayerCnt)
	}
	if v.GetPlayer(-1) != nil || v.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
	}
	if v.GetTricksWon(-1) != 0 || v.GetBelow(99) != 0 || v.GetAbove(99) != 0 || v.GetGamesWon(99) != 0 {
		t.Error("out-of-range values must be 0, not a panic")
	}
	if got := len(v.GetTrick()); got != 0 {
		t.Errorf("the trick starts empty, got %d", got)
	}
	if len(v.GetActionLog()) == 0 {
		t.Error("dealing writes to the action log")
	}
	if got := len(v.GetBids()); got != 0 {
		t.Errorf("the bid history starts empty, got %d", got)
	}
	cfg := v.GetConfig()
	v.SetConfig(cfg)
	if v.GetConfig() != cfg {
		t.Error("SetConfig must take effect")
	}
}

func TestVintConfigValidate(t *testing.T) {
	if err := DefaultVintConfig().Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	if err := (VintConfig{CpuDifficulty: 9}).Validate(); err == nil {
		t.Error("a bad difficulty must not validate")
	}
}

func TestVintRoundTripsThroughJSON(t *testing.T) {
	v := NewDefaultVint()
	v.Reset()
	_ = v.Bid(1, 3, VintDenomHeart)

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Vint
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetHighBid() == nil || restored.GetHighBid().Level != 3 {
		t.Error("the standing bid did not survive the round trip")
	}
	if got := restored.GetPlayer(0).GetCardsSize(); got != VintHandSize {
		t.Errorf("the restored hand holds %d, want %d", got, VintHandSize)
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestVintRejectsBadJSON(t *testing.T) {
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
		{"bad trump suit", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":-1,"ts":9}`},
		{"oversized trick", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":-1,"tk":[{},{},{},{},{}]}`},
		{"bad high bid level", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":-1,"hb":{"Level":9}}`},
		{"bad high bid denomination", `{"pl":[{},{},{},{}],"ph":0,"di":0,"ci":0,"bi":0,"de":-1,"tl":-1,"wt":-1,"hb":{"Level":3,"Denom":9}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var v Vint
			if err := json.Unmarshal([]byte(tc.body), &v); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}

// TestVintAccessorsBeforeTheAuctionSettles は落札前に公開アクセサを呼んでも
// 落ちないことを確かめる。
//
// **declarerIdx は競りが決まるまで -1。**プレゼンターは宣言フェーズでも状態を
// 送るので、素で添字にする経路があると最初の応答で panic する。
func TestVintAccessorsBeforeTheAuctionSettles(t *testing.T) {
	v := NewDefaultVint()
	v.Reset()
	if got := v.GetDeclarerIdx(); got != -1 {
		t.Fatalf("declarer = %d, want -1 during the auction", got)
	}
	// **範囲外の席は -1 を返す。**Go の剰余は -1 % 2 = -1 なので素通しは危険。
	if got := VintTeamOf(-1); got != -1 {
		t.Errorf("VintTeamOf(-1) = %d, want -1", got)
	}
	if got := VintTeamOf(99); got != -1 {
		t.Errorf("VintTeamOf(99) = %d, want -1", got)
	}
	// 落札前でも全部のアクセサが落ちずに既定値を返す。
	for _, get := range []func() int{
		func() int { return v.VintTeamTricks(VintTeamOf(v.GetDeclarerIdx())) },
		func() int { return v.GetBelow(VintTeamOf(v.GetDeclarerIdx())) },
		func() int { return v.GetAbove(VintTeamOf(v.GetDeclarerIdx())) },
		func() int { return v.GetGamesWon(VintTeamOf(v.GetDeclarerIdx())) },
		func() int { return v.GetTricksWon(v.GetDeclarerIdx()) },
	} {
		if got := get(); got != 0 {
			t.Errorf("an accessor returned %d before the auction settles, want 0", got)
		}
	}
}

func TestVintPlayerTeam(t *testing.T) {
	p := NewVintPlayer(true)
	if p.GetTeam(0) != p.GetTeam(2) {
		t.Error("seats 0 and 2 are partners")
	}
	if p.GetTeam(0) == p.GetTeam(1) {
		t.Error("seats 0 and 1 are opponents")
	}
}
