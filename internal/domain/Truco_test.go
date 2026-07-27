package domain

import (
	"encoding/json"
	"testing"
)

// --- helpers ---

// twoHumanTruco は両プレイヤーを人間扱いにした Truco を返す。
// PlayerPlay でどちらの手番も駆動できるため、CPU の乱択を介さずに
// ルールエンジンを決定論的にテストできる。
func twoHumanTruco() *Truco {
	players := []*TrucoPlayer{NewTrucoPlayer(true), NewTrucoPlayer(true)}
	g := NewTruco(NewTrumpCardsBriscola(), players, DefaultTrucoConfig())
	g.Reset()
	return g
}

func trucoSetHand(p *TrucoPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playByCard は指定プレイヤーに、手札中の (value, design) のカードを出させる。
func playByCard(t *testing.T, g *Truco, playerIdx, value, design int) {
	t.Helper()
	p := g.GetPlayer(playerIdx)
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetValue() == value && c.GetDesign() == design {
			g.SetCurrentPlayerIdx(playerIdx)
			if err := g.PlayerPlay(i); err != nil {
				t.Fatalf("PlayerPlay(%d) p%d: %v", i, playerIdx, err)
			}
			return
		}
	}
	t.Fatalf("player %d has no card (v=%d d=%d)", playerIdx, value, design)
}

// --- card strength ---

func TestTrucoCardStrength(t *testing.T) {
	tests := []struct {
		name  string
		card  *Card
		power int
	}{
		{"1 espadas (top)", NewCard(CardDesignSpade, 1, false), 14},
		{"1 bastos", NewCard(CardDesignClover, 1, false), 13},
		{"7 espadas", NewCard(CardDesignSpade, 7, false), 12},
		{"7 oros", NewCard(CardDesignDiamond, 7, false), 11},
		{"any 3", NewCard(CardDesignHeart, 3, false), 10},
		{"any 2", NewCard(CardDesignClover, 2, false), 9},
		{"false ace (copas)", NewCard(CardDesignHeart, 1, false), 8},
		{"false ace (oros)", NewCard(CardDesignDiamond, 1, false), 8},
		{"rey", NewCard(CardDesignSpade, 13, false), 7},
		{"caballo", NewCard(CardDesignSpade, 12, false), 6},
		{"sota", NewCard(CardDesignSpade, 11, false), 5},
		{"false 7 (copas)", NewCard(CardDesignHeart, 7, false), 4},
		{"false 7 (bastos)", NewCard(CardDesignClover, 7, false), 4},
		{"6", NewCard(CardDesignSpade, 6, false), 3},
		{"5", NewCard(CardDesignSpade, 5, false), 2},
		{"4", NewCard(CardDesignSpade, 4, false), 1},
		{"nil", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrucoCardStrength(tt.card); got != tt.power {
				t.Errorf("strength = %d, want %d", got, tt.power)
			}
		})
	}
}

func TestTrucoCardStrengthTotalOrder(t *testing.T) {
	// matadores はスート固有: 1♠ > 1♣ > 7♠ > 7♦ で、偽のエース/7 より強い。
	if TrucoCardStrength(NewCard(CardDesignSpade, 1, false)) <=
		TrucoCardStrength(NewCard(CardDesignClover, 1, false)) {
		t.Error("1 espadas should beat 1 bastos")
	}
	if TrucoCardStrength(NewCard(CardDesignSpade, 7, false)) <=
		TrucoCardStrength(NewCard(CardDesignDiamond, 7, false)) {
		t.Error("7 espadas should beat 7 oros")
	}
	if TrucoCardStrength(NewCard(CardDesignDiamond, 7, false)) <=
		TrucoCardStrength(NewCard(CardDesignHeart, 7, false)) {
		t.Error("7 oros (matador) should beat false 7")
	}
}

func TestTrucoLevelValue(t *testing.T) {
	cases := map[int]int{
		TrucoLevelNone:       1,
		TrucoLevelTruco:      2,
		TrucoLevelRetruco:    3,
		TrucoLevelValeCuatro: 4,
		-5:                   1,
		99:                   TrucoMaxLevel + 1,
	}
	for level, want := range cases {
		if got := TrucoLevelValue(level); got != want {
			t.Errorf("TrucoLevelValue(%d) = %d, want %d", level, got, want)
		}
	}
}

// --- baza winner ---

func TestTrucoBazaWinner(t *testing.T) {
	strong := NewCard(CardDesignSpade, 1, false) // 14
	weak := NewCard(CardDesignSpade, 4, false)   // 1
	tieA := NewCard(CardDesignHeart, 5, false)   // 2
	tieB := NewCard(CardDesignDiamond, 5, false) // 2

	if got := trucoBazaWinner([]*TrickCard{{0, strong}, {1, weak}}); got != 0 {
		t.Errorf("strong-vs-weak winner = %d, want 0", got)
	}
	if got := trucoBazaWinner([]*TrickCard{{0, weak}, {1, strong}}); got != 1 {
		t.Errorf("weak-vs-strong winner = %d, want 1", got)
	}
	if got := trucoBazaWinner([]*TrickCard{{0, tieA}, {1, tieB}}); got != -1 {
		t.Errorf("equal strength = %d, want -1 (parda)", got)
	}
	if got := trucoBazaWinner([]*TrickCard{{0, strong}}); got != -1 {
		t.Errorf("incomplete baza = %d, want -1", got)
	}
}

// --- mano resolution (pure) ---

func TestResolveMano(t *testing.T) {
	const elder = 1
	tests := []struct {
		name        string
		results     []int
		wantDecided bool
		wantWinner  int
	}{
		{"empty", []int{}, false, -1},
		{"one win", []int{0}, false, -1},
		{"two straight wins p0", []int{0, 0}, true, 0},
		{"two straight wins p1", []int{1, 1}, true, 1},
		{"win then tie -> first winner", []int{0, -1}, true, 0},
		{"tie then win -> second winner", []int{-1, 1}, true, 1},
		{"split -> undecided after two", []int{0, 1}, false, -1},
		{"both parda -> undecided after two", []int{-1, -1}, false, -1},
		{"split then third decides", []int{0, 1, 1}, true, 1},
		{"split then third parda -> first winner", []int{0, 1, -1}, true, 0},
		{"split (1,0) then third parda -> first winner", []int{1, 0, -1}, true, 1},
		{"all parda -> elder", []int{-1, -1, -1}, true, elder},
		{"two parda then win", []int{-1, -1, 0}, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decided, winner := resolveMano(tt.results, elder)
			if decided != tt.wantDecided {
				t.Errorf("decided = %v, want %v", decided, tt.wantDecided)
			}
			if decided && winner != tt.wantWinner {
				t.Errorf("winner = %d, want %d", winner, tt.wantWinner)
			}
		})
	}
}

// --- full mano play (deterministic, two humans) ---

func setupBaza(g *Truco, lead int) {
	g.SetPhase(TrucoPhasePlay)
	g.SetLeadPlayerIdx(lead)
	g.SetCurrentPlayerIdx(lead)
	g.SetTrickNumber(1)
	g.SetTrickResults(nil)
	g.SetCurrentTrick(nil)
	g.SetAcceptedLevel(TrucoLevelNone)
	g.SetHandStake(1)
}

func TestTrucoManoTwoStraightWins(t *testing.T) {
	g := twoHumanTruco()
	trucoSetHand(g.GetPlayer(0), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 2, false), NewCard(CardDesignHeart, 1, false))
	trucoSetHand(g.GetPlayer(1), NewCard(CardDesignDiamond, 4, false), NewCard(CardDesignDiamond, 5, false), NewCard(CardDesignDiamond, 6, false))
	setupBaza(g, 0)

	// baza 1: 3♥ (10) vs 4♦ (1) -> p0
	playByCard(t, g, 0, 3, CardDesignHeart)
	playByCard(t, g, 1, 4, CardDesignDiamond)
	if g.GetPhase() != TrucoPhaseTrickEnd {
		t.Fatalf("after baza1 phase = %d, want TrickEnd", g.GetPhase())
	}
	g.Next()
	if g.GetPhase() != TrucoPhasePlay || g.GetTrickNumber() != 2 {
		t.Fatalf("baza2 start: phase=%d trick=%d", g.GetPhase(), g.GetTrickNumber())
	}
	if g.GetLeadPlayerIdx() != 0 {
		t.Errorf("baza2 lead = %d, want 0 (baza1 winner)", g.GetLeadPlayerIdx())
	}

	// baza 2: 2♥ (9) vs 5♦ (2) -> p0 wins -> mano decided
	playByCard(t, g, 0, 2, CardDesignHeart)
	playByCard(t, g, 1, 5, CardDesignDiamond)
	g.Next()
	if g.GetPhase() != TrucoPhaseHandEnd {
		t.Fatalf("after 2 wins phase = %d, want HandEnd", g.GetPhase())
	}
	if g.GetHandWinnerIdx() != 0 {
		t.Errorf("hand winner = %d, want 0", g.GetHandWinnerIdx())
	}
}

func TestTrucoManoTieFirstThenWin(t *testing.T) {
	g := twoHumanTruco()
	trucoSetHand(g.GetPlayer(0), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignHeart, 3, false))
	trucoSetHand(g.GetPlayer(1), NewCard(CardDesignDiamond, 5, false), NewCard(CardDesignDiamond, 4, false))
	setupBaza(g, 0)

	// baza1: 5♥ vs 5♦ -> parda
	playByCard(t, g, 0, 5, CardDesignHeart)
	playByCard(t, g, 1, 5, CardDesignDiamond)
	g.Next()
	if len(g.GetTrickResults()) != 1 || g.GetTrickResults()[0] != -1 {
		t.Fatalf("baza1 result = %v, want [-1]", g.GetTrickResults())
	}
	// baza2: 3♥ (10) vs 4♦ (1) -> p0 -> tie-then-win => p0 wins mano
	playByCard(t, g, 0, 3, CardDesignHeart)
	playByCard(t, g, 1, 4, CardDesignDiamond)
	g.Next()
	if g.GetPhase() != TrucoPhaseHandEnd || g.GetHandWinnerIdx() != 0 {
		t.Fatalf("phase=%d winner=%d, want HandEnd/0", g.GetPhase(), g.GetHandWinnerIdx())
	}
}

func TestTrucoManoAllPardaGoesToElder(t *testing.T) {
	g := twoHumanTruco()
	g.SetManoIdx(1)
	trucoSetHand(g.GetPlayer(0), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignHeart, 6, false))
	trucoSetHand(g.GetPlayer(1), NewCard(CardDesignDiamond, 4, false), NewCard(CardDesignDiamond, 5, false), NewCard(CardDesignDiamond, 6, false))
	setupBaza(g, 0)

	for _, v := range []int{4, 5, 6} {
		playByCard(t, g, 0, v, CardDesignHeart)
		playByCard(t, g, 1, v, CardDesignDiamond)
		g.Next()
	}
	if g.GetPhase() != TrucoPhaseHandEnd {
		t.Fatalf("phase = %d, want HandEnd", g.GetPhase())
	}
	if g.GetHandWinnerIdx() != 1 {
		t.Errorf("all-parda winner = %d, want elder (1)", g.GetHandWinnerIdx())
	}
}

func TestTrucoManoSplitThenThird(t *testing.T) {
	g := twoHumanTruco()
	trucoSetHand(g.GetPlayer(0), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignSpade, 1, false))
	trucoSetHand(g.GetPlayer(1), NewCard(CardDesignDiamond, 4, false), NewCard(CardDesignClover, 3, false), NewCard(CardDesignDiamond, 6, false))
	setupBaza(g, 0)

	// baza1: 3♥(10) vs 4♦(1) -> p0
	playByCard(t, g, 0, 3, CardDesignHeart)
	playByCard(t, g, 1, 4, CardDesignDiamond)
	g.Next()
	// baza2: lead is p0; 4♥(1) vs 3♣(10) -> p1 -> split
	playByCard(t, g, 0, 4, CardDesignHeart)
	playByCard(t, g, 1, 3, CardDesignClover)
	g.Next()
	if g.GetPhase() != TrucoPhasePlay || g.GetTrickNumber() != 3 {
		t.Fatalf("expected baza3: phase=%d trick=%d", g.GetPhase(), g.GetTrickNumber())
	}
	// baza3: 1♠(14) vs 6♦(3) -> p0 wins the mano
	playByCard(t, g, 0, 1, CardDesignSpade)
	playByCard(t, g, 1, 6, CardDesignDiamond)
	g.Next()
	if g.GetPhase() != TrucoPhaseHandEnd || g.GetHandWinnerIdx() != 0 {
		t.Fatalf("phase=%d winner=%d, want HandEnd/0", g.GetPhase(), g.GetHandWinnerIdx())
	}
}

// --- truco escalation (unit on internal call/respond) ---

func TestTrucoCallAndAccept(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	g.callTruco(1)
	if g.GetPhase() != TrucoPhaseRespond {
		t.Fatalf("phase = %d, want Respond", g.GetPhase())
	}
	if g.GetResponderIdx() != 0 || g.GetTrucoCallerIdx() != 1 || g.GetPendingLevel() != TrucoLevelTruco {
		t.Fatalf("responder=%d caller=%d pending=%d", g.GetResponderIdx(), g.GetTrucoCallerIdx(), g.GetPendingLevel())
	}
	g.respond(0, true)
	if g.GetPhase() != TrucoPhasePlay {
		t.Fatalf("phase after accept = %d, want Play", g.GetPhase())
	}
	if g.GetAcceptedLevel() != TrucoLevelTruco || g.GetHandStake() != 2 {
		t.Errorf("acceptedLevel=%d handStake=%d, want 1/2", g.GetAcceptedLevel(), g.GetHandStake())
	}
	if g.GetResponderIdx() != -1 || g.GetTrucoCallerIdx() != -1 || g.GetPendingLevel() != 0 {
		t.Errorf("pending state not cleared after accept")
	}
}

func TestTrucoCallAndDeclineAwardsPriorStake(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	// no truco accepted yet -> decline awards 1 pt to caller
	g.callTruco(1)
	g.respond(0, false)
	if g.GetPhase() != TrucoPhaseHandEnd {
		t.Fatalf("phase = %d, want HandEnd", g.GetPhase())
	}
	if g.GetHandWinnerIdx() != 1 {
		t.Errorf("hand winner = %d, want caller 1", g.GetHandWinnerIdx())
	}
	if g.GetHandStake() != 1 {
		t.Errorf("decline at base awards %d, want 1", g.GetHandStake())
	}
}

func TestTrucoDeclineAfterAcceptedTruco(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	g.callTruco(0)      // p0 calls Truco
	g.respond(1, true)  // p1 accepts -> stake 2
	g.callTruco(0)      // p0 calls Retruco (pending=2)
	g.respond(1, false) // p1 declines -> p0 wins at prior stake (2)
	if g.GetHandWinnerIdx() != 0 {
		t.Errorf("hand winner = %d, want 0", g.GetHandWinnerIdx())
	}
	if g.GetHandStake() != 2 {
		t.Errorf("decline after accepted Truco awards %d, want 2", g.GetHandStake())
	}
}

func TestTrucoReRaiseChain(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	g.callTruco(0) // p0 Truco (pending 1, responder 1)
	if g.GetResponderIdx() != 1 {
		t.Fatalf("responder = %d, want 1", g.GetResponderIdx())
	}
	g.callTruco(1) // p1 re-raises to Retruco (pending 2, responder 0)
	if g.GetPendingLevel() != TrucoLevelRetruco || g.GetResponderIdx() != 0 || g.GetTrucoCallerIdx() != 1 {
		t.Fatalf("after re-raise pending=%d responder=%d caller=%d", g.GetPendingLevel(), g.GetResponderIdx(), g.GetTrucoCallerIdx())
	}
	g.respond(0, true)
	if g.GetAcceptedLevel() != TrucoLevelRetruco || g.GetHandStake() != 3 {
		t.Errorf("acceptedLevel=%d stake=%d, want 2/3", g.GetAcceptedLevel(), g.GetHandStake())
	}
}

func TestTrucoCanDeclareMaxLevel(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	g.SetAcceptedLevel(TrucoLevelValeCuatro)
	if g.canDeclare(0) {
		t.Error("canDeclare at Vale Cuatro should be false")
	}
	// pending at max in respond phase -> cannot re-raise
	g.SetPhase(TrucoPhaseRespond)
	g.SetPendingLevel(TrucoMaxLevel)
	g.SetResponderIdx(0)
	if g.canDeclare(0) {
		t.Error("canDeclare with pending at max should be false")
	}
}

func TestTrucoDeclareGuards(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	// human idx 0 leads -> declareActor returns 0 -> allowed
	if !g.CanDeclareTruco() {
		t.Error("human should be able to declare at start of own turn")
	}
	if err := g.DeclareTruco(); err != nil {
		t.Fatalf("DeclareTruco: %v", err)
	}
	// now responder is 1 (not human-as-0); CanDeclareTruco should be false for actor!=0
	if g.GetResponderIdx() != 1 {
		t.Fatalf("responder = %d", g.GetResponderIdx())
	}
}

// --- match accumulation ---

func TestTrucoMatchEndOnTarget(t *testing.T) {
	g := twoHumanTruco()
	g.SetPlayerMatchPoints(0, 14)
	g.SetHandWinnerIdx(0)
	g.SetHandStake(2)
	g.SetPhase(TrucoPhaseHandEnd)
	g.Next()
	if !g.GetGameEndFlag() {
		t.Fatal("expected game end after reaching target")
	}
	if g.GetWinnerIdx() != 0 {
		t.Errorf("winner = %d, want 0", g.GetWinnerIdx())
	}
	if g.GetPhase() != TrucoPhaseGameEnd {
		t.Errorf("phase = %d, want GameEnd", g.GetPhase())
	}
	if g.GetPlayerMatchPoints(0) != 16 {
		t.Errorf("points = %d, want 16", g.GetPlayerMatchPoints(0))
	}
}

func TestTrucoMatchContinuesBelowTarget(t *testing.T) {
	g := twoHumanTruco()
	startHand := g.GetHandNumber()
	g.SetPlayerMatchPoints(0, 5)
	g.SetHandWinnerIdx(0)
	g.SetHandStake(2)
	g.SetPhase(TrucoPhaseHandEnd)
	g.Next()
	if g.GetGameEndFlag() {
		t.Fatal("game should continue below target")
	}
	if g.GetPhase() != TrucoPhasePlay {
		t.Errorf("phase = %d, want Play (new hand)", g.GetPhase())
	}
	if g.GetHandNumber() != startHand+1 {
		t.Errorf("hand number = %d, want %d", g.GetHandNumber(), startHand+1)
	}
	if g.GetPlayer(0).GetCardsSize() != TrucoHandSize {
		t.Errorf("new hand size = %d, want %d", g.GetPlayer(0).GetCardsSize(), TrucoHandSize)
	}
	// dealer/mano should have alternated
	if g.GetDealerIdx() == 0 {
		t.Error("dealer should have flipped from 0")
	}
}

// --- error / guard paths ---

func TestTrucoPlayerPlayErrors(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	trucoSetHand(g.GetPlayer(0), NewCard(CardDesignSpade, 1, false))

	g.SetCurrentPlayerIdx(0)
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected error for out-of-range index")
	}
	g.SetPhase(TrucoPhaseTrickEnd)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected ErrWrongPhase")
	}
	g.SetPhase(TrucoPhasePlay)
	g.SetGameEndFlag(true)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected ErrGameEnded")
	}
}

func TestTrucoPlayerPlayNotHumanTurn(t *testing.T) {
	players := []*TrucoPlayer{NewTrucoPlayer(true), NewTrucoPlayer(false)}
	g := NewTruco(NewTrumpCardsBriscola(), players, DefaultTrucoConfig())
	g.Reset()
	setupBaza(g, 1) // CPU leads
	trucoSetHand(g.GetPlayer(1), NewCard(CardDesignSpade, 1, false))
	g.SetCurrentPlayerIdx(1)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected ErrNotHumanTurn when CPU is current")
	}
}

func TestTrucoRespondGuards(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	if err := g.RespondTruco(true); err == nil {
		t.Error("expected ErrWrongPhase when not in Respond")
	}
	g.callTruco(1) // responder = 0
	g.SetResponderIdx(1)
	if err := g.RespondTruco(true); err == nil {
		t.Error("expected ErrNotHumanTurn when responder != 0")
	}
	g.SetGameEndFlag(true)
	if err := g.RespondTruco(true); err == nil {
		t.Error("expected ErrGameEnded")
	}
}

func TestTrucoDeclareWrongPhase(t *testing.T) {
	g := twoHumanTruco()
	g.SetPhase(TrucoPhaseTrickEnd)
	if err := g.DeclareTruco(); err == nil {
		t.Error("expected error declaring in TrickEnd")
	}
}

// --- getters ---

func TestTrucoGettersAndValidIndices(t *testing.T) {
	g := twoHumanTruco()
	if g.GetPlayerCnt() != 2 {
		t.Errorf("player count = %d, want 2", g.GetPlayerCnt())
	}
	if g.GetPlayer(99) != nil {
		t.Error("out-of-range player should be nil")
	}
	if g.GetMatchTarget() != TrucoDefaultMatchTarget {
		t.Errorf("match target = %d, want %d", g.GetMatchTarget(), TrucoDefaultMatchTarget)
	}
	if g.GetPlayerMatchPoints(99) != 0 {
		t.Error("out-of-range match points should be 0")
	}
	idxs := g.GetValidPlayIndices(0)
	if len(idxs) != g.GetPlayer(0).GetCardsSize() {
		t.Errorf("valid indices = %d, want %d", len(idxs), g.GetPlayer(0).GetCardsSize())
	}
	if g.GetValidPlayIndices(99) != nil {
		t.Error("invalid player valid indices should be nil")
	}
	if len(g.GetActionLog()) == 0 {
		t.Error("action log should contain deal entry")
	}
}

func TestTrucoHint(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	trucoSetHand(g.GetPlayer(0), NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 4, false))
	g.SetCurrentPlayerIdx(0)
	h := g.GetHint()
	if h == nil {
		t.Fatal("expected hint during human play turn")
	}
	if h.Action != "call" && h.Action != "play" {
		t.Errorf("hint action = %q, want call/play", h.Action)
	}

	// respond-phase hint
	g.callTruco(1)
	g.SetResponderIdx(0)
	rh := g.GetHint()
	if rh == nil || (rh.Action != "accept" && rh.Action != "decline") {
		t.Errorf("respond hint = %+v", rh)
	}

	// no hint outside actionable phases
	g.SetPhase(TrucoPhaseTrickEnd)
	if g.GetHint() != nil {
		t.Error("expected nil hint in TrickEnd")
	}
}

// --- CPU ---

func TestTrucoCpuSelectPlayCard(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 1)
	// leading -> weakest card.
	// Hand: 3 espadas (str 10), false 4 (str 1), rey (str 7). Deliberately no
	// matador, so the 1 bastos played below is genuinely unbeatable by this hand.
	trucoSetHand(g.GetPlayer(1), NewCard(CardDesignSpade, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 13, false))
	g.SetCurrentTrick(nil)
	if idx := g.cpuSelectPlayCard(1); g.GetPlayer(1).GetCard(idx).GetValue() != 4 {
		t.Errorf("lead pick value = %d, want 4 (weakest)", g.GetPlayer(1).GetCard(idx).GetValue())
	}
	// following a mid card -> smallest winner
	g.SetCurrentTrick([]*TrickCard{{0, NewCard(CardDesignDiamond, 12, false)}}) // caballo str 6
	idx := g.cpuSelectPlayCard(1)
	got := TrucoCardStrength(g.GetPlayer(1).GetCard(idx))
	if got != 7 { // rey (13) is smallest card that beats caballo(6) in this hand
		t.Errorf("follow pick strength = %d, want 7 (smallest winner)", got)
	}
	// following an unbeatable card -> dump weakest
	g.SetCurrentTrick([]*TrickCard{{0, NewCard(CardDesignClover, 1, false)}}) // 1 bastos str 13
	idx = g.cpuSelectPlayCard(1)
	if g.GetPlayer(1).GetCard(idx).GetValue() != 4 {
		t.Errorf("dump pick value = %d, want 4 (weakest)", g.GetPlayer(1).GetCard(idx).GetValue())
	}
}

func TestTrucoCpuDecisionsReturnValid(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 1)
	trucoSetHand(g.GetPlayer(1), NewCard(CardDesignSpade, 1, false), NewCard(CardDesignClover, 1, false), NewCard(CardDesignSpade, 7, false))
	// strong hand never declines
	g.SetPhase(TrucoPhaseRespond)
	g.SetPendingLevel(TrucoLevelTruco)
	declined := false
	for i := 0; i < 200; i++ {
		if g.cpuRespondDecision(1) == "decline" {
			declined = true
		}
	}
	if declined {
		t.Error("strong hand should never decline")
	}

	// weak hand: decline appears
	trucoSetHand(g.GetPlayer(1), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignDiamond, 6, false))
	weakDeclined := false
	for i := 0; i < 200; i++ {
		if g.cpuRespondDecision(1) == "decline" {
			weakDeclined = true
			break
		}
	}
	if !weakDeclined {
		t.Error("weak hand should decline at least sometimes")
	}
}

func TestTrucoFullMatchViaCpu(t *testing.T) {
	for game := 0; game < 40; game++ {
		players := []*TrucoPlayer{NewTrucoPlayer(true), NewTrucoPlayer(false)}
		cfg := DefaultTrucoConfig()
		cfg.MatchTarget = 3 // short match for speed + branch coverage
		g := NewTruco(NewTrumpCardsBriscola(), players, cfg)
		g.Reset()

		steps := 0
		for !g.GetGameEndFlag() && steps < 5000 {
			steps++
			switch g.GetPhase() {
			case TrucoPhasePlay:
				if g.IsHumanTurn() {
					if g.CanDeclareTruco() && steps%9 == 0 {
						_ = g.DeclareTruco()
					} else {
						idxs := g.GetValidPlayIndices(0)
						if len(idxs) == 0 {
							t.Fatalf("human has no cards mid-play")
						}
						if err := g.PlayerPlay(idxs[0]); err != nil {
							t.Fatalf("PlayerPlay: %v", err)
						}
					}
				} else {
					g.CpuStep()
				}
			case TrucoPhaseRespond:
				if g.IsHumanTurn() {
					if g.CanDeclareTruco() && steps%5 == 0 {
						_ = g.DeclareTruco() // re-raise
					} else {
						_ = g.RespondTruco(steps%3 != 0)
					}
				} else {
					g.CpuStep()
				}
			case TrucoPhaseTrickEnd, TrucoPhaseHandEnd:
				g.Next()
			default:
				// GameEnd
			}
		}
		if !g.GetGameEndFlag() {
			t.Fatalf("game %d did not finish in %d steps", game, steps)
		}
		w := g.GetWinnerIdx()
		if w != 0 && w != 1 {
			t.Fatalf("game %d invalid winner %d", game, w)
		}
		if g.GetPlayerMatchPoints(w) < cfg.MatchTarget {
			t.Fatalf("game %d winner points %d < target %d", game, g.GetPlayerMatchPoints(w), cfg.MatchTarget)
		}
	}
}

func TestTrucoCpuStepNoOpWhenHuman(t *testing.T) {
	g := twoHumanTruco()
	setupBaza(g, 0)
	before := g.GetPlayer(0).GetCardsSize()
	g.CpuStep() // current player is human -> no-op
	if g.GetPlayer(0).GetCardsSize() != before {
		t.Error("CpuStep should not act on human turn")
	}
	g.SetGameEndFlag(true)
	g.CpuStep() // game ended -> no-op (no panic)
}

// --- JSON ---

func TestTrucoJSONRoundTrip(t *testing.T) {
	g := NewDefaultTruco()
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Truco
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetPlayerCnt() != TrucoPlayerCnt {
		t.Errorf("players = %d, want %d", got.GetPlayerCnt(), TrucoPlayerCnt)
	}
	if got.GetMatchTarget() != g.GetMatchTarget() {
		t.Errorf("match target = %d, want %d", got.GetMatchTarget(), g.GetMatchTarget())
	}
	if got.GetPhase() != g.GetPhase() {
		t.Errorf("phase = %d, want %d", got.GetPhase(), g.GetPhase())
	}
}

func TestTrucoUnmarshalValidation(t *testing.T) {
	bad := []string{
		`{"ps":[{}]}`,                           // wrong player count
		`{"ps":[{},{}],"ct":[{},{},{}]}`,        // too many trick cards
		`{"ps":[{},{}],"tr":[0,1,0,1]}`,         // too many trick results
		`{"ps":[{},{}],"pp":[1,2,3]}`,           // wrong match-points length
		`{"ps":[null,{}]}`,                      // nil player
		`{"ps":[{},{}],"ct":[{"pi":0}]}`,        // nil trick card payload
		`{"ps":[{},{}],"ct":[{"pi":5,"c":{}}]}`, // trick playerIdx out of bounds
		`{"ps":[{},{}],"ci":99}`,                // currentPlayerIdx out of bounds
		`{"ps":[{},{}],"ri":-2}`,                // responderIdx below sentinel
		`{"ps":[{},{}],"wi":2}`,                 // winnerIdx out of bounds
	}
	for _, s := range bad {
		var g Truco
		if err := json.Unmarshal([]byte(s), &g); err == nil {
			t.Errorf("expected error unmarshalling %s", s)
		}
	}
	// nil-safe defaults
	var g Truco
	if err := json.Unmarshal([]byte(`{"ps":[{},{}]}`), &g); err != nil {
		t.Fatalf("unmarshal valid minimal: %v", err)
	}
	if g.GetMatchTarget() != TrucoDefaultMatchTarget {
		t.Errorf("match target default = %d, want %d", g.GetMatchTarget(), TrucoDefaultMatchTarget)
	}
	if g.GetActionLog() == nil {
		t.Error("action log should default to non-nil")
	}

	// out-of-range value fields are clamped (not rejected) to their invariants
	var c Truco
	if err := json.Unmarshal([]byte(`{"ps":[{},{}],"mt":999,"hs":99,"al":-5,"pl":42}`), &c); err != nil {
		t.Fatalf("unmarshal clampable payload: %v", err)
	}
	if c.GetMatchTarget() != TrucoDefaultMatchTarget {
		t.Errorf("matchTarget clamp = %d, want %d", c.GetMatchTarget(), TrucoDefaultMatchTarget)
	}
	if c.GetHandStake() != 1 {
		t.Errorf("handStake clamp = %d, want 1", c.GetHandStake())
	}
	if c.GetAcceptedLevel() != TrucoLevelNone {
		t.Errorf("acceptedLevel clamp = %d, want %d", c.GetAcceptedLevel(), TrucoLevelNone)
	}
	if c.GetPendingLevel() != TrucoLevelNone {
		t.Errorf("pendingLevel clamp = %d, want %d", c.GetPendingLevel(), TrucoLevelNone)
	}
}
