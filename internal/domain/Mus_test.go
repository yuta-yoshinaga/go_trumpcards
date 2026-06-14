//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func musCard(design, value int) *Card { return NewCard(design, value, false) }

func newMusGame(human bool) *Mus {
	players := make([]*MusPlayer, MusPlayerCnt)
	players[0] = NewMusPlayer(human)
	for i := 1; i < MusPlayerCnt; i++ {
		players[i] = NewMusPlayer(false)
	}
	return NewMus(NewTrumpCardsBriscola(), players, DefaultMusConfig())
}

func musSetHand(p *MusPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestMusConfig_Validate(t *testing.T) {
	if err := DefaultMusConfig().Validate(); err != nil {
		t.Fatalf("default invalid: %v", err)
	}
	bad := []MusConfig{
		{CpuDifficulty: -1, TargetAmarrakos: 40},
		{CpuDifficulty: 9, TargetAmarrakos: 40},
		{CpuDifficulty: 1, TargetAmarrakos: 0},
	}
	for i, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestMus_ResetDeals4Each(t *testing.T) {
	g := newMusGame(true)
	g.Reset()
	if g.GetPhase() != MusPhaseMus {
		t.Fatalf("phase = %v, want Mus", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != MusHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, MusHandSize)
		}
	}
	if g.GetMusTurn() != g.GetManoIdx() {
		t.Errorf("mus turn should start at mano")
	}
}

func TestMus_TeamOf(t *testing.T) {
	if MusTeamOf(0) != 0 || MusTeamOf(2) != 0 || MusTeamOf(1) != 1 || MusTeamOf(3) != 1 {
		t.Error("team assignment wrong")
	}
}

func TestMus_CardHelpers(t *testing.T) {
	if musCardRank(13) != 10 || musCardRank(12) != 9 || musCardRank(11) != 8 || musCardRank(1) != 1 || musCardRank(7) != 7 {
		t.Error("musCardRank wrong")
	}
	if musCardPoints(13) != 10 || musCardPoints(11) != 10 || musCardPoints(1) != 1 || musCardPoints(7) != 7 {
		t.Error("musCardPoints wrong")
	}
}

func TestMus_ParesCategory(t *testing.T) {
	cases := []struct {
		ranks []int
		want  int
	}{
		{[]int{10, 9, 8, 7}, 0},    // no pair
		{[]int{10, 10, 8, 7}, 1},   // one pair
		{[]int{10, 10, 10, 7}, 2},  // medias (three)
		{[]int{10, 10, 8, 8}, 3},   // duples (two pair)
		{[]int{10, 10, 10, 10}, 3}, // four → duples
	}
	for _, c := range cases {
		if got := musParesCategory(c.ranks); got != c.want {
			t.Errorf("paresCategory(%v) = %d, want %d", c.ranks, got, c.want)
		}
	}
}

func TestMus_GrandeChicaEncoding(t *testing.T) {
	high := []int{10, 10, 9, 8} // K K Q J
	low := []int{1, 1, 2, 3}    // A A 2 3
	if musEncodeDesc(high) <= musEncodeDesc(low) {
		t.Error("grande: high hand should encode higher")
	}
	if musEncodeAsc(low) >= musEncodeAsc(high) {
		t.Error("chica: low hand should encode lower")
	}
}

func TestMus_JuegoKey(t *testing.T) {
	// 31 best, then 32, then 40..33; punto (<31) ranked by points.
	if musJuegoKey(31) <= musJuegoKey(32) {
		t.Error("31 should beat 32")
	}
	if musJuegoKey(32) <= musJuegoKey(40) {
		t.Error("32 should beat 40")
	}
	if musJuegoKey(40) <= musJuegoKey(33) {
		t.Error("40 should beat 33")
	}
	if musJuegoKey(31) <= musJuegoKey(30) {
		t.Error("juego should beat punto")
	}
	if musJuegoKey(30) <= musJuegoKey(20) {
		t.Error("punto 30 should beat punto 20")
	}
}

func TestMus_MusPhaseAllMusThenDiscard(t *testing.T) {
	g := newMusGame(false) // all CPU for resolveMus driving
	g.Reset()
	g.SetManoIdx(0)
	g.SetMusTurn(0)
	g.musAgreed = 0
	for i := 0; i < 4; i++ {
		g.resolveMus(true)
	}
	if g.GetPhase() != MusPhaseDiscard {
		t.Errorf("phase = %v, want Discard after all mus", g.GetPhase())
	}
}

func TestMus_MusCutStartsBetting(t *testing.T) {
	g := newMusGame(true)
	g.Reset()
	g.SetManoIdx(0)
	g.SetMusTurn(0)
	if err := g.PlayerMus(false); err != nil { // cut
		t.Fatalf("mus err: %v", err)
	}
	if g.GetPhase() != MusPhaseGrande {
		t.Errorf("phase = %v, want Grande after cut", g.GetPhase())
	}
}

func TestMus_PlayerMusErrors(t *testing.T) {
	g := newMusGame(true)
	g.Reset()
	g.SetMusTurn(1) // CPU
	if err := g.PlayerMus(true); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
	g.SetMusTurn(0)
	g.SetPhase(MusPhaseShowdown)
	if err := g.PlayerMus(true); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
	g.gameEndFlag = true
	if err := g.PlayerMus(true); !errors.Is(err, ErrGameEnded) {
		t.Errorf("err = %v, want ErrGameEnded", err)
	}
}

func TestMus_BettingPasoPasoDeferred(t *testing.T) {
	g := newMusGame(true)
	g.SetPhase(MusPhaseGrande)
	g.SetManoIdx(0)
	g.startBetRound() // betTeam = team0 (mano)
	// team0 (human) paso
	if err := g.PlayerBet(MusActionPaso, 0); err != nil {
		t.Fatalf("paso err: %v", err)
	}
	if g.GetBetTeam() != 1 {
		t.Errorf("after first paso betTeam should be 1")
	}
	// team1 (CPU) paso via resolveBet directly
	_ = g.resolveBet(MusActionPaso, 0)
	if g.GetResult(0).Kind != MusResultDeferred {
		t.Errorf("both paso should defer Grande, got kind %d", g.GetResult(0).Kind)
	}
	if g.GetPhase() != MusPhaseChica {
		t.Errorf("should advance to Chica")
	}
}

func TestMus_BettingEnvidoQuieroAccepted(t *testing.T) {
	g := newMusGame(true)
	g.SetPhase(MusPhaseGrande)
	g.SetManoIdx(0)
	g.startBetRound()
	// team0 envido 2
	if err := g.PlayerBet(MusActionEnvido, 2); err != nil {
		t.Fatalf("envido err: %v", err)
	}
	if g.GetPendingStake() != 2 || g.GetLastBettorTeam() != 0 {
		t.Errorf("pending stake/bettor wrong")
	}
	// team1 quiero
	_ = g.resolveBet(MusActionQuiero, 0)
	if g.GetResult(0).Kind != MusResultAccepted || g.GetResult(0).Stake != 2 {
		t.Errorf("Grande should be accepted for 2, got %+v", g.GetResult(0))
	}
}

func TestMus_BettingNoQuieroAwards(t *testing.T) {
	g := newMusGame(true)
	g.SetPhase(MusPhaseGrande)
	g.SetManoIdx(0)
	g.startBetRound()
	_ = g.resolveBet(MusActionEnvido, 4) // team0 bets
	before := g.GetAmarrakos()
	_ = g.resolveBet(MusActionNoQuiero, 0) // team1 declines
	after := g.GetAmarrakos()
	if after[0] != before[0]+1 {
		t.Errorf("no quiero should give bettor team0 +1: %v -> %v", before, after)
	}
	if g.GetResult(0).Kind != MusResultAwarded {
		t.Errorf("kind should be awarded")
	}
}

func TestMus_OrdagoEndsGame(t *testing.T) {
	g := newMusGame(true)
	g.SetPhase(MusPhaseGrande)
	g.SetManoIdx(0)
	g.startBetRound()
	// Give team0 the strongest Grande hand so it wins the ordago showdown.
	musSetHand(g.GetPlayer(0), musCard(1, 13), musCard(2, 13), musCard(3, 12), musCard(4, 12)) // K K Q Q
	musSetHand(g.GetPlayer(2), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 4))
	musSetHand(g.GetPlayer(1), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 5))
	musSetHand(g.GetPlayer(3), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 6))
	_ = g.resolveBet(MusActionOrdago, 0) // team0 ordago
	_ = g.resolveBet(MusActionQuiero, 0) // team1 accepts
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("ordago: team0 should win game, end=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerTeam())
	}
}

func TestMus_ParesSkippedWhenNoPairs(t *testing.T) {
	g := newMusGame(true)
	g.SetManoIdx(0)
	g.SetPhase(MusPhasePares)
	// Nobody has a pair, and nobody has juego (all <31) so Juego becomes a Punto
	// betting round and the cascade stops there (phase == Juego).
	musSetHand(g.GetPlayer(0), musCard(1, 13), musCard(2, 12), musCard(3, 7), musCard(4, 2)) // K Q 7 2 = 29
	musSetHand(g.GetPlayer(1), musCard(1, 1), musCard(2, 3), musCard(3, 4), musCard(4, 5))   // 13
	musSetHand(g.GetPlayer(2), musCard(1, 6), musCard(2, 11), musCard(3, 1), musCard(4, 3))  // 20
	musSetHand(g.GetPlayer(3), musCard(1, 2), musCard(2, 4), musCard(3, 6), musCard(4, 11))  // 22
	g.startBetRound()
	if g.GetResult(2).Kind != MusResultSkipped {
		t.Errorf("Pares should be skipped with no pairs, got %d", g.GetResult(2).Kind)
	}
	if g.GetPhase() != MusPhaseJuego {
		t.Errorf("should advance past Pares to Juego (Punto betting), got phase %v", g.GetPhase())
	}
}

func TestMus_ParesAutoAwardOneTeam(t *testing.T) {
	g := newMusGame(true)
	g.SetManoIdx(0)
	g.SetPhase(MusPhasePares)
	// Only team0 (player 0) has a pair; nobody has juego (all <31) so the Juego
	// round is Punto betting and does not auto-award (isolates the Pares award).
	musSetHand(g.GetPlayer(0), musCard(1, 13), musCard(2, 13), musCard(3, 1), musCard(4, 2)) // K K A 2 = pair, 23
	musSetHand(g.GetPlayer(2), musCard(1, 1), musCard(2, 2), musCard(3, 4), musCard(4, 6))   // no pair, 13
	musSetHand(g.GetPlayer(1), musCard(1, 3), musCard(2, 4), musCard(3, 5), musCard(4, 7))   // no pair, 19
	musSetHand(g.GetPlayer(3), musCard(1, 2), musCard(2, 5), musCard(3, 6), musCard(4, 7))   // no pair, 20
	g.startBetRound()
	if g.GetResult(2).Kind != MusResultAwarded || g.GetResult(2).Team != 0 {
		t.Errorf("Pares should auto-award team0, got %+v", g.GetResult(2))
	}
	if g.GetAmarrakos()[0] != 1 {
		t.Errorf("team0 should have 1 amarrako, got %v", g.GetAmarrakos())
	}
}

func TestMus_ShowdownAwardsAndRoundWinner(t *testing.T) {
	g := newMusGame(true)
	g.SetManoIdx(0)
	g.SetPhase(MusPhaseShowdown)
	// Grande accepted for 3; team0 has higher cards → team0 wins.
	g.results[0] = MusRoundResult{Kind: MusResultAccepted, Stake: 3, Team: -1}
	musSetHand(g.GetPlayer(0), musCard(1, 13), musCard(2, 13), musCard(3, 12), musCard(4, 12))
	musSetHand(g.GetPlayer(2), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 4))
	musSetHand(g.GetPlayer(1), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 5))
	musSetHand(g.GetPlayer(3), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 6))
	g.Showdown()
	if g.GetAmarrakos()[0] != 3 {
		t.Errorf("team0 should win Grande +3, got %v", g.GetAmarrakos())
	}
	if g.GetPhase() != MusPhaseRoundEnd {
		t.Errorf("phase should be RoundEnd after showdown")
	}
}

func TestMus_ShowdownReachesTargetEndsGame(t *testing.T) {
	g := newMusGame(true)
	g.SetManoIdx(0)
	cfg := g.GetConfig()
	cfg.TargetAmarrakos = 2
	g.SetConfig(cfg)
	g.SetPhase(MusPhaseShowdown)
	g.results[0] = MusRoundResult{Kind: MusResultAccepted, Stake: 5, Team: -1}
	musSetHand(g.GetPlayer(0), musCard(1, 13), musCard(2, 13), musCard(3, 12), musCard(4, 12))
	musSetHand(g.GetPlayer(2), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 4))
	musSetHand(g.GetPlayer(1), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 5))
	musSetHand(g.GetPlayer(3), musCard(1, 1), musCard(2, 2), musCard(3, 3), musCard(4, 6))
	g.Showdown()
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("game should end with team0 winner")
	}
}

func TestMus_NextRound(t *testing.T) {
	g := newMusGame(false)
	g.Reset()
	g.SetPhase(MusPhaseRoundEnd)
	m := g.GetManoIdx()
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetManoIdx() != (m+1)%MusPlayerCnt {
		t.Errorf("next round/mano wrong: round=%d mano=%d", g.GetRoundNumber(), g.GetManoIdx())
	}
	if g.GetPhase() != MusPhaseMus {
		t.Errorf("phase = %v, want Mus", g.GetPhase())
	}
}

func TestMus_CpuFullGameProgresses(t *testing.T) {
	g := newMusGame(false) // all CPU
	g.Reset()
	for steps := 0; steps < 5000; steps++ {
		switch g.GetPhase() {
		case MusPhaseMus, MusPhaseDiscard, MusPhaseGrande, MusPhaseChica, MusPhasePares, MusPhaseJuego:
			g.CpuPlay()
		case MusPhaseShowdown:
			g.Showdown()
		case MusPhaseRoundEnd:
			g.NextRound()
		case MusPhaseGameEnd:
			if g.GetWinnerTeam() < 0 {
				t.Error("game ended without winner")
			}
			return
		}
		if g.GetRoundNumber() > 500 {
			t.Fatal("game did not converge")
		}
	}
	t.Fatal("game did not reach end in step budget")
}

func TestMus_HintPerPhase(t *testing.T) {
	g := newMusGame(true)
	g.Reset()
	g.SetManoIdx(0)
	g.SetMusTurn(0)
	if h := g.GetHint(); h == nil || (h.Reason != "mus_cut" && h.Reason != "mus_exchange") {
		t.Errorf("mus hint missing: %+v", h)
	}
	g.SetPhase(MusPhaseGrande)
	g.startBetRound()
	if h := g.GetHint(); h == nil {
		t.Errorf("bet hint missing")
	}
}

func TestMus_IsHumanTurn(t *testing.T) {
	g := newMusGame(true)
	g.SetPhase(MusPhaseMus)
	g.SetMusTurn(0)
	if !g.IsHumanTurn() {
		t.Error("human mus turn")
	}
	g.SetMusTurn(1)
	if g.IsHumanTurn() {
		t.Error("CPU mus turn should not be human")
	}
	g.SetPhase(MusPhaseGrande)
	g.SetBetTeam(0)
	if !g.IsHumanTurn() {
		t.Error("team0 bet should be human turn")
	}
	g.SetBetTeam(1)
	if g.IsHumanTurn() {
		t.Error("team1 bet should not be human turn")
	}
}

func TestMus_JSONRoundTrip(t *testing.T) {
	g := newMusGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Mus
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPhase() != g.GetPhase() || g2.GetPlayerCnt() != MusPlayerCnt {
		t.Error("round trip mismatch")
	}
}

func TestMus_UnmarshalOversized(t *testing.T) {
	var g Mus
	bad := `{"al":[`
	for i := 0; i < musMaxSliceLen+1; i++ {
		if i > 0 {
			bad += ","
		}
		bad += `{"TurnNumber":0,"PlayerIdx":-1,"ActionType":"x","Detail":"y","Cards":null}`
	}
	bad += `]}`
	if err := json.Unmarshal([]byte(bad), &g); !errors.Is(err, errMusOversized) {
		t.Errorf("err = %v, want errMusOversized", err)
	}
}
