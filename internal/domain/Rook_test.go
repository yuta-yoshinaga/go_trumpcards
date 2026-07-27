//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func rookNewGame() *domain.Rook {
	players := []*domain.RookPlayer{
		domain.NewRookPlayer(true, 0),
		domain.NewRookPlayer(false, 1),
		domain.NewRookPlayer(false, 0),
		domain.NewRookPlayer(false, 1),
	}
	return domain.NewRook(players, domain.DefaultRookConfig())
}

func rookCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func rookBird() *domain.Card {
	return domain.NewCard(domain.RookBirdDesign, domain.RookBirdValue, false)
}

func TestRookResetDealsFullDeck(t *testing.T) {
	g := rookNewGame()
	g.Reset()
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if sz := g.GetPlayer(i).GetCardsSize(); sz != domain.RookHandSize {
			t.Errorf("player %d hand = %d, want %d", i, sz, domain.RookHandSize)
		}
		total += g.GetPlayer(i).GetCardsSize()
	}
	total += len(g.GetNest())
	if total != domain.RookColorCnt*14+1 {
		t.Errorf("total cards = %d, want 57", total)
	}
	if len(g.GetNest()) != domain.RookNestSize {
		t.Errorf("nest = %d, want %d", len(g.GetNest()), domain.RookNestSize)
	}
	if g.GetPhase() != domain.RookPhaseBid {
		t.Errorf("phase = %d, want bid", g.GetPhase())
	}
}

func TestRookCardPoints(t *testing.T) {
	g := rookNewGame()
	cases := []struct {
		card *domain.Card
		want int
	}{
		{rookCard(1, 5), 5},
		{rookCard(2, 10), 10},
		{rookCard(3, 14), 10},
		{rookCard(4, 1), 15},
		{rookBird(), 20},
		{rookCard(1, 7), 0},
		{nil, 0},
	}
	for _, c := range cases {
		if got := g.CardPointsPublic(c.card); got != c.want {
			t.Errorf("points(%v) = %d, want %d", c.card, got, c.want)
		}
	}
	// Full-deck point total = 4 colors * 40 + 20 rook = 180.
	sum := 0
	for color := 1; color <= 4; color++ {
		for v := 1; v <= 14; v++ {
			sum += g.CardPointsPublic(rookCard(color, v))
		}
	}
	sum += g.CardPointsPublic(rookBird())
	if sum != 180 {
		t.Errorf("deck point total = %d, want 180", sum)
	}
}

func TestRookCardRankOneIsHigh(t *testing.T) {
	g := rookNewGame()
	g.SetTrumpColor(-1) // no trump yet
	// 1 outranks 14 outranks 13 ... outranks 2 within a plain suit.
	if g.CardRankPublic(rookCard(1, 1)) <= g.CardRankPublic(rookCard(1, 14)) {
		t.Errorf("1 should outrank 14")
	}
	if g.CardRankPublic(rookCard(1, 14)) <= g.CardRankPublic(rookCard(1, 13)) {
		t.Errorf("14 should outrank 13")
	}
	if g.CardRankPublic(rookCard(1, 3)) <= g.CardRankPublic(rookCard(1, 2)) {
		t.Errorf("3 should outrank 2")
	}
	// Rook bird is the highest of all.
	g.SetTrumpColor(2)
	if g.CardRankPublic(rookBird()) <= g.CardRankPublic(rookCard(2, 1)) {
		t.Errorf("rook bird should outrank the highest trump")
	}
	// A trump outranks any plain card.
	if g.CardRankPublic(rookCard(2, 2)) <= g.CardRankPublic(rookCard(1, 1)) {
		t.Errorf("trump 2 should outrank plain 1")
	}
	// Rook bird's effective suit is the trump color.
	if g.EffectiveSuitPublic(rookBird()) != 2 {
		t.Errorf("rook bird effective suit should be trump color")
	}
}

// rookSetupTrick puts the given trick on the table with trump declared and
// returns the winner index by resolving it directly through play mechanics.
func TestRookTrickWinner(t *testing.T) {
	g := rookNewGame()
	g.SetTrumpColor(4) // black is trump
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.RookPhasePlay)
	g.SetTrickNumber(1)
	g.SetLeadPlayerIdx(0)

	t.Run("highest of led suit wins", func(t *testing.T) {
		g.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: rookCard(1, 10)},
			{PlayerIdx: 1, Card: rookCard(1, 1)}, // 1 is high in led suit
			{PlayerIdx: 2, Card: rookCard(1, 13)},
			{PlayerIdx: 3, Card: rookCard(2, 1)}, // off-suit, cannot win
		})
		g.SetPhase(domain.RookPhaseTrickEnd)
		g.ResolveTrick()
		if g.GetPlayer(1).GetTrickCount() != 1 {
			t.Errorf("player 1 should win with the high card of the led suit")
		}
	})

	t.Run("trump beats led suit, rook bird beats trump", func(t *testing.T) {
		g2 := rookNewGame()
		g2.SetTrumpColor(4)
		g2.SetDeclarerIdx(0)
		g2.SetLeadPlayerIdx(0)
		g2.SetTrickNumber(1)
		g2.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: rookCard(1, 1)},  // led plain suit, high
			{PlayerIdx: 1, Card: rookCard(4, 2)},  // low trump
			{PlayerIdx: 2, Card: rookBird()},      // rook bird, highest trump
			{PlayerIdx: 3, Card: rookCard(4, 14)}, // high trump
		})
		g2.SetPhase(domain.RookPhaseTrickEnd)
		g2.ResolveTrick()
		if g2.GetPlayer(2).GetTrickCount() != 1 {
			t.Errorf("rook bird should win the trick")
		}
		// Captured points on winner: rook(20)+1(15)+color4-2(0)+color4-14(10) = 45.
		// (The 14 scores 10 even as a trump — point cards score their value regardless of suit.)
		if g2.GetPlayer(2).GetPoints() != 45 {
			t.Errorf("winner points = %d, want 45", g2.GetPlayer(2).GetPoints())
		}
	})
}

func TestRookBiddingFlow(t *testing.T) {
	g := rookNewGame()
	g.Reset()
	g.SetBidPlayerIdx(0) // force human's turn
	if err := g.PlayerBid(75); err != nil {
		t.Fatalf("bid failed: %v", err)
	}
	if g.GetHighestBid() != 75 || g.GetHighestBidder() != 0 {
		t.Errorf("highest bid not recorded")
	}
	// invalid: not higher than current
	g.SetBidPlayerIdx(0)
	if err := g.PlayerBid(75); err == nil {
		t.Errorf("bid must exceed current highest")
	}
	// invalid: out of step
	if err := g.PlayerBid(77); err == nil {
		t.Errorf("bid must be a multiple of step")
	}
	// invalid: out of range
	if err := g.PlayerBid(200); err == nil {
		t.Errorf("bid must be within range")
	}
}

func TestRookAllPassRedeals(t *testing.T) {
	g := rookNewGame()
	g.Reset()
	// Everyone passes -> redeal, still in bid phase, no bidder.
	for i := 0; i < 8 && g.GetPhase() == domain.RookPhaseBid; i++ {
		if g.IsHumanBidTurn() {
			_ = g.PlayerPass()
		} else {
			g.CpuBid()
		}
		if g.GetPhase() == domain.RookPhaseNestExchange {
			break
		}
	}
	// Either someone bid (exchange phase) or a redeal happened (still bid phase).
	if g.GetPhase() != domain.RookPhaseBid && g.GetPhase() != domain.RookPhaseNestExchange {
		t.Errorf("unexpected phase after bidding: %d", g.GetPhase())
	}
}

// rookHumanDeclarerExchange sets up a deterministic nest-exchange state with the
// human (player 0) as declarer holding 18 known cards.
func rookHumanDeclarerExchange() *domain.Rook {
	g := rookNewGame()
	g.SetDeclarerIdx(0)
	g.SetContractBid(70)
	g.GetPlayer(0).SetIsDeclarer(true)
	p := g.GetPlayer(0)
	p.Reset()
	// indices 0-4: point cards (5+10+10+15+5 = 45); 5-17: filler.
	for _, c := range []*domain.Card{
		rookCard(1, 5), rookCard(1, 10), rookCard(1, 14), rookCard(1, 1), rookCard(2, 5),
	} {
		p.AddCard(c)
	}
	for i := 0; i < 13; i++ {
		p.AddCard(rookCard(3, (i%12)+2))
	}
	g.SetPhase(domain.RookPhaseNestExchange)
	g.SetCurrentPlayerIdx(0)
	return g
}

func TestRookNestExchangeAndTrump(t *testing.T) {
	g := rookHumanDeclarerExchange()
	// Human holds 18 cards (13 + 5 nest).
	if g.GetPlayer(0).GetCardsSize() != domain.RookHandSize+domain.RookNestSize {
		t.Fatalf("declarer hand size = %d", g.GetPlayer(0).GetCardsSize())
	}
	// Discard wrong count -> error.
	if err := g.PlayerExchangeNest([]int{0, 1, 2}, 1); err == nil {
		t.Errorf("must discard exactly 5")
	}
	// Invalid trump color.
	if err := g.PlayerExchangeNest([]int{0, 1, 2, 3, 4}, 9); err == nil {
		t.Errorf("trump color must be 1..4")
	}
	// Duplicate index.
	if err := g.PlayerExchangeNest([]int{0, 0, 1, 2, 3}, 1); err == nil {
		t.Errorf("duplicate discard index rejected")
	}
	// Valid discard + trump declaration.
	if err := g.PlayerExchangeNest([]int{0, 1, 2, 3, 4}, 3); err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if g.GetTrumpColor() != 3 {
		t.Errorf("trump color = %d, want 3", g.GetTrumpColor())
	}
	if g.GetPlayer(0).GetCardsSize() != domain.RookHandSize {
		t.Errorf("declarer hand back to 13")
	}
	if g.GetPhase() != domain.RookPhasePlay {
		t.Errorf("should be in play phase")
	}
	if g.GetLeadPlayerIdx() != 0 {
		t.Errorf("declarer leads")
	}
}

func TestRookValidatePlayFollowSuit(t *testing.T) {
	g := rookNewGame()
	g.SetTrumpColor(4)
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.RookPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetLeadPlayerIdx(1)
	// Trick led with color 1; player 0 holds color 1 and must follow.
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: rookCard(1, 8)}})
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(rookCard(1, 5)) // idx 0 follows suit
	p.AddCard(rookCard(2, 9)) // idx 1 off-suit
	p.AddCard(rookBird())     // idx 2 rook bird, always legal
	valid := g.GetValidPlayIndices(0)
	// Must-follow: only idx 0 (color 1) and idx 2 (rook bird) are legal.
	assert.Contains(t, valid, 0)
	assert.Contains(t, valid, 2)
	assert.NotContains(t, valid, 1)
	// Playing the off-suit card is rejected.
	if err := g.PlayerPlay(1); err == nil {
		t.Errorf("off-suit play should be rejected when able to follow")
	}
	// Rook bird can always be played.
	if err := g.PlayerPlay(2); err != nil {
		t.Errorf("rook bird should be playable: %v", err)
	}
}

func TestRookScoreRoundMadeAndSet(t *testing.T) {
	// Declarer team makes the bid.
	g := rookNewGame()
	g.SetDeclarerIdx(0)
	g.SetContractBid(80)
	g.SetPhase(domain.RookPhaseRoundEnd)
	g.GetPlayer(0).SetPoints(90) // team 0
	g.GetPlayer(1).SetPoints(70) // team 1
	g.ScoreRound()
	if g.GetTeamScore(0) != 90 {
		t.Errorf("made: team0 = %d, want 90", g.GetTeamScore(0))
	}
	if g.GetTeamScore(1) != 70 {
		t.Errorf("made: team1 = %d, want 70", g.GetTeamScore(1))
	}

	// Declarer team is set.
	g2 := rookNewGame()
	g2.SetDeclarerIdx(1) // team 1 declares
	g2.SetContractBid(100)
	g2.SetPhase(domain.RookPhaseRoundEnd)
	g2.GetPlayer(1).SetPoints(60) // team 1 captured only 60 < 100
	g2.GetPlayer(0).SetPoints(120)
	g2.ScoreRound()
	if g2.GetTeamScore(1) != -100 {
		t.Errorf("set: team1 = %d, want -100", g2.GetTeamScore(1))
	}
	if g2.GetTeamScore(0) != 120 {
		t.Errorf("set: team0 = %d, want 120", g2.GetTeamScore(0))
	}
}

func TestRookGameEnd(t *testing.T) {
	g := rookNewGame()
	g.SetDeclarerIdx(0)
	g.SetContractBid(70)
	g.SetTeamScore(0, 450)
	g.SetPhase(domain.RookPhaseRoundEnd)
	g.GetPlayer(0).SetPoints(100) // pushes team 0 to 550 >= 500
	g.ScoreRound()
	if !g.GetGameEndFlag() {
		t.Fatalf("game should have ended")
	}
	if g.GetWinnerTeam() != 0 {
		t.Errorf("winner = %d, want 0", g.GetWinnerTeam())
	}
}

func TestRookNestPointsToLastTrick(t *testing.T) {
	// Discard 5 known point cards, then verify the last-trick winner is
	// credited the nest points on top of the trick's captured points.
	g := rookHumanDeclarerExchange()
	if err := g.PlayerExchangeNest([]int{0, 1, 2, 3, 4}, 3); err != nil {
		t.Fatalf("exchange: %v", err)
	}
	nestPts := g.GetNestPoints()
	if nestPts != 45 {
		t.Errorf("nest points = %d, want 45", nestPts)
	}

	// Resolve the final trick and confirm the winner gets the nest points.
	g.SetTrumpColor(4)
	g.SetLeadPlayerIdx(0)
	g.SetTrickNumber(domain.RookTrickCnt)
	before := g.GetPlayer(1).GetPoints()
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: rookCard(1, 8)},
		{PlayerIdx: 1, Card: rookCard(1, 1)}, // wins the led suit
		{PlayerIdx: 2, Card: rookCard(1, 7)},
		{PlayerIdx: 3, Card: rookCard(2, 9)},
	})
	g.SetPhase(domain.RookPhaseTrickEnd)
	g.ResolveTrick()
	gained := g.GetPlayer(1).GetPoints() - before
	if gained < nestPts {
		t.Errorf("last-trick winner gained %d, expected at least nest %d", gained, nestPts)
	}
	if g.GetPhase() != domain.RookPhaseRoundEnd {
		t.Errorf("after last trick should be round end")
	}
}

func TestRookCpuHelpers(t *testing.T) {
	g := rookNewGame()
	g.SetPhase(domain.RookPhaseBid)
	g.SetBidPlayerIdx(1)
	// Give CPU a strong hand so it bids.
	p := g.GetPlayer(1)
	p.Reset()
	for _, c := range []*domain.Card{
		rookBird(), rookCard(1, 1), rookCard(1, 14), rookCard(1, 13),
		rookCard(1, 5), rookCard(1, 10), rookCard(2, 1), rookCard(2, 14),
		rookCard(3, 1), rookCard(4, 1), rookCard(1, 12), rookCard(1, 11), rookCard(1, 9),
	} {
		p.AddCard(c)
	}
	g.CpuBid()
	// After a CPU bid or pass, the bid turn advanced.
	if g.GetPhase() != domain.RookPhaseBid && g.GetPhase() != domain.RookPhaseNestExchange {
		t.Errorf("unexpected phase after cpu bid")
	}
}

func TestRookFullGameViaDomain(t *testing.T) {
	g := rookNewGame()
	g.Reset()
	const maxSteps = 200000
	steps := 0
	for !g.GetGameEndFlag() && steps < maxSteps {
		steps++
		switch g.GetPhase() {
		case domain.RookPhaseBid:
			if g.IsHumanBidTurn() {
				_ = g.PlayerPass() // human always passes; CPUs carry the auction
			} else {
				g.CpuBid()
			}
		case domain.RookPhaseNestExchange:
			g.CpuExchange() // only reached if a CPU is declarer
		case domain.RookPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				if len(valid) == 0 {
					t.Fatalf("human has no valid plays")
				}
				_ = g.PlayerPlay(valid[0])
			} else {
				g.CpuPlay()
			}
		case domain.RookPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == domain.RookPhaseTrickEnd {
				g.NextTrick()
			}
		case domain.RookPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		default:
			t.Fatalf("unexpected phase %d", g.GetPhase())
		}
	}
	if !g.GetGameEndFlag() {
		t.Fatalf("game did not finish within %d steps", maxSteps)
	}
	if g.GetWinnerTeam() < 0 {
		t.Errorf("winner team not set")
	}
}

func TestRookJSONRoundTrip(t *testing.T) {
	g := rookNewGame()
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored domain.Rook
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetPlayerCnt() != domain.RookPlayerCnt {
		t.Errorf("restored player count = %d", restored.GetPlayerCnt())
	}
	if restored.GetPhase() != g.GetPhase() {
		t.Errorf("phase mismatch")
	}
}

func TestRookUnmarshalValidation(t *testing.T) {
	cases := []string{
		`{"ps":[]}`,                    // wrong player count
		`{"ps":[null,null,null,null]}`, // nil player
		`{"ps":[` + rookOnePlayer(0) + `,` + rookOnePlayer(1) + `,` + rookOnePlayer(0) + `,` + rookOnePlayer(1) + `],"cf":{"cd":99,"ts":0}}`, // invalid config
	}
	for i, c := range cases {
		var g domain.Rook
		if err := json.Unmarshal([]byte(c), &g); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}

	// Invalid deck card value.
	bad := `{"ps":[` + rookOnePlayer(0) + `,` + rookOnePlayer(1) + `,` + rookOnePlayer(0) + `,` + rookOnePlayer(1) + `],"cf":{"cd":1,"ts":500},"dk":[{"d":1,"v":99,"w":false}]}`
	var g domain.Rook
	if err := json.Unmarshal([]byte(bad), &g); err == nil {
		t.Errorf("expected invalid deck card error")
	}
}

func rookOnePlayer(team int) string {
	p := domain.NewRookPlayer(false, team)
	b, _ := json.Marshal(p)
	return string(b)
}

func TestRookConfigValidate(t *testing.T) {
	if err := domain.DefaultRookConfig().Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
	if err := (domain.RookConfig{CpuDifficulty: 99, TargetScore: 500}).Validate(); err == nil {
		t.Errorf("bad difficulty should fail")
	}
	if err := (domain.RookConfig{CpuDifficulty: 1, TargetScore: 0}).Validate(); err == nil {
		t.Errorf("bad target should fail")
	}
}

func TestRookNextRoundGuards(t *testing.T) {
	g := rookNewGame()
	g.Reset()
	// NextRound only advances from RoundEnd.
	round := g.GetRoundNumber()
	g.NextRound()
	if g.GetRoundNumber() != round {
		t.Errorf("NextRound should be a no-op outside RoundEnd")
	}
	// GetHint should not panic in bid phase.
	g.SetBidPlayerIdx(0)
	_ = g.GetHint()
}
