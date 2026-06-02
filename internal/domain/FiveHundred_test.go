//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newFiveHundredForTest() *domain.FiveHundred {
	players := []*domain.FiveHundredPlayer{
		domain.NewFiveHundredPlayer(true, 0),
		domain.NewFiveHundredPlayer(false, 1),
		domain.NewFiveHundredPlayer(false, 0),
		domain.NewFiveHundredPlayer(false, 1),
	}
	return domain.NewFiveHundred(domain.NewTrumpCardsFiveHundred(), players, domain.DefaultFiveHundredConfig())
}

func TestNewTrumpCardsFiveHundred_DeckComposition(t *testing.T) {
	tc := domain.NewTrumpCardsFiveHundred()
	if got := tc.GetTotalCount(); got != 43 {
		t.Fatalf("deck size = %d, want 43", got)
	}
	// Draw every card and tally.
	suitVals := map[int][]int{}
	jokers := 0
	for {
		c := tc.DrawCard()
		if c == nil {
			break
		}
		if c.GetDesign() == domain.CardDesignJoker {
			jokers++
			continue
		}
		suitVals[c.GetDesign()] = append(suitVals[c.GetDesign()], c.GetValue())
	}
	if jokers != 1 {
		t.Errorf("jokers = %d, want 1", jokers)
	}
	if len(suitVals[domain.CardDesignHeart]) != 11 || len(suitVals[domain.CardDesignDiamond]) != 11 {
		t.Errorf("red suit counts = %d/%d, want 11/11",
			len(suitVals[domain.CardDesignHeart]), len(suitVals[domain.CardDesignDiamond]))
	}
	if len(suitVals[domain.CardDesignSpade]) != 10 || len(suitVals[domain.CardDesignClover]) != 10 {
		t.Errorf("black suit counts = %d/%d, want 10/10",
			len(suitVals[domain.CardDesignSpade]), len(suitVals[domain.CardDesignClover]))
	}
	// Black suits must not contain values 2,3,4.
	for _, v := range suitVals[domain.CardDesignSpade] {
		if v == 2 || v == 3 || v == 4 {
			t.Errorf("spade contains removed value %d", v)
		}
	}
	// Red suits must not contain values 2,3.
	for _, v := range suitVals[domain.CardDesignHeart] {
		if v == 2 || v == 3 {
			t.Errorf("heart contains removed value %d", v)
		}
	}
}

func TestFiveHundredBid_ValueAndOrder(t *testing.T) {
	cases := []struct {
		name      string
		bid       domain.FiveHundredBid
		wantValue int
		wantValid bool
	}{
		{"6 spades", domain.FiveHundredBid{Kind: domain.FiveHundredContractSuit, Tricks: 6, Suit: domain.CardDesignSpade}, 40, true},
		{"7 spades", domain.FiveHundredBid{Kind: domain.FiveHundredContractSuit, Tricks: 7, Suit: domain.CardDesignSpade}, 140, true},
		{"10 hearts", domain.FiveHundredBid{Kind: domain.FiveHundredContractSuit, Tricks: 10, Suit: domain.CardDesignHeart}, 500, true},
		{"6 NT", domain.FiveHundredBid{Kind: domain.FiveHundredContractNoTrump, Tricks: 6}, 120, true},
		{"10 NT", domain.FiveHundredBid{Kind: domain.FiveHundredContractNoTrump, Tricks: 10}, 520, true},
		{"misere", domain.FiveHundredBid{Kind: domain.FiveHundredContractMisere}, 250, true},
		{"open misere", domain.FiveHundredBid{Kind: domain.FiveHundredContractOpenMisere}, 520, true},
		{"bad tricks", domain.FiveHundredBid{Kind: domain.FiveHundredContractSuit, Tricks: 5, Suit: domain.CardDesignSpade}, 40, false},
	}
	for _, c := range cases {
		if got := c.bid.Value(); got != c.wantValue {
			t.Errorf("%s Value() = %d, want %d", c.name, got, c.wantValue)
		}
	}
	// Order: open misere outranks 10NT; misere sits above 8 spades(240).
	tenNT := domain.FiveHundredBid{Kind: domain.FiveHundredContractNoTrump, Tricks: 10}
	openMis := domain.FiveHundredBid{Kind: domain.FiveHundredContractOpenMisere}
	mis := domain.FiveHundredBid{Kind: domain.FiveHundredContractMisere}
	eightSpades := domain.FiveHundredBid{Kind: domain.FiveHundredContractSuit, Tricks: 8, Suit: domain.CardDesignSpade}
	if !(openMis.Order() > tenNT.Order()) {
		t.Errorf("open misere order %d should exceed 10NT order %d", openMis.Order(), tenNT.Order())
	}
	if !(mis.Order() > eightSpades.Order()) {
		t.Errorf("misere order %d should exceed 8 spades order %d", mis.Order(), eightSpades.Order())
	}
}

func TestFiveHundred_CardRank_BowersAndJoker(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)

	joker := domain.NewCard(domain.CardDesignJoker, 1, false)
	rightBower := domain.NewCard(domain.CardDesignSpade, 11, false)
	leftBower := domain.NewCard(domain.CardDesignClover, 11, false) // same color as spade
	trumpAce := domain.NewCard(domain.CardDesignSpade, 1, false)
	heartAce := domain.NewCard(domain.CardDesignHeart, 1, false)

	if !(g.CardRankPublic(joker) > g.CardRankPublic(rightBower)) {
		t.Errorf("joker should outrank right bower")
	}
	if !(g.CardRankPublic(rightBower) > g.CardRankPublic(leftBower)) {
		t.Errorf("right bower should outrank left bower")
	}
	if !(g.CardRankPublic(leftBower) > g.CardRankPublic(trumpAce)) {
		t.Errorf("left bower should outrank trump ace")
	}
	if !(g.CardRankPublic(trumpAce) > g.CardRankPublic(heartAce)) {
		t.Errorf("trump ace should outrank off-suit ace")
	}
	// Left bower's effective suit is the trump suit.
	if g.EffectiveSuitPublic(leftBower) != domain.CardDesignSpade {
		t.Errorf("left bower effective suit = %d, want spade", g.EffectiveSuitPublic(leftBower))
	}
}

func TestFiveHundred_TrickWinner_TrumpBeatsLead(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g.SetCurrentTrick([]*domain.FiveHundredTrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},  // lead heart A
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 13, false)}, // heart K
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},  // trump 5
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignDiamond, 1, false)},
	})
	g.SetPhase(domain.FiveHundredPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.ResolveTrick()
	if g.GetPlayer(2).GetTrickCount() != 1 {
		t.Errorf("trump player should win the trick")
	}
}

func TestFiveHundred_TrickWinner_JokerWinsNoTrump(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractNoTrump, 7, -1)
	g.SetCurrentTrick([]*domain.FiveHundredTrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)}, // lead heart A
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignJoker, 1, false)}, // joker
	})
	g.SetPhase(domain.FiveHundredPhaseTrickEnd)
	g.SetCurrentPlayerIdx(0)
	// Only 2 active is wrong for NT; force resolve via a 4-card trick instead.
	g.SetCurrentTrick([]*domain.FiveHundredTrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignJoker, 1, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 12, false)},
	})
	g.SetTrickNumber(1)
	g.ResolveTrick()
	if g.GetPlayer(1).GetTrickCount() != 1 {
		t.Errorf("joker should win the trick in no-trump")
	}
}

func TestFiveHundred_Reset_DealsHands(t *testing.T) {
	g := newFiveHundredForTest()
	g.Reset()
	if g.GetPhase() != domain.FiveHundredPhaseBid {
		t.Errorf("phase = %d, want Bid", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != domain.FiveHundredHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, domain.FiveHundredHandSize)
		}
	}
	if len(g.GetKitty()) != domain.FiveHundredKittySize {
		t.Errorf("kitty = %d, want %d", len(g.GetKitty()), domain.FiveHundredKittySize)
	}
	if g.GetBidPlayerIdx() != 1 {
		t.Errorf("first bidder = %d, want 1 (left of dealer)", g.GetBidPlayerIdx())
	}
}

func TestFiveHundred_ScoreRound_SuitContract(t *testing.T) {
	tests := []struct {
		name       string
		bidTricks  int
		declTricks int // tricks for declarer team (0,2)
		wantTeam0  int
		wantTeam1  int
	}{
		{"made", 7, 7, 140, 30},
		{"set", 7, 5, -140, 50},
		{"slam bonus", 6, 10, 250, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newFiveHundredForTest()
			g.SetContract(domain.FiveHundredContractSuit, tt.bidTricks, domain.CardDesignSpade)
			g.SetDeclarerIdx(0)
			// Assign tricks: declarer team = players 0,2; defenders = 1,3.
			defTricks := domain.FiveHundredTrickCnt - tt.declTricks
			for i := 0; i < tt.declTricks; i++ {
				g.GetPlayer(i % 2 * 2).AddTrick([]*domain.Card{nil})
			}
			for i := 0; i < defTricks; i++ {
				g.GetPlayer(1).AddTrick([]*domain.Card{nil})
			}
			g.SetPhase(domain.FiveHundredPhaseRoundEnd)
			g.ScoreRound()
			if g.GetTeamScore(0) != tt.wantTeam0 {
				t.Errorf("team0 = %d, want %d", g.GetTeamScore(0), tt.wantTeam0)
			}
			if g.GetTeamScore(1) != tt.wantTeam1 {
				t.Errorf("team1 = %d, want %d", g.GetTeamScore(1), tt.wantTeam1)
			}
		})
	}
}

func TestFiveHundred_ScoreRound_Misere(t *testing.T) {
	// Made: declarer takes 0 tricks → +250.
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractMisere, 0, -1)
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.FiveHundredPhaseRoundEnd)
	g.ScoreRound()
	if g.GetTeamScore(0) != 250 {
		t.Errorf("misere made: team0 = %d, want 250", g.GetTeamScore(0))
	}

	// Failed: declarer takes a trick → -250.
	g2 := newFiveHundredForTest()
	g2.SetContract(domain.FiveHundredContractMisere, 0, -1)
	g2.SetDeclarerIdx(0)
	g2.GetPlayer(0).AddTrick([]*domain.Card{nil})
	g2.SetPhase(domain.FiveHundredPhaseRoundEnd)
	g2.ScoreRound()
	if g2.GetTeamScore(0) != -250 {
		t.Errorf("misere failed: team0 = %d, want -250", g2.GetTeamScore(0))
	}
}

func TestFiveHundred_ScoreRound_GameEnd(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetTeamScore(0, 400)
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g.SetDeclarerIdx(0)
	for i := 0; i < 7; i++ {
		g.GetPlayer(i % 2 * 2).AddTrick([]*domain.Card{nil})
	}
	for i := 0; i < 3; i++ {
		g.GetPlayer(1).AddTrick([]*domain.Card{nil})
	}
	g.SetPhase(domain.FiveHundredPhaseRoundEnd)
	g.ScoreRound()
	if !g.GetGameEndFlag() {
		t.Fatalf("expected game to end at >= 500")
	}
	if g.GetWinnerTeam() != 0 {
		t.Errorf("winner = %d, want 0", g.GetWinnerTeam())
	}
}

func TestFiveHundred_KittyExchange_Validation(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetPhase(domain.FiveHundredPhaseKittyExchange)
	g.SetDeclarerIdx(0)
	// Give the human declarer 13 cards.
	for i := 0; i < 13; i++ {
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, (i%10)+4, false))
	}
	if err := g.PlayerExchangeKitty([]int{0, 1}); err == nil {
		t.Errorf("expected error for wrong discard count")
	}
	if err := g.PlayerExchangeKitty([]int{0, 0, 1}); err == nil {
		t.Errorf("expected error for duplicate indices")
	}
	if err := g.PlayerExchangeKitty([]int{0, 1, 99}); err == nil {
		t.Errorf("expected error for out-of-range index")
	}
	if err := g.PlayerExchangeKitty([]int{0, 1, 2}); err != nil {
		t.Fatalf("valid exchange failed: %v", err)
	}
	if g.GetPlayer(0).GetCardsSize() != 10 {
		t.Errorf("hand after exchange = %d, want 10", g.GetPlayer(0).GetCardsSize())
	}
	if g.GetPhase() != domain.FiveHundredPhasePlay {
		t.Errorf("phase after exchange = %d, want Play", g.GetPhase())
	}
}

func TestFiveHundred_PlayTrick_FollowSuitAndResolve(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g.SetDeclarerIdx(0)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))  // human leads heart A
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))  // follows
	g.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))  // trump
	g.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignHeart, 13, false)) // follows
	g.SetPhase(domain.FiveHundredPhasePlay)
	g.SetTrickNumber(1)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)

	if err := g.PlayerPlay(0, -1); err != nil {
		t.Fatalf("human play failed: %v", err)
	}
	for g.GetPhase() == domain.FiveHundredPhasePlay {
		g.CpuPlay()
	}
	if g.GetPhase() != domain.FiveHundredPhaseTrickEnd {
		t.Fatalf("phase = %d, want TrickEnd", g.GetPhase())
	}
	g.ResolveTrick()
	if g.GetPlayer(2).GetTrickCount() != 1 {
		t.Errorf("trump player (2) should win the trick")
	}
}

func TestFiveHundred_FollowSuitEnforced(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractNoTrump, 7, -1)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
	g.SetPhase(domain.FiveHundredPhasePlay)
	g.SetTrickNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.FiveHundredTrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 13, false)}, // heart led
	})
	// Player holds a heart, so playing diamond (idx 1) must be rejected.
	if err := g.PlayerPlay(1, -1); err == nil {
		t.Errorf("expected follow-suit violation when holding the lead suit")
	}
}

func TestFiveHundred_BidFlow_FinalizesContract(t *testing.T) {
	g := newFiveHundredForTest()
	g.Reset()
	steps := 0
	for g.GetPhase() == domain.FiveHundredPhaseBid && steps < 5000 {
		idx := g.GetBidPlayerIdx()
		if g.GetPlayer(idx).GetIsHuman() {
			_ = g.PlayerPass()
		} else {
			g.CpuBid()
		}
		steps++
	}
	if g.GetPhase() == domain.FiveHundredPhaseBid {
		t.Fatalf("bidding did not resolve within %d steps", steps)
	}
	// A resolved contract must hand the kitty to a declarer.
	if g.GetPhase() == domain.FiveHundredPhaseKittyExchange {
		if g.GetDeclarerIdx() < 0 {
			t.Errorf("declarer not set after bidding")
		}
		if g.GetPlayer(g.GetDeclarerIdx()).GetCardsSize() != 13 {
			t.Errorf("declarer hand = %d, want 13 after taking kitty",
				g.GetPlayer(g.GetDeclarerIdx()).GetCardsSize())
		}
	}
}

func TestFiveHundred_JSONRoundTrip(t *testing.T) {
	g := newFiveHundredForTest()
	g.Reset()
	g.SetTeamScore(0, 120)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored domain.FiveHundred
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetPhase() != g.GetPhase() {
		t.Errorf("phase mismatch: %d vs %d", restored.GetPhase(), g.GetPhase())
	}
	if restored.GetTeamScore(0) != 120 {
		t.Errorf("team score not restored: %d", restored.GetTeamScore(0))
	}
	if restored.GetPlayerCnt() != g.GetPlayerCnt() {
		t.Errorf("player count mismatch")
	}
}

func TestFiveHundred_Config_Validate(t *testing.T) {
	cfg := domain.DefaultFiveHundredConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("default config invalid: %v", err)
	}
	bad := domain.FiveHundredConfig{CpuDifficulty: 99, TargetScore: 0}
	if err := bad.Validate(); err == nil {
		t.Errorf("expected invalid config error")
	}
}
