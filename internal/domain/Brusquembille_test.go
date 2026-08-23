//go:build !js || !wasm || classic

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestBrusquembille() *domain.Brusquembille {
	return domain.NewDefaultBrusquembille()
}

func TestNewDefaultBrusquembille(t *testing.T) {
	b := domain.NewDefaultBrusquembille()
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

func TestBrusquembille_Reset(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()

	if b.GetPhase() != domain.BrusquembillePhasePlay {
		t.Errorf("phase = %d, want Play", b.GetPhase())
	}
	if b.GetTrickNumber() != 1 {
		t.Errorf("trick = %d, want 1", b.GetTrickNumber())
	}
	if b.GetPlayer(0).GetCardsSize() != domain.BrusquembilleHandSize {
		t.Errorf("p0 hand size = %d, want %d", b.GetPlayer(0).GetCardsSize(), domain.BrusquembilleHandSize)
	}
	if b.GetPlayer(1).GetCardsSize() != domain.BrusquembilleHandSize {
		t.Errorf("p1 hand size = %d, want %d", b.GetPlayer(1).GetCardsSize(), domain.BrusquembilleHandSize)
	}
	if b.GetTrumpCard() == nil {
		t.Fatal("trump card should be set after Reset")
	}
	if b.GetTrumpSuit() != b.GetTrumpCard().GetDesign() {
		t.Errorf("trump suit %d != trump card suit %d", b.GetTrumpSuit(), b.GetTrumpCard().GetDesign())
	}
	// **32 枚デッキ。** クローン元のブリスコラは 40 枚のイタリア式だが、
	// ブリュスカンビーユは 32 枚のフランス式 (7-8-9-10-J-Q-K-A)。
	// Stock = 32 - 6 dealt - 1 trumpCard drawn = 25
	if got := b.GetStockRemaining(); got != 25 {
		t.Errorf("stock remaining = %d, want 25 (32 - 6 - 1)", got)
	}
	// Lead is dealer's left (dealer=0 → lead=1)
	if b.GetLeadPlayerIdx() != 1 {
		t.Errorf("lead = %d, want 1", b.GetLeadPlayerIdx())
	}
}

func TestBrusquembille_CardPointsAndRank(t *testing.T) {
	// A > 10 > K > Q > J > 9 > 8 > 7、点は A=11 / 10=10 / K=4 / Q=3 / J=2。
	//
	// **クローン元のブリスコラの表とは別物。** ブリスコラは 40 枚デッキで
	// 「3」が 10 点かつ A に次ぐ強さだが、32 枚デッキに 3 は無い。逆に
	// ブリスコラの表は **10 を一度も挙げていない**ので、そのまま使うと 10 が
	// 既定値 0 に落ちて盤で一番弱い札になる —— このゲームで A に次ぐ札が。
	cases := []struct {
		val  int
		pts  int
		rank int
	}{
		{1, 11, 8},  // As
		{10, 10, 7}, // Dix
		{13, 4, 6},  // Roi
		{12, 3, 5},  // Dame
		{11, 2, 4},  // Valet
		{9, 0, 3},
		{8, 0, 2},
		{7, 0, 1},
	}
	for _, c := range cases {
		card := domain.NewCard(domain.CardDesignSpade, c.val, false)
		if got := domain.BrusquembilleCardPoints(card); got != c.pts {
			t.Errorf("BrusquembilleCardPoints(val=%d) = %d, want %d", c.val, got, c.pts)
		}
		if got := domain.BrusquembilleRankOrder(card); got != c.rank {
			t.Errorf("BrusquembilleRankOrder(val=%d) = %d, want %d", c.val, got, c.rank)
		}
	}
	if domain.BrusquembilleCardPoints(nil) != 0 || domain.BrusquembilleRankOrder(nil) != 0 {
		t.Error("nil card should return 0 for points and rank")
	}
}

// TestBrusquembille_TablesCoverTheWholeDeck は、点数表と強さ表が
// **実際に配られる 32 枚すべて**を説明していることを見る。
//
// **これが無いと、表に無い札が静かに「0」になる。** クローン元の強さ表は
// 2/4/5/6 という 32 枚デッキに存在しない札を並べる一方、10 を挙げていない。
// 個別のケースを並べるだけのテストは「書いた分だけ」しか見ないので、
// 抜けた 1 枚に気づけない —— デッキ側から数える。
func TestBrusquembille_TablesCoverTheWholeDeck(t *testing.T) {
	deck := domain.NewTrumpCards32()
	total, seen := 0, 0
	ranks := map[int]bool{}

	for {
		c := deck.DrawCard()
		if c == nil {
			break
		}
		seen++
		total += domain.BrusquembilleCardPoints(c)

		r := domain.BrusquembilleRankOrder(c)
		if r <= 0 {
			t.Fatalf("value %d has no rank order — it would be the weakest card on the table", c.GetValue())
		}
		ranks[c.GetValue()] = true
	}

	if seen != 32 {
		t.Fatalf("dealt %d cards, want 32", seen)
	}
	if len(ranks) != 8 {
		t.Fatalf("%d distinct ranks got an order, want 8", len(ranks))
	}
	// A(11)+10(10)+K(4)+Q(3)+J(2) = 30 per suit × 4 = 120。
	// **合計が 120 にならなければ、点のある札が盤から消えている。**
	if total != 120 {
		t.Fatalf("the whole deck is worth %d points, want 120", total)
	}
}

func TestBrusquembille_ValidatePlay_AlwaysAllows(t *testing.T) {
	// must-follow がないことを示す: トリックにリードがあっても、
	// 手札のどのカードでも有効。
	b := newTestBrusquembille()
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

func TestBrusquembille_TrickWinner(t *testing.T) {
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
			challenge: domain.NewCard(domain.CardDesignSpade, 7, false),  // 7 of trumps (weakest card)
			want:      1,
		},
		{
			// **10 は K より強い。** ブリュスカンビーユの序列は
			// A > 10 > K > Q > J > 9 > 8 > 7 で、10 が 2 番目に強い。
			// クローン元 (ブリスコラ) の表をそのまま使うと 10 は序列に載らず、
			// 盤で一番弱い札に落ちるので、ここが逆さまになる。
			name:      "the ten beats the king in the trump suit",
			trumpSuit: domain.CardDesignSpade,
			lead:      domain.NewCard(domain.CardDesignSpade, 13, false), // K trumps
			challenge: domain.NewCard(domain.CardDesignSpade, 10, false), // 10 trumps
			want:      1,
		},
		{
			name:      "the ace still beats the ten",
			trumpSuit: domain.CardDesignSpade,
			lead:      domain.NewCard(domain.CardDesignSpade, 1, false),  // A trumps
			challenge: domain.NewCard(domain.CardDesignSpade, 10, false), // 10 trumps
			want:      0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTestBrusquembille()
			b.SetTrumpSuit(tt.trumpSuit)
			b.SetCurrentTrick([]*domain.TrickCard{
				{PlayerIdx: 0, Card: tt.lead},
				{PlayerIdx: 1, Card: tt.challenge},
			})
			b.SetPhase(domain.BrusquembillePhaseTrickEnd)
			b.SetLeadPlayerIdx(0)
			b.ResolveTrick()
			gotWinner := b.GetLeadPlayerIdx()
			if gotWinner != tt.want {
				t.Errorf("winner = %d, want %d", gotWinner, tt.want)
			}
		})
	}
}

func TestBrusquembille_ResolveTrick_AwardsPoints(t *testing.T) {
	b := newTestBrusquembille()
	b.SetTrumpSuit(domain.CardDesignSpade)
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 1, false)},  // A=11pt, lead
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 13, false)}, // K=4pt
	})
	b.SetPhase(domain.BrusquembillePhaseTrickEnd)
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

func TestBrusquembille_PlayerPlay_HappyPath(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	// Force human (idx 0) to be on turn for deterministic test
	b.SetCurrentPlayerIdx(0)
	b.SetLeadPlayerIdx(0)
	if err := b.PlayerPlay(0); err != nil {
		t.Fatalf("PlayerPlay: %v", err)
	}
	if b.GetPlayer(0).GetCardsSize() != domain.BrusquembilleHandSize-1 {
		t.Errorf("hand size after play = %d, want %d", b.GetPlayer(0).GetCardsSize(), domain.BrusquembilleHandSize-1)
	}
	if len(b.GetCurrentTrick()) != 1 {
		t.Errorf("trick has %d cards, want 1", len(b.GetCurrentTrick()))
	}
	if b.GetCurrentPlayerIdx() != 1 {
		t.Errorf("turn passed to %d, want 1", b.GetCurrentPlayerIdx())
	}
}

func TestBrusquembille_PlayerPlay_Errors(t *testing.T) {
	b := newTestBrusquembille()
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
		b2 := newTestBrusquembille()
		b2.Reset()
		b2.SetGameEndFlag(true)
		if err := b2.PlayerPlay(0); err == nil {
			t.Error("want ErrGameEnded")
		}
	})

	t.Run("wrong phase", func(t *testing.T) {
		b2 := newTestBrusquembille()
		b2.Reset()
		b2.SetPhase(domain.BrusquembillePhaseTrickEnd)
		if err := b2.PlayerPlay(0); err == nil {
			t.Error("want ErrWrongPhase")
		}
	})

	t.Run("not human turn", func(t *testing.T) {
		b2 := newTestBrusquembille()
		b2.Reset()
		b2.SetCurrentPlayerIdx(1) // CPU
		if err := b2.PlayerPlay(0); err == nil {
			t.Error("want ErrNotHumanTurn")
		}
	})
}

func TestBrusquembille_CpuPlay(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	b.SetCurrentPlayerIdx(1) // CPU
	beforeSize := b.GetPlayer(1).GetCardsSize()
	b.CpuPlay()
	if b.GetPlayer(1).GetCardsSize() != beforeSize-1 {
		t.Errorf("CPU hand size = %d, want %d", b.GetPlayer(1).GetCardsSize(), beforeSize-1)
	}
}

func TestBrusquembille_CpuPlay_NoOpWhenNotCPUTurn(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	b.SetCurrentPlayerIdx(0) // human
	beforeSize := b.GetPlayer(0).GetCardsSize()
	b.CpuPlay()
	if b.GetPlayer(0).GetCardsSize() != beforeSize {
		t.Error("CpuPlay should be no-op when human's turn")
	}
}

func TestBrusquembille_NextTrick_DrawReplenish(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	// Play one trick: human leads, CPU follows
	b.SetCurrentPlayerIdx(0)
	b.SetLeadPlayerIdx(0)
	if err := b.PlayerPlay(0); err != nil {
		t.Fatalf("p0 play: %v", err)
	}
	b.CpuPlay() // p1 plays; trick should now be at TrickEnd
	if b.GetPhase() != domain.BrusquembillePhaseTrickEnd {
		t.Fatalf("phase = %d, want TrickEnd", b.GetPhase())
	}
	stockBefore := b.GetStockRemaining()
	b.ResolveTrick()
	b.NextTrick()
	// After draw, both hands should be back to BrusquembilleHandSize, stock down by 2
	if got := b.GetPlayer(0).GetCardsSize(); got != domain.BrusquembilleHandSize {
		t.Errorf("p0 hand after draw = %d, want %d", got, domain.BrusquembilleHandSize)
	}
	if got := b.GetPlayer(1).GetCardsSize(); got != domain.BrusquembilleHandSize {
		t.Errorf("p1 hand after draw = %d, want %d", got, domain.BrusquembilleHandSize)
	}
	if b.GetStockRemaining() != stockBefore-2 {
		t.Errorf("stock = %d, want %d", b.GetStockRemaining(), stockBefore-2)
	}
	if b.GetTrickNumber() != 2 {
		t.Errorf("trick = %d, want 2", b.GetTrickNumber())
	}
}

func TestBrusquembille_FullHand_GameEnds(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	// Drive the game to completion by always letting human play idx 0
	for !b.GetGameEndFlag() {
		switch b.GetPhase() {
		case domain.BrusquembillePhasePlay:
			if b.IsHumanTurn() {
				if err := b.PlayerPlay(0); err != nil {
					t.Fatalf("PlayerPlay: %v", err)
				}
			} else {
				b.CpuPlay()
			}
		case domain.BrusquembillePhaseTrickEnd:
			b.ResolveTrick()
			b.NextTrick()
		default:
			t.Fatalf("unexpected phase %d", b.GetPhase())
		}
	}
	if b.GetPhase() != domain.BrusquembillePhaseGameEnd {
		t.Errorf("phase after game = %d, want GameEnd", b.GetPhase())
	}
	if got := b.GetPlayerPoints(0) + b.GetPlayerPoints(1); got != domain.BrusquembilleTotalPoints {
		t.Errorf("total points = %d, want %d", got, domain.BrusquembilleTotalPoints)
	}
	// Trump card and stock should be exhausted
	if b.GetStockRemaining() != 0 {
		t.Errorf("stock = %d, want 0", b.GetStockRemaining())
	}
	if b.GetPlayer(0).GetCardsSize() != 0 || b.GetPlayer(1).GetCardsSize() != 0 {
		t.Error("hands should be empty after game")
	}
}

func TestBrusquembilleDetermineWinner(t *testing.T) {
	cases := []struct {
		name   string
		points []int
		want   int
	}{
		// 2 人卓は従来どおり。合計 120 点を二人で分けるので、単独最多は必ず 60 点超。
		{"p0 wins clearly", []int{80, 40}, 0},
		{"p1 wins clearly", []int{40, 80}, 1},
		{"p0 wins by 1", []int{61, 59}, 0},
		{"tie 60-60", []int{60, 60}, -1},
		{"both below threshold (degenerate)", []int{30, 30}, -1},

		// **3 席以上を見る。** クローン元の「席 0 か席 1 か」の形のままだと、
		// 席 2 以降がどれだけ取っても勝者にならない。
		{"seat 2 takes the most at a 3-seat table", []int{30, 40, 50}, 2},
		{"seat 4 takes the most at a 5-seat table", []int{20, 20, 20, 20, 40}, 4},
		{"a three-way tie has no winner", []int{40, 40, 40}, -1},
		{"the leader wins without passing 60", []int{50, 35, 35}, 0},
		{"two seats level at the top is a draw", []int{50, 50, 20}, -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := domain.BrusquembilleDetermineWinner(c.points); got != c.want {
				t.Errorf("BrusquembilleDetermineWinner(%v) = %d, want %d", c.points, got, c.want)
			}
		})
	}
}

func TestBrusquembille_GetHint(t *testing.T) {
	b := newTestBrusquembille()
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

func TestBrusquembille_GetHint_NoneWhenNotPlayPhase(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	b.SetPhase(domain.BrusquembillePhaseTrickEnd)
	if got := b.GetHint(); got != nil {
		t.Errorf("expected nil hint, got %+v", got)
	}
}

func TestBrusquembille_GetHint_NoneWhenCpuTurn(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	b.SetCurrentPlayerIdx(1)
	if got := b.GetHint(); got != nil {
		t.Errorf("expected nil hint, got %+v", got)
	}
}

func TestBrusquembille_GetValidPlayIndices(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	indices := b.GetValidPlayIndices(0)
	if len(indices) != domain.BrusquembilleHandSize {
		t.Errorf("valid count = %d, want %d", len(indices), domain.BrusquembilleHandSize)
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

func TestBrusquembille_Getters(t *testing.T) {
	b := newTestBrusquembille()
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
	if cfg.CpuDifficulty != domain.BrusquembilleCpuDifficultyNormal {
		t.Error("default config not preserved")
	}
}

func TestBrusquembille_SetConfig(t *testing.T) {
	b := newTestBrusquembille()
	b.SetConfig(domain.BrusquembilleConfig{CpuDifficulty: domain.BrusquembilleCpuDifficultyNormal})
	if b.GetConfig().CpuDifficulty != domain.BrusquembilleCpuDifficultyNormal {
		t.Error("SetConfig did not persist")
	}
}

func TestBrusquembille_JSONRoundtrip(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()
	// Mutate state to make the roundtrip meaningful
	b.SetPlayerPoints(0, 30)
	b.SetPlayerPoints(1, 15)

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got domain.Brusquembille
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

func TestBrusquembille_UnmarshalJSON_RejectsBadShape(t *testing.T) {
	cases := []struct {
		name    string
		payload string
	}{
		{"oversized players array", `{"ps":[` + brusqRepeat(`null,`, 1001) + `null]}`},
		// **2〜5 人。** クローン元は 2 人固定だった。
		{"too many players", `{"ps":[null,null,null,null,null,null]}`},
		{"one player", `{"ps":[null]}`},
		{"empty players array", `{"ps":[]}`},
		{"too many trick cards", `{"ps":[null,null],"ct":[{"pi":0,"c":null},{"pi":1,"c":null},{"pi":0,"c":null}]}`},
		{"player points length disagrees with the seats", `{"ps":[null,null],"pp":[0,0,0]}`},
		{"oversized action log", `{"ps":[null,null],"al":[` + brusqRepeat(`null,`, 1001) + `null]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b domain.Brusquembille
			if err := json.Unmarshal([]byte(tc.payload), &b); err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func brusqRepeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for range n {
		out = append(out, s...)
	}
	return string(out)
}

// TestBrusquembille_FollowBecomesRequiredWhenTheStockRuns は、このゲームの
// 肝である二相構造を、実際に打ち切って確かめる。
//
// **前半 (山札あり) は自由出し、後半 (山札切れ) は追従必須。** クローン元の
// ブリスコラは最後まで自由出しなので、この切り替えを足すのが実装の主眼。
// 述語を直接叩くのではなく本当に打ち切るのは、切り替わる瞬間 —— 山札が 0 に
// なっても場に出ている切札を誰かが取るまでは前半 —— を跨ぐため。
func TestBrusquembille_FollowBecomesRequiredWhenTheStockRuns(t *testing.T) {
	b := newTestBrusquembille()
	b.Reset()

	if b.IsFollowRequired() {
		t.Fatal("配った直後は山札が残っているので自由出しのはず")
	}

	sawFreePhase, sawFollowPhase := false, false
	for step := 0; step < 400 && !b.GetGameEndFlag(); step++ {
		if b.GetPhase() == domain.BrusquembillePhaseTrickEnd {
			b.ResolveTrick()
			b.NextTrick()
			continue
		}
		cur := b.GetCurrentPlayerIdx()
		valid := b.GetValidPlayIndices(cur)
		hand := b.GetPlayer(cur).GetCardsSize()
		if hand == 0 {
			break
		}
		if len(valid) == 0 {
			t.Fatalf("step %d: 手札 %d 枚あるのに合法手が 0 —— 誰も打てなくなる", step, hand)
		}

		if b.IsFollowRequired() {
			sawFollowPhase = true
			// **持っているなら追従しか許されない。**
			if lead := leadSuitOf(b); lead >= 0 && playerHasSuit(b, cur, lead) {
				for _, idx := range valid {
					if c := b.GetPlayer(cur).GetCard(idx); c.GetDesign() != lead {
						t.Fatalf("step %d: 追従できるのに別スート (%v) が合法とされた", step, c.GetDesign())
					}
				}
			}
		} else {
			sawFreePhase = true
			if len(valid) != hand {
				t.Fatalf("step %d: 前半は自由出しのはずが %d/%d しか合法でない", step, len(valid), hand)
			}
		}

		if cur == 0 {
			if err := b.PlayerPlay(valid[0]); err != nil {
				t.Fatalf("step %d: %v", step, err)
			}
		} else {
			b.CpuPlay()
		}
	}

	if !sawFreePhase {
		t.Error("前半 (自由出し) を一度も通っていない")
	}
	if !sawFollowPhase {
		t.Error("後半 (追従必須) を一度も通っていない —— 切り替えが起きていない")
	}
}

// leadSuitOf は現在のトリックのリードスートを返す (未リードなら -1)。
func leadSuitOf(b *domain.Brusquembille) int {
	trick := b.GetCurrentTrick()
	if len(trick) == 0 || trick[0].Card == nil {
		return -1
	}
	return trick[0].Card.GetDesign()
}

// playerHasSuit は seat がそのスートを持っているか。
func playerHasSuit(b *domain.Brusquembille, seat, suit int) bool {
	p := b.GetPlayer(seat)
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && c.GetDesign() == suit {
			return true
		}
	}
	return false
}

// TestBrusquembille_PlaysAtEveryTableSize は、2〜5 人のどの席数でも
// 最後まで打ち切れることを見る。
//
// **クローン元は 2 人固定だった。** トリックの長さも補充の順番も
// 「勝者ともう一人」で書かれていたので、席数を可変にした箇所が
// 一つでも取り残されていると、3 人以上で手番が回らなくなるか、
// 補充が偏るか、どちらかで止まる。
func TestBrusquembille_PlaysAtEveryTableSize(t *testing.T) {
	for cnt := domain.BrusquembilleMinPlayerCnt; cnt <= domain.BrusquembilleMaxPlayerCnt; cnt++ {
		cfg := domain.DefaultBrusquembilleConfig()
		cfg.PlayerCnt = cnt
		b := domain.NewBrusquembille(domain.NewTrumpCards32(),
			domain.NewBrusquembillePlayersForTable(cnt), cfg)
		b.Reset()

		if got := b.GetPlayerCnt(); got != cnt {
			t.Fatalf("cnt=%d: GetPlayerCnt() = %d", cnt, got)
		}

		dealt := 0
		for i := 0; i < cnt; i++ {
			dealt += b.GetPlayer(i).GetCardsSize()
		}
		// 配った枚数 + 山札 + 表向きの切札 = デッキ全部。どこにも消えていないこと。
		//
		// **席数で割り切れるまで低い札を抜く**ので、3 人卓・5 人卓は 30 枚に
		// なる (32 % 3 = 2、32 % 5 = 2)。抜かないと最後に手札が残って
		// 打ち切れない。
		want := 32 - 32%cnt
		total := dealt + b.GetStockRemaining()
		if b.GetTrumpCard() != nil {
			total++
		}
		if total != want {
			t.Fatalf("cnt=%d: %d cards accounted for, want %d", cnt, total, want)
		}
		if total%cnt != 0 {
			t.Fatalf("cnt=%d: %d cards do not divide by %d seats — the hands cannot empty evenly",
				cnt, total, cnt)
		}

		for step := 0; step < 600 && !b.GetGameEndFlag(); step++ {
			if b.GetPhase() == domain.BrusquembillePhaseTrickEnd {
				b.ResolveTrick()
				b.NextTrick()
				continue
			}
			cur := b.GetCurrentPlayerIdx()
			if b.GetPlayer(cur).GetCardsSize() == 0 {
				break
			}
			valid := b.GetValidPlayIndices(cur)
			if len(valid) == 0 {
				t.Fatalf("cnt=%d step=%d: seat %d has cards but no legal play", cnt, step, cur)
			}
			if cur == 0 {
				if err := b.PlayerPlay(valid[0]); err != nil {
					t.Fatalf("cnt=%d step=%d: %v", cnt, step, err)
				}
			} else {
				b.CpuPlay()
			}
		}

		if !b.GetGameEndFlag() {
			t.Errorf("cnt=%d: game did not finish", cnt)
		}
	}
}

// TestBrusquembille_SeatCountIsReachableThroughConfig は、席数が
// **設定を通して実際に効く**ことを見る。
//
// **ドメインが席数可変でも、設定が卓を組み直さなければ届かない。** 代入する
// だけだと config は 4 人卓と言っているのに players は 2 人のままで、Reset
// しても 2 人卓が始まる —— 設定が効いていないのに効いたように見える。
//
// 併せて、勝者が**全席**から選ばれることも見る。席 0/1 だけ比べる実装だと、
// 席 2 以降がどれだけ取っても勝者にならない。
func TestBrusquembille_SeatCountIsReachableThroughConfig(t *testing.T) {
	for cnt := domain.BrusquembilleMinPlayerCnt; cnt <= domain.BrusquembilleMaxPlayerCnt; cnt++ {
		b := domain.NewDefaultBrusquembille() // 既定は 2 人卓
		cfg := b.GetConfig()
		cfg.PlayerCnt = cnt
		if err := cfg.Validate(); err != nil {
			t.Fatalf("cnt=%d: config rejected: %v", cnt, err)
		}
		b.SetConfig(cfg)
		b.Reset()

		if got := b.GetPlayerCnt(); got != cnt {
			t.Fatalf("cnt=%d: SetConfig did not rebuild the table (got %d seats)", cnt, got)
		}

		for step := 0; step < 600 && !b.GetGameEndFlag(); step++ {
			if b.GetPhase() == domain.BrusquembillePhaseTrickEnd {
				b.ResolveTrick()
				b.NextTrick()
				continue
			}
			cur := b.GetCurrentPlayerIdx()
			if b.GetPlayer(cur).GetCardsSize() == 0 {
				break
			}
			valid := b.GetValidPlayIndices(cur)
			if len(valid) == 0 {
				t.Fatalf("cnt=%d step=%d: seat %d has cards but no legal play", cnt, step, cur)
			}
			if cur == 0 {
				if err := b.PlayerPlay(valid[0]); err != nil {
					t.Fatalf("cnt=%d step=%d: %v", cnt, step, err)
				}
			} else {
				b.CpuPlay()
			}
		}

		if !b.GetGameEndFlag() {
			t.Fatalf("cnt=%d: game did not finish", cnt)
		}

		// **勝者は全席から選ばれる。** 合計 120 点が卓に配られるので、
		// 引き分け (-1) でなければ勝者は必ず有効な席で、最多得点でなければならない。
		w := b.GetWinnerIdx()
		if w == -1 {
			continue
		}
		if w < 0 || w >= cnt {
			t.Fatalf("cnt=%d: winner %d is not a seat at this table", cnt, w)
		}
		for seat := 0; seat < cnt; seat++ {
			if seat != w && b.GetPlayerPoints(seat) > b.GetPlayerPoints(w) {
				t.Fatalf("cnt=%d: seat %d has %d points but seat %d was declared the winner with %d",
					cnt, seat, b.GetPlayerPoints(seat), w, b.GetPlayerPoints(w))
			}
		}
	}
}
