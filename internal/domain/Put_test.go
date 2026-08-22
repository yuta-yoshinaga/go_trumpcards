//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"testing"
)

// --- helpers ---

// twoHumanPut は両プレイヤーを人間扱いにした Put を返す。
// PlayerPlay でどちらの手番も駆動できるため、CPU の乱択を介さずに
// ルールエンジンを決定論的にテストできる。
func twoHumanPut() *Put {
	players := []*PutPlayer{NewPutPlayer(true), NewPutPlayer(true)}
	g := NewPut(NewTrumpCardsBriscola(), players, DefaultPutConfig())
	g.Reset()
	return g
}

func putSetHand(p *PutPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// putPlayByCard は指定プレイヤーに、手札中の (value, design) のカードを出させる。
func putPlayByCard(t *testing.T, g *Put, playerIdx, value, design int) {
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

func TestPutCardStrength(t *testing.T) {
	// プットの序列は 3-2-A-K-Q-J-10-9-8-7-6-5-4。3 が最強、4 が最弱。
	tests := []struct {
		name  string
		card  *Card
		power int
	}{
		{"3 (top)", NewCard(CardDesignHeart, 3, false), 13},
		{"2", NewCard(CardDesignClover, 2, false), 12},
		{"ace", NewCard(CardDesignSpade, 1, false), 11},
		{"king", NewCard(CardDesignSpade, 13, false), 10},
		{"queen", NewCard(CardDesignSpade, 12, false), 9},
		{"jack", NewCard(CardDesignSpade, 11, false), 8},
		{"ten", NewCard(CardDesignSpade, 10, false), 7},
		{"nine", NewCard(CardDesignSpade, 9, false), 6},
		{"eight", NewCard(CardDesignSpade, 8, false), 5},
		{"seven", NewCard(CardDesignSpade, 7, false), 4},
		{"six", NewCard(CardDesignSpade, 6, false), 3},
		{"five", NewCard(CardDesignSpade, 5, false), 2},
		{"four (bottom)", NewCard(CardDesignSpade, 4, false), 1},
		{"nil", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PutCardStrength(tt.card); got != tt.power {
				t.Errorf("strength = %d, want %d", got, tt.power)
			}
		})
	}
}

// TestPutCardStrengthIgnoresSuit は、スートが強さに一切効かないことを見る。
//
// **クローン元のトゥルコはここがスート依存だった** (1♠ > 1♣ > 7♠ > 7♦)。
// 移植でその分岐が残っていると、同じ額面の札に上下が付いてしまい、引き分けの
// トリックが起きなくなる。
func TestPutCardStrengthIgnoresSuit(t *testing.T) {
	suits := []int{CardDesignSpade, CardDesignHeart, CardDesignDiamond, CardDesignClover}
	for v := 1; v <= 13; v++ {
		want := PutCardStrength(NewCard(suits[0], v, false))
		for _, s := range suits[1:] {
			if got := PutCardStrength(NewCard(s, v, false)); got != want {
				t.Errorf("value %d: suit %d gives %d, suit %d gives %d — suits must not matter",
					v, s, got, suits[0], want)
			}
		}
	}
}

// TestPutCardStrengthIsATotalOrderOverTheDeck は、13 階級がぴったり 1..13 に
// 並ぶことを見る。**穴や重複があると、同位でないはずの札が引き分けになる。**
func TestPutCardStrengthIsATotalOrderOverTheDeck(t *testing.T) {
	seen := map[int]int{}
	for v := 1; v <= 13; v++ {
		seen[PutCardStrength(NewCard(CardDesignSpade, v, false))]++
	}
	if len(seen) != 13 {
		t.Fatalf("13 の額面が %d 段階にしか割り当たっていない: %v", len(seen), seen)
	}
	for s := 1; s <= 13; s++ {
		if seen[s] != 1 {
			t.Errorf("強さ %d に対応する額面が %d 枚 (1 枚であるべき)", s, seen[s])
		}
	}
	// 順序そのもの: 3 > 2 > A > K
	order := []int{3, 2, 1, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4}
	for i := 1; i < len(order); i++ {
		hi := PutCardStrength(NewCard(CardDesignSpade, order[i-1], false))
		lo := PutCardStrength(NewCard(CardDesignSpade, order[i], false))
		if hi <= lo {
			t.Errorf("%d は %d より強いはず (%d <= %d)", order[i-1], order[i], hi, lo)
		}
	}
}

func TestPutLevelValue(t *testing.T) {
	// プットの宣言は 1 段だけ (None=1点 / Put=2点)。クローン元のトゥルコの
	// Reput / Vale Cuatro は存在しない。
	cases := map[int]int{
		PutLevelNone: 1,
		PutLevelPut:  2,
		-5:           1,
		99:           PutMaxLevel + 1,
	}
	for level, want := range cases {
		if got := PutLevelValue(level); got != want {
			t.Errorf("PutLevelValue(%d) = %d, want %d", level, got, want)
		}
	}
}

// --- baza winner ---

func TestPutBazaWinner(t *testing.T) {
	strong := NewCard(CardDesignSpade, 1, false) // 14
	weak := NewCard(CardDesignSpade, 4, false)   // 1
	tieA := NewCard(CardDesignHeart, 5, false)   // 2
	tieB := NewCard(CardDesignDiamond, 5, false) // 2

	if got := putBazaWinner([]*TrickCard{{0, strong}, {1, weak}}); got != 0 {
		t.Errorf("strong-vs-weak winner = %d, want 0", got)
	}
	if got := putBazaWinner([]*TrickCard{{0, weak}, {1, strong}}); got != 1 {
		t.Errorf("weak-vs-strong winner = %d, want 1", got)
	}
	if got := putBazaWinner([]*TrickCard{{0, tieA}, {1, tieB}}); got != -1 {
		t.Errorf("equal strength = %d, want -1 (parda)", got)
	}
	if got := putBazaWinner([]*TrickCard{{0, strong}}); got != -1 {
		t.Errorf("incomplete baza = %d, want -1", got)
	}
}

// --- mano resolution (pure) ---

func TestPutResolveMano(t *testing.T) {
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
			decided, winner := putResolveMano(tt.results, elder)
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

func putSetupBaza(g *Put, lead int) {
	g.SetPhase(PutPhasePlay)
	g.SetLeadPlayerIdx(lead)
	g.SetCurrentPlayerIdx(lead)
	g.SetTrickNumber(1)
	g.SetTrickResults(nil)
	g.SetCurrentTrick(nil)
	g.SetAcceptedLevel(PutLevelNone)
	g.SetHandStake(1)
}

func TestPutManoTwoStraightWins(t *testing.T) {
	g := twoHumanPut()
	putSetHand(g.GetPlayer(0), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 2, false), NewCard(CardDesignHeart, 1, false))
	putSetHand(g.GetPlayer(1), NewCard(CardDesignDiamond, 4, false), NewCard(CardDesignDiamond, 5, false), NewCard(CardDesignDiamond, 6, false))
	putSetupBaza(g, 0)

	// baza 1: 3♥ (10) vs 4♦ (1) -> p0
	putPlayByCard(t, g, 0, 3, CardDesignHeart)
	putPlayByCard(t, g, 1, 4, CardDesignDiamond)
	if g.GetPhase() != PutPhaseTrickEnd {
		t.Fatalf("after baza1 phase = %d, want TrickEnd", g.GetPhase())
	}
	g.Next()
	if g.GetPhase() != PutPhasePlay || g.GetTrickNumber() != 2 {
		t.Fatalf("baza2 start: phase=%d trick=%d", g.GetPhase(), g.GetTrickNumber())
	}
	if g.GetLeadPlayerIdx() != 0 {
		t.Errorf("baza2 lead = %d, want 0 (baza1 winner)", g.GetLeadPlayerIdx())
	}

	// baza 2: 2♥ (9) vs 5♦ (2) -> p0 wins -> mano decided
	putPlayByCard(t, g, 0, 2, CardDesignHeart)
	putPlayByCard(t, g, 1, 5, CardDesignDiamond)
	g.Next()
	if g.GetPhase() != PutPhaseHandEnd {
		t.Fatalf("after 2 wins phase = %d, want HandEnd", g.GetPhase())
	}
	if g.GetHandWinnerIdx() != 0 {
		t.Errorf("hand winner = %d, want 0", g.GetHandWinnerIdx())
	}
}

func TestPutManoTieFirstThenWin(t *testing.T) {
	g := twoHumanPut()
	putSetHand(g.GetPlayer(0), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignHeart, 3, false))
	putSetHand(g.GetPlayer(1), NewCard(CardDesignDiamond, 5, false), NewCard(CardDesignDiamond, 4, false))
	putSetupBaza(g, 0)

	// baza1: 5♥ vs 5♦ -> parda
	putPlayByCard(t, g, 0, 5, CardDesignHeart)
	putPlayByCard(t, g, 1, 5, CardDesignDiamond)
	g.Next()
	if len(g.GetTrickResults()) != 1 || g.GetTrickResults()[0] != -1 {
		t.Fatalf("baza1 result = %v, want [-1]", g.GetTrickResults())
	}
	// baza2: 3♥ (10) vs 4♦ (1) -> p0 -> tie-then-win => p0 wins mano
	putPlayByCard(t, g, 0, 3, CardDesignHeart)
	putPlayByCard(t, g, 1, 4, CardDesignDiamond)
	g.Next()
	if g.GetPhase() != PutPhaseHandEnd || g.GetHandWinnerIdx() != 0 {
		t.Fatalf("phase=%d winner=%d, want HandEnd/0", g.GetPhase(), g.GetHandWinnerIdx())
	}
}

func TestPutManoAllPardaGoesToElder(t *testing.T) {
	g := twoHumanPut()
	g.SetManoIdx(1)
	putSetHand(g.GetPlayer(0), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignHeart, 6, false))
	putSetHand(g.GetPlayer(1), NewCard(CardDesignDiamond, 4, false), NewCard(CardDesignDiamond, 5, false), NewCard(CardDesignDiamond, 6, false))
	putSetupBaza(g, 0)

	for _, v := range []int{4, 5, 6} {
		putPlayByCard(t, g, 0, v, CardDesignHeart)
		putPlayByCard(t, g, 1, v, CardDesignDiamond)
		g.Next()
	}
	if g.GetPhase() != PutPhaseHandEnd {
		t.Fatalf("phase = %d, want HandEnd", g.GetPhase())
	}
	if g.GetHandWinnerIdx() != 1 {
		t.Errorf("all-parda winner = %d, want elder (1)", g.GetHandWinnerIdx())
	}
}

func TestPutManoSplitThenThird(t *testing.T) {
	g := twoHumanPut()
	putSetHand(g.GetPlayer(0), NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignSpade, 1, false))
	putSetHand(g.GetPlayer(1), NewCard(CardDesignDiamond, 4, false), NewCard(CardDesignClover, 3, false), NewCard(CardDesignDiamond, 6, false))
	putSetupBaza(g, 0)

	// baza1: 3♥(10) vs 4♦(1) -> p0
	putPlayByCard(t, g, 0, 3, CardDesignHeart)
	putPlayByCard(t, g, 1, 4, CardDesignDiamond)
	g.Next()
	// baza2: lead is p0; 4♥(1) vs 3♣(10) -> p1 -> split
	putPlayByCard(t, g, 0, 4, CardDesignHeart)
	putPlayByCard(t, g, 1, 3, CardDesignClover)
	g.Next()
	if g.GetPhase() != PutPhasePlay || g.GetTrickNumber() != 3 {
		t.Fatalf("expected baza3: phase=%d trick=%d", g.GetPhase(), g.GetTrickNumber())
	}
	// baza3: 1♠(14) vs 6♦(3) -> p0 wins the mano
	putPlayByCard(t, g, 0, 1, CardDesignSpade)
	putPlayByCard(t, g, 1, 6, CardDesignDiamond)
	g.Next()
	if g.GetPhase() != PutPhaseHandEnd || g.GetHandWinnerIdx() != 0 {
		t.Fatalf("phase=%d winner=%d, want HandEnd/0", g.GetPhase(), g.GetHandWinnerIdx())
	}
}

// --- put escalation (unit on internal call/respond) ---

func TestPutCallAndAccept(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	g.callPut(1)
	if g.GetPhase() != PutPhaseRespond {
		t.Fatalf("phase = %d, want Respond", g.GetPhase())
	}
	if g.GetResponderIdx() != 0 || g.GetPutCallerIdx() != 1 || g.GetPendingLevel() != PutLevelPut {
		t.Fatalf("responder=%d caller=%d pending=%d", g.GetResponderIdx(), g.GetPutCallerIdx(), g.GetPendingLevel())
	}
	g.respond(0, true)
	if g.GetPhase() != PutPhasePlay {
		t.Fatalf("phase after accept = %d, want Play", g.GetPhase())
	}
	if g.GetAcceptedLevel() != PutLevelPut || g.GetHandStake() != 2 {
		t.Errorf("acceptedLevel=%d handStake=%d, want 1/2", g.GetAcceptedLevel(), g.GetHandStake())
	}
	if g.GetResponderIdx() != -1 || g.GetPutCallerIdx() != -1 || g.GetPendingLevel() != 0 {
		t.Errorf("pending state not cleared after accept")
	}
}

func TestPutCallAndDeclineAwardsPriorStake(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	// no put accepted yet -> decline awards 1 pt to caller
	g.callPut(1)
	g.respond(0, false)
	if g.GetPhase() != PutPhaseHandEnd {
		t.Fatalf("phase = %d, want HandEnd", g.GetPhase())
	}
	if g.GetHandWinnerIdx() != 1 {
		t.Errorf("hand winner = %d, want caller 1", g.GetHandWinnerIdx())
	}
	if g.GetHandStake() != 1 {
		t.Errorf("decline at base awards %d, want 1", g.GetHandStake())
	}
}

func TestPutDeclineAfterAcceptedPut(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	g.callPut(0)        // p0 calls Put
	g.respond(1, true)  // p1 accepts -> stake 2
	g.callPut(0)        // p0 declares Put (pending=1)
	g.respond(1, false) // p1 declines -> p0 wins at prior stake (2)
	if g.GetHandWinnerIdx() != 0 {
		t.Errorf("hand winner = %d, want 0", g.GetHandWinnerIdx())
	}
	if g.GetHandStake() != 2 {
		t.Errorf("decline after accepted Put awards %d, want 2", g.GetHandStake())
	}
}

func TestPutReRaiseChain(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	g.callPut(0) // p0 Put (pending 1, responder 1)
	if g.GetResponderIdx() != 1 {
		t.Fatalf("responder = %d, want 1", g.GetResponderIdx())
	}
	// **プットに再宣言は無い。** 「Put」と言われた側は受けるか降りるかの
	// 二択で、賭けを引き上げ返す手はない (クローン元のトゥルコは Retruco へ
	// 伸ばせる)。応答フェーズで相手が更に宣言できないことを見る。
	if g.canDeclare(1) {
		t.Error("プットでは応答側が賭けを引き上げ返せない")
	}
	g.respond(1, true)
	if g.GetAcceptedLevel() != PutLevelPut || g.GetHandStake() != 2 {
		t.Errorf("acceptedLevel=%d stake=%d, want %d/2", g.GetAcceptedLevel(), g.GetHandStake(), PutLevelPut)
	}
}

func TestPutCanDeclareMaxLevel(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	g.SetAcceptedLevel(PutMaxLevel)
	if g.canDeclare(0) {
		t.Error("すでに Put が受諾済みなら、それ以上は宣言できない")
	}
	// pending at max in respond phase -> cannot re-raise
	g.SetPhase(PutPhaseRespond)
	g.SetPendingLevel(PutMaxLevel)
	g.SetResponderIdx(0)
	if g.canDeclare(0) {
		t.Error("canDeclare with pending at max should be false")
	}
}

func TestPutDeclareGuards(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	// human idx 0 leads -> declareActor returns 0 -> allowed
	if !g.CanDeclarePut() {
		t.Error("human should be able to declare at start of own turn")
	}
	if err := g.DeclarePut(); err != nil {
		t.Fatalf("DeclarePut: %v", err)
	}
	// now responder is 1 (not human-as-0); CanDeclarePut should be false for actor!=0
	if g.GetResponderIdx() != 1 {
		t.Fatalf("responder = %d", g.GetResponderIdx())
	}
}

// --- match accumulation ---

func TestPutMatchEndOnTarget(t *testing.T) {
	g := twoHumanPut()
	g.SetPlayerMatchPoints(0, 14)
	g.SetHandWinnerIdx(0)
	g.SetHandStake(2)
	g.SetPhase(PutPhaseHandEnd)
	g.Next()
	if !g.GetGameEndFlag() {
		t.Fatal("expected game end after reaching target")
	}
	if g.GetWinnerIdx() != 0 {
		t.Errorf("winner = %d, want 0", g.GetWinnerIdx())
	}
	if g.GetPhase() != PutPhaseGameEnd {
		t.Errorf("phase = %d, want GameEnd", g.GetPhase())
	}
	if g.GetPlayerMatchPoints(0) != 16 {
		t.Errorf("points = %d, want 16", g.GetPlayerMatchPoints(0))
	}
}

func TestPutMatchContinuesBelowTarget(t *testing.T) {
	g := twoHumanPut()
	startHand := g.GetHandNumber()
	g.SetPlayerMatchPoints(0, 5)
	g.SetHandWinnerIdx(0)
	g.SetHandStake(2)
	g.SetPhase(PutPhaseHandEnd)
	g.Next()
	if g.GetGameEndFlag() {
		t.Fatal("game should continue below target")
	}
	if g.GetPhase() != PutPhasePlay {
		t.Errorf("phase = %d, want Play (new hand)", g.GetPhase())
	}
	if g.GetHandNumber() != startHand+1 {
		t.Errorf("hand number = %d, want %d", g.GetHandNumber(), startHand+1)
	}
	if g.GetPlayer(0).GetCardsSize() != PutHandSize {
		t.Errorf("new hand size = %d, want %d", g.GetPlayer(0).GetCardsSize(), PutHandSize)
	}
	// dealer/mano should have alternated
	if g.GetDealerIdx() == 0 {
		t.Error("dealer should have flipped from 0")
	}
}

// --- error / guard paths ---

func TestPutPlayerPlayErrors(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	putSetHand(g.GetPlayer(0), NewCard(CardDesignSpade, 1, false))

	g.SetCurrentPlayerIdx(0)
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected error for out-of-range index")
	}
	g.SetPhase(PutPhaseTrickEnd)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected ErrWrongPhase")
	}
	g.SetPhase(PutPhasePlay)
	g.SetGameEndFlag(true)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected ErrGameEnded")
	}
}

func TestPutPlayerPlayNotHumanTurn(t *testing.T) {
	players := []*PutPlayer{NewPutPlayer(true), NewPutPlayer(false)}
	g := NewPut(NewTrumpCardsBriscola(), players, DefaultPutConfig())
	g.Reset()
	putSetupBaza(g, 1) // CPU leads
	putSetHand(g.GetPlayer(1), NewCard(CardDesignSpade, 1, false))
	g.SetCurrentPlayerIdx(1)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected ErrNotHumanTurn when CPU is current")
	}
}

func TestPutRespondGuards(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	if err := g.RespondPut(true); err == nil {
		t.Error("expected ErrWrongPhase when not in Respond")
	}
	g.callPut(1) // responder = 0
	g.SetResponderIdx(1)
	if err := g.RespondPut(true); err == nil {
		t.Error("expected ErrNotHumanTurn when responder != 0")
	}
	g.SetGameEndFlag(true)
	if err := g.RespondPut(true); err == nil {
		t.Error("expected ErrGameEnded")
	}
}

func TestPutDeclareWrongPhase(t *testing.T) {
	g := twoHumanPut()
	g.SetPhase(PutPhaseTrickEnd)
	if err := g.DeclarePut(); err == nil {
		t.Error("expected error declaring in TrickEnd")
	}
}

// --- getters ---

func TestPutGettersAndValidIndices(t *testing.T) {
	g := twoHumanPut()
	if g.GetPlayerCnt() != 2 {
		t.Errorf("player count = %d, want 2", g.GetPlayerCnt())
	}
	if g.GetPlayer(99) != nil {
		t.Error("out-of-range player should be nil")
	}
	if g.GetMatchTarget() != PutDefaultMatchTarget {
		t.Errorf("match target = %d, want %d", g.GetMatchTarget(), PutDefaultMatchTarget)
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

func TestPutHint(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	putSetHand(g.GetPlayer(0), NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 4, false))
	g.SetCurrentPlayerIdx(0)
	h := g.GetHint()
	if h == nil {
		t.Fatal("expected hint during human play turn")
	}
	if h.Action != "call" && h.Action != "play" {
		t.Errorf("hint action = %q, want call/play", h.Action)
	}

	// respond-phase hint
	g.callPut(1)
	g.SetResponderIdx(0)
	rh := g.GetHint()
	if rh == nil || (rh.Action != "accept" && rh.Action != "decline") {
		t.Errorf("respond hint = %+v", rh)
	}

	// no hint outside actionable phases
	g.SetPhase(PutPhaseTrickEnd)
	if g.GetHint() != nil {
		t.Error("expected nil hint in TrickEnd")
	}
}

// --- CPU ---

func TestPutCpuSelectPlayCard(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 1)

	// 手札: 3♠(強さ13) / 4♥(1) / K♥(10)。プットの序列は 3-2-A-K-...-4 で
	// スートは効かないので、強さは額面だけで決まる。
	putSetHand(g.GetPlayer(1), NewCard(CardDesignSpade, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 13, false))

	// 先手 -> 一番弱い札を出す。
	g.SetCurrentTrick(nil)
	if idx := g.cpuSelectPlayCard(1); g.GetPlayer(1).GetCard(idx).GetValue() != 4 {
		t.Errorf("lead pick value = %d, want 4 (weakest)", g.GetPlayer(1).GetCard(idx).GetValue())
	}

	// Q(強さ9) に後手 -> 勝てる中で一番小さい札 = K(強さ10)。
	// 3(13) でも勝てるが、それは温存する。
	g.SetCurrentTrick([]*TrickCard{{0, NewCard(CardDesignDiamond, 12, false)}})
	idx := g.cpuSelectPlayCard(1)
	if got := PutCardStrength(g.GetPlayer(1).GetCard(idx)); got != 10 {
		t.Errorf("follow pick strength = %d, want 10 (smallest winner)", got)
	}

	// **勝てない場面では一番弱い札を捨てる。**
	// プットでは 3 が最強なので、「絶対に勝てない」を作るには相手が 3 を出し、
	// こちらの手札に 3 が無い必要がある。クローン元 (トゥルコ) はスート付きの
	// 特別札があったため 1♣ が絶対札だったが、プットではただの A で、
	// 手札の 3 に負ける。
	putSetHand(g.GetPlayer(1), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 13, false), NewCard(CardDesignSpade, 5, false))
	g.SetCurrentTrick([]*TrickCard{{0, NewCard(CardDesignDiamond, 3, false)}})
	idx = g.cpuSelectPlayCard(1)
	if v := g.GetPlayer(1).GetCard(idx).GetValue(); v != 4 {
		t.Errorf("dump pick value = %d, want 4 (weakest)", v)
	}
}

func TestPutCpuDecisionsReturnValid(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 1)
	putSetHand(g.GetPlayer(1), NewCard(CardDesignSpade, 1, false), NewCard(CardDesignClover, 1, false), NewCard(CardDesignSpade, 7, false))
	// strong hand never declines
	g.SetPhase(PutPhaseRespond)
	g.SetPendingLevel(PutLevelPut)
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
	putSetHand(g.GetPlayer(1), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignDiamond, 6, false))
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

func TestPutFullMatchViaCpu(t *testing.T) {
	for game := 0; game < 40; game++ {
		players := []*PutPlayer{NewPutPlayer(true), NewPutPlayer(false)}
		cfg := DefaultPutConfig()
		cfg.MatchTarget = 3 // short match for speed + branch coverage
		g := NewPut(NewTrumpCardsBriscola(), players, cfg)
		g.Reset()

		steps := 0
		for !g.GetGameEndFlag() && steps < 5000 {
			steps++
			switch g.GetPhase() {
			case PutPhasePlay:
				if g.IsHumanTurn() {
					if g.CanDeclarePut() && steps%9 == 0 {
						_ = g.DeclarePut()
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
			case PutPhaseRespond:
				if g.IsHumanTurn() {
					if g.CanDeclarePut() && steps%5 == 0 {
						_ = g.DeclarePut() // re-raise
					} else {
						_ = g.RespondPut(steps%3 != 0)
					}
				} else {
					g.CpuStep()
				}
			case PutPhaseTrickEnd, PutPhaseHandEnd:
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

func TestPutCpuStepNoOpWhenHuman(t *testing.T) {
	g := twoHumanPut()
	putSetupBaza(g, 0)
	before := g.GetPlayer(0).GetCardsSize()
	g.CpuStep() // current player is human -> no-op
	if g.GetPlayer(0).GetCardsSize() != before {
		t.Error("CpuStep should not act on human turn")
	}
	g.SetGameEndFlag(true)
	g.CpuStep() // game ended -> no-op (no panic)
}

// --- JSON ---

func TestPutJSONRoundTrip(t *testing.T) {
	g := NewDefaultPut()
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Put
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetPlayerCnt() != PutPlayerCnt {
		t.Errorf("players = %d, want %d", got.GetPlayerCnt(), PutPlayerCnt)
	}
	if got.GetMatchTarget() != g.GetMatchTarget() {
		t.Errorf("match target = %d, want %d", got.GetMatchTarget(), g.GetMatchTarget())
	}
	if got.GetPhase() != g.GetPhase() {
		t.Errorf("phase = %d, want %d", got.GetPhase(), g.GetPhase())
	}
}

func TestPutUnmarshalValidation(t *testing.T) {
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
		var g Put
		if err := json.Unmarshal([]byte(s), &g); err == nil {
			t.Errorf("expected error unmarshalling %s", s)
		}
	}
	// nil-safe defaults
	var g Put
	if err := json.Unmarshal([]byte(`{"ps":[{},{}]}`), &g); err != nil {
		t.Fatalf("unmarshal valid minimal: %v", err)
	}
	if g.GetMatchTarget() != PutDefaultMatchTarget {
		t.Errorf("match target default = %d, want %d", g.GetMatchTarget(), PutDefaultMatchTarget)
	}
	if g.GetActionLog() == nil {
		t.Error("action log should default to non-nil")
	}

	// out-of-range value fields are clamped (not rejected) to their invariants
	var c Put
	if err := json.Unmarshal([]byte(`{"ps":[{},{}],"mt":999,"hs":99,"al":-5,"pl":42}`), &c); err != nil {
		t.Fatalf("unmarshal clampable payload: %v", err)
	}
	if c.GetMatchTarget() != PutDefaultMatchTarget {
		t.Errorf("matchTarget clamp = %d, want %d", c.GetMatchTarget(), PutDefaultMatchTarget)
	}
	if c.GetHandStake() != 1 {
		t.Errorf("handStake clamp = %d, want 1", c.GetHandStake())
	}
	if c.GetAcceptedLevel() != PutLevelNone {
		t.Errorf("acceptedLevel clamp = %d, want %d", c.GetAcceptedLevel(), PutLevelNone)
	}
	if c.GetPendingLevel() != PutLevelNone {
		t.Errorf("pendingLevel clamp = %d, want %d", c.GetPendingLevel(), PutLevelNone)
	}
}
