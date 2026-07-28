//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

// ───── helpers ─────

// pq creates a fresh Piquet with two CPU players for deterministic test setup.
// First-deal Elder is idx=0.
func newPiquetForTest() *Piquet {
	players := []*PiquetPlayer{
		NewPiquetPlayer(false),
		NewPiquetPlayer(false),
	}
	return NewPiquet(NewTrumpCardsBelote(), players, DefaultPiquetConfig())
}

// addHand replaces a player's hand with the given cards.
func addHand(pl *PiquetPlayer, cards ...*Card) {
	pl.Reset()
	for _, c := range cards {
		pl.AddCard(c)
	}
}

// card shorthand: A=1, K=13, Q=12, J=11
func sp(v int) *Card { return NewCard(CardDesignSpade, v, false) }
func cl(v int) *Card { return NewCard(CardDesignClover, v, false) }
func he(v int) *Card { return NewCard(CardDesignHeart, v, false) }
func di(v int) *Card { return NewCard(CardDesignDiamond, v, false) }

// ───── Deal / Reset ─────

func TestNewDefaultPiquet(t *testing.T) {
	p := NewDefaultPiquet()
	if p == nil {
		t.Fatal("nil game")
	}
	if p.config.DealsPerPartie != 6 {
		t.Errorf("DealsPerPartie = %d, want 6", p.config.DealsPerPartie)
	}
	if !p.players[0].GetIsHuman() {
		t.Error("first player should be human")
	}
	if p.players[1].GetIsHuman() {
		t.Error("second player should be CPU")
	}
}

func TestPiquetReset(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	if p.GetPhase() != PiquetPhaseExchange {
		t.Errorf("phase = %d, want PiquetPhaseExchange", p.GetPhase())
	}
	if p.GetDealNumber() != 1 {
		t.Errorf("dealNumber = %d, want 1", p.GetDealNumber())
	}
	if p.GetElderIdx() != 0 {
		t.Errorf("elderIdx = %d, want 0", p.GetElderIdx())
	}
	for i := range PiquetPlayerCnt {
		if p.players[i].GetCardsSize() != PiquetHandSize {
			t.Errorf("player %d hand size = %d, want %d", i, p.players[i].GetCardsSize(), PiquetHandSize)
		}
	}
	// Talon should have 8 cards before any exchange
	totalRevealed := len(p.GetElderTalon()) + len(p.GetYoungerTalon())
	if totalRevealed != PiquetTalonSize {
		t.Errorf("talon size = %d, want %d", totalRevealed, PiquetTalonSize)
	}
}

// ───── Card rank/pip helpers ─────

func TestPiquetCardRank(t *testing.T) {
	tests := []struct {
		value int
		want  int
	}{
		{1, 8}, {13, 7}, {12, 6}, {11, 5}, {10, 4}, {9, 3}, {8, 2}, {7, 1},
		{0, 0}, // unknown
	}
	for _, tt := range tests {
		if got := piquetCardRank(tt.value); got != tt.want {
			t.Errorf("piquetCardRank(%d) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestPiquetPipValue(t *testing.T) {
	tests := []struct{ v, want int }{
		{1, 11}, {13, 10}, {12, 10}, {11, 10}, {10, 10}, {9, 9}, {8, 8}, {7, 7},
	}
	for _, tt := range tests {
		if got := piquetPipValue(tt.v); got != tt.want {
			t.Errorf("piquetPipValue(%d) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

// ───── Carte Blanche ─────

func TestCarteBlancheAutoDetected(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	// Force player 0 hand with no court cards (only 7,8,9,10,A)
	addHand(p.players[0],
		sp(7), sp(8), sp(9), sp(10), sp(1),
		cl(7), cl(8), cl(9), cl(10), cl(1),
		he(7), he(8),
	)
	// Force player 1 hand with court cards
	addHand(p.players[1],
		sp(11), sp(12), sp(13),
		cl(11), cl(12), cl(13),
		he(11), he(12), he(13),
		di(11), di(12), di(13),
	)
	// Re-run carte blanche detection
	p.carteBlanche = [PiquetPlayerCnt]bool{}
	p.players[0].SetDeclScore(0)
	p.players[1].SetDeclScore(0)
	p.firstScorerIdx = -1
	for i := range PiquetPlayerCnt {
		if hasNoCourtCards(p.players[i]) {
			p.carteBlanche[i] = true
			p.players[i].AddDeclScore(PiquetCarteBlancheBonus)
			p.recordFirstScorer(i)
		}
	}

	if !p.GetCarteBlanche(0) {
		t.Error("player 0 should have carte blanche")
	}
	if p.GetCarteBlanche(1) {
		t.Error("player 1 should NOT have carte blanche")
	}
	if p.players[0].GetDeclScore() != PiquetCarteBlancheBonus {
		t.Errorf("player 0 decl score = %d, want %d", p.players[0].GetDeclScore(), PiquetCarteBlancheBonus)
	}
}

func TestHasNoCourtCards(t *testing.T) {
	pl := NewPiquetPlayer(false)
	pl.AddCard(sp(7))
	pl.AddCard(sp(10))
	pl.AddCard(sp(1))
	if !hasNoCourtCards(pl) {
		t.Error("expected no court cards")
	}
	pl.AddCard(sp(11)) // Jack added
	if hasNoCourtCards(pl) {
		t.Error("expected court card detected")
	}
}

// ───── Exchange ─────

func TestExchangeElderValid(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	beforeHand := p.players[0].GetCardsSize()
	err := p.ExchangeElder([]int{0, 1, 2})
	if err != nil {
		t.Fatalf("ExchangeElder: %v", err)
	}
	if p.players[0].GetCardsSize() != beforeHand {
		t.Errorf("hand size should be unchanged after exchange, got %d", p.players[0].GetCardsSize())
	}
	if p.GetElderExchangedCnt() != 3 {
		t.Errorf("elderExchangedCnt = %d, want 3", p.GetElderExchangedCnt())
	}
	if p.GetExchangeTurn() != PiquetExchangeTurnYounger {
		t.Errorf("turn = %d, want Younger", p.GetExchangeTurn())
	}
}

func TestExchangeElderRejectsZero(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	if err := p.ExchangeElder([]int{}); err == nil {
		t.Error("expected error for zero-card exchange")
	}
}

func TestExchangeElderRejectsOverMax(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	if err := p.ExchangeElder([]int{0, 1, 2, 3, 4, 5}); err == nil {
		t.Error("expected error for 6-card exchange")
	}
}

func TestExchangeElderRejectsDuplicate(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	if err := p.ExchangeElder([]int{0, 0, 1}); err == nil {
		t.Error("expected error for duplicate index")
	}
}

func TestExchangeElderRejectsOutOfRange(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	if err := p.ExchangeElder([]int{99}); err == nil {
		t.Error("expected error for out-of-range index")
	}
}

func TestExchangeElderRejectsWrongPhase(t *testing.T) {
	p := newPiquetForTest()
	// Without Reset, phase is zero value (PiquetPhaseExchange = 0). Force phase.
	p.phase = PiquetPhasePlay
	if err := p.ExchangeElder([]int{0}); err == nil {
		t.Error("expected error for wrong phase")
	}
}

func TestExchangeYoungerValid(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	if err := p.ExchangeElder([]int{0, 1, 2, 3, 4}); err != nil {
		t.Fatal(err)
	}
	if err := p.ExchangeYounger([]int{0, 1, 2}); err != nil {
		t.Fatalf("ExchangeYounger: %v", err)
	}
	if p.GetYoungerExchangedCnt() != 3 {
		t.Errorf("youngerExchangedCnt = %d, want 3", p.GetYoungerExchangedCnt())
	}
	if p.GetPhase() != PiquetPhaseDeclaration {
		t.Errorf("phase = %d, want Declaration", p.GetPhase())
	}
}

func TestExchangeYoungerZeroAllowed(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	_ = p.ExchangeElder([]int{0})
	if err := p.ExchangeYounger([]int{}); err != nil {
		t.Errorf("Younger 0-exchange should be allowed: %v", err)
	}
}

func TestExchangeYoungerRejectsOverMax(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	_ = p.ExchangeElder([]int{0})
	if err := p.ExchangeYounger([]int{0, 1, 2, 3}); err == nil {
		t.Error("expected error for 4-card exchange")
	}
}

func TestExchangeYoungerBeforeElder(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	if err := p.ExchangeYounger([]int{0}); err == nil {
		t.Error("expected error: younger cannot exchange before elder")
	}
}

// ───── Best Point ─────

func TestBestPointSingleSuit(t *testing.T) {
	pl := NewPiquetPlayer(false)
	pl.AddCard(sp(1))
	pl.AddCard(sp(13))
	pl.AddCard(sp(12))
	pl.AddCard(sp(11))
	pl.AddCard(sp(10))
	pl.AddCard(cl(7))
	claim := bestPoint(pl)
	if claim == nil {
		t.Fatal("nil claim")
	}
	if claim.Length != 5 {
		t.Errorf("Length = %d, want 5", claim.Length)
	}
	// 11+10+10+10+10 = 51
	if claim.PipTotal != 51 {
		t.Errorf("PipTotal = %d, want 51", claim.PipTotal)
	}
	if claim.Suit != CardDesignSpade {
		t.Errorf("Suit = %d, want Spade", claim.Suit)
	}
}

func TestBestPointTieBreakByPip(t *testing.T) {
	// Two suits with 4 cards each — pip total decides
	pl := NewPiquetPlayer(false)
	// Spades: A, K, Q, J (11+10+10+10 = 41)
	pl.AddCard(sp(1))
	pl.AddCard(sp(13))
	pl.AddCard(sp(12))
	pl.AddCard(sp(11))
	// Hearts: 7, 8, 9, 10 (7+8+9+10 = 34)
	pl.AddCard(he(7))
	pl.AddCard(he(8))
	pl.AddCard(he(9))
	pl.AddCard(he(10))
	claim := bestPoint(pl)
	if claim.Length != 4 || claim.Suit != CardDesignSpade {
		t.Errorf("want best=spades (41 pip); got suit=%d length=%d", claim.Suit, claim.Length)
	}
}

// ───── Best Sequence ─────

func TestBestSequenceQuint(t *testing.T) {
	pl := NewPiquetPlayer(false)
	pl.AddCard(sp(7))
	pl.AddCard(sp(8))
	pl.AddCard(sp(9))
	pl.AddCard(sp(10))
	pl.AddCard(sp(11))
	pl.AddCard(cl(13))
	claim := bestSequence(pl)
	if claim == nil {
		t.Fatal("nil claim")
	}
	if claim.Length != 5 {
		t.Errorf("Length = %d, want 5", claim.Length)
	}
	if claim.TopRank != 5 { // J ranks 5
		t.Errorf("TopRank = %d, want 5", claim.TopRank)
	}
	if got := claimScore(PiquetDeclKindSequence, claim); got != 15 {
		t.Errorf("Quint score = %d, want 15", got)
	}
}

func TestBestSequenceNoneShortRun(t *testing.T) {
	pl := NewPiquetPlayer(false)
	pl.AddCard(sp(7))
	pl.AddCard(sp(8))
	pl.AddCard(cl(7))
	pl.AddCard(cl(9))
	if got := bestSequence(pl); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestBestSequenceTieByTopRank(t *testing.T) {
	pl := NewPiquetPlayer(false)
	// Spades: 7-8-9 (top=9 → rank 3)
	pl.AddCard(sp(7))
	pl.AddCard(sp(8))
	pl.AddCard(sp(9))
	// Hearts: 11-12-13 (top=K → rank 7)
	pl.AddCard(he(11))
	pl.AddCard(he(12))
	pl.AddCard(he(13))
	claim := bestSequence(pl)
	if claim.Suit != CardDesignHeart {
		t.Errorf("expected hearts (higher top), got suit=%d top=%d", claim.Suit, claim.TopRank)
	}
}

// ───── Sets ─────

func TestAllSetsTrioAndQuatorze(t *testing.T) {
	pl := NewPiquetPlayer(false)
	// Quatorze of Aces
	pl.AddCard(sp(1))
	pl.AddCard(cl(1))
	pl.AddCard(he(1))
	pl.AddCard(di(1))
	// Trio of Kings
	pl.AddCard(sp(13))
	pl.AddCard(cl(13))
	pl.AddCard(he(13))
	// Junk (excluded: 9s are below set rank)
	pl.AddCard(sp(9))
	pl.AddCard(cl(9))
	pl.AddCard(he(9))

	sets := allSets(pl)
	if len(sets) != 2 {
		t.Fatalf("len(sets) = %d, want 2", len(sets))
	}
	// First should be aces (length 4)
	if sets[0].Length != 4 {
		t.Errorf("first set length = %d, want 4", sets[0].Length)
	}
	if sets[1].Length != 3 {
		t.Errorf("second set length = %d, want 3", sets[1].Length)
	}
}

func TestClaimScoreSet(t *testing.T) {
	if got := claimScore(PiquetDeclKindSet, &PiquetClaim{Length: 4}); got != 14 {
		t.Errorf("quatorze score = %d, want 14", got)
	}
	if got := claimScore(PiquetDeclKindSet, &PiquetClaim{Length: 3}); got != 3 {
		t.Errorf("trio score = %d, want 3", got)
	}
	if got := claimScore(PiquetDeclKindSet, &PiquetClaim{Length: 2}); got != 0 {
		t.Errorf("pair score = %d, want 0", got)
	}
}

// ───── ResolveDeclaration ─────

func TestResolveDeclarationsFullFlow(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	addHand(p.players[0],
		sp(1), sp(13), sp(12), sp(11), sp(10), sp(9), // 6 spades → sixième (rank A high) = 16 pts
		cl(1), he(1), di(1), // quatorze of Aces (with sp(1))
		cl(13), he(13), di(13), // quatorze of Kings (with sp(13))
	)
	addHand(p.players[1],
		// Younger weaker: scattered, no sequences ≥3 in any single suit
		sp(7), sp(8), cl(7), cl(8), he(7), he(8), di(7), di(8),
		sp(11), sp(12), cl(9), he(9),
	)
	p.phase = PiquetPhaseDeclaration
	p.declStage = PiquetDeclKindPoint

	// Point comparison
	res, err := p.ResolveDeclaration()
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != PiquetDeclKindPoint {
		t.Errorf("expected Point first, got kind=%d", res.Kind)
	}
	if res.ScoredBy != 0 {
		t.Errorf("expected Elder (0) to win Point, got scoredBy=%d", res.ScoredBy)
	}

	// Sequence comparison
	res, _ = p.ResolveDeclaration()
	if res.Kind != PiquetDeclKindSequence {
		t.Errorf("expected Sequence second, got kind=%d", res.Kind)
	}

	// Set comparison
	res, _ = p.ResolveDeclaration()
	if res.Kind != PiquetDeclKindSet {
		t.Errorf("expected Set third, got kind=%d", res.Kind)
	}
	// After all 3 → Play
	if p.GetPhase() != PiquetPhasePlay {
		t.Errorf("phase after 3 decls = %d, want Play", p.GetPhase())
	}
}

func TestResolveDeclarationTie(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	// Identical point in two suits (4 cards each, same pip)
	addHand(p.players[0],
		sp(1), sp(13), sp(12), sp(11), // 4 spades, pip 41
		he(7), he(8), he(9), he(10), // 4 hearts (pip 34) — best is spades
		cl(7), cl(8), di(7), di(8),
	)
	addHand(p.players[1],
		sp(7), sp(8), sp(9), sp(10), // 4 spades, pip 34
		he(1), he(13), he(12), he(11), // 4 hearts, pip 41 — best is hearts
		cl(9), cl(10), di(9), di(10),
	)
	p.phase = PiquetPhaseDeclaration
	p.declStage = PiquetDeclKindPoint
	res, _ := p.ResolveDeclaration()
	if res == nil {
		t.Fatal("nil result")
	}
	if res.Score != 0 {
		t.Errorf("tie should score 0, got %d", res.Score)
	}
}

// ───── Repique ─────

func TestRepiqueAwardedWhenDeclScoreReaches30Alone(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	// Force Elder to 30 decl points without Younger scoring
	p.players[p.elderIdx].SetDeclScore(35)
	p.players[p.GetYoungerIdx()].SetDeclScore(0)
	p.checkRepique()
	if p.players[p.elderIdx].GetBonusScore() != PiquetRepiqueBonus {
		t.Errorf("repique bonus = %d, want %d", p.players[p.elderIdx].GetBonusScore(), PiquetRepiqueBonus)
	}
}

func TestRepiqueNotAwardedIfOpponentScored(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.players[p.elderIdx].SetDeclScore(35)
	p.players[p.GetYoungerIdx()].SetDeclScore(3)
	p.checkRepique()
	if p.players[p.elderIdx].GetBonusScore() != 0 {
		t.Errorf("repique bonus should be 0 when opponent scored, got %d", p.players[p.elderIdx].GetBonusScore())
	}
}

// ───── Play phase ─────

func TestGetLegalPlayIndicesFollowSuit(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.players[1].Reset()
	p.players[1].AddCard(sp(7))
	p.players[1].AddCard(cl(7))
	p.players[1].AddCard(he(7))
	p.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: sp(13)},
	}
	legal := p.GetLegalPlayIndices(1)
	if len(legal) != 1 || legal[0] != 0 {
		t.Errorf("expected follow-suit indices [0], got %v", legal)
	}
}

func TestGetLegalPlayIndicesAllWhenNoSuit(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.players[1].Reset()
	p.players[1].AddCard(cl(7))
	p.players[1].AddCard(he(7))
	p.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: sp(13)},
	}
	legal := p.GetLegalPlayIndices(1)
	if len(legal) != 2 {
		t.Errorf("expected 2 legal cards (no follow-suit), got %v", legal)
	}
}

func TestPlayCardLeadScoresAndAdvancesTurn(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.declStage = PiquetDeclKindPoint // reset
	p.trickNumber = 0
	p.tricksWon = [PiquetPlayerCnt]int{}
	p.currentTrick = nil
	p.currentPlayerIdx = 0
	p.leadPlayerIdx = 0
	p.firstScorerIdx = -1
	for i := range 2 {
		p.players[i].SetDeclScore(0)
		p.players[i].SetTrickScore(0)
		p.players[i].SetBonusScore(0)
	}
	addHand(p.players[0], sp(13))
	addHand(p.players[1], sp(7))

	if err := p.PlayCard(0); err != nil {
		t.Fatalf("lead: %v", err)
	}
	if p.players[0].GetTrickScore() != PiquetTrickLeadPoint {
		t.Errorf("leader did not get +1: got %d", p.players[0].GetTrickScore())
	}
	if p.GetCurrentPlayerIdx() != 1 {
		t.Errorf("turn not advanced: got %d", p.GetCurrentPlayerIdx())
	}
}

func TestPlayTrickFullResolution(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.trickNumber = 0
	p.tricksWon = [PiquetPlayerCnt]int{}
	p.currentTrick = nil
	p.currentPlayerIdx = 0
	p.leadPlayerIdx = 0
	p.firstScorerIdx = -1
	for i := range 2 {
		p.players[i].SetDeclScore(0)
		p.players[i].SetTrickScore(0)
		p.players[i].SetBonusScore(0)
	}
	// Single-trick setup
	addHand(p.players[0], sp(7))
	addHand(p.players[1], sp(13))
	// trickNumber must be set to last trick to avoid endRoundScoring complications
	p.trickNumber = PiquetTricksPerRound - 1

	if err := p.PlayCard(0); err != nil { // lead 7♠
		t.Fatal(err)
	}
	if err := p.PlayCard(0); err != nil { // follow K♠
		t.Fatal(err)
	}
	// Winner is player 1 (K beats 7)
	if p.tricksWon[1] != 1 {
		t.Errorf("player 1 should win trick, tricksWon=%v", p.tricksWon)
	}
	// player 1 should get +1 (won non-lead) + +1 (last trick) = 2
	if p.players[1].GetTrickScore() != 2 {
		t.Errorf("player 1 trickScore = %d, want 2", p.players[1].GetTrickScore())
	}
	// player 0 should keep +1 (led — but did not win, so no last-trick bonus)
	if p.players[0].GetTrickScore() != 1 {
		t.Errorf("player 0 trickScore = %d, want 1", p.players[0].GetTrickScore())
	}
}

// ───── Capot / Cards bonus ─────

func TestEndRoundCardsBonus(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.tricksWon = [PiquetPlayerCnt]int{8, 4}
	p.dealNumber = 1
	p.config.DealsPerPartie = 6
	p.trickNumber = PiquetTricksPerRound
	for i := range 2 {
		p.players[i].SetBonusScore(0)
		p.players[i].SetMatchScore(0)
	}
	p.endRoundScoring()
	if p.players[0].GetBonusScore() != PiquetCardsBonus {
		t.Errorf("player 0 cards bonus = %d, want %d", p.players[0].GetBonusScore(), PiquetCardsBonus)
	}
	if p.players[1].GetBonusScore() != 0 {
		t.Errorf("player 1 cards bonus = %d, want 0", p.players[1].GetBonusScore())
	}
	if p.GetPhase() != PiquetPhaseScore {
		t.Errorf("phase after endRound = %d, want Score", p.GetPhase())
	}
}

func TestEndRoundCapotBonus(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.tricksWon = [PiquetPlayerCnt]int{12, 0}
	p.config.DealsPerPartie = 6
	p.trickNumber = PiquetTricksPerRound
	for i := range 2 {
		p.players[i].SetBonusScore(0)
		p.players[i].SetMatchScore(0)
	}
	p.endRoundScoring()
	if p.players[0].GetBonusScore() != PiquetCapotBonus {
		t.Errorf("capot bonus = %d, want %d", p.players[0].GetBonusScore(), PiquetCapotBonus)
	}
}

func TestEndRoundTrickTieNoBonus(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.tricksWon = [PiquetPlayerCnt]int{6, 6}
	p.config.DealsPerPartie = 6
	p.trickNumber = PiquetTricksPerRound
	for i := range 2 {
		p.players[i].SetBonusScore(0)
	}
	p.endRoundScoring()
	if p.players[0].GetBonusScore() != 0 || p.players[1].GetBonusScore() != 0 {
		t.Errorf("tied tricks should award no bonus, got %d / %d", p.players[0].GetBonusScore(), p.players[1].GetBonusScore())
	}
}

// ───── Pique ─────

func TestPiqueAwardedToElder(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.elderReached30InPlay = false
	p.players[p.elderIdx].SetDeclScore(28)
	p.players[p.elderIdx].SetTrickScore(0)
	p.players[p.GetYoungerIdx()].SetDeclScore(0)
	p.players[p.GetYoungerIdx()].SetTrickScore(0)
	p.players[p.elderIdx].SetBonusScore(0)
	// Elder scores +2 in play → total 30
	p.players[p.elderIdx].AddTrickScore(2)
	p.checkPique(p.elderIdx)
	if p.players[p.elderIdx].GetBonusScore() != PiquetPiqueBonus {
		t.Errorf("pique bonus = %d, want %d", p.players[p.elderIdx].GetBonusScore(), PiquetPiqueBonus)
	}
}

func TestPiqueNotAwardedWhenYoungerHasScore(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.elderReached30InPlay = false
	p.players[p.elderIdx].SetDeclScore(30)
	p.players[p.GetYoungerIdx()].SetDeclScore(1)
	p.players[p.elderIdx].SetBonusScore(0)
	p.checkPique(p.elderIdx)
	if p.players[p.elderIdx].GetBonusScore() != 0 {
		t.Errorf("pique should not be awarded, got %d", p.players[p.elderIdx].GetBonusScore())
	}
}

func TestPiqueOnlyForElder(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	younger := p.GetYoungerIdx()
	p.players[younger].SetDeclScore(30)
	p.players[p.elderIdx].SetDeclScore(0)
	p.players[younger].SetBonusScore(0)
	p.checkPique(younger)
	if p.players[younger].GetBonusScore() != 0 {
		t.Errorf("pique should only be elder; got %d for younger", p.players[younger].GetBonusScore())
	}
}

// ───── NextDeal / Partie ─────

func TestNextDealSwapsElder(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhaseScore
	prevElder := p.elderIdx
	p.NextDeal()
	if p.elderIdx == prevElder {
		t.Errorf("Elder did not swap after NextDeal")
	}
	if p.dealNumber != 2 {
		t.Errorf("dealNumber = %d, want 2", p.dealNumber)
	}
	if p.phase != PiquetPhaseExchange {
		t.Errorf("phase after NextDeal = %d, want Exchange", p.phase)
	}
}

func TestNextDealNoOpOutsideScorePhase(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	prev := p.dealNumber
	p.NextDeal() // phase=Exchange → no op
	if p.dealNumber != prev {
		t.Errorf("NextDeal should no-op outside Score phase")
	}
}

func TestPartieFinishesAfterAllDeals(t *testing.T) {
	p := newPiquetForTest()
	p.config.DealsPerPartie = 1
	p.Reset()
	p.phase = PiquetPhaseScore
	p.NextDeal()
	if !p.GetGameEndFlag() {
		t.Error("expected gameEndFlag after final deal")
	}
	if p.GetPhase() != PiquetPhaseGameEnd {
		t.Errorf("phase = %d, want GameEnd", p.GetPhase())
	}
}

func TestRubiconAddsBothScoresAsBonus(t *testing.T) {
	p := newPiquetForTest()
	p.config.DealsPerPartie = 1
	p.Reset()
	// Force scores: winner=80, loser=50 (rubicon)
	p.players[0].SetMatchScore(80)
	p.players[1].SetMatchScore(50)
	p.phase = PiquetPhaseScore
	p.NextDeal()
	// Bonus = 100 + 80 + 50 = 230 → winner becomes 80+230=310
	if p.players[0].GetMatchScore() != 310 {
		t.Errorf("rubicon winner score = %d, want 310", p.players[0].GetMatchScore())
	}
	// Loser unchanged
	if p.players[1].GetMatchScore() != 50 {
		t.Errorf("rubicon loser score changed: %d", p.players[1].GetMatchScore())
	}
}

func TestNormalPartieAdds100PlusDifference(t *testing.T) {
	p := newPiquetForTest()
	p.config.DealsPerPartie = 1
	p.Reset()
	p.players[0].SetMatchScore(150)
	p.players[1].SetMatchScore(120)
	p.phase = PiquetPhaseScore
	p.NextDeal()
	// Bonus = 100 + 150 - 120 = 130 → winner=280
	if p.players[0].GetMatchScore() != 280 {
		t.Errorf("normal winner = %d, want 280", p.players[0].GetMatchScore())
	}
}

func TestTiedPartieNoBonusNoWinner(t *testing.T) {
	p := newPiquetForTest()
	p.config.DealsPerPartie = 1
	p.Reset()
	p.players[0].SetMatchScore(100)
	p.players[1].SetMatchScore(100)
	p.phase = PiquetPhaseScore
	p.NextDeal()
	if p.GetWinnerIdx() != -1 {
		t.Errorf("tied partie winner = %d, want -1", p.GetWinnerIdx())
	}
}

// ───── Full game with CPU ─────

func TestCpuPlaysFullDeal(t *testing.T) {
	players := []*PiquetPlayer{
		NewPiquetPlayer(false),
		NewPiquetPlayer(false),
	}
	p := NewPiquet(NewTrumpCardsBelote(), players, PiquetConfig{DealsPerPartie: 1, CpuDifficulty: PiquetCpuDifficultyNormal})
	p.Reset()

	// Exchange phase: drive both CPUs
	for p.phase == PiquetPhaseExchange {
		p.CpuPlay()
	}
	// Declarations: resolve 3 stages
	for p.phase == PiquetPhaseDeclaration {
		if _, err := p.ResolveDeclaration(); err != nil {
			t.Fatal(err)
		}
	}
	// Play: 12 tricks
	for p.phase == PiquetPhasePlay {
		p.CpuPlay()
	}
	// Round end → next deal is final (DealsPerPartie=1) → game end
	if p.phase != PiquetPhaseScore && p.phase != PiquetPhaseGameEnd {
		t.Errorf("phase after play = %d, want Score/GameEnd", p.phase)
	}
	total := p.tricksWon[0] + p.tricksWon[1]
	if total != PiquetTricksPerRound {
		t.Errorf("total tricks = %d, want %d", total, PiquetTricksPerRound)
	}
}

// ───── JSON round-trip ─────

func TestPiquetJSONRoundTrip(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.players[0].AddDeclScore(20)
	p.players[1].AddTrickScore(5)
	p.players[0].AddMatchScore(50)

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	restored := &Piquet{}
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if restored.GetPhase() != p.GetPhase() {
		t.Errorf("phase mismatch")
	}
	if restored.players[0].GetDeclScore() != p.players[0].GetDeclScore() {
		t.Errorf("decl score mismatch: %d vs %d",
			restored.players[0].GetDeclScore(), p.players[0].GetDeclScore())
	}
	if restored.players[0].GetMatchScore() != 50 {
		t.Errorf("match score mismatch: %d vs 50", restored.players[0].GetMatchScore())
	}
}

// ───── Hint ─────

func TestGetHintInExchangePhase(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	hint := p.GetHint(p.elderIdx)
	if hint == nil {
		t.Fatal("expected hint in exchange phase")
	}
	if len(hint.DiscardIndices) == 0 {
		t.Error("expected non-empty discard recommendation")
	}
}

func TestGetHintInPlayPhase(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	p.currentTrick = nil
	p.currentPlayerIdx = 0
	hint := p.GetHint(0)
	if hint == nil || hint.CardIndex == nil {
		t.Fatalf("expected card hint, got %+v", hint)
	}
}

// ───── Edge: invalid PlayCard ─────

func TestPlayCardRejectsWrongPhase(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	if err := p.PlayCard(0); err == nil {
		t.Error("expected error: PlayCard in Exchange phase")
	}
}

func TestPlayCardRejectsIllegalFollow(t *testing.T) {
	p := newPiquetForTest()
	p.Reset()
	p.phase = PiquetPhasePlay
	addHand(p.players[0], sp(7), cl(8))
	addHand(p.players[1], sp(13), he(12))
	p.currentTrick = nil
	p.currentPlayerIdx = 0
	if err := p.PlayCard(0); err != nil { // lead 7♠
		t.Fatal(err)
	}
	// Player 1 must follow spade; trying cl(?) should fail
	// Player 1 has sp(13) at idx 0 → playing he(12) at idx 1 is illegal
	if err := p.PlayCard(1); err == nil {
		t.Error("expected error: not following suit")
	}
}

// ───── Domain error helpers ─────

func TestValidateUniqueRange(t *testing.T) {
	tests := []struct {
		name    string
		idx     []int
		max     int
		wantErr bool
	}{
		{"ok", []int{0, 1, 2}, 10, false},
		{"empty", []int{}, 10, false},
		{"out of range", []int{0, 11}, 10, true},
		{"negative", []int{-1}, 10, true},
		{"duplicate", []int{0, 0}, 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUniqueRange(tt.idx, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUniqueRange err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
