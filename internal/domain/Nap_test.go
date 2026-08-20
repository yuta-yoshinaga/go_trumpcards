//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func napCard(design, value int) *Card { return NewCard(design, value, false) }

func newNapGame(human bool) *Nap {
	players := make([]*NapPlayer, NapPlayerCnt)
	players[0] = NewNapPlayer(human)
	for i := 1; i < NapPlayerCnt; i++ {
		players[i] = NewNapPlayer(false)
	}
	return NewNap(NewTrumpCards(0), players, DefaultNapConfig())
}

func newNapAllHuman() *Nap {
	players := make([]*NapPlayer, NapPlayerCnt)
	for i := 0; i < NapPlayerCnt; i++ {
		players[i] = NewNapPlayer(true)
	}
	return NewNap(NewTrumpCards(0), players, DefaultNapConfig())
}

func napSetHand(p *NapPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNapConfig_Validate(t *testing.T) {
	if err := DefaultNapConfig().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
	if err := (NapConfig{CpuDifficulty: 9, TargetPoints: 20}).Validate(); err == nil {
		t.Error("expected difficulty out-of-range error")
	}
	if err := (NapConfig{CpuDifficulty: NapCpuDifficultyNormal, TargetPoints: 0}).Validate(); err == nil {
		t.Error("expected target-points min error")
	}
}

func TestNap_ResetDealsFiveAndBids(t *testing.T) {
	g := newNapGame(true)
	g.Reset()
	if g.GetPhase() != NapPhaseBid {
		t.Errorf("phase = %d, want Bid", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.GetPlayer(i).GetCardsSize() != NapHandSize {
			t.Errorf("player %d dealt %d cards, want %d", i, g.GetPlayer(i).GetCardsSize(), NapHandSize)
		}
	}
}

func TestNap_BiddingResolvesHighestDeclarer(t *testing.T) {
	g := newNapAllHuman()
	g.Reset() // dealer 0, forehand P1 bids first
	// P1 Two, P2 Four (outbids), P3 Pass, P0 Pass -> declarer P2 / Four.
	if err := g.PlayerBid(NapBidTwo); err != nil {
		t.Fatalf("P1 two err: %v", err)
	}
	if err := g.PlayerBid(NapBidFour); err != nil {
		t.Fatalf("P2 four err: %v", err)
	}
	if err := g.PlayerBid(NapBidPass); err != nil {
		t.Fatalf("P3 pass err: %v", err)
	}
	if err := g.PlayerBid(NapBidPass); err != nil {
		t.Fatalf("P0 pass err: %v", err)
	}
	if g.GetDeclarerIdx() != 2 || g.GetContract() != NapBidFour {
		t.Errorf("declarer=%d contract=%d, want 2 / Four", g.GetDeclarerIdx(), g.GetContract())
	}
	if g.GetPhase() != NapPhasePlay {
		t.Errorf("phase after bidding = %d, want Play", g.GetPhase())
	}
	if g.GetTrumpSuit() < 1 || g.GetTrumpSuit() > 4 {
		t.Errorf("trump = %d, want 1-4", g.GetTrumpSuit())
	}
	if g.GetLeadPlayerIdx() != 2 {
		t.Errorf("declarer should lead, lead=%d want 2", g.GetLeadPlayerIdx())
	}
}

func TestNap_CannotUnderbid(t *testing.T) {
	g := newNapAllHuman()
	g.Reset()
	if err := g.PlayerBid(NapBidNap); err != nil {
		t.Fatalf("nap err: %v", err)
	}
	if err := g.PlayerBid(NapBidThree); err == nil {
		t.Error("expected under-bid rejection")
	}
}

func TestNap_AllPassVoidsRound(t *testing.T) {
	g := newNapAllHuman()
	g.Reset()
	for i := 0; i < NapPlayerCnt; i++ {
		if err := g.PlayerBid(NapBidPass); err != nil {
			t.Fatalf("pass %d err: %v", i, err)
		}
	}
	if g.GetDeclarerIdx() != -1 || g.GetPhase() != NapPhaseRoundEnd {
		t.Errorf("all-pass: declarer=%d phase=%d, want -1 / RoundEnd", g.GetDeclarerIdx(), g.GetPhase())
	}
}

func TestNap_TrickWinnerTrumpBeatsLead(t *testing.T) {
	g := newNapGame(false)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: napCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: napCard(CardDesignDiamond, 2)},
		{PlayerIdx: 2, Card: napCard(CardDesignClover, 13)},
		{PlayerIdx: 3, Card: napCard(CardDesignClover, 12)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("trick winner = %d, want 1 (trump beats high plain)", w)
	}
}

func TestNap_MustFollow(t *testing.T) {
	g := newNapGame(true)
	g.SetPhase(NapPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: napCard(CardDesignClover, 1)}})
	napSetHand(g.GetPlayer(0), napCard(CardDesignClover, 13), napCard(CardDesignDiamond, 7))
	if err := g.PlayerPlay(1); err == nil {
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestNap_ScoreRoundMadeAndFailed(t *testing.T) {
	// Made: declarer bid Three, took 3 -> +3.
	g := newNapGame(false)
	g.SetPhase(NapPhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(NapBidThree)
	g.SetRoundTricks([NapPlayerCnt]int{3, 1, 1, 0})
	g.ScoreRound()
	if sc := g.GetPlayerScores(); sc[0] != 3 {
		t.Errorf("declarer score = %d, want 3 on made Three", sc[0])
	}
	// Failed: bid Four, took 2 -> each defender +4.
	g2 := newNapGame(false)
	g2.SetPhase(NapPhaseRoundEnd)
	g2.SetDeclarerIdx(0)
	g2.SetContract(NapBidFour)
	g2.SetRoundTricks([NapPlayerCnt]int{2, 1, 1, 1})
	g2.ScoreRound()
	sc := g2.GetPlayerScores()
	if sc[0] != 0 || sc[1] != 4 || sc[2] != 4 || sc[3] != 4 {
		t.Errorf("scores = %v, want [0 4 4 4] on failed Four", sc)
	}
}

func TestNap_NapScoringAsymmetry(t *testing.T) {
	// Nap made -> +10.
	g := newNapGame(false)
	g.SetPhase(NapPhaseRoundEnd)
	g.SetDeclarerIdx(1)
	g.SetContract(NapBidNap)
	g.SetRoundTricks([NapPlayerCnt]int{0, 5, 0, 0})
	g.ScoreRound()
	if sc := g.GetPlayerScores(); sc[1] != 10 {
		t.Errorf("declarer score = %d, want 10 on made Nap", sc[1])
	}
	// Nap failed -> each defender +5.
	g2 := newNapGame(false)
	g2.SetPhase(NapPhaseRoundEnd)
	g2.SetDeclarerIdx(1)
	g2.SetContract(NapBidNap)
	g2.SetRoundTricks([NapPlayerCnt]int{2, 3, 0, 0})
	g2.ScoreRound()
	sc := g2.GetPlayerScores()
	if sc[0] != 5 || sc[2] != 5 || sc[3] != 5 || sc[1] != 0 {
		t.Errorf("scores = %v, want defenders +5 each on failed Nap", sc)
	}
}

func TestNap_ScoreRoundEndsMatch(t *testing.T) {
	g := newNapGame(false)
	g.SetPhase(NapPhaseRoundEnd)
	g.SetDeclarerIdx(0)
	g.SetContract(NapBidNap)
	g.SetPlayerScores([NapPlayerCnt]int{12, 0, 0, 0})
	g.SetRoundTricks([NapPlayerCnt]int{5, 0, 0, 0}) // +10 -> 22 >= 20
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerPlayer() != 0 {
		t.Errorf("expected match end with player 0, flag=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerPlayer())
	}
}

func TestNap_NextRound(t *testing.T) {
	g := newNapGame(false)
	g.SetPhase(NapPhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetPhase() != NapPhaseBid {
		t.Errorf("after NextRound: round=%d phase=%d, want 2 / Bid", g.GetRoundNumber(), g.GetPhase())
	}
}

func TestNap_HintAndPlayable(t *testing.T) {
	g := newNapGame(true)
	g.SetPhase(NapPhasePlay)
	g.SetTrumpSuit(CardDesignDiamond)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	napSetHand(g.GetPlayer(0), napCard(CardDesignClover, 7), napCard(CardDesignClover, 1))
	if idxs := g.GetPlayableIndices(0); len(idxs) != 2 {
		t.Errorf("playable = %v, want 2", idxs)
	}
	if h := g.GetHint(); h == nil || len(h.CardIndices) == 0 {
		t.Error("expected a lead hint")
	}
}

func TestNap_CpuFullMatchProgresses(t *testing.T) {
	g := newNapGame(false) // all CPU
	g.Reset()
	for guard := 0; guard < 6000 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case NapPhaseBid:
			g.CpuBid()
		case NapPhasePlay:
			g.CpuPlay()
			if g.GetPhase() == NapPhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == NapPhaseTrickEnd {
					g.NextTrick()
				}
			}
		case NapPhaseRoundEnd:
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

func TestNap_PlayerPlayErrors(t *testing.T) {
	g := newNapGame(true)
	g.SetPhase(NapPhasePlay)
	g.SetCurrentPlayerIdx(0)
	napSetHand(g.GetPlayer(0), napCard(CardDesignClover, 7))
	if err := g.PlayerPlay(-1); err == nil {
		t.Error("expected out-of-range error")
	}
	g.SetPhase(NapPhaseBid)
	if err := g.PlayerPlay(0); err == nil {
		t.Error("expected wrong-phase error")
	}
}

func TestNap_JSONRoundTrip(t *testing.T) {
	g := newNapGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Nap
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() || g2.GetPhase() != g.GetPhase() {
		t.Error("round-trip mismatch")
	}
}

func TestNap_UnmarshalErrors(t *testing.T) {
	var g Nap
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

// **CUI は宣言者が何トリック取ったかを一切知らせていなかった (#4763)。**
// Web は nap-declarer-progress で常時出している。
func TestNap_GetDeclarerProgress(t *testing.T) {
	card := func(v int) *Card { return NewCard(CardDesignSpade, v, false) }
	// declarer が won トリック、他プレイヤーが others トリック取った局面。
	setup := func(contract NapBid, won, others int) *Nap {
		g := NewDefaultNap()
		g.Reset()
		g.SetPhase(NapPhasePlay)
		g.SetDeclarerIdx(0)
		g.SetContract(contract)
		for i := 0; i < won; i++ {
			g.GetPlayer(0).AddTrick([]*Card{card(2)})
		}
		for i := 0; i < others; i++ {
			g.GetPlayer(1).AddTrick([]*Card{card(3)})
		}
		return g
	}

	t.Run("nothing to report before a declarer is settled", func(t *testing.T) {
		g := setup(NapBidThree, 0, 0)
		g.SetDeclarerIdx(-1)
		assert.Nil(t, g.GetDeclarerProgress())
	})

	t.Run("nothing to report outside the play phases", func(t *testing.T) {
		g := setup(NapBidThree, 1, 1)
		g.SetPhase(NapPhaseRoundEnd)
		assert.Nil(t, g.GetDeclarerProgress())
	})

	t.Run("counts the declarer's tricks and what is left", func(t *testing.T) {
		pr := setup(NapBidThree, 1, 1).GetDeclarerProgress()
		if assert.NotNil(t, pr) {
			assert.Equal(t, 1, pr.Won)
			assert.Equal(t, 3, pr.Needed)
			assert.Equal(t, NapTrickCount-2, pr.Remaining, "全員のトリックを引く")
			assert.False(t, pr.Unreachable)
		}
	})

	// **「もう届かない」は残りを全部取っても足りないときだけ。**早すぎる判定は
	// まだ勝てるラウンドを投げさせる。
	t.Run("declares failure only when even every remaining trick falls short", func(t *testing.T) {
		// 契約5、宣言者0勝ち、相手1勝ち → 残り4。5-0=5 > 4 で不可能。
		unreachable := setup(NapBidNap, 0, 1).GetDeclarerProgress()
		if assert.NotNil(t, unreachable) {
			assert.True(t, unreachable.Unreachable)
		}
		// 契約5、まだ誰も取っていない → 残り5。ちょうど届く見込みがある。
		alive := setup(NapBidNap, 0, 0).GetDeclarerProgress()
		if assert.NotNil(t, alive) {
			assert.False(t, alive.Unreachable, "ちょうど届くうちは不可能にしない")
		}
	})

	t.Run("stays reachable once the contract is already made", func(t *testing.T) {
		pr := setup(NapBidThree, 3, 1).GetDeclarerProgress()
		if assert.NotNil(t, pr) {
			assert.False(t, pr.Unreachable)
		}
	})
}

// #5651: Nap 契約だけ賭け金が非対称 (達成 +10 / 失敗は相手が各 +5)。他の契約は
// 達成・失敗とも契約数と同じ数が動く。**この非対称さが Nap を宣言するか否かの
// 判断そのもの**なのに、入札ボタンには契約名しか出ていなかった。
func TestNapBidPayout(t *testing.T) {
	cases := []struct {
		bid            NapBid
		wantMake, fail int
	}{
		{NapBidTwo, 2, 2},
		{NapBidThree, 3, 3},
		{NapBidFour, 4, 4},
		// **ここだけ非対称。**達成すれば 10 だが、失敗しても相手が得るのは各 5。
		{NapBidNap, 10, 5},
	}
	for _, c := range cases {
		gotMake, gotFail := NapBidPayout(c.bid)
		if gotMake != c.wantMake || gotFail != c.fail {
			t.Errorf("NapBidPayout(%v) = %d/%d, want %d/%d", c.bid, gotMake, gotFail, c.wantMake, c.fail)
		}
	}
}

// 表と実際の精算が食い違わないこと。ラウンドを実際に精算させて突き合わせる。
func TestNapBidPayoutMatchesTheSettlement(t *testing.T) {
	for _, bid := range []NapBid{NapBidTwo, NapBidThree, NapBidFour, NapBidNap} {
		wantMake, wantFail := NapBidPayout(bid)

		// 達成: 宣言者が必要トリック数を取った状態で精算する。
		g := newNapGame(true)
		g.SetPhase(NapPhaseRoundEnd)
		g.SetDeclarerIdx(0)
		g.SetContract(bid)
		tricks := [NapPlayerCnt]int{}
		tricks[0] = napBidTarget(bid)
		g.SetRoundTricks(tricks)
		before := g.GetPlayerScores()
		g.ScoreRound()
		if got := g.GetPlayerScores()[0] - before[0]; got != wantMake {
			t.Errorf("%v made: declarer gained %d, want %d", bid, got, wantMake)
		}

		// 失敗: 1 トリックも取らせずに精算する。相手それぞれが failValue を得る。
		g2 := newNapGame(true)
		g2.SetPhase(NapPhaseRoundEnd)
		g2.SetDeclarerIdx(0)
		g2.SetContract(bid)
		g2.SetRoundTricks([NapPlayerCnt]int{})
		before2 := g2.GetPlayerScores()
		g2.ScoreRound()
		after2 := g2.GetPlayerScores()
		if got := after2[0] - before2[0]; got != 0 {
			t.Errorf("%v failed: declarer moved by %d, want 0 (相手が得る側)", bid, got)
		}
		if got := after2[1] - before2[1]; got != wantFail {
			t.Errorf("%v failed: opponent gained %d, want %d", bid, got, wantFail)
		}
	}
}
