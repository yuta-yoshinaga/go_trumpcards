//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func sfCard(design, value int) *Card { return NewCard(design, value, false) }

func newSfGame(human bool) *SpoilFive {
	players := make([]*SpoilFivePlayer, SpoilFivePlayerCnt)
	players[0] = NewSpoilFivePlayer(human)
	for i := 1; i < SpoilFivePlayerCnt; i++ {
		players[i] = NewSpoilFivePlayer(false)
	}
	return NewSpoilFive(NewTrumpCards(0), players, DefaultSpoilFiveConfig())
}

func sfSetHand(p *SpoilFivePlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func sfTrick(g *SpoilFive, cards ...*Card) {
	t := make([]*TrickCard, len(cards))
	for i, c := range cards {
		t[i] = &TrickCard{PlayerIdx: i, Card: c}
	}
	g.SetCurrentTrick(t)
}

func TestSpoilFiveConfig_Validate(t *testing.T) {
	if err := DefaultSpoilFiveConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (SpoilFiveConfig{CpuDifficulty: 9, TargetPoints: 30}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (SpoilFiveConfig{CpuDifficulty: SpoilFiveCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestSpoilFive_ResetDealsFiveAndAntes(t *testing.T) {
	g := newSfGame(true)
	g.Reset()
	if g.GetPhase() != SpoilFivePhasePlay {
		t.Errorf("phase = %d, want Play", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetCardsSize() != SpoilFiveHandSize {
			t.Errorf("player %d dealt %d cards, want %d", i, g.GetPlayer(i).GetCardsSize(), SpoilFiveHandSize)
		}
	}
	if g.GetTrumpSuit() < 1 || g.GetTrumpSuit() > 4 {
		t.Errorf("trump = %d, want 1-4", g.GetTrumpSuit())
	}
	if g.GetPot() != SpoilFiveAntePerRound {
		t.Errorf("pot = %d, want %d", g.GetPot(), SpoilFiveAntePerRound)
	}
}

func TestSpoilFive_FixedTopTrumpRank(t *testing.T) {
	g := newSfGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	// 5♦ > J♦ > ♥A > A♦ > K♦
	cards := []*Card{
		sfCard(CardDesignDiamond, 5), sfCard(CardDesignDiamond, 11),
		sfCard(CardDesignHeart, 1), sfCard(CardDesignDiamond, 1), sfCard(CardDesignDiamond, 13),
	}
	for i := 1; i < len(cards); i++ {
		if g.spoilRank(cards[i-1]) <= g.spoilRank(cards[i]) {
			t.Errorf("rank(%v) should beat rank(%v)", cards[i-1], cards[i])
		}
	}
}

func TestSpoilFive_HeartAceIsTopTrumpEvenWhenNotTrump(t *testing.T) {
	g := newSfGame(false)
	g.SetTrumpSuit(CardDesignSpade) // hearts is NOT trump
	// Lead spade (trump) low; ♥A is a trump-tier card and beats a low trump.
	sfTrick(g, sfCard(CardDesignSpade, 7), sfCard(CardDesignHeart, 1))
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1 (♥A is always a top trump)", w)
	}
}

func TestSpoilFive_RenegingTopTrumpNotForced(t *testing.T) {
	g := newSfGame(true)
	g.SetPhase(SpoilFivePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	// Trump (diamond) led low. Human holds only a top trump (5♦) + an off-suit club.
	// Reneging: may decline to play the top trump and play the club instead.
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: sfCard(CardDesignDiamond, 7)}})
	sfSetHand(g.GetPlayer(0), sfCard(CardDesignDiamond, 5), sfCard(CardDesignClover, 9))
	if err := g.PlayerPlay(1); err != nil { // play the club (renege the 5♦) — allowed
		t.Fatalf("reneging a top trump should be allowed, got: %v", err)
	}
}

func TestSpoilFive_MustFollowOrdinaryTrump(t *testing.T) {
	g := newSfGame(true)
	g.SetPhase(SpoilFivePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: sfCard(CardDesignDiamond, 7)}})
	// Holds an ordinary trump (K♦, not a top trump) → must follow.
	sfSetHand(g.GetPlayer(0), sfCard(CardDesignDiamond, 13), sfCard(CardDesignClover, 9))
	if err := g.PlayerPlay(1); err == nil { // club while holding K♦
		t.Error("expected must-follow-trump error")
	}
	if err := g.PlayerPlay(0); err != nil { // K♦ follows
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestSpoilFive_ResolveTrickThreeTricksWins(t *testing.T) {
	g := newSfGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(SpoilFivePhaseTrickEnd)
	g.SetTrickNumber(3)
	// Give player 0 two prior tricks, then win this one -> 3 -> immediate round win.
	g.GetPlayer(0).IncRoundTricks()
	g.GetPlayer(0).IncRoundTricks()
	sfTrick(g,
		sfCard(CardDesignDiamond, 5), // p0 plays the best trump -> wins
		sfCard(CardDesignClover, 9), sfCard(CardDesignClover, 8),
		sfCard(CardDesignClover, 7), sfCard(CardDesignClover, 6))
	g.ResolveTrick()
	if g.GetRoundWinnerIdx() != 0 || g.GetPhase() != SpoilFivePhaseRoundEnd {
		t.Errorf("expected immediate round win for p0, winner=%d phase=%d", g.GetRoundWinnerIdx(), g.GetPhase())
	}
}

func TestSpoilFive_SpoilWhenNoOneReachesThree(t *testing.T) {
	g := newSfGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(SpoilFivePhaseTrickEnd)
	g.SetTrickNumber(5) // final trick, nobody has 2 priors here
	sfTrick(g,
		sfCard(CardDesignClover, 13), sfCard(CardDesignClover, 9), sfCard(CardDesignClover, 8),
		sfCard(CardDesignClover, 7), sfCard(CardDesignClover, 6))
	g.ResolveTrick()
	if g.GetRoundWinnerIdx() != -1 || g.GetPhase() != SpoilFivePhaseRoundEnd {
		t.Errorf("expected Spoil (winner -1), winner=%d phase=%d", g.GetRoundWinnerIdx(), g.GetPhase())
	}
}

func TestSpoilFive_ScoreRoundWinnerTakesPotElseCarries(t *testing.T) {
	// Winner takes pot.
	g := newSfGame(false)
	g.SetPhase(SpoilFivePhaseRoundEnd)
	g.SetPot(15)
	g.currentTrick = nil
	g.roundWinnerIdx = 2
	g.ScoreRound()
	if g.GetPlayer(2).GetScore() != 15 || g.GetPot() != 0 {
		t.Errorf("winner score=%d pot=%d, want 15 / 0", g.GetPlayer(2).GetScore(), g.GetPot())
	}
	// Spoil: pot carries over.
	g2 := newSfGame(false)
	g2.SetPhase(SpoilFivePhaseRoundEnd)
	g2.SetPot(20)
	g2.roundWinnerIdx = -1
	g2.ScoreRound()
	if g2.GetPot() != 20 {
		t.Errorf("spoil pot=%d, want 20 (carried)", g2.GetPot())
	}
}

func TestSpoilFive_ScoreRoundEndsMatch(t *testing.T) {
	g := newSfGame(false)
	g.SetPhase(SpoilFivePhaseRoundEnd)
	g.SetPot(10)
	g.GetPlayer(1).SetScore(25)
	g.roundWinnerIdx = 1
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerPlayer() != 1 {
		t.Errorf("expected match end with player 1, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerPlayer())
	}
}

func TestSpoilFive_NextRound(t *testing.T) {
	g := newSfGame(false)
	g.SetPhase(SpoilFivePhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetPhase() != SpoilFivePhasePlay {
		t.Errorf("after NextRound: round=%d phase=%d, want 2 / Play", g.GetRoundNumber(), g.GetPhase())
	}
}

func TestSpoilFive_HintAndPlayable(t *testing.T) {
	g := newSfGame(true)
	g.SetPhase(SpoilFivePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	sfSetHand(g.GetPlayer(0), sfCard(CardDesignClover, 7), sfCard(CardDesignClover, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestSpoilFive_CpuFullMatchProgresses(t *testing.T) {
	g := newSfGame(false) // all CPU
	g.Reset()
	for guard := 0; guard < 6000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case SpoilFivePhasePlay:
			g.CpuPlay()
			if g.GetPhase() == SpoilFivePhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == SpoilFivePhaseTrickEnd {
					g.NextTrick()
				}
			}
		case SpoilFivePhaseRoundEnd:
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

func TestSpoilFive_PlayerPlayErrors(t *testing.T) {
	g := newSfGame(true)
	g.SetPhase(SpoilFivePhasePlay)
	g.SetCurrentPlayerIdx(0)
	sfSetHand(g.GetPlayer(0), sfCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(SpoilFivePhaseRoundEnd)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestSpoilFive_JSONRoundTrip(t *testing.T) {
	g := newSfGame(true)
	g.Reset()
	g.GetPlayer(1).SetScore(7)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 SpoilFive
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() || g2.GetPot() != g.GetPot() {
		t.Error("round-trip mismatch")
	}
	if g2.GetPlayer(1).GetScore() != 7 {
		t.Error("score not preserved across round-trip")
	}
}

func TestSpoilFive_UnmarshalErrors(t *testing.T) {
	var g SpoilFive
	if err := g.UnmarshalJSON([]byte(`{"ps":[]}`)); err == nil {
		t.Error("expected invalid-player-count error")
	}
	if err := g.UnmarshalJSON([]byte(`not json`)); err == nil {
		t.Error("expected json syntax error")
	}
	if err := g.UnmarshalJSON([]byte(`{"ps":[null,null,null,null,null]}`)); err == nil {
		t.Error("expected nil-player error")
	}
}
