//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func marCard(design, value int) *Card { return NewCard(design, value, false) }

func newMarGame(human bool) *Marias {
	players := make([]*MariasPlayer, MariasPlayerCnt)
	players[0] = NewMariasPlayer(human)
	for i := 1; i < MariasPlayerCnt; i++ {
		players[i] = NewMariasPlayer(false)
	}
	return NewMarias(NewTrumpCardsBelote(), players, DefaultMariasConfig())
}

func marSetHand(p *MariasPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestMariasConfig_Validate(t *testing.T) {
	if err := DefaultMariasConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (MariasConfig{CpuDifficulty: 9, TargetPoints: 10}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (MariasConfig{CpuDifficulty: MariasCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestMarias_ResetDealsAndSetsSoloistTrump(t *testing.T) {
	g := newMarGame(true)
	g.Reset()
	if g.GetPhase() != MariasPhasePlay {
		t.Errorf("phase = %d, want Play", g.GetPhase())
	}
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
	}
	if total != MariasPlayerCnt*MariasHandSize {
		t.Errorf("dealt %d cards, want %d", total, MariasPlayerCnt*MariasHandSize)
	}
	if g.GetSoloistIdx() != (g.GetDealerIdx()+1)%MariasPlayerCnt {
		t.Errorf("soloist = %d, want forehand", g.GetSoloistIdx())
	}
	if g.GetTrumpSuit() < 1 || g.GetTrumpSuit() > 4 {
		t.Errorf("trump suit = %d, want 1-4", g.GetTrumpSuit())
	}
}

func TestMarias_StrengthAndPoints(t *testing.T) {
	order := []int{1, 10, 13, 12, 11, 9, 8, 7}
	for i := 1; i < len(order); i++ {
		if mariasStrength(order[i-1]) <= mariasStrength(order[i]) {
			t.Errorf("strength(%d) should beat strength(%d)", order[i-1], order[i])
		}
	}
	cases := map[int]int{1: 11, 10: 10, 13: 4, 12: 3, 11: 2, 9: 0, 8: 0, 7: 0}
	total := 0
	for v, want := range cases {
		if got := mariasCardPoints(marCard(CardDesignSpade, v)); got != want {
			t.Errorf("points(%d) = %d, want %d", v, got, want)
		}
		total += want
	}
	if total*4 != 120 {
		t.Errorf("deck card points = %d, want 120", total*4)
	}
}

func TestMarias_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newMarGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: marCard(CardDesignClover, 1)},  // Ace lead
		{PlayerIdx: 1, Card: marCard(CardDesignDiamond, 7)}, // low trump beats it
		{PlayerIdx: 2, Card: marCard(CardDesignClover, 10)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("trick winner = %d, want 1 (trump beats high plain)", w)
	}
}

func TestMarias_MustFollowAndTrumpWhenVoid(t *testing.T) {
	g := newMarGame(true)
	g.SetPhase(MariasPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: marCard(CardDesignClover, 1)}})
	// Holds a club -> must follow.
	marSetHand(g.GetPlayer(0), marCard(CardDesignClover, 13), marCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(1); err == nil {
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("valid follow err: %v", err)
	}
	// Void in clubs but holds a trump -> must trump.
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: marCard(CardDesignClover, 1)}})
	g.SetCurrentPlayerIdx(0)
	marSetHand(g.GetPlayer(0), marCard(CardDesignHeart, 13), marCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(0); err == nil { // heart discard while holding trump
		t.Error("expected must-trump error")
	}
	if err := g.PlayerPlay(1); err != nil { // 7♦ trump
		t.Fatalf("valid trump err: %v", err)
	}
}

func TestMarias_MarriageDetection(t *testing.T) {
	g := newMarGame(false)
	g.SetTrumpSuit(CardDesignHeart)
	// Plain marriage K+Q of spades = 20; trump marriage K+Q of hearts = 40.
	marSetHand(g.GetPlayer(0),
		marCard(CardDesignSpade, 13), marCard(CardDesignSpade, 12),
		marCard(CardDesignHeart, 13), marCard(CardDesignHeart, 12))
	g.SetRoundMarriage([MariasPlayerCnt]int{})
	g.detectMarriages()
	rm := g.GetRoundMarriage()
	if rm[0] != 60 {
		t.Errorf("player 0 marriage = %d, want 60 (20 plain + 40 trump)", rm[0])
	}
}

func TestMarias_ResolveTrickPointsAndLastBonus(t *testing.T) {
	g := newMarGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(MariasPhaseTrickEnd)
	g.SetTrickNumber(MariasTrickCount) // final trick -> +10 bonus
	// A♣(11) + 10♣(10) + K♣(4) = 25; A♣ wins for player 0, +10 last-trick bonus = 35.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: marCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: marCard(CardDesignClover, 10)},
		{PlayerIdx: 2, Card: marCard(CardDesignClover, 13)},
	})
	g.ResolveTrick()
	pts := g.GetRoundCardPoints()
	if pts[0] != 35 {
		t.Errorf("player 0 round points = %d, want 35 (25 + 10 last)", pts[0])
	}
	if g.GetPhase() != MariasPhaseRoundEnd {
		t.Errorf("phase after final trick = %d, want RoundEnd", g.GetPhase())
	}
}

func TestMarias_ScoreRoundSoloistWinAndLoss(t *testing.T) {
	// Soloist wins: soloist total > defense.
	g := newMarGame(false)
	g.SetPhase(MariasPhaseRoundEnd)
	g.SetSoloistIdx(0)
	g.SetRoundCardPoints([MariasPlayerCnt]int{80, 20, 20})
	g.ScoreRound()
	if sc := g.GetPlayerScores(); sc[0] != 2 {
		t.Errorf("soloist score = %d, want 2 on win", sc[0])
	}
	// Soloist loses: each defender +2.
	g2 := newMarGame(false)
	g2.SetPhase(MariasPhaseRoundEnd)
	g2.SetSoloistIdx(0)
	g2.SetRoundCardPoints([MariasPlayerCnt]int{10, 60, 50})
	g2.ScoreRound()
	sc := g2.GetPlayerScores()
	if sc[0] != 0 || sc[1] != 2 || sc[2] != 2 {
		t.Errorf("scores = %v, want [0 2 2] on soloist loss", sc)
	}
}

func TestMarias_ScoreRoundEndsMatch(t *testing.T) {
	g := newMarGame(false)
	g.SetPhase(MariasPhaseRoundEnd)
	g.SetSoloistIdx(0)
	g.SetPlayerScores([MariasPlayerCnt]int{9, 0, 0})
	g.SetRoundCardPoints([MariasPlayerCnt]int{90, 10, 10})
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerPlayer() != 0 {
		t.Errorf("expected match end with player 0 winning, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerPlayer())
	}
}

func TestMarias_NextRound(t *testing.T) {
	g := newMarGame(false)
	g.SetPhase(MariasPhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("round number = %d, want 2", g.GetRoundNumber())
	}
	if g.GetPhase() != MariasPhasePlay {
		t.Errorf("phase after NextRound = %d, want Play", g.GetPhase())
	}
}

func TestMarias_HintAndPlayable(t *testing.T) {
	g := newMarGame(true)
	g.SetPhase(MariasPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	marSetHand(g.GetPlayer(0), marCard(CardDesignClover, 7), marCard(CardDesignClover, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestMarias_CpuFullMatchProgresses(t *testing.T) {
	g := newMarGame(false)
	g.Reset()
	for guard := 0; guard < 3000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case MariasPhasePlay:
			g.CpuPlay()
			if g.GetPhase() == MariasPhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == MariasPhaseTrickEnd {
					g.NextTrick()
				}
			}
		case MariasPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		default:
			g.NextTrick()
		}
	}
	if !g.GetGameEndFlag() {
		t.Error("all-CPU match did not reach game end within guard")
	}
}

func TestMarias_PlayerPlayErrors(t *testing.T) {
	g := newMarGame(true)
	g.SetPhase(MariasPhasePlay)
	g.SetCurrentPlayerIdx(0)
	marSetHand(g.GetPlayer(0), marCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(MariasPhaseRoundEnd)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestMarias_JSONRoundTrip(t *testing.T) {
	g := newMarGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Marias
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetTrumpSuit() != g.GetTrumpSuit() || g2.GetSoloistIdx() != g.GetSoloistIdx() {
		t.Error("round-trip mismatch")
	}
}

func TestMarias_UnmarshalErrors(t *testing.T) {
	var g Marias
	if err := g.UnmarshalJSON([]byte(`{"ps":[]}`)); err == nil {
		t.Error("expected invalid-player-count error")
	}
	if err := g.UnmarshalJSON([]byte(`not json`)); err == nil {
		t.Error("expected json syntax error")
	}
	if err := g.UnmarshalJSON([]byte(`{"ps":[null,null,null]}`)); err == nil {
		t.Error("expected nil-player error")
	}
}
