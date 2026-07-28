//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func suecaCard(design, value int) *Card { return NewCard(design, value, false) }

func newSuecaGame(human bool) *Sueca {
	players := make([]*SuecaPlayer, SuecaPlayerCnt)
	players[0] = NewSuecaPlayer(human)
	for i := 1; i < SuecaPlayerCnt; i++ {
		players[i] = NewSuecaPlayer(false)
	}
	return NewSueca(NewTrumpCardsBriscola(), players, DefaultSuecaConfig())
}

func suecaSetHand(p *SuecaPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestSuecaConfig_Validate(t *testing.T) {
	if err := DefaultSuecaConfig().Validate(); err != nil {
		t.Fatalf("default invalid: %v", err)
	}
	bad := []SuecaConfig{
		{CpuDifficulty: -1, TargetGamePoints: 4},
		{CpuDifficulty: 9, TargetGamePoints: 4},
		{CpuDifficulty: 1, TargetGamePoints: 0},
	}
	for i, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestSueca_ResetDealsAndSetsTrump(t *testing.T) {
	g := newSuecaGame(true)
	g.Reset()
	if g.GetPhase() != SuecaPhasePlay {
		t.Fatalf("phase = %v, want Play", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != SuecaHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, SuecaHandSize)
		}
	}
	if g.GetTrumpSuit() < CardDesignSpade || g.GetTrumpSuit() > CardDesignMax {
		t.Errorf("trump suit not set: %d", g.GetTrumpSuit())
	}
}

func TestSueca_StrengthAndPoints(t *testing.T) {
	order := []int{1, 7, 13, 11, 12, 6, 5, 4, 3, 2} // A>7>K>J>Q>6>5>4>3>2
	for i := 1; i < len(order); i++ {
		if suecaStrength(order[i-1]) <= suecaStrength(order[i]) {
			t.Errorf("strength order broken at %d", i)
		}
	}
	pts := map[int]int{1: 11, 7: 10, 13: 4, 11: 3, 12: 2, 2: 0, 6: 0}
	for v, p := range pts {
		if suecaCardPoints(v) != p {
			t.Errorf("points(%d)=%d want %d", v, suecaCardPoints(v), p)
		}
	}
	total := 0
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 11, 12, 13} {
		total += suecaCardPoints(v) * 4
	}
	if total != 120 {
		t.Errorf("deck points = %d, want 120", total)
	}
}

func TestSueca_GamePoints(t *testing.T) {
	if suecaGamePoints(61) != 1 || suecaGamePoints(90) != 1 {
		t.Error("61-90 should be 1")
	}
	if suecaGamePoints(91) != 2 || suecaGamePoints(119) != 2 {
		t.Error("91-119 should be 2")
	}
	if suecaGamePoints(120) != 4 {
		t.Error("120 should be 4")
	}
}

func TestSueca_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newSuecaGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: suecaCard(CardDesignClover, 1)},  // A♣ lead (strong fail)
		{PlayerIdx: 1, Card: suecaCard(CardDesignDiamond, 2)}, // 2♦ trump
		{PlayerIdx: 2, Card: suecaCard(CardDesignClover, 7)},  // 7♣ fail
		{PlayerIdx: 3, Card: suecaCard(CardDesignSpade, 1)},   // off-suit
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1 (trump beats fail lead)", w)
	}
}

func TestSueca_TrickWinnerHighOfLeadSuit(t *testing.T) {
	g := newSuecaGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// No trump; A♣ beats 7♣ beats K♣ (A>7>K).
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: suecaCard(CardDesignClover, 13)}, // K♣
		{PlayerIdx: 1, Card: suecaCard(CardDesignClover, 7)},  // 7♣
		{PlayerIdx: 2, Card: suecaCard(CardDesignClover, 1)},  // A♣ (strongest)
		{PlayerIdx: 3, Card: suecaCard(CardDesignSpade, 1)},   // off-suit ignored
	})
	if w := g.trickWinner(); w != 2 {
		t.Errorf("winner = %d, want 2 (A♣)", w)
	}
}

func TestSueca_MustFollowSuit(t *testing.T) {
	g := newSuecaGame(true)
	g.SetPhase(SuecaPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: suecaCard(CardDesignClover, 1)}})
	suecaSetHand(g.GetPlayer(0), suecaCard(CardDesignClover, 13), suecaCard(CardDesignSpade, 1))
	if err := g.PlayerPlay(1); err == nil { // spade while holding club
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // club ok
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestSueca_ResolveTrickAddsPoints(t *testing.T) {
	g := newSuecaGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(SuecaPhaseTrickEnd)
	g.SetTrickNumber(1)
	// A♣(11)+7♣(10)+K♣(4)+2♣(0)=25, A♣ wins for team 0.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: suecaCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: suecaCard(CardDesignClover, 7)},
		{PlayerIdx: 2, Card: suecaCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: suecaCard(CardDesignClover, 2)},
	})
	g.ResolveTrick()
	if g.GetRoundCardPoints()[0] != 25 {
		t.Errorf("team0 should have 25 pts, got %v", g.GetRoundCardPoints())
	}
	if g.GetLeadPlayerIdx() != 0 {
		t.Error("lead should follow winner")
	}
}

func TestSueca_ScoreRoundGamePointsAndMatchEnd(t *testing.T) {
	// Team 0 takes 95 → 2 game points; target 2 → match ends.
	g := newSuecaGame(false)
	g.SetPhase(SuecaPhaseRoundEnd)
	cfg := g.GetConfig()
	cfg.TargetGamePoints = 2
	g.SetConfig(cfg)
	g.SetRoundCardPoints([SuecaTeamCnt]int{95, 25})
	g.ScoreRound()
	if g.GetRoundWinnerTeam() != 0 || g.GetRoundGamePoints() != 2 {
		t.Errorf("team0 should win 2 game pts, got winner=%d pts=%d", g.GetRoundWinnerTeam(), g.GetRoundGamePoints())
	}
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("match should end with team0 winner")
	}
}

func TestSueca_ScoreRoundDraw(t *testing.T) {
	g := newSuecaGame(false)
	g.SetPhase(SuecaPhaseRoundEnd)
	g.SetRoundCardPoints([SuecaTeamCnt]int{60, 60})
	g.ScoreRound()
	if g.GetRoundWinnerTeam() != -1 {
		t.Errorf("60-60 should be a draw, got winner %d", g.GetRoundWinnerTeam())
	}
	if g.GetTeamGamePoints() != [SuecaTeamCnt]int{0, 0} {
		t.Errorf("draw should award no game pts, got %v", g.GetTeamGamePoints())
	}
	if g.GetPhase() != SuecaPhaseRoundEnd {
		t.Error("draw should not end the match")
	}
}

func TestSueca_NextRound(t *testing.T) {
	g := newSuecaGame(false)
	g.Reset()
	g.SetPhase(SuecaPhaseRoundEnd)
	d := g.GetDealerIdx()
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetDealerIdx() != (d+1)%SuecaPlayerCnt {
		t.Errorf("next round/dealer wrong")
	}
	if g.GetPhase() != SuecaPhasePlay {
		t.Errorf("phase = %v, want Play", g.GetPhase())
	}
}

func TestSueca_CpuFullMatchProgresses(t *testing.T) {
	g := newSuecaGame(false)
	g.Reset()
	for steps := 0; steps < 600; steps++ {
		switch g.GetPhase() {
		case SuecaPhasePlay:
			g.CpuPlay()
		case SuecaPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == SuecaPhaseTrickEnd {
				g.NextTrick()
			}
		case SuecaPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case SuecaPhaseGameEnd:
			if g.GetWinnerTeam() < 0 {
				t.Error("match ended without winner")
			}
			return
		}
		if g.GetRoundNumber() > 100 {
			break
		}
	}
}

func TestSueca_HintAndPlayable(t *testing.T) {
	g := newSuecaGame(true)
	g.SetPhase(SuecaPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	suecaSetHand(g.GetPlayer(0), suecaCard(CardDesignClover, 2), suecaCard(CardDesignSpade, 3))
	if h := g.GetHint(); h == nil || len(h.CardIndices) != 1 {
		t.Errorf("play hint missing: %+v", h)
	}
	if idx := g.GetPlayableIndices(0); len(idx) != 2 {
		t.Errorf("playable = %v, want 2 (lead)", idx)
	}
	g.SetPhase(SuecaPhaseRoundEnd)
	if idx := g.GetPlayableIndices(0); idx != nil {
		t.Error("playable should be nil outside play phase")
	}
}

func TestSueca_PlayerPlayErrors(t *testing.T) {
	g := newSuecaGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(1) // CPU
	if err := g.PlayerPlay(0); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(SuecaPhaseRoundEnd)
	if err := g.PlayerPlay(0); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
	g.gameEndFlag = true
	if err := g.PlayerPlay(0); !errors.Is(err, ErrGameEnded) {
		t.Errorf("err = %v, want ErrGameEnded", err)
	}
}

func TestSueca_JSONRoundTrip(t *testing.T) {
	g := newSuecaGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Sueca
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPhase() != g.GetPhase() || g2.GetPlayerCnt() != SuecaPlayerCnt || g2.GetTrumpSuit() != g.GetTrumpSuit() {
		t.Error("round trip mismatch")
	}
}

func TestSueca_UnmarshalOversized(t *testing.T) {
	var g Sueca
	bad := `{"al":[`
	for i := 0; i < suecaMaxSliceLen+1; i++ {
		if i > 0 {
			bad += ","
		}
		bad += `{"TurnNumber":0,"PlayerIdx":-1,"ActionType":"x","Detail":"y","Cards":null}`
	}
	bad += `]}`
	if err := json.Unmarshal([]byte(bad), &g); !errors.Is(err, errSuecaOversized) {
		t.Errorf("err = %v, want errSuecaOversized", err)
	}
}

func TestSueca_UnmarshalInvalidPlayers(t *testing.T) {
	var g Sueca
	if err := json.Unmarshal([]byte(`{"ps":[]}`), &g); !errors.Is(err, errSuecaInvalidPlayers) {
		t.Errorf("err = %v, want errSuecaInvalidPlayers", err)
	}
}
