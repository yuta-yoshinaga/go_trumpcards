//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

// ─── Deck Factory ───────────────────────────────────────

func TestNewTrumpCardsPinochle(t *testing.T) {
	deck := NewTrumpCardsPinochle()
	if deck.GetTotalCount() != 48 {
		t.Errorf("expected 48 cards, got %d", deck.GetTotalCount())
	}

	// カードの分布を確認: 各 {suit, value} ペアが2枚ずつ
	type sv struct{ suit, value int }
	counts := make(map[sv]int)
	for i := 0; i < 48; i++ {
		card := deck.DrawCard()
		if card == nil {
			t.Fatalf("expected card at index %d, got nil", i)
		}
		counts[sv{card.GetDesign(), card.GetValue()}]++
	}

	expectedValues := []int{1, 9, 10, 11, 12, 13}
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for _, s := range suits {
		for _, v := range expectedValues {
			key := sv{s, v}
			if counts[key] != 2 {
				t.Errorf("expected 2 cards for suit=%d value=%d, got %d", s, v, counts[key])
			}
		}
	}

	// 49枚目はnil
	if deck.DrawCard() != nil {
		t.Error("expected nil for 49th draw")
	}
}

// ─── Config ─────────────────────────────────────────────

func TestPinochleConfig_Default(t *testing.T) {
	config := DefaultPinochleConfig()
	if config.CpuDifficulty != PinochleCpuDifficultyNormal {
		t.Errorf("expected Normal difficulty, got %d", config.CpuDifficulty)
	}
	if config.PointLimit != 1500 {
		t.Errorf("expected 1500 point limit, got %d", config.PointLimit)
	}
}

func TestPinochleConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  PinochleConfig
		wantErr bool
	}{
		{"valid default", DefaultPinochleConfig(), false},
		{"valid easy", PinochleConfig{CpuDifficulty: PinochleCpuDifficultyEasy, PointLimit: 100}, false},
		{"valid hard", PinochleConfig{CpuDifficulty: PinochleCpuDifficultyHard, PointLimit: 5000}, false},
		{"invalid difficulty", PinochleConfig{CpuDifficulty: 99, PointLimit: 1500}, true},
		{"invalid point limit", PinochleConfig{CpuDifficulty: PinochleCpuDifficultyNormal, PointLimit: 0}, true},
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

// ─── Player ─────────────────────────────────────────────

func TestPinochlePlayer_ResetRound(t *testing.T) {
	p := NewPinochlePlayer(true, 0)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.SetBid(25)
	p.SetHasPassed(true)
	p.SetMeldScore(100)
	p.SetTrickPoints(50)
	p.AddTrick([]*Card{NewCard(CardDesignSpade, 1, false)})

	p.ResetRound()

	if p.GetCardsSize() != 0 {
		t.Error("expected empty hand after reset")
	}
	if p.GetBid() != 0 {
		t.Error("expected bid 0 after reset")
	}
	if p.GetHasPassed() {
		t.Error("expected hasPassed false after reset")
	}
	if p.GetMeldScore() != 0 {
		t.Error("expected meld score 0 after reset")
	}
	if p.GetTrickPoints() != 0 {
		t.Error("expected trick points 0 after reset")
	}
	if p.GetTrickCount() != 0 {
		t.Error("expected trick count 0 after reset")
	}
}

func TestPinochlePlayer_JSON(t *testing.T) {
	p := NewPinochlePlayer(true, 1)
	p.AddCard(NewCard(CardDesignHeart, 1, false))
	p.SetBid(25)
	p.SetMeldScore(40)

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var p2 PinochlePlayer
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if p2.GetTeam() != 1 {
		t.Errorf("expected team 1, got %d", p2.GetTeam())
	}
	if p2.GetBid() != 25 {
		t.Errorf("expected bid 25, got %d", p2.GetBid())
	}
	if p2.GetMeldScore() != 40 {
		t.Errorf("expected meld score 40, got %d", p2.GetMeldScore())
	}
	if p2.GetCardsSize() != 1 {
		t.Errorf("expected 1 card, got %d", p2.GetCardsSize())
	}
}

// ─── Game Init & Reset ─────────────────────────────────

func newTestPinochle() *Pinochle {
	players := []*PinochlePlayer{
		NewPinochlePlayer(true, 0),
		NewPinochlePlayer(false, 1),
		NewPinochlePlayer(false, 0),
		NewPinochlePlayer(false, 1),
	}
	return NewPinochle(NewTrumpCardsPinochle(), players, DefaultPinochleConfig())
}

func TestPinochle_Reset(t *testing.T) {
	g := newTestPinochle()
	g.Reset()

	if g.GetPhase() != PinochlePhaseBid {
		t.Errorf("expected Bid phase, got %d", g.GetPhase())
	}
	if g.GetRoundNumber() != 1 {
		t.Errorf("expected round 1, got %d", g.GetRoundNumber())
	}

	// 全プレイヤーに12枚配られている
	for i, p := range g.GetPlayers() {
		if p.GetCardsSize() != PinochleHandSize {
			t.Errorf("player %d expected %d cards, got %d", i, PinochleHandSize, p.GetCardsSize())
		}
	}

	// ビッドはディーラーの左隣から
	expectedBidder := (g.GetDealerIdx() + 1) % PinochlePlayerCnt
	if g.GetBidPlayerIdx() != expectedBidder {
		t.Errorf("expected bid player %d, got %d", expectedBidder, g.GetBidPlayerIdx())
	}
}

// ─── Meld Evaluation ───────────────────────────────────

func TestEvaluateMelds_Dix(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 9, false), // 9 of trump
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldDix {
			found = true
			if m.Points != 10 {
				t.Errorf("expected 10 points for Dix, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Dix meld")
	}
}

func TestEvaluateMelds_DoubleDix(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 9, false),
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	dixCount := 0
	for _, m := range melds {
		if m.Type == PinochleMeldDix {
			dixCount++
		}
	}
	if dixCount != 2 {
		t.Errorf("expected 2 Dix melds, got %d", dixCount)
	}
}

func TestEvaluateMelds_CommonMarriage(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignHeart, 13, false), // K♥
		NewCard(CardDesignHeart, 12, false), // Q♥
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldCommonMarriage {
			found = true
			if m.Points != 20 {
				t.Errorf("expected 20 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Common Marriage meld")
	}
}

func TestEvaluateMelds_RoyalMarriage(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 13, false), // K♠
		NewCard(CardDesignSpade, 12, false), // Q♠
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldRoyalMarriage {
			found = true
			if m.Points != 40 {
				t.Errorf("expected 40 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Royal Marriage meld")
	}
}

func TestEvaluateMelds_Pinochle(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignDiamond, 11, false), // J♦
		NewCard(CardDesignSpade, 12, false),   // Q♠
	}
	melds := evaluateMelds(hand, CardDesignHeart)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldPinochle {
			found = true
			if m.Points != 40 {
				t.Errorf("expected 40 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Pinochle meld")
	}
}

func TestEvaluateMelds_DoublePinochle(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 12, false),
	}
	melds := evaluateMelds(hand, CardDesignHeart)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldDoublePinochle {
			found = true
			if m.Points != 300 {
				t.Errorf("expected 300 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Double Pinochle meld")
	}
}

func TestEvaluateMelds_AcesAround(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignDiamond, 1, false),
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldAcesAround {
			found = true
			if m.Points != 100 {
				t.Errorf("expected 100 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Aces Around meld")
	}
}

func TestEvaluateMelds_DoubleAcesAround(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignDiamond, 1, false),
		NewCard(CardDesignDiamond, 1, false),
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldDoubleAcesAround {
			found = true
			if m.Points != 1000 {
				t.Errorf("expected 1000 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Double Aces Around meld")
	}
}

func TestEvaluateMelds_KingsAround(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignDiamond, 13, false),
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldKingsAround {
			found = true
			if m.Points != 80 {
				t.Errorf("expected 80 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Kings Around meld")
	}
}

func TestEvaluateMelds_QueensAround(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignDiamond, 12, false),
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldQueensAround {
			found = true
			if m.Points != 60 {
				t.Errorf("expected 60 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Queens Around meld")
	}
}

func TestEvaluateMelds_JacksAround(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignClover, 11, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignDiamond, 11, false),
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldJacksAround {
			found = true
			if m.Points != 40 {
				t.Errorf("expected 40 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Jacks Around meld")
	}
}

func TestEvaluateMelds_Run(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 1, false),  // A♠
		NewCard(CardDesignSpade, 10, false), // 10♠
		NewCard(CardDesignSpade, 13, false), // K♠
		NewCard(CardDesignSpade, 12, false), // Q♠
		NewCard(CardDesignSpade, 11, false), // J♠
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldRun {
			found = true
			if m.Points != 150 {
				t.Errorf("expected 150 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Run meld")
	}
	// RunにはRoyal Marriageが含まれるため別途加算されない
	for _, m := range melds {
		if m.Type == PinochleMeldRoyalMarriage {
			t.Error("Royal Marriage should not be added when Run is present")
		}
	}
}

func TestEvaluateMelds_DoubleRun(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 11, false),
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	found := false
	for _, m := range melds {
		if m.Type == PinochleMeldDoubleRun {
			found = true
			if m.Points != 1500 {
				t.Errorf("expected 1500 points, got %d", m.Points)
			}
		}
	}
	if !found {
		t.Error("expected Double Run meld")
	}
}

func TestEvaluateMelds_NoMelds(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 9, false),    // 9♠
		NewCard(CardDesignClover, 9, false),   // 9♣
		NewCard(CardDesignHeart, 9, false),    // 9♥
		NewCard(CardDesignDiamond, 10, false), // 10♦
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	// 9♠はDixなのでメルドあり
	total := meldTotalPoints(melds)
	if total != 10 {
		t.Errorf("expected 10 points (Dix only), got %d", total)
	}
}

func TestEvaluateMelds_MultipleMeldTypes(t *testing.T) {
	// Run + Aces Around + Pinochle を同時に
	hand := []*Card{
		// Run in spades
		NewCard(CardDesignSpade, 1, false),  // A♠
		NewCard(CardDesignSpade, 10, false), // 10♠
		NewCard(CardDesignSpade, 13, false), // K♠
		NewCard(CardDesignSpade, 12, false), // Q♠ (also part of Pinochle)
		NewCard(CardDesignSpade, 11, false), // J♠
		// Aces Around (remaining aces)
		NewCard(CardDesignClover, 1, false),  // A♣
		NewCard(CardDesignHeart, 1, false),   // A♥
		NewCard(CardDesignDiamond, 1, false), // A♦
		// Pinochle
		NewCard(CardDesignDiamond, 11, false), // J♦
	}
	melds := evaluateMelds(hand, CardDesignSpade)
	types := make(map[PinochleMeldType]bool)
	for _, m := range melds {
		types[m.Type] = true
	}
	if !types[PinochleMeldRun] {
		t.Error("expected Run meld")
	}
	if !types[PinochleMeldAcesAround] {
		t.Error("expected Aces Around meld")
	}
	if !types[PinochleMeldPinochle] {
		t.Error("expected Pinochle meld")
	}
}

// ─── Card Ranking ───────────────────────────────────────

func TestPinochleRankValue(t *testing.T) {
	// A > 10 > K > Q > J > 9
	tests := []struct {
		value    int
		expected int
	}{
		{1, 6},  // Ace
		{10, 5}, // 10
		{13, 4}, // King
		{12, 3}, // Queen
		{11, 2}, // Jack
		{9, 1},  // 9
	}
	for _, tt := range tests {
		got := pinochleRankValue(tt.value)
		if got != tt.expected {
			t.Errorf("pinochleRankValue(%d) = %d, want %d", tt.value, got, tt.expected)
		}
	}
}

func TestPinochleCardPointValue(t *testing.T) {
	tests := []struct {
		value    int
		expected int
	}{
		{1, 11}, // Ace
		{10, 10},
		{13, 4}, // King
		{12, 3}, // Queen
		{11, 2}, // Jack
		{9, 0},  // 9
	}
	for _, tt := range tests {
		card := NewCard(CardDesignSpade, tt.value, false)
		got := pinochleCardPointValue(card)
		if got != tt.expected {
			t.Errorf("pinochleCardPointValue(value=%d) = %d, want %d", tt.value, got, tt.expected)
		}
	}
}

// ─── Bidding ────────────────────────────────────────────

func TestPinochle_Bidding(t *testing.T) {
	g := newTestPinochle()
	g.Reset()

	// Player 0はhuman, bidPlayerIdxは1 (dealer 0の左隣)
	// CPU 1, 2, 3がパスして、player 0がビッドする
	// まずCPUにパスさせるため、difficulty をeasyにしてテスト

	// 代わりに手動でビッドを制御
	// bidPlayerIdx = 1 (CPU)
	if g.GetBidPlayerIdx() != 1 {
		t.Fatalf("expected bid player 1, got %d", g.GetBidPlayerIdx())
	}

	// CPU1がパス
	_ = g.doPass(1)
	// CPU2がパス
	_ = g.doPass(2)
	// CPU3がパス
	_ = g.doPass(3)

	// 全員パス → ディーラー (player 0) が強制ビッド
	if g.GetPhase() != PinochlePhaseTrump {
		t.Errorf("expected Trump phase after all pass, got %d", g.GetPhase())
	}
	if g.GetHighestBid() != PinochleMinBid {
		t.Errorf("expected forced bid %d, got %d", PinochleMinBid, g.GetHighestBid())
	}
}

func TestPinochle_BidValidation(t *testing.T) {
	g := newTestPinochle()
	g.Reset()

	// 最低ビッド未満
	err := g.doBid(g.GetBidPlayerIdx(), PinochleMinBid-1)
	if err == nil {
		t.Error("expected error for bid below minimum")
	}

	// 最低ビッドでビッド成功
	err = g.doBid(g.GetBidPlayerIdx(), PinochleMinBid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 現在のビッド以下
	err = g.doBid(g.GetBidPlayerIdx(), PinochleMinBid)
	if err == nil {
		t.Error("expected error for bid not exceeding current")
	}
}

func TestPinochle_PlayerBidWrongPhase(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.phase = PinochlePhasePlay // force wrong phase
	err := g.PlayerBid(20)
	if err == nil {
		t.Error("expected wrong phase error")
	}
}

func TestPinochle_PlayerPassWrongPhase(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.phase = PinochlePhasePlay
	err := g.PlayerPass()
	if err == nil {
		t.Error("expected wrong phase error")
	}
}

// ─── Trump Declaration ─────────────────────────────────

func TestPinochle_CallTrump(t *testing.T) {
	g := newTestPinochle()
	g.Reset()

	// 全CPUパス → 強制ビッド → Trump phase
	_ = g.doPass(1)
	_ = g.doPass(2)
	_ = g.doPass(3)

	if g.GetPhase() != PinochlePhaseTrump {
		t.Fatalf("expected Trump phase, got %d", g.GetPhase())
	}

	// Player 0がトランプ宣言
	err := g.PlayerCallTrump(CardDesignHeart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.GetTrumpSuit() != CardDesignHeart {
		t.Errorf("expected trump suit Heart, got %d", g.GetTrumpSuit())
	}
	if g.GetPhase() != PinochlePhaseMeld {
		t.Errorf("expected Meld phase, got %d", g.GetPhase())
	}
}

func TestPinochle_CallTrumpInvalidSuit(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	_ = g.doPass(1)
	_ = g.doPass(2)
	_ = g.doPass(3)

	err := g.doCallTrump(0, 99)
	if err == nil {
		t.Error("expected error for invalid suit")
	}
}

func TestPinochle_CallTrumpWrongPhase(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	err := g.PlayerCallTrump(CardDesignSpade)
	if err == nil {
		t.Error("expected wrong phase error")
	}
}

// ─── Trick-Taking ───────────────────────────────────────

func setupPlayPhase(t *testing.T) *Pinochle {
	t.Helper()
	g := newTestPinochle()
	g.Reset()

	// 手札を手動設定
	for _, pl := range g.players {
		pl.Reset()
	}

	// Player 0 (human, team 0): A♠, 10♠, K♥
	g.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	g.players[0].AddCard(NewCard(CardDesignSpade, 10, false))
	g.players[0].AddCard(NewCard(CardDesignHeart, 13, false))

	// Player 1 (CPU, team 1): Q♠, J♠, 9♥
	g.players[1].AddCard(NewCard(CardDesignSpade, 12, false))
	g.players[1].AddCard(NewCard(CardDesignSpade, 11, false))
	g.players[1].AddCard(NewCard(CardDesignHeart, 9, false))

	// Player 2 (CPU, team 0): K♠, 9♠, A♥
	g.players[2].AddCard(NewCard(CardDesignSpade, 13, false))
	g.players[2].AddCard(NewCard(CardDesignSpade, 9, false))
	g.players[2].AddCard(NewCard(CardDesignHeart, 1, false))

	// Player 3 (CPU, team 1): A♦, 10♦, K♦
	g.players[3].AddCard(NewCard(CardDesignDiamond, 1, false))
	g.players[3].AddCard(NewCard(CardDesignDiamond, 10, false))
	g.players[3].AddCard(NewCard(CardDesignDiamond, 13, false))

	g.trumpSuit = CardDesignHeart
	g.highestBid = 20
	g.highestBidder = 0
	g.phase = PinochlePhasePlay
	g.trickNumber = 1
	g.leadPlayerIdx = 0
	g.currentPlayerIdx = 0
	g.currentTrick = nil

	return g
}

func TestPinochle_ValidPlay_FollowSuit(t *testing.T) {
	g := setupPlayPhase(t)

	// Player 0 leads A♠
	err := g.PlayerPlay(0) // A♠
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Player 1 must follow spade
	validIndices := g.getValidPlayIndices(1)
	for _, vi := range validIndices {
		card := g.players[1].GetCard(vi)
		if card.GetDesign() != CardDesignSpade {
			t.Errorf("expected only spade cards valid, got suit=%d", card.GetDesign())
		}
	}
}

func TestPinochle_ValidPlay_MustTrump(t *testing.T) {
	g := setupPlayPhase(t)

	// Player 0 leads A♠
	_ = g.doPlay(0, 0) // A♠

	// Player 3 has no spades → must play trump (heart) if they had it
	// Player 3 only has diamonds, so they can play anything
	// Let's setup a scenario where must-trump applies

	g2 := setupPlayPhase(t)
	// Give player 3 a heart (trump) instead of K♦
	g2.players[3].Reset()
	g2.players[3].AddCard(NewCard(CardDesignDiamond, 1, false))
	g2.players[3].AddCard(NewCard(CardDesignHeart, 10, false)) // trump
	g2.players[3].AddCard(NewCard(CardDesignDiamond, 13, false))

	_ = g2.doPlay(0, 0) // A♠ lead
	_ = g2.doPlay(1, 0) // Q♠ follow
	_ = g2.doPlay(2, 0) // K♠ follow

	// Player 3 has no spades, has trump (heart) → must play trump
	validIndices := g2.getValidPlayIndices(3)
	for _, vi := range validIndices {
		card := g2.players[3].GetCard(vi)
		if card.GetDesign() != CardDesignHeart {
			t.Errorf("expected only trump cards valid when void in lead suit, got suit=%d", card.GetDesign())
		}
	}
}

func TestPinochle_ValidPlay_MustWin(t *testing.T) {
	g := setupPlayPhase(t)

	// Player 0 leads 10♠ (rank 5)
	_ = g.doPlay(0, 1) // 10♠

	// Player 1 has Q♠ (rank 3) and J♠ (rank 2)
	// Neither can beat 10♠ (rank 5) → both are valid (must follow suit but can't win)
	validIndices := g.getValidPlayIndices(1)
	if len(validIndices) != 2 {
		t.Errorf("expected 2 valid indices (Q♠, J♠), got %d", len(validIndices))
	}

	// Now test must-win: Player 0 leads K♥ (trump, rank 404)
	g2 := setupPlayPhase(t)
	g2.players[0].Reset()
	g2.players[0].AddCard(NewCard(CardDesignHeart, 13, false)) // K♥
	g2.players[0].AddCard(NewCard(CardDesignSpade, 10, false))
	g2.players[0].AddCard(NewCard(CardDesignSpade, 1, false))

	g2.players[2].Reset()
	g2.players[2].AddCard(NewCard(CardDesignHeart, 1, false)) // A♥ (can beat K♥)
	g2.players[2].AddCard(NewCard(CardDesignHeart, 9, false)) // 9♥ (cannot beat K♥)
	g2.players[2].AddCard(NewCard(CardDesignSpade, 9, false))

	_ = g2.doPlay(0, 0) // K♥ lead (trump)
	_ = g2.doPlay(1, 2) // Player 1 has no hearts - play whatever
	// Player 2 has A♥ and 9♥. Must follow heart and must win → only A♥ valid
	validIndices = g2.getValidPlayIndices(2)
	if len(validIndices) != 1 {
		t.Errorf("expected 1 valid index (A♥ to beat K♥), got %d", len(validIndices))
	}
}

func TestPinochle_TrickWinner(t *testing.T) {
	g := setupPlayPhase(t)

	// Play a full trick: A♠ > 10♠ > K♠ > some off-suit
	_ = g.doPlay(0, 0) // A♠
	_ = g.doPlay(1, 0) // Q♠
	_ = g.doPlay(2, 0) // K♠
	_ = g.doPlay(3, 0) // A♦ (off-suit, doesn't count)

	winner := g.trickWinner()
	if winner != 0 {
		t.Errorf("expected player 0 (A♠) to win, got player %d", winner)
	}
}

func TestPinochle_TrickWinner_TrumpBeatsLead(t *testing.T) {
	g := setupPlayPhase(t)

	// Lead spade, but player with trump heart wins
	g.players[1].Reset()
	g.players[1].AddCard(NewCard(CardDesignHeart, 9, false)) // 9♥ (trump)

	_ = g.doPlay(0, 0) // A♠ lead
	_ = g.doPlay(1, 0) // 9♥ (trump) beats A♠
	_ = g.doPlay(2, 0) // K♠
	_ = g.doPlay(3, 0) // A♦

	winner := g.trickWinner()
	if winner != 1 {
		t.Errorf("expected player 1 (trump) to win, got player %d", winner)
	}
}

func TestPinochle_PlayerPlayWrongPhase(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	err := g.PlayerPlay(0)
	if err == nil {
		t.Error("expected wrong phase error")
	}
}

func TestPinochle_PlayerPlayInvalidCard(t *testing.T) {
	g := setupPlayPhase(t)
	err := g.PlayerPlay(99)
	if err == nil {
		t.Error("expected invalid card error")
	}
}

// ─── Scoring ────────────────────────────────────────────

func TestPinochle_ScoreRound_BidSuccess(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.highestBid = 20
	g.highestBidder = 0 // team 0
	g.trumpSuit = CardDesignSpade

	// Player 0のメルド: 10点
	g.players[0].SetMeldScore(10)
	// Player 2のメルド (same team): 0点
	g.players[2].SetMeldScore(0)
	// Player 1のメルド: 20点
	g.players[1].SetMeldScore(20)
	// Player 3のメルド: 0点

	// チーム0がトリックポイント15点、チーム1が15点
	g.players[0].SetTrickPoints(15)
	g.players[0].AddTrick([]*Card{NewCard(CardDesignSpade, 1, false)}) // 1 trick
	g.players[2].SetTrickPoints(0)
	g.players[1].SetTrickPoints(10)
	g.players[1].AddTrick([]*Card{NewCard(CardDesignSpade, 9, false)}) // 1 trick
	g.players[3].SetTrickPoints(5)
	g.players[3].AddTrick([]*Card{NewCard(CardDesignSpade, 9, false)}) // 1 trick

	g.phase = PinochlePhaseRoundEnd
	g.scoreRound()

	// チーム0: trickPoints(15) + meldPoints(10) = 25 >= bid(20) → 25加算
	if g.teamScores[0] != 25 {
		t.Errorf("expected team 0 score 25, got %d", g.teamScores[0])
	}
	// チーム1: trickPoints(15) + meldPoints(20) = 35 → 常に加算
	if g.teamScores[1] != 35 {
		t.Errorf("expected team 1 score 35, got %d", g.teamScores[1])
	}
}

func TestPinochle_ScoreRound_BidFailure(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.highestBid = 50
	g.highestBidder = 0 // team 0
	g.trumpSuit = CardDesignSpade

	g.players[0].SetMeldScore(10)
	g.players[0].SetTrickPoints(15)
	g.players[0].AddTrick([]*Card{NewCard(CardDesignSpade, 1, false)})
	g.players[2].SetMeldScore(0)
	g.players[2].SetTrickPoints(0)

	g.players[1].SetMeldScore(0)
	g.players[1].SetTrickPoints(10)
	g.players[1].AddTrick([]*Card{NewCard(CardDesignSpade, 9, false)})
	g.players[3].SetMeldScore(0)
	g.players[3].SetTrickPoints(5)
	g.players[3].AddTrick([]*Card{NewCard(CardDesignSpade, 9, false)})

	g.phase = PinochlePhaseRoundEnd
	g.scoreRound()

	// チーム0: total 25 < bid 50 → -50
	if g.teamScores[0] != -50 {
		t.Errorf("expected team 0 score -50, got %d", g.teamScores[0])
	}
	// チーム1: 15点
	if g.teamScores[1] != 15 {
		t.Errorf("expected team 1 score 15, got %d", g.teamScores[1])
	}
}

func TestPinochle_ScoreRound_MeldForfeiture(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.highestBid = 20
	g.highestBidder = 0
	g.trumpSuit = CardDesignSpade

	// チーム0: メルド100点だがトリック0
	g.players[0].SetMeldScore(100)
	g.players[0].SetTrickPoints(0)
	// no tricks for team 0
	g.players[2].SetMeldScore(0)
	g.players[2].SetTrickPoints(0)

	// チーム1: トリックあり
	g.players[1].SetMeldScore(0)
	g.players[1].SetTrickPoints(30)
	g.players[1].AddTrick([]*Card{NewCard(CardDesignSpade, 1, false)})
	g.players[3].SetTrickPoints(0)
	g.players[3].AddTrick([]*Card{NewCard(CardDesignSpade, 9, false)})

	g.phase = PinochlePhaseRoundEnd
	g.scoreRound()

	// チーム0: メルド没収 (trickCount=0), total=0 < bid=20 → -20
	if g.teamScores[0] != -20 {
		t.Errorf("expected team 0 score -20 (meld forfeited), got %d", g.teamScores[0])
	}
}

// ─── Game End ───────────────────────────────────────────

func TestPinochle_GameEnd(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.config.PointLimit = 100
	g.teamScores[0] = 90
	g.highestBid = 20
	g.highestBidder = 0
	g.trumpSuit = CardDesignSpade

	g.players[0].SetMeldScore(0)
	g.players[0].SetTrickPoints(15)
	g.players[0].AddTrick([]*Card{NewCard(CardDesignSpade, 1, false)})
	g.players[2].SetMeldScore(0)
	g.players[2].SetTrickPoints(10)
	g.players[2].AddTrick([]*Card{NewCard(CardDesignSpade, 9, false)})

	g.players[1].SetMeldScore(0)
	g.players[1].SetTrickPoints(5)
	g.players[1].AddTrick([]*Card{NewCard(CardDesignSpade, 9, false)})
	g.players[3].SetTrickPoints(0)

	g.phase = PinochlePhaseRoundEnd
	g.scoreRound()

	// チーム0: 90 + 25 = 115 >= 100
	if !g.GetGameEndFlag() {
		t.Error("expected game to end")
	}
	if g.GetWinnerTeam() != 0 {
		t.Errorf("expected team 0 to win, got team %d", g.GetWinnerTeam())
	}
	if g.GetPhase() != PinochlePhaseGameEnd {
		t.Errorf("expected GameEnd phase, got %d", g.GetPhase())
	}
}

// ─── JSON Round-Trip ────────────────────────────────────

func TestPinochle_JSON(t *testing.T) {
	g := newTestPinochle()
	g.Reset()

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var g2 Pinochle
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if g2.GetPhase() != g.GetPhase() {
		t.Errorf("phase mismatch: %d vs %d", g2.GetPhase(), g.GetPhase())
	}
	if g2.GetRoundNumber() != g.GetRoundNumber() {
		t.Errorf("round mismatch: %d vs %d", g2.GetRoundNumber(), g.GetRoundNumber())
	}
	if len(g2.GetPlayers()) != PinochlePlayerCnt {
		t.Errorf("expected %d players, got %d", PinochlePlayerCnt, len(g2.GetPlayers()))
	}
}

// ─── NextRound ──────────────────────────────────────────

func TestPinochle_NextRound(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.phase = PinochlePhaseRoundEnd

	oldDealer := g.GetDealerIdx()
	g.NextRound()

	if g.GetPhase() != PinochlePhaseBid {
		t.Errorf("expected Bid phase after NextRound, got %d", g.GetPhase())
	}
	expectedDealer := (oldDealer + 1) % PinochlePlayerCnt
	if g.GetDealerIdx() != expectedDealer {
		t.Errorf("expected dealer %d, got %d", expectedDealer, g.GetDealerIdx())
	}
	if g.GetRoundNumber() != 2 {
		t.Errorf("expected round 2, got %d", g.GetRoundNumber())
	}
}

func TestPinochle_NextRound_WrongPhase(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	// phase is Bid, NextRound should do nothing
	g.NextRound()
	if g.GetPhase() != PinochlePhaseBid {
		t.Errorf("expected Bid phase unchanged, got %d", g.GetPhase())
	}
}

// ─── Hint ───────────────────────────────────────────────

func TestPinochle_Hint_Bid(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	// **Hint は人間の席にしか答えない (#4585)。**Reset 後の入札手番が人間とは
	// 限らないので、明示する。
	g.bidPlayerIdx = g.findHumanIdxForTest()
	hint := g.Hint()
	if hint == nil {
		t.Fatal("expected hint, got nil")
	}
	// Should either recommend a bid or a pass
	if hint.BidAmount == nil && hint.Pass == nil {
		t.Error("expected either BidAmount or Pass in hint")
	}
}

func TestPinochle_Hint_Trump(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	_ = g.doPass(1)
	_ = g.doPass(2)
	_ = g.doPass(3)

	hint := g.Hint()
	if hint == nil {
		t.Fatal("expected hint, got nil")
	}
	if hint.Suit == nil {
		t.Error("expected Suit in hint")
	}
}

func TestPinochle_Hint_Play(t *testing.T) {
	g := setupPlayPhase(t)
	hint := g.Hint()
	if hint == nil {
		t.Fatal("expected hint, got nil")
	}
	if hint.CardIndex == nil {
		t.Error("expected CardIndex in hint")
	}
}

// ─── ConfirmMelds ───────────────────────────────────────

func TestPinochle_ConfirmMelds(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	_ = g.doPass(1)
	_ = g.doPass(2)
	_ = g.doPass(3)
	_ = g.doCallTrump(0, CardDesignSpade)

	if g.GetPhase() != PinochlePhaseMeld {
		t.Fatalf("expected Meld phase, got %d", g.GetPhase())
	}

	g.ConfirmMelds()
	if g.GetPhase() != PinochlePhasePlay {
		t.Errorf("expected Play phase after ConfirmMelds, got %d", g.GetPhase())
	}
	if g.GetTrickNumber() != 1 {
		t.Errorf("expected trick number 1, got %d", g.GetTrickNumber())
	}
}

// ─── ResolveTrick ───────────────────────────────────────

func TestPinochle_ResolveTrick(t *testing.T) {
	g := setupPlayPhase(t)

	_ = g.doPlay(0, 0) // A♠
	_ = g.doPlay(1, 0) // Q♠
	_ = g.doPlay(2, 0) // K♠
	_ = g.doPlay(3, 0) // A♦

	if g.GetPhase() != PinochlePhaseTrickEnd {
		t.Fatalf("expected TrickEnd phase, got %d", g.GetPhase())
	}

	g.ResolveTrick()

	// Winner is player 0 (A♠)
	// Points: A=11 + Q=3 + K=4 + A=11 = 29 (A♦ is off-suit but still has point value)
	if g.players[0].GetTrickPoints() != 29 {
		t.Errorf("expected 29 trick points for player 0, got %d", g.players[0].GetTrickPoints())
	}
	if g.players[0].GetTrickCount() != 1 {
		t.Errorf("expected 1 trick for player 0, got %d", g.players[0].GetTrickCount())
	}
}

// ─── CPU ────────────────────────────────────────────────

func TestPinochle_CpuBid(t *testing.T) {
	// Ensure CPU can bid/pass without error across multiple iterations
	for i := 0; i < 100; i++ {
		g2 := newTestPinochle()
		g2.config.CpuDifficulty = PinochleCpuDifficultyNormal
		g2.Reset()
		// Run bids until the phase changes or the loop is exhausted.
		//
		// **上限 200 は競りの実測に足りていなかった。** 人間は降りられない席で
		// 1 点ずつ上げさせられ、CPU がそれに乗ると入札が長く伸びる。20000 局
		// 回した実測で最悪 185 回、200 回だと ~0.2% で使い切って Bid のまま
		// 抜けていた (highestBid=223 で失敗したのがそれ)。2000 なら 0/20000。
		for j := 0; j < 2000 && g2.GetPhase() == PinochlePhaseBid; j++ {
			bidder := g2.GetBidPlayerIdx()
			if g2.players[bidder].GetIsHuman() {
				// The last remaining active bidder cannot pass (doPass returns
				// ErrCannotPass) and is forced to take the bid. Ignoring that
				// error left the human active forever, spinning the loop to its
				// cap and leaving the phase stuck at Bid — the source of the
				// flake. Force a valid bid so bidding always resolves to Trump.
				if err := g2.doPass(bidder); err != nil {
					forced := PinochleMinBid
					if g2.GetHighestBid() >= forced {
						forced = g2.GetHighestBid() + 1
					}
					if bidErr := g2.doBid(bidder, forced); bidErr != nil {
						t.Fatalf("iteration %d: human forced bid failed: %v", i, bidErr)
					}
				}
			} else {
				g2.CpuBid()
			}
		}
		if g2.GetPhase() != PinochlePhaseTrump {
			t.Errorf("iteration %d: expected Trump phase, got %d (highestBid=%d)",
				i, g2.GetPhase(), g2.GetHighestBid())
		}
	}
}

func TestPinochle_CpuPlay(t *testing.T) {
	g := setupPlayPhase(t)
	g.currentPlayerIdx = 1 // CPU player

	g.CpuPlay()
	// Should have played a card
	if len(g.currentTrick) != 1 {
		t.Errorf("expected 1 card in trick after CPU play, got %d", len(g.currentTrick))
	}
}

func TestPinochle_CpuCallTrump(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.phase = PinochlePhaseTrump
	g.currentPlayerIdx = 1 // CPU
	g.highestBidder = 1

	g.CpuCallTrump()
	if g.GetPhase() != PinochlePhaseMeld {
		t.Errorf("expected Meld phase after CPU call trump, got %d", g.GetPhase())
	}
	if g.GetTrumpSuit() < CardDesignSpade || g.GetTrumpSuit() > CardDesignDiamond {
		t.Errorf("expected valid trump suit, got %d", g.GetTrumpSuit())
	}
}

// ─── SortHand ───────────────────────────────────────────

func TestPinochle_SortHand(t *testing.T) {
	g := newTestPinochle()
	g.Reset()
	g.SortHand(0)

	player := g.GetPlayers()[0]
	for i := 1; i < player.GetCardsSize(); i++ {
		prev := player.GetCard(i - 1)
		curr := player.GetCard(i)
		if prev.GetDesign() > curr.GetDesign() {
			t.Errorf("cards not sorted by suit at index %d", i)
		}
		if prev.GetDesign() == curr.GetDesign() {
			if pinochleRankValue(prev.GetValue()) < pinochleRankValue(curr.GetValue()) {
				t.Errorf("cards not sorted by rank within suit at index %d", i)
			}
		}
	}
}

// ─── NextTrick ──────────────────────────────────────────

func TestPinochle_NextTrick(t *testing.T) {
	g := setupPlayPhase(t)
	g.trickNumber = 1

	_ = g.doPlay(0, 0) // A♠
	_ = g.doPlay(1, 0) // Q♠
	_ = g.doPlay(2, 0) // K♠
	_ = g.doPlay(3, 0) // A♦

	if g.GetPhase() != PinochlePhaseTrickEnd {
		t.Fatalf("expected TrickEnd, got %d", g.GetPhase())
	}

	// ResolveTrick must be called before NextTrick (interactor pattern)
	g.ResolveTrick()
	g.NextTrick()

	if g.GetPhase() != PinochlePhasePlay {
		t.Errorf("expected Play phase after NextTrick, got %d", g.GetPhase())
	}
	if g.GetTrickNumber() != 2 {
		t.Errorf("expected trick number 2, got %d", g.GetTrickNumber())
	}
}

// **CPU の席のヒントは返さない。**返すと、Output() が毎レスポンスで呼ぶ以上、
// CPU の手番に CPU 自身の手が「推奨手」として人間に見える (#4585 のレビュー指摘)。
func TestPinochleHintRefusesTheCpuTurn(t *testing.T) {
	p := newTestPinochle()
	p.Reset()

	// 人間の手番では出る。
	p.phase = PinochlePhasePlay
	p.currentPlayerIdx = p.findHumanIdxForTest()
	if p.Hint() == nil {
		t.Error("the human's own turn must still get a hint")
	}

	// CPU の手番では出ない。
	p.currentPlayerIdx = (p.findHumanIdxForTest() + 1) % PinochlePlayerCnt
	if got := p.Hint(); got != nil {
		t.Errorf("Hint must not describe a CPU seat's move, got %+v", got)
	}
}

// findHumanIdxForTest は人間の席を返す。
func (p *Pinochle) findHumanIdxForTest() int {
	for i, pl := range p.players {
		if pl.GetIsHuman() {
			return i
		}
	}
	return -1
}

// #5519: 早見表はメルド定義そのものから引くこと。別に書き写した表は、
// 点数を1つ直したときに黙って食い違う。
func TestPinochleMeldTable(t *testing.T) {
	table := PinochleMeldTable()

	if len(table) != len(pinochleMeldPoints) {
		t.Fatalf("expected %d entries, got %d", len(pinochleMeldPoints), len(table))
	}
	seen := map[PinochleMeldType]bool{}
	for _, e := range table {
		if seen[e.Type] {
			t.Errorf("duplicate entry for meld %d", int(e.Type))
		}
		seen[e.Type] = true
		// **点数は表示用に書き写した値ではなく、加点に使う値そのもの。**
		if want := pinochleMeldPoints[e.Type]; e.Points != want {
			t.Errorf("meld %d: table says %d, scoring says %d", int(e.Type), e.Points, want)
		}
	}

	// 安い順に並ぶこと。並びが不定だと表示のたびに行が入れ替わる。
	for i := 1; i < len(table); i++ {
		if table[i-1].Points > table[i].Points {
			t.Errorf("entry %d (%d pts) must not come after %d pts", i, table[i].Points, table[i-1].Points)
		}
	}
	if table[0].Type != PinochleMeldDix || table[0].Points != 10 {
		t.Errorf("first entry should be the 10-point dix, got type=%d points=%d", int(table[0].Type), table[0].Points)
	}
	last := table[len(table)-1]
	if last.Type != PinochleMeldDoubleRun || last.Points != 1500 {
		t.Errorf("last entry should be the 1500-point double run, got type=%d points=%d", int(last.Type), last.Points)
	}

	// 同点のメルドは種類の順で並ぶ (40点が3種類ある)。
	var at40 []PinochleMeldType
	for _, e := range table {
		if e.Points == 40 {
			at40 = append(at40, e.Type)
		}
	}
	want40 := []PinochleMeldType{PinochleMeldRoyalMarriage, PinochleMeldPinochle, PinochleMeldJacksAround}
	if len(at40) != len(want40) {
		t.Fatalf("expected %d melds worth 40, got %d", len(want40), len(at40))
	}
	for i := range want40 {
		if at40[i] != want40[i] {
			t.Errorf("40-point order[%d]: got %d, want %d", i, int(at40[i]), int(want40[i]))
		}
	}
}
