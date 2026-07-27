//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func ffCard(design, value int) *Card { return NewCard(design, value, false) }

func newFfGame(human bool) *FortyFives {
	players := make([]*FortyFivesPlayer, FortyFivesPlayerCnt)
	players[0] = NewFortyFivesPlayer(human)
	for i := 1; i < FortyFivesPlayerCnt; i++ {
		players[i] = NewFortyFivesPlayer(false)
	}
	return NewFortyFives(NewTrumpCards(0), players, DefaultFortyFivesConfig())
}

func newFfAllHuman() *FortyFives {
	players := make([]*FortyFivesPlayer, FortyFivesPlayerCnt)
	for i := 0; i < FortyFivesPlayerCnt; i++ {
		players[i] = NewFortyFivesPlayer(true)
	}
	return NewFortyFives(NewTrumpCards(0), players, DefaultFortyFivesConfig())
}

func ffSetHand(p *FortyFivesPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestFortyFivesConfig_Validate(t *testing.T) {
	if err := DefaultFortyFivesConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (FortyFivesConfig{CpuDifficulty: 9, TargetPoints: 45}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (FortyFivesConfig{CpuDifficulty: FortyFivesCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestFortyFives_ResetDealsFiveAndBids(t *testing.T) {
	g := newFfGame(true)
	g.Reset()
	if g.GetPhase() != FortyFivesPhaseBid {
		t.Errorf("phase = %d, want Bid", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetCardsSize() != FortyFivesHandSize {
			t.Errorf("player %d dealt %d cards, want %d", i, g.GetPlayer(i).GetCardsSize(), FortyFivesHandSize)
		}
	}
}

func TestFortyFives_BiddingResolvesHighestDeclarer(t *testing.T) {
	g := newFfAllHuman()
	g.Reset() // dealer 0, forehand P1 bids first
	// P1 15, P2 20 (outbids), P3 Pass, P0 Pass -> declarer P2 / 20.
	if err := g.PlayerBid(FortyFivesBidFifteen); err != nil {
		t.Fatalf("P1 15 err: %v", err)
	}
	if err := g.PlayerBid(FortyFivesBidTwenty); err != nil {
		t.Fatalf("P2 20 err: %v", err)
	}
	if err := g.PlayerBid(FortyFivesBidPass); err != nil {
		t.Fatalf("P3 pass err: %v", err)
	}
	if err := g.PlayerBid(FortyFivesBidPass); err != nil {
		t.Fatalf("P0 pass err: %v", err)
	}
	if g.GetDeclarerIdx() != 2 || g.GetContract() != FortyFivesBidTwenty {
		t.Errorf("declarer=%d contract=%d, want 2 / 20", g.GetDeclarerIdx(), g.GetContract())
	}
	if g.GetPhase() != FortyFivesPhasePlay {
		t.Errorf("phase after bidding = %d, want Play", g.GetPhase())
	}
}

func TestFortyFives_CannotUnderbid(t *testing.T) {
	g := newFfAllHuman()
	g.Reset()
	if err := g.PlayerBid(FortyFivesBidTwenty); err != nil {
		t.Fatalf("20 err: %v", err)
	}
	if err := g.PlayerBid(FortyFivesBidFifteen); err == nil {
		t.Error("expected under-bid rejection")
	}
}

func TestFortyFives_AllPassVoidsRound(t *testing.T) {
	g := newFfAllHuman()
	g.Reset()
	for i := 0; i < FortyFivesPlayerCnt; i++ {
		if err := g.PlayerBid(FortyFivesBidPass); err != nil {
			t.Fatalf("pass %d err: %v", i, err)
		}
	}
	if g.GetDeclarerIdx() != -1 || g.GetPhase() != FortyFivesPhaseRoundEnd {
		t.Errorf("all-pass: declarer=%d phase=%d, want -1 / RoundEnd", g.GetDeclarerIdx(), g.GetPhase())
	}
}

func TestFortyFives_FixedTrumpRankAndHeartAce(t *testing.T) {
	g := newFfGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// 5♦ > J♦ > ♥A > A♦ > K♦
	cards := []*Card{
		ffCard(CardDesignDiamond, 5), ffCard(CardDesignDiamond, 11),
		ffCard(CardDesignHeart, 1), ffCard(CardDesignDiamond, 1), ffCard(CardDesignDiamond, 13),
	}
	for i := 1; i < len(cards); i++ {
		if g.fortyFivesRank(cards[i-1]) <= g.fortyFivesRank(cards[i]) {
			t.Errorf("rank(%v) should beat rank(%v)", cards[i-1], cards[i])
		}
	}
	// ♥A is a top trump even when hearts isn't trump.
	g.SetTrumpSuit(CardDesignSpade)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: ffCard(CardDesignSpade, 7)},
		{PlayerIdx: 1, Card: ffCard(CardDesignHeart, 1)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1 (♥A top trump)", w)
	}
}

func TestFortyFives_RenegingTopTrump(t *testing.T) {
	g := newFfGame(true)
	g.SetPhase(FortyFivesPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: ffCard(CardDesignDiamond, 7)}})
	// Only trump is the 5♦ (top trump) -> may renege and play a club.
	ffSetHand(g.GetPlayer(0), ffCard(CardDesignDiamond, 5), ffCard(CardDesignClover, 9))
	if err := g.PlayerPlay(1); err != nil {
		t.Fatalf("reneging a top trump should be allowed, got: %v", err)
	}
}

func TestFortyFives_ResolveTrickFivePointsPerTrick(t *testing.T) {
	g := newFfGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(FortyFivesPhaseTrickEnd)
	g.SetTrickNumber(1)
	// p0 (team 0) wins with the best trump.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: ffCard(CardDesignDiamond, 5)},
		{PlayerIdx: 1, Card: ffCard(CardDesignClover, 9)},
		{PlayerIdx: 2, Card: ffCard(CardDesignClover, 8)},
		{PlayerIdx: 3, Card: ffCard(CardDesignClover, 7)},
	})
	g.ResolveTrick()
	if pts := g.GetRoundTeamPoints(); pts[0] != FortyFivesPointsPerTrick {
		t.Errorf("team 0 round points = %d, want %d", pts[0], FortyFivesPointsPerTrick)
	}
}

func TestFortyFives_ScoreRoundMadeAndFailed(t *testing.T) {
	// Made: bid 15, team got 20 -> +20.
	g := newFfGame(false)
	g.SetPhase(FortyFivesPhaseRoundEnd)
	g.SetDeclarerIdx(0) // team 0
	g.SetContract(FortyFivesBidFifteen)
	g.SetRoundTeamPoints([FortyFivesTeamCnt]int{20, 5})
	g.ScoreRound()
	sc := g.GetTeamScores()
	if sc[0] != 20 || sc[1] != 5 {
		t.Errorf("scores = %v, want [20 5] on made bid", sc)
	}
	// Failed: bid 20, team got 10 -> -20; other team +their points.
	g2 := newFfGame(false)
	g2.SetPhase(FortyFivesPhaseRoundEnd)
	g2.SetDeclarerIdx(0)
	g2.SetContract(FortyFivesBidTwenty)
	g2.SetRoundTeamPoints([FortyFivesTeamCnt]int{10, 15})
	g2.ScoreRound()
	sc = g2.GetTeamScores()
	if sc[0] != -20 || sc[1] != 15 {
		t.Errorf("scores = %v, want [-20 15] on failed bid", sc)
	}
}

func TestFortyFives_JinkSweepWins(t *testing.T) {
	g := newFfGame(false)
	g.SetPhase(FortyFivesPhaseRoundEnd)
	g.SetDeclarerIdx(0) // team 0
	g.SetContract(FortyFivesBidTwentyFive)
	// Team 0 (seats 0,2) won all 5 tricks.
	g.GetPlayer(0).IncRoundTricks()
	g.GetPlayer(0).IncRoundTricks()
	g.GetPlayer(0).IncRoundTricks()
	g.GetPlayer(2).IncRoundTricks()
	g.GetPlayer(2).IncRoundTricks()
	g.SetRoundTeamPoints([FortyFivesTeamCnt]int{25, 0})
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("expected Jink win for team 0, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerTeam())
	}
}

func TestFortyFives_ScoreRoundEndsMatch(t *testing.T) {
	g := newFfGame(false)
	g.SetPhase(FortyFivesPhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(FortyFivesBidFifteen)
	g.SetTeamScores([FortyFivesTeamCnt]int{40, 10})
	g.SetRoundTeamPoints([FortyFivesTeamCnt]int{15, 10}) // +15 -> 55 >= 45
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerTeam() != 0 {
		t.Errorf("expected match end with team 0, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerTeam())
	}
}

func TestFortyFives_NextRound(t *testing.T) {
	g := newFfGame(false)
	g.SetPhase(FortyFivesPhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetPhase() != FortyFivesPhaseBid {
		t.Errorf("after NextRound: round=%d phase=%d, want 2 / Bid", g.GetRoundNumber(), g.GetPhase())
	}
}

func TestFortyFives_HintAndPlayable(t *testing.T) {
	g := newFfGame(true)
	g.SetPhase(FortyFivesPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	ffSetHand(g.GetPlayer(0), ffCard(CardDesignClover, 7), ffCard(CardDesignClover, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestFortyFives_CpuFullMatchProgresses(t *testing.T) {
	g := newFfGame(false) // all CPU
	g.Reset()
	for guard := 0; guard < 6000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case FortyFivesPhaseBid:
			g.CpuBid()
		case FortyFivesPhasePlay:
			g.CpuPlay()
			if g.GetPhase() == FortyFivesPhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == FortyFivesPhaseTrickEnd {
					g.NextTrick()
				}
			}
		case FortyFivesPhaseRoundEnd:
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

func TestFortyFives_PlayerPlayErrors(t *testing.T) {
	g := newFfGame(true)
	g.SetPhase(FortyFivesPhasePlay)
	g.SetCurrentPlayerIdx(0)
	ffSetHand(g.GetPlayer(0), ffCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(FortyFivesPhaseBid)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestFortyFives_JSONRoundTrip(t *testing.T) {
	g := newFfGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 FortyFives
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() || g2.GetPhase() != g.GetPhase() {
		t.Error("round-trip mismatch")
	}
}

func TestFortyFives_UnmarshalErrors(t *testing.T) {
	var g FortyFives
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
