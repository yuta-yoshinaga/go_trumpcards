//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func klavCard(design, value int) *Card { return NewCard(design, value, false) }

func newKlavGame(human bool) *Klaverjas {
	players := make([]*KlaverjasPlayer, KlaverjasPlayerCnt)
	players[0] = NewKlaverjasPlayer(human)
	for i := 1; i < KlaverjasPlayerCnt; i++ {
		players[i] = NewKlaverjasPlayer(false)
	}
	return NewKlaverjas(NewTrumpCardsBelote(), players, DefaultKlaverjasConfig())
}

func klavSetHand(p *KlaverjasPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestKlaverjasConfig_Validate(t *testing.T) {
	if err := DefaultKlaverjasConfig().Validate(); err != nil {
		t.Fatalf("default invalid: %v", err)
	}
	bad := []KlaverjasConfig{
		{CpuDifficulty: -1, TargetPoints: 1501},
		{CpuDifficulty: 9, TargetPoints: 1501},
		{CpuDifficulty: 1, TargetPoints: 0},
	}
	for i, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestKlaverjas_ResetDealsAndSetsTrump(t *testing.T) {
	g := newKlavGame(true)
	g.Reset()
	if g.GetPhase() != KlaverjasPhasePlay {
		t.Fatalf("phase = %v, want Play", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != KlaverjasHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, KlaverjasHandSize)
		}
	}
	if g.GetTrumpSuit() < CardDesignSpade || g.GetTrumpSuit() > CardDesignMax {
		t.Errorf("trump suit not set: %d", g.GetTrumpSuit())
	}
}

func TestKlaverjas_TrumpStrengthJassHigh(t *testing.T) {
	g := newKlavGame(false)
	// trump strength: J>9>A>10>K>Q>8>7
	order := []int{11, 9, 1, 10, 13, 12, 8, 7}
	for i := 1; i < len(order); i++ {
		if g.trumpStrength(order[i-1]) <= g.trumpStrength(order[i]) {
			t.Errorf("trump strength order broken at %d", i)
		}
	}
	// plain strength: A>10>K>Q>J>9>8>7
	plain := []int{1, 10, 13, 12, 11, 9, 8, 7}
	for i := 1; i < len(plain); i++ {
		if klaverjasPlainStrength(plain[i-1]) <= klaverjasPlainStrength(plain[i]) {
			t.Errorf("plain strength order broken at %d", i)
		}
	}
}

func TestKlaverjas_CardPoints(t *testing.T) {
	g := newKlavGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// trump J=20, 9=14, A=11, 10=10, K=4, Q=3
	if g.cardPoints(klavCard(CardDesignDiamond, 11)) != 20 || g.cardPoints(klavCard(CardDesignDiamond, 9)) != 14 {
		t.Error("trump J/9 points wrong")
	}
	// non-trump A=11, 10=10, K=4, Q=3, J=2, 9=0
	if g.cardPoints(klavCard(CardDesignClover, 11)) != 2 || g.cardPoints(klavCard(CardDesignClover, 9)) != 0 {
		t.Error("plain J/9 points wrong")
	}
	if g.cardPoints(klavCard(CardDesignClover, 1)) != 11 {
		t.Error("plain ace points wrong")
	}
}

func TestKlaverjas_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newKlavGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// Lead A♣; trump J♦ (top trump) overtrumps and wins even vs another trump.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: klavCard(CardDesignClover, 1)},   // A♣ lead
		{PlayerIdx: 1, Card: klavCard(CardDesignDiamond, 9)},  // 9♦ Nel trump
		{PlayerIdx: 2, Card: klavCard(CardDesignDiamond, 11)}, // J♦ Jas (top trump)
		{PlayerIdx: 3, Card: klavCard(CardDesignClover, 10)},
	})
	if w := g.trickWinner(); w != 2 {
		t.Errorf("winner = %d, want 2 (J♦ top trump)", w)
	}
}

func TestKlaverjas_MustOvertrump(t *testing.T) {
	g := newKlavGame(true)
	g.SetPhase(KlaverjasPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	// Lead A♣ by p3; a trump 9♦ already played by p1. Human is void in clubs but
	// holds J♦ (beats 9♦) and 7♦ — must play the higher trump (J♦), not 7♦.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 3, Card: klavCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: klavCard(CardDesignDiamond, 9)},
	})
	klavSetHand(g.GetPlayer(0), klavCard(CardDesignDiamond, 11), klavCard(CardDesignDiamond, 7), klavCard(CardDesignHeart, 1))
	if err := g.PlayerPlay(1); err == nil { // 7♦ — cannot, must overtrump with J♦
		t.Error("expected must-overtrump error for the low trump")
	}
	if err := g.PlayerPlay(2); err == nil { // heart discard — has trump, must trump
		t.Error("expected must-trump error for the off-suit discard")
	}
	if err := g.PlayerPlay(0); err != nil { // J♦ overtrump — valid
		t.Fatalf("overtrump play err: %v", err)
	}
}

func TestKlaverjas_MustFollowAndTrumpWhenVoid(t *testing.T) {
	g := newKlavGame(true)
	g.SetPhase(KlaverjasPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: klavCard(CardDesignClover, 1)}})
	// Has a club → must follow (no trump on table yet, no overtrump needed).
	klavSetHand(g.GetPlayer(0), klavCard(CardDesignClover, 13), klavCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(1); err == nil { // diamond while holding club
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // K♣ follows
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestKlaverjas_RoemDetection(t *testing.T) {
	// 4-card sequence (A-K-Q-J of spades) = 50.
	p := NewKlaverjasPlayer(false)
	klavSetHand(p, klavCard(CardDesignSpade, 1), klavCard(CardDesignSpade, 13), klavCard(CardDesignSpade, 12), klavCard(CardDesignSpade, 11))
	if r := klaverjasHandRoem(p); r != 50 {
		t.Errorf("A-K-Q-J sequence roem = %d, want 50", r)
	}
	// 3-card sequence = 20.
	klavSetHand(p, klavCard(CardDesignHeart, 13), klavCard(CardDesignHeart, 12), klavCard(CardDesignHeart, 11), klavCard(CardDesignClover, 1))
	if r := klaverjasHandRoem(p); r != 20 {
		t.Errorf("K-Q-J sequence roem = %d, want 20", r)
	}
	// 4 of a kind (4 aces) = 100.
	klavSetHand(p, klavCard(CardDesignSpade, 1), klavCard(CardDesignClover, 1), klavCard(CardDesignHeart, 1), klavCard(CardDesignDiamond, 1))
	if r := klaverjasHandRoem(p); r != 100 {
		t.Errorf("four aces roem = %d, want 100", r)
	}
	// No roem.
	klavSetHand(p, klavCard(CardDesignSpade, 7), klavCard(CardDesignClover, 10), klavCard(CardDesignHeart, 13))
	if r := klaverjasHandRoem(p); r != 0 {
		t.Errorf("no-roem hand = %d, want 0", r)
	}
}

func TestKlaverjas_MustOvertrumpOnTrumpLead(t *testing.T) {
	// When the lead suit IS the trump suit, a player holding a higher trump must
	// still overtrump — playing a lower trump is illegal (regression: the lead==trump
	// case previously bypassed the overtrump check).
	g := newKlavGame(true)
	g.SetPhase(KlaverjasPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: klavCard(CardDesignDiamond, 9)}})        // 9♦ trump lead
	klavSetHand(g.GetPlayer(0), klavCard(CardDesignDiamond, 11), klavCard(CardDesignDiamond, 7)) // J♦ beats, 7♦ does not
	if err := g.PlayerPlay(1); err == nil {                                                      // 7♦ — must overtrump with J♦
		t.Error("expected must-overtrump error when a higher trump is held on a trump lead")
	}
	if err := g.PlayerPlay(0); err != nil { // J♦ overtrumps the led 9♦
		t.Fatalf("overtrump-on-trump-lead play err: %v", err)
	}
}

func TestKlaverjas_RoemMultipleRunsSameSuit(t *testing.T) {
	// Two disjoint runs in the same suit each score: 7-8-9 (20) + J-Q-K (20) = 40.
	p := NewKlaverjasPlayer(false)
	klavSetHand(p,
		klavCard(CardDesignSpade, 7), klavCard(CardDesignSpade, 8), klavCard(CardDesignSpade, 9),
		klavCard(CardDesignSpade, 11), klavCard(CardDesignSpade, 12), klavCard(CardDesignSpade, 13))
	if r := klaverjasHandRoem(p); r != 40 {
		t.Errorf("two same-suit runs roem = %d, want 40", r)
	}
}

func TestKlaverjas_ResolveTrickPointsAndLastBonus(t *testing.T) {
	g := newKlavGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(KlaverjasPhaseTrickEnd)
	g.SetTrickNumber(1)
	// trump J♦(20)+9♦(14) + A♣(11) + 7♣(0); J♦ wins for team 0.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: klavCard(CardDesignDiamond, 11)},
		{PlayerIdx: 1, Card: klavCard(CardDesignDiamond, 9)},
		{PlayerIdx: 2, Card: klavCard(CardDesignClover, 1)},
		{PlayerIdx: 3, Card: klavCard(CardDesignClover, 7)},
	})
	g.ResolveTrick()
	if g.GetRoundCardPoints()[0] != 45 {
		t.Errorf("team0 should have 45 pts, got %v", g.GetRoundCardPoints())
	}
	// Last trick +10.
	g.SetPhase(KlaverjasPhaseTrickEnd)
	g.SetTrickNumber(KlaverjasTrickCount)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: klavCard(CardDesignClover, 13)}, // K♣ 4pts wins
		{PlayerIdx: 1, Card: klavCard(CardDesignClover, 7)},
		{PlayerIdx: 2, Card: klavCard(CardDesignClover, 8)},
		{PlayerIdx: 3, Card: klavCard(CardDesignClover, 9)},
	})
	before := g.GetRoundCardPoints()[0]
	g.ResolveTrick()
	if g.GetRoundCardPoints()[0] != before+4+KlaverjasLastTrickBonus {
		t.Errorf("last trick should add 4+%d, got %v (before %d)", KlaverjasLastTrickBonus, g.GetRoundCardPoints(), before)
	}
	if g.GetPhase() != KlaverjasPhaseRoundEnd {
		t.Errorf("phase = %v, want RoundEnd", g.GetPhase())
	}
}

func TestKlaverjas_ScoreRoundAddsRoemAndEnds(t *testing.T) {
	g := newKlavGame(false)
	g.SetPhase(KlaverjasPhaseRoundEnd)
	cfg := g.GetConfig()
	cfg.TargetPoints = 100
	g.SetConfig(cfg)
	g.SetRoundCardPoints([KlaverjasTeamCnt]int{90, 72})
	g.roundRoem = [KlaverjasTeamCnt]int{20, 0}
	g.ScoreRound()
	if g.GetTeamScores()[0] != 110 {
		t.Errorf("team0 should have 90+20=110, got %v", g.GetTeamScores())
	}
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("team0 should win match")
	}
}

func TestKlaverjas_NextRound(t *testing.T) {
	g := newKlavGame(false)
	g.Reset()
	g.SetPhase(KlaverjasPhaseRoundEnd)
	d := g.GetDealerIdx()
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetDealerIdx() != (d+1)%KlaverjasPlayerCnt {
		t.Errorf("next round/dealer wrong")
	}
	if g.GetPhase() != KlaverjasPhasePlay {
		t.Errorf("phase = %v, want Play", g.GetPhase())
	}
}

func TestKlaverjas_CpuFullMatchProgresses(t *testing.T) {
	g := newKlavGame(false)
	g.Reset()
	for steps := 0; steps < 2000; steps++ {
		switch g.GetPhase() {
		case KlaverjasPhasePlay:
			g.CpuPlay()
		case KlaverjasPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == KlaverjasPhaseTrickEnd {
				g.NextTrick()
			}
		case KlaverjasPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case KlaverjasPhaseGameEnd:
			if g.GetWinnerTeam() < 0 {
				t.Error("match ended without winner")
			}
			return
		}
		if g.GetRoundNumber() > 200 {
			break
		}
	}
}

func TestKlaverjas_HintAndPlayable(t *testing.T) {
	g := newKlavGame(true)
	g.SetPhase(KlaverjasPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	klavSetHand(g.GetPlayer(0), klavCard(CardDesignClover, 7), klavCard(CardDesignSpade, 8))
	if h := g.GetHint(); h == nil || len(h.CardIndices) != 1 {
		t.Errorf("play hint missing: %+v", h)
	}
	if idx := g.GetPlayableIndices(0); len(idx) != 2 {
		t.Errorf("playable = %v, want 2 (lead)", idx)
	}
	g.SetPhase(KlaverjasPhaseRoundEnd)
	if idx := g.GetPlayableIndices(0); idx != nil {
		t.Error("playable should be nil outside play phase")
	}
}

func TestKlaverjas_PlayerPlayErrors(t *testing.T) {
	g := newKlavGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(1) // CPU
	if err := g.PlayerPlay(0); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KlaverjasPhaseRoundEnd)
	if err := g.PlayerPlay(0); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
	g.gameEndFlag = true
	if err := g.PlayerPlay(0); !errors.Is(err, ErrGameEnded) {
		t.Errorf("err = %v, want ErrGameEnded", err)
	}
}

func TestKlaverjas_JSONRoundTrip(t *testing.T) {
	g := newKlavGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Klaverjas
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPhase() != g.GetPhase() || g2.GetPlayerCnt() != KlaverjasPlayerCnt || g2.GetTrumpSuit() != g.GetTrumpSuit() {
		t.Error("round trip mismatch")
	}
}

func TestKlaverjas_UnmarshalErrors(t *testing.T) {
	var g Klaverjas
	if err := json.Unmarshal([]byte(`{"ps":[]}`), &g); !errors.Is(err, errKlaverjasInvalidPlayers) {
		t.Errorf("err = %v, want errKlaverjasInvalidPlayers", err)
	}
	bad := `{"al":[`
	for i := 0; i < klaverjasMaxSliceLen+1; i++ {
		if i > 0 {
			bad += ","
		}
		bad += `{"TurnNumber":0,"PlayerIdx":-1,"ActionType":"x","Detail":"y","Cards":null}`
	}
	bad += `]}`
	if err := json.Unmarshal([]byte(bad), &g); !errors.Is(err, errKlaverjasOversized) {
		t.Errorf("err = %v, want errKlaverjasOversized", err)
	}
}
