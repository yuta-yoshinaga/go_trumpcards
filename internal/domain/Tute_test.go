//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tuteCard(design, value int) *Card { return NewCard(design, value, false) }

func newTuteGame(human bool) *Tute {
	players := make([]*TutePlayer, TutePlayerCnt)
	players[0] = NewTutePlayer(human)
	for i := 1; i < TutePlayerCnt; i++ {
		players[i] = NewTutePlayer(false)
	}
	return NewTute(NewTrumpCardsBriscola(), players, DefaultTuteConfig())
}

func tuteSetHand(p *TutePlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestTuteConfig_Validate(t *testing.T) {
	if err := DefaultTuteConfig().Validate(); err != nil {
		t.Fatalf("default invalid: %v", err)
	}
	bad := []TuteConfig{
		{CpuDifficulty: -1, TargetPoints: 121},
		{CpuDifficulty: 9, TargetPoints: 121},
		{CpuDifficulty: 1, TargetPoints: 0},
	}
	for i, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestTute_ResetDealsAndSetsTrump(t *testing.T) {
	g := newTuteGame(true)
	g.Reset()
	if g.GetPhase() != TutePhasePlay {
		t.Fatalf("phase = %v, want Play", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != TuteHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, TuteHandSize)
		}
	}
	if g.GetTrumpSuit() < CardDesignSpade || g.GetTrumpSuit() > CardDesignMax {
		t.Errorf("trump suit not set: %d", g.GetTrumpSuit())
	}
}

func TestTute_TeamOf(t *testing.T) {
	if TuteTeamOf(0) != 0 || TuteTeamOf(2) != 0 || TuteTeamOf(1) != 1 || TuteTeamOf(3) != 1 {
		t.Error("team assignment wrong")
	}
}

func TestTute_StrengthAndPoints(t *testing.T) {
	order := []int{1, 3, 13, 12, 11, 7, 6, 5, 4, 2} // A>3>K>Q>J>7>6>5>4>2
	for i := 1; i < len(order); i++ {
		if tuteStrength(order[i-1]) <= tuteStrength(order[i]) {
			t.Errorf("strength order broken at %d", i)
		}
	}
	pts := map[int]int{1: 11, 3: 10, 13: 4, 12: 3, 11: 2, 2: 0, 7: 0}
	for v, p := range pts {
		if tuteCardPoints(v) != p {
			t.Errorf("points(%d)=%d want %d", v, tuteCardPoints(v), p)
		}
	}
	// 1 スート 30 点 × 4 = 120。
	total := 0
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 11, 12, 13} {
		total += tuteCardPoints(v) * 4
	}
	if total != 120 {
		t.Errorf("deck points = %d, want 120", total)
	}
}

func TestTute_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newTuteGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// Lead ♣ A (strongest non-trump); a ♦ (trump) overtrumps and wins.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: tuteCard(CardDesignClover, 1)},  // A♣ lead, strong fail
		{PlayerIdx: 1, Card: tuteCard(CardDesignDiamond, 2)}, // 2♦ trump (weak but trump)
		{PlayerIdx: 2, Card: tuteCard(CardDesignClover, 3)},  // 3♣ fail
		{PlayerIdx: 3, Card: tuteCard(CardDesignSpade, 1)},   // off-suit
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1 (trump beats fail lead)", w)
	}
}

func TestTute_TrickWinnerHighOfLeadSuit(t *testing.T) {
	g := newTuteGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// No trump played; highest of lead suit (♣) wins: A♣.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: tuteCard(CardDesignClover, 13)}, // K♣
		{PlayerIdx: 1, Card: tuteCard(CardDesignClover, 1)},  // A♣ (strongest)
		{PlayerIdx: 2, Card: tuteCard(CardDesignClover, 3)},  // 3♣
		{PlayerIdx: 3, Card: tuteCard(CardDesignSpade, 1)},   // off-suit ignored
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1", w)
	}
}

func TestTute_MustFollowSuit(t *testing.T) {
	g := newTuteGame(true)
	g.SetPhase(TutePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: tuteCard(CardDesignClover, 1)}})
	tuteSetHand(g.GetPlayer(0), tuteCard(CardDesignClover, 13), tuteCard(CardDesignSpade, 1))
	if err := g.PlayerPlay(1); err == nil { // spade while holding club
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // club ok
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestTute_MarriageDeclaration(t *testing.T) {
	g := newTuteGame(true)
	g.SetPhase(TutePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentTrick(nil)
	// Human (team 0) holds K+Q of clubs (non-trump → 20) and K+Q of diamonds (trump → 40).
	tuteSetHand(g.GetPlayer(0),
		tuteCard(CardDesignClover, 13), tuteCard(CardDesignClover, 12),
		tuteCard(CardDesignDiamond, 13), tuteCard(CardDesignDiamond, 12))
	if !g.CanHumanDeclareMarriage() {
		t.Fatal("human should be able to declare a marriage")
	}
	if err := g.PlayerDeclareMarriage(CardDesignClover); err != nil {
		t.Fatalf("marriage err: %v", err)
	}
	if g.GetRoundTeamPoints()[0] != TuteMarriagePlain {
		t.Errorf("plain marriage should add %d, got %v", TuteMarriagePlain, g.GetRoundTeamPoints())
	}
	// Trump marriage = 40.
	if err := g.PlayerDeclareMarriage(CardDesignDiamond); err != nil {
		t.Fatalf("trump marriage err: %v", err)
	}
	if g.GetRoundTeamPoints()[0] != TuteMarriagePlain+TuteMarriageTrump {
		t.Errorf("trump marriage should add %d, got %v", TuteMarriageTrump, g.GetRoundTeamPoints())
	}
	// Re-declaring the same suit fails.
	if err := g.PlayerDeclareMarriage(CardDesignClover); err == nil {
		t.Error("re-declaring a suit should fail")
	}
	// Cannot declare while not leading (mid-trick).
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: tuteCard(CardDesignSpade, 1)}})
	if g.canDeclareMarriage(0, CardDesignHeart) {
		t.Error("cannot declare mid-trick")
	}
}

func TestTute_GetHumanDeclarableMarriageSuits(t *testing.T) {
	g := newTuteGame(true)
	g.SetPhase(TutePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentTrick(nil)
	// Human holds K+Q of clubs and diamonds (both declarable); hearts is only a K.
	tuteSetHand(g.GetPlayer(0),
		tuteCard(CardDesignClover, 13), tuteCard(CardDesignClover, 12),
		tuteCard(CardDesignDiamond, 13), tuteCard(CardDesignDiamond, 12),
		tuteCard(CardDesignHeart, 13))
	got := g.GetHumanDeclarableMarriageSuits()
	assert.ElementsMatch(t, []int{CardDesignClover, CardDesignDiamond}, got)

	// After declaring clubs, only diamonds remains.
	if err := g.PlayerDeclareMarriage(CardDesignClover); err != nil {
		t.Fatalf("marriage err: %v", err)
	}
	assert.Equal(t, []int{CardDesignDiamond}, g.GetHumanDeclarableMarriageSuits())

	// Mid-trick (not leading) → nothing is declarable.
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: tuteCard(CardDesignSpade, 1)}})
	assert.Empty(t, g.GetHumanDeclarableMarriageSuits())
}

func TestTute_TuteInstantWin(t *testing.T) {
	g := newTuteGame(true)
	g.SetPhase(TutePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	// Human holds all four Kings → Tute → instant win for team 0.
	tuteSetHand(g.GetPlayer(0),
		tuteCard(CardDesignSpade, 13), tuteCard(CardDesignClover, 13),
		tuteCard(CardDesignHeart, 13), tuteCard(CardDesignDiamond, 13))
	if !g.CanHumanDeclareTute() {
		t.Fatal("human with 4 kings should be able to declare Tute")
	}
	if err := g.PlayerDeclareTute(); err != nil {
		t.Fatalf("tute err: %v", err)
	}
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("Tute should end the game with team 0 winner")
	}
}

func TestTute_ResolveTrickAddsPointsAndLastBonus(t *testing.T) {
	g := newTuteGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(TutePhaseTrickEnd)
	g.SetTrickNumber(1)
	// A♣(11) + 3♣(10) + K♣(4) + 2♣(0) = 25 pts, A♣ wins for team 0 (player 0).
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: tuteCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: tuteCard(CardDesignClover, 3)},
		{PlayerIdx: 2, Card: tuteCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: tuteCard(CardDesignClover, 2)},
	})
	g.ResolveTrick()
	if g.GetRoundTeamPoints()[0] != 25 {
		t.Errorf("team0 should have 25 pts, got %v", g.GetRoundTeamPoints())
	}
	if g.GetLeadPlayerIdx() != 0 {
		t.Error("lead should follow winner")
	}
	// Last trick gives +10.
	g.SetPhase(TutePhaseTrickEnd)
	g.SetTrickNumber(TuteTrickCount)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: tuteCard(CardDesignClover, 11)}, // J♣ 2pts wins
		{PlayerIdx: 1, Card: tuteCard(CardDesignClover, 4)},
		{PlayerIdx: 2, Card: tuteCard(CardDesignClover, 5)},
		{PlayerIdx: 3, Card: tuteCard(CardDesignClover, 6)},
	})
	before := g.GetRoundTeamPoints()[0]
	g.ResolveTrick()
	if g.GetRoundTeamPoints()[0] != before+2+TuteLastTrickBonus {
		t.Errorf("last trick should add 2+%d, got %v (before %d)", TuteLastTrickBonus, g.GetRoundTeamPoints(), before)
	}
	if g.GetPhase() != TutePhaseRoundEnd {
		t.Errorf("phase = %v, want RoundEnd after last trick", g.GetPhase())
	}
}

func TestTute_ScoreRoundAndGameEnd(t *testing.T) {
	g := newTuteGame(false)
	g.SetPhase(TutePhaseRoundEnd)
	cfg := g.GetConfig()
	cfg.TargetPoints = 50
	g.SetConfig(cfg)
	g.roundTeamPts = [TuteTeamCnt]int{60, 20}
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("team0 should win: end=%v winner=%d scores=%v", g.GetGameEndFlag(), g.GetWinnerTeam(), g.GetTeamScores())
	}
}

func TestTute_NextRound(t *testing.T) {
	g := newTuteGame(false)
	g.Reset()
	g.SetPhase(TutePhaseRoundEnd)
	d := g.GetDealerIdx()
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetDealerIdx() != (d+1)%TutePlayerCnt {
		t.Errorf("next round/dealer wrong")
	}
	if g.GetPhase() != TutePhasePlay {
		t.Errorf("phase = %v, want Play", g.GetPhase())
	}
}

func TestTute_CpuFullRoundProgresses(t *testing.T) {
	g := newTuteGame(false)
	g.Reset()
	for steps := 0; steps < 400; steps++ {
		switch g.GetPhase() {
		case TutePhasePlay:
			g.CpuPlay()
		case TutePhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == TutePhaseTrickEnd {
				g.NextTrick()
			}
		case TutePhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case TutePhaseGameEnd:
			if g.GetWinnerTeam() < 0 {
				t.Error("game ended without winner")
			}
			return
		}
		if g.GetRoundNumber() > 100 {
			break
		}
	}
}

func TestTute_HintAndPlayable(t *testing.T) {
	g := newTuteGame(true)
	g.SetPhase(TutePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	tuteSetHand(g.GetPlayer(0), tuteCard(CardDesignClover, 7), tuteCard(CardDesignSpade, 2))
	if h := g.GetHint(); h == nil || len(h.CardIndices) != 1 {
		t.Errorf("play hint missing: %+v", h)
	}
	if idx := g.GetPlayableIndices(0); len(idx) != 2 {
		t.Errorf("playable = %v, want 2 (lead)", idx)
	}
	g.SetPhase(TutePhaseRoundEnd)
	if idx := g.GetPlayableIndices(0); idx != nil {
		t.Error("playable should be nil outside play phase")
	}
	// Marriage hint when holding K+Q.
	g.SetPhase(TutePhasePlay)
	tuteSetHand(g.GetPlayer(0), tuteCard(CardDesignClover, 13), tuteCard(CardDesignClover, 12))
	if h := g.GetHint(); h == nil || h.Marriage != CardDesignClover {
		t.Errorf("marriage hint missing: %+v", h)
	}
}

func TestTute_PlayerPlayErrors(t *testing.T) {
	g := newTuteGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(1) // CPU
	if err := g.PlayerPlay(0); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(TutePhaseRoundEnd)
	if err := g.PlayerPlay(0); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
	g.gameEndFlag = true
	if err := g.PlayerPlay(0); !errors.Is(err, ErrGameEnded) {
		t.Errorf("err = %v, want ErrGameEnded", err)
	}
}

func TestTute_JSONRoundTrip(t *testing.T) {
	g := newTuteGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Tute
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPhase() != g.GetPhase() || g2.GetPlayerCnt() != TutePlayerCnt || g2.GetTrumpSuit() != g.GetTrumpSuit() {
		t.Error("round trip mismatch")
	}
}

func TestTute_UnmarshalOversized(t *testing.T) {
	var g Tute
	bad := `{"al":[`
	for i := 0; i < tuteMaxSliceLen+1; i++ {
		if i > 0 {
			bad += ","
		}
		bad += `{"TurnNumber":0,"PlayerIdx":-1,"ActionType":"x","Detail":"y","Cards":null}`
	}
	bad += `]}`
	if err := json.Unmarshal([]byte(bad), &g); !errors.Is(err, errTuteOversized) {
		t.Errorf("err = %v, want errTuteOversized", err)
	}
}

func TestTute_UnmarshalInvalidPlayers(t *testing.T) {
	var g Tute
	if err := json.Unmarshal([]byte(`{"ps":[]}`), &g); !errors.Is(err, errTuteInvalidPlayers) {
		t.Errorf("err = %v, want errTuteInvalidPlayers", err)
	}
}
