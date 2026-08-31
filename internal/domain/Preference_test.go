//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func prefCard(design, value int) *Card { return NewCard(design, value, false) }

func newPrefGame(human bool) *Preference {
	players := make([]*PreferencePlayer, PreferencePlayerCnt)
	players[0] = NewPreferencePlayer(human)
	for i := 1; i < PreferencePlayerCnt; i++ {
		players[i] = NewPreferencePlayer(false)
	}
	return NewPreference(NewTrumpCardsBelote(), players, DefaultPreferenceConfig())
}

func newPrefAllHuman() *Preference {
	players := make([]*PreferencePlayer, PreferencePlayerCnt)
	for i := 0; i < PreferencePlayerCnt; i++ {
		players[i] = NewPreferencePlayer(true)
	}
	return NewPreference(NewTrumpCardsBelote(), players, DefaultPreferenceConfig())
}

func prefSetHand(p *PreferencePlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestPreferenceConfig_Validate(t *testing.T) {
	if err := DefaultPreferenceConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (PreferenceConfig{CpuDifficulty: 9, TargetPoints: 30}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (PreferenceConfig{CpuDifficulty: PreferenceCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestPreference_ResetDealsTenAndBids(t *testing.T) {
	g := newPrefGame(true)
	g.Reset()
	if g.GetPhase() != PreferencePhaseBid {
		t.Errorf("phase = %d, want Bid", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetCardsSize() != PreferenceHandSize {
			t.Errorf("player %d dealt %d cards, want %d", i, g.GetPlayer(i).GetCardsSize(), PreferenceHandSize)
		}
	}
}

func TestPreference_BiddingResolvesHighestDeclarer(t *testing.T) {
	g := newPrefAllHuman()
	g.Reset() // dealer 0, forehand P1 bids first
	// P1 Six, P2 Seven (outbids), P0 Pass -> declarer P2 / Seven.
	if err := g.PlayerBid(PreferenceBidSix); err != nil {
		t.Fatalf("P1 six err: %v", err)
	}
	if err := g.PlayerBid(PreferenceBidSeven); err != nil {
		t.Fatalf("P2 seven err: %v", err)
	}
	if err := g.PlayerBid(PreferenceBidPass); err != nil {
		t.Fatalf("P0 pass err: %v", err)
	}
	if g.GetDeclarerIdx() != 2 || g.GetContract() != PreferenceBidSeven {
		t.Errorf("declarer=%d contract=%d, want 2 / Seven", g.GetDeclarerIdx(), g.GetContract())
	}
	if g.GetPhase() != PreferencePhasePlay {
		t.Errorf("phase after bidding = %d, want Play", g.GetPhase())
	}
}

func TestPreference_MisereHasNoTrump(t *testing.T) {
	g := newPrefAllHuman()
	g.Reset()
	if err := g.PlayerBid(PreferenceBidMisere); err != nil { // P1
		t.Fatalf("misere err: %v", err)
	}
	if err := g.PlayerBid(PreferenceBidPass); err != nil { // P2
		t.Fatalf("pass err: %v", err)
	}
	if err := g.PlayerBid(PreferenceBidPass); err != nil { // P0
		t.Fatalf("pass err: %v", err)
	}
	if g.GetContract() != PreferenceBidMisere || g.GetTrumpSuit() != 0 {
		t.Errorf("Misère contract=%d trump=%d, want Misère / 0", g.GetContract(), g.GetTrumpSuit())
	}
}

func TestPreference_CannotUnderbid(t *testing.T) {
	g := newPrefAllHuman()
	g.Reset()
	if err := g.PlayerBid(PreferenceBidEight); err != nil {
		t.Fatalf("eight err: %v", err)
	}
	if err := g.PlayerBid(PreferenceBidSix); err == nil {
		t.Error("expected under-bid rejection")
	}
}

func TestPreference_AllPassVoidsRound(t *testing.T) {
	g := newPrefAllHuman()
	g.Reset()
	for i := 0; i < PreferencePlayerCnt; i++ {
		if err := g.PlayerBid(PreferenceBidPass); err != nil {
			t.Fatalf("pass %d err: %v", i, err)
		}
	}
	if g.GetDeclarerIdx() != -1 || g.GetPhase() != PreferencePhaseRoundEnd {
		t.Errorf("all-pass: declarer=%d phase=%d, want -1 / RoundEnd", g.GetDeclarerIdx(), g.GetPhase())
	}
}

func TestPreference_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newPrefGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: prefCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: prefCard(CardDesignDiamond, 7)},
		{PlayerIdx: 2, Card: prefCard(CardDesignClover, 13)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("trick winner = %d, want 1 (trump beats high plain)", w)
	}
}

func TestPreference_MustFollow(t *testing.T) {
	g := newPrefGame(true)
	g.SetPhase(PreferencePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: prefCard(CardDesignClover, 1)}})
	prefSetHand(g.GetPlayer(0), prefCard(CardDesignClover, 13), prefCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(1); err == nil {
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestPreference_ScoreRoundMadeAndFailed(t *testing.T) {
	// Made: bid Six, took 6 -> declarer +6.
	g := newPrefGame(false)
	g.SetPhase(PreferencePhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(PreferenceBidSix)
	g.SetRoundTricks([PreferencePlayerCnt]int{6, 2, 2})
	g.ScoreRound()
	if sc := g.GetPlayerScores(); sc[0] != 6 {
		t.Errorf("declarer score = %d, want 6 on made Six", sc[0])
	}
	// Failed: bid Seven, took 5 -> each defender +7.
	g2 := newPrefGame(false)
	g2.SetPhase(PreferencePhaseRoundEnd)
	g2.SetDeclarerIdx(0)
	g2.SetContract(PreferenceBidSeven)
	g2.SetRoundTricks([PreferencePlayerCnt]int{5, 3, 2})
	g2.ScoreRound()
	sc := g2.GetPlayerScores()
	if sc[0] != 0 || sc[1] != 7 || sc[2] != 7 {
		t.Errorf("scores = %v, want [0 7 7] on failed Seven", sc)
	}
}

func TestPreference_MisereMadeWhenZeroTricks(t *testing.T) {
	g := newPrefGame(false)
	g.SetPhase(PreferencePhaseRoundEnd)
	g.SetDeclarerIdx(1)
	g.SetContract(PreferenceBidMisere)
	g.SetRoundTricks([PreferencePlayerCnt]int{6, 0, 4})
	g.ScoreRound()
	if sc := g.GetPlayerScores(); sc[1] != 10 {
		t.Errorf("declarer score = %d, want 10 on made Misère", sc[1])
	}
}

func TestPreference_ScoreRoundEndsMatch(t *testing.T) {
	g := newPrefGame(false)
	g.SetPhase(PreferencePhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(PreferenceBidEight)
	g.SetPlayerScores([PreferencePlayerCnt]int{25, 0, 0})
	g.SetRoundTricks([PreferencePlayerCnt]int{8, 1, 1}) // +8 -> 33 >= 30
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerPlayer() != 0 {
		t.Errorf("expected match end with player 0, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerPlayer())
	}
}

func TestPreference_NextRound(t *testing.T) {
	g := newPrefGame(false)
	g.SetPhase(PreferencePhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetPhase() != PreferencePhaseBid {
		t.Errorf("after NextRound: round=%d phase=%d, want 2 / Bid", g.GetRoundNumber(), g.GetPhase())
	}
}

func TestPreference_HintAndPlayable(t *testing.T) {
	g := newPrefGame(true)
	g.SetPhase(PreferencePhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	prefSetHand(g.GetPlayer(0), prefCard(CardDesignClover, 7), prefCard(CardDesignClover, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestPreference_CpuFullMatchProgresses(t *testing.T) {
	g := newPrefGame(false) // all CPU
	g.Reset()
	for guard := 0; guard < 6000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case PreferencePhaseBid:
			g.CpuBid()
		case PreferencePhasePlay:
			g.CpuPlay()
			if g.GetPhase() == PreferencePhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == PreferencePhaseTrickEnd {
					g.NextTrick()
				}
			}
		case PreferencePhaseRoundEnd:
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

func TestPreference_PlayerPlayErrors(t *testing.T) {
	g := newPrefGame(true)
	g.SetPhase(PreferencePhasePlay)
	g.SetCurrentPlayerIdx(0)
	prefSetHand(g.GetPlayer(0), prefCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(PreferencePhaseBid)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestPreference_JSONRoundTrip(t *testing.T) {
	g := newPrefGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Preference
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() || g2.GetPhase() != g.GetPhase() {
		t.Error("round-trip mismatch")
	}
}

func TestPreference_UnmarshalErrors(t *testing.T) {
	var g Preference
	if err := g.UnmarshalJSON([]byte(`{"ps":[]}`)); err == nil {
		t.Error("expected invalid-player-count error")
	}
	if err := g.UnmarshalJSON([]byte(`not json`)); err == nil {
		t.Error("expected json syntax error")
	}
	if err := g.UnmarshalJSON([]byte(`{"ps":[null,null,null]}`)); err == nil {
		t.Error("expected nil-player error")
	}
}

func TestPreference_GetDeclarerProgress(t *testing.T) {
	newGame := func() *Preference {
		g := newPrefGame(true)
		g.Reset()
		return g
	}

	t.Run("nil outside Play and TrickEnd phases", func(t *testing.T) {
		for _, ph := range []PreferencePhase{
			PreferencePhaseBid,
			PreferencePhaseRoundEnd,
			PreferencePhaseGameEnd,
		} {
			g := newGame()
			g.SetPhase(ph)
			g.SetDeclarerIdx(0)
			g.SetContract(PreferenceBidSix)
			if p := g.GetDeclarerProgress(); p != nil {
				t.Errorf("phase %v: got %+v, want nil", ph, p)
			}
		}
	})

	t.Run("nil when declarer is undeclared or invalid", func(t *testing.T) {
		for _, ph := range []PreferencePhase{PreferencePhasePlay, PreferencePhaseTrickEnd} {
			g := newGame()
			g.SetPhase(ph)
			g.SetDeclarerIdx(-1)
			g.SetContract(PreferenceBidSix)
			if p := g.GetDeclarerProgress(); p != nil {
				t.Errorf("declarer -1, phase %v: got %+v, want nil", ph, p)
			}

			g.SetDeclarerIdx(PreferencePlayerCnt)
			if p := g.GetDeclarerProgress(); p != nil {
				t.Errorf("declarer %d, phase %v: got %+v, want nil", PreferencePlayerCnt, ph, p)
			}
		}
	})

	t.Run("nil when contract is Pass", func(t *testing.T) {
		g := newGame()
		g.SetPhase(PreferencePhasePlay)
		g.SetDeclarerIdx(0)
		g.SetContract(PreferenceBidPass)
		if p := g.GetDeclarerProgress(); p != nil {
			t.Errorf("contract pass: got %+v, want nil", p)
		}
	})

	cases := []struct {
		name       string
		contract   PreferenceBid
		won        int
		played     int
		wantNeeded int
		wantRem    int
		wantMisere bool
		wantMade   bool
		wantFailed bool
	}{
		// Six contract (needed: 6)
		{"six in progress", PreferenceBidSix, 3, 5, 6, 5, false, false, false},
		{"six just enough remaining", PreferenceBidSix, 5, 9, 6, 1, false, false, false},
		{"six made when target reached", PreferenceBidSix, 6, 8, 6, 2, false, true, false},
		{"six made with surplus", PreferenceBidSix, 7, 8, 6, 2, false, true, false},
		{"six failed when remaining cannot reach", PreferenceBidSix, 3, 8, 6, 2, false, false, true},
		{"six failed on 0 won with 5 played", PreferenceBidSix, 0, 5, 6, 5, false, false, true},

		// Seven contract (needed: 7)
		{"seven in progress", PreferenceBidSeven, 4, 6, 7, 4, false, false, false},
		{"seven made", PreferenceBidSeven, 7, 9, 7, 1, false, true, false},
		{"seven failed", PreferenceBidSeven, 5, 9, 7, 1, false, false, true},

		// Eight contract (needed: 8)
		{"eight in progress", PreferenceBidEight, 5, 6, 8, 4, false, false, false},
		{"eight made", PreferenceBidEight, 8, 8, 8, 2, false, true, false},
		{"eight failed", PreferenceBidEight, 6, 9, 8, 1, false, false, true},

		// Misère contract (needed: 0)
		{"misere in progress clean sheet", PreferenceBidMisere, 0, 5, 0, 5, true, false, false},
		{"misere failed the moment a trick is taken", PreferenceBidMisere, 1, 3, 0, 7, true, false, true},
		{"misere failed with multiple tricks taken", PreferenceBidMisere, 3, 5, 0, 5, true, false, true},
		{"misere made when 0 tricks and 0 remaining", PreferenceBidMisere, 0, 10, 0, 0, true, true, false},
		{"misere failed at end with 1 trick", PreferenceBidMisere, 1, 10, 0, 0, true, false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, ph := range []PreferencePhase{PreferencePhasePlay, PreferencePhaseTrickEnd} {
				g := newGame()
				g.SetPhase(ph)
				g.SetDeclarerIdx(0)
				g.SetContract(tc.contract)

				var tr [PreferencePlayerCnt]int
				tr[0] = tc.won
				tr[1] = tc.played - tc.won
				g.SetRoundTricks(tr)

				pr := g.GetDeclarerProgress()
				if pr == nil {
					t.Fatalf("phase %v: GetDeclarerProgress = nil", ph)
				}
				if pr.Won != tc.won {
					t.Errorf("phase %v: Won = %d, want %d", ph, pr.Won, tc.won)
				}
				if pr.Needed != tc.wantNeeded {
					t.Errorf("phase %v: Needed = %d, want %d", ph, pr.Needed, tc.wantNeeded)
				}
				if pr.Remaining != tc.wantRem {
					t.Errorf("phase %v: Remaining = %d, want %d", ph, pr.Remaining, tc.wantRem)
				}
				if pr.IsMisere != tc.wantMisere {
					t.Errorf("phase %v: IsMisere = %v, want %v", ph, pr.IsMisere, tc.wantMisere)
				}
				if pr.Made != tc.wantMade {
					t.Errorf("phase %v: Made = %v, want %v", ph, pr.Made, tc.wantMade)
				}
				if pr.Failed != tc.wantFailed {
					t.Errorf("phase %v: Failed = %v, want %v", ph, pr.Failed, tc.wantFailed)
				}
				if pr.Made && pr.Failed {
					t.Errorf("phase %v: Made and Failed both true", ph)
				}
			}
		})
	}
}
