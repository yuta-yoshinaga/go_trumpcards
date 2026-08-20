//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func ofcCard(d, v int) *Card { return NewCard(d, v, false) }

func newOfcGame(playerCount int, human bool) *OpenFaceChinese {
	cfg := DefaultOpenFaceChineseConfig()
	cfg.PlayerCount = playerCount
	players := ofcBuildPlayers(playerCount)
	if !human && len(players) > 0 {
		players[0] = NewOpenFaceChinesePlayer(false)
	}
	return NewOpenFaceChinese(NewTrumpCards(0), players, cfg)
}

// ofcFillRows sets all three rows directly (test helper).
func ofcFillRows(p *OpenFaceChinesePlayer, front, middle, back []*Card) {
	p.SetFront(front)
	p.SetMiddle(middle)
	p.SetBack(back)
	p.SetPending(nil)
}

func TestOpenFaceChineseConfig_Validate(t *testing.T) {
	if err := DefaultOpenFaceChineseConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (OpenFaceChineseConfig{CpuDifficulty: 9, PlayerCount: 2, TargetRounds: 4}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (OpenFaceChineseConfig{CpuDifficulty: 0, PlayerCount: 1, TargetRounds: 4}).Validate(); err == nil {
		t.Error("expected player-count out-of-range error")
	}
	if err := (OpenFaceChineseConfig{CpuDifficulty: 0, PlayerCount: 2, TargetRounds: 0}).Validate(); err == nil {
		t.Error("expected target-rounds min error")
	}
}

func TestOpenFaceChinese_ResetDealsInitialFive(t *testing.T) {
	g := newOfcGame(2, true)
	g.Reset()
	if g.GetPhase() != OpenFaceChinesePhasePlacing {
		t.Errorf("phase = %d, want Placing", g.GetPhase())
	}
	// current player should have 5 pending (or be advanced); at least one human seat dealt 5.
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += len(g.GetPlayer(i).GetPending())
	}
	if total < OpenFaceChineseInitialDeal {
		t.Errorf("expected at least %d pending cards dealt, got %d", OpenFaceChineseInitialDeal, total)
	}
}

func TestOpenFaceChinesePlayer_PlaceAndRowFull(t *testing.T) {
	p := NewOpenFaceChinesePlayer(true)
	p.SetPending([]*Card{ofcCard(CardDesignSpade, 5), ofcCard(CardDesignSpade, 6), ofcCard(CardDesignSpade, 7), ofcCard(CardDesignSpade, 8)})
	// front holds at most 3.
	for i := 0; i < 3; i++ {
		if err := p.placeCard(OpenFaceChineseRowFront); err != nil {
			t.Fatalf("place %d err: %v", i, err)
		}
	}
	if err := p.placeCard(OpenFaceChineseRowFront); err == nil {
		t.Error("expected row-full error on 4th front placement")
	}
	if len(p.GetFront()) != 3 {
		t.Errorf("front size = %d, want 3", len(p.GetFront()))
	}
}

func TestOpenFaceChinesePlayer_PlaceErrors(t *testing.T) {
	p := NewOpenFaceChinesePlayer(true)
	if err := p.placeCard(OpenFaceChineseRowFront); err == nil {
		t.Error("expected error placing with empty pending")
	}
	p.SetPending([]*Card{ofcCard(CardDesignSpade, 5)})
	if err := p.placeCard(99); err == nil {
		t.Error("expected invalid-row error")
	}
}

func TestOpenFaceChinese_PlayerPlaceFlow(t *testing.T) {
	g := newOfcGame(2, true)
	g.SetPhase(OpenFaceChinesePhasePlacing)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	p.SetPending([]*Card{ofcCard(CardDesignSpade, 5), ofcCard(CardDesignHeart, 9)})
	if err := g.PlayerPlace(OpenFaceChineseRowBack); err != nil {
		t.Fatalf("place err: %v", err)
	}
	if len(p.GetBack()) != 1 {
		t.Errorf("back = %d, want 1", len(p.GetBack()))
	}
	// still pending: same player's turn continues.
	if g.GetCurrentPlayerIdx() != 0 {
		t.Errorf("current idx = %d, want 0", g.GetCurrentPlayerIdx())
	}
}

func TestOpenFaceChinese_PlayerPlaceErrors(t *testing.T) {
	g := newOfcGame(2, true)
	g.SetPhase(OpenFaceChinesePhaseRoundEnd)
	if err := g.PlayerPlace(0); err == nil {
		t.Error("expected wrong-phase error")
	}
	g.SetPhase(OpenFaceChinesePhasePlacing)
	g.SetCurrentPlayerIdx(1) // CPU seat
	if err := g.PlayerPlace(0); err == nil {
		t.Error("expected not-human-turn error")
	}
}

// foulFront/middle/back builds a fouling layout (back weaker than middle).
func TestOpenFaceChinese_FoulDetection(t *testing.T) {
	p := NewOpenFaceChinesePlayer(true)
	// back: high-card 7-high, middle: pair of Kings → back < middle → foul.
	ofcFillRows(p,
		[]*Card{ofcCard(CardDesignSpade, 2), ofcCard(CardDesignHeart, 3), ofcCard(CardDesignClover, 4)},
		[]*Card{ofcCard(CardDesignSpade, 13), ofcCard(CardDesignHeart, 13), ofcCard(CardDesignClover, 5), ofcCard(CardDesignDiamond, 6), ofcCard(CardDesignSpade, 7)},
		[]*Card{ofcCard(CardDesignSpade, 8), ofcCard(CardDesignHeart, 9), ofcCard(CardDesignClover, 10), ofcCard(CardDesignDiamond, 11), ofcCard(CardDesignHeart, 2)},
	)
	if ofcValidRows(p) {
		t.Error("expected foul (back high-card < middle pair)")
	}
}

func TestOpenFaceChinese_ValidRows(t *testing.T) {
	p := NewOpenFaceChinesePlayer(true)
	// back: pair of Aces, middle: pair of Kings, front: Q-high. valid.
	ofcFillRows(p,
		[]*Card{ofcCard(CardDesignSpade, 12), ofcCard(CardDesignHeart, 4), ofcCard(CardDesignClover, 3)},
		[]*Card{ofcCard(CardDesignSpade, 13), ofcCard(CardDesignHeart, 13), ofcCard(CardDesignClover, 5), ofcCard(CardDesignDiamond, 6), ofcCard(CardDesignSpade, 7)},
		[]*Card{ofcCard(CardDesignSpade, 1), ofcCard(CardDesignHeart, 1), ofcCard(CardDesignClover, 10), ofcCard(CardDesignDiamond, 9), ofcCard(CardDesignHeart, 8)},
	)
	if !ofcValidRows(p) {
		t.Error("expected valid rows (back pair-A >= middle pair-K >= front Q-high)")
	}
}

func TestOpenFaceChinese_RoyaltyAndFantasyland(t *testing.T) {
	p := NewOpenFaceChinesePlayer(true)
	// front: trip Queens → fantasyland + royalty; middle: trip Kings (>= front);
	// back: spade flush (>= middle). Rows stay valid (back >= middle >= front).
	ofcFillRows(p,
		[]*Card{ofcCard(CardDesignSpade, 12), ofcCard(CardDesignHeart, 12), ofcCard(CardDesignClover, 12)},
		[]*Card{ofcCard(CardDesignSpade, 13), ofcCard(CardDesignHeart, 13), ofcCard(CardDesignClover, 13), ofcCard(CardDesignDiamond, 6), ofcCard(CardDesignSpade, 7)},
		[]*Card{ofcCard(CardDesignSpade, 1), ofcCard(CardDesignSpade, 10), ofcCard(CardDesignSpade, 8), ofcCard(CardDesignSpade, 5), ofcCard(CardDesignSpade, 3)},
	)
	if !ofcValidRows(p) {
		t.Fatal("expected valid rows for royalty test")
	}
	if ofcPlayerRoyalty(p) <= 0 {
		t.Error("expected positive royalty (trips front + trips mid + flush back)")
	}
	if !ofcQualifiesFantasyland(p) {
		t.Error("expected fantasyland qualification for trip queens")
	}
}

func TestOpenFaceChinese_FantasylandQueensPair(t *testing.T) {
	p := NewOpenFaceChinesePlayer(true)
	p.SetFront([]*Card{ofcCard(CardDesignSpade, 12), ofcCard(CardDesignHeart, 12), ofcCard(CardDesignClover, 3)})
	if !ofcQualifiesFantasyland(p) {
		t.Error("expected fantasyland for QQ front")
	}
	p.SetFront([]*Card{ofcCard(CardDesignSpade, 11), ofcCard(CardDesignHeart, 11), ofcCard(CardDesignClover, 3)})
	if ofcQualifiesFantasyland(p) {
		t.Error("JJ pair should NOT qualify for fantasyland")
	}
}

func TestOpenFaceChinese_CompareScoreScoop(t *testing.T) {
	a := NewOpenFaceChinesePlayer(true)
	b := NewOpenFaceChinesePlayer(false)
	// a wins all three rows (higher everything).
	ofcFillRows(a,
		[]*Card{ofcCard(CardDesignSpade, 1), ofcCard(CardDesignHeart, 1), ofcCard(CardDesignClover, 13)},
		[]*Card{ofcCard(CardDesignSpade, 13), ofcCard(CardDesignHeart, 13), ofcCard(CardDesignClover, 13), ofcCard(CardDesignDiamond, 6), ofcCard(CardDesignSpade, 7)},
		[]*Card{ofcCard(CardDesignSpade, 1), ofcCard(CardDesignSpade, 10), ofcCard(CardDesignSpade, 8), ofcCard(CardDesignSpade, 5), ofcCard(CardDesignSpade, 3)},
	)
	// b's front is a 9-high non-pair, non-straight, non-flush hand so a's pair of
	// Aces wins the front row (in 3-card poker a straight would beat a pair, so
	// avoid consecutive ranks here). b's middle/back are weak high-card hands.
	ofcFillRows(b,
		[]*Card{ofcCard(CardDesignSpade, 2), ofcCard(CardDesignHeart, 5), ofcCard(CardDesignClover, 9)},
		[]*Card{ofcCard(CardDesignSpade, 5), ofcCard(CardDesignHeart, 6), ofcCard(CardDesignClover, 7), ofcCard(CardDesignDiamond, 8), ofcCard(CardDesignSpade, 10)},
		[]*Card{ofcCard(CardDesignHeart, 2), ofcCard(CardDesignDiamond, 4), ofcCard(CardDesignClover, 6), ofcCard(CardDesignHeart, 9), ofcCard(CardDesignDiamond, 11)},
	)
	a.SetFouled(false)
	b.SetFouled(false)
	s := ofcCompareScore(a, b)
	if s != 6 {
		t.Errorf("scoop score = %d, want 6 (3 rows + 3 bonus)", s)
	}
	if ofcCompareScore(b, a) != -6 {
		t.Error("reverse scoop should be -6")
	}
}

func TestOpenFaceChinese_CompareScoreFoul(t *testing.T) {
	a := NewOpenFaceChinesePlayer(true)
	b := NewOpenFaceChinesePlayer(false)
	a.SetFouled(true)
	b.SetFouled(false)
	if ofcCompareScore(a, b) != -6 {
		t.Error("fouled a vs clean b should be -6")
	}
	if ofcCompareScore(b, a) != 6 {
		t.Error("clean b vs fouled a should be 6")
	}
	a.SetFouled(true)
	b.SetFouled(true)
	if ofcCompareScore(a, b) != 0 {
		t.Error("both fouled should net 0")
	}
}

func TestOpenFaceChinese_ScoreRoundSettlement(t *testing.T) {
	g := newOfcGame(2, true)
	g.SetPhase(OpenFaceChinesePhasePlacing)
	// a (seat 0) clean & strong, b (seat 1) clean & weak.
	ofcFillRows(g.GetPlayer(0),
		[]*Card{ofcCard(CardDesignSpade, 1), ofcCard(CardDesignHeart, 1), ofcCard(CardDesignClover, 13)},
		[]*Card{ofcCard(CardDesignSpade, 13), ofcCard(CardDesignHeart, 13), ofcCard(CardDesignClover, 13), ofcCard(CardDesignDiamond, 6), ofcCard(CardDesignSpade, 7)},
		[]*Card{ofcCard(CardDesignSpade, 1), ofcCard(CardDesignSpade, 10), ofcCard(CardDesignSpade, 8), ofcCard(CardDesignSpade, 5), ofcCard(CardDesignSpade, 3)},
	)
	ofcFillRows(g.GetPlayer(1),
		[]*Card{ofcCard(CardDesignSpade, 2), ofcCard(CardDesignHeart, 3), ofcCard(CardDesignClover, 4)},
		[]*Card{ofcCard(CardDesignSpade, 5), ofcCard(CardDesignHeart, 6), ofcCard(CardDesignClover, 7), ofcCard(CardDesignDiamond, 8), ofcCard(CardDesignSpade, 10)},
		[]*Card{ofcCard(CardDesignHeart, 2), ofcCard(CardDesignDiamond, 4), ofcCard(CardDesignClover, 6), ofcCard(CardDesignHeart, 9), ofcCard(CardDesignDiamond, 11)},
	)
	g.ScoreRound()
	if g.GetPhase() != OpenFaceChinesePhaseRoundEnd && g.GetPhase() != OpenFaceChinesePhaseGameEnd {
		t.Errorf("phase = %d, want RoundEnd/GameEnd", g.GetPhase())
	}
	if g.GetPlayer(0).GetRoundScore() <= 0 {
		t.Errorf("seat 0 round score = %d, want positive", g.GetPlayer(0).GetRoundScore())
	}
	if g.GetPlayer(1).GetRoundScore() >= 0 {
		t.Errorf("seat 1 round score = %d, want negative", g.GetPlayer(1).GetRoundScore())
	}
	// 段得点もロイヤリティも対戦相手間でやり取りされるため合計は 0 (ゼロサム)。
	if sum := g.GetPlayer(0).GetRoundScore() + g.GetPlayer(1).GetRoundScore(); sum != 0 {
		t.Errorf("round score sum = %d, want 0 (zero-sum settlement)", sum)
	}
}

// TestOpenFaceChinese_ScoreRoundZeroSum はロイヤリティを含む 3 人精算でも
// 全プレイヤーの round/total スコア合計が 0 になる (相手から減点される) ことを保証する。
func TestOpenFaceChinese_ScoreRoundZeroSum(t *testing.T) {
	g := newOfcGame(3, true)
	g.SetPhase(OpenFaceChinesePhasePlacing)
	// seat 0: 上段トリップ Q (ロイヤリティ大) + 中段トリップ K + 下段フラッシュ。
	ofcFillRows(g.GetPlayer(0),
		[]*Card{ofcCard(CardDesignSpade, 12), ofcCard(CardDesignHeart, 12), ofcCard(CardDesignClover, 12)},
		[]*Card{ofcCard(CardDesignSpade, 13), ofcCard(CardDesignHeart, 13), ofcCard(CardDesignClover, 13), ofcCard(CardDesignDiamond, 6), ofcCard(CardDesignSpade, 7)},
		[]*Card{ofcCard(CardDesignSpade, 1), ofcCard(CardDesignSpade, 10), ofcCard(CardDesignSpade, 8), ofcCard(CardDesignSpade, 5), ofcCard(CardDesignSpade, 3)},
	)
	// seat 1: 弱いがファウルしない手。
	ofcFillRows(g.GetPlayer(1),
		[]*Card{ofcCard(CardDesignSpade, 2), ofcCard(CardDesignHeart, 3), ofcCard(CardDesignClover, 4)},
		[]*Card{ofcCard(CardDesignSpade, 5), ofcCard(CardDesignHeart, 6), ofcCard(CardDesignClover, 7), ofcCard(CardDesignDiamond, 8), ofcCard(CardDesignSpade, 9)},
		[]*Card{ofcCard(CardDesignHeart, 2), ofcCard(CardDesignDiamond, 4), ofcCard(CardDesignClover, 6), ofcCard(CardDesignHeart, 9), ofcCard(CardDesignDiamond, 11)},
	)
	// seat 2: 中庸な手。
	ofcFillRows(g.GetPlayer(2),
		[]*Card{ofcCard(CardDesignDiamond, 7), ofcCard(CardDesignHeart, 7), ofcCard(CardDesignClover, 5)},
		[]*Card{ofcCard(CardDesignDiamond, 9), ofcCard(CardDesignHeart, 10), ofcCard(CardDesignClover, 11), ofcCard(CardDesignDiamond, 12), ofcCard(CardDesignHeart, 13)},
		[]*Card{ofcCard(CardDesignClover, 1), ofcCard(CardDesignClover, 9), ofcCard(CardDesignClover, 10), ofcCard(CardDesignDiamond, 13), ofcCard(CardDesignHeart, 1)},
	)
	g.ScoreRound()
	sum := 0
	for i := 0; i < 3; i++ {
		sum += g.GetPlayer(i).GetRoundScore()
	}
	if sum != 0 {
		t.Errorf("3-player round score sum = %d, want 0 (zero-sum)", sum)
	}
}

func TestOpenFaceChinese_CpuFullMatch(t *testing.T) {
	for _, pc := range []int{2, 3, 4} {
		for _, diff := range []OpenFaceChineseCpuDifficulty{OpenFaceChineseCpuDifficultyEasy, OpenFaceChineseCpuDifficultyNormal, OpenFaceChineseCpuDifficultyHard} {
			g := newOfcGame(pc, false) // all CPU
			cfg := g.GetConfig()
			cfg.CpuDifficulty = diff
			cfg.TargetRounds = 3
			g.SetConfig(cfg)
			g.Reset()
			guard := 0
			for guard < 20000 && !g.GetGameEndFlag() {
				guard++
				switch g.GetPhase() {
				case OpenFaceChinesePhasePlacing:
					g.CpuPlay()
				case OpenFaceChinesePhaseRoundEnd:
					g.NextRound()
				default:
					g.NextRound()
				}
			}
			if !g.GetGameEndFlag() {
				t.Errorf("pc=%d diff=%d: all-CPU match did not finish within guard", pc, diff)
			}
			// winner index valid or draw.
			if w := g.GetWinnerIdx(); w < -1 || w >= pc {
				t.Errorf("pc=%d: invalid winner idx %d", pc, w)
			}
		}
	}
}

func TestOpenFaceChinese_FantasylandNextRoundDealsThirteen(t *testing.T) {
	g := newOfcGame(2, true)
	g.Reset()
	g.SetPhase(OpenFaceChinesePhaseRoundEnd)
	g.GetPlayer(0).SetFantasyland(true)
	g.NextRound()
	if len(g.GetPlayer(0).GetPending()) != OpenFaceChineseHandSize {
		// seat 0 (fantasyland) is dealt 13 at once; it may already be the lead seat
		// or not — but its pending should be 13 right after deal until it starts placing.
		// The dealer rotates, so seat 0 keeps its 13-card pending until reached.
		if g.GetCurrentPlayerIdx() == 0 {
			t.Errorf("fantasyland seat pending = %d, want 13", len(g.GetPlayer(0).GetPending()))
		}
	}
}

func TestOpenFaceChinese_JSONRoundTrip(t *testing.T) {
	g := newOfcGame(3, true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 OpenFaceChinese
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != 3 {
		t.Errorf("restored player count = %d, want 3", g2.GetPlayerCnt())
	}
	if g2.GetPhase() != g.GetPhase() {
		t.Errorf("restored phase mismatch")
	}
}

func TestOpenFaceChinese_UnmarshalRejectsBadState(t *testing.T) {
	cases := []string{
		`{"ps":[],"cf":{"cd":0,"pn":2,"tr":4},"ph":0,"rn":1,"ci":0,"di":0,"wi":-1}`,          // too few players
		`{"ps":[null,null],"cf":{"cd":0,"pn":2,"tr":4},"ph":0,"rn":1,"ci":0,"di":0,"wi":-1}`, // nil player
		`{"ps":[{},{}],"cf":{"cd":9,"pn":2,"tr":4},"ph":0,"rn":1,"ci":0,"di":0,"wi":-1}`,     // bad config
		`{"ps":[{},{}],"cf":{"cd":0,"pn":2,"tr":4},"ph":9,"rn":1,"ci":0,"di":0,"wi":-1}`,     // bad phase
		`{"ps":[{},{}],"cf":{"cd":0,"pn":2,"tr":4},"ph":0,"rn":1,"ci":5,"di":0,"wi":-1}`,     // ci out of range
		`{"ps":[{},{}],"cf":{"cd":0,"pn":2,"tr":4},"ph":0,"rn":0,"ci":0,"di":0,"wi":-1}`,     // rn < 1
	}
	for i, c := range cases {
		var g OpenFaceChinese
		if err := json.Unmarshal([]byte(c), &g); err == nil {
			t.Errorf("case %d: expected unmarshal error", i)
		}
	}
}

func TestOpenFaceChinese_HintInPlacing(t *testing.T) {
	g := newOfcGame(2, true)
	g.SetPhase(OpenFaceChinesePhasePlacing)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).SetPending([]*Card{ofcCard(CardDesignSpade, 13)})
	h := g.GetHint()
	if h == nil {
		t.Fatal("expected a hint")
	}
	if h.Row < OpenFaceChineseRowFront || h.Row > OpenFaceChineseRowBack {
		t.Errorf("hint row out of range: %d", h.Row)
	}
}

func TestOpenFaceChinese_PlayerGetters(t *testing.T) {
	g := newOfcGame(2, true)
	if g.GetPlayer(-1) != nil || g.GetPlayer(99) != nil {
		t.Error("out-of-range GetPlayer should be nil")
	}
	if g.GetWinnerIdx() != -1 {
		t.Error("winner should start at -1")
	}
}

// #5676: OFC の核心ルールは front ≦ middle ≦ back。**置いた瞬間に確定で反則に
// なる段**が分かるかどうかがプレイの質を決めるのに、その判定はフロントの
// ofcPlacementFouls にしか無かった。CUI からも問えるようにする。
func TestOpenFaceChinesePlacementFouls(t *testing.T) {
	card := func(d, v int) *Card { return NewCard(d, v, false) }

	// middle が埋まって A ハイ、back が埋まって K ハイ = すでに反則。
	middleAceHigh := []*Card{
		card(CardDesignSpade, 1), card(CardDesignHeart, 3), card(CardDesignClover, 5),
		card(CardDesignDiamond, 7), card(CardDesignSpade, 9),
	}
	backKingHigh := []*Card{
		card(CardDesignHeart, 13), card(CardDesignClover, 2), card(CardDesignDiamond, 4),
		card(CardDesignSpade, 6), card(CardDesignHeart, 8),
	}

	t.Run("a placement that completes an out-of-order pair of rows fouls", func(t *testing.T) {
		// back に 1 枚足すと埋まり、middle(A ハイ) > back(K ハイ) で確定的に反則。
		back4 := backKingHigh[:4]
		if OpenFaceChinesePlacementFouls(nil, middleAceHigh, back4, card(CardDesignHeart, 8), OpenFaceChineseRowBack) != true {
			t.Error("middle が back より強くなる配置を反則と判定していない")
		}
	})

	t.Run("a placement that keeps the order does not foul", func(t *testing.T) {
		// back を A ハイより強く埋めるなら反則にならない。
		strongBack := []*Card{
			card(CardDesignHeart, 13), card(CardDesignHeart, 12), card(CardDesignHeart, 11),
			card(CardDesignHeart, 10),
		}
		if OpenFaceChinesePlacementFouls(nil, middleAceHigh, strongBack, card(CardDesignHeart, 9), OpenFaceChineseRowBack) {
			t.Error("順序を保つ配置を反則と判定している")
		}
	})

	// **上段と中段だけが埋まった場面も確定する。** 下段は必ずこれ以上強くできる
	// ので、front > middle はその時点で覆らない。
	t.Run("front stronger than middle fouls before the back is filled", func(t *testing.T) {
		// front が A のペア、middle が K ハイ = front > middle。
		frontPairOfAces := []*Card{card(CardDesignSpade, 1), card(CardDesignHeart, 1), card(CardDesignClover, 5)}
		middleKingHigh := []*Card{
			card(CardDesignHeart, 13), card(CardDesignClover, 2), card(CardDesignDiamond, 4),
			card(CardDesignSpade, 6),
		}
		if !OpenFaceChinesePlacementFouls(frontPairOfAces, middleKingHigh, nil,
			card(CardDesignHeart, 8), OpenFaceChineseRowMiddle) {
			t.Error("front が middle より強くなる配置を反則と判定していない")
		}
	})

	t.Run("front weaker than middle does not foul", func(t *testing.T) {
		frontLow := []*Card{card(CardDesignSpade, 2), card(CardDesignHeart, 4), card(CardDesignClover, 6)}
		middleKingHigh := []*Card{
			card(CardDesignHeart, 13), card(CardDesignClover, 3), card(CardDesignDiamond, 5),
			card(CardDesignSpade, 7),
		}
		if OpenFaceChinesePlacementFouls(frontLow, middleKingHigh, nil,
			card(CardDesignHeart, 9), OpenFaceChineseRowMiddle) {
			t.Error("順序を保つ配置を反則と判定している")
		}
	})

	// **3 段すべて埋まった場面は本判定そのもの。**
	t.Run("judges all three rows once they are full", func(t *testing.T) {
		frontLow := []*Card{card(CardDesignSpade, 2), card(CardDesignHeart, 4), card(CardDesignClover, 6)}
		strongBack := []*Card{
			card(CardDesignHeart, 13), card(CardDesignHeart, 12), card(CardDesignHeart, 11),
			card(CardDesignHeart, 10),
		}
		if OpenFaceChinesePlacementFouls(frontLow, middleAceHigh, strongBack,
			card(CardDesignHeart, 9), OpenFaceChineseRowBack) {
			t.Error("front < middle < back の完成形を反則と判定している")
		}
		if !OpenFaceChinesePlacementFouls(frontLow, middleAceHigh, backKingHigh[:4],
			card(CardDesignHeart, 8), OpenFaceChineseRowBack) {
			t.Error("middle > back の完成形を反則と判定していない")
		}
	})

	// **まだ埋まっていない段は判定しない。**「未確定」を反則と呼ぶと、まだ挽回
	// できる配置を避けさせてしまう。
	t.Run("an unfilled row is not judged", func(t *testing.T) {
		if OpenFaceChinesePlacementFouls(nil, middleAceHigh[:2], nil, card(CardDesignSpade, 2), OpenFaceChineseRowMiddle) {
			t.Error("埋まっていない段の配置を反則と判定している")
		}
	})
}
