//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"
)

// kjCard は 32 枚デッキの札を作る (A は値 1)。
func kjCard(suit, value int) *Card { return NewCard(suit, value, true) }

// kjReady は切札を決めた状態のプレイ局面を作る。
func kjReady(t *testing.T, trump int) *Klaberjass {
	t.Helper()
	k := NewDefaultKlaberjass()
	k.Reset()
	k.SetPhaseForTest(KlaberjassPhasePlay)
	k.SetTrumpForTest(trump)
	k.SetMakerForTest(0)
	k.SetCurrentPlayerForTest(0)
	k.SetTrickLeaderForTest(0)
	return k
}

// TestKlaberjassCardPoints pins the point ladder the issue only half states.
func TestKlaberjassCardPoints(t *testing.T) {
	const trump = CardDesignSpade
	for _, tc := range []struct {
		name string
		card *Card
		want int
	}{
		{"trump jack is the Jass", kjCard(trump, 11), 20},
		{"trump nine is the Menel", kjCard(trump, 9), 14},
		{"trump ace", kjCard(trump, 1), 11},
		{"trump ten", kjCard(trump, 10), 10},
		{"trump king", kjCard(trump, 13), 4},
		{"trump queen", kjCard(trump, 12), 3},
		{"trump eight", kjCard(trump, 8), 0},
		{"plain ace", kjCard(CardDesignHeart, 1), 11},
		{"plain ten", kjCard(CardDesignHeart, 10), 10},
		{"plain king", kjCard(CardDesignHeart, 13), 4},
		{"plain queen", kjCard(CardDesignHeart, 12), 3},
		// **平の J は 2 点。**0 ではない。
		{"plain jack", kjCard(CardDesignHeart, 11), 2},
		// **平の 9 は 0 点。**切札のときだけ 14 になる。
		{"plain nine", kjCard(CardDesignHeart, 9), 0},
		{"plain seven", kjCard(CardDesignHeart, 7), 0},
		{"nil", nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := KlaberjassCardPoints(tc.card, trump); got != tc.want {
				t.Errorf("points = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestKlaberjassPackTotalIsOneSixtyTwo は issue の «163点» が誤りであることを
// 算術で示す。
//
// **カード 152 点 + 最終トリック 10 点 = 162 点。**163 にはならない。
func TestKlaberjassPackTotalIsOneSixtyTwo(t *testing.T) {
	const trump = CardDesignSpade
	total := 0
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for _, v := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
			total += KlaberjassCardPoints(kjCard(suit, v), trump)
		}
	}
	if total != KlaberjassCardPointsTotal {
		t.Fatalf("the pack holds %d card points, want %d", total, KlaberjassCardPointsTotal)
	}
	if got := total + KlaberjassLastTrickBonus; got != 162 {
		t.Errorf("total with the last trick = %d, want 162 (the issue says 163)", got)
	}
	// 切札スートだけで 62 点、平のスートは各 30 点。
	trumpOnly := 0
	for _, v := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
		trumpOnly += KlaberjassCardPoints(kjCard(trump, v), trump)
	}
	if trumpOnly != 62 {
		t.Errorf("the trump suit holds %d, want 62", trumpOnly)
	}
	if (total-trumpOnly)%3 != 0 || (total-trumpOnly)/3 != 30 {
		t.Errorf("each plain suit should hold 30, got %d across three", total-trumpOnly)
	}
}

// TestKlaberjassNotEveryCardIsDealt は、2 人戦では固定の総点を争わないことを
// 示す。issue の «163点を両者で争う» が成り立たない理由。
func TestKlaberjassNotEveryCardIsDealt(t *testing.T) {
	k := NewDefaultKlaberjass()
	k.Reset()
	_ = k.AcceptTrump(k.GetBidPlayerIdx())

	dealt := 0
	for i := range KlaberjassPlayerCnt {
		got := k.GetPlayer(i).GetCardsSize()
		if got != KlaberjassHandSize {
			t.Fatalf("seat %d holds %d cards, want %d", i, got, KlaberjassHandSize)
		}
		dealt += got
	}
	if dealt != 18 {
		t.Fatalf("%d cards are dealt, want 18", dealt)
	}
	// 32 枚のうち 14 枚は表向きカードを含めて場に出ない。
	if dealt >= 32 {
		t.Error("a two-player deal must leave most of the pack out of play")
	}
}

// TestKlaberjassTrickRankPutsJassAndMenelOnTop covers the ordering that makes
// this family distinctive.
func TestKlaberjassTrickRankPutsJassAndMenelOnTop(t *testing.T) {
	const trump = CardDesignSpade
	descending := []int{11, 9, 1, 10, 13, 12, 8, 7}
	for i := 1; i < len(descending); i++ {
		hi := klaberjassTrickRank(kjCard(trump, descending[i-1]), trump)
		lo := klaberjassTrickRank(kjCard(trump, descending[i]), trump)
		if hi <= lo {
			t.Errorf("trump %d must beat trump %d", descending[i-1], descending[i])
		}
	}
	// 平のスートでは通常の並び。J は Q より下。
	plain := []int{1, 10, 13, 12, 11, 9, 8, 7}
	for i := 1; i < len(plain); i++ {
		hi := klaberjassTrickRank(kjCard(CardDesignHeart, plain[i-1]), trump)
		lo := klaberjassTrickRank(kjCard(CardDesignHeart, plain[i]), trump)
		if hi <= lo {
			t.Errorf("plain %d must beat plain %d", plain[i-1], plain[i])
		}
	}
	if klaberjassTrickRank(nil, trump) != 0 {
		t.Error("a nil card has no rank")
	}
}

// TestKlaberjassSequencesUseTheNaturalOrder は実装上いちばん間違えやすい点を
// 押さえる。
//
// **並びは点数順ではない。**切札の J が 20 点でも、J は 10 と Q の間にある。
func TestKlaberjassSequencesUseTheNaturalOrder(t *testing.T) {
	const trump = CardDesignSpade
	// 切札の 10-J-Q は連続。点数順 (J=20, 10=10, Q=3) ではばらばらである。
	seqs := klaberjassFindSequences([]*Card{
		kjCard(trump, 10), kjCard(trump, 11), kjCard(trump, 12),
	})
	if len(seqs) != 1 {
		t.Fatalf("found %d sequences, want 1", len(seqs))
	}
	if seqs[0].Length != 3 || seqs[0].Points != KlaberjassTerzPoints {
		t.Errorf("got length %d / %d points, want 3 / %d", seqs[0].Length, seqs[0].Points, KlaberjassTerzPoints)
	}

	// **A は最上位。**値は 1 だが 7-8-9 の下ではなく K の上に付く。
	high := klaberjassFindSequences([]*Card{
		kjCard(CardDesignHeart, 12), kjCard(CardDesignHeart, 13), kjCard(CardDesignHeart, 1),
	})
	if len(high) != 1 {
		t.Fatalf("Q-K-A must be a sequence, found %d", len(high))
	}
	if high[0].TopValue != 14 {
		t.Errorf("the ace tops the run at %d, want 14", high[0].TopValue)
	}
	// A-7-8 は並びではない。
	if got := klaberjassFindSequences([]*Card{
		kjCard(CardDesignHeart, 1), kjCard(CardDesignHeart, 7), kjCard(CardDesignHeart, 8),
	}); len(got) != 0 {
		t.Errorf("the ace must not wrap round to the seven, found %d sequences", len(got))
	}
}

func TestKlaberjassSequenceScoring(t *testing.T) {
	const s = CardDesignClover
	four := klaberjassFindSequences([]*Card{
		kjCard(s, 7), kjCard(s, 8), kjCard(s, 9), kjCard(s, 10),
	})
	if len(four) != 1 || four[0].Points != KlaberjassFiftyPoints {
		t.Fatalf("a four-card run scores %v, want %d", four, KlaberjassFiftyPoints)
	}
	// **5 枚でも 50 点のまま。**長さに応じて増えない。
	five := klaberjassFindSequences([]*Card{
		kjCard(s, 7), kjCard(s, 8), kjCard(s, 9), kjCard(s, 10), kjCard(s, 11),
	})
	if len(five) != 1 || five[0].Points != KlaberjassFiftyPoints {
		t.Errorf("a five-card run scores %v, want %d", five, KlaberjassFiftyPoints)
	}
	// 2 枚は役にならない。
	if got := klaberjassFindSequences([]*Card{kjCard(s, 7), kjCard(s, 8)}); len(got) != 0 {
		t.Errorf("two cards are not a sequence, found %d", len(got))
	}
	// スートを跨いだ並びは役にならない。
	if got := klaberjassFindSequences([]*Card{
		kjCard(s, 7), kjCard(CardDesignHeart, 8), kjCard(s, 9),
	}); len(got) != 0 {
		t.Errorf("a run must stay in one suit, found %d", len(got))
	}
	// 同じ手札に 2 つの並びがあれば両方拾う。
	two := klaberjassFindSequences([]*Card{
		kjCard(s, 7), kjCard(s, 8), kjCard(s, 9),
		kjCard(CardDesignHeart, 11), kjCard(CardDesignHeart, 12), kjCard(CardDesignHeart, 13),
	})
	if len(two) != 2 {
		t.Errorf("found %d sequences, want 2", len(two))
	}
	if klaberjassSequencePoints(2) != 0 {
		t.Error("a two-card run is worth nothing")
	}
	if klaberjassBestSequence(nil) != nil {
		t.Error("no sequences means no best sequence")
	}
}

// TestKlaberjassOnlyTheBetterSequenceScoresAndScoresAllOfThem covers the rule
// the issue states only half of.
func TestKlaberjassOnlyTheBetterSequenceScoresAndScoresAllOfThem(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	// 席 0: ♠7-8-9-10 (50) と ♥J-Q-K (20)。
	k.SetHandForTest(0, []*Card{
		kjCard(CardDesignSpade, 7), kjCard(CardDesignSpade, 8),
		kjCard(CardDesignSpade, 9), kjCard(CardDesignSpade, 10),
		kjCard(CardDesignHeart, 11), kjCard(CardDesignHeart, 12), kjCard(CardDesignHeart, 13),
	})
	// 席 1: ♦J-Q-K (20) だけ。
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignDiamond, 11), kjCard(CardDesignDiamond, 12), kjCard(CardDesignDiamond, 13),
	})
	k.SetHandPointsForTest(0, 0)
	k.SetHandPointsForTest(1, 0)
	k.CollectSequencesForTest()

	if got := k.GetSequenceWinner(); got != 0 {
		t.Fatalf("sequence winner = %d, want 0", got)
	}
	// **勝った側は自分の役を全部数える。**50 + 20 = 70。
	if got := k.GetHandPoints(0); got != 70 {
		t.Errorf("seat 0 scores %d, want 70 (all of its sequences, not just the best)", got)
	}
	// **負けた側は 1 点も入らない。**
	if got := k.GetHandPoints(1); got != 0 {
		t.Errorf("seat 1 scores %d, want 0", got)
	}
}

func TestKlaberjassSequenceTieScoresNobody(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	k.SetHandForTest(0, []*Card{
		kjCard(CardDesignHeart, 11), kjCard(CardDesignHeart, 12), kjCard(CardDesignHeart, 13),
	})
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignDiamond, 11), kjCard(CardDesignDiamond, 12), kjCard(CardDesignDiamond, 13),
	})
	k.SetHandPointsForTest(0, 0)
	k.SetHandPointsForTest(1, 0)
	k.CollectSequencesForTest()

	if got := k.GetSequenceWinner(); got != -1 {
		t.Errorf("sequence winner = %d, want -1 on an exact tie", got)
	}
	if k.GetHandPoints(0) != 0 || k.GetHandPoints(1) != 0 {
		t.Errorf("nobody scores an exact tie, got %d / %d", k.GetHandPoints(0), k.GetHandPoints(1))
	}
}

func TestKlaberjassLongerSequenceBeatsHigherOne(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	// 席 0 は低い 4 枚、席 1 は高い 3 枚。**長い方が勝つ。**
	k.SetHandForTest(0, []*Card{
		kjCard(CardDesignHeart, 7), kjCard(CardDesignHeart, 8),
		kjCard(CardDesignHeart, 9), kjCard(CardDesignHeart, 10),
	})
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignDiamond, 12), kjCard(CardDesignDiamond, 13), kjCard(CardDesignDiamond, 1),
	})
	k.SetHandPointsForTest(0, 0)
	k.SetHandPointsForTest(1, 0)
	k.CollectSequencesForTest()
	if got := k.GetSequenceWinner(); got != 0 {
		t.Errorf("sequence winner = %d, want 0 (four beats three however high)", got)
	}
}

// TestKlaberjassBelaNeedNotBePlayedConsecutively は issue の «連続して出すと»
// が誤りであることを示す。
func TestKlaberjassBelaNeedNotBePlayedConsecutively(t *testing.T) {
	const trump = CardDesignSpade
	k := kjReady(t, trump)
	k.SetHandForTest(0, []*Card{
		kjCard(trump, 13), kjCard(CardDesignHeart, 7), kjCard(trump, 12),
	})
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignHeart, 8), kjCard(CardDesignHeart, 9), kjCard(CardDesignHeart, 10),
	})
	k.FindBelaForTest()
	if got := k.GetBelaHolder(); got != 0 {
		t.Fatalf("bela holder = %d, want 0", got)
	}
	before := k.GetHandPoints(0)

	// 切札 K を出す。
	if err := k.PlayCard(0, 0); err != nil {
		t.Fatalf("play K: %v", err)
	}
	if k.IsBelaScored() {
		t.Fatal("one half of the bela is not the bela")
	}
	k.SetCurrentPlayerForTest(1)
	if err := k.PlayCard(1, 0); err != nil {
		t.Fatalf("opponent play: %v", err)
	}

	// **間に別の札を挟んでから** Q を出す。
	k.SetPhaseForTest(KlaberjassPhasePlay)
	k.SetCurrentPlayerForTest(0)
	k.SetTrickLeaderForTest(0)
	k.trick = nil
	if err := k.PlayCard(0, 0); err != nil { // ♥7
		t.Fatalf("play the filler: %v", err)
	}
	k.SetCurrentPlayerForTest(1)
	if err := k.PlayCard(1, 0); err != nil {
		t.Fatalf("opponent play: %v", err)
	}
	k.SetPhaseForTest(KlaberjassPhasePlay)
	k.SetCurrentPlayerForTest(0)
	k.SetTrickLeaderForTest(0)
	k.trick = nil
	if err := k.PlayCard(0, 0); err != nil { // 切札 Q
		t.Fatalf("play Q: %v", err)
	}

	if !k.IsBelaScored() {
		t.Fatal("the bela scores once both halves are played, consecutive or not")
	}
	if got := k.GetHandPoints(0) - before; got < KlaberjassBelaBonus {
		t.Errorf("the bela added %d, want at least %d", got, KlaberjassBelaBonus)
	}
}

func TestKlaberjassBelaNeedsBothHalvesInOneHand(t *testing.T) {
	const trump = CardDesignSpade
	k := kjReady(t, trump)
	k.SetHandForTest(0, []*Card{kjCard(trump, 13)})
	k.SetHandForTest(1, []*Card{kjCard(trump, 12)})
	k.FindBelaForTest()
	if got := k.GetBelaHolder(); got != -1 {
		t.Errorf("bela holder = %d, want -1: the halves are split between the seats", got)
	}
}

// TestKlaberjassFollowingAndTrumpingAreCompulsory covers the play restrictions.
func TestKlaberjassFollowingAndTrumpingAreCompulsory(t *testing.T) {
	const trump = CardDesignSpade
	k := kjReady(t, trump)
	k.SetHandForTest(0, []*Card{kjCard(CardDesignHeart, 7)})
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignHeart, 8), kjCard(trump, 7), kjCard(CardDesignDiamond, 1),
	})
	if err := k.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	// 追随できるなら追随のみ。
	valid := k.KlaberjassValidPlays(1)
	if len(valid) != 1 || valid[0] != 0 {
		t.Errorf("valid = %v, want only the heart", valid)
	}

	// フォローできなければ切札を出さねばならない。
	k2 := kjReady(t, trump)
	k2.SetHandForTest(0, []*Card{kjCard(CardDesignHeart, 7)})
	k2.SetHandForTest(1, []*Card{kjCard(trump, 7), kjCard(CardDesignDiamond, 1)})
	if err := k2.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	valid2 := k2.KlaberjassValidPlays(1)
	if len(valid2) != 1 || valid2[0] != 0 {
		t.Errorf("valid = %v, want the forced trump", valid2)
	}

	// 切札も無ければ何でも出せる。
	k3 := kjReady(t, trump)
	k3.SetHandForTest(0, []*Card{kjCard(CardDesignHeart, 7)})
	k3.SetHandForTest(1, []*Card{kjCard(CardDesignDiamond, 1), kjCard(CardDesignClover, 8)})
	if err := k3.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := len(k3.KlaberjassValidPlays(1)); got != 2 {
		t.Errorf("with neither the suit nor a trump, all %d cards are legal, got %d", 2, got)
	}
}

// TestKlaberjassMustOvertrumpOnATrumpLead covers the third restriction.
func TestKlaberjassMustOvertrumpOnATrumpLead(t *testing.T) {
	const trump = CardDesignSpade
	k := kjReady(t, trump)
	// 切札 Q (rank 3) がリード。9 (Menel, rank 7) で勝てるので 7 は選べない。
	k.SetHandForTest(0, []*Card{kjCard(trump, 12)})
	k.SetHandForTest(1, []*Card{kjCard(trump, 7), kjCard(trump, 9)})
	if err := k.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	valid := k.KlaberjassValidPlays(1)
	if len(valid) != 1 || valid[0] != 1 {
		t.Errorf("valid = %v, want only the winning trump", valid)
	}

	// 勝てる切札が無ければ、切札のどれでもよい。
	k2 := kjReady(t, trump)
	k2.SetHandForTest(0, []*Card{kjCard(trump, 11)}) // Jass、最強
	k2.SetHandForTest(1, []*Card{kjCard(trump, 7), kjCard(trump, 8)})
	if err := k2.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := len(k2.KlaberjassValidPlays(1)); got != 2 {
		t.Errorf("with no winning trump both are legal, got %d", got)
	}
}

func TestKlaberjassRejectsAnIllegalPlay(t *testing.T) {
	const trump = CardDesignSpade
	k := kjReady(t, trump)
	k.SetHandForTest(0, []*Card{kjCard(CardDesignHeart, 7)})
	k.SetHandForTest(1, []*Card{kjCard(CardDesignHeart, 8), kjCard(CardDesignDiamond, 1)})
	if err := k.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if err := k.PlayCard(1, 1); err == nil {
		t.Error("discarding while able to follow must be refused")
	}
	if err := k.PlayCard(0, 0); err == nil {
		t.Error("playing out of turn must be refused")
	}
	if err := k.PlayCard(1, 99); err == nil {
		t.Error("an out-of-range index must be refused")
	}
}

// TestKlaberjassTrumpWinsAndTakesThePoints covers trick resolution.
func TestKlaberjassTrumpWinsAndTakesThePoints(t *testing.T) {
	const trump = CardDesignSpade
	k := kjReady(t, trump)
	k.SetHandForTest(0, []*Card{kjCard(CardDesignHeart, 1)}) // 11 点
	k.SetHandForTest(1, []*Card{kjCard(trump, 7)})           // 0 点だが切札
	k.SetHandPointsForTest(0, 0)
	k.SetHandPointsForTest(1, 0)
	if err := k.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if err := k.PlayCard(1, 0); err != nil {
		t.Fatalf("trump it: %v", err)
	}
	// 最後の札だったので精算まで走る。切札を出した側が 11 点 + 最終トリック 10 点。
	if got := k.GetHandPoints(1); got != 11+KlaberjassLastTrickBonus {
		t.Errorf("seat 1 took %d, want %d", got, 11+KlaberjassLastTrickBonus)
	}
	if got := k.GetHandPoints(0); got != 0 {
		t.Errorf("seat 0 took %d, want 0", got)
	}
}

// TestKlaberjassBeteGivesEverythingToTheOpponent covers the settlement.
func TestKlaberjassBeteGivesEverythingToTheOpponent(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	k.SetMakerForTest(0)
	k.SetHandPointsForTest(0, 40)
	k.SetHandPointsForTest(1, 90)
	k.lastTrickWinner = -1
	k.FinishHandForTest()

	if !k.IsBete() {
		t.Fatal("the maker scored less; that is bete")
	}
	if got := k.GetScore(0); got != 0 {
		t.Errorf("the bete maker scores %d, want 0", got)
	}
	// **相手は両者の合計を取る。**自分の分だけではない。
	if got := k.GetScore(1); got != 130 {
		t.Errorf("the opponent scores %d, want 130 (both totals)", got)
	}
}

// TestKlaberjassTieIsBete は「多く取れなければ」の境界を押さえる。
//
// **同点はメイカーの負け。**
func TestKlaberjassTieIsBete(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	k.SetMakerForTest(0)
	k.SetHandPointsForTest(0, 60)
	k.SetHandPointsForTest(1, 60)
	k.lastTrickWinner = -1
	k.FinishHandForTest()

	if !k.IsBete() {
		t.Fatal("an exact tie is bete: the maker must score MORE")
	}
	if got := k.GetScore(1); got != 120 {
		t.Errorf("the opponent scores %d, want 120", got)
	}
}

func TestKlaberjassMakerSucceedsAndBothScore(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	k.SetMakerForTest(0)
	k.SetHandPointsForTest(0, 90)
	k.SetHandPointsForTest(1, 40)
	k.lastTrickWinner = -1
	k.FinishHandForTest()

	if k.IsBete() {
		t.Fatal("the maker scored more; that is not bete")
	}
	if got := k.GetScore(0); got != 90 {
		t.Errorf("the maker scores %d, want 90", got)
	}
	// **成功したときは相手も自分の分を取る。**取り上げるのはベートのときだけ。
	if got := k.GetScore(1); got != 40 {
		t.Errorf("the defender scores %d, want 40", got)
	}
}

func TestKlaberjassLastTrickBonusGoesToItsWinner(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	k.SetMakerForTest(0)
	k.SetHandPointsForTest(0, 50)
	k.SetHandPointsForTest(1, 45)
	k.lastTrickWinner = 1
	k.FinishHandForTest()

	// 45 + 10 = 55 > 50 なのでメイカーはベート。
	if !k.IsBete() {
		t.Errorf("the last trick flipped the hand; the maker should be bete")
	}
	if got := k.GetScore(1); got != 105 {
		t.Errorf("the opponent scores %d, want 105", got)
	}
}

// TestKlaberjassBidding covers both rounds and the redeal.
func TestKlaberjassBidding(t *testing.T) {
	t.Run("the non-dealer bids first", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		if got := k.GetBidPlayerIdx(); got != 1 {
			t.Errorf("bid seat = %d, want 1 (the dealer is seat 0)", got)
		}
		if got := k.GetPhase(); got != KlaberjassPhaseBidTurnUp {
			t.Errorf("phase = %v, want the turn-up round", got)
		}
	})

	t.Run("accepting fixes the turn-up suit and deals up to nine", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		want := k.GetTurnUpCard().GetDesign()
		bidder := k.GetBidPlayerIdx()
		if err := k.AcceptTrump(bidder); err != nil {
			t.Fatalf("AcceptTrump: %v", err)
		}
		if got := k.GetTrumpSuit(); got != want {
			t.Errorf("trump = %d, want %d", got, want)
		}
		if got := k.GetMakerIdx(); got != bidder {
			t.Errorf("maker = %d, want %d", got, bidder)
		}
		if got := k.GetPhase(); got != KlaberjassPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
		// **リードは非ディーラー。**メイカーではない。
		if got := k.GetTrickLeaderIdx(); got != 1 {
			t.Errorf("leader = %d, want the non-dealer", got)
		}
	})

	t.Run("two passes open the free round", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		if err := k.Pass(1); err != nil {
			t.Fatalf("pass: %v", err)
		}
		if got := k.GetPhase(); got != KlaberjassPhaseBidTurnUp {
			t.Fatalf("one pass must not end the round, phase = %v", got)
		}
		if err := k.Pass(0); err != nil {
			t.Fatalf("pass: %v", err)
		}
		if got := k.GetPhase(); got != KlaberjassPhaseBidFree {
			t.Errorf("phase = %v, want the free round", got)
		}
		if got := k.GetBidPlayerIdx(); got != 1 {
			t.Errorf("the free round restarts with the non-dealer, got %d", got)
		}
	})

	t.Run("the refused turn-up suit may not be called", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		refused := k.GetTurnUpCard().GetDesign()
		_ = k.Pass(1)
		_ = k.Pass(0)
		if err := k.CallTrump(1, refused); err == nil {
			t.Error("the suit both players refused must not come back")
		}
		other := refused%4 + 1
		if err := k.CallTrump(1, other); err != nil {
			t.Errorf("CallTrump: %v", err)
		}
	})

	t.Run("four passes redeal without advancing the deal number", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		deal := k.GetDealNumber()
		_ = k.Pass(1)
		_ = k.Pass(0)
		_ = k.Pass(1)
		_ = k.Pass(0)
		if got := k.GetPhase(); got != KlaberjassPhaseBidTurnUp {
			t.Errorf("phase = %v, want a fresh turn-up round", got)
		}
		if got := k.GetDealNumber(); got != deal {
			t.Errorf("deal number = %d, want %d: an abandoned deal does not count", got, deal)
		}
		for i := range KlaberjassPlayerCnt {
			if got := k.GetPlayer(i).GetCardsSize(); got != KlaberjassFirstDealSize {
				t.Errorf("seat %d holds %d after the redeal, want %d", i, got, KlaberjassFirstDealSize)
			}
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		if err := k.AcceptTrump(0); err == nil {
			t.Error("bidding out of turn must be refused")
		}
		if err := k.CallTrump(1, CardDesignSpade); err == nil {
			t.Error("calling before the free round must be refused")
		}
		_ = k.Pass(1)
		_ = k.Pass(0)
		if err := k.CallTrump(1, 99); err == nil {
			t.Error("a bad suit must be refused")
		}
		if err := k.AcceptTrump(1); err == nil {
			t.Error("the turn-up is no longer on offer in the free round")
		}
	})
}

// TestKlaberjassDixExchangesTheTrumpSeven covers the rule the issue omits.
func TestKlaberjassDixExchangesTheTrumpSeven(t *testing.T) {
	k := NewDefaultKlaberjass()
	k.Reset()
	turnUp := kjCard(CardDesignSpade, 1)
	k.SetTurnUpForTest(turnUp)
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignSpade, 7), kjCard(CardDesignHeart, 8),
		kjCard(CardDesignHeart, 9), kjCard(CardDesignHeart, 10),
		kjCard(CardDesignHeart, 11), kjCard(CardDesignHeart, 12),
	})
	if err := k.AcceptTrump(1); err != nil {
		t.Fatalf("AcceptTrump: %v", err)
	}
	if !k.IsDixUsed() {
		t.Fatal("the holder of the trump seven exchanges it for the turn-up")
	}
	found := false
	p := k.GetPlayer(1)
	for i := range p.GetCardsSize() {
		if c := p.GetCard(i); c != nil && c.GetDesign() == CardDesignSpade && c.GetValue() == 1 {
			found = true
		}
	}
	if !found {
		t.Error("the seven's holder must now hold the turn-up card")
	}
	if got := k.GetTurnUpCard(); got == nil || got.GetValue() != 7 {
		t.Error("the seven takes the turn-up's place")
	}
}

// TestKlaberjassDixOnlyWhenTheTurnUpSuitIsTrump は交換できない場合を押さえる。
func TestKlaberjassDixOnlyWhenTheTurnUpSuitIsTrump(t *testing.T) {
	k := NewDefaultKlaberjass()
	k.Reset()
	k.SetTurnUpForTest(kjCard(CardDesignSpade, 1))
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignHeart, 7), kjCard(CardDesignHeart, 8),
		kjCard(CardDesignHeart, 9), kjCard(CardDesignHeart, 10),
		kjCard(CardDesignHeart, 11), kjCard(CardDesignHeart, 12),
	})
	_ = k.Pass(1)
	_ = k.Pass(0)
	if err := k.CallTrump(1, CardDesignHeart); err != nil {
		t.Fatalf("CallTrump: %v", err)
	}
	if k.IsDixUsed() {
		t.Error("no exchange is allowed once a different suit becomes trump")
	}
}

// TestKlaberjassSchmeiss covers the option the issue omits.
func TestKlaberjassSchmeiss(t *testing.T) {
	t.Run("accepting throws the deal in", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		deal := k.GetDealNumber()
		if err := k.Schmeiss(1); err != nil {
			t.Fatalf("Schmeiss: %v", err)
		}
		if got := k.GetPhase(); got != KlaberjassPhaseSchmeiss {
			t.Fatalf("phase = %v, want the schmeiss answer", got)
		}
		if err := k.AnswerSchmeiss(0, true); err != nil {
			t.Fatalf("AnswerSchmeiss: %v", err)
		}
		if got := k.GetPhase(); got != KlaberjassPhaseBidTurnUp {
			t.Errorf("phase = %v, want a fresh deal", got)
		}
		if got := k.GetDealNumber(); got != deal {
			t.Errorf("deal number = %d, want %d", got, deal)
		}
	})

	// **拒めば提案者がメイカーにされる。**投げ得にはならない。
	t.Run("refusing makes the thrower the maker", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		want := k.GetTurnUpCard().GetDesign()
		if err := k.Schmeiss(1); err != nil {
			t.Fatalf("Schmeiss: %v", err)
		}
		if err := k.AnswerSchmeiss(0, false); err != nil {
			t.Fatalf("AnswerSchmeiss: %v", err)
		}
		if got := k.GetMakerIdx(); got != 1 {
			t.Errorf("maker = %d, want the thrower (1)", got)
		}
		if got := k.GetTrumpSuit(); got != want {
			t.Errorf("trump = %d, want the turn-up suit %d", got, want)
		}
		if got := k.GetPhase(); got != KlaberjassPhasePlay {
			t.Errorf("phase = %v, want play", got)
		}
	})

	t.Run("bad input is refused", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		k.Reset()
		if err := k.AnswerSchmeiss(0, true); err == nil {
			t.Error("answering with no schmeiss pending must be refused")
		}
		if err := k.Schmeiss(0); err == nil {
			t.Error("throwing out of turn must be refused")
		}
		if err := k.Schmeiss(1); err != nil {
			t.Fatalf("Schmeiss: %v", err)
		}
		if err := k.AnswerSchmeiss(1, true); err == nil {
			t.Error("the thrower does not answer its own throw")
		}
	})

	t.Run("disabled by config", func(t *testing.T) {
		k := NewDefaultKlaberjass()
		cfg := k.GetConfig()
		cfg.AllowSchmeiss = false
		k.SetConfig(cfg)
		k.Reset()
		if err := k.Schmeiss(1); err == nil {
			t.Error("schmeiss must be refused when the config disables it")
		}
	})
}

func TestKlaberjassNextDealAlternatesTheDealer(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	k.SetMakerForTest(0)
	k.SetHandPointsForTest(0, 90)
	k.SetHandPointsForTest(1, 40)
	k.lastTrickWinner = -1
	k.FinishHandForTest()

	dealer := k.GetDealerIdx()
	if err := k.NextDeal(); err != nil {
		t.Fatalf("NextDeal: %v", err)
	}
	if got := k.GetDealerIdx(); got == dealer {
		t.Errorf("dealer stayed at %d; it must alternate", got)
	}
	for i := range KlaberjassPlayerCnt {
		if got := k.GetPlayer(i).GetCardsSize(); got != KlaberjassFirstDealSize {
			t.Errorf("seat %d holds %d, want %d", i, got, KlaberjassFirstDealSize)
		}
	}
	// **山札を戻さないと 2 ディール目で札が尽きる。**
	if k.GetTurnUpCard() == nil {
		t.Error("the second deal must still produce a turn-up card")
	}
}

func TestKlaberjassNextDealGuards(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	if err := k.NextDeal(); err == nil {
		t.Error("dealing again mid-hand must be refused")
	}
}

func TestKlaberjassGameEnd(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	k.SetMakerForTest(0)
	k.SetScoreForTest(0, KlaberjassTargetScoreDefault-10)
	k.SetHandPointsForTest(0, 90)
	k.SetHandPointsForTest(1, 40)
	k.lastTrickWinner = -1
	k.FinishHandForTest()

	if !k.GetGameEndFlag() {
		t.Fatal("passing the target ends the game")
	}
	if got := k.GetWinnerIdx(); got != 0 {
		t.Errorf("winner = %d, want 0", got)
	}
	if got := k.GetPhase(); got != KlaberjassPhaseGameEnd {
		t.Errorf("phase = %v, want game end", got)
	}
	if err := k.NextDeal(); err == nil {
		t.Error("dealing after the game is over must be refused")
	}
}

// 両者が同時に目標を超えたら点の多い方、同点ならメイカーの勝ち。
func TestKlaberjassGameEndTieBreak(t *testing.T) {
	k := kjReady(t, CardDesignSpade)
	k.SetMakerForTest(1)
	k.SetScoreForTest(0, KlaberjassTargetScoreDefault)
	k.SetScoreForTest(1, KlaberjassTargetScoreDefault)
	k.checkGameEnd()
	if got := k.GetWinnerIdx(); got != 1 {
		t.Errorf("winner = %d, want the maker (1) on an exact tie", got)
	}
}

func TestKlaberjassIsHumanTurnAndCpuPlay(t *testing.T) {
	k := NewDefaultKlaberjass()
	k.Reset()
	// 席 1 は CPU で、ビッドの先手。
	if k.IsHumanTurn() {
		t.Error("the non-dealer bids first and it is the CPU")
	}
	k.CpuPlay()
	if k.GetPhase() == KlaberjassPhaseBidTurnUp && k.GetBidPlayerIdx() == 1 {
		t.Error("CpuPlay must move the bid along")
	}

	// ゲームが終わっていれば誰の手番でもない。
	k2 := NewDefaultKlaberjass()
	k2.Reset()
	k2.SetPhaseForTest(KlaberjassPhaseGameEnd)
	k2.gameEndFlag = true
	if k2.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	k2.CpuPlay() // 何も起きない
}

// **CPU だけで最後まで走り切れること。**途中で手番が止まると詰む。
func TestKlaberjassCpuDrivesAFullDeal(t *testing.T) {
	for attempt := range 30 {
		k := NewDefaultKlaberjass()
		k.Reset()
		// 人間席も CPU 扱いにして全自動で回す。
		k.GetPlayer(0).SetIsHuman(false)
		for step := 0; step < 400; step++ {
			if k.GetPhase() == KlaberjassPhaseHandEnd || k.GetGameEndFlag() {
				break
			}
			k.CpuPlay()
		}
		if k.GetPhase() != KlaberjassPhaseHandEnd && !k.GetGameEndFlag() {
			t.Fatalf("attempt %d: the deal never finished (phase %v)", attempt, k.GetPhase())
		}
		total := k.GetHandPoints(0) + k.GetHandPoints(1)
		if total <= 0 {
			t.Fatalf("attempt %d: a finished deal must have scored something", attempt)
		}
	}
}

func TestKlaberjassCpuBid(t *testing.T) {
	k := NewDefaultKlaberjass()
	k.Reset()
	k.SetTurnUpForTest(kjCard(CardDesignSpade, 1))
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignSpade, 7), kjCard(CardDesignSpade, 8), kjCard(CardDesignSpade, 9),
		kjCard(CardDesignHeart, 10), kjCard(CardDesignHeart, 11), kjCard(CardDesignHeart, 12),
	})
	if action, _ := k.KlaberjassCpuBid(1); action != "accept" {
		t.Errorf("action = %q, want accept with three of the turn-up suit", action)
	}

	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignSpade, 7), kjCard(CardDesignHeart, 8), kjCard(CardDesignClover, 9),
		kjCard(CardDesignDiamond, 10), kjCard(CardDesignHeart, 11), kjCard(CardDesignClover, 12),
	})
	if action, _ := k.KlaberjassCpuBid(1); action != "pass" {
		t.Errorf("action = %q, want pass with a scattered hand", action)
	}

	// 第 2 ラウンドでは一番長いスートを指名する。
	k.SetPhaseForTest(KlaberjassPhaseBidFree)
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignHeart, 7), kjCard(CardDesignHeart, 8), kjCard(CardDesignHeart, 9),
		kjCard(CardDesignClover, 10), kjCard(CardDesignClover, 11), kjCard(CardDesignDiamond, 12),
	})
	action, suit := k.KlaberjassCpuBid(1)
	if action != "call" || suit != CardDesignHeart {
		t.Errorf("action = %q suit = %d, want call hearts", action, suit)
	}

	if action, _ := k.KlaberjassCpuBid(99); action != "pass" {
		t.Errorf("an unknown seat passes, got %q", action)
	}
}

func TestKlaberjassCpuPlayPrefersToWin(t *testing.T) {
	const trump = CardDesignSpade
	k := kjReady(t, trump)
	k.SetHandForTest(0, []*Card{kjCard(CardDesignHeart, 13)}) // ♥K, 4 点
	k.SetHandForTest(1, []*Card{
		kjCard(CardDesignHeart, 7), kjCard(CardDesignHeart, 1), // ♥A で勝てる
	})
	if err := k.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := k.KlaberjassCpuPlay(1); got != 1 {
		t.Errorf("cpu played index %d, want the winning ace", got)
	}

	// 勝てないなら一番安い札。
	k2 := kjReady(t, trump)
	k2.SetHandForTest(0, []*Card{kjCard(CardDesignHeart, 1)}) // ♥A
	k2.SetHandForTest(1, []*Card{
		kjCard(CardDesignHeart, 13), kjCard(CardDesignHeart, 7), // K(4) と 7(0)
	})
	if err := k2.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := k2.KlaberjassCpuPlay(1); got != 1 {
		t.Errorf("cpu played index %d, want the cheapest card", got)
	}

	if got := k2.KlaberjassCpuPlay(99); got != -1 {
		t.Errorf("an unknown seat has no play, got %d", got)
	}
}

func TestKlaberjassAccessors(t *testing.T) {
	k := NewDefaultKlaberjass()
	k.Reset()
	if got := k.GetDealNumber(); got != 1 {
		t.Errorf("deal number = %d, want 1", got)
	}
	if got := k.GetWinnerIdx(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if got := k.GetMakerIdx(); got != -1 {
		t.Errorf("maker = %d, want -1 before anyone bids", got)
	}
	if got := k.GetSchmeissBy(); got != -1 {
		t.Errorf("schmeiss = %d, want -1", got)
	}
	if got := len(k.GetPlayers()); got != KlaberjassPlayerCnt {
		t.Errorf("%d seats, want %d", got, KlaberjassPlayerCnt)
	}
	if k.GetPlayer(-1) != nil || k.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
	}
	if k.GetHandPoints(-1) != 0 || k.GetScore(99) != 0 {
		t.Error("out-of-range scores must be 0, not a panic")
	}
	if k.GetSequences(99) != nil {
		t.Error("out-of-range sequences must be nil")
	}
	if got := len(k.GetTrick()); got != 0 {
		t.Errorf("the trick starts empty, got %d", got)
	}
	if got := k.GetTrickNumber(); got != 0 {
		t.Errorf("trick number = %d, want 0", got)
	}
	if len(k.GetActionLog()) == 0 {
		t.Error("dealing writes to the action log")
	}
	if got := k.GetConfig().TargetScore; got != KlaberjassTargetScoreDefault {
		t.Errorf("target = %d, want %d", got, KlaberjassTargetScoreDefault)
	}
}

func TestKlaberjassConfigValidate(t *testing.T) {
	cfg := DefaultKlaberjassConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	for _, bad := range []KlaberjassConfig{
		{CpuDifficulty: 9, TargetScore: 501},
		{CpuDifficulty: KlaberjassCpuDifficultyNormal, TargetScore: 0},
		{CpuDifficulty: KlaberjassCpuDifficultyNormal, TargetScore: 99999},
	} {
		if err := bad.Validate(); err == nil {
			t.Errorf("%+v must not validate", bad)
		}
	}
}

func TestKlaberjassRoundTripsThroughJSON(t *testing.T) {
	k := NewDefaultKlaberjass()
	k.Reset()
	_ = k.AcceptTrump(k.GetBidPlayerIdx())

	data, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Klaberjass
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetTrumpSuit() != k.GetTrumpSuit() {
		t.Errorf("trump = %d, want %d", restored.GetTrumpSuit(), k.GetTrumpSuit())
	}
	if restored.GetMakerIdx() != k.GetMakerIdx() {
		t.Errorf("maker = %d, want %d", restored.GetMakerIdx(), k.GetMakerIdx())
	}
	if restored.GetPlayer(0).GetCardsSize() != KlaberjassHandSize {
		t.Errorf("the restored hand holds %d, want %d", restored.GetPlayer(0).GetCardsSize(), KlaberjassHandSize)
	}
	// **復元後もプレイできる。**山札が nil のままだと次のディールで落ちる。
	if err := restored.NextDeal(); err == nil {
		t.Error("NextDeal mid-hand is still refused after a restore")
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestKlaberjassRejectsBadJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[],"ph":0,"di":0,"ci":0,"bi":0}`},
		{"bad phase", `{"pl":[{},{}],"ph":99,"di":0,"ci":0,"bi":0}`},
		{"bad dealer", `{"pl":[{},{}],"ph":0,"di":5,"ci":0,"bi":0}`},
		{"bad current", `{"pl":[{},{}],"ph":0,"di":0,"ci":-1,"bi":0}`},
		{"bad maker", `{"pl":[{},{}],"ph":0,"di":0,"ci":0,"bi":0,"mi":7}`},
		{"bad trump suit", `{"pl":[{},{}],"ph":0,"di":0,"ci":0,"bi":0,"mi":-1,"ts":9}`},
		{"oversized trick", `{"pl":[{},{}],"ph":0,"di":0,"ci":0,"bi":0,"mi":-1,"tk":[{},{},{}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var k Klaberjass
			if err := json.Unmarshal([]byte(tc.body), &k); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}
