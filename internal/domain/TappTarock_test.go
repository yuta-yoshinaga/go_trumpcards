//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTappTarockDealUsesThreeHandsOfSixteenAndSixCardTalon(t *testing.T) {
	g := NewDefaultTappTarock()
	g.Reset()
	if g.GetPlayerCnt() != TappTarockPlayerCnt || TappTarockPlayerCnt != 3 {
		t.Fatalf("expected three players, got %d", g.GetPlayerCnt())
	}
	for i := 0; i < TappTarockPlayerCnt; i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != TappTarockHandSize {
			t.Errorf("player %d has %d cards", i, got)
		}
	}
	if got := g.GetTalonSize(); got != TappTarockTalonSize {
		t.Errorf("talon has %d cards", got)
	}
}

func TestTappTarockPlaysSixteenTricks(t *testing.T) {
	players := []*TappTarockPlayer{NewTappTarockPlayer(true), NewTappTarockPlayer(false), NewTappTarockPlayer(false)}
	g := NewTappTarock(players, DefaultTappTarockConfig())
	deck := buildKoenigrufenDeck()
	for i, c := range deck[:TappTarockPlayerCnt*TappTarockHandSize] {
		players[i%TappTarockPlayerCnt].AddCard(c)
	}
	g.startPlay()
	for trick := 0; trick < TappTarockTrickCount && g.GetPhase() == TappTarockPhasePlay; trick++ {
		for seat := 0; seat < TappTarockPlayerCnt; seat++ {
			idxs := g.GetValidPlayIndices(g.currentPlayerIdx)
			if len(idxs) == 0 {
				t.Fatalf("no legal card at trick %d seat %d", trick, seat)
			}
			g.playCard(g.currentPlayerIdx, idxs[0])
		}
		g.NextTrick()
	}
	totalTricks := 0
	for i := 0; i < TappTarockPlayerCnt; i++ {
		totalTricks += g.GetPlayer(i).GetTrickCount()
	}
	for i := 0; i < TappTarockPlayerCnt; i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != 0 {
			t.Fatalf("player %d still has %d cards after the round", i, got)
		}
	}
	if totalTricks != TappTarockTrickCount || g.GetPhase() != TappTarockPhaseRoundEnd {
		t.Fatalf("expected %d completed tricks and round end, got %d/%d", TappTarockTrickCount, totalTricks, g.GetPhase())
	}
}

func TestTappTarockTrullBonusRequiresAllThreeTrullCards(t *testing.T) {
	g := NewDefaultTappTarock()
	trull := []*Card{
		NewCard(KoenigrufenTrumpDesign, KoenigrufenPagatValue, false),
		NewCard(KoenigrufenTrumpDesign, KoenigrufenMaxTrump, false),
		NewCard(KoenigrufenSkusDesign, KoenigrufenSkusValue, false),
	}
	for _, c := range trull {
		g.players[0].AddTrick([]*Card{c})
	}
	if got, want := g.GetCardPoints(0), 15+TappTarockTrullBonus; got != want {
		t.Fatalf("trull score = %d, want %d", got, want)
	}
}

func TestTappTarockContractOutcomeChangesWithDeclarerMajority(t *testing.T) {
	winning := NewDefaultTappTarock()
	winning.contract = TappTarockBidDreier
	winning.declarerIdx = 0
	winning.players[0].AddTrick(buildKoenigrufenDeck())
	if breakdown := winning.scoreContract(); !breakdown.Won || breakdown.Seats[0] <= 0 {
		t.Fatalf("expected declarer majority to win, got %+v", breakdown)
	}

	losing := NewDefaultTappTarock()
	losing.contract = TappTarockBidDreier
	losing.declarerIdx = 0
	losing.players[1].AddTrick(buildKoenigrufenDeck())
	if breakdown := losing.scoreContract(); breakdown.Won || breakdown.Seats[0] >= 0 {
		t.Fatalf("expected declarer minority to lose, got %+v", breakdown)
	}
}

func tappTarockPlayers() []*TappTarockPlayer {
	return []*TappTarockPlayer{NewTappTarockPlayer(true), NewTappTarockPlayer(false), NewTappTarockPlayer(false)}
}

func newTappTarockTestGame(t *testing.T) *TappTarock {
	t.Helper()
	g := NewTappTarock(tappTarockPlayers(), TappTarockConfig{TargetDeals: 1})
	g.Reset()
	return g
}

func tappTarockAtTalon(t *testing.T, bid TappTarockBid) *TappTarock {
	t.Helper()
	g := newTappTarockTestGame(t)
	g.highestBid, g.highestBidder = bid, 0
	g.finalizeBid()
	require.Equal(t, TappTarockPhaseTalon, g.phase)
	return g
}

func tappTarockAtPlay(t *testing.T) *TappTarock {
	t.Helper()
	g := tappTarockAtTalon(t, TappTarockBidDreier)
	// Arrange the real dealer state so startPlay makes the human lead naturally.
	g.dealerIdx = 2
	indices := g.GetDiscardableIndices()[:TappTarockTalonSize]
	require.NoError(t, g.PlayerDiscard(indices))
	return g
}

func tappTarockStep(t *testing.T, g *TappTarock) {
	t.Helper()
	switch g.phase {
	case TappTarockPhaseBid:
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerPass())
		} else {
			g.CpuBid()
		}
	case TappTarockPhaseTalon:
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerDiscard(g.GetDiscardableIndices()[:6]))
		} else {
			g.CpuDiscard()
		}
	case TappTarockPhasePlay:
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NotNil(t, h)
			require.NoError(t, g.PlayerPlayCard(*h.CardIndex))
		} else {
			g.CpuPlayCard()
		}
	case TappTarockPhaseTrickEnd:
		g.NextTrick()
	case TappTarockPhaseRoundEnd:
		g.NextRound()
	}
}

func TestTappTarockConfigValidateAndAccessors(t *testing.T) {
	for _, tc := range []struct {
		name string
		c    TappTarockConfig
		ok   bool
	}{
		{"default", DefaultTappTarockConfig(), true}, {"min", TappTarockConfig{TargetDeals: 1}, true},
		{"max", TappTarockConfig{TargetDeals: 12}, true}, {"zero", TappTarockConfig{TargetDeals: 0}, false},
		{"too many", TappTarockConfig{TargetDeals: 13}, false}, {"difficulty", TappTarockConfig{CpuDifficulty: 9, TargetDeals: 1}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ok {
				assert.NoError(t, tc.c.Validate())
			} else {
				assert.Error(t, tc.c.Validate())
			}
		})
	}
	g := newTappTarockTestGame(t)
	c := TappTarockConfig{TargetDeals: 2}
	g.SetConfig(c)
	assert.Equal(t, c, g.GetConfig())
	assert.Len(t, g.GetPlayers(), 3)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(3))
	assert.Equal(t, 0, g.GetPlayerScore(-1))
	assert.Equal(t, 0, g.GetCardPoints(99))
	assert.Equal(t, 0, g.HumanSeat())
}

func TestTappTarockBiddingContractsAndPass(t *testing.T) {
	g := newTappTarockTestGame(t)
	g.bidPlayerIdx, g.currentPlayerIdx = 0, 0
	for !g.IsHumanTurn() {
		g.CpuBid()
	}
	assert.Error(t, g.PlayerBid(TappTarockBidPass))
	assert.Error(t, g.PlayerBid(TappTarockBidTrischaken))
	require.NoError(t, g.PlayerBid(TappTarockBidDreier))
	assert.Equal(t, TappTarockBidDreier, g.GetHighestBid())
	assert.Error(t, g.PlayerBid(TappTarockBidDreier))
	g = newTappTarockTestGame(t)
	for g.phase == TappTarockPhaseBid {
		g.applyPass(g.bidPlayerIdx)
	}
	assert.Equal(t, TappTarockBidTrischaken, g.contract)
	assert.Equal(t, -1, g.declarerIdx)
	assert.Len(t, g.stash, 6)
	assert.Empty(t, g.talon)
	g = newTappTarockTestGame(t)
	g.highestBid, g.highestBidder = TappTarockBidSolo, 0
	g.finalizeBid()
	assert.Equal(t, TappTarockPhasePlay, g.phase)
	assert.Equal(t, 0, g.declarerIdx)
	assert.Equal(t, 1, g.stashOwner)
	assert.Len(t, g.stash, 6)
}

func TestTappTarockDiscardValidationAndCPUSelection(t *testing.T) {
	g := tappTarockAtTalon(t, TappTarockBidDreier)
	p := g.players[0]
	p.Reset()
	for _, c := range []*Card{NewCard(1, 1, false), NewCard(1, KoenigrufenKingValue, false), NewCard(KoenigrufenTrumpDesign, 1, false), NewCard(KoenigrufenTrumpDesign, 10, false), NewCard(KoenigrufenTrumpDesign, KoenigrufenMaxTrump, false), NewCard(KoenigrufenSkusDesign, KoenigrufenSkusValue, false), NewCard(2, 2, false)} {
		p.AddCard(c)
	}
	assert.Error(t, g.PlayerDiscard([]int{0}))
	assert.Error(t, g.PlayerDiscard([]int{0, 0, 1, 2, 3, 4}))
	assert.Error(t, g.PlayerDiscard([]int{0, 1, 2, 3, 4, 99}))
	assert.ErrorContains(t, g.PlayerDiscard([]int{1, 2, 3, 4, 5, 6}), "king")
	assert.Equal(t, []int{0, 3, 6}, g.GetDiscardableIndices())
	assert.Empty(t, newTappTarockTestGame(t).GetDiscardableIndices())
	// CPU's fallback also handles a hand made entirely of protected cards.
	g.players[1].Reset()
	for i := 1; i <= 7; i++ {
		g.players[1].AddCard(NewCard(KoenigrufenTrumpDesign, i, false))
	}
	assert.Len(t, g.cpuSelectDiscards(1), 6)
}

func TestTappTarockValidPlaysWinnerAndIllegalActions(t *testing.T) {
	g := tappTarockAtPlay(t)
	lead := g.currentPlayerIdx
	assert.Len(t, g.GetValidPlayIndices(lead), g.players[lead].GetCardsSize())
	assert.Nil(t, g.GetValidPlayIndices(-1))
	g.players[(lead+1)%3].Reset()
	g.players[(lead+1)%3].AddCard(NewCard(KoenigrufenTrumpDesign, 5, false))
	g.players[(lead+1)%3].AddCard(NewCard(2, 3, false))
	g.players[lead].Reset()
	g.players[lead].AddCard(NewCard(1, 4, false))
	g.currentPlayerIdx = lead
	g.playCard(lead, 0)
	valid := g.GetValidPlayIndices((lead + 1) % 3)
	require.Len(t, valid, 1)
	assert.True(t, koenigrufenIsTrumpLike(g.players[(lead+1)%3].GetCard(valid[0])))
	assert.Error(t, g.PlayerPlayCard(99))
	g.phase = TappTarockPhaseBid
	assert.Error(t, g.PlayerPlayCard(0))
	g.phase = TappTarockPhasePlay
	g.currentPlayerIdx = 1
	assert.ErrorIs(t, g.PlayerPlayCard(0), ErrNotHumanTurn)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPlayCard(0), ErrGameEnded)

	g = tappTarockAtPlay(t)
	for i := 0; i < 3; i++ {
		g.playCard(g.currentPlayerIdx, g.GetValidPlayIndices(g.currentPlayerIdx)[0])
	}
	assert.Equal(t, TappTarockPhaseTrickEnd, g.phase)
	assert.Len(t, g.lastTrickCards, 3)
	g.NextTrick()
	assert.Equal(t, 2, g.trickNumber)
}

func TestTappTarockPlaysAllSixteenTricksFromStartPlay(t *testing.T) {
	g := NewTappTarock(tappTarockPlayers(), TappTarockConfig{TargetDeals: 1})
	deck := buildKoenigrufenDeck()
	for i, c := range deck[:48] {
		g.players[i%3].AddCard(c)
	}
	g.startPlay()
	for steps := 0; steps < 100 && (g.phase == TappTarockPhasePlay || g.phase == TappTarockPhaseTrickEnd); steps++ {
		if g.phase == TappTarockPhasePlay {
			valid := g.GetValidPlayIndices(g.currentPlayerIdx)
			require.NotEmpty(t, valid)
			g.playCard(g.currentPlayerIdx, valid[0])
		} else {
			g.NextTrick()
		}
	}
	assert.Equal(t, TappTarockPhaseRoundEnd, g.phase)
	assert.Equal(t, 16, g.trickNumber)
	for _, p := range g.players {
		assert.Equal(t, 0, p.GetCardsSize())
	}
}

func TestTappTarockScoringTrischakenStashAndGameEnd(t *testing.T) {
	g := newTappTarockTestGame(t)
	g.contract = TappTarockBidTrischaken
	g.declarerIdx = -1
	g.stash = append([]*Card(nil), g.talon...)
	g.talon = nil
	g.stashOwner = -1
	g.lastTrickWinner = 2
	g.players[1].AddTrick([]*Card{NewCard(1, KoenigrufenKingValue, false)})
	bd := g.scoreTrischaken()
	assert.Equal(t, 1, bd.Loser)
	assert.Equal(t, -3, bd.Seats[1])
	g.assignStash()
	assert.Empty(t, g.stash)
	g.phase = TappTarockPhaseTrickEnd
	g.trickNumber = 16
	g.finishRound()
	assert.Equal(t, TappTarockPhaseGameEnd, g.phase)
	assert.True(t, g.gameEndFlag)
	assert.Equal(t, -1, g.winnerPlayer)
}

func TestTappTarockHintsAndJSONRoundTrip(t *testing.T) {
	g := tappTarockAtPlay(t)
	h := g.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	b, err := json.Marshal(g)
	require.NoError(t, err)
	var back TappTarock
	require.NoError(t, json.Unmarshal(b, &back))
	assert.Equal(t, g.GetPhase(), back.GetPhase())
	assert.Equal(t, g.GetTrickNumber(), back.GetTrickNumber())
	require.NotNil(t, back.GetPlayers())
	tappTarockStep(t, &back)
	var p TappTarockPlayer
	original := NewTappTarockPlayer(true)
	original.AddCard(NewCard(1, 1, false))
	original.AddTrick([]*Card{NewCard(1, KoenigrufenKingValue, false)})
	pb, _ := json.Marshal(original)
	require.NoError(t, json.Unmarshal(pb, &p))
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 1, p.GetCardsSize())
	assert.Equal(t, 1, p.GetTrickCount())
	assert.Equal(t, "trischaken", TappTarockBidName(TappTarockBidTrischaken))
	assert.Equal(t, "dreier", TappTarockBidName(TappTarockBidDreier))
	assert.Equal(t, "solo", TappTarockBidName(TappTarockBidSolo))
	assert.Equal(t, "pass", TappTarockBidName(TappTarockBidPass))
}

func TestTappTarockUnmarshalRejectsInvalidState(t *testing.T) {
	g := tappTarockAtPlay(t)
	raw, err := json.Marshal(g)
	require.NoError(t, err)
	for _, tc := range []struct {
		k   string
		v   any
		msg string
	}{
		{"ph", 9, "invalid phase"}, {"cp", 3, "current player"}, {"di", -1, "dealer"}, {"dc", 3, "declarer"}, {"rn", 0, "round"}, {"tn", 17, "trick"}, {"co", 9, "invalid contract"},
	} {
		t.Run(tc.k, func(t *testing.T) {
			var m map[string]any
			require.NoError(t, json.Unmarshal(raw, &m))
			m[tc.k] = tc.v
			b, _ := json.Marshal(m)
			var out TappTarock
			assert.ErrorContains(t, json.Unmarshal(b, &out), tc.msg)
		})
	}
	var out TappTarock
	assert.Error(t, json.Unmarshal([]byte(`{"ph":"bad"}`), &out))
}

func TestTappTarockCPUPathsRoundEndAndGetters(t *testing.T) {
	g := newTappTarockTestGame(t)
	g.bidPlayerIdx = 1
	g.players[1].Reset()
	for i := 1; i <= 9; i++ {
		g.players[1].AddCard(NewCard(KoenigrufenTrumpDesign, i, false))
	}
	g.players[1].AddCard(NewCard(KoenigrufenTrumpDesign, KoenigrufenMaxTrump, false))
	g.players[1].AddCard(NewCard(KoenigrufenSkusDesign, KoenigrufenSkusValue, false))
	for i := 1; i <= 3; i++ {
		g.players[1].AddCard(NewCard(i, KoenigrufenKingValue, false))
	}
	g.highestBid = TappTarockBidPass
	g.CpuBid()
	assert.Equal(t, TappTarockBidSolo, g.highestBid)
	g = newTappTarockTestGame(t)
	g.highestBid, g.highestBidder = TappTarockBidDreier, 1
	g.finalizeBid()
	g.CpuDiscard()
	assert.Equal(t, TappTarockPhasePlay, g.phase)
	g = tappTarockAtPlay(t)
	g.currentPlayerIdx = 1
	g.CpuPlayCard()
	assert.Len(t, g.currentTrick, 1)
	assert.Equal(t, g.roundNumber, g.GetRoundNumber())
	assert.Equal(t, g.trickNumber, g.GetTrickNumber())
	assert.Equal(t, g.currentPlayerIdx, g.GetCurrentPlayerIdx())
	assert.Equal(t, g.dealerIdx, g.GetDealerIdx())
	assert.Equal(t, g.bidPlayerIdx, g.GetBidPlayerIdx())
	assert.Equal(t, g.contract, g.GetContract())
	assert.Equal(t, g.declarerIdx, g.GetDeclarerIdx())
	assert.Equal(t, g.lastTrickWinner, g.GetLastTrickWinner())
	assert.Equal(t, g.lastTrickCards, g.GetLastTrickCards())
	assert.Equal(t, g.outcome, g.GetOutcome())
	assert.Equal(t, g.breakdown, g.GetBreakdown())
	assert.Equal(t, g.gameEndFlag, g.GetGameEndFlag())
	assert.Equal(t, g.winnerPlayer, g.GetWinnerPlayer())
	g.phase = TappTarockPhaseRoundEnd
	g.gameEndFlag = false
	g.NextRound()
	assert.Equal(t, 2, g.roundNumber)
	g.gameEndFlag = true
	g.NextRound()
	assert.Equal(t, 2, g.roundNumber)
}

func TestTappTarockHintBranchesAndValidationHelpers(t *testing.T) {
	g := newTappTarockTestGame(t)
	g.bidPlayerIdx = 0
	g.currentPlayerIdx = 0
	g.players[0].Reset()
	g.players[0].AddCard(NewCard(1, 1, false))
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "pass_weak_hand", h.Reason)
	g.players[0].Reset()
	for i := 1; i <= 10; i++ {
		g.players[0].AddCard(NewCard(KoenigrufenTrumpDesign, i, false))
	}
	h = g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "bid_strong_trumps", h.Reason)
	g = tappTarockAtPlay(t)
	g.contract = TappTarockBidTrischaken
	g.currentPlayerIdx = 0
	h = g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "avoid_points", h.Reason)
	g.gameEndFlag = true
	assert.Nil(t, g.GetHint())
	assert.Error(t, tapptarockValidateContract(&tapptarockJSON{Contract: 9}))
	assert.Error(t, tapptarockValidateContract(&tapptarockJSON{Contract: TappTarockBidTrischaken, DeclarerIdx: 0}))
	assert.Error(t, tapptarockValidateCards([]*Card{nil}))
	assert.Error(t, tapptarockValidateCards([]*Card{NewCard(9, 1, false)}))
}

func TestTappTarockPassAndEmptyJSONBranches(t *testing.T) {
	g := newTappTarockTestGame(t)
	g.bidPlayerIdx, g.currentPlayerIdx = 0, 0
	require.NoError(t, g.PlayerPass())
	assert.Empty(t, g.GetCurrentTrick())
	g.phase = TappTarockPhasePlay
	assert.Error(t, g.PlayerPass())
	var p TappTarockPlayer
	require.NoError(t, json.Unmarshal([]byte(`{"gp":null,"th":null}`), &p))
	assert.False(t, p.GetIsHuman())
	assert.Error(t, json.Unmarshal([]byte(`{"gp":`), &p))
	g = newTappTarockTestGame(t)
	g.SetConfig(TappTarockConfig{TargetDeals: 0})
	g.Reset()
	assert.Equal(t, TappTarockDefaultDeals, g.GetConfig().TargetDeals)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPass(), ErrGameEnded)
	g.phase = TappTarockPhasePlay
	g.CpuBid()
	g.CpuDiscard()
	g.CpuPlayCard()
}
