//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"testing"
)

func knCard(suit, value int) *Card { return NewCard(suit, value, true) }

// knPlaying puts a game into the play phase with a fixed chosen suit.
func knPlaying(t *testing.T, chosenSuit int) *Karnoffel {
	t.Helper()
	k := NewDefaultKarnoffel()
	k.Reset()
	k.SetPhaseForTest(KarnoffelPhasePlay)
	k.SetChosenSuitForTest(chosenSuit)
	k.SetCurrentPlayerForTest(0)
	k.SetTrickLeaderForTest(0)
	// 第 1 トリックの悪魔制限を外して、比較そのものを見る。
	k.SetTrickNumberForTest(1)
	return k
}

// TestKarnoffelDealsFiveEachWithFaceUpCards は issue の
// 「12枚配布し、最後の1枚を表向きにして選ばれたスートを決定する」が誤りである
// ことを押さえる。
//
// **1 人 5 枚。**48 ÷ 4 = 12 だと配り切ってしまい、めくる 1 枚が残らない。
func TestKarnoffelDealsFiveEachWithFaceUpCards(t *testing.T) {
	if got := len(newKarnoffelDeck()); got != KarnoffelDeckSize {
		t.Fatalf("the deck holds %d cards, want %d", got, KarnoffelDeckSize)
	}
	if KarnoffelHandSize != 5 {
		t.Fatalf("KarnoffelHandSize = %d, want 5 — the issue's 12 is wrong", KarnoffelHandSize)
	}
	// **issue の 12 枚ずつだとデッキを使い切る。**めくる札が残らない。
	if KarnoffelPlayerCnt*12 != KarnoffelDeckSize {
		t.Fatalf("the issue's 12 each would not even consume the pack; check the constants")
	}
	// 5 トリック制と手札枚数は一致していなければならない。
	if KarnoffelTricks != KarnoffelHandSize {
		t.Errorf("%d tricks against %d cards — the issue's own 5-trick rule contradicts its 12-card deal",
			KarnoffelTricks, KarnoffelHandSize)
	}

	k := NewDefaultKarnoffel()
	k.Reset()
	for i := range KarnoffelPlayerCnt {
		if got := k.GetPlayer(i).GetCardsSize(); got != KarnoffelHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, KarnoffelHandSize)
		}
		// **各席に 1 枚だけ表向きの札がある。**
		up := k.GetUpCard(i)
		if up == nil {
			t.Errorf("seat %d has no face-up card", i)
			continue
		}
		// 表向きの札は手札の一部でもある。
		found := false
		for j := range k.GetPlayer(i).GetCardsSize() {
			c := k.GetPlayer(i).GetCard(j)
			if c.GetDesign() == up.GetDesign() && c.GetValue() == up.GetValue() {
				found = true
			}
		}
		if !found {
			t.Errorf("seat %d's face-up card is not in its hand", i)
		}
	}
	if k.GetUpCard(-1) != nil || k.GetUpCard(99) != nil {
		t.Error("an out-of-range seat has no face-up card")
	}
}

// TestKarnoffelDeckHasNoAces は issue の「8を除く」が誤りであることを押さえる。
//
// **A を除いた 48 枚。**A が無いことが平スートの序列の前提。
func TestKarnoffelDeckHasNoAces(t *testing.T) {
	for _, c := range newKarnoffelDeck() {
		if c.GetValue() == 1 {
			t.Fatal("the pack must have no aces — the ranking assumes they are gone")
		}
		if c.GetValue() == 8 {
			// 8 は残る。issue の「8を除く」とは逆。
			return
		}
	}
	t.Error("the eights must stay in the pack")
}

// **切札は表向きの 4 枚のうち最も低い札が決める。**「最後の 1 枚」ではない。
func TestKarnoffelChosenSuitComesFromTheLowestFaceUpCard(t *testing.T) {
	up := []*Card{
		knCard(CardDesignSpade, 13),
		knCard(CardDesignHeart, 3), // これが最低
		knCard(CardDesignClover, 10),
		knCard(CardDesignDiamond, 9),
	}
	if got := karnoffelChooseSuit(up); got != CardDesignHeart {
		t.Errorf("chosen suit = %d, want hearts — the LOWEST face-up card decides", got)
	}

	// nil 混入でも落ちない。
	if got := karnoffelChooseSuit([]*Card{nil, knCard(CardDesignClover, 2)}); got != CardDesignClover {
		t.Errorf("chosen suit = %d, want clubs", got)
	}
	if got := karnoffelChooseSuit([]*Card{nil}); got != 0 {
		t.Errorf("chosen suit = %d, want none", got)
	}

	// 実際に配っても、選ばれたスートは表向きの札のどれかのスートになる。
	k := NewDefaultKarnoffel()
	k.Reset()
	seen := false
	for i := range KarnoffelPlayerCnt {
		if c := k.GetUpCard(i); c != nil && c.GetDesign() == k.GetChosenSuit() {
			seen = true
		}
	}
	if !seen {
		t.Error("the chosen suit must be one of the face-up cards' suits")
	}
}

// TestKarnoffelTheKarnoffelIsTheJack は issue の「Karnöffel（6）が最強」が
// 誤りであることを押さえる。
//
// **カルニッフェルは J。6 は法王（Pope）で別の札。**
func TestKarnoffelTheKarnoffelIsTheJack(t *testing.T) {
	if KarnoffelKarnoffel != 11 {
		t.Fatalf("the Karnöffel is card %d, want the jack (11) — the issue's 6 is the Pope", KarnoffelKarnoffel)
	}
	if KarnoffelPope != 6 {
		t.Fatalf("the Pope is card %d, want 6", KarnoffelPope)
	}

	const chosen = CardDesignHeart
	karnoffel := knCard(chosen, KarnoffelKarnoffel)
	pope := knCard(chosen, KarnoffelPope)

	// **カルニッフェルは法王に勝つ。**issue のとおりなら逆になる。
	if !KarnoffelBeats(karnoffel, pope, pope, chosen, false, true) {
		t.Error("the jack (Karnöffel) beats the six (Pope)")
	}
	if KarnoffelBeats(pope, karnoffel, karnoffel, chosen, false, true) {
		t.Error("the six must not beat the jack")
	}

	// カルニッフェルはリードされた悪魔にも勝つ。
	devil := knCard(chosen, KarnoffelDevil)
	if !KarnoffelBeats(karnoffel, devil, devil, chosen, false, true) {
		t.Error("the Karnöffel beats even a led devil")
	}
	// 平スートの K にも勝つ。
	if !KarnoffelBeats(karnoffel, knCard(CardDesignSpade, 13), knCard(CardDesignSpade, 13), chosen, false, true) {
		t.Error("the Karnöffel beats any plain card")
	}
}

// TestKarnoffelDevilOnlyWinsWhenLed covers the rule the issue omits entirely.
//
// **悪魔（7）はリードされたときだけ強く、追随して出すとあらゆる札に負ける。**
func TestKarnoffelDevilOnlyWinsWhenLed(t *testing.T) {
	const chosen = CardDesignHeart
	devil := knCard(chosen, KarnoffelDevil)
	pope := knCard(chosen, KarnoffelPope)
	weak := knCard(CardDesignSpade, 2)

	// リードされた悪魔は法王に勝つ。
	if KarnoffelBeats(pope, devil, devil, chosen, false, true) {
		t.Error("a LED devil beats the Pope")
	}
	// 追随して出した悪魔は、いちばん弱い平札にすら勝てない。
	if KarnoffelBeats(devil, weak, weak, chosen, false, true) {
		t.Error("a devil played in follow loses to EVERY card")
	}
	// 逆向きでも同じ。
	if !KarnoffelBeats(weak, devil, weak, chosen, false, false) {
		t.Error("any card beats a devil that was not led")
	}
}

// **第 1 トリックのリードに悪魔は使えない。**
func TestKarnoffelDevilCannotLeadTheFirstTrick(t *testing.T) {
	k := NewDefaultKarnoffel()
	k.Reset()
	k.SetChosenSuitForTest(CardDesignHeart)
	k.SetTrickNumberForTest(0)
	k.SetCurrentPlayerForTest(0)
	k.SetTrickLeaderForTest(0)
	k.SetHandForTest(0, []*Card{
		knCard(CardDesignHeart, KarnoffelDevil),
		knCard(CardDesignSpade, 9),
	})

	valid := k.KarnoffelValidPlays(0)
	if len(valid) != 1 || valid[0] != 1 {
		t.Errorf("valid = %v, want only the plain card — the devil cannot lead the first trick", valid)
	}
	if err := k.PlayCard(0, 0); err == nil {
		t.Error("leading the first trick with the devil must be refused")
	}
	if err := k.PlayCard(0, 1); err != nil {
		t.Errorf("the plain card must be accepted: %v", err)
	}

	// 2 トリック目以降なら出せる。
	k2 := knPlaying(t, CardDesignHeart)
	k2.SetHandForTest(0, []*Card{knCard(CardDesignHeart, KarnoffelDevil)})
	if got := len(k2.KarnoffelValidPlays(0)); got != 1 {
		t.Errorf("%d plays are legal from the second trick on, want 1", got)
	}
	if err := k2.PlayCard(0, 0); err != nil {
		t.Errorf("the devil may lead a later trick: %v", err)
	}

	// **悪魔しか無ければ出さざるを得ない。**手が止まってはいけない。
	k3 := NewDefaultKarnoffel()
	k3.Reset()
	k3.SetChosenSuitForTest(CardDesignHeart)
	k3.SetTrickNumberForTest(0)
	k3.SetCurrentPlayerForTest(0)
	k3.SetTrickLeaderForTest(0)
	k3.SetHandForTest(0, []*Card{knCard(CardDesignHeart, KarnoffelDevil)})
	if got := len(k3.KarnoffelValidPlays(0)); got != 1 {
		t.Errorf("%d plays are legal, want the devil — there is nothing else to lead", got)
	}
}

// TestKarnoffelPartialTrumps covers the 3/4/5 rule the issue omits.
//
// **3 は K に、4 は K・Q に、5 は絵札すべてに負ける。**
func TestKarnoffelPartialTrumps(t *testing.T) {
	const chosen = CardDesignHeart
	for _, tc := range []struct {
		name       string
		value      int
		losesTo    []int
		stillBeats []int
	}{
		{"the three loses to kings", KarnoffelOberstecher, []int{13}, []int{12, 11, 10, 9}},
		{"the four loses to kings and queens", KarnoffelUnterstecher, []int{13, 12}, []int{11, 10, 9}},
		{"the five loses to every face card", KarnoffelFarbenstecher, []int{13, 12, 11}, []int{10, 9, 8}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			partial := knCard(chosen, tc.value)
			for _, v := range tc.losesTo {
				plain := knCard(CardDesignSpade, v)
				if KarnoffelBeats(partial, plain, plain, chosen, false, true) {
					t.Errorf("the %d must lose to a plain %d", tc.value, v)
				}
			}
			for _, v := range tc.stillBeats {
				plain := knCard(CardDesignSpade, v)
				if !KarnoffelBeats(partial, plain, plain, chosen, false, true) {
					t.Errorf("the %d must beat a plain %d", tc.value, v)
				}
			}
		})
	}

	// nil 相手や部分切札でない値でも落ちない。
	if !karnoffelPartialBeats(KarnoffelOberstecher, nil) {
		t.Error("a partial trump beats nothing at all")
	}
	if !karnoffelPartialBeats(KarnoffelKarnoffel, knCard(CardDesignSpade, 13)) {
		t.Error("the Karnöffel is not a partial trump and is not held back")
	}
	// デッキに無い値は役職序列を持たない。
	if got := karnoffelChosenRankOf(1); got != 0 {
		t.Errorf("an ace ranks %d in the chosen suit, want 0", got)
	}

	// 部分切札どうしは通常の切札序列で決まる (3 > 4 > 5)。
	three := knCard(chosen, KarnoffelOberstecher)
	four := knCard(chosen, KarnoffelUnterstecher)
	five := knCard(chosen, KarnoffelFarbenstecher)
	if !KarnoffelBeats(three, four, three, chosen, true, false) {
		t.Error("the three outranks the four")
	}
	if !KarnoffelBeats(four, five, four, chosen, true, false) {
		t.Error("the four outranks the five")
	}
}

// 選ばれたスート内の序列。
func TestKarnoffelChosenSuitRanking(t *testing.T) {
	const chosen = CardDesignHeart
	// **J > 6 > 2 > 3 > 4 > 5 > K > Q > 10 > 9 > 8。**悪魔は文脈依存なので別扱い。
	descending := []int{KarnoffelKarnoffel, KarnoffelPope, KarnoffelKaiser,
		KarnoffelOberstecher, KarnoffelUnterstecher, KarnoffelFarbenstecher, 13, 12, 10, 9, 8}
	for i := 1; i < len(descending); i++ {
		hi := karnoffelChosenRank(knCard(chosen, descending[i-1]))
		lo := karnoffelChosenRank(knCard(chosen, descending[i]))
		if hi <= lo {
			t.Errorf("chosen-suit %d must outrank %d", descending[i-1], descending[i])
		}
	}
	if karnoffelChosenRank(nil) != 0 {
		t.Error("a nil card has no rank")
	}
}

// 平スートの序列は A 抜きの K > Q > J > 10 > … > 2。
func TestKarnoffelPlainRanking(t *testing.T) {
	descending := []int{13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2}
	for i := 1; i < len(descending); i++ {
		hi := karnoffelPlainRank(knCard(CardDesignSpade, descending[i-1]))
		lo := karnoffelPlainRank(knCard(CardDesignSpade, descending[i]))
		if hi <= lo {
			t.Errorf("%d must outrank %d", descending[i-1], descending[i])
		}
	}
	// A はデッキに無いので序列も持たない。
	if got := karnoffelPlainRank(knCard(CardDesignSpade, 1)); got != 0 {
		t.Errorf("an ace ranks %d, want 0 — it is not in this pack", got)
	}
	if karnoffelPlainRank(nil) != 0 {
		t.Error("a nil card has no rank")
	}
	if KarnoffelIsFaceCard(nil) {
		t.Error("a nil card is not a face card")
	}
	if !KarnoffelIsFaceCard(knCard(CardDesignSpade, 11)) {
		t.Error("the jack is a face card")
	}
	if KarnoffelIsFaceCard(knCard(CardDesignSpade, 10)) {
		t.Error("the ten is not a face card")
	}
}

// **勝者判定は 1 箇所に集約されている。**途中経過 (CPU の読み) と最終判定で
// 同じ規則を使うので、片方だけ直して食い違うことがない。
func TestKarnoffelLeadingCardIsSharedBetweenReadAndResolve(t *testing.T) {
	const chosen = CardDesignHeart
	// リードされた悪魔は、あとから出たどの札にも抜かれない。
	trick := []*Card{
		knCard(chosen, KarnoffelDevil),
		knCard(chosen, KarnoffelPope),
		knCard(CardDesignSpade, 13),
	}
	off, best := karnoffelLeadingCard(trick, chosen)
	if off != 0 || best != trick[0] {
		t.Errorf("offset = %d, want the led devil to still be winning", off)
	}

	// 追随して出した悪魔は逆に勝てない。
	trick2 := []*Card{
		knCard(CardDesignSpade, 2),
		knCard(chosen, KarnoffelDevil),
	}
	off2, _ := karnoffelLeadingCard(trick2, chosen)
	if off2 != 0 {
		t.Errorf("offset = %d, want the plain lead to hold — a followed devil loses to everything", off2)
	}

	// カルニッフェルはリードされた悪魔をも抜く。
	trick3 := []*Card{
		knCard(chosen, KarnoffelDevil),
		knCard(chosen, KarnoffelKarnoffel),
	}
	off3, _ := karnoffelLeadingCard(trick3, chosen)
	if off3 != 1 {
		t.Errorf("offset = %d, want the Karnöffel to beat even a led devil", off3)
	}

	// 空のトリックでも落ちない。
	if off, best := karnoffelLeadingCard(nil, chosen); off != 0 || best != nil {
		t.Error("an empty trick has no leader")
	}
}

// 平札どうしはリードスートに追随したものだけが争う。
func TestKarnoffelPlainTrickResolution(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	k.SetHandForTest(0, []*Card{knCard(CardDesignSpade, 9)})
	k.SetHandForTest(1, []*Card{knCard(CardDesignSpade, 13)})
	k.SetHandForTest(2, []*Card{knCard(CardDesignClover, 13)}) // 追随していない
	k.SetHandForTest(3, []*Card{knCard(CardDesignSpade, 10)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := k.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}
	if got := k.GetTricksWon(1); got != 1 {
		t.Errorf("the king of the led suit takes it, seat 1 has %d", got)
	}
	// **追随の義務は無いが、追随しない札は勝てない。**
	if got := k.GetTricksWon(2); got != 0 {
		t.Errorf("seat 2 discarded off-suit and must not win, has %d", got)
	}
}

// **追随の義務は無い。**好きな札を出せる。
func TestKarnoffelHasNoFollowSuitRequirement(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	k.SetHandForTest(0, []*Card{knCard(CardDesignSpade, 9)})
	k.SetHandForTest(1, []*Card{knCard(CardDesignSpade, 13), knCard(CardDesignClover, 2)})
	if err := k.PlayCard(0, 0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if got := len(k.KarnoffelValidPlays(1)); got != 2 {
		t.Errorf("%d plays are legal, want both — there is no obligation to follow", got)
	}
	if err := k.PlayCard(1, 1); err != nil {
		t.Errorf("discarding while holding the led suit must be allowed: %v", err)
	}
	if k.KarnoffelValidPlays(99) != nil {
		t.Error("an unknown seat has no legal plays")
	}
}

// **3 トリック取った時点で局は決まる。**5 トリック全部を打つ必要はない。
func TestKarnoffelHandEndsAtThreeTricks(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	// チーム 0 (席 0 と 2) が既に 2 トリック。
	k.SetTricksWonForTest(0, 1)
	k.SetTricksWonForTest(2, 1)
	k.SetTrickNumberForTest(2)

	k.SetHandForTest(0, []*Card{knCard(CardDesignSpade, 13)})
	k.SetHandForTest(1, []*Card{knCard(CardDesignSpade, 2)})
	k.SetHandForTest(2, []*Card{knCard(CardDesignSpade, 3)})
	k.SetHandForTest(3, []*Card{knCard(CardDesignSpade, 4)})
	for _, seat := range []int{0, 1, 2, 3} {
		if err := k.PlayCard(seat, 0); err != nil {
			t.Fatalf("PlayCard(%d): %v", seat, err)
		}
	}

	if got := k.GetPhase(); got != KarnoffelPhaseHandEnd {
		t.Fatalf("phase = %v, want the hand to end at three tricks", got)
	}
	res := k.GetLastResult()
	if res == nil || res.WinnerTeam != 0 {
		t.Fatalf("winner = %v, want team 0", res)
	}
	if got := res.Tricks[0]; got != KarnoffelTricksToWin {
		t.Errorf("team 0 took %d, want %d", got, KarnoffelTricksToWin)
	}
	if got := k.GetHandsWon(0); got != 1 {
		t.Errorf("team 0 has %d hands, want 1", got)
	}
	if got := res.ChosenSuit; got != CardDesignHeart {
		t.Errorf("the result records suit %d, want hearts", got)
	}
}

func TestKarnoffelTeamTricks(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	k.SetTricksWonForTest(0, 2)
	k.SetTricksWonForTest(2, 1)
	k.SetTricksWonForTest(1, 1)
	k.SetTricksWonForTest(3, 1)
	if got := k.KarnoffelTeamTricks(0); got != 3 {
		t.Errorf("team 0 took %d, want 3", got)
	}
	if got := k.KarnoffelTeamTricks(1); got != 2 {
		t.Errorf("team 1 took %d, want 2", got)
	}
	if k.KarnoffelTeamTricks(99) != 0 {
		t.Error("an out-of-range team took nothing")
	}
	// **パートナーは向かい合わせ。**
	if KarnoffelTeamOf(0) != KarnoffelTeamOf(2) || KarnoffelTeamOf(1) != KarnoffelTeamOf(3) {
		t.Error("seats 0/2 and 1/3 are partners")
	}
	// **範囲外は -1。**Go の剰余は -1 % 2 = -1 なので素通しは危険。
	if got := KarnoffelTeamOf(-1); got != -1 {
		t.Errorf("KarnoffelTeamOf(-1) = %d, want -1", got)
	}
	if got := KarnoffelTeamOf(99); got != -1 {
		t.Errorf("KarnoffelTeamOf(99) = %d, want -1", got)
	}
}

func TestKarnoffelPlayGuards(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	k.SetHandForTest(0, []*Card{knCard(CardDesignSpade, 9)})
	if err := k.PlayCard(1, 0); err == nil {
		t.Error("playing out of turn must be refused")
	}
	if err := k.PlayCard(0, 99); err == nil {
		t.Error("an out-of-range index must be refused")
	}
	k.SetPhaseForTest(KarnoffelPhaseHandEnd)
	if err := k.PlayCard(0, 0); err == nil {
		t.Error("playing outside the play phase must be refused")
	}
	k.SetHandForTest(99, nil)
}

func TestKarnoffelNextHandRotatesTheDealer(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	k.SetTricksWonForTest(0, 3)
	k.FinishHandForTest()

	dealer := k.GetDealerIdx()
	if err := k.NextHand(); err != nil {
		t.Fatalf("NextHand: %v", err)
	}
	if got := k.GetDealerIdx(); got == dealer {
		t.Errorf("the dealer stayed at %d; it must rotate", got)
	}
	for i := range KarnoffelPlayerCnt {
		if got := k.GetPlayer(i).GetCardsSize(); got != KarnoffelHandSize {
			t.Errorf("seat %d holds %d, want %d", i, got, KarnoffelHandSize)
		}
	}
}

func TestKarnoffelNextHandGuards(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	if err := k.NextHand(); err == nil {
		t.Error("dealing again mid-hand must be refused")
	}
}

// 規定局数を取ったチームが勝つ。
func TestKarnoffelGameEnd(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	k.SetHandsWonForTest(0, k.GetConfig().TargetHands-1)
	k.SetTricksWonForTest(0, 3)
	k.FinishHandForTest()

	if !k.GetGameEndFlag() {
		t.Fatal("reaching the target ends the game")
	}
	if got := k.GetWinnerTeam(); got != 0 {
		t.Errorf("winner = %d, want 0", got)
	}
	if err := k.NextHand(); err == nil {
		t.Error("dealing after the game is over must be refused")
	}
	if got := k.GetPhase(); got != KarnoffelPhaseGameEnd {
		t.Errorf("phase = %v, want game end", got)
	}
}

// 5 トリック打ち切っても 3 に届かなければ勝者なし。
func TestKarnoffelHandWithoutAWinner(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	k.SetTricksWonForTest(0, 2)
	k.SetTricksWonForTest(1, 2)
	k.FinishHandForTest()

	res := k.GetLastResult()
	if res == nil {
		t.Fatal("the hand must settle")
	}
	if res.WinnerTeam != -1 {
		t.Errorf("winner = %d, want none — neither side reached three", res.WinnerTeam)
	}
	if k.GetHandsWon(0) != 0 || k.GetHandsWon(1) != 0 {
		t.Error("nobody banks a hand when neither side reached three")
	}
}

func TestKarnoffelIsHumanTurnAndCpuPlay(t *testing.T) {
	k := NewDefaultKarnoffel()
	k.Reset()
	// リードは親の左 (席 1) なので CPU。
	if k.IsHumanTurn() {
		t.Error("the seat to the dealer's left leads and it is a CPU")
	}
	k.CpuPlay()
	if len(k.GetTrick()) == 0 {
		t.Error("CpuPlay must put a card on the table")
	}

	over := NewDefaultKarnoffel()
	over.Reset()
	over.SetPhaseForTest(KarnoffelPhaseGameEnd)
	over.gameEndFlag = true
	if over.IsHumanTurn() {
		t.Error("a finished game is nobody's turn")
	}
	over.CpuPlay()

	settled := knPlaying(t, CardDesignHeart)
	settled.SetPhaseForTest(KarnoffelPhaseHandEnd)
	if settled.IsHumanTurn() {
		t.Error("the settlement is nobody's turn")
	}
	settled.CpuPlay()
}

// **CPU だけで 1 局を回し切れること。**途中で止まると詰む。
func TestKarnoffelCpuDrivesAFullHand(t *testing.T) {
	for attempt := range 50 {
		k := NewDefaultKarnoffel()
		k.Reset()
		for step := 0; step < 200; step++ {
			if k.GetPhase() != KarnoffelPhasePlay {
				break
			}
			if k.IsHumanTurn() {
				idx := k.GetCurrentPlayerIdx()
				if i := k.KarnoffelCpuPlay(idx); i >= 0 {
					_ = k.PlayCard(idx, i)
				}
				continue
			}
			k.CpuPlay()
		}
		if k.GetPhase() == KarnoffelPhasePlay {
			t.Fatalf("attempt %d: the hand never finished", attempt)
		}
		// **3 トリックで打ち切るので、5 トリック全部とは限らない。**
		played := k.GetTrickNumber()
		if played < KarnoffelTricksToWin || played > KarnoffelTricks {
			t.Fatalf("attempt %d: %d tricks played, want %d-%d",
				attempt, played, KarnoffelTricksToWin, KarnoffelTricks)
		}
		total := 0
		for i := range KarnoffelPlayerCnt {
			total += k.GetTricksWon(i)
		}
		if total != played {
			t.Fatalf("attempt %d: %d tricks accounted for, want %d", attempt, total, played)
		}
	}
}

func TestKarnoffelCpuEdges(t *testing.T) {
	k := knPlaying(t, CardDesignHeart)
	k.SetHandForTest(0, []*Card{})
	if got := k.KarnoffelCpuPlay(0); got != -1 {
		t.Errorf("an empty hand has no play, got %d", got)
	}
	if got := k.KarnoffelCpuPlay(99); got != -1 {
		t.Errorf("an unknown seat has no play, got %d", got)
	}

	// **悪魔はリードしてこそ強い。**第 1 トリック以外なら真っ先に出す。
	lead := knPlaying(t, CardDesignHeart)
	lead.SetHandForTest(0, []*Card{
		knCard(CardDesignSpade, 13),
		knCard(CardDesignHeart, KarnoffelDevil),
	})
	if got := lead.KarnoffelCpuPlay(0); got != 1 {
		t.Errorf("the CPU led index %d, want the devil", got)
	}
}

func TestKarnoffelAccessors(t *testing.T) {
	k := NewDefaultKarnoffel()
	k.Reset()
	if got := k.GetHandNumber(); got != 1 {
		t.Errorf("hand number = %d, want 1", got)
	}
	if got := k.GetWinnerTeam(); got != -1 {
		t.Errorf("winner = %d, want -1", got)
	}
	if k.GetLastResult() != nil {
		t.Error("there is no result before the first hand settles")
	}
	if got := len(k.GetPlayers()); got != KarnoffelPlayerCnt {
		t.Errorf("%d seats, want %d", got, KarnoffelPlayerCnt)
	}
	if k.GetPlayer(-1) != nil || k.GetPlayer(99) != nil {
		t.Error("an out-of-range seat must be nil")
	}
	if got := len(k.GetTrick()); got != 0 {
		t.Errorf("the trick starts empty, got %d", got)
	}
	// リード席は親の左から始まる。
	if got := k.GetTrickLeaderIdx(); got != (k.GetDealerIdx()+1)%KarnoffelPlayerCnt {
		t.Errorf("leader = %d, want the seat to the dealer's left", got)
	}
	k.SetDealerForTest(2)
	if got := k.GetDealerIdx(); got != 2 {
		t.Errorf("dealer = %d, want 2", got)
	}
	if len(k.GetActionLog()) == 0 {
		t.Error("dealing writes to the action log")
	}
	for _, idx := range []int{-1, KarnoffelPlayerCnt} {
		if k.GetTricksWon(idx) != 0 {
			t.Errorf("seat %d must read as zero", idx)
		}
	}
	for _, team := range []int{-1, KarnoffelTeamCnt} {
		if k.GetHandsWon(team) != 0 {
			t.Errorf("team %d must read as zero", team)
		}
	}
	k.SetTricksWonForTest(99, 5)
	k.SetHandsWonForTest(99, 5)

	cfg := k.GetConfig()
	if cfg.TargetHands != KarnoffelDefaultTarget {
		t.Errorf("target = %d, want %d", cfg.TargetHands, KarnoffelDefaultTarget)
	}
	cfg.TargetHands = 5
	k.SetConfig(cfg)
	if k.GetConfig().TargetHands != 5 {
		t.Error("SetConfig must take effect")
	}
	// プレイヤーのチーム判定。
	p := k.GetPlayer(0)
	if p.GetTeam(0) != p.GetTeam(2) || p.GetTeam(0) == p.GetTeam(1) {
		t.Error("seats 0/2 are partners and 0/1 are opponents")
	}
}

func TestKarnoffelConfigValidate(t *testing.T) {
	if err := DefaultKarnoffelConfig().Validate(); err != nil {
		t.Errorf("the default config must validate: %v", err)
	}
	if err := (KarnoffelConfig{CpuDifficulty: 9, TargetHands: 3}).Validate(); err == nil {
		t.Error("a bad difficulty must not validate")
	}
	for _, n := range []int{KarnoffelMinTarget - 1, KarnoffelMaxTarget + 1} {
		if err := (KarnoffelConfig{TargetHands: n}).Validate(); err == nil {
			t.Errorf("%d hands must not validate", n)
		}
	}
}

func TestKarnoffelRoundTripsThroughJSON(t *testing.T) {
	k := NewDefaultKarnoffel()
	k.Reset()

	data, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Karnoffel
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := restored.GetChosenSuit(); got != k.GetChosenSuit() {
		t.Errorf("chosen suit = %d, want %d", got, k.GetChosenSuit())
	}
	if got := restored.GetPlayer(0).GetCardsSize(); got != KarnoffelHandSize {
		t.Errorf("the restored hand holds %d, want %d", got, KarnoffelHandSize)
	}
	// **表向きの札も往復する。**切札の根拠が見えなくなると盤面が読めない。
	if restored.GetUpCard(0) == nil {
		t.Error("the face-up cards did not survive the round trip")
	}
}

// **壊れた状態を弾く。**KV から戻る値なので、範囲外のまま受け入れると詰む。
func TestKarnoffelRejectsBadJSON(t *testing.T) {
	base := `"pl":[{},{},{},{}],"cf":{"cd":0,"th":3},"ph":0,"di":0,"ci":0,"tl":0,"wt":-1,"cs":0`
	for _, tc := range []struct {
		name string
		body string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[],"cf":{"cd":0,"th":3},"ph":0}`},
		{"bad phase", `{` + base + `,"ph":99}`},
		{"bad dealer", `{` + base + `,"di":9}`},
		{"bad current seat", `{` + base + `,"ci":9}`},
		{"bad trick leader", `{` + base + `,"tl":9}`},
		{"bad winner team", `{` + base + `,"wt":9}`},
		{"bad chosen suit", `{` + base + `,"cs":9}`},
		{"oversized trick", `{` + base + `,"tk":[{},{},{},{},{}]}`},
		{"too many face-up cards", `{` + base + `,"uc":[{},{},{},{},{}]}`},
		{"bad trick number", `{` + base + `,"tn":99}`},
		{"bad config", `{"pl":[{},{},{},{}],"cf":{"cd":0,"th":99},"ph":0,"di":0,"ci":0,"tl":0,"wt":-1,"cs":0}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var k Karnoffel
			if err := json.Unmarshal([]byte(tc.body), &k); err == nil {
				t.Error("must be rejected")
			}
		})
	}
}
