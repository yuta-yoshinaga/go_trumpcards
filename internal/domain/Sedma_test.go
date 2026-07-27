//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func sedCard(design, value int) *Card { return NewCard(design, value, false) }

func newSedGame(human bool) *Sedma {
	players := make([]*SedmaPlayer, SedmaPlayerCnt)
	players[0] = NewSedmaPlayer(human)
	for i := 1; i < SedmaPlayerCnt; i++ {
		players[i] = NewSedmaPlayer(false)
	}
	return NewSedma(NewTrumpCardsBelote(), players, DefaultSedmaConfig())
}

func sedSetHand(p *SedmaPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestSedmaConfig_Validate(t *testing.T) {
	if err := DefaultSedmaConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (SedmaConfig{CpuDifficulty: 9, TargetPoints: 101}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (SedmaConfig{CpuDifficulty: SedmaCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestSedma_ResetDeals(t *testing.T) {
	g := newSedGame(true)
	g.Reset()
	if g.GetPhase() != SedmaPhasePlay {
		t.Errorf("phase = %d, want Play", g.GetPhase())
	}
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
	}
	if total != SedmaPlayerCnt*SedmaHandSize {
		t.Errorf("dealt %d cards, want %d", total, SedmaPlayerCnt*SedmaHandSize)
	}
}

func TestSedma_Points(t *testing.T) {
	cases := map[int]int{1: 10, 10: 10, 13: 0, 12: 0, 11: 0, 9: 0, 8: 0, 7: 0}
	total := 0
	for v, want := range cases {
		if got := sedmaCardPoints(sedCard(CardDesignSpade, v)); got != want {
			t.Errorf("points(%d) = %d, want %d", v, got, want)
		}
		total += want
	}
	if total*4 != 80 {
		t.Errorf("deck card points = %d, want 80", total*4)
	}
}

func TestSedma_TrickWinnerSameRankCaptures(t *testing.T) {
	g := newSedGame(false)
	// Lead K♣; p2 plays another K (same rank) -> captures; p3 discards.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: sedCard(CardDesignClover, 13)},
		{PlayerIdx: 1, Card: sedCard(CardDesignHeart, 9)},
		{PlayerIdx: 2, Card: sedCard(CardDesignSpade, 13)},
		{PlayerIdx: 3, Card: sedCard(CardDesignDiamond, 8)},
	})
	if w := g.trickWinner(); w != 2 {
		t.Errorf("trick winner = %d, want 2 (same-rank capture)", w)
	}
}

func TestSedma_TrickWinnerSevenCaptures(t *testing.T) {
	g := newSedGame(false)
	// Lead A♣; p3 plays a 7 (wild) -> captures even though no one matched rank.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: sedCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: sedCard(CardDesignHeart, 9)},
		{PlayerIdx: 2, Card: sedCard(CardDesignSpade, 8)},
		{PlayerIdx: 3, Card: sedCard(CardDesignDiamond, 7)},
	})
	if w := g.trickWinner(); w != 3 {
		t.Errorf("trick winner = %d, want 3 (7 wild capture)", w)
	}
}

func TestSedma_TrickWinnerNoCaptureLeadWins(t *testing.T) {
	g := newSedGame(false)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: sedCard(CardDesignClover, 13)},
		{PlayerIdx: 1, Card: sedCard(CardDesignHeart, 9)},
		{PlayerIdx: 2, Card: sedCard(CardDesignSpade, 8)},
		{PlayerIdx: 3, Card: sedCard(CardDesignDiamond, 10)},
	})
	if w := g.trickWinner(); w != 0 {
		t.Errorf("trick winner = %d, want 0 (no capture, lead wins)", w)
	}
}

func TestSedma_TrickWinnerLastCaptureWins(t *testing.T) {
	g := newSedGame(false)
	// Lead K; p1 captures with K; p3 captures again with a 7 -> last capturer wins.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: sedCard(CardDesignClover, 13)},
		{PlayerIdx: 1, Card: sedCard(CardDesignHeart, 13)},
		{PlayerIdx: 2, Card: sedCard(CardDesignSpade, 8)},
		{PlayerIdx: 3, Card: sedCard(CardDesignDiamond, 7)},
	})
	if w := g.trickWinner(); w != 3 {
		t.Errorf("trick winner = %d, want 3 (last capture wins)", w)
	}
}

func TestSedma_TrickWinnerSevenLeadOnlySevenCaptures(t *testing.T) {
	// Lead is itself a 7: only another 7 captures it (a King does not).
	g := newSedGame(false)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: sedCard(CardDesignClover, 7)},
		{PlayerIdx: 1, Card: sedCard(CardDesignHeart, 13)},
		{PlayerIdx: 2, Card: sedCard(CardDesignSpade, 7)},
		{PlayerIdx: 3, Card: sedCard(CardDesignDiamond, 8)},
	})
	if w := g.trickWinner(); w != 2 {
		t.Errorf("trick winner = %d, want 2 (only a 7 captures a 7 lead)", w)
	}
}

func TestSedma_TrickWinnerSevenLeadNoCapture(t *testing.T) {
	// Lead 7, nobody else plays a 7 -> the lead player keeps the trick.
	g := newSedGame(false)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: sedCard(CardDesignClover, 7)},
		{PlayerIdx: 1, Card: sedCard(CardDesignHeart, 13)},
		{PlayerIdx: 2, Card: sedCard(CardDesignSpade, 1)},
		{PlayerIdx: 3, Card: sedCard(CardDesignDiamond, 8)},
	})
	if w := g.trickWinner(); w != 0 {
		t.Errorf("trick winner = %d, want 0 (7 lead, no other 7)", w)
	}
}

func TestSedma_AnyCardIsPlayable(t *testing.T) {
	g := newSedGame(true)
	g.SetPhase(SedmaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: sedCard(CardDesignClover, 13)}})
	sedSetHand(g.GetPlayer(0), sedCard(CardDesignHeart, 9), sedCard(CardDesignDiamond, 8), sedCard(CardDesignSpade, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 3 {
		t.Errorf("playable = %v, want all 3 (no follow obligation)", idxs)
	}
	if err := g.PlayerPlay(0); err != nil { // off-rank discard is legal
		t.Fatalf("any card should be legal, got: %v", err)
	}
}

func TestSedma_ResolveTrickPointsAndLastBonus(t *testing.T) {
	g := newSedGame(false)
	g.SetPhase(SedmaPhaseTrickEnd)
	g.SetTrickNumber(SedmaTrickCount) // final trick -> +10 bonus
	// A♣(10) + 10♥(10) + K♠(0) + 7♦(0); 7♦ captures for player 3 (team 1). pts 20 + 10 last.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: sedCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: sedCard(CardDesignHeart, 10)},
		{PlayerIdx: 2, Card: sedCard(CardDesignSpade, 13)},
		{PlayerIdx: 3, Card: sedCard(CardDesignDiamond, 7)},
	})
	g.ResolveTrick()
	pts := g.GetRoundCardPoints()
	if pts[1] != 30 {
		t.Errorf("team 1 round points = %d, want 30 (20 + 10 last)", pts[1])
	}
	if g.GetPhase() != SedmaPhaseRoundEnd {
		t.Errorf("phase after final trick = %d, want RoundEnd", g.GetPhase())
	}
}

func TestSedma_ScoreRoundAddsAndEnds(t *testing.T) {
	g := newSedGame(false)
	g.SetPhase(SedmaPhaseRoundEnd)
	g.SetRoundCardPoints([SedmaTeamCnt]int{60, 30})
	g.SetTeamScores([SedmaTeamCnt]int{50, 20})
	g.ScoreRound()
	sc := g.GetTeamScores()
	if sc[0] != 110 || sc[1] != 50 {
		t.Errorf("team scores = %v, want [110 50]", sc)
	}
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Error("expected match end with team 0 winning")
	}
}

func TestSedma_NextRound(t *testing.T) {
	g := newSedGame(false)
	g.SetPhase(SedmaPhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("round number = %d, want 2", g.GetRoundNumber())
	}
	if g.GetPhase() != SedmaPhasePlay {
		t.Errorf("phase after NextRound = %d, want Play", g.GetPhase())
	}
}

func TestSedma_HintAndPlayable(t *testing.T) {
	g := newSedGame(true)
	g.SetPhase(SedmaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	sedSetHand(g.GetPlayer(0), sedCard(CardDesignClover, 8), sedCard(CardDesignClover, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestSedma_CpuFullMatchProgresses(t *testing.T) {
	g := newSedGame(false)
	g.Reset()
	for guard := 0; guard < 3000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case SedmaPhasePlay:
			g.CpuPlay()
			if g.GetPhase() == SedmaPhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == SedmaPhaseTrickEnd {
					g.NextTrick()
				}
			}
		case SedmaPhaseRoundEnd:
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

func TestSedma_PlayerPlayErrors(t *testing.T) {
	g := newSedGame(true)
	g.SetPhase(SedmaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	sedSetHand(g.GetPlayer(0), sedCard(CardDesignClover, 8))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(SedmaPhaseRoundEnd)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestSedma_JSONRoundTrip(t *testing.T) {
	g := newSedGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Sedma
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() || g2.GetRoundNumber() != g.GetRoundNumber() {
		t.Error("round-trip mismatch")
	}
}

func TestSedma_UnmarshalErrors(t *testing.T) {
	var g Sedma
	if err := g.UnmarshalJSON([]byte(`{"ps":[]}`)); err == nil {
		t.Error("expected invalid-player-count error")
	}
	if err := g.UnmarshalJSON([]byte(`not json`)); err == nil {
		t.Error("expected json syntax error")
	}
	if err := g.UnmarshalJSON([]byte(`{"ps":[null,null,null,null]}`)); err == nil {
		t.Error("expected nil-player error")
	}
}
