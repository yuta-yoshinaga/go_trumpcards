//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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
		if c.wantValid {
			if got := c.bid.Value(); got != c.wantValue {
				t.Errorf("%s Value() = %d, want %d", c.name, got, c.wantValue)
			}
		}
	}
	// Order: open misere outranks 10NT; misere sits above 8 spades(240).
	tenNT := domain.FiveHundredBid{Kind: domain.FiveHundredContractNoTrump, Tricks: 10}
	openMis := domain.FiveHundredBid{Kind: domain.FiveHundredContractOpenMisere}
	mis := domain.FiveHundredBid{Kind: domain.FiveHundredContractMisere}
	eightSpades := domain.FiveHundredBid{Kind: domain.FiveHundredContractSuit, Tricks: 8, Suit: domain.CardDesignSpade}
	if openMis.Order() <= tenNT.Order() {
		t.Errorf("open misere order %d should exceed 10NT order %d", openMis.Order(), tenNT.Order())
	}
	if mis.Order() <= eightSpades.Order() {
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

	if g.CardRankPublic(joker) <= g.CardRankPublic(rightBower) {
		t.Errorf("joker should outrank right bower")
	}
	if g.CardRankPublic(rightBower) <= g.CardRankPublic(leftBower) {
		t.Errorf("right bower should outrank left bower")
	}
	if g.CardRankPublic(leftBower) <= g.CardRankPublic(trumpAce) {
		t.Errorf("left bower should outrank trump ace")
	}
	if g.CardRankPublic(trumpAce) <= g.CardRankPublic(heartAce) {
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
	g.SetCurrentTrick([]*domain.TrickCard{
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
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)}, // lead heart A
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignJoker, 1, false)}, // joker
	})
	g.SetPhase(domain.FiveHundredPhaseTrickEnd)
	g.SetCurrentPlayerIdx(0)
	// Only 2 active is wrong for NT; force resolve via a 4-card trick instead.
	g.SetCurrentTrick([]*domain.TrickCard{
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

// **内訳を読める形で返す (#4809)。**画面が「成立/セット」「必要トリック数」
// 「増減点」を言えるようにする。ScoreRound と同じ計算を共有しているので、表示と
// 実際の加算がずれることはない。
func TestFiveHundred_GetRoundResult(t *testing.T) {
	tests := []struct {
		name       string
		contract   domain.FiveHundredContractKind
		bidTricks  int
		declTricks int
		wantMade   bool
		wantSlam   bool
		wantNeed   int
	}{
		{"made", domain.FiveHundredContractSuit, 7, 7, true, false, 7},
		{"set", domain.FiveHundredContractSuit, 7, 5, false, false, 7},
		{"slam", domain.FiveHundredContractSuit, 6, 10, true, true, 6},
		{"misere made", domain.FiveHundredContractMisere, 0, 0, true, false, 0},
		{"misere failed", domain.FiveHundredContractMisere, 0, 1, false, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newFiveHundredForTest()
			g.SetContract(tt.contract, tt.bidTricks, domain.CardDesignSpade)
			g.SetDeclarerIdx(0)
			for i := 0; i < tt.declTricks; i++ {
				g.GetPlayer(i % 2 * 2).AddTrick([]*domain.Card{nil})
			}
			for i := 0; i < domain.FiveHundredTrickCnt-tt.declTricks; i++ {
				g.GetPlayer(1).AddTrick([]*domain.Card{nil})
			}
			g.SetPhase(domain.FiveHundredPhaseRoundEnd)

			r := g.GetRoundResult()
			if r == nil {
				t.Fatal("ラウンド終了フェーズでは内訳が返る")
			}
			if r.Made != tt.wantMade {
				t.Errorf("Made = %v, want %v", r.Made, tt.wantMade)
			}
			if r.Slam != tt.wantSlam {
				t.Errorf("Slam = %v, want %v", r.Slam, tt.wantSlam)
			}
			if r.NeedTricks != tt.wantNeed {
				t.Errorf("NeedTricks = %d, want %d", r.NeedTricks, tt.wantNeed)
			}
			if r.DeclarerTricks != tt.declTricks {
				t.Errorf("DeclarerTricks = %d, want %d", r.DeclarerTricks, tt.declTricks)
			}
			if r.DeclarerTeam == r.DefenderTeam {
				t.Errorf("宣言側と守備側が同じチーム: %d", r.DeclarerTeam)
			}

			// 表示した増減が、実際に加算される値と一致する。
			before := [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
			g.ScoreRound()
			if got := g.GetTeamScore(r.DeclarerTeam); got != before[r.DeclarerTeam]+r.DeclarerDelta {
				t.Errorf("declarer score = %d, want %d", got, before[r.DeclarerTeam]+r.DeclarerDelta)
			}
			if got := g.GetTeamScore(r.DefenderTeam); got != before[r.DefenderTeam]+r.DefenderDelta {
				t.Errorf("defender score = %d, want %d", got, before[r.DefenderTeam]+r.DefenderDelta)
			}
		})
	}

	// ラウンド終了フェーズ以外、および契約が決まっていないときは nil。
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.FiveHundredPhasePlay)
	if g.GetRoundResult() != nil {
		t.Error("プレイ中は内訳を返さない")
	}
	g2 := newFiveHundredForTest()
	g2.SetPhase(domain.FiveHundredPhaseRoundEnd)
	if g2.GetRoundResult() != nil {
		t.Error("落札者がいなければ内訳はない")
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
	g.SetCurrentTrick([]*domain.TrickCard{
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
			// Open with the lowest bid when nobody has bid yet, else pass.
			// This guarantees a standing bid so the auction always resolves
			// (avoids relying on a CPU bidding, which can all-pass and redeal).
			if g.GetHighestBid() == nil {
				_ = g.PlayerBid(domain.FiveHundredContractSuit, 6, domain.CardDesignSpade)
			} else {
				_ = g.PlayerPass()
			}
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

func fhCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func TestFiveHundred_MisereContract_PartnerSkipped(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractMisere, 0, -1)
	g.SetDeclarerIdx(0) // partner = player 2 is skipped
	g.GetPlayer(0).AddCard(fhCard(domain.CardDesignHeart, 5))
	g.GetPlayer(1).AddCard(fhCard(domain.CardDesignHeart, 6))
	g.GetPlayer(3).AddCard(fhCard(domain.CardDesignHeart, 7))
	g.SetPhase(domain.FiveHundredPhasePlay)
	g.SetTrickNumber(1)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)

	if err := g.PlayerPlay(0, -1); err != nil {
		t.Fatalf("human play failed: %v", err)
	}
	steps := 0
	for g.GetPhase() == domain.FiveHundredPhasePlay && steps < 10 {
		g.CpuPlay()
		steps++
	}
	if g.GetPhase() != domain.FiveHundredPhaseTrickEnd {
		t.Fatalf("misère trick did not complete with 3 active players, phase=%d", g.GetPhase())
	}
	// Player 2 (declarer's partner) must not have played.
	if g.GetPlayer(2).GetCardsSize() != 0 {
		t.Errorf("skipped partner should hold no test cards")
	}
	g.ResolveTrick()
}

func TestFiveHundred_GetHint_BidAndExchange(t *testing.T) {
	// Bid phase, weak hand -> pass recommended.
	g := newFiveHundredForTest()
	g.SetPhase(domain.FiveHundredPhaseBid)
	g.SetBidPlayerIdx(0)
	for _, v := range []int{5, 6, 7, 8, 9} {
		g.GetPlayer(0).AddCard(fhCard(domain.CardDesignHeart, v))
	}
	hint := g.GetHint()
	if hint == nil || hint.Pass == nil || !*hint.Pass {
		t.Errorf("weak bid hand should recommend pass, got %+v", hint)
	}

	// Kitty exchange phase -> 3 discard indices.
	g2 := newFiveHundredForTest()
	g2.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g2.SetDeclarerIdx(0)
	g2.SetPhase(domain.FiveHundredPhaseKittyExchange)
	for i := 0; i < 13; i++ {
		g2.GetPlayer(0).AddCard(fhCard(domain.CardDesignHeart, (i%9)+5))
	}
	h2 := g2.GetHint()
	if h2 == nil || len(h2.DiscardIndices) != 3 {
		t.Errorf("exchange hint should suggest 3 discards, got %+v", h2)
	}
}

func TestFiveHundred_JokerLeadNoTrump_NominatesSuit(t *testing.T) {
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractNoTrump, 7, -1)
	g.SetDeclarerIdx(0)
	g.GetPlayer(0).AddCard(fhCard(domain.CardDesignJoker, 1))
	g.GetPlayer(1).AddCard(fhCard(domain.CardDesignHeart, 6))
	g.GetPlayer(2).AddCard(fhCard(domain.CardDesignHeart, 7))
	g.GetPlayer(3).AddCard(fhCard(domain.CardDesignHeart, 8))
	g.SetPhase(domain.FiveHundredPhasePlay)
	g.SetTrickNumber(1)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)

	if err := g.PlayerPlay(0, domain.CardDesignHeart); err != nil {
		t.Fatalf("joker lead failed: %v", err)
	}
	if g.GetJokerLeadSuit() != domain.CardDesignHeart {
		t.Errorf("joker lead suit = %d, want heart", g.GetJokerLeadSuit())
	}
	for g.GetPhase() == domain.FiveHundredPhasePlay {
		g.CpuPlay()
	}
	g.ResolveTrick()
	// Joker is the highest card in no-trump, so player 0 wins.
	if g.GetPlayer(0).GetTrickCount() != 1 {
		t.Errorf("joker should win the no-trump trick")
	}
}

func TestFiveHundred_OpenMisereScoring(t *testing.T) {
	// Made (0 tricks) -> +520.
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractOpenMisere, 0, -1)
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.FiveHundredPhaseRoundEnd)
	g.ScoreRound()
	if g.GetTeamScore(0) != 520 {
		t.Errorf("open misère made: team0 = %d, want 520", g.GetTeamScore(0))
	}

	// Failed (a trick) -> -520.
	g2 := newFiveHundredForTest()
	g2.SetContract(domain.FiveHundredContractOpenMisere, 0, -1)
	g2.SetDeclarerIdx(0)
	g2.GetPlayer(0).AddTrick([]*domain.Card{nil})
	g2.SetPhase(domain.FiveHundredPhaseRoundEnd)
	g2.ScoreRound()
	if g2.GetTeamScore(0) != -520 {
		t.Errorf("open misère failed: team0 = %d, want -520", g2.GetTeamScore(0))
	}
}

func TestFiveHundred_NextRound_AdvancesAndRotates(t *testing.T) {
	g := newFiveHundredForTest()
	g.Reset() // round 1, dealer 0
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g.SetDeclarerIdx(0)
	for i := 0; i < 5; i++ {
		g.GetPlayer(i % 2 * 2).AddTrick([]*domain.Card{nil})
	}
	for i := 0; i < 5; i++ {
		g.GetPlayer(1).AddTrick([]*domain.Card{nil})
	}
	g.SetPhase(domain.FiveHundredPhaseRoundEnd)
	g.ScoreRound()
	if g.GetGameEndFlag() {
		t.Skip("unexpected early game end")
	}
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("round number = %d, want 2", g.GetRoundNumber())
	}
	if g.GetPhase() != domain.FiveHundredPhaseBid {
		t.Errorf("phase after NextRound = %d, want Bid", g.GetPhase())
	}
	if g.GetDealerIdx() != 1 {
		t.Errorf("dealer = %d, want 1 (rotated)", g.GetDealerIdx())
	}
}

func TestFiveHundred_PlayerBid_Variants(t *testing.T) {
	g := newFiveHundredForTest()
	g.Reset()
	g.SetBidPlayerIdx(0)
	if err := g.PlayerBid(domain.FiveHundredContractNoTrump, 6, -1); err != nil {
		t.Fatalf("NT bid failed: %v", err)
	}
	if g.GetHighestBid() == nil || g.GetHighestBid().Kind != domain.FiveHundredContractNoTrump {
		t.Errorf("highest bid not set to NT")
	}

	// Misère bid path.
	g2 := newFiveHundredForTest()
	g2.Reset()
	g2.SetBidPlayerIdx(0)
	if err := g2.PlayerBid(domain.FiveHundredContractMisere, 0, -1); err != nil {
		t.Fatalf("misère bid failed: %v", err)
	}

	// Error paths.
	g3 := newFiveHundredForTest()
	g3.Reset()
	g3.SetBidPlayerIdx(0)
	if err := g3.PlayerBid(domain.FiveHundredContractSuit, 5, domain.CardDesignSpade); err == nil {
		t.Errorf("expected error for invalid (5-trick) bid")
	}
	g3.SetBidPlayerIdx(1) // CPU's turn
	if err := g3.PlayerBid(domain.FiveHundredContractNoTrump, 6, -1); err == nil {
		t.Errorf("expected ErrNotHumanTurn when it is not the human's bid turn")
	}
}

func TestFiveHundred_Unmarshal_GuardsAndOversize(t *testing.T) {
	// Empty object must be rejected: player count is not 4.
	var g domain.FiveHundred
	if err := json.Unmarshal([]byte(`{}`), &g); err == nil {
		t.Fatalf("expected error for empty JSON (no players), got nil")
	}

	// Oversize player array must be rejected.
	var g2 domain.FiveHundred
	oversize := `{"ps":[` + strings.TrimSuffix(strings.Repeat("{},", 1001), ",") + `]}`
	if err := json.Unmarshal([]byte(oversize), &g2); err == nil {
		t.Errorf("expected oversize-array error")
	}

	// Malformed JSON returns an error.
	if err := json.Unmarshal([]byte(`{`), new(domain.FiveHundred)); err == nil {
		t.Errorf("expected error for malformed JSON")
	}

	// Nil player element must be rejected.
	var g3 domain.FiveHundred
	if err := json.Unmarshal([]byte(`{"ps":[null,null,null,null]}`), &g3); err == nil {
		t.Errorf("expected error for nil player element")
	}

	// Invalid config must be rejected.
	data, _ := json.Marshal(newFiveHundredForTest())
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	raw["cf"] = json.RawMessage(`{"cd":99,"ts":0}`)
	invalidCfg, _ := json.Marshal(raw)
	var g4 domain.FiveHundred
	if err := json.Unmarshal(invalidCfg, &g4); err == nil {
		t.Errorf("expected error for invalid config")
	}
}

func TestFiveHundred_CpuBid_Difficulties(t *testing.T) {
	strong := func(p *domain.FiveHundredPlayer) {
		p.Reset()
		p.AddCard(fhCard(domain.CardDesignJoker, 1))
		p.AddCard(fhCard(domain.CardDesignSpade, 11))  // right bower
		p.AddCard(fhCard(domain.CardDesignClover, 11)) // left bower
		p.AddCard(fhCard(domain.CardDesignSpade, 1))
		p.AddCard(fhCard(domain.CardDesignSpade, 13))
		p.AddCard(fhCard(domain.CardDesignSpade, 12))
		p.AddCard(fhCard(domain.CardDesignSpade, 10))
		p.AddCard(fhCard(domain.CardDesignSpade, 9))
		p.AddCard(fhCard(domain.CardDesignHeart, 1))
		p.AddCard(fhCard(domain.CardDesignDiamond, 1))
	}
	for _, diff := range []domain.FiveHundredCpuDifficulty{
		domain.FiveHundredCpuDifficultyEasy,
		domain.FiveHundredCpuDifficultyNormal,
		domain.FiveHundredCpuDifficultyHard,
	} {
		g := newFiveHundredForTest()
		cfg := domain.DefaultFiveHundredConfig()
		cfg.CpuDifficulty = diff
		g.SetConfig(cfg)
		g.Reset()
		strong(g.GetPlayer(1))
		g.SetPhase(domain.FiveHundredPhaseBid)
		g.SetBidPlayerIdx(1)
		g.CpuBid() // exercises the difficulty branch in cpuSelectBid
		if g.GetHighestBid() == nil {
			t.Errorf("difficulty %d: strong CPU hand should bid", diff)
		}
	}
}

func TestFiveHundred_GetHint_StrongBidAndJokerLead(t *testing.T) {
	// Strong hand on the human's bid turn -> a concrete bid hint.
	g := newFiveHundredForTest()
	g.SetPhase(domain.FiveHundredPhaseBid)
	g.SetBidPlayerIdx(0)
	for _, c := range []*domain.Card{
		fhCard(domain.CardDesignJoker, 1),
		fhCard(domain.CardDesignSpade, 11),
		fhCard(domain.CardDesignClover, 11),
		fhCard(domain.CardDesignSpade, 1),
		fhCard(domain.CardDesignSpade, 13),
		fhCard(domain.CardDesignSpade, 12),
		fhCard(domain.CardDesignSpade, 10),
		fhCard(domain.CardDesignSpade, 9),
		fhCard(domain.CardDesignHeart, 1),
		fhCard(domain.CardDesignDiamond, 1),
	} {
		g.GetPlayer(0).AddCard(c)
	}
	hint := g.GetHint()
	if hint == nil || hint.BidKind == nil {
		t.Errorf("strong hand should yield a bid hint, got %+v", hint)
	}

	// No-trump play hint when the human leads holding the joker -> JokerSuit set.
	g2 := newFiveHundredForTest()
	g2.SetContract(domain.FiveHundredContractNoTrump, 7, -1)
	g2.GetPlayer(0).AddCard(fhCard(domain.CardDesignJoker, 1))
	g2.SetPhase(domain.FiveHundredPhasePlay)
	g2.SetCurrentPlayerIdx(0)
	h2 := g2.GetHint()
	if h2 == nil || h2.CardIndex == nil || h2.JokerSuit == nil {
		t.Errorf("joker-lead play hint should set JokerSuit, got %+v", h2)
	}
}

func TestFiveHundred_Redeal_AllPass(t *testing.T) {
	g := newFiveHundredForTest()
	g.Reset()
	// Give everyone a weak hand so all CPUs (and the human) pass.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		p.Reset()
		for v := 5; v <= 9; v++ {
			p.AddCard(fhCard(domain.CardDesignHeart, v))
			p.AddCard(fhCard(domain.CardDesignDiamond, v))
		}
	}
	g.SetPhase(domain.FiveHundredPhaseBid)
	g.SetBidPlayerIdx(1)

	steps := 0
	for steps < 8 {
		idx := g.GetBidPlayerIdx()
		if g.GetPlayer(idx).GetIsHuman() {
			_ = g.PlayerPass()
		} else {
			g.CpuBid()
		}
		steps++
		redealt := false
		for _, e := range g.GetActionLog() {
			if e.ActionType == "redeal" {
				redealt = true
			}
		}
		if redealt {
			return // redeal branch exercised
		}
	}
	t.Errorf("expected an all-pass redeal")
}

func TestFiveHundred_CpuPlay_FollowScenarios(t *testing.T) {
	// CPU player 1 (team 1) follows when its partner (player 3, team 1) is
	// already winning the trick -> it sheds its weakest card.
	g := newFiveHundredForTest()
	g.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: fhCard(domain.CardDesignHeart, 1)}, // partner of player 1 winning (heart A)
	})
	g.GetPlayer(1).AddCard(fhCard(domain.CardDesignHeart, 5))
	g.GetPlayer(1).AddCard(fhCard(domain.CardDesignHeart, 6))
	g.SetPhase(domain.FiveHundredPhasePlay)
	g.SetTrickNumber(1)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	if len(g.GetCurrentTrick()) != 2 {
		t.Errorf("CPU should follow when partner is winning")
	}

	// CPU over-cards when it can beat the current winner.
	g2 := newFiveHundredForTest()
	g2.SetContract(domain.FiveHundredContractSuit, 7, domain.CardDesignSpade)
	g2.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: fhCard(domain.CardDesignHeart, 10)},
	})
	g2.GetPlayer(1).AddCard(fhCard(domain.CardDesignHeart, 1)) // can win
	g2.GetPlayer(1).AddCard(fhCard(domain.CardDesignHeart, 5)) // or duck
	g2.SetPhase(domain.FiveHundredPhasePlay)
	g2.SetTrickNumber(1)
	g2.SetCurrentPlayerIdx(1)
	g2.CpuPlay()
	if len(g2.GetCurrentTrick()) != 2 {
		t.Errorf("CPU should have played a card")
	}
}

func TestFiveHundred_AccessorsAndOpenMisereBid(t *testing.T) {
	g := newFiveHundredForTest()
	g.Reset()
	// Exercise the test-only setters/getters.
	g.SetTrickNumber(3)
	g.SetCurrentPlayerIdx(2)
	g.SetLeadPlayerIdx(1)
	g.SetDealerIdx(2)
	g.SetTrumpSuit(domain.CardDesignHeart)
	if g.GetTrickNumber() != 3 || g.GetCurrentPlayerIdx() != 2 || g.GetLeadPlayerIdx() != 1 ||
		g.GetDealerIdx() != 2 || g.GetTrumpSuit() != domain.CardDesignHeart {
		t.Errorf("setter/getter round trip mismatch")
	}
	g.SetContract(domain.FiveHundredContractNoTrump, 8, -1)
	if g.GetContractTricks() != 8 || g.GetContractValue() != 320 {
		t.Errorf("NT contract getters wrong: tricks=%d value=%d", g.GetContractTricks(), g.GetContractValue())
	}
	_ = g.GetJokerLeadSuit()
	_ = g.GetHighestBidder()
	if g.GetTeamScore(99) != 0 {
		t.Errorf("out-of-range team score should be 0")
	}

	// Open-misère bid via the human path exercises the open-misère bid label.
	g2 := newFiveHundredForTest()
	g2.Reset()
	g2.SetBidPlayerIdx(0)
	if err := g2.PlayerBid(domain.FiveHundredContractOpenMisere, 0, -1); err != nil {
		t.Fatalf("open misère bid failed: %v", err)
	}
	if g2.GetHighestBid() == nil || g2.GetHighestBid().Kind != domain.FiveHundredContractOpenMisere {
		t.Errorf("open misère bid not recorded")
	}
	// A lower bid afterwards must be rejected (covers the not-higher branch).
	g2.SetBidPlayerIdx(0)
	if err := g2.PlayerBid(domain.FiveHundredContractSuit, 6, domain.CardDesignSpade); err == nil {
		t.Errorf("expected rejection of a lower bid after open misère")
	}
}

// #5626: 指名スートは画面が読む値なので、設定と取得が同じフィールドを指すこと
// だけは固定しておく (取り違えると、指名と違うスートを画面が出す)。
func TestFiveHundredJokerLeadSuitRoundTrips(t *testing.T) {
	g := newFiveHundredForTest()
	g.Reset()
	// Reset 直後は「指名なし」。
	assert.Equal(t, -1, g.GetJokerLeadSuit())

	g.SetJokerLeadSuit(domain.CardDesignHeart)
	assert.Equal(t, domain.CardDesignHeart, g.GetJokerLeadSuit())
}
