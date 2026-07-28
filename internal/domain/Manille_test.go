//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func manCard(design, value int) *Card { return NewCard(design, value, false) }

func newManGame(human bool) *Manille {
	players := make([]*ManillePlayer, ManillePlayerCnt)
	players[0] = NewManillePlayer(human)
	for i := 1; i < ManillePlayerCnt; i++ {
		players[i] = NewManillePlayer(false)
	}
	return NewManille(NewTrumpCardsBelote(), players, DefaultManilleConfig())
}

func manSetHand(p *ManillePlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestManilleConfig_Validate(t *testing.T) {
	if err := DefaultManilleConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (ManilleConfig{CpuDifficulty: 9, TargetPoints: 101}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (ManilleConfig{CpuDifficulty: ManilleCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestManille_ResetDealsAndSetsTrump(t *testing.T) {
	g := newManGame(true)
	g.Reset()
	if g.GetPhase() != ManillePhasePlay {
		t.Errorf("phase = %d, want Play", g.GetPhase())
	}
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
	}
	if total != ManillePlayerCnt*ManilleHandSize {
		t.Errorf("dealt %d cards, want %d", total, ManillePlayerCnt*ManilleHandSize)
	}
	if g.GetTrumpSuit() < 1 || g.GetTrumpSuit() > 4 {
		t.Errorf("trump suit = %d, want 1-4", g.GetTrumpSuit())
	}
}

func TestManille_StrengthAndPoints(t *testing.T) {
	// 10 > A > K > Q > J > 9 > 8 > 7
	order := []int{10, 1, 13, 12, 11, 9, 8, 7}
	for i := 1; i < len(order); i++ {
		if manilleStrength(order[i-1]) <= manilleStrength(order[i]) {
			t.Errorf("strength(%d) should beat strength(%d)", order[i-1], order[i])
		}
	}
	cases := map[int]int{10: 5, 1: 4, 13: 3, 12: 2, 11: 1, 9: 0, 8: 0, 7: 0}
	for v, want := range cases {
		if got := manilleCardPoints(manCard(CardDesignSpade, v)); got != want {
			t.Errorf("points(%d) = %d, want %d", v, got, want)
		}
	}
}

func TestManille_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newManGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// Lead 10♣ (Manille, strongest plain) but a low trump 7♦ beats it.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: manCard(CardDesignClover, 10)},
		{PlayerIdx: 1, Card: manCard(CardDesignDiamond, 7)},
		{PlayerIdx: 2, Card: manCard(CardDesignClover, 1)},
		{PlayerIdx: 3, Card: manCard(CardDesignClover, 13)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("trick winner = %d, want 1 (low trump beats high plain)", w)
	}
}

func TestManille_TenBeatsAceInPlainSuit(t *testing.T) {
	g := newManGame(false)
	g.SetTrumpSuit(CardDesignDiamond) // no trump in trick
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: manCard(CardDesignClover, 1)},  // Ace lead
		{PlayerIdx: 1, Card: manCard(CardDesignClover, 10)}, // Manille beats Ace
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("trick winner = %d, want 1 (10 outranks A)", w)
	}
}

func TestManille_MustFollow(t *testing.T) {
	g := newManGame(true)
	g.SetPhase(ManillePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: manCard(CardDesignClover, 1)}})
	manSetHand(g.GetPlayer(0), manCard(CardDesignClover, 13), manCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(1); err == nil { // diamond while holding club
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // K♣ follows
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestManille_PartnerWinningTrumpExemption(t *testing.T) {
	g := newManGame(true)
	g.SetPhase(ManillePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0) // human plays last in this trick
	// Trick led by p1; the human's partner p2 (team 0) is currently winning with 10♣.
	// The human is void in clubs but holds a trump — partner-winning exemption lets them discard.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 1, Card: manCard(CardDesignClover, 7)},
		{PlayerIdx: 2, Card: manCard(CardDesignClover, 10)},
		{PlayerIdx: 3, Card: manCard(CardDesignClover, 8)},
	})
	manSetHand(g.GetPlayer(0), manCard(CardDesignHeart, 7), manCard(CardDesignDiamond, 8))
	if err := g.PlayerPlay(0); err != nil { // discard heart while partner winning — allowed
		t.Fatalf("partner-winning exemption should allow discard, got: %v", err)
	}
}

func TestManille_MustTrumpWhenOpponentWinning(t *testing.T) {
	g := newManGame(true)
	g.SetPhase(ManillePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0) // human plays last
	// Opponent p1 (team 1) is winning with 10♣. Human is void in clubs but holds a trump → must trump.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 1, Card: manCard(CardDesignClover, 10)},
		{PlayerIdx: 2, Card: manCard(CardDesignClover, 7)},
		{PlayerIdx: 3, Card: manCard(CardDesignClover, 8)},
	})
	manSetHand(g.GetPlayer(0), manCard(CardDesignHeart, 7), manCard(CardDesignDiamond, 8))
	if err := g.PlayerPlay(0); err == nil { // heart discard while opponent winning and holding trump
		t.Error("expected must-trump error when opponent is winning")
	}
	if err := g.PlayerPlay(1); err != nil { // 8♦ trump
		t.Fatalf("valid trump play err: %v", err)
	}
}

func TestManille_ResolveTrickPoints(t *testing.T) {
	g := newManGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(ManillePhaseTrickEnd)
	g.SetTrickNumber(1)
	// 10♣(5) + A♣(4) + K♣(3) + 7♣(0) = 12; 10♣ wins for team 0 (p0).
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: manCard(CardDesignClover, 10)},
		{PlayerIdx: 1, Card: manCard(CardDesignClover, 1)},
		{PlayerIdx: 2, Card: manCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: manCard(CardDesignClover, 7)},
	})
	g.ResolveTrick()
	pts := g.GetRoundCardPoints()
	if pts[0] != 12 {
		t.Errorf("team 0 round points = %d, want 12", pts[0])
	}
	if g.GetPhase() != ManillePhaseTrickEnd {
		t.Errorf("phase after non-final trick = %d, want TrickEnd", g.GetPhase())
	}
}

func TestManille_ScoreRoundAddsAndEnds(t *testing.T) {
	g := newManGame(false)
	g.SetPhase(ManillePhaseRoundEnd)
	g.SetRoundCardPoints([ManilleTeamCnt]int{40, 20})
	g.SetTeamScores([ManilleTeamCnt]int{70, 30})
	g.ScoreRound()
	sc := g.GetTeamScores()
	if sc[0] != 110 || sc[1] != 50 {
		t.Errorf("team scores = %v, want [110 50]", sc)
	}
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("expected game end with team 0 winning (>=101)")
	}
}

func TestManille_NextRound(t *testing.T) {
	g := newManGame(false)
	g.SetPhase(ManillePhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("round number = %d, want 2", g.GetRoundNumber())
	}
	if g.GetPhase() != ManillePhasePlay {
		t.Errorf("phase after NextRound = %d, want Play", g.GetPhase())
	}
}

func TestManille_HintAndPlayable(t *testing.T) {
	g := newManGame(true)
	g.SetPhase(ManillePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	manSetHand(g.GetPlayer(0), manCard(CardDesignClover, 7), manCard(CardDesignClover, 10))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2 leading options", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestManille_CpuFullMatchProgresses(t *testing.T) {
	g := newManGame(false) // all CPU
	g.Reset()
	for guard := 0; guard < 2000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case ManillePhasePlay:
			g.CpuPlay()
			if g.GetPhase() == ManillePhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == ManillePhaseTrickEnd {
					g.NextTrick()
				}
			}
		case ManillePhaseRoundEnd:
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

func TestManille_PlayerPlayErrors(t *testing.T) {
	g := newManGame(true)
	g.SetPhase(ManillePhasePlay)
	g.SetCurrentPlayerIdx(0)
	manSetHand(g.GetPlayer(0), manCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(ManillePhaseRoundEnd)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestManille_JSONRoundTrip(t *testing.T) {
	g := newManGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Manille
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetTrumpSuit() != g.GetTrumpSuit() || g2.GetPlayerCnt() != g.GetPlayerCnt() {
		t.Error("round-trip mismatch")
	}
}

func TestManille_UnmarshalErrors(t *testing.T) {
	var g Manille
	if err := g.UnmarshalJSON([]byte(`{"ps":[]}`)); err == nil {
		t.Error("expected invalid-player-count error")
	}
	if err := g.UnmarshalJSON([]byte(`not json`)); err == nil {
		t.Error("expected json syntax error")
	}
	// A 4-element players array with a nil entry is rejected.
	if err := g.UnmarshalJSON([]byte(`{"ps":[null,null,null,null]}`)); err == nil {
		t.Error("expected nil-player error")
	}
}
