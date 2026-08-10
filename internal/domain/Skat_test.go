//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSkatForTest sets up a Skat game with the human at index 0. Reset()
// performs the deal so callers can immediately interact with the bidding
// phase.
func newSkatForTest(t *testing.T, cfg SkatConfig) *Skat {
	t.Helper()
	players := []*SkatPlayer{
		NewSkatPlayer(true),
		NewSkatPlayer(false),
		NewSkatPlayer(false),
	}
	g := NewSkat(newSkatDeck(), players, cfg)
	g.Reset()
	return g
}

// resetForControlledPhase clears the dealt hands so the test can set up an
// arbitrary game state.
func resetForControlledPhase(g *Skat) {
	for _, p := range g.players {
		p.Reset()
	}
}

func TestSkatResetSetsUpPositionsAndPhase(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	if g.GetPhase() != SkatPhaseBid {
		t.Fatalf("phase = %v, want Bid", g.GetPhase())
	}
	if g.GetForehandIdx() == g.GetMiddlehandIdx() || g.GetMiddlehandIdx() == g.GetRearhandIdx() {
		t.Fatalf("positions overlap: F=%d M=%d R=%d", g.GetForehandIdx(), g.GetMiddlehandIdx(), g.GetRearhandIdx())
	}
	for i := 0; i < SkatPlayerCnt; i++ {
		if g.GetPlayer(i).GetCardsSize() != SkatHandSize {
			t.Fatalf("player %d hand size = %d, want %d", i, g.GetPlayer(i).GetCardsSize(), SkatHandSize)
		}
	}
	if len(g.GetSkat()) != SkatSkatSize {
		t.Fatalf("skat size = %d, want %d", len(g.GetSkat()), SkatSkatSize)
	}
}

func TestSkatBidAllPassMakesForehandDeclarer(t *testing.T) {
	// In our simplified flow, when both middle and rear pass without calling,
	// forehand becomes declarer at 18 (the lowest legal bid). Fully-passed-out
	// hands are not modelled separately.
	g := newSkatForTest(t, DefaultSkatConfig())
	for g.GetActiveBidActorIdx() >= 0 && g.GetPhase() == SkatPhaseBid {
		idx := g.GetActiveBidActorIdx()
		g.applyBidStep(idx, false)
	}
	if g.GetPhase() != SkatPhaseSkatPickup {
		t.Fatalf("phase = %v, want SkatPickup", g.GetPhase())
	}
	if g.GetDeclarerIdx() != g.GetForehandIdx() {
		t.Fatalf("declarer = %d, want forehand %d", g.GetDeclarerIdx(), g.GetForehandIdx())
	}
	if g.GetCurrentBid() != SkatBidLadder[0] {
		t.Fatalf("current bid = %d, want %d", g.GetCurrentBid(), SkatBidLadder[0])
	}
}

func TestSkatBidRoundFlow(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	// Middle calls 18.
	g.applyBidStep(g.GetMiddlehandIdx(), true)
	if g.GetCurrentBid() != SkatBidLadder[0] {
		t.Fatalf("current bid = %d, want %d", g.GetCurrentBid(), SkatBidLadder[0])
	}
	// Forehand passes → middle is round 1 survivor; round 2 begins with rear vs middle.
	g.applyBidStep(g.GetForehandIdx(), false)
	if g.GetActiveBidActorIdx() != g.GetMiddlehandIdx() {
		t.Fatalf("expected middle to respond, got %d", g.GetActiveBidActorIdx())
	}
	// Middle passes (drops out) → rearhand is declarer.
	g.applyBidStep(g.GetMiddlehandIdx(), false)
	if g.GetDeclarerIdx() != g.GetRearhandIdx() {
		t.Fatalf("declarer = %d, want rearhand %d", g.GetDeclarerIdx(), g.GetRearhandIdx())
	}
	if g.GetPhase() != SkatPhaseSkatPickup {
		t.Fatalf("phase = %v, want SkatPickup", g.GetPhase())
	}
}

func TestSkatPlayerPickSkatWrongPhaseError(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	if err := g.PlayerPickSkat(true); err == nil {
		t.Fatal("expected error during bid phase")
	}
}

func TestSkatHumanFlowSuitGameWin(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	// Force the human to be declarer with a known strong hand.
	resetForControlledPhase(g)
	human := g.GetPlayer(0)
	cpu1 := g.GetPlayer(1)
	cpu2 := g.GetPlayer(2)

	// Strong human hand: 3 jacks + AS + TS + KS + QS + 9S + 8S + 7S
	humanCards := []*Card{
		NewCard(CardDesignClover, skatValueJack, false),
		NewCard(CardDesignSpade, skatValueJack, false),
		NewCard(CardDesignHeart, skatValueJack, false),
		NewCard(CardDesignSpade, skatValueAce, false),
		NewCard(CardDesignSpade, skatValueTen, false),
		NewCard(CardDesignSpade, skatValueKing, false),
		NewCard(CardDesignSpade, skatValueQueen, false),
		NewCard(CardDesignSpade, skatValueNine, false),
		NewCard(CardDesignSpade, skatValueEight, false),
		NewCard(CardDesignSpade, skatValueSeven, false),
	}
	for _, c := range humanCards {
		human.AddCard(c)
	}
	cpu1Cards := []*Card{
		NewCard(CardDesignDiamond, skatValueJack, false),
		NewCard(CardDesignHeart, skatValueAce, false),
		NewCard(CardDesignHeart, skatValueTen, false),
		NewCard(CardDesignHeart, skatValueKing, false),
		NewCard(CardDesignHeart, skatValueQueen, false),
		NewCard(CardDesignHeart, skatValueNine, false),
		NewCard(CardDesignHeart, skatValueEight, false),
		NewCard(CardDesignHeart, skatValueSeven, false),
		NewCard(CardDesignClover, skatValueAce, false),
		NewCard(CardDesignClover, skatValueTen, false),
	}
	for _, c := range cpu1Cards {
		cpu1.AddCard(c)
	}
	cpu2Cards := []*Card{
		NewCard(CardDesignClover, skatValueKing, false),
		NewCard(CardDesignClover, skatValueQueen, false),
		NewCard(CardDesignClover, skatValueNine, false),
		NewCard(CardDesignClover, skatValueEight, false),
		NewCard(CardDesignClover, skatValueSeven, false),
		NewCard(CardDesignDiamond, skatValueAce, false),
		NewCard(CardDesignDiamond, skatValueTen, false),
		NewCard(CardDesignDiamond, skatValueKing, false),
		NewCard(CardDesignDiamond, skatValueQueen, false),
		NewCard(CardDesignDiamond, skatValueNine, false),
	}
	for _, c := range cpu2Cards {
		cpu2.AddCard(c)
	}
	g.round.skat = []*Card{
		NewCard(CardDesignDiamond, skatValueEight, false),
		NewCard(CardDesignDiamond, skatValueSeven, false),
	}

	// Skip auction: declare human declarer manually.
	g.round.currentBid = 18
	g.declareDeclarer(0)
	if g.GetPhase() != SkatPhaseSkatPickup {
		t.Fatalf("phase = %v, want SkatPickup", g.GetPhase())
	}
	if err := g.PlayerPickSkat(true); err != nil {
		t.Fatalf("PlayerPickSkat: %v", err)
	}
	if human.GetCardsSize() != SkatHandSize+SkatSkatSize {
		t.Fatalf("hand size after pickup = %d, want %d", human.GetCardsSize(), SkatHandSize+SkatSkatSize)
	}

	// Discard the lowest two: 7D and 8D.
	idx7D, idx8D := -1, -1
	for i := 0; i < human.GetCardsSize(); i++ {
		c := human.GetCard(i)
		if c.GetDesign() == CardDesignDiamond && c.GetValue() == skatValueSeven {
			idx7D = i
		}
		if c.GetDesign() == CardDesignDiamond && c.GetValue() == skatValueEight {
			idx8D = i
		}
	}
	if idx7D < 0 || idx8D < 0 {
		t.Fatal("could not find 7D/8D in hand")
	}
	if err := g.PlayerDiscard(idx7D, idx8D); err != nil {
		t.Fatalf("PlayerDiscard: %v", err)
	}
	if human.GetCardsSize() != SkatHandSize {
		t.Fatalf("hand after discard = %d, want %d", human.GetCardsSize(), SkatHandSize)
	}
	if g.GetPhase() != SkatPhaseGameDeclaration {
		t.Fatalf("phase = %v, want GameDeclaration", g.GetPhase())
	}

	if err := g.PlayerDeclareGame(SkatGameSuit, CardDesignSpade); err != nil {
		t.Fatalf("DeclareGame: %v", err)
	}
	if g.GetPhase() != SkatPhasePlay {
		t.Fatalf("phase = %v, want Play", g.GetPhase())
	}
	if g.GetCurrentPlayerIdx() != g.GetForehandIdx() {
		t.Fatalf("first turn = %d, want forehand %d", g.GetCurrentPlayerIdx(), g.GetForehandIdx())
	}

	// Run all 10 tricks via CPU/human auto-play loop.
	for trick := 0; trick < SkatTricksPerRound; trick++ {
		for played := 0; played < SkatPlayerCnt; played++ {
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				if len(valid) == 0 {
					t.Fatal("no valid plays for human")
				}
				if err := g.PlayerPlay(valid[0]); err != nil {
					t.Fatalf("PlayerPlay: %v", err)
				}
			} else {
				g.CpuPlay()
			}
		}
		if g.GetPhase() != SkatPhaseTrickEnd {
			t.Fatalf("expected trick end after %d, got %v", trick, g.GetPhase())
		}
		g.ResolveTrick()
		if trick < SkatTricksPerRound-1 {
			g.NextTrick()
		}
	}

	if g.GetPhase() != SkatPhaseRoundEnd {
		t.Fatalf("phase after last trick = %v, want RoundEnd", g.GetPhase())
	}
	g.ScoreRound()
	if g.GetWinnerSide() != SkatWinnerDeclarer {
		t.Fatalf("declarer expected to win, got winnerSide=%d (declPts=%d)",
			g.GetWinnerSide(), g.GetDeclarerCardPoints())
	}
	if g.GetGameValue() <= 0 {
		t.Fatalf("game value = %d, want > 0", g.GetGameValue())
	}
}

func TestSkatNextRoundAdvancesDealer(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	prevDealer := g.GetDealerIdx()
	prevRound := g.GetRoundNumber()
	g.SetPhase(SkatPhaseRoundEnd)
	g.NextRound()
	if g.GetDealerIdx() != (prevDealer+1)%SkatPlayerCnt {
		t.Fatalf("dealer did not rotate: prev=%d cur=%d", prevDealer, g.GetDealerIdx())
	}
	if g.GetRoundNumber() != prevRound+1 {
		t.Fatalf("round = %d, want %d", g.GetRoundNumber(), prevRound+1)
	}
}

func TestSkatTrickWinnerSuit(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, skatValueAce, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, skatValueKing, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, skatValueSeven, false)}, // trump
	}
	if got := g.trickWinner(); got != 2 {
		t.Fatalf("trump should win: got %d", got)
	}

	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, skatValueAce, false)}, // lead trump
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, skatValueJack, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, skatValueTen, false)},
	}
	if got := g.trickWinner(); got != 1 {
		t.Fatalf("CJ (top trump) should win: got %d", got)
	}
}

func TestSkatTrickWinnerNull(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.gameType = SkatGameNull
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, skatValueSeven, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, skatValueAce, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, skatValueAce, false)},
	}
	if got := g.trickWinner(); got != 1 {
		t.Fatalf("HA should win null trick: got %d", got)
	}
}

func TestSkatValidatePlayMustFollowSuit(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.currentPlayerIdx = 0
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, skatValueKing, false)},
	}
	human := g.GetPlayer(0)
	human.AddCard(NewCard(CardDesignHeart, skatValueAce, false))
	human.AddCard(NewCard(CardDesignDiamond, skatValueSeven, false))

	if err := g.validatePlay(0, NewCard(CardDesignDiamond, skatValueSeven, false)); err == nil {
		t.Fatal("expected suit-follow violation")
	}
	if err := g.validatePlay(0, NewCard(CardDesignHeart, skatValueAce, false)); err != nil {
		t.Fatalf("legal follow rejected: %v", err)
	}
}

func TestSkatGameValueGrand(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	declarer := g.GetPlayer(0)
	declarer.AddCard(NewCard(CardDesignClover, skatValueJack, false))
	declarer.AddCard(NewCard(CardDesignSpade, skatValueJack, false))

	g.round.declarerIdx = 0
	g.round.gameType = SkatGameGrand
	g.round.pickedSkat = true
	g.round.currentBid = 24
	g.round.declarerCardPts = 75
	g.round.defendersCardPts = 45

	val, won := g.computeRoundResult()
	if !won {
		t.Fatalf("declarer should win at 75 pts")
	}
	if val <= 0 {
		t.Fatalf("game value = %d, want > 0", val)
	}
}

func TestSkatGameValueSuitLoss(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	declarer := g.GetPlayer(0)
	declarer.AddCard(NewCard(CardDesignClover, skatValueJack, false))
	g.round.declarerIdx = 0
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.pickedSkat = true
	g.round.currentBid = 18
	g.round.declarerCardPts = 40
	g.round.defendersCardPts = 80

	val, won := g.computeRoundResult()
	if won {
		t.Fatal("declarer should lose at 40 pts")
	}
	if val <= 0 {
		t.Fatalf("loss value = %d, want > 0", val)
	}
}

func TestSkatGameValueNullWinAndLoss(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.declarerIdx = 0
	g.round.gameType = SkatGameNull

	// Win: declarer takes no tricks.
	g.GetPlayer(0).ResetTricks()
	val, won := g.computeRoundResult()
	if !won || val <= 0 {
		t.Fatalf("expected null win, got won=%v val=%d", won, val)
	}

	// Loss: declarer takes a trick.
	g.GetPlayer(0).AddTrick([]*Card{NewCard(CardDesignSpade, skatValueAce, false)})
	val2, won2 := g.computeRoundResult()
	if won2 {
		t.Fatalf("expected null loss")
	}
	if val2 <= 0 {
		t.Fatalf("null loss value = %d, want > 0", val2)
	}
}

func TestSkatScoreRoundDeclarerWinsAccumulates(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	declarer := g.GetPlayer(0)
	declarer.AddCard(NewCard(CardDesignClover, skatValueJack, false))
	g.round.declarerIdx = 0
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.pickedSkat = true
	g.round.currentBid = 18
	g.round.skat = []*Card{NewCard(CardDesignDiamond, skatValueSeven, false), NewCard(CardDesignDiamond, skatValueEight, false)}
	declarer.SetCardPoints(75)
	g.round.phase = SkatPhaseRoundEnd
	g.ScoreRound()
	if g.GetWinnerSide() != SkatWinnerDeclarer {
		t.Fatalf("expected declarer win, got %d", g.GetWinnerSide())
	}
	if declarer.GetCumulativeScore() <= 0 {
		t.Fatalf("cumulative score not updated: %d", declarer.GetCumulativeScore())
	}
}

func TestSkatHumanInputValidation(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	g.round.phase = SkatPhaseGameEnd
	g.round.gameEndFlag = true
	if err := g.PlayerBid(true); err != ErrGameEnded {
		t.Fatalf("PlayerBid after game end err=%v, want ErrGameEnded", err)
	}
	if err := g.PlayerPickSkat(true); err != ErrGameEnded {
		t.Fatalf("PlayerPickSkat err=%v", err)
	}
	if err := g.PlayerDiscard(0, 1); err != ErrGameEnded {
		t.Fatalf("PlayerDiscard err=%v", err)
	}
	if err := g.PlayerDeclareGame(SkatGameSuit, CardDesignSpade); err != ErrGameEnded {
		t.Fatalf("PlayerDeclareGame err=%v", err)
	}
	if err := g.PlayerPlay(0); err != ErrGameEnded {
		t.Fatalf("PlayerPlay err=%v", err)
	}
}

func TestSkatHintReturnsActionableInfo(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	// Make human the active bidder.
	for g.GetActiveBidActorIdx() >= 0 && !g.GetPlayer(g.GetActiveBidActorIdx()).GetIsHuman() {
		g.applyBidStep(g.GetActiveBidActorIdx(), false)
		if g.GetPhase() != SkatPhaseBid {
			break
		}
	}
	if g.GetPhase() == SkatPhaseBid && g.IsHumanBidTurn() {
		hint := g.GetHint()
		if hint == nil {
			t.Fatal("expected bid hint")
		}
		if hint.Bid == nil {
			t.Fatal("expected bid value in hint")
		}
	}
}

func TestSkatCpuFlowAuto(t *testing.T) {
	cfg := DefaultSkatConfig()
	cfg.CpuDifficulty = SkatCpuDifficultyHard
	g := newSkatForTest(t, cfg)

	// Drive bidding: human always passes, CPUs use their own strategy.
	for g.GetPhase() == SkatPhaseBid {
		idx := g.GetActiveBidActorIdx()
		if idx < 0 {
			break
		}
		if g.GetPlayer(idx).GetIsHuman() {
			g.applyBidStep(idx, false)
			continue
		}
		g.CpuBid()
	}
	if g.GetDeclarerIdx() < 0 {
		t.Fatalf("expected a declarer, phase=%v", g.GetPhase())
	}

	// If the human is declarer, drive the post-auction phases manually.
	if g.GetPlayer(g.GetDeclarerIdx()).GetIsHuman() {
		if err := g.PlayerPickSkat(true); err != nil {
			t.Fatalf("PlayerPickSkat: %v", err)
		}
		if err := g.PlayerDiscard(0, 1); err != nil {
			t.Fatalf("PlayerDiscard: %v", err)
		}
		if err := g.PlayerDeclareGame(SkatGameSuit, CardDesignSpade); err != nil {
			t.Fatalf("DeclareGame: %v", err)
		}
	} else {
		if g.GetPhase() == SkatPhaseSkatPickup {
			g.CpuPickSkat()
		}
		if g.GetPhase() == SkatPhaseDiscard {
			g.CpuDiscard()
		}
		if g.GetPhase() == SkatPhaseGameDeclaration {
			g.CpuDeclareGame()
		}
	}
	if g.GetPhase() != SkatPhasePlay {
		t.Fatalf("expected Play after declaration, got %v", g.GetPhase())
	}
}

func TestSkatHandStrengthAndCpuPicks(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	p := g.GetPlayer(0)
	for _, c := range []*Card{
		NewCard(CardDesignClover, skatValueJack, false),
		NewCard(CardDesignSpade, skatValueJack, false),
		NewCard(CardDesignHeart, skatValueJack, false),
		NewCard(CardDesignSpade, skatValueAce, false),
	} {
		p.AddCard(c)
	}
	if got := g.handStrength(0); got <= 0 {
		t.Fatalf("hand strength = %d, want > 0", got)
	}
	gt, _ := g.cpuPickGame(0)
	if gt != SkatGameGrand {
		t.Fatalf("3 jacks should pick Grand, got %v", gt)
	}
}

func TestSkatNullRankAndCardPoints(t *testing.T) {
	if nullRank(NewCard(CardDesignSpade, skatValueAce, false)) != 8 {
		t.Fatal("ace should rank 8 in null")
	}
	if nullRank(NewCard(CardDesignSpade, skatValueSeven, false)) != 1 {
		t.Fatal("seven should rank 1 in null")
	}
	if skatCardPoints(NewCard(CardDesignSpade, skatValueAce, false)) != 11 {
		t.Fatal("ace = 11 points")
	}
	if skatCardPoints(NewCard(CardDesignSpade, skatValueSeven, false)) != 0 {
		t.Fatal("seven = 0 points")
	}
}

func TestSkatGetValidPlayIndicesNoFollow(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	resetForControlledPhase(g)
	g.round.gameType = SkatGameSuit
	g.round.trumpSuit = CardDesignSpade
	g.round.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, skatValueAce, false)},
	}
	p := g.GetPlayer(0)
	p.AddCard(NewCard(CardDesignDiamond, skatValueSeven, false))
	p.AddCard(NewCard(CardDesignSpade, skatValueAce, false)) // trump
	valid := g.GetValidPlayIndices(0)
	if len(valid) != 2 {
		t.Fatalf("no hearts → all cards legal, got %v", valid)
	}
}

func TestSkatGetterDefaults(t *testing.T) {
	g := newSkatForTest(t, DefaultSkatConfig())
	if g.GetPlayerCnt() != SkatPlayerCnt {
		t.Fatalf("player cnt = %d", g.GetPlayerCnt())
	}
	if g.GetPlayer(-1) != nil || g.GetPlayer(99) != nil {
		t.Fatal("out-of-range GetPlayer should return nil")
	}
	if g.GetConfig().TargetScore != 500 {
		t.Fatalf("config target = %d", g.GetConfig().TargetScore)
	}
	g.SetConfig(SkatConfig{CpuDifficulty: SkatCpuDifficultyHard, TargetScore: 100})
	if g.GetConfig().TargetScore != 100 {
		t.Fatal("SetConfig did not apply")
	}
	g.SetCurrentPlayerIdx(2)
	if g.GetCurrentPlayerIdx() != 2 {
		t.Fatal("SetCurrentPlayerIdx did not apply")
	}
}

// **CUI にも安全ビッド上限が要る。**Web は `skatBidEstimate.ts` で常時出すのに、
// Go 側には対応するものが無かった (#4905)。
func TestSkat_BidEstimates(t *testing.T) {
	card := func(d, v int) *Card { return NewCard(d, v, false) }

	// ♣J だけ持つ = with 1。どの契約でも matadors 1 → (1+1)×base。
	hand := []*Card{card(CardDesignClover, 11)}
	est := SkatBidEstimates(hand)
	require.Len(t, est, 5)
	for _, e := range est {
		assert.Equal(t, 1, e.Matadors, "holding only the top trump is with 1")
		assert.Equal(t, (e.Matadors+1)*e.Base, e.Value)
	}
	// グランドが基礎点 24 で最大。
	best := SkatBestBidEstimate(hand)
	assert.Equal(t, SkatGameGrand, best.GameType)
	assert.Equal(t, 24, best.Base)
	assert.Equal(t, 48, best.Value)

	// **♣J を持たない = without。**次に何枚欠けているかで数える。
	// ♠J だけの手札は ♣J が無いので without 1。
	without := []*Card{card(CardDesignSpade, 11)}
	for _, e := range SkatBidEstimates(without) {
		assert.Equal(t, 1, e.Matadors, "missing only the top trump is without 1")
	}

	// ♣J ♠J ♥J を持つ = with 3。♦J が無いのでそこで止まる。
	three := []*Card{
		card(CardDesignClover, 11), card(CardDesignSpade, 11), card(CardDesignHeart, 11),
	}
	assert.Equal(t, 4*24, SkatBestBidEstimate(three).Value, "with 3 on grand is (3+1)x24")

	// 空の手札でも壊れない。ジャックを 1 枚も持たないので without 4 以上。
	assert.Positive(t, SkatBestBidEstimate(nil).Value)
}
