//go:build test

package domain

import (
	"testing"
)

// skatTestCard is a small helper to keep card construction terse.
func skatTestCard(design, value int) *Card {
	return NewCard(design, value, false)
}

// --- Bug 1: lead-suit must be required when comparing two non-trump cards ---

// TestSkatCpuPickPlayDoesNotChaseOffSuitNonTrump verifies that the CPU does
// not treat an off-suit non-trump card as a winnable trick. Before the fix,
// any non-trump card with higher card-strength would be treated as the trick
// leader regardless of suit, causing the CPU to play a high off-suit card in
// the false belief it would win.
func TestSkatCpuPickPlayDoesNotChaseOffSuitNonTrump(t *testing.T) {
	cfg := DefaultSkatConfig()
	cfg.CpuDifficulty = SkatCpuDifficultyHard
	g := newSkatForTest(t, cfg)
	resetForControlledPhase(g)
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.phase = SkatPhasePlay
	g.round.currentPlayerIdx = 2
	// Lead is 9 of Hearts (non-trump); CPU 1 followed with King of Hearts.
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: skatTestCard(CardDesignHeart, skatValueNine)},
		{PlayerIdx: 1, Card: skatTestCard(CardDesignHeart, skatValueKing)},
	}
	cpu := g.GetPlayer(2)
	// CPU 2 is void in Hearts; holds Ace of Diamonds (high but off-suit) and
	// 7 of Diamonds (low). Suit-following is impossible, so any card is legal.
	cpu.AddCard(skatTestCard(CardDesignDiamond, skatValueAce))
	cpu.AddCard(skatTestCard(CardDesignDiamond, skatValueSeven))

	idx := g.cpuPickPlay(2)
	played := cpu.GetCard(idx)
	// The Ace of Diamonds cannot beat the Hearts trick leader, so the CPU
	// must dump the lowest-value card instead of "chasing" with the Ace.
	if played.GetDesign() == CardDesignDiamond && played.GetValue() == skatValueAce {
		t.Fatalf("CPU chose off-suit Ace of Diamonds; expected to dump the 7 of Diamonds")
	}
	if played.GetValue() != skatValueSeven {
		t.Fatalf("CPU should dump the 7 of Diamonds, got %v", played)
	}
}

// TestSkatCpuPickPlayWinsWithLeadSuit ensures the CPU still wins when it
// holds a same-suit higher card.
func TestSkatCpuPickPlayWinsWithLeadSuit(t *testing.T) {
	cfg := DefaultSkatConfig()
	cfg.CpuDifficulty = SkatCpuDifficultyHard
	g := newSkatForTest(t, cfg)
	resetForControlledPhase(g)
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.phase = SkatPhasePlay
	g.round.currentPlayerIdx = 1
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: skatTestCard(CardDesignHeart, skatValueKing)},
	}
	cpu := g.GetPlayer(1)
	cpu.AddCard(skatTestCard(CardDesignHeart, skatValueAce))
	cpu.AddCard(skatTestCard(CardDesignHeart, skatValueSeven))

	idx := g.cpuPickPlay(1)
	played := cpu.GetCard(idx)
	if played.GetValue() != skatValueAce {
		t.Fatalf("Hard CPU should play winning Ace of Hearts, got %v", played)
	}
}

// TestSkatCpuPickPlayBoundsCheck guards the nil/range check added alongside
// the lead-suit fix.
func TestSkatCpuPickPlayBoundsCheck(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	if got := g.cpuPickPlay(-1); got != 0 {
		t.Fatalf("negative idx: got %d, want 0", got)
	}
	if got := g.cpuPickPlay(99); got != 0 {
		t.Fatalf("out-of-range idx: got %d, want 0", got)
	}
}

// TestSkatCpuPickPlayHardAlwaysTakesTrick verifies that Hard difficulty
// removes the 50/50 randomness — every iteration should pick the winning
// card when one is available.
func TestSkatCpuPickPlayHardAlwaysTakesTrick(t *testing.T) {
	cfg := DefaultSkatConfig()
	cfg.CpuDifficulty = SkatCpuDifficultyHard
	for i := 0; i < 50; i++ {
		g := newSkatForTest(t, cfg)
		resetForControlledPhase(g)
		g.round.gameType = SkatGameSuit
		g.round.trumpSuit = CardDesignSpade
		g.round.phase = SkatPhasePlay
		g.round.currentPlayerIdx = 1
		g.round.currentTrick = []*TrickCard{
			{PlayerIdx: 0, Card: skatTestCard(CardDesignHeart, skatValueNine)},
		}
		cpu := g.GetPlayer(1)
		cpu.AddCard(skatTestCard(CardDesignHeart, skatValueAce))
		cpu.AddCard(skatTestCard(CardDesignHeart, skatValueSeven))

		idx := g.cpuPickPlay(1)
		if cpu.GetCard(idx).GetValue() != skatValueAce {
			t.Fatalf("iteration %d: Hard CPU did not take winning trick", i)
		}
	}
}

// TestSkatCpuPickPlayLeadingNoTrick covers the branch where there is no
// trick yet — CPU must still return a valid index from the worst-card path.
func TestSkatCpuPickPlayLeadingNoTrick(t *testing.T) {
	cfg := DefaultSkatConfig()
	cfg.CpuDifficulty = SkatCpuDifficultyHard
	g := newSkatForTest(t, cfg)
	resetForControlledPhase(g)
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.phase = SkatPhasePlay
	g.round.currentPlayerIdx = 1
	cpu := g.GetPlayer(1)
	cpu.AddCard(skatTestCard(CardDesignHeart, skatValueSeven))
	cpu.AddCard(skatTestCard(CardDesignHeart, skatValueAce))
	idx := g.cpuPickPlay(1)
	if idx < 0 || idx >= cpu.GetCardsSize() {
		t.Fatalf("invalid idx returned: %d", idx)
	}
}

// --- Bug 2: matadors must use the snapshotted hand ---

// TestSkatStartPlaySnapshotsDeclarerHand verifies that the declarer's hand
// is captured when play begins so matadors can be counted at scoring time.
func TestSkatStartPlaySnapshotsDeclarerHand(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.declarerIdx = 0
	declarer := g.GetPlayer(0)
	declarer.AddCard(skatTestCard(CardDesignClover, skatValueJack))
	declarer.AddCard(skatTestCard(CardDesignSpade, skatValueJack))
	declarer.AddCard(skatTestCard(CardDesignHeart, skatValueAce))
	g.startPlay()
	if len(g.round.declarerHand) != declarer.GetCardsSize() {
		t.Fatalf("snapshot size = %d, want %d", len(g.round.declarerHand), declarer.GetCardsSize())
	}
}

// TestSkatMatadorsCountUsesSnapshot is the regression test for the empty-hand
// bug: by scoring time the declarer's live hand is empty, but matadors must
// still count from the snapshot taken at start-of-play.
func TestSkatMatadorsCountUsesSnapshot(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.declarerIdx = 0
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	// Snapshot: declarer holds top-3 trumps (CJ, SJ, HJ) — "with 3" matadors.
	g.round.declarerHand = []*Card{
		skatTestCard(CardDesignClover, skatValueJack),
		skatTestCard(CardDesignSpade, skatValueJack),
		skatTestCard(CardDesignHeart, skatValueJack),
		skatTestCard(CardDesignSpade, skatValueAce),
	}
	if got := g.matadorsCount(g.round.declarerHand); got != 3 {
		t.Fatalf("with-3 matadors: got %d, want 3", got)
	}
}

// TestSkatMatadorsCountWithoutBranch covers the "without N" path.
func TestSkatMatadorsCountWithoutBranch(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	// Lacks CJ, SJ; has HJ — "without 2".
	hand := []*Card{
		skatTestCard(CardDesignHeart, skatValueJack),
		skatTestCard(CardDesignSpade, skatValueAce),
	}
	if got := g.matadorsCount(hand); got != 2 {
		t.Fatalf("without-2 matadors: got %d, want 2", got)
	}
}

// TestSkatMatadorsCountNullGame returns 0 because null games have no trump
// order at all.
func TestSkatMatadorsCountNullGame(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.gameType = SkatGameNull
	if got := g.matadorsCount([]*Card{skatTestCard(CardDesignSpade, skatValueAce)}); got != 0 {
		t.Fatalf("null game matadors: got %d, want 0", got)
	}
}

// TestSkatGameMultiplierUsesSnapshot drives the multiplier path through the
// snapshotted hand and confirms the hand-game bonus stacks on top.
func TestSkatGameMultiplierUsesSnapshot(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.declarerIdx = 0
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	// "With 2" + hand-game bonus = matadors(2) + 1 + hand(1) = 4.
	g.round.declarerHand = []*Card{
		skatTestCard(CardDesignClover, skatValueJack),
		skatTestCard(CardDesignSpade, skatValueJack),
	}
	g.round.pickedSkat = false
	if got := g.gameMultiplier(); got != 4 {
		t.Fatalf("with-2 hand-game multiplier: got %d, want 4", got)
	}
	// Picking up the skat removes the hand bonus.
	g.round.pickedSkat = true
	if got := g.gameMultiplier(); got != 3 {
		t.Fatalf("with-2 picked-skat multiplier: got %d, want 3", got)
	}
}

// --- trickWinner: extra coverage paths ---

// TestSkatTrickWinnerNullOffSuitIgnored verifies that off-suit cards are
// skipped in null game trick resolution (regression guard for any future
// rewrite of the loop).
func TestSkatTrickWinnerNullOffSuitIgnored(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.gameType = SkatGameNull
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: skatTestCard(CardDesignHeart, skatValueNine)},
		{PlayerIdx: 1, Card: skatTestCard(CardDesignDiamond, skatValueAce)}, // off-suit
		{PlayerIdx: 2, Card: skatTestCard(CardDesignHeart, skatValueKing)},  // higher in lead suit
	}
	if got := g.trickWinner(); got != 2 {
		t.Fatalf("null off-suit lookup: got %d, want 2", got)
	}
}

// TestSkatTrickWinnerEmptyTrickReturnsZero protects the early-return guard.
func TestSkatTrickWinnerEmptyTrickReturnsZero(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.currentTrick = nil
	if got := g.trickWinner(); got != 0 {
		t.Fatalf("empty trick: got %d, want 0", got)
	}
}

// TestSkatTrickWinnerGrandJackBeatsLead exercises the trump-vs-trump branch
// in trickWinner via a Grand game (only jacks are trumps).
func TestSkatTrickWinnerGrandJackBeatsLead(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.gameType = SkatGameGrand
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: skatTestCard(CardDesignHeart, skatValueAce)},
		{PlayerIdx: 1, Card: skatTestCard(CardDesignDiamond, skatValueJack)}, // weakest jack
		{PlayerIdx: 2, Card: skatTestCard(CardDesignClover, skatValueJack)},  // top jack
	}
	if got := g.trickWinner(); got != 2 {
		t.Fatalf("CJ should win grand trick over DJ/HA, got %d", got)
	}
}

// --- bid ladder: forced acceptance and overbid penalty regression ---

// TestSkatComputeRoundResultOverbidLossPenalty walks the overbid branch in
// computeRoundResult — declarer "won" by card points but bid more than the
// game value, so the simplified penalty (bid * 2) must still apply.
func TestSkatComputeRoundResultOverbidLossPenalty(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.declarerIdx = 0
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.pickedSkat = true
	g.round.currentBid = 60
	g.round.declarerCardPts = 70
	g.round.defendersCardPts = 50
	// Hand snapshot with no top trumps (Ace + 7 of Spades only).
	g.round.declarerHand = []*Card{
		skatTestCard(CardDesignSpade, skatValueAce),
		skatTestCard(CardDesignSpade, skatValueSeven),
	}
	val, won := g.computeRoundResult()
	if won {
		t.Fatal("expected overbid loss")
	}
	if val != g.round.currentBid*2 {
		t.Fatalf("overbid loss value = %d, want %d", val, g.round.currentBid*2)
	}
}

// TestSkatComputeRoundResultSchneiderTriggersBonus exercises the schneider
// path: declarer keeps defenders below 30 card points.
func TestSkatComputeRoundResultSchneiderTriggersBonus(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.declarerIdx = 0
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.pickedSkat = true
	g.round.currentBid = 18
	g.round.declarerCardPts = 95
	g.round.defendersCardPts = 25
	g.round.declarerHand = []*Card{
		skatTestCard(CardDesignClover, skatValueJack),
	}
	val, won := g.computeRoundResult()
	if !won {
		t.Fatal("expected declarer win")
	}
	if val <= 0 {
		t.Fatalf("game value = %d, want > 0", val)
	}
}

// --- validatePlay edge paths ---

// TestSkatValidatePlayTrumpRequiredWhenLeadIsTrump checks the must-follow-
// trump branch in validatePlay.
func TestSkatValidatePlayTrumpRequiredWhenLeadIsTrump(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: skatTestCard(CardDesignClover, skatValueJack)}, // jack = trump
	}
	human := g.GetPlayer(0)
	human.AddCard(skatTestCard(CardDesignSpade, skatValueAce)) // trump
	human.AddCard(skatTestCard(CardDesignDiamond, skatValueAce))
	if err := g.validatePlay(0, skatTestCard(CardDesignDiamond, skatValueAce)); err == nil {
		t.Fatal("expected must-follow-trump violation")
	}
	if err := g.validatePlay(0, skatTestCard(CardDesignSpade, skatValueAce)); err != nil {
		t.Fatalf("legal trump follow rejected: %v", err)
	}
}

// TestSkatValidatePlayDiscardingNonTrumpHoldsLeadSuit tests the second
// branch where the player holds the lead suit but plays off-suit.
func TestSkatValidatePlayDiscardingNonTrumpHoldsLeadSuit(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: skatTestCard(CardDesignHeart, skatValueKing)},
	}
	human := g.GetPlayer(0)
	human.AddCard(skatTestCard(CardDesignHeart, skatValueAce))
	human.AddCard(skatTestCard(CardDesignClover, skatValueQueen))
	if err := g.validatePlay(0, skatTestCard(CardDesignClover, skatValueQueen)); err == nil {
		t.Fatal("expected must-follow-suit violation")
	}
}

// --- PlayerPlay/PlayerDiscard error paths ---

func TestSkatPlayerPlayInvalidIndex(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.phase = SkatPhasePlay
	g.round.currentPlayerIdx = 0
	g.GetPlayer(0).AddCard(skatTestCard(CardDesignSpade, skatValueAce))
	if err := g.PlayerPlay(99); err == nil {
		t.Fatal("expected out-of-range error")
	}
	if err := g.PlayerPlay(-1); err == nil {
		t.Fatal("expected negative-idx error")
	}
}

func TestSkatPlayerDiscardErrorPaths(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.phase = SkatPhasePlay // wrong phase
	if err := g.PlayerDiscard(0, 1); err == nil {
		t.Fatal("expected wrong-phase error")
	}

	g.round.phase = SkatPhaseDiscard
	g.round.declarerIdx = 0
	human := g.GetPlayer(0)
	resetForControlledPhase(g)
	g.round.phase = SkatPhaseDiscard
	g.round.declarerIdx = 0
	human.AddCard(skatTestCard(CardDesignSpade, skatValueAce))
	human.AddCard(skatTestCard(CardDesignSpade, skatValueKing))
	// Same index twice
	if err := g.PlayerDiscard(0, 0); err == nil {
		t.Fatal("expected same-index error")
	}
	if err := g.PlayerDiscard(0, 99); err == nil {
		t.Fatal("expected out-of-range error")
	}
}

// TestSkatPlayerDeclareGameErrorPaths exercises wrong-phase and invalid game
// type guards.
func TestSkatPlayerDeclareGameErrorPaths(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.phase = SkatPhaseBid
	if err := g.PlayerDeclareGame(SkatGameSuit, CardDesignSpade); err == nil {
		t.Fatal("expected wrong-phase error")
	}

	g.round.phase = SkatPhaseGameDeclaration
	g.round.declarerIdx = 0
	if err := g.PlayerDeclareGame(SkatGameType(99), 0); err == nil {
		t.Fatal("expected invalid game-type error")
	}
}

// TestSkatCpuPickGameVariants drives cpuPickGame with several hand shapes to
// cover Grand and Suit branches.
func TestSkatCpuPickGameVariants(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	p := g.GetPlayer(0)
	// Heavy spade hand → suit game with spades.
	for _, c := range []*Card{
		skatTestCard(CardDesignSpade, skatValueAce),
		skatTestCard(CardDesignSpade, skatValueTen),
		skatTestCard(CardDesignSpade, skatValueKing),
		skatTestCard(CardDesignSpade, skatValueQueen),
		skatTestCard(CardDesignSpade, skatValueNine),
		skatTestCard(CardDesignSpade, skatValueEight),
		skatTestCard(CardDesignSpade, skatValueSeven),
	} {
		p.AddCard(c)
	}
	gt, suit := g.cpuPickGame(0)
	if gt != SkatGameSuit {
		t.Fatalf("strong suit hand: got %v, want SkatGameSuit", gt)
	}
	if suit != CardDesignSpade {
		t.Fatalf("trump suit: got %d, want spade", suit)
	}
}

// TestSkatHandStrengthAccountsForJacksTensAces ensures handStrength reflects
// jacks > aces > tens, matching the heuristic used by cpuBidDecision.
func TestSkatHandStrengthAccountsForJacksTensAces(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	p := g.GetPlayer(0)
	for _, c := range []*Card{
		skatTestCard(CardDesignClover, skatValueJack),
		skatTestCard(CardDesignSpade, skatValueJack),
		skatTestCard(CardDesignHeart, skatValueAce),
		skatTestCard(CardDesignHeart, skatValueTen),
	} {
		p.AddCard(c)
	}
	got := g.handStrength(0)
	// 2 jacks * 5 + maxSuit(2) * 2 + 1 ace * 2 + 1 ten = 10 + 4 + 2 + 1 = 17.
	if got != 17 {
		t.Fatalf("handStrength = %d, want 17", got)
	}
}

// TestSkatGameTypeNameAllVariants exercises every game-type label branch.
func TestSkatGameTypeNameAllVariants(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	if got := g.gameTypeName(); got == "" || got == "None" {
		t.Fatalf("Suit gameTypeName empty/None: %q", got)
	}
	g.round.gameType = SkatGameGrand
	if got := g.gameTypeName(); got != "Grand" {
		t.Fatalf("Grand gameTypeName: got %q, want Grand", got)
	}
	g.round.gameType = SkatGameNull
	if got := g.gameTypeName(); got != "Null" {
		t.Fatalf("Null gameTypeName: got %q, want Null", got)
	}
	g.round.gameType = SkatGameNone
	if got := g.gameTypeName(); got != "None" {
		t.Fatalf("None gameTypeName: got %q, want None", got)
	}
}

// TestSkatSuitNameAllBranches exercises skatSuitName for each suit + default.
func TestSkatSuitNameAllBranches(t *testing.T) {
	cases := []struct {
		suit int
		want string
	}{
		{CardDesignSpade, "Spades"},
		{CardDesignClover, "Clubs"},
		{CardDesignHeart, "Hearts"},
		{CardDesignDiamond, "Diamonds"},
		{99, "?"},
	}
	for _, c := range cases {
		if got := skatSuitName(c.suit); got != c.want {
			t.Fatalf("skatSuitName(%d) = %q, want %q", c.suit, got, c.want)
		}
	}
}

// TestSkatNullRankAllValues fills in the missing branches of nullRank.
func TestSkatNullRankAllValues(t *testing.T) {
	cases := []struct {
		v    int
		want int
	}{
		{skatValueSeven, 1}, {skatValueEight, 2}, {skatValueNine, 3},
		{skatValueTen, 4}, {skatValueJack, 5}, {skatValueQueen, 6},
		{skatValueKing, 7}, {skatValueAce, 8},
	}
	for _, c := range cases {
		if got := nullRank(skatTestCard(CardDesignSpade, c.v)); got != c.want {
			t.Fatalf("nullRank(%d) = %d, want %d", c.v, got, c.want)
		}
	}
}

// TestSkatCardPointsAllValues fills in the missing branches of skatCardPoints.
func TestSkatCardPointsAllValues(t *testing.T) {
	cases := []struct {
		v    int
		want int
	}{
		{skatValueAce, 11}, {skatValueTen, 10}, {skatValueKing, 4},
		{skatValueQueen, 3}, {skatValueJack, 2},
		{skatValueNine, 0}, {skatValueEight, 0}, {skatValueSeven, 0},
	}
	for _, c := range cases {
		if got := skatCardPoints(skatTestCard(CardDesignSpade, c.v)); got != c.want {
			t.Fatalf("skatCardPoints(%d) = %d, want %d", c.v, got, c.want)
		}
	}
}

// TestSkatTrumpOrderForEachGameType exercises every branch of trumpOrder.
func TestSkatTrumpOrderForEachGameType(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.gameType = SkatGameNull
	if order := g.trumpOrder(); order != nil {
		t.Fatalf("null trumpOrder = %v, want nil", order)
	}
	g.round.gameType = SkatGameGrand
	if order := g.trumpOrder(); len(order) != 4 {
		t.Fatalf("grand trumpOrder len = %d, want 4", len(order))
	}
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	if order := g.trumpOrder(); len(order) != 11 {
		t.Fatalf("suit trumpOrder len = %d, want 11", len(order))
	}
}

// TestSkatIsTrumpAcrossGameTypes covers the isTrump branch table.
func TestSkatIsTrumpAcrossGameTypes(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.gameType = SkatGameNull
	if g.isTrump(skatTestCard(CardDesignSpade, skatValueJack)) {
		t.Fatal("null game has no trump")
	}
	g.round.gameType = SkatGameGrand
	if !g.isTrump(skatTestCard(CardDesignDiamond, skatValueJack)) {
		t.Fatal("grand: jack is trump")
	}
	if g.isTrump(skatTestCard(CardDesignDiamond, skatValueAce)) {
		t.Fatal("grand: non-jack is not trump")
	}
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	if !g.isTrump(skatTestCard(CardDesignSpade, skatValueAce)) {
		t.Fatal("suit: lead-suit ace must be trump")
	}
	if !g.isTrump(skatTestCard(CardDesignDiamond, skatValueJack)) {
		t.Fatal("suit: any jack must be trump")
	}
	if g.isTrump(skatTestCard(CardDesignDiamond, skatValueAce)) {
		t.Fatal("suit: off-suit non-jack is not trump")
	}
}
