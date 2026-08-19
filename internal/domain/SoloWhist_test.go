//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func swCard(design, value int) *Card { return NewCard(design, value, false) }

func newSwGame(human bool) *SoloWhist {
	players := make([]*SoloWhistPlayer, SoloWhistPlayerCnt)
	players[0] = NewSoloWhistPlayer(human)
	for i := 1; i < SoloWhistPlayerCnt; i++ {
		players[i] = NewSoloWhistPlayer(false)
	}
	return NewSoloWhist(NewTrumpCards(0), players, DefaultSoloWhistConfig())
}

func newSwAllHuman() *SoloWhist {
	players := make([]*SoloWhistPlayer, SoloWhistPlayerCnt)
	for i := 0; i < SoloWhistPlayerCnt; i++ {
		players[i] = NewSoloWhistPlayer(true)
	}
	return NewSoloWhist(NewTrumpCards(0), players, DefaultSoloWhistConfig())
}

func swSetHand(p *SoloWhistPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestSoloWhistConfig_Validate(t *testing.T) {
	if err := DefaultSoloWhistConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (SoloWhistConfig{CpuDifficulty: 9, TargetPoints: 21}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (SoloWhistConfig{CpuDifficulty: SoloWhistCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestSoloWhist_ResetDealsAndBids(t *testing.T) {
	g := newSwGame(true)
	g.Reset()
	if g.GetPhase() != SoloWhistPhaseBid {
		t.Errorf("phase = %d, want Bid", g.GetPhase())
	}
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
	}
	if total != SoloWhistPlayerCnt*SoloWhistHandSize {
		t.Errorf("dealt %d cards, want %d", total, SoloWhistPlayerCnt*SoloWhistHandSize)
	}
}

func TestSoloWhist_BiddingResolvesHighestDeclarer(t *testing.T) {
	g := newSwAllHuman()
	g.Reset() // dealer 0, forehand (P1) bids first
	// P1 Solo, P2 Misère (outbids), P3 Pass, P0 Pass -> declarer P2, contract Misère.
	if err := g.PlayerBid(SoloWhistBidSolo); err != nil {
		t.Fatalf("P1 solo err: %v", err)
	}
	if err := g.PlayerBid(SoloWhistBidMisere); err != nil {
		t.Fatalf("P2 misere err: %v", err)
	}
	if err := g.PlayerBid(SoloWhistBidPass); err != nil {
		t.Fatalf("P3 pass err: %v", err)
	}
	if err := g.PlayerBid(SoloWhistBidPass); err != nil {
		t.Fatalf("P0 pass err: %v", err)
	}
	if g.GetDeclarerIdx() != 2 || g.GetContract() != SoloWhistBidMisere {
		t.Errorf("declarer=%d contract=%d, want 2 / Misère", g.GetDeclarerIdx(), g.GetContract())
	}
	if g.GetTrumpSuit() != 0 {
		t.Errorf("Misère trump = %d, want 0 (no trump)", g.GetTrumpSuit())
	}
	if g.GetPhase() != SoloWhistPhasePlay {
		t.Errorf("phase after bidding = %d, want Play", g.GetPhase())
	}
}

func TestSoloWhist_BiddingCannotUnderbid(t *testing.T) {
	g := newSwAllHuman()
	g.Reset()
	if err := g.PlayerBid(SoloWhistBidAbundance); err != nil { // P1 high bid
		t.Fatalf("abundance err: %v", err)
	}
	if err := g.PlayerBid(SoloWhistBidSolo); err == nil { // P2 cannot bid lower
		t.Error("expected under-bid rejection")
	}
}

func TestSoloWhist_AllPassVoidsRound(t *testing.T) {
	g := newSwAllHuman()
	g.Reset()
	for i := 0; i < SoloWhistPlayerCnt; i++ {
		if err := g.PlayerBid(SoloWhistBidPass); err != nil {
			t.Fatalf("pass %d err: %v", i, err)
		}
	}
	if g.GetDeclarerIdx() != -1 {
		t.Errorf("declarer = %d, want -1 on all-pass", g.GetDeclarerIdx())
	}
	if g.GetPhase() != SoloWhistPhaseRoundEnd {
		t.Errorf("phase = %d, want RoundEnd on all-pass", g.GetPhase())
	}
}

func TestSoloWhist_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newSwGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: swCard(CardDesignClover, 1)},  // Ace lead
		{PlayerIdx: 1, Card: swCard(CardDesignDiamond, 2)}, // low trump beats it
		{PlayerIdx: 2, Card: swCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: swCard(CardDesignClover, 12)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("trick winner = %d, want 1 (trump beats high plain)", w)
	}
}

func TestSoloWhist_TrickWinnerNoTrumpMisere(t *testing.T) {
	g := newSwGame(false)
	g.SetTrumpSuit(0) // Misère: no trump
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: swCard(CardDesignClover, 13)},
		{PlayerIdx: 1, Card: swCard(CardDesignDiamond, 1)}, // off-suit ace cannot win
		{PlayerIdx: 2, Card: swCard(CardDesignClover, 1)},  // ace of lead suit wins
		{PlayerIdx: 3, Card: swCard(CardDesignClover, 12)},
	})
	if w := g.trickWinner(); w != 2 {
		t.Errorf("trick winner = %d, want 2 (no trump, high of lead suit)", w)
	}
}

func TestSoloWhist_MustFollow(t *testing.T) {
	g := newSwGame(true)
	g.SetPhase(SoloWhistPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: swCard(CardDesignClover, 1)}})
	swSetHand(g.GetPlayer(0), swCard(CardDesignClover, 13), swCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(1); err == nil { // diamond while holding club
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // K♣ follows
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestSoloWhist_ResolveTrickCountsTricks(t *testing.T) {
	g := newSwGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetPhase(SoloWhistPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: swCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: swCard(CardDesignClover, 5)},
		{PlayerIdx: 2, Card: swCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: swCard(CardDesignClover, 12)},
	})
	g.ResolveTrick()
	rt := g.GetRoundTricks()
	if rt[0] != 1 {
		t.Errorf("player 0 tricks = %d, want 1", rt[0])
	}
}

func TestSoloWhist_ScoreRoundContractMadeAndFailed(t *testing.T) {
	// Solo made: declarer takes >= 8 tricks -> declarer +2.
	g := newSwGame(false)
	g.SetPhase(SoloWhistPhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(SoloWhistBidSolo)
	g.SetRoundTricks([SoloWhistPlayerCnt]int{8, 2, 2, 1})
	g.ScoreRound()
	if sc := g.GetPlayerScores(); sc[0] != 2 {
		t.Errorf("declarer score = %d, want 2 on made Solo", sc[0])
	}
	// Solo failed: declarer < 8 -> each defender +2.
	g2 := newSwGame(false)
	g2.SetPhase(SoloWhistPhaseRoundEnd)
	g2.SetDeclarerIdx(0)
	g2.SetContract(SoloWhistBidSolo)
	g2.SetRoundTricks([SoloWhistPlayerCnt]int{5, 3, 3, 2})
	g2.ScoreRound()
	sc := g2.GetPlayerScores()
	if sc[0] != 0 || sc[1] != 2 || sc[2] != 2 || sc[3] != 2 {
		t.Errorf("scores = %v, want [0 2 2 2] on failed Solo", sc)
	}
}

func TestSoloWhist_MisereContractMade(t *testing.T) {
	g := newSwGame(false)
	g.SetPhase(SoloWhistPhaseRoundEnd)
	g.SetDeclarerIdx(1)
	g.SetContract(SoloWhistBidMisere)
	g.SetRoundTricks([SoloWhistPlayerCnt]int{5, 0, 4, 4}) // declarer 0 tricks -> made
	g.ScoreRound()
	if sc := g.GetPlayerScores(); sc[1] != 3 {
		t.Errorf("declarer score = %d, want 3 on made Misère", sc[1])
	}
}

func TestSoloWhist_ScoreRoundEndsMatch(t *testing.T) {
	g := newSwGame(false)
	g.SetPhase(SoloWhistPhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(SoloWhistBidAbundance)
	g.SetPlayerScores([SoloWhistPlayerCnt]int{20, 0, 0, 0})
	g.SetRoundTricks([SoloWhistPlayerCnt]int{9, 2, 1, 1}) // abundance made -> +4 -> 24
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerPlayer() != 0 {
		t.Errorf("expected match end with player 0, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerPlayer())
	}
}

func TestSoloWhist_NextRound(t *testing.T) {
	g := newSwGame(false)
	g.SetPhase(SoloWhistPhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("round number = %d, want 2", g.GetRoundNumber())
	}
	if g.GetPhase() != SoloWhistPhaseBid {
		t.Errorf("phase after NextRound = %d, want Bid", g.GetPhase())
	}
}

func TestSoloWhist_HintAndPlayable(t *testing.T) {
	g := newSwGame(true)
	g.SetPhase(SoloWhistPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	swSetHand(g.GetPlayer(0), swCard(CardDesignClover, 7), swCard(CardDesignClover, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestSoloWhist_CpuFullMatchProgresses(t *testing.T) {
	g := newSwGame(false) // all CPU
	g.Reset()
	for guard := 0; guard < 6000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case SoloWhistPhaseBid:
			g.CpuBid()
		case SoloWhistPhasePlay:
			g.CpuPlay()
			if g.GetPhase() == SoloWhistPhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == SoloWhistPhaseTrickEnd {
					g.NextTrick()
				}
			}
		case SoloWhistPhaseRoundEnd:
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

func TestSoloWhist_PlayerPlayErrors(t *testing.T) {
	g := newSwGame(true)
	g.SetPhase(SoloWhistPhasePlay)
	g.SetCurrentPlayerIdx(0)
	swSetHand(g.GetPlayer(0), swCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(SoloWhistPhaseBid)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestSoloWhist_JSONRoundTrip(t *testing.T) {
	g := newSwGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 SoloWhist
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() || g2.GetPhase() != g.GetPhase() {
		t.Error("round-trip mismatch")
	}
}

func TestSoloWhist_UnmarshalErrors(t *testing.T) {
	var g SoloWhist
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

// #5649: ミゼールは1トリック取った瞬間に失敗が確定する。Web は
// solowhist-contract-progress で常時出しているのに、CUI はラウンドが終わるまで
// 達成可否を出していなかった。同じ入札制の Nap は #4763 で解決済み。
func TestSoloWhistGetDeclarerProgress(t *testing.T) {
	cases := []struct {
		name        string
		contract    SoloWhistBid
		won         int
		played      int
		wantNeeded  int
		wantMisere  bool
		wantUnreach bool
		wantMade    bool
	}{
		{"ソロ: まだ届く", SoloWhistBidSolo, 5, 8, 8, false, false, false},
		{"ソロ: 達成", SoloWhistBidSolo, 8, 10, 8, false, false, true},
		// 残り 13-11=2 トリック。5+2=7 < 8 なのでこの時点で不成立が確定する。
		{"ソロ: 残り全部取っても届かない", SoloWhistBidSolo, 5, 11, 8, false, true, false},
		// ちょうど届く境界: 残り 3 で 5+3 = 8。まだ失敗にしてはいけない。
		{"ソロ: ちょうど届く境界", SoloWhistBidSolo, 5, 10, 8, false, false, false},
		{"ミゼール: 無傷", SoloWhistBidMisere, 0, 5, 0, true, false, false},
		// **1トリックでも取れば即失敗。**残りトリックは関係ない。
		{"ミゼール: 1トリックで即失敗", SoloWhistBidMisere, 1, 5, 0, true, true, false},
		{"ミゼール: 最後まで無傷なら達成", SoloWhistBidMisere, 0, SoloWhistTrickCount, 0, true, false, true},
		{"アバンダンス: 達成", SoloWhistBidAbundance, 9, 11, 9, false, false, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := newSwGame(true)
			g.SetPhase(SoloWhistPhasePlay)
			g.SetDeclarerIdx(0)
			g.SetContract(c.contract)
			tricks := [SoloWhistPlayerCnt]int{}
			tricks[0] = c.won
			// 残りは「配られた13から消化済みを引いた分」。宣言者以外の取得分は
			// 席1にまとめる (どの席が取ったかは進捗に影響しない)。
			tricks[1] = c.played - c.won
			g.SetRoundTricks(tricks)

			p := g.GetDeclarerProgress()
			if p == nil {
				t.Fatal("GetDeclarerProgress = nil, want a progress")
			}
			if p.Won != c.won || p.Needed != c.wantNeeded {
				t.Errorf("Won/Needed = %d/%d, want %d/%d", p.Won, p.Needed, c.won, c.wantNeeded)
			}
			if p.IsMisere != c.wantMisere {
				t.Errorf("IsMisere = %v, want %v", p.IsMisere, c.wantMisere)
			}
			if p.Unreachable != c.wantUnreach {
				t.Errorf("Unreachable = %v, want %v", p.Unreachable, c.wantUnreach)
			}
			if p.Made != c.wantMade {
				t.Errorf("Made = %v, want %v", p.Made, c.wantMade)
			}
		})
	}
}

// 入札中やラウンド終了後、宣言者未確定なら進捗そのものが無い。
func TestSoloWhistGetDeclarerProgressNilOutsidePlay(t *testing.T) {
	g := newSwGame(true)
	g.SetDeclarerIdx(0)
	g.SetContract(SoloWhistBidSolo)
	for _, ph := range []SoloWhistPhase{SoloWhistPhaseBid, SoloWhistPhaseRoundEnd} {
		g.SetPhase(ph)
		if p := g.GetDeclarerProgress(); p != nil {
			t.Errorf("phase %v: got %+v, want nil", ph, p)
		}
	}

	g.SetPhase(SoloWhistPhasePlay)
	g.SetDeclarerIdx(-1)
	if p := g.GetDeclarerProgress(); p != nil {
		t.Errorf("no declarer: got %+v, want nil", p)
	}
}
