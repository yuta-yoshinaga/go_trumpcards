//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func zwCard(design, value int) *Card { return NewCard(design, value, true) }

// zwJoker returns the joker whose fixed matching value is 15/20/25.
func zwJoker(which int) *Card { return NewCard(CardDesignJoker, which, true) }

// zwReady puts a fresh game into the play phase for the given seat.
func zwReady(t *testing.T, seat int) *Zwicker {
	t.Helper()
	z := NewDefaultZwicker()
	z.Reset()
	z.SetCurrentPlayerForTest(seat)
	z.SetPhaseForTest(ZwickerPhasePlay)
	return z
}

// setZwHand replaces a seat's hand outright.
func setZwHand(z *Zwicker, seat int, cards []*Card) {
	p := z.GetPlayer(seat)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// TestZwickerCardValues pins the rule the issue got backwards: the jokers are
// fixed and it is the aces and court cards that carry a choice.
func TestZwickerCardValues(t *testing.T) {
	tests := []struct {
		name string
		card *Card
		want []int
	}{
		{"ace is one or eleven", zwCard(CardDesignSpade, 1), []int{1, 11}},
		{"jack is two or twelve", zwCard(CardDesignHeart, 11), []int{2, 12}},
		{"queen is three or thirteen", zwCard(CardDesignClover, 12), []int{3, 13}},
		{"king is four or fourteen", zwCard(CardDesignDiamond, 13), []int{4, 14}},
		{"a number card is only itself", zwCard(CardDesignSpade, 7), []int{7}},
		{"the small joker is fixed at fifteen", zwJoker(1), []int{ZwickerJokerSmall}},
		{"the middle joker is fixed at twenty", zwJoker(2), []int{ZwickerJokerMiddle}},
		{"the large joker is fixed at twenty-five", zwJoker(3), []int{ZwickerJokerLarge}},
		{"an empty card has no value", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ZwickerCardValues(tt.card)
			if len(got) != len(tt.want) {
				t.Fatalf("values = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("values = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestZwickerHasValue(t *testing.T) {
	q := zwCard(CardDesignClover, 12)
	if !ZwickerHasValue(q, 3) || !ZwickerHasValue(q, 13) {
		t.Error("a queen must be usable as either 3 or 13")
	}
	if ZwickerHasValue(q, 12) {
		t.Error("a queen is not worth its rank")
	}
	if ZwickerHasValue(nil, 1) {
		t.Error("an empty card has no value")
	}
}

// TestZwickerScoreTotalsThirty is the arithmetic check that the scoring table
// is the real one: the card points plus the majority bonus come to exactly 30.
func TestZwickerScoreTotalsThirty(t *testing.T) {
	sum := ZwickerScoreMajority
	for _, c := range []*Card{zwJoker(1), zwJoker(2), zwJoker(3)} {
		sum += ZwickerScoreOfCard(c)
	}
	sum += ZwickerScoreOfCard(zwCard(CardDesignDiamond, 10))
	sum += ZwickerScoreOfCard(zwCard(CardDesignSpade, 10))
	sum += ZwickerScoreOfCard(zwCard(CardDesignSpade, 2))
	for _, d := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		sum += ZwickerScoreOfCard(zwCard(d, 1))
	}
	if sum != 30 {
		t.Fatalf("the scoring table adds to %d, want 30", sum)
	}
	if ZwickerScoreTotal != 30 {
		t.Fatalf("ZwickerScoreTotal = %d, want 30", ZwickerScoreTotal)
	}
	if got := ZwickerScoreOfCard(zwCard(CardDesignHeart, 9)); got != 0 {
		t.Errorf("a plain card scores %d, want 0", got)
	}
	if got := ZwickerScoreOfCard(nil); got != 0 {
		t.Errorf("an empty card scores %d, want 0", got)
	}
}

// TestZwickerDealExhaustsThePack is the arithmetic that forces three jokers:
// 55 - 3 on the table = 52, which is 13 each for four players (4+4+5).
// With two jokers the final deal would not come out even.
func TestZwickerDealExhaustsThePack(t *testing.T) {
	z := NewDefaultZwicker()
	z.Reset()

	if got := len(z.GetTableCards()); got != ZwickerInitialTableSize {
		t.Fatalf("table = %d, want %d", got, ZwickerInitialTableSize)
	}
	for i, p := range z.GetPlayers() {
		if got := p.GetCardsSize(); got != zwickerDealSizes[0] {
			t.Errorf("seat %d holds %d, want %d", i, got, zwickerDealSizes[0])
		}
	}
	dealt := ZwickerInitialTableSize + zwickerDealSizes[0]*ZwickerPlayerCnt
	total := 52 + ZwickerJokerCnt
	if got := z.GetStockCount(); got != total-dealt {
		t.Fatalf("stock = %d, want %d", got, total-dealt)
	}

	sum := ZwickerInitialTableSize
	for _, n := range zwickerDealSizes {
		sum += n * ZwickerPlayerCnt
	}
	if sum != total {
		t.Fatalf("the three deals use %d cards but the pack holds %d", sum, total)
	}
}

func TestZwickerTeamOf(t *testing.T) {
	// 向かい合わせが味方。
	if ZwickerTeamOf(0) != ZwickerTeamOf(2) || ZwickerTeamOf(1) != ZwickerTeamOf(3) {
		t.Error("seats across the table must be partners")
	}
	if ZwickerTeamOf(0) == ZwickerTeamOf(1) {
		t.Error("adjacent seats must be opponents")
	}
}

func TestZwickerTakeByExactMatch(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 7), zwCard(CardDesignClover, 9)})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 7), zwCard(CardDesignSpade, 3)})

	if err := z.Take(0, 0, 7, []int{0}, nil); err != nil {
		t.Fatalf("Take: %v", err)
	}
	if got := len(z.GetTableCards()); got != 1 {
		t.Errorf("table = %d, want 1", got)
	}
	if got := len(z.GetPlayer(0).GetCaptured()); got != 2 {
		t.Errorf("captured = %d, want 2 (the played card counts)", got)
	}
	if got := z.GetPlayer(0).GetZwicks(); got != 0 {
		t.Errorf("zwicks = %d; the table was not cleared", got)
	}
}

// TestZwickerTakeSeveralGroupsAtOnce is the capture the issue mislabelled as
// "Zwick": a ten taking 7+3 and 6+4 in the same play.
func TestZwickerTakeSeveralGroupsAtOnce(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{
		zwCard(CardDesignHeart, 7), zwCard(CardDesignClover, 3),
		zwCard(CardDesignDiamond, 6), zwCard(CardDesignSpade, 4),
	})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 10)})

	if err := z.Take(0, 0, 10, []int{0, 1, 2, 3}, nil); err != nil {
		t.Fatalf("Take: %v", err)
	}
	if got := len(z.GetPlayer(0).GetCaptured()); got != 5 {
		t.Errorf("captured = %d, want 5", got)
	}
	// **場が空になったので Zwick。**同時取りそのものが Zwick なのではない。
	if got := z.GetPlayer(0).GetZwicks(); got != 1 {
		t.Errorf("zwicks = %d, want 1", got)
	}
}

// TestZwickerZwickIsClearingTheTable pins the distinction directly: a
// multi-group capture that leaves a card behind earns no Zwick.
func TestZwickerZwickIsClearingTheTable(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{
		zwCard(CardDesignHeart, 7), zwCard(CardDesignClover, 3),
		zwCard(CardDesignDiamond, 6), zwCard(CardDesignSpade, 4),
		zwCard(CardDesignHeart, 9),
	})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 10)})

	if err := z.Take(0, 0, 10, []int{0, 1, 2, 3}, nil); err != nil {
		t.Fatalf("Take: %v", err)
	}
	if got := z.GetPlayer(0).GetZwicks(); got != 0 {
		t.Errorf("zwicks = %d; a card was left on the table", got)
	}
}

// TestZwickerCourtCardsJoinSums is what rules out reusing the Cassino engine:
// there a court card matches by rank only.
func TestZwickerCourtCardsJoinSums(t *testing.T) {
	z := zwReady(t, 0)
	// ♥K を 4 として使い、♦3 + ♣A(=1) で取る。
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignDiamond, 3), zwCard(CardDesignClover, 1)})
	setZwHand(z, 0, []*Card{zwCard(CardDesignHeart, 13)})
	if err := z.Take(0, 0, 4, []int{0, 1}, nil); err != nil {
		t.Fatalf("Take with a king as 4: %v", err)
	}

	// 同じ K を 14 として使い、♠Q(=13) + ♠A(=1) で取る。
	z2 := zwReady(t, 0)
	z2.SetTableCardsForTest([]*Card{zwCard(CardDesignSpade, 12), zwCard(CardDesignSpade, 1)})
	setZwHand(z2, 0, []*Card{zwCard(CardDesignHeart, 13)})
	if err := z2.Take(0, 0, 14, []int{0, 1}, nil); err != nil {
		t.Fatalf("Take with a king as 14: %v", err)
	}
}

func TestZwickerTakeRejections(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 7), zwCard(CardDesignClover, 9)})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 7)})

	if err := z.Take(1, 0, 7, []int{0}, nil); err == nil {
		t.Error("expected an error when it is not that player's turn")
	}
	if err := z.Take(0, 9, 7, []int{0}, nil); err == nil {
		t.Error("expected an error for an out-of-range hand index")
	}
	// 数札は 1 つの値しか持たない。
	if err := z.Take(0, 0, 8, []int{0}, nil); err == nil {
		t.Error("expected an error playing a seven as an eight")
	}
	if err := z.Take(0, 0, 7, nil, nil); err == nil {
		t.Error("expected an error capturing nothing")
	}
	if err := z.Take(0, 0, 7, []int{0, 0}, nil); err == nil {
		t.Error("expected an error for a repeated table index")
	}
	if err := z.Take(0, 0, 7, []int{5}, nil); err == nil {
		t.Error("expected an error for an out-of-range table index")
	}
	if err := z.Take(0, 0, 7, []int{1}, nil); err == nil {
		t.Error("expected an error taking a nine with a seven")
	}
	if got := len(z.GetTableCards()); got != 2 {
		t.Errorf("table = %d, want it untouched at 2", got)
	}
}

func TestZwickerBuild(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 4)})
	// ♠5 と場の ♥4 で 9 を組み、9 で取れる ♦9 を残す。
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 5), zwCard(CardDesignDiamond, 9)})

	if err := z.Build(0, 0, []int{0}, 9); err != nil {
		t.Fatalf("Build: %v", err)
	}
	builds := z.GetBuilds()
	if len(builds) != 1 {
		t.Fatalf("builds = %d, want 1", len(builds))
	}
	if builds[0].Value != 9 || builds[0].Owner != 0 || len(builds[0].Cards) != 2 {
		t.Errorf("build = %+v, want value 9 owned by seat 0 with two cards", builds[0])
	}
	if len(z.GetTableCards()) != 0 {
		t.Error("the table card should have gone into the build")
	}
}

// TestZwickerBuildNeedsAMatchingCardInHand は、取れない値のビルドを作らせない
// ことを確かめる。作れてしまうと相手に献上するだけになる。
func TestZwickerBuildNeedsAMatchingCardInHand(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 4)})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 5), zwCard(CardDesignDiamond, 2)})

	if err := z.Build(0, 0, []int{0}, 9); err == nil {
		t.Error("expected an error building a 9 with no 9 left in hand")
	}
	if len(z.GetBuilds()) != 0 {
		t.Error("no build should have been made")
	}
}

func TestZwickerBuildRejections(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 4)})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 5), zwCard(CardDesignDiamond, 9)})

	if err := z.Build(0, 0, nil, 9); err == nil {
		t.Error("expected an error building on nothing")
	}
	if err := z.Build(0, 0, []int{0}, 0); err == nil {
		t.Error("expected an error for a non-positive build value")
	}
	if err := z.Build(0, 0, []int{0}, 12); err == nil {
		t.Error("expected an error when the cards do not add to the declared value")
	}
	if err := z.Build(0, 0, []int{7}, 9); err == nil {
		t.Error("expected an error for an out-of-range table index")
	}
}

// TestZwickerTakeABuild は、ビルドは宣言値ちょうどでしか取れないことを確かめる。
func TestZwickerTakeABuild(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 4)})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 5), zwCard(CardDesignDiamond, 9)})
	if err := z.Build(0, 0, []int{0}, 9); err != nil {
		t.Fatalf("Build: %v", err)
	}

	z.SetCurrentPlayerForTest(0)
	setZwHand(z, 0, []*Card{zwCard(CardDesignDiamond, 9), zwCard(CardDesignSpade, 8)})
	if err := z.Take(0, 1, 8, nil, []int{0}); err == nil {
		t.Error("expected an error taking a 9-build with an 8")
	}
	if err := z.Take(0, 0, 9, nil, []int{0}); err != nil {
		t.Fatalf("Take the build: %v", err)
	}
	if len(z.GetBuilds()) != 0 {
		t.Error("the build should be gone")
	}
	if got := len(z.GetPlayer(0).GetCaptured()); got != 3 {
		t.Errorf("captured = %d, want 3 (two build cards plus the played one)", got)
	}
	if got := z.GetPlayer(0).GetZwicks(); got != 1 {
		t.Errorf("zwicks = %d, want 1: the table is empty", got)
	}
}

func TestZwickerTrail(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest(nil)
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 5)})
	for i := 1; i < ZwickerPlayerCnt; i++ {
		setZwHand(z, i, []*Card{zwCard(CardDesignHeart, 2)})
	}

	if err := z.Trail(0, 0); err != nil {
		t.Fatalf("Trail: %v", err)
	}
	if got := len(z.GetTableCards()); got != 1 {
		t.Errorf("table = %d, want 1", got)
	}
	if got := z.GetCurrentPlayerIdx(); got != 1 {
		t.Errorf("current = %d, want 1", got)
	}
	if err := z.Trail(0, 0); err == nil {
		t.Error("expected an error playing out of turn")
	}
}

// TestZwickerDealsAgainWhenHandsRunOut は 3 段階の配り直しを確かめる。
func TestZwickerDealsAgainWhenHandsRunOut(t *testing.T) {
	z := NewDefaultZwicker()
	z.Reset()
	for i := range ZwickerPlayerCnt {
		setZwHand(z, i, []*Card{zwCard(CardDesignHeart, 2)})
	}
	z.SetCurrentPlayerForTest(0)
	for i := range ZwickerPlayerCnt {
		if err := z.Trail(i, 0); err != nil {
			t.Fatalf("Trail seat %d: %v", i, err)
		}
	}
	if z.GetPhase() != ZwickerPhasePlay {
		t.Fatalf("phase = %v, want Play: there are two more deals", z.GetPhase())
	}
	for i, p := range z.GetPlayers() {
		if p.GetCardsSize() == 0 {
			t.Errorf("seat %d was not dealt again", i)
		}
	}
}

// TestZwickerScoresThirtyPointsPerDeal は、1 ディールで配られる得点札が
// ちょうど 30 点ぶん精算されることを確かめる。
func TestZwickerScoresThirtyPointsPerDeal(t *testing.T) {
	z := NewDefaultZwicker()
	z.Reset()
	z.SetDealStageForTest(len(zwickerDealSizes))
	z.SetTableCardsForTest(nil)

	// 全 55 枚を席 0 と席 1 に分け、Zwick は与えない。
	all := make([]*Card, 0, 55)
	for _, d := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for v := 1; v <= 13; v++ {
			all = append(all, zwCard(d, v))
		}
	}
	all = append(all, zwJoker(1), zwJoker(2), zwJoker(3))
	z.GetPlayer(0).AddCaptured(all[:40])
	z.GetPlayer(1).AddCaptured(all[40:])
	for i := range ZwickerPlayerCnt {
		z.GetPlayer(i).Reset()
	}

	z.finishRound()
	got := z.GetLastRoundScore()
	if got == nil {
		t.Fatal("the deal should have been scored")
	}
	if total := got.Total[0] + got.Total[1]; total != 30 {
		t.Fatalf("the deal paid out %d, want exactly 30", total)
	}
	if got.MajorityTeam != 0 {
		t.Errorf("majority = %d, want team 0 (40 cards to 15)", got.MajorityTeam)
	}
	if got.Cards[0]+got.Cards[1] != 55 {
		t.Errorf("counted %d cards, want 55", got.Cards[0]+got.Cards[1])
	}
}

// TestZwickerMajorityIsUnawardedOnATie は枚数が並んだときに 3 点が誰にも
// 行かないことを確かめる。片方に倒すと合計が 30 を超える。
func TestZwickerMajorityIsUnawardedOnATie(t *testing.T) {
	z := NewDefaultZwicker()
	z.Reset()
	z.SetDealStageForTest(len(zwickerDealSizes))
	z.SetTableCardsForTest(nil)
	for i := range ZwickerPlayerCnt {
		z.GetPlayer(i).Reset()
	}
	z.GetPlayer(0).AddCaptured([]*Card{zwCard(CardDesignHeart, 3), zwCard(CardDesignHeart, 4)})
	z.GetPlayer(1).AddCaptured([]*Card{zwCard(CardDesignClover, 3), zwCard(CardDesignClover, 4)})

	z.finishRound()
	got := z.GetLastRoundScore()
	if got.MajorityTeam != -1 {
		t.Errorf("majority = %d, want -1 on a tie", got.MajorityTeam)
	}
	if got.Total[0] != 0 || got.Total[1] != 0 {
		t.Errorf("totals = %v, want no points at all", got.Total)
	}
}

// TestZwickerLeftoversGoToTheLastCapture pins the convention documented in
// finishRound -- it moves the three-point majority bonus, so it must not be
// silent.
func TestZwickerLeftoversGoToTheLastCapture(t *testing.T) {
	z := zwReady(t, 0)
	z.SetDealStageForTest(len(zwickerDealSizes))
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 7)})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 7)})
	for i := 1; i < ZwickerPlayerCnt; i++ {
		z.GetPlayer(i).Reset()
	}

	if err := z.Take(0, 0, 7, []int{0}, nil); err != nil {
		t.Fatalf("Take: %v", err)
	}
	// 手札が尽きて精算される。場は空なので残り物はない。
	if z.GetPhase() != ZwickerPhaseRoundEnd {
		t.Fatalf("phase = %v, want RoundEnd", z.GetPhase())
	}

	// 今度は場に札を残して精算する。
	z2 := zwReady(t, 0)
	z2.SetDealStageForTest(len(zwickerDealSizes))
	z2.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 7), zwCard(CardDesignDiamond, 10)})
	setZwHand(z2, 0, []*Card{zwCard(CardDesignSpade, 7)})
	for i := 1; i < ZwickerPlayerCnt; i++ {
		z2.GetPlayer(i).Reset()
	}
	if err := z2.Take(0, 0, 7, []int{0}, nil); err != nil {
		t.Fatalf("Take: %v", err)
	}
	// ♦10 は 3 点。最後に取った席 0 に流れる。
	if got := z2.GetLastRoundScore().CardPoints[0]; got != ZwickerScoreDiamondTen {
		t.Errorf("team 0 card points = %d, want %d from the leftover", got, ZwickerScoreDiamondTen)
	}
}

func TestZwickerNextRoundAndGameEnd(t *testing.T) {
	z := NewDefaultZwicker()
	z.Reset()
	if err := z.NextRound(); err == nil {
		t.Error("expected an error while the deal is still live")
	}

	z.SetPhaseForTest(ZwickerPhaseRoundEnd)
	if err := z.NextRound(); err != nil {
		t.Fatalf("NextRound: %v", err)
	}
	if z.GetPhase() != ZwickerPhasePlay {
		t.Errorf("phase = %v, want Play", z.GetPhase())
	}
	// ディーラーが 1 つ進むので、先手も 1 つ進む。
	if got := z.GetCurrentPlayerIdx(); got != 2 {
		t.Errorf("current = %d, want 2", got)
	}

	z.SetTeamScoreForTest(1, z.GetConfig().TargetScore)
	z.SetPhaseForTest(ZwickerPhaseRoundEnd)
	z.checkGameEnd()
	if !z.GetGameEndFlag() || z.GetWinnerTeam() != 1 {
		t.Errorf("gameEnd=%v winner=%d, want true/1", z.GetGameEndFlag(), z.GetWinnerTeam())
	}
	if err := z.NextRound(); err == nil {
		t.Error("expected an error dealing after the game is over")
	}
	if err := z.Trail(0, 0); err == nil {
		t.Error("expected an error playing after the game is over")
	}
}

// TestZwickerGameDoesNotEndOnASimultaneousTie は、両チームが同時に目標点へ
// 達し、しかも同点なら決着させないことを確かめる。
func TestZwickerGameDoesNotEndOnASimultaneousTie(t *testing.T) {
	z := NewDefaultZwicker()
	z.Reset()
	target := z.GetConfig().TargetScore
	z.SetTeamScoreForTest(0, target)
	z.SetTeamScoreForTest(1, target)
	z.checkGameEnd()
	if z.GetGameEndFlag() {
		t.Error("a tie at the target must not decide the game")
	}

	z.SetTeamScoreForTest(0, target+1)
	z.checkGameEnd()
	if !z.GetGameEndFlag() || z.GetWinnerTeam() != 0 {
		t.Errorf("gameEnd=%v winner=%d, want true/0", z.GetGameEndFlag(), z.GetWinnerTeam())
	}
}

func TestZwickerCpuDecide(t *testing.T) {
	z := zwReady(t, 1)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignDiamond, 10), zwCard(CardDesignHeart, 3)})
	setZwHand(z, 1, []*Card{zwCard(CardDesignSpade, 10), zwCard(CardDesignClover, 6)})

	act := z.ZwickerCpuDecide(1)
	if act.Type != "take" {
		t.Fatalf("the CPU should take the scoring ten, got %+v", act)
	}
	if err := z.Take(1, act.HandIdx, act.Value, act.TableIdxs, nil); err != nil {
		t.Fatalf("the CPU's capture was rejected: %v", err)
	}
}

func TestZwickerCpuTrailsWhenItCannotTake(t *testing.T) {
	z := zwReady(t, 1)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 9)})
	setZwHand(z, 1, []*Card{zwCard(CardDesignSpade, 1), zwCard(CardDesignClover, 5)})

	act := z.ZwickerCpuDecide(1)
	if act.Type != "trail" {
		t.Fatalf("nothing is takeable, got %+v", act)
	}
	// **エースは 1 点。**捨てるなら点にならない ♣5 のほう。
	if act.HandIdx != 1 {
		t.Errorf("HandIdx = %d, want 1 (keep the ace)", act.HandIdx)
	}
	if err := z.Trail(1, act.HandIdx); err != nil {
		t.Fatalf("the CPU's trail was rejected: %v", err)
	}
}

func TestZwickerCpuDecideEmptyHand(t *testing.T) {
	z := zwReady(t, 1)
	setZwHand(z, 1, nil)
	if act := z.ZwickerCpuDecide(1); act.HandIdx != -1 {
		t.Errorf("HandIdx = %d, want -1 for an empty hand", act.HandIdx)
	}
	if z.ZwickerCpuDecide(9).HandIdx != -1 {
		t.Error("an unknown seat must not produce a move")
	}
}

// TestZwickerCpuDrivesDealsToAnEnd checks that four CPUs always finish a deal.
func TestZwickerCpuDrivesDealsToAnEnd(t *testing.T) {
	for trial := range 20 {
		z := NewDefaultZwicker()
		z.Reset()
		for range 500 {
			if z.GetPhase() != ZwickerPhasePlay {
				break
			}
			cur := z.GetCurrentPlayerIdx()
			act := z.ZwickerCpuDecide(cur)
			if act.HandIdx < 0 {
				t.Fatalf("trial %d: seat %d had nothing to do", trial, cur)
			}
			var err error
			if act.Type == "take" {
				err = z.Take(cur, act.HandIdx, act.Value, act.TableIdxs, nil)
			} else {
				err = z.Trail(cur, act.HandIdx)
			}
			if err != nil {
				t.Fatalf("trial %d: %v", trial, err)
			}
		}
		if z.GetPhase() != ZwickerPhaseRoundEnd && z.GetPhase() != ZwickerPhaseGameEnd {
			t.Fatalf("trial %d: the deal never ended (phase = %v)", trial, z.GetPhase())
		}
		if total := z.GetLastRoundScore().Total[0] + z.GetLastRoundScore().Total[1]; total < 30 {
			t.Fatalf("trial %d: the deal paid out %d, want at least the 30 card points", trial, total)
		}
	}
}

func TestZwickerActionLog(t *testing.T) {
	z := NewDefaultZwicker()
	z.Reset()
	if len(z.GetActionLog()) == 0 {
		t.Fatal("the deal should be logged")
	}
	cur := z.GetCurrentPlayerIdx()
	if err := z.Trail(cur, 0); err != nil {
		t.Fatalf("Trail: %v", err)
	}
	log := z.GetActionLog()
	last := log[len(log)-1]
	if last.ActionType != "trail" || last.PlayerIdx != cur {
		t.Errorf("last entry = %+v, want a trail by seat %d", last, cur)
	}
	if last.TurnNumber != len(log) {
		t.Errorf("TurnNumber = %d, want %d", last.TurnNumber, len(log))
	}
}

func TestZwickerConfigValidate(t *testing.T) {
	if err := DefaultZwickerConfig().Validate(); err != nil {
		t.Fatalf("the default config should validate: %v", err)
	}
	if err := (ZwickerConfig{CpuDifficulty: 5, TargetScore: 61}).Validate(); err == nil {
		t.Error("expected an error for an unknown difficulty")
	}
	if err := (ZwickerConfig{TargetScore: 0}).Validate(); err == nil {
		t.Error("expected an error for a zero target")
	}
	z := NewDefaultZwicker()
	z.SetConfig(ZwickerConfig{TargetScore: 30})
	if z.GetConfig().TargetScore != 30 {
		t.Error("SetConfig/GetConfig disagree")
	}
}

func TestZwickerAccessorBounds(t *testing.T) {
	z := NewDefaultZwicker()
	z.Reset()
	if z.GetPlayer(-1) != nil || z.GetPlayer(ZwickerPlayerCnt) != nil {
		t.Error("GetPlayer should return nil outside the table")
	}
	if z.GetTeamScore(-1) != 0 || z.GetTeamScore(2) != 0 {
		t.Error("GetTeamScore should return 0 outside the two teams")
	}
}

func TestZwickerJSONRoundTrip(t *testing.T) {
	z := zwReady(t, 0)
	z.SetTableCardsForTest([]*Card{zwCard(CardDesignHeart, 4)})
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 5), zwCard(CardDesignDiamond, 9)})
	if err := z.Build(0, 0, []int{0}, 9); err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := json.Marshal(z)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Zwicker
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got.GetBuilds()) != 1 || got.GetBuilds()[0].Value != 9 {
		t.Errorf("builds = %+v, want the 9-build back", got.GetBuilds())
	}
	if got.GetStockCount() != z.GetStockCount() {
		t.Errorf("stock = %d, want %d", got.GetStockCount(), z.GetStockCount())
	}
	if got.GetCurrentPlayerIdx() != z.GetCurrentPlayerIdx() {
		t.Errorf("current = %d, want %d", got.GetCurrentPlayerIdx(), z.GetCurrentPlayerIdx())
	}
	if len(got.GetActionLog()) != len(z.GetActionLog()) {
		t.Errorf("action log = %d entries, want %d", len(got.GetActionLog()), len(z.GetActionLog()))
	}
}

func TestZwickerUnmarshalRejectsGarbage(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"not json", `{`},
		{"wrong player count", `{"pl":[null,null],"cfg":{"cd":0,"ts":61},"ph":0}`},
		{"bad config", `{"pl":[{},{},{},{}],"cfg":{"cd":9,"ts":61},"ph":0}`},
		{"unknown phase", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"ts":61},"ph":9}`},
		// **dealStage が範囲外だと配り直しが走り続ける。**
		{"deal stage out of range", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"ts":61},"ph":0,"st":9}`},
		{"negative deal stage", `{"pl":[{},{},{},{}],"cfg":{"cd":0,"ts":61},"ph":0,"st":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var z Zwicker
			if err := json.Unmarshal([]byte(tt.data), &z); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestZwickerUnmarshalClampsIndices(t *testing.T) {
	var z Zwicker
	data := `{"pl":[{},{},{},{}],"cfg":{"cd":0,"ts":61},"ph":0,"cur":99,"dl":-3,"lc":42,"wt":9,` +
		`"bd":[null,{"Owner":0,"Value":0,"Cards":[{}]},{"Owner":99,"Value":9,"Cards":[{}]},` +
		`{"Owner":1,"Value":9,"Cards":[{}]}]}`
	if err := json.Unmarshal([]byte(data), &z); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := z.GetCurrentPlayerIdx(); got != 0 {
		t.Errorf("current = %d, want it clamped to 0", got)
	}
	if got := z.GetWinnerTeam(); got != -1 {
		t.Errorf("winnerTeam = %d, want -1", got)
	}
	if got := len(z.GetBuilds()); got != 1 {
		t.Errorf("builds = %d, want the malformed ones dropped", got)
	}
	if z.GetStockCount() == 0 {
		t.Error("a missing pack should be rebuilt, not left nil")
	}
}

// TestZwickerCaptureSearchIsBounded は、細工した大きな tableIndices が探索を
// 爆発させないことを確かめる。/zwicker/exec からは任意の添字集合を送れるので、
// 素直な指数探索だと Worker のリクエスト時間を食い潰せてしまう。
//
// **意図的に「ほぼ割り切れる」形にしてある。**41 枚の A (1 として使える) を 4
// ずつの組に分けようとすると、10 組できて 1 枚余る。どの組み合わせを試しても
// 最後に必ず失敗するので、枝刈りだけでは全通りを舐めることになる。
func TestZwickerCaptureSearchIsBounded(t *testing.T) {
	z := zwReady(t, 0)
	table := make([]*Card, 0, 41)
	for range 41 {
		table = append(table, zwCard(CardDesignHeart, 1))
	}
	z.SetTableCardsForTest(table)
	setZwHand(z, 0, []*Card{zwCard(CardDesignSpade, 4)})

	idxs := make([]int, 0, len(table))
	for i := range table {
		idxs = append(idxs, i)
	}
	done := make(chan error, 1)
	go func() { done <- z.Take(0, 0, 4, idxs, nil) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("41 cards cannot split into groups of four; it must be rejected")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the capture search did not terminate; the budget is not being applied")
	}
	if got := len(z.GetTableCards()); got != len(table) {
		t.Errorf("table = %d, want it untouched at %d", got, len(table))
	}
}
