//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func koCard(design, value int) *Card { return NewCard(design, value, false) }

func newKoGame(human bool) *KnockoutWhist {
	players := make([]*KnockoutWhistPlayer, KnockoutWhistPlayerCnt)
	players[0] = NewKnockoutWhistPlayer(human)
	for i := 1; i < KnockoutWhistPlayerCnt; i++ {
		players[i] = NewKnockoutWhistPlayer(false)
	}
	return NewKnockoutWhist(NewTrumpCards(0), players, DefaultKnockoutWhistConfig())
}

func koSetHand(p *KnockoutWhistPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestKnockoutWhistConfig_Validate(t *testing.T) {
	if err := DefaultKnockoutWhistConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (KnockoutWhistConfig{CpuDifficulty: 9}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
}

func TestKnockoutWhist_ResetDealsSevenAndArmsDogbones(t *testing.T) {
	g := newKoGame(true)
	g.Reset()
	if g.GetPhase() != KnockoutWhistPhasePlay {
		t.Errorf("phase = %d, want Play", g.GetPhase())
	}
	if g.GetHandSize() != 7 {
		t.Errorf("round-1 hand size = %d, want 7", g.GetHandSize())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		if p.GetCardsSize() != 7 {
			t.Errorf("player %d dealt %d cards, want 7", i, p.GetCardsSize())
		}
		if p.GetDogbones() != KnockoutWhistStartingDogbones {
			t.Errorf("player %d dogbones = %d, want %d", i, p.GetDogbones(), KnockoutWhistStartingDogbones)
		}
		if p.GetEliminated() {
			t.Errorf("player %d eliminated at start", i)
		}
	}
	if g.GetActiveCount() != KnockoutWhistPlayerCnt {
		t.Errorf("active = %d, want %d", g.GetActiveCount(), KnockoutWhistPlayerCnt)
	}
}

func TestKnockoutWhist_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newKoGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: koCard(CardDesignClover, 1)},  // Ace lead
		{PlayerIdx: 1, Card: koCard(CardDesignDiamond, 2)}, // low trump beats it
		{PlayerIdx: 2, Card: koCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: koCard(CardDesignClover, 12)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("trick winner = %d, want 1 (trump beats high plain)", w)
	}
}

func TestKnockoutWhist_MustFollow(t *testing.T) {
	g := newKoGame(true)
	g.SetPhase(KnockoutWhistPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: koCard(CardDesignClover, 1)}})
	koSetHand(g.GetPlayer(0), koCard(CardDesignClover, 13), koCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(1); err == nil { // diamond while holding club
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // K♣ follows
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestKnockoutWhist_ResolveTrickCountsRoundTricks(t *testing.T) {
	g := newKoGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(KnockoutWhistPhaseTrickEnd)
	g.SetHandSize(2)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: koCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: koCard(CardDesignClover, 5)},
		{PlayerIdx: 2, Card: koCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: koCard(CardDesignClover, 12)},
	})
	g.ResolveTrick()
	if g.GetPlayer(0).GetRoundTricks() != 1 {
		t.Errorf("player 0 round tricks = %d, want 1", g.GetPlayer(0).GetRoundTricks())
	}
	if g.GetPhase() != KnockoutWhistPhaseTrickEnd {
		t.Errorf("phase after trick 1 of 2 = %d, want TrickEnd", g.GetPhase())
	}
}

func TestKnockoutWhist_ScoreRoundDogboneThenEliminate(t *testing.T) {
	g := newKoGame(false)
	g.SetPhase(KnockoutWhistPhaseRoundEnd)
	g.SetRoundNumber(1)
	// p0 won a trick -> survives. p1 0 tricks, has dogbone -> spends it. p2 0 tricks,
	// no dogbone -> eliminated. p3 won a trick -> survives.
	g.GetPlayer(0).IncRoundTricks()
	g.GetPlayer(3).IncRoundTricks()
	g.GetPlayer(2).SetDogbones(0)
	g.ScoreRound()
	if g.GetPlayer(1).GetDogbones() != 0 || g.GetPlayer(1).GetEliminated() {
		t.Errorf("p1 should have spent a dogbone and survived (dogbones=%d elim=%v)",
			g.GetPlayer(1).GetDogbones(), g.GetPlayer(1).GetEliminated())
	}
	if !g.GetPlayer(2).GetEliminated() {
		t.Error("p2 should be eliminated (0 tricks, no dogbone)")
	}
	if g.GetGameEndFlag() {
		t.Error("game should not end with 3 players still active")
	}
}

func TestKnockoutWhist_ScoreRoundLastStandingWins(t *testing.T) {
	g := newKoGame(false)
	g.SetPhase(KnockoutWhistPhaseRoundEnd)
	g.SetRoundNumber(3)
	// p0 wins the round; p1,p2,p3 have 0 tricks and 0 dogbones -> all eliminated -> p0 wins.
	g.GetPlayer(0).IncRoundTricks()
	for i := 1; i < KnockoutWhistPlayerCnt; i++ {
		g.GetPlayer(i).SetDogbones(0)
	}
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerPlayer() != 0 {
		t.Errorf("expected game end with player 0, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerPlayer())
	}
}

func TestKnockoutWhist_NextRoundShrinksHand(t *testing.T) {
	g := newKoGame(false)
	g.Reset() // round 1, hand 7
	g.SetPhase(KnockoutWhistPhaseRoundEnd)
	g.GetPlayer(0).IncRoundTricks() // ensure a round winner exists
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("round = %d, want 2", g.GetRoundNumber())
	}
	if g.GetHandSize() != 6 {
		t.Errorf("round-2 hand size = %d, want 6", g.GetHandSize())
	}
}

func TestKnockoutWhist_HintAndPlayable(t *testing.T) {
	g := newKoGame(true)
	g.SetPhase(KnockoutWhistPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	koSetHand(g.GetPlayer(0), koCard(CardDesignClover, 7), koCard(CardDesignClover, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestKnockoutWhist_CpuFullMatchProgresses(t *testing.T) {
	g := newKoGame(false) // all CPU
	g.Reset()
	for guard := 0; guard < 4000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case KnockoutWhistPhasePlay:
			g.CpuPlay()
			if g.GetPhase() == KnockoutWhistPhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == KnockoutWhistPhaseTrickEnd {
					g.NextTrick()
				}
			}
		case KnockoutWhistPhaseRoundEnd:
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
	if g.GetWinnerPlayer() < 0 {
		t.Error("game ended without a winner")
	}
}

func TestKnockoutWhist_PlayerPlayErrors(t *testing.T) {
	g := newKoGame(true)
	g.SetPhase(KnockoutWhistPhasePlay)
	g.SetCurrentPlayerIdx(0)
	koSetHand(g.GetPlayer(0), koCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(KnockoutWhistPhaseRoundEnd)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestKnockoutWhist_JSONRoundTrip(t *testing.T) {
	g := newKoGame(true)
	g.Reset()
	g.GetPlayer(1).SetDogbones(0)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 KnockoutWhist
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() || g2.GetHandSize() != g.GetHandSize() {
		t.Error("round-trip mismatch")
	}
	if g2.GetPlayer(1).GetDogbones() != 0 {
		t.Error("dogbone state not preserved across round-trip")
	}
}

func TestKnockoutWhist_UnmarshalErrors(t *testing.T) {
	var g KnockoutWhist
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

func TestKnockoutWhist_NextRoundHumanWinnerEntersTrumpSelect(t *testing.T) {
	g := newKoGame(true) // human at index 0
	g.SetPhase(KnockoutWhistPhaseRoundEnd)
	g.SetRoundNumber(2)
	g.GetPlayer(0).IncRoundTricks() // human takes the most tricks this round
	g.ScoreRound()                  // sets roundWinnerIdx = 0 (human)
	g.NextRound()                   // human winner leads -> TrumpSelect phase
	if g.GetPhase() != KnockoutWhistPhaseTrumpSelect {
		t.Fatalf("want TrumpSelect phase, got %v", g.GetPhase())
	}
	if err := g.PlayerSelectTrump(CardDesignHeart); err != nil {
		t.Fatalf("PlayerSelectTrump: %v", err)
	}
	if g.GetPhase() != KnockoutWhistPhasePlay {
		t.Errorf("want Play phase after select, got %v", g.GetPhase())
	}
	if g.GetTrumpSuit() != CardDesignHeart {
		t.Errorf("want trump %d, got %d", CardDesignHeart, g.GetTrumpSuit())
	}
}

func TestKnockoutWhist_NextRoundCpuWinnerAutoSelectsTrump(t *testing.T) {
	g := newKoGame(true)
	g.SetPhase(KnockoutWhistPhaseRoundEnd)
	g.SetRoundNumber(2)
	g.GetPlayer(1).IncRoundTricks() // CPU idx1 takes the most tricks
	g.ScoreRound()
	g.NextRound()
	if g.GetPhase() != KnockoutWhistPhasePlay {
		t.Fatalf("CPU winner should auto-select trump and enter Play, got %v", g.GetPhase())
	}
}

func TestKnockoutWhist_PlayerSelectTrumpErrors(t *testing.T) {
	g := newKoGame(true)

	g.SetPhase(KnockoutWhistPhasePlay)
	if err := g.PlayerSelectTrump(CardDesignHeart); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("want ErrWrongPhase, got %v", err)
	}

	g.SetPhase(KnockoutWhistPhaseTrumpSelect)
	g.SetLeadPlayerIdx(1) // CPU is the lead/winner
	if err := g.PlayerSelectTrump(CardDesignHeart); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("want ErrNotHumanTurn, got %v", err)
	}

	g.SetLeadPlayerIdx(0) // human lead
	if err := g.PlayerSelectTrump(0); !errors.Is(err, ErrInvalidPlay) {
		t.Errorf("want ErrInvalidPlay for suit 0, got %v", err)
	}
	if err := g.PlayerSelectTrump(5); !errors.Is(err, ErrInvalidPlay) {
		t.Errorf("want ErrInvalidPlay for suit 5, got %v", err)
	}
}

func TestKnockoutWhist_JSONRoundTripTrumpSelect(t *testing.T) {
	g := newKoGame(true)
	g.SetPhase(KnockoutWhistPhaseTrumpSelect)
	g.SetRoundNumber(2)
	g.SetHandSize(6)
	g.SetTrickNumber(1)
	g.SetTrumpSuit(CardDesignSpade)
	g.SetLeadPlayerIdx(0)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 KnockoutWhist
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPhase() != KnockoutWhistPhaseTrumpSelect {
		t.Errorf("phase not preserved: got %v", g2.GetPhase())
	}
}
