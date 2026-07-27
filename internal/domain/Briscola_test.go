package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestBriscola() *domain.Briscola {
	return domain.NewDefaultBriscola()
}

func TestNewDefaultBriscola(t *testing.T) {
	b := domain.NewDefaultBriscola()
	if got := b.GetPlayerCnt(); got != 2 {
		t.Errorf("GetPlayerCnt() = %d, want 2", got)
	}
	if !b.GetPlayer(0).GetIsHuman() {
		t.Error("player 0 should be human")
	}
	if b.GetPlayer(1).GetIsHuman() {
		t.Error("player 1 should be CPU")
	}
}

func TestBriscola_Reset(t *testing.T) {
	b := newTestBriscola()
	b.Reset()

	if b.GetPhase() != domain.BriscolaPhasePlay {
		t.Errorf("phase = %d, want Play", b.GetPhase())
	}
	if b.GetTrickNumber() != 1 {
		t.Errorf("trick = %d, want 1", b.GetTrickNumber())
	}
	if b.GetPlayer(0).GetCardsSize() != domain.BriscolaHandSize {
		t.Errorf("p0 hand size = %d, want %d", b.GetPlayer(0).GetCardsSize(), domain.BriscolaHandSize)
	}
	if b.GetPlayer(1).GetCardsSize() != domain.BriscolaHandSize {
		t.Errorf("p1 hand size = %d, want %d", b.GetPlayer(1).GetCardsSize(), domain.BriscolaHandSize)
	}
	if b.GetTrumpCard() == nil {
		t.Fatal("trump card should be set after Reset")
	}
	if b.GetTrumpSuit() != b.GetTrumpCard().GetDesign() {
		t.Errorf("trump suit %d != trump card suit %d", b.GetTrumpSuit(), b.GetTrumpCard().GetDesign())
	}
	// Stock = 40 - 6 dealt - 1 trumpCard drawn = 33
	if got := b.GetStockRemaining(); got != 33 {
		t.Errorf("stock remaining = %d, want 33 (40 - 6 - 1)", got)
	}
	// Lead is dealer's left (dealer=0 → lead=1)
	if b.GetLeadPlayerIdx() != 1 {
		t.Errorf("lead = %d, want 1", b.GetLeadPlayerIdx())
	}
}

func TestBriscola_CardPointsAndRank(t *testing.T) {
	cases := []struct {
		val  int
		pts  int
		rank int
	}{
		{1, 11, 10},
		{3, 10, 9},
		{13, 4, 8},
		{12, 3, 7},
		{11, 2, 6},
		{7, 0, 5},
		{6, 0, 4},
		{5, 0, 3},
		{4, 0, 2},
		{2, 0, 1},
	}
	for _, c := range cases {
		card := domain.NewCard(domain.CardDesignSpade, c.val, false)
		if got := domain.BriscolaCardPoints(card); got != c.pts {
			t.Errorf("BriscolaCardPoints(val=%d) = %d, want %d", c.val, got, c.pts)
		}
		if got := domain.BriscolaRankOrder(card); got != c.rank {
			t.Errorf("BriscolaRankOrder(val=%d) = %d, want %d", c.val, got, c.rank)
		}
	}
	if domain.BriscolaCardPoints(nil) != 0 || domain.BriscolaRankOrder(nil) != 0 {
		t.Error("nil card should return 0 for points and rank")
	}
}

func TestBriscola_ValidatePlay_AlwaysAllows(t *testing.T) {
	// must-follow がないことを示す: トリックにリードがあっても、
	// 手札のどのカードでも有効。
	b := newTestBriscola()
	b.Reset()
	// 任意の状態でも全カードがプレイ可能
	human := b.GetPlayer(0)
	for i := 0; i < human.GetCardsSize(); i++ {
		if card := human.GetCard(i); card == nil {
			t.Fatalf("card at %d is nil", i)
		}
	}
	indices := b.GetValidPlayIndices(0)
	if len(indices) != human.GetCardsSize() {
		t.Errorf("valid indices = %d, want %d", len(indices), human.GetCardsSize())
	}
}

func TestBriscola_TrickWinner(t *testing.T) {
	tests := []struct {
		name      string
		trumpSuit int
		lead      *domain.Card
		challenge *domain.Card
		want      int // expected winnerIdx (0 = lead, 1 = challenge)
	}{
		{
			name:      "lead suit higher rank wins",
			trumpSuit: domain.CardDesignSpade,
			lead:      domain.NewCard(domain.CardDesignClover, 1, false),  // A clovers (lead)
			challenge: domain.NewCard(domain.CardDesignClover, 13, false), // K clovers
			want:      0,                                                  // A>K so lead wins
		},
		{
			name:      "lead suit lower rank loses",
			trumpSuit: domain.CardDesignSpade,
			lead:      domain.NewCard(domain.CardDesignClover, 11, false), // J clovers
			challenge: domain.NewCard(domain.CardDesignClover, 1, false),  // A clovers
			want:      1,
		},
		{
			name:      "off-suit non-trump cannot win",
			trumpSuit: domain.CardDesignSpade,
			lead:      domain.NewCard(domain.CardDesignClover, 2, false), // weakest clover
			challenge: domain.NewCard(domain.CardDesignHeart, 1, false),  // A hearts (off-suit)
			want:      0,
		},
		{
			name:      "trump beats non-trump",
			trumpSuit: domain.CardDesignSpade,
			lead:      domain.NewCard(domain.CardDesignClover, 1, false), // A clovers
			challenge: domain.NewCard(domain.CardDesignSpade, 2, false),  // 2 of trumps
			want:      1,
		},
		{
			name:      "higher trump wins over lower trump",
			trumpSuit: domain.CardDesignSpade,
			lead:      domain.NewCard(domain.CardDesignSpade, 13, false), // K trumps
			challenge: domain.NewCard(domain.CardDesignSpade, 3, false),  // 3 trumps (rank 9)
			want:      1,
		},
		{
			name:      "lead trump wins over weaker trump",
			trumpSuit: domain.CardDesignSpade,
			lead:      domain.NewCard(domain.CardDesignSpade, 1, false), // A trumps
			challenge: domain.NewCard(domain.CardDesignSpade, 3, false), // 3 trumps
			want:      0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTestBriscola()
			b.SetTrumpSuit(tt.trumpSuit)
			b.SetCurrentTrick([]*domain.TrickCard{
				{PlayerIdx: 0, Card: tt.lead},
				{PlayerIdx: 1, Card: tt.challenge},
			})
			b.SetPhase(domain.BriscolaPhaseTrickEnd)
			b.SetLeadPlayerIdx(0)
			b.ResolveTrick()
			gotWinner := b.GetLeadPlayerIdx()
			if gotWinner != tt.want {
				t.Errorf("winner = %d, want %d", gotWinner, tt.want)
			}
		})
	}
}

func TestBriscola_ResolveTrick_AwardsPoints(t *testing.T) {
	b := newTestBriscola()
	b.SetTrumpSuit(domain.CardDesignSpade)
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 1, false)},  // A=11pt, lead
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 13, false)}, // K=4pt
	})
	b.SetPhase(domain.BriscolaPhaseTrickEnd)
	b.SetLeadPlayerIdx(0)
	b.ResolveTrick()
	// p0 wins (A > K) and gets 11 + 4 = 15 points
	if pts := b.GetPlayerPoints(0); pts != 15 {
		t.Errorf("p0 points = %d, want 15", pts)
	}
	if pts := b.GetPlayerPoints(1); pts != 0 {
		t.Errorf("p1 points = %d, want 0", pts)
	}
	if got := b.GetPlayer(0).GetTrickCount(); got != 1 {
		t.Errorf("p0 trick count = %d, want 1", got)
	}
}

func TestBriscola_PlayerPlay_HappyPath(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	// Force human (idx 0) to be on turn for deterministic test
	b.SetCurrentPlayerIdx(0)
	b.SetLeadPlayerIdx(0)
	if err := b.PlayerPlay(0); err != nil {
		t.Fatalf("PlayerPlay: %v", err)
	}
	if b.GetPlayer(0).GetCardsSize() != domain.BriscolaHandSize-1 {
		t.Errorf("hand size after play = %d, want %d", b.GetPlayer(0).GetCardsSize(), domain.BriscolaHandSize-1)
	}
	if len(b.GetCurrentTrick()) != 1 {
		t.Errorf("trick has %d cards, want 1", len(b.GetCurrentTrick()))
	}
	if b.GetCurrentPlayerIdx() != 1 {
		t.Errorf("turn passed to %d, want 1", b.GetCurrentPlayerIdx())
	}
}

func TestBriscola_PlayerPlay_Errors(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	b.SetCurrentPlayerIdx(0)

	t.Run("invalid card index", func(t *testing.T) {
		if err := b.PlayerPlay(-1); err == nil {
			t.Error("want error for negative index")
		}
		if err := b.PlayerPlay(99); err == nil {
			t.Error("want error for out-of-range index")
		}
	})

	t.Run("game ended", func(t *testing.T) {
		b2 := newTestBriscola()
		b2.Reset()
		b2.SetGameEndFlag(true)
		if err := b2.PlayerPlay(0); err == nil {
			t.Error("want ErrGameEnded")
		}
	})

	t.Run("wrong phase", func(t *testing.T) {
		b2 := newTestBriscola()
		b2.Reset()
		b2.SetPhase(domain.BriscolaPhaseTrickEnd)
		if err := b2.PlayerPlay(0); err == nil {
			t.Error("want ErrWrongPhase")
		}
	})

	t.Run("not human turn", func(t *testing.T) {
		b2 := newTestBriscola()
		b2.Reset()
		b2.SetCurrentPlayerIdx(1) // CPU
		if err := b2.PlayerPlay(0); err == nil {
			t.Error("want ErrNotHumanTurn")
		}
	})
}

func TestBriscola_CpuPlay(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	b.SetCurrentPlayerIdx(1) // CPU
	beforeSize := b.GetPlayer(1).GetCardsSize()
	b.CpuPlay()
	if b.GetPlayer(1).GetCardsSize() != beforeSize-1 {
		t.Errorf("CPU hand size = %d, want %d", b.GetPlayer(1).GetCardsSize(), beforeSize-1)
	}
}

func TestBriscola_CpuPlay_NoOpWhenNotCPUTurn(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	b.SetCurrentPlayerIdx(0) // human
	beforeSize := b.GetPlayer(0).GetCardsSize()
	b.CpuPlay()
	if b.GetPlayer(0).GetCardsSize() != beforeSize {
		t.Error("CpuPlay should be no-op when human's turn")
	}
}

func TestBriscola_NextTrick_DrawReplenish(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	// Play one trick: human leads, CPU follows
	b.SetCurrentPlayerIdx(0)
	b.SetLeadPlayerIdx(0)
	if err := b.PlayerPlay(0); err != nil {
		t.Fatalf("p0 play: %v", err)
	}
	b.CpuPlay() // p1 plays; trick should now be at TrickEnd
	if b.GetPhase() != domain.BriscolaPhaseTrickEnd {
		t.Fatalf("phase = %d, want TrickEnd", b.GetPhase())
	}
	stockBefore := b.GetStockRemaining()
	b.ResolveTrick()
	b.NextTrick()
	// After draw, both hands should be back to BriscolaHandSize, stock down by 2
	if got := b.GetPlayer(0).GetCardsSize(); got != domain.BriscolaHandSize {
		t.Errorf("p0 hand after draw = %d, want %d", got, domain.BriscolaHandSize)
	}
	if got := b.GetPlayer(1).GetCardsSize(); got != domain.BriscolaHandSize {
		t.Errorf("p1 hand after draw = %d, want %d", got, domain.BriscolaHandSize)
	}
	if b.GetStockRemaining() != stockBefore-2 {
		t.Errorf("stock = %d, want %d", b.GetStockRemaining(), stockBefore-2)
	}
	if b.GetTrickNumber() != 2 {
		t.Errorf("trick = %d, want 2", b.GetTrickNumber())
	}
}

func TestBriscola_FullHand_GameEnds(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	// Drive the game to completion by always letting human play idx 0
	for !b.GetGameEndFlag() {
		switch b.GetPhase() {
		case domain.BriscolaPhasePlay:
			if b.IsHumanTurn() {
				if err := b.PlayerPlay(0); err != nil {
					t.Fatalf("PlayerPlay: %v", err)
				}
			} else {
				b.CpuPlay()
			}
		case domain.BriscolaPhaseTrickEnd:
			b.ResolveTrick()
			b.NextTrick()
		default:
			t.Fatalf("unexpected phase %d", b.GetPhase())
		}
	}
	if b.GetPhase() != domain.BriscolaPhaseGameEnd {
		t.Errorf("phase after game = %d, want GameEnd", b.GetPhase())
	}
	if got := b.GetPlayerPoints(0) + b.GetPlayerPoints(1); got != domain.BriscolaTotalPoints {
		t.Errorf("total points = %d, want %d", got, domain.BriscolaTotalPoints)
	}
	// Trump card and stock should be exhausted
	if b.GetStockRemaining() != 0 {
		t.Errorf("stock = %d, want 0", b.GetStockRemaining())
	}
	if b.GetPlayer(0).GetCardsSize() != 0 || b.GetPlayer(1).GetCardsSize() != 0 {
		t.Error("hands should be empty after game")
	}
}

func TestBriscolaDetermineWinner(t *testing.T) {
	cases := []struct {
		name   string
		p0, p1 int
		want   int
	}{
		{"p0 wins clearly", 80, 40, 0},
		{"p1 wins clearly", 40, 80, 1},
		{"p0 wins by 1", 61, 59, 0},
		{"tie 60-60", 60, 60, -1},
		{"both below threshold (degenerate)", 30, 30, -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := domain.BriscolaDetermineWinner(c.p0, c.p1); got != c.want {
				t.Errorf("BriscolaDetermineWinner(%d, %d) = %d, want %d", c.p0, c.p1, got, c.want)
			}
		})
	}
}

func TestBriscola_GetHint(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	b.SetCurrentPlayerIdx(0)
	hint := b.GetHint()
	if hint == nil || hint.CardIndex == nil {
		t.Fatal("expected non-nil hint with card index")
	}
	idx := *hint.CardIndex
	if idx < 0 || idx >= b.GetPlayer(0).GetCardsSize() {
		t.Errorf("hint index %d out of range", idx)
	}
	if hint.Reason == "" {
		t.Error("expected non-empty hint reason")
	}
}

func TestBriscola_GetHint_NoneWhenNotPlayPhase(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	b.SetPhase(domain.BriscolaPhaseTrickEnd)
	if got := b.GetHint(); got != nil {
		t.Errorf("expected nil hint, got %+v", got)
	}
}

func TestBriscola_GetHint_NoneWhenCpuTurn(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	b.SetCurrentPlayerIdx(1)
	if got := b.GetHint(); got != nil {
		t.Errorf("expected nil hint, got %+v", got)
	}
}

func TestBriscola_GetValidPlayIndices(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	indices := b.GetValidPlayIndices(0)
	if len(indices) != domain.BriscolaHandSize {
		t.Errorf("valid count = %d, want %d", len(indices), domain.BriscolaHandSize)
	}
	for i, idx := range indices {
		if idx != i {
			t.Errorf("index[%d] = %d, want %d", i, idx, i)
		}
	}
	if got := b.GetValidPlayIndices(99); got != nil {
		t.Errorf("out-of-range player should return nil, got %v", got)
	}
}

func TestBriscola_Getters(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	if b.GetActionLog() == nil {
		t.Error("ActionLog should be non-nil after Reset")
	}
	if got := b.GetPlayer(99); got != nil {
		t.Error("out-of-range player should return nil")
	}
	if got := b.GetPlayerPoints(99); got != 0 {
		t.Error("out-of-range player points should return 0")
	}
	cfg := b.GetConfig()
	if cfg.CpuDifficulty != domain.BriscolaCpuDifficultyNormal {
		t.Error("default config not preserved")
	}
}

func TestBriscola_SetConfig(t *testing.T) {
	b := newTestBriscola()
	b.SetConfig(domain.BriscolaConfig{CpuDifficulty: domain.BriscolaCpuDifficultyNormal})
	if b.GetConfig().CpuDifficulty != domain.BriscolaCpuDifficultyNormal {
		t.Error("SetConfig did not persist")
	}
}

func TestBriscola_JSONRoundtrip(t *testing.T) {
	b := newTestBriscola()
	b.Reset()
	// Mutate state to make the roundtrip meaningful
	b.SetPlayerPoints(0, 30)
	b.SetPlayerPoints(1, 15)

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got domain.Briscola
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.GetPlayerCnt() != 2 {
		t.Errorf("player count after restore = %d, want 2", got.GetPlayerCnt())
	}
	if got.GetPlayerPoints(0) != 30 || got.GetPlayerPoints(1) != 15 {
		t.Errorf("points = %d/%d, want 30/15", got.GetPlayerPoints(0), got.GetPlayerPoints(1))
	}
	if got.GetTrumpCard() == nil {
		t.Error("trump card lost in roundtrip")
	}
}

func TestBriscola_UnmarshalJSON_RejectsBadShape(t *testing.T) {
	cases := []struct {
		name    string
		payload string
	}{
		{"oversized players array", `{"ps":[` + repeat(`null,`, 1001) + `null]}`},
		{"wrong player count", `{"ps":[null,null,null]}`},
		{"empty players array", `{"ps":[]}`},
		{"too many trick cards", `{"ps":[null,null],"ct":[{"pi":0,"c":null},{"pi":1,"c":null},{"pi":0,"c":null}]}`},
		{"wrong player points length", `{"ps":[null,null],"pp":[0,0,0]}`},
		{"oversized action log", `{"ps":[null,null],"al":[` + repeat(`null,`, 1001) + `null]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b domain.Briscola
			if err := json.Unmarshal([]byte(tc.payload), &b); err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func repeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for range n {
		out = append(out, s...)
	}
	return string(out)
}
