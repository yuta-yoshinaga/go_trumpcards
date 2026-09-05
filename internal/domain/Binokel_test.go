//go:build test

package domain

import (
	"encoding/json"
	"fmt"
	"testing"
)

func newTestBinokel(cfg ...BinokelConfig) *Binokel {
	config := DefaultBinokelConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}
	players := []*BinokelPlayer{
		NewBinokelPlayer(true),
		NewBinokelPlayer(false),
		NewBinokelPlayer(false),
	}
	return NewBinokel(NewTrumpCardsBinokel(), players, config)
}

// ─── 1. Deck Verification ─────────────────────────────────

func TestBinokel_Deck(t *testing.T) {
	deck := NewTrumpCardsBinokel()
	if deck.GetTotalCount() != 48 {
		t.Fatalf("expected 48 cards, got %d", deck.GetTotalCount())
	}

	type sv struct{ suit, value int }
	counts := make(map[sv]int)
	for i := 0; i < 48; i++ {
		card := deck.DrawCard()
		if card == nil {
			t.Fatalf("expected card at index %d, got nil", i)
		}
		counts[sv{card.GetDesign(), card.GetValue()}]++
	}

	// 49枚目はnil
	if deck.DrawCard() != nil {
		t.Error("expected nil for 49th draw")
	}

	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}

	// 7 は各スート 2 枚 (計 8 枚)
	totalSevens := 0
	for _, s := range suits {
		c7 := counts[sv{s, 7}]
		if c7 != 2 {
			t.Errorf("suit %d: expected 2 sevens, got %d", s, c7)
		}
		totalSevens += c7
	}
	if totalSevens != 8 {
		t.Errorf("expected 8 sevens total, got %d", totalSevens)
	}

	// 9 は 1 枚も含まれない (Len == 0)
	for _, s := range suits {
		if c9 := counts[sv{s, 9}]; c9 != 0 {
			t.Errorf("suit %d: expected 0 nines, got %d", s, c9)
		}
	}

	// 各スート・ランクの組み合わせが正確に 2 枚ずつ (A=1, 7, 10, J=11, Q=12, K=13)
	expectedValues := []int{1, 7, 10, 11, 12, 13}
	for _, s := range suits {
		for _, v := range expectedValues {
			if counts[sv{s, v}] != 2 {
				t.Errorf("suit %d value %d: expected 2 cards, got %d", s, v, counts[sv{s, v}])
			}
		}
	}
}

// ─── 2. Config & Player ───────────────────────────────────

func TestBinokelConfig_Default(t *testing.T) {
	config := DefaultBinokelConfig()
	if config.CpuDifficulty != BinokelCpuDifficultyNormal {
		t.Errorf("expected Normal difficulty, got %d", config.CpuDifficulty)
	}
	if config.PointLimit != 1000 {
		t.Errorf("expected 1000 point limit, got %d", config.PointLimit)
	}
	if err := config.Validate(); err != nil {
		t.Errorf("default config should be valid, got: %v", err)
	}
}

func TestBinokelConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BinokelConfig
		wantErr bool
	}{
		{"valid default", DefaultBinokelConfig(), false},
		{"valid easy", BinokelConfig{CpuDifficulty: BinokelCpuDifficultyEasy, PointLimit: 100}, false},
		{"valid hard", BinokelConfig{CpuDifficulty: BinokelCpuDifficultyHard, PointLimit: 5000}, false},
		{"invalid difficulty low", BinokelConfig{CpuDifficulty: -1, PointLimit: 1000}, true},
		{"invalid difficulty high", BinokelConfig{CpuDifficulty: 99, PointLimit: 1000}, true},
		{"invalid point limit zero", BinokelConfig{CpuDifficulty: BinokelCpuDifficultyNormal, PointLimit: 0}, true},
		{"invalid point limit negative", BinokelConfig{CpuDifficulty: BinokelCpuDifficultyNormal, PointLimit: -100}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBinokelPlayer_Basics(t *testing.T) {
	p := NewBinokelPlayer(true, 0)
	if !p.GetIsHuman() {
		t.Error("expected human player")
	}
	if p.GetTeam() != 0 {
		t.Errorf("expected team 0 (compat), got %d", p.GetTeam())
	}

	// Cards
	c1 := NewCard(CardDesignSpade, 1, false)
	c2 := NewCard(CardDesignHeart, 10, false)
	p.AddCard(c1)
	p.AddCard(c2)
	if p.GetCardsSize() != 2 {
		t.Fatalf("expected 2 cards, got %d", p.GetCardsSize())
	}
	if p.GetCard(0) != c1 || p.GetCard(1) != c2 {
		t.Error("card mismatch")
	}
	if p.GetCard(-1) != nil || p.GetCard(2) != nil {
		t.Error("expected nil for out-of-range card index")
	}
	removed := p.RemoveCard(0)
	if removed != c1 || p.GetCardsSize() != 1 {
		t.Error("failed to remove card at index 0")
	}
	if p.RemoveCard(-1) != nil || p.RemoveCard(5) != nil {
		t.Error("expected nil when removing out-of-range index")
	}

	// Bids & status
	p.SetBid(160)
	if p.GetBid() != 160 {
		t.Errorf("expected bid 160, got %d", p.GetBid())
	}
	p.SetHasPassed(true)
	if !p.GetHasPassed() {
		t.Error("expected hasPassed true")
	}
	p.SetMeldScore(120)
	if p.GetMeldScore() != 120 {
		t.Errorf("expected meld score 120, got %d", p.GetMeldScore())
	}
	p.SetTrickPoints(85)
	if p.GetTrickPoints() != 85 {
		t.Errorf("expected trick points 85, got %d", p.GetTrickPoints())
	}

	trick := []*Card{c1, c2}
	p.AddTrick(trick)
	if p.GetTrickCount() != 1 || len(p.GetTricksTaken()) != 1 {
		t.Errorf("expected 1 trick taken, got %d", p.GetTrickCount())
	}

	// ResetRound
	p.ResetRound()
	if p.GetCardsSize() != 0 {
		t.Error("expected 0 cards after ResetRound")
	}
	if p.GetBid() != 0 {
		t.Error("expected bid 0 after ResetRound")
	}
	if p.GetHasPassed() {
		t.Error("expected hasPassed false after ResetRound")
	}
	if p.GetMeldScore() != 0 {
		t.Error("expected meld score 0 after ResetRound")
	}
	if p.GetTrickPoints() != 0 {
		t.Error("expected trick points 0 after ResetRound")
	}
	if p.GetTrickCount() != 0 {
		t.Error("expected 0 tricks taken after ResetRound")
	}
}

func TestBinokelPlayer_JSON(t *testing.T) {
	p := NewBinokelPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.SetBid(170)
	p.SetMeldScore(40)
	p.SetTrickPoints(60)

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	p2 := &BinokelPlayer{}
	if err := json.Unmarshal(data, p2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !p2.GetIsHuman() || p2.GetBid() != 170 || p2.GetMeldScore() != 40 || p2.GetTrickPoints() != 60 || p2.GetCardsSize() != 1 {
		t.Errorf("unmarshaled player state mismatch: %+v", p2)
	}

	// Invalid JSON
	if err := json.Unmarshal([]byte("invalid json"), &BinokelPlayer{}); err == nil {
		t.Error("expected unmarshal error for invalid JSON")
	}
}

// ─── 3. Deal Verification ─────────────────────────────────

func TestBinokel_Deal(t *testing.T) {
	game := newTestBinokel()
	game.Reset()

	// 3 プレイヤーの手札が各 15 枚、Dabb が 3 枚、計 48 枚
	totalCards := 0
	for i := 0; i < BinokelPlayerCnt; i++ {
		size := game.GetPlayer(i).GetCardsSize()
		if size != BinokelHandSize {
			t.Errorf("player %d hand size: expected %d, got %d", i, BinokelHandSize, size)
		}
		totalCards += size
	}
	dabb := game.GetDabb()
	if len(dabb) != BinokelDabbSize {
		t.Errorf("dabb size: expected %d, got %d", BinokelDabbSize, len(dabb))
	}
	totalCards += len(dabb)
	if totalCards != 48 {
		t.Fatalf("expected 48 total cards, got %d", totalCards)
	}

	// カードの重複や欠落がない
	type sv struct{ suit, value int }
	counts := make(map[sv]int)
	for i := 0; i < BinokelPlayerCnt; i++ {
		p := game.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			counts[sv{c.GetDesign(), c.GetValue()}]++
		}
	}
	for _, c := range dabb {
		counts[sv{c.GetDesign(), c.GetValue()}]++
	}

	expectedValues := []int{1, 7, 10, 11, 12, 13}
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for _, s := range suits {
		for _, v := range expectedValues {
			if counts[sv{s, v}] != 2 {
				t.Errorf("suit %d value %d: expected exactly 2 cards across deal, got %d", s, v, counts[sv{s, v}])
			}
		}
	}
}

// ─── 4. Dabb Exchange and Discard ─────────────────────────

func TestBinokel_DabbExchangeAndDiscard(t *testing.T) {
	game := newTestBinokel()
	game.Reset()
	// Set seat 0 to be first bidder
	game.dealerIdx = 2
	game.bidPlayerIdx = 0

	// Human bids, CPU passes -> Human wins auction
	if err := game.PlayerBid(150); err != nil {
		t.Fatalf("PlayerBid failed: %v", err)
	}
	_ = game.doPass(1)
	_ = game.doPass(2)

	// Should be in Dabb phase
	if game.GetPhase() != BinokelPhaseDabb {
		t.Fatalf("expected phase Dabb, got %d", game.GetPhase())
	}
	if game.GetHighestBidder() != 0 {
		t.Fatalf("expected highest bidder 0, got %d", game.GetHighestBidder())
	}
	if !game.IsHumanDabbTurn() {
		t.Error("expected IsHumanDabbTurn to be true")
	}

	// ビッド勝者が Dabb 3 枚を受け取って手札 18 枚になること
	p0 := game.GetPlayer(0)
	if p0.GetCardsSize() != 18 {
		t.Fatalf("expected hand size 18 after Dabb, got %d", p0.GetCardsSize())
	}

	// 不正な捨て札: 枚数が3枚でない (2枚)
	if err := game.PlayerDiscardToDabb([]int{0, 1}); err == nil {
		t.Error("expected error when discarding 2 cards")
	}
	// 不正な捨て札: 枚数が3枚でない (4枚)
	if err := game.PlayerDiscardToDabb([]int{0, 1, 2, 3}); err == nil {
		t.Error("expected error when discarding 4 cards")
	}
	// 不正な捨て札: 範囲外インデックス
	if err := game.PlayerDiscardToDabb([]int{0, 1, 99}); err == nil {
		t.Error("expected error for out of range index")
	}
	if err := game.PlayerDiscardToDabb([]int{-1, 1, 2}); err == nil {
		t.Error("expected error for negative index")
	}
	// 不正な捨て札: 重複インデックス
	if err := game.PlayerDiscardToDabb([]int{1, 1, 2}); err == nil {
		t.Error("expected error for duplicate index")
	}

	// 正常な捨て札: 3枚選んで捨てる
	c0 := p0.GetCard(0)
	c1 := p0.GetCard(1)
	c2 := p0.GetCard(2)
	expectedDiscardPoints := binokelCardPointValue(c0) + binokelCardPointValue(c1) + binokelCardPointValue(c2)

	if err := game.PlayerDiscardToDabb([]int{0, 1, 2}); err != nil {
		t.Fatalf("PlayerDiscardToDabb failed: %v", err)
	}

	// 手札が15枚に戻ること
	if p0.GetCardsSize() != 15 {
		t.Fatalf("expected hand size 15 after discard, got %d", p0.GetCardsSize())
	}
	// フェーズがTrump宣言フェーズへ進むこと
	if game.GetPhase() != BinokelPhaseTrump {
		t.Fatalf("expected phase Trump, got %d", game.GetPhase())
	}

	// 捨て札の検証
	discarded := game.GetDabbDiscarded()
	if len(discarded) != 3 {
		t.Fatalf("expected 3 discarded cards, got %d", len(discarded))
	}
	actualPoints := 0
	for _, c := range discarded {
		actualPoints += binokelCardPointValue(c)
	}
	if actualPoints != expectedDiscardPoints {
		t.Errorf("discard points: expected %d, got %d", expectedDiscardPoints, actualPoints)
	}
}

func TestBinokel_CpuDabbDiscard(t *testing.T) {
	game := newTestBinokel()
	game.Reset()
	// Round 1: dealer 0, player 1 bids first
	_ = game.doBid(1, 150)
	_ = game.doPass(2)
	_ = game.doPass(0)

	if game.GetPhase() != BinokelPhaseDabb {
		t.Fatalf("expected phase Dabb, got %d", game.GetPhase())
	}
	if game.GetHighestBidder() != 1 {
		t.Fatalf("expected highest bidder 1, got %d", game.GetHighestBidder())
	}
	if game.IsHumanDabbTurn() {
		t.Error("expected IsHumanDabbTurn false for CPU bidder")
	}

	p1 := game.GetPlayer(1)
	if p1.GetCardsSize() != 18 {
		t.Fatalf("expected hand size 18, got %d", p1.GetCardsSize())
	}

	game.CpuDiscardToDabb()

	if p1.GetCardsSize() != 15 {
		t.Fatalf("expected hand size 15, got %d", p1.GetCardsSize())
	}
	if len(game.GetDabbDiscarded()) != 3 {
		t.Fatalf("expected 3 discarded cards, got %d", len(game.GetDabbDiscarded()))
	}
	if game.GetPhase() != BinokelPhaseTrump {
		t.Errorf("expected phase Trump, got %d", game.GetPhase())
	}
}

// ─── 5. Melds Verification ────────────────────────────────

func TestBinokel_EvaluateMelds_Dix(t *testing.T) {
	trumpSuit := CardDesignSpade

	// Dix = トランプの 7 (10点)
	hand1 := []*Card{
		NewCard(CardDesignSpade, 7, false),
	}
	melds1 := evaluateBinokelMelds(hand1, trumpSuit)
	if len(melds1) != 1 || melds1[0].Type != BinokelMeldDix || melds1[0].Points != 10 {
		t.Fatalf("expected Dix meld with 10 pts, got: %+v", melds1)
	}

	// 9 はトランプであっても Dix ではない (0点)
	hand9 := []*Card{
		NewCard(CardDesignSpade, 9, false),
	}
	melds9 := evaluateBinokelMelds(hand9, trumpSuit)
	if len(melds9) != 0 {
		t.Fatalf("expected no melds for 9, got: %+v", melds9)
	}

	// 非トランプの 7 は 0点
	handNonTrump7 := []*Card{
		NewCard(CardDesignHeart, 7, false),
	}
	meldsNonTrump := evaluateBinokelMelds(handNonTrump7, trumpSuit)
	if len(meldsNonTrump) != 0 {
		t.Fatalf("expected 0 melds for non-trump 7, got: %+v", meldsNonTrump)
	}

	// 2枚の Dix (トランプ 7 が 2枚) = 20点
	handDouble := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 7, false),
	}
	meldsDouble := evaluateBinokelMelds(handDouble, trumpSuit)
	if len(meldsDouble) != 2 || binokelMeldTotalPoints(meldsDouble) != 20 {
		t.Fatalf("expected 2 Dix melds with 20 pts total, got: %+v", meldsDouble)
	}
}

func TestBinokel_EvaluateMelds_RunAndFamily(t *testing.T) {
	trumpSuit := CardDesignSpade

	// Trump Run (A-10-K-Q-J of trump): 150 点
	trumpRunHand := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
	}
	trumpMelds := evaluateBinokelMelds(trumpRunHand, trumpSuit)
	foundTrumpRun := false
	for _, m := range trumpMelds {
		if m.Type == BinokelMeldRun && m.Points == 150 {
			foundTrumpRun = true
		}
	}
	if !foundTrumpRun {
		t.Fatalf("expected Trump Run with 150 points, got: %+v", trumpMelds)
	}

	// Non-trump Run / Family (A-10-K-Q-J of Heart): 100 点
	nonTrumpRunHand := []*Card{
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignHeart, 11, false),
	}
	nonTrumpMelds := evaluateBinokelMelds(nonTrumpRunHand, trumpSuit)
	foundNonTrumpRun := false
	for _, m := range nonTrumpMelds {
		if m.Type == BinokelMeldNonTrumpRun && m.Points == 100 {
			foundNonTrumpRun = true
		}
	}
	if !foundNonTrumpRun {
		t.Fatalf("expected Non-Trump Run with 100 points, got: %+v", nonTrumpMelds)
	}

	// Double Non-trump Run: 1000 点
	doubleNonTrumpHand := append(nonTrumpRunHand, []*Card{
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignHeart, 11, false),
	}...)
	doubleNonTrumpMelds := evaluateBinokelMelds(doubleNonTrumpHand, trumpSuit)
	foundDoubleNonTrump := false
	for _, m := range doubleNonTrumpMelds {
		if m.Type == BinokelMeldDoubleNonTrumpRun && m.Points == 1000 {
			foundDoubleNonTrump = true
		}
	}
	if !foundDoubleNonTrump {
		t.Fatalf("expected Double Non-Trump Run with 1000 points, got: %+v", doubleNonTrumpMelds)
	}
}

func TestBinokel_EvaluateMelds_Rundgang(t *testing.T) {
	trumpSuit := CardDesignSpade

	// Rundgang: 4 スートの K+Q
	rundgangHand := []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignClover, 12, false),
	}

	melds := evaluateBinokelMelds(rundgangHand, trumpSuit)

	// ルントガングは 240 点であり、個別のペア (40 + 20 + 20 + 20 = 100) を二重加算しないこと
	foundRundgang := false
	for _, m := range melds {
		if m.Type == BinokelMeldRundgang {
			foundRundgang = true
			if m.Points != 240 {
				t.Errorf("expected Rundgang to be 240 points, got %d", m.Points)
			}
		}
		// 個別マリッジ (Royal / Common) が重複して含まれていないこと
		if m.Type == BinokelMeldRoyalMarriage || m.Type == BinokelMeldCommonMarriage {
			t.Errorf("individual marriage should NOT be included with Rundgang, found: %v", m.Type)
		}
	}
	if !foundRundgang {
		t.Error("expected Rundgang meld not found")
	}

	// マリッジ群としての点数が 240 点であり、個別のペア (100点) と二重加算されて 340点などにならないことを検証
	marriageCategoryPoints := 0
	for _, m := range melds {
		if m.Type == BinokelMeldRundgang || m.Type == BinokelMeldRoyalMarriage || m.Type == BinokelMeldCommonMarriage {
			marriageCategoryPoints += m.Points
		}
	}
	if marriageCategoryPoints != 240 {
		t.Errorf("expected marriage category points to be exactly 240, got %d", marriageCategoryPoints)
	}
}

func TestBinokel_EvaluateMelds_OtherMelds(t *testing.T) {
	trumpSuit := CardDesignSpade

	// Royal Marriage (Spade K+Q): 40 pts
	royalMarriageHand := []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
	}
	rmMelds := evaluateBinokelMelds(royalMarriageHand, trumpSuit)
	if binokelMeldTotalPoints(rmMelds) != 40 {
		t.Errorf("expected Royal Marriage 40 pts, got %d", binokelMeldTotalPoints(rmMelds))
	}

	// Common Marriage (Heart K+Q): 20 pts
	commonMarriageHand := []*Card{
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 12, false),
	}
	cmMelds := evaluateBinokelMelds(commonMarriageHand, trumpSuit)
	if binokelMeldTotalPoints(cmMelds) != 20 {
		t.Errorf("expected Common Marriage 20 pts, got %d", binokelMeldTotalPoints(cmMelds))
	}

	// Binokel (Spade Q + Diamond J): 40 pts
	binokelHand := []*Card{
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignDiamond, 11, false),
	}
	bMelds := evaluateBinokelMelds(binokelHand, trumpSuit)
	if binokelMeldTotalPoints(bMelds) != 40 {
		t.Errorf("expected Binokel 40 pts, got %d", binokelMeldTotalPoints(bMelds))
	}

	// Double Binokel: 300 pts
	doubleBinokelHand := append(binokelHand, []*Card{
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignDiamond, 11, false),
	}...)
	dbMelds := evaluateBinokelMelds(doubleBinokelHand, trumpSuit)
	if binokelMeldTotalPoints(dbMelds) != 300 {
		t.Errorf("expected Double Binokel 300 pts, got %d", binokelMeldTotalPoints(dbMelds))
	}

	// Arounds: Aces Around (100 pts), Double Aces Around (1000 pts)
	acesHand := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignDiamond, 1, false),
	}
	acesMelds := evaluateBinokelMelds(acesHand, trumpSuit)
	if binokelMeldTotalPoints(acesMelds) != 100 {
		t.Errorf("expected Aces Around 100 pts, got %d", binokelMeldTotalPoints(acesMelds))
	}
}

// ─── 6. Scoring Asymmetry ─────────────────────────────────

func TestBinokel_ScoringAsymmetry(t *testing.T) {
	// ビッダーと非ビッダーを同一テーブルで検証する
	// Scenario A: Bidder achieves bid -> adds meld + trick points
	t.Run("Bidder achieves bid", func(t *testing.T) {
		game := newTestBinokel()
		game.Reset()

		// Bidder = 0, bid = 150
		game.highestBidder = 0
		game.highestBid = 150

		// Player 0 (Bidder): meld 60, trickPoints 80, discardedPoints 20 -> total 160 >= 150
		game.players[0].SetMeldScore(60)
		game.players[0].SetTrickPoints(80)
		game.dabbDiscarded = []*Card{
			NewCard(CardDesignSpade, 10, false), // 10 pts
			NewCard(CardDesignHeart, 10, false), // 10 pts
			NewCard(CardDesignHeart, 7, false),  // 0 pts
		} // 20 pts from Dabb discard

		// Player 1 (Non-bidder): meld 20, trickPoints 70 -> total 90
		game.players[1].SetMeldScore(20)
		game.players[1].SetTrickPoints(70)

		// Player 2 (Non-bidder): meld 40, trickPoints 40 -> total 80
		game.players[2].SetMeldScore(40)
		game.players[2].SetTrickPoints(40)

		game.scoreRound()

		// Bidder: 60 + 80 + 20 = 160 points added
		if game.GetScore(0) != 160 {
			t.Errorf("player 0 (bidder) score: expected 160, got %d", game.GetScore(0))
		}
		// Non-bidder 1: 20 + 70 = 90 points added
		if game.GetScore(1) != 90 {
			t.Errorf("player 1 (non-bidder) score: expected 90, got %d", game.GetScore(1))
		}
		// Non-bidder 2: 40 + 40 = 80 points added
		if game.GetScore(2) != 80 {
			t.Errorf("player 2 (non-bidder) score: expected 80, got %d", game.GetScore(2))
		}
	})

	// Scenario B: Bidder fails bid -> loses bid amount (-bid), non-bidders still gain their points
	t.Run("Bidder fails bid", func(t *testing.T) {
		game := newTestBinokel()
		game.Reset()

		// Bidder = 1, bid = 200
		game.highestBidder = 1
		game.highestBid = 200

		// Player 0 (Non-bidder): meld 40, trickPoints 50 -> total 90
		game.players[0].SetMeldScore(40)
		game.players[0].SetTrickPoints(50)

		// Player 1 (Bidder): meld 50, trickPoints 80, discardedPoints 10 -> total 140 < 200 (FAILED)
		game.players[1].SetMeldScore(50)
		game.players[1].SetTrickPoints(80)
		game.dabbDiscarded = []*Card{
			NewCard(CardDesignSpade, 10, false), // 10 pts
			NewCard(CardDesignHeart, 7, false),  // 0 pts
			NewCard(CardDesignSpade, 7, false),  // 0 pts
		} // 10 pts from Dabb discard

		// Player 2 (Non-bidder): meld 20, trickPoints 70 -> total 90
		game.players[2].SetMeldScore(20)
		game.players[2].SetTrickPoints(70)

		game.scoreRound()

		// Player 0 (Non-bidder): gains 90
		if game.GetScore(0) != 90 {
			t.Errorf("player 0 (non-bidder) score: expected 90, got %d", game.GetScore(0))
		}
		// Player 1 (Bidder failed): loses 200 (-200)
		if game.GetScore(1) != -200 {
			t.Errorf("player 1 (failed bidder) score: expected -200, got %d", game.GetScore(1))
		}
		// Player 2 (Non-bidder): gains 90
		if game.GetScore(2) != 90 {
			t.Errorf("player 2 (non-bidder) score: expected 90, got %d", game.GetScore(2))
		}
	})
}

// ─── 7. Statistical Auction Verification (2000 trials) ────

func TestBinokel_Auction_Statistical2000(t *testing.T) {
	const trials = 2000

	handsEstimate150Plus := 0
	totalHands := trials * BinokelPlayerCnt

	winnerSeats := [BinokelPlayerCnt]int{}
	forcedWinners := 0
	seat0VisitedCount := 0

	for trial := 0; trial < trials; trial++ {
		game := newTestBinokel()
		game.Reset()
		// Rotate dealer each trial
		game.roundNumber = trial + 1
		game.dealerIdx = trial % BinokelPlayerCnt
		game.bidPlayerIdx = (game.dealerIdx + 1) % BinokelPlayerCnt

		// Estimate >= 150 check for all 3 players
		for i := 0; i < BinokelPlayerCnt; i++ {
			est := game.cpuEstimateBid(i)
			if est >= BinokelMinBid {
				handsEstimate150Plus++
			}
		}

		// Run auction until completed
		seat0HadTurn := false
		for game.GetPhase() == BinokelPhaseBid {
			if game.IsHumanBidTurn() {
				seat0HadTurn = true
				est := game.cpuEstimateBid(0)
				nextBid := game.GetHighestBid() + BinokelBidStep
				if game.GetHighestBid() == 0 {
					nextBid = BinokelMinBid
				}
				if est >= nextBid {
					_ = game.PlayerBid(nextBid)
				} else {
					_ = game.PlayerPass()
				}
			} else {
				game.CpuBid()
			}
		}

		if seat0HadTurn {
			seat0VisitedCount++
		}

		winner := game.GetHighestBidder()
		if winner >= 0 && winner < BinokelPlayerCnt {
			winnerSeats[winner]++
		}

		// Check if forced winner (all passed):
		if game.GetPlayer(winner).GetHasPassed() {
			forcedWinners++
		}
	}

	estRate := float64(handsEstimate150Plus) / float64(totalHands) * 100.0
	forcedRate := float64(forcedWinners) / float64(trials) * 100.0
	seat0WinRate := float64(winnerSeats[0]) / float64(trials) * 100.0
	seat1WinRate := float64(winnerSeats[1]) / float64(trials) * 100.0
	seat2WinRate := float64(winnerSeats[2]) / float64(trials) * 100.0

	t.Logf("=== 2000-Trial Auction Statistics ===")
	t.Logf("Hands with Estimate >= 150: %d / %d (%.2f%%)", handsEstimate150Plus, totalHands, estRate)
	t.Logf("Forced Winner (All Pass) Rate: %d / %d (%.2f%%)", forcedWinners, trials, forcedRate)
	t.Logf("Seat 0 Win Rate: %d / %d (%.2f%%)", winnerSeats[0], trials, seat0WinRate)
	t.Logf("Seat 1 Win Rate: %d / %d (%.2f%%)", winnerSeats[1], trials, seat1WinRate)
	t.Logf("Seat 2 Win Rate: %d / %d (%.2f%%)", winnerSeats[2], trials, seat2WinRate)
	t.Logf("Seat 0 Turn Visited: %d / %d", seat0VisitedCount, trials)

	// 各プレイヤーの手札見積もり値 (CPU bid estimate) が 150 (最低ビッド) 以上になる割合が 15%〜45% の範囲
	if estRate < 15.0 || estRate > 45.0 {
		t.Errorf("expected estimate >= 150 rate in [15%%, 45%%], got %.2f%%", estRate)
	}

	// 最終的なビッド勝者の席番号 (0, 1, 2) の分布が、各席 15%〜50% の範囲に収まる
	rates := []float64{seat0WinRate, seat1WinRate, seat2WinRate}
	for i, r := range rates {
		if r < 15.0 || r > 50.0 {
			t.Errorf("seat %d win rate expected in [15%%, 50%%], got %.2f%%", i, r)
		}
	}

	// 全員パスによる強制勝者率が 30% 以下
	if forcedRate > 30.0 {
		t.Errorf("forced winner rate expected <= 30%%, got %.2f%%", forcedRate)
	}

	// 席 0 (プレイヤー) がビッド勝者になる割合が 15% 以上
	if seat0WinRate < 15.0 {
		t.Errorf("seat 0 win rate expected >= 15%%, got %.2f%%", seat0WinRate)
	}

	// 各試行で席 0 のビッド手番が最低 1 回は回ってくること
	if seat0VisitedCount != trials {
		t.Errorf("expected seat 0 to be visited in all %d trials, got %d", trials, seat0VisitedCount)
	}
}

// ─── 8. JSON Roundtrip & Mutation ─────────────────────────

func TestBinokel_JSONRoundTrip(t *testing.T) {
	game := newTestBinokel()
	game.Reset()
	game.dealerIdx = 2
	game.bidPlayerIdx = 0

	// Advance to Dabb phase
	_ = game.PlayerBid(150)
	game.CpuBid()
	game.CpuBid()

	data, err := json.Marshal(game)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	game2 := &Binokel{}
	if err := json.Unmarshal(data, game2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if game2.GetPhase() != game.GetPhase() {
		t.Errorf("phase mismatch: expected %d, got %d", game.GetPhase(), game2.GetPhase())
	}
	if game2.GetHighestBid() != game.GetHighestBid() {
		t.Errorf("highest bid mismatch: expected %d, got %d", game.GetHighestBid(), game2.GetHighestBid())
	}
	if game2.GetHighestBidder() != game.GetHighestBidder() {
		t.Errorf("highest bidder mismatch: expected %d, got %d", game.GetHighestBidder(), game2.GetHighestBidder())
	}
	if len(game2.GetDabb()) != len(game.GetDabb()) {
		t.Errorf("dabb length mismatch: expected %d, got %d", len(game.GetDabb()), len(game2.GetDabb()))
	}
	for i := 0; i < BinokelPlayerCnt; i++ {
		if game2.GetPlayer(i).GetCardsSize() != game.GetPlayer(i).GetCardsSize() {
			t.Errorf("player %d hand size mismatch", i)
		}
	}
}

func TestBinokel_JSONMutation(t *testing.T) {
	// Test 1: Invalid JSON syntax fails unmarshal
	game := &Binokel{}
	if err := json.Unmarshal([]byte("{invalid-json}"), game); err == nil {
		t.Error("expected error unmarshaling invalid JSON")
	}

	// Test 2: Incomplete JSON with mutated/missing fields
	validGame := newTestBinokel()
	validGame.Reset()
	// Set highestBidder to 1
	validGame.highestBidder = 1
	data, _ := json.Marshal(validGame)

	var rawMap map[string]interface{}
	_ = json.Unmarshal(data, &rawMap)

	// Mutate: delete hd (highestBidder)
	delete(rawMap, "hd")
	mutatedData, _ := json.Marshal(rawMap)

	game2 := &Binokel{}
	_ = json.Unmarshal(mutatedData, game2)
	// Mutated JSON leaves highestBidder as default 0 rather than original 1
	if game2.GetHighestBidder() != 0 {
		t.Errorf("expected highest bidder to default to 0 when omitted, got %d", game2.GetHighestBidder())
	}
	if game2.GetHighestBidder() == validGame.GetHighestBidder() {
		t.Error("mutated game should not match original highest bidder")
	}
}

// ─── 9. Bidding Edge Cases ────────────────────────────────

func TestBinokel_BiddingEdgeCases(t *testing.T) {
	game := newTestBinokel()
	game.Reset()
	game.dealerIdx = 2
	game.bidPlayerIdx = 0

	// Wrong phase: play when in bid phase
	if err := game.PlayerPlay(0); err == nil {
		t.Error("expected error playing card in bid phase")
	}

	// PlayerBid: below min bid (140)
	if err := game.PlayerBid(140); err == nil {
		t.Error("expected error bidding below min bid (140)")
	}

	// PlayerBid: not a multiple of 10 (155)
	if err := game.PlayerBid(155); err == nil {
		t.Error("expected error bidding not in step of 10")
	}

	// PlayerBid: valid bid
	if err := game.PlayerBid(150); err != nil {
		t.Fatalf("PlayerBid(150) failed: %v", err)
	}

	// Not human's turn now (player 1's turn)
	if err := game.PlayerBid(160); err == nil {
		t.Error("expected error bidding when not human's turn")
	}
	if err := game.PlayerPass(); err == nil {
		t.Error("expected error passing when not human's turn")
	}

	// CPU 1 bids or passes
	game.CpuBid()

	// CPU 2 bids or passes
	game.CpuBid()

	// If game returned to human, bid lower or equal to current highest
	if game.GetPhase() == BinokelPhaseBid && game.IsHumanBidTurn() {
		if err := game.PlayerBid(game.GetHighestBid()); err == nil {
			t.Error("expected error bidding <= current highest bid")
		}
	}
}

func TestBinokel_AllPassForcedBid(t *testing.T) {
	for dealer := 0; dealer < BinokelPlayerCnt; dealer++ {
		t.Run(fmt.Sprintf("dealer_%d", dealer), func(t *testing.T) {
			game := newTestBinokel()
			game.Reset()
			game.dealerIdx = dealer
			game.bidPlayerIdx = (dealer + 1) % BinokelPlayerCnt

			// All players pass in turn order:
			for i := 0; i < BinokelPlayerCnt; i++ {
				current := game.GetBidPlayerIdx()
				if game.GetPlayer(current).GetIsHuman() {
					if err := game.PlayerPass(); err != nil {
						t.Fatalf("PlayerPass failed for player %d: %v", current, err)
					}
				} else {
					if err := game.doPass(current); err != nil {
						t.Fatalf("doPass failed for player %d: %v", current, err)
					}
				}
			}

			// Phase should now be Dabb with forced winner
			if game.GetPhase() != BinokelPhaseDabb {
				t.Fatalf("expected phase Dabb after all pass, got %d", game.GetPhase())
			}
			if game.GetHighestBid() != BinokelMinBid {
				t.Errorf("expected highest bid %d, got %d", BinokelMinBid, game.GetHighestBid())
			}
			// The forced bidder must be the dealer (last speaker)
			winner := game.GetHighestBidder()
			if winner != dealer {
				t.Errorf("expected forced bidder to be dealer %d, got %d", dealer, winner)
			}
			// The forced bidder should have hand size 18 (15 original + 3 dabb)
			if game.GetPlayer(winner).GetCardsSize() != 18 {
				t.Errorf("forced bidder hand size expected 18, got %d", game.GetPlayer(winner).GetCardsSize())
			}

			// Verify action log contains "forced_bid" for the forced bidder
			foundForcedBidLog := false
			for _, entry := range game.GetActionLog() {
				if entry.ActionType == "forced_bid" && entry.PlayerIdx == dealer {
					foundForcedBidLog = true
					break
				}
			}
			if !foundForcedBidLog {
				t.Errorf("expected forced_bid action log entry for dealer %d", dealer)
			}
		})
	}
}

// ─── 10. Trump Calling & Meld Phase ───────────────────────

func TestBinokel_TrumpCallAndMelds(t *testing.T) {
	game := newTestBinokel()
	game.Reset()
	game.dealerIdx = 2
	game.bidPlayerIdx = 0

	// Force human to win bid
	_ = game.PlayerBid(150)
	_ = game.doPass(1)
	_ = game.doPass(2)

	if game.GetPhase() != BinokelPhaseDabb {
		t.Fatalf("expected phase Dabb, got %d", game.GetPhase())
	}

	// Discard 3 cards
	_ = game.PlayerDiscardToDabb([]int{0, 1, 2})

	if game.GetPhase() != BinokelPhaseTrump {
		t.Fatalf("expected phase Trump, got %d", game.GetPhase())
	}

	// Invalid trump suit
	if err := game.PlayerCallTrump(99); err == nil {
		t.Error("expected error for invalid trump suit")
	}

	// Valid trump suit
	if err := game.PlayerCallTrump(CardDesignSpade); err != nil {
		t.Fatalf("PlayerCallTrump failed: %v", err)
	}

	if game.GetTrumpSuit() != CardDesignSpade {
		t.Errorf("expected trump suit Spade, got %d", game.GetTrumpSuit())
	}

	// Should be in Meld phase
	if game.GetPhase() != BinokelPhaseMeld {
		t.Fatalf("expected phase Meld, got %d", game.GetPhase())
	}

	// ConfirmMelds should advance to Play phase
	game.ConfirmMelds()
	if game.GetPhase() != BinokelPhasePlay {
		t.Fatalf("expected phase Play, got %d", game.GetPhase())
	}
	if game.GetTrickNumber() != 1 {
		t.Errorf("expected trick number 1, got %d", game.GetTrickNumber())
	}
	if game.GetLeadPlayerIdx() != 0 {
		t.Errorf("expected lead player 0 (bidder), got %d", game.GetLeadPlayerIdx())
	}
}

func TestBinokel_CpuTrumpCall(t *testing.T) {
	game := newTestBinokel()
	game.Reset()

	// CPU 1 wins bid
	_ = game.doBid(1, 150)
	_ = game.doPass(2)
	_ = game.doPass(0)

	if game.GetPhase() != BinokelPhaseDabb {
		t.Fatalf("expected phase Dabb, got %d", game.GetPhase())
	}

	// CPU discards
	game.CpuDiscardToDabb()

	if game.GetPhase() != BinokelPhaseTrump {
		t.Fatalf("expected phase Trump, got %d", game.GetPhase())
	}

	game.CpuCallTrump()

	if game.GetTrumpSuit() < CardDesignSpade || game.GetTrumpSuit() > CardDesignDiamond {
		t.Errorf("invalid trump suit called by CPU: %d", game.GetTrumpSuit())
	}
	if game.GetPhase() != BinokelPhaseMeld {
		t.Errorf("expected phase Meld, got %d", game.GetPhase())
	}
}

// ─── 11. Trick-Taking & Rules ─────────────────────────────

func TestBinokel_TrickRules(t *testing.T) {
	game := newTestBinokel()
	game.Reset()

	// Set up game state directly in play phase
	game.phase = BinokelPhasePlay
	game.trumpSuit = CardDesignSpade
	game.leadPlayerIdx = 0
	game.currentPlayerIdx = 0
	game.trickNumber = 1

	// Setup custom hands to test trick-taking rules:
	// Player 0: Spade A (trump), Heart 10
	// Player 1: Heart A, Heart 7
	// Player 2: Heart K, Spade 7 (trump)
	for i := 0; i < BinokelPlayerCnt; i++ {
		game.players[i].ResetRound()
	}
	game.players[0].AddCard(NewCard(CardDesignHeart, 10, false))
	game.players[0].AddCard(NewCard(CardDesignSpade, 1, false))

	game.players[1].AddCard(NewCard(CardDesignHeart, 1, false))
	game.players[1].AddCard(NewCard(CardDesignHeart, 7, false))

	game.players[2].AddCard(NewCard(CardDesignHeart, 13, false))
	game.players[2].AddCard(NewCard(CardDesignSpade, 7, false))

	// Player 0 leads Heart 10 (index 0)
	if err := game.PlayerPlay(0); err != nil {
		t.Fatalf("PlayerPlay failed: %v", err)
	}

	// Player 1's turn: must follow Heart and must beat Heart 10 if able.
	// Player 1 has Heart A (rank beats 10) and Heart 7 (rank loses to 10).
	// Valid plays must be restricted to Heart A!
	validIndicesP1 := game.GetValidPlayIndices(1)
	if len(validIndicesP1) != 1 || validIndicesP1[0] != 0 {
		t.Errorf("player 1 must play Heart A to win trick, valid: %v", validIndicesP1)
	}
	game.CpuPlay() // plays Heart A

	// Player 2's turn: has Heart K and Spade 7.
	// Lead is Heart. Current winning card is Heart A.
	// Player 2 has Heart K, so must follow Heart (cannot trump even though Heart K loses to Heart A).
	validIndicesP2 := game.GetValidPlayIndices(2)
	if len(validIndicesP2) != 1 || validIndicesP2[0] != 0 {
		t.Errorf("player 2 must follow Heart K, valid: %v", validIndicesP2)
	}
	game.CpuPlay() // plays Heart K

	// Trick should be 3 cards
	if len(game.GetCurrentTrick()) != 3 {
		t.Fatalf("expected 3 cards in trick, got %d", len(game.GetCurrentTrick()))
	}

	// Resolve trick
	game.ResolveTrick()

	// Winner should be Player 1 (Heart A beats Heart 10 and Heart K)
	if game.GetLeadPlayerIdx() != 1 {
		t.Errorf("expected trick winner to be player 1, got %d", game.GetLeadPlayerIdx())
	}
	// Points: Heart 10 (10) + Heart A (11) + Heart K (4) = 25
	if game.players[1].GetTrickPoints() != 25 {
		t.Errorf("expected 25 trick points for player 1, got %d", game.players[1].GetTrickPoints())
	}

	// Next trick
	game.NextTrick()
	if game.GetTrickNumber() != 2 {
		t.Errorf("expected trick number 2, got %d", game.GetTrickNumber())
	}
	if game.GetCurrentPlayerIdx() != 1 {
		t.Errorf("expected currentPlayerIdx 1, got %d", game.GetCurrentPlayerIdx())
	}
}

func TestBinokel_MustTrumpRule(t *testing.T) {
	game := newTestBinokel()
	game.Reset()

	game.phase = BinokelPhasePlay
	game.trumpSuit = CardDesignSpade
	game.leadPlayerIdx = 0
	game.currentPlayerIdx = 0

	for i := 0; i < BinokelPlayerCnt; i++ {
		game.players[i].ResetRound()
	}
	// Player 0 leads Heart 10
	// Player 1 has NO Hearts, but has Spade 7 (trump) and Diamond A.
	// Must trump!
	game.players[0].AddCard(NewCard(CardDesignHeart, 10, false))
	game.players[1].AddCard(NewCard(CardDesignDiamond, 1, false))
	game.players[1].AddCard(NewCard(CardDesignSpade, 7, false))
	game.players[2].AddCard(NewCard(CardDesignHeart, 1, false))

	_ = game.PlayerPlay(0)

	validP1 := game.GetValidPlayIndices(1)
	// Must play trump (Spade 7, index 1)
	if len(validP1) != 1 || validP1[0] != 1 {
		t.Errorf("player 1 must trump when void in lead suit, valid: %v", validP1)
	}
}

// ─── 12. Complete Play & Game End ─────────────────────────

func TestBinokel_FullPlayToGameEnd(t *testing.T) {
	cfg := DefaultBinokelConfig()
	cfg.PointLimit = 200 // Low limit to test game completion
	game := newTestBinokel(cfg)
	game.Reset()

	// Simulate completing rounds until game end
	for round := 1; round <= 20 && !game.GetGameEndFlag(); round++ {
		if game.GetPhase() == BinokelPhaseBid {
			// Fast finish auction
			_ = game.doBid(0, 150)
			_ = game.doPass(1)
			_ = game.doPass(2)
		}
		if game.GetPhase() == BinokelPhaseDabb {
			_ = game.PlayerDiscardToDabb([]int{0, 1, 2})
		}
		if game.GetPhase() == BinokelPhaseTrump {
			_ = game.PlayerCallTrump(CardDesignSpade)
		}
		if game.GetPhase() == BinokelPhaseMeld {
			game.ConfirmMelds()
		}
		if game.GetPhase() == BinokelPhasePlay {
			for game.GetPhase() == BinokelPhasePlay {
				if len(game.GetCurrentTrick()) == BinokelPlayerCnt {
					game.ResolveTrick()
					game.NextTrick()
					continue
				}
				if game.IsHumanTurn() {
					valid := game.GetValidPlayIndices(game.GetCurrentPlayerIdx())
					_ = game.PlayerPlay(valid[0])
				} else {
					game.CpuPlay()
				}
			}
		}
		if game.GetPhase() == BinokelPhaseRoundEnd && !game.GetGameEndFlag() {
			game.NextRound()
		}
	}

	if !game.GetGameEndFlag() {
		t.Log("game did not end in 20 rounds, but play loop ran without deadlock")
	} else {
		if game.GetWinnerPlayer() < 0 || game.GetWinnerPlayer() >= BinokelPlayerCnt {
			t.Errorf("expected valid winner player, got %d", game.GetWinnerPlayer())
		}
	}
}

// ─── 13. Hint & Other Methods ─────────────────────────────

func TestBinokel_Hints(t *testing.T) {
	game := newTestBinokel()
	game.Reset()
	game.dealerIdx = 2
	game.bidPlayerIdx = 0

	// Hint in Bid Phase
	hint := game.GetHint()
	if hint == nil {
		t.Fatal("expected non-nil hint during human bid turn")
	}
	if hint.BidAmount == nil && hint.Pass == nil {
		t.Error("expected hint to suggest either bid or pass")
	}

	// Hint when not human turn
	game.bidPlayerIdx = 1
	if h := game.GetHint(); h != nil {
		t.Errorf("expected nil hint when not human turn, got: %+v", h)
	}

	// Hint in Trump Phase
	game.phase = BinokelPhaseTrump
	game.highestBidder = 0
	game.currentPlayerIdx = 0
	hintTrump := game.GetHint()
	if hintTrump == nil || hintTrump.Suit == nil {
		t.Fatal("expected trump suit hint")
	}

	// Hint in Play Phase
	game.phase = BinokelPhasePlay
	game.currentPlayerIdx = 0
	hintPlay := game.GetHint()
	if hintPlay == nil || hintPlay.CardIndex == nil {
		t.Fatal("expected play card hint")
	}
}

func TestBinokel_GettersAndCompatibility(t *testing.T) {
	game := newTestBinokel()
	game.Reset()

	if game.GetPlayerCnt() != 3 {
		t.Errorf("expected 3 players, got %d", game.GetPlayerCnt())
	}
	if len(game.GetPlayers()) != 3 {
		t.Errorf("expected 3 players slice, got %d", len(game.GetPlayers()))
	}
	if game.GetRoundNumber() != 1 {
		t.Errorf("expected round number 1, got %d", game.GetRoundNumber())
	}
	if game.GetDealerIdx() != 0 {
		t.Errorf("expected dealer 0, got %d", game.GetDealerIdx())
	}

	// Compatibility stubs
	if game.GetTeamScore(0) != game.GetScore(0) {
		t.Error("GetTeamScore compatibility stub mismatch")
	}
	if game.GetWinnerTeam() != game.GetWinnerPlayer() {
		t.Error("GetWinnerTeam compatibility stub mismatch")
	}

	// SortHand
	game.SortHand(0)
	p0 := game.GetPlayer(0)
	for i := 1; i < p0.GetCardsSize(); i++ {
		prev := p0.GetCard(i - 1)
		curr := p0.GetCard(i)
		if prev.GetDesign() > curr.GetDesign() {
			t.Error("hand not sorted by suit")
		}
	}

	// ActionLog after an action
	_ = game.doBid(1, 150)
	logs := game.GetActionLog()
	if len(logs) == 0 {
		t.Error("expected non-empty action log after bid")
	}

	// Additional getters & setters
	_ = game.GetBidPlayerIdx()
	_ = game.GetScores()
	_ = game.GetScore(-1) // out of bounds
	_ = game.GetScore(99) // out of bounds
	cfg := game.GetConfig()
	game.SetConfig(cfg)
	_ = game.GetPlayerMelds()

	table := BinokelMeldTable()
	if len(table) == 0 {
		t.Error("expected non-empty meld table")
	}
}

func TestBinokel_DifficultiesAndCpuStrategies(t *testing.T) {
	// Easy CPU
	cfgEasy := DefaultBinokelConfig()
	cfgEasy.CpuDifficulty = BinokelCpuDifficultyEasy
	gameEasy := newTestBinokel(cfgEasy)
	gameEasy.Reset()
	// CPU 1 bids (Easy)
	gameEasy.CpuBid()

	// Hard CPU
	cfgHard := DefaultBinokelConfig()
	cfgHard.CpuDifficulty = BinokelCpuDifficultyHard
	gameHard := newTestBinokel(cfgHard)
	gameHard.Reset()
	// CPU 1 bids (Hard)
	gameHard.CpuBid()

	// Play with Easy
	gameEasy.phase = BinokelPhasePlay
	gameEasy.trumpSuit = CardDesignSpade
	gameEasy.currentPlayerIdx = 1
	gameEasy.CpuPlay()

	// Play with Hard
	gameHard.phase = BinokelPhasePlay
	gameHard.trumpSuit = CardDesignSpade
	gameHard.currentPlayerIdx = 1
	gameHard.CpuPlay()
}

func TestBinokel_PhaseErrorBranches(t *testing.T) {
	game := newTestBinokel()
	game.Reset()

	// Wrong phase errors
	game.phase = BinokelPhaseTrump
	if err := game.PlayerBid(150); err == nil {
		t.Error("expected error PlayerBid in Trump phase")
	}
	if err := game.PlayerPass(); err == nil {
		t.Error("expected error PlayerPass in Trump phase")
	}

	game.phase = BinokelPhaseBid
	if err := game.PlayerDiscardToDabb([]int{0, 1, 2}); err == nil {
		t.Error("expected error PlayerDiscardToDabb in Bid phase")
	}
	if err := game.PlayerCallTrump(CardDesignSpade); err == nil {
		t.Error("expected error PlayerCallTrump in Bid phase")
	}

	// PlayerDiscardToDabb when bidder is CPU (player 1)
	game.phase = BinokelPhaseDabb
	game.highestBidder = 1
	game.currentPlayerIdx = 1
	if err := game.PlayerDiscardToDabb([]int{0, 1, 2}); err == nil {
		t.Error("expected error when human discards but CPU is bidder")
	}

	// PlayerCallTrump when human is not highest bidder
	game.phase = BinokelPhaseTrump
	game.highestBidder = 1
	game.currentPlayerIdx = 1
	if err := game.PlayerCallTrump(CardDesignSpade); err == nil {
		t.Error("expected error when human calls trump but CPU is bidder")
	}

	// CpuCallTrump when human is highest bidder
	game.highestBidder = 0
	game.currentPlayerIdx = 0
	game.CpuCallTrump() // should no-op
	if game.GetPhase() != BinokelPhaseTrump {
		t.Error("CpuCallTrump should no-op when human is bidder")
	}

	// ConfirmMelds when not in Meld phase
	game.phase = BinokelPhaseBid
	game.ConfirmMelds()
	if game.GetPhase() != BinokelPhaseBid {
		t.Error("ConfirmMelds should no-op when not in Meld phase")
	}

	// PlayerPlay when not play phase
	game.phase = BinokelPhaseBid
	if err := game.PlayerPlay(0); err == nil {
		t.Error("expected error PlayerPlay when not in Play phase")
	}

	// PlayerPlay when not human turn
	game.phase = BinokelPhasePlay
	game.currentPlayerIdx = 1
	if err := game.PlayerPlay(0); err == nil {
		t.Error("expected error PlayerPlay when not human turn")
	}

	// PlayerPlay with out of bounds card index
	game.currentPlayerIdx = 0
	if err := game.PlayerPlay(-1); err == nil {
		t.Error("expected error PlayerPlay with negative index")
	}
	if err := game.PlayerPlay(99); err == nil {
		t.Error("expected error PlayerPlay with out of bounds index")
	}

	// CpuPlay when not play phase or human turn
	game.phase = BinokelPhaseBid
	game.CpuPlay()
	game.phase = BinokelPhasePlay
	game.currentPlayerIdx = 0
	game.CpuPlay() // should no-op

	// ResolveTrick when trick is not full
	game.currentTrick = nil
	game.ResolveTrick() // should no-op

	// NextTrick when trick is empty
	game.NextTrick() // should no-op

	// NextRound when phase is not RoundEnd
	game.phase = BinokelPhasePlay
	game.NextRound() // should no-op
	if game.GetRoundNumber() != 1 {
		t.Error("NextRound should not advance when not RoundEnd")
	}

	// NextRound when game has ended
	game.phase = BinokelPhaseRoundEnd
	game.gameEndFlag = true
	game.NextRound() // should no-op

	// IsHumanTurn branches
	game.phase = BinokelPhaseTrump
	game.currentPlayerIdx = 0
	_ = game.IsHumanTurn()
	game.phase = BinokelPhasePlay
	game.currentPlayerIdx = 0
	_ = game.IsHumanTurn()
	game.phase = BinokelPhaseBid
	_ = game.IsHumanTurn()

	// IsHumanBidTurn when not in Bid phase
	game.phase = BinokelPhasePlay
	if game.IsHumanBidTurn() {
		t.Error("IsHumanBidTurn should be false when not in Bid phase")
	}

	// IsHumanDabbTurn when not in Dabb phase
	game.phase = BinokelPhasePlay
	if game.IsHumanDabbTurn() {
		t.Error("IsHumanDabbTurn should be false when not in Dabb phase")
	}

	// Hint when not in bid/trump/play phase
	game.phase = BinokelPhaseRoundEnd
	if game.Hint() != nil {
		t.Error("Hint should be nil when in RoundEnd phase")
	}
}

func TestBinokel_GameEndAndScoringScenarios(t *testing.T) {
	// Game ends when player reaches PointLimit
	game := newTestBinokel()
	game.Reset()
	game.scores[0] = 950
	game.highestBidder = 0
	game.highestBid = 150
	game.players[0].SetMeldScore(100)
	game.players[0].SetTrickPoints(100)

	game.scoreRound()

	if !game.GetGameEndFlag() {
		t.Error("expected game to end when player score exceeds PointLimit (1000)")
	}
	if game.GetWinnerPlayer() != 0 {
		t.Errorf("expected winner player 0, got %d", game.GetWinnerPlayer())
	}

	// Tie-breaking check
	gameTie := newTestBinokel()
	gameTie.Reset()
	gameTie.scores[0] = 1050
	gameTie.scores[1] = 1050
	gameTie.scores[2] = 800
	gameTie.highestBidder = 1
	gameTie.highestBid = 150
	gameTie.players[1].SetMeldScore(150)
	gameTie.players[1].SetTrickPoints(0)

	gameTie.scoreRound()
	if !gameTie.GetGameEndFlag() {
		t.Error("expected game to end on tie")
	}
	if gameTie.GetWinnerPlayer() < 0 {
		t.Error("expected winner player on tie")
	}
}
