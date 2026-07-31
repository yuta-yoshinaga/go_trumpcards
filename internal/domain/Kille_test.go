//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"testing"
)

// kiReady puts a fresh game into the exchange round for the given seat.
func kiReady(t *testing.T, seat int) *Kille {
	t.Helper()
	k := NewDefaultKille()
	k.Reset()
	k.SetPhaseForTest(KillePhaseExchange)
	k.SetCurrentPlayerForTest(seat)
	// ディーラーを最後の席に固定して、seat 0 が「隣と交換」経路を通るようにする。
	k.SetDealerForTest(KillePlayerCnt - 1)
	return k
}

// TestKilleDeckIsTwentyOneByTwo pins the deck the issue only calls "42 picture
// cards": it is 21 denominations in two copies, one suit.
func TestKilleDeckIsTwentyOneByTwo(t *testing.T) {
	if int(KilleRankCount) != 21 {
		t.Fatalf("KilleRankCount = %d, want 21", KilleRankCount)
	}
	if KilleDeckSize != 42 {
		t.Fatalf("KilleDeckSize = %d, want 42", KilleDeckSize)
	}
	deck := newKilleDeck()
	if len(deck) != 42 {
		t.Fatalf("deck holds %d, want 42", len(deck))
	}
	counts := map[KilleRank]int{}
	for _, c := range deck {
		if c.GetDesign() != KilleDesign {
			t.Fatalf("a Kille card must use the single Kille design, got %d", c.GetDesign())
		}
		counts[KilleRankOf(c)]++
	}
	if len(counts) != 21 {
		t.Fatalf("deck holds %d distinct ranks, want 21", len(counts))
	}
	for r, n := range counts {
		if n != KilleCopies {
			t.Errorf("%s appears %d times, want %d", KilleRankName(r), n, KilleCopies)
		}
	}
	// 6 絵札 + 12 数札 + 3 下位札 = 21。
	pictures, numbers, lows := 0, 0, 0
	for r := KilleMask; r <= KilleHarlequin; r++ {
		switch {
		case KilleIsPicture(r):
			pictures++
		case r >= KilleNum1 && r <= KilleNum12:
			numbers++
		default:
			lows++
		}
	}
	if pictures != 6 || numbers != 12 || lows != 3 {
		t.Errorf("split = %d pictures / %d numbers / %d low, want 6 / 12 / 3", pictures, numbers, lows)
	}
}

// TestKilleRankingOrder pins the order the issue never states.
func TestKilleRankingOrder(t *testing.T) {
	descending := []KilleRank{
		KilleHarlequin, KilleCuckoo, KilleHussar, KillePig, KilleCavalier, KilleInn,
		KilleNum12, KilleNum11, KilleNum10, KilleNum9, KilleNum8, KilleNum7,
		KilleNum6, KilleNum5, KilleNum4, KilleNum3, KilleNum2, KilleNum1,
		KilleWreath, KilleFlowerpot, KilleMask,
	}
	if len(descending) != 21 {
		t.Fatalf("the ladder lists %d ranks, want 21", len(descending))
	}
	for i := 1; i < len(descending); i++ {
		if descending[i-1] <= descending[i] {
			t.Errorf("%s must outrank %s", KilleRankName(descending[i-1]), KilleRankName(descending[i]))
		}
	}
	// 数札は絵札より弱く、下位 3 種より強い。
	if KilleNum12 >= KilleInn {
		t.Error("the highest number must still rank below the Inn")
	}
	if KilleNum1 <= KilleWreath {
		t.Error("the lowest number must still outrank the Wreath")
	}
}

func TestKilleRankNamesAndGlyphs(t *testing.T) {
	for r, want := range map[KilleRank]string{
		KilleHarlequin: "Harlequin", KilleCuckoo: "Cuckoo", KilleHussar: "Hussar",
		KillePig: "Pig", KilleCavalier: "Cavalier", KilleInn: "Inn",
		KilleWreath: "Wreath", KilleFlowerpot: "Flowerpot", KilleMask: "Mask",
	} {
		if got := KilleRankName(r); got != want {
			t.Errorf("name = %q, want %q", got, want)
		}
	}
	// 数札は数字がそのまま名前になる。
	if got := KilleRankName(KilleNum1); got != "1" {
		t.Errorf("name = %q, want %q", got, "1")
	}
	if got := KilleRankName(KilleNum12); got != "12" {
		t.Errorf("name = %q, want %q", got, "12")
	}
	if got := KilleRankGlyph(KilleNum7); got != "7" {
		t.Errorf("glyph = %q, want %q", got, "7")
	}
	if got := KilleRankGlyph(KilleHarlequin); got == "" {
		t.Error("a picture card needs a glyph")
	}
	if got := KilleRankName(KilleRank(99)); got != "?" {
		t.Errorf("unknown rank name = %q, want %q", got, "?")
	}
	if got := KilleRankGlyph(KilleRank(99)); got != "?" {
		t.Errorf("unknown rank glyph = %q, want %q", got, "?")
	}
	// 効果を持つ札は色で見分けられる。
	if KilleRankColor(KilleHarlequin) == KilleRankColor(KilleNum5) {
		t.Error("the Harlequin must not share the plain colour")
	}
}

func TestKilleRankOf(t *testing.T) {
	if got := KilleRankOf(NewKilleCard(KillePig)); got != KillePig {
		t.Errorf("rank = %v, want the Pig", got)
	}
	if got := KilleRankOf(nil); got != 0 {
		t.Errorf("an empty card has rank %v, want 0", got)
	}
	// 標準デッキの札は Kille 札ではない。
	if got := KilleRankOf(NewCard(CardDesignSpade, 5, true)); got != 0 {
		t.Errorf("a French-suited card has rank %v, want 0", got)
	}
	if got := KilleRankOf(NewCard(KilleDesign, 99, true)); got != 0 {
		t.Errorf("an out-of-range value has rank %v, want 0", got)
	}
}

func TestKilleDealGivesOneCardEach(t *testing.T) {
	k := NewDefaultKille()
	k.Reset()
	for i, p := range k.GetPlayers() {
		if got := p.GetCardsSize(); got != 1 {
			t.Errorf("seat %d holds %d cards, want 1", i, got)
		}
	}
	if got := k.GetStockCount(); got != KilleDeckSize-KillePlayerCnt {
		t.Errorf("stock = %d, want %d", got, KilleDeckSize-KillePlayerCnt)
	}
	if got := k.GetPot(); got != KillePlayerCnt*k.GetConfig().Stake {
		t.Errorf("pot = %d, want %d", got, KillePlayerCnt*k.GetConfig().Stake)
	}
	// ディーラーの左隣が先手。
	if got := k.GetCurrentPlayerIdx(); got != 1 {
		t.Errorf("current = %d, want 1", got)
	}
}

// TestKilleCuckooEndsTheRoundImmediately is the first of the two effects the
// issue has backwards: it calls the Cuckoo a *forced* exchange.
func TestKilleCuckooEndsTheRoundImmediately(t *testing.T) {
	k := kiReady(t, 0)
	k.SetHandForTest(0, KilleNum3)
	k.SetHandForTest(1, KilleCuckoo)

	if err := k.Exchange(0); err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	// **交換されず、その場でラウンドが終わる。**
	if k.GetPhase() != KillePhaseShowdown {
		t.Fatalf("phase = %v, want Showdown: nobody swaps the Cuckoo", k.GetPhase())
	}
	if got := KilleRankOf(k.GetPlayer(1).GetCard(0)); got != KilleCuckoo {
		t.Errorf("seat 1 holds %s, want it to have kept the Cuckoo", KilleRankName(got))
	}
	if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KilleNum3 {
		t.Errorf("seat 0 holds %s, want its own card back", KilleRankName(got))
	}
}

// TestKilleHarlequinFlipsWhenSwapped is the second reversed effect: the issue
// says the Harlequin may *refuse* an exchange. It does not — it is exchanged
// face up and its strength inverts.
func TestKilleHarlequinFlipsWhenSwapped(t *testing.T) {
	// 配られた Harlequin は最強。
	dealt := kiReady(t, 0)
	dealt.SetHandForTest(0, KilleHarlequin)
	if got := dealt.KilleStrength(0); got != int(KilleHarlequin) {
		t.Fatalf("a dealt Harlequin has strength %d, want %d", got, int(KilleHarlequin))
	}

	// 交換で渡ってきた Harlequin は最弱。
	k := kiReady(t, 0)
	k.SetHandForTest(0, KilleNum3)
	k.SetHandForTest(1, KilleHarlequin)
	if err := k.Exchange(0); err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KilleHarlequin {
		t.Fatalf("seat 0 holds %s, want the Harlequin after the swap", KilleRankName(got))
	}
	if !k.GetPlayer(0).IsHarlequinSwapped() {
		t.Fatal("the Harlequin arrived by exchange; it must be marked")
	}
	if got := k.KilleStrength(0); got != 0 {
		t.Errorf("a swapped Harlequin has strength %d, want 0 (lowest)", got)
	}
}

// TestKilleHussarStrikesTheChallenger covers an effect the issue omits.
func TestKilleHussarStrikesTheChallenger(t *testing.T) {
	k := kiReady(t, 0)
	k.SetHandForTest(0, KilleNum3)
	k.SetHandForTest(1, KilleHussar)

	if err := k.Exchange(0); err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	// **仕掛けた側**が落ちる。
	if !k.GetPlayer(0).IsOut() {
		t.Error("the challenger must be struck down by the Hussar")
	}
	if got := k.GetPlayer(0).GetKnockedBy(); got != KilleKnockHussar {
		t.Errorf("knocked by %q, want %q", got, KilleKnockHussar)
	}
	if k.GetPlayer(1).IsOut() {
		t.Error("the Hussar's holder stays in")
	}
	// 札は動かない。
	if got := KilleRankOf(k.GetPlayer(1).GetCard(0)); got != KilleHussar {
		t.Errorf("seat 1 holds %s, want it to have kept the Hussar", KilleRankName(got))
	}
}

// TestKillePigBitesBackKnocksItsOwnHolder covers the effect the issue omits
// entirely. **The challenger survives and the Pig's own holder goes out.**
func TestKillePigBitesBackKnocksItsOwnHolder(t *testing.T) {
	k := kiReady(t, 0)
	k.SetHandForTest(0, KilleNum3)
	k.SetHandForTest(1, KillePig)

	if err := k.Exchange(0); err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if !k.GetPlayer(1).IsOut() {
		t.Error("the Pig's own holder goes out")
	}
	if got := k.GetPlayer(1).GetKnockedBy(); got != KilleKnockPig {
		t.Errorf("knocked by %q, want %q", got, KilleKnockPig)
	}
	if k.GetPlayer(0).IsOut() {
		t.Error("the challenger is not the one who goes out to a Pig")
	}
	// 交換自体は取り消される。両者とも自分の札のまま。
	if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KilleNum3 {
		t.Errorf("seat 0 holds %s, want its own card kept", KilleRankName(got))
	}
	if got := KilleRankOf(k.GetPlayer(1).GetCard(0)); got != KillePig {
		t.Errorf("seat 1 holds %s, want it to have kept the Pig", KilleRankName(got))
	}
}

// TestKillePigRevertsOnlySwapsOfItsOwnCard は «all prior swaps involving that
// card reverse» の範囲を確かめる。**巻き戻るのは Pig 自身が動いた交換だけ**で、
// 無関係な交換はそのまま残る。
func TestKillePigRevertsOnlySwapsOfItsOwnCard(t *testing.T) {
	k := kiReady(t, 0)
	k.SetHandForTest(0, KilleNum3)
	k.SetHandForTest(1, KilleNum5)
	k.SetHandForTest(2, KillePig)

	// 0 と 1 が交換する。Pig は関係しない。
	if err := k.Exchange(0); err != nil {
		t.Fatalf("first Exchange: %v", err)
	}
	if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KilleNum5 {
		t.Fatalf("seat 0 holds %s, want the 5 after the swap", KilleRankName(got))
	}

	// 1 が Pig に仕掛ける。
	k.SetCurrentPlayerForTest(1)
	if err := k.Exchange(1); err != nil {
		t.Fatalf("second Exchange: %v", err)
	}
	if !k.GetPlayer(2).IsOut() {
		t.Fatal("the Pig's holder goes out")
	}
	// **Pig は一度も動いていないので、0↔1 の交換は巻き戻らない。**
	if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KilleNum5 {
		t.Errorf("seat 0 holds %s, want the unrelated swap left standing", KilleRankName(got))
	}
	if got := KilleRankOf(k.GetPlayer(1).GetCard(0)); got != KilleNum3 {
		t.Errorf("seat 1 holds %s, want the unrelated swap left standing", KilleRankName(got))
	}
}

// TestKillePigUnwindsItsOwnTravel は、Pig が交換で移動したあとに噛んだ場合に
// その移動が巻き戻ることを確かめる。
//
// 4 人 1 周では自然に起きにくい形なので、履歴を直接組んで確かめる。
func TestKillePigUnwindsItsOwnTravel(t *testing.T) {
	k := kiReady(t, 0)
	// 0 が Pig を持っていて、0↔1 の交換で 1 に渡ったという履歴を作る。
	k.SetHandForTest(0, KilleNum5)
	k.SetHandForTest(1, KillePig)
	k.events = append(k.events, &KilleEvent{Kind: "swap", Actor: 0, Target: 1})

	k.revertSwapsInvolving(1)

	// 巻き戻って Pig は 0 に戻る。
	if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KillePig {
		t.Errorf("seat 0 holds %s, want the Pig back where it started", KilleRankName(got))
	}
	if got := KilleRankOf(k.GetPlayer(1).GetCard(0)); got != KilleNum5 {
		t.Errorf("seat 1 holds %s, want the 5 back", KilleRankName(got))
	}
	// 巻き戻した交換は履歴から消える。
	for _, e := range k.GetEvents() {
		if e.Kind == "swap" {
			t.Error("a reverted swap must not stay in the history")
		}
	}
}

// TestKilleCavalierAndInnPassAlong covers the third omitted effect.
func TestKilleCavalierAndInnPassAlong(t *testing.T) {
	for name, r := range map[string]KilleRank{"cavalier": KilleCavalier, "inn": KilleInn} {
		t.Run(name, func(t *testing.T) {
			k := kiReady(t, 0)
			k.SetHandForTest(0, KilleNum3)
			k.SetHandForTest(1, r)
			k.SetHandForTest(2, KilleNum9)

			if err := k.Exchange(0); err != nil {
				t.Fatalf("Exchange: %v", err)
			}
			// 1 は飛ばされ、**2 と交換**される。
			if got := KilleRankOf(k.GetPlayer(1).GetCard(0)); got != r {
				t.Errorf("seat 1 holds %s, want it to have kept the %s", KilleRankName(got), name)
			}
			if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KilleNum9 {
				t.Errorf("seat 0 holds %s, want the 9 from seat 2", KilleRankName(got))
			}
			if got := KilleRankOf(k.GetPlayer(2).GetCard(0)); got != KilleNum3 {
				t.Errorf("seat 2 holds %s, want the 3", KilleRankName(got))
			}
		})
	}
}

// TestKilleAllPassMeansNoExchange は Cavalier / Inn で一周した場合に詰まらない
// ことを確かめる。
func TestKilleAllPassMeansNoExchange(t *testing.T) {
	k := kiReady(t, 0)
	k.SetHandForTest(0, KilleNum3)
	for i := 1; i < KillePlayerCnt; i++ {
		k.SetHandForTest(i, KilleCavalier)
	}
	if err := k.Exchange(0); err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KilleNum3 {
		t.Errorf("seat 0 holds %s, want its own card kept", KilleRankName(got))
	}
}

// TestKilleDealerExchangesWithTheStock covers the rule the issue omits.
func TestKilleDealerExchangesWithTheStock(t *testing.T) {
	k := kiReady(t, 0)
	k.SetDealerForTest(0)
	k.SetCurrentPlayerForTest(0)
	k.SetHandForTest(0, KilleNum3)
	k.SetStockForTest([]*Card{NewKilleCard(KilleHarlequin)})

	if err := k.Exchange(0); err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if got := KilleRankOf(k.GetPlayer(0).GetCard(0)); got != KilleHarlequin {
		t.Fatalf("seat 0 holds %s, want the card drawn from the stock", KilleRankName(got))
	}
	// **山から引いた Harlequin は最強のまま。**交換で渡ったものだけが弱い。
	if k.GetPlayer(0).IsHarlequinSwapped() {
		t.Error("a Harlequin drawn from the stock must not count as swapped")
	}
	if got := k.KilleStrength(0); got != int(KilleHarlequin) {
		t.Errorf("strength = %d, want %d", got, int(KilleHarlequin))
	}
}

func TestKilleSatisfiedPassesTheTurn(t *testing.T) {
	k := kiReady(t, 0)
	if err := k.Satisfied(0); err != nil {
		t.Fatalf("Satisfied: %v", err)
	}
	if !k.GetPlayer(0).IsSatisfied() {
		t.Error("the seat should be recorded as satisfied")
	}
	if got := k.GetCurrentPlayerIdx(); got != 1 {
		t.Errorf("current = %d, want 1", got)
	}
}

func TestKilleTurnRejections(t *testing.T) {
	k := kiReady(t, 0)
	if err := k.Exchange(1); err == nil {
		t.Error("expected an error acting out of turn")
	}
	if err := k.Satisfied(1); err == nil {
		t.Error("expected an error acting out of turn")
	}
	k.GetPlayer(0).SetOut(KilleKnockHussar)
	if err := k.Exchange(0); err == nil {
		t.Error("expected an error acting after being knocked out")
	}
	k.SetPhaseForTest(KillePhaseShowdown)
	k.GetPlayer(0).out = false
	if err := k.Exchange(0); err == nil {
		t.Error("expected an error acting outside the exchange round")
	}
}

// TestKilleLowestApartFromHarlequinLoses covers the settlement rule the issue
// states without its exception.
func TestKilleLowestApartFromHarlequinLoses(t *testing.T) {
	k := kiReady(t, 0)
	k.SetDealerForTest(0)
	k.SetCurrentPlayerForTest(0)
	k.SetHandForTest(0, KilleNum9)
	k.SetHandForTest(1, KilleMask) // 最弱
	k.SetHandForTest(2, KilleNum5)
	k.SetHandForTest(3, KilleNum7)

	if err := k.Satisfied(0); err != nil {
		t.Fatalf("Satisfied: %v", err)
	}
	if k.GetPhase() != KillePhaseShowdown {
		t.Fatalf("phase = %v, want Showdown after the dealer acts", k.GetPhase())
	}
	if !k.GetPlayer(1).IsOut() {
		t.Error("the Mask is the lowest card; that seat goes out")
	}
	if got := k.GetPlayer(1).GetKnockedBy(); got != KilleKnockLowest {
		t.Errorf("knocked by %q, want the lowest-card marker", got)
	}
}

// TestKilleKnockedPlayersGoOutTogetherWithTheLowest は、Hussar / Pig で落ちた
// 人も最弱の人と一緒に脱落することを確かめる。
func TestKilleKnockedPlayersGoOutTogetherWithTheLowest(t *testing.T) {
	k := kiReady(t, 0)
	k.SetHandForTest(0, KilleNum3)
	k.SetHandForTest(1, KilleHussar)
	k.SetHandForTest(2, KilleMask)
	k.SetHandForTest(3, KilleNum9)

	// 0 が Hussar に仕掛けて落ちる。
	if err := k.Exchange(0); err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	// 残りを回してラウンドを終える。
	for k.GetPhase() == KillePhaseExchange {
		cur := k.GetCurrentPlayerIdx()
		if err := k.Satisfied(cur); err != nil {
			t.Fatalf("Satisfied seat %d: %v", cur, err)
		}
	}
	losers := k.GetLoserIdxs()
	if len(losers) != 2 {
		t.Fatalf("losers = %v, want the Hussar's victim and the lowest card", losers)
	}
	if !k.GetPlayer(0).IsOut() || !k.GetPlayer(2).IsOut() {
		t.Error("both the struck challenger and the Mask holder must be out")
	}
}

// TestKilleReentryCostsEscalate pins the buy-back the issue replaces with lives.
func TestKilleReentryCostsEscalate(t *testing.T) {
	k := kiReady(t, 0)
	k.SetPotForTest(40)
	k.SetPhaseForTest(KillePhaseShowdown)
	p := k.GetPlayer(1)
	p.SetOut(KilleKnockLowest)

	// 1 回目は 1 口。
	if got := k.KilleReentryCost(1); got != k.GetConfig().Stake {
		t.Errorf("first buy-back costs %d, want the stake %d", got, k.GetConfig().Stake)
	}
	if err := k.Reenter(1); err != nil {
		t.Fatalf("first Reenter: %v", err)
	}
	if p.IsOut() {
		t.Error("the seat should be back in")
	}

	// 2 回目はポットの半分。
	p.SetOut(KilleKnockLowest)
	k.SetPotForTest(40)
	if got := k.KilleReentryCost(1); got != 20 {
		t.Errorf("second buy-back costs %d, want half the pot (20)", got)
	}
	if err := k.Reenter(1); err != nil {
		t.Fatalf("second Reenter: %v", err)
	}

	// 3 回目はポット全額。
	p.SetOut(KilleKnockLowest)
	k.SetPotForTest(40)
	if got := k.KilleReentryCost(1); got != 40 {
		t.Errorf("third buy-back costs %d, want the whole pot (40)", got)
	}
	if err := k.Reenter(1); err != nil {
		t.Fatalf("third Reenter: %v", err)
	}

	// **4 回目は無い。**
	p.SetOut(KilleKnockLowest)
	if p.CanReenter() {
		t.Fatalf("a seat may buy back only %d times", KilleMaxReentries)
	}
	if got := k.KilleReentryCost(1); got != 0 {
		t.Errorf("a fourth buy-back costs %d, want 0 (not offered)", got)
	}
	if err := k.Reenter(1); err == nil {
		t.Error("expected an error on a fourth buy-back")
	}
}

func TestKilleReenterRejections(t *testing.T) {
	k := kiReady(t, 0)
	if err := k.Reenter(0); err == nil {
		t.Error("expected an error while the round is still live")
	}
	k.SetPhaseForTest(KillePhaseShowdown)
	if err := k.Reenter(0); err == nil {
		t.Error("expected an error for a seat that is not out")
	}
	if err := k.Reenter(99); err == nil {
		t.Error("expected an error for an unknown seat")
	}
}

func TestKilleNextRoundAndGameEnd(t *testing.T) {
	k := NewDefaultKille()
	k.Reset()
	if err := k.NextRound(); err == nil {
		t.Error("expected an error while the round is still live")
	}

	k.SetPhaseForTest(KillePhaseShowdown)
	if err := k.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	if k.GetPhase() != KillePhaseExchange {
		t.Errorf("phase = %v, want Exchange", k.GetPhase())
	}
	// ディーラーが 1 つ進むので、先手も 1 つ進む。
	if got := k.GetCurrentPlayerIdx(); got != 2 {
		t.Errorf("current = %d, want 2", got)
	}

	// 買い戻せない状態まで落として決着させる。
	for i := 1; i < KillePlayerCnt; i++ {
		p := k.GetPlayer(i)
		p.SetOut(KilleKnockLowest)
		p.reentries = KilleMaxReentries
	}
	k.SetPhaseForTest(KillePhaseShowdown)
	k.checkGameEnd()
	if !k.GetGameEndFlag() {
		t.Fatal("with one seat able to continue the game is decided")
	}
	if got := k.GetWinnerIdx(); got != 0 {
		t.Errorf("winnerIdx = %d, want 0", got)
	}
	if err := k.NextRound(); err == nil {
		t.Error("expected an error dealing after the game is over")
	}
	if err := k.Exchange(0); err == nil {
		t.Error("expected an error acting after the game is over")
	}
	if err := k.Reenter(1); err == nil {
		t.Error("expected an error buying back after the game is over")
	}
}

func TestKilleCpuDecide(t *testing.T) {
	k := kiReady(t, 1)
	k.SetHandForTest(1, KilleMask)
	if act := k.KilleCpuDecide(1); act.Type != "exchange" {
		t.Errorf("with the lowest card the CPU should exchange, got %+v", act)
	}
	k.SetHandForTest(1, KilleCuckoo)
	if act := k.KilleCpuDecide(1); act.Type != "satisfied" {
		t.Errorf("with a top card the CPU should stand pat, got %+v", act)
	}
	// **交換で渡ってきた Harlequin は最弱。**強い札に見えても交換したがる。
	k.SetHandForTest(1, KilleHarlequin)
	k.GetPlayer(1).SetHarlequinSwapped(true)
	if act := k.KilleCpuDecide(1); act.Type != "exchange" {
		t.Errorf("a swapped Harlequin is the lowest card; the CPU should exchange, got %+v", act)
	}
	if act := k.KilleCpuDecide(99); act.Type != "satisfied" {
		t.Errorf("an unknown seat must not produce a move, got %+v", act)
	}
}

// TestKilleCpuDrivesRoundsToAnEnd checks that four CPUs always finish a round.
func TestKilleCpuDrivesRoundsToAnEnd(t *testing.T) {
	for trial := range 30 {
		k := NewDefaultKille()
		k.Reset()
		for range 200 {
			if k.GetPhase() != KillePhaseExchange {
				break
			}
			cur := k.GetCurrentPlayerIdx()
			act := k.KilleCpuDecide(cur)
			var err error
			if act.Type == "exchange" {
				err = k.Exchange(cur)
			} else {
				err = k.Satisfied(cur)
			}
			if err != nil {
				t.Fatalf("trial %d: %v", trial, err)
			}
		}
		if k.GetPhase() == KillePhaseExchange {
			t.Fatalf("trial %d: the round never ended", trial)
		}
		if len(k.GetLoserIdxs()) == 0 {
			t.Fatalf("trial %d: somebody must go out", trial)
		}
	}
}

func TestKilleActionLog(t *testing.T) {
	k := NewDefaultKille()
	k.Reset()
	log := k.GetActionLog()
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

func TestKilleConfigValidate(t *testing.T) {
	if err := DefaultKilleConfig().Validate(); err != nil {
		t.Fatalf("the default config should validate: %v", err)
	}
	if err := (KilleConfig{CpuDifficulty: 5, Stake: 1}).Validate(); err == nil {
		t.Error("expected an error for an unknown difficulty")
	}
	if err := (KilleConfig{Stake: 0}).Validate(); err == nil {
		t.Error("expected an error for a zero stake")
	}
	k := NewDefaultKille()
	k.SetConfig(KilleConfig{Stake: 5})
	if k.GetConfig().Stake != 5 {
		t.Error("SetConfig/GetConfig disagree")
	}
}

func TestKilleAccessorBounds(t *testing.T) {
	k := NewDefaultKille()
	k.Reset()
	if k.GetPlayer(-1) != nil || k.GetPlayer(KillePlayerCnt) != nil {
		t.Error("GetPlayer should return nil outside the table")
	}
	if got := k.KilleStrength(99); got != 0 {
		t.Errorf("strength of an unknown seat = %d, want 0", got)
	}
	if got := k.GetRoundNumber(); got != 0 {
		t.Errorf("GetRoundNumber = %d, want 0", got)
	}
}

func TestKilleJSONRoundTrip(t *testing.T) {
	k := kiReady(t, 0)
	k.SetHandForTest(0, KilleNum3)
	k.SetHandForTest(1, KilleNum5)
	if err := k.Exchange(0); err != nil {
		t.Fatalf("Exchange: %v", err)
	}

	data, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Kille
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.GetPot() != k.GetPot() {
		t.Errorf("pot = %d, want %d", got.GetPot(), k.GetPot())
	}
	if len(got.GetEvents()) != len(k.GetEvents()) {
		t.Errorf("events = %d, want %d", len(got.GetEvents()), len(k.GetEvents()))
	}
	if got.GetStockCount() != k.GetStockCount() {
		t.Errorf("stock = %d, want %d", got.GetStockCount(), k.GetStockCount())
	}
	if len(got.GetActionLog()) != len(k.GetActionLog()) {
		t.Errorf("action log = %d entries, want %d", len(got.GetActionLog()), len(k.GetActionLog()))
	}
}

func TestKilleUnmarshalRejectsGarbage(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[null,null],"cfg":{"cd":0,"st":1},"ph":0}`},
		{"bad config", `{"pl":[{},{},{},{}],"cfg":{"cd":9,"st":1},"ph":0}`},
		{"unknown phase", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"st":1},"ph":9}`},
		{"negative phase", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"st":1},"ph":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var k Kille
			if err := json.Unmarshal([]byte(tt.data), &k); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestKilleUnmarshalClampsIndices(t *testing.T) {
	var k Kille
	data := `{"pl":[{"re":99},{},{},{}],"cfg":{"cd":0,"st":1},"ph":0,"cur":99,"dl":-3,` +
		`"pt":-5,"wi":42,"ls":[0,99],` +
		`"ev":[null,{"Kind":"swap","Actor":99,"Target":0},{"Kind":"swap","Actor":0,"Target":99},` +
		`{"Kind":"stock","Actor":1,"Target":-1}]}`
	if err := json.Unmarshal([]byte(data), &k); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := k.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want it clamped to 0", got)
	}
	if got := k.GetWinnerIdx(); got != -1 {
		t.Errorf("winnerIdx = %d, want -1", got)
	}
	if got := k.GetPot(); got != 0 {
		t.Errorf("pot = %d, want a negative pot clamped to 0", got)
	}
	// **Target の -1 は「山札との交換」という意味を持つ**ので残す。
	if got := len(k.GetEvents()); got != 1 {
		t.Errorf("events = %d, want only the stock exchange kept", got)
	}
	if got := len(k.GetLoserIdxs()); got != 1 {
		t.Errorf("losers = %d, want the out-of-range one dropped", got)
	}
	// **買い戻し回数は上限で頭打ち。**超えた値が残ると無限に買い戻せる。
	if got := k.GetPlayer(0).GetReentries(); got != KilleMaxReentries {
		t.Errorf("reentries = %d, want it capped at %d", got, KilleMaxReentries)
	}
}
