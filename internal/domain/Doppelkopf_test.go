//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func dkCard(design, value int) *Card { return NewCard(design, value, false) }

func newDKGame(human bool) *Doppelkopf {
	cfg := DefaultDoppelkopfConfig()
	players := make([]*DoppelkopfPlayer, DoppelkopfPlayerCnt)
	players[0] = NewDoppelkopfPlayer(human, cfg.StartChips)
	for i := 1; i < DoppelkopfPlayerCnt; i++ {
		players[i] = NewDoppelkopfPlayer(false, cfg.StartChips)
	}
	return NewDoppelkopf(NewTrumpCardsPinochle(), players, cfg)
}

func dkSetHand(p *DoppelkopfPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestDoppelkopfConfig_Validate(t *testing.T) {
	if err := DefaultDoppelkopfConfig().Validate(); err != nil {
		t.Fatalf("default invalid: %v", err)
	}
	bad := []DoppelkopfConfig{
		{CpuDifficulty: -1, BaseChips: 2, StartChips: 20, TargetChips: 40},
		{CpuDifficulty: 9, BaseChips: 2, StartChips: 20, TargetChips: 40},
		{CpuDifficulty: 1, BaseChips: 0, StartChips: 20, TargetChips: 40},
		{CpuDifficulty: 1, BaseChips: 2, StartChips: 0, TargetChips: 40},
		{CpuDifficulty: 1, BaseChips: 2, StartChips: 20, TargetChips: 20},
	}
	for i, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestDoppelkopf_ResetDealsAndAssignsTeams(t *testing.T) {
	g := newDKGame(true)
	g.Reset()
	if g.GetPhase() != DoppelkopfPhasePlay {
		t.Fatalf("phase = %v, want Play", g.GetPhase())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != DoppelkopfHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, DoppelkopfHandSize)
		}
	}
	// Exactly two ♣Q exist → either two Re players, or one solo Re.
	reCount := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if g.IsRe(i) {
			reCount++
		}
	}
	if g.IsSoloRe() {
		if reCount != 1 {
			t.Errorf("solo Re should have 1 Re player, got %d", reCount)
		}
	} else if reCount != 2 {
		t.Errorf("normal game should have 2 Re players, got %d", reCount)
	}
}

func TestDoppelkopf_TrumpStrengthOrdering(t *testing.T) {
	// ♥10 > Q♣>Q♠>Q♥>Q♦ > J♣>J♠>J♥>J♦ > A♦>10♦>K♦>9♦ > fail A
	order := []*Card{
		dkCard(CardDesignHeart, 10), // Dulle
		dkCard(CardDesignClover, 12), dkCard(CardDesignSpade, 12), dkCard(CardDesignHeart, 12), dkCard(CardDesignDiamond, 12),
		dkCard(CardDesignClover, 11), dkCard(CardDesignSpade, 11), dkCard(CardDesignHeart, 11), dkCard(CardDesignDiamond, 11),
		dkCard(CardDesignDiamond, 1), dkCard(CardDesignDiamond, 10), dkCard(CardDesignDiamond, 13), dkCard(CardDesignDiamond, 9),
		dkCard(CardDesignClover, 1),
	}
	for i := 1; i < len(order); i++ {
		if dkStrength(order[i-1]) <= dkStrength(order[i]) {
			t.Errorf("order broken at %d: %s !> %s", i, cardStr(order[i-1]), cardStr(order[i]))
		}
	}
}

func TestDoppelkopf_Classification(t *testing.T) {
	trumps := []*Card{
		dkCard(CardDesignHeart, 10), dkCard(CardDesignDiamond, 1), dkCard(CardDesignClover, 12), dkCard(CardDesignSpade, 11),
	}
	for _, c := range trumps {
		if !dkIsTrump(c) {
			t.Errorf("%s should be trump", cardStr(c))
		}
	}
	fails := []*Card{dkCard(CardDesignHeart, 1), dkCard(CardDesignHeart, 13), dkCard(CardDesignClover, 1), dkCard(CardDesignSpade, 10)}
	for _, c := range fails {
		if dkIsTrump(c) {
			t.Errorf("%s should be fail", cardStr(c))
		}
	}
}

func TestDoppelkopf_CardPointsTotal240(t *testing.T) {
	want := map[int]int{1: 11, 10: 10, 13: 4, 12: 3, 11: 2, 9: 0}
	for v, p := range want {
		if dkCardPoints(v) != p {
			t.Errorf("points(%d)=%d want %d", v, dkCardPoints(v), p)
		}
	}
	total := 0
	for _, v := range []int{1, 9, 10, 11, 12, 13} {
		total += dkCardPoints(v) * 2 * 4 // 2 copies × 4 suits
	}
	if total != DoppelkopfTotalPoints {
		t.Errorf("deck total = %d, want %d", total, DoppelkopfTotalPoints)
	}
}

func TestDoppelkopf_DulleBeatsAllAndTieFirstWins(t *testing.T) {
	g := newDKGame(false)
	g.SetPhase(DoppelkopfPhaseTrickEnd)
	g.SetTrickNumber(1)
	// Two Dulle (♥10) played by p1 then p3; first (p1) wins the tie.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: dkCard(CardDesignClover, 12)}, // Q♣ trump
		{PlayerIdx: 1, Card: dkCard(CardDesignHeart, 10)},  // Dulle (highest)
		{PlayerIdx: 2, Card: dkCard(CardDesignDiamond, 1)}, // A♦ trump
		{PlayerIdx: 3, Card: dkCard(CardDesignHeart, 10)},  // Dulle (second copy)
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1 (first Dulle wins tie)", w)
	}
}

func TestDoppelkopf_TrickWinnerHighFail(t *testing.T) {
	g := newDKGame(false)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: dkCard(CardDesignClover, 13)}, // K♣
		{PlayerIdx: 1, Card: dkCard(CardDesignClover, 1)},  // A♣ (strongest fail in suit)
		{PlayerIdx: 2, Card: dkCard(CardDesignSpade, 1)},   // off-suit, ignored
		{PlayerIdx: 3, Card: dkCard(CardDesignClover, 9)},
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1", w)
	}
}

func TestDoppelkopf_TrumpBeatsFailLead(t *testing.T) {
	// A fail suit is led; a player void in it overruffs with trump. The trump
	// must win even though its suit ID (trump) differs from the lead suit.
	g := newDKGame(false)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: dkCard(CardDesignClover, 1)},   // A♣ fail lead
		{PlayerIdx: 1, Card: dkCard(CardDesignDiamond, 13)}, // K♦ trump
		{PlayerIdx: 2, Card: dkCard(CardDesignClover, 10)},  // 10♣ fail
		{PlayerIdx: 3, Card: dkCard(CardDesignClover, 9)},   // 9♣ fail
	})
	if w := g.trickWinner(); w != 1 {
		t.Errorf("winner = %d, want 1 (K♦ trump beats the ♣ fail lead)", w)
	}
	// And the Dulle (♥10) beats a lower trump played earlier.
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: dkCard(CardDesignClover, 1)},  // A♣ fail lead
		{PlayerIdx: 1, Card: dkCard(CardDesignDiamond, 1)}, // A♦ trump
		{PlayerIdx: 2, Card: dkCard(CardDesignHeart, 10)},  // Dulle (top trump)
		{PlayerIdx: 3, Card: dkCard(CardDesignSpade, 12)},  // Q♠ trump
	})
	if w := g.trickWinner(); w != 2 {
		t.Errorf("winner = %d, want 2 (Dulle is the highest trump)", w)
	}
}

func TestDoppelkopf_TrickTopStrengthGuard(t *testing.T) {
	g := newDKGame(false)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: dkCard(CardDesignClover, 1)},
	})
	// A winner not present in the current trick must yield the sentinel, not panic.
	if got := g.trickTopStrength(99); got != -1<<30 {
		t.Errorf("trickTopStrength(absent) = %d, want sentinel", got)
	}
	if got := g.trickTopStrength(0); got != dkStrength(dkCard(CardDesignClover, 1)) {
		t.Errorf("trickTopStrength(0) = %d, want A♣ strength", got)
	}
}

func TestDoppelkopf_MustFollowSuit(t *testing.T) {
	g := newDKGame(true) // human is player 0
	g.SetPhase(DoppelkopfPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: dkCard(CardDesignClover, 1)}}) // lead ♣ fail
	// Human (player 0) has a club → must follow.
	dkSetHand(g.GetPlayer(0), dkCard(CardDesignClover, 13), dkCard(CardDesignSpade, 1))
	if err := g.PlayerPlay(1); err == nil { // try spade
		t.Error("expected must-follow error")
	}
	if err := g.PlayerPlay(0); err != nil { // club ok
		t.Fatalf("valid follow err: %v", err)
	}
}

func TestDoppelkopf_ResolveAndNextTrick(t *testing.T) {
	g := newDKGame(false)
	g.SetPhase(DoppelkopfPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: dkCard(CardDesignHeart, 10)},
		{PlayerIdx: 1, Card: dkCard(CardDesignClover, 12)},
		{PlayerIdx: 2, Card: dkCard(CardDesignSpade, 1)},
		{PlayerIdx: 3, Card: dkCard(CardDesignSpade, 13)},
	})
	g.ResolveTrick()
	if g.GetPlayer(0).GetTrickCount() != 1 {
		t.Error("Dulle should win the trick for player 0")
	}
	if g.GetLeadPlayerIdx() != 0 {
		t.Error("lead should follow winner")
	}
	g.NextTrick()
	if g.GetTrickNumber() != 2 || g.GetPhase() != DoppelkopfPhasePlay {
		t.Errorf("next trick not started: trick=%d phase=%v", g.GetTrickNumber(), g.GetPhase())
	}
}

func TestDoppelkopf_LastTrickToRoundEnd(t *testing.T) {
	g := newDKGame(false)
	g.SetPhase(DoppelkopfPhaseTrickEnd)
	g.SetTrickNumber(DoppelkopfTrickCount)
	g.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: dkCard(CardDesignClover, 1)},
		{PlayerIdx: 1, Card: dkCard(CardDesignClover, 13)},
		{PlayerIdx: 2, Card: dkCard(CardDesignClover, 10)},
		{PlayerIdx: 3, Card: dkCard(CardDesignClover, 9)},
	})
	g.ResolveTrick()
	if g.GetPhase() != DoppelkopfPhaseRoundEnd {
		t.Errorf("phase = %v, want RoundEnd", g.GetPhase())
	}
}

func TestDoppelkopf_Announce(t *testing.T) {
	g := newDKGame(true)
	g.SetPhase(DoppelkopfPhasePlay)
	g.SetTrickNumber(1)
	g.SetReTeam([DoppelkopfPlayerCnt]bool{true, false, true, false}) // human (0) is Re
	if !g.CanHumanAnnounce() {
		t.Fatal("human should be able to announce in trick 1")
	}
	if err := g.PlayerAnnounce(); err != nil {
		t.Fatalf("announce err: %v", err)
	}
	if !g.IsReAnnounced() {
		t.Error("Re should be announced")
	}
	// After trick 1 the window closes.
	g.SetTrickNumber(2)
	if g.CanHumanAnnounce() {
		t.Error("announce window should be closed after trick 1")
	}
	if err := g.PlayerAnnounce(); err == nil {
		t.Error("expected error announcing past window")
	}
}

func TestDoppelkopf_ScoreRoundReWinsZeroSum(t *testing.T) {
	g := newDKGame(false)
	g.SetPhase(DoppelkopfPhaseRoundEnd)
	g.SetReTeam([DoppelkopfPlayerCnt]bool{true, false, true, false})
	// Re players (0,2) collect 130 pts.
	g.GetPlayer(0).AddTrick([]*Card{dkCard(CardDesignClover, 1), dkCard(CardDesignClover, 1), dkCard(CardDesignSpade, 1)})                                  // 33
	g.GetPlayer(2).AddTrick([]*Card{dkCard(CardDesignHeart, 1), dkCard(CardDesignDiamond, 1), dkCard(CardDesignSpade, 10)})                                 // 32
	g.GetPlayer(0).AddTrick([]*Card{dkCard(CardDesignHeart, 10), dkCard(CardDesignDiamond, 10), dkCard(CardDesignClover, 10), dkCard(CardDesignSpade, 13)}) // 24+? 10+10+10+4
	g.GetPlayer(2).AddTrick([]*Card{dkCard(CardDesignClover, 13), dkCard(CardDesignSpade, 13), dkCard(CardDesignHeart, 13)})                                // 12
	// Kontra (1,3) collect the rest.
	g.GetPlayer(1).AddTrick([]*Card{dkCard(CardDesignHeart, 1)}) // some pts
	before := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		before += g.GetPlayer(i).GetChips()
	}
	g.ScoreRound()
	after := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		after += g.GetPlayer(i).GetChips()
	}
	if before != after {
		t.Errorf("chips not zero-sum: before=%d after=%d", before, after)
	}
	if !g.AreTeamsRevealed() {
		t.Error("teams should be revealed at round end")
	}
}

func TestDoppelkopf_GamePointsAndMultiplier(t *testing.T) {
	if dkGamePoints(100, false) != 1 {
		t.Error("100 pts loser → 1 game pt")
	}
	if dkGamePoints(89, false) != 2 {
		t.Error("<90 → 2")
	}
	if dkGamePoints(59, false) != 3 {
		t.Error("<60 → 3")
	}
	if dkGamePoints(29, false) != 4 {
		t.Error("<30 → 4")
	}
	if dkGamePoints(0, true) != 5 {
		t.Error("schwarz → 5")
	}
}

func TestDoppelkopf_SoloReSettlement(t *testing.T) {
	g := newDKGame(false)
	g.SetPhase(DoppelkopfPhaseRoundEnd)
	g.SetReTeam([DoppelkopfPlayerCnt]bool{true, false, false, false})
	g.soloRe = true
	// Solo Re player 0 takes everything (240) → wins, schwarz.
	all := []*Card{}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for _, v := range []int{1, 9, 10, 11, 12, 13} {
			all = append(all, dkCard(suit, v), dkCard(suit, v))
		}
	}
	g.GetPlayer(0).AddTrick(all)
	g.ScoreRound()
	if !g.GetRoundReWon() {
		t.Errorf("solo Re should win with %d pts", g.GetRoundRePoints())
	}
	// Player 0 gains 3×, others lose 1× each → zero-sum.
	sum := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		sum += g.GetPlayer(i).GetChips()
	}
	if sum != 4*g.GetConfig().StartChips {
		t.Errorf("solo settlement not zero-sum: %d", sum)
	}
}

func TestDoppelkopf_NextRoundAndGameEnd(t *testing.T) {
	g := newDKGame(false)
	g.Reset()
	g.SetPhase(DoppelkopfPhaseRoundEnd)
	d := g.GetDealerIdx()
	g.NextRound()
	if g.GetRoundNumber() != 2 || g.GetDealerIdx() != (d+1)%DoppelkopfPlayerCnt {
		t.Errorf("next round/dealer wrong: round=%d dealer=%d", g.GetRoundNumber(), g.GetDealerIdx())
	}
	// Force game end via target chips.
	cfg := g.GetConfig()
	cfg.TargetChips = cfg.StartChips + 1
	g.SetConfig(cfg)
	g.SetPhase(DoppelkopfPhaseRoundEnd)
	g.SetReTeam([DoppelkopfPlayerCnt]bool{true, true, false, false})
	g.GetPlayer(0).AddTrick([]*Card{dkCard(CardDesignClover, 1), dkCard(CardDesignClover, 1), dkCard(CardDesignSpade, 1), dkCard(CardDesignSpade, 1), dkCard(CardDesignHeart, 1), dkCard(CardDesignHeart, 1), dkCard(CardDesignDiamond, 1), dkCard(CardDesignDiamond, 1), dkCard(CardDesignClover, 10), dkCard(CardDesignClover, 10), dkCard(CardDesignSpade, 10), dkCard(CardDesignSpade, 10)})
	g.ScoreRound()
	if !g.GetGameEndFlag() || g.GetWinnerIdx() < 0 {
		t.Errorf("game should end: end=%v winner=%d", g.GetGameEndFlag(), g.GetWinnerIdx())
	}
}

func TestDoppelkopf_CpuFullRoundProgresses(t *testing.T) {
	g := newDKGame(false)
	g.Reset()
	for steps := 0; steps < 300; steps++ {
		switch g.GetPhase() {
		case DoppelkopfPhasePlay:
			g.CpuPlay()
		case DoppelkopfPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == DoppelkopfPhaseTrickEnd {
				g.NextTrick()
			}
		case DoppelkopfPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case DoppelkopfPhaseGameEnd:
			return
		}
		if g.GetRoundNumber() > 60 {
			break
		}
	}
}

func TestDoppelkopf_HintAndPlayable(t *testing.T) {
	g := newDKGame(true)
	g.SetPhase(DoppelkopfPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	dkSetHand(g.GetPlayer(0), dkCard(CardDesignClover, 9), dkCard(CardDesignSpade, 1))
	if h := g.GetHint(); h == nil || len(h.CardIndices) != 1 {
		t.Errorf("hint missing: %+v", h)
	}
	if idx := g.GetPlayableIndices(0); len(idx) != 2 {
		t.Errorf("playable = %v, want 2 (lead)", idx)
	}
	g.SetPhase(DoppelkopfPhaseRoundEnd)
	if idx := g.GetPlayableIndices(0); idx != nil {
		t.Error("playable should be nil outside play phase")
	}
}

func TestDoppelkopf_JSONRoundTrip(t *testing.T) {
	g := newDKGame(true)
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var g2 Doppelkopf
	if err := json.Unmarshal(data, &g2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if g2.GetPhase() != g.GetPhase() || g2.GetPlayerCnt() != DoppelkopfPlayerCnt {
		t.Error("round trip mismatch")
	}
}

func TestDoppelkopf_UnmarshalOversized(t *testing.T) {
	var g Doppelkopf
	bad := `{"al":[`
	for i := 0; i < doppelkopfMaxSliceLen+1; i++ {
		if i > 0 {
			bad += ","
		}
		bad += `{"TurnNumber":0,"PlayerIdx":-1,"ActionType":"x","Detail":"y","Cards":null}`
	}
	bad += `]}`
	if err := json.Unmarshal([]byte(bad), &g); !errors.Is(err, errDoppelkopfOversized) {
		t.Errorf("err = %v, want errDoppelkopfOversized", err)
	}
}

func TestDoppelkopf_PlayerPlayErrors(t *testing.T) {
	g := newDKGame(true)
	g.Reset()
	g.SetCurrentPlayerIdx(1) // CPU
	if err := g.PlayerPlay(0); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(DoppelkopfPhaseRoundEnd)
	if err := g.PlayerPlay(0); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
	g.gameEndFlag = true
	if err := g.PlayerPlay(0); !errors.Is(err, ErrGameEnded) {
		t.Errorf("err = %v, want ErrGameEnded", err)
	}
}

// #5639: 切り札は「♦全部 + 全 Q + 全 J + ♥10」という複合ルールで、手札を見ても
// 一目では分からない。Web は isDoppelkopfTrump でバッジを付けているのに、CUI は
// 判定手段そのものを持っていなかった (dkIsTrump は非公開)。
func TestDoppelkopfGetTrumpIndices(t *testing.T) {
	g := newDKGame(true)
	p := g.GetPlayer(0)
	p.ResetRound()
	// ♦9(切り札) / ♠A(フェイル) / ♣Q(切り札) / ♥10(Dulle=切り札) / ♠10(フェイル)
	p.AddCard(dkCard(CardDesignDiamond, 9))
	p.AddCard(dkCard(CardDesignSpade, 1))
	p.AddCard(dkCard(CardDesignClover, 12))
	p.AddCard(dkCard(CardDesignHeart, 10))
	p.AddCard(dkCard(CardDesignSpade, 10))

	got := g.GetTrumpIndices(0)
	want := []int{0, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("GetTrumpIndices = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("GetTrumpIndices = %v, want %v", got, want)
		}
	}
}

// 範囲外のプレイヤーは nil。呼び出し側で境界を気にせず使えるようにする。
func TestDoppelkopfGetTrumpIndicesOutOfRange(t *testing.T) {
	g := newDKGame(true)
	// **境界そのものを踏む。**99 だけだと `>=` を `>` に緩めても通ってしまい、
	// 実際に落ちる playerIdx == len(players) が素通りする。
	for _, idx := range []int{-1, DoppelkopfPlayerCnt, 99} {
		if got := g.GetTrumpIndices(idx); got != nil {
			t.Fatalf("GetTrumpIndices(%d) = %v, want nil", idx, got)
		}
	}
}
