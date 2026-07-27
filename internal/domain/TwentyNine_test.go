//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func tnCard(design, value int) *Card { return NewCard(design, value, false) }

func newTnGame(human bool) *TwentyNine {
	players := make([]*TwentyNinePlayer, TwentyNinePlayerCnt)
	players[0] = NewTwentyNinePlayer(human)
	for i := 1; i < TwentyNinePlayerCnt; i++ {
		players[i] = NewTwentyNinePlayer(false)
	}
	return NewTwentyNine(NewTrumpCardsBelote(), players, DefaultTwentyNineConfig())
}

func newTnAllHuman() *TwentyNine {
	players := make([]*TwentyNinePlayer, TwentyNinePlayerCnt)
	for i := 0; i < TwentyNinePlayerCnt; i++ {
		players[i] = NewTwentyNinePlayer(true)
	}
	return NewTwentyNine(NewTrumpCardsBelote(), players, DefaultTwentyNineConfig())
}

func tnSetHand(p *TwentyNinePlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestTwentyNineConfig_Validate(t *testing.T) {
	if err := DefaultTwentyNineConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (TwentyNineConfig{CpuDifficulty: 9, TargetPoints: 6}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (TwentyNineConfig{CpuDifficulty: TwentyNineCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestTwentyNine_ResetDealsEightAndBids(t *testing.T) {
	g := newTnGame(true)
	g.Reset()
	if g.GetPhase() != TwentyNinePhaseBid {
		t.Errorf("phase = %d, want Bid", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetCardsSize() != TwentyNineHandSize {
			t.Errorf("player %d dealt %d cards, want %d", i, g.GetPlayer(i).GetCardsSize(), TwentyNineHandSize)
		}
	}
}

func TestTwentyNine_StrengthAndPoints(t *testing.T) {
	order := []int{11, 9, 1, 10, 13, 12, 8, 7} // J>9>A>10>K>Q>8>7
	for i := 1; i < len(order); i++ {
		if twentyNineStrength(order[i-1]) <= twentyNineStrength(order[i]) {
			t.Errorf("strength(%d) should beat strength(%d)", order[i-1], order[i])
		}
	}
	cases := map[int]int{11: 3, 9: 2, 1: 1, 10: 1, 13: 0, 12: 0, 8: 0, 7: 0}
	total := 0
	for v, want := range cases {
		if got := twentyNineCardPoints(tnCard(CardDesignSpade, v)); got != want {
			t.Errorf("points(%d) = %d, want %d", v, got, want)
		}
		total += want
	}
	if total*4 != 28 {
		t.Errorf("deck card points = %d, want 28", total*4)
	}
}

func TestTwentyNine_BiddingResolvesHighestDeclarer(t *testing.T) {
	g := newTnAllHuman()
	g.Reset() // dealer 0, forehand P1 bids first
	// P1 16, P2 20, P3 Pass, P0 Pass -> declarer P2 / 20, hidden trump set.
	if err := g.PlayerBid(TwentyNineBidSixteen); err != nil {
		t.Fatalf("P1 16 err: %v", err)
	}
	if err := g.PlayerBid(TwentyNineBidTwenty); err != nil {
		t.Fatalf("P2 20 err: %v", err)
	}
	if err := g.PlayerBid(TwentyNineBidPass); err != nil {
		t.Fatalf("P3 pass err: %v", err)
	}
	if err := g.PlayerBid(TwentyNineBidPass); err != nil {
		t.Fatalf("P0 pass err: %v", err)
	}
	if g.GetDeclarerIdx() != 2 || g.GetContract() != TwentyNineBidTwenty {
		t.Errorf("declarer=%d contract=%d, want 2 / 20", g.GetDeclarerIdx(), g.GetContract())
	}
	if g.GetTrumpSuit() < 1 || g.GetTrumpSuit() > 4 {
		t.Errorf("trump = %d, want 1-4", g.GetTrumpSuit())
	}
	if g.GetTrumpRevealed() {
		t.Error("trump should start hidden")
	}
	if g.GetPhase() != TwentyNinePhasePlay {
		t.Errorf("phase after bidding = %d, want Play", g.GetPhase())
	}
}

func TestTwentyNine_CannotUnderbid(t *testing.T) {
	g := newTnAllHuman()
	g.Reset()
	if err := g.PlayerBid(TwentyNineBidTwenty); err != nil {
		t.Fatalf("20 err: %v", err)
	}
	if err := g.PlayerBid(TwentyNineBidSixteen); err == nil {
		t.Error("expected under-bid rejection")
	}
}

func TestTwentyNine_TrickWinnerJHighAndTrump(t *testing.T) {
	g := newTnGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// Lead A♣; J♣ beats it (J high in plain suit); a low trump 7♦ beats all plain.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: tnCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: tnCard(CardDesignClover, 11)},
		{PlayerIdx: 2, Card: tnCard(CardDesignDiamond, 7)},
		{PlayerIdx: 3, Card: tnCard(CardDesignClover, 13)},
	})
	if w := g.trickWinner(); w != 2 {
		t.Errorf("trick winner = %d, want 2 (trump beats J-high plain)", w)
	}
}

func TestTwentyNine_HiddenTrumpRevealsOnOffSuit(t *testing.T) {
	g := newTnGame(true)
	g.SetPhase(TwentyNinePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetTrumpRevealed(false)
	g.SetCurrentPlayerIdx(0)
	// Lead was a club (by p3); human is void in clubs, plays a diamond (trump) -> reveal.
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: tnCard(CardDesignClover, 1)}})
	tnSetHand(g.GetPlayer(0), tnCard(CardDesignDiamond, 11), tnCard(CardDesignHeart, 7))
	if err := g.PlayerPlay(0); err != nil { // play J♦ (off-suit) — legal (void in clubs)
		t.Fatalf("off-suit play err: %v", err)
	}
	if !g.GetTrumpRevealed() {
		t.Error("trump should be revealed after an off-suit play")
	}
}

func TestTwentyNine_MustFollow(t *testing.T) {
	g := newTnGame(true)
	g.SetPhase(TwentyNinePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: tnCard(CardDesignClover, 1)}})
	tnSetHand(g.GetPlayer(0), tnCard(CardDesignClover, 13), tnCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(1); err == nil {
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestTwentyNine_ResolveTrickPointsAndLastBonus(t *testing.T) {
	g := newTnGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(TwentyNinePhaseTrickEnd)
	g.SetTrickNumber(TwentyNineTrickCount) // final trick -> +1
	// J♣(3) + 9♣(2) + A♣(1) + 10♣(1) = 7; J♣ wins for team 0 (p0). +1 last = 8.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: tnCard(CardDesignClover, 11)},
		{PlayerIdx: 1, Card: tnCard(CardDesignClover, 9)},
		{PlayerIdx: 2, Card: tnCard(CardDesignClover, 1)},
		{PlayerIdx: 3, Card: tnCard(CardDesignClover, 10)},
	})
	g.ResolveTrick()
	if pts := g.GetRoundTeamPoints(); pts[0] != 8 {
		t.Errorf("team 0 round points = %d, want 8 (7 + 1 last)", pts[0])
	}
}

func TestTwentyNine_ScoreRoundMadeAndSet(t *testing.T) {
	// Made: bid 16, team got 18 -> bid team +1 game point.
	g := newTnGame(false)
	g.SetPhase(TwentyNinePhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(TwentyNineBidSixteen)
	g.SetRoundTeamPoints([TwentyNineTeamCnt]int{18, 11})
	g.ScoreRound()
	if sc := g.GetTeamScores(); sc[0] != 1 || sc[1] != 0 {
		t.Errorf("scores = %v, want [1 0] on made bid", sc)
	}
	// Set: bid 20, team got 12 -> other team +1.
	g2 := newTnGame(false)
	g2.SetPhase(TwentyNinePhaseRoundEnd)
	g2.SetDeclarerIdx(0)
	g2.SetContract(TwentyNineBidTwenty)
	g2.SetRoundTeamPoints([TwentyNineTeamCnt]int{12, 17})
	g2.ScoreRound()
	if sc := g2.GetTeamScores(); sc[0] != 0 || sc[1] != 1 {
		t.Errorf("scores = %v, want [0 1] on set bid", sc)
	}
}

func TestTwentyNine_ScoreRoundEndsMatch(t *testing.T) {
	g := newTnGame(false)
	g.SetPhase(TwentyNinePhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(TwentyNineBidSixteen)
	g.SetTeamScores([TwentyNineTeamCnt]int{5, 2})
	g.SetRoundTeamPoints([TwentyNineTeamCnt]int{20, 9}) // made -> +1 -> 6 >= 6
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("expected match end with team 0, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerTeam())
	}
}

func TestTwentyNine_NextRound(t *testing.T) {
	g := newTnGame(false)
	g.SetPhase(TwentyNinePhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetPhase() != TwentyNinePhaseBid {
		t.Errorf("after NextRound: round=%d phase=%d, want 2 / Bid", g.GetRoundNumber(), g.GetPhase())
	}
}

func TestTwentyNine_HintAndPlayable(t *testing.T) {
	g := newTnGame(true)
	g.SetPhase(TwentyNinePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	tnSetHand(g.GetPlayer(0), tnCard(CardDesignClover, 7), tnCard(CardDesignClover, 11))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestTwentyNine_CpuFullMatchProgresses(t *testing.T) {
	g := newTnGame(false) // all CPU
	g.Reset()
	for guard := 0; guard < 6000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case TwentyNinePhaseBid:
			g.CpuBid()
		case TwentyNinePhasePlay:
			g.CpuPlay()
			if g.GetPhase() == TwentyNinePhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == TwentyNinePhaseTrickEnd {
					g.NextTrick()
				}
			}
		case TwentyNinePhaseRoundEnd:
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

func TestTwentyNine_PlayerPlayErrors(t *testing.T) {
	g := newTnGame(true)
	g.SetPhase(TwentyNinePhasePlay)
	g.SetCurrentPlayerIdx(0)
	tnSetHand(g.GetPlayer(0), tnCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(TwentyNinePhaseBid)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestTwentyNine_JSONRoundTrip(t *testing.T) {
	g := newTnGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 TwentyNine
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() || g2.GetPhase() != g.GetPhase() {
		t.Error("round-trip mismatch")
	}
}

func TestTwentyNine_UnmarshalErrors(t *testing.T) {
	var g TwentyNine
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

func TestTwentyNine_UnmarshalRejectsInvalidBid(t *testing.T) {
	base := newTnGame(true)
	base.Reset()
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	// Inject an out-of-range bid value (99 is not a valid TwentyNineBid constant).
	tampered := strings.Replace(string(data), `"bd":[0,0,0,0]`, `"bd":[99,0,0,0]`, 1)
	if tampered == string(data) {
		t.Fatalf("test setup: expected to find the default bid array in %s", data)
	}
	var g TwentyNine
	if err := g.UnmarshalJSON([]byte(tampered)); err == nil {
		t.Error("expected invalid-bid error")
	}
	// A tampered contract value must also be rejected.
	tampered2 := strings.Replace(string(data), `"co":0`, `"co":7`, 1)
	if tampered2 != string(data) {
		var g2 TwentyNine
		if err := g2.UnmarshalJSON([]byte(tampered2)); err == nil {
			t.Error("expected invalid-contract error")
		}
	}
}
